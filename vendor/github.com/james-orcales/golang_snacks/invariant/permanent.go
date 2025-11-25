package invariant

import (
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"path"
	"runtime"
	"strings"
)

var IsRunningUnderGoTest = func() bool {
	v := false
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "-test.") {
			v = true
			break
		}
	}
	return v
}()

var IsRunningUnderGoFuzz = func() bool {
	v := false
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "-test.fuzz") {
			v = true
			break
		}
	}
	return v
}()

var IsRunningUnderGoBenchmark = func() bool {
	v := false
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "-test.bench") {
			v = true
			break
		}
	}
	return v
}()

// FprintStackTrace writes a formatted stack trace to the given io.Writer.
//
// It collects up to 15 caller program counters starting at `callerLocation`
// (relative to the call site).
//
// Frames corresponding to Go test harness files (_testmain.go) and common
// runtime or invariant helper functions are excluded from the output.
//
//	~/code/golang_snacks git:(main)
//	$ go test ./invariant/examples/02_math/ -v
//	=== RUN   TestAdd
//	=== PAUSE TestAdd
//	=== RUN   TestSubtract
//	=== PAUSE TestSubtract
//	=== RUN   TestMultiply
//	=== PAUSE TestMultiply
//	=== CONT  TestAdd
//	=== CONT  TestMultiply
//	--- PASS: TestAdd (0.00s)
//	=== CONT  TestSubtract
//	--- PASS: TestSubtract (0.00s)
//	02_math.Multiply          | /Users/my_username/code/golang_snacks/invariant/examples/02_math/math.go:125
//	02_math_test.TestMultiply | /Users/my_username/code/golang_snacks/invariant/examples/02_math/math_test.go:59
//	🚨 Assertion Failure 🚨: Your assertion message here.
//	FAIL    golang_snacks/invariant/examples/02_math        0.330s
//	FAIL
func FprintStackTrace(w io.Writer, callerLocation int) {
	const depth = 15
	var pcs [depth]uintptr
	skip := 2 + max(0, callerLocation)

	n := runtime.Callers(skip, pcs[:])
	fs := runtime.CallersFrames(pcs[:n])

	var frames [depth]runtime.Frame
	i := 0
	for {
		frame, ok := fs.Next()
		if !ok || i >= len(frames) {
			break
		}
		frame.Function = path.Base(frame.Function)
		frames[i] = frame
		i++
	}

	maxFn := 0
	for j := 0; j < i; j++ {
		n := len(frames[j].Function)
		if n > maxFn {
			maxFn = n
		}
	}

	for j := 0; j < i; j++ {
		frame := frames[j]
		if frame.File == "_testmain.go" {
			continue
		}
		switch frame.Function {
		case "runtime.main", "invariant.init.func1", "testing.tRunner",
			"invariant.Always", "invariant.AlwaysNil", "invariant.AlwaysErrIs", "invariant.AlwaysErrIsNot",
			"invariant.XAlways", "invariant.XAlwaysNil", "invariant.XAlwaysErrIs", "invariant.XAlwaysErrIsNot",
			"invariant.Sometimes", "invariant.XSometimes":
			continue
		}
		fmt.Fprintf(w,
			"%-*s | %s:%d\n",
			maxFn,
			frame.Function,
			frame.File,
			frame.Line,
		)
	}
}

// InjectFault randomly triggers a fault for testing purposes, based on the given percentage.
// Only active when running under Go fuzzing (`IsRunningUnderGoFuzz`).
//
// Example usage:
//
//	err := fallible()
//	if err != nil || invariant.InjectFault(30) {
//	    handleError()
//	}
//
// The `percent` argument specifies the probability (0–100) that a fault is injected.
func InjectFault(percent int) bool {
	return IsRunningUnderGoFuzz && rand.IntN(100) < percent
}
