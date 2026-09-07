# AEP-9970: Per-VPA Target Percentiles

<!-- toc -->
- [Summary](#summary)
- [Motivation](#motivation)
  - [Goals](#goals)
  - [Non-Goals](#non-goals)
- [Proposal](#proposal)
- [Design Details](#design-details)
  - [API Changes](#api-changes)
  - [Effective-Value Resolution](#effective-value-resolution)
  - [Recommender Integration](#recommender-integration)
  - [Interaction with Lower and Upper Bounds](#interaction-with-lower-and-upper-bounds)
  - [Validation](#validation)
  - [Feature Enablement and Rollback](#feature-enablement-and-rollback)
  - [Version Skew](#version-skew)
  - [Kubernetes Version Compatibility](#kubernetes-version-compatibility)
- [Test Plan](#test-plan)
- [Examples](#examples)
  - [Latency-sensitive service: aggressive CPU target](#latency-sensitive-service-aggressive-cpu-target)
  - [Batch workload: relaxed range for all containers](#batch-workload-relaxed-range-for-all-containers)
- [Future Work](#future-work)
- [Alternatives Considered](#alternatives-considered)
- [Implementation History](#implementation-history)
<!-- /toc -->

## Summary

Add optional per-container percentile overrides to `ContainerResourcePolicy`: the lower-bound, target, and upper-bound percentiles for CPU and memory, each overriding the Recommender's corresponding global `--*-percentile` flag for that container. This extends the per-VPA configuration mechanism of [AEP-8026](../8026-per-vpa-component-configuration/README.md) to the recommendation percentiles: the fields live on `ContainerResourcePolicy`, are gated behind the `PerVPAConfig` feature gate, and fall back to the global flags when unset. For a resource, the three percentiles are set together so the admission controller can enforce `lower ≤ target ≤ upper`.

## Motivation

The recommendation percentiles are among the most workload-dependent knobs the Recommender has: a latency-sensitive service may want an aggressive p95 CPU target, while a batch job on the same cluster is happy at p50. The lower- and upper-bound percentiles matter just as much — they set the range the Updater treats as acceptable, i.e. how much drift a workload tolerates before it is resized. Today all of these are cluster-wide Recommender flags, so operators either pick compromise values for the whole cluster or run separate Recommender instances per profile via [AEP-3919](../3919-customized-recommender-vpa/README.md), with all the operational overhead that brings.

The three percentiles for a resource are configured as a group. A per-VPA target on its own could drift outside the cluster-wide bounds, which the Updater then cannot act on coherently; setting lower/target/upper together keeps the range consistent and lets admission validate it.

### Goals

- Override the lower-bound, target, and upper-bound percentiles per container (CPU and memory) via `ContainerResourcePolicy`, including through `containerName: "*"`.
- Match the global flags' semantics: each percentile affects only its own resource and role.
- Keep the feature additive and off by default: a resource whose percentiles are unset falls back to the global flags.

### Non-Goals

- **Per-VPA overrides for other recommender parameters** (histogram decay, confidence, and similar) are out of scope.
- **Changing the recommendation model.** The histogram, decay, and confidence computations are untouched; only the percentiles at which the recommendations are read change.

## Proposal

Add six optional fields to `ContainerResourcePolicy` (autoscaling.k8s.io/v1) — the lower-bound, target, and upper-bound percentiles for CPU and memory:

```go
// The six fields below override this container's recommendation percentiles.
// When set, each overrides the Recommender's corresponding global
// --*-percentile flag. Values are integer percentiles in [1, 100] (e.g. 95
// for p95). Only honored when the PerVPAConfig feature gate is enabled.
// For a resource, the lower-bound, target, and upper-bound percentiles must
// be set together and satisfy lower <= target <= upper.

// +optional
// +kubebuilder:validation:Minimum=1
// +kubebuilder:validation:Maximum=100
LowerBoundCPUPercentile *int32 `json:"lowerBoundCPUPercentile,omitempty"`
// +optional
// +kubebuilder:validation:Minimum=1
// +kubebuilder:validation:Maximum=100
TargetCPUPercentile *int32 `json:"targetCPUPercentile,omitempty"`
// +optional
// +kubebuilder:validation:Minimum=1
// +kubebuilder:validation:Maximum=100
UpperBoundCPUPercentile *int32 `json:"upperBoundCPUPercentile,omitempty"`

// +optional
// +kubebuilder:validation:Minimum=1
// +kubebuilder:validation:Maximum=100
LowerBoundMemoryPercentile *int32 `json:"lowerBoundMemoryPercentile,omitempty"`
// +optional
// +kubebuilder:validation:Minimum=1
// +kubebuilder:validation:Maximum=100
TargetMemoryPercentile *int32 `json:"targetMemoryPercentile,omitempty"`
// +optional
// +kubebuilder:validation:Minimum=1
// +kubebuilder:validation:Maximum=100
UpperBoundMemoryPercentile *int32 `json:"upperBoundMemoryPercentile,omitempty"`
```

When set, the Recommender reads each recommendation (lower bound, target, upper bound) for that container at the declared percentile instead of the global flag's; the rest of the pipeline is unchanged.

## Design Details

### API Changes

Only `ContainerResourcePolicy` changes. No status, condition, or metric changes: the effective percentiles are fully determined by the spec and the Recommender flags, and the resulting recommendation is already observable in `status.recommendation`.

Each field is a plain integer percentile (`*int32`, `[1, 100]`) rather than a `resource.Quantity` — it's a unit, not a resource quantity. This keeps per-field validation to a simple `Minimum`/`Maximum` on the CRD. The Recommender divides the value by 100 to get the `(0, 1]` fraction its estimators use, matching the global flags.

### Effective-Value Resolution

Percentiles resolve per resource, with the standard Phase 1 precedence:

1. `containerPolicies` entry matching the container's name, if it sets the resource's percentiles.
2. `containerPolicies` entry with `containerName: "*"`, if it sets them.
3. The corresponding global Recommender flags.

CPU and memory resolve independently: a policy may set the CPU triple and inherit the global memory percentiles, or vice versa.

### Recommender Integration

Today the percentile estimators are constructed once, at Recommender startup, with the global flag values baked in (e.g. `NewPercentileCPUEstimator(config.TargetCPUPercentile)` in `pkg/recommender/logic/recommender.go`). A single construction-time value cannot express per-container percentiles.

The estimators become parameterized: the effective percentiles are carried on the `AggregateContainerState` — the same vehicle Phase 1 uses for `OOMBumpUpRatio` — and the lower-bound, target, and upper-bound estimators for each resource read their percentile at estimation time, falling back to the construction-time global value when no per-container override is present.

`AggregateContainerState` already receives the VPA's `ContainerResourcePolicy` during aggregation, so populating the effective percentiles alongside `OOMBumpUpRatio` needs no new plumbing between the API and model layers.

### Interaction with Lower and Upper Bounds

The lower- and upper-bound percentiles drive the lower/upper bound recommendations, which define the range the Updater treats as acceptable: current usage outside it triggers a resize. Making them per-VPA lets a workload pick its own eviction sensitivity — a wide range for a tolerant batch job, a tight one for a latency-sensitive service. Because the admission controller enforces `lower ≤ target ≤ upper` (see [Validation](#validation)), a container's percentiles are always internally consistent.

### Validation

Each field is an integer in `[1, 100]`, enforced by the CRD schema:

```go
// +kubebuilder:validation:Minimum=1
// +kubebuilder:validation:Maximum=100
```

The admission webhook (`pkg/admission-controller/resource/vpa/validation.go`) enforces the cross-field rules the CRD schema cannot express:

- For each resource (CPU, memory), the lower-bound, target, and upper-bound percentiles are either all set or all unset.
- When set, `lower ≤ target ≤ upper`.

Requiring the three together lets admission validate the ordering without knowing the Recommender's global flag values. The webhook also rejects the fields when the `PerVPAConfig` gate is disabled, matching the Phase 1 fields.

### Feature Enablement and Rollback

Feature gate: **`PerVPAConfig`** (existing, introduced by AEP-8026). No new gate.

- **Enabled:** the admission controller accepts the fields on new/updated VPAs; the Recommender honours them.
- **Disabled:** the admission controller rejects new VPAs that set the fields with a descriptive error; the Recommender ignores the fields on existing objects and uses the global flags (fail-open, identical to the Phase 1 fields' rollback semantics).

### Version Skew

Only the Recommender consumes the fields; the Updater and the Admission Controller (beyond validation) are unaffected. An older Recommender ignores the fields entirely and applies the global flags — the same behaviour as the gate-disabled path. No skew combination causes errors or corrupted state.

### Kubernetes Version Compatibility

The feature is entirely internal to the VPA controllers and depends on no new Kubernetes APIs. It is compatible with any Kubernetes version supported by the corresponding VPA release.

## Test Plan

**Unit tests:**

- Effective-value resolution: named-container policy wins over `"*"`, which wins over the global flags; CPU and memory resolve independently.
- Estimator behaviour: each of the lower/target/upper estimators uses the per-container percentile when present on the `AggregateContainerState` and the global value otherwise.
- Validation: values outside `[1, 100]` rejected; a partially-set resource triple rejected; `lower > target` or `target > upper` rejected; fields rejected when the gate is disabled.

**Integration tests** (Recommender):

- Two VPAs with identical usage histories, one setting the CPU triple with a low target and one unset — the first receives a lower CPU target; memory is unchanged.
- The equivalent scenario for the memory triple.
- Fields set via `containerName: "*"` apply to all containers not covered by a named policy.
- Gate-disabled path: fields on an existing object are ignored and the global flags apply.

## Examples

### Latency-sensitive service: aggressive CPU target

```yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: api-gateway-vpa
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: api-gateway
  resourcePolicy:
    containerPolicies:
    - containerName: gateway
      lowerBoundCPUPercentile: 60
      targetCPUPercentile: 95
      upperBoundCPUPercentile: 98
```

The `gateway` container's CPU recommendations are read at these percentiles instead of the cluster defaults; its memory percentiles are unchanged.

### Batch workload: relaxed range for all containers

```yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: nightly-report-vpa
spec:
  targetRef:
    apiVersion: batch/v1
    kind: CronJob
    name: nightly-report
  resourcePolicy:
    containerPolicies:
    - containerName: "*"
      lowerBoundCPUPercentile: 25
      targetCPUPercentile: 50
      upperBoundCPUPercentile: 75
      lowerBoundMemoryPercentile: 25
      targetMemoryPercentile: 50
      upperBoundMemoryPercentile: 75
```

Every container targets the median with a wide bound range, trading headroom for density and tolerating more drift before a resize.

## Future Work

- **Loosen the all-set requirement.** The three percentiles for a resource must currently be set together so admission can validate `lower ≤ target ≤ upper` without reading the Recommender's flags. If a safe way to validate a partially-set triple against the global flags emerges, this could be relaxed.

## Alternatives Considered

**1. Multiple Recommender instances (AEP-3919).** The status quo escape hatch: run one Recommender per percentile profile and point each VPA at one. Works, but each additional Recommender is another deployment to size, monitor, and upgrade, and workloads must be partitioned into a small number of static profiles. Differing percentiles are the canonical reason operators end up here; making the percentiles declarative removes the most common need for the pattern.

## Implementation History

- (issue filed) 2026-07-11 — Issue [kubernetes/autoscaler#9970](https://github.com/kubernetes/autoscaler/issues/9970).
- (scope agreed) 2026-07-14 — SIG feedback on the issue: start with the target percentiles.
- (scope expanded) 2026-08-28 — PR review: include the lower-bound, target, and upper-bound percentiles as a required set per resource.
- (initial implementation) TBD.
