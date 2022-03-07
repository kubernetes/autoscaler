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
	"os"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/services/cce/v3/model"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/klog/v2"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
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
	// key: nodePool.Name
	configNg map[string]*dynamic.NodeGroupSpec
	// key: Node.Uid  value: nodePool
	nodePoolForNodeUID map[string]*NodeGroup
}

// Name returns the name of the cloud provider.
func (hcp *huaweicloudCloudProvider) Name() string {
	return cloudprovider.HuaweicloudProviderName
}

// NodeGroups returns all node groups managed by this cloud provider.
func (hcp *huaweicloudCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	groups := make([]cloudprovider.NodeGroup, len(hcp.nodeGroups))
	for i := 0; i < len(hcp.nodeGroups); i++ {
		groups[i] = &hcp.nodeGroups[i]
	}
	return groups
}

// NodeGroupForNode returns the node group that a given node belongs to.
func (hcp *huaweicloudCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	if _, found := node.ObjectMeta.Labels["node-role.kubernetes.io/master"]; found {
		return nil, nil
	}

	if nodePool, ok := hcp.nodePoolForNodeUID[node.Spec.ProviderID]; ok {
		return nodePool, nil
	}
	return nil, nil
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
// Currently prints debug information and update Cluster Infos.
func (hcp *huaweicloudCloudProvider) Refresh() error {
	initialTargetSize := make(map[string]int)
	keepTemplate := make(map[string]*schedulerframework.NodeInfo)
	for _, nodegroup := range hcp.nodeGroups {
		initialTargetSize[nodegroup.nodePoolId] = *nodegroup.targetSize
		keepTemplate[nodegroup.nodePoolId] = nodegroup.nodeInfoTemplate
		klog.V(3).Info(nodegroup.Debug())
	}

	listNodePoolReq := &model.ListNodePoolsRequest{
		ClusterId: hcp.huaweiCloudManager.clusterName,
	}
	allNodePools, err := hcp.huaweiCloudManager.clusterClient.ListNodePools(listNodePoolReq)

	if err != nil {
		klog.Errorf("failed to get node pools information of a cluster: %v\n", err)
		return err
	}

	nodePoolForNodePoolUID := make(map[string]*NodeGroup)
	hcp.nodeGroups = make([]NodeGroup, 0)
	for _, pool := range *allNodePools.Items {
		nodePool := pool
		if !*nodePool.Spec.Autoscaling.Enable {
			klog.Warningf("NodePool(%s) not enable autoscaling, skip", nodePool.Metadata.Name)
			continue
		}

		spec, ok := hcp.configNg[nodePool.Metadata.Name]
		if len(hcp.configNg) > 0 && !ok {
			klog.Warningf("NodePool(%s) not in config file, skip", nodePool.Metadata.Name)
			continue
		}

		nodeGroup := getNodeGroups(nodePool, hcp.huaweiCloudManager, spec)
		nodePoolForNodePoolUID[*nodePool.Metadata.Uid] = nodeGroup
		hcp.fixNodePool(nodeGroup, initialTargetSize[*nodePool.Metadata.Uid])
		nodeGroup.nodeInfoTemplate = keepTemplate[nodeGroup.nodePoolId]
		hcp.nodeGroups = append(hcp.nodeGroups, *nodeGroup)
	}

	err = hcp.updateNodePoolForNodeUID(nodePoolForNodePoolUID)

	return err
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
func getAutoscaleNodePools(manager *huaweicloudCloudManager, configNg map[string]*dynamic.NodeGroupSpec) *[]NodeGroup {
	listNodePoolReq := &model.ListNodePoolsRequest{
		ClusterId: manager.clusterName,
	}
	nodePools, err := manager.clusterClient.ListNodePools(listNodePoolReq)

	if err != nil {
		klog.Fatalf("failed to get node pools information of a cluster: %v\n", err)
	}

	var nodePoolsWithAutoscalingEnabled []NodeGroup

	for _, nodePool := range *nodePools.Items {
		if !*nodePool.Spec.Autoscaling.Enable {
			continue
		}

		spec, ok := configNg[nodePool.Metadata.Name]
		if !ok && len(configNg) > 0 {
			continue
		}

		klog.V(4).Infof("adding node pool: %q, name: %s, min: %d, max: %d",
			nodePool.Metadata.Uid, nodePool.Metadata.Name, nodePool.Spec.Autoscaling.MinNodeCount, nodePool.Spec.Autoscaling.MaxNodeCount)

		nodePoolsWithAutoscalingEnabled = append(nodePoolsWithAutoscalingEnabled, *getNodeGroups(nodePool, manager, spec))
	}

	if len(nodePoolsWithAutoscalingEnabled) == 0 {
		klog.V(4).Info("cluster-autoscaler is disabled Because no node pools has Autoscaling enabled in CCE cluster")
	}
	for _, nodepool := range nodePoolsWithAutoscalingEnabled {
		klog.Info(nodepool.nodePoolName)
	}
	return &nodePoolsWithAutoscalingEnabled
}

// BuildHuaweiCloud is called by the autoscaler/cluster-autoscaler/builder to build a huaweicloud cloud provider.
// The manager and nodegroups are created here based on the specs provided via the command line parameters in the deployment file
func BuildHuaweiCloud(opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
	manager := buildHuaweiCloudManager(opts, do)

	provider, err := buildhuaweicloudCloudProvider(manager, rl)
	if err != nil {
		klog.Fatalf("failed to create huaweicloud cloud provider: %v", err)
	}

	configNg := getConfigNg(do)
	nodePoolsWithAutoscalingEnabled := getAutoscaleNodePools(manager, configNg)
	provider.(*huaweicloudCloudProvider).Append(*nodePoolsWithAutoscalingEnabled)
	provider.(*huaweicloudCloudProvider).configNg = configNg

	return provider
}
