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
- [Alternatives](#alternatives)
<!-- /toc -->

## Summary

This proposal adds support for native sidecar containers (init containers with `restartPolicy: Always`) in Vertical Pod Autoscaler.

Kubernetes 1.28 introduced native sidecar containers. These are init containers that start before the main containers and continue running during the lifecycle of the Pod. VPA currently supports standard containers and regular init containers, but it should also support recommending resources for these new native sidecar containers to ensure they are right-sized.
Addresses [issue #7229](https://github.com/kubernetes/autoscaler/issues/7229)

### Goals

- Allow VPA Recommender to generate resource recommendations for native sidecar containers.
- Ensure VPA Updater and Admission Controller can apply recommendations to native sidecar containers.

### Non-Goals

- Support for sidecar containers in Kubernetes versions older than 1.28.

## Proposal

The proposal is to introduce a new feature gate `NativeSidecar` in VPA. When enabled, VPA components will recognize and handle native sidecar containers.

## Design Details

### Recommendations

The Recommender component identifies native sidecar containers by examining init containers with `restartPolicy: Always` in the `SpecClient`. These containers are assigned the `ContainerTypeInitSidecar` type.

When the `NativeSidecar` feature gate is enabled, the `ClusterFeeder` processes native sidecars similarly to standard containers:
- They are stored in a separate `InitSidecarsContainers` map in the pod state
- Resource usage samples are collected and aggregated for recommendations
- Recommendations are generated using the same logic as standard containers

The VPA custom resource definition remains unchanged. Native sidecar recommendations are included in the `containerRecommendations` array alongside standard container recommendations, using the unique container name to identify them.

### Update / Admission

Both Updater and Admission Controller components retrieve recommendations for init containers (including native sidecars) separately from standard containers using `GetContainersResourcesForPod`.
The patch generation logic is updated to target `/spec/initContainers` for native sidecar containers, while normal containers will continue to update `/spec/containers`.

### Test Plan

The following test scenarios will be added to e2e tests.

- Admission applies recommendations to native sidecars.
- Updater will update sidecar container resources in-place or evict.
- Admission will patch sidecar container resources.
- When the feature gate `NativeSidecar` is false VPA components will not modify native sidecars.

### Upgrade / Downgrade Strategy

#### Upgrade

On upgrade of the VPA to 1.6.0 (tentative release version), nothing will change,
VPAs will continue to work as before.

Users can use the new `NativeSidecar` by enabling the alpha Feature Gate (which defaults to disabled)
by passing `--feature-gates=NativeSidecar=true` to the VPA components.

#### Downgrade

On downgrade of VPA from 1.6.0 (tentative release version), nothing will change.
VPAs will continue to work as previously. Checkpoints may contain sidecar resource information until updated, but updater and admission will modify sidecar resources.

## Alternatives

### Treat as Standard Containers

We could treat them as standard containers, but they are technically init containers in the Pod spec, so the patch path would be incorrect (`/spec/containers` vs `/spec/initContainers`).
