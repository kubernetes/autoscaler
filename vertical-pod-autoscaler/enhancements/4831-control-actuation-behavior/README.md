# AEP-4831: Control VPA eviction behavior based on scaling direction and resource

<!-- toc -->
- [Summary](#summary)
- [Motivation](#motivation)
   - [Goals](#goals)
- [Proposal](#proposal)
- [Design Details](#design-details)
   - [Test Plan](#test-plan)
- [Alternatives](#alternatives)
- [Implementation History](#implementation-history)

<!-- /toc -->

## Summary
VPA provides different `UpdateMode`s, describing if/how the VPA applies new recommendations to the Pods. They range from
"recommendation only, don't make any changes" (`Off`), over "apply new recommendations when pods are re-scheduled, but
don't evict" (`Initial`) to "evict a Pod to apply a new recommendation" (`Auto`). These existing `UpdateModes` work in
the same way for scaling up and scaling down and are the same for all resources controlled by the VPA.

[Kubernetes 1.27 supports](https://github.com/kubernetes/kubernetes/blob/master/CHANGELOG/CHANGELOG-1.27.md#api-change-3)
[changing container resources in place](https://github.com/kubernetes/enhancements/issues/1287) as an alpha feature.
VPA will use that as
[an alternative way to actuate its recommendations](https://github.com/kubernetes/autoscaler/issues/4016).

We propose adding a new functionality which allows to control the conditions in which recommendations can be applied to
containers based on the changes (are we scaling up or scaling down?) for each resource individually.

## Motivation
For some workloads, each eviction introduces disruptions for users. Examples include anything running as a singleton 
(e.g. etcd) or workload that needs a lot of time to process data on startup (e.g. prometheus). Therefore, operators may
want to limit the amount of evictions. We can do this by introducing a new tradeoff: Overspend on resources to increase 
uptime. It is possible to turn off eviction for downscaling entirely or have it managed by an external component to 
allow downscaling only during certain time windows, days, or based on other criteria.

Scaling containers in place might cause immediate disruption (when containers specify `RestartPolicy: RestartContainer`
for one of the changed resources) which operators might want to avoid. Even if in place scale down doesn't cause
immediate disruption there is no guarantee that container will be able to scale up later without eviction. 

### Goals
* For a specific VPA object, allow turning off actuation for downscaling entirely while still allow for upscaling
* Allow for resource-specific decisions: The desired policy may be different for CPU and Memory

## Proposal
Add a new field `ActuationRequirements` to
[`PodUpdatePolicy`](https://github.com/kubernetes/autoscaler/blob/2f4385b72e304216cf745893747da45ef314898f/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1/types.go#L109)
of type `[]*ActuationRequirement`. A single `ActuationRequirement` defines a condition which must be `true` to allow
actuation for the corresponding `Container`. When multiple `ActuationRequirements` are specified for a `Pod`, all of
them must evaluate to `true` to allow eviction. VPA may partially actuate its recommendation in place, obeying all the
`ActuationRequirements` while not applying parts of the recommendation that would violate any `ActuationRequirement`.

A single `ActuationRequirement` specifies `Resources` and a `ChangeRequirement` comparing the new recommendation
(`Target`) with the existing requests on a Pod (`Requests`). Possible values for `Resources` are `[CPU]` and `[Memory]`
or both `[CPU,Memory]`. If `Resources: [CPU, Memory]`, the condition must be true for either of the two resources to
allow for actuation. Possible values for `ChangeRequirement` are `TargetHigherThanRequests`,
`TargetHigherThanOrEqualToRequests`, `TargetLowerThanRequests` and `TargetLowerThanOrEqualToRequests`.

Add validation to prevent users from adding `ActuationRequirements` which can never evaluate to `true`:
* Reject if more than one `ActuationRequirement` for a single resource is found
* Reject if `Resource: [CPU, Memory]` is specified on one `ActuationRequirement` together with `Resource: [CPU]` or
* `Resource: [Memory]` on another `ActuationRequirement`

## Design Details
### Test Plan
* Add automated E2E tests verifying that VPA only evicts when all `ActuationRequirements` evaluate to `true`
* Keep existing tests which verify that nothing changes when users don't specify any `ActuationRequirements`
* Add tests verifying the validation

## Alternatives
### Run VPA in recommendation only mode
We currently achieve a similar functionality by running VPA in `spec.updatePolicy.updateMode: "Off"` and having a 
different controller inspect the recommendations and apply them based on certain criteria like scaling direction and 
time.
With this approach you either end up re-building half of the VPA (updater/webhook), or use a different mechanism to
apply the recommendations, such as modifying the requests in the Pod owning Object – which has its own drawbacks.

### Independent controls for eviction and in-place actuation
We could introduce independent fields controlling scaling by eviction and scaling in place. This seems redundant.

The following analysis focuses on the `InPlaceOrRecreate` mode. In other modes having separate controls for in-place and
eviction actuation doesn't offer any more control than single setting, because VPA won't actively actuate (`Off` and
`Initial`) or because it will use only one mode of actuation (`Recreate` and `InPlaceOnly`). It also assumes that scale
downs always succeed.

VPA will attempt to scale in-place before attempting to evict. Since scale downs succeed VPA will never consider
evicting  pods to scale them down so configuring scale down in place separately from scale down by eviction is
unnecessary.

If user wants to scale up in place only, not evict they should use `InPlaceOnly` mode (since we assume scale downs in
place always succeed `InPlaceOnly` and `InPlaceOrRecreate` are the same when scaling down).

If user prefers to scale up in place and fall back to scale up by eviction if that fails then that's behavior of
`InPlaceOrRecreate` for scale ups when `ActuationRequirements` allow scale ups.

`ActuationRequirements` placing the same requirements on in-place and eviction actuation doesn't allow:

1. Scaling in different modes, depending on which containers change,
2. Scaling up by eviction only but scaling down only in place.

However, those seem unnecessary because if some containers need to restart when resources change they can specify that.

## Implementation History

* 2022-09-08: Initial version
* 2023-05-11: Adjusted for in-place updates

