# Provisioning Request CRD

author: kisieland

## Background

Currently CA does not provide any way to express that a group of pods would like
to have a capacity available.
This is caused by the fact that each CA loop picks a group of unschedulable pods
and works on provisioning capacity for them, meaning that the grouping is random
(as it depends on the kube-scheduler and CA loop interactions).
This is especially problematic in couple of cases:

  - Users would like to have all-or-nothing semantics for their workloads.
    Currently CA will try to provision this capacity and if it is partially
    successful it will leave it in cluster until user removes the workload.
  - Users would like to lower e2e scale-up latency for huge scale-ups (100
    nodes+). Due to CA nature and kube-scheduler throughput, CA will create
    partial scale-ups, e.g. `0->200->400->600` rather than one `0->600`. This
    significantly increases the e2e latency as there is non-negligible time tax
    on each scale-up operation.

## Proposal

### High level

Provisioning Request (abbr. ProvReq) is a new namespaced Custom Resource that
aims to allow users to ask CA for capacity for groups of pods.
It allows users to express the fact that group of pods is connected and should
be threated as one entity.
This AEP proposes an API that can have multiple provisioning classes and can be
extended by cloud provider specific ones.
This object is meant as one-shot request to CA, so that if CA fails to provision
the capacity it is up to users to retry (such retry functionality can be added
later on).

### ProvisioningRequest CRD

