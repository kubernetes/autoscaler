/*
Copyright 2018 The Kubernetes Authors.

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

package nodegroups

import (
	"reflect"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
)

// NodeGroupManager is responsible for creating/deleting node groups.
type NodeGroupManager interface {
	// CreateNodeGroup creates node group and returns all of the results.
	// Creating a node group may result in multiple node group creations, as the behavior is
	// cloud provider dependent.
	CreateNodeGroup(context *context.AutoscalingContext, nodeGroup cloudprovider.NodeGroup) (CreateNodeGroupResult, errors.AutoscalerError)

	// CreateNodeGroupAsync similar to CreateNodeGroup method but creates node group asynchronously.
	// Immediately returns upcoming node group that may be used for scale ups and scale up simulations.
	CreateNodeGroupAsync(context *context.AutoscalingContext, nodeGroup cloudprovider.NodeGroup, nodeGroupInitializer AsyncNodeGroupInitializer) (CreateNodeGroupResult, errors.AutoscalerError)

	RemoveUnneededNodeGroups(context *context.AutoscalingContext) (removedNodeGroups []cloudprovider.NodeGroup, err error)

	CleanUp()
}

// AsyncNodeGroupCreationResult captures result of NodeGroupManager.CreateNodeGroupAsync call.
type AsyncNodeGroupCreationResult struct {
	TargetSizes    map[string]int
	CreationResult CreateNodeGroupResult
	Error          errors.AutoscalerError
}

// AsyncNodeGroupInitializer is responsible for initializing asynchronously created node groups.
// In most cases node group initialization should involve scaling up newly created node groups.
type AsyncNodeGroupInitializer interface {
	// InitializeNodeGroup initializes asynchronously created node group.
	InitializeNodeGroup(result AsyncNodeGroupCreationResult)
}

// NoOpNodeGroupManager is a no-op implementation of NodeGroupManager.
// It does not remove any node groups and its CreateNodeGroup method always returns an error.
// To be used together with NoOpNodeGroupListProcessor.
type NoOpNodeGroupManager struct {
}

// CreateNodeGroupResult result captures result of successful NodeGroupManager.CreateNodeGroup call.
type CreateNodeGroupResult struct {
	// Main created node group, matching the requested node group passed to CreateNodeGroup call
	MainCreatedNodeGroup cloudprovider.NodeGroup

	// List of extra node groups created by CreateNodeGroup call. Non-empty if due manager specific
	// constraints creating one node group requires creating other ones (e.g. matching node group
	// must exist in each zone for multizonal deployments)
	ExtraCreatedNodeGroups []cloudprovider.NodeGroup
}

// AllCreatedNodeGroups returns all created node groups.
func (r CreateNodeGroupResult) AllCreatedNodeGroups() []cloudprovider.NodeGroup {
	var result []cloudprovider.NodeGroup
	if r.MainCreatedNodeGroup != nil && !reflect.ValueOf(r.MainCreatedNodeGroup).IsNil() {
		result = append(result, r.MainCreatedNodeGroup)
	}
	result = append(result, r.ExtraCreatedNodeGroups...)
	return result
}

// CreateNodeGroup always returns internal error. It must not be called on NoOpNodeGroupManager.
func (*NoOpNodeGroupManager) CreateNodeGroup(context *context.AutoscalingContext, nodeGroup cloudprovider.NodeGroup) (CreateNodeGroupResult, errors.AutoscalerError) {
	return CreateNodeGroupResult{}, errors.NewAutoscalerError(errors.InternalError, "not implemented")
}

// CreateNodeGroupAsync always returns internal error. It must not be called on NoOpNodeGroupManager.
func (*NoOpNodeGroupManager) CreateNodeGroupAsync(context *context.AutoscalingContext, nodeGroup cloudprovider.NodeGroup, nodeGroupInitializer AsyncNodeGroupInitializer) (CreateNodeGroupResult, errors.AutoscalerError) {
	return CreateNodeGroupResult{}, errors.NewAutoscalerError(errors.InternalError, "not implemented")
}

// RemoveUnneededNodeGroups does nothing in NoOpNodeGroupManager
func (*NoOpNodeGroupManager) RemoveUnneededNodeGroups(context *context.AutoscalingContext) (removedNodeGroups []cloudprovider.NodeGroup, err error) {
	return nil, nil
}

// CleanUp does nothing in NoOpNodeGroupManager
func (*NoOpNodeGroupManager) CleanUp() {}

// NewDefaultNodeGroupManager creates an instance of NodeGroupManager.
func NewDefaultNodeGroupManager() NodeGroupManager {
	return &NoOpNodeGroupManager{}
}
