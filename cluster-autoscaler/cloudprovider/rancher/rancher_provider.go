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

package rancher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	autoscalererrors "k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"

	provisioningv1 "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/rancher/provisioning.cattle.io/v1"
	klog "k8s.io/klog/v2"
)

const (
	// providerName is the cloud provider name for rancher
	providerName = "rancher"

	rancherProvisioningGroup       = "provisioning.cattle.io"
	rancherProvisioningVersion     = "v1"
	rancherLocalClusterPath        = "/k8s/clusters/local"
	rancherMachinePoolNameLabelKey = "rke.cattle.io/rke-machine-pool-name"

	minSizeAnnotation                  = "cluster.provisioning.cattle.io/autoscaler-min-size"
	maxSizeAnnotation                  = "cluster.provisioning.cattle.io/autoscaler-max-size"
	resourceCPUAnnotation              = "cluster.provisioning.cattle.io/autoscaler-resource-cpu"
	resourceMemoryAnnotation           = "cluster.provisioning.cattle.io/autoscaler-resource-memory"
	resourceEphemeralStorageAnnotation = "cluster.provisioning.cattle.io/autoscaler-resource-ephemeral-storage"
)

// RancherCloudProvider implements CloudProvider interface for rancher
type RancherCloudProvider struct {
	resourceLimiter *cloudprovider.ResourceLimiter
	client          dynamic.Interface
	nodeGroups      []*nodeGroup
	config          *cloudConfig
}

// BuildRancher builds rancher cloud provider.
func BuildRancher(opts config.AutoscalingOptions, _ cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
	provider, err := newRancherCloudProvider(opts.CloudConfig, rl)
	if err != nil {
		klog.Fatalf("failed to create rancher cloud provider: %v", err)
	}
	return provider
}

func newRancherCloudProvider(cloudConfig string, resourceLimiter *cloudprovider.ResourceLimiter) (*RancherCloudProvider, error) {
	config, err := newConfig(cloudConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create cloud config: %w", err)
	}

	restConfig := &rest.Config{
		Host:        config.URL,
		APIPath:     rancherLocalClusterPath,
		BearerToken: config.Token,
	}

	client, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create dynamic client: %w", err)
	}

	discovery, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create discovery client: %w", err)
	}

	if config.ClusterAPIVersion == "" {
		// automatically discover cluster API version
		clusterAPIVersion, err := getAPIGroupPreferredVersion(discovery, clusterAPIGroup)
		if err != nil {
			return nil, err
		}

		config.ClusterAPIVersion = clusterAPIVersion
	}

	return &RancherCloudProvider{
		resourceLimiter: resourceLimiter,
		client:          client,
		config:          config,
	}, nil
}

// Name returns name of the cloud provider.
func (provider *RancherCloudProvider) Name() string {
	return providerName
}

// GPULabel returns the label added to nodes with GPU resource.
func (provider *RancherCloudProvider) GPULabel() string {
	return ""
}

// GetAvailableGPUTypes return all available GPU types cloud provider supports
func (provider *RancherCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	// TODO: implement GPU support
	return nil
}

// GetNodeGpuConfig returns the label, type and resource name for the GPU added to node. If node doesn't have
// any GPUs, it returns nil.
func (provider *RancherCloudProvider) GetNodeGpuConfig(node *corev1.Node) *cloudprovider.GpuConfig {
	return gpu.GetNodeGPUFromCloudProvider(provider, node)
}

// NodeGroups returns all node groups configured for this cloud provider.
func (provider *RancherCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	nodeGroups := make([]cloudprovider.NodeGroup, len(provider.nodeGroups))
	for i, ng := range provider.nodeGroups {
		nodeGroups[i] = ng
	}
	return nodeGroups
}

