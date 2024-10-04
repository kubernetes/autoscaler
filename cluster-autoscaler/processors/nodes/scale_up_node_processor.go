/*
Copyright 2024 The Kubernetes Authors.

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

package nodes

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/context"
)

// ScaleUpNodeProcessor is used to process the nodes of the cluster after cloud provider is refreshed and before scale-up.
type ScaleUpNodeProcessor interface {
	// Process processes the nodes of the cluster after cloud provider is refreshed and before scale-up.
	Process(context *context.AutoscalingContext, allNodes, readyNodes []*apiv1.Node) ([]*apiv1.Node, []*apiv1.Node)
	// CleanUp is called at CA termination
	CleanUp()
}

// NewDefaultScaleUpNodeProcessor creates a default instance of ScaleUpNodeProcessor that does not alter the nodes when processing.
func NewDefaultScaleUpNodeProcessor() ScaleUpNodeProcessor {
	return &defaultScaleUpNodeProcessor{}
}

type defaultScaleUpNodeProcessor struct{}

func (p *defaultScaleUpNodeProcessor) Process(context *context.AutoscalingContext, allNodes, readyNodes []*apiv1.Node) ([]*apiv1.Node, []*apiv1.Node) {
	return allNodes, readyNodes
}

func (p *defaultScaleUpNodeProcessor) CleanUp() {}
