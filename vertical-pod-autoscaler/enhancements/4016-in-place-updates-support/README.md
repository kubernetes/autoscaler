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
    - [Test Plan](#test-plan)
- [Implementation History](#implementation-history)
<!-- /toc -->

## Summary


VPA applies its recommendations with a mutating webhook, during pod creation. It can also evict
pods expecting that it will apply the recommendation when the pod is recreated. This is a
disruptive process so VPA has some mechanism to avoid too frequent disruptive updates. 

This proposal allows VPA to apply its recommendations more frequently, with less disruption by
using in-place update feature which is an alpha feature available in Kubernetes 1.27.

This proposal enables only core uses of  in place updates in VPA with intention of gathering more feedback. Any more
advanced uses of in place updates in VPA (like applying different recommendations during pod initialization) will be
introduced as separate enhancement proposals.

### Goals

* Allow VPA to actuate without disruption,
* Allow VPA to actuate more frequently, when it can actuate without disruption,
* Allow VPA to actuate in situations where actuation by eviction doesn't work.

### Non-Goals

* Improved handling of injected sidecars
  * Separate AEP will improve VPAs handling of injected sidecars.

## Proposal

Add new supported values of
[`UpdateMode`](https://github.com/kubernetes/autoscaler/blob/71b489f5aec3899157b37472cdf36a1de223d011/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1/types.go#L124):

* `InPlaceOnly` and
* `InPlaceOrRecreate`.

## Context

[In-place update of pod resources KEP](https://github.com/kubernetes/enhancements/issues/1287) is available in alpha in
[Kubernetes 1.27](https://github.com/kubernetes/kubernetes/blob/master/CHANGELOG/CHANGELOG-1.27.md#api-change-3). The
feature allows changing container resources while the container is running. It also adds 
[`ResizePolicy`](https://github.com/kubernetes/api/blob/8360d82aecbc72aa039281a394ebed2eaf0c0ccc/core/v1/types.go#L2448)
field to Container. This field indicates that the container should be restarted when its resources

## Design Details

VPA Updater will execute in-place updates or eviction-updates in the newly added `InPlaceOnly` and `InPlaceOrRecreate`
modes. 

### Applying Updates During Pod Admission

For VPAs in `InPlaceOnly` and `InPlaceOrRecreate` modes VPA Admission Controller will apply updates to starting pods,
like it does for VPAs in `Initial`, `Auto`, and `Recreate` modes.

### Applying Disruption-free Updates

When update changes only resources for which container indicates that it doesn't require a restart then VPA can attempt
to actuate the change without disrupting the pod. VPA Updater will:
* attempt to actuate such updates in place,
* attempt to actuate them if difference between recommendation and request is at least 10%
  * even if pod has been running for less than 24h,
* if necessary execute only partial updates (if some changes would require a restart).

### Applying Disruptive Updates

In both `InPlaceOnly` and `InPlaceOrRecreate` modes VPA updater will attempt to apply updates that require container
restart in place. It will update them under conditions that would trigger update with `Recreate` mode. That is it will
apply disruptive updates if:

* Sum of pods requests is below sum of recommendations `LowerBound` or
* Sum of pods requests is above sum of recommendations `UppelBound` or
* Difference between sum of pods requests and sum of recommendation `Target`s is more than 10% and the pod has been
  running undisrupted for at least 24h.
  * Successful disruptive update counts as disruption here (and prevents further disruptive updates to the pod for 24h).

In `InPlaceOrRecreate` mode (but not in `InPlaceOnly` mode) VPA updater will evict pod to actuate a recommendation if it
attempted to apply the recommendation in place and failed.

### Test Plan

The following test scenarios will be added to e2e tests. Both `InPlaceOnly` and `InPlaceOrRecreate` modes will be tested
and they should behave the same:

* Admission controller applies recommendation to pod controlled by VPA. 
* Disruption-free in-place update applied to all containers of a pod (request in recommendation bounds).
* Partial disruption-free update applied to some containers of a pod, some disruptive changes skipped (request in
  recommendation bounds).
* Disruptive in-place update applied to all containers of a pod (request out ouf recommendation bounds).

There will be also scenario testing difference between `InPlaceOnly` and `InPlaceOrRecreate` modes. Disruptive in-place
update will fail. In `InPlaceOnly` pod should not be evicted, in the `InPlaceOrRecreate` pod should be evicted and a the
recommendation applied.

## Implementation History

- 2023-05-10: initial version