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

package estimator

import (
	"sort"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	schedulerUtils "k8s.io/autoscaler/cluster-autoscaler/utils/scheduler"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"
)

const oldBinPackingEstimatorDeprecationMessage = "old binpacking estimator is deprecated. It will be removed in Cluster Autoscaler 1.15."

// OldBinpackingNodeEstimator uses the same bin packing logic as in BinPackingEstimator, but, also
// packs in upcoming nodes
//
// Deprecated:
// TODO(vivekbagade): Remove in 1.15
type OldBinpackingNodeEstimator struct {
	predicateChecker *simulator.PredicateChecker
}

// NewOldBinpackingNodeEstimator builds a new OldBinpackingNodeEstimator.
func NewOldBinpackingNodeEstimator(predicateChecker *simulator.PredicateChecker) *OldBinpackingNodeEstimator {
	return &OldBinpackingNodeEstimator{
		predicateChecker: predicateChecker,
	}
}

// Estimate number of nodes needed using bin packing
func (estimator *OldBinpackingNodeEstimator) Estimate(pods []*apiv1.Pod, nodeTemplate *schedulernodeinfo.NodeInfo,
	upcomingNodes []*schedulernodeinfo.NodeInfo) int {

	podInfos := calculatePodScore(pods, nodeTemplate)
	sort.Slice(podInfos, func(i, j int) bool { return podInfos[i].score > podInfos[j].score })

	newNodes := make([]*schedulernodeinfo.NodeInfo, 0)
	newNodes = append(newNodes, upcomingNodes...)

	for _, podInfo := range podInfos {
		found := false
		for i, nodeInfo := range newNodes {
			if err := estimator.predicateChecker.CheckPredicates(podInfo.pod, nil, nodeInfo); err == nil {
				found = true
				newNodes[i] = schedulerUtils.NodeWithPod(nodeInfo, podInfo.pod)
				break
			}
		}
		if !found {
			newNodes = append(newNodes, schedulerUtils.NodeWithPod(nodeTemplate, podInfo.pod))
		}
	}
	return len(newNodes) - len(upcomingNodes)
}
