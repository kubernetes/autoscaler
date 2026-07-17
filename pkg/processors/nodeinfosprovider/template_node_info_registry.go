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
// It is updated once per autoscaling loop iteration via Recompute() and provides a consistent view of node templates to all processors.
type TemplateNodeInfoRegistry struct {
	provider TemplateNodeInfoProvider

	lock      sync.RWMutex
	nodeInfos map[string]*framework.NodeInfo
}

// NewTemplateNodeInfoRegistry creates a new TemplateNodeInfoRegistry.
func NewTemplateNodeInfoRegistry(provider TemplateNodeInfoProvider) *TemplateNodeInfoRegistry {
	return &TemplateNodeInfoRegistry{
		provider:  provider,
		nodeInfos: make(map[string]*framework.NodeInfo),
	}
}

// Recompute calls the embedded provider to update the cached node infos.
func (r *TemplateNodeInfoRegistry) Recompute(autoscalingCtx *ca_context.AutoscalingContext, nodes []*apiv1.Node, daemonsets []*appsv1.DaemonSet, taintConfig taints.TaintConfig, currentTime time.Time) errors.AutoscalerError {
	nodeInfos, err := r.provider.Process(autoscalingCtx, nodes, daemonsets, taintConfig, currentTime)
	if err != nil {
		return err
	}

	r.lock.Lock()
	defer r.lock.Unlock()
	r.nodeInfos = nodeInfos
	return nil
}

// CleanUp cleans up the embedded provider.
func (r *TemplateNodeInfoRegistry) CleanUp() {
	r.provider.CleanUp()
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
	return maps.Clone(r.nodeInfos)
}
