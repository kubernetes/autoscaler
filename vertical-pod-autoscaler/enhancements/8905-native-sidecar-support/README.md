# KEP-8905: Native Sidecar Support

<!-- toc -->
- [Summary](#summary)
   - [Goals](#goals)
   - [Non-Goals](#non-goals)
- [Proposal](#proposal)
- [Design Details](#design-details)
   - [Recommendations](#recommendations)
   - [Update / Admission](#update--admission)
   - [Test Plan](#test-plan)
   - [Upgrade / Downgrade Strategy](#upgrade--downgrade-strategy)
- [Implementation History](#implementation-history)
- [Alternatives](#alternatives)
<!-- /toc -->

## Summary

This proposal adds support for native sidecar containers (init containers with `restartPolicy: Always`) in Vertical Pod Autoscaler.

Kubernetes 1.28 introduced native sidecar containers. These are init containers that start before the main containers and continue running during the lifecycle of the Pod. VPA currently supports standard containers, but it should also support recommending resources for these new native sidecar containers to ensure they are right-sized. Standard init containers will continue to be ignored.
Addresses [issue #7229](https://github.com/kubernetes/autoscaler/issues/7229)

### Goals

- Allow VPA Recommender to generate resource recommendations for native sidecar containers.
- Ensure VPA Updater and Admission Controller can apply recommendations to native sidecar containers.

### Non-Goals

- Support for sidecar containers in Kubernetes versions older than 1.28.
- Support for regular init containers.

## Proposal

The proposal is to extend VPA to support native sidecar containers. This involves updating the Recommender to generate recommendations for these containers, and updating the Updater and Admission Controller to apply these recommendations.

This functionality will be guarded by a new feature gate `NativeSidecar` to allow for safe rollout.

## Design Details

### Recommendations

The Recommender component identifies native sidecar containers by examining init containers with `restartPolicy: Always` in the `SpecClient`. These containers are assigned the `ContainerTypeInitSidecar` type.

When the `NativeSidecar` feature gate is enabled, the `ClusterFeeder` processes native sidecars similarly to standard containers:
- They are stored in a separate `InitSidecarsContainers` map in the pod state
- Resource usage samples are collected and aggregated for recommendations
- Recommendations are generated using the same logic as standard containers
```go
for _, container := range pod.Containers {
  if err := feeder.clusterState.AddOrUpdateContainer(container.ID, container.Request, container.ContainerType); err != nil {
    klog.V(0).InfoS("Failed to add container", "container", container.ID, "error", err)
  }
}
for _, initContainer := range pod.InitContainers {
  if features.Enabled(features.NativeSidecar) && initContainer.ContainerType == model.ContainerTypeInitSidecar {
    if err := feeder.clusterState.AddOrUpdateContainer(initContainer.ID, initContainer.Request, initContainer.ContainerType); err != nil {
      klog.V(0).InfoS("Failed to add initContainer", "container", initContainer.ID, "error", err)
    }
  } else {
    // existing init container logic...
  }
}
```

The VPA custom resource definition remains unchanged. Native sidecar recommendations are included in the `containerRecommendations` array alongside standard container recommendations, using the unique container name to identify them.

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

On upgrade of the VPA to 1.6.0 (tentative release version), nothing will change,
VPAs will continue to work as before.

Users can use the new `NativeSidecar` by enabling the alpha Feature Gate (which defaults to disabled)
by passing `--feature-gates=NativeSidecar=true` to the VPA components.

#### Downgrade

On downgrade of VPA from 1.6.0 (tentative release version), nothing will change.
VPAs will continue to work as previously. Checkpoints may contain sidecar resource information until updated, but updater and admission will not modify sidecar resources.

## Implementation History

- 2025-12-08: Initial proposal

## Alternatives

### Treat as Standard Containers

We could treat them as standard containers, but they are technically init containers in the Pod spec, so the patch path would be incorrect (`/spec/containers` vs `/spec/initContainers`).
