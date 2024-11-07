/*
Copyright 2019 The Kubernetes Authors.

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

package magnum

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum/gophercloud/openstack/containerinfra/v1/nodegroups"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	klog "k8s.io/klog/v2"
)

const (
	// GPULabel is the label added to nodes with GPU resource.
	GPULabel = "magnum.openstack.org/gpu"

	// Refresh interval for node group auto discovery
	discoveryRefreshInterval = 1 * time.Minute
)

var (
	availableGPUTypes = map[string]struct{}{}
)

// magnumCloudProvider implements CloudProvider interface from cluster-autoscaler/cloudprovider module.
type magnumCloudProvider struct {
	magnumManager   magnumManager
	resourceLimiter *cloudprovider.ResourceLimiter

	nodeGroups []*magnumNodeGroup

	// To be locked when modifying or reading the node groups slice.
	nodeGroupsLock *sync.Mutex

	// To be locked when modifying or reading the cluster state from Magnum.
	clusterUpdateLock *sync.Mutex

	usingAutoDiscovery   bool
	autoDiscoveryConfigs []magnumAutoDiscoveryConfig
	lastDiscoveryRefresh time.Time
}

func buildMagnumCloudProvider(magnumManager magnumManager, resourceLimiter *cloudprovider.ResourceLimiter) (*magnumCloudProvider, error) {
	mcp := &magnumCloudProvider{
		magnumManager:   magnumManager,
		resourceLimiter: resourceLimiter,
		nodeGroups:      []*magnumNodeGroup{},
		nodeGroupsLock:  &sync.Mutex{},
	}
	return mcp, nil
}

// Name returns the name of the cloud provider.
func (mcp *magnumCloudProvider) Name() string {
	return cloudprovider.MagnumProviderName
}

// GPULabel returns the label added to nodes with GPU resource.
func (mcp *magnumCloudProvider) GPULabel() string {
	return GPULabel
}

// GetAvailableGPUTypes return all available GPU types cloud provider supports
func (mcp *magnumCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	return availableGPUTypes
}

// GetNodeGpuConfig returns the label, type and resource name for the GPU added to node. If node doesn't have
// any GPUs, it returns nil.
func (mcp *magnumCloudProvider) GetNodeGpuConfig(node *apiv1.Node) *cloudprovider.GpuConfig {
	return gpu.GetNodeGPUFromCloudProvider(mcp, node)
}

// NodeGroups returns all node groups managed by this cloud provider.
func (mcp *magnumCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	mcp.nodeGroupsLock.Lock()
	defer mcp.nodeGroupsLock.Unlock()

	// Have to convert to a slice of the NodeGroup interface type.
	groups := make([]cloudprovider.NodeGroup, len(mcp.nodeGroups))
	for i, group := range mcp.nodeGroups {
		groups[i] = group
	}
	return groups
}

// AddNodeGroup appends a node group to the list of node groups managed by this cloud provider.
func (mcp *magnumCloudProvider) AddNodeGroup(group *magnumNodeGroup) {
	mcp.nodeGroupsLock.Lock()
	defer mcp.nodeGroupsLock.Unlock()
	mcp.nodeGroups = append(mcp.nodeGroups, group)
}

// NodeGroupForNode returns the node group that a given node belongs to.
func (mcp *magnumCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	mcp.nodeGroupsLock.Lock()
	defer mcp.nodeGroupsLock.Unlock()

	// Ignore master node
	if _, found := node.ObjectMeta.Labels["node-role.kubernetes.io/master"]; found {
		return nil, nil
	}
	// Ignore control-plane nodes
	if _, found := node.ObjectMeta.Labels["node-role.kubernetes.io/control-plane"]; found {
		return nil, nil
	}

	ngUUID, err := mcp.magnumManager.nodeGroupForNode(node)
	if err != nil {
		return nil, fmt.Errorf("error finding node group UUID for node %s: %v", node.Spec.ProviderID, err)
	}

	for _, group := range mcp.nodeGroups {
		if group.UUID == ngUUID {
			klog.V(4).Infof("Node %s belongs to node group %s", node.Spec.ProviderID, group.Id())
			return group, nil
		}
	}

	klog.V(4).Infof("Node %s is not part of an autoscaled node group", node.Spec.ProviderID)

	return nil, nil
}

// HasInstance returns whether a given node has a corresponding instance in this cloud provider
func (mcp *magnumCloudProvider) HasInstance(node *apiv1.Node) (bool, error) {
	return true, cloudprovider.ErrNotImplemented
}

// Pricing is not implemented.
func (mcp *magnumCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetAvailableMachineTypes is not implemented.
func (mcp *magnumCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return []string{}, nil
}

// NewNodeGroup is not implemented.
func (mcp *magnumCloudProvider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string,
	taints []apiv1.Taint, extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns resource constraints for the cloud provider
func (mcp *magnumCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return mcp.resourceLimiter, nil
}

// Refresh is called before every autoscaler main loop.
//
// Debug information for each node group is printed with logging level >= 5.
// Every 60 seconds the node group state on the Magnum side is checked,
// to see if there are any node groups that need to be added/removed/updated.
func (mcp *magnumCloudProvider) Refresh() error {
	mcp.nodeGroupsLock.Lock()
	for _, nodegroup := range mcp.nodeGroups {
		klog.V(5).Info(nodegroup.Debug())
	}
	mcp.nodeGroupsLock.Unlock()

	if mcp.usingAutoDiscovery {
		if time.Since(mcp.lastDiscoveryRefresh) > discoveryRefreshInterval {
			mcp.lastDiscoveryRefresh = time.Now()
			err := mcp.refreshNodeGroups()
			if err != nil {
				return fmt.Errorf("error refreshing node groups: %v", err)
			}
		}
	}

	return nil
}

// Cleanup currently does nothing.
func (mcp *magnumCloudProvider) Cleanup() error {
	return nil
}

// refreshNodeGroups gets the list of node groups which meet the requirements for autoscaling,
// creates magnumNodeGroups for any that do not exist in the cloud provider,
// and drops any node groups which are present in the cloud provider but not in the
// list of node groups that should be autoscaled.
//
// Any node groups which have had their min/max node count updated in Magnum
// are updated with the new limits.
func (mcp *magnumCloudProvider) refreshNodeGroups() error {
	mcp.clusterUpdateLock.Lock()
	defer mcp.clusterUpdateLock.Unlock()

	// Get the list of node groups that match the auto discovery configuration and
	// meet the requirements for autoscaling.
	nodeGroups, err := mcp.magnumManager.autoDiscoverNodeGroups(mcp.autoDiscoveryConfigs)
	if err != nil {
		return fmt.Errorf("could not discover node groups: %v", err)
	}

	// Track names of node groups which are added or removed (for logging).
	var newNodeGroupNames []string
	var droppedNodeGroupNames []string

	// Use maps for easier lookups of node group names.

	// Node group names as registered in the autoscaler.
	registeredNGs := make(map[string]*magnumNodeGroup)
	mcp.nodeGroupsLock.Lock()
	for _, ng := range mcp.nodeGroups {
		registeredNGs[ng.UUID] = ng
	}
	mcp.nodeGroupsLock.Unlock()

	// Node group names that exist on the cloud side and should be autoscaled.
	autoscalingNGs := make(map[string]string)

	for _, nodeGroup := range nodeGroups {
		name := uniqueName(nodeGroup)

		// Just need the name in the key.
		autoscalingNGs[nodeGroup.UUID] = ""

		if ng, alreadyRegistered := registeredNGs[nodeGroup.UUID]; alreadyRegistered {
			// Node group exists in autoscaler and in cloud, only need to check if min/max node count have changed.
			if ng.minSize != nodeGroup.MinNodeCount {
				ng.minSize = nodeGroup.MinNodeCount
				klog.V(2).Infof("Node group %s min node count changed to %d", nodeGroup.Name, ng.minSize)
			}
			// Node groups with unset max node count are not eligible for autoscaling, so this dereference is safe.
			if ng.maxSize != *nodeGroup.MaxNodeCount {
				ng.maxSize = *nodeGroup.MaxNodeCount
				klog.V(2).Infof("Node group %s max node count changed to %d", nodeGroup.Name, ng.maxSize)
			}
			continue
		}

		// The node group is not known to the autoscaler, so create it.
		ng := &magnumNodeGroup{
			magnumManager:     mcp.magnumManager,
			id:                name,
			UUID:              nodeGroup.UUID,
			clusterUpdateLock: mcp.clusterUpdateLock,
			minSize:           nodeGroup.MinNodeCount,
			maxSize:           *nodeGroup.MaxNodeCount,
			targetSize:        nodeGroup.NodeCount,
			deletedNodes:      make(map[string]time.Time),
			nodeTemplate:      getMagnumNodeTemplate(nodeGroup, mcp.magnumManager),
		}
		mcp.AddNodeGroup(ng)
		mcp.magnumManager.fetchNodeGroupStackIDs(ng.UUID)
		newNodeGroupNames = append(newNodeGroupNames, name)
	}

	// Drop any node groups that should not be autoscaled either
	// because they were deleted or had their maximum node count unset.
	// Done by copying all node groups to a buffer, clearing the original
	// node groups and copying back only the ones that should still exist.
	mcp.nodeGroupsLock.Lock()
	buffer := make([]*magnumNodeGroup, len(mcp.nodeGroups))
	copy(buffer, mcp.nodeGroups)

	mcp.nodeGroups = nil

	for _, ng := range buffer {
		if _, ok := autoscalingNGs[ng.UUID]; ok {
			mcp.nodeGroups = append(mcp.nodeGroups, ng)
		} else {
			droppedNodeGroupNames = append(droppedNodeGroupNames, ng.id)
		}
	}

	mcp.nodeGroupsLock.Unlock()

	// Log whatever actions were taken
	if len(newNodeGroupNames) == 0 && len(droppedNodeGroupNames) == 0 {
		klog.V(3).Info("No nodegroups added or removed")
		return nil
	}

	if len(newNodeGroupNames) > 0 {
		klog.V(2).Infof("Discovered %d new node groups for autoscaling: %s", len(newNodeGroupNames),
			strings.Join(newNodeGroupNames, ", "))
	}
	if len(droppedNodeGroupNames) > 0 {
		klog.V(2).Infof("Dropped %d node groups which should no longer be autoscaled: %s",
			len(droppedNodeGroupNames), strings.Join(droppedNodeGroupNames, ", "))
	}

	return nil
}

// BuildMagnum is called by the autoscaler to build a magnum cloud provider.
//
// The magnumManager is created here, and the initial node groups are created
// based on the static or auto discovery specs provided via the command line parameters.
func BuildMagnum(opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
	var config io.ReadCloser

	// Should be loaded with --cloud-config /etc/kubernetes/kube_openstack_config from master node.
	if opts.CloudConfig != "" {
		var err error
		config, err = os.Open(opts.CloudConfig)
		if err != nil {
			klog.Fatalf("Couldn't open cloud provider configuration from %s: %#v", opts.CloudConfig, err)
		}
		defer config.Close()
	}

	// Check that one of static node group discovery or auto discovery are specified.
	if !do.DiscoverySpecified() {
		klog.Fatal("no node group discovery options specified")
	}
	if do.StaticDiscoverySpecified() && do.AutoDiscoverySpecified() {
		klog.Fatal("can not use both static node group discovery and node group auto discovery")
	}

	manager, err := createMagnumManager(config, do, opts)
	if err != nil {
		klog.Fatalf("Failed to create magnum manager: %v", err)
	}

	provider, err := buildMagnumCloudProvider(manager, rl)
	if err != nil {
		klog.Fatalf("Failed to create magnum cloud provider: %v", err)
	}

	clusterUpdateLock := sync.Mutex{}
	provider.clusterUpdateLock = &clusterUpdateLock

	// Handle initial node group discovery.
	if do.StaticDiscoverySpecified() {
		for _, nodegroupSpec := range do.NodeGroupSpecs {
			// Parse a node group spec in the form min:max:name
			spec, err := dynamic.SpecFromString(nodegroupSpec, scaleToZeroSupported)
			if err != nil {
				klog.Fatalf("Could not parse node group spec %s: %v", nodegroupSpec, err)
			}

			ng := &magnumNodeGroup{
				magnumManager:     manager,
				id:                spec.Name,
				clusterUpdateLock: &clusterUpdateLock,
				minSize:           spec.MinSize,
				maxSize:           spec.MaxSize,
				targetSize:        1,
				deletedNodes:      make(map[string]time.Time),
			}

			// Lookup the nodegroup with this name and create a unique name for it based on the UUID.
			name, uuid, err := ng.magnumManager.uniqueNameAndIDForNodeGroup(ng.id)
			if err != nil {
				klog.Fatalf("could not get unique name and UUID for node group %s: %v", spec.Name, err)
			}
			ng.id = name
			ng.UUID = uuid

			// Fetch the current size of this node group.
			ng.targetSize, err = ng.magnumManager.nodeGroupSize(ng.UUID)
			if err != nil {
				klog.Fatalf("Could not get current number of nodes in node group %s: %v", spec.Name, err)
			}

			provider.AddNodeGroup(ng)
			manager.(*magnumManagerImpl).fetchNodeGroupStackIDs(ng.UUID)
		}
	} else if do.AutoDiscoverySpecified() {
		provider.usingAutoDiscovery = true
		cfgs, err := parseMagnumAutoDiscoverySpecs(do)
		if err != nil {
			klog.Fatalf("Could not parse auto discovery specs: %v", err)
		}
		provider.autoDiscoveryConfigs = cfgs

		err = provider.refreshNodeGroups()
		if err != nil {
			klog.Fatalf("Initial node group discovery failed: %v", err)
		}
		provider.lastDiscoveryRefresh = time.Now()
	}

	return provider
}

func getMagnumNodeTemplate(nodegroup *nodegroups.NodeGroup, client magnumManager) *MagnumNodeTemplate {
	template := &MagnumNodeTemplate{}
	flavor, err := client.getFlavorById(nodegroup.FlavorID)

	if err != nil {
		klog.V(5).ErrorS(err, "Failed to build MagnumNodeTemplate. We return a fake template with 4 cores, 4GB ram and 50GB disk.")
		template.CPUCores = 4
		template.RAMMegabytes = 4096
		template.DiskGigabytes = 50
	} else {
		template.CPUCores = flavor.VCPUs
		template.RAMMegabytes = flavor.RAM
		template.DiskGigabytes = flavor.Disk
	}

	if len(nodegroup.Labels) == 0 {
		template.Labels = make(map[string]string)
	} else {
		template.Labels = nodegroup.Labels
	}
	template.Labels["magnum.openstack.org/nodegroup"] = nodegroup.Name

	return template
}
