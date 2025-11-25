//go:build !disable_assertions

/*
Package invariant provides a property testing framework for Go, designed to
encode and enforce assumptions (invariants) directly within your code.

# Assertion Types

   - **Always** — The property must hold true for *all* inputs or executions.
     To disprove an Always assertion, you only need *one counterexample* where
     the condition is false.

   - **Sometimes** — The property is expected to be *occasionally true* across
     runs. To prove a Sometimes assertion, you need *one example* where the
     condition is true. If it never evaluates to true, the property is disproven.
     Sometimes assertions are only meaningful in testing environments, since
     observing their absence requires multiple executions of the program.


# Implementation

In production, assertions immediately crash the program on violation.
In test environments, invariant activates a global tracker that records every
assertion that evaluated to true.

During setup, you register the packages to analyze (usually just the current
test’s target package). As tests execute, successful assertions (true
evaluations) are recorded in a package-global tracker keyed by file and line
number. After all tests finish, the analyzer then parses all Go files in the
registered packages and locates every assertion in the source. It’s crucial that
the package qualifier name remains `invariant`, since the parser relies on this
identifier to detect assertions.

Next, the analyzer cross-references parsed assertions with those observed at
runtime. Any assertion that never evaluated to true (frequency = 0) is reported
as a “missed invariant,” causing the test suite to fail. If all assertions were
exercised, the analyzer prints a summary showing up to 20 of the *least
exercised* assertions.

Invariant therefore provides actionable, frequency-based insight into how
thoroughly your properties have been exercised, revealing the true scope and
effectiveness of your testing suite. By tracking which invariants fire and
which remain untested, it exposes gaps that ordinary unit tests alone would
never highlight, giving a far more comprehensive view of software reliability
than coverage metrics or ad-hoc tests can offer. Additionally, by making the
assumptions of your software explicit, it improves developer experience and
accelerates onboarding, helping new contributors understand the intended
behavior and constraints of the system.
*/

package invariant

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// AssertionFailureCallback lets you override the default hard assertion
// behavior (which crashes the program on failure) with custom logic. Assigning
// a non-crashing callback allows users to handle assertion failures gracefully,
// for example by logging or recovering in long-running applications like web
// servers. See examples/backend for usage.
var (
	AssertionFailureCallback = DefaultAssertionFailureCallbackFatal
	/*
		Used to detect panics caused by assertion failures
		defer func() {
			if err := recover(); err != nil {
				if strErr, ok := err.(string); ok && strings.HasPrefix(strErr, invariant.AssertionFailureMsgPrefix) {
					// handle assertion failure
				}
			}
		}()

	*/
	AssertionFailureMsgPrefix = "🚨 Assertion Failure 🚨"

	DefaultAssertionFailureCallbackFatal = func(msg string) {
		FprintStackTrace(os.Stderr, 1)
		fmt.Fprintln(os.Stderr, msg)
		os.Exit(1)
	}

	DefaultAssertionFailureCallbackPanic = func(msg string) {
		panic(msg)
	}

	// packagesToAnalyze defaults to the current testing package.
	packagesToAnalyze = []string{"."}
	// assertionTracker globally tracks true assertions inside packagesToAnalyze.
	assertionTracker        = make(map[string]*metadata, maxAssertionsPerPackage*len(packagesToAnalyze))
	assertionFrequencyMutex = sync.Mutex{}
)

type metadata struct {
	Frequency int
	Message   string
	Kind      string
}

// registerAssertion records in the package-global assertion tracker that an
// assertion was evaluated as true at the current source location. It tracks the
// file and line number of the call and increments a counter in
// assertionTracker.
//
// This function is a noop outside of a test environment.
//
// It is concurrency-safe and can be called from multiple goroutines.
//
//go:noinline
func registerAssertion(kind, msg string) {
	if IsRunningUnderGoBenchmark || IsRunningUnderGoFuzz {
		return
	}
	if msg == "" {
		msg = "<empty>"
	}
	callers := [1]uintptr{}
	count := runtime.Callers(3, callers[:])
	frame, _ := runtime.CallersFrames(callers[:count]).Next()

	arr := [assertionIDLength]byte{}
	buf := arr[:0]
	buf = append(buf, frame.File...)
	buf = append(buf, ':')
	buf = strconv.AppendInt(buf, int64(frame.Line), 10)

	id := string(buf)
	assertionFrequencyMutex.Lock()
	a, ok := assertionTracker[id]
	if ok {
		a.Frequency++
	}
	assertionFrequencyMutex.Unlock()
}

