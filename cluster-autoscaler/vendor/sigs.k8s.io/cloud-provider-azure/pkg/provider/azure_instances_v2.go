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

package provider

import (
	"context"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	cloudprovider "k8s.io/cloud-provider"
	"k8s.io/klog/v2"
)

var _ cloudprovider.InstancesV2 = (*Cloud)(nil)

// InstanceExists returns true if the instance for the given node exists according to the cloud provider.
// Use the node.name or node.spec.providerID field to find the node in the cloud provider.
func (az *Cloud) InstanceExists(ctx context.Context, node *v1.Node) (bool, error) {
	if node == nil {
		return false, nil
	}
	unmanaged, err := az.IsNodeUnmanaged(node.Name)
	if err != nil {
		return false, err
	}
	if unmanaged {
		klog.V(4).Infof("InstanceExists: omitting unmanaged node %q", node.Name)
		return true, nil
	}

	providerID := node.Spec.ProviderID
	if providerID == "" {
		var err error
		providerID, err = cloudprovider.GetInstanceProviderID(ctx, az, types.NodeName(node.Name))
		if err != nil {
			if strings.Contains(err.Error(), cloudprovider.InstanceNotFound.Error()) {
				return false, nil
			}

			klog.Errorf("InstanceExists: failed to get the provider ID by node name %s: %v", node.Name, err)
			return false, err
		}
	}

	return az.InstanceExistsByProviderID(ctx, providerID)
}

// InstanceShutdown returns true if the instance is shutdown according to the cloud provider.
// Use the node.name or node.spec.providerID field to find the node in the cloud provider.
func (az *Cloud) InstanceShutdown(ctx context.Context, node *v1.Node) (bool, error) {
	if node == nil {
		return false, nil
	}
	unmanaged, err := az.IsNodeUnmanaged(node.Name)
	if err != nil {
		return false, err
	}
	if unmanaged {
		klog.V(4).Infof("InstanceShutdown: omitting unmanaged node %q", node.Name)
		return false, nil
	}
	providerID := node.Spec.ProviderID
	if providerID == "" {
		var err error
		providerID, err = cloudprovider.GetInstanceProviderID(ctx, az, types.NodeName(node.Name))
		if err != nil {
			// Returns false, so the controller manager will continue to check InstanceExistsByProviderID().
			if strings.Contains(err.Error(), cloudprovider.InstanceNotFound.Error()) {
				return false, nil
			}

			klog.Errorf("InstanceShutdown: failed to get the provider ID by node name %s: %v", node.Name, err)
			return false, err
		}
	}

	return az.InstanceShutdownByProviderID(ctx, providerID)
}

// InstanceMetadata returns the instance's metadata. The values returned in InstanceMetadata are
// translated into specific fields in the Node object on registration.
// Use the node.name or node.spec.providerID field to find the node in the cloud provider.
func (az *Cloud) InstanceMetadata(ctx context.Context, node *v1.Node) (*cloudprovider.InstanceMetadata, error) {
	meta := cloudprovider.InstanceMetadata{}
	if node == nil {
		return &meta, nil
	}
	unmanaged, err := az.IsNodeUnmanaged(node.Name)
	if err != nil {
		return &meta, err
	}
	if unmanaged {
		klog.V(4).Infof("InstanceMetadata: omitting unmanaged node %q", node.Name)
		return &meta, nil
	}

	if node.Spec.ProviderID != "" {
		meta.ProviderID = node.Spec.ProviderID
	} else {
		providerID, err := cloudprovider.GetInstanceProviderID(ctx, az, types.NodeName(node.Name))
		if err != nil {
			klog.Errorf("InstanceMetadata: failed to get the provider ID by node name %s: %v", node.Name, err)
			return nil, err
		}
		meta.ProviderID = providerID
	}

	instanceType, err := az.InstanceType(ctx, types.NodeName(node.Name))
	if err != nil {
		klog.Errorf("InstanceMetadata: failed to get the instance type of %s: %v", node.Name, err)
		return &cloudprovider.InstanceMetadata{}, err
	}
	meta.InstanceType = instanceType

	nodeAddresses, err := az.NodeAddresses(ctx, types.NodeName(node.Name))
	if err != nil {
		klog.Errorf("InstanceMetadata: failed to get the node address of %s: %v", node.Name, err)
		return &cloudprovider.InstanceMetadata{}, err
	}
	meta.NodeAddresses = nodeAddresses

	zone, err := az.GetZoneByNodeName(ctx, types.NodeName(node.Name))
	if err != nil {
		klog.Errorf("InstanceMetadata: failed to get the node zone of %s: %v", node.Name, err)
		return &cloudprovider.InstanceMetadata{}, err
	}
	meta.Zone = zone.FailureDomain
	meta.Region = zone.Region

	return &meta, nil
}
