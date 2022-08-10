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
	"k8s.io/klog/v2"

	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

const (
	// DatadogLocalStorageLabel is "true" on nodes offering local storage
	DatadogLocalStorageLabel = "nodegroups.datadoghq.com/local-storage"

	// DatadogLocalDataResource is a virtual resource placed on new or future
	// nodes offering local storage, and currently injected as requests on
	// Pending pods having a PVC for local-data volumes.
	DatadogLocalDataResource apiv1.ResourceName = "storageclass/local-data"

	// DatadogLocalStorageProvisionerLabel is indicating which technology will be used to provide local storage
	DatadogLocalStorageProvisionerLabel = "nodegroups.datadoghq.com/local-storage-provisioner"
	// DatadogInitialStorageCapacityLabel is storing the amount of local storage a new node will have in the beginning
	// e.g. nodegroups.datadoghq.com/initial-storage-capacity=100Gi
	DatadogInitialStorageCapacityLabel = "nodegroups.datadoghq.com/initial-storage-capacity"

	// DatadogStorageProvisionerTopoLVM is the storage provisioner label value to use for topolvm implementation
	DatadogStorageProvisionerTopoLVM = "topolvm"
	// DatadogStorageProvisionerOpenEBS is the storage provisioner label value to use for openebs implementation
	DatadogStorageProvisionerOpenEBS = "openebs-lvm"
)

var (
	// DatadogLocalDataQuantity is the default amount of DatadogLocalDataResource
	DatadogLocalDataQuantity = resource.NewQuantity(1, resource.DecimalSI)
)

// NodeHasLocalData returns true if the node holds a local-storage:true or local-storage-provisioner:<any> label
func NodeHasLocalData(node *apiv1.Node) bool {
	if node == nil {
		return false
	}

	labels := node.GetLabels()

	_, newStorageOk := labels[DatadogLocalStorageProvisionerLabel]
	value, ok := labels[DatadogLocalStorageLabel]

	// the node should have either the local-stoarge or local-storage-provisioner label
	return (ok && value == "true") || newStorageOk
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

	provisioner, _ := node.Labels[DatadogLocalStorageProvisionerLabel]
	switch provisioner {
	case DatadogStorageProvisionerTopoLVM, DatadogStorageProvisionerOpenEBS:
		capacity, _ := node.Labels[DatadogInitialStorageCapacityLabel]
		capacityResource, err := resource.ParseQuantity(capacity)
		if err == nil {
			node.Status.Capacity[DatadogLocalDataResource] = capacityResource.DeepCopy()
			node.Status.Allocatable[DatadogLocalDataResource] = capacityResource.DeepCopy()
		} else {
			klog.Warningf("failed to attach capacity information (%s) to node (%s): %v", capacity, node.Name, err)
		}
	default:
		// The old local-storage provisioner is using a different label for identification.
		// So if we cannot find any of the new options, we should check if it's using the old system and otherwise print a warning.
		if val, ok := node.Labels[DatadogLocalStorageLabel]; ok && val == "true" {
			node.Status.Capacity[DatadogLocalDataResource] = DatadogLocalDataQuantity.DeepCopy()
			node.Status.Allocatable[DatadogLocalDataResource] = DatadogLocalDataQuantity.DeepCopy()
		} else {
			klog.Warningf("this should never be reached. local storage provisioner (%s) is unknown and cannot be used on node: %s", provisioner, node.Name)
		}
	}

	nodeInfo.SetNode(node)
}
