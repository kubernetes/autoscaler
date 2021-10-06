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

package csi

import (
	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/client-go/informers"
	storagev1beta1listers "k8s.io/client-go/listers/storage/v1beta1"
	"k8s.io/klog/v2"
)

// CSIProcessor checks whether a node is ready for applications using volumes
// provided by a CSI driver. This is relevant when the autoscaler has been
// configured to check storage capacity.
//
// Without this processor, the following happens:
// - autoscaler determines that it needs a new node to get volumes
//   for a pending pod created
// - the new node starts and is ready to run pods, but the CSI driver
//   itself hasn't started running on it yet
// - autoscaler checks for pending pods, finds that the pod still
//   cannot run and asks for another node
// - the CSI driver starts, creates volumes and the pod runs
// => the extra node is redundant
//
// To determine whether a node will have a CSI driver, a heuristic is used: if
// a template node derived from the node has a CSIStorageCapacity object, then
// the node itself should also have one, otherwise it is not ready.
type CSIProcessor interface {
	// FilterOutNodesWithUnreadyResources removes nodes that should have a CSI
	// driver, but don't have CSIStorageCapacity information yet.
	FilterOutNodesWithUnreadyResources(context *context.AutoscalingContext, allNodes, readyNodes []*apiv1.Node) ([]*apiv1.Node, []*apiv1.Node)

	// CleanUp frees resources.
	CleanUp()
}

type csiProcessor struct {
	csiStorageCapacityLister storagev1beta1listers.CSIStorageCapacityLister
}

// FilterOutNodesWithUnreadyResources removes nodes that should have a CSI
// driver, but don't have CSIStorageCapacity information yet.
func (p csiProcessor) FilterOutNodesWithUnreadyResources(context *context.AutoscalingContext, allNodes, readyNodes []*apiv1.Node) ([]*apiv1.Node, []*apiv1.Node) {
	newAllNodes := make([]*apiv1.Node, 0, len(allNodes))
	newReadyNodes := make([]*apiv1.Node, 0, len(readyNodes))
	nodesWithUnreadyCSI := make(map[string]*apiv1.Node)
	for _, node := range readyNodes {
		// TODO: short-circuit this check if the node has been in the
		// ready state long enough? If all of the tests below hit the
		// API server to query CSIDriver and CSIStorageCapacity objects
		// and all nodes in a cluster get checked again during each
		// scale up run, then this might create a lot of additional
		// load.
		klog.V(3).Infof("checking CSIStorageCapacity of node %s", node.Name)
		if p.isReady(context, node) {
			newReadyNodes = append(newReadyNodes, node)
		} else {
			nodesWithUnreadyCSI[node.Name] = kubernetes.GetUnreadyNodeCopy(node)
		}
	}
	// Override any node with unready CSI with its "unready" copy
	for _, node := range allNodes {
		if newNode, found := nodesWithUnreadyCSI[node.Name]; found {
			newAllNodes = append(newAllNodes, newNode)
		} else {
			newAllNodes = append(newAllNodes, node)
		}
	}
	return newAllNodes, newReadyNodes
}

func (p csiProcessor) isReady(context *context.AutoscalingContext, node *v1.Node) bool {
	cloudProvider := context.CloudProvider
	nodeGroup, err := cloudProvider.NodeGroupForNode(node)
	if err != nil || nodeGroup == nil {
		// Not a node that is part of a node group? Assume that the normal
		// ready state applies and continue.
		klog.V(3).Infof("node %s has no node group, skip CSI check (error: %v)", node.Name, err)
		return true
	}
	nodeInfo, err := nodeGroup.TemplateNodeInfo()
	if err != nil {
		// Again, ignore the node.
		klog.V(3).Infof("node %s has no node info, skip CSI check: %v", node.Name, err)
		return true
	}
	templateNode := nodeInfo.Node()
	expected := p.numStorageCapacityObjects(templateNode)
	if expected == 0 {
		// Node cannot be unready because no CSI storage is expected.
		klog.V(3).Infof("node %s is not expected to have CSI storage: %v", node.Name)
		return true
	}
	actual := p.numStorageCapacityObjects(node)
	if expected <= actual {
		klog.V(3).Infof("node %s has enough CSIStorageCapacity objects (expected %d, have %d)",
			node.Name, expected, actual)
		return true
	}

	// CSI driver should have published capacity information and
	// hasn't done it yet -> treat the node as not ready yet.
	klog.V(3).Infof("node %s is expected to have %d CSIStorageCapacity objects, only has %d -> treat it as unready",
		node.Name, expected, actual)
	return false
}

func (p csiProcessor) numStorageCapacityObjects(node *v1.Node) int {
	count := 0
	capacities, err := p.csiStorageCapacityLister.List(labels.Everything())
	if err != nil {
		klog.Error(err, "list CSIStorageCapacity")
		return 0
	}
	for _, capacity := range capacities {
		// match labels
		if capacity.NodeTopology == nil {
			continue
		}
		selector, err := metav1.LabelSelectorAsSelector(capacity.NodeTopology)
		if err != nil {
			// Invalid capacity object? Ignore it.
			continue
		}
		if selector.Matches(labels.Set(node.Labels)) {
			count++
		}
	}
	return count
}

func (p csiProcessor) CleanUp() {}

// NewDefaultCSIProcessor returns a default instance of CSIProcessor.
func NewDefaultCSIProcessor(informerFactory informers.SharedInformerFactory) CSIProcessor {
	return csiProcessor{
		csiStorageCapacityLister: informerFactory.Storage().V1beta1().CSIStorageCapacities().Lister(),
	}
}
