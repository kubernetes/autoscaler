/*
Copyright 2021 The Kubernetes Authors.

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

package customresources

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
)

// CustomResourceTarget contains information about targeted custom resources
type CustomResourceTarget struct {
	// ResourceType is a type of targeted resources
	ResourceType string
	// ResourceCount is a count of targeted resources
	ResourceCount int64
}

// CustomResourcesProcessor is interface defining handling custom resources
type CustomResourcesProcessor interface {
	// FilterOutNodesWithUnreadyResources removes nodes that should have a custom resource, but don't have
	// it in allocatable from ready nodes list and updates their status to unready on all nodes list.
	FilterOutNodesWithUnreadyResources(context *context.AutoscalingContext, allNodes, readyNodes []*apiv1.Node) ([]*apiv1.Node, []*apiv1.Node)
	// GetNodeResourceTargets returns mapping of resource names to their targets.
	GetNodeResourceTargets(context *context.AutoscalingContext, node *apiv1.Node, nodeGroup cloudprovider.NodeGroup) ([]CustomResourceTarget, errors.AutoscalerError)
	// CleanUp cleans up processor's internal structures.
	CleanUp()
}

// NewDefaultCustomResourcesProcessor returns a default instance of CustomResourcesProcessor.
func NewDefaultCustomResourcesProcessor() CustomResourcesProcessor {
	return &GpuCustomResourcesProcessor{}
}