// RegisterPackagesForAnalysis ensures that only assertions from the tested
// directories are tracked for frequency analysis. Dirs is relative to the
// directory of the caller.
func RegisterPackagesForAnalysis(dirs ...string) {
	Always(IsRunningUnderGoTest, "RegisterPackagesForAnalysis is only used in testing environments")
	Always(len(packagesToAnalyze) == 1 && packagesToAnalyze[0] == ".", "packagesToAnalyze was set to the current testing package by default")
	if IsRunningUnderGoBenchmark {
		return
	}
	if len(dirs) > 0 {
		packagesToAnalyze = dirs
	}
	// === Absolute Path Conversion ===
	for i, path := range packagesToAnalyze {
		path, err := filepath.Abs(path)
		if err != nil || path == "" {
			panic(fmt.Sprintf("Failed to convert package path to absolute path: %s\n", err))
		}
		packagesToAnalyze[i] = path
	}

	// ===Collection===
	filesArray := [maxGoFilesPerPackage]string{}
	files := filesArray[:0]
	for _, dir := range packagesToAnalyze {
		before := len(files)
		err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() || filepath.Ext(path) != ".go" {
				return err
			}
			path, err = filepath.Abs(path)
			if err != nil {
				return err
			}
			if len(path) > len("_test.go") && strings.HasSuffix(path, "_test.go") {
				return nil
			}
			files = append(files, path)
			return nil
		})
		if err != nil {
			panic(fmt.Sprintf("Collecting all files to find missed invariants: %s\n", err))
		}
		after := len(files)
		Always(before < after, "The directory contains go files")
	}
	Always(len(files) > 0, "There's at least one file to parse")

	// ===Parsing===
	semaphore := make(chan struct{}, runtime.NumCPU())
	var wg sync.WaitGroup

	for _, path := range files {
		wg.Add(1)
		semaphore <- struct{}{}
		go func(path string) {
			defer func() { <-semaphore; wg.Done() }()
			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
			if err != nil {
				return
			}
			ast.Inspect(node, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				ident, ok := sel.X.(*ast.Ident)
				if !ok {
					return true
				}
				if ident.Name != "invariant" {
					return true
				}
				msg := "<empty>"
				switch sel.Sel.Name {
				case "Sometimes", "XSometimes", "Always", "AlwaysNil", "AlwaysErrIs", "AlwaysErrIsNot",
					"XAlways", "XAlwaysNil", "XAlwaysErrIs", "XAlwaysErrIsNot":
					pos := fset.Position(call.Lparen)
					key := path + ":" + strconv.Itoa(pos.Line)
					Always(len(call.Args) >= 2, "All of the matched assertions have at least two parameters")
					if literal, ok := call.Args[1].(*ast.BasicLit); ok && literal.Kind == token.STRING {
						msg = literal.Value[1 : len(literal.Value)-1] // remove quotes
					}
					assertionTracker[key] = &metadata{
						Kind:    sel.Sel.Name,
						Message: msg,
					}
				}
				return true
			})
		}(path)
	}

	wg.Wait()
}

// AnalyzeAssertionFrequency scans the given directories for Sometimes, Always*,
// XAlways* calls that have never evaluated to true and returns their source
// locations and respective messages. In your *_test.go, add the following
// snippet:
//
//	func TestMain(m *testing.M) {
//		invariant.RegisterPackagesForAnalysis()
//		code := m.Run()
//		if code == 0 {
//			invariant.AnalyzeAssertionFrequency()
//		}
//		os.Exit(code)
//	}
//
// Then run your tests with the `-v` flag so you can see the frequency analysis
// printed at the end: `go test ./mypackage -v`
//
// It is critical that you import this package under the name "invariant" as it
// is hardcoded in the analyzer to look for this identifier.
func AnalyzeAssertionFrequency() {
	Always(IsRunningUnderGoTest, "AnalyzeAssertionFrequency is only used for testing")
	Always(len(packagesToAnalyze) > 0, "At least one package was registered for analysis")
	if IsRunningUnderGoBenchmark || IsRunningUnderGoFuzz {
		return
	}

	longestKindWord := 0
	longestMessageLength := 0
	missed := make([]string, 0, len(assertionTracker))
	for location, metadata := range assertionTracker {
		Always(location != "", "All assertion records have a location")
		if metadata.Frequency == 0 {
			longestKindWord = max(longestKindWord, len(metadata.Kind))
			longestMessageLength = max(longestMessageLength, len(metadata.Message))
			missed = append(missed, location)
		}
	}
	if len(missed) > 0 {
		fmt.Printf("🚨 %d assertions were never true. 🚨\n", len(missed))
		for _, id := range missed {
			fmt.Printf(
				"\t%*s | %-*s | %s\n",
				longestKindWord, assertionTracker[id].Kind,
				longestMessageLength, assertionTracker[id].Message,
				id,
			)
		}
		os.Exit(1)
	}

	// ===Analysis===
	//
	// This is a very simple and “dumb” analysis based solely on absolute
	// assertion frequency. It does not attempt to correlate assertions with
	// specific unit tests or cluster them with each other; any attempt to
	// do so is futile. For that kind of analysis, you’d need
	// instrumentation/static analysis combined with machine learning.
	{
		type scored struct {
			key   string
			count int
		}
		const n = 20
		fmt.Printf("Showing up to %d of the least-exercised invariants:\n", n)

		longestMessageLength := 0
		longestKindWord := 0

		h := make([]scored, 0, n)
		for key, assertion := range assertionTracker {
			Always(key != "", "Assertion location must not be empty")
			longestMessageLength = max(longestMessageLength, len(assertion.Message))
			longestKindWord = max(longestKindWord, len(assertion.Kind))

			if len(h) < n {
				h = append(h, scored{key, assertion.Frequency})
				if len(h) == n {
					sort.Slice(h, func(i, j int) bool {
						return h[i].count > h[j].count
					})
				}
				continue
			}
			if assertion.Frequency < h[0].count {
				h[0] = scored{key, assertion.Frequency}
				i := 0
				for {
					l, r := 2*i+1, 2*i+2
					if l >= n {
						break
					}
					small := l
					if r < n && h[r].count > h[l].count {
						small = r
					}
					if h[i].count >= h[small].count {
						break
					}
					h[i], h[small] = h[small], h[i]
					i = small
				}
			}
		}

		sort.Slice(h, func(i, j int) bool {
			return h[i].count < h[j].count
		})

		for _, v := range h {
			a := assertionTracker[v.key]
			fmt.Printf(
				"count=%-4d | %-*s | %-*s | %s\n",
				a.Frequency,
				longestKindWord, a.Kind,
				longestMessageLength, a.Message,
				v.key,
			)
		}
	}
}

