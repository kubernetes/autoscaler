# AEP-9970: Per-VPA Target Percentiles

<!-- toc -->
- [Summary](#summary)
- [Motivation](#motivation)
  - [Goals](#goals)
  - [Non-Goals](#non-goals)
- [Proposal](#proposal)
- [Design Details](#design-details)
  - [API Changes](#api-changes)
  - [Why Three Percentiles Per Resource](#why-three-percentiles-per-resource)
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

- **Per-VPA overrides for other recommender parameters** are out of scope.
- **Changing the recommendation model.** The histogram, decay, and confidence computations are untouched; only the percentiles at which the recommendations are read change.

## Proposal

Add one optional field to `ContainerResourcePolicy` (autoscaling.k8s.io/v1), grouping the lower-bound, target, and upper-bound percentiles per resource:

```go
// recommendationPercentiles overrides this container's recommendation
// percentiles, replacing the Recommender's global --*-percentile flags.
// Set per resource (cpu, memory); within a resource the three percentiles
// are required together and must satisfy lowerBound <= target <= upperBound.
// Only honored when the PerVPAConfig feature gate is enabled.
// +optional
RecommendationPercentiles *RecommendationPercentiles `json:"recommendationPercentiles,omitempty"`

type RecommendationPercentiles struct {
	// +optional
	CPU *ResourcePercentiles `json:"cpu,omitempty"`
	// +optional
	Memory *ResourcePercentiles `json:"memory,omitempty"`
}

// Each percentile is an integer in [1, 100] (e.g. 95 for p95).
// +kubebuilder:validation:XValidation:rule="self.lowerBound <= self.target && self.target <= self.upperBound",message="percentiles must satisfy lowerBound <= target <= upperBound"
type ResourcePercentiles struct {
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=100
	LowerBound int32 `json:"lowerBound"`
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=100
	Target int32 `json:"target"`
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=100
	UpperBound int32 `json:"upperBound"`
}
```

When set, the Recommender reads each recommendation (lower bound, target, upper bound) for that container at the declared percentile instead of the global flag's; the rest of the pipeline is unchanged.

## Design Details

### API Changes

Only `ContainerResourcePolicy` changes. No status, condition, or metric changes: the effective percentiles are fully determined by the spec and the Recommender flags, and the resulting recommendation is already observable in `status.recommendation`.

Each percentile is a plain integer (`int32`, `[1, 100]`) rather than a `resource.Quantity` — it's a unit, not a resource quantity. Grouping the three into `ResourcePercentiles` lets the API schema carry the invariants directly: the three are required together (non-pointer fields of an optional struct), bounded by `Minimum`/`Maximum`, and ordered by a CEL rule on the struct. The Recommender divides each value by 100 to get the `(0, 1]` fraction its estimators use, matching the global flags.

### Why three percentiles per resource

The Recommender already computes three separate values for each resource, each read at its own percentile from a distinct global flag:

| Recommendation | Global flag | Default | Drives |
|---|---|---|---|
| lower bound | `--recommendation-lower-bound-{cpu,memory}-percentile` | 0.5 | bottom of the range the Updater tolerates |
| target | `--target-{cpu,memory}-percentile` | 0.9 | the request injected at admission and aimed for by the Updater |
| upper bound | `--recommendation-upper-bound-{cpu,memory}-percentile` | 0.95 | top of the range the Updater tolerates |

These are three independent knobs, not a min/max pair around one value. The target is what actually gets applied; the lower and upper bounds define the range the Updater treats as acceptable — it evicts a pod whose current request falls below the lower bound or above the upper bound (`pkg/updater/priority/priority_processor.go`). So the per-VPA API mirrors the three existing flags: one override each.

Overriding only the target is not enough, and can produce an incoherent config. With the bounds left at their global percentiles, a per-VPA target can land outside them: e.g. a target at the 99th percentile while the global upper bound stays at the 95th gives `target > upperBound`. The Updater then reads every pod's request (set from the target) as above the upper bound and evicts it on every pass.

Exposing all three lets an operator shift the whole range together, and lets the API reject inconsistent configs. For a resource the three are required together and must satisfy `lower ≤ target ≤ upper` (see [Validation](#validation)); a resource whose percentiles are unset falls back entirely to the global flags.

### Effective-Value Resolution

Percentiles resolve per resource:

1. `containerPolicies` entry matching the container's name, if it sets the resource's percentiles.
2. `containerPolicies` entry with `containerName: "*"`, if it sets them.
3. The corresponding global Recommender flags.

CPU and memory resolve independently: a policy may set the CPU triple and inherit the global memory percentiles, or vice versa.

### Recommender Integration

Today the percentile estimators are constructed once, at Recommender startup, with the global flag values baked in (e.g. `NewPercentileCPUEstimator(config.TargetCPUPercentile)` in `pkg/recommender/logic/recommender.go`). A single construction-time value cannot express per-container percentiles.

The estimators become parameterized: the effective percentiles are carried on the `AggregateContainerState` — the same vehicle Phase 1 uses for `OOMBumpUpRatio` — and the lower-bound, target, and upper-bound estimators for each resource read their percentile at estimation time, falling back to the construction-time global value when no per-container override is present.

`AggregateContainerState` already receives the VPA's `ContainerResourcePolicy` during aggregation, so populating the effective percentiles alongside `OOMBumpUpRatio` needs no new plumbing between the API and model layers.

### Interaction with Lower and Upper Bounds

The lower- and upper-bound percentiles drive the lower/upper bound recommendations, which define the range the Updater treats as acceptable: current usage outside it triggers a resize. Making them per-VPA lets a workload pick its own eviction sensitivity — a wide range for a tolerant batch job, a tight one for a latency-sensitive service. Because the API enforces `lower ≤ target ≤ upper` (see [Validation](#validation)), a container's percentiles are always internally consistent.

### Validation

The invariants are enforced by the API schema, so any object that reaches the Recommender or Updater is already well-formed:

- Range: each percentile is an integer in `[1, 100]` (`Minimum`/`Maximum`).
- All-or-none per resource: `lowerBound`, `target` and `upperBound` are required fields of `ResourcePercentiles`, so a resource is either fully specified or omitted.
- Ordering: a CEL rule on `ResourcePercentiles` requires `lowerBound ≤ target ≤ upperBound`.

```go
// +kubebuilder:validation:XValidation:rule="self.lowerBound <= self.target && self.target <= self.upperBound",message="percentiles must satisfy lowerBound <= target <= upperBound"
```

The admission webhook (`pkg/admission-controller/resource/vpa/validation.go`) only gates the feature: it rejects `recommendationPercentiles` when the `PerVPAConfig` gate is disabled, matching the Phase 1 fields. It does not re-check the invariants above.

### Feature Enablement and Rollback

Feature gate: **`PerVPAConfig`** (existing, introduced by AEP-8026). No new gate.

- **Enabled:** the admission controller accepts the fields on new/updated VPAs; the Recommender honours them.
- **Disabled:** the admission controller rejects new VPAs that set `recommendationPercentiles` with a descriptive error; the Recommender ignores it on existing objects and uses the global flags (fail-open, identical to the Phase 1 fields' rollback semantics).

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
      recommendationPercentiles:
        cpu:
          lowerBound: 60
          target: 95
          upperBound: 98
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
      recommendationPercentiles:
        cpu:
          lowerBound: 25
          target: 50
          upperBound: 75
        memory:
          lowerBound: 25
          target: 50
          upperBound: 75
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
