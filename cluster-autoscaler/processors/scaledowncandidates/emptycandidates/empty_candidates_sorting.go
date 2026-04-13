/*
Copyright 2023 The Kubernetes Authors.

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

package emptycandidates

import (
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/drainability/rules"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/options"
)

type nodeInfoGetter interface {
	GetNodeInfo(nodeName string) (*framework.NodeInfo, error)
}

type nodeInfoGetterImpl struct {
	c clustersnapshot.ClusterSnapshot
}

func (n *nodeInfoGetterImpl) GetNodeInfo(nodeName string) (*framework.NodeInfo, error) {
	return n.c.GetNodeInfo(nodeName)
}

// NewNodeInfoGetter limits ClusterSnapshot interface to NodeInfoGet() method.
func NewNodeInfoGetter(c clustersnapshot.ClusterSnapshot) *nodeInfoGetterImpl {
	return &nodeInfoGetterImpl{c}
}

// EmptySorting is sorting scale down candidates so that empty nodes appear first.
type EmptySorting struct {
	nodeInfoGetter
	deleteOptions     options.NodeDeleteOptions
	drainabilityRules rules.Rules
	isEmptyCache      map[string]emptyInfo
}

// NewEmptySortingProcessor return EmptySorting struct.
func NewEmptySortingProcessor(n nodeInfoGetter, deleteOptions options.NodeDeleteOptions, drainabilityRules rules.Rules) *EmptySorting {
	return &EmptySorting{
		nodeInfoGetter:    n,
		deleteOptions:     deleteOptions,
		drainabilityRules: drainabilityRules,
		isEmptyCache:      make(map[string]emptyInfo),
	}
}

// ScaleDownEarlierThan return true if node1 is empty and node2 isn't, and differentiates by on-completion pods.
func (p *EmptySorting) ScaleDownEarlierThan(node1, node2 *apiv1.Node) bool {
	n1EmptyInfo := p.getNodeEmptyInfo(node1)
	n2EmptyInfo := p.getNodeEmptyInfo(node2)

	if n1EmptyInfo.IsEmpty != n2EmptyInfo.IsEmpty {
		return n1EmptyInfo.IsEmpty
	}
	return !n1EmptyInfo.HasOnCompletionPods && n2EmptyInfo.HasOnCompletionPods
}

type emptyInfo struct {
	IsEmpty             bool
	HasOnCompletionPods bool
}

// ResetState resets internal state before every sorting.
func (p *EmptySorting) ResetState() {
	p.isEmptyCache = make(map[string]emptyInfo)
}

func (p *EmptySorting) getNodeEmptyInfo(node *apiv1.Node) emptyInfo {
	if val, ok := p.isEmptyCache[node.Name]; ok {
		return val
	}
	val := p.isNodeEmptyNoCache(node)
	p.isEmptyCache[node.Name] = val
	return val
}

func (p *EmptySorting) isNodeEmptyNoCache(node *apiv1.Node) emptyInfo {
	nodeInfo, err := p.nodeInfoGetter.GetNodeInfo(node.Name)
	if err != nil {
		return emptyInfo{IsEmpty: false}
	}
	podMoveInfo, err := simulator.GetPodsToMove(nodeInfo, p.deleteOptions, p.drainabilityRules, nil, nil, time.Now())
	if err == nil && len(podMoveInfo.Pods) == 0 {
		return emptyInfo{
			IsEmpty:             true,
			HasOnCompletionPods: len(podMoveInfo.OnCompletionPods) > 0,
		}
	}
	return emptyInfo{IsEmpty: false}
}
