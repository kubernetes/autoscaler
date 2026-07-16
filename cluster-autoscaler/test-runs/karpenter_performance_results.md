# Karpenter Simulator Performance Benchmark Results

This document presents the performance comparison between the default Cluster Autoscaler (CA) simulator and the new Karpenter-based simulator under heavy pod affinity and anti-affinity constraints at scale.

## Benchmark Environment
- **OS**: Linux
- **Arch**: amd64
- **CPU**: AMD EPYC 7B13 (64 cores used for benchmark run)
- **Go version**: go1.21+ (assumed from CA codebase)

## Benchmark Configuration
- **Initial Nodes**: 1000 nodes (10 groups of 100 nodes each)
- **Pods per Node**: 10 pods initially scheduled (nodes are 100% full)
- **Surge Pods**: 2000 unschedulable pods to be scaled up
- **Affinity Type**: 
  - Zonal Pod Anti-Affinity (on `topology.kubernetes.io/zone`)
  - Hostname Pod Anti-Affinity (on `kubernetes.io/hostname`)
- **Benchmark Command**:
  ```bash
  go test -bench=BenchmarkRunOnceAffinitySurge -benchmem -run=^$ ./core/bench/... -benchtime=1s -nodes-count=1000 -nodes-per-node-group=100 -surge-count=2000
  ```

## Results Summary

| Benchmark | Simulator | Execution Time (ns/op) | Time (s) | Allocated Memory (B/op) | Memory (GB) | Allocations (allocs/op) | Status |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **Hostname Affinity** | Default | 86,819,265,931 | 86.82s | 23,189,042,616 | 23.19 GB | 79,044,046 | PASS |
| | Karpenter | 58,578,641,338 | 58.58s | 13,758,482,672 | 13.76 GB | 79,907,302 | PASS |
| | **Improvement** | | **32.5% faster** | | **40.7% reduction** | **1% increase** | |
| **Zonal Affinity** | Default | 69,892,881,200 | 69.89s | 13,100,645,560 | 13.10 GB | 41,000,779 | PASS |
| | Karpenter | 39,621,702,906 | 39.62s | 7,881,731,288 | 7.88 GB | 78,595,194 | PASS |
| | **Improvement** | | **43.3% faster** | | **39.8% reduction** | **91.7% increase** | |

## Key Findings

1. **Significant Speedup**: Karpenter's scheduler is **32% to 43% faster** than the default CA simulator. This is because Karpenter uses a more optimized topology-aware scheduling algorithm that handles pod affinity/anti-affinity constraints more efficiently than CA's filter-based approach at scale.
2. **Reduced Memory Footprint**: Karpenter simulator uses **40% less memory** (allocated bytes) during the simulation run.
3. **Allocation Trade-off**: In the Zonal Affinity scenario, Karpenter has **91% more allocations**, but they are significantly smaller on average (100 bytes vs 317 bytes), resulting in lower total memory usage and faster execution (less GC pressure if GC were enabled, and faster allocation time).

## Bugs Resolved During Performance Testing

During the scale-up verification, several critical bugs were identified and fixed in the Karpenter simulator implementation:

1. **Zonal Anti-Affinity Failure**: 
   - *Issue*: Well-known labels (specifically `topology.kubernetes.io/zone`) were being stripped from the generated `NodePool` requirements during translation. This caused Karpenter's scheduler to fail with `unsatisfiable topology constraint` because it couldn't guarantee zonal placement constraints.
   - *Fix*: Modified `Convert` to only skip `RestrictedLabels` (e.g., `kubernetes.io/hostname`) and preserve `WellKnownLabels`.
2. **Hostname Anti-Affinity "Missing Pods" Bug**:
   - *Issue*: Simulated pods committed to the `ClusterSnapshot` in intermediate batches did not have `pod.Spec.NodeName` set. Karpenter's scheduler relies on `NodeName` to bind pods to nodes. Without it, Karpenter saw the simulated nodes as empty and scheduled subsequent pods on them, leading to missing claims and verification failures.
   - *Fix*: Explicitly set `p.Spec.NodeName = simNodeName` before calling `ForceAddPod` when committing simulated pods.
3. **Winner InstanceType Selection Bug**:
   - *Issue*: `mapResultsToOptions` was only evaluating the first option in `nc.InstanceTypeOptions`. If Karpenter returned multiple options and the first one failed physical node group matching (e.g. due to zone constraints), the claim was discarded, even if other options would have matched.
   - *Fix*: Modified `mapResultsToOptions` to loop over all `InstanceTypeOptions` and select the first one that successfully matches a physical node group, respecting Karpenter's price preference sorting.
