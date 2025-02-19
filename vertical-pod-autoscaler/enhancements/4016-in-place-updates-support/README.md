# AEP-4016: Support for in place updates in VPA

<!-- toc -->
- [Summary](#summary)
    - [Goals](#goals)
    - [Non-Goals](#non-goals)
- [Proposal](#proposal)
- [Context](#context)
- [Design Details](#design-details)
    - [1. Applying Updates During Pod Admission](#pod-admission)
    - [2. In-Place Updates (**NEW**)](#in-place)
    - [Comparison of `UpdateMode`s](#comparison-of-updatemodes)
    - [Test Plan](#test-plan)
- [Implementation History](#implementation-history)
<!-- /toc -->

## Summary

VPA applies its recommendations with a mutating webhook, during pod creation. It can also evict
pods expecting that it will apply the recommendation when the pod is recreated. Today, this process
is potentially disruptive as any changes in recommendations requires a pod to be recreated.

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
* Allow VPA to actuate in situations where actuation by eviction is not desirable.

### Non-Goals

* Allow VPA to actuate more frequently.
* Allow VPA to operate with NO disruptions, see the [note above](#disruptions).
* Improved handling of injected sidecars
  * Separate AEP will improve VPAs handling of injected sidecars.

## Proposal

Add a new supported value of [`UpdateMode`]:

* `InPlaceOrRecreate`

Here we specify `InPlaceOrRecreate` to make sure the user explicitly knows that the existing pod
may be replaced.

[`UpdateMode`]: https://github.com/kubernetes/autoscaler/blob/71b489f5aec3899157b37472cdf36a1de223d011/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1/types.go#L124

For the initial release of in-place updates with VPA, in-place updates will only be available
using the `InPlaceOrRecreate` mode. In the future, once the SIG feels that the feature is
mature enough, this behavior will become the default behavior for the `Auto` mode. See the [`Auto`
mode documentation].

[`Auto` mode documentation]: https://github.com/kubernetes/autoscaler/blob/78c8173b979316f892327022d53369760b000210/vertical-pod-autoscaler/docs/api.md#updatemode

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

Note that in the initial Beta version of in-place updates, [memory limit downscaling is forbidden]
for pods with `resizePolicy: PreferNoRestart`. This means that when VPA will attempt to apply the
patch, it will fail and VPA will need to fallback to a regular eviction (see below).

[In-place update of pod resources KEP]: https://github.com/kubernetes/enhancements/issues/1287
[`/resize` subresource]:https://github.com/kubernetes/enhancements/tree/master/keps/sig-node/1287-in-place-update-pod-resources#api-changes
[`ResizePolicy`]: https://github.com/kubernetes/api/blob/4dccc5e86b957cea946a63c4f052ee7dec3946ce/core/v1/types.go#L2636
[memory limit downscaling is forbidden]: https://github.com/kubernetes/enhancements/pull/5089

## Design Details

Prior to this AEP, only the VPA admission controller was responsible for changing the pod spec.

The VPA updater is responsible for evicting pods so that the admission controller can change them
during admission.

In the newly added `InPlaceOrRecreate` mode, the VPA Updater will attempt to execute in-place
updates _FIRST_. If it is unable to process an in-place update in time, it will evict the pod to
force a change.

This will effectively match the current behavior in `Auto` except that resizes will first be
attempted in-place.

In the future, this logic may be improved to:
* Provide more frequent resizes.
* Make changes that are only attempted using in-place resizes and wouldn't ultimately result in an
  eviction on failure.
* In the case of failure, make smaller updates to circumvent a node that does not have enough
  headroom to accept the full resize but could accomodate a smaller one.

We classify two types of updates in the context of this new mode:

1. Updates on pod admission
2. In-place updates

### 1. Applying Updates During Pod Admission {#pod-admission}

For VPAs using the new `InPlaceOrRecreate` mode, the VPA Admission Controller will apply updates to
starting pods just as it does for VPAs in `Initial`, `Auto`, and `Recreate` modes.

### 2. In-Place Updates (**NEW**) {#in-place}

In the `InPlaceOrRecreate` modes, and for updates that require a container restart, the VPA updater
will attempt to apply updates in place. It will update them under the same conditions that would
trigger an update with `Recreate` mode. That is it will apply an in-place update if:

* Any container has a request below the corresponding `LowerBound` or
* Any container has a request above the corresponding `UpperBound` or
* Difference between sum of pods requests and sum of recommendation `Target`s is more than 10% and the pod has been
  running undisrupted for at least 12h.
  * NOTE: A successful update counts as disruption here (and prevents further disruptive updates to the pod for 12h).

(NEW!) In addition, VPA will attempt an in-place update in some cases where we NORMALLY would not
be able to perform an eviction, including:

* If `CanEvict` is false.
* If any of the `EvictionRequirements` on the VPA are not true.

These additional resizes can be attempted because the eviction fallback would fail anyway.

The VPA updater will evict a pod to actuate a recommendation if it attempted to apply the
recommendation in place and failed.

VPA updater will consider that the update failed if:
* The pod has condition `PodResizePending` with reason `Infeasible` or
* The pod has condition `PodResizePending` with reason `Deferred` and more than 5 minutes elapsed
  since the update or
* The pod has condition `PodResizing` and more than 1 hour elapsed since the update or
* Patch attempt returns an error.

Note that in the initial version of In-Place updates, memory limit downscaling will always fail
the patch operation. This means VPA will need to evict the pod normally for this change to happen.

#### A note on `ResizePolicy`.

VPA does not care and should not care about a container's `ResizePolicy` setting. In the new mode,
it will simply issue the `/resize` request and let the underlying machinery apply the resize
operation in a way that complies with the user's specification.

### Comparison of `UpdateMode`s

Today, VPA updater considers the following conditions when deciding if it should apply an update:
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
   * `EvictionRequirements` are all true.
 
`InPlaceOrRecreate` will attempt to apply an update in place if it meets at least one
of the following conditions:
* Quick OOM,
* Outside recommended range,
* Long-lived Significant change.
* [`CanEvict`] won't be checked and
* [`EvictionRequirements`] won't be checked


[`CanEvict`]: https://github.com/kubernetes/autoscaler/blob/114a35961a85efdf3f36859350764e5e2c0c7013/vertical-pod-autoscaler/pkg/updater/eviction/pods_eviction_restriction.go#LL100C10-L100C37
[by default less than 10 minutes]: https://github.com/kubernetes/autoscaler/blob/114a35961a85efdf3f36859350764e5e2c0c7013/vertical-pod-autoscaler/pkg/updater/priority/update_priority_calculator.go#L37
[`UpdatePriorityCalculator.AddPod`]: https://github.com/kubernetes/autoscaler/blob/114a35961a85efdf3f36859350764e5e2c0c7013/vertical-pod-autoscaler/pkg/updater/priority/update_priority_calculator.go#L81
[by default 12h]: https://github.com/kubernetes/autoscaler/blob/114a35961a85efdf3f36859350764e5e2c0c7013/vertical-pod-autoscaler/pkg/updater/priority/update_priority_calculator.go#L35
[by default 10%]: https://github.com/kubernetes/autoscaler/blob/114a35961a85efdf3f36859350764e5e2c0c7013/vertical-pod-autoscaler/pkg/updater/priority/update_priority_calculator.go#L33
[Outside recommended range]: https://github.com/kubernetes/autoscaler/blob/114a35961a85efdf3f36859350764e5e2c0c7013/vertical-pod-autoscaler/pkg/updater/priority/priority_processor.go#L73
[`EvictionRequirements`]: https://github.com/kubernetes/autoscaler/blob/54fe60ed4d4bb4cb89fe4abe11284d1bd6b06390/vertical-pod-autoscaler/pkg/updater/priority/scaling_direction_pod_eviction_admission.go

### Test Plan

The following test scenarios will be added to e2e tests. The `InPlaceOrRecreate` mode will be
tested in the following scenarios:

* Admission controller applies recommendation to pod controlled by VPA. 
* In-place update applied to all containers of a pod.
* Partial updates applied to some containers of a pod, some changes skipped (request in
  recommendation bounds).
* In-place update will fail. Pod should be evicted and the recommendation applied.
* In-place update will fail but `CanEvict` is false, pod should not be evicted.
* In-place update will fail but `EvictionRequirements` are false, pod should not be evicted.

### Details still to consider

#### Careful with memory scale down

Downsizing memory may have to be done slowly to prevent OOMs if application starts to allocate rapidly.
Needs more research on how to scale down on memory safely.

## Implementation History

- 2023-05-10: initial version
- 2025-02-19: Updates to align with latest changes to [KEP-1287](https://github.com/kubernetes/enhancements/issues/1287).