//go:noinline
func Unimplemented(msg string) {
	if msg == "" {
		msg = "<empty>"
	}
	AssertionFailureCallback(fmt.Sprintf("%s: %s\n", AssertionFailureMsgPrefix, msg))
}

//go:noinline
func Unreachable(msg string) {
	if msg == "" {
		msg = "<empty>"
	}
	AssertionFailureCallback(fmt.Sprintf("%s: %s\n", AssertionFailureMsgPrefix, msg))
}

// Always calls AssertionFailureCallback if cond is false. If you need
// Always(item == nil), use AlwaysNil(item) instead.
//
// Note 1: When deferring assertions, enclose them in a closure. Otherwise, cond
// is evaluated immediately.
//
//	// Correct: deferred assertion evaluates cond later
//	defer func() { invariant.Always(x > 0, "x must be positive") }()
//
//	// Incorrect: cond evaluated immediately
//	defer invariant.Always(x > 0, "x must be positive")
//
// Note 2: The analyzer tracks assertions by source line, so multiple assertions
// on the same line via loops are treated as a single assertion. However, its
// frequency counter increments on every evaluation. To avoid inflating
// frequency statistics, wrap the loop in a closure that returns a single bool,
// so the assertion is evaluated once for the aggregate condition.
//
//	// Aggregate check over a loop
//	invariant.Always(func() bool {
//	    for _, k := range keys {
//	        if k == forbidden {
//	            return false
//	        }
//	    }
//	    return true
//	}(), "No forbidden keys present")
//
//go:noinline
func Always(cond bool, msg string) {
	if cond {
		registerAssertion("Always", msg)
	} else {
		AssertionFailureCallback(fmt.Sprintf("%s: %s\n", AssertionFailureMsgPrefix, msg))
	}
}

// Sometimes records that a condition was true at least once throughout the test
// run. It only has an effect in test environments; outside of tests it is a
// no-op.
//
// Technically, the term "invariant.Sometimes" is a misnomer. It's more accurately
// described as a property check. But I wanted to name this library "invariant"
// instead of "assert" or "property" for better grep-ability... maybe also
// because it sounds cool.
//
//go:noinline
func Sometimes(ok bool, msg string) {
	if !IsRunningUnderGoTest || !ok {
		return
	}
	registerAssertion("Sometimes", msg)
}

// AlwaysNil calls AssertionFailureCallback if x is NOT nil and prints the
// non-null object. Prefer this over Always(x == nil) so that the value of x can
// be logged.
//
//go:noinline
func AlwaysNil(x interface{}, msg string) {
	if x == nil {
		registerAssertion("AlwaysNil", msg)
	} else {
		AssertionFailureCallback(fmt.Sprintf("%s: expected nil. got %v. %s\n", AssertionFailureMsgPrefix, x, msg))
	}
}

