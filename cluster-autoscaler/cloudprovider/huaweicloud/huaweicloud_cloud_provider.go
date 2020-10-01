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

package huaweicloud

import (
	"io"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huawei-cloud-sdk-go/openstack/cce/v3/clusters"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	klog "k8s.io/klog/v2"
	"os"
	"sync"
)

const (
	// GPULabel is the label added to nodes with GPU resource.
	GPULabel = "cloud.google.com/gke-accelerator"
)

var (
	availableGPUTypes = map[string]struct{}{
		"nvidia-tesla-k80":  {},
		"nvidia-tesla-p100": {},
		"nvidia-tesla-v100": {},
	}
)

// huaweicloudCloudProvider implements CloudProvider interface defined in autoscaler/cluster-autoscaler/cloudprovider/cloud_provider.go
type huaweicloudCloudProvider struct {
	huaweiCloudManager *huaweicloudCloudManager
	resourceLimiter    *cloudprovider.ResourceLimiter
	nodeGroups         []NodeGroup
}

// Name returns the name of the cloud provider.
func (hcp *huaweicloudCloudProvider) Name() string {
	return cloudprovider.HuaweicloudProviderName
}

// NodeGroups returns all node groups managed by this cloud provider.
func (hcp *huaweicloudCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	groups := make([]cloudprovider.NodeGroup, len(hcp.nodeGroups))
	for i, group := range hcp.nodeGroups {
		groups[i] = &group
	}
	return groups
}

// NodeGroupForNode returns the node group that a given node belongs to.
// Since only a single node group is currently supported in huaweicloudprovider, the first node group is always returned.
func (hcp *huaweicloudCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	// currently there is ONLY one nodegroup
	if _, found := node.ObjectMeta.Labels["node-role.kubernetes.io/master"]; found {
		return nil, nil
	}
	return &(hcp.nodeGroups[0]), nil
}

