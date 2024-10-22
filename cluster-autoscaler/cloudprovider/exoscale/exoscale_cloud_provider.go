/*
Copyright 2021 The Kubernetes Authors.

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

package exoscale

import (
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	egoscale "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/exoscale/internal/github.com/exoscale/egoscale/v2"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
)

var _ cloudprovider.CloudProvider = (*exoscaleCloudProvider)(nil)

const exoscaleProviderIDPrefix = "exoscale://"

type exoscaleCloudProvider struct {
	manager         *Manager
	resourceLimiter *cloudprovider.ResourceLimiter
}

func newExoscaleCloudProvider(manager *Manager, rl *cloudprovider.ResourceLimiter) (*exoscaleCloudProvider, error) {
	return &exoscaleCloudProvider{
		manager:         manager,
		resourceLimiter: rl,
	}, nil
}

// Name returns name of the cloud provider.
func (e *exoscaleCloudProvider) Name() string {
	return cloudprovider.ExoscaleProviderName
}

// NodeGroups returns all node groups configured for this cloud provider.
func (e *exoscaleCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	return e.manager.nodeGroups
}

// NodeGroupForNode returns the node group for the given node, nil if the node
// should not be processed by cluster autoscaler, or non-nil error if such
// occurred. Must be implemented.
func (e *exoscaleCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	instancePool, err := e.instancePoolFromNode(node)
	if err != nil {
		if err == errNoInstancePool {
			// If the node is not member of an Instance Pool, don't autoscale it.
			return nil, nil
		}

		return nil, err
	}

	var nodeGroup cloudprovider.NodeGroup
	if instancePool.Manager != nil && instancePool.Manager.Type == "sks-nodepool" {
		// SKS-managed Instance Pool (Nodepool)
		var (
			sksCluster  *egoscale.SKSCluster
			sksNodepool *egoscale.SKSNodepool
		)

		sksClusters, err := e.manager.client.ListSKSClusters(e.manager.ctx, e.manager.zone)
		if err != nil {
			errorf("unable to list SKS clusters: %v", err)
			return nil, err
		}
		for _, c := range sksClusters {
			for _, n := range c.Nodepools {
				if *n.ID == instancePool.Manager.ID {
					sksCluster = c
					sksNodepool = n
					break
				}
			}
		}
		if sksNodepool == nil {
			return nil, fmt.Errorf(
				"no SKS Nodepool found with ID %s in zone %s",
				instancePool.Manager.ID,
				e.manager.zone,
			)
		}

		// nodeGroupSpec contains the configuration spec from the '--nodes' flag
		// which includes the min and max size of the node group.
		var nodeGroupSpec *dynamic.NodeGroupSpec
		for _, spec := range e.manager.discoveryOpts.NodeGroupSpecs {
			s, err := dynamic.SpecFromString(spec, scaleToZeroSupported)
			if err != nil {
				return nil, fmt.Errorf("failed to parse node group spec: %v", err)
			}

			if s.Name == *sksNodepool.Name {
				nodeGroupSpec = s
				break
			}
		}
		var minSize, maxSize int
		if nodeGroupSpec != nil {
			minSize = nodeGroupSpec.MinSize
			maxSize = nodeGroupSpec.MaxSize
		} else {
			minSize = 1
			maxSize, err = e.manager.computeInstanceQuota()
			if err != nil {
				return nil, err
			}
		}

		nodeGroup = &sksNodepoolNodeGroup{
			sksNodepool: sksNodepool,
			sksCluster:  sksCluster,
			m:           e.manager,
			minSize:     minSize,
			maxSize:     maxSize,
		}
		debugf("found node %s belonging to SKS Nodepool %s", toNodeID(node.Spec.ProviderID), *sksNodepool.ID)
	} else {
		// Standalone Instance Pool
		nodeGroup = &instancePoolNodeGroup{
			instancePool: instancePool,
			m:            e.manager,
		}
		debugf("found node %s belonging to Instance Pool %s", toNodeID(node.Spec.ProviderID), *instancePool.ID)
	}

	found := false
	for i, ng := range e.manager.nodeGroups {
		if ng.Id() == nodeGroup.Id() {
			e.manager.nodeGroups[i] = nodeGroup
			found = true
			break
		}
	}
	if !found {
		e.manager.nodeGroups = append(e.manager.nodeGroups, nodeGroup)
	}

	if err := e.manager.Refresh(); err != nil {
		return nil, err
	}

	return nodeGroup, nil
}

// HasInstance returns whether a given node has a corresponding instance in this cloud provider
func (e *exoscaleCloudProvider) HasInstance(node *apiv1.Node) (bool, error) {
	return true, cloudprovider.ErrNotImplemented
}

// Pricing returns pricing model for this cloud provider or error if not available.
// Implementation optional.
func (e *exoscaleCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetAvailableMachineTypes get all machine types that can be requested from the cloud provider.
// Implementation optional.
func (e *exoscaleCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return []string{}, nil
}

// NewNodeGroup builds a theoretical node group based on the node definition provided. The node group is not automatically
// created on the cloud provider side. The node group is not returned by NodeGroups() until it is created.
// Implementation optional.
func (e *exoscaleCloudProvider) NewNodeGroup(
	_ string,
	_,
	_ map[string]string,
	_ []apiv1.Taint,
	_ map[string]resource.Quantity,
) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns struct containing limits (max, min) for resources (cores, memory etc.).
func (e *exoscaleCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return e.resourceLimiter, nil
}

// GPULabel returns the label added to nodes with GPU resource.
func (e *exoscaleCloudProvider) GPULabel() string {
	return ""
}

// GetAvailableGPUTypes return all available GPU types cloud provider supports.
func (e *exoscaleCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	return nil
}

// GetNodeGpuConfig returns the label, type and resource name for the GPU added to node. If node doesn't have
// any GPUs, it returns nil.
func (e *exoscaleCloudProvider) GetNodeGpuConfig(node *apiv1.Node) *cloudprovider.GpuConfig {
	return gpu.GetNodeGPUFromCloudProvider(e, node)
}

// Cleanup cleans up open resources before the cloud provider is destroyed, i.e. go routines etc.
func (e *exoscaleCloudProvider) Cleanup() error {
	return nil
}

// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
// In particular the list of node groups returned by NodeGroups can change as a result of CloudProvider.Refresh().
func (e *exoscaleCloudProvider) Refresh() error {
	debugf("refreshing node groups cache")
	return e.manager.Refresh()
}

// BuildExoscale builds the Exoscale cloud provider.
func BuildExoscale(_ config.AutoscalingOptions, discoveryOpts cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
	manager, err := newManager(discoveryOpts)
	if err != nil {
		fatalf("failed to initialize manager: %v", err)
	}

	// The cloud provider automatically uses all Instance Pools in the k8s cluster.
	// The flag '--nodes=1:5:nodepoolname' may be specified to limit the size of a nodepool.
	// The flag '--node-group-auto-discovery' is not implemented.
	provider, err := newExoscaleCloudProvider(manager, rl)
	if err != nil {
		fatalf("failed to create Exoscale cloud provider: %v", err)
	}

	return provider
}

func (e *exoscaleCloudProvider) instancePoolFromNode(node *apiv1.Node) (*egoscale.InstancePool, error) {
	nodeID := toNodeID(node.Spec.ProviderID)
	if nodeID == "" {
		// Some K8s deployments result in the master Node not having a provider ID set
		// (e.g. if it isn't managed by the Exoscale Cloud Controller Manager), therefore
		// if we detect the Node indeed has a master taint we skip it.
		var isMasterNode bool
		for _, taint := range node.Spec.Taints {
			if taint.Key == "node-role.kubernetes.io/master" {
				isMasterNode = true
				break
			}
		}
		if isMasterNode {
			return nil, errNoInstancePool
		}

		return nil, fmt.Errorf("unable to retrieve instance ID from Node %q", node.Spec.ProviderID)
	}

	debugf("looking up node group for node ID %s", nodeID)

	instance, err := e.manager.client.GetInstance(e.manager.ctx, e.manager.zone, nodeID)
	if err != nil {
		return nil, err
	}

	if instance.Manager == nil || instance.Manager.Type != "instance-pool" {
		return nil, errNoInstancePool
	}

	return e.manager.client.GetInstancePool(e.manager.ctx, e.manager.zone, instance.Manager.ID)
}
