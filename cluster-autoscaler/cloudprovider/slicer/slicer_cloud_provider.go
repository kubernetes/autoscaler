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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"

	kubeclient "k8s.io/client-go/kubernetes"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	coreoptions "k8s.io/autoscaler/cluster-autoscaler/core/options"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"

	"github.com/docker/go-units"
	sdk "github.com/slicervm/sdk"
	klog "k8s.io/klog/v2"
)

// SlicerCloudProvider implements the CloudProvider interface for slicer.
type SlicerCloudProvider struct {
	resourceLimiter *cloudprovider.ResourceLimiter
	nodeGroups      []*SlicerNodeGroup
	kubeClient      kubeclient.Interface
}

// BuildSlicer constructs a new SlicerCloudProvider.
func BuildSlicer(opts *coreoptions.AutoscalerOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
	if opts.CloudConfig == "" {
		klog.Fatalf("No config file provided, please specify it via the --cloud-config flag")
	}

	configFile, err := os.Open(opts.CloudConfig)
	if err != nil {
		klog.Fatalf("Could not open cloud provider configuration file %q, error: %v", opts.CloudConfig, err)
	}
	defer configFile.Close()

	cfg, err := buildConfig(configFile)
	if err != nil {
		klog.Fatalf("Failed to parse config, error: %v", err)
	}

	klog.V(2).Infof("Slicer: K3S URL=%s", cfg.k3sURL)
	klog.V(2).Infof("Slicer: K3S token length=%d", len(cfg.k3sToken))
	if cfg.caBundle != "" {
		klog.V(2).Infof("Slicer: Using custom CA bundle: %s", cfg.caBundle)
	}

	groups := []*SlicerNodeGroup{}
	for nodeGroupName, nodeGroup := range cfg.nodeGroupCfg {
		userAgent := "cluster-autoscaler/dev"

		// Create custom HTTP client with CA bundle if specified
		httpClient, err := createHTTPClientWithCABundle(cfg.caBundle)
		if err != nil {
			klog.Fatalf("Failed to create HTTP client with CA bundle for %s: %v", nodeGroupName, err)
		}

		apiClient := sdk.NewSlicerClient(nodeGroup.slicerUrl, nodeGroup.slicerToken, userAgent, httpClient)

		restGroups, err := apiClient.GetHostGroups(context.Background())

		if err != nil {
			klog.Fatalf("Failed to fetch node groups from %s, error: %v", nodeGroup.slicerUrl, err)
		}

		var restGroup *sdk.SlicerHostGroup
		klog.Infof("Slicer: Fetched %d host groups from API", len(restGroups))
		for _, g := range restGroups {
			if nodeGroupName == g.Name {
				restGroup = &g
			}
		}

		if restGroup == nil {
			klog.Fatalf("Failed to find slicer host group with name %s", nodeGroupName)
		}
		klog.Infof("Slicer: Host group: Name=%s, RAM=%s, VCPU=%d, Arch=%s", restGroup.Name, units.BytesSize(float64(restGroup.RamBytes)), restGroup.CPUs, restGroup.Arch)

		arch := restGroup.Arch
		if len(nodeGroup.arch) > 0 {
			arch = nodeGroup.arch
		}

		newGroup := &SlicerNodeGroup{
			id:         nodeGroupName,
			minSize:    nodeGroup.MinSize(),
			maxSize:    nodeGroup.MaxSize(), // Should be adjust according to available memory and CPU on the slicer host.
			targetSize: 1,                   // Start with 1 node (the existing one)
			apiClient:  apiClient,
			k3sUrl:     cfg.k3sURL,
			k3sToken:   cfg.k3sToken,
			arch:       arch,
		}

		groups = append(groups, newGroup)
		klog.Infof("Slicer: Created node group with ID: %s", newGroup.Id())
	}

	provider := &SlicerCloudProvider{
		resourceLimiter: rl,
		nodeGroups:      groups,
		kubeClient:      opts.KubeClient,
	}

	// Set the provider reference in each node group so they can access kubeClient
	for _, g := range groups {
		g.provider = provider
		klog.V(2).Infof("Slicer: Initializing node group: %s", g.Id())
		nodes, err := g.Nodes()
		if err != nil {
			klog.Errorf("Slicer: Failed to get nodes for group %s: %v", g.Id(), err)
		} else {
			klog.V(2).Infof("Slicer: Node group %s has %d nodes from API", g.Id(), len(nodes))
			for i, node := range nodes {
				klog.V(2).Infof("Slicer: API Node %d: %s", i, node.Id)
			}
		}
		g.targetSize = len(nodes)
		klog.V(2).Infof("Slicer: Set target size for group %s to %d", g.Id(), g.targetSize)
	}

	provider.nodeGroups = groups
	return provider
}

// Name returns the name of the cloud provider.
func (s *SlicerCloudProvider) Name() string {
	return "slicer"
}

// NodeGroups returns all node groups configured for this cloud provider.
func (s *SlicerCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	result := make([]cloudprovider.NodeGroup, len(s.nodeGroups))
	for i, g := range s.nodeGroups {
		result[i] = g
	}
	return result
}

