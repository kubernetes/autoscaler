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
pods expecting that it will apply the recommendation when the pod is recreated. Today, this process
is very disruptive as any changes in recommendations requires a pod to be recreated.

We can instead reduce the amount of disruption by leveraging the [in-place update feature] which is
currently an [alpha feature since 1.27] and graduating to [beta in 1.33].

This proposal enables only core uses of in place updates in VPA with intention of providing the
foundational pieces. Further advanced uses of in place updates in VPA (like applying different
recommendations during pod initialization or providing more frequent smaller updates) will be
introduced as separate enhancement proposals.

[in-place update feature]: https://github.com/kubernetes/enhancements/tree/master/keps/sig-node/1287-in-place-update-pod-resources
[alpha feature since 1.27]: https://github.com/kubernetes/kubernetes/blob/master/CHANGELOG/CHANGELOG-1.27.md#api-change-3
[beta in 1.33]: https://github.com/orgs/kubernetes/projects/178/views/1

### A Note On Disruptions {#disruptions}

It is important to note that **VPA cannot guarantee NO disruptions**. This is because the
underlying container runtime is responsible for actuating the resize operation and there are no
guarantees provided (see [this thread] for more information). However, in practice if the
underlying container runtime supports it, we expect these disruptions to be minimal and that MOST
of the time the updates will be done in-place.

This proposal therefore focuses on *reducing* disruptions while still harnessing the benefits of
VPA.

[this thread]: https://github.com/kubernetes/autoscaler/issues/7722#issue-2796215055

### Goals

* Allow VPA to actuate with reduced disruption.
* Allow VPA to actuate more frequently.
* Allow VPA to actuate in situations where actuation by eviction is not desirable.

### Non-Goals

