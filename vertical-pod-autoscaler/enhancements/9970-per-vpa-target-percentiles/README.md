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
  - [Batch workload: relaxed targets for all containers](#batch-workload-relaxed-targets-for-all-containers)
- [Future Work](#future-work)
- [Alternatives Considered](#alternatives-considered)
- [Implementation History](#implementation-history)
<!-- /toc -->

## Summary

Add two optional per-container fields to `ContainerResourcePolicy` — `targetCPUPercentile` and `targetMemoryPercentile` — that override the Recommender's global `--target-cpu-percentile` and `--target-memory-percentile` flags for that container. This extends the per-VPA configuration mechanism introduced by [AEP-8026](../8026-per-vpa-component-configuration/README.md) to the target recommendation percentiles, following the same conventions: fields live on `ContainerResourcePolicy`, use `resource.Quantity` values, are gated behind the `PerVPAConfig` feature gate, and fall back to the corresponding global flag when unset.

## Motivation

The target percentiles are among the most workload-dependent knobs the Recommender has: a latency-sensitive service may want an aggressive p95 CPU target, while a batch job on the same cluster is happy at p50. Today both values are cluster-wide Recommender flags (defaulting to 0.9), so operators either pick compromise values for the whole cluster or run separate Recommender instances per profile via [AEP-3919](../3919-customized-recommender-vpa/README.md) — differing percentiles being the canonical motivation for running multiple recommenders, with all the operational overhead that brings.

Per [SIG discussion on the tracking issue](https://github.com/kubernetes/autoscaler/issues/9970), the scope is deliberately limited to the two target percentiles. The lower- and upper-bound percentiles (which drive the Updater's eviction decisions) are deferred until there is user feedback on the target fields — see [Future Work](#future-work).

### Goals

- Allow overriding `--target-cpu-percentile` and `--target-memory-percentile` per container via `ContainerResourcePolicy`, with per-VPA configuration expressible through `containerName: "*"`.
- Match the global flags' semantics exactly: the CPU target percentile affects only the CPU target recommendation (not CPU bounds, not memory), and likewise for memory.
- Each field independently falls back to its global flag when unset, keeping the feature additive and off by default.

### Non-Goals

- **Other per-VPA recommender parameters** (such as lower/upper bound percentiles) are out of scope.
- **Changing the recommendation model.** The histogram, decay, and confidence computations are untouched; only the percentile at which the target is read changes.

## Proposal

Add two optional fields to `ContainerResourcePolicy` (autoscaling.k8s.io/v1):

```go
// TargetCPUPercentile, when set, overrides the global
// --target-cpu-percentile flag for this container: the CPU usage
// percentile used as the base for the CPU target recommendation.
// Expressed as a Quantity in (0, 1], e.g. "0.95" or "950m" for p95.
// Falls back to the global flag when unset.
// Only honored when the PerVPAConfig feature gate is enabled.
// +optional
TargetCPUPercentile *resource.Quantity `json:"targetCPUPercentile,omitempty"`

// TargetMemoryPercentile, when set, overrides the global
// --target-memory-percentile flag for this container: the memory usage
// percentile used as the base for the memory target recommendation.
// Expressed as a Quantity in (0, 1], e.g. "0.95" or "950m" for p95.
// Falls back to the global flag when unset.
// Only honored when the PerVPAConfig feature gate is enabled.
// +optional
TargetMemoryPercentile *resource.Quantity `json:"targetMemoryPercentile,omitempty"`
```

Behaviour, in one sentence: **when set, the Recommender reads the target recommendation for that container at the declared percentile instead of the global flag's percentile; everything else about the recommendation pipeline is unchanged.**

## Design Details

### API Changes

Only `ContainerResourcePolicy` changes, as shown in [Proposal](#proposal). No status, condition, or metric changes: the effective percentile is fully determined by the spec and the Recommender flags, and the resulting recommendation is already observable in `status.recommendation`.

`resource.Quantity` is used instead of a float for consistency with `oomBumpUpRatio`.

### Effective-Value Resolution

The effective percentile for a container is resolved with the standard Phase 1 precedence:

1. `containerPolicies` entry matching the container's name, if it sets the field.
2. `containerPolicies` entry with `containerName: "*"`, if it sets the field.
3. The corresponding global Recommender flag.

Each field resolves independently: a policy may set only `targetMemoryPercentile` and inherit the global CPU percentile, or vice versa.

### Recommender Integration

Today the percentile estimators are constructed once, at Recommender startup, with the global flag values baked in (`NewPercentileCPUEstimator(config.TargetCPUPercentile)` in `pkg/recommender/logic/recommender.go`). A single construction-time value cannot express per-container percentiles.

The target estimators become parameterized: the effective target percentile is carried on the `AggregateContainerState` — the same vehicle Phase 1 uses for `OOMBumpUpRatio` — and the CPU/memory target estimators read it at estimation time from the state passed to `GetResourceEstimation`, using the construction-time global value as the fallback when no per-container override is present. The lower- and upper-bound estimators are untouched and keep their construction-time global percentiles.

The `AggregateContainerState` already receives the VPA's `ContainerResourcePolicy` during aggregation, so populating the two effective percentiles alongside `OOMBumpUpRatio` requires no new plumbing between the API layer and the model layer.

### Interaction with Lower and Upper Bounds

The bound percentiles remain global, so a per-VPA target percentile can be configured above the global upper-bound percentile (or below the lower-bound one), producing a target outside the `[lowerBound, upperBound]` interval. This is not a new failure mode — the global flags permit the same misordering today, and no ordering validation exists between them — but per-VPA configuration makes it easier to reach accidentally.

The admission controller therefore emits a warning (not a rejection) when a declared target percentile is above the global upper-bound percentile or below the global lower-bound percentile at validation time.

### Validation

Each field must be greater than `0` and at most `1` (`(0, 1]`), matching the value range the global flags accept. The bound is enforced at the CRD level with a CEL validation rule:

```go
// +kubebuilder:validation:XValidation:rule="self > quantity('0') && self <= quantity('1')",message="percentile must be in (0, 1]"
```

The admission controller's VPA validation (`pkg/admission-controller/resource/vpa/validation.go`) validates the same range to return more descriptive error messages, and is also where the fields are rejected when the `PerVPAConfig` feature gate is disabled — matching the Phase 1 fields' handling.

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

- Effective-value resolution: named-container policy wins over `"*"`, which wins over the global flag; each field resolves independently.
- Estimator behaviour: target estimation uses the per-container percentile when present on the `AggregateContainerState` and the global value otherwise; bound estimations are unaffected by the fields.
- Validation: values outside `(0, 1]` rejected; fields rejected when the feature gate is disabled; warning emitted when a target percentile lies outside the global bound percentiles.

**Integration tests** (Recommender):

- Two VPAs targeting identical workloads with identical usage histories, one with `targetCPUPercentile: "0.5"` and one unset — the first receives a lower CPU target; memory targets are identical.
- The equivalent scenario for `targetMemoryPercentile`.
- A VPA setting the fields via `containerName: "*"` applies them to all containers not covered by a named policy.
- Feature-gate-disabled path: fields present on an existing object are ignored and the global flags apply.

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
      targetCPUPercentile: "0.95"
```

The `gateway` container's CPU target is read at p95 instead of the cluster default (p90 unless the flag is changed); its memory target and all bounds are unchanged.

### Batch workload: relaxed targets for all containers

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
      targetCPUPercentile: "0.5"
      targetMemoryPercentile: "0.5"
```

Every container in the job is targeted at the median, trading headroom for density on a throughput-insensitive workload.

## Future Work

- **Lower/upper bound percentiles per-VPA** (`lowerBoundCPUPercentile`, `upperBoundCPUPercentile`, and memory equivalents): the natural completion of this work, enabling per-workload eviction sensitivity. Deferred pending user feedback on the target fields, per SIG discussion on [#9970](https://github.com/kubernetes/autoscaler/issues/9970). Adding them also makes cross-field ordering validation (`lowerBound <= target <= upperBound`) meaningful.

## Alternatives Considered

**1. Multiple Recommender instances (AEP-3919).** The status quo escape hatch: run one Recommender per percentile profile and point each VPA at one. Works, but each additional Recommender is another deployment to size, monitor, and upgrade, and workloads must be partitioned into a small number of static profiles. Differing percentiles are the canonical reason operators end up here; making the percentile declarative removes the most common need for the pattern.

**2. Float fields instead of `resource.Quantity`.** More natural to read (`0.95` vs `"950m"`), but the Kubernetes API conventions discourage floating-point fields, and Phase 1 already established `Quantity` for fractional per-VPA values (`oomBumpUpRatio`). Consistency wins.

## Implementation History

- (issue filed) 2026-07-11 — Issue [kubernetes/autoscaler#9970](https://github.com/kubernetes/autoscaler/issues/9970).
- (scope agreed) 2026-07-14 — SIG feedback on the issue: limit to the two target percentiles.
- (AEP PR opened) TBD.
- (initial implementation) TBD.
