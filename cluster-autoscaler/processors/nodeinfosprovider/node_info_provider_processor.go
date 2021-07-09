/*
Copyright 2020 The Kubernetes Authors.

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

package nodeinfosprovider

import (
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"

	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

// NodeInfoProviderProcessor is provides the initial nodeInfos set.
type NodeInfoProviderProcessor interface {
	// Process returns a map of nodeInfos for node groups.
	Process(nodes []*apiv1.Node, ctx *context.AutoscalingContext, daemonsets []*appsv1.DaemonSet, ignoredTaints taints.TaintKeySet) (map[string]*schedulerframework.NodeInfo, errors.AutoscalerError)
	// CleanUp cleans up processor's internal structures.
	CleanUp()
}

// NoOpNodeInfoProviderProcessor doesn't change nodeInfos.
type NoOpNodeInfoProviderProcessor struct {
}

// Process returns empty nodeInfos map.
func (p *NoOpNodeInfoProviderProcessor) Process(nodes []*apiv1.Node, ctx *context.AutoscalingContext, daemonsets []*appsv1.DaemonSet, ignoredTaints taints.TaintKeySet) (map[string]*schedulerframework.NodeInfo, errors.AutoscalerError) {
	return map[string]*schedulerframework.NodeInfo{}, nil
}

// CleanUp cleans up processor's internal structures.
func (p *NoOpNodeInfoProviderProcessor) CleanUp() {
}

// NewDefaultNodeInfoProviderProcessor returns a default NodeInfoProviderProcessor.
func NewDefaultNodeInfoProviderProcessor() NodeInfoProviderProcessor {
	return NewMixedNodeInfoProviderProcessor()
}
