/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cloudprovider

import (
	"fmt"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"
)

// CloudProvider contains configuration info and functions for interacting with
// cloud provider (GCE, AWS, etc).
type CloudProvider interface {
	// Name returns name of the cloud provider.
	Name() string

	// NodeGroups returns all node groups configured for this cloud provider.
	NodeGroups() []NodeGroup

	// NodeGroupForNode returns the node group for the given node, nil if the node
	// should not be processed by cluster autoscaler, or non-nil error if such
	// occurred. Must be implemented.
	NodeGroupForNode(*apiv1.Node) (NodeGroup, error)

	// Pricing returns pricing model for this cloud provider or error if not available.
	// Implementation optional.
	Pricing() (PricingModel, errors.AutoscalerError)

	// GetAvailableMachineTypes get all machine types that can be requested from the cloud provider.
	// Implementation optional.
	GetAvailableMachineTypes() ([]string, error)

	// NewNodeGroup builds a theoretical node group based on the node definition provided. The node group is not automatically
	// created on the cloud provider side. The node group is not returned by NodeGroups() until it is created.
	// Implementation optional.
	NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string,
		taints []apiv1.Taint, extraResources map[string]resource.Quantity) (NodeGroup, error)

	// GetResourceLimiter returns struct containing limits (max, min) for resources (cores, memory etc.).
	GetResourceLimiter() (*ResourceLimiter, error)

	// GPULabel returns the label added to nodes with GPU resource.
	GPULabel() string

	// GetAvailableGPUTypes return all available GPU types cloud provider supports.
	GetAvailableGPUTypes() map[string]struct{}

	// Cleanup cleans up open resources before the cloud provider is destroyed, i.e. go routines etc.
	Cleanup() error

	// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
	// In particular the list of node groups returned by NodeGroups can change as a result of CloudProvider.Refresh().
	Refresh() error
}

// ErrNotImplemented is returned if a method is not implemented.
var ErrNotImplemented = errors.NewAutoscalerError(errors.InternalError, "Not implemented")

// ErrAlreadyExist is returned if a method is not implemented.
var ErrAlreadyExist = errors.NewAutoscalerError(errors.InternalError, "Already exist")

// ErrIllegalConfiguration is returned when trying to create NewNodeGroup with
// configuration that is not supported by cloudprovider.
var ErrIllegalConfiguration = errors.NewAutoscalerError(errors.InternalError, "Configuration not allowed by cloud provider")

// NodeGroup contains configuration info and functions to control a set
// of nodes that have the same capacity and set of labels.
type NodeGroup interface {
	// MaxSize returns maximum size of the node group.
	MaxSize() int

	// MinSize returns minimum size of the node group.
	MinSize() int

	// TargetSize returns the current target size of the node group. It is possible that the
	// number of nodes in Kubernetes is different at the moment but should be equal
	// to Size() once everything stabilizes (new nodes finish startup and registration or
	// removed nodes are deleted completely). Implementation required.
	TargetSize() (int, error)

	// IncreaseSize increases the size of the node group. To delete a node you need
	// to explicitly name it and use DeleteNode. This function should wait until
	// node group size is updated. Implementation required.
	IncreaseSize(delta int) error

	// DeleteNodes deletes nodes from this node group. Error is returned either on
	// failure or if the given node doesn't belong to this node group. This function
	// should wait until node group size is updated. Implementation required.
	DeleteNodes([]*apiv1.Node) error

	// DecreaseTargetSize decreases the target size of the node group. This function
	// doesn't permit to delete any existing node and can be used only to reduce the
	// request for new nodes that have not been yet fulfilled. Delta should be negative.
	// It is assumed that cloud provider will not delete the existing nodes when there
	// is an option to just decrease the target. Implementation required.
	DecreaseTargetSize(delta int) error

	// Id returns an unique identifier of the node group.
	Id() string

	// Debug returns a string containing all information regarding this node group.
	Debug() string

	// Nodes returns a list of all nodes that belong to this node group.
	// It is required that Instance objects returned by this method have Id field set.
	// Other fields are optional.
	Nodes() ([]Instance, error)

	// TemplateNodeInfo returns a schedulernodeinfo.NodeInfo structure of an empty
	// (as if just started) node. This will be used in scale-up simulations to
	// predict what would a new node look like if a node group was expanded. The returned
	// NodeInfo is expected to have a fully populated Node object, with all of the labels,
	// capacity and allocatable information as well as all pods that are started on
	// the node by default, using manifest (most likely only kube-proxy). Implementation optional.
	TemplateNodeInfo() (*schedulernodeinfo.NodeInfo, error)

	// Exist checks if the node group really exists on the cloud provider side. Allows to tell the
	// theoretical node group from the real one. Implementation required.
	Exist() bool

	// Create creates the node group on the cloud provider side. Implementation optional.
	Create() (NodeGroup, error)

	// Delete deletes the node group on the cloud provider side.
	// This will be executed only for autoprovisioned node groups, once their size drops to 0.
	// Implementation optional.
	Delete() error

	// Autoprovisioned returns true if the node group is autoprovisioned. An autoprovisioned group
	// was created by CA and can be deleted when scaled to 0.
	Autoprovisioned() bool
}

