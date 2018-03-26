/*
Copyright 2018 The Kubernetes Authors.

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

package aztools

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aztools/az"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/kubernetes/plugin/pkg/scheduler/schedulercache"

	"github.com/golang/glog"
)

const (
	// ProviderName is the cloud provider name for Azure
	ProviderName = "aztools"

	scaleToZeroSupported = false
)

// OnScaleUpFunc is a function called on node group increase in AzToolsCloudProvider.
// First parameter is the NodeGroup id, second is the increase delta.
type OnScaleUpFunc func(string, int) error

// OnScaleDownFunc is a function called on cluster scale down
type OnScaleDownFunc func(string, string) error

// OnNodeGroupCreateFunc is a fuction called when a new node group is created.
type OnNodeGroupCreateFunc func(string) error

// OnNodeGroupDeleteFunc is a function called when a node group is deleted.
type OnNodeGroupDeleteFunc func(string) error

// AzToolsCloudProvider is a cloud provider to be used with az_tools.
type AzToolsCloudProvider struct {
	sync.Mutex
	nodes             map[string]string // {providerID: groupName}
	groups            map[string]cloudprovider.NodeGroup
	onScaleUp         func(string, int) error
	onScaleDown       func(string, string) error
	onNodeGroupCreate func(string) error
	onNodeGroupDelete func(string) error
	machineTypes      []string
	machineTemplates  map[string]*schedulercache.NodeInfo
	resourceLimiter   *cloudprovider.ResourceLimiter
}

func BuildAzToolsCloudProvider(
	clusterName string,
	discoveryOpts cloudprovider.NodeGroupDiscoveryOptions,
	rl *cloudprovider.ResourceLimiter,
) (*AzToolsCloudProvider, error) {

	for _, file := range []string{"./az_tools.py", "./deploy.py", "./config.yaml", "./cluster.yaml"} {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return nil, fmt.Errorf("%v is not found. Please make sure you are under `DLworkspace/src/ClusterBootstrap`",
				file,
			)
		}
	}

	provider := NewAzToolsCloudProvider(az.OnScaleUp, az.OnScaleDown, rl)

	grouNames := []string{}

	for i, spec := range discoveryOpts.NodeGroupSpecs {
		if i > 0 {
			glog.Warningf("multiple node groups detected: this is not supported for az tools for now")
		}
		s, err := dynamic.SpecFromString(spec, scaleToZeroSupported)
		if err != nil {
			return nil, fmt.Errorf("failed to parse node group spec: %v: %v", spec, err)
		}

		grpID := s.Name

		grouNames = append(grouNames, grpID)

		// Fetch nodes from ./cluster.yaml
		workers, err := az.GetWorkerList(grpID)
		if err != nil {
			return nil, err
		}

		// Add given node in node group.
		for _, nodeName := range workers {
			provider.AddNode(grpID, nodeName)
		}

		provider.AddNodeGroup(grpID, s.MinSize, s.MaxSize)
	}

	if err := az.InitScalerFromConfig(grouNames); err != nil {
		return nil, err
	}

	return provider, nil
}

// NewAzToolsCloudProvider builds new AzToolsCloudProvider
func NewAzToolsCloudProvider(
	onScaleUp OnScaleUpFunc,
	onScaleDown OnScaleDownFunc,
	rl *cloudprovider.ResourceLimiter,
) *AzToolsCloudProvider {
	return &AzToolsCloudProvider{
		nodes:           make(map[string]string),
		groups:          make(map[string]cloudprovider.NodeGroup),
		onScaleUp:       onScaleUp,
		onScaleDown:     onScaleDown,
		resourceLimiter: rl,
	}
}

// Name returns name of the cloud provider.
func (azcp *AzToolsCloudProvider) Name() string {
	return ProviderName
}

// NodeGroups returns all node groups configured for this cloud provider.
func (azcp *AzToolsCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	azcp.Lock()
	defer azcp.Unlock()

	result := make([]cloudprovider.NodeGroup, 0)
	for _, group := range azcp.groups {
		result = append(result, group)
	}
	return result
}

// GetNodeGroup returns node group with the given name.
func (azcp *AzToolsCloudProvider) GetNodeGroup(name string) cloudprovider.NodeGroup {
	azcp.Lock()
	defer azcp.Unlock()
	return azcp.groups[name]
}

var validRefIdRegex = regexp.MustCompile(`^aztools\:\/\/[-0-9a-z]*$`)

// NodeGroupForNode returns the node group for the given node, nil if the node
// should not be processed by cluster autoscaler, or non-nil error if such
// occurred.
func (azcp *AzToolsCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	azcp.Lock()
	defer azcp.Unlock()

	id := node.Spec.ProviderID
	if validRefIdRegex.FindStringSubmatch(id) == nil {
		return nil, fmt.Errorf("Wrong id: expected provider ID with format 'aztools://<node_name>', but got: %v", id)
	}

	groupName, found := azcp.nodes[id]

	if !found {
		return nil, nil
	}
	group, found := azcp.groups[groupName]
	if !found {
		return nil, nil
	}
	return group, nil
}

// Pricing returns pricing model for this cloud provider or error if not available.
func (azcp *AzToolsCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetAvailableMachineTypes get all machine types that can be requested from the cloud provider.
func (azcp *AzToolsCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return azcp.machineTypes, nil
}

// NOTE(harry): we will not call this since autoprovisioned is not supported.
// NewNodeGroup builds a theoretical node group based on the node definition provided. The node group is not automatically
// created on the cloud provider side. The node group is not returned by NodeGroups() until it is created.
func (azcp *AzToolsCloudProvider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string,
	extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	return &AzToolsNodeGroup{
		cloudProvider:   azcp,
		id:              "autoprovisioned-" + machineType,
		minSize:         0,
		maxSize:         1000,
		exist:           false,
		autoprovisioned: true,
		machineType:     machineType,
	}, nil
}

// AddNodeGroup adds node group to test cloud provider.
func (azcp *AzToolsCloudProvider) AddNodeGroup(id string, min int, max int) {
	azcp.Lock()
	defer azcp.Unlock()

	azcp.groups[id] = &AzToolsNodeGroup{
		cloudProvider:   azcp,
		id:              id,
		minSize:         min,
		maxSize:         max,
		exist:           true,
		autoprovisioned: false,
	}
}

// getProviderID returns node name prefixed by cloud provider.
// It is assumed to be the same as node.spec.providerID, otherwise, this node will be considered as
// non-registered node.
func getProviderID(nodeName string) string {
	return "aztools://" + nodeName
}

// AddNode adds the given node to the group.
func (azcp *AzToolsCloudProvider) AddNode(nodeGroupId string, nodeName string) {
	azcp.Lock()
	defer azcp.Unlock()
	azcp.nodes[getProviderID(nodeName)] = nodeGroupId
}

// GetResourceLimiter returns struct containing limits (max, min) for resources (cores, memory etc.).
func (azcp *AzToolsCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return azcp.resourceLimiter, nil
}

// SetResourceLimiter sets resource limiter.
func (azcp *AzToolsCloudProvider) SetResourceLimiter(resourceLimiter *cloudprovider.ResourceLimiter) {
	azcp.resourceLimiter = resourceLimiter
}

// Cleanup this is a function to close resources associated with the cloud provider
func (azcp *AzToolsCloudProvider) Cleanup() error {
	return nil
}

// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
// In particular the list of node groups returned by NodeGroups can change as a result of CloudProvider.Refresh().
func (azcp *AzToolsCloudProvider) Refresh() error {
	for _, nodeGroup := range azcp.NodeGroups() {
		grpID := nodeGroup.Id()

		// Fetch nodes from ./cluster.yaml
		workers, err := az.GetWorkerList(grpID)
		if err != nil {
			return err
		}

		oldGroup := azcp.GetNodeGroup(grpID)
		oldNodes, err := oldGroup.Nodes()
		if err != nil {
			return err
		}

		changed := false

		if len(workers) != len(oldNodes) {
			changed = true
		} else {
			// If length is equal, check the contents of node list.
			nodeNameSet := sets.NewString()
			for _, nodeID := range oldNodes {
				nodeNameSet.Insert(strings.Split(nodeID, "://")[1])
			}
			if !nodeNameSet.HasAll(workers...) {
				changed = true
			}
		}

		// If current state of node group is not identical with cluster.yaml
		// use cluster.yaml to recover it.
		if changed {
			for _, oldNode := range oldNodes {
				// Old node name already has provider id as prefix.
				delete(azcp.nodes, oldNode)
			}

			for _, nodeName := range workers {
				// AddNode will automatically add provider id as prefix.
				azcp.AddNode(grpID, nodeName)
			}

			azcp.AddNodeGroup(grpID, oldGroup.MinSize(), oldGroup.MaxSize())
			glog.V(4).Infof("Regenerated cached nodes list from: %v, to: %v, in node group: %v.",
				oldNodes, workers, grpID)
		}
	}
	return nil
}

// AzToolsNodeGroup is a node group used by AzToolsCloudProvider.
type AzToolsNodeGroup struct {
	sync.Mutex
	cloudProvider   *AzToolsCloudProvider
	id              string
	maxSize         int
	minSize         int
	exist           bool
	autoprovisioned bool
	machineType     string
}

// MaxSize returns maximum size of the node group.
func (aztng *AzToolsNodeGroup) MaxSize() int {
	aztng.Lock()
	defer aztng.Unlock()

	return aztng.maxSize
}

// MinSize returns minimum size of the node group.
func (aztng *AzToolsNodeGroup) MinSize() int {
	aztng.Lock()
	defer aztng.Unlock()

	return aztng.minSize
}

// TargetSize returns the current target size of the node group. It is possible that the
// number of nodes in Kubernetes is different at the moment but should be equal
// to Size() once everything stabilizes (new nodes finish startup and registration or
// removed nodes are deleted completely)
func (aztng *AzToolsNodeGroup) TargetSize() (int, error) {
	aztng.Lock()
	id := aztng.id
	defer aztng.Unlock()

	nodes, err := az.GetWorkerList(id)
	if err != nil {
		return 0, err
	}
	return len(nodes), nil
}

// IncreaseSize increases the size of the node group. To delete a node you need
// to explicitly name it and use DeleteNode. This function should wait until
// node group size is updated.
func (aztng *AzToolsNodeGroup) IncreaseSize(delta int) error {
	aztng.Lock()
	id := aztng.id
	defer aztng.Unlock()

	return aztng.cloudProvider.onScaleUp(id, delta)
}

// Exist checks if the node group really exists on the cloud provider side. Allows to tell the
// theoretical node group from the real one.
func (aztng *AzToolsNodeGroup) Exist() bool {
	aztng.Lock()
	defer aztng.Unlock()
	return aztng.exist
}

// Create creates the node group on the cloud provider side.
func (aztng *AzToolsNodeGroup) Create() error {
	return fmt.Errorf("unimplemented")
}

// Delete deletes the node group on the cloud provider side.
// This will be executed only for autoprovisioned node groups, once their size drops to 0.
func (aztng *AzToolsNodeGroup) Delete() error {
	return aztng.cloudProvider.onNodeGroupDelete(aztng.id)
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
func (aztng *AzToolsNodeGroup) DecreaseTargetSize(delta int) error {
	// TODO(harry): this is used to fix wrong cluster.yaml: remove delta machines in cluster.yaml
	if delta >= 0 {
		return fmt.Errorf("size decrease must be negative")
	}
	// size, err := mig.gceManager.GetMigSize(mig)
	// if err != nil {
	// 	return err
	// }
	// nodes, err := mig.gceManager.GetMigNodes(mig)
	// if err != nil {
	// 	return err
	// }
	// if int(size)+delta < len(nodes) {
	// 	return fmt.Errorf("attempt to delete existing nodes targetSize:%d delta:%d existingNodes: %d",
	// 		size, delta, len(nodes))
	// }
	// return mig.gceManager.SetMigSize(mig, size+int64(delta))
	return nil
}

// DeleteNodes deletes nodes from this node group. Error is returned either on
// failure or if the given node doesn't belong to this node group. This function
// should wait until node group size is updated.
func (aztng *AzToolsNodeGroup) DeleteNodes(nodes []*apiv1.Node) error {
	aztng.Lock()
	id := aztng.id
	defer aztng.Unlock()

	for _, node := range nodes {
		err := aztng.cloudProvider.onScaleDown(id, node.Name)
		if err != nil {
			return err
		}
	}
	return nil
}

// Id returns an unique identifier of the node group.
func (aztng *AzToolsNodeGroup) Id() string {
	aztng.Lock()
	defer aztng.Unlock()

	return aztng.id
}

// Debug returns a string containing all information regarding this node group.
func (aztng *AzToolsNodeGroup) Debug() string {
	aztng.Lock()
	defer aztng.Unlock()

	return fmt.Sprintf("group: %s: min:%d max:%d", aztng.id, aztng.minSize, aztng.maxSize)
}

// Nodes returns a list of all nodes that belong to this node group.
func (aztng *AzToolsNodeGroup) Nodes() ([]string, error) {
	aztng.Lock()
	defer aztng.Unlock()

	result := make([]string, 0)
	for node, nodegroup := range aztng.cloudProvider.nodes {
		if nodegroup == aztng.id {
			result = append(result, node)
		}
	}
	return result, nil
}

// Autoprovisioned returns true if the node group is autoprovisioned.
func (aztng *AzToolsNodeGroup) Autoprovisioned() bool {
	return aztng.autoprovisioned
}

// TemplateNodeInfo returns a node template for this node group.
func (aztng *AzToolsNodeGroup) TemplateNodeInfo() (*schedulercache.NodeInfo, error) {
	if aztng.cloudProvider.machineTemplates == nil {
		return nil, cloudprovider.ErrNotImplemented
	}
	if aztng.autoprovisioned {
		template, found := aztng.cloudProvider.machineTemplates[aztng.machineType]
		if !found {
			return nil, fmt.Errorf("No template declared for %s", aztng.machineType)
		}
		return template, nil
	}
	template, found := aztng.cloudProvider.machineTemplates[aztng.id]
	if !found {
		return nil, fmt.Errorf("No template declared for %s", aztng.id)
	}
	return template, nil
}
