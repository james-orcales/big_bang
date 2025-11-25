//go:build disable_assertions

package invariant

var (
	AssertionFailureCallback = func(msg string) {
	}

	DefaultAssertionFailureCallbackFatal = func(msg string) {
	}

	DefaultAssertionFailureCallbackPanic = func(msg string) {
	}
)

func RegisterPackagesForAnalysis(dirs ...string) {
}

func AnalyzeAssertionFrequency() {
}

func Unreachable(msg string) {
}

func Unimplemented(msg string) {
}

func Sometimes(ok bool, msg string) {
}

func Always(cond bool, msg string) {
}

func AlwaysNil(x any, msg string) {
}

func AlwaysErrIs(actual error, msg string, targets ...error) {
}

func AlwaysErrIsNot(actual error, msg string, targets ...error) {
}

func XSometimes(ok func() bool, msg string) {
}

func XAlways(fn func() bool, msg string) {
}

func XAlwaysNil(fn func() any, msg string) {
}

func XAlwaysErrIs(fn func() error, msg string, targets ...error) {
}

func XAlwaysErrIsNot(fn func() error, msg string, targets ...error) {
}
