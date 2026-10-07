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
  - [Batch workload: median target for all containers](#batch-workload-median-target-for-all-containers)
- [Future Work](#future-work)
- [Alternatives Considered](#alternatives-considered)
- [Implementation History](#implementation-history)
<!-- /toc -->

## Summary

Add an optional per-container target percentile override to `ContainerResourcePolicy`, for CPU and memory, replacing the Recommender's global `--target-{cpu,memory}-percentile` flag for that container. This extends the per-VPA configuration mechanism of [AEP-8026](../8026-per-vpa-component-configuration/README.md): the field lives on `ContainerResourcePolicy`, is gated behind the `PerVPAConfig` feature gate, and falls back to the global flag when unset.

The lower and upper bound percentiles stay global. If a per-VPA target falls outside them, the Recommender sets a new `ConfigInvalid` condition on the VPA and the Updater stops updating its pods.

## Motivation

The target percentile is one of the most workload-dependent knobs the Recommender has: a latency-sensitive service may want a p95 CPU target, while a batch job on the same cluster is fine at p50. Today it is a cluster-wide flag, so operators either pick a compromise value for the whole cluster or run separate Recommender instances per profile via [AEP-3919](../3919-customized-recommender-vpa/README.md), with all the operational overhead that brings.

### Goals

- Override the target percentile per container (CPU and memory) via `ContainerResourcePolicy`, including through `containerName: "*"`.
- Match the global flags' semantics: each target percentile affects only its own resource.
- Keep the feature additive and off by default: a resource whose target is unset falls back to the global flag.
- Never let a per-VPA target cause an eviction loop.

### Non-Goals

- **Per-VPA lower and upper bound percentiles.** See [Future Work](#future-work).
- **Per-VPA overrides for other recommender parameters.**
- **Changing the recommendation model.** The histogram, decay, and confidence computations are untouched; only the percentile at which the target is read changes.

## Proposal

Add one optional field to `ContainerResourcePolicy` (autoscaling.k8s.io/v1):

```go
// recommendationPercentiles overrides the recommender's target percentile for
// this container, per resource (cpu, memory), replacing the corresponding
// global Recommender flag. The lower and upper bounds keep using the global
// percentiles; if a target falls outside them, the Recommender sets the
// ConfigInvalid condition and the Updater does not update the VPA's pods.
// Only honored when the PerVPAConfig feature gate is enabled.
// +optional
RecommendationPercentiles *RecommendationPercentiles `json:"recommendationPercentiles,omitempty"`

type RecommendationPercentiles struct {
	// +optional
	CPU *ResourcePercentiles `json:"cpu,omitempty"`
	// +optional
	Memory *ResourcePercentiles `json:"memory,omitempty"`
}

type ResourcePercentiles struct {
	// target is the usage percentile used for the target recommendation, as an
	// integer in [1, 100] (e.g. 95 for p95).
	// +required
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=100
	Target int32 `json:"target"`
}
```

And one condition type:

```go
// ConfigInvalid indicates that this VPA configuration is invalid for the
// recommender serving it, for example a per-VPA target percentile outside
// the recommender's lower and upper bound percentiles. Recommendations are
// still provided, but the updater does not update the VPA's pods.
ConfigInvalid VerticalPodAutoscalerConditionType = "ConfigInvalid"
```

The per-resource struct leaves room to add lower and upper bound percentiles later without another API shape.

## Design Details

### API Changes

`ContainerResourcePolicy` gets the `recommendationPercentiles` field, and `VerticalPodAutoscalerConditionType` gets `ConfigInvalid`, in the same family as `ConfigDeprecated` and `ConfigUnsupported`.

The target is a plain integer (`int32`, `[1, 100]`) rather than a `resource.Quantity`, since it's a percentile and not a resource amount. The Recommender divides it by 100 to get the `(0, 1]` fraction its estimators use, matching the global flags.

### Effective-Value Resolution

Targets resolve per resource:

1. `containerPolicies` entry matching the container's name, if it sets the resource's target.
2. `containerPolicies` entry with `containerName: "*"`, if it sets it.
3. The global `--target-{cpu,memory}-percentile` flag.

CPU and memory resolve independently.

### Recommender Integration

Today the percentile estimators are built once at startup with the global flag values (e.g. `NewPercentileCPUEstimator(config.TargetCPUPercentile)` in `pkg/recommender/logic/recommender.go`), which can't express a per-container value.

The target estimators become parameterized: the effective target is carried on the `AggregateContainerState`, the same way as `OOMBumpUpRatio`, and the target estimator reads it at estimation time, falling back to the global value when unset. `AggregateContainerState` already receives the VPA's `ContainerResourcePolicy`, so no new plumbing is needed between the API and model layers. The lower and upper bound estimators are unchanged.

### Interaction with Lower and Upper Bounds

The Updater evicts a pod when its current request is below the lower bound or above the upper bound recommendation (`pkg/updater/priority/priority_processor.go`). The target is what gets applied to the pod. So if the target is read at a percentile above the upper bound percentile (or below the lower one), every pod resized to the target is immediately outside the bounds again, and the Updater evicts it on every pass.

This can happen today with the global flags alone, since nothing checks their order. A per-VPA target makes it easier to hit: a target of 99 with the default upper bound of 95 is enough.

The check needs the Recommender's flags, so it can't be done in the API schema or the admission webhook. Instead:

- **Recommender:** on each status update, it compares every per-container target against its own `--recommendation-lower-bound-*-percentile` and `--recommendation-upper-bound-*-percentile` flags. If one is out of range, it sets `ConfigInvalid=True` with a message naming the container and resource, e.g. `container "app": cpu target percentile 99 is outside the recommender's bound percentiles [50, 95]`. Otherwise it removes the condition. Recommendations are still computed and published.
- **Updater:** skips VPAs with `ConfigInvalid=True`, the same way it skips `Off` and `Initial` modes. No evictions or in-place updates happen until the target is fixed.

The check is based only on percentiles, so it doesn't need to wait for any history. The condition shows up on the first Recommender loop after the VPA is created or changed. The Admission Controller still applies the recommendation to new pods, so pods created by the workload itself (scale-up, rollout) get the per-VPA target.

### Validation

The API schema enforces the range: `target` is required and must be in `[1, 100]`.

The admission webhook (`pkg/admission-controller/resource/vpa/validation.go`) only gates the feature: it rejects `recommendationPercentiles` when `PerVPAConfig` is disabled.

The ordering against the bound percentiles is checked by the Recommender, as described above.

### Feature Enablement and Rollback

Feature gate: **`PerVPAConfig`** (existing, introduced by AEP-8026). No new gate.

- **Enabled:** the admission controller accepts the field; the Recommender honours it and sets or clears `ConfigInvalid`.
- **Disabled:** the admission controller rejects new VPAs that set `recommendationPercentiles`. The Recommender ignores the field on existing objects, uses the global flags, and removes any `ConfigInvalid` condition it set earlier.

### Version Skew

- **Older Recommender:** ignores the field and never sets `ConfigInvalid`. Same as the gate-disabled path.
- **Older Updater:** doesn't know about `ConfigInvalid`, so it keeps updating pods even when the target is out of bounds, which can cause the eviction loop above. Upgrade the Updater together with or before the Recommender when using this field.
- **Older Admission Controller:** doesn't know the field, so it doesn't gate it. The Recommender's gate still applies.

### Kubernetes Version Compatibility

The feature is internal to the VPA controllers and depends on no new Kubernetes APIs.

## Test Plan

**Unit tests:**

- Effective-value resolution: the target is set from the policy, reset when removed, and ignored when the gate is disabled; CPU and memory resolve independently.
- Estimator behaviour: the target estimator uses the per-container percentile when set and the global value otherwise; the lower and upper bound estimators ignore it.
- Recommender: `ConfigInvalid` is set for a target above the upper bound or below the lower bound, not set at or inside the bounds, cleared once the target is fixed, and cleared when the gate is disabled.
- Updater: no evictions for a VPA with `ConfigInvalid=True`; normal evictions when it's `False` or absent.
- Validation: the field is rejected when the gate is disabled.

**E2E tests:**

- A VPA with a CPU target inside the bounds gets a different CPU target than an identical VPA without the override; memory is unchanged.
- A VPA with a target outside the bounds gets `ConfigInvalid=True` and its pods are not evicted.

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
          target: 95
```

The `gateway` container's CPU target is read at p95 instead of the cluster default. With the default upper bound percentile of 95 this is valid; a target of 99 would set `ConfigInvalid`.

### Batch workload: median target for all containers

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
          target: 50
        memory:
          target: 50
```

Every container targets the median, trading headroom for density.

## Future Work

- **Per-VPA lower and upper bounds.** Let a VPA set its own bound percentiles, either as explicit `lowerBound`/`upperBound` fields next to `target` or as a deviation around the target. This would let a workload choose how much drift it tolerates before a resize, and avoid `ConfigInvalid` for targets near the global bounds.
- **Checking the global flags at startup.** The Recommender could reject flag values where the target is outside the bounds, which would also catch the existing misconfiguration without per-VPA settings.

## Alternatives Considered

**1. Multiple Recommender instances (AEP-3919).** The current workaround: run one Recommender per percentile profile and point each VPA at one. Works, but each extra Recommender is another deployment to size, monitor, and upgrade, and workloads have to be split into a few static profiles.

**2. Require lower, target and upper together.** An earlier version of this AEP had all three percentiles per resource, required together, with a CEL rule enforcing `lowerBound <= target <= upperBound`. This keeps every VPA consistent at admission time, but makes the common case (just change the target) more verbose, and adds two fields most users don't need. It was dropped in review in favour of the target-only field with a Recommender-side check; it can still be added later (see [Future Work](#future-work)).

**3. Validate in the admission webhook.** The webhook doesn't know the Recommender's flags, and a cluster can run several Recommenders with different flags (AEP-3919), so only the Recommender serving the VPA can check the target against its bounds.

## Implementation History

- (issue filed) 2026-07-11 — Issue [kubernetes/autoscaler#9970](https://github.com/kubernetes/autoscaler/issues/9970).
- (scope agreed) 2026-07-14 — SIG feedback on the issue: start with the target percentiles.
- (scope expanded) 2026-08-28 — PR review: include the lower-bound, target, and upper-bound percentiles as a required set per resource.
- (scope reduced) 2026-10-07 — Implementation review: back to target-only, with the Recommender setting `ConfigInvalid` and the Updater skipping invalid VPAs.
