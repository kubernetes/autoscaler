# AEP-7571: Support Pod-Level Resources

<!-- toc -->
- [Summary](#summary)
  - [Goals](#goals)
  - [Non-Goals](#non-goals)
- [Glossary](#glossary)
- [Design Details](#design-details)
  - [Notes/Constraints/Caveats](#notesconstraintscaveats)
  - [Design Principles](#design-principles)
    - [Container-level resources](#container-level-resources)
    - [Pod-level resources](#pod-level-resources)
    - [Pod and Container-Level Resources](#pod-and-container-level-resources)
      - [Option 1: VPA Controls Only Pod-Level Resources](#option-1-vpa-controls-only-pod-level-resources)
      - [[Selected] Option 2: VPA controls pod-level resources and the initially set container-level resources](#selected-option-2-vpa-controls-pod-level-resources-and-the-initially-set-container-level-resources)
    - [No resources stanza exists at the pod and container level](#no-resources-stanza-exists-at-the-pod-and-container-level)
      - [[Selected] Option 1: VPA controls only the container-level resources](#selected-option-1-vpa-controls-only-the-container-level-resources)
      - [Option 2: VPA controls only the pod-level resources](#option-2-vpa-controls-only-the-pod-level-resources)
  - [Proposal](#proposal)
    - [Recommender](#recommender)
      - [Pod-level Policies](#pod-level-policies)
      - [New Global Pod maximums](#new-global-pod-maximums)
    - [Admission Controller](#admission-controller)
      - [Patch Generation Algorithm](#patch-generation-algorithm)
    - [Updater](#updater)
    - [Pod Annotations](#pod-annotations)
    - [LimitRange objects](#limitrange-objects)
      - [Container LimitRange objects](#container-limitrange-objects)
      - [Pod LimitRange objects](#pod-limitrange-objects)
  - [Validation](#validation)
    - [Static Validation](#static-validation)
    - [Dynamic Validation](#dynamic-validation)
      - [Validating controlledResources and podPolicies](#validating-controlledresources-and-podpolicies)
  - [Test Plan](#test-plan)
  - [Upgrade / Downgrade Strategy](#upgrade--downgrade-strategy)
    - [Upgrade](#upgrade)
    - [Downgrade](#downgrade)
  - [Kubernetes Version Compatibility](#kubernetes-version-compatibility)
- [Implementation History](#implementation-history)
<!-- /toc -->

## Summary

Starting with Kubernetes v1.34 (enabled by default), it is possible to specify CPU and memory `resources` for Pods at the pod level, in addition to the existing container-level `resources` specifications. For example:


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

Currently, before this AEP, VPA computes recommendations only at the container level, and those recommendations are applied exclusively at the container level. With the new pod-level resources specifications, VPA should be able to read from the pod-level `resources` stanza, calculate pod-level recommendations, and scale at the pod level when users define pod-level `resources`.

### Goals

- Add support for the pod-level resources stanza in VPA
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

This section describes how VPA reacts based on where resources are defined (pod level, container level or both).

Before this AEP, the recommender computes recommendations only at the container level. With this proposal, the recommender also computes pod-level recommendations in addition to container-level ones. Pod-level recommendations derive from per-container recommendations. Container-level policy (the `containerPolicies` stanza) influences pod-level recommendations: setting `mode: Off` in `spec.resourcePolicy.containerPolicies` excludes a container from pod-level recommendations, and `minAllowed` and `maxAllowed` bounds continue to apply.

This AEP extends the VPA CRD `spec.resourcePolicy` with a new `podPolicies` stanza that influences pod-level recommendations. The AEP also introduces two global pod-level flags `pod-recommendation-max-allowed-cpu` and `pod-recommendation-max-allowed-memory`. Details are covered in the [Proposal section](#proposal).

Today, the updater makes decisions based solely on container-level resource stanzas, and both the updater and the admission controller modify resources only at the container level. This proposal extends the updater to make eviction or in-place update decisions for Pods that define pod-level resources, and adds support for the admission controller to compute and apply pod-level resource patches as well.

**This AEP suggests that when a workload defines pod-level resources, VPA should manage those by default because pod-level resources offer benefits over container-only settings** - see the "Better resource utilization" section in [KEP-2837](https://github.com/kubernetes/enhancements/blob/master/keps/sig-node/2837-pod-level-resource-spec/README.md#better-resource-utilization) for details.

Scenarios with no resources defined, or with both pod-level and container-level values present, require clear defaulting rules and are discussed in the options below. Note: community feedback should determine the default behavior. 

#### Container-level resources

For workloads that define only container-level resources, VPA should continue controlling resources at the container level, consistent with current behavior prior to this AEP (i.e. version 1.6.0). In other words, for a multi-container Pod that does not define pod-level resources but has at least one container specifying resources, VPA should, by default, manage only container-level resources.

#### Pod-level resources

For workloads that define only pod-level resources, VPA will control resources only at the pod level.

#### Pod and Container-Level Resources

This part of the AEP covers workloads that define resources both at the pod level and for at least one container. To demonstrate multiple implementation options for how VPA should handle such workloads by default, consider the following manifest. It defines three containers:  
  * `ide` - the main workload container  
  * `tool1` and `tool2` - non-critical sidecar containers (that is, a non-native sidecar pattern is used)

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

##### Option 1: VPA Controls Only Pod-Level Resources

With this option, VPA manages only the pod-level resources stanza. To follow this approach, the initially defined container-level resources for `ide` must be removed so that changes in usage are reflected only in pod-level recommendations.

**Pros**:
* VPA does not need to track which container-level resources were initially set.
* Enables shared headroom across containers in the same Pod. With container-only limits, a sidecar (`tool1` or `tool2`) or the main workload (`ide` container) hitting its own CPU limit could get throttled even if other containers in the Pod have idle CPU. Pod-level resources allow a container experiencing a spike to access idle resources from others, optimizing overall utilization.

**Cons**: 
* A downside of this approach is that the most important container (`ide`) may be recreated without container-level resources, leading to an `oom_score_adj` that matches other sidecars in the Pod, as a result the OOM killer may target all containers more evenly under node memory pressure. For details on how `oom_score_adj` is computed when pod-level resources are present, see the [KEP section](https://github.com/kubernetes/enhancements/blob/master/keps/sig-node/2837-pod-level-resource-spec/README.md#oom-score-adjustment) on OOM score adjustment.

##### [Selected] Option 2: VPA controls pod-level resources and the initially set container-level resources

With this option, VPA controls pod-level resources and the container-level resources that were initially set by the user. The resources for containers `tool1` and `tool2` are not updated by VPA, however their usage is still observed and contributes to the overall pod-level recommendation.

**Pros**:
- The primary container (`ide`) is less likely to be killed under memory pressure because sidecars (`tool1`, `tool2`) have higher `oom_score_adj` values, the OOM killer targets them first during node pressure evictions. See the [updated OOM killer behavior and formula](https://github.com/kubernetes/enhancements/blob/master/keps/sig-node/2837-pod-level-resource-spec/README.md#oom-killer-behavior) when pod-level resources are present.
- Sidecars, such as logging agents or mesh proxies (like `tool1` or `tool2`), that don't use container-level limits can borrow idle CPU from other containers in the pod when they experience a spike in usage. Pod-level resources allow a container experiencing a spike to access idle resources from others, optimizing overall utilization.

**Cons**:
- Existing VPA users may find the behavior surprising because VPA does not control all container-level resources stanzas - only those initially set.

#### No resources stanza exists at the pod and container level

When a workload is created without any resources defined at either the pod or container level, there are two options:

##### [Selected] Option 1: VPA controls only the container-level resources

This option mirrors current VPA behavior by managing only container-level resources. In this mode, pod-level recommendations are not computed and therefore not applied.

**Pros**:
- Familiar to existing users because it does not change current VPA behavior.

**Cons**:
- No cross-container resource sharing: in multi-container Pods, a container can hit its own limit and be throttled even if sibling containers are idle.

##### Option 2: VPA controls only the pod-level resources

With this option, VPA computes and applies only pod-level recommendations.

**Pros**:
- Enables shared headroom across containers in the same Pod, previously with container-only limits, a sidecar (e.g. `tool1` or `tool2` container) hitting its own CPU limit could throttle the Pod even if other containers had spare CPU. Pod-level resources allows one container experiencing a spike to access idle resources from others, optimizing overall utilization.

**Cons**:
- Choosing this option is a breaking change because it modifies the default behavior. The following occurs:
  1. Users enable the new VPA feature gate.
  2. VPA recreates existing Pods whose higher-level controller does not define a resources stanza, applying a pod-level resources stanza.

At the same time, this option benefits users who provision workloads without a resources stanza and expect VPA to manage only pod-level resources. However, this document does not pursue this option.

### Proposal

- Add a new VPA feature flag named `PodLevelResources` to each VPA component: the recommender, updater, and admission controller.

- Extend the VPA object:
  1. Add a new `spec.resourcePolicy.podPolicies` stanza. This stanza is user-modifiable and allows setting constraints for pod-level recommendations:
     - `controlledResources`: Specifies which resource types are recommended (and possibly applied). Valid values are `cpu`, `memory`, or both. If not specified, both resource types are controlled by VPA.

     - `controlledValues`: Specifies which resource values are controlled. Valid values are `RequestsAndLimits` and `RequestsOnly`. The default is `RequestsAndLimits`.

     - `minAllowed`: Specifies the minimum amount of resources that can be recommended for the Pod. The default is no minimum. This Pod-level field must be enforced by the recommender as it is recommendation constraint. For more details about this field, see the [Pod-level Policies](#pod-level-policies) section.

     - `maxAllowed`: Specifies the maximum amount of resources that can be recommended for the Pod. The default is no maximum. For more details about this field, see the [Pod-level Policies](#pod-level-policies) section.

  2. Add a new `status.recommendation.podRecommendation` stanza. The field is populated by the VPA Recommender and stores Pod-level recommendations, including Pod-level `Target`, `LowerBound`, and `UpperBound` for CPU and memory. The consumers of these new fields include not only end users, but also the updater and the admission controller. For more details about how VPA calculates pod-level recommendations, see the [recommender](#recommender) section.

#### Recommender

This section and its subsections describe the proposed modifications to the recommender:

- Modify the [PodState](https://github.com/kubernetes/autoscaler/blob/master/vertical-pod-autoscaler/pkg/recommender/model/cluster.go#L116) struct to store pod-level resource requests when present:

```go
type PodState struct {
	// Unique id of the Pod.
	ID PodID
	// Set of labels attached to the Pod.
	labelSetKey labelSetKey
	// Containers that belong to the Pod, keyed by the container name.
	Containers map[string]*ContainerState
  // Current Pod-level requests (!NEW)
	PodLevelRequest map[ResourceName]ResourceAmount
	// InitContainers is a list of init containers names which belong to the Pod.
	InitContainers []string
	// PodPhase describing current life cycle phase of the Pod.
	Phase apiv1.PodPhase
}
```

- Add a new method that populates the `PodLevelRequest` field. Invoke this method from the [LoadPods method](https://github.com/kubernetes/autoscaler/blob/ce55a882414e127c49d3ceb6b0f1790062950fce/vertical-pod-autoscaler/pkg/recommender/input/cluster_feeder.go#L467).

- Update the [processVPAUpdate](https://github.com/kubernetes/autoscaler/blob/master/vertical-pod-autoscaler/pkg/recommender/routines/recommender.go#L83) logic so that the VPA status stanza includes both container-level and pod-level recommendations. The recommender determines whether to calculate pod-level recommendations by inspecting the `PodLevelRequest` field. If this field is empty, the recommender skips pod-level processing. If the `PodLevelRequest` field contains at least one resource request, the recommender calculates pod-level recommendations derived from container-level recommendations.

  This new code path operates as follows:
  1. Collect container-level recommendations for CPU and memory. Users may disable recommendations for specific containers or resource types (see the [Validating controlledResources and podPolicies](#validating-controlledresources-and-podpolicies) section for examples).
  2. Aggregate the collected container-level recommendations by summing values per resource type and per recommendation type. For example, summing all container-level CPU `lowerBound` values produces the pod-level CPU `lowerBound` recommendation.

  The following example shows the VPA status when the user defines pod-level resource requests for both CPU and memory in the higher-level controller (for example, a Deployment):

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
  resourcePolicy: {} # No restrictions on which recommendations are computed
status:
  recommendation:
    podRecommendation:
      lowerBound:
        memory: 100Mi # 80 + 20
        cpu: 30m # 20 + 10
      target:
        memory: 125Mi # 100 + 25
        cpu: 50m # 30 + 20
      upperBound:
        memory: 150Mi # 120 + 30
        cpu: 80m # 50 + 30
    containerRecommendations:
    - containerName: main
      lowerBound:
        memory: 80Mi
        cpu: 20m
      target:
        memory: 100Mi
        cpu: 30m
      uncappedTarget:
        memory: 120Mi
        cpu: 50m
      upperBound:
        memory: 120Mi
        cpu: 50m
    - containerName: sidecar1
      lowerBound:
        memory: 20Mi
        cpu: 10m
      target:
        memory: 25Mi
        cpu: 20m
      uncappedTarget:
        memory: 30Mi
        cpu: 30m
      upperBound:
        memory: 30Mi
        cpu: 30m
```

##### Pod-level Policies

The new pod-level fields `podPolicies.minAllowed` and `podPolicies.maxAllowed` serve as recommendation constraints. When a calculated pod-level `Target` recommendation violates these constraints, the recommender adjusts the pod-level recommendations (`Target`, `LowerBound`, and `UpperBound`) and proportionally updates the container-level recommendations.

###### podPolicies.minAllowed:

If the calculated pod-level `Target` recommendation is less than `podPolicies.minAllowed`, the recommender sets the pod-level `Target` recommendation to `minAllowed` and records it in the VPA status stanza. Each container's recommendation is then increased proportionally, using the same formula applied by the updater and the admission controller when a Pod LimitRange with a `min` field is present.

The proposed algorithm works as follows:

- For each resource type in `podPolicies.minAllowed` (e.g. CPU or memory):
  - For each container in the Pod:
    - Retrieve the container-level recommendation (i.e. container `Target`) for the resource.
    - Call the appropriate proportional scaling function for the resource type.  
      ```go
      // Example for memory:
      scaleQuantityProportionallyMem(
          containerTargetMemRecommendation,
          podLevelTargetMemRecommendation,
          podLevelMemMinAllowed,
      )
      ```
    - Record the result from the previous step as the new container-level `Target` recommendation.
    - Update the container-level `LowerBound` and `UpperBound` recommendations based on the scaled value (`Target`).
  - After all containers are processed, sum the updated container-level `LowerBound` and `UpperBound` values to compute the pod-level `LowerBound` and `UpperBound` recommendations.

###### podPolicies.maxAllowed

If the calculated pod-level `Target` recommendation exceeds `podPolicies.maxAllowed`, the recommender sets the pod-level `Target` recommendation to `maxAllowed` and records it in the VPA status stanza. Each container's recommendation is then adjusted proportionally using the same formula as before, with the only difference being that the constraint comes from `podPolicies.maxAllowed`.

##### New Global Pod maximums

This AEP proposes adding two new recommender flags to cap pod-level recommendations:
- `--pod-recommendation-max-allowed-cpu`: Sets the global maximum CPU the recommender may recommend at Pod scope.
- `--pod-recommendation-max-allowed-memory`: Sets the global maximum memory the recommender may recommend at Pod scope.

Per-container recommendations must be adjusted so that their total does not exceed these global constraints, using the same enforcement formula described in the `podPolicies.minAllowed` and `podPolicies.maxAllowed` sections. The VerticalPodAutoscaler maximum (Pod-level or container-level `maxAllowed`) takes precedence over these global Pod maximums when both are present.

#### Admission Controller

The admission controller must be extended to generate pod-level resource patches alongside container-level patches. The container-level patches applied when a pod-level resources stanza is present are determined by the user's intent, as expressed in the higher-level controller (for example, a Deployment). For instance:

Suppose an user creates a Deployment with a Pod containing three containers. The user intends for the VPA (i.e. admission controller) to manage the pod-level resources stanza and only one container-level resources stanza (for the `main` container). To achieve this, the user deploys the following manifest to the cluster along with a VPA object in `Recreate` mode:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: workload1
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: workload1
  template:
    metadata:
      labels:
        app: workload1
    spec:
      resources:
       requests:
         cpu: 100m
         memory: 100Mi
       limits:
         cpu: 200m
         memory: 200Mi
      containers:
        - name: main
          image: registry.k8s.io/pause:3.1
          resources:
            requests:
              cpu: 80m
              memory: 100Mi
            limits:
              cpu: 80m
              memory: 100Mi
        - name: sidecar1
          image: registry.k8s.io/pause:3.1
        - name: sidecar1
          image: registry.k8s.io/pause:3.1
```

When a Pod creation event occurs, the admission controller checks the VPA status and the Pod's resources stanza. For each resource request, it retrieves the corresponding recommendation from the VPA status and generates the appropriate patches. The algorithm in pseudocode is as follows:

##### Patch Generation Algorithm

1. Retrieve all recommendations from the VPA status stanza for the Pod, including both pod-level and container-level recommendations.
2. Retrieve the resource stanzas from the Pod spec. For each resource request (pod-level or container-level):
   1. Find the matching recommendation (`Target`) in the VPA status and **record it as a patch**.
      - If no recommendation is found, skip this resource request, log a message, and continue with the next resource request in the list. Log message formats:
        * `"No recommendation found for container, skipping" container="<container_name>"` or:
        * `"No recommendation found for pod, skipping" pod="<pod_name>"`
   2. If a corresponding limit exists, scale it proportionally to preserve the original request-to-limit ratio. **Record the scaled limit as a patch**.
   3. Repeat steps 2.1–2.2 for all remaining resource requests defined in the Pod spec.
3. Apply all recorded patches to the Pod.

The [CalculatePatches](https://github.com/kubernetes/autoscaler/blob/master/vertical-pod-autoscaler/pkg/admission-controller/resource/pod/patch/resource_updates.go#L54) method should be extended to support the new logic described above.

Note: The current patch calculation algorithm in the latest VPA does not filter recommendations stored in the VPA status stanza based on the resource requests present in the Pod spec. This AEP adds a new condition: when a pod-level resources stanza is present, VPA follows the proposed algorithm described above, otherwise VPA follows the existing behavior.

#### Updater

With the introduction of pod-level recommendations, the updater must make eviction and in-place update decisions for these Pods.

This section and its subsections describe the proposed modifications to the updater. Parts marked with `[DOESN'T CHANGE]` remain unchanged by this AEP:

1. Extend the [GetUpdatePriority](https://github.com/kubernetes/autoscaler/blob/master/vertical-pod-autoscaler/pkg/updater/priority/priority_processor.go#L45) method to also evaluate pod-level recommendations. The updated method verifies whether pod-level recommendations fall outside the recommended range and calculates the pod-level `resourceDiff`. These checks occur only when container-level recommendations do not set `OutsideRecommendedRange` to true and the container-level `resourceDiff` remains below the threshold.
2. `[DOESN'T CHANGE]` When the updater adds a Pod to the `UpdatePriorityCalculator`, it marks the Pod for eviction or in-place update based on the VPA mode:
   1. `[DOESN'T CHANGE]` If the updater evicts the Pod, control passes to the admission controller, and the updater proceeds to the next Pod in the list.
   2. If the updater selects the Pod for an in-place update, it applies the pod-level and container-level recommendations directly to the running Pod using the in-place mechanism, based on the presence of resource requests in the Pod spec. The algorithm follows the approach proposed in the admission controller subsection - see [Patch Generation Algorithm](#patch-generation-algorithm). If the in-place update fails, the updater falls back to the eviction path.

#### Pod Annotations

Define new pod-level annotations that the admission controller and the updater (when using in-place) apply as needed:
* Annotate the Pod when Pod-level requests and limits are managed. Explicitly indicate when a Pod-level limit is not set.
* Annotate the Pod when a Pod-level resource limit (CPU or memory) is capped to fit within the `Min` or `Max` defined in a Pod LimitRange object.

#### LimitRange objects

The admission controller and the updater should respect constraints defined by LimitRange objects in the cluster.

##### Container LimitRange objects

This AEP recommends that Pods with Pod-level resources should not be deployed in namespaces that contain a Container LimitRange object. The built-in LimitRanger admission controller automatically sets default container-level requests and limits for these Pods, which is undesirable when Pod-level resources are in use.

This behavior is also described in the [Pod-level Resource Spec KEP](https://github.com/kubernetes/enhancements/blob/master/keps/sig-node/2837-pod-level-resource-spec/README.md#admission-controller), which suggests placing Pods in namespaces without any Container LimitRange objects. Therefore this AEP proposes that creation of Pods with Pod-level resources in namespaces containing Container LimitRange objects should be rejected. This validation is further detailed in the [Validation section](#dynamic-validation).

In summary Pods with Pod-level resources should not be validated against Container LimitRange objects.

##### Pod LimitRange objects

For the admission controller, the most relevant recommendation type is the Target, since the controller calculates requests and limits. For the updater, the other Pod-level recommendation types - LowerBound and UpperBound - are used to determine which Pods should be evicted.

This AEP proposes applying the same logic for Pod-level recommendations as defined in the [capProportionallyToPodLimitRange](https://github.com/kubernetes/autoscaler/blob/722902dce0d3fba2a1eb9904a4d91e448496a7f8/vertical-pod-autoscaler/pkg/utils/vpa/capping.go#L514) method. In other words, the Pod-level Target, UpperBound, and LowerBound should be increased or decreased proportionally to comply with Pod LimitRange constraints.

For example:
* Pod-level `LowerBound` = 100 Mi
* Pod-level `Target` = 120 Mi
* Pod-level `UpperBound` = 150 Mi
* Pod LimitRange `min` = 200 Mi

In this case, the Pod-level `LowerBound`, `Target`, and `UpperBound` should all be increased to 200 Mi. Similarly if a maximum field in the Pod LimitRange is violated, the amounts should be proportionally decreased.

For Pods with Pod-level resources that also define container-level resources stanzas, container-level resource limits and recommendations should be checked against Pod LimitRange objects in the same way as in the latest VPA using `capProportionallyToPodLimitRange` method.

For example:
* Pod with two containers: c1 and c2
* Pod-level resources stanza is present
* Container-level stanza is present only for c1
* Calculated container-level `Target`: c1 = 120 Mi, c2 = 30 Mi
* Pod LimitRange `min` = 200 Mi

To respect the Pod LimitRange, container-level targets are increased proportionally:
* c1 -> 160 Mi (calculated as`(120/150)×200`)
* c2 -> 40 Mi (calculated as `(30/150)×200`)

### Validation

#### Static Validation

- The new fields under `spec.resourcePolicy.podPolicies` will be validated consistently, following the same approach we use for the existing `containerPolicies` stanza.
- This AEP proposes validating the new global flags, `--pod-recommendation-max-allowed-cpu` and `--pod-recommendation-max-allowed-memory`, as Kubernetes `Quantity` values using `resource.ParseQuantity`.

#### Dynamic Validation

Admission Controller:

* The pod-level `minAllowed` must be equal to or greater than the sum of the container-level `minAllowed` values. For example, if an end user has a pod with three containers - C1, C2, and C3 - and defines `minAllowed` for C1 and C2, then the sum of these values must be equal to or lower than the pod-level `minAllowed` specified in the VPA object. The same rule applies to the pod-level `maxAllowed` value as well.
* Creation of Pods with Pod-level resources should be rejected in namespaces that contain Container LimitRange objects. See the [Container LimitRange](#container-limitrange-objects) section for more details.

##### Validating controlledResources and podPolicies

The admission controller validates the correctness of the `controlledResources` fields at both the `containerPolicies` and the new `podPolicies` levels. Because pod-level recommendations derive from container-level recommendations, the admission controller rejects a VPA object when the user configures incompatible resource scopes.  

For example, in the following VPA object, the user instructs VPA to calculate container-level recommendations only for CPU. As a result, VPA cannot derive a pod-level memory recommendation, and the admission controller must reject this configuration:

```yaml
# Invalid VPA object
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
        controlledResources:
          - "cpu"
    podPolicies:
      controlledResources:
        - "memory"
```

The rule is as follows: when a user defines a `podPolicies` stanza with `controlledResources`, each resource listed in `podPolicies.controlledResources` must appear in the `controlledResources` field of at least one entry in `containerPolicies`. If this condition is met, the admission controller validates the VPA object, even if the configuration provides limited practical value.

For example, the following VPA object is valid because one container policy enables memory recommendations, which allows VPA to derive a pod-level memory recommendation:

```yaml
# Target Pods have three containers (c1, c2, and c3).
# Valid VPA object, but impractical
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
      - containerName: "c1"
        mode: "Off"
      - containerName: "c2"
        mode: "Off"
      - containerName: "c3"
        mode: "Auto"
        controlledResources:
          - "memory" # Changing this field to "cpu" causes validation to fail
    podPolicies:
      controlledResources:
        - "memory"
```

The following example illustrates a more appropriate configuration. The user chooses not to calculate recommendations for the `sidecar` container. As a result, VPA excludes the `sidecar` container from pod-level recommendation calculations. VPA also omits container-level recommendations for this container from the VPA status stanza:

```yaml
# Target Pod has three containers (c1, c2, and sidecar).
# Valid VPA object
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
      - containerName: "sidecar"
        mode: "Off"
```

### Test Plan

This AEP proposes comprehensive unit test coverage for all new code paths and additionally plans to introduce the following e2e tests:
- Recommender: Process a Kubernetes Deployment whose Pod template defines multiple containers, includes pod-level resources, and specifies container-level resources for at least one container. Verify that the generated VPA status stanza contains both container-level and pod-level recommendations.
- Updater: Process a Kubernetes Deployment whose Pod template defines multiple containers, includes pod-level resources, and specifies container-level resources for at least one container. Validate behavior in at least the `Recreate` and `InPlaceOrRecreate` modes.
- Admission Controller: Process a Kubernetes Deployment whose Pod template defines multiple containers, includes pod-level resources, and specifies container-level resources for at least one container. Test at least the `Recreate` VPA mode, ensuring that both the pod-level recommendation and the applicable container-level recommendation are applied.

The use case where a workload defines only container-level resource stanzas, or no resource stanzas at all, does not need to be covered by this AEP. Existing e2e tests already validate this behavior, and their outcomes are expected to remain unchanged after this proposal.

### Upgrade / Downgrade Strategy

#### Upgrade

Use a VPA release that includes this feature across all three components and pass `--feature-gates=PodLevelResources=true` to each component. Begin deploying new workloads that specify a pod-level resources stanza.

#### Downgrade

Downgrading VPA from a version that includes this feature should not immediately disrupt existing workloads. Pod-level resource specifications will remain in Pod specs, but VPA will revert to managing container-level resources only.

However, once VPA stops managing pod-level resources and continues updating only container-level resources, newly applied container-level requests or limits may violate the unmanaged pod-level constraints. For this reason, it is recommended to update the higher-level controller to remove pod-level resource specifications before performing the downgrade.

### Kubernetes Version Compatibility

`PodLevelResources` (i.e. the new VPA flag) is supported starting from Kubernetes v1.34, where the beta version of the [PodLevelResources](https://kubernetes.io/docs/reference/command-line-tools-reference/feature-gates/#PodLevelResources) feature gate is enabled by default. In this version, all VPA modes are supported except `InPlaceOrRecreate`.

To use the `InPlaceOrRecreate` VPA mode, Kubernetes v1.35 or later is required, and the alpha feature gate [InPlacePodLevelResourcesVerticalScaling](https://kubernetes.io/docs/reference/command-line-tools-reference/feature-gates/#InPlacePodLevelResourcesVerticalScaling) must be enabled (introduced in [In-Place Pod-Level Resources Resizing](https://github.com/kubernetes/enhancements/tree/master/keps/sig-node/5419-pod-level-resources-in-place-resize)). If this feature gate is disabled or the user runs an earlier Kubernetes version, in-place updates for pod-level resources will fail, and the updater will fall back to eviction.

## Implementation History

- 2025-09-29: initial version