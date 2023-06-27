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

package brightbox

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	brightbox "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/brightbox/gobrightbox"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/brightbox/gobrightbox/status"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/brightbox/k8ssdk"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	klog "k8s.io/klog/v2"
)

const (
	// GPULabel is added to nodes with GPU resource
	GPULabel = "cloud.brightbox.com/gpu-node"
)

var (
	availableGPUTypes = map[string]struct{}{}
)

// brightboxCloudProvider implements cloudprovider.CloudProvider interface
type brightboxCloudProvider struct {
	resourceLimiter *cloudprovider.ResourceLimiter
	ClusterName     string
	nodeGroups      []cloudprovider.NodeGroup
	nodeMap         map[string]string
	*k8ssdk.Cloud
}

// Name returns name of the cloud provider.
func (b *brightboxCloudProvider) Name() string {
	klog.V(4).Info("Name")
	return cloudprovider.BrightboxProviderName
}

// NodeGroups returns all node groups configured for this cloud provider.
func (b *brightboxCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	klog.V(4).Info("NodeGroups")
	// Duplicate the stored nodegroup elements and return it
	//return append(b.nodeGroups[:0:0], b.nodeGroups...)
	// Or just return the stored nodegroup elements by reference
	return b.nodeGroups
}

// NodeGroupForNode returns the node group for the given node, nil if
// the node should not be processed by cluster autoscaler, or non-nil
// error if such occurred. Must be implemented.
func (b *brightboxCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	klog.V(4).Info("NodeGroupForNode")
	klog.V(4).Infof("Looking for %v", node.Spec.ProviderID)
	groupID, ok := b.nodeMap[k8ssdk.MapProviderIDToServerID(node.Spec.ProviderID)]
	if ok {
		klog.V(4).Infof("Found in group %v", groupID)
		return b.findNodeGroup(groupID), nil
	}
	klog.V(4).Info("Not found")
	return nil, nil
}

// HasInstance returns whether a given node has a corresponding instance in this cloud provider
func (b *brightboxCloudProvider) HasInstance(node *apiv1.Node) (bool, error) {
	return true, cloudprovider.ErrNotImplemented
}

// Refresh is before every main loop and can be used to dynamically
// update cloud provider state.
// In particular the list of node groups returned by NodeGroups can
// change as a result of CloudProvider.Refresh().
func (b *brightboxCloudProvider) Refresh() error {
	klog.V(4).Info("Refresh")
	configmaps, err := b.GetConfigMaps()
	if err != nil {
		return err
	}
	clusterSuffix := "." + b.ClusterName
	nodeGroups := make([]cloudprovider.NodeGroup, 0)
	nodeMap := make(map[string]string)
	for _, configMapOutline := range configmaps {
		if !strings.HasSuffix(configMapOutline.Name, clusterSuffix) {
			klog.V(4).Infof("name %q doesn't match suffix %q. Ignoring %q", configMapOutline.Name, clusterSuffix, configMapOutline.Id)
			continue
		}
		configMap, err := b.GetConfigMap(configMapOutline.Id)
		if err != nil {
			return err
		}
		klog.V(6).Infof("ConfigMap %+v", configMap)
		mapData := make(map[string]string)
		for k, v := range configMap.Data {
			element, ok := v.(string)
			if !ok {
				return fmt.Errorf("Unexpected value for key %q in configMap %q", k, configMap.Id)
			}
			mapData[k] = element
		}
		klog.V(6).Infof("MapData: %+v", mapData)
		minSize, err := strconv.Atoi(mapData["min"])
		if err != nil {
			klog.V(4).Info("Unable to retrieve minimum size. Ignoring")
			continue
		}
		maxSize, err := strconv.Atoi(mapData["max"])
		if err != nil {
			klog.V(4).Info("Unable to retrieve maximum size. Ignoring")
			continue
		}
		if minSize == maxSize {
			klog.V(4).Infof("Group %q has a fixed size %d. Ignoring", mapData["server_group"], minSize)
			continue
		}
		klog.V(4).Infof("Group %q: Node defaults found in %q. Adding to node group list", configMap.Data["server_group"], configMap.Id)
		newNodeGroup, err := makeNodeGroupFromAPIDetails(
			defaultServerName(configMap.Name),
			mapData,
			minSize,
			maxSize,
			b.Cloud,
		)
		if err != nil {
			return err
		}
		group, err := b.GetServerGroup(newNodeGroup.Id())
		if err != nil {
			return err
		}
		for _, server := range group.Servers {
			nodeMap[server.Id] = group.Id
		}
		nodeGroups = append(nodeGroups, newNodeGroup)
	}
	b.nodeGroups = nodeGroups
	b.nodeMap = nodeMap
	klog.V(4).Infof("Refresh located %v node(s) over %v group(s)", len(nodeMap), len(nodeGroups))
	return nil
}

