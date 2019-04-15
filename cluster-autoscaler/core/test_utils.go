/*
Copyright 2019 The Kubernetes Authors.

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

package core

import (
	"k8s.io/autoscaler/cluster-autoscaler/processors"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroups"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroupset"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
)

// NewTestProcessors returns a set of simple processors for use in tests.
func NewTestProcessors() *processors.AutoscalingProcessors {
	return &processors.AutoscalingProcessors{
		PodListProcessor:       NewFilterOutSchedulablePodListProcessor(),
		NodeGroupListProcessor: &nodegroups.NoOpNodeGroupListProcessor{},
		NodeGroupSetProcessor:  &nodegroupset.BalancingNodeGroupSetProcessor{},
		// TODO(bskiba): change scale up test so that this can be a NoOpProcessor
		ScaleUpStatusProcessor:     &status.EventingScaleUpStatusProcessor{},
		ScaleDownStatusProcessor:   &status.NoOpScaleDownStatusProcessor{},
		AutoscalingStatusProcessor: &status.NoOpAutoscalingStatusProcessor{},
		NodeGroupManager:           nodegroups.NewDefaultNodeGroupManager(),
	}
}
