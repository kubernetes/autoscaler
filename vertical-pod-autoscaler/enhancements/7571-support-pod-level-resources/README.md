# AEP-7571: Support Pod-Level Resources

<!-- toc -->
- [Summary](#summary)
    - [Goals](#goals)
- [Design Details](#design-details)
    - [Design Principles](#design-principles)
    - [Container-level resources](#container-level-resources)
    - [Pod-level resources](#pod-level-resources)
    - [Pod and Container-Level Resources](#pod-and-container-level-resources)
    - [Proposal](#proposal)
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
  initContainers:
    - image: tool1:latest
      name: tool1
      restartPolicy: Always
    - image: tool2:latest
      name: tool2
      restartPolicy: Always
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
```

The benefits and implementation details of pod-level `resources` are described in [KEP-2837](https://github.com/kubernetes/enhancements/blob/master/keps/sig-node/2837-pod-level-resource-spec/README.md). A related article is also available in the [Kubernetes documentation](https://kubernetes.io/docs/tasks/configure-pod-container/assign-pod-level-resources/).

Currently, before this AEP, VPA computes recommendations only at the container level, and those recommendations are applied exclusively at the container level. With the new pod-level resources specifications, VPA should be able to read from the pod-level `resources` stanza, calculate pod-level recommendations, and scale at the pod level when users define pod-level `resources`.

To address this, this KEP proposes extending the VPA object's `spec` and `status` fields, and introducing two new pod-level flags to set constraints directly at the pod level. For more details check the [Proposal](#proposal) section.

### Goals

* Add support for the pod-level resources stanza in VPA:
  * Read pod-level values
  * Calculate pod-level recommendations
  * Apply recommendations at the pod level
* Thoroughly document the new feature, focusing on areas that change default behaviors, in the [VPA documentation](https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler/docs).

## Design Details

### Design Principles

This section describes how VPA reacts based on where resources are defined (pod level, container level or both).

Before this KEP, the recommender computes recommendations only at the container level, and VPA applies changes only to container-level fields. With this proposal, the recommender also computes pod-level recommendations in addition to container-level ones. Pod-level recommendations are derived from per-container usage and recommendations, typically by aggregating container recommendations. Container-level policy still influences pod-level output: setting `mode: Off` in `spec.resourcePolicy.containerPolicies` excludes a container from recommendations, and `minAllowed`/`maxAllowed` bounds continue to apply.

This KEP extends the VPA CRD `spec.resourcePolicy` with a new `podPolicies` stanza that influences pod-level recommendations. The KEP also introduces two global pod-level flags `pod-recommendation-max-allowed-cpu` and `pod-recommendation-max-allowed-memory`. Details are covered in the [Proposal section](#proposal).

Today, the updater and admission controller update resources only at the container level. This proposal enables VPA components to update resources at the pod level as well.

**This KEP suggests that when a workload defines pod-level resources, VPA should manage those by default because pod-level resources offer benefits over container-only settings** - see the "Better resource utilization" section in [KEP-2837](https://github.com/kubernetes/enhancements/blob/master/keps/sig-node/2837-pod-level-resource-spec/README.md#better-resource-utilization) for details. 


Scenarios with no resources defined, or with both pod-level and container-level values present, require clear defaulting rules and are discussed in the options below. Note: community feedback should determine the default behavior. 

#### Container-level resources

For workloads that define only container-level resources, VPA should continue controlling resources at the container level, consistent with current behavior prior to this KEP. In other words, for a multi-container Pod without pod-level resources but with at least one container specifying resources, VPA should by default autoscale all containers.

#### Pod-level resources

For workloads that define only pod-level resources, VPA will control resources at the pod level. At the time of writing, [in-place pod-level resource resizing](https://github.com/kubernetes/enhancements/tree/master/keps/sig-node/5419-pod-level-resources-in-place-resize) is not available for pod-level fields, so applying pod-level recommendations requires evicting Pods. 

When [in-place pod-level resource resizing](https://github.com/kubernetes/enhancements/tree/master/keps/sig-node/5419-pod-level-resources-in-place-resize) becomes available, VPA should attempt to apply pod-level recommendations in place first and fall back to eviction if in-place updates fail, mirroring the current `InPlaceOrRecreate` behavior used for container-level updates.

#### Pod and Container-Level Resources

This part of the KEP covers workloads that define resources both at the pod level and for at least one container. To demonstrate multiple implementation options for how VPA should handle such workloads by default, consider the following manifest. It defines three containers:  
  * `ide` - the main workload container  
  * `tool1` and `tool2` - non-critical sidecar containers  


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
    - image: tool1:latest
      name: tool1
    - image: tool2:latest
      name: tool2
```

##### Option 1: VPA Controls Only Pod-Level Resources

With this option, VPA manages only the pod-level resources stanza. To follow this approach, the initially defined container-level resources for `ide` must be removed so that changes in usage are reflected only in pod-level recommendations.

**Pros**:
* VPA does not need to track which container-level resources were initially set.
* Straightforward for users: only the pod-level resources stanza is updated, while container-level stanzas are dropped.
* Enables shared headroom across containers in the same Pod. With container-only limits, a sidecar (`tool1` or `tool2`) or the main workload (`ide` container) hitting its own CPU limit could get throttled even if other containers in the Pod have idle CPU. Pod-level resources allow a container experiencing a spike to access idle resources from others, optimizing overall utilization.

**Cons**:
* This option currently requires evicting the Pod, because in-place pod-level resource resizing is not yet available, additionally container-level resources cannot be removed via the `resize` subresource (only changed to new values). 
* A downside of this approach is that the most important container (`ide`) may be recreated without container-level resources, leading to an `oom_score_adj` that matches other sidecars in the Pod, as a result the OOM killer may target all containers more evenly under node memory pressure. For details on how `oom_score_adj` is computed when pod-level resources are present, see the [KEP section](https://github.com/kubernetes/enhancements/blob/master/keps/sig-node/2837-pod-level-resource-spec/README.md#oom-score-adjustment) on OOM score adjustment.

##### Option 2: VPA controls pod-level resources and the initially set container-level resources

With this option, VPA controls pod-level resources and the container-level resources that were initially set. Using the earlier manifest as an example, VPA first applies recommendations at the pod level, then applies container-level recommendations for `ide`. Containers `tool1` and `tool2` are not updated by VPA, however their usage is still observed and contributes to the overall pod-level recommendation.

**Pros**:
- The primary container (`ide`) is less likely to be killed under memory pressure because sidecars (`tool1`, `tool2`) have higher `oom_score_adj` values, the OOM killer targets them first during node pressure evictions. See the [updated OOM killer behavior and formula](https://github.com/kubernetes/enhancements/blob/master/keps/sig-node/2837-pod-level-resource-spec/README.md#oom-killer-behavior) when pod-level resources are present.
- Sidecars, such as logging agents or mesh proxies (like `tool1` or `tool2`), that don't use container-level limits can borrow idle CPU from other containers in the pod when they experience a spike in usage. Pod-level resources allow a container experiencing a spike to access idle resources from others, optimizing overall utilization.

**Cons**:
- Applying both pod-level and container-level recommendations requires eviction because [in-place pod-level resource resizing](https://github.com/kubernetes/enhancements/tree/master/keps/sig-node/5419-pod-level-resources-in-place-resize) is not yet available.
- This option adds complexity: VPA must track which container-level resources are under its control by default and avoid mutating others.
- Existing VPA users may find the behavior surprising because VPA does not control all container-level resources stanzas - only those initially set - unless configured otherwise.

#### No resources stanza exists at the pod and container level

When a workload is created without any resources defined at either the pod or container level, there are two options:

##### Option 1: VPA controls only the container-level resources

This option mirrors current VPA behavior by managing only container-level resources, preserving benefits like in-place container-level resource resize. In this mode, pod-level recommendations are not computed and therefore not applied.

**Pros**:
- Familiar to existing users because it does not change current VPA behavior.

**Cons**:
- No cross-container resource sharing: in multi-container Pods, a container can hit its own limit and be throttled even if sibling containers are idle.

##### Option 2: VPA controls only the pod-level resources

With this option, VPA computes and applies only pod-level recommendations.

**Pros**:
- Enables shared headroom across containers in the same Pod, previously with container-only limits, a sidecar (e.g. `tool1` or `tool2` container) hitting its own CPU limit could throttle the Pod even if other containers had spare CPU. Pod-level resources allows one container experiencing a spike to access idle resources from others, optimizing overall utilization.
- Simple to adopt and remains straightforward for users if documented clearly in official documentation.

**Cons**:
- [In-place pod-level resource resize](https://github.com/kubernetes/enhancements/tree/master/keps/sig-node/5419-pod-level-resources-in-place-resize) is not yet available, so the updater would need to evict Pods for the admission controller to apply pod-level recommendations, future proposals can add in-place support when available.

## Proposal

- Add a new feature flag named `PodLevelResources`. Because this proposal introduces new code paths across all three VPA components, this flag will be added to each component.

- Extend the VPA object:
  1. Add a new `spec.resourcePolicy.podPolicies` stanza. This stanza is user-modifiable and allows setting constraints for pod-level recommendations:
     - `controlledResources`: Specifies which resource types are recommended (and possibly applied). Valid values are `cpu`, `memory`, or both. If not specified, both resource types are controlled by VPA.
     - `controlledValues`: Specifies which resource values are controlled. Valid values are `RequestsAndLimits` and `RequestsOnly`. The default is `RequestsAndLimits`.
     - `minAllowed`: Specifies the minimum resources that will be recommended for the Pod. The default is no minimum.
     - `maxAllowed`: Specifies the maximum resources that will be recommended for the Pod. The default is no maximum. To ensure per-container recommendations do not exceed the Pod's defined maximum, apply the formula to adjust the recommendations for containers proposed by @omerap12 (see [discussion](https://github.com/kubernetes/autoscaler/issues/7147#issuecomment-2515296024)). This field takes precedence over the global Pod maximum set by the new flags (see "Global Pod maximums").
  2. Add a new `status.recommendation.podRecommendation` stanza. This field is not user-modifiable, it is populated by the VPA recommender and stores the Pod-level recommendations. The updater and admission controller use this stanza to read Pod-level recommendations. The updater may evict Pods to apply the recommendation, the admission controller applies the recommendation when the Pod is recreated.

- Global Pod maximums: add two new recommender flags to constrain the maximum CPU and memory recommended at the Pod level. These flags are the Pod-level equivalents of `container-recommendation-max-allowed-cpu` and `container-recommendation-max-allowed-memory`, and will be named `pod-recommendation-max-allowed-cpu` and `pod-recommendation-max-allowed-memory`. Use the same enforcement formula referenced in the `maxAllowed` section. The VerticalPodAutoscaler-level maximum (that is, `maxAllowed`) takes precedence over the global maximum.

### Notes/Constraints/Caveats

- Pod-level resources support in VPA is opt-in and does not change the behavior of existing workload APIs (e.g. Deployments) unless explicitly enabled and the workload is recreated with a pod-level resources stanza.
- Because [in-place pod-level resource resize](https://github.com/kubernetes/enhancements/tree/master/keps/sig-node/5419-pod-level-resources-in-place-resize) is not yet available for Pod-level fields, the updater must evict Pods to apply pod-level recommendations.

### Validation

#### Static Validation

- The new fields under `spec.resourcePolicy.podPolicies` will be validated consistently, following the same approach we use for the existing `containerPolicies` stanza.
- This KEP proposes validating the new global flags, `--pod-recommendation-max-allowed-cpu` and `--pod-recommendation-max-allowed-memory`, as Kubernetes `Quantity` values using `resource.ParseQuantity`.

#### Dynamic Validation via Admission Controller

- When a pod-level resources stanza exists in the workload API, avoid using the `InPlaceOrRecreate` mode because it implies that in-place updates are possible - [in-place updates are not currently supported for pod-level resources](https://github.com/kubernetes/enhancements/tree/master/keps/sig-node/5419-pod-level-resources-in-place-resize). Therefore this mode should be rejected by the admission controller when a pod-level resources stanza is present.

### Test Plan

TODO

### Upgrade / Downgrade Strategy

### Upgrade

Use a VPA release that includes this feature across all three components and pass `--feature-gates=PodLevelResources=true` to each component. Begin deploying new workloads that specify a pod-level resources stanza.

### Downgrade

Downgrading VPA from a version that includes this feature should not disrupt existing workloads. Existing pod-level resource specifications remain in Pod specs, and VPA reverts to controlling container-level resources only.

### Examples

TODO

### Kubernetes version compatibility

This feature targets Kubernetes v1.34 or newer, with the beta version of [KEP-2837](https://github.com/kubernetes/enhancements/blob/master/keps/sig-node/2837-pod-level-resource-spec/README.md) (Pod-level resource specifications) enabled.

Kubernetes v1.32 is not recommended because the alpha implementation of Pod-level resources rejects container-level resource updates when Pod-level resources are set, see the [validation and defaulting rules for details](https://github.com/kubernetes/enhancements/blob/master/keps/sig-node/2837-pod-level-resource-spec/README.md#proposed-validation--defaulting-rules).

## Implementation History

- 2025-09-29: initial version