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

package common

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

const (
	// DatadogLocalStorageLabel is "true" on nodes offering local storage
	DatadogLocalStorageLabel = "nodegroups.datadoghq.com/local-storage"

	// DatadogLocalDataResource is a virtual resource placed on new or future
	// nodes offering local storage, and currently injected as requests on
	// Pending pods having a PVC for local-data volumes.
	DatadogLocalDataResource apiv1.ResourceName = "storageclass/local-data"
)

var (
	// DatadogLocalDataQuantity is the default amount of DatadogLocalDataResource
	DatadogLocalDataQuantity = resource.NewQuantity(1, resource.DecimalSI)
)

// NodeHasLocalData returns true if the node holds a local-storage:true label
func NodeHasLocalData(node *apiv1.Node) bool {
	if node == nil {
		return false
	}
	value, ok := node.GetLabels()[DatadogLocalStorageLabel]
	return ok && value == "true"
}

// SetNodeLocalDataResource updates a NodeInfo with the DatadogLocalDataResource resource
func SetNodeLocalDataResource(nodeInfo *schedulerframework.NodeInfo) {
	if nodeInfo == nil || nodeInfo.Node() == nil {
		return
	}

	node := nodeInfo.Node()
	nodeInfo.RemoveNode()
	if node.Status.Allocatable == nil {
		node.Status.Allocatable = apiv1.ResourceList{}
	}
	if node.Status.Capacity == nil {
		node.Status.Capacity = apiv1.ResourceList{}
	}
	node.Status.Capacity[DatadogLocalDataResource] = DatadogLocalDataQuantity.DeepCopy()
	node.Status.Allocatable[DatadogLocalDataResource] = DatadogLocalDataQuantity.DeepCopy()
	nodeInfo.SetNode(node)
}
