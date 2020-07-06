/*
Copyright 2020 The Kubernetes Authors.

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

package taints

import (
	"strings"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/utils/deletetaint"
	"k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	cloudproviderapi "k8s.io/cloud-provider/api"

	klog "k8s.io/klog/v2"
)

const (
	// ReschedulerTaintKey is the name of the taint created by rescheduler.
	ReschedulerTaintKey = "CriticalAddonsOnly"
	// IgnoreTaintPrefix any taint starting with it will be filtered out from autoscaler template node.
	IgnoreTaintPrefix = "ignore-taint.cluster-autoscaler.kubernetes.io/"

	gkeNodeTerminationHandlerTaint = "cloud.google.com/impending-node-termination"

	// AWS: Indicates that a node has volumes stuck in attaching state and hence it is not fit for scheduling more pods
	awsNodeWithImpairedVolumesTaint = "NodeWithImpairedVolumes"
)

// TaintKeySet is a set of taint key
type TaintKeySet map[string]bool

var (
	// NodeConditionTaints lists taint keys used as node conditions
	NodeConditionTaints = TaintKeySet{
		apiv1.TaintNodeNotReady:                     true,
		apiv1.TaintNodeUnreachable:                  true,
		apiv1.TaintNodeUnschedulable:                true,
		apiv1.TaintNodeMemoryPressure:               true,
		apiv1.TaintNodeDiskPressure:                 true,
		apiv1.TaintNodeNetworkUnavailable:           true,
		apiv1.TaintNodePIDPressure:                  true,
		cloudproviderapi.TaintExternalCloudProvider: true,
		cloudproviderapi.TaintNodeShutdown:          true,
		gkeNodeTerminationHandlerTaint:              true,
		awsNodeWithImpairedVolumesTaint:             true,
	}
)

// SanitizeTaints returns filtered taints
func SanitizeTaints(taints []apiv1.Taint, ignoredTaints TaintKeySet) []apiv1.Taint {
	var newTaints []apiv1.Taint
	for _, taint := range taints {
		// Rescheduler can put this taint on a node while evicting non-critical pods.
		// New nodes will not have this taint and so we should strip it when creating
		// template node.
		switch taint.Key {
		case ReschedulerTaintKey:
			klog.V(4).Info("Removing rescheduler taint when creating template")
			continue
		case deletetaint.ToBeDeletedTaint:
			klog.V(4).Infof("Removing autoscaler taint when creating template from node")
			continue
		case deletetaint.DeletionCandidateTaint:
			klog.V(4).Infof("Removing autoscaler soft taint when creating template from node")
			continue
		}

		// ignore conditional taints as they represent a transient node state.
		if exists := NodeConditionTaints[taint.Key]; exists {
			klog.V(4).Infof("Removing node condition taint %s, when creating template from node", taint.Key)
			continue
		}

		if _, exists := ignoredTaints[taint.Key]; exists {
			klog.V(4).Infof("Removing ignored taint %s, when creating template from node", taint.Key)
			continue
		}

		if strings.HasPrefix(taint.Key, IgnoreTaintPrefix) {
			klog.V(4).Infof("Removing taint %s based on prefix, when creation template from node", taint.Key)
			continue
		}

		newTaints = append(newTaints, taint)
	}
	return newTaints
}

// FilterOutNodesWithIgnoredTaints override the condition status of the given nodes to mark them as NotReady when they have
// filtered taints.
func FilterOutNodesWithIgnoredTaints(ignoredTaints TaintKeySet, allNodes, readyNodes []*apiv1.Node) ([]*apiv1.Node, []*apiv1.Node) {
	newAllNodes := make([]*apiv1.Node, 0)
	newReadyNodes := make([]*apiv1.Node, 0)
	nodesWithIgnoredTaints := make(map[string]*apiv1.Node)
	for _, node := range readyNodes {
		if len(node.Spec.Taints) == 0 {
			newReadyNodes = append(newReadyNodes, node)
			continue
		}
		ready := true
		for _, t := range node.Spec.Taints {
			_, hasIgnoredTaint := ignoredTaints[t.Key]
			if hasIgnoredTaint || strings.HasPrefix(t.Key, IgnoreTaintPrefix) {
				ready = false
				nodesWithIgnoredTaints[node.Name] = kubernetes.GetUnreadyNodeCopy(node)
				klog.V(3).Infof("Overriding status of node %v, which seems to have ignored taint %q", node.Name, t.Key)
				break
			}
		}
		if ready {
			newReadyNodes = append(newReadyNodes, node)
		}
	}
	// Override any node with ignored taint with its "unready" copy
	for _, node := range allNodes {
		if newNode, found := nodesWithIgnoredTaints[node.Name]; found {
			newAllNodes = append(newAllNodes, newNode)
		} else {
			newAllNodes = append(newAllNodes, node)
		}
	}
	return newAllNodes, newReadyNodes
}
