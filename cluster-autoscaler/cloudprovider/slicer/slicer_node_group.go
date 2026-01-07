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

package slicer

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/docker/go-units"
	sdk "github.com/slicervm/sdk"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	klog "k8s.io/klog/v2"
)

// SlicerNodeGroup implements cloudprovider.NodeGroup for the slicer REST API.
type SlicerNodeGroup struct {
	id         string
	minSize    int
	maxSize    int
	provider   *SlicerCloudProvider
	targetSize int // Track the desired target size locally
	apiClient  *sdk.SlicerClient
	k3sUrl     string
	k3sToken   string
	arch       string
}

// MaxSize returns the maximum size of the node group.
func (g *SlicerNodeGroup) MaxSize() int { return g.maxSize }

// MinSize returns the minimum size of the node group.
func (g *SlicerNodeGroup) MinSize() int { return g.minSize }

// TargetSize returns the current target size of the node group.
func (g *SlicerNodeGroup) TargetSize() (int, error) {
	// Return our local target size, not the actual backend count
	return g.targetSize, nil
}

// IncreaseSize increases the size of the node group by the given delta.
func (g *SlicerNodeGroup) IncreaseSize(delta int) error {
	klog.V(2).Infof("Slicer: IncreaseSize called with delta=%d, current targetSize=%d", delta, g.targetSize)

	if delta <= 0 {
		return fmt.Errorf("delta must be positive")
	}
	if g.targetSize+delta > g.maxSize {
		return fmt.Errorf("size increase too large: current=%d, delta=%d, max=%d", g.targetSize, delta, g.maxSize)
	}

	userdata := fmt.Sprintf(`
#!/bin/bash

HOSTNAME=$(hostname)

# Populate with the join token from the master node
export K3S_TOKEN="$(cat /run/slicer/secrets/k3s-token)"

# Join k3s agent with random node ID but label with Slicer hostname for mapping
curl -sfL https://get.k3s.io | K3S_URL=%s sh -s - --with-node-id --node-label "slicer/hostgroup=%s" --node-label "k3sup.dev/node-type=agent" --node-label "slicer/hostname=$HOSTNAME"
`, g.k3sUrl, g.id)

	klog.V(2).Infof("Slicer: About to create %d nodes via API", delta)

	secrets, err := g.apiClient.ListSecrets(context.Background())
	if err != nil {
		klog.Errorf("Slicer: Failed to list secrets: %v", err)
		return fmt.Errorf("failed to list secrets: %w", err)
	}

	foundK3sToken := false
	for _, secret := range secrets {
		if secret.Name == "k3s-token" {
			foundK3sToken = true
			break
		}
	}

	if !foundK3sToken {
		// Ensure the join token secret exists
		err := g.apiClient.CreateSecret(context.Background(), sdk.CreateSecretRequest{
			Name: "k3s-token",
			Data: base64.StdEncoding.EncodeToString([]byte(g.k3sToken)),
		})
		if err != nil && !errors.Is(err, sdk.ErrSecretExists) {
			klog.Errorf("Slicer: Failed to create k3s token secret: %v", err)
			return fmt.Errorf("failed to create k3s-token secret: %w", err)
		}
	}

	// Create the nodes via API
	for i := 0; i < delta; i++ {
		payload := sdk.SlicerCreateNodeRequest{
			Userdata: userdata,
			Secrets:  []string{"k3s-token"},
		}

		klog.V(2).Infof("Slicer: Creating node via API client")

		result, err := g.apiClient.CreateNode(context.Background(), g.id, payload)
		if err != nil {
			klog.Errorf("Slicer: Failed to create node: %v", err)
			return fmt.Errorf("failed to create node: %w", err)
		}

		klog.V(2).Infof("Slicer: Successfully created node: %s", result.Hostname)
	}

	// Update our local target size
	g.targetSize += delta
	klog.V(2).Infof("Slicer: Increased target size by %d to %d", delta, g.targetSize)
	return nil
}

// AtomicIncreaseSize is not implemented.
func (g *SlicerNodeGroup) AtomicIncreaseSize(delta int) error { return cloudprovider.ErrNotImplemented }

// ForceDeleteNodes is not implemented.
func (g *SlicerNodeGroup) ForceDeleteNodes(nodes []*apiv1.Node) error {
	return cloudprovider.ErrNotImplemented
}

