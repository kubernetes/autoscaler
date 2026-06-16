# KEP-8905: Native Sidecar Support

<!-- toc -->
- [Summary](#summary)
  - [Goals](#goals)
  - [Non-Goals](#non-goals)
- [Proposal](#proposal)
- [Design Details](#design-details)
  - [Recommendations](#recommendations)
    - [Startup Ordering](#startup-ordering)
    - [Pod Resource Calculations](#pod-resource-calculations)
  - [Update / Admission](#update--admission)
  - [Test Plan](#test-plan)
  - [Upgrade / Downgrade Strategy](#upgrade--downgrade-strategy)
    - [Upgrade](#upgrade)
    - [Downgrade](#downgrade)
- [Implementation History](#implementation-history)
- [Alternatives](#alternatives)
  - [Treat as Standard Containers](#treat-as-standard-containers)
<!-- /toc -->

## Summary

This proposal adds support for native sidecar containers (init containers with `restartPolicy: Always`) in Vertical Pod Autoscaler.

Native sidecar containers (introduced via the `SidecarContainers` feature gate) are init containers that start before the main containers and continue running during the lifecycle of the Pod. VPA currently supports standard containers, but it should also support recommending resources for these new native sidecar containers to ensure they are right-sized. Standard init containers will continue to be ignored.
Addresses [issue #7229](https://github.com/kubernetes/autoscaler/issues/7229)

### Goals

- Allow VPA Recommender to generate resource recommendations for native sidecar containers.
- Ensure VPA Updater and Admission Controller can apply recommendations to native sidecar containers.

### Non-Goals

- Support for clusters where the `SidecarContainers` feature gate is not available or enabled.
- Support for regular init containers.

## Proposal

The proposal is to extend VPA to support native sidecar containers. This involves updating the Recommender to generate recommendations for these containers, and updating the Updater and Admission Controller to apply these recommendations.

This functionality will be guarded by a new feature gate `NativeSidecar` to allow for safe rollout.

## Design Details

### Recommendations

The Recommender component identifies native sidecar containers by examining init containers with `restartPolicy: Always` in the [`SpecClient`](https://github.com/kubernetes/autoscaler/blob/d9d867a15e96dc50573c59e071f84df5491c03db/vertical-pod-autoscaler/pkg/recommender/input/spec/spec_client.go#L53-L56). These containers are assigned the `ContainerTypeNativeSidecar` type.

When the `NativeSidecar` feature gate is enabled, the `ClusterFeeder` processes native sidecars similarly to standard containers:
- Resource usage samples are collected and aggregated for recommendations
- Recommendations are generated using the same logic as standard containers
```go
for _, container := range pod.Containers {
  if err := feeder.clusterState.AddOrUpdateContainer(container.ID, container.Request, container.ContainerType); err != nil {
    klog.V(4).ErrorS(err, "Failed to add container", "container", container.ID)
  }
}
for _, initContainer := range pod.InitContainers {
  if features.Enabled(features.NativeSidecar) && initContainer.ContainerType == model.ContainerTypeNativeSidecar {
    if err := feeder.clusterState.AddOrUpdateContainer(initContainer.ID, initContainer.Request, initContainer.ContainerType); err != nil {
      klog.V(4).ErrorS(err, "Failed to add initContainer", "container", initContainer.ID)
    }
  } else {
    // existing init container logic...
  }
}
```

The VPA custom resource definition remains unchanged. Native sidecar recommendations are treated exactly like standard container recommendations; the unique container names allow us to identify them.

#### Startup Ordering

Native sidecars are defined in `initContainers` and start before regular containers. This startup ordering does not affect VPA's recommendation logic because VPA collects ongoing resource usage samples over the container's lifetime regardless of when it started. The only difference is the JSON patch path used when applying recommendations (`/spec/initContainers` vs `/spec/containers`), which is handled in the patch generation logic described below.

#### Pod Resource Calculations

With native sidecars, the effective pod resource request is calculated as:
`max(max(init containers), sum(regular containers) + sum(native sidecars))`

VPA generates per-container recommendations independently and does not need to account for this aggregate calculation since the scheduler handles pod-level resource accounting. However, users should be aware that increasing a native sidecar's resource request increases the pod's overall resource footprint additively (unlike regular init containers, which only matter if they exceed the sum of running containers).

### Update / Admission

`GetContainersResourcesForPod` has been updated to retrieve recommendations for native sidecars, Updater and Admission Controller both use this method so the logic is fairly contained.

```go
GetContainersResourcesForPod(pod *core.Pod, vpa *vpa_types.VerticalPodAutoscaler) ([]vpa_api_util.ContainerResources /*containers*/, []vpa_api_util.ContainerResources /*native sidecars*/, vpa_api_util.ContainerToAnnotationsMap, error)
```

The patch generation logic is updated to target `/spec/initContainers` for native sidecar containers, while normal containers will continue to update `/spec/containers`.

### Test Plan

The following test scenarios will be added to e2e tests.

- Admission applies recommendations to native sidecars.
- Updater will update sidecar container resources in-place or evict.
- Admission will patch sidecar container resources.
- When the feature gate `NativeSidecar` is false VPA components will not modify native sidecars.
- Verify that regular init containers (without `restartPolicy: Always`) are not treated as sidecars and do not receive recommendations intended for sidecars.

### Upgrade / Downgrade Strategy

#### Upgrade

Upgrading to VPA 1.8.0 (tentative release version) does not change behavior by default. Existing VPAs continue to work as before.

To use native sidecar support, enable the (alpha) `NativeSidecar` feature gate (disabled by default) by passing `--feature-gates=NativeSidecar=true` to all VPA components.
When enabled, the Recommender will generate recommendations for native sidecars, and the Updater and Admission Controller will apply them.

#### Downgrade

If the `NativeSidecar` feature gate was never enabled, downgrading from VPA 1.8.0 (tentative release version) has no behavioral impact.

If `NativeSidecar` was enabled before the downgrade:

- Existing VPA behavior remains the same as in versions without native sidecar support; the Updater and Admission Controller will not modify native sidecar resources.
- Checkpoints may still contain native sidecar data until they are refreshed.
- Pods that already had native sidecar resources applied will keep running with those last applied values. Manual intervention may be needed to revert native sidecar resources if you want to return to pre-feature-gate behavior.

## Implementation History

- 2025-12-08: Initial proposal

## Alternatives

### Treat as Standard Containers

We could treat them as standard containers, but they are technically init containers in the Pod spec, so the patch path would be incorrect (`/spec/containers` vs `/spec/initContainers`).
