/*
Copyright 2025 The Kubernetes Authors.

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
	"maps"
	"sync"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"

	ca_context "k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
)

// TemplateNodeInfoRegistry is a component that stores and exposes template NodeInfos.
type TemplateNodeInfoRegistry struct {
	processor TemplateNodeInfoProvider
	nodeInfos map[string]*framework.NodeInfo
	lock      sync.RWMutex
}

// NewTemplateNodeInfoRegistry creates a new TemplateNodeInfoRegistry.
func NewTemplateNodeInfoRegistry(processor TemplateNodeInfoProvider) *TemplateNodeInfoRegistry {
	return &TemplateNodeInfoRegistry{
		processor: processor,
		nodeInfos: make(map[string]*framework.NodeInfo),
	}
}

// Process calls the embedded processor and updates the cache.
func (r *TemplateNodeInfoRegistry) Process(autoscalingCtx *ca_context.AutoscalingContext, nodes []*apiv1.Node, daemonsets []*appsv1.DaemonSet, taintConfig taints.TaintConfig, currentTime time.Time) (map[string]*framework.NodeInfo, errors.AutoscalerError) {
	nodeInfos, err := r.processor.Process(autoscalingCtx, nodes, daemonsets, taintConfig, currentTime)
	if err != nil {
		return nil, err
	}

	r.lock.Lock()
	defer r.lock.Unlock()
	r.nodeInfos = nodeInfos
	return nodeInfos, nil
}

// CleanUp cleans up the embedded processor.
func (r *TemplateNodeInfoRegistry) CleanUp() {
	r.processor.CleanUp()
}

// GetNodeInfo returns the template NodeInfo for the given node group id.
func (r *TemplateNodeInfoRegistry) GetNodeInfo(id string) (*framework.NodeInfo, bool) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	nodeInfo, found := r.nodeInfos[id]
	return nodeInfo, found
}

// GetNodeInfos returns a copy of the full map of template NodeInfos.
func (r *TemplateNodeInfoRegistry) GetNodeInfos() map[string]*framework.NodeInfo {
	r.lock.RLock()
	defer r.lock.RUnlock()
	result := make(map[string]*framework.NodeInfo, len(r.nodeInfos))
	maps.Copy(result, r.nodeInfos)
	return result
}