// DecreaseTargetSize decreases the target size of the node group by the given delta.
func (g *SlicerNodeGroup) DecreaseTargetSize(delta int) error {
	klog.V(2).Infof("Slicer: DecreaseTargetSize called with delta=%d, current targetSize=%d", delta, g.targetSize)

	if delta >= 0 {
		return fmt.Errorf("size decrease must be negative")
	}

	// Get actual nodes that exist
	nodes, err := g.Nodes()
	if err != nil {
		return fmt.Errorf("failed to get existing nodes: %w", err)
	}

	newSize := g.targetSize + delta
	if newSize < len(nodes) {
		return fmt.Errorf("attempt to delete existing nodes targetSize:%d delta:%d existingNodes: %d",
			g.targetSize, delta, len(nodes))
	}

	if newSize < g.MinSize() {
		return fmt.Errorf("size decrease too large, desired:%d min:%d", newSize, g.MinSize())
	}

	// Update our local target size
	g.targetSize = newSize
	klog.V(2).Infof("Slicer: Decreased target size by %d to %d", delta, g.targetSize)

	return nil
}

// DeleteNodes deletes the given nodes from the node group.
func (g *SlicerNodeGroup) DeleteNodes(nodes []*apiv1.Node) error {
	klog.V(2).Infof("Slicer: DeleteNodes called with %d nodes", len(nodes))

	if len(nodes) == 0 {
		return nil
	}

	for _, node := range nodes {
		// Use the Slicer hostname from labels, not the K3s node name
		slicerHostname := node.Name // Default to node name
		if hostname, exists := node.Labels["slicer/hostname"]; exists {
			slicerHostname = hostname
		}

		klog.V(2).Infof("Slicer: Deleting node - K3s name: %s, Slicer hostname: %s", node.Name, slicerHostname)

		// First delete from Slicer API
		err := g.apiClient.DeleteNode(g.id, slicerHostname)
		if err != nil {
			klog.Errorf("Slicer: Failed to delete node %s from Slicer API: %v", slicerHostname, err)
			// Continue to try deleting from Kubernetes anyway
		} else {
			klog.V(2).Infof("Slicer: Successfully deleted node from Slicer API: %s", slicerHostname)
		}

		// Then delete from Kubernetes cluster using the actual K3s node name
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := g.provider.kubeClient.CoreV1().Nodes().Delete(ctx, node.Name, metav1.DeleteOptions{}); err != nil {
			// Check if node was already deleted (not found error)
			if strings.Contains(err.Error(), "not found") {
				klog.V(2).Infof("Slicer: Node %s already deleted from Kubernetes", node.Name)
			} else {
				klog.Errorf("Slicer: Failed to delete node %s from Kubernetes: %v", node.Name, err)
				return fmt.Errorf("failed to delete node from Kubernetes: %w", err)
			}
		} else {
			klog.V(2).Infof("Slicer: Successfully deleted node from Kubernetes API: %s", node.Name)
		}

		g.targetSize--
	}

	return nil
}

// Id returns the unique node pool id.
func (g *SlicerNodeGroup) Id() string { return g.id }

// Debug returns a debug string for the NodeGroup.
func (g *SlicerNodeGroup) Debug() string { return g.id }

// Nodes returns a list of all nodes that belong to this node group.
func (g *SlicerNodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	klog.V(4).Infof("Slicer: Fetching nodes for group %s", g.id)
	nodes, err := g.apiClient.GetHostGroupNodes(context.Background(), g.id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch nodes: %w", err)
	}

	klog.V(4).Infof("Slicer: API returned %d nodes for group %s", len(nodes), g.id)

	// Get all Kubernetes nodes to map Slicer API names to actual K8s node names
	kubeNodes, err := g.provider.kubeClient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list Kubernetes nodes: %w", err)
	}

	// Create a map from Slicer hostname to Kubernetes node name
	slicerToK8sNodeMap := make(map[string]string)
	for _, kubeNode := range kubeNodes.Items {
		if hostgroup, hasHostgroup := kubeNode.Labels["slicer/hostgroup"]; hasHostgroup && hostgroup == g.id {
			if slicerHostname, hasHostname := kubeNode.Labels["slicer/hostname"]; hasHostname {
				slicerToK8sNodeMap[slicerHostname] = kubeNode.Name
				klog.V(4).Infof("Slicer: Mapped slicer hostname %s to K8s node %s", slicerHostname, kubeNode.Name)
			}
		}
	}

	instances := make([]cloudprovider.Instance, 0, len(nodes))
	for i, n := range nodes {
		klog.V(4).Infof("Slicer: Node %d: hostname=%s, ip=%s, created=%s", i, n.Hostname, n.IP, n.CreatedAt)

		// Map Slicer API hostname to actual Kubernetes node name
		k8sNodeName, exists := slicerToK8sNodeMap[n.Hostname]
		if !exists {
			klog.V(4).Infof("Slicer: No matching Kubernetes node found for Slicer hostname %s - skipping", n.Hostname)
			continue // Skip nodes that don't have corresponding K8s nodes
		}

		klog.V(4).Infof("Slicer: Using K8s node name %s for Slicer hostname %s", k8sNodeName, n.Hostname)
		instances = append(instances, cloudprovider.Instance{
			Id: k8sNodeName, // Use the actual Kubernetes node name
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceRunning,
			},
		})
	}

	klog.V(4).Infof("Slicer: Returning %d mapped instances for group %s", len(instances), g.id)
	return instances, nil
}

