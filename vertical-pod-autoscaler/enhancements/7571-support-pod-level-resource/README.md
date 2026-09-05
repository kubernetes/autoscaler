<!--
**Note:** When your AEP is complete, all of these comment blocks should be
removed.

AEPs (Autoscaler Enhancement Proposals) are a lightweight version of Kubernetes
KEPs, scoped to the Vertical Pod Autoscaler subproject. Use this template as
the skeleton for new AEPs so reviewers see a consistent structure.

To get started:

- [ ] Open a tracking issue in kubernetes/autoscaler describing the problem.
- [ ] Copy this directory to `NNNN-short-descriptive-title`, where `NNNN` is
      the issue number.
- [ ] Fill in `Summary` and `Motivation` first — these are enough to start a
      design discussion.
- [ ] Open a PR for the new AEP and iterate. Merging an AEP does not mean it
      is approved or complete; aim for tightly-scoped PRs per topic.
- [ ] Fill in the remaining sections as the design firms up.

One AEP corresponds to one "feature" or "enhancement" for its whole lifecycle.
If major changes emerge after implementation, edit the AEP rather than opening
a new one.
-->

# Pod-level resources support in VPA - Phase 1

<!--
Keep the title short and descriptive. It is used in the TOC, commit messages,
and PR titles.
-->