// Pricing returns pricing model for this cloud provider or error if
// not available.
// Implementation optional.
func (b *brightboxCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	klog.V(4).Info("Pricing")
	return nil, cloudprovider.ErrNotImplemented
}

// GetAvailableMachineTypes get all machine types that can be requested
// from the cloud provider.
// Implementation optional.
func (b *brightboxCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	klog.V(4).Info("GetAvailableMachineTypes")
	return nil, cloudprovider.ErrNotImplemented
}

// NewNodeGroup builds a theoretical node group based on the node
// definition provided. The node group is not automatically created on
// the cloud provider side. The node group is not returned by NodeGroups()
// until it is created.
// Implementation optional.
func (b *brightboxCloudProvider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string, taints []apiv1.Taint, extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	klog.V(4).Info("newNodeGroup")
	return nil, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns struct containing limits (max, min) for
// resources (cores, memory etc.).
func (b *brightboxCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	klog.V(4).Info("GetResourceLimiter")
	return b.resourceLimiter, nil
}

// GPULabel returns the label added to nodes with GPU resource.
func (b *brightboxCloudProvider) GPULabel() string {
	klog.V(4).Info("GPULabel")
	return GPULabel
}

// GetAvailableGPUTypes return all available GPU types cloud provider
// supports.
func (b *brightboxCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	klog.V(4).Info("GetAvailableGPUTypes")
	return availableGPUTypes
}

// GetNodeGpuConfig returns the label, type and resource name for the GPU added to node. If node doesn't have
// any GPUs, it returns nil.
func (b *brightboxCloudProvider) GetNodeGpuConfig(node *apiv1.Node) *cloudprovider.GpuConfig {
	klog.V(4).Info("GetNodeGpuConfig")
	return gpu.GetNodeGPUFromCloudProvider(b, node)
}

// Cleanup cleans up open resources before the cloud provider is
// destroyed, i.e. go routines etc.
func (b *brightboxCloudProvider) Cleanup() error {
	klog.V(4).Info("Cleanup")
	return nil
}

// BuildBrightbox builds the Brightbox provider
func BuildBrightbox(
	opts config.AutoscalingOptions,
	do cloudprovider.NodeGroupDiscoveryOptions,
	rl *cloudprovider.ResourceLimiter,
) cloudprovider.CloudProvider {
	klog.V(4).Info("BuildBrightbox")
	klog.V(4).Infof("Config: %+v", opts)
	klog.V(4).Infof("Discovery Options: %+v", do)
	if opts.CloudConfig != "" {
		klog.Warning("supplied config is not read by this version. Using environment")
	}
	if opts.ClusterName == "" {
		klog.Fatal("Set the cluster name option to the Fully Qualified Internal Domain Name of the cluster")
	}
	newCloudProvider := &brightboxCloudProvider{
		ClusterName:     opts.ClusterName,
		resourceLimiter: rl,
		Cloud:           &k8ssdk.Cloud{},
	}
	_, err := newCloudProvider.CloudClient()
	if err != nil {
		klog.Fatalf("Failed to create Brightbox Cloud Client: %v", err)
	}
	return newCloudProvider
}

//private

func (b *brightboxCloudProvider) findNodeGroup(groupID string) cloudprovider.NodeGroup {
	klog.V(4).Info("findNodeGroup")
	klog.V(4).Infof("Looking for %q", groupID)
	for _, nodeGroup := range b.nodeGroups {
		if nodeGroup.Id() == groupID {
			return nodeGroup
		}
	}
	return nil
}

func defaultServerName(name string) string {
	klog.V(4).Info("defaultServerName")
	klog.V(4).Infof("group name is %q", name)
	return "auto." + name
}

func fetchDefaultGroup(groups []brightbox.ServerGroup, clusterName string) string {
	klog.V(4).Info("findDefaultGroup")
	klog.V(4).Infof("for cluster %q", clusterName)
	for _, group := range groups {
		if group.Name == clusterName {
			return group.Id
		}
	}
	klog.Warningf("Unable to detect main group for cluster %q", clusterName)
	return ""
}

type idWithStatus struct {
	id     string
	status string
}

func (b *brightboxCloudProvider) extractGroupDefaults(servers []brightbox.Server) (string, string, string, error) {
	klog.V(4).Info("extractGroupDefaults")
	const zoneSentinel string = "dummyValue"
	zoneID := zoneSentinel
	var serverType, image idWithStatus
	for _, serverSummary := range servers {
		server, err := b.GetServer(
			context.Background(),
			serverSummary.Id,
			serverNotFoundError(serverSummary.Id),
		)
		if err != nil {
			return "", "", "", err
		}
		image = checkForChange(image, idWithStatus{server.Image.Id, server.Image.Status}, "Group has multiple Image Ids")
		serverType = checkForChange(serverType, idWithStatus{server.ServerType.Id, server.ServerType.Status}, "Group has multiple ServerType Ids")
		zoneID = checkZoneForChange(zoneID, server.Zone.Id, zoneSentinel)
	}
	switch {
	case serverType.id == "":
		return "", "", "", fmt.Errorf("Unable to determine Server Type details from Group")
	case image.id == "":
		return "", "", "", fmt.Errorf("Unable to determine Image details from Group")
	case zoneID == zoneSentinel:
		return "", "", "", fmt.Errorf("Unable to determine Zone details from Group")
	case image.status == status.Deprecated:
		klog.Warningf("Selected image %q is deprecated. Please update to an available version", image.id)
	}
	return serverType.id, image.id, zoneID, nil
}

func checkZoneForChange(zoneID string, newZoneID string, sentinel string) string {
	klog.V(4).Info("checkZoneForChange")
	klog.V(4).Infof("new %q, existing %q", newZoneID, zoneID)
	switch zoneID {
	case newZoneID, sentinel:
		return newZoneID
	default:
		klog.V(4).Info("Group is zone balanced")
		return ""
	}
}

func checkForChange(current idWithStatus, newDetails idWithStatus, errorMessage string) idWithStatus {
	klog.V(4).Info("checkForChange")
	klog.V(4).Infof("new %v, existing %v", newDetails, current)
	switch {
	case newDetails == current:
		// Skip to end
	case newDetails.status == status.Available:
		if current.id == "" || current.status == status.Deprecated {
			klog.V(4).Infof("Object %q is available. Selecting", newDetails.id)
			return newDetails
		}
		// Multiple ids
		klog.Warning(errorMessage)
	case newDetails.status == status.Deprecated:
		if current.id == "" {
			klog.V(4).Infof("Object %q is deprecated, but selecting anyway", newDetails.id)
			return newDetails
		}
		// Multiple ids
		klog.Warning(errorMessage)
	default:
		klog.Warningf("Object %q is no longer available. Ignoring.", newDetails.id)
	}
	return current
}
