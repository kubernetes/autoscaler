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

Many workloads scale their memory requirements proportionally to CPU. Today, VPA independently recommends CPU and memory, which can result in skewed recommendations (too much memory for small CPU, or too little memory for high CPU).

By introducing `memoryPerCPU`, users can enforce a predictable ratio between CPU and memory, reducing risk of misconfiguration and simplifying tuning for ratio-based workloads.

In addition, some environments or organizations prefer to keep a fixed CPU-to-memory ratio for reasons such as:
* **Billing models** – Many cloud providers price instances based on predefined CPU/memory bundles. Enforcing a fixed ratio makes VPA recommendations align better with billing units, avoiding unexpected cost patterns.
* **Operational simplicity** – A consistent CPU/memory ratio across workloads reduces variability and simplifies capacity planning.

### Goals

* Allow users to specify a `memoryPerCPU` ratio in `VerticalPodAutoscaler` objects.  
* Ensure VPA recommendations respect the ratio across Target, LowerBound, UpperBound, and UncappedTarget.  
* Provide a feature gate to enable/disable the feature cluster-wide.

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
* If ratio cannot be applied (e.g., missing CPU), fallback to standard recommendations.  
* With feature gate OFF: recommendations are unaffected.

### Feature Enablement and Rollback

#### How can this feature be enabled / disabled in a live cluster?

* Feature gate name: `MemoryPerCPURatio`  
* Default: Off (Alpha)  
* Components depending on the feature gate:
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

### Test Plan

* Unit tests ensuring ratio enforcement logic.  

## Implementation History

* 2025-08-19: Initial proposal

## Future Work


## Alternatives

