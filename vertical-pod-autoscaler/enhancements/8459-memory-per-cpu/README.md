# AEP-8459: MemoryPerCPU

<!-- toc -->
- [Summary](#summary)
- [Motivation](#motivation)
  - [Goals](#goals)
  - [Non-Goals](#non-goals)
- [Proposal](#proposal)
- [Design Details](#design-details)
  - [API Changes](#api-changes)
  - [Behavior](#behavior)
  - [Feature Enablement and Rollback](#feature-enablement-and-rollback)
    - [How can this feature be enabled / disabled in a live cluster?](#how-can-this-feature-be-enabled--disabled-in-a-live-cluster)
  - [Kubernetes Version Compatibility](#kubernetes-version-compatibility)
  - [Validation](#validation)
  - [Test Plan](#test-plan)
- [Implementation History](#implementation-history)
- [Future Work](#future-work)
- [Alternatives](#alternatives)
  <!-- /toc -->

## Summary

This AEP proposes a new feature to allow enforcing a fixed memory-per-CPU ratio (`memoryPerCPU`) in Vertical Pod Autoscaler (VPA) recommendations.  
The feature is controlled by a new alpha feature gate `MemoryPerCPURatio` (default off).

## Motivation

Many workloads scale their memory requirements proportionally to CPU, but today VPA generates independent CPU and memory recommendations. This can lead to skewed configurations — for example, too much memory for a small CPU allocation, or too little memory for a high CPU allocation.

The `memoryPerCPU` field addresses this by enforcing a predictable CPU-to-memory ratio in recommendations. This reduces the risk of misconfiguration, ensures consistency, and simplifies tuning for workloads where CPU and memory usage are tightly coupled.

This feature is particularly useful in environments where services are billed primarily on CPU with a fixed CPU-to-memory ratio. In such cases, it allows VPA to be used for automatic vertical scaling while preserving the existing billing model and guarantees to customers.

### Goals

* Allow users to specify a `memoryPerCPU` ratio in `VerticalPodAutoscaler` objects.  
* Ensure VPA recommendations respect the ratio across Target, LowerBound, UpperBound, and UncappedTarget.  

### Non-Goals

* Redesign of the VPA recommender algorithm beyond enforcing the ratio.  
* Supporting multiple ratio policies per container (only one `memoryPerCPU` is supported).  
* Retroactive migration of existing VPAs without explicit user opt-in.

## Proposal

Extend `ContainerResourcePolicy` with a new optional field:

```yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: my-app
spec:
  resourcePolicy:
    containerPolicies:
      - containerName: app
        minAllowed:
          cpu: 1
          memory: 4Gi
        maxAllowed:
          cpu: 4
          memory: 16Gi
        controlledResources: ["cpu", "memory"]
        controlledValues: RequestsAndLimits
        memoryPerCPU: "4Gi"
```

When enabled, VPA will adjust CPU or memory recommendations to maintain:

```
memory_bytes = cpu_cores * memoryPerCPU
```

## Design Details

### API Changes

* New field `memoryPerCPU` (`resource.Quantity`) in `ContainerResourcePolicy`.  
* Feature gate: `MemoryPerCPURatio` (alpha, default off).

### Behavior

* If both CPU and memory are controlled, VPA enforces the ratio.  
* Applies to Target, LowerBound, UpperBound, and UncappedTarget.
* Ratio enforcement is strict:
  * If the memory recommendation would exceed `cpu * memoryPerCPU`, then **CPU is increased** to satisfy the ratio.
  * If the CPU recommendation would exceed `memory / memoryPerCPU`, then **memory is increased** to satisfy the ratio.
* With the `MemoryPerCPURatio` feature gate disabled, the `memoryPerCPU` field is ignored and recommendations fall back to standard VPA behavior.

> [!IMPORTANT]
> The enforced ratio values will be capped by
> [`--container-recommendation-max-allowed-cpu`](https://github.com/kubernetes/autoscaler/blob/4d294562e505431d518a81e8833accc0ec99c9b8/vertical-pod-autoscaler/pkg/recommender/main.go#L122)
> and
> [`--container-recommendation-max-allowed-memory`](https://github.com/kubernetes/autoscaler/blob/4d294562e505431d518a81e8833accc0ec99c9b8/vertical-pod-autoscaler/pkg/recommender/main.go#L123)
> flag values, if set.

#### Examples

* Example 1: `memoryPerCPU = 4Gi`
  * Baseline recommendation: 1 CPU, 8Gi memory
  * UncappedTarget (ratio enforced): 2 CPUs, 8Gi
  * Target (after policy/caps): 2 CPUs, 8Gi

* Example 2: `memoryPerCPU = 4Gi`
  * Baseline recommendation: 2 CPUs, 4Gi memory
  * UncappedTarget (ratio enforced): 2 CPUs, 8Gi
  * Target (after policy/caps): 2 CPUs, 8Gi

* Example 3: `memoryPerCPU = 4Gi`, with `--container-recommendation-max-allowed-memory=7Gi` or with `maxAllowed.memory=6Gi` set in VPA object
  * Baseline recommendation: 2 CPUs, 4Gi memory
  * UncappedTarget (ratio enforced): 2 CPUs, 8Gi
  * Target (capped): 2 CPUs, 6Gi  ← memory capped by max-allowed-memory; ratio not fully satisfied

### Feature Enablement and Rollback

#### How can this feature be enabled / disabled in a live cluster?

* Feature gate name: `MemoryPerCPURatio`  
* Default: Off (Alpha)  
* Components depending on the feature gate:
  * admission-controller
  * recommender

**When enabled**:  
* VPA honors `memoryPerCPU` in recommendations.  

**When disabled**:  
* `memoryPerCPU` is ignored.  
* Recommendations behave as before.

### Kubernetes Version Compatibility

The `memoryPerCPU` feature requires VPA version 1.5.0 or higher. The feature is being introduced as alpha and will follow the standard Kubernetes feature gate graduation process:
- Alpha: v1.5.0 (default off)
- Beta: TBD (default on)
- GA: TBD (default on)

### Validation

* `memoryPerCPU` must be > 0.  
* Value must be a valid `resource.Quantity` (e.g., `512Mi`, `4Gi`).
* If `memoryPerCPU` is set, controlledResources must include both `cpu` and `memory`.
* Admission ensures that memoryPerCPU is reachable within the VPA bounds.
  * Reject the object if `minAllowed.cpu` × `memoryPerCPU` > `maxAllowed.memory`.
  * Reject the object if `maxAllowed.cpu` × `memoryPerCPU` < `minAllowed.memory`.
    * Example: `minAllowed.cpu=1`, `maxAllowed.memory=4Gi`, `memoryPerCPU`=5Gi ⇒ invalid (1×5Gi > 4Gi).
    * Example: `minAllowed.cpu=1`, `maxAllowed.memory=4Gi`, `memoryPerCPU=2Gi` ⇒ valid (1×2Gi ≤ 4Gi).

### Test Plan

* Unit tests covering:
  - ensuring ratio enforcement logic,
  - ensuring that when the feature gate is on or off the values and validation are applied accordingly.
* E2E tests comparing behavior with different configurations.

## Implementation History

* 2025-08-19: Initial proposal

## Future Work


## Alternatives

