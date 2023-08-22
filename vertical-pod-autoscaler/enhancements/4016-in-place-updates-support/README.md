# AEP-4016: Support for in place updates in VPA

<!-- toc -->
- [Summary](#summary)
    - [Goals](#goals)
    - [Non-Goals](#non-goals)
- [Proposal](#proposal)
- [Context](#context)
- [Design Details](#design-details)
    - [Applying Updates During Pod Admission](#applying-updates-during-pod-admission)
    - [Applying Disruption-free Updates](#Applying-disruption-free-updates)
    - [Applying Disruptive Updates](#applying-disruptive-updates)
    - [Comparison of `UpdateMode`s](#comparison-of-updatemodes)
    - [Test Plan](#test-plan)
- [Implementation History](#implementation-history)
<!-- /toc -->

## Summary


VPA applies its recommendations with a mutating webhook, during pod creation. It can also evict
pods expecting that it will apply the recommendation when the pod is recreated. This is a
disruptive process so VPA has some mechanism to avoid too frequent disruptive updates. 

This proposal allows VPA to apply its recommendations more frequently, with less disruption by
using the
[in-place update feature] which is an alpha feature [available in Kubernetes 1.27.] This proposal enables only core uses
of in place updates in VPA with intention of gathering more feedback. Any more advanced uses of in place updates in VPA
(like applying different recommendations during pod initialization) will be introduced as separate enhancement
proposals.

[in-place update feature]: https://github.com/kubernetes/enhancements/tree/master/keps/sig-node/1287-in-place-update-pod-resources
[available in Kubernetes 1.27.]: https://github.com/kubernetes/kubernetes/blob/master/CHANGELOG/CHANGELOG-1.27.md#api-change-3

### Goals

* Allow VPA to actuate without disruption,
* Allow VPA to actuate more frequently, when it can actuate without disruption,
* Allow VPA to actuate in situations where actuation by eviction doesn't work.

### Non-Goals

* Improved handling of injected sidecars
  * Separate AEP will improve VPAs handling of injected sidecars.

## Proposal

Add new supported values of [`UpdateMode`]:

* `InPlaceOnly` and
* `InPlaceOrRecreate`.

[`UpdateMode`]: https://github.com/kubernetes/autoscaler/blob/71b489f5aec3899157b37472cdf36a1de223d011/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1/types.go#L124

## Context

[In-place update of pod resources KEP] is available in alpha in [Kubernetes 1.27]. The feature allows changing container
resources while the container is running. It also adds [`ResizePolicy`] field to Container. This field indicates for an
individual resource if a container needs to be restarted by kubelet when the resource is changed. For example it may be
the case that a Container automatically adapts to a change in CPU, but needs to be restarted for a change in Memory to
take effect.

[In-place update of pod resources KEP]: https://github.com/kubernetes/enhancements/issues/1287
[Kubernetes 1.27]: https://github.com/kubernetes/kubernetes/blob/master/CHANGELOG/CHANGELOG-1.27.md#api-change-3
[`ResizePolicy`]: https://github.com/kubernetes/api/blob/8360d82aecbc72aa039281a394ebed2eaf0c0ccc/core/v1/types.go#L2448

## Design Details

In the update modes existing before (`Initial` and `Recreate`) only admission controller was changing pod spec (updater
was responsible for evicting pods so that admission controller could change them during admission).

In the newly added `InPlaceOnly` and `InPlaceOrRecreate` modes VPA Updater will execute in-place updates (which change
pod spec without involving admission controller).

### Applying Updates During Pod Admission

For VPAs in `InPlaceOnly` and `InPlaceOrRecreate` modes VPA Admission Controller will apply updates to starting pods,
like it does for VPAs in `Initial`, `Auto`, and `Recreate` modes.

### Applying Disruption-free Updates

When an update only changes resources for which a container indicates that it doesn't require a restart, then VPA can
attempt to actuate the change without disrupting the pod. VPA Updater will:
* attempt to actuate such updates in place,
* attempt to actuate them if difference between recommendation and request is at least 10%
  * even if pod has been running for less than 12h,
* if necessary execute only partial updates (if some changes would require a restart).

### Applying Disruptive Updates

In both `InPlaceOnly` and `InPlaceOrRecreate` modes VPA updater will attempt to apply updates that require container
restart in place. It will update them under conditions that would trigger update with `Recreate` mode. That is it will
apply disruptive updates if:

* Any container has a request below the corresponding `LowerBound` or
* Any container has a request above the corresponding `UpperBound` or
* Difference between sum of pods requests and sum of recommendation `Target`s is more than 10% and the pod has been
  running undisrupted for at least 12h.
  * Successful disruptive update counts as disruption here (and prevents further disruptive updates to the pod for 12h).

In `InPlaceOrRecreate` mode (but not in `InPlaceOnly` mode) VPA updater will evict pod to actuate a recommendation if it
attempted to apply the recommendation in place and failed.

VPA updater will consider that the update failed if:
* The pod has `.status.resize: Infeasible` or
* The pod has `.status.resize: Deferred` and more than 1 minute elapsed since the update or
* The pod has `.status.resize: InProgress` and more than 1 hour elapsed since the update:
  * There seems to be a bug where containers that say they need to be restarted get stuck in update, hopefully it gets
    fixed and we don't have to worry about this by beta.
* Patch attempt will return an error.
  * If the attempt fails because it would change pods QoS:
    * `InPlaceOrRecreate` will treat it as any other failure and consider evicting the pod.
    * `InPlaceOnly` will consider applying request slightly lower than the limit.

Those failure modes shouldn't disrupt pod operations, only update. If there are problems that can disrupt pod operation
we should consider not implementing the `InPlaceOnly` mode.

### Comparison of `UpdateMode`s

VPA updater considers the following conditions when deciding if it should apply an update:
- [`CanEvict`]:
  - Pod is `Pending` or
  - There are enough running pods in the controller.
- Quick OOM:
  - Any container in the pod had a quick OOM ([by default less than 10 minutes] after the container started) and
  - There is difference between sum of recommendations and sum of current requests over containers (see
    [`defaultPriorityProcessor.GetUpdatePriority`]).
- Long-lived pod - started enough time ago ([by default 12h])
- Significant change - difference between sum of requests over containers and sum of recommendations are different
  enough ([by default 10%]).
- [Outside recommended range]:
  - At least one container has at least one resource request lower than the lower bound of the corresponding
    recommendation or
  - At least one container has at least one resource request higher than the upper bound of the corresponding
    recommendation.
- Disruption-free update - doesn't change any resources for which the relevant container specifies
  `RestartPolicy: RestartContainer`.

`Auto` / `Recreate` evicts pod if:
 * [`CanEvict`] returns true for the pod, and it meets at least one of the following conditions:
   * Quick OOM,
   * Outside recommended range,
   * Long-lived pod with significant change.
 
`InPlaceOnly` and `InPlaceOrRecreate` will attempt to apply a disruption-free update in place if it meets at least one
of the following conditions:
* Quick OOM,
* Outside recommended range,
* Significant change.

`InPlaceOnly` and `InPlaceOrRecreate` when considering a disruption-free update in place ignore some conditions that
influence eviction decission in the `Recreate` mode:
* [`CanEvict`] won't be checked and
* Pods with significant change can be updated even if they are not long-lived.

`InPlaceOnly` and `InPlaceOrRecreate` will attempt to apply updates that are **not** disruption-free in place under
the same conditions that apply to updates in the `Recreate` mode.

`InPlaceOrRecreate` will attempt to apply updates that by eviction when:
* VPA already attempted to apply the update in-place and failed and
* it meets conditions for applying in the `Recreate` mode.

[`CanEvict`]: https://github.com/kubernetes/autoscaler/blob/114a35961a85efdf3f36859350764e5e2c0c7013/vertical-pod-autoscaler/pkg/updater/eviction/pods_eviction_restriction.go#LL100C10-L100C37
[by default less than 10 minutes]: https://github.com/kubernetes/autoscaler/blob/114a35961a85efdf3f36859350764e5e2c0c7013/vertical-pod-autoscaler/pkg/updater/priority/update_priority_calculator.go#L37
[`UpdatePriorityCalculator.AddPod`]: https://github.com/kubernetes/autoscaler/blob/114a35961a85efdf3f36859350764e5e2c0c7013/vertical-pod-autoscaler/pkg/updater/priority/update_priority_calculator.go#L81
[by default 12h]: https://github.com/kubernetes/autoscaler/blob/114a35961a85efdf3f36859350764e5e2c0c7013/vertical-pod-autoscaler/pkg/updater/priority/update_priority_calculator.go#L35
[by default 10%]: https://github.com/kubernetes/autoscaler/blob/114a35961a85efdf3f36859350764e5e2c0c7013/vertical-pod-autoscaler/pkg/updater/priority/update_priority_calculator.go#L33
[Outside recommended range]: https://github.com/kubernetes/autoscaler/blob/114a35961a85efdf3f36859350764e5e2c0c7013/vertical-pod-autoscaler/pkg/updater/priority/priority_processor.go#L73

### Test Plan

The following test scenarios will be added to e2e tests. Both `InPlaceOnly` and `InPlaceOrRecreate` modes will be tested
and they should behave the same:

* Admission controller applies recommendation to pod controlled by VPA. 
* Disruption-free in-place update applied to all containers of a pod (request in recommendation bounds).
* Partial disruption-free update applied to some containers of a pod, some disruptive changes skipped (request in
  recommendation bounds).
* Disruptive in-place update applied to all containers of a pod (request out ouf recommendation bounds). 

There will be also scenarios testing differences between `InPlaceOnly` and `InPlaceOrRecreate` modesL
* Disruptive in-place update will fail. In `InPlaceOnly` pod should not be evicted, in the `InPlaceOrRecreate` pod
  should be evicted and the recommendation applied.
* VPA attempts an update that would change Pods QoS (`RequestsOnly` scaling, request initially < limit, recommendation
  equal to limit). In `InPlaceOnly` pod should not be evicted, request slightly lower than the recommendation will be
  applied. In the `InPlaceOrRecreate` pod should be evicted and the recommendation applied.

### Details still to consider

#### Ensure in-place resize request doesn't cause restarts

Currently the container [resize policy](https://kubernetes.io/docs/tasks/configure-pod-container/resize-container-resources/#container-resize-policies)
can be either `NotRequired` or `RestartContainer`. With `NotRequired` in-place update could still end up
restarting the container if in-place update is not possible, depending on kubelet and container
runtime implementation. However in the proposed design it should be VPA's decision whether to fall back
to restarts or not.

Extending or changing the existing API for in-place updates is possible, e.g. adding a new
`MustNotRestart` container resize policy.

#### Should `InPlaceOnly` mode be dropped

The use case for `InPlaceOnly` is not understood yet. Unless we have a strong signal it solves real
needs we should not implement it. Also VPA cannot ensure no restart would happen unless
*Ensure in-place resize request doesn't cause restarts* (see above) is solved.

#### Careful with memory scale down

Downsizing memory may have to be done slowly to prevent OOMs if application starts to allocate rapidly.
Needs more research on how to scale down on memory safely.

## Implementation History

- 2023-05-10: initial version
