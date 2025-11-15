# AEP-7571: Support Pod-Level Resources

<!-- toc -->
- [Summary](#summary)
    - [Goals](#goals)
    - [Non-Goals](#non-goals)
- [Design Details](#design-details)
    - [Notes/Constraints/Caveats](#notesconstraintscaveats)
    - [Design Principles](#design-principles)
      - [Container-level resources](#container-level-resources)
      - [Pod-level resources](#pod-level-resources)
      - [Pod and Container-Level Resources](#pod-and-container-level-resources)
    - [Terminology and Clarifications](#terminology-and-clarifications)
    - [Proposal](#proposal)
      - [New Global Pod maximums](#new-global-pod-maximums)
      - [Recommender](#recommender)
      - [Updater](#updater)
      - [Admission Controller](#admission-controller)
        - [Pod Annotations](#pod-annotations)
        - [Pod-level Resource Limits](#pod-level-resource-limits)
      - [Pod-level Policies](#pod-level-policies)
      - [LimitRange object](#limitrange-object)
    - [Validation](#validation)
    - [Test Plan](#test-plan)
    - [Upgrade / Downgrade Strategy](#upgrade--downgrade-strategy)
    - [Feature Enablement and Rollback](#feature-enablement-and-rollback)
    - [Kubernetes version compatibility](#kubernetes-version-compatibility)
- [Implementation History](#implementation-history)
<!-- /toc -->

## Summary

Starting with Kubernetes version 1.34, it is now possible to specify CPU and memory `resources` for Pods at the pod level in addition to the existing container-level `resources` specifications. For example:

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

To address this, this AEP proposes extending the VPA object's `spec` and `status` fields, and introducing two new pod-level flags to set constraints directly at the pod level. For more details check the [Proposal](#proposal) section.

### Goals

* Add support for the pod-level resources stanza in VPA:
  * Read pod-level values
  * Calculate pod-level recommendations
  * Apply recommendations at the pod level
* Thoroughly document the new feature, focusing on areas that change default behaviors, in the [VPA documentation](https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler/docs).

### Non-Goals

* Since the latest VPA does not support `initContainers` ([the native way to use sidecar containers](https://kubernetes.io/blog/2023/08/25/native-sidecar-containers/)), this AEP does not aim to implement support for them. In other words, `initContainers` are simply ignored when calculating pod-level recommendations. Support for `initContainers` may be explored in a future proposal.

## Design Details

### Notes/Constraints/Caveats

- Pod-level resources support in VPA is opt-in and does not change the behavior of existing workload APIs (e.g. Deployments) unless explicitly enabled and the workload is recreated with a pod-level resources stanza.
- At the time of writing this AEP, the [In-Place Pod-Level Resources Resizing](https://github.com/kubernetes/enhancements/tree/master/keps/sig-node/5419-pod-level-resources-in-place-resize) is not available for pod level fields, so applying pod-level recommendations requires evicting Pods. When [In-Place Pod-Level Resources Resizing](https://github.com/kubernetes/enhancements/tree/master/keps/sig-node/5419-pod-level-resources-in-place-resize) becomes available, VPA should attempt to apply pod-level recommendations in place first and fall back to eviction if in-place updates fail, mirroring the current `InPlaceOrRecreate` mode behavior used for container-level updates. This new mechanism should be addressed in a separate proposal.

### Design Principles

This section describes how VPA reacts based on where resources are defined (pod level, container level or both).

Before this AEP, the recommender computes recommendations only at the container level, and the admission controller applies changes only to container-level fields. With this proposal, the recommender also computes Pod-level recommendations in addition to container-level ones. In later sections, it is shown that Pod-level recommendations must also be calculated by both the updater and the admission controller. Pod-level recommendations are derived from per-container usage and recommendations, typically by aggregating container recommendations. Container-level policy still influences pod-level output: setting `mode: Off` in `spec.resourcePolicy.containerPolicies` excludes a container from recommendations, and `minAllowed`/`maxAllowed` bounds continue to apply.

This AEP extends the VPA CRD `spec.resourcePolicy` with a new `podPolicies` stanza that influences pod-level recommendations. The AEP also introduces two global pod-level flags `pod-recommendation-max-allowed-cpu` and `pod-recommendation-max-allowed-memory`. Details are covered in the [Proposal section](#proposal).

Today, the updater makes decisions based on the container-level resources stanzas, and both the updater and the admission controller modify resources only at the container level. This proposal enables the updater to evict pods based on their pod-level resources stanzas and allows the admission controller to update resources at the pod level as well.

**This AEP suggests that when a workload defines pod-level resources, VPA should manage those by default because pod-level resources offer benefits over container-only settings** - see the "Better resource utilization" section in [KEP-2837](https://github.com/kubernetes/enhancements/blob/master/keps/sig-node/2837-pod-level-resource-spec/README.md#better-resource-utilization) for details.

Scenarios with no resources defined, or with both pod-level and container-level values present, require clear defaulting rules and are discussed in the options below. Note: community feedback should determine the default behavior. 

#### Container-level resources

For workloads that define only container-level resources, VPA should continue controlling resources at the container level, consistent with current behavior prior to this AEP (i.e. version 1.6.0). In other words, for a multi-container Pod without pod-level resources but with at least one container specifying resources, VPA should by default autoscale all containers.

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
* Straightforward for users: only the pod-level resources stanza is updated, while container-level stanzas are dropped.
* Enables shared headroom across containers in the same Pod. With container-only limits, a sidecar (`tool1` or `tool2`) or the main workload (`ide` container) hitting its own CPU limit could get throttled even if other containers in the Pod have idle CPU. Pod-level resources allow a container experiencing a spike to access idle resources from others, optimizing overall utilization.

**Cons**: 
* A downside of this approach is that the most important container (`ide`) may be recreated without container-level resources, leading to an `oom_score_adj` that matches other sidecars in the Pod, as a result the OOM killer may target all containers more evenly under node memory pressure. For details on how `oom_score_adj` is computed when pod-level resources are present, see the [KEP section](https://github.com/kubernetes/enhancements/blob/master/keps/sig-node/2837-pod-level-resource-spec/README.md#oom-score-adjustment) on OOM score adjustment.

##### [Selected] Option 2: VPA controls pod-level resources and the initially set container-level resources

With this option, VPA controls pod-level resources and the container-level resources that were initially set. The resources for containers `tool1` and `tool2` are not updated by VPA, however their usage is still observed and contributes to the overall pod-level recommendation.

**Pros**:
- The primary container (`ide`) is less likely to be killed under memory pressure because sidecars (`tool1`, `tool2`) have higher `oom_score_adj` values, the OOM killer targets them first during node pressure evictions. See the [updated OOM killer behavior and formula](https://github.com/kubernetes/enhancements/blob/master/keps/sig-node/2837-pod-level-resource-spec/README.md#oom-killer-behavior) when pod-level resources are present.
- Sidecars, such as logging agents or mesh proxies (like `tool1` or `tool2`), that don't use container-level limits can borrow idle CPU from other containers in the pod when they experience a spike in usage. Pod-level resources allow a container experiencing a spike to access idle resources from others, optimizing overall utilization.

**Cons**:
- VPA must track which container-level resources are under its control by default and avoid mutating others.
- Existing VPA users may find the behavior surprising because VPA does not control all container-level resources stanzas - only those initially set - unless configured otherwise.

#### No resources stanza exists at the pod and container level

When a workload is created without any resources defined at either the pod or container level, there are two options:

##### Option 1: VPA controls only the container-level resources

This option mirrors current VPA behavior by managing only container-level resources, preserving benefits like in-place container-level resource resize. In this mode, pod-level recommendations are not computed and therefore not applied.

**Pros**:
- Familiar to existing users because it does not change current VPA behavior.

**Cons**:
- No cross-container resource sharing: in multi-container Pods, a container can hit its own limit and be throttled even if sibling containers are idle.

##### [Selected] Option 2: VPA controls only the pod-level resources

With this option, VPA computes and applies only pod-level recommendations.

**Pros**:
- Enables shared headroom across containers in the same Pod, previously with container-only limits, a sidecar (e.g. `tool1` or `tool2` container) hitting its own CPU limit could throttle the Pod even if other containers had spare CPU. Pod-level resources allows one container experiencing a spike to access idle resources from others, optimizing overall utilization.
- Simple to adopt and remains straightforward for users if documented clearly in official documentation.

### Terminology and Clarifications

The following terms and definitions are used throughout this AEP:

* **Pods with Pod-level resources** - The user defines a resources stanza at the Pod level and may optionally define container-level resources for some containers. The VPA manages only the Pod-level resources stanza, and container-level resources only for containers that were initially defined within the workload API.
* **Pods with Container-level resources** - The user defines resources stanzas within individual containers. The VPA manages only container-level resources, and by default manages all containers in this type of Pod.
* **Pods without any resources stanza** - These Pods are treated as Pod-level by default, and the VPA manages the Pod-level resources stanza. In this AEP, the term "Pods with Pod-level resources" also includes "Pods without any resources stanza".
* **Pod LimitRange object** - Refers to a `LimitRange` object with `type=Pod`.
* **Container LimitRange object** - Refers to a `LimitRange` object with `type=Container`.
* **Latest VPA version** - In this AEP, the term "latest VPA version" refers to version 1.6.0.

### Proposal

- Add a new feature flag named `PodLevelResources`. Because this proposal introduces new code paths across all three VPA components, this flag will be added to each component.

- Extend the VPA object:
  1. Add a new `spec.resourcePolicy.podPolicies` stanza. This stanza is user-modifiable and allows setting constraints for pod-level recommendations:
     - `controlledResources`: Specifies which resource types are recommended (and possibly applied). Valid values are `cpu`, `memory`, or both. If not specified, both resource types are controlled by VPA.
     - `controlledValues`: Specifies which resource values are controlled. Valid values are `RequestsAndLimits` and `RequestsOnly`. The default is `RequestsAndLimits`.
     - `minAllowed`: Specifies the minimum amount of resources that can be recommended for the Pod. The default is no minimum. This Pod-level field must be enforced by all VPA components: the recommender, the updater, and the admission controller. If the sum of per-container recommendations does not meet the Pod-level minAllowed value, each container's recommendation should be increased proportionally, following the formula proposed by @omerap12 (see [discussion](https://github.com/kubernetes/autoscaler/issues/7147#issuecomment-2515296024)).
     - `maxAllowed`: Specifies the maximum amount of resources that can be recommended for the Pod. The default is no maximum. To ensure that per-container recommendations do not exceed the Pod's defined maximum, apply the adjustment formula proposed by @omerap12. This field takes precedence over the global Pod maximum set by the new flags (see [Global Pod maximums](#new-global-pod-maximums) section). Like `minAllowed`, this Pod-level field must be enforced by all VPA components: the recommender, the updater, and the admission controller.
  2. Add a new `status.recommendation.podRecommendation` stanza. This field is not user modifiable, it is populated by the VPA Recommender and stores Pod-level recommendations, including Pod-level `target`, `lowerBound`, and `upperBound` for CPU and memory. The consumers of these new fields include not only end users, but also the updater and the admission controller. Here is the newly proposed `status` field, including the new `podRecommendation` section:

```yaml
status:
  conditions:
  - lastTransitionTime: "2025-11-14T10:53:08Z"
    status: "True"
    type: RecommendationProvided
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

#### New Global Pod maximums

Add two new recommender flags to cap Pod-level recommendations:
- `--pod-recommendation-max-allowed-cpu`: Sets the global maximum CPU the recommender may recommend at Pod scope.
- `--pod-recommendation-max-allowed-memory`: Sets the global maximum memory the recommender may recommend at Pod scope.

Per-container recommendations must be adjusted so that their total does not exceed these global constraints, using the same enforcement formula described in the `minAllowed` and `maxAllowed` sections. The VerticalPodAutoscaler maximum (Pod-level or container-level `maxAllowed`) takes precedence over these global Pod maximums when both are present.

#### Recommender

This AEP proposes extending the recommender so that it can compute pod-level recommendations for Pods that define pod-level resources. The proposed calculation method is to sum all container-level targets for CPU and Memory, as well as sum all container-level UpperBound and LowerBound values.

In other words, if the user defines a Deployment with pod-level resources and three containers, the resulting pod-level target is the sum of the target values from container 1, container 2, and container 3. For example, the formula for the pod-level CPU target is:

```
Pod-level CPU Target = c1 CPU target
                     + c2 CPU target
                     + c3 CPU target
```

The same approach applies to memory targets.

If a container is not managed by VPA - i.e. its `containerPolicies` mode is set to "Off" - then the recommender does not compute recommendations for that container (target, UpperBound, or LowerBound). As a result that container is excluded from the pod-level recommendation calculation.

The computed pod-level recommendations are consumed by both the admission controller and the updater:
* The admission controller uses them to generate pod-level patches.
* The updater uses them to make eviction-related decisions.

#### Updater

With the introduction of pod-level recommendations, the updater must be able to make eviction decisions for Pods with Pod-level resources that are managed by the VPA. In the latest VPA version, the updater does not consider evicting a Pod if container-level recommendations are missing (i.e. the recommender did not compute them and they are not present in the VPA object). The same behavior should apply to Pods with Pod-level resources. Therefore, if the corresponding pod-level recommendations do not exist for a Pod, the updater skips that Pod and proceeds to the next one.

This AEP also proposes using the same eviction logic that the current VPA implements in the [AddPod](https://github.com/kubernetes/autoscaler/blob/7fe519cb69ebd51d91d4fbb5bba2d2b7d8c9946f/vertical-pod-autoscaler/pkg/updater/priority/update_priority_calculator.go#L84) method, as defined in the latest implementation:
* **The Pod-level request is outside the Pod-level recommended range** (as stored in the VPA object). For example, if only memory is managed and the current Pod-level request is 200 Mi while the calculated Pod-level LowerBound is 50 Mi and the UpperBound is 100 Mi, the updater should evict the Pod so it can be recreated with the new Pod-level target.
* **A managed container's current request is outside its container-level recommended range**. A managed container is one whose resource stanza is defined in the workload API and is therefore under VPA control.
* The Pod has been running for at least 12 hours, and the resource difference is >= `MinChangePriority`.
* A VPA-scaled container experienced an OOM event within the `evictAfterOOMThreshold` window.

When the updater decides to evict a Pod with Pod-level resources, the admission controller is responsible for generating the required patches. This includes both Pod-level resource patches and any necessary container-level patches.

#### Admission Controller

The admission controller needs to be extended so that it can compute Pod-level resource patches for Pods with Pod-level resources. To clarify how the admission controller should calculate these patches, the following describes the required logic.

1. For each container in the Pod, determine the applicable recommendation based on the following rules:

Table Legend:
* CR = Container-level `Requests` as specified in the Pod Spec
* CL = Container-level `Limits` as specified in the Pod Spec
* CMIN = Container-level `minAllowed` from the VPA object's `containerPolicies` stanza
* CMAX = Container-level `maxAllowed` from the VPA object's `containerPolicies` stanza
* CT = Container-level `target` from the VPA object's `containerRecommendations` stanza
* recommendation = The resulting value determined based on the user's configuration in the VPA policy, the workload API resources stanza, or the calculated container-level recommendations.

| Sr.No. | requests | limit | minAllowed | maxAllowed | target |  recommendation   |
| :----: | :------: | :---: | :--------: | :--------: | :----: | :---------------: |
|   1    |  unset   | unset |   unset    |   unset    | unset  |      0 `[a]`      |
|   2    |    CR    | unset |   unset    |   unset    | unset  |        CR         |
|   3    |  unset   |  CL   |   unset    |   unset    | unset  |      CL`[b]`      |
|   4    |    CR    |  CL   |   unset    |   unset    | unset  |        CR         |
|   5    |  unset   | unset |   unset    |   unset    |   CT   |        CT         |
|   6    |  unset   | unset |    CMIN    |   unset    | unset  |     CMIN `[c]`    |
|   7    |    CR    | unset |   unset    |   unset    |   CT   |      CT `[d]`     |
|   8    |  unset   | unset |   unset    |    CMAX    |   CT   | CMAX if CT > CMAX |
|   9    |  unset   | unset |    CMIN    |   unset    |   CT   | CMIN if CT < CMIN |
|   10   |    CR    | unset |   unset    |    CMAX    | unset  | CMAX if CR > CMAX `[c]` |
|   11   |    CR    | unset |    CMIN    |   unset    | unset  | CMIN if CR < CMIN `[c]` |

* Annotation `[a]` - In this case the admission controller cannot determine a recommendation for the container. As a result the container is excluded from the Pod-level recommendation calculation.
* Annotation `[b]` - For container-level resources, when a request is not specified but a limit is set, Kubernetes defaults the request to equal the limit (see [upstream defaulting behavior](https://github.com/kubernetes/kubernetes/blob/005f184ab631e52195ed6d129969ff3914d51c98/pkg/apis/core/v1/defaults.go#L167-L228)). This AEP proposes that such limits should be considered when calculating Pod-level recommendations.
* Annotation `[c]` - This AEP proposes updating the admission controller to ensure that the conditions referenced by this annotation are enforced.
* Annotation `[d]` - The container's target value from the VPA object takes precedence over any container-level requests specified in the workload API.

2. Calculate the Pod-level recommendation using the values from the recommendation column in the previous table, separately for CPU and memory. For example, the CPU calculation can be expressed as:

```
Pod-level CPU recommendation = container[1].cpu.recommendation
                             + container[2].cpu.recommendation
                             + ...
                             + container[N].cpu.recommendation
```

In summary, the admission controller must be extended to compute Pod-level resource patches in addition to container-level patches. For example, when a user creates a workload using the following manifest:

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
    requests:
      memory: "512Mi"
      cpu: "2"
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

After implementing this AEP, we expect the admission controller to generate Pod-level patches for the Deployment above - based on the formula specified in the previous step - while generating container-level patches only for the container named `ide`.

For workload API objects where the user does not define either Pod-level or container-level resources stanzas, the admission controller is expected to generate only Pod-level patches.

##### Pod Annotations

New Pod-level annotations should be defined and applied as needed:
* Annotate the Pod when Pod-level requests and limits are managed. Explicitly indicate when a Pod-level limit is not set.
* Annotate the Pod when a Pod-level resource request (CPU or memory) is capped to the Pod-level new `minAllowed` or `maxAllowed`.
* Annotate the Pod when a Pod-level resource limit (CPU or memory) is capped to fit within the `Min` or `Max` defined in a Pod LimitRange object.

##### Pod-level Resource Limits

The Pod-level request-to-limit ratio must be maintained according to the original ratio defined in the workload API's Pod-level resources stanza.

If Pod-level limits are omitted, the new Pod-level limits cannot be determined. In this case, Pods managed by the VPA are recreated without Pod-level limits.

### Pod-level Policies

If the Pod-level policies (i.e. `podPolicies.minAllowed` and `podPolicies.maxAllowed`) are violated, the recommender and the admission controller should proportionally adjust the Pod-level recommendations:
  * If the Pod-level recommendation is less than `podPolicies.minAllowed`, use the minAllowed value as the new Pod-level recommendation.
  * If the Pod-level recommendation is greater than `podPolicies.maxAllowed`, use the maxAllowed value as the new Pod-level recommendation.

The updater skips Pods for which no Pod-level recommendations exist in the corresponding VPA object. Consequently the updater may skip validating `podPolicies.minAllowed` and `podPolicies.maxAllowed` for such Pods.

### LimitRange object

The admission controller and the updater should respect constraints defined by LimitRange objects in the cluster.

#### Container LimitRange object

This AEP recommends that Pods with Pod-level resources should not be deployed in namespaces that contain a Container LimitRange object. The built-in LimitRanger admission controller automatically sets default container-level requests and limits for these Pods, which is undesirable when Pod-level resources are in use.

This behavior is also described in the [Pod-level Resource Spec KEP](https://github.com/kubernetes/enhancements/blob/master/keps/sig-node/2837-pod-level-resource-spec/README.md#admission-controller), which suggests placing Pods in namespaces without any Container LimitRange objects. Therefore this AEP proposes that creation of Pods with Pod-level resources in namespaces containing Container LimitRange objects should be rejected. This validation is further detailed in the [Validation section](#dynamic-validation).

In summary Pods with Pod-level resources should not be validated against Container LimitRange objects.

#### Pod LimitRange object

For the admission controller, the most relevant recommendation type is the Target, since the controller calculates requests and limits. For the updater, the other Pod-level recommendation types - LowerBound and UpperBound - are used to determine which Pods should be evicted.

This AEP proposes applying the same logic for Pod-level recommendations as defined in the [capProportionallyToPodLimitRange](https://github.com/kubernetes/autoscaler/blob/722902dce0d3fba2a1eb9904a4d91e448496a7f8/vertical-pod-autoscaler/pkg/utils/vpa/capping.go#L514) method. In other words, the Pod-level Target, UpperBound, and LowerBound should be increased or decreased proportionally to comply with Pod LimitRange constraints.

For example:
* Pod-level `LowerBound` = 100 Mi
* Pod-level `Target` = 120 Mi
* Pod-level `UpperBound` = 150 Mi
* Pod LimitRange `min` = 200 Mi

In this case, the Pod-level `LowerBound`, `Target`, and `UpperBound` should all be increased to 200 Mi. Similarly if a maximum field in the Pod LimitRange is violated, the amounts should be proportionally decreased.

For Pods with Pod-level resources that also define container-level resources stanzas, managed container-level recommendations should be checked against Pod LimitRange objects in the same way as in the latest VPA using `capProportionallyToPodLimitRange` method.

For example:
* Pod with two containers: c1 and c2
* Pod-level resources stanza is present
* Container-level stanza is present only for c1
* Calculated container-level Target: c1 = 120 Mi, c2 = 30 Mi
* Pod LimitRange min = 200 Mi

To respect the Pod LimitRange, container-level targets are increased proportionally:
* c1 -> 160 Mi (calculated as`(120/150)×200`)
* c2 -> 40 Mi (calculated as `(30/150)×200`)

### Validation

#### Static Validation

- The new fields under `spec.resourcePolicy.podPolicies` will be validated consistently, following the same approach we use for the existing `containerPolicies` stanza.
- This AEP proposes validating the new global flags, `--pod-recommendation-max-allowed-cpu` and `--pod-recommendation-max-allowed-memory`, as Kubernetes `Quantity` values using `resource.ParseQuantity`.

#### Dynamic Validation

Admission Controller:

* When a pod-level resources stanza exists in the workload API, avoid using the `InPlaceOrRecreate` mode because it implies that in-place updates are possible - [in-place updates are not currently supported for pod-level resources](https://github.com/kubernetes/enhancements/tree/master/keps/sig-node/5419-pod-level-resources-in-place-resize). Therefore this mode should be rejected by the admission controller when a pod-level resources stanza is present.
* The pod-level `minAllowed` must be equal to or greater than the sum of the container-level `minAllowed` values. For example, if an end user has a pod with three containers - C1, C2, and C3 - and defines `minAllowed` for C1 and C2, then the sum of these values must be equal to or lower than the pod-level `minAllowed` specified in the VPA object. The same rule applies to the pod-level `maxAllowed` value as well.
* Creation of Pods with Pod-level resources should be rejected in namespaces that contain Container LimitRange objects. See the [LimitRange](#limitrange-object) section for more details.

### Test Plan

This AEP aims to propose thoughtful unit test coverage for new code paths and additionally intends to develop the following e2e tests:
* Enable VPA for a workload that doesn't define any pod- or container-level resources stanzas.
The expected outcome is that pod-level recommendations are calculated and applied only at the pod level.
* Enable VPA for a workload that defines pod-level and at least one container-level resources stanza.
The expected outcome is that both the pod-level stanza and the initially defined container-level stanzas are updated.
* Test use cases where the user adds or removes a container from a workload managed by VPA.

The use case in which the workload contains only container-level resources stanzas doesn't need to be tested in this AEP, as an existing e2e test already covers it. The outcome of that test should be the same as before this AEP.

### Upgrade / Downgrade Strategy

#### Upgrade

Use a VPA release that includes this feature across all three components and pass `--feature-gates=PodLevelResources=true` to each component. Begin deploying new workloads that specify a pod-level resources stanza.

#### Downgrade

Downgrading VPA from a version that includes this feature should not disrupt existing workloads. Existing pod-level resource specifications remain in Pod specs, and VPA reverts to controlling container-level resources only.

### Kubernetes version compatibility

This feature targets Kubernetes v1.34 or newer, with the beta version of [KEP-2837](https://github.com/kubernetes/enhancements/blob/master/keps/sig-node/2837-pod-level-resource-spec/README.md) (Pod-level resource specifications) enabled.

Kubernetes v1.32 is not recommended because the alpha implementation of Pod-level resources rejects container-level resource updates when Pod-level resources are set, see the [validation and defaulting rules for details](https://github.com/kubernetes/enhancements/blob/master/keps/sig-node/2837-pod-level-resource-spec/README.md#proposed-validation--defaulting-rules).

## Implementation History

- 2025-09-29: initial version