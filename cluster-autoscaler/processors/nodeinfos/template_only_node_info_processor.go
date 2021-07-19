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

package nodeinfos

import (
	"math/rand"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/utils"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	klog "k8s.io/klog/v2"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

const nodeInfoRefreshInterval = 15 * time.Minute

type nodeInfoCacheEntry struct {
	nodeInfo    *schedulerframework.NodeInfo
	lastRefresh time.Time
}

// TemplateOnlyNodeInfoProcessor return NodeInfos built from node group templates.
type TemplateOnlyNodeInfoProcessor struct {
	nodeInfoCache map[string]*nodeInfoCacheEntry
}

// Process returns nodeInfos built from node groups templates.
func (p *TemplateOnlyNodeInfoProcessor) Process(ctx *context.AutoscalingContext, nodeInfosForNodeGroups map[string]*schedulerframework.NodeInfo, daemonsets []*appsv1.DaemonSet, ignoredTaints taints.TaintKeySet) (map[string]*schedulerframework.NodeInfo, error) {
	result := make(map[string]*schedulerframework.NodeInfo)
	seenGroups := make(map[string]bool)

	for _, nodeGroup := range ctx.CloudProvider.NodeGroups() {
		id := nodeGroup.Id()
		seenGroups[id] = true

		splay := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(int(nodeInfoRefreshInterval.Seconds() + 1))
		lastRefresh := time.Now().Add(-time.Second * time.Duration(splay))
		if ng, ok := p.nodeInfoCache[id]; ok {
			if ng.lastRefresh.Add(nodeInfoRefreshInterval).After(time.Now()) {
				result[id] = ng.nodeInfo
				continue
			}
			lastRefresh = time.Now()
		}

		nodeInfo, err := utils.GetNodeInfoFromTemplate(nodeGroup, daemonsets, ctx.PredicateChecker, ignoredTaints)
		if err != nil {
			if err == cloudprovider.ErrNotImplemented {
				klog.Warningf("Running in template only mode, but template isn't implemented for group %s", id)
				continue
			} else {
				klog.Errorf("Unable to build proper template node for %s: %v", id, err)
				return map[string]*schedulerframework.NodeInfo{},
					errors.ToAutoscalerError(errors.CloudProviderError, err)
			}
		}

		p.nodeInfoCache[id] = &nodeInfoCacheEntry{
			nodeInfo:    nodeInfo,
			lastRefresh: lastRefresh,
		}
		result[id] = nodeInfo
	}

	for id := range p.nodeInfoCache {
		if _, ok := seenGroups[id]; !ok {
			delete(p.nodeInfoCache, id)
		}
	}

	return result, nil
}

// CleanUp cleans up processor's internal structures.
func (p *TemplateOnlyNodeInfoProcessor) CleanUp() {
}

// NewTemplateOnlyNodeInfoProcessor returns a NodeInfoProcessor generating NodeInfos from node group templates.
func NewTemplateOnlyNodeInfoProcessor() *TemplateOnlyNodeInfoProcessor {
	return &TemplateOnlyNodeInfoProcessor{
		nodeInfoCache: make(map[string]*nodeInfoCacheEntry),
	}
}
