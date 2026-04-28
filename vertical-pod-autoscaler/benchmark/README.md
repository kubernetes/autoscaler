# VPA Performance Benchmark

Measures VPA component latencies using KWOK (Kubernetes WithOut Kubelet) to simulate pods without real resource consumption.

<!-- toc -->
- [Prerequisites](#prerequisites)
- [Quick Start (Local)](#quick-start-local)
- [Manual Setup](#manual-setup)
- [What It Does](#what-it-does)
- [Profiles](#profiles)
- [Flags](#flags)
- [Metrics Collected](#metrics-collected)
  - [Recommender Metrics](#recommender-metrics)
  - [Updater Metrics](#updater-metrics)
  - [Admission Controller Metrics](#admission-controller-metrics)
- [Scripts](#scripts)
- [Cleanup](#cleanup)
- [Notes](#notes)
  - [Performance Optimizations](#performance-optimizations)
  - [Caveats](#caveats)
<!-- /toc -->

## Prerequisites

- Go 1.25+
- kubectl
- Kind
- Helm

## Quick Start (Local)

The `full-benchmark.sh` script handles everything end-to-end: creates a Kind cluster, deploys VPA, installs KWOK, configures VPA for benchmarking, builds and runs the benchmark.

```bash
cd vertical-pod-autoscaler
./benchmark/hack/full-benchmark.sh --profile=small --output=results.csv
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
   - Scales down all VPA components and cleans up previous benchmark resources
   - Creates ReplicaSets with fake pods assigned directly to KWOK node (bypasses scheduler)
   - Creates noise ReplicaSets (if `--noise-percentage` > 0) — these are not managed by any VPA
   - Creates VPAs targeting managed ReplicaSets only
   - Scales up recommender and admission controller, waits for recommendations
   - Scrapes recommender execution latency metrics
   - Scales up updater, waits for its loop to complete
   - Scrapes updater and admission controller execution latency metrics
2. Outputs per-run tables (with Avg column when multiple runs) and cross-profile summary tables to stdout and/or a CSV file

> [!NOTE]
> Recommender and updater latencies are cumulative sums from a single loop. Admission controller latencies are per-request averages (sum divided by request count), since it handles many requests per benchmark run.

e.g., of output using this command: `bin/vpa-benchmark --profile=small,large,xxlarge`

```
========== Results [Recommender] ==========
┌─────────────────────┬───────────────┬────────────────┬───────────────────┐
│        STEP         │ SMALL  ( 25 ) │ LARGE  ( 250 ) │ XXLARGE  ( 1000 ) │
├─────────────────────┼───────────────┼────────────────┼───────────────────┤
│ LoadVPAs            │ 0.0005s       │ 0.0022s        │ 0.0099s           │
│ LoadPods            │ 0.0007s       │ 0.0138s        │ 0.1869s           │
│ LoadMetrics         │ 0.0031s       │ 0.0055s        │ 0.0036s           │
│ UpdateVPAs          │ 0.0142s       │ 0.5050s        │ 8.0046s           │
│ MaintainCheckpoints │ 0.0174s       │ 3.0046s        │ 18.0054s          │
│ GarbageCollect      │ 0.0001s       │ 0.0055s        │ 0.0426s           │
│ total               │ 0.0361s       │ 3.5367s        │ 26.2529s          │
└─────────────────────┴───────────────┴────────────────┴───────────────────┘

========== Results [Updater] ==========
┌───────────────┬───────────────┬────────────────┬───────────────────┐
│     STEP      │ SMALL  ( 25 ) │ LARGE  ( 250 ) │ XXLARGE  ( 1000 ) │
├───────────────┼───────────────┼────────────────┼───────────────────┤
│ ListVPAs      │ 0.0021s       │ 0.0020s        │ 0.0023s           │
│ ListPods      │ 0.0001s       │ 0.0004s        │ 0.0022s           │
│ FilterPods    │ 0.0001s       │ 0.0016s        │ 0.0242s           │
│ AdmissionInit │ 0.0000s       │ 0.0001s        │ 0.0003s           │
│ EvictPods     │ 2.3205s       │ 24.5523s       │ 98.5502s          │
│ total         │ 2.3229s       │ 24.5565s       │ 98.5792s          │
└───────────────┴───────────────┴────────────────┴───────────────────┘

========== Results [Admission Controller] ==========
┌────────────────┬───────────────┬────────────────┬───────────────────┐
│      STEP      │ SMALL  ( 25 ) │ LARGE  ( 250 ) │ XXLARGE  ( 1000 ) │
├────────────────┼───────────────┼────────────────┼───────────────────┤
│ read_request   │ 0.0000s       │ 0.0000s        │ 0.0000s           │
│ admit          │ 0.0004s       │ 0.0005s        │ 0.0007s           │
│ build_response │ 0.0000s       │ 0.0000s        │ 0.0000s           │
│ write_response │ 0.0000s       │ 0.0000s        │ 0.0000s           │
│ request_count  │ 26            │ 251            │ 1001              │
│ total          │ 0.0005s       │ 0.0005s        │ 0.0007s           │
└────────────────┴───────────────┴────────────────┴───────────────────┘
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

All metrics are scraped from each component's `/metrics` endpoint via port-forwarding. Values are parsed from `vpa_<component>_execution_latency_seconds` histograms. Admission controller values are per-request averages.

### Recommender Metrics

Steps are listed in execution order.

| Step | Description |
| ---- | ----------- |
| `LoadVPAs` | Load VPA objects |
| `LoadPods` | Load pods matching VPA targets |
| `LoadMetrics` | Load metrics from metrics-server |
| `UpdateVPAs` | Compute and write recommendations |
| `MaintainCheckpoints` | Create/update VPA checkpoints |
| `GarbageCollect` | Clean up stale data |
| `total` | Total loop time |

### Updater Metrics

| Step | Description |
| ---- | ----------- |
| `ListVPAs` | List VPA objects |
| `ListPods` | List pods matching VPA targets |
| `FilterPods` | Filter evictable pods |
| `AdmissionInit` | Verify admission controller status |
| `EvictPods` | Evict pods needing updates |
| `total` | Total loop time |

### Admission Controller Metrics

| Step | Description |
| ---- | ----------- |
| `read_request` | Parse incoming admission request |
| `admit` | Compute resource recommendations for the pod |
| `build_response` | Build admission response |
| `write_response` | Write response back to API server |
| `request_count` | Total number of admission requests handled |
| `total` | Total per-request time |

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
| `KIND_CLUSTER_NAME` | `kind` | `full-benchmark.sh` |

## Cleanup

```bash
kind delete cluster
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

- The updater uses `time.Tick` which waits the full interval before the first tick, so the benchmark polls for up to 5 minutes waiting for the updater's `total` metric to appear.
- The benchmark uses Recreate update mode. In-place scaling is not supported on KWOK pods.
- The benchmark scales down all VPA components at the start of each run, so that any caching is not a factor.
