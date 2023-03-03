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
	"k8s.io/autoscaler/cluster-autoscaler/core/utils"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"

	klog "k8s.io/klog/v2"
)

const stabilizationDelay = 1 * time.Minute
const maxCacheExpireTime = 87660 * time.Hour

type cacheItem struct {
	*schedulerframework.NodeInfo
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
func (p *MixedTemplateNodeInfoProvider) Process(ctx *context.AutoscalingContext, nodes []*apiv1.Node, daemonsets []*appsv1.DaemonSet, ignoredTaints taints.TaintKeySet, now time.Time) (map[string]*schedulerframework.NodeInfo, errors.AutoscalerError) {
	// TODO(mwielgus): This returns map keyed by url, while most code (including scheduler) uses node.Name for a key.
	// TODO(mwielgus): Review error policy - sometimes we may continue with partial errors.
	result := make(map[string]*schedulerframework.NodeInfo)
	seenGroups := make(map[string]bool)

	podsForNodes, err := getPodsForNodes(ctx.ListerRegistry)
	if err != nil {
		return map[string]*schedulerframework.NodeInfo{}, err
	}

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
			// Build nodeInfo.
			nodeInfo, err := simulator.BuildNodeInfoForNode(node, podsForNodes[node.Name], daemonsets, p.forceDaemonSets)
			sanitizedNodeInfo, err := utils.SanitizeNodeInfo(nodeInfo, id, ignoredTaints)
			if err != nil {
				return false, "", err
			}
			result[id] = sanitizedNodeInfo
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
			return map[string]*schedulerframework.NodeInfo{}, typedErr
		}
		if added && p.nodeInfoCache != nil {
			nodeInfoCopy := utils.DeepCopyNodeInfo(result[id])
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
					result[id] = utils.DeepCopyNodeInfo(cacheItem.NodeInfo)
					continue
				}
			}
		}

		// No good template, trying to generate one. This is called only if there are no
		// working nodes in the node groups. By default CA tries to use a real-world example.
		nodeInfo, err := utils.GetNodeInfoFromTemplate(nodeGroup, daemonsets, ignoredTaints)
		if err != nil {
			if err == cloudprovider.ErrNotImplemented {
				continue
			} else {
				klog.Errorf("Unable to build proper template node for %s: %v", id, err)
				return map[string]*schedulerframework.NodeInfo{}, errors.ToAutoscalerError(errors.CloudProviderError, err)
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
			return map[string]*schedulerframework.NodeInfo{}, typedErr
		}
		nodeGroup, err := ctx.CloudProvider.NodeGroupForNode(node)
		if err != nil {
			return map[string]*schedulerframework.NodeInfo{}, errors.ToAutoscalerError(
				errors.CloudProviderError, err)
		}
		if added {
			klog.Warningf("Built template for %s based on unready/unschedulable node %s", nodeGroup.Id(), node.Name)
		}
	}

	return result, nil
}

func getPodsForNodes(listers kube_util.ListerRegistry) (map[string][]*apiv1.Pod, errors.AutoscalerError) {
	pods, err := listers.ScheduledPodLister().List()
	if err != nil {
		return nil, errors.ToAutoscalerError(errors.ApiCallError, err)
	}
	podsForNodes := map[string][]*apiv1.Pod{}
	for _, p := range pods {
		podsForNodes[p.Spec.NodeName] = append(podsForNodes[p.Spec.NodeName], p)
	}
	return podsForNodes, nil
}

func isNodeGoodTemplateCandidate(node *apiv1.Node, now time.Time) bool {
	ready, lastTransitionTime, _ := kube_util.GetReadinessState(node)
	stable := lastTransitionTime.Add(stabilizationDelay).Before(now)
	schedulable := !node.Spec.Unschedulable
	return ready && stable && schedulable
}