// AlwaysErrIs calls AssertionFailureCallback if actual is NOT one of the
// specified targets. Must provide at least one target. All targets must not be
// nil.
//
//go:noinline
func AlwaysErrIs(actual error, msg string, targets ...error) {
	Always(len(targets) > 0, "invariant.AlwaysErrIs requires at least one target")
	for _, t := range targets {
		Always(t != nil, "All invariant.AlwaysErrIs targets must not be nil")
		if errors.Is(actual, t) {
			registerAssertion("AlwaysErrIs", msg)
			return
		}
	}
	AssertionFailureCallback(fmt.Sprintf("%s: error did not match any targets. got %q. %s\n", AssertionFailureMsgPrefix, actual, msg))
}

// AlwaysErrIsNot calls AssertionFailureCallback if actual is one of the
// specified targets. Must provide at least one target. All targets must not be
// nil.
//
//go:noinline
func AlwaysErrIsNot(actual error, msg string, targets ...error) {
	Always(len(targets) > 0, "invariant.AlwaysErrIsNot() must have at least one target")
	for _, t := range targets {
		Always(t != nil, "invariant.AlwaysErrIsNot() targets must not be nil")
		if errors.Is(actual, t) {
			AssertionFailureCallback(fmt.Sprintf("%s: error unexpectedly matched a target. got %q. %s\n", AssertionFailureMsgPrefix, actual, msg))
			return
		}
	}
	registerAssertion("AlwaysErrIsNot", msg)
}

/*
XAlways evaluates fn and calls AssertionFailureCallback if it returns false. It
is designed for use cases where you want to perform expensive validations that
can be disabled in production builds using the `disable_assertions`
build tag.

	expensiveFn := func() bool { ... }
	// expensiveFn is still evaluated but boolean check is a noop under disable_assertions
	invariant.Always(expensiveFn())


	// expensiveFn itself will be a noop under disable_assertions
	invariant.XAlways(expensiveFn)

Be wary of this if you rely on side effects produced by fn. Rule of thumb would
be to ensure that fn is pure or idempotent.
*/
//go:noinline
func XAlways(fn func() bool, msg string) {
	if fn() {
		registerAssertion("XAlways", msg)
	} else {
		AssertionFailureCallback(fmt.Sprintf("%s: %s\n", AssertionFailureMsgPrefix, msg))
	}
}

func XSometimes(fn func() bool, msg string) {
	if !IsRunningUnderGoTest || !fn() {
		return
	}
	registerAssertion("Sometimes", msg)
}

// XAlwaysNil evaluates fn and calls AssertionFailureCallback if the result is not nil.
//
//go:noinline
func XAlwaysNil(fn func() interface{}, msg string) {
	x := fn()
	if x == nil {
		registerAssertion("XAlwaysNil", msg)
	} else {
		AssertionFailureCallback(fmt.Sprintf("%s: expected nil. got %v. %s\n", AssertionFailureMsgPrefix, x, msg))
	}
}

// XAlwaysErrIs evaluates fn and calls AssertionFailureCallback if the returned error is not in targets.
//
//go:noinline
func XAlwaysErrIs(fn func() error, msg string, targets ...error) {
	Always(len(targets) > 0, "invariant.XAlwaysErrIs requires at least one target")
	for _, t := range targets {
		Always(t != nil, "All invariant.XAlwaysErrIs targets must not be nil")
	}
	actual := fn()
	for _, t := range targets {
		if errors.Is(actual, t) {
			registerAssertion("XAlwaysErrIs", msg)
			return
		}
	}
	AssertionFailureCallback(fmt.Sprintf("%s: error did not match any targets. got %q. %s\n", AssertionFailureMsgPrefix, actual, msg))
}

// XAlwaysErrIsNot evaluates fn and calls AssertionFailureCallback if the returned error matches any target.
//
//go:noinline
func XAlwaysErrIsNot(fn func() error, msg string, targets ...error) {
	Always(len(targets) > 0, "invariant.XAlwaysErrIsNot requires at least one target")
	for _, t := range targets {
		Always(t != nil, "All invariant.XAlwaysErrIsNot targets must not be nil")
	}
	actual := fn()
	for _, t := range targets {
		if errors.Is(actual, t) {
			AssertionFailureCallback(fmt.Sprintf("%s: error unexpectedly matched a target. got %q. %s\n", AssertionFailureMsgPrefix, actual, msg))
			return
		}
	}
	registerAssertion("XAlwaysErrIsNot", msg)
}

// TODO: func RandomInt

// Be generous with these constants. This library minimizes heap allocations by
// preferring fixed-size backing arrays for slices. These values determine the
// size of those arrays and should accommodate typical codebases without
// frequent resizing.
const (
	// maxAssertionsPerPackage is the maximum number of Sometimes, XSometimes, Always*, and XAlways* calls in a single package.
	maxAssertionsPerPackage = 2048
	maxGoFilesPerPackage    = 1024
	maxFilePath             = 260
	maxFileLines            = 5 // In digits (99,999 lines)
	assertionIDLength       = maxFilePath + 1 + maxFileLines
)
