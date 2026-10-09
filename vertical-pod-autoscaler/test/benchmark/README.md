# VPA Performance Benchmark

Measures VPA component latencies using KWOK (Kubernetes WithOut Kubelet) to simulate pods without real resource consumption.

> **Note:** The cluster benchmark currently only collects updater metrics. Recommender metrics are planned for the future. The hot paths of the recommender and the admission-controller are covered by [Go micro-benchmarks](#micro-benchmarks-no-cluster-required) that run without a cluster.

<!-- toc -->
- [Prerequisites](#prerequisites)
- [Quick Start (Local)](#quick-start-local)
- [Manual Setup](#manual-setup)
- [What It Does](#what-it-does)
- [Profiles](#profiles)
- [Flags](#flags)
- [Metrics Collected](#metrics-collected)
  - [Updater Metrics](#updater-metrics)
- [Scripts](#scripts)
- [Cleanup](#cleanup)
- [Micro-benchmarks (No Cluster Required)](#micro-benchmarks-no-cluster-required)
- [Notes](#notes)
  - [Performance Optimizations](#performance-optimizations)
  - [Caveats](#caveats)
<!-- /toc -->

## Prerequisites

- Go 1.21+
- kubectl
- Kind
- Helm

## Quick Start (Local)

The `full-benchmark.sh` script handles everything end-to-end: creates a Kind cluster, deploys VPA, installs KWOK, configures VPA for benchmarking, builds and runs the benchmark.

```bash
cd vertical-pod-autoscaler
./test/benchmark/hack/full-benchmark.sh --profile=small --output=results.csv
```

You can pass any benchmark flags directly:

```bash
./benchmark/hack/full-benchmark.sh --profile=small,medium,large --runs=3 --output=results.csv
```

## Manual Setup

If you prefer to run each step individually (or if the cluster already exists):

```bash
# 1. Create a Kind cluster (the full-benchmark.sh script appends benchmark-specific
#    kubeadmConfigPatches to .github/kind-config.yaml automatically; for manual setup,
#    create a cluster with the base config or your own)
kind create cluster --config=.github/kind-config.yaml

# 2. Deploy VPA for benchmark
EXTRA_HELM_VALUES=./benchmark/hack/values.yaml ./hack/deploy-for-e2e-locally.sh full-vpa

# 3. Install KWOK and create fake node
./benchmark/hack/install-kwok.sh

# 4. Build and run
go build -C benchmark -o ../bin/vpa-benchmark .
./bin/vpa-benchmark --profile=small --output=results.csv
```

## What It Does

The benchmark program (`main.go`) assumes the cluster is already set up with VPA, KWOK, and the fake node. It then:

1. For each profile run:
   - Scales down VPA components
   - Cleans up previous benchmark resources
   - Creates ReplicaSets with fake pods assigned directly to KWOK node (bypasses scheduler)
   - Creates noise ReplicaSets (if `--noise-ratio` > 0) — these are not managed by any VPA
   - Creates VPAs targeting managed ReplicaSets only
   - Scales up recommender, waits for recommendations
   - Scales up updater, waits for its loop to complete
   - Scrapes `vpa_updater_execution_latency_seconds_sum` metrics
2. Outputs results to stdout and/or a CSV file if specified

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

When `--noise-percentage=P` is set, each profile also creates `P%` additional noise ReplicaSets (not managed by any VPA). For example, `--profile=medium --noise-percentage=50` creates 100 managed RS (200 pods) + 50 noise RS (100 pods) = 300 total pods.

## Flags

| Flag | Default | Description |
| ---- | ------- | ----------- |
| `--profile` | small | Comma-separated profiles to run. You can run multiple profiles at once. (e.g., `--profile=small,medium`) |
| `--runs` | 1 | Iterations per profile. This is used for averaging multiple runs. |
| `--output` | "" | Path to output file for results table (CSV format). Output will always be printed to stdout. |
| `--kubeconfig` | "" | Path to kubeconfig. Required if not using KUBECONFIG env var or ~/.kube/config. |
| `--noise-percentage` | 0% | Percentage of additional noise (unmanaged) ReplicaSets relative to managed ReplicaSets. Set to 0% for no noise. Noise pods increase `FilterPods` and `ListPods` costs without adding VPAs. |

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

## Scripts

| Script | Purpose |
| ------ | ------- |
| `hack/full-benchmark.sh` | Full local workflow (Kind + VPA + KWOK + configure + benchmark) |
| `hack/install-kwok.sh` | Install KWOK controller and create fake node |
| `hack/values.yaml` | Helm values for benchmark-specific VPA configuration |

Environment variables accepted by the scripts:

| Variable | Default | Used by |
| -------- | ------- | ------- |
| `KWOK_VERSION` | `v0.7.0` | `install-kwok.sh` |
| `KWOK_NAMESPACE` | `kube-system` | `install-kwok.sh` |
| `KWOK_NODE_NAME` | `kwok-node` | `install-kwok.sh` |
| `VPA_NAMESPACE` | `kube-system` | `configure-vpa.sh` |
| `KIND_CLUSTER_NAME` | `kind` | `full-benchmark.sh` |

## Cleanup

```bash
kind delete cluster
```

## Micro-benchmarks (No Cluster Required)

Standard Go benchmarks (`testing.B`) cover the recommender and
admission-controller hot paths with synthetic VPA and pod fixtures. They run in
seconds on a developer machine, need no cluster, and are the quickest way to
check whether a code change regresses these code paths.

| Package | Benchmark | What it measures |
| ------- | --------- | ---------------- |
| `pkg/recommender/model` | `BenchmarkClusterStateAddOrUpdatePod` | Feeding pods into the cluster state: steady-state resync of tracked pods and pod churn scaling with the number of VPAs |
| `pkg/recommender/model` | `BenchmarkClusterStateAddSample` | Aggregating a single CPU/memory usage sample (runs for every container on every metrics resync) |
| `pkg/recommender/model` | `BenchmarkVpaAggregateStateByContainerName` | Merging the aggregations contributing to a VPA (runs for every VPA on every recommender loop) |
| `pkg/recommender/logic` | `BenchmarkGetRecommendedPodResources` | Computing the recommendation for one VPA from aggregated usage histograms, with the default production estimator chain |
| `pkg/admission-controller/logic` | `BenchmarkAdmissionServerServe` | Full per-request cost of the pod admission webhook: request decoding, VPA matching, patch calculation and response encoding |

Run all of them:

```bash
cd vertical-pod-autoscaler
go test -run='^$' -bench=. -benchmem \
  ./pkg/recommender/model/ ./pkg/recommender/logic/ ./pkg/admission-controller/logic/
```

To compare a change against `master`, collect repeated runs of the relevant
benchmark on both branches and feed them to
[benchstat](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat):

```bash
go test -run='^$' -bench=BenchmarkAdmissionServerServe -benchmem -count=10 \
  ./pkg/admission-controller/logic/ | tee /tmp/new.txt
# ... check out the baseline, repeat with | tee /tmp/old.txt ...
benchstat /tmp/old.txt /tmp/new.txt
```

## Notes

### Performance Optimizations

The benchmark includes several performance optimizations:

- `values.yaml` configures VPA components via Helm:
  - Sets `--kube-api-qps=100` and `--kube-api-burst=200` on all three components
  - Sets `--updater-interval=2m` on the updater (default is 60s)
  - Sets `--memory-saver=true` on the recommender
- Pods are assigned directly to the KWOK node via `nodeName`, bypassing the scheduler for faster creation
- The benchmark script appends `kubeadmConfigPatches` to the base `.github/kind-config.yaml` to increase API server limits (`max-requests-inflight`, `max-mutating-requests-inflight`) and kube-controller-manager client QPS to handle the large number of API calls
- Uses ReplicaSets instead of Deployments to skip the Deployment controller layer and speed up pod creation, but keep a targetRef for VPA

### Caveats

- The updater uses `time.Tick` which waits the full interval before the first tick, so the benchmark sleeps 2 minutes before polling for metrics
- The benchmark uses Recreate update mode. In-place scaling is not supported on KWOK pods.
- The benchmark scales down all VPA components at the start of each run, so that any caching is not a factor.
