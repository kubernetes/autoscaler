#!/bin/bash
set -e

log_file="./test-runs/benchmark_run.log"
echo "Starting benchmarks run..." > "$log_file"

run_bench() {
  echo "RUNNING: $@" >> "$log_file"
  $@ >> "$log_file" 2>&1
  echo "----------------------------------------" >> "$log_file"
}

# 1. Multi-Iteration
echo "1. Running Multi-Iteration Benchmarks..."
run_bench go test -bench=BenchmarkMultiIteration -benchmem -run=^$ ./core/bench/... -benchtime=1s -target-nodes-count=100 -steps-count=5
run_bench go test -bench=BenchmarkMultiIteration -benchmem -run=^$ ./core/bench/... -benchtime=1s -target-nodes-count=500 -steps-count=5
run_bench go test -bench=BenchmarkMultiIteration -benchmem -run=^$ ./core/bench/... -benchtime=1s -target-nodes-count=1000 -steps-count=5

# 2. Hostname Affinity
echo "2. Running Hostname Affinity Benchmarks..."
run_bench go test -bench=BenchmarkRunOnceAffinitySurge_Hostname -benchmem -run=^$ ./core/bench/... -benchtime=1s -nodes-count=100 -nodes-per-node-group=10 -surge-count=20
run_bench go test -bench=BenchmarkRunOnceAffinitySurge_Hostname -benchmem -run=^$ ./core/bench/... -benchtime=1s
run_bench go test -bench=BenchmarkRunOnceAffinitySurge_Hostname -benchmem -run=^$ ./core/bench/... -benchtime=1s -nodes-count=1000 -nodes-per-node-group=100 -surge-count=200

# 3. Zonal Affinity
echo "3. Running Zonal Affinity Benchmarks..."
run_bench go test -bench=BenchmarkRunOnceAffinitySurge_Zonal -benchmem -run=^$ ./core/bench/... -benchtime=1s -nodes-count=100 -nodes-per-node-group=10 -surge-count=20
run_bench go test -bench=BenchmarkRunOnceAffinitySurge_Zonal -benchmem -run=^$ ./core/bench/... -benchtime=1s
run_bench go test -bench=BenchmarkRunOnceAffinitySurge_Zonal -benchmem -run=^$ ./core/bench/... -benchtime=1s -nodes-count=1000 -nodes-per-node-group=100 -surge-count=200

echo "All benchmarks finished!"
