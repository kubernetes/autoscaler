/*
Copyright 2017 The Kubernetes Authors.

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

package scheduler

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/scheduler/schedulercache"
)

const (
	// NominatedNodeAnnotationKey is used to annotate a pod that has preempted other pods.
	// The scheduler uses the annotation to find that the pod shouldn't preempt more pods
	// when it gets to the head of scheduling queue again.
	// See podEligibleToPreemptOthers() for more information.
	NominatedNodeAnnotationKey = "NominatedNodeName"
)

// CreateNodeNameToInfoMap obtains a list of pods and pivots that list into a map where the keys are node names
// and the values are the aggregated information for that node. Pods waiting lower priority pods preemption
// (annotated with NominatedNodeAnnotationKey) are also added to list of pods for a node.
func CreateNodeNameToInfoMap(pods []*apiv1.Pod, nodes []*apiv1.Node) map[string]*schedulercache.NodeInfo {

	nodeNameToNodeInfo := make(map[string]*schedulercache.NodeInfo)
	for _, pod := range pods {
		nodeName := pod.Spec.NodeName
		if nodeName == "" {
			nodeName = pod.Annotations[NominatedNodeAnnotationKey]
		}
		if _, ok := nodeNameToNodeInfo[nodeName]; !ok {
			nodeNameToNodeInfo[nodeName] = schedulercache.NewNodeInfo()
		}
		nodeNameToNodeInfo[nodeName].AddPod(pod)
	}

	for _, node := range nodes {
		if _, ok := nodeNameToNodeInfo[node.Name]; !ok {
			nodeNameToNodeInfo[node.Name] = schedulercache.NewNodeInfo()
		}
		nodeNameToNodeInfo[node.Name].SetNode(node)
	}

	// Some pods may be out of sync with node lists. Removing incomplete node infos.
	keysToRemove := make([]string, 0)
	for key, nodeInfo := range nodeNameToNodeInfo {
		if nodeInfo.Node() == nil {
			keysToRemove = append(keysToRemove, key)
		}
	}
	for _, key := range keysToRemove {
		delete(nodeNameToNodeInfo, key)
	}

	return nodeNameToNodeInfo
}
