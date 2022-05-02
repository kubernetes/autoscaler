# KEP-4831: VPA 'upscaling only' UpdateMode

<!-- toc -->
- [Summary](#summary)
- [Motivation](#motivation)
   - [Goals](#goals)
   - [Non-Goals](#non-goals)
- [Proposal](#proposal)
- [Design Details](#design-details)
   - [Test Plan](#test-plan)
- [Alternatives](#alternatives)
<!-- /toc -->

## Summary
VPA provides different `UpdateMode`s, describing if/how the VPA updater applies new recommendations to the Pods. They range from "recommendation only, don't make any changes" (`Off`), over "apply new recommendations when pods are re-scheduled, but don't evict" (`Initial`) to "evict a Pod to apply a new recommendation" (`Auto`). These existing UpdateModes work in the same way for scaling up and scaling down. We propose adding a new `UpdateMode` that disables active downscaling while still allowing for scaling up.

## Motivation
For some workloads, each scaling operation introduces disruptions for users. Examples include anything running as a singleton (e.g. etcd) or workload that needs a lot of time to process data on startup after each eviction (e.g. prometheus). Therefore, operators may want to limit the amount of scaling events without causing disruptions due to insufficient resources. This introduces a new tradeoff: Overspend on resources to increase uptime. It is possible to turn off active downscaling entirely or have it managed by an external component to allow downscaling only during certain time windows, days, or based on other criteria.


### Goals
* For a specific VPA object, allow to turn off active downscaling entirely while still allow for upscaling
* Allow for changes of `UpdateMode` during the lifetime of a VPA object
### Non-goals
* Have VPA take care of dynamically changing the `UpdateMode`: Reasons and already existing mechanism achieving a similar functionality will vary widely, this is not what VPA should be concerned with

## Proposal
Add a new `UpdateMode` called `UpscaleOnly`, which works similar to `Auto` when scaling up, allowing the VPA to evict Pods and increase their resource requests, but prevents the VPA from evicting Pods when the new recommendation is smaller than the current one. Similarly to how the `Initial` mode works, this means that a smaller recommendation can still be applied to a Pod if it is recreated due to other reasons.

Since the change is backward-compatible the suggestion is to extend `v1` version of VPA API, avoiding the hassle of introducing a new API version.

## Design Details
### Test Plan
Add automated E2E tests with `spec.updatePolicy.updateMode: "UpscaleOnly"` verify that the VPA Updater does evict for higher recommendations, but not for lower ones.

## Alternatives
### Run VPA in recommendation only mode
We currently achieve a similar functionality by running VPA in `spec.updatePolicy.updateMode: "Off"` and having a different controller inspect the recommendations and apply them based on certain criteria like scaling direction and time.
With this approach you either end up re-building half of the VPA (updater/webhook), or use a different mechanism to apply the recommendations, such as modifying the requests in the Pod owning Object – which has its own drawbacks.
