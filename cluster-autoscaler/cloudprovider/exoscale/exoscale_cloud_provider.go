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

package exoscale

import (
	err "errors"
	"fmt"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/exoscale/internal/github.com/exoscale/egoscale"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/exoscale/internal/k8s.io/klog"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
)

var _ cloudprovider.CloudProvider = (*exoscaleCloudProvider)(nil)

var errNoInstancePool = err.New("not an Instance Pool member")

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
	nodeGroups := make([]cloudprovider.NodeGroup, len(e.manager.nodeGroups))
	for i, ng := range e.manager.nodeGroups {
		nodeGroups[i] = ng
	}
	return nodeGroups
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

	nodeGroup := &NodeGroup{
		id:           instancePool.ID.String(),
		manager:      e.manager,
		instancePool: instancePool,
	}

	found := false
	for i, ng := range e.manager.nodeGroups {
		if ng.id == nodeGroup.id {
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
func (e *exoscaleCloudProvider) NewNodeGroup(_ string, _, _ map[string]string, _ []apiv1.Taint, _ map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
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

// Cleanup cleans up open resources before the cloud provider is destroyed, i.e. go routines etc.
func (e *exoscaleCloudProvider) Cleanup() error {
	return nil
}

// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
// In particular the list of node groups returned by NodeGroups can change as a result of CloudProvider.Refresh().
func (e *exoscaleCloudProvider) Refresh() error {
	klog.V(4).Info("Refreshing node group cache")
	return e.manager.Refresh()
}

// BuildExoscale builds the Exoscale cloud provider.
func BuildExoscale(_ config.AutoscalingOptions, _ cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
	manager, err := newManager()
	if err != nil {
		klog.Fatalf("Failed to create Exoscale manager: %v", err)
	}

	// The cloud provider automatically uses all Instance Pools in the k8s cluster.
	// This means we don't use the cloudprovider.NodeGroupDiscoveryOptions
	// flags (which can be set via '--node-group-auto-discovery' or '-nodes')
	provider, err := newExoscaleCloudProvider(manager, rl)
	if err != nil {
		klog.Fatalf("Failed to create Exoscale cloud provider: %v", err)
	}

	return provider
}

// toProviderID returns a provider ID from the given node ID.
func toProviderID(nodeID string) string {
	return fmt.Sprintf("%s%s", exoscaleProviderIDPrefix, nodeID)
}

// toNodeID returns a node or Compute instance ID from the given provider ID.
func toNodeID(providerID string) string {
	return strings.TrimPrefix(providerID, exoscaleProviderIDPrefix)
}

func (e *exoscaleCloudProvider) instancePoolFromNode(node *apiv1.Node) (*egoscale.InstancePool, error) {
	providerID := node.Spec.ProviderID
	nodeID := toNodeID(providerID)

	klog.V(4).Infof("Looking up node group for node ID %q", nodeID)

	id, err := egoscale.ParseUUID(nodeID)
	if err != nil {
		return nil, err
	}

	resp, err := e.manager.client.Get(egoscale.VirtualMachine{ID: id})
	if err != nil {
		if csError, ok := err.(*egoscale.ErrorResponse); ok && csError.ErrorCode == egoscale.ParamError {
			return nil, errNoInstancePool
		}

		return nil, err
	}

	instance := resp.(*egoscale.VirtualMachine)

	if instance.Manager != "instancepool" {
		return nil, errNoInstancePool
	}

	zone, err := e.zoneFromNode(node)
	if err != nil {
		return nil, err
	}

	resp, err = e.manager.client.Request(egoscale.GetInstancePool{
		ID:     instance.ManagerID,
		ZoneID: zone.ID,
	})
	if csError, ok := err.(*egoscale.ErrorResponse); ok && csError.ErrorCode == egoscale.NotFound {
		return nil, errNoInstancePool
	} else if err != nil {
		return nil, err
	}

	return &resp.(*egoscale.GetInstancePoolResponse).InstancePools[0], nil
}

func (e *exoscaleCloudProvider) zoneFromNode(node *apiv1.Node) (*egoscale.Zone, error) {
	zoneName, ok := node.Labels["topology.kubernetes.io/region"]
	if !ok {
		return nil, fmt.Errorf("zone not found")
	}

	resp, err := e.manager.client.Get(egoscale.Zone{Name: zoneName})
	if err != nil {
		return nil, err
	}

	return resp.(*egoscale.Zone), nil
}