// Instance represents a cloud-provider node. The node does not necessarily map to k8s node
// i.e it does not have to be registered in k8s cluster despite being returned by NodeGroup.Nodes()
// method. Also it is sane to have Instance object for nodes which are being created or deleted.
type Instance struct {
	// Id is instance id.
	Id string
	// Status represents status of node. (Optional)
	Status *InstanceStatus
}

// InstanceStatus represents instance status.
type InstanceStatus struct {
	// State tells if instance is running, being created or being deleted
	State InstanceState
	// ErrorInfo is not nil if there is error condition related to instance.
	// E.g instance cannot be created.
	ErrorInfo *InstanceErrorInfo
}

// InstanceState tells if instance is running, being created or being deleted
type InstanceState int

const (
	// InstanceRunning means instance is running
	InstanceRunning InstanceState = 1
	// InstanceCreating means instance is being created
	InstanceCreating InstanceState = 2
	// InstanceDeleting means instance is being deleted
	InstanceDeleting InstanceState = 3
)

// InstanceErrorInfo provides information about error condition on instance
type InstanceErrorInfo struct {
	// ErrorClass tells what is class of error on instance
	ErrorClass InstanceErrorClass
	// ErrorCode is cloud-provider specific error code for error condition
	ErrorCode string
	// ErrorMessage is human readable description of error condition
	ErrorMessage string
}

// InstanceErrorClass defines class of error condition
type InstanceErrorClass int

const (
	// OutOfResourcesErrorClass means that error is related to lack of resources (e.g. due to
	// stockout or quota-exceeded situation)
	OutOfResourcesErrorClass InstanceErrorClass = 1
	// OtherErrorClass means some non-specific error situation occurred
	OtherErrorClass InstanceErrorClass = 99
)

func (c InstanceErrorClass) String() string {
	switch c {
	case OutOfResourcesErrorClass:
		return "OutOfResource"
	case OtherErrorClass:
		return "Other"
	default:
		return fmt.Sprintf("%d", c)
	}
}

// PricingModel contains information about the node price and how it changes in time.
type PricingModel interface {
	// NodePrice returns a price of running the given node for a given period of time.
	// All prices returned by the structure should be in the same currency.
	NodePrice(node *apiv1.Node, startTime time.Time, endTime time.Time) (float64, error)

	// PodPrice returns a theoretical minimum price of running a pod for a given
	// period of time on a perfectly matching machine.
	PodPrice(pod *apiv1.Pod, startTime time.Time, endTime time.Time) (float64, error)
}

const (
	// ResourceNameCores is string name for cores. It's used by ResourceLimiter.
	ResourceNameCores = "cpu"
	// ResourceNameMemory is string name for memory. It's used by ResourceLimiter.
	// Memory should always be provided in bytes.
	ResourceNameMemory = "memory"
)

// IsGpuResource checks if given resource name point denotes a gpu type
func IsGpuResource(resourceName string) bool {
	// hack: we assume anything which is not cpu/memory to be a gpu.
	// we are not getting anything more that a map string->limits from the user
	return resourceName != ResourceNameCores && resourceName != ResourceNameMemory
}

// ContainsGpuResources returns true iff given list contains any resource name denoting a gpu type
func ContainsGpuResources(resources []string) bool {
	for _, resource := range resources {
		if IsGpuResource(resource) {
			return true
		}
	}
	return false
}
