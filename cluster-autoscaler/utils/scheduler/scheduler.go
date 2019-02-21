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
	"k8s.io/klog"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"
)

// CreateNodeNameToInfoMap obtains a list of pods and pivots that list into a map where the keys are node names
// and the values are the aggregated information for that node. Pods waiting lower priority pods preemption
// (pod.Status.NominatedNodeName is set) are also added to list of pods for a node.
func CreateNodeNameToInfoMap(pods []*apiv1.Pod, nodes []*apiv1.Node) map[string]*schedulernodeinfo.NodeInfo {
	nodeNameToNodeInfo := make(map[string]*schedulernodeinfo.NodeInfo)
	for _, pod := range pods {
		nodeName := pod.Spec.NodeName
		if nodeName == "" {
			nodeName = pod.Status.NominatedNodeName
		}
		if _, ok := nodeNameToNodeInfo[nodeName]; !ok {
			nodeNameToNodeInfo[nodeName] = schedulernodeinfo.NewNodeInfo()
		}
		nodeNameToNodeInfo[nodeName].AddPod(pod)
	}

	for _, node := range nodes {
		if _, ok := nodeNameToNodeInfo[node.Name]; !ok {
			nodeNameToNodeInfo[node.Name] = schedulernodeinfo.NewNodeInfo()
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

// NodeWithPod function returns NodeInfo, which is a copy of nodeInfo argument with an additional pod scheduled on it.
func NodeWithPod(nodeInfo *schedulernodeinfo.NodeInfo, pod *apiv1.Pod) *schedulernodeinfo.NodeInfo {
	podsOnNode := nodeInfo.Pods()
	podsOnNode = append(podsOnNode, pod)
	newNodeInfo := schedulernodeinfo.NewNodeInfo(podsOnNode...)
	if err := newNodeInfo.SetNode(nodeInfo.Node()); err != nil {
		klog.Errorf("error setting node for NodeInfo %s, because of %s", nodeInfo.Node().Name, err.Error())
	}
	return newNodeInfo
}
