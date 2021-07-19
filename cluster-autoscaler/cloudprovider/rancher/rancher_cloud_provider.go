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

package rancher

import (
	"errors"
	"fmt"
	"os"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	caerrors "k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/klog/v2"
)

var (
	availableGPUTypes map[string]struct{}
)

var (
	errAutodiscoveryNotSupported = errors.New("only support static discovery scaling group for now")
	errNodeSpecsCannotEmpty      = errors.New("node group specs must be specified")
)

// rancherProvider implements CloudProvider interface.
type rancherProvider struct {
	manager         *manager
	nodePools       []*NodePool
	resourceLimiter *cloudprovider.ResourceLimiter
}

func newRancherProvider(manager *manager, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) (*rancherProvider, error) {
	klog.V(5).Info("Build rancher autoscaler CloudProvider")
	if do.AutoDiscoverySpecified() {
		return nil, errAutodiscoveryNotSupported
	}

	if len(do.NodeGroupSpecs) == 0 {
		return nil, errNodeSpecsCannotEmpty
	}

	clusterNPS, err := manager.getNodePools()
	if err != nil {
		return nil, fmt.Errorf("error trying to get nodePools from rancher: %w", err)
	}

	nps := make([]*NodePool, len(do.NodeGroupSpecs))
	for i, spec := range do.NodeGroupSpecs {
		ns, err := dynamic.SpecFromString(spec, false)
		if err != nil {
			return nil, fmt.Errorf("failed to parse node group spec: %v", err)
		}

		np, ok := clusterNPS[ns.Name]
		if !ok {
			return nil, fmt.Errorf("nodePool: %s does not exist", ns.Name)
		}

		if !np.DrainBeforeDelete {
			return nil, fmt.Errorf("nodePool: %s must have DrainBeforeDelete enabled", ns.Name)
		}

		nps[i] = &NodePool{
			manager:   manager,
			id:        ns.Name,
			minSize:   ns.MinSize,
			maxSize:   ns.MaxSize,
			rancherNP: np,
		}
		klog.Infof("Inserting NodePool: %#v", nps[i])
	}

	return &rancherProvider{
		manager:         manager,
		nodePools:       nps,
		resourceLimiter: rl,
	}, nil
}

// Name returns name of the cloud provider.
func (u rancherProvider) Name() string {
	return cloudprovider.RancherProviderName
}

// NodeGroups returns all node pools configured for this cloud provider.
func (u rancherProvider) NodeGroups() []cloudprovider.NodeGroup {
	nodeGroups := make([]cloudprovider.NodeGroup, len(u.nodePools))
	for i, ng := range u.nodePools {
		nodeGroups[i] = ng
	}
	return nodeGroups
}

// NodeGroupForNode returns the node group for the given node, nil if the node
// should not be processed by cluster autoscaler, or non-nil error if such
// occurred. Must be implemented.
func (u rancherProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	np, err := u.manager.getNode(node)
	if err != nil {
		return nil, err
	}

	klog.V(4).Infof("checking node pool for node: %s - %q", node.Name, np.ID)
	for _, pool := range u.nodePools {
		if np.NodePoolID == pool.id {
			klog.Infof("Node %q belongs to NodePool %q", np.ID, pool.id)
			return pool, nil
		}
	}

	return nil, nil
}

// Pricing returns pricing model for this cloud provider or error if not available.
// Implementation optional.
func (u rancherProvider) Pricing() (cloudprovider.PricingModel, caerrors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetAvailableMachineTypes get all machine types that can be requested from the cloud provider.
// Implementation optional.
func (u rancherProvider) GetAvailableMachineTypes() ([]string, error) {
	return []string{}, nil
}

// NewNodeGroup builds a theoretical node group based on the node definition provided. The node group is not automatically
// created on the cloud provider side. The node group is not returned by NodeGroups() until it is created.
// Implementation optional.
func (u rancherProvider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string, taints []apiv1.Taint, extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns struct containing limits (max, min) for resources (cores, memory etc.).
func (u rancherProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return u.resourceLimiter, nil
}

// GPULabel returns the label added to nodes with GPU resource.
func (u rancherProvider) GPULabel() string {
	return "gpu-image"
}

// GetAvailableGPUTypes return all available GPU types cloud provider supports.
func (u rancherProvider) GetAvailableGPUTypes() map[string]struct{} {
	return availableGPUTypes
}

// Cleanup cleans up open resources before the cloud provider is destroyed, i.e. go routines etc.
func (u rancherProvider) Cleanup() error {
	return nil
}

// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
// In particular the list of node groups returned by NodeGroups can change as a result of CloudProvider.Refresh().
func (u rancherProvider) Refresh() error {
	klog.V(4).Info("Refreshing node group")
	clusterNPS, err := u.manager.getNodePools()
	if err != nil {
		return err
	}

	for _, np := range u.nodePools {
		clusterNode, ok := clusterNPS[np.id]
		if !ok {
			return fmt.Errorf("nodePool: %q does not exist", np.id)
		}

		if !clusterNode.DrainBeforeDelete {
			return fmt.Errorf("nodePool: %q must have DrainBeforeDelete enabled", clusterNode.Name)
		}

		np.rancherNP = clusterNode
	}

	return nil
}

// BuildRancher builds the Rancher cloud provider.
func BuildRancher(opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
	configFile, err := os.Open(opts.CloudConfig)
	if err != nil {
		klog.Fatalf("Couldn't open cloud provider configuration %s: %#v", opts.CloudConfig, err)
	}
	defer configFile.Close()

	manager, err := newManager(configFile)
	if err != nil {
		klog.Fatalf("Failed to create rancher manager: %v", err)
	}

	if _, err := manager.client.ClusterByID(manager.clusterID); err != nil {
		klog.Fatalf("Failed to create rancher provider: error retrieving cluster %v", err)
	}

	provider, err := newRancherProvider(manager, do, rl)
	if err != nil {
		klog.Fatalf("Failed to create rancher provider: %v", err)
	}

	return provider
}
