# AEP-8818: Non-Disruptive In-Place Updates in VPA

<!-- toc -->
- [Summary](#summary)
- [Motivation](#motivation)
- [Goals](#goals)
- [Non-Goals](#non-goals)
- [Proposal](#proposal)
- [Design Details](#design-details)
	- [Resize Status Handling](#resize-status-handling)
- [Risk Mitigation](#risk-mitigation)
	- [Memory Limit Downsize Risk](#memory-limit-downsize-risk)
	- [Mitigation Strategies](#mitigation-strategies)
- [Test Plan](#test-plan)
- [Graduation Criteria](#graduation-criteria)
	- [Alpha](#alpha)
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
- Change the behavior of existing update modes (Off, Recreate, InPlaceOrRecreate)
- Eliminate all possible disruption scenarios (see [Risk Mitigation](#risk-mitigation) for details on memory limit downsizing risks)


## Proposal

Add a new supported value of UpdateMode: `InPlace`
This mode will:
- Apply recommendations during pod admission (like all other modes)
- Attempt in-place updates for running pods under the same conditions as `InPlaceOrRecreate`
- Never add pods to `podsForEviction` if in-place updates fail
- Continue monitoring for recommendation changes and re-attempt updates when new recommendations differ from the current `spec.resources` (see [Resize Status Handling](#resize-status-handling) for details)

## Design Details

Add `UpdateModeInPlace` to the VPA types:

```golang
// In pkg/apis/autoscaling.k8s.io/v1/types.go
const (
    // ... existing modes ...
    // UpdateModeInPlace means that VPA will only attempt to update pods in-place
    // and will never evict them. If in-place update fails, VPA will rely on
    // Kubelet's automatic retry mechanism.
    UpdateModeInPlace UpdateMode = "InPlace"
)
```

### Resize Status Handling

The `InPlace` mode handles different resize statuses with distinct behaviors. Critically, before checking the resize status, VPA first compares the current recommendation against the pod's `spec.resources`. If they differ, VPA attempts to apply the new recommendation regardless of the current resize status, as the new recommendation may be feasible (e.g., a smaller resource request that fits on the node).

When the current recommendation matches `spec.resources`, VPA handles each resize status as follows:

- **`Deferred`**: VPA takes no action and waits. The kubelet has accepted the resize but is waiting for the right conditions to apply it (e.g., waiting for container restart). VPA continues to monitor for recommendation changes—if a new recommendation arrives that differs from `spec.resources`, VPA will attempt to apply it.

- **`Infeasible`**: VPA takes no action and waits. The node cannot accommodate the current resize request. VPA continues to monitor for recommendation changes—if a new recommendation arrives that differs from `spec.resources` (e.g., a lower resource request), VPA will attempt to apply it, as the new values may be feasible on the node.

- **`InProgress`**: VPA takes no action and waits for completion. The resize is actively being applied by kubelet.

- **Error Handling**: There are two types of errors that can occur:
  1. **API Server Error**: The patch request to update `spec.resources` fails (e.g., due to conflicts, validation errors, or transient API server issues). In this case, VPA will retry the patch operation in the next reconciliation loop.
  2. **Resize Condition Error**: The resize status indicates an error occurred during the kubelet's attempt to apply the resize (reflected in the pod's status conditions). In this case, VPA will retry by re-applying the patch, as the error may be transient.

**Key Difference from `InPlaceOrRecreate`**: In `InPlace` mode, both `Deferred` and `Infeasible` statuses are treated the same way—VPA waits and monitors for recommendation changes without taking any action. In contrast, `InPlaceOrRecreate` mode treats these statuses differently, potentially falling back to eviction after a timeout. This unified handling in `InPlace` mode ensures that pods are never evicted, regardless of how long they remain in a non-updatable state.

"Retry" in this context means VPA will attempt to patch `spec.resources` with the current recommendation. VPA does not artificially adjust recommendations to work around constraints—the Recommender component is solely responsible for generating recommendations. If the recommendation changes (because the Recommender produces a new value), VPA will attempt to apply the new recommendation, which may succeed where the previous one failed (e.g., if the new recommendation requests fewer resources).

Modify the `CanInPlaceUpdate` to accommodate the new update mode:

```golang
// CanInPlaceUpdate checks if pod can be safely updated
func (ip *PodsInPlaceRestrictionImpl) CanInPlaceUpdate(pod *apiv1.Pod, updateMode vpa_types.UpdateMode) utils.InPlaceDecision {
	switch updateMode {
	case vpa_types.UpdateModeInPlaceOrRecreate:
		if !features.Enabled(features.InPlaceOrRecreate) {
			return utils.InPlaceEvict
		}
	case vpa_types.UpdateModeInPlace:
		if !features.Enabled(features.InPlace) {
			return utils.InPlaceDeferred
		}
	case vpa_types.UpdateModeAuto:
		// Auto mode is deprecated but still supports in-place updates
		// when the feature gate is enabled
		if !features.Enabled(features.InPlaceOrRecreate) {
			return utils.InPlaceEvict
		}
	default:
		// UpdateModeOff, UpdateModeInitial, UpdateModeRecreate, etc.
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
				resizeStatus := getResizeStatus(pod)
				// For InPlace mode: wait for Deferred, retry for Infeasible (no backoff for alpha)
				if updateMode == vpa_types.UpdateModeInPlace {
					switch resizeStatus {
					case utils.ResizeStatusInfeasible:
						// Infeasible means node can't accommodate the resize.
						// For alpha, retry with no backoff.
						klog.V(4).InfoS("In-place update infeasible, will retry", "pod", klog.KObj(pod))
						return utils.InPlaceInfeasible
					case utils.ResizeStatusDeferred:
						// Deferred means kubelet is waiting to apply the resize.
						// Do nothing, wait for kubelet to proceed.
						klog.V(4).InfoS("In-place update deferred by kubelet, waiting", "pod", klog.KObj(pod))
						return utils.InPlaceDeferred
					case utils.ResizeStatusInProgress:
						// Resize is actively being applied, wait for completion.
						klog.V(4).InfoS("In-place update in progress, waiting for completion", "pod", klog.KObj(pod))
						return utils.InPlaceDeferred
					case utils.ResizeStatusError:
						// Error during resize, retry
						klog.V(4).InfoS("In-place update error, will retry", "pod", klog.KObj(pod))
						return utils.InPlaceInfeasible
					default:
						klog.V(4).InfoS("In-place update status unknown, waiting", "pod", klog.KObj(pod), "status", resizeStatus)
						return utils.InPlaceDeferred
					}
				}
				// For InPlaceOrRecreate mode, check timeout
				canEvict := CanEvictInPlacingPod(pod, singleGroupStats, ip.lastInPlaceAttemptTimeMap, ip.clock)
				if canEvict {
					return utils.InPlaceEvict
				}
				return utils.InPlaceDeferred
			}
			if singleGroupStats.isPodDisruptable() {
				return utils.InPlaceApproved
			}
		}
	}
	klog.V(4).InfoS("Can't in-place update pod, but not falling back to eviction. Waiting for next loop", "pod", klog.KObj(pod))
	return utils.InPlaceDeferred
}
```

### Behavior when Feature Gate is Disabled

- When `InPlace` feature gate is disabled and a VPA is configured with `UpdateMode: InPlace`, the updater will skip processing that VPA entirely (not fall back to eviction).
- In contrast, `InPlaceOrRecreate` with its feature gate disabled will fall back to eviction mode.

This design ensures that `InPlace` mode truly guarantees no evictions, even in misconfiguration scenarios.

## Risk Mitigation

### Memory Limit Downsize Risk

While `InPlace` mode prevents pod eviction and eliminates the disruption associated with pod recreation, it is still subject to the behavior of Kubernetes' `InPlacePodVerticalScaling` feature.
When a memory limit is decreased in-place, there is a small but non-zero risk of `OOMKill` if the container's current memory usage exceeds the new lower limit at the moment the resize is applied.
This is an inherent limitation of in-place resource updates documented in [KEP-1287](https://github.com/kubernetes/enhancements/blob/master/keps/sig-node/1287-in-place-update-pod-resources/README.md), not a VPA-specific behavior.
This risk may be unacceptable for workloads with strict SLO requirements where even brief disruptions (including `OOMKills`) cannot be tolerated.

### Mitigation Strategies

For workloads where even unintended OOMKills are unacceptable, users should implement one or more of the following strategies:

- Disable memory limits for critical containers - configure your VPA's ResourcePolicy to prevent VPA from managing memory limits entirely.
- Use conservative memory limit recommendations - if memory limits must be managed by VPA, configure generous bounds and safety buffers

## Test Plan

The following test scenarios will be added to e2e tests. The InPlace mode will be tested in the following scenarios:

- Basic In-Place Update: Pod successfully updated in-place with InPlace mode
- No Eviction on Failure: Update fails due to node capacity, verify no eviction occurs and pod remains running
- Feature Gate Disabled: Verify InPlace mode is rejected when feature gate is disabled
- Indefinite Wait for In-Progress Updates: Update is in-progress for extended period, verify no timeout/eviction occurs (unlike `InPlaceOrRecreate`)
- Failed Update - Retry Success: Update fails initially, conditions improve, verify successful retry
- Infeasible Resize Handling: Pod resize marked as infeasible, verify pod is deferred (not evicted)

## Upgrade / Downgrade Strategy

### Upgrade

On upgrade to VPA 1.6.0 (tentative release version), users can opt into the new `InPlace` mode by enabling the alpha Feature Gate (which defaults to disabled) by passing `--feature-gates=InPlace=true` to the updater and admission-controller components and setting their VPA UpdateMode to use `InPlace`.
Existing VPAs will continue to work as before.

### Downgrade

On downgrade of VPA from 1.6.0 (tentative release version), nothing will change. VPAs will continue to work as previously, unless, the user had enabled the feature gate. In which case downgrade could break their VPA that uses `InPlace`.

## Graduation Criteria

### Alpha

- Feature gate InPlace is disabled by default
- Basic functionality implemented including InPlace update mode accepted by admission controller, updater attempts in-place updates and never evicts pods, retry behavior for Infeasible status with no backoff, and deferred behavior for Deferred, InProgress, and unknown statuses
- Unit tests covering core logic
- E2E tests for basic scenarios
- Documentation updated

## Feature Enablement and Rollback

### How can this feature be enabled / disabled in a live cluster?

- Feature gate name: `InPlace`
- Components depending on the feature gate:
    - admission-controller
    - updater

Disabling of feature gate `InPlace` will cause the following to happen:
- admission-controller will:
	- Reject new VPA objects being created with `InPlace` configured
    	- A descriptive error message should be returned to the user letting them know that they are using a feature gated feature
	- Continue to apply recommendations at pod admission time for existing VPAs configured with InPlace mode (behaving similarly to Initial mode)
		- This ensures that when pods are deleted and recreated (e.g., by a deployment rollout or manual deletion), they receive the latest resource recommendations
		- Only the in-place update functionality is disabled; admission-time updates remain functional
- updater will:
	- Skip processing VPAs with `InPlace` mode (no in-place updates or evictions will be attempted)
	- Effectively treating these VPAs as if they were in Initial or Off mode for running pods

Enabling of feature gate `InPlace` will cause the following to happen:
- admission-controller to accept new VPA objects being created with `InPlace` configured
- updater will attempt to perform an in-place **only** adjustment for VPAs configured with `InPlace`

## Kubernetes version compatibility

`InPlace` is being built assuming that it will be running on a Kubernetes version of at least 1.33 with the beta version of [KEP-1287: In-Place Update of Pod Resources](https://github.com/kubernetes/enhancements/issues/1287) enabled.
Should these conditions not be true, the VPA shall not be able to scale your workload at all.

## Implementation History

- 2025-15-11: initial version
