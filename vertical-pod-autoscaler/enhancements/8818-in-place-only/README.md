# AEP-8818: Eviction-Free In-Place Updates in VPA

<!-- toc -->
- [Summary](#summary)
- [Motivation](#motivation)
- [Goals](#goals)
- [Non-Goals](#non-goals)
- [Proposal](#proposal)
- [Design Details](#design-details)
  - [Infeasible Attempt Tracking](#infeasible-attempt-tracking)
    - [Cleanup of Stale Entries](#cleanup-of-stale-entries)
    - [Limitations of In-Memory Storage](#limitations-of-in-memory-storage)
  - [Version-Agnostic Infeasibility Detection](#version-agnostic-infeasibility-detection)
    - [Detection Path 1: Resize Status (Pre-admission-check versions)](#detection-path-1-resize-status-pre-admission-check-versions)
    - [Detection Path 2: Patch Rejection (Post-admission-check versions)](#detection-path-2-patch-rejection-post-admission-check-versions)
  - [Complete Flow Description](#complete-flow-description)
  - [Behavior when Feature Gate is Disabled](#behavior-when-feature-gate-is-disabled)
- [Risk Mitigation](#risk-mitigation)
  - [Memory Limit Downsize Risk](#memory-limit-downsize-risk)
  - [Mitigation Strategies](#mitigation-strategies)
- [Test Plan](#test-plan)
- [Upgrade / Downgrade Strategy](#upgrade--downgrade-strategy)
  - [Upgrade](#upgrade)
  - [Downgrade](#downgrade)
- [Graduation Criteria](#graduation-criteria)
  - [Alpha](#alpha)
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

- Provide a VPA update mode that never evicts pods and only attempts in-place updates
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
- Continue monitoring for recommendation changes and re-attempt updates when new recommendations differ from the current `spec.resources` (see [Resize Status Handling](#complete-flow-description) for details)

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

### Infeasible Attempt Tracking

VPA must track `infeasible` resize attempts to prevent infinite retry loops. This is necessary because infeasibility can be detected at different points depending on the Kubernetes version:
| Kubernetes Version                  | When Infeasibility Is Detected                     | `spec.resources` After Attempt | How VPA Learns            |
|-------------------------------------|----------------------------------------------------|--------------------------------|---------------------------|
|  < 1.36 (or later if KEP slips)     | After patch succeeds, kubelet reports status       | Updated to attempted value     | Resize status = Infeasible|
|  >= 1.36 (targeted, not guaranteed) | At patch time, API server rejects                  | Unchanged (old value)          | Patch error response      |

> **Note:** The Kubernetes sig-node team is *targeting* admission-time feasibility checks for 1.36 ([kubernetes/kubernetes#136043](https://github.com/kubernetes/kubernetes/pull/136043)), but this timeline is not guaranteed and may slip to a later release. VPA implements version-agnostic detection that works correctly regardless of which version introduces this change.

To handle both cases uniformly, VPA maintains a map of infeasible attempts:

```golang
type updater struct {
    // ... existing fields ...

    // infeasibleAttempts maps pod UID to the last resource values
    // that were determined to be infeasible. This prevents retrying the same
    // infeasible values repeatedly.
    infeasibleAttempts           map[types.UID]*vpa_types.RecommendedPodResources
    infeasibleMu                 sync.RWMutex
}
```

Using the pod UID as the key ensures that entries are uniquely identified even if pods with the same name are recreated.

#### Cleanup of Stale Entries

At the beginning of each updater cycle, VPA cleans up entries from the `infeasibleAttempts` map for pods that no longer exist. This prevents memory leaks from accumulating stale entries:

```golang
// CleanupInfeasibleAttempts removes entries from infeasibleAttempts for pods that no longer exist.
// This should be called at the beginning of each updater cycle with the list of all live pods.
func (u *updater) CleanupInfeasibleAttempts(livePods []*apiv1.Pod) {
    u.infeasibleMu.Lock()
    defer u.infeasibleMu.Unlock()

    // Build a set of existing pod UIDs
    seenPods := sets.New[types.UID]()
    for _, pod := range livePods {
        seenPods.Insert(pod.UID)
    }

    // Remove entries for pods that no longer exist
    for podUID := range u.infeasibleAttempts {
        if !seenPods.Has(podUID) {
            delete(u.infeasibleAttempts, podUID)
            klog.V(4).InfoS("Cleaned up infeasible attempt for non-existent pod", "podUID", podUID)
        }
    }
}
```


#### Limitations of In-Memory Storage

The `infeasibleAttempts` map is stored in-memory within the updater component. This has the following implications:

- When running the updater with multiple replicas in HA mode, or when the updater pod is restarted/replaced, the in-memory `infeasibleAttempts` data is lost. This means:
  1. The new updater instance will not know which resize attempts were previously determined to be infeasible
  2. The updater may retry a previously-infeasible resize attempt once before re-learning it is infeasible
  3. This results in at most one wasted update cycle per pod after an updater restart


### Version-Agnostic Infeasibility Detection

Since VPA cannot know at runtime which Kubernetes version is running (and the admission-time check may ship in 1.36 or later), VPA implements dual detection paths that handle both scenarios simultaneously:

#### Detection Path 1: Resize Status (Pre-admission-check versions)

VPA attempts patch → Patch succeeds → `spec.resources` updated → Kubelet evaluates → Status = Infeasible → VPA stores `spec.resources`

In this path:
- The patch succeeds and `spec.resources` is updated to the attempted values
- Kubelet later determines the resize is infeasible and sets the resize status
- VPA detects the `Infeasible` status during reconciliation
- VPA stores `spec.resources` (which contains the infeasible values)


#### Detection Path 2: Patch Rejection (Post-admission-check versions)

VPA attempts patch → API server rejects with infeasibility error → VPA stores attempted recommendation

In this path:
- The API server validates feasibility at admission time via the podresize admission plugin
- If infeasible due to node capacity constraints, the patch is rejected with an error
- The error includes a `StatusCause` with `Type: NodeCapacity` to indicate the specific reason
- `spec.resources` remains unchanged (the patch was rejected)
- VPA detects the error, checks for the `NodeCapacity` cause, and stores the recommendation it attempted to apply

These paths are mutually exclusive for any given update attempt:
- If the patch succeeds, you're on Path 1 (status-based detection)
- If the patch fails with infeasibility, you're on Path 2 (error-based detection)

VPA implements both paths unconditionally, so it works correctly on any Kubernetes version without needing version detection.


### Complete Flow Description

The reconciliation loop follows this sequence of steps for each pod:

VPA retrieves the current recommendation for the pod from the VPA object. Then the VPA checks if there is a stored infeasible attempt for this pod in the `infeasibleAttempts` map.

- If stored attempt exists AND matches current recommendation: Skip this pod entirely. The same values were already determined to be infeasible, so retrying would fail again. Wait for the next reconciliation loop in case the recommendation changes.
- If stored attempt exists AND current recommendation has at least one resource value lower than the stored infeasible attempt: The recommendation may now be feasible since it requires fewer resources in at least one dimension. Clear the stored attempt and proceed with the update.
- If stored attempt exists AND current recommendation has all resource values greater than or equal to the stored infeasible attempt: Skip this pod. The new recommendation requires at least as many resources as the previously infeasible attempt, so it would also be infeasible. The stored attempt is retained.
- If no stored attempt exists: Proceed to next step.

If the pod is currently undergoing an in-place resize (i.e., `spec.resources` differs from `status.resources`), check the resize status:

| Status       | Action                                             | Rationale                                                                                                                   |
| ------------ | -------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------- |
| `InProgress` | Wait, take no action                               | Resize is actively being applied by kubelet                                                                                 |
| `Deferred`   | Wait, take no action                               | Kubelet has accepted resize but is waiting for the right conditions                                                         |
| `Infeasible` | Store `spec.resources` as infeasible, skip pod     | Pre-admission-check path: kubelet determined resize cannot be accommodated. `spec.resources` contains the attempted values. |
| `Error`      | Take no action                                     | Kubelet will automatically retry                                                                                            |


VPA attempts to patch the pod's `spec.resources` with the current recommendation:

| Outcome                                                  | Action                                                     | Rationale                                                                             |
| -------------------------------------------------------- | ---------------------------------------------------------- | ------------------------------------------------------------------------------------- |
| Patch succeeds                                           | Clear any stored infeasible attempt                        | Update is now in progress; previous infeasibility no longer relevant                  |
| Patch fails with **error containing NodeCapacity cause** | Store attempted recommendation as infeasible, skip pod     | Post-admission-check path: API server rejected because resources exceed node capacity |
| Patch fails with **transient error**     | Do nothing                                                 | Will retry in next reconciliation loop                                                |

Periodically, VPA removes entries from the infeasibleAttempts map for pods that no longer exist. This prevents memory leaks from accumulating stale entries. This cleanup behavior is targeted for beta.

**Key Difference from `InPlaceOrRecreate`**: In `InPlace` mode, `Deferred`, `Infeasible`, and `InProgress` statuses all result in waiting—VPA never falls back to eviction. In contrast, `InPlaceOrRecreate` mode may fall back to eviction after a timeout. This ensures that `InPlace` mode pods are never evicted, regardless of how long they remain in a non-updatable state.

Modify the `CanInPlaceUpdate` to accommodate the new update mode:

```golang
// CanInPlaceUpdate checks if pod can be safely updated
func (ip *PodsInPlaceRestrictionImpl) CanInPlaceUpdate(pod *apiv1.Pod, updateMode vpa_types.UpdateMode) utils.InPlaceDecision {
	// Feature gate checks based on update mode
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
	if !present {
		klog.V(4).InfoS("Can't in-place update pod, but not falling back to eviction. Waiting for next loop", "pod", klog.KObj(pod))
		return utils.InPlaceDeferred
	}

	if pod.Status.Phase == apiv1.PodPending {
		return utils.InPlaceDeferred
	}

	singleGroupStats, present := ip.creatorToSingleGroupStatsMap[cr]
	if !present {
		klog.V(4).InfoS("Can't in-place update pod, but not falling back to eviction. Waiting for next loop", "pod", klog.KObj(pod))
		return utils.InPlaceDeferred
	}

	if isInPlaceUpdating(pod) {
		// For InPlace mode: wait for all non-terminal statuses, never evict.
		// Infeasible attempts are tracked and only retried when recommendation changes.
		if updateMode == vpa_types.UpdateModeInPlace {
			resizeStatus := getResizeStatus(pod)
			switch resizeStatus {
			case utils.ResizeStatusInfeasible:
				// Infeasible means node can't accommodate the resize.
				// Store spec.resources and wait for recommendation to change before retrying.
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

		// For InPlaceOrRecreate/Auto modes, check timeout and potentially evict
		canEvict := CanEvictInPlacingPod(pod, singleGroupStats, ip.lastInPlaceAttemptTimeMap, ip.clock)
		if canEvict {
			return utils.InPlaceEvict
		}
		return utils.InPlaceDeferred
	}

	if ip.inPlaceSkipDisruptionBudget && utils.IsNonDisruptiveResize(pod) {
		klog.V(4).InfoS("in-place-skip-disruption-budget enabled, skipping disruption budget check for in-place update")
		return utils.InPlaceApproved
	}

	if ip.inPlaceSkipDisruptionBudget {
		klog.V(4).InfoS("in-place-skip-disruption-budget enabled, but pod has RestartContainer resize policy", "pod", klog.KObj(pod))
	}

	if singleGroupStats.isPodDisruptable() {
		return utils.InPlaceApproved
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

Infeasible Attempt Tracking Tests:
- Infeasible Status Storage (pre-1.36): When resize status is `Infeasible`, verify `spec.resources` is stored and pod is skipped until recommendation changes
- Infeasible Patch Rejection with NodeCapacity Cause (1.36+): When patch is rejected with an error containing `NodeCapacity` cause, verify attempted recommendation is stored and pod is skipped until recommendation changes
- No Retry for Unchanged Recommendation: After infeasible attempt is stored, verify same recommendation is not retried
- Retry After Recommendation Change: After infeasible attempt is stored, verify changed recommendation triggers a new attempt and clears stored value
- Successful Update Clears Stored Attempt: After successful patch, verify any stored infeasible attempt is cleared
- Stale Entry Cleanup: Verify infeasible attempts for deleted pods are cleaned up (targeted for beta)

## Upgrade / Downgrade Strategy

### Upgrade

On upgrade to VPA 1.6.0 (tentative release version), users can opt into the new `InPlace` mode by enabling the alpha Feature Gate (which defaults to disabled) by passing `--feature-gates=InPlace=true` to the updater and admission-controller components and setting their VPA UpdateMode to use `InPlace`.
Existing VPAs will continue to work as before.

### Downgrade

On downgrade of VPA from 1.6.0 (tentative release version), nothing will change. VPAs will continue to work as previously, unless, the user had enabled the feature gate. In which case downgrade could break their VPA that uses `InPlace`.

## Graduation Criteria

### Alpha

- Feature gate InPlace is disabled by default
- Basic functionality implemented including:
  - InPlace update mode accepted by admission controller
  - Updater attempts in-place updates and never evicts pods
  - Deferred behavior for Deferred, InProgress, and unknown statuses
  - Infeasible attempt tracking for both patch failure (1.36+) and status-based (pre-1.36) scenarios
- Unit tests covering core logic including infeasible attempt tracking
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

`InPlace` is being built assuming that it will be running on a Kubernetes version of at least 1.34 with the beta version of [KEP-1287: In-Place Update of Pod Resources](https://github.com/kubernetes/enhancements/issues/1287) enabled.
Should these conditions not be true, the VPA shall not be able to scale your workload at all.

**Kubernetes 1.36+ Considerations**: Starting with Kubernetes 1.36, infeasible resize requests are rejected at the API server level rather than being accepted and later marked as `Infeasible` by the kubelet. The `InPlace` mode handles both scenarios through unified infeasible attempt tracking:
- **Pre-1.36**: When resize status is `Infeasible`, VPA stores `spec.resources` (which reflects the attempted values)
- **1.36+**: When patch fails with an error containing a `NodeCapacity` cause, VPA stores the attempted recommendation

In both cases, VPA only retries when the recommendation changes from the stored values. This ensures consistent behavior across Kubernetes versions without requiring user configuration.

## Implementation History

- 2025-15-11: initial version