// Pricing returns pricing model for this cloud provider or error if not available.
func (provider *RancherCloudProvider) Pricing() (cloudprovider.PricingModel, autoscalererrors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

// NodeGroupForNode returns the node group for the given node.
func (provider *RancherCloudProvider) NodeGroupForNode(node *corev1.Node) (cloudprovider.NodeGroup, error) {
	machineName, ok := node.Annotations[machineNodeAnnotationKey]
	if !ok {
		klog.V(4).Infof("skipping NodeGroupForNode %q as the annotation %q is missing", node.Name, machineNodeAnnotationKey)
		return nil, nil
	}

	for _, group := range provider.nodeGroups {
		machine, err := group.machineByName(machineName)
		if err != nil {
			klog.V(6).Infof("node %q is not part of node group %q", node.Name, group.name)
			continue
		}

		pool, ok := machine.GetLabels()[rancherMachinePoolNameLabelKey]
		if !ok {
			return nil, fmt.Errorf("machine %q is missing the label %q", machine.GetName(), rancherMachinePoolNameLabelKey)
		}

		klog.V(4).Infof("found pool %q via machine %q", pool, machine.GetName())

		if group.name == pool {
			return group, nil
		}
	}

	// if node is not in one of our scalable nodeGroups, we return nil so it
	// won't be processed further by the CA.
	return nil, nil
}

// HasInstance returns whether a given node has a corresponding instance in this cloud provider
func (provider *RancherCloudProvider) HasInstance(node *corev1.Node) (bool, error) {
	return true, cloudprovider.ErrNotImplemented
}

// GetAvailableMachineTypes get all machine types that can be requested from the cloud provider.
// Implementation optional.
func (provider *RancherCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return []string{}, cloudprovider.ErrNotImplemented
}

// NewNodeGroup builds a theoretical node group based on the node definition provided.
func (provider *RancherCloudProvider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string,
	taints []corev1.Taint,
	extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns struct containing limits (max, min) for resources (cores, memory etc.).
func (provider *RancherCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return provider.resourceLimiter, nil
}

// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
// In particular the list of node groups returned by NodeGroups can change as a result of CloudProvider.Refresh().
func (provider *RancherCloudProvider) Refresh() error {
	nodeGroups, err := provider.scalableNodeGroups()
	if err != nil {
		return fmt.Errorf("unable to get node groups from cluster: %w", err)
	}

	provider.nodeGroups = nodeGroups
	return nil
}

// Cleanup cleans up all resources before the cloud provider is removed
func (provider *RancherCloudProvider) Cleanup() error {
	return nil
}

func (provider *RancherCloudProvider) scalableNodeGroups() ([]*nodeGroup, error) {
	var result []*nodeGroup

	pools, err := provider.getMachinePools()
	if err != nil {
		return nil, err
	}

	for _, pool := range pools {
		nodeGroup, err := newNodeGroupFromMachinePool(provider, pool)
		if err != nil {
			if isNotScalable(err) {
				klog.V(4).Infof("ignoring machine pool %s as it does not have min/max annotations", pool.Name)
				continue
			}

			return nil, fmt.Errorf("error getting node group from machine pool: %w", err)
		}

		klog.V(4).Infof("scalable node group found: %s", nodeGroup.Debug())

		result = append(result, nodeGroup)
	}

	return result, err
}

func clusterGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    rancherProvisioningGroup,
		Version:  rancherProvisioningVersion,
		Resource: "clusters",
	}
}

func (provider *RancherCloudProvider) getMachinePools() ([]provisioningv1.RKEMachinePool, error) {
	res, err := provider.client.Resource(clusterGVR()).
		Namespace(provider.config.ClusterNamespace).
		Get(context.TODO(), provider.config.ClusterName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting cluster: %w", err)
	}

	machinePools, ok, err := unstructured.NestedFieldNoCopy(res.Object, "spec", "rkeConfig", "machinePools")
	if !ok {
		return nil, fmt.Errorf("unable to find machinePools of cluster %s", provider.config.ClusterName)
	}
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(machinePools)
	if err != nil {
		return nil, err
	}

	var pools []provisioningv1.RKEMachinePool
	err = json.Unmarshal(data, &pools)

	return pools, err
}

func (provider *RancherCloudProvider) updateMachinePools(machinePools []provisioningv1.RKEMachinePool) error {
	cluster, err := provider.client.Resource(clusterGVR()).
		Namespace(provider.config.ClusterNamespace).
		Get(context.TODO(), provider.config.ClusterName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error getting cluster: %w", err)
	}

	pools, err := machinePoolsToUnstructured(machinePools)
	if err != nil {
		return err
	}

	if err := unstructured.SetNestedSlice(cluster.Object, pools, "spec", "rkeConfig", "machinePools"); err != nil {
		return err
	}

	_, err = provider.client.Resource(clusterGVR()).Namespace(provider.config.ClusterNamespace).
		Update(context.TODO(), &unstructured.Unstructured{Object: cluster.Object}, metav1.UpdateOptions{})
	return err
}

// converts machinePools into a usable form for the unstructured client.
// unstructured.SetNestedSlice expects types produced by json.Unmarshal(),
// so we marshal and unmarshal again before passing it on.
func machinePoolsToUnstructured(machinePools []provisioningv1.RKEMachinePool) ([]interface{}, error) {
	data, err := json.Marshal(machinePools)
	if err != nil {
		return nil, err
	}

	var pools []interface{}
	if err := json.Unmarshal(data, &pools); err != nil {
		return nil, err
	}

	return pools, nil
}

func isNotScalable(err error) bool {
	return errors.Is(err, errMissingMinSizeAnnotation) || errors.Is(err, errMissingMaxSizeAnnotation)
}