// NodeGroupForNode returns the node group for the given node, or nil if not found.
func (s *SlicerCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	// Add more verbose logging at Info level to see what's happening
	klog.Infof("Slicer: NodeGroupForNode called for node %s", node.Name)
	klog.Infof("Slicer: Node labels: %v", node.Labels)
	klog.Infof("Slicer: Number of configured node groups: %d", len(s.nodeGroups))

	if len(s.nodeGroups) == 0 {
		klog.Infof("Slicer: No node groups configured")
		return nil, nil
	}

	// List all node group IDs for debugging
	var groupIds []string
	for _, group := range s.nodeGroups {
		groupIds = append(groupIds, group.Id())
	}
	klog.Infof("Slicer: Available node group IDs: %v", groupIds)

	// Only match nodes that have the proper slicer/hostgroup label
	// This prevents us from claiming ownership of phantom or stale nodes
	if hostGroupLabel, exists := node.Labels["slicer/hostgroup"]; exists {
		klog.Infof("Slicer: Found slicer/hostgroup label: %s", hostGroupLabel)
		for _, group := range s.nodeGroups {
			klog.Infof("Slicer: Comparing hostgroup label '%s' with node group ID '%s'", hostGroupLabel, group.Id())
			if group.Id() == hostGroupLabel {
				klog.Infof("Slicer: Matched node %s to node group %s via label", node.Name, group.Id())
				return group, nil
			}
		}
		klog.Infof("Slicer: No matching node group found for hostgroup label: %s", hostGroupLabel)
	} else {
		klog.Infof("Slicer: Node %s missing slicer/hostgroup label - not a Slicer-managed node", node.Name)
	}

	klog.Infof("Slicer: No node group found for node %s", node.Name)
	return nil, nil
}

// HasInstance returns whether the node has corresponding instance in cloud provider.
func (s *SlicerCloudProvider) HasInstance(node *apiv1.Node) (bool, error) {
	klog.V(4).Infof("Slicer: Checking if node %s has instance in cloud provider", node.Name)

	// Check if the node has the slicer/hostgroup label to identify if it's managed by slicer
	if hostGroupLabel, exists := node.Labels["slicer/hostgroup"]; exists {
		klog.V(4).Infof("Slicer: Node %s has slicer/hostgroup label: %s", node.Name, hostGroupLabel)
		// Check if this hostgroup exists in our node groups
		for _, group := range s.nodeGroups {
			if group.Id() == hostGroupLabel {
				klog.V(4).Infof("Slicer: Node %s belongs to managed hostgroup %s", node.Name, hostGroupLabel)

				// Also verify the node exists in the Slicer API
				nodes, err := group.apiClient.GetHostGroupNodes(context.Background(), group.Id())
				if err != nil {
					klog.V(4).Infof("Slicer: Failed to get nodes for group %s: %v", group.Id(), err)
					return true, nil // Assume it exists if we can't check
				}

				// Check if the node actually exists in Slicer by comparing the slicer/hostname label
				// Note: We should NOT compare node.Name directly with n.Hostname because:
				// - n.Hostname comes from Slicer API (stable name like "k3s-agent-1")
				// - node.Name comes from Kubernetes (dynamic name like "k3s-agent-1-149257af")
				// - The mapping is done via the slicer/hostname label
				for _, n := range nodes {
					if slicerHostname, exists := node.Labels["slicer/hostname"]; exists && n.Hostname == slicerHostname {
						klog.V(4).Infof("Slicer: Confirmed node %s (hostname: %s) exists in Slicer API", node.Name, slicerHostname)
						return true, nil
					}
				}

				klog.V(4).Infof("Slicer: Node %s has slicer label but not found in Slicer API - may have been deleted", node.Name)
				return false, nil // Node was deleted from Slicer but still exists in K8s
			}
		}
		klog.V(4).Infof("Slicer: Hostgroup %s not found in managed groups", hostGroupLabel)
		return false, nil
	}

	// Special case: If a node has no labels at all, it's likely a stale/ghost node
	// Don't claim ownership of such nodes to avoid confusion
	if len(node.Labels) == 0 {
		klog.V(4).Infof("Slicer: Node %s has no labels - treating as non-Slicer node", node.Name)
		return false, nil
	}

	// If we get here, the node has some labels but no slicer/hostgroup label
	// This means it's not a Slicer-managed node
	klog.V(4).Infof("Slicer: Node %s not managed by Slicer (missing slicer/hostgroup label)", node.Name)
	return false, nil
}

// Pricing returns pricing model for this cloud provider or error if not available.
func (s *SlicerCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetAvailableMachineTypes get all machine types that can be requested from the cloud provider.
func (s *SlicerCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// NewNodeGroup builds a theoretical node group based on the node definition provided.
func (s *SlicerCloudProvider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string, taints []apiv1.Taint, extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns struct containing limits (max, min) for resources (cores, memory etc.).
func (s *SlicerCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return s.resourceLimiter, nil
}

// GPULabel returns the label added to nodes with GPU resource.
func (s *SlicerCloudProvider) GPULabel() string {
	return ""
}

// GetAvailableGPUTypes return all available GPU types cloud provider supports.
func (s *SlicerCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	return nil
}

// GetNodeGpuConfig returns the label, type and resource name for the GPU added to node. If node doesn't have any GPUs, it returns nil.
func (s *SlicerCloudProvider) GetNodeGpuConfig(node *apiv1.Node) *cloudprovider.GpuConfig {
	return nil
}

// Cleanup cleans up open resources before the cloud provider is destroyed, i.e. go routines etc.
func (s *SlicerCloudProvider) Cleanup() error {
	return nil
}

// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
func (s *SlicerCloudProvider) Refresh() error {
	return nil
}

// createHTTPClientWithCABundle creates an HTTP client with optional custom CA bundle
func createHTTPClientWithCABundle(caBundlePath string) (*http.Client, error) {
	if caBundlePath == "" {
		return http.DefaultClient, nil
	}

	// Read the CA bundle file
	caBundlePEM, err := os.ReadFile(caBundlePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA bundle file %s: %w", caBundlePath, err)
	}

	// Create a certificate pool and add the CA bundle
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caBundlePEM) {
		return nil, fmt.Errorf("failed to parse CA bundle from %s", caBundlePath)
	}

	// Create HTTP client with custom TLS config
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs: caCertPool,
		},
	}

	return &http.Client{Transport: transport}, nil
}
