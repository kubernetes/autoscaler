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

package nodeinfosprovider

import (
	"reflect"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"

	klog "k8s.io/klog/v2"
)

const stabilizationDelay = 1 * time.Minute
const maxCacheExpireTime = 87660 * time.Hour

type cacheItem struct {
	*framework.NodeInfo
	added time.Time
}

// MixedTemplateNodeInfoProvider build nodeInfos from the cluster's nodes and node groups.
type MixedTemplateNodeInfoProvider struct {
	nodeInfoCache   map[string]cacheItem
	ttl             time.Duration
	forceDaemonSets bool
}

// NewMixedTemplateNodeInfoProvider returns a NodeInfoProvider processor building
// NodeInfos from real-world nodes when available, otherwise from node groups templates.
func NewMixedTemplateNodeInfoProvider(t *time.Duration, forceDaemonSets bool) *MixedTemplateNodeInfoProvider {
	ttl := maxCacheExpireTime
	if t != nil {
		ttl = *t
	}
	return &MixedTemplateNodeInfoProvider{
		nodeInfoCache:   make(map[string]cacheItem),
		ttl:             ttl,
		forceDaemonSets: forceDaemonSets,
	}
}

func (p *MixedTemplateNodeInfoProvider) isCacheItemExpired(added time.Time) bool {
	return time.Now().Sub(added) > p.ttl
}

// CleanUp cleans up processor's internal structures.
func (p *MixedTemplateNodeInfoProvider) CleanUp() {
}

// Process returns the nodeInfos set for this cluster
func (p *MixedTemplateNodeInfoProvider) Process(ctx *context.AutoscalingContext, nodes []*apiv1.Node, daemonsets []*appsv1.DaemonSet, taintConfig taints.TaintConfig, now time.Time) (map[string]*framework.NodeInfo, errors.AutoscalerError) {
	// TODO(mwielgus): This returns map keyed by url, while most code (including scheduler) uses node.Name for a key.
	// TODO(mwielgus): Review error policy - sometimes we may continue with partial errors.
	result := make(map[string]*framework.NodeInfo)
	seenGroups := make(map[string]bool)

	// processNode returns information whether the nodeTemplate was generated and if there was an error.
	processNode := func(node *apiv1.Node) (bool, string, errors.AutoscalerError) {
		nodeGroup, err := ctx.CloudProvider.NodeGroupForNode(node)
		if err != nil {
			return false, "", errors.ToAutoscalerError(errors.CloudProviderError, err)
		}
		if nodeGroup == nil || reflect.ValueOf(nodeGroup).IsNil() {
			return false, "", nil
		}
		id := nodeGroup.Id()
		if _, found := result[id]; !found {
			nodeInfo, err := ctx.ClusterSnapshot.GetNodeInfo(node.Name)
			if err != nil {
				return false, "", errors.NewAutoscalerError(errors.InternalError, "error while retrieving node %s from cluster snapshot - this shouldn't happen: %v", node.Name, err)
			}
			templateNodeInfo, caErr := simulator.TemplateNodeInfoFromExampleNodeInfo(nodeInfo, id, daemonsets, p.forceDaemonSets, taintConfig)
			if err != nil {
				return false, "", caErr
			}
			result[id] = templateNodeInfo
			return true, id, nil
		}
		return false, "", nil
	}

	for _, node := range nodes {
		// Broken nodes might have some stuff missing. Skipping.
		if !isNodeGoodTemplateCandidate(node, now) {
			continue
		}
		added, id, typedErr := processNode(node)
		if typedErr != nil {
			return map[string]*framework.NodeInfo{}, typedErr
		}
		if added && p.nodeInfoCache != nil {
			nodeInfoCopy := simulator.DeepCopyNodeInfo(result[id])
			p.nodeInfoCache[id] = cacheItem{NodeInfo: nodeInfoCopy, added: time.Now()}
		}
	}
	for _, nodeGroup := range ctx.CloudProvider.NodeGroups() {
		id := nodeGroup.Id()
		seenGroups[id] = true
		if _, found := result[id]; found {
			continue
		}

		// No good template, check cache of previously running nodes.
		if p.nodeInfoCache != nil {
			if cacheItem, found := p.nodeInfoCache[id]; found {
				if p.isCacheItemExpired(cacheItem.added) {
					delete(p.nodeInfoCache, id)
				} else {
					result[id] = simulator.DeepCopyNodeInfo(cacheItem.NodeInfo)
					continue
				}
			}
		}

		// No good template, trying to generate one. This is called only if there are no
		// working nodes in the node groups. By default CA tries to use a real-world example.
		nodeInfo, err := simulator.TemplateNodeInfoFromNodeGroupTemplate(nodeGroup, daemonsets, taintConfig)
		if err != nil {
			if err == cloudprovider.ErrNotImplemented {
				continue
			} else {
				klog.Errorf("Unable to build proper template node for %s: %v", id, err)
				return map[string]*framework.NodeInfo{}, errors.ToAutoscalerError(errors.CloudProviderError, err)
			}
		}
		result[id] = nodeInfo
	}

	// Remove invalid node groups from cache
	for id := range p.nodeInfoCache {
		if _, ok := seenGroups[id]; !ok {
			delete(p.nodeInfoCache, id)
		}
	}

	// Last resort - unready/unschedulable nodes.
	for _, node := range nodes {
		// Allowing broken nodes
		if isNodeGoodTemplateCandidate(node, now) {
			continue
		}
		added, _, typedErr := processNode(node)
		if typedErr != nil {
			return map[string]*framework.NodeInfo{}, typedErr
		}
		nodeGroup, err := ctx.CloudProvider.NodeGroupForNode(node)
		if err != nil {
			return map[string]*framework.NodeInfo{}, errors.ToAutoscalerError(
				errors.CloudProviderError, err)
		}
		if added {
			klog.Warningf("Built template for %s based on unready/unschedulable node %s", nodeGroup.Id(), node.Name)
		}
	}

	return result, nil
}

func isNodeGoodTemplateCandidate(node *apiv1.Node, now time.Time) bool {
	ready, lastTransitionTime, _ := kube_util.GetReadinessState(node)
	stable := lastTransitionTime.Add(stabilizationDelay).Before(now)
	schedulable := !node.Spec.Unschedulable
	return ready && stable && schedulable
}
