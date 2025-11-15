# AEP-8818: Non-Disruptive In-Place Updates in VPA

<!-- toc -->
- [Summary](#summary)
- [Motivation](#motivation)
- [Goals](#goals)
- [Non-Goals](#non-goals)
- [Proposal](#proposal)
- [Design Details](#design-details)
- [Test Plan](#test-plan)
- [Upgrade / Downgrade Strategy](#upgrade--downgrade-strategy)
  - [Upgrade](#upgrade)
  - [Downgrade](#downgrade)
- [Feature Enablement and Rollback](#feature-enablement-and-rollback)
  - [How can this feature be enabled / disabled in a live cluster?](#how-can-this-feature-be-enabled--disabled-in-a-live-cluster)
- [Kubernetes version compatibility](#kubernetes-version-compatibility)
- [Implementation History](#implementation-history)
<!-- /toc -->

## Summary

[AEP-4016](https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler/enhancements/4016-in-place-updates-support) introduced the `InPlaceOrRecreate` update mode which attempts in-place updates first but falls back to pod eviction if the in-place update fails. However, for certain workloads, any disruption is unacceptable, and users would prefer to retry in-place updates indefinitely rather than evict and recreate pods.

This proposal introduces a new update mode that only attempts in-place updates and retries on failure without ever falling back to eviction.

## Motivation

There are several use cases where pod disruption should be avoided at all costs:

- Stateful workloads: Pods managing critical state where restart would cause data loss or lengthy recovery.
- Long-running computations: Jobs or services performing computations that cannot be checkpointed and would need to restart from the beginning.
- Strict SLO requirements: Services with stringent availability requirements where even brief disruptions are unacceptable.
In these scenarios, users would prefer:

- To operate with current (potentially suboptimal) resource allocations until an in-place update becomes feasible
- To receive clear signals when updates cannot be applied
- To have VPA continuously retry updates as cluster conditions change

## Goals

- Provide a truly non-disruptive VPA update mode that never evicts pods
- Allow VPA to eventually apply updates when cluster conditions improve
- Respect the existing in-place update infrastructure from AEP-4016

## Non-Goals

- Guarantee that all updates will eventually succeed (node capacity constraints may prevent this)
- Provide mechanisms to automatically increase node capacity to accommodate updates
- Change the behavior of existing update modes (Auto, Recreate, InPlaceOrRecreate)

## Proposal

Add a new supported value of UpdateMode: `InPlace`
This mode will:
- Apply recommendations during pod admission (like all other modes)
- Attempt in-place updates for running pods under the same conditions as `InPlaceOrRecreate`
- Never add pods to `podsForEviction` if in-place updates fail
- Continuously retry failed in-place update

## Design Details

Add `UpdateModeInPlace` to the VPA types:

```golang
// In pkg/apis/autoscaling.k8s.io/v1/types.go
const (
    // ... existing modes ...
    // UpdateModeInPlace means that VPA will only attempt to update pods in-place
    // and will never evict them. If in-place update fails, VPA will retry later.
    UpdateModeInPlace UpdateMode = "InPlace"
)
```

We will enhance `inplace_restriction.go` to support the new mode:

```golang
// Update CanInPlaceUpdate to accept update mode
func (ip *PodsInPlaceRestrictionImpl) CanInPlaceUpdate(pod *apiv1.Pod, updateMode vpa_types.UpdateMode) utils.InPlaceDecision {
    if !features.Enabled(features.InPlaceOrRecreate) {
        return utils.InPlaceEvict
    }

    cr, present := ip.podToReplicaCreatorMap[getPodID(pod)]
    if present {
        singleGroupStats, present := ip.creatorToSingleGroupStatsMap[cr]
        if pod.Status.Phase == apiv1.PodPending {
            return utils.InPlaceDeferred
        }
        if present {
            if isInPlaceUpdating(pod) {
                canEvict := CanEvictInPlacingPod(pod, singleGroupStats, ip.lastInPlaceAttemptTimeMap, ip.clock)
                if canEvict {
                    // For InPlace mode, never suggest eviction
                    if updateMode == vpa_types.UpdateModeInPlace {
                        return utils.InPlaceDeferred
                    }
                    return utils.InPlaceEvict
                }
                return utils.InPlaceDeferred
            }
            if singleGroupStats.isPodDisruptable() {
                return utils.InPlaceApproved
            }
        }
    }
    klog.V(4).InfoS("Can't in-place update pod, waiting for next loop", "pod", klog.KObj(pod))
    return utils.InPlaceDeferred
}
```

The retry logic is implicitly handled by the existing `CanInPlaceUpdate` decision system. Specifically:
- Deferred State: When `CanInPlaceUpdate` returns `utils.InPlaceDeferred`, the pod is skipped in the current loop and will be reconsidered in the next iteration
- Loop Frequency: The updater's main loop runs periodically (default every 1 minute), providing natural retry behavior
- Condition-Based Decisions: The `CanEvictInPlacingPod` function already tracks state via `lastInPlaceAttemptTimeMap`

```golang
for vpa, livePods := range controlledPods {
    // ... existing setup code ...

    podsForInPlace := make([]*apiv1.Pod, 0)
    podsForEviction := make([]*apiv1.Pod, 0)

    if updateMode == vpa_types.UpdateModeInPlace && inPlaceFeatureEnable {
        // New mode: only in-place, never evict
        podsForInPlace = u.getPodsUpdateOrder(filterNonInPlaceUpdatablePods(livePods, inPlaceLimiter), vpa)
        inPlaceUpdatablePodsCounter.Add(vpaSize, len(podsForInPlace))
    } // rest of the code

    // ... existing counters ...

for _, pod := range podsForInPlace {
    withInPlaceUpdatable = true
    decision := inPlaceLimiter.CanInPlaceUpdate(pod, updateMode)

    if decision == utils.InPlaceDeferred {
        klog.V(2).InfoS("In-place update deferred, will retry in next loop", "pod", klog.KObj(pod))
        continue
    } else if decision == utils.InPlaceEvict {
        // Only add to eviction list if NOT in InPlace mode
        if updateMode != vpa_types.UpdateModeInPlace {
            podsForEviction = append(podsForEviction, pod)
        } else {
            klog.V(2).InfoS("In-place update would require eviction, but InPlace mode prevents it. Will retry later.", "pod", klog.KObj(pod))
            metrics_updater.RecordDeferredInPlaceUpdate(vpaSize, vpa.Name, vpa.Namespace, "EvictionPrevented")
        }
        continue
    }
    // rest of the code
}
```

## Test Plan

The following test scenarios will be added to e2e tests. The InPlace mode will be tested in the following scenarios:

- Basic In-Place Update: Pod successfully updated in-place with InPlace mode
- Failed Update - No Eviction: Update fails due to node capacity, verify no eviction occurs and pod remains running
- Failed Update - Retry Success: Update fails initially, conditions improve, verify successful retry

## Upgrade / Downgrade Strategy

### Upgrade

On upgrade to VPA 1.6.0 (tentative release version), users can opt into the new `InPlace` mode by enabling the alpha Feature Gate (which defaults to disabled) by passing `--feature-gates=InPlace=true` to the updater and admission-controller components and setting their VPA UpdateMode to use `InPlace`.
Existing VPAs will continue to work as before.

### Downgrade

On downgrade of VPA from 1.6.0 (tentative release version), nothing will change. VPAs will continue to work as previously, unless, the user had enabled the feature gate. In which case downgrade could break their VPA that uses `InPlace`.

## Feature Enablement and Rollback

### How can this feature be enabled / disabled in a live cluster?

- Feature gate name: `InPlace`
- Components depending on the feature gate:
    - admission-controller
    - updater

Disabling of feature gate `InPlace` will cause the following to happen:
- admission-controller to reject new VPA objects being created with `InPlace` configured
    - A descriptive error message should be returned to the user letting them know that they are using a feature gated feature

Enabling of feature gate `InPlace` will cause the following to happen:
- admission-controller to accept new VPA objects being created with `InPlace` configured
- updater will attempt to perform an in-place **only** adjustment for VPAs configured with `InPlace`

## Kubernetes version compatibility

`InPlace` is being built assuming that it will be running on a Kubernetes version of at least 1.33 with the beta version of [KEP-1287: In-Place Update of Pod Resources](https://github.com/kubernetes/enhancements/issues/1287) enabled.
Should these conditions not be true, the VPA shall not be able to scale your workload at all.

## Implementation History

- 2025-15-11: initial version
