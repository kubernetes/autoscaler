# API Reference

## Packages
- [autoscaling.x-k8s.io/v1](#autoscalingx-k8siov1)
- [autoscaling.x-k8s.io/v1alpha1](#autoscalingx-k8siov1alpha1)
- [autoscaling.x-k8s.io/v1beta1](#autoscalingx-k8siov1beta1)


## autoscaling.x-k8s.io/v1

Package v1 contains definitions of Provisioning Request related objects.

Package v1 contains definitions of Provisioning Request related objects.

Package v1 contains definitions of Provisioning Request related objects.



#### Detail

_Underlying type:_ _string_

Detail is limited to 32768 characters.

_Validation:_
- MaxLength: 32768

_Appears in:_
- [ProvisioningRequestStatus](#provisioningrequeststatus)



#### Parameter

_Underlying type:_ _string_

Parameter is limited to 255 characters.

_Validation:_
- MaxLength: 255

_Appears in:_
- [ProvisioningRequestSpec](#provisioningrequestspec)



#### PodSet



PodSet represents one group of pods for Provisioning Request to provision capacity.



_Appears in:_
- [ProvisioningRequestSpec](#provisioningrequestspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `podTemplateRef` _[Reference](#reference)_ | PodTemplateRef is a reference to a PodTemplate object that is representing pods<br />that will consume this reservation (must be within the same namespace).<br />Users need to make sure that the  fields relevant to scheduler (e.g. node selector tolerations)<br />are consistent between this template and actual pods consuming the Provisioning Request. |  | Required: \{\} <br /> |
| `count` _integer_ | Count contains the number of pods that will be created with a given<br />template. |  | Minimum: 1 <br /> |


#### ProvisioningRequest



ProvisioningRequest is a way to express additional capacity
that we would like to provision in the cluster. Cluster Autoscaler
can use this information in its calculations and signal if the capacity
is available in the cluster or actively add capacity if needed.



_Appears in:_
- [ProvisioningRequestList](#provisioningrequestlist)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `kind` _string_ | Kind is a string value representing the REST resource this object represents.<br />Servers may infer this from the endpoint the client submits requests to.<br />Cannot be updated.<br />In CamelCase.<br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds |  | Optional: \{\} <br /> |
| `apiVersion` _string_ | APIVersion defines the versioned schema of this representation of an object.<br />Servers should convert recognized schemas to the latest internal value, and<br />may reject unrecognized values.<br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources |  | Optional: \{\} <br /> |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.36/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  | Optional: \{\} <br /> |
| `spec` _[ProvisioningRequestSpec](#provisioningrequestspec)_ | Spec contains specification of the ProvisioningRequest object.<br />More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status.<br />The spec is immutable, to make changes to the request users are expected to delete an existing<br />and create a new object with the corrected fields. |  | Required: \{\} <br /> |
| `status` _[ProvisioningRequestStatus](#provisioningrequeststatus)_ | Status of the ProvisioningRequest. CA constantly reconciles this field. |  | Optional: \{\} <br /> |




#### ProvisioningRequestSpec



ProvisioningRequestSpec is a specification of additional pods for which we
would like to provision additional resources in the cluster.



_Appears in:_
- [ProvisioningRequest](#provisioningrequest)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `podSets` _[PodSet](#podset) array_ | PodSets lists groups of pods for which we would like to provision<br />resources. |  | MaxItems: 32 <br />MinItems: 1 <br />Required: \{\} <br /> |
| `provisioningClassName` _string_ | ProvisioningClassName describes the different modes of provisioning the resources.<br />Currently there is no support for 'ProvisioningClass' objects.<br />Supported values:<br />* check-capacity.autoscaling.x-k8s.io - check if current cluster state can fullfil this request,<br />  do not reserve the capacity. Users should provide a reference to a valid PodTemplate object.<br />  CA will check if there is enough capacity in cluster to fulfill the request and put<br />  the answer in 'CapacityAvailable' condition.<br />* best-effort-atomic-scale-up.autoscaling.x-k8s.io - provision the resources in an atomic manner.<br />  Users should provide a reference to a valid PodTemplate object.<br />  CA will try to create the VMs in an atomic manner, clean any partially provisioned VMs<br />  and re-try the operation in a exponential back-off manner. Users can configure the timeout<br />  duration after which the request will fail by 'ValidUntilSeconds' key in 'Parameters'.<br />  CA will set 'Failed=true' or 'Provisioned=true' condition according to the outcome.<br />* ... - potential other classes that are specific to the cloud providers.<br />'kubernetes.io' suffix is reserved for the modes defined in Kubernetes projects. |  | MaxLength: 253 <br />Pattern: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$` <br />Required: \{\} <br /> |
| `parameters` _object (keys:string, values:[Parameter](#parameter))_ | Parameters contains all other parameters classes may require.<br />'best-effort-atomic-scale-up.autoscaling.x-k8s.io' supports 'ValidUntilSeconds' parameter, which should contain<br /> a string denoting duration for which we should retry (measured since creation fo the CR). |  | MaxProperties: 100 <br />Optional: \{\} <br /> |


#### ProvisioningRequestStatus



ProvisioningRequestStatus represents the status of the resource reservation.



_Appears in:_
- [ProvisioningRequest](#provisioningrequest)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `conditions` _[Condition](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.36/#condition-v1-meta) array_ | Conditions represent the observations of a Provisioning Request's<br />current state. Those will contain information whether the capacity<br />was found/created or if there were any issues. The condition types<br />may differ between different provisioning classes. |  | Optional: \{\} <br /> |
| `provisioningClassDetails` _object (keys:string, values:[Detail](#detail))_ | ProvisioningClassDetails contains all other values custom provisioning classes may<br />want to pass to end users. |  | MaxProperties: 64 <br />Optional: \{\} <br /> |


#### Reference



Reference represents reference to an object within the same namespace.



_Appears in:_
- [PodSet](#podset)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name of the referenced object.<br />More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names#names |  | MaxLength: 253 <br />Pattern: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$` <br />Required: \{\} <br /> |



## autoscaling.x-k8s.io/v1alpha1

Package v1alpha1 contains the v1alpha1 API for the autoscaling.x-k8s.io group.
This API group defines custom resources used by the Cluster Autoscaler
for managing buffer capacity.

### Resource Types
- [CapacityBuffer](#capacitybuffer)
- [CapacityQuota](#capacityquota)
- [CapacityQuotaList](#capacityquotalist)



#### CapacityBuffer



CapacityBuffer is the configuration that an autoscaler can use to provision buffer capacity within a cluster.
This buffer is represented by placeholder pods that trigger the Cluster Autoscaler to scale up nodes in advance,
ensuring that there is always spare capacity available to handle sudden workload spikes or to speed up scaling events.



_Appears in:_
- [CapacityBufferList](#capacitybufferlist)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `autoscaling.x-k8s.io/v1alpha1` | | |
| `kind` _string_ | `CapacityBuffer` | | |
| `kind` _string_ | Kind is a string value representing the REST resource this object represents.<br />Servers may infer this from the endpoint the client submits requests to.<br />Cannot be updated.<br />In CamelCase.<br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds |  | Optional: \{\} <br /> |
| `apiVersion` _string_ | APIVersion defines the versioned schema of this representation of an object.<br />Servers should convert recognized schemas to the latest internal value, and<br />may reject unrecognized values.<br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources |  | Optional: \{\} <br /> |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.36/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[CapacityBufferSpec](#capacitybufferspec)_ | Spec defines the desired characteristics of the buffer. |  | Required: \{\} <br /> |
| `status` _[CapacityBufferStatus](#capacitybufferstatus)_ | Status represents the current state of the buffer and its readiness for autoprovisioning. |  | Optional: \{\} <br /> |




#### CapacityBufferSpec



CapacityBufferSpec defines the desired state of CapacityBuffer.



_Appears in:_
- [CapacityBuffer](#capacitybuffer)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `provisioningStrategy` _string_ | ProvisioningStrategy defines how the buffer is utilized.<br />"buffer.x-k8s.io/active-capacity" is the default strategy, where the buffer actively scales up the cluster by creating placeholder pods. | buffer.x-k8s.io/active-capacity | Optional: \{\} <br /> |
| `podTemplateRef` _[LocalObjectRef](#localobjectref)_ | PodTemplateRef is a reference to a PodTemplate resource in the same namespace<br />that declares the shape of a single chunk of the buffer. The pods created<br />from this template will be used as placeholder pods for the buffer capacity.<br />Exactly one of `podTemplateRef`, `scalableRef` should be specified. |  | Optional: \{\} <br /> |
| `scalableRef` _[ScalableRef](#scalableref)_ | ScalableRef is a reference to an object of a kind that has a scale subresource<br />and specifies its label selector field. This allows the CapacityBuffer to<br />manage the buffer by scaling an existing scalable resource.<br />Exactly one of `podTemplateRef`, `scalableRef` should be specified. |  | Optional: \{\} <br /> |
| `replicas` _integer_ | Replicas defines the desired number of buffer chunks to provision.<br />If neither `replicas` nor `percentage` is set, as many chunks as fit within<br />defined resource limits (if any) will be created. If both are set, the maximum<br />of the two will be used. |  | ExclusiveMinimum: false <br />Minimum: 0 <br />Optional: \{\} <br /> |
| `percentage` _integer_ | Percentage defines the desired buffer capacity as a percentage of the<br />`scalableRef`'s current replicas. This is only applicable if `scalableRef` is set.<br />The absolute number of replicas is calculated from the percentage by rounding up to a minimum of 1.<br />For example, if `scalableRef` has 10 replicas and `percentage` is 20, 2 buffer chunks will be created. |  | ExclusiveMinimum: false <br />Minimum: 0 <br />Optional: \{\} <br /> |
| `limits` _[ResourceList](#resourcelist)_ | Limits, if specified, will limit the number of chunks created for this buffer<br />based on total resource requests (e.g., CPU, memory). If there are no other<br />limitations for the number of chunks (i.e., `replicas` or `percentage` are not set),<br />this will be used to create as many chunks as fit into these limits. |  | Optional: \{\} <br /> |


#### CapacityBufferStatus



CapacityBufferStatus defines the observed state of CapacityBuffer.



_Appears in:_
- [CapacityBuffer](#capacitybuffer)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `podTemplateRef` _[LocalObjectRef](#localobjectref)_ | PodTemplateRef is the observed reference to the PodTemplate that was used<br />to provision the buffer. If this field is not set, and the `conditions`<br />indicate an error, it provides details about the error state. |  | Optional: \{\} <br /> |
| `replicas` _integer_ | Replicas is the actual number of buffer chunks currently provisioned. |  | Optional: \{\} <br /> |
| `podTemplateGeneration` _integer_ | PodTemplateGeneration is the observed generation of the PodTemplate, used<br />to determine if the status is up-to-date with the desired `spec.podTemplateRef`. |  | Optional: \{\} <br /> |
| `conditions` _[Condition](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.36/#condition-v1-meta) array_ | Conditions provide a standard mechanism for reporting the buffer's state.<br />The "Ready" condition indicates if the buffer is successfully provisioned<br />and active. Other conditions may report on various aspects of the buffer's<br />health and provisioning process. |  | Optional: \{\} <br /> |
| `provisioningStrategy` _string_ | ProvisioningStrategy defines how the buffer should be utilized. |  | Optional: \{\} <br /> |


#### CapacityQuota



CapacityQuota limits the amount of resources that can be provisioned in the cluster
by the node autoscaler. Resources used are calculated by summing up resources
reported in the status.capacity field of each node passing the configured
label selector. When making a provisioning decision, node autoscaler will
take all CapacityQuota objects that match the labels of the upcoming node.
If provisioning that node would exceed any of the matching quotas, node
autoscaler will not provision it. Quotas are best-effort, and it is possible
that in rare circumstances node autoscaler will exceed them, for example
due to stale caches.
More info: https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/proposals/granular-resource-limits.md



_Appears in:_
- [CapacityQuotaList](#capacityquotalist)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `autoscaling.x-k8s.io/v1alpha1` | | |
| `kind` _string_ | `CapacityQuota` | | |
| `kind` _string_ | Kind is a string value representing the REST resource this object represents.<br />Servers may infer this from the endpoint the client submits requests to.<br />Cannot be updated.<br />In CamelCase.<br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds |  | Optional: \{\} <br /> |
| `apiVersion` _string_ | APIVersion defines the versioned schema of this representation of an object.<br />Servers should convert recognized schemas to the latest internal value, and<br />may reject unrecognized values.<br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources |  | Optional: \{\} <br /> |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.36/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  | Optional: \{\} <br /> |
| `spec` _[CapacityQuotaSpec](#capacityquotaspec)_ | spec defines the desired state of CapacityQuota |  | Required: \{\} <br /> |
| `status` _[CapacityQuotaStatus](#capacityquotastatus)_ | status defines the observed state of CapacityQuota |  | Optional: \{\} <br /> |


#### CapacityQuotaLimits



CapacityQuotaLimits define quota limits.



_Appears in:_
- [CapacityQuotaSpec](#capacityquotaspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `resources` _[ResourceList](#resourcelist)_ | Resources define resource limits of this quota.<br />Currently supported built-in resources: cpu, memory. Additionally,<br />nodes key can be used to limit the number of existing nodes.<br />All resource quantities must be non-negative integers. Binary and decimal units are allowed as long as the quantity<br />can be converted to an integer.<br />Example allowed quantities: "32Gi", "2k", "10"<br />Example invalid quantities: "3.67Gi", "500m", "0.3"<br />**Caveat**: milli quantities are not supported even if they represent an integer, for example: "1000m".<br />Node autoscaler implementations and cloud providers can support custom<br />resources, such as GPU. |  | MaxProperties: 20 <br />Type: object <br />Required: \{\} <br /> |


#### CapacityQuotaList



CapacityQuotaList contains a list of CapacityQuota





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `autoscaling.x-k8s.io/v1alpha1` | | |
| `kind` _string_ | `CapacityQuotaList` | | |
| `kind` _string_ | Kind is a string value representing the REST resource this object represents.<br />Servers may infer this from the endpoint the client submits requests to.<br />Cannot be updated.<br />In CamelCase.<br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds |  | Optional: \{\} <br /> |
| `apiVersion` _string_ | APIVersion defines the versioned schema of this representation of an object.<br />Servers should convert recognized schemas to the latest internal value, and<br />may reject unrecognized values.<br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources |  | Optional: \{\} <br /> |
| `metadata` _[ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.36/#listmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `items` _[CapacityQuota](#capacityquota) array_ |  |  |  |


#### CapacityQuotaSpec



CapacityQuotaSpec defines the desired state of CapacityQuota



_Appears in:_
- [CapacityQuota](#capacityquota)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `selector` _[LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.36/#labelselector-v1-meta)_ | Selector is a label selector selecting the nodes to which the quota applies.<br />Empty or nil selector matches all nodes. |  | Optional: \{\} <br /> |
| `limits` _[CapacityQuotaLimits](#capacityquotalimits)_ | Limits define quota limits. |  | Required: \{\} <br /> |


#### CapacityQuotaStatus



CapacityQuotaStatus defines the observed state of CapacityQuota.



_Appears in:_
- [CapacityQuota](#capacityquota)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `used` _[CapacityQuotaUsage](#capacityquotausage)_ | Used shows the current usage of the quota. |  | Optional: \{\} <br /> |
| `conditions` _[Condition](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.36/#condition-v1-meta) array_ | Conditions provide a standard mechanism for reporting the quota's state.<br />Cluster Autoscaler manages cluster-autoscaler.kubernetes.io/valid condition, and will enforce<br />the quota only if the status of the condition is True. Note that this condition is not considered a part<br />of the public API. |  | Optional: \{\} <br /> |
| `observedGeneration` _integer_ | ObservedGeneration is the last generation observed by the controller. |  | Optional: \{\} <br /> |


#### CapacityQuotaUsage



CapacityQuotaUsage shows the current usage of the quota.



_Appears in:_
- [CapacityQuotaStatus](#capacityquotastatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `resources` _[ResourceList](#resourcelist)_ | Resources shows the current usage of the resources defined in the quota limits. |  |  |


#### LocalObjectRef



LocalObjectRef contains the name of the object being referred to.



_Appears in:_
- [CapacityBufferSpec](#capacitybufferspec)
- [CapacityBufferStatus](#capacitybufferstatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name of the object. |  | MinLength: 1 <br />Required: \{\} <br /> |


#### ResourceList

_Underlying type:_ _[Quantity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.36/#quantity-resource-api)_

ResourceList is a set of (resource name, quantity) pairs.



_Appears in:_
- [CapacityQuotaLimits](#capacityquotalimits)
- [CapacityQuotaUsage](#capacityquotausage)



#### ResourceName

_Underlying type:_ _string_

ResourceName is the name identifying a resource mirroring k8s.io/api/core/v1.ResourceName.



_Appears in:_
- [ResourceList](#resourcelist)

| Field | Description |
| --- | --- |
| `cpu` | ResourceCPU - CPU, in cores. (500m = .5 cores)<br /> |
| `memory` | ResourceMemory - memory in bytes. (500Gi = 500GiB = 500 * 1024 * 1024 * 1024)<br /> |
| `nodes` | ResourceNodes - number of nodes, in units.<br /> |


#### ScalableRef



ScalableRef contains name, kind and API group of an object that can be scaled.



_Appears in:_
- [CapacityBufferSpec](#capacitybufferspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiGroup` _string_ | APIGroup of the scalable object.<br />Empty string for the core API group |  | Optional: \{\} <br /> |
| `kind` _string_ | Kind of the scalable object (e.g., "Deployment", "StatefulSet"). |  | MinLength: 1 <br />Required: \{\} <br /> |
| `name` _string_ | Name of the scalable object. |  | MinLength: 1 <br />Required: \{\} <br /> |



## autoscaling.x-k8s.io/v1beta1

Package v1beta1 contains the v1beta1 API for the autoscaling.x-k8s.io group.
This API group defines custom resources used by the Cluster Autoscaler
for managing buffer capacity.

### Resource Types
- [CapacityBuffer](#capacitybuffer)



#### CapacityBuffer



CapacityBuffer is the configuration that an autoscaler can use to provision buffer capacity within a cluster.
This buffer is represented by placeholder pods that trigger the Cluster Autoscaler to scale up nodes in advance,
ensuring that there is always spare capacity available to handle sudden workload spikes or to speed up scaling events.



_Appears in:_
- [CapacityBufferList](#capacitybufferlist)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `autoscaling.x-k8s.io/v1beta1` | | |
| `kind` _string_ | `CapacityBuffer` | | |
| `kind` _string_ | Kind is a string value representing the REST resource this object represents.<br />Servers may infer this from the endpoint the client submits requests to.<br />Cannot be updated.<br />In CamelCase.<br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds |  | Optional: \{\} <br /> |
| `apiVersion` _string_ | APIVersion defines the versioned schema of this representation of an object.<br />Servers should convert recognized schemas to the latest internal value, and<br />may reject unrecognized values.<br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources |  | Optional: \{\} <br /> |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.36/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[CapacityBufferSpec](#capacitybufferspec)_ | Spec defines the desired characteristics of the buffer. |  | Required: \{\} <br /> |
| `status` _[CapacityBufferStatus](#capacitybufferstatus)_ | Status represents the current state of the buffer and its readiness for autoprovisioning. |  | Optional: \{\} <br /> |




#### CapacityBufferSpec



CapacityBufferSpec defines the desired state of CapacityBuffer.



_Appears in:_
- [CapacityBuffer](#capacitybuffer)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `provisioningStrategy` _string_ | ProvisioningStrategy defines how the buffer is utilized.<br />"buffer.x-k8s.io/active-capacity" is the default strategy, where the buffer actively scales up the cluster by creating placeholder pods. | buffer.x-k8s.io/active-capacity | Optional: \{\} <br /> |
| `podTemplateRef` _[LocalObjectRef](#localobjectref)_ | PodTemplateRef is a reference to a PodTemplate resource in the same namespace<br />that declares the shape of a single chunk of the buffer. The pods created<br />from this template will be used as placeholder pods for the buffer capacity.<br />Exactly one of `podTemplateRef`, `scalableRef` should be specified. |  | Optional: \{\} <br /> |
| `scalableRef` _[ScalableRef](#scalableref)_ | ScalableRef is a reference to an object of a kind that has a scale subresource<br />and specifies its label selector field. This allows the CapacityBuffer to<br />manage the buffer by scaling an existing scalable resource.<br />Exactly one of `podTemplateRef`, `scalableRef` should be specified. |  | Optional: \{\} <br /> |
| `replicas` _integer_ | Replicas defines the desired number of buffer chunks to provision.<br />If neither `replicas` nor `percentage` is set, as many chunks as fit within<br />defined resource limits (if any) will be created. If both are set, the maximum<br />of the two will be used. |  | ExclusiveMinimum: false <br />Minimum: 0 <br />Optional: \{\} <br /> |
| `percentage` _integer_ | Percentage defines the desired buffer capacity as a percentage of the<br />`scalableRef`'s current replicas. This is only applicable if `scalableRef` is set.<br />The absolute number of replicas is calculated from the percentage by rounding up to a minimum of 1.<br />For example, if `scalableRef` has 10 replicas and `percentage` is 20, 2 buffer chunks will be created. |  | ExclusiveMinimum: false <br />Minimum: 0 <br />Optional: \{\} <br /> |
| `limits` _[ResourceList](#resourcelist)_ | Limits, if specified, will limit the number of chunks created for this buffer<br />based on total resource requests (e.g., CPU, memory). If there are no other<br />limitations for the number of chunks (i.e., `replicas` or `percentage` are not set),<br />this will be used to create as many chunks as fit into these limits. |  | Optional: \{\} <br /> |


#### CapacityBufferStatus



CapacityBufferStatus defines the observed state of CapacityBuffer.



_Appears in:_
- [CapacityBuffer](#capacitybuffer)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `podTemplateRef` _[LocalObjectRef](#localobjectref)_ | PodTemplateRef is the observed reference to the PodTemplate that was used<br />to provision the buffer. If this field is not set, and the `conditions`<br />indicate an error, it provides details about the error state. |  | Optional: \{\} <br /> |
| `replicas` _integer_ | Replicas is the actual number of buffer chunks currently provisioned. |  | Optional: \{\} <br /> |
| `podTemplateGeneration` _integer_ | PodTemplateGeneration is the observed generation of the PodTemplate, used<br />to determine if the status is up-to-date with the desired `spec.podTemplateRef`. |  | Optional: \{\} <br /> |
| `conditions` _[Condition](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.36/#condition-v1-meta) array_ | Conditions provide a standard mechanism for reporting the buffer's state.<br />The "Ready" condition indicates if the buffer is successfully provisioned<br />and active. Other conditions may report on various aspects of the buffer's<br />health and provisioning process. |  | Optional: \{\} <br /> |
| `provisioningStrategy` _string_ | ProvisioningStrategy defines how the buffer should be utilized. |  | Optional: \{\} <br /> |


#### Detail

_Underlying type:_ _string_

Detail is limited to 32768 characters.

_Validation:_
- MaxLength: 32768

_Appears in:_
- [ProvisioningRequestStatus](#provisioningrequeststatus)



#### LocalObjectRef



LocalObjectRef contains the name of the object being referred to.



_Appears in:_
- [CapacityBufferSpec](#capacitybufferspec)
- [CapacityBufferStatus](#capacitybufferstatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name of the object. |  | MinLength: 1 <br />Required: \{\} <br /> |


#### Parameter

_Underlying type:_ _string_

Parameter is limited to 255 characters.

_Validation:_
- MaxLength: 255

_Appears in:_
- [ProvisioningRequestSpec](#provisioningrequestspec)



#### PodSet



PodSet represents one group of pods for Provisioning Request to provision capacity.



_Appears in:_
- [ProvisioningRequestSpec](#provisioningrequestspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `podTemplateRef` _[Reference](#reference)_ | PodTemplateRef is a reference to a PodTemplate object that is representing pods<br />that will consume this reservation (must be within the same namespace).<br />Users need to make sure that the  fields relevant to scheduler (e.g. node selector tolerations)<br />are consistent between this template and actual pods consuming the Provisioning Request. |  | Required: \{\} <br /> |
| `count` _integer_ | Count contains the number of pods that will be created with a given<br />template. |  | Minimum: 1 <br /> |


#### ProvisioningRequest



ProvisioningRequest is a way to express additional capacity
that we would like to provision in the cluster. Cluster Autoscaler
can use this information in its calculations and signal if the capacity
is available in the cluster or actively add capacity if needed.



_Appears in:_
- [ProvisioningRequestList](#provisioningrequestlist)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `kind` _string_ | Kind is a string value representing the REST resource this object represents.<br />Servers may infer this from the endpoint the client submits requests to.<br />Cannot be updated.<br />In CamelCase.<br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds |  | Optional: \{\} <br /> |
| `apiVersion` _string_ | APIVersion defines the versioned schema of this representation of an object.<br />Servers should convert recognized schemas to the latest internal value, and<br />may reject unrecognized values.<br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources |  | Optional: \{\} <br /> |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.36/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  | Optional: \{\} <br /> |
| `spec` _[ProvisioningRequestSpec](#provisioningrequestspec)_ | Spec contains specification of the ProvisioningRequest object.<br />More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status.<br />The spec is immutable, to make changes to the request users are expected to delete an existing<br />and create a new object with the corrected fields. |  | Required: \{\} <br /> |
| `status` _[ProvisioningRequestStatus](#provisioningrequeststatus)_ | Status of the ProvisioningRequest. CA constantly reconciles this field. |  | Optional: \{\} <br /> |




#### ProvisioningRequestSpec



ProvisioningRequestSpec is a specification of additional pods for which we
would like to provision additional resources in the cluster.



_Appears in:_
- [ProvisioningRequest](#provisioningrequest)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `podSets` _[PodSet](#podset) array_ | PodSets lists groups of pods for which we would like to provision<br />resources. |  | MaxItems: 32 <br />MinItems: 1 <br />Required: \{\} <br /> |
| `provisioningClassName` _string_ | ProvisioningClassName describes the different modes of provisioning the resources.<br />Currently there is no support for 'ProvisioningClass' objects.<br />Supported values:<br />* check-capacity.autoscaling.x-k8s.io - check if current cluster state can fullfil this request,<br />  do not reserve the capacity. Users should provide a reference to a valid PodTemplate object.<br />  CA will check if there is enough capacity in cluster to fulfill the request and put<br />  the answer in 'CapacityAvailable' condition.<br />* best-effort-atomic-scale-up.autoscaling.x-k8s.io - provision the resources in an atomic manner.<br />  Users should provide a reference to a valid PodTemplate object.<br />  CA will try to create the VMs in an atomic manner, clean any partially provisioned VMs<br />  and re-try the operation in a exponential back-off manner. Users can configure the timeout<br />  duration after which the request will fail by 'ValidUntilSeconds' key in 'Parameters'.<br />  CA will set 'Failed=true' or 'Provisioned=true' condition according to the outcome.<br />* ... - potential other classes that are specific to the cloud providers.<br />'kubernetes.io' suffix is reserved for the modes defined in Kubernetes projects. |  | MaxLength: 253 <br />Pattern: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$` <br />Required: \{\} <br /> |
| `parameters` _object (keys:string, values:[Parameter](#parameter))_ | Parameters contains all other parameters classes may require.<br />'best-effort-atomic-scale-up.autoscaling.x-k8s.io' supports 'ValidUntilSeconds' parameter, which should contain<br /> a string denoting duration for which we should retry (measured since creation fo the CR). |  | MaxProperties: 100 <br />Optional: \{\} <br /> |


#### ProvisioningRequestStatus



ProvisioningRequestStatus represents the status of the resource reservation.



_Appears in:_
- [ProvisioningRequest](#provisioningrequest)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `conditions` _[Condition](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.36/#condition-v1-meta) array_ | Conditions represent the observations of a Provisioning Request's<br />current state. Those will contain information whether the capacity<br />was found/created or if there were any issues. The condition types<br />may differ between different provisioning classes. |  | Optional: \{\} <br /> |
| `provisioningClassDetails` _object (keys:string, values:[Detail](#detail))_ | ProvisioningClassDetails contains all other values custom provisioning classes may<br />want to pass to end users. |  | MaxProperties: 64 <br />Optional: \{\} <br /> |


#### Reference



Reference represents reference to an object within the same namespace.



_Appears in:_
- [PodSet](#podset)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name of the referenced object.<br />More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names#names |  | MaxLength: 253 <br />Pattern: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$` <br />Required: \{\} <br /> |


#### ResourceList

_Underlying type:_ _[Quantity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.36/#quantity-resource-api)_

ResourceList is a set of (resource name, quantity) pairs.
This is mirroring k8s.io/api/core/v1.ResourceList to avoid direct dependency.



_Appears in:_
- [CapacityBufferSpec](#capacitybufferspec)



#### ResourceName

_Underlying type:_ _string_

ResourceName is the name identifying a resource mirroring k8s.io/api/core/v1.ResourceName.



_Appears in:_
- [ResourceList](#resourcelist)



#### ScalableRef



ScalableRef contains name, kind and API group of an object that can be scaled.



_Appears in:_
- [CapacityBufferSpec](#capacitybufferspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiGroup` _string_ | APIGroup of the scalable object.<br />Empty string for the core API group |  | Optional: \{\} <br /> |
| `kind` _string_ | Kind of the scalable object (e.g., "Deployment", "StatefulSet"). |  | MinLength: 1 <br />Required: \{\} <br /> |
| `name` _string_ | Name of the scalable object. |  | MinLength: 1 <br />Required: \{\} <br /> |