// Pricing returns pricing model for this cloud provider or error if not available. Not implemented.
func (hcp *huaweicloudCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetAvailableMachineTypes get all machine types that can be requested from the cloud provider. Not implemented.
func (hcp *huaweicloudCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return []string{}, nil
}

// NewNodeGroup builds a theoretical node group based on the node definition provided. The node group is not automatically
// created on the cloud provider side. The node group is not returned by NodeGroups() until it is created. Not implemented.
func (hcp *huaweicloudCloudProvider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string,
	taints []apiv1.Taint, extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns struct containing limits (max, min) for resources (cores, memory etc.).
func (hcp *huaweicloudCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return hcp.resourceLimiter, nil
}

// GPULabel returns the label added to nodes with GPU resource.
func (hcp *huaweicloudCloudProvider) GPULabel() string {
	return GPULabel
}

// GetAvailableGPUTypes returns all available GPU types cloud provider supports.
func (hcp *huaweicloudCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	return availableGPUTypes
}

// Cleanup currently does nothing.
func (hcp *huaweicloudCloudProvider) Cleanup() error {
	return nil
}

// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
// In particular the list of node groups returned by NodeGroups can change as a result of CloudProvider.Refresh().
// Currently only prints debug information.
func (hcp *huaweicloudCloudProvider) Refresh() error {
	for _, nodegroup := range hcp.nodeGroups {
		klog.V(3).Info(nodegroup.Debug())
	}
	return nil
}

// Append appends a node group to the list of node groups managed by this cloud provider.
func (hcp *huaweicloudCloudProvider) Append(group []NodeGroup) {
	hcp.nodeGroups = append(hcp.nodeGroups, group...) // append slice to another
}

// GetInstanceID returns the unique id of a specified node.
func (hcp *huaweicloudCloudProvider) GetInstanceID(node *apiv1.Node) string {
	return node.Spec.ProviderID
}

// buildhuaweicloudCloudProvider returns a new instance of type huaweicloudCloudProvider.
func buildhuaweicloudCloudProvider(huaweiCloudManager *huaweicloudCloudManager, resourceLimiter *cloudprovider.ResourceLimiter) (cloudprovider.CloudProvider, error) {
	hcp := &huaweicloudCloudProvider{
		huaweiCloudManager: huaweiCloudManager,
		resourceLimiter:    resourceLimiter,
		nodeGroups:         []NodeGroup{},
	}
	return hcp, nil
}

// buildHuaweiCloudManager checks the command line arguments and build the huaweicloudCloudManager.
func buildHuaweiCloudManager(opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions) *huaweicloudCloudManager {
	var conf io.ReadCloser

	// check the command line passed-in parameters i.e. settings in the deployment yaml file
	// CloudConfig is the path to the cloud provider configuration file. Empty string for no configuration file.
	// Should be loaded with --cloud-config flag.
	if opts.CloudConfig != "" {
		var err error
		conf, err = os.Open(opts.CloudConfig)
		if err != nil {
			klog.Fatalf("couldn't open cloud provider configuration (cloud-config) %s: %#v", opts.CloudConfig, err)
		}

		defer func() {
			err = conf.Close()
			if err != nil {
				klog.Warningf("failed to close config: %v\n", err)
			}
		}()
	}

	if opts.ClusterName == "" {
		klog.Fatalf("the cluster-name parameter must be set in the deployment file and the value must be <clusterID>")
	}

	if opts.CloudProviderName == "" {
		klog.Fatalf("the cloud-provider parameter must be set in the deployment file and the value must be huaweicloud")
	}

	manager, err := buildManager(conf, do, opts)
	if err != nil {
		klog.Fatalf("failed to create huaweicloud manager: %v", err)
	}
	return manager
}

// getAutoscaleNodePools returns a slice of NodeGroup with Autoscaler label enabled.
func getAutoscaleNodePools(manager *huaweicloudCloudManager, opts config.AutoscalingOptions) *[]NodeGroup {
	nodePools, err := clusters.GetNodePools(manager.clusterClient, opts.ClusterName).Extract()
	if err != nil {
		klog.Fatalf("failed to get node pools information of a cluster: %v\n", err)
	}

	clusterUpdateLock := sync.Mutex{}

	// Given our current implementation just support single node pool,
	// please make sure there is only one node pool with Autoscaling flag turned on in CCE cluster
	var nodePoolsWithAutoscalingEnabled []NodeGroup
	for _, nodePool := range nodePools.Items {
		if !nodePool.Spec.Autoscaling.Enable {
			continue
		}

		klog.V(4).Infof("adding node pool: %q, name: %s, min: %d, max: %d",
			nodePool.Metadata.Uid, nodePool.Metadata.Name, nodePool.Spec.Autoscaling.MinNodeCount, nodePool.Spec.Autoscaling.MaxNodeCount)

		nodePoolsWithAutoscalingEnabled = append(nodePoolsWithAutoscalingEnabled, NodeGroup{
			huaweiCloudManager: manager,
			clusterUpdateMutex: &clusterUpdateLock,
			nodePoolName:       nodePool.Metadata.Name,
			nodePoolId:         nodePool.Metadata.Uid,
			clusterName:        opts.ClusterName,
			autoscalingEnabled: nodePool.Spec.Autoscaling.Enable,
			minNodeCount:       nodePool.Spec.Autoscaling.MinNodeCount,
			maxNodeCount:       nodePool.Spec.Autoscaling.MaxNodeCount,
			targetSize:         &nodePool.NodePoolStatus.CurrentNode,
		})
	}

	if len(nodePoolsWithAutoscalingEnabled) == 0 {
		klog.V(4).Info("cluster-autoscaler is disabled Because no node pools has Autoscaling enabled in CCE cluster")
	}
	return &nodePoolsWithAutoscalingEnabled
}

// BuildHuaweiCloud is called by the autoscaler/cluster-autoscaler/builder to build a huaweicloud cloud provider.
// The manager and nodegroups are created here based on the specs provided via the command line parameters in the deployment file
func BuildHuaweiCloud(opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
	manager := buildHuaweiCloudManager(opts, do)

	if len(do.NodeGroupSpecs) == 0 {
		klog.Fatalf("must specify at least one node group with --nodes=<min>:<max>:<name>,...")
	}

	// TODO: Currently, only support single nodegroup (a.k.a node pool in Huawei CCE)
	if len(do.NodeGroupSpecs) > 1 {
		klog.Fatalf("HuaweiCloud autoscaler only supports a single nodegroup for now")
	}

	provider, err := buildhuaweicloudCloudProvider(manager, rl)
	if err != nil {
		klog.Fatalf("failed to create huaweicloud cloud provider: %v", err)
	}

	nodePoolsWithAutoscalingEnabled := getAutoscaleNodePools(manager, opts)
	provider.(*huaweicloudCloudProvider).Append(*nodePoolsWithAutoscalingEnabled)

	return provider
}
