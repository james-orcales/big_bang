# itlog

A zero-allocation logger for Go that keeps the Zerolog feel and speed in a tiny codebase, with
human-readable logfmt-style output.

Its accomplishments are:

1. 94% smaller footprint (~650 SLOC vs Zerolog's ~11.5k SLOC) excluding `*_test.go` and assertions
2. Performant for human-readable format (50 million logs per second) and zero heap allocations
4. Similar core API to Zerolog
5. Easily vendorable (MIT licensed)

```go
package main

import (
    "os"
    "github.com/james-orcales/golang_snacks/itlog"
)

func main() {
	lgr := itlog.New(os.Stdout, itlog.LevelInfo)
    // Notice that this wasn't assigned to a variable.
	lgr.Clone().WithStr("name", "James").WithStr("email", "iwillhackyou@proton.mail")
	{
		lgr := lgr.Clone().WithInt("FAANG_companies_hacked", 369)
		lgr.Info().Msg("This is a deep copy of the parent logger.")
	}
	lgr.Info().Msg("This logger's buffer was not mutated")
}

// $ go run .
// 2025-11-05T23:10:33Z|INF|This is a deep copy of the parent logger.                                       |FAANG_companies_hacked=369|
// 2025-11-05T23:10:33Z|INF|This logger's buffer was not mutated                                            |
```

## Benchmarks

These are the EXACT same benchmarks used by Zerolog. Note some benchmarks were
excluded since they're unsupported yet.

```
$ go test ./itlog   -bench=. -count=20 -benchtime=0.01s -benchmem -tags=disable_assertions > benchitlog
$ go test ./zerolog -bench=. -count=20 -benchtime=0.01s -benchmem > benchzerolog
$ benchstat itlog_bench zerolog_bench
goos: darwin
goarch: arm64
pkg: golang_snacks/itlog
cpu: Apple M4
                                       │ benchzerolog  │                benchitlog                 │
                                       │    sec/op     │    sec/op      vs base                    │
LogEmpty-10                               6.679n ±  3%   18.505n ± 27%   +177.04% (p=0.000 n=20)
Disabled-10                              0.3120n ±  2%   0.2959n ±  1%     -5.18% (p=0.000 n=20)
Info-10                                   13.73n ±  7%    24.04n ± 17%    +75.03% (p=0.000 n=20)
ContextFields-10                          14.76n ±  7%    21.62n ± 25%    +46.56% (p=0.000 n=20)
ContextAppend-10                          3.291n ±  1%   41.260n ±  2%  +1153.91% (p=0.000 n=20)
LogFields-10                              25.52n ±  8%    38.82n ±  5%    +52.13% (p=0.000 n=20)
LogArrayObject-10                         153.9n ±  3%    424.6n ±  1%   +175.89% (p=0.000 n=20)

                                       │  benchzerolog  │            benchitlog            │
                                       │      B/op      │    B/op     vs base              │
LogEmpty-10                                0.000 ± 0%     0.000 ± 0%  ~ (p=1.000 n=20) ¹
Disabled-10                                0.000 ± 0%     0.000 ± 0%  ~ (p=1.000 n=20) ¹
Info-10                                    0.000 ± 0%     0.000 ± 0%  ~ (p=1.000 n=20) ¹
ContextFields-10                           0.000 ± 0%     0.000 ± 0%  ~ (p=1.000 n=20) ¹
ContextAppend-10                             0.0 ± 0%     256.0 ± 0%  ? (p=0.000 n=20)
LogFields-10                               0.000 ± 0%     0.000 ± 0%  ~ (p=1.000 n=20) ¹
LogArrayObject-10                           0.00 ± 0%     72.00 ± 0%  ? (p=0.000 n=20)

                                       │ benchzerolog │            benchitlog            │
                                       │  allocs/op   │ allocs/op   vs base              │
LogEmpty-10                              0.000 ± 0%     0.000 ± 0%  ~ (p=1.000 n=20) ¹
Disabled-10                              0.000 ± 0%     0.000 ± 0%  ~ (p=1.000 n=20) ¹
Info-10                                  0.000 ± 0%     0.000 ± 0%  ~ (p=1.000 n=20) ¹
ContextFields-10                         0.000 ± 0%     0.000 ± 0%  ~ (p=1.000 n=20) ¹
ContextAppend-10                         0.000 ± 0%     2.000 ± 0%  ? (p=0.000 n=20)
LogFields-10                             0.000 ± 0%     0.000 ± 0%  ~ (p=1.000 n=20) ¹
LogArrayObject-10                        0.000 ± 0%     3.000 ± 0%  ? (p=0.000 n=20)

```

## Design


### Format

```
.......Header.....|.Body..\n
time|level|message|context\n
```

- The format is split into the Header and Body.
- The header is a fixed-width chunk, holding fields that are present in all
  logs. It is split into sub-components (1) time, (2) log level, (3) and
  message, all separated by the ComponentSeparator.
- The body is a variable-width chunk, holding dynamically appended fields. It
  consists of key=value pairs separated by ComponentSeparator and delimited by
  ContextKeyValueSeparator. Keys with multiple values are appended as individual
  pairs len(values) times in the context, sharing the same key.

- Context keys only permit the following characters `[a-zA-Z._]`
- Context string values are surrounded by double quotes `my_foo="your baz"`

- Float context values are either `+Inf`, `-Inf`, `NaN`, or an integer that ALWAYS at least one
  decimal point.

### Encoding

Itlog uses a lossy encoding.

**Message:**  

Raw null bytes (`'\x00'`) and newlines (`'\n'`) are replaced with whitespace `' '`.

**Context:**  
Raw null bytes (`'\x00'`) and newlines (`'\n'`) are encoded as the string
literals `\0` and `\n`. 

### No colored output

At first, I implemented colored output. In practice however, the colors are not
all that useful when the context gets large and your terminal screen is filled
with text that hard wrap, breaking visual alignment of the logs. It's better to
save the logs in a file and explore them in your editor. Another benefit is that
this simplifies the implementation further.


