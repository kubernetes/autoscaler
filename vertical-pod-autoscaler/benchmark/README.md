# VPA Performance Benchmark

Measures VPA component latencies using KWOK (Kubernetes WithOut Kubelet) to simulate pods without real resource consumption.

> **Note:** Currently only updater metrics are collected. Recommender metrics are planned for the future.

<!-- toc -->
- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [What It Does](#what-it-does)
- [Profiles](#profiles)
- [Flags](#flags)
- [Metrics Collected](#metrics-collected)
  - [Updater Metrics](#updater-metrics)
- [Cleanup](#cleanup)
- [Notes](#notes)
  - [Performance Optimizations](#performance-optimizations)
  - [Caveats](#caveats)
<!-- /toc -->

## Prerequisites

- Go 1.21+
- kubectl
- Kind
- yq

## Quick Start

```bash
# Create a Kind cluster with tuned settings
kind create cluster --config=benchmark/kind-config.yaml

# Deploy VPA
./hack/deploy-for-e2e-locally.sh full-vpa

# Build and run
go build -C benchmark -o ../bin/vpa-benchmark .
# requires --kubeconfig flag, KUBECONFIG env var, or ~/.kube/config
bin/vpa-benchmark --profile=small --output=results.txt
```

## What It Does

1. Installs KWOK and creates a fake node
2. Configures VPA deployments with higher QPS/burst limits
3. For each profile run:
   - Scales down VPA components
   - Cleans up previous benchmark resources
   - Creates ReplicaSets with fake pods assigned directly to KWOK node (bypasses scheduler)
   - Creates VPAs targeting those ReplicaSets
   - Scales up recommender, waits for recommendations
   - Scales up updater, waits for its loop to complete
   - Scrapes `vpa_updater_execution_latency_seconds_sum` metrics
4. Outputs results to stdout and/or a CSV file if specified

e.g., of output using this command: `bin/vpa-benchmark --profile=small,large,xxlarge`

```bash
========== Results ==========
┌───────────────┬───────────────┬────────────────┬───────────────────┐
│     STEP      │ SMALL  ( 25 ) │ LARGE  ( 250 ) │ XXLARGE  ( 1000 ) │
├───────────────┼───────────────┼────────────────┼───────────────────┤
│ AdmissionInit │ 0.0000s       │ 0.0001s        │ 0.0004s           │
│ EvictPods     │ 2.4239s       │ 24.5535s       │ 98.6963s          │
│ FilterPods    │ 0.0002s       │ 0.0020s        │ 0.0925s           │
│ ListPods      │ 0.0001s       │ 0.0006s        │ 0.0025s           │
│ ListVPAs      │ 0.0024s       │ 0.0030s        │ 0.0027s           │
│ total         │ 2.4267s       │ 24.5592s       │ 98.7945s          │
└───────────────┴───────────────┴────────────────┴───────────────────┘
```

We can then compare the results of a code change with the results of the main branch.
Ideally the benchmark would be done on the same machine (or a similar one), with the same benchmark settings (profiles and runs).

## Profiles

| Profile | VPAs | ReplicaSets | Pods |
| ------- | ---- | ----------- | ---- |
| small   | 25   | 25          | 50   |
| medium  | 100  | 100         | 200  |
| large   | 250  | 250         | 500  |
| xlarge  | 500  | 500         | 1000 |
| xxlarge | 1000 | 1000        | 2000 |

## Flags

| Flag | Default | Description |
| ---- | ------- | ----------- |
| `--profile` | small | Comma-separated profiles to run. You can run multiple profiles at once. (e.g., `--profile=small,medium`) |
| `--runs` | 1 | Iterations per profile. This is used for averaging multiple runs. |
| `--output` | "" | Path to output file for results table (CSV format). Output will always be printed to stdout. |
| `--kubeconfig` | "" | Path to kubeconfig. Required if not using KUBECONFIG env var or ~/.kube/config. |

## Metrics Collected

### Updater Metrics

| Metric | Description |
| ------ | ----------- |
| `ListVPAs` | List VPA objects |
| `ListPods` | List pods matching VPA targets |
| `FilterPods` | Filter evictable pods |
| `AdmissionInit` | Verify admission controller status |
| `EvictPods` | Evict pods needing updates |
| `total` | Total loop time |

## Cleanup

```bash
kind delete cluster
```

## Notes

### Performance Optimizations

The benchmark includes several performance optimizations:

- Modifies VPA deployments at startup using `yq`:
  - Sets `--kube-api-qps=100` and `--kube-api-burst=200` on all three components
  - Sets `--updater-interval=2m` on the updater (default is 60s)
- Pods are assigned directly to the KWOK node via `nodeName`, bypassing the scheduler for faster creation
- The `kind-config.yaml` increases API server limits (`max-requests-inflight`, `max-mutating-requests-inflight`) and kube-controller-manager client QPS to handle the large number of API calls
- Uses ReplicaSets instead of Deployments to skip the Deployment controller layer and speed up pod creation, but keep a targetRef for VPA

### Caveats

- The updater uses `time.Tick` which waits the full interval before the first tick, so the benchmark sleeps 2 minutes before polling for metrics
- The benchmark uses Recreate update mode. In-place scaling is not supported on KWOK pods.
- The benchmark scales down all VPA components at the start of each run, so that any caching is not a factor.
