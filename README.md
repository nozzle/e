# e - opinionated Go errors

This has been in production for several years at [Nozzle](https://nozzle.io), with millions of errors pushed
through it. This is a copy from our internal monorepo package. For now, we don't plan to use this package
directly, but if there is interest and an appetite for public use, we're open to collaboration.

## Expected Usage

[Google Slides presentation](https://docs.google.com/presentation/d/1mmIQKy7124MlMboJGI3YEkGDsJafRQQDGYGCbf-0iEA/)

## Future Work

- [ ] remove hardcoded Nozzle use cases
- [ ] make stack trace truncation / frame skipping configurable
- [ ] make error log printout + repo base url configurable
- [ ] make frame class assignment configurable
- [ ] add a context labeling interface to add user id, trace id, etc
- [ ] make error reporters pluggable vs hardcoding Sentry and Google Error Reporting

## Benchmarks

The initial wrap is heavy due to the stacktrace/memstats collection, parsing, and allocation,
while subsequent operations are much lighter. There hasn't been much focus on performance,
so there is likely some low-hanging fruit.

```
cpu: Intel(R) Core(TM) i9-9880H CPU @ 2.30GHz
BenchmarkInitialWrap
BenchmarkInitialWrap-16               	    720	   1451582 ns/op  420843 B/op    9061 allocs/op
BenchmarkAlreadyWrappedNoOpts
BenchmarkAlreadyWrappedNoOpts-16      275561215	     4.321 ns/op       0 B/op       0 allocs/op
BenchmarkAlreadyWrappedWithVars
BenchmarkAlreadyWrappedWithVars-16       666651	      1529 ns/op     540 B/op       6 allocs/op
```
