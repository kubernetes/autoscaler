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

package nodeinfosprovider

import (
	"math/rand"
	"sync"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core"
	"k8s.io/autoscaler/cluster-autoscaler/core/utils"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/processors/datadog/common"
	"k8s.io/autoscaler/cluster-autoscaler/processors/datadog/nodeinfosprovider/podtemplate"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/utils/daemonset"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/labels"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	klog "k8s.io/klog/v2"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

const (
	nodeInfoRefreshInterval                       = 5 * time.Minute
	templateOnlyFuncLabel   metrics.FunctionLabel = "TemplateOnlyNodeInfoProvider"
)

type nodeInfoCacheEntry struct {
	nodeInfo    *schedulerframework.NodeInfo
	lastRefresh time.Time
}

// TemplateOnlyNodeInfoProvider return NodeInfos built from node group templates.
type TemplateOnlyNodeInfoProvider struct {
	sync.Mutex
	nodeInfoCache map[string]*nodeInfoCacheEntry
	cloudProvider cloudprovider.CloudProvider
	interrupt     chan struct{}

	podTemplateProcessor podtemplate.Interface
}

// Process returns nodeInfos built from node groups (ASGs, MIGs, VMSS) templates only, not real-world nodes.
func (p *TemplateOnlyNodeInfoProvider) Process(ctx *context.AutoscalingContext, nodes []*apiv1.Node, daemonsets []*appsv1.DaemonSet, ignoredTaints taints.TaintKeySet, currentTime time.Time) (map[string]*schedulerframework.NodeInfo, errors.AutoscalerError) {
	defer metrics.UpdateDurationFromStart(templateOnlyFuncLabel, time.Now())
	p.init(ctx.CloudProvider)

	p.Lock()
	defer p.Unlock()

	result := make(map[string]*schedulerframework.NodeInfo)
	for _, nodeGroup := range p.cloudProvider.NodeGroups() {
		var err error
		var nodeInfo *schedulerframework.NodeInfo

		id := nodeGroup.Id()
		if cacheEntry, found := p.nodeInfoCache[id]; found {
			nodeInfo, err = p.GetFullNodeInfoFromBase(id, cacheEntry.nodeInfo, daemonsets, ctx.PredicateChecker, ignoredTaints)
		} else {
			// new nodegroup: this can be slow (locked) but allows discovering new nodegroups faster
			klog.V(4).Infof("No cached base NodeInfo for %s yet", id)
			nodeInfo, err = p.GetNodeInfoFromTemplate(nodeGroup, daemonsets, ctx.PredicateChecker, ignoredTaints)
			if common.NodeHasLocalData(nodeInfo.Node()) {
				common.SetNodeLocalDataResource(nodeInfo)
			}
		}
		if err != nil {
			klog.Warningf("Failed to build NodeInfo template for %s: %v", id, err)
			continue
		}

		labels.UpdateDeprecatedLabels(nodeInfo.Node().ObjectMeta.Labels)
		result[id] = nodeInfo
	}

	return result, nil
}

// init starts a background refresh loop (and a shutdown channel).
// we unfortunately can't do or call that from NewTemplateOnlyNodeInfoProvider(),
// because don't have cloudProvider yet at New time.
func (p *TemplateOnlyNodeInfoProvider) init(cloudProvider cloudprovider.CloudProvider) {
	if p.interrupt != nil {
		return
	}

	p.interrupt = make(chan struct{})
	p.cloudProvider = cloudProvider
	p.refresh()
	go wait.Until(func() {
		p.refresh()
	}, 10*time.Second, p.interrupt)
}

func (p *TemplateOnlyNodeInfoProvider) refresh() {
	result := make(map[string]*nodeInfoCacheEntry)

	for _, nodeGroup := range p.cloudProvider.NodeGroups() {
		id := nodeGroup.Id()

		splay := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(int(nodeInfoRefreshInterval.Seconds() + 1))
		lastRefresh := time.Now().Add(-time.Second * time.Duration(splay))

		if ng, ok := p.nodeInfoCache[id]; ok {
			if ng.lastRefresh.Add(nodeInfoRefreshInterval).After(time.Now()) {
				result[id] = ng
				continue
			}
			lastRefresh = time.Now()
		}

		nodeInfo, err := nodeGroup.TemplateNodeInfo()
		if err != nil {
			klog.Warningf("Unable to build template node for %s: %v", id, err)
			continue
		}

		// Virtual nodes in NodeInfo templates (built from ASG / MIGS / VMSS) having the
		// local-storage:true label now also gets the Datadog local-storage custom resource
		if common.NodeHasLocalData(nodeInfo.Node()) {
			common.SetNodeLocalDataResource(nodeInfo)
		}

		result[id] = &nodeInfoCacheEntry{
			nodeInfo:    nodeInfo,
			lastRefresh: lastRefresh,
		}
	}

	p.Lock()
	p.nodeInfoCache = result
	p.Unlock()
}

// GetNodeInfoFromTemplate returns NodeInfo object built base on TemplateNodeInfo returned by NodeGroup.TemplateNodeInfo().
func (p *TemplateOnlyNodeInfoProvider) GetNodeInfoFromTemplate(nodeGroup cloudprovider.NodeGroup, daemonsets []*appsv1.DaemonSet, predicateChecker simulator.PredicateChecker, ignoredTaints taints.TaintKeySet) (*schedulerframework.NodeInfo, errors.AutoscalerError) {
	id := nodeGroup.Id()
	baseNodeInfo, err := nodeGroup.TemplateNodeInfo()
	if err != nil {
		return nil, errors.ToAutoscalerError(errors.CloudProviderError, err)
	}

	labels.UpdateDeprecatedLabels(baseNodeInfo.Node().ObjectMeta.Labels)

	return p.GetFullNodeInfoFromBase(id, baseNodeInfo, daemonsets, predicateChecker, ignoredTaints)
}

// GetFullNodeInfoFromBase returns a new NodeInfo object built from provided base TemplateNodeInfo
// differs from utils.GetNodeInfoFromTemplate() in that it takes a nodeInfo as arg instead of a
// nodegroup, and doesn't need to call nodeGroup.TemplateNodeInfo() -> we can reuse a cached nodeInfo.
func (p *TemplateOnlyNodeInfoProvider) GetFullNodeInfoFromBase(nodeGroupId string, baseNodeInfo *schedulerframework.NodeInfo, daemonsets []*appsv1.DaemonSet, predicateChecker simulator.PredicateChecker, ignoredTaints taints.TaintKeySet) (*schedulerframework.NodeInfo, errors.AutoscalerError) {
	pods, err := daemonset.GetDaemonSetPodsForNode(baseNodeInfo, daemonsets, predicateChecker)
	if err != nil {
		return nil, errors.ToAutoscalerError(errors.InternalError, err)
	}

	podTpls, err := p.podTemplateProcessor.GetDaemonSetPodsFromPodTemplateForNode(baseNodeInfo, predicateChecker, ignoredTaints)
	if err != nil {
		return nil, errors.ToAutoscalerError(errors.InternalError, err)
	}
	pods = append(pods, podTpls...)

	for _, podInfo := range baseNodeInfo.Pods {
		pods = append(pods, podInfo.Pod)
	}
	fullNodeInfo := schedulerframework.NewNodeInfo(pods...)
	fullNodeInfo.SetNode(baseNodeInfo.Node())
	sanitizedNodeInfo, typedErr := utils.SanitizeNodeInfo(fullNodeInfo, nodeGroupId, ignoredTaints)
	if typedErr != nil {
		return nil, typedErr
	}
	return sanitizedNodeInfo, nil
}

// CleanUp cleans up processor's internal structures.
func (p *TemplateOnlyNodeInfoProvider) CleanUp() {
	p.podTemplateProcessor.CleanUp()
	close(p.interrupt)
}

// NewTemplateOnlyNodeInfoProvider returns a NodeInfoProcessor generating NodeInfos from node group templates.
func NewTemplateOnlyNodeInfoProvider(opts *core.AutoscalerOptions) *TemplateOnlyNodeInfoProvider {
	return &TemplateOnlyNodeInfoProvider{
		nodeInfoCache:        make(map[string]*nodeInfoCacheEntry),
		podTemplateProcessor: podtemplate.NewPodTemplateProcessor(opts),
	}
}
