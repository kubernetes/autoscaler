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

package nodeinfos

import (
	"k8s.io/autoscaler/cluster-autoscaler/context"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

// NodeInfoProcessor processes nodeInfos after they're created.
type NodeInfoProcessor interface {
	// Process processes a map of nodeInfos for node groups.
	Process(ctx *context.AutoscalingContext, nodeInfosForNodeGroups map[string]*schedulerframework.NodeInfo) (map[string]*schedulerframework.NodeInfo, error)
	// CleanUp cleans up processor's internal structures.
	CleanUp()
}

// NoOpNodeInfoProcessor doesn't change nodeInfos.
type NoOpNodeInfoProcessor struct {
}

// Process returns unchanged nodeInfos.
func (p *NoOpNodeInfoProcessor) Process(ctx *context.AutoscalingContext, nodeInfosForNodeGroups map[string]*schedulerframework.NodeInfo) (map[string]*schedulerframework.NodeInfo, error) {
	return nodeInfosForNodeGroups, nil
}

// CleanUp cleans up processor's internal structures.
func (p *NoOpNodeInfoProcessor) CleanUp() {
}

// NewDefaultNodeInfoProcessor returns a default instance of NodeInfoProcessor.
func NewDefaultNodeInfoProcessor() NodeInfoProcessor {
	return &NoOpNodeInfoProcessor{}
}