* Allow VPA to operate with NO disruptions, see the [note above](#disruptions).
* Improved handling of injected sidecars
  * Separate AEP will improve VPAs handling of injected sidecars.

## Proposal

Add a new supported value of [`UpdateMode`]:

* `InPlaceOrRecreate`

Here we specify `OrRecreate` to make sure the user explicitly knows that the pod may be
restarted.

[`UpdateMode`]: https://github.com/kubernetes/autoscaler/blob/71b489f5aec3899157b37472cdf36a1de223d011/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1/types.go#L124

## Context

[In-place update of pod resources KEP] is available in alpha in 1.27 and graduating to beta in
1.33. The feature allows changing container resources while the container is running. It adds
two key features:

* A [`/resize` subresource] that can be used to mutate the `Pod.Spec.Containers[i].Resources`
  field.
* A [`ResizePolicy`] field to Container. This field allows to the user to specify the behavior when
  modifying a resource value. Currently it has two modes:
  - `PreferNoRestart` (default) which indicates to the container runtime that it should try to resize
    the container without restarting. However, it does not guarantee that a restart will not
    happen.
  - `RestartContainer` which indicates that any mutation to the resource requires a restart (for
    example, this is important for Java apps using the `-xmxN` which are unable to resize memory
    without restarting).

Note that resize operations will NOT change the pod's quality of service (QoS) class.

[In-place update of pod resources KEP]: https://github.com/kubernetes/enhancements/issues/1287
[`/resize` subresource]:https://github.com/kubernetes/enhancements/tree/master/keps/sig-node/1287-in-place-update-pod-resources#api-changes
[`ResizePolicy`]: https://github.com/kubernetes/api/blob/4dccc5e86b957cea946a63c4f052ee7dec3946ce/core/v1/types.go#L2636

## Design Details

Today, only the VPA admission controller is responsible for changing the pod spec.

The VPA updater is responsible for evicting pods so that the admission controller can change them
during admission.

In the newly added `InPlaceOrRecreate` mode, the VPA Updater will attempt to execute in-place
updates FIRST. In certain situations, if it is unable to process an in-place update in time, it
will evict the pod to force a change.

We classify three types of updates in the context of this new mode:

1. Updates on pod admission
2. Disruption-free updates
3. Disruptive updates

**NOTE:** Here "disruptive" is from the perspective of VPA. As mentioned above, the container
runtime MAY choose to restart a container or pod as needed.

More on these two types of changes below.

### 1. Applying Updates During Pod Admission

For VPAs using the new `InPlaceOrRecreate` mode, the VPA Admission Controller will apply updates to
starting pods just as it does for VPAs in `Initial`, `Auto`, and `Recreate` modes.

### 2. Applying Disruption-free Updates (**NEW**)

In the `InPlaceOrRecreate` mode when an update only changes resources for which a container
indicates that it doesn't require a restart, then VPA can attempt to actuate the change without
disrupting the pod. VPA Updater will:
* attempt to actuate such updates in place,
* attempt to actuate them if difference between recommendation and request is at least 10%
  * even if pod has been running for less than 12h,
* if necessary execute only partial updates (if some changes would require a restart).

If VPA updater fails to update the size of a container, we ignore the failure and try again. We
will not intentionally evict the pod unless the conditions call for a disruptive update (see
below).

### 3. Applying Disruptive Updates

In the `InPlaceOrRecreate` modes VPA updater will attempt to apply updates that require container
restart in place. It will update them under the same conditions that would trigger update with
`Recreate` mode. That is it will apply disruptive updates if:

* Any container has a request below the corresponding `LowerBound` or
* Any container has a request above the corresponding `UpperBound` or
* Difference between sum of pods requests and sum of recommendation `Target`s is more than 10% and the pod has been
  running undisrupted for at least 12h.
  * Successful disruptive update counts as disruption here (and prevents further disruptive updates to the pod for 12h).

For disruptive updates, the VPA updater will evict a pod to actuate a recommendation if it
attempted to apply the recommendation in place and failed.

VPA updater will consider that the update failed if:
* The pod has `.status.resize: Infeasible` or
* The pod has `.status.resize: Deferred` and more than 1 minute elapsed since the update or
* The pod has `.status.resize: InProgress` and more than 1 hour elapsed since the update:
  * There seems to be a bug where containers that say they need to be restarted get stuck in update, hopefully it gets
    fixed and we don't have to worry about this by beta.
* Patch attempt returns an error.

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
- **NEW** Disruption-free update - doesn't change any resources for which the relevant container specifies
  `RestartPolicy: RestartContainer`.

`Auto` / `Recreate` evicts pod if:
 * [`CanEvict`] returns true for the pod, and it meets at least one of the following conditions:
   * Quick OOM,
   * Outside recommended range,
   * Long-lived pod with significant change.
 
`InPlaceOrRecreate` will attempt to apply a disruption-free update in place if it meets at least one
of the following conditions:
* Quick OOM,
* Outside recommended range,
* Significant change.

`InPlaceOrRecreate` when considering a disruption-free update in place ignore some conditions that
influence eviction decission in the `Recreate` mode:
* [`CanEvict`] won't be checked and
* Pods with significant change can be updated even if they are not long-lived.

`InPlaceOrRecreate` will attempt to apply updates that are **not** disruption-free in place under
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

The following test scenarios will be added to e2e tests. The `InPlaceOrRecreate` mode will be
tested in the following scenarios:

* Admission controller applies recommendation to pod controlled by VPA. 
* Disruption-free in-place update applied to all containers of a pod (request in recommendation bounds).
* Partial disruption-free update applied to some containers of a pod, some disruptive changes skipped (request in
  recommendation bounds).
* Disruptive in-place update applied to all containers of a pod (request out ouf recommendation bounds). 
* Disruptive in-place update will fail. Pod should be evicted and the recommendation applied.
* Disruption free in-place updates fail, pod should not be evicted.

### Details still to consider

#### Careful with memory scale down

Downsizing memory may have to be done slowly to prevent OOMs if application starts to allocate rapidly.
Needs more research on how to scale down on memory safely.

## Implementation History

- 2023-05-10: initial version
- 2025-02-07: Updates to align with latest changes to [KEP-1287](https://github.com/kubernetes/enhancements/issues/1287).
