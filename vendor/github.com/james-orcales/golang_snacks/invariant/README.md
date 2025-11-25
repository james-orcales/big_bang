# invariant

## Long term plan

Include go scripts that instrument your code. These reduce boilerplate, improve runtime performance, and enable deeper test insights.

- Metaprogramming to hardcode additional metadata for each assertion call
- Count function:assertion ratio. Ideally, it's at least equal to len(input params)+len(return values)
- Export metadata to CSV then do simple machine learning analysis on that