// TemplateNodeInfo returns a node template for this node group.
func (g *SlicerNodeGroup) TemplateNodeInfo() (*framework.NodeInfo, error) {

	klog.V(2).Infof("Slicer: Fetching hostgroup sizing for %s", g.id)

	// Fetch hostgroup sizing
	groups, err := g.apiClient.GetHostGroups(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch hostgroups: %w", err)
	}

	var groupInfo *sdk.SlicerHostGroup
	for _, hg := range groups {
		if hg.Name == g.id {
			groupInfo = &hg
			break
		}
	}

	if groupInfo == nil {
		return nil, fmt.Errorf("hostgroup %s not found", g.id)
	}

	cpu := groupInfo.CPUs
	ramBytes := groupInfo.RamBytes
	if cpu <= 0 || ramBytes <= 0 {
		return nil, fmt.Errorf("invalid cpu or ram from hostgroup: cpu=%d ram=%s", cpu, units.BytesSize(float64(ramBytes)))
	}

	nodeName := "slicer-node-template"
	ramQty := resource.NewQuantity(ramBytes, resource.BinarySI)
	cpuQty := resource.NewQuantity(int64(cpu), resource.DecimalSI)
	labels := map[string]string{
		"kubernetes.io/arch":          g.arch,
		"kubernetes.io/os":            "linux",
		"kubernetes.io/instance-type": "k3s",
		"slicer/hostgroup":            g.id,
		"k3sup.dev/node-type":         "agent",
		"slicer/hostname":             fmt.Sprintf("slicer-node-%d", rand.Intn(1000000)), // Random hostname for uniqueness
	}

	node := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   nodeName,
			Labels: labels,
		},
		Status: apiv1.NodeStatus{
			Capacity: apiv1.ResourceList{
				apiv1.ResourceCPU:    *cpuQty,
				apiv1.ResourceMemory: *ramQty,
				apiv1.ResourcePods:   *resource.NewQuantity(110, resource.DecimalSI),
			},
			Allocatable: apiv1.ResourceList{
				apiv1.ResourceCPU:    *cpuQty,
				apiv1.ResourceMemory: *ramQty,
				apiv1.ResourcePods:   *resource.NewQuantity(110, resource.DecimalSI),
			},
			Conditions: cloudprovider.BuildReadyConditions(),
		},
	}
	// Defensive: check node has required fields
	if node.Name == "" {
		return nil, fmt.Errorf("constructed node is missing name")
	}

	nodeInfo := framework.NewNodeInfo(node, nil, &framework.PodInfo{Pod: cloudprovider.BuildKubeProxy(g.id)})

	return nodeInfo, nil
}

// Exist returns true if the node group exists.
func (g *SlicerNodeGroup) Exist() bool { return true }

// Create is not implemented.
func (g *SlicerNodeGroup) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Delete is not implemented.
func (g *SlicerNodeGroup) Delete() error { return cloudprovider.ErrNotImplemented }

// Autoprovisioned returns true if the node group is autoprovisioned.
func (g *SlicerNodeGroup) Autoprovisioned() bool { return false }

// GetOptions is not implemented.
func (g *SlicerNodeGroup) GetOptions(defaults config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
	return nil, cloudprovider.ErrNotImplemented
}
