/*
Copyright 2022 The Kubernetes Authors.

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

package scaleway

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/scaleway/scalewaygo"
	coreoptions "k8s.io/autoscaler/cluster-autoscaler/core/options"
	ca_errors "k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/klog/v2"
)

const (
	// GPULabel is the label added to GPU nodes
	GPULabel = "k8s.scw.cloud/gpu"

	// DefaultRefreshInterval is the default refresh interval for the cloud provider
	DefaultRefreshInterval = 60 * time.Second
)

type scalewayCloudProvider struct {
	// client talks to Kapsule API
	client scalewaygo.Client
	// ClusterID is the cluster id where the Autoscaler is running.
	clusterID string
	// nodeGroups is an abstraction around the Pool object returned by the API
	// key is the Pool ID
	nodeGroups map[string]*NodeGroup
	// providerNodeGroups is a pre-converted slice of node groups for NodeGroups() method
	providerNodeGroups []cloudprovider.NodeGroup
	// refreshInterval is the minimum duration between refreshes
	refreshInterval time.Duration
	// lastRefresh is the last time the nodes and node groups were refreshed from the API
	lastRefresh time.Time
	// lastRefreshError stores the error from the last refresh, if any
	lastRefreshError error

	resourceLimiter *cloudprovider.ResourceLimiter
}

func readConf(config *scalewaygo.Config, configFile io.Reader) error {
	body, err := io.ReadAll(configFile)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, config)
	return err
}

func newScalewayCloudProvider(configFile io.Reader, defaultUserAgent string, rl *cloudprovider.ResourceLimiter) (*scalewayCloudProvider, error) {
	getenvOr := func(key, defaultValue string) string {
		value := os.Getenv(key)
		if value != "" {
			return value
		}
		return defaultValue
	}

	// Config file passed with `cloud-config` flag
	cfg := scalewaygo.Config{}
	if configFile != nil {
		err := readConf(&cfg, configFile)
		if err != nil {
			klog.Errorf("failed to read/parse scaleway config file: %s", err)
		}
	}

	// env takes precedence over config passed by command-line
	cfg.ClusterID = getenvOr("CLUSTER_ID", cfg.ClusterID)
	cfg.SecretKey = getenvOr("SCW_SECRET_KEY", cfg.SecretKey)
	cfg.Region = getenvOr("SCW_REGION", cfg.Region)
	cfg.ApiUrl = getenvOr("SCW_API_URL", cfg.ApiUrl)
	cfg.DefaultCacheControl = DefaultRefreshInterval

	cfg.UserAgent = defaultUserAgent

	client, err := scalewaygo.NewClient(cfg)
	if err != nil {
		klog.Fatalf("failed to create scaleway cloud provider: %v", err)
	}

	klog.V(4).Infof("Scaleway Cloud Provider built; ClusterId=%s,Region=%s,ApiURL=%s", cfg.ClusterID, client.Region(), client.ApiURL())

	provider := &scalewayCloudProvider{
		client:          client,
		clusterID:       cfg.ClusterID,
		resourceLimiter: rl,
		refreshInterval: DefaultRefreshInterval,
	}

	// Perform initial refresh to populate node groups cache
	if err := provider.Refresh(); err != nil {
		klog.Errorf("Failed to perform initial refresh: %v", err)
		return nil, err
	}

	return provider, nil
}

// BuildScaleway returns CloudProvider implementation for Scaleway.
func BuildScaleway(opts *coreoptions.AutoscalerOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
	var configFile io.Reader

	if opts.CloudConfig != "" {
		configFile, err := os.Open(opts.CloudConfig)

		if err != nil {
			klog.Errorf("could not open scaleway configuration %s: %s", opts.CloudConfig, err)
		} else {
			defer func() {
				err = configFile.Close()
				if err != nil {
					klog.Errorf("failed to close scaleway config file: %s", err)
				}
			}()
		}
	}

	provider, err := newScalewayCloudProvider(configFile, opts.UserAgent, rl)
	if err != nil {
		klog.Fatalf("Failed to create Scaleway cloud provider: %v", err)
	}
	return provider
}

// Name returns name of the cloud provider.
func (*scalewayCloudProvider) Name() string {
	return cloudprovider.ScalewayProviderName
}

// NodeGroups returns all node groups configured for this cloud provider.
func (scw *scalewayCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	klog.V(4).Info("NodeGroups,ClusterID=", scw.clusterID)

	return scw.providerNodeGroups
}

// NodeGroupForNode returns the node group for the given node, nil if the node
// should not be processed by cluster autoscaler, or non-nil error if such
// occurred. Must be implemented.
func (scw *scalewayCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	klog.V(4).Infof("NodeGroupForNode,NodeSpecProviderID=%s", node.Spec.ProviderID)

	for _, ng := range scw.nodeGroups {
		if _, ok := ng.nodes[node.Spec.ProviderID]; ok {
			return ng, nil
		}
	}

	return nil, nil
}

// HasInstance returns whether the node has corresponding instance in cloud provider,
// true if the node has an instance, false if it no longer exists
func (scw *scalewayCloudProvider) HasInstance(node *apiv1.Node) (bool, error) {
	return node.Spec.ProviderID != "", nil
}

// Pricing returns pricing model for this cloud provider or error if not available.
// Implementation optional.
func (scw *scalewayCloudProvider) Pricing() (cloudprovider.PricingModel, ca_errors.AutoscalerError) {
	klog.V(4).Info("Pricing,called")
	return scw, nil
}

// GetAvailableMachineTypes get all machine types that can be requested from the cloud provider.
// Implementation optional.
func (scw *scalewayCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return []string{}, nil
}

// NewNodeGroup builds a theoretical node group based on the node definition provided. The node group is not automatically
// created on the cloud provider side. The node group is not returned by NodeGroups() until it is created.
// Implementation optional.
func (scw *scalewayCloudProvider) NewNodeGroup(
	machineType string,
	labels map[string]string,
	systemLabels map[string]string,
	taints []apiv1.Taint,
	extraResources map[string]resource.Quantity,
) (cloudprovider.NodeGroup, error) {
	klog.V(4).Info("NewNodeGroup,called")
	return nil, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns struct containing limits (max, min) for resources (cores, memory etc.).
func (scw *scalewayCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	klog.V(4).Info("GetResourceLimiter,called")
	return scw.resourceLimiter, nil
}

// GPULabel returns the label added to nodes with GPU resource.
func (scw *scalewayCloudProvider) GPULabel() string {
	klog.V(6).Info("GPULabel,called")
	return GPULabel
}

// GetAvailableGPUTypes return all available GPU types cloud provider supports.
func (scw *scalewayCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	klog.V(4).Info("GetAvailableGPUTypes,called")
	return nil
}

// GetNodeGpuConfig returns the label, type and resource name for the GPU added to node. If node doesn't have
// any GPUs, it returns nil.
func (scw *scalewayCloudProvider) GetNodeGpuConfig(node *apiv1.Node) *cloudprovider.GpuConfig {
	klog.V(6).Info("GetNodeGpuConfig,called")
	return gpu.GetNodeGPUFromCloudProvider(scw, node)
}

// Cleanup cleans up open resources before the cloud provider is destroyed, i.e. go routines etc.
func (scw *scalewayCloudProvider) Cleanup() error {
	klog.V(4).Info("Cleanup,called")
	return nil
}

// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
// In particular the list of node groups returned by NodeGroups can change as a result of CloudProvider.Refresh().
func (scw *scalewayCloudProvider) Refresh() error {
	klog.V(4).Info("Refresh,ClusterID=", scw.clusterID)

	// Only skip refresh if lastRefresh is non-zero and interval has not elapsed
	if !scw.lastRefresh.IsZero() && time.Since(scw.lastRefresh) < scw.refreshInterval {
		klog.V(4).Infof("Refresh,ClusterID=%s,skipping refresh, last refresh was %s ago", scw.clusterID, time.Since(scw.lastRefresh))
		return scw.lastRefreshError
	}

	cc, pools, err := scw.client.ListPools(context.Background(), scw.clusterID)
	if err != nil {
		klog.Errorf("Refresh,failed to list pools for cluster %s: %s", scw.clusterID, err)
		scw.lastRefresh = time.Now()
		scw.lastRefreshError = err
		return err
	}
	// Update refresh interval based on Cache-Control header from listPools response
	scw.refreshInterval = cc

	_, nodes, err := scw.client.ListNodes(context.Background(), scw.clusterID)
	if err != nil {
		klog.Errorf("Refresh,failed to list nodes for cluster %s: %s", scw.clusterID, err)
		scw.lastRefresh = time.Now()
		scw.lastRefreshError = err
		return err
	}

	// Build NodeGroups
	nodeGroups := make(map[string]*NodeGroup)
	for _, pool := range pools {
		if !pool.Autoscaling {
			klog.V(4).Infof("Refresh,ClusterID=%s,skipping pool %s (autoscaling disabled)", scw.clusterID, pool.ID)
			continue
		}

		nodeGroup := &NodeGroup{
			Client: scw.client,
			nodes:  make(map[string]*scalewaygo.Node),
			pool:   pool,
		}

		nodeGroups[pool.ID] = nodeGroup
	}

	// Assign nodes to NodeGroups
	for _, node := range nodes {
		_, ok := nodeGroups[node.PoolID]
		if !ok {
			klog.V(4).Infof("Refresh,ClusterID=%s,node %s found for PoolID=%s which does not exist in nodeGroups, skipping", scw.clusterID, node.ProviderID, node.PoolID)
			continue
		}

		nodeGroups[node.PoolID].nodes[node.ProviderID] = &node
	}

	scw.nodeGroups = nodeGroups

	// Pre-convert nodeGroups map to slice for NodeGroups() method
	// This is to avoid converting the map to a slice on every call to NodeGroups()
	// which happens quite often
	scw.providerNodeGroups = make([]cloudprovider.NodeGroup, 0, len(nodeGroups))
	for _, ng := range nodeGroups {
		scw.providerNodeGroups = append(scw.providerNodeGroups, ng)
	}

	klog.V(4).Infof("Refresh,ClusterID=%s,%d pools found", scw.clusterID, len(nodeGroups))

	scw.lastRefresh = time.Now()
	scw.lastRefreshError = nil

	return nil
}

// NodePrice returns a price of running the given node for a given period of time.
// All prices returned by the structure should be in the same currency.
func (scw *scalewayCloudProvider) NodePrice(node *apiv1.Node, startTime time.Time, endTime time.Time) (float64, error) {
	var nodeGroup *NodeGroup
	for _, ng := range scw.nodeGroups {
		if _, ok := ng.nodes[node.Spec.ProviderID]; ok {
			nodeGroup = ng
		}
	}

	if nodeGroup == nil {
		return 0.0, fmt.Errorf("node group not found for node %s", node.Spec.ProviderID)
	}

	d := endTime.Sub(startTime)
	hours := math.Ceil(d.Hours())

	return hours * float64(nodeGroup.pool.NodePricePerHour), nil
}

// PodPrice returns a theoretical minimum price of running a pod for a given
// period of time on a perfectly matching machine.
func (scw *scalewayCloudProvider) PodPrice(pod *apiv1.Pod, startTime time.Time, endTime time.Time) (float64, error) {
	return 0.0, nil
}
