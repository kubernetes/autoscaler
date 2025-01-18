# API Reference

## Packages
- [autoscaling.k8s.io/v1](#autoscalingk8siov1)


## autoscaling.k8s.io/v1

Package v1 contains definitions of Vertical Pod Autoscaler related objects.



#### ContainerControlledValues

_Underlying type:_ _string_

ContainerControlledValues controls which resource value should be autoscaled.

_Validation:_
- Enum: [RequestsAndLimits RequestsOnly]

_Appears in:_
- [ContainerResourcePolicy](#containerresourcepolicy)

| Field | Description |
| --- | --- |
| `RequestsAndLimits` | ContainerControlledValuesRequestsAndLimits means resource request and limits<br />are scaled automatically. The limit is scaled proportionally to the request.<br /> |
| `RequestsOnly` | ContainerControlledValuesRequestsOnly means only requested resource is autoscaled.<br /> |


#### ContainerResourcePolicy



ContainerResourcePolicy controls how autoscaler computes the recommended
resources for a specific container.



_Appears in:_
- [PodResourcePolicy](#podresourcepolicy)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `containerName` _string_ | Name of the container or DefaultContainerResourcePolicy, in which<br />case the policy is used by the containers that don't have their own<br />policy specified. |  |  |
| `mode` _[ContainerScalingMode](#containerscalingmode)_ | Whether autoscaler is enabled for the container. The default is "Auto". |  | Enum: [Auto Off] <br /> |
| `minAllowed` _[ResourceList](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#resourcelist-v1-core)_ | Specifies the minimal amount of resources that will be recommended<br />for the container. The default is no minimum. |  |  |
| `maxAllowed` _[ResourceList](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#resourcelist-v1-core)_ | Specifies the maximum amount of resources that will be recommended<br />for the container. The default is no maximum. |  |  |
| `controlledResources` _[ResourceName](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#resourcename-v1-core)_ | Specifies the type of recommendations that will be computed<br />(and possibly applied) by VPA.<br />If not specified, the default of [ResourceCPU, ResourceMemory] will be used. |  |  |
| `controlledValues` _[ContainerControlledValues](#containercontrolledvalues)_ | Specifies which resource values should be controlled.<br />The default is "RequestsAndLimits". |  | Enum: [RequestsAndLimits RequestsOnly] <br /> |


#### ContainerScalingMode

_Underlying type:_ _string_

ContainerScalingMode controls whether autoscaler is enabled for a specific
container.

_Validation:_
- Enum: [Auto Off]

_Appears in:_
- [ContainerResourcePolicy](#containerresourcepolicy)

| Field | Description |
| --- | --- |
| `Auto` | ContainerScalingModeAuto means autoscaling is enabled for a container.<br /> |
| `Off` | ContainerScalingModeOff means autoscaling is disabled for a container.<br /> |


#### EvictionChangeRequirement

_Underlying type:_ _string_

EvictionChangeRequirement refers to the relationship between the new target recommendation for a Pod and its current requests, what kind of change is necessary for the Pod to be evicted

_Validation:_
- Enum: [TargetHigherThanRequests TargetLowerThanRequests]

_Appears in:_
- [EvictionRequirement](#evictionrequirement)

| Field | Description |
| --- | --- |
| `TargetHigherThanRequests` | TargetHigherThanRequests means the new target recommendation for a Pod is higher than its current requests, i.e. the Pod is scaled up<br /> |
| `TargetLowerThanRequests` | TargetLowerThanRequests means the new target recommendation for a Pod is lower than its current requests, i.e. the Pod is scaled down<br /> |


#### EvictionRequirement



EvictionRequirement defines a single condition which needs to be true in
order to evict a Pod



_Appears in:_
- [PodUpdatePolicy](#podupdatepolicy)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `resources` _[ResourceName](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#resourcename-v1-core) array_ | Resources is a list of one or more resources that the condition applies<br />to. If more than one resource is given, the EvictionRequirement is fulfilled<br />if at least one resource meets `changeRequirement`. |  |  |
| `changeRequirement` _[EvictionChangeRequirement](#evictionchangerequirement)_ |  |  | Enum: [TargetHigherThanRequests TargetLowerThanRequests] <br /> |


#### HistogramCheckpoint



HistogramCheckpoint contains data needed to reconstruct the histogram.



_Appears in:_
- [VerticalPodAutoscalerCheckpointStatus](#verticalpodautoscalercheckpointstatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `referenceTimestamp` _[Time](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#time-v1-meta)_ | Reference timestamp for samples collected within this histogram. |  |  |
| `bucketWeights` _object (keys:integer, values:integer)_ | Map from bucket index to bucket weight. |  | Type: object <br />XPreserveUnknownFields: \{\} <br /> |
| `totalWeight` _float_ | Sum of samples to be used as denominator for weights from BucketWeights. |  |  |


#### PodResourcePolicy



PodResourcePolicy controls how autoscaler computes the recommended resources
for containers belonging to the pod. There can be at most one entry for every
named container and optionally a single wildcard entry with `containerName` = '*',
which handles all containers that don't have individual policies.



_Appears in:_
- [VerticalPodAutoscalerSpec](#verticalpodautoscalerspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `containerPolicies` _[ContainerResourcePolicy](#containerresourcepolicy) array_ | Per-container resource policies. |  |  |


#### PodUpdatePolicy



PodUpdatePolicy describes the rules on how changes are applied to the pods.



_Appears in:_
- [VerticalPodAutoscalerSpec](#verticalpodautoscalerspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `updateMode` _[UpdateMode](#updatemode)_ | Controls when autoscaler applies changes to the pod resources.<br />The default is 'Auto'. |  | Enum: [Off Initial Recreate Auto] <br /> |
| `minReplicas` _integer_ | Minimal number of replicas which need to be alive for Updater to attempt<br />pod eviction (pending other checks like PDB). Only positive values are<br />allowed. Overrides global '--min-replicas' flag. |  |  |
| `evictionRequirements` _[EvictionRequirement](#evictionrequirement) array_ | EvictionRequirements is a list of EvictionRequirements that need to<br />evaluate to true in order for a Pod to be evicted. If more than one<br />EvictionRequirement is specified, all of them need to be fulfilled to allow eviction. |  |  |


#### RecommendedContainerResources



RecommendedContainerResources is the recommendation of resources computed by
autoscaler for a specific container. Respects the container resource policy
if present in the spec. In particular the recommendation is not produced for
containers with `ContainerScalingMode` set to 'Off'.



_Appears in:_
- [RecommendedPodResources](#recommendedpodresources)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `containerName` _string_ | Name of the container. |  |  |
| `target` _[ResourceList](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#resourcelist-v1-core)_ | Recommended amount of resources. Observes ContainerResourcePolicy. |  |  |
| `lowerBound` _[ResourceList](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#resourcelist-v1-core)_ | Minimum recommended amount of resources. Observes ContainerResourcePolicy.<br />This amount is not guaranteed to be sufficient for the application to operate in a stable way, however<br />running with less resources is likely to have significant impact on performance/availability. |  |  |
| `upperBound` _[ResourceList](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#resourcelist-v1-core)_ | Maximum recommended amount of resources. Observes ContainerResourcePolicy.<br />Any resources allocated beyond this value are likely wasted. This value may be larger than the maximum<br />amount of application is actually capable of consuming. |  |  |
| `uncappedTarget` _[ResourceList](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#resourcelist-v1-core)_ | The most recent recommended resources target computed by the autoscaler<br />for the controlled pods, based only on actual resource usage, not taking<br />into account the ContainerResourcePolicy.<br />May differ from the Recommendation if the actual resource usage causes<br />the target to violate the ContainerResourcePolicy (lower than MinAllowed<br />or higher that MaxAllowed).<br />Used only as status indication, will not affect actual resource assignment. |  |  |


#### RecommendedPodResources



RecommendedPodResources is the recommendation of resources computed by
autoscaler. It contains a recommendation for each container in the pod
(except for those with `ContainerScalingMode` set to 'Off').



_Appears in:_
- [VerticalPodAutoscalerStatus](#verticalpodautoscalerstatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `containerRecommendations` _[RecommendedContainerResources](#recommendedcontainerresources) array_ | Resources recommended by the autoscaler for each container. |  |  |


#### UpdateMode

_Underlying type:_ _string_

UpdateMode controls when autoscaler applies changes to the pod resources.

_Validation:_
- Enum: [Off Initial Recreate Auto]

_Appears in:_
- [PodUpdatePolicy](#podupdatepolicy)

| Field | Description |
| --- | --- |
| `Off` | UpdateModeOff means that autoscaler never changes Pod resources.<br />The recommender still sets the recommended resources in the<br />VerticalPodAutoscaler object. This can be used for a "dry run".<br /> |
| `Initial` | UpdateModeInitial means that autoscaler only assigns resources on pod<br />creation and does not change them during the lifetime of the pod.<br /> |
| `Recreate` | UpdateModeRecreate means that autoscaler assigns resources on pod<br />creation and additionally can update them during the lifetime of the<br />pod by deleting and recreating the pod.<br /> |
| `Auto` | UpdateModeAuto means that autoscaler assigns resources on pod creation<br />and additionally can update them during the lifetime of the pod,<br />using any available update method. Currently this is equivalent to<br />Recreate, which is the only available update method.<br /> |


#### VerticalPodAutoscaler



VerticalPodAutoscaler is the configuration for a vertical pod
autoscaler, which automatically manages pod resources based on historical and
real time resource utilization.



_Appears in:_
- [VerticalPodAutoscalerList](#verticalpodautoscalerlist)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `kind` _string_ | Kind is a string value representing the REST resource this object represents.<br />Servers may infer this from the endpoint the client submits requests to.<br />Cannot be updated.<br />In CamelCase.<br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds |  |  |
| `apiVersion` _string_ | APIVersion defines the versioned schema of this representation of an object.<br />Servers should convert recognized schemas to the latest internal value, and<br />may reject unrecognized values.<br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources |  |  |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[VerticalPodAutoscalerSpec](#verticalpodautoscalerspec)_ | Specification of the behavior of the autoscaler.<br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status. |  |  |
| `status` _[VerticalPodAutoscalerStatus](#verticalpodautoscalerstatus)_ | Current information about the autoscaler. |  |  |


#### VerticalPodAutoscalerCheckpoint



VerticalPodAutoscalerCheckpoint is the checkpoint of the internal state of VPA that
is used for recovery after recommender's restart.



_Appears in:_
- [VerticalPodAutoscalerCheckpointList](#verticalpodautoscalercheckpointlist)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `kind` _string_ | Kind is a string value representing the REST resource this object represents.<br />Servers may infer this from the endpoint the client submits requests to.<br />Cannot be updated.<br />In CamelCase.<br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds |  |  |
| `apiVersion` _string_ | APIVersion defines the versioned schema of this representation of an object.<br />Servers should convert recognized schemas to the latest internal value, and<br />may reject unrecognized values.<br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources |  |  |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[VerticalPodAutoscalerCheckpointSpec](#verticalpodautoscalercheckpointspec)_ | Specification of the checkpoint.<br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status. |  |  |
| `status` _[VerticalPodAutoscalerCheckpointStatus](#verticalpodautoscalercheckpointstatus)_ | Data of the checkpoint. |  |  |




#### VerticalPodAutoscalerCheckpointSpec



VerticalPodAutoscalerCheckpointSpec is the specification of the checkpoint object.



_Appears in:_
- [VerticalPodAutoscalerCheckpoint](#verticalpodautoscalercheckpoint)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `vpaObjectName` _string_ | Name of the VPA object that stored VerticalPodAutoscalerCheckpoint object. |  |  |
| `containerName` _string_ | Name of the checkpointed container. |  |  |


#### VerticalPodAutoscalerCheckpointStatus



VerticalPodAutoscalerCheckpointStatus contains data of the checkpoint.



_Appears in:_
- [VerticalPodAutoscalerCheckpoint](#verticalpodautoscalercheckpoint)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `lastUpdateTime` _[Time](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#time-v1-meta)_ | The time when the status was last refreshed. |  |  |
| `version` _string_ | Version of the format of the stored data. |  |  |
| `cpuHistogram` _[HistogramCheckpoint](#histogramcheckpoint)_ | Checkpoint of histogram for consumption of CPU. |  |  |
| `memoryHistogram` _[HistogramCheckpoint](#histogramcheckpoint)_ | Checkpoint of histogram for consumption of memory. |  |  |
| `firstSampleStart` _[Time](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#time-v1-meta)_ | Timestamp of the fist sample from the histograms. |  |  |
| `lastSampleStart` _[Time](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#time-v1-meta)_ | Timestamp of the last sample from the histograms. |  |  |
| `totalSamplesCount` _integer_ | Total number of samples in the histograms. |  |  |


#### VerticalPodAutoscalerCondition



VerticalPodAutoscalerCondition describes the state of
a VerticalPodAutoscaler at a certain point.



_Appears in:_
- [VerticalPodAutoscalerStatus](#verticalpodautoscalerstatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `type` _[VerticalPodAutoscalerConditionType](#verticalpodautoscalerconditiontype)_ | type describes the current condition |  |  |
| `status` _[ConditionStatus](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#conditionstatus-v1-core)_ | status is the status of the condition (True, False, Unknown) |  |  |
| `lastTransitionTime` _[Time](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#time-v1-meta)_ | lastTransitionTime is the last time the condition transitioned from<br />one status to another |  |  |
| `reason` _string_ | reason is the reason for the condition's last transition. |  |  |
| `message` _string_ | message is a human-readable explanation containing details about<br />the transition |  |  |


#### VerticalPodAutoscalerConditionType

_Underlying type:_ _string_

VerticalPodAutoscalerConditionType are the valid conditions of
a VerticalPodAutoscaler.



_Appears in:_
- [VerticalPodAutoscalerCondition](#verticalpodautoscalercondition)





#### VerticalPodAutoscalerRecommenderSelector



VerticalPodAutoscalerRecommenderSelector points to a specific Vertical Pod Autoscaler recommender.
In the future it might pass parameters to the recommender.



_Appears in:_
- [VerticalPodAutoscalerSpec](#verticalpodautoscalerspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name of the recommender responsible for generating recommendation for this object. |  |  |


#### VerticalPodAutoscalerSpec



VerticalPodAutoscalerSpec is the specification of the behavior of the autoscaler.



_Appears in:_
- [VerticalPodAutoscaler](#verticalpodautoscaler)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `targetRef` _[CrossVersionObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#crossversionobjectreference-v1-autoscaling)_ | TargetRef points to the controller managing the set of pods for the<br />autoscaler to control - e.g. Deployment, StatefulSet. VerticalPodAutoscaler<br />can be targeted at controller implementing scale subresource (the pod set is<br />retrieved from the controller's ScaleStatus) or some well known controllers<br />(e.g. for DaemonSet the pod set is read from the controller's spec).<br />If VerticalPodAutoscaler cannot use specified target it will report<br />ConfigUnsupported condition.<br />Note that VerticalPodAutoscaler does not require full implementation<br />of scale subresource - it will not use it to modify the replica count.<br />The only thing retrieved is a label selector matching pods grouped by<br />the target resource. |  |  |
| `updatePolicy` _[PodUpdatePolicy](#podupdatepolicy)_ | Describes the rules on how changes are applied to the pods.<br />If not specified, all fields in the `PodUpdatePolicy` are set to their<br />default values. |  |  |
| `resourcePolicy` _[PodResourcePolicy](#podresourcepolicy)_ | Controls how the autoscaler computes recommended resources.<br />The resource policy may be used to set constraints on the recommendations<br />for individual containers.<br />If any individual containers need to be excluded from getting the VPA recommendations, then<br />it must be disabled explicitly by setting mode to "Off" under containerPolicies.<br />If not specified, the autoscaler computes recommended resources for all containers in the pod,<br />without additional constraints. |  |  |
| `recommenders` _[VerticalPodAutoscalerRecommenderSelector](#verticalpodautoscalerrecommenderselector) array_ | Recommender responsible for generating recommendation for this object.<br />List should be empty (then the default recommender will generate the<br />recommendation) or contain exactly one recommender. |  |  |


#### VerticalPodAutoscalerStatus



VerticalPodAutoscalerStatus describes the runtime state of the autoscaler.



_Appears in:_
- [VerticalPodAutoscaler](#verticalpodautoscaler)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `recommendation` _[RecommendedPodResources](#recommendedpodresources)_ | The most recently computed amount of resources recommended by the<br />autoscaler for the controlled pods. |  |  |
| `conditions` _[VerticalPodAutoscalerCondition](#verticalpodautoscalercondition) array_ | Conditions is the set of conditions required for this autoscaler to scale its target,<br />and indicates whether or not those conditions are met. |  |  |


