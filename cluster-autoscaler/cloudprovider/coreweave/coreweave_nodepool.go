/*
Copyright 2025 The Kubernetes Authors.

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

package coreweave

import (
	"encoding/json"
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

// CoreWeaveNodePool represents a node pool in the CoreWeave cloud provider.
type CoreWeaveNodePool struct {
	nodepool      *unstructured.Unstructured
	name          string
	dynamicClient dynamic.Interface
	client        kubernetes.Interface
}

const (
	maxSizeSpecField    = "maxNodes"
	minSizeSpecField    = "minNodes"
	targetSizeSpecField = "targetNodes"
)

// NewCoreWeaveNodePool creates a new CoreWeaveNodePool instance.
func NewCoreWeaveNodePool(nodepool *unstructured.Unstructured, dynamicClient dynamic.Interface, client kubernetes.Interface) (*CoreWeaveNodePool, error) {
	if nodepool == nil {
		return nil, fmt.Errorf("nodepool cannot be nil")
	}
	name, found, _ := unstructured.NestedString(nodepool.Object, "metadata", "name")
	if !found || name == "" {
		return nil, fmt.Errorf("nodepool name cannot be empty")
	}
	// Ensure the dynamic client is not nil
	if dynamicClient == nil {
		return nil, fmt.Errorf("dynamic client cannot be nil")
	}
	// Ensure the client is not nil
	if client == nil {
		return nil, fmt.Errorf("kubernetes client cannot be nil")
	}
	// Create the CoreWeaveNodePool instance
	return &CoreWeaveNodePool{
		nodepool:      nodepool,
		name:          name,
		dynamicClient: dynamicClient,
		client:        client,
	}, nil
}

// GetName returns the name of the node pool.
func (np *CoreWeaveNodePool) GetName() string {
	return np.name
}

// GetNodePool returns the underlying unstructured node pool object.
func (np *CoreWeaveNodePool) GetNodePool() *unstructured.Unstructured {
	return np.nodepool
}

// GetUID returns the unique identifier (UID) of the node pool.
func (np *CoreWeaveNodePool) GetUID() string {
	uid, _, _ := unstructured.NestedString(np.nodepool.Object, "metadata", "uid")
	return uid
}

// GetAutoscalingEnabled returns whether autoscaling is enabled for the node pool.
func (np *CoreWeaveNodePool) GetAutoscalingEnabled() bool {
	autoscalingEnabled, found, _ := unstructured.NestedBool(np.nodepool.Object, "spec", "autoscaling")
	if !found {
		return false // Default to false if not found
	}
	return autoscalingEnabled
}

// GetMinSize returns the minimum size for autoscaling.
func (np *CoreWeaveNodePool) GetMinSize() int {
	minSize, found, _ := unstructured.NestedInt64(np.nodepool.Object, "spec", minSizeSpecField)
	if !found {
		return 0 // Default to 0 if not found
	}
	return int(minSize)
}

// GetMaxSize returns the maximum size for autoscaling.
func (np *CoreWeaveNodePool) GetMaxSize() int {
	maxSize, found, _ := unstructured.NestedInt64(np.nodepool.Object, "spec", maxSizeSpecField)
	if !found {
		return 0 // Default to 0 if not found
	}
	return int(maxSize)
}

// GetTargetSize returns the target size for the node pool.
func (np *CoreWeaveNodePool) GetTargetSize() int {
	targetSize, found, _ := unstructured.NestedInt64(np.nodepool.Object, "spec", targetSizeSpecField)
	if !found {
		return 0 // Default to 0 if not found
	}
	return int(targetSize)
}

// GetNodes returns the list of nodes in the node pool.
func (np *CoreWeaveNodePool) GetNodes() ([]apiv1.Node, error) {
	nodes, err := GetCoreWeaveNodesByNodeGroupUID(np.client, np.GetUID())
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes for node pool %s: %v", np.GetUID(), err)
	}
	return nodes, nil
}

// SetSize sets the target size of the node pool.
// It validates the size against the minimum and maximum limits defined in the node pool.
// If the size is already set to the desired value, it returns nil without making any changes
func (np *CoreWeaveNodePool) SetSize(size int) error {
	if size < 0 {
		return fmt.Errorf("size cannot be negative")
	}
	// check if new size is within the min and max limits
	minSize := np.GetMinSize()
	maxSize := np.GetMaxSize()
	if size < minSize || size > maxSize {
		return fmt.Errorf("size %d is out of bounds: min %d, max %d", size, minSize, maxSize)
	}
	// Check if the target size is already set to the desired size
	currentTargetSize := np.GetTargetSize()
	if currentTargetSize == size {
		return nil // No change needed
	}
	// If the target size is different, update it
	// Ensure the node pool is not nil
	if np.nodepool == nil {
		return fmt.Errorf("node pool is nil")
	}

	ctx, cancel := GetCoreWeaveContext()
	defer cancel()

	// Set the target size in the patch payload
	patch := map[string]any{
		"spec": map[string]any{
			targetSizeSpecField: size,
		},
	}
	// Convert the patch to JSON bytes
	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return fmt.Errorf("failed to marshal patch: %v", err)
	}

	resource := np.dynamicClient.Resource(CoreWeaveNodeGroupResource).Namespace(np.nodepool.GetNamespace())
	_, err = resource.Patch(ctx, np.nodepool.GetName(), types.MergePatchType, patchBytes, metav1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("failed to update node pool: %v", err)
	}

	//Refresh the node pool object after the update
	updatedNodePool, err := resource.Get(ctx, np.nodepool.GetName(), metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get updated node pool: %v", err)
	}
	// Update the local node pool object with the updated one
	np.nodepool = updatedNodePool
	// Log the successful update
	klog.V(4).Infof("Successfully updated node pool %s target size to %d, desired size to %d", np.GetName(), np.GetTargetSize(), size)
	return nil
}

// ValidateNodes checks if the provided nodes belong to the node pool.
func (np *CoreWeaveNodePool) ValidateNodes(nodes []*apiv1.Node) error {
	if len(nodes) == 0 {
		return fmt.Errorf("no nodes provided for validation")
	}
	for _, node := range nodes {
		if node.Labels == nil || node.Labels[coreWeaveNodePoolUID] != np.GetUID() {
			return fmt.Errorf("node %s does not belong to node pool %s", node.Name, np.GetName())
		}
	}
	return nil
}

// GetInstanceType returns the instance type of the node pool.
func (np *CoreWeaveNodePool) GetInstanceType() string {
	instanceType, found, _ := unstructured.NestedString(np.nodepool.Object, "spec", "instanceType")
	if !found {
		return ""
	}
	return instanceType
}

// GetNodeLabels returns the node labels defined in the node pool spec.
func (np *CoreWeaveNodePool) GetNodeLabels() map[string]string {
	labels, found, _ := unstructured.NestedStringMap(np.nodepool.Object, "spec", "nodeLabels")
	if !found {
		return map[string]string{} // Return empty map if not found
	}
	return labels
}

// GetNodeTaints returns the node taints defined in the node pool spec.
func (np *CoreWeaveNodePool) GetNodeTaints() []apiv1.Taint {
	taintsRaw, found, _ := unstructured.NestedSlice(np.nodepool.Object, "spec", "nodeTaints")
	if !found || len(taintsRaw) == 0 {
		return []apiv1.Taint{} // Return empty slice if not found
	}

	taints := make([]apiv1.Taint, 0, len(taintsRaw))
	for _, t := range taintsRaw {
		taintMap, ok := t.(map[string]any)
		if !ok {
			continue
		}

		taint := apiv1.Taint{}

		if key, ok := taintMap["key"].(string); ok {
			taint.Key = key
		}
		if value, ok := taintMap["value"].(string); ok {
			taint.Value = value
		}
		if effect, ok := taintMap["effect"].(string); ok {
			taint.Effect = apiv1.TaintEffect(effect)
		}

		taints = append(taints, taint)
	}

	return taints
}
