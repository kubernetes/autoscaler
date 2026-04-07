# AEP-7571: Support Pod-Level Resources

<!-- toc -->
- [Summary](#summary)
  - [Goals](#goals)
  - [Non-Goals](#non-goals)
- [Glossary](#glossary)
- [Design Details](#design-details)
  - [Notes/Constraints/Caveats](#notesconstraintscaveats)
  - [Design Principles](#design-principles)
  - [Proposal](#proposal)
    - [Recommender](#recommender)
      - [New Global Pod maximums](#new-global-pod-maximums)
    - [Admission Controller](#admission-controller)
    - [capping.go](#cappinggo)
      - [Container LimitRange Objects](#container-limitrange-objects)
    - [Pod Annotations](#pod-annotations)
    - [Updater](#updater)
    - [Validation](#validation)
      - [Static Validation](#static-validation)
      - [Dynamic Validation](#dynamic-validation)
    - [Test Plan](#test-plan)
  - [Upgrade / Downgrade Strategy](#upgrade--downgrade-strategy)
    - [Upgrade](#upgrade)
    - [Downgrade](#downgrade)
  - [Kubernetes Version Compatibility](#kubernetes-version-compatibility)
  - [Examples](#examples)
    - [Example 1](#example-1)
    - [Example 2](#example-2)
    - [Example 3](#example-3)
- [Implementation History](#implementation-history)
<!-- /toc -->

## Summary

Starting with Kubernetes v1.34 (enabled by default, feature gate `PodLevelResources`), it is possible to specify CPU and memory `resources` for Pods at the pod level, in addition to the existing container-level `resources` specifications. For example:


```yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx
  namespace: default
spec:
  resources:
    requests:
      memory: "100Mi"
    limits:
      memory: "200Mi"
  containers:
  - name: container1
    image: nginx
```

It is also possible to combine both pod-level and container-level specifications. In this case, one container can define its own resource constraints - the `ide` container, while other containers (`tool1` and `tool2`) can dynamically use any remaining resources within the Pod's overall limit:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: workload
  namespace: default
spec:
  resources:
    limits:
      memory: "1024Mi"
      cpu: "4"
  containers:
    - name: ide
      image: theia:latest
      resources:
        requests:
          memory: "128Mi"
          cpu: "0.5"
        limits:
          memory: "256Mi"
          cpu: "1" 
    - name: tool1
      image: tool1:latest
    - name: tool2
      image: tool2:latest
```

The benefits and implementation details of pod-level `resources` are described in [KEP-2837](https://github.com/kubernetes/enhancements/blob/master/keps/sig-node/2837-pod-level-resource-spec/README.md). A related article is also available in the [Kubernetes documentation](https://kubernetes.io/docs/tasks/configure-pod-container/assign-pod-level-resources/).

Furthermore, the In-Place Pod Resize (IPPR) functionality has been extended to support pod-level resources, as defined in [KEP-5419](https://github.com/kubernetes/enhancements/blob/master/keps/sig-node/5419-pod-level-resources-in-place-resize/README.md).

Prior to this AEP, VPA computes recommendations only at the container level, and those recommendations are applied exclusively at the container level. With the new pod-level resource specifications, VPA should be capable of reading the pod-level resources stanza, calculating pod-level recommendations, and managing pod-level resource stanzas using any of the available VPA modes, including `InPlaceOrRecreate, InPlace etc`.

### Goals

- The recommender should be able to calculate pod-level recommendations when the user opts in.  
- The updater should be able to make eviction or in-place update decisions based on the pod-level `resources` stanza and pod-level recommendations.  
- The admission controller should be able to apply pod-level resources at admission.
- Support all existing VPA modes
- Document the new functionality in the [VPA documentation](https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler/docs) and on [kubernetes.io](https://kubernetes.io/docs/concepts/workloads/autoscaling/vertical-pod-autoscale/).

### Non-Goals

- Since the latest VPA does not support `initContainers` ([the native way to use sidecar containers](https://kubernetes.io/blog/2023/08/25/native-sidecar-containers/)), this AEP does not aim to implement support for them. In other words, `initContainers` are simply ignored when calculating pod-level recommendations. Support for `initContainers` may be explored in a future proposal.
- This AEP does not define a complete mechanism for updating pod-level resource stanzas without disruption. With this proposal in place, when VPA runs in `InPlaceOrRecreate` mode, the updater can apply pod-level recommendations in place, but this does not guarantee disruption-free updates.
- This proposal does not implement **in-place partial updates**. In-place partial updates refer to sending resize requests only to containers that require them, for example, when a container's current resource request falls outside the recommended range or when `resourceDiff` exceeds the threshold. This proposal does not change this existing behavior.

## Glossary

The following terms and definitions are used throughout this AEP:

* **Pod LimitRange object** - Refers to a `LimitRange` object with `type=Pod`.
* **Container LimitRange object** - Refers to a `LimitRange` object with `type=Container`.
* **Latest VPA version** - In this AEP, the term "latest VPA version" refers to version 1.6.0.

## Design Details

### Notes/Constraints/Caveats

- Since this new feature is an opt-in feature, merging the feature related changes won't impact existing workloads. Moreover, the feature will be rolled out gradually, beginning with an alpha release for testing and gathering feedback. This will be followed by beta and GA releases as the feature matures and potential problems and improvements are addressed.

### Design Principles

Before this AEP, the recommender computes recommendations only at the container level. With this proposal, the recommender also computes pod-level recommendations in addition to container-level ones. Pod-level recommendations derive from per-container recommendations. Container-level policy (the `containerPolicies` stanza) influences pod-level recommendations: setting `mode: Off` in `spec.resourcePolicy.containerPolicies` excludes a container from pod-level recommendations, and `minAllowed` and `maxAllowed` bounds continue to apply.

This AEP extends the VPA CRD `spec.resourcePolicy` with a new `podPolicies` stanza that influences pod-level recommendations (and possibly container-level recommendations as well). The AEP also introduces two global pod-level flags `pod-recommendation-max-allowed-cpu` and `pod-recommendation-max-allowed-memory`. Details are covered in later section.

Today, the updater makes decisions based solely on container-level resource stanzas and recommendations, and both the updater and admission controller modify resources only at the container level. This proposal extends the updater to make eviction or in-place update decisions using pod-level resources and recommendations, and adds support for the admission controller to compute and apply pod-level resource patches.

### Proposal

- Add a new VPA feature flag named `VPAPodLevelResources` to each VPA component: the recommender, updater, and admission controller.

- Extend the VPA object:

  1. Add a new enum to the container resource policy's `mode` field with the value `RecommendationOnly`. This enum allows users to indicate that the container-level recommendation should be calculated by the recommender, while the updater and admission controller ignore it. For example, the updater will not evict the pod based on this container-level recommendation. The primary purpose of this enum is to give users finer control compared to the `Auto` mode, which instructs VPA to calculate the container-level recommendation and manage the corresponding resource stanza. The AEP includes example VPA objects and behaviors in the [Examples](#examples) section.

  2. Add a new `spec.resourcePolicy.podPolicies` stanza. This stanza allows setting constraints for pod-level recommendations:

    - `mode`: This new field is similar to the one that already exists at the container level, with two possible values: `Off` and `Auto`. It indicates whether the autoscaler is enabled at the pod level. The default is `Off`, so the autoscaler does not compute recommendations at the pod level, and the updater and admission controller ignore the pod-level `resources` stanza.

     - `controlledResources`: Specifies which resource types are recommended (and possibly applied). Valid values are `cpu`, `memory`, or both. If not specified, both resource types are controlled by VPA.

     - `controlledValues`: Specifies which resource values are controlled. Valid values are `RequestsAndLimits` and `RequestsOnly`. The default is `RequestsAndLimits`.

     - `minAllowed`: Specifies the minimum amount of resources that the autoscaler recommends at the pod level. The default is no minimum.

     - `maxAllowed`: Specifies the maximum amount of resources that the autoscaler recommends at the pod level. The default is no maximum.

  3. Add a new `status.recommendation.podRecommendations` stanza. The field is populated by the VPA Recommender and stores Pod-level recommendations, including Pod-level `Target`, `LowerBound`, `UpperBound` and `UncappedTarget` for CPU and memory. For more details about how VPA calculates pod-level recommendations, see the [recommender](#recommender) section.

#### Recommender

This section and its subsections describe the proposed modifications to the recommender:

- Implement a new function that calculates pod-level recommendations when the `mode` field is set to `Auto` by the user in the `spec.resourcePolicy.podPolicies` stanza. The calculation loops through all container-level recommendations and sums values of the same type. For example, summing all container-level targets produces the pod-level target.

- Update the `ApplyVPAPolicy` function to enforce the new recommender-level constraints, such as the `pod-recommendation-max-allowed-cpu` and `pod-recommendation-max-allowed-memory` flags, as well as the pod-level `minAllowed` and `maxAllowed` values for both **pod-level and container-level recommendations**. To enforce pod-level constraints on container-level recommendations, this document proposes using the formula introduced by @omerap12 here: [see discussion](https://github.com/kubernetes/autoscaler/issues/7147#issuecomment-2515296024).

##### New Global Pod maximums

This AEP proposes adding two new recommender flags to cap pod-level recommendations:
- `--pod-recommendation-max-allowed-cpu`: Sets the global maximum CPU the recommender may recommend at Pod scope.
- `--pod-recommendation-max-allowed-memory`: Sets the global maximum memory the recommender may recommend at Pod scope.

Per-container recommendations must be adjusted so that their total does not exceed these global constraints. The Pod-level `maxAllowed` takes precedence over these global Pod maximums when both are present.

#### Admission Controller

The admission controller must be extended to generate pod-level resource patches alongside container-level patches. The admission controller should follow these rules after implementing this proposal when pod-level scaling is enabled in the VPA object:

1. Calculate only container-level patches for containers whose `mode` is set to `Auto`. Ignore containers with `Off` mode, as well as containers whose `mode` is set to `RecommendationOnly`.

2. Calculate pod-level patches based on pod-level recommendations.

#### capping.go

The `capping.go` logic must be updated so that the admission controller and the updater can adjust pod-level recommendations based on Pod LimitRange objects in the cluster.

##### Container LimitRange Objects

This AEP proposes to skip enforcing constraints defined by container LimitRange objects for pod-level recommendations. The built-in LimitRanger admission controller automatically sets default container-level requests and limits for these Pods, which is undesirable when pod-level resources are in use.

This behavior is also described in the [Pod-level Resource Spec KEP](https://github.com/kubernetes/enhancements/blob/master/keps/sig-node/2837-pod-level-resource-spec/README.md#admission-controller), which recommends placing Pods in namespaces without container LimitRange objects.

To handle container LimitRange objects for Pods whose pod-level resources stanza is managed by VPA, this AEP proposes the following:
- Inform the user through logs produced by the updater or admission controller that a container LimitRange object exists in the namespace, and suggest that the user remove it from the cluster. The implementation does not enforce these constraints on pod-level recommendations in this case.

#### Pod Annotations

Define new pod-level annotations that the admission controller and the updater apply as needed:
* Annotate the Pod when Pod-level requests and limits are managed. Explicitly indicate when a Pod-level limit is not set.
* Annotate the Pod when a Pod-level resource limit (CPU or memory) is capped to fit within the `Min` or `Max` defined in a Pod LimitRange object.

#### Updater

With the introduction of pod-level recommendations and pod-level resources, the updater must make eviction and in-place update decisions for these Pods. Achieve this by making the following updates to its source code:

1. Extend the [GetUpdatePriority](https://github.com/kubernetes/autoscaler/blob/master/vertical-pod-autoscaler/pkg/updater/priority/priority_processor.go#L45) method to evaluate pod-level recommendations as well. The updated method verifies whether pod-level recommendations fall outside the recommended range, calculates the pod-level `resourceDiff`, and checks whether a scale-up occurs. The function should still return a single `PodPriority` struct. For example, if a multi-container Pod requires resource management for only one container and at the pod level, run `GetUpdatePriority` for that container and for the pod level.

#### Validation

##### Static Validation

- The new fields under `spec.resourcePolicy.podPolicies` will be validated consistently, following the same approach we use for the existing `containerPolicies` stanza.
- This AEP proposes validating the new global flags, `--pod-recommendation-max-allowed-cpu` and `--pod-recommendation-max-allowed-memory`, as Kubernetes `Quantity` values using `resource.ParseQuantity`.

##### Dynamic Validation

The admission controller should validate the following rules. If a rule is violated, it should return an error to the user:
  * When the pod-level `minAllowed` is set by the user, the sum of container-level `minAllowed` values must be equal to or greater than the pod-level constraint.
  * When the pod-level `maxAllowed` is set by the user, the sum of container-level `maxAllowed` values must be equal to or less than the pod-level constraint.

#### Test Plan

This AEP proposes comprehensive unit test coverage for all new code paths and additionally plans to introduce the following e2e tests:

- Recommender: Process a Kubernetes Deployment whose Pod template defines multiple containers, has pod-level scaling set to `Auto`, and sets the container-level scaling mode to `RecommendationOnly` for every container. Verify that the generated VPA status stanza contains pod-level recommendations.

- Updater: Process a Kubernetes Deployment whose Pod template defines multiple containers, includes pod-level resources, has pod-level scaling set to `Auto`, and sets the container-level scaling mode to `RecommendationOnly`. Validate behavior in at least the `Recreate` and `InPlaceOrRecreate` modes.

- Admission Controller: Process a Kubernetes Deployment whose Pod template defines multiple containers, includes pod-level resources, has pod-level scaling set to `Auto`, and sets the container-level scaling mode to `RecommendationOnly` for all containers. Test at least the `Recreate` VPA mode, ensuring that pod-level recommendations are applied.

- Add a full e2e test in which all VPA components are used in these scenarios:
  - Multi-container Pod with pod-level scaling enabled, where at least one container's scaling mode is set to `Auto` and others are set to `RecommendationOnly`. In this case, pod-level resources and the `Auto` container-level resources should be managed.
  - Multi-container Pod with pod-level scaling enabled, where all containers have scaling mode set to `RecommendationOnly`. In this case, only pod-level resources should be managed.

### Upgrade / Downgrade Strategy

#### Upgrade

Use a VPA release that includes this feature across all three components, and pass `--feature-gates=VPAPodLevelResources=true` to each component. Begin deploying new workloads where the VPA object has pod-level scaling enabled, and update container-level scaling modes appropriately to reflect user requirements.

#### Downgrade

Downgrading VPA from a version that includes this feature does not impact running workloads where pod-level scaling is not enabled.

After the downgrade, workloads with pod-level scaling enabled are not disrupted immediately. Pod-level resource specifications remain in Pod specs, but VPA reverts to managing container-level resources only.

However, once VPA stops managing pod-level resources and updates only container-level resources, newly applied container-level requests or limits may violate the unmanaged pod-level constraints. To avoid such violations, follow these steps for downgrade:

1. Disable pod-level scaling in the VPA object and update container-level scaling modes, for example by changing `RecommendationOnly` mode to `Auto` or `Off`.
2. Remove the pod-level `resources` stanza from the higher-level controller, if present.
3. Downgrade the VPA version and disable the feature flag at the same time.

### Kubernetes Version Compatibility

This new feature, proposed by this AEP, relies on two feature gates:
* `PodLevelResources` - implements pod-level resources.
* `InPlacePodLevelResourcesVerticalScaling` – implements in-place scaling at the pod level.

Therefore, it is recommended to use a Kubernetes version that includes both of these feature gates (use [this page](https://kubernetes.io/docs/reference/command-line-tools-reference/feature-gates) to verify this information). The alpha version of `InPlacePodLevelResourcesVerticalScaling` was released in v1.35, so at least this Kubernetes version should be used.

### Examples

#### Example 1

Based on the manifest below, the user directs the recommender to calculate container-level recommendations for all containers. Summing these recommendations produces the pod-level recommendation. Recommendations at any level are stored under the VPA object's `status` stanza.

The updater makes eviction decisions based only on the pod-level `resources` stanza and the pod-level recommendation. When the pod is recreated, the admission controller updates only the pod-level resources.

The recommended approach for creating the higher-level controller (for example, a Kubernetes Deployment) is to avoid specifying container-level resources, as these are not managed by VPA. This allows VPA to provide pod-level recommendations without conflicting with pre-set container-level resource stanzas. The pod-level resources can be set by the user to ensure that VPA can calculate the pod-level request-to-limit ratio, update the pod-level requests accordingly, and set the corresponding limits.

```yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: workload1
  namespace: default
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: workload1
  updatePolicy:
    updateMode: 'Recreate'
  resourcePolicy:
    containerPolicies:
      - containerName: "*"
        mode: "RecommendationOnly"
    podPolicies:
      mode: "Auto"
```

#### Example 2

Using the manifest below, the user directs the recommender to calculate container-level recommendations for all containers and the pod-level recommendation as well. The results, as before, are stored under the `status` stanza.

The updater may now make eviction decisions based on the resource stanza and recommendation for the container named `main`, or based on the pod-level `resources` stanza and the pod-level recommendation.

The admission controller sets the container-level `resources` stanza for the `main` container. If the request-to-limit ratio can be determined for the container, it adjusts the request according to the ratio and sets the container limit as well. Additionally, it sets the pod-level resources including the limits if possible.

The recommended approach, if the user wants to maintain a specific request-to-limit ratio for the `main` container and at the pod level, is to set both `resources` stanzas initially when the workload is created.

```yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: workload1
  namespace: default
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: workload1
  updatePolicy:
    updateMode: 'Recreate'
  resourcePolicy:
    containerPolicies:
      - containerName: "main"
        mode: "Auto"
      - containerName: "*"
        mode: "RecommendationOnly"
    podPolicies:
      mode: "Auto"
```

#### Example 3

The manifest below is an example of incorrect usage of this new feature, as the user directs the updater and admission controller to continue managing container-level `resources` stanzas for all containers (because the container-level scaling mode defaults to `Auto`), preventing any benefit from the pod-level `resources` stanza.

```yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: workload1
  namespace: default
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: workload1
  updatePolicy:
    updateMode: 'Recreate'
  resourcePolicy:
    podPolicies:
      mode: "Auto"
```

## Implementation History

- 2025-09-29: initial version