<!-- toc -->
- [Summary](#summary)
- [Motivation](#motivation)
  - [Goals](#goals)
  - [Non-Goals](#non-goals)
- [Proposal](#proposal)
- [Design Details](#design-details)
  - [API Changes](#api-changes)
  - [Recommender](#recommender)
    - [Calculating Pod-Level Recommendations](#calculating-pod-level-recommendations)
      - [Leveraging Container-Level Histograms](#leveraging-container-level-histograms)
      - [Using New Pod-Level Histograms](#using-new-pod-level-histograms)
    - [Maintaining checkpoints](#maintaining-checkpoints)
    - [New Global Pod maximums](#new-global-pod-maximums)
    - [New Flags for Configuring Pod Level Percentiles](#new-flags-for-configuring-pod-level-percentiles)
  - [Admission Controller](#admission-controller)
    - [Dynamic Validation](#dynamic-validation)
  - [Updater](#updater)
  - [capping.go](#cappinggo)
  - [Container LimitRange Objects](#container-limitrange-objects)
  - [Pod Annotations](#pod-annotations)
  - [Handling Container Addition and Removal](#handling-container-addition-and-removal)
  - [VPAPodLevelResources Feature Gate](#vpapodlevelresources-feature-gate)
    - [Recommender](#recommender-1)
    - [Updater](#updater-1)
    - [Admission Controller](#admission-controller-1)
  - [Test Plan](#test-plan)
    - [Recommender](#recommender-2)
    - [Admission Controller](#admission-controller-2)
    - [Updater](#updater-2)
  - [Feature Enablement and Rollback](#feature-enablement-and-rollback)
    - [Feature Enablement](#feature-enablement)
    - [Rollback](#rollback)
  - [Graduation Criteria](#graduation-criteria)
    - [Phase 1: Alpha (target VPA 1.9.0 version)](#phase-1-alpha-target-vpa-190-version)
    - [Phase 2: Beta (target VPA 1.10.0 version)](#phase-2-beta-target-vpa-1100-version)
    - [GA (stable)](#ga-stable)
  - [Version Skew](#version-skew)
  - [Kubernetes Version Compatibility](#kubernetes-version-compatibility)
  - [VPA Object Examples](#vpa-object-examples)
    - [Example 1 - Observe Pod Level recommendations](#example-1---observe-pod-level-recommendations)
    - [Example 2 - Incorrect VPA Object](#example-2---incorrect-vpa-object)
    - [Example 3](#example-3)
- [Implementation History](#implementation-history)
<!-- /toc -->

## Summary

<!--
A paragraph or two that captures what this AEP is about and why it matters.
Write this section so that someone unfamiliar with the VPA internals can read
it and understand the shape of the proposal.
-->

Starting with Kubernetes v1.34 (enabled by default, feature gate PodLevelResources), it is possible to specify CPU and memory resources for Pods at the pod-level, in addition to the existing container-level resources specifications. For example:


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

For more details about the pod-level resource stanza, refer to these documents:
* [KEP-2837](https://github.com/kubernetes/enhancements/blob/master/keps/sig-node/2837-pod-level-resource-spec/README.md)
* [KEP-5419](https://github.com/kubernetes/enhancements/blob/master/keps/sig-node/5419-pod-level-resources-in-place-resize/README.md)

This AEP proposes the implementation details for Phase 1 of pod-level resource support in VPA. The feature is gated by a feature gate. When the feature gate is enabled and a user opts in by creating or updating a VPA object, VPA computes pod-level resource requests and limits and applies them either through in-place Pod updates or during Pod admission.

In Phase 1, this feature cannot be used together with the existing VPA behavior of managing container-level resource stanzas. Support for simultaneously managing both container-level and pod-level resource stanzas will be investigated after the completion of Phase 1.

## Motivation

<!--
Explain why this change is worth doing. Link to issues, user reports, or
previous discussions that show the problem is real. Reviewers will weigh the
cost of the change against the motivation described here, so be concrete.
-->

The goal of this proposal is to autoscale pod-level resource requests and limits. For workloads whose containers do not reach peak resource usage simultaneously, pod-level resources can improve resource utilization compared to container-level resource requests and limits.

### Goals

<!--
Bullet list of what this AEP is trying to achieve. Keep each goal testable —
something a reviewer could point at later to decide whether the AEP succeeded.
-->

* The recommender should be able to calculate pod-level recommendations when the user opts in.  
* The updater should be able to make eviction or in-place update decisions based on the pod-level resources stanza and pod-level recommendations.  
* The admission controller should be able to apply pod-level resources at admission.
* The pod-level request-to-limit ratio is determined from the pod spec in the same way as in the latest VPA release at the container-level.
* Support all existing VPA modes

### Non-Goals

<!--
What is explicitly out of scope. Listing non-goals is often more valuable than
listing goals — it keeps the discussion focused and prevents scope creep during
review.
-->

* Managing container-level resource stanzas when pod-level autoscaling is enabled
* Since the latest VPA does not support initContainers ([the native way to use sidecar containers](https://kubernetes.io/blog/2023/08/25/native-sidecar-containers/)), this AEP does not aim to implement support for them. In other words, initContainers are simply ignored when calculating pod-level recommendations. Support for initContainers is outside the scope of this proposal.
* Support for the startupBoost feature at the pod-level is outside the scope of this proposal.

## Proposal

<!--
Describe the proposed change at a high level. This is the "what", not the
"how" — implementation details belong in Design Details below. A reviewer
should be able to read this section and understand the user-visible behavior
without reading any code.
-->

This proposal extends VPA to autoscale (that is, manage) pod-level resource stanzas. This requires updating the Recommender to compute pod-level recommendations (LowerBound, Target, UpperBound, and UncappedTarget) according to the configuration specified in the new `spec.resourcePolicy.podPolicy` stanza. It also requires updating the Updater and Admission Controller to apply these recommendations.

## Design Details

<!--
The "how". Include enough detail that a reader can evaluate whether the
approach is sound. API types, flag names, component interactions, and any
non-obvious behavior belong here. Code snippets and YAML examples are welcome
when they clarify intent.
-->

To enable a safe and incremental rollout of pod-level autoscaling feature, this proposal introduces a new VPA feature gate named `VPAPodLevelResources` for each VPA component: the recommender, updater, and admission controller. Note for reviewers: The `PodLevelResources` feature gate name is already used in Kubernetes (k/k), which is why the `VPA` prefix was added.

### API Changes

<!--
If this AEP adds or modifies fields in the VPA API (`autoscaling.k8s.io/v1`),
describe the new types, their validation rules, and default behavior. Include
the Go struct definitions when possible. If there are no API changes, remove
this subsection.
-->

* Add a new `spec.resourcePolicy.podPolicy` stanza with the following fields. While most pod-level fields function identically to their container-level counterparts, the key difference is that the fields listed below apply directly to pod-level resource recommendations:
  * `mode`: This new field is similar to the one that already exists at the container-level, with two possible values: 
      * `Off`: The default value is "Off", meaning that the autoscaler does not compute pod-level recommendations and ignores management of the pod-level resource stanzas for the targeted pods.
      * `Auto`: It indicates whether the autoscaler is enabled at the pod-level.
  * `controlledResources`: Specifies which resource types are recommended (and possibly applied). Valid values are "cpu", "memory", or both. If not specified, both resource types are controlled by VPA.
  * `controlledValues`: Specifies which resource values are controlled. Valid values are "RequestsAndLimits" and "RequestsOnly". The default is "RequestsAndLimits".
  * `minAllowed`: Specifies the minimum amount of resources that the autoscaler recommends at the pod-level. The default is no minimum.
  * `maxAllowed`: Specifies the maximum amount of resources that the autoscaler recommends at the pod-level. The default is no maximum.
  * `oomBumpUpRatio` is the ratio by which the recommender increases memory when a pod experiences an OOM event, i.e. when the combined memory usage of all containers within a pod exceeds the pod-level memory limit. Optional field, therefore there is no default value. This new field overrides the existing recommender flag --oom-bump-up-ratio.
  * `oomMinBumpUp` is the minimum amount by which the recommender increases memory when a pod experiences an OOM event, i.e. when the combined memory usage of all containers within a pod exceeds the pod-level memory limit. Optional field, therefore there is no default value. This new field overrides the existing recommender flag --oom-min-bump-up-bytes.

Extend the VPA object with a new `status.recommendation.podRecommendations` field. The field is populated by the VPA Recommender and stores Pod-level recommendations, including pod-level "Target", "LowerBound", "UpperBound" and "UncappedTarget" for CPU and memory.

Here are the proposed changes to the VPA API that would allow us to avoid an API bump. However, this would come at the cost of breaking struct type naming consistency. For example, the types of the two fields in `RecommendedPodResources` would no longer follow a consistent naming convention:

```go
// RecommendedPodResources contains the resource recommendations computed by
// the autoscaler. Resource recommendations can exist at either the container
// level or the pod level.
type RecommendedPodResources struct {
  // Resources recommended by the autoscaler for each container.
  // +optional
  ContainerRecommendations []RecommendedContainerResources `json:"containerRecommendations,omitempty"`
  // Resources recommended by the autoscaler at the pod level.
  // +optional
  PodRecommendations *RecommendedPodLevelResources `json:"podRecommendations,omitempty"` // (!NEW)
}

// (!NEW) RecommendedPodLevelResources is the recommendation computed by the autoscaler for Pod Level Resources.
type RecommendedPodLevelResources struct {
  // Pod-level recommended amount of resources. Observes PodResourcePolicies.
  Target corev1.ResourceList `json:"target"`
  // Pod-level minimum recommended amount of resources. Observes PodResourcePolicies.
  // This amount is not guaranteed to be sufficient for the application to operate in a stable way, however
  // running with less resources is likely to have significant impact on performance/availability.
  // +optional
  LowerBound corev1.ResourceList `json:"lowerBound,omitempty"`
  // Pod-level maximum recommended amount of resources. Observes PodResourcePolicies.
  // Any resources allocated beyond this value are likely wasted. This value may be larger than the maximum
  // amount of application is actually capable of consuming.
  // +optional
  UpperBound corev1.ResourceList `json:"upperBound,omitempty"`
  // The most recent recommended resources target computed by the autoscaler
  // at the pod level, based only on actual resource usage, not taking
  // into account the PodResourcePolicies.
  // May differ from the Recommendation if the actual resource usage causes
  // the target to violate the PodResourcePolicies (lower than MinAllowed
  // or higher that MaxAllowed).
  // Used only as status indication, will not affect actual resource assignment.
  // +optional
  UncappedTarget corev1.ResourceList `json:"uncappedTarget,omitempty"`
}

// PodResourcePolicy controls how autoscaler computes the recommended resources
// for containers belonging to the pod. There can be at most one entry for every
// named container and optionally a single wildcard entry with `containerName` = '*',
// which handles all containers that don't have individual policies.
// Additionally, it controls how pod-level recommendations are calculated.
type PodResourcePolicy struct {
  // Per-container resource policies.
  // +optional
  // +patchMergeKey=containerName
  // +patchStrategy=merge
  ContainerPolicies []ContainerResourcePolicy `json:"containerPolicies,omitempty" patchStrategy:"merge" patchMergeKey:"containerName"`
  // Pod-level resource policy. It controls how Pod Level Resources are calculated by the autoscaler.
  // +optional
  PodPolicy *PodLevelResourcePolicy `json:"podPolicy,omitempty"` // (!NEW)
}

// (!NEW) PodLevelResourcePolicy controls how the autoscaler computes recommended resources at the pod level.
// In other words, it controls how Pod Level Resources are calculated by the autoscaler.
type PodLevelResourcePolicy struct {
  // Indicates whether the autoscaler is enabled at the pod level.
  // The default is "Off", so the autoscaler does not compute recommendations at the pod level.
  // +optional
  Mode *PodScalingMode `json:"mode,omitempty"`
  // Specifies the minimum amount of resources that the autoscaler recommends at the pod level.
  // The default is no minimum.
  // +optional
  MinAllowed corev1.ResourceList `json:"minAllowed,omitempty"`
  // Specifies the maximum amount of resources that the autoscaler recommends at the pod level.
  // The default is no maximum.
  // +optional
  MaxAllowed corev1.ResourceList `json:"maxAllowed,omitempty"`
  // Specifies the resource types (CPU, memory, or both) that VPA computes at the pod level and may apply.
  // If not specified, the default of [ResourceCPU, ResourceMemory] will be used.
  // +patchStrategy=merge
  ControlledResources *[]corev1.ResourceName `json:"controlledResources,omitempty" patchStrategy:"merge"`
  // Specifies which resource values VPA controls at the pod level.
  // The default is "RequestsAndLimits".
  // +optional
  ControlledValues *ContainerControlledValues `json:"controlledValues,omitempty"`
  // oomBumpUpRatio is the ratio by which memory is increased when an OOM is detected.
  // +optional
  OOMBumpUpRatio *resource.Quantity `json:"oomBumpUpRatio,omitempty"`
  // oomMinBumpUp is the minimum memory increase applied when an OOM is detected.
  // +optional
  OOMMinBumpUp *resource.Quantity `json:"oomMinBumpUp,omitempty"`
}
```

### Recommender

#### Calculating Pod-Level Recommendations

The goal of this section is to discuss the different methods considered for calculating pod-level recommendations.

##### Leveraging Container-Level Histograms

This section discusses whether the new pod-level autoscaling feature could leverage the existing container-level histograms. By default, these histograms are populated with usage samples for all regular containers in the cluster unless the `memory-saver` flag is enabled.

The following approaches do not produce valid pod-level recommendations:

* **Summing container-level recommendations**: Summing values such as the lower bounds of container-level recommendations does not produce a valid pod-level recommendation because percentiles are not additive.

* **Merging container-level histograms**: Adding the weights of buckets at the same index across multiple container-level histograms and then calculating a percentile does not produce a valid pod-level recommendation. To sum the bucket weights, we could use the [MergeContainerState](https://github.com/kubernetes/autoscaler/blob/6a616ea0c5ea0cb6111240073a9273b3467c064e/vertical-pod-autoscaler/pkg/recommender/model/aggregate_container_state.go#L197) method. However, this approach is invalid because it merges independent container-level observations, whereas the result should reflect the pod's resource usage as a whole.

##### Using New Pod-Level Histograms

This proposal recommends using new pod-level histograms as the source for calculating pod-level recommendations.

The recommender would continue to collect container usage samples as before. The most significant change within the recommender would be the introduction of two new pod-level decaying histograms - one for each resource type - for every group of pods that shares the same namespace and labels.

For example, if a user defines a Kubernetes Deployment with two replicas, each containing three regular containers, all collected samples in this case (three samples - i.e. three [ContainerMetricsSnapshot objects](https://github.com/kubernetes/autoscaler/blob/14ed794bee002422d0256a4fdd853681399c78cb/vertical-pod-autoscaler/pkg/recommender/input/metrics/metrics_client.go#L33) - per pod) would be aggregated into two new pod-level histograms during each recommender loop. Overall, this would result in two separate pod-level histograms: one for CPU samples and another for memory samples.

The differences in how resource sample aggregation (both CPU and memory) works at the pod level versus the container level are as follows, at the pod level:
* All container usage samples for each resource type belonging to a pod are discarded if a sample for any of its running containers is missing.
* For CPU usage, the recommender sums the CPU samples from all running regular containers in the pod, calculates the weight of the aggregated sample, and adds it to the appropriate bucket in the pod-level CPU histogram.
* For memory usage, the recommender sums the memory samples from all running regular containers in the pod. It adds the resulting sample to the pod-level memory histogram only when the sum exceeds the current pod-level peak within the current aggregation interval.

All other mechanisms remain unchanged, including:
* The initialization of the new pod-level histograms follows the same approach as the existing container-level histograms. Specifically, they use the same [cpuHistogramOptions](https://github.com/kubernetes/autoscaler/blob/6a616ea0c5ea0cb6111240073a9273b3467c064e/vertical-pod-autoscaler/pkg/recommender/model/aggregations_config.go#L89) and memoryHistogramOptions. These are the planned parameters for the alpha release. Based on community feedback and testing, we may adjust these parameters in future iterations.
* Calculating the sample weight by using the decayFactor
* Determining the histogram bucket index
* Applying the confidence multiplier

#### Maintaining checkpoints

The new pod-level histograms are then periodically stored in VerticalPodAutoscalerCheckpoint objects (i.e. pod-level checkpoint objects) based on the --recommender-interval flag. As a result, the pod-level CPU and memory histograms for a targeted controller are stored together within a single pod-level checkpoint object.

Since the current container-level checkpoints are garbage collected, the new pod-level checkpoint objects would follow the same garbage collection behavior. Additionally, the existing checkpoint flags (e.g. checkpoints-gc-interval) would work the same way for pod-level checkpoint objects.

A container-level checkpoint object name consists of the [VPA object name and the container name](https://github.com/kubernetes/autoscaler/blob/6d05073583657186351fd5df942980041ee8e71b/vertical-pod-autoscaler/pkg/recommender/checkpoint/checkpoint_writer.go#L84). The pod-level checkpoint object name should be based only on the VPA object name.

Since the ContainerName field in the [VerticalPodAutoscalerCheckpointSpec](https://github.com/kubernetes/autoscaler/blob/6d05073583657186351fd5df942980041ee8e71b/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1/types.go#L500) struct is optional, we can reuse the struct for pod-level checkpoint objects by not setting the ContainerName field. In this regard, there is no need for an API change for pod-level checkpoints within types.go.

#### New Global Pod maximums

This document proposes adding two new recommender flags to cap pod-level recommendations, following the same behavior as the existing container-level flags (e.g. `--container-recommendation-max-allowed-cpu`):
* `--pod-recommendation-max-allowed-cpu`: sets the global maximum CPU that the recommender may recommend at the pod level.
* `--pod-recommendation-max-allowed-memory`: sets the global maximum memory that the recommender may recommend at the pod level.
The Pod-level maxAllowed value defined in the podPolicy stanza takes precedence over the global Pod maximums when both are specified.

#### New Flags for Configuring Pod Level Percentiles

This document proposes adding new recommender flags that allow customization of the percentiles used for pod-level LowerBound, Target, and UpperBound recommendations. The following flags are introduced:
* `--pod-level-recommendation-lower-bound-cpu-percentile` (defaults to 0.5)
* `--pod-level-recommendation-lower-bound-memory-percentile` (defaults to 0.5)
* `--pod-level-target-cpu-percentile` (defaults to 0.9)
* `--pod-level-target-memory-percentile` (defaults to 0.9)
* `--pod-level-recommendation-upper-bound-cpu-percentile` (defaults to 0.95)
* `--pod-level-recommendation-upper-bound-memory-percentile` (defaults to 0.95)

### Admission Controller

The admission controller is extended to be capable of calculating pod-level resource patches (requests and limits) based on the pod-level Targets from the VPA object's status stanza, and applying them at admission when a targeted pod is recreated.

#### Dynamic Validation

The admission controller would reject VPA objects that define both `containerPolicies` and the new `podPolicy` stanza, because in Phase 1 the VPA cannot manage both container-level and pod-level resource stanzas simultaneously.

### Updater

With the introduction of pod-level recommendations and pod-level resources, the updater must make eviction and in-place update decisions for these pods. The decision making logic remains consistent with the latest VPA behavior. In other words:
* Evict a pod or apply a pod-level in-place update when the pod-level resource request in the pod spec is lower than the LowerBound or higher than the UpperBound.
* Evict a pod or apply a pod-level in-place update when the [resourceDiff](https://github.com/kubernetes/autoscaler/blob/14ed794bee002422d0256a4fdd853681399c78cb/vertical-pod-autoscaler/pkg/updater/priority/priority_processor.go#L97) is greater than 0.1 and the pod is not categorized as a "short-lived pod".
* Evict a pod or apply a pod-level in-place update if any container in the pod has experienced an OOM kill.

### capping.go

The capping.go logic must be updated so that both the admission controller and the updater can adjust pod-level recommendations based on LimitRange objects with Pod type in the cluster. Additionally, both components must enforce constraints defined by the new pod-level minAllowed and maxAllowed fields.

### Container LimitRange Objects

This AEP proposes to skip enforcing constraints defined by container LimitRange objects for pod-level recommendations. The built-in LimitRanger admission controller automatically sets default container-level requests and limits for these Pods, which is undesirable when pod-level resources are in use.

This behavior is also described in the [pod-level Resource Spec KEP](https://github.com/kubernetes/enhancements/blob/master/keps/sig-node/2837-pod-level-resource-spec/README.md#admission-controller), which recommends placing Pods in namespaces without container LimitRange objects.

To handle container LimitRange objects for Pods whose pod-level resources stanza is managed by VPA, this AEP proposes the following:
* Inform the user through logs produced by the updater and admission controller that a Container LimitRange object exists in the namespace, and suggest that the user either remove it from the cluster or move the workload to a namespace where such a LimitRange object does not exist. The implementation does not enforce these constraints on pod-level recommendations in this case.

### Pod Annotations

Define new pod-level annotations that the admission controller and the updater apply as needed:
* Annotate the Pod when Pod-level requests and limits are managed. Explicitly indicate when a Pod-level limit is not set.
* Annotate the Pod when a Pod-level resource limit (CPU or memory) is capped to fit within the Min or Max defined in a Pod LimitRange object.

### Handling Container Addition and Removal

When a user adds a new container to a pod, the pod-level recommendation will not immediately reflect the additional resource usage. Instead, the recommendation will gradually adapt as the recommender collects new usage samples over subsequent recommendation cycles. During this period, the pod-level recommendation may underestimate the pod's resource requirements, potentially leading to CPU starvation or out-of-memory (OOM) kills.

To mitigate this, users can increase `spec.resourcePolicy.podPolicy.minAllowed` before adding the new container. When the pod is recreated, the admission controller applies the pod-level recommendation while ensuring that the configured `spec.resourcePolicy.podPolicy.minAllowed` values are respected.

Removing a container from a pod that uses pod-level autoscaling naturally causes older recommendations to become less relevant. As new usage samples are collected from the remaining containers, the pod-level histograms gradually converge to reflect the new resource usage pattern, causing previous samples to be forgotten over time.

### VPAPodLevelResources Feature Gate

This section describes the behavior of the VPA components when the proposed feature gate is disabled.

#### Recommender

* The recommender ignores the podPolicy stanza.
* The recommender does not create new Pod-level histograms or aggregate usage samples into existing Pod-level histograms.
* The recommender does not calculate or store pod-level recommendations in the VPA status stanza.
* When the feature gate is disabled, any existing pod-level recommendations are removed from the VPA object's status stanza during the next recommender cycle.
* The recommender ignores the new pod-level flags introduced by this proposal.
* When the feature gate is disabled, the recommender stops aggregating resource usage samples into pod-level histograms. Over time, these histograms may be garbage collected, following the same garbage collection behavior as `AggregateContainerStates`. For more details, see the `RateLimitedGarbageCollectAggregateCollectionStates` method.
* When the feature gate is disabled, the recommender stops persisting pod-level histograms from the cluster state to checkpoint objects during subsequent recommender intervals. Existing pod-level checkpoint objects are garbage collected once they no longer have a corresponding VPA object, following the same behavior as the garbage collection of container-level checkpoint objects.

#### Updater

* The updater ignores the podPolicy stanza.
* The updater stops managing pod-level resource stanzas. In other words, existing pod-level resources remain unchanged, which may be an issue, as the updater falls back to managing only container-level resources, potentially violating the "frozen" pod-level resource stanza at some point.

#### Admission Controller

* The admission controller rejects VPA objects that include the `podPolicy` stanza.
* The admission controller stops managing pod-level resource stanzas. In other words, existing pod-level resources remain unchanged, which may be an issue, as the admission controller falls back to managing only container-level resources, potentially violating the "frozen" pod-level resource stanza at some point.

### Test Plan

<!--
Describe how this change will be tested. At minimum, reviewers expect:
- unit tests for new logic,
- e2e tests for user-visible behavior (what scenarios will be covered?).
Integration tests are not required for most AEPs, but mention them if they
apply.
-->

This AEP proposes comprehensive unit test coverage for all new code paths and additionally plans to introduce the following e2e tests:

#### Recommender

* Create tests to verify that the recommender can calculate Pod-level recommendations and store them in the VPA object's status stanza.

#### Admission Controller

* Create tests to verify the behavior of the new Pod-level minAllowed and maxAllowed fields.
* Create tests to verify that, when a Pod LimitRange is present, the admission controller correctly enforces the configured minimum and maximum resource constraints.

#### Updater

* Create tests to verify that the updater can evict Pods based on Pod-level recommendations and Pod-level resource stanzas.
* Create tests to verify that the updater can apply Pod-level recommendations through in-place updates.

### Feature Enablement and Rollback

<!--
Answer the following if this AEP is gated by a feature flag:

- Feature gate name:
- Components depending on the feature gate (e.g. updater, admission-controller,
  recommender):
- What happens when the gate is enabled?
- What happens when the gate is disabled after being enabled? In particular,
  what happens to VPA objects already configured with the new field?

If the change is not gated (for example, a backward-compatible API extension
with safe defaults), state that explicitly and explain why no gate is needed.
-->

#### Feature Enablement 

Use a VPA release that includes this feature, and enable it by setting `--feature-gates=VPAPodLevelResources=true` across all components. Then begin deploying new workloads targeted by a VPA object where pod-level scaling is enabled via the new podPolicy stanza.

#### Rollback

Downgrading VPA from a version that includes this feature does not impact running workloads where pod-level scaling is NOT enabled.

After the downgrade, workloads with pod-level scaling enabled are not disrupted immediately. Pod-level resource specifications remain in Pod specs, but VPA reverts to managing container-level resources only.

However, once VPA stops managing pod-level resources and updates only container-level resources, newly applied container-level requests or limits may violate the unmanaged pod-level resource stanzas.

To avoid such violations, follow these steps for downgrade:
* Recreate the VPA object using only the containerPolicies stanza, fully removing the new podPolicy stanza.
* Downgrade the VPA version and disable the feature flag at the same time. Wait until container-level recommendations are present in the VPA object's status stanza.
* Remove the pod-level resources stanza from the higher level controller, if present. If the pod-level resource stanza is not present, trigger a restart using the `kubectl rollout restart` command to remove the stale pod-level resource stanza from the Pod specifications.

### Graduation Criteria

<!--
A few bullets describing what needs to be true to move the feature from alpha
to beta and from beta to GA. For most AEPs this is short — typical signals are
"tests are stable for N releases", "no open bugs against the feature gate",
and "positive user feedback". Remove this subsection if the change does not go
through a graduation lifecycle (e.g. a pure bug fix).
-->

#### Phase 1: Alpha (target VPA 1.9.0 version)

* Feature gate is disabled by default. This is an opt-in feature that can be enabled by setting the `VPAPodLevelResources` feature gate across all VPA components.
* Support the ability for the recommender to calculate Pod-level recommendations, and enable both the updater and the admission controller to use these recommendations for all existing VPA behaviors, including evicting pods, applying in-place Pod-level updates, and applying recommendations at admission.
* Enabling the feature in alpha by setting `spec.resourcePolicy.podPolicy.mode` to `Auto` causes the recommender to aggregate CPU and memory usage samples into pod-level histograms for the targeted controller. Disabling the feature stops further aggregation into these histograms. This behavior may change in future alpha releases based on community feedback and performance testing. For example, we may decide to maintain pod-level histograms for all controllers by default if the overhead is acceptable. In the initial alpha implementation, enabling the feature for a running workload that previously had pod-level aggregation disabled will not immediately produce optimal pod-level recommendations, as sufficient usage samples must first be collected.
* Unit test coverage.
* E2E tests.
* Documentation mentioning high level design.

#### Phase 2: Beta (target VPA 1.10.0 version)

* The feature is enabled by default. In this phase, the feature flag `VPAPodLevelResources` is still present, so cluster operators can opt out if they want to.
* This phase includes implementing community feedback, addressing bugs, and incorporating improvements as needed.
* Adds additional unit tests and e2e tests to cover the newly introduced code paths.
* Update the documentation if applicable.

#### GA (stable)

* Feature is enabled by default. The `VPAPodLevelResources` feature gate is removed from all VPA components.
* No major bugs reported for 3 months.

### Version Skew

<!--
The VPA ships multiple components (recommender, updater, admission-controller).
Describe what happens when they are not all running the same version during a
rollout — for example, a new recommender writing a field that an older updater
does not understand. If the feature gate fully mitigates skew (the gate must
be enabled on all components before the behavior takes effect), state that.
Remove this subsection if only one component is affected.
-->

All new functionality is controlled by the `VPAPodLevelResources` feature gate. If any VPA component is running a version where the feature is disabled or unavailable, the component will fall back to container-level behavior.

For users who want to use this feature, it is their responsibility to run the same version of all VPA components to ensure correct VPA behavior. If this requirement is not met, users should expect missing functionality and bug fixes introduced in newer releases.

### Kubernetes Version Compatibility

<!--
Call out any minimum Kubernetes version this feature requires, and what the
VPA does when running on an older version. Fill this in when the AEP depends
on an upstream Kubernetes feature (e.g. KEP-1287 for in-place updates).
Otherwise, remove this subsection.
-->

This new feature, proposed by this document, relies on two feature gates:
* `PodLevelResources` - implements pod-level resources.
* `InPlacePodLevelResourcesVerticalScaling` - implements in-place scaling at the pod-level.

Therefore, it is recommended to use a Kubernetes version that includes both of these feature gates (use [this page](https://kubernetes.io/docs/reference/command-line-tools-reference/feature-gates) to verify this information). The alpha version of `InPlacePodLevelResourcesVerticalScaling` was released in v1.35, so at least this Kubernetes version should be used.

### VPA Object Examples

#### Example 1 - Observe Pod Level recommendations

If the user only wants to observe pod-level recommendations from the VPA object's status stanza, and wants the updater and admission controller to skip managing the pod, the VPA object can be configured as follows:

```yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: workload1
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: workload1
  updatePolicy:
    updateMode: Off
  resourcePolicy:
    podPolicy:
      mode: Auto
```

#### Example 2 - Incorrect VPA Object

When the admission controller is started with the VPAPodLevelResources feature flag enabled, it rejects the VPA object below, because in Phase 1 the containerPolicies and podPolicy stanzas cannot be used together:

```yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: workload1
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
        mode: "Auto"
        controlledResources:
          - "memory"
          - "cpu"
    podPolicy:
      mode: "Auto"
```

#### Example 3

With the VPA object below, the user instructs the recommender to start calculating pod-level recommendations and enables the updater and admission controller to manage pod-level resources for the targeted pods.

```yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: workload1
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: workload1
  updatePolicy:
    updateMode: 'Recreate'
  resourcePolicy:
    podPolicy:
      mode: "Auto"
```


## Implementation History

<!--
Track major milestones using absolute dates (YYYY-MM-DD):
- initial version
- significant design changes
- the first VPA release where the feature shipped
- graduation to beta / GA
-->

- 2026-07-15: initial version