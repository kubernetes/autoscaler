# GCE Autoscaler Benchmarks

This directory contains benchmarks for the GCE Autoscaler.


## Running Benchmarks

To run the GCE MIG lookup benchmark:

```bash
go test -v -bench=BenchmarkRunOnceWithGce -run=^$ -benchmem -count=6 -benchtime=10x ./core/bench/ > results.txt
```

### Benchmark Flags

*   `-count=N`: Runs the benchmark `N` times. Each run yields a separate data point for `benchstat`. `benchstat` needs at least 4-6 samples to compute reliable statistical confidence.
*   `-benchtime=Nx`: Runs exactly `N` iterations per sample.  For heavy benchmarks, specifying a fixed count (like `10x`) prevents it from running for too long.

## Comparing Results
To compare results of 2 benchmarks, install `benchstat`:

```bash
GOBIN=$PWD/core/bench/bin go install golang.org/x/perf/cmd/benchstat@latest
```

To compare two benchmark results (e.g., before and after optimization):

```bash
$PWD/core/bench/bin/benchstat before.txt after.txt
```