The following code snippets assume [kubebuilder](https://book.kubebuilder.io/)
is used to generate the CRD:

```go
// ProvisioningRequest is a way to express additional capacity
// that we would like to provision in the cluster. Cluster Autoscaler
// can use this information in its calculations and signal if the capacity
// is available in the cluster or actively add capacity if needed.
type ProvisioningRequest struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	//
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Spec contains specification of the ProvisioningRequest object.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status.
	//
	// +kubebuilder:validation:Required
	Spec ProvisioningRequestSpec `json:"spec"`
	// Status of the ProvisioningRequest. CA constantly reconciles this field.
	//
	// +optional
	Status ProvisioningRequestStatus `json:"status,omitempty"`
}

// ProvisioningRequestList is a object for list of ProvisioningRequest.
type ProvisioningRequestList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata.
	//
	// +optional
	metav1.ListMeta `json:"metadata"`
	// Items, list of ProvisioningRequest returned from API.
	//
	// +optional
	Items []ProvisioningRequest `json:"items"`
}

// ProvisioningRequestSpec is a specification of additional pods for which we
// would like to provision additional resources in the cluster.
type ProvisioningRequestSpec struct {
	// PodSets lists groups of pods for which we would like to provision
	// resources.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=32
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	PodSets []PodSet `json:"podSets"`

	// ProvisioningClass describes the different modes of provisioning the resources.
	// Supported values:
	// * check-capacity.kubernetes.io - check if current cluster state can fullfil this request,
	//   do not reserve the capacity.
	// * atomic-scale-up.kubernetes.io - provision the resources in an atomic manner
    // * ... - potential other classes that are specific to the cloud providers
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	ProvisioningClass string `json:"provisioningClass"`

	// Parameters contains all other parameters custom classes may require.
	//
	// +optional
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	Parameters map[string]string `json:"Parameters"`
}

type PodSet struct {
	// PodTemplateRef is a reference to a PodTemplate object that is representing pods
	// that will consume this reservation (must be within the same namespace).
	// Users need to make sure that the  fields relevant to scheduler (e.g. node selector tolerations)
	// are consistent between this template and actual pods consuming the Provisioning Request.
	//
	// +kubebuilder:validation:Required
	PodTemplateRef Reference `json:"podTemplateRef"`
	// Count contains the number of pods that will be created with a given
	// template.
	//
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=16384
	Count int32 `json:"count"`
}

type Reference struct {
	// Name of the referenced object.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names#names
	//
	// +kubebuilder:validation:Required
	Name string `json:"name,omitempty"`
}

// ProvisioningRequestStatus represents the status of the resource reservation.
type ProvisioningRequestStatus struct {
	// Conditions represent the observations of a Provisioning Request's
	// current state. Those will contain information whether the capacity
    // was found/created or if there were any issues. The condition types
    // may differ between different provisioning classes.
	//
	// +listType=map
    // +listMapKey=type
    // +patchStrategy=merge
    // +patchMergeKey=type
    // +optional
	Conditions []metav1.Condition `json:"conditions"`

	// Statuses contains all other status values custom provisioning classes may require.
	//
	// +optional
	// +kubebuilder:validation:MaxItems=64
	Statuses map[string]string `json:"statuses"`
}
```

### Provisioning Classes

#### check-capacity.kubernetes.io class

The `check-capacity.kubernetes.io` is one-off check to verify that the in the cluster
there is enough capacity to provision given set of pods.

Note: If two of such objects are created around the same time, CA will consider
them independently and place no guards for the capacity.
Also the capacity is not reserved in any manner so it may be scaled-down.

#### atomic-scale-up.kubernetes.io class

The `atomic-scale-up.kubernetes.io` aims to provision the resources required for the
specified pods in an atomic way. The proposed logic is to:
1. Try to provision required VMs in one loop.
2. If it failed, remove the partially provisioned VMs and back-off.
3. Stop the back-off after a given duration (optional), which would be passed
   via `Parameters` field, using `ValidUntilSeconds` key and would contain string
   denoting duration for which we should retry (measured since creation fo the CR).

Note: that the VMs created in this mode are subject to the scale-down logic.
So the duration during which users need to create the Pods is equal to the
value of `--scale-down-unneeded-time` flag.

### Adding pods that consume given ProvisioningRequest

To avoid generating double scale-ups and exclude pods that are meant to consume
given capacity CA should be able to differentiate those from all other pods.
To achieve this users need to specify the following pod annotations (those are
not required in ProvReq’s template, though can be specified):

```yaml
annotations:
    "cluster-autoscaler.kubernetes.io/provisioning-class-name": "provreq-class-name"
    "cluster-autoscaler.kubernetes.io/consume-provisioning-request": "provreq-name"
```

If those are provided for the pods that consume the ProvReq with `check-capacity.kubernetes.io` class,
the CA will not provision the capacity, even if it was needed (as some other pods might have been
scheduled on it) and will result in visibility events passed to the ProvReq and pods.
If those are not passed the CA will behave normally and just provision the capacity if it needed.
Both annotation are required and CA will not work correctly if only one of them is passed.

Note: CA will match all pods with this annotation to a corresponding ProvReq and
ignore them when executing a scale-up loop (so that is up to users to make sure
that the ProvReq count is matching the number of created pods).
If the ProvReq is missing, all of the pods that consume it will be unschedulable indefinitely.

### CRD lifecycle

1.  A ProvReq will be created either by the end user or by a framework.
    At this point needed PodTemplate objects should be also created.
2.  CA will pick it up, choose a nodepool (or create a new one if NAP is
    enabled), and try to create nodes.
3.  If CA successfully creates capacity, ProvReq will receive information about
    this fact in `Conditions` field.
4.  At this moment, users can create pods in that will consume the ProvReq (in
    the same namespace), those will be scheduled on the capacity that was
    created by the CA.
5.  Once all of the pods are scheduled users can delete the ProvReq object,
    otherwise it will be garbage collected after some time.
6.  When pods finish the work and nodes become unused the CA will scale them
    down.

Note: Users can create a ProvReq and pods consuming them at the same time (in a
"fire and forget" manner), but this may result in the pods being unschedulable
and triggering user configured alerts.

### Canceling the requests

To cancel a pending Provisioning Request with atomic class, all that the users need to do is
to delete the Provisioning Request object.
After that the CA will no longer guard the nodes from deletion and proceed with standard scale-down logic.

### Conditions

The following Condition states should encode the states of the ProvReq:

  - Provisioned - VMs were created successfully (Atomic class)
  - CapacityAvailable - cluster contains enough capacity to schedule pods (Check
    class)
    * `CapacityAvailable=true` will denote that cluster contains enough capacity to schedule pods
	* `CapacityAvailable=false` will denote that cluster does not contain enough capacity to schedule pods
  - Failed - failed to create or check capacity (both classes)

The Reasons and Messages will contain more details about why the specific
condition was triggered.

Providers of the custom classes should reuse the conditions where available or create their own ones
if items from the above list cannot be used to denote a specific situation.

### CA implementation details

The proposed implementation is to handle each ProvReq in a separate scale-up
loop. This will require changes in multiple parts of CA:

1.  Listing unschedulable pods where:
      - pods that consume ProvReq need to filtered-out
      - pods that are represented by the ProvReq need to be injected (we need to
        ensure those are threated as one group by the sharding logic)
2.  Scale-up logic, which as of now has no notion atomicity and grouping of
    pods. This is simplified as the ScaleUp logic was recently put [behind an
    interface](https://github.com/kubernetes/autoscaler/pull/5597).
      - This is a place where the biggest part of the change will be made. Here
        many parts of the logic are assuming best-effort semantics and the scale
        up size is lowered in many situations:
          - Estimation logic, which stops after some time-out or number of
            pods/nodes.
          - Size limiting, which caps the scale-up to match the size
            restrictions (on node group or cluster level).
3.  Node creation, which needs to support atomic resize. Either via native cloud
    provider APIs or best effort with node removal if CA is unable to fulfill
    the scale-up.
      - This is also quite substantial change, we can provide a generic
        best-effort implementation that will try to scale up and clean-up nodes
        if it is unsuccessful, but it is up to cloud providers to integrate with
        provider specific APIs.
4.  Scale down path is not expected to change much. But users should follow
    [best
    practices](https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/FAQ.md#what-types-of-pods-can-prevent-ca-from-removing-a-node)
    to avoid CA disturbing their workloads.

## Testing

The following e2e test scenarios will be created to check whether ProvReq
handling works as expected:

1.  A new ProvReq with `check-capacity.kubernetes.io` provisioning class is created, CA
    checks if there is enough capacity in cluster to provision specified pods.
2.  A new ProvReq with `atomic-scale-up.kubernetes.io` provisioning class is created, CA
    picks an appropriate node group scales it up atomically.
3.  A new atomic ProvReq is created for which a NAP needs to provision a new
    node group. NAP creates it CA scales it atomically.
      - Here we should cover some of the different reasons why NAP may be
        required.
4.  An atomic ProvReq fails due to node group size limits and NAP CPU and/or RAM
    limits.
5.  Scalability tests.
      - Scenario in which many small ProvReqs are created (strain on the number
        of scale-up loops).
      - Scenario in which big ProvReq is created (strain on a single scale-up
        loop).

## Limitations

The current Cluster Autoscaler implementation is not taking into account [Resource Quotas](https://kubernetes.io/docs/concepts/policy/resource-quotas/). \
The current proposal is to not include handling of the Resource Quotas, but it could be added later on.

## Future Expansions

### ProvisioningClass CRD

One of the expansion of this approach is to introduce the ProvisioningClass CRD,
which follows the same approach as
[StorageClass object](https://kubernetes.io/docs/concepts/storage/storage-classes/).
Such approach would allow administrators of the cluster to introduce a list of allowed
ProvisioningClasses. Such CRD can also contain a pre set configuration, i.e.
administrators may set that `atomic-scale-up.kubernetes.io` would retry up to `2h`.

Possible CRD definition:
```go
// ProvisioningClass is a way to express provisioning classes available in the cluster.
type ProvisioningClass struct {
	// Name denotes the name of the object, which is to be used in the ProvisioningClass
	// field in Provisioning Request CRD.
	//
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Parameters contains all other parameters custom classes may require.
	//
	// +optional
	Parameters map[string]string `json:"Parameters"`
}
```
