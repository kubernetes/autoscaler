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
	"context"
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	coreWeaveGroup        = "compute.coreweave.com"
	coreWeaveVersion      = "v1alpha1"
	coreWeaveResource     = "nodepools"
	coreWeaveNodePoolUID  = "compute.coreweave.com/node-pool-uid"
	coreWeaveNodePoolName = "compute.coreweave.com/node-pool"
	coreWeaveRemoveNode   = "compute.coreweave.com/remove-node"
)

// CoreWeaveNodeGroupResource is the GroupVersionResource for CoreWeave NodeGroup
var CoreWeaveNodeGroupResource = schema.GroupVersionResource{
	Group:    coreWeaveGroup,
	Version:  coreWeaveVersion,
	Resource: coreWeaveResource,
}

// GetCoreWeaveClient returns a Kubernetes client for CoreWeave using the provided rest.Config
func GetCoreWeaveClient(config *rest.Config) (kubernetes.Interface, error) {
	if config == nil {
		return nil, fmt.Errorf("rest.Config is nil")
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %v", err)
	}
	return clientset, nil
}

// GetCoreWeaveDynamicClient returns a dynamic client for CoreWeave using the provided rest.Config
func GetCoreWeaveDynamicClient(config *rest.Config) (dynamic.Interface, error) {
	if config == nil {
		return nil, fmt.Errorf("rest.Config is nil")
	}
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %v", err)
	}
	return dynamicClient, nil
}

// GetCoreWeaveContext returns a context for CoreWeave operations
func GetCoreWeaveContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	return ctx, cancel
}

// GetCoreWeaveNodesByNodeGroupUID retrieves all nodes in a specific CoreWeave node group by its UID
// It returns a slice of Node objects and an error if any occurs.
func GetCoreWeaveNodesByNodeGroupUID(client kubernetes.Interface, nodeGroupUID string) ([]apiv1.Node, error) {
	// Get the context for CoreWeave operations
	ctx, cancel := GetCoreWeaveContext()
	defer cancel()
	if client == nil {
		return nil, fmt.Errorf("kubernetes client is nil")
	}
	// Validate nodeGroupUID
	if nodeGroupUID == "" {
		return nil, fmt.Errorf("node group UID cannot be empty")
	}
	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", coreWeaveNodePoolUID, nodeGroupUID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes for node group %s: %v", nodeGroupUID, err)
	}
	return nodes.Items, nil
}
