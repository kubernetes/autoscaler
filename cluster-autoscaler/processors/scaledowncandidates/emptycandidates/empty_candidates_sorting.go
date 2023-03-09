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
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

type nodeInfoGetter interface {
	GetNodeInfo(nodeName string) (*schedulerframework.NodeInfo, error)
}

type nodeInfoGetterImpl struct {
	c clustersnapshot.ClusterSnapshot
}

func (n *nodeInfoGetterImpl) GetNodeInfo(nodeName string) (*schedulerframework.NodeInfo, error) {
	return n.c.NodeInfos().Get(nodeName)
}

// NewNodeInfoGetter limits ClusterSnapshot interface to NodeInfoGet() method.
func NewNodeInfoGetter(c clustersnapshot.ClusterSnapshot) *nodeInfoGetterImpl {
	return &nodeInfoGetterImpl{c}
}

// EmptySorting is sorting scale down candidates so that empty nodes appear first.
type EmptySorting struct {
	nodeInfoGetter
	deleteOptions simulator.NodeDeleteOptions
}

// NewEmptySortingProcessor return EmptySorting struct.
func NewEmptySortingProcessor(opts *config.AutoscalingOptions, n nodeInfoGetter) *EmptySorting {
	deleteOptions := simulator.NodeDeleteOptions{
		SkipNodesWithSystemPods:   opts.SkipNodesWithSystemPods,
		SkipNodesWithLocalStorage: opts.SkipNodesWithLocalStorage,
		MinReplicaCount:           opts.MinReplicaCount,
	}
	return &EmptySorting{n, deleteOptions}
}

// ScaleDownEarlierThan return true if node1 is empty and node2 isn't.
func (p *EmptySorting) ScaleDownEarlierThan(node1, node2 *apiv1.Node) bool {
	if p.isNodeEmpty(node1) && !p.isNodeEmpty(node2) {
		return true
	}
	return false
}

func (p *EmptySorting) isNodeEmpty(node *apiv1.Node) bool {
	nodeInfo, err := p.nodeInfoGetter.GetNodeInfo(node.Name)
	if err != nil {
		return false
	}
	podsToRemove, _, _, err := simulator.GetPodsToMove(nodeInfo, p.deleteOptions, nil, nil, time.Now())
	if err == nil && len(podsToRemove) == 0 {
		return true
	}
	return false
}
