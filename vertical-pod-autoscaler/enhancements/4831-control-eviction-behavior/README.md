# KEP-4831: Control VPA eviction behavior based on scaling direction and resource

<!-- toc -->
- [Summary](#summary)
- [Motivation](#motivation)
   - [Goals](#goals)
- [Proposal](#proposal)
- [Design Details](#design-details)
   - [Test Plan](#test-plan)
- [Alternatives](#alternatives)
<!-- /toc -->

## Summary
VPA provides different `UpdateMode`s, describing if/how the VPA applies new recommendations to the Pods. They range from "recommendation only, don't make any changes" (`Off`), over "apply new recommendations when pods are re-scheduled, but don't evict" (`Initial`) to "evict a Pod to apply a new recommendation" (`Auto`). These existing `UpdateModes` work in the same way for scaling up and scaling down and are the same for all resources controlled by the VPA. We propose adding a new functionality which allows to control the conditions in which Pods can be evicted to apply a new recommendation based on the changes (are we scaling up or scaling down?) for each resource individually.

## Motivation
For some workloads, each eviction introduces disruptions for users. Examples include anything running as a singleton (e.g. etcd) or workload that needs a lot of time to process data on startup (e.g. prometheus). Therefore, operators may want to limit the amount of evictions. We can do this by introducing a new tradeoff: Overspend on resources to increase uptime. It is possible to turn off eviction for downscaling entirely or have it managed by an external component to allow downscaling only during certain time windows, days, or based on other criteria.

### Goals
* For a specific VPA object, allow to turn off eviction for downscaling entirely while still allow for upscaling
* Allow for resource-specific decisions: The desired policy may be different for CPU and Memory

## Proposal
Add a new field `EvictionRequirements` to [`PodUpdatePolicy`](https://github.com/kubernetes/autoscaler/blob/2f4385b72e304216cf745893747da45ef314898f/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1/types.go#L109) of type `[]*EvictionRequirement`. A single `EvictionRequirement` defines a condition which must be `true` to allow eviction for the corresponding `Pod`. When multiple `EvictionRequirements` are specified for a `Pod`, all of them must evaluate to `true` to allow eviction.

A single `EvictionRequirement` specifies `Resources` and a `ChangeRequirement` comparing the new recommendation (`Target`) with the existing requests on a Pod (`Requests`). Possible values for `Resources` are `[CPU]` and `[Memory]` or both `[CPU,Memory]`. If `Resources: [CPU, Memory]`, the condition must be true for either of the two resources to allow for eviction. Possible values for `ChangeRequirement` are `TargetHigherThanRequests`, `TargetHigherThanOrEqualToRequests`, `TargetLowerThanRequests` and `TargetLowerThanOrEqualToRequests`.

Add validation to prevent users from adding `EvictionRequirements` which can never evaluate to `true`:
* Reject if more than one `EvictionRequirement` for a single resource is found
* Reject if `Resource: [CPU, Memory]` is specified on one `EvictionRequirement` together with `Resource: [CPU]` or `Resource: [Memory]` on another `EvictionRequirement`

## Design Details
### Test Plan
* Add automated E2E tests verifying that VPA only evicts when all `EvictionRequirements` evaluate to `true`
* Keep existing tests which verify that nothing changes when users don't specify any `EvictionRequirements`
* Add tests verifying the validation

## Alternatives
### Run VPA in recommendation only mode
We currently achieve a similar functionality by running VPA in `spec.updatePolicy.updateMode: "Off"` and having a different controller inspect the recommendations and apply them based on certain criteria like scaling direction and time.
With this approach you either end up re-building half of the VPA (updater/webhook), or use a different mechanism to apply the recommendations, such as modifying the requests in the Pod owning Object – which has its own drawbacks.
