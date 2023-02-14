/*
Copyright 2021 Oracle and/or its affiliates.

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

package oci

import (
	"time"

	"github.com/pkg/errors"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	caerrors "k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

const (
	ociUseInstancePrincipalEnvVar       = "OCI_USE_INSTANCE_PRINCIPAL"
	ociUseNonPoolMemberAnnotationEnvVar = "OCI_USE_NON_POOL_MEMBER_ANNOTATION"
	ociCompartmentEnvVar                = "OCI_COMPARTMENT_ID"
	ociRegionEnvVar                     = "OCI_REGION"
	ociRefreshInterval                  = "OCI_REFRESH_INTERVAL"
	ociAnnotationCompartmentID          = "oci.oraclecloud.com/compartment-id"
	// ResourceGPU is the GPU resource type
	ResourceGPU            apiv1.ResourceName = "nvidia.com/gpu"
	defaultRefreshInterval                    = 5 * time.Minute
)

// OciCloudProvider implements the CloudProvider interface for OCI. It contains an
// instance pool manager to interact with OCI instance pools.
type OciCloudProvider struct {
	rl          *cloudprovider.ResourceLimiter
	poolManager InstancePoolManager
}

// CloudConfig holds the cloud config for OCI provider.
type CloudConfig struct {
	Global struct {
		RefreshInterval        time.Duration `gcfg:"refresh-interval"`
		CompartmentID          string        `gcfg:"compartment-id"`
		Region                 string        `gcfg:"region"`
		UseInstancePrinciples  bool          `gcfg:"use-instance-principals"`
		UseNonMemberAnnotation bool          `gcfg:"use-non-member-annotation"`
	}
}

// Name returns name of the cloud provider.
func (ocp *OciCloudProvider) Name() string {
	return cloudprovider.OracleCloudProviderName
}

// NodeGroups returns all node groups configured for this cloud provider.
func (ocp *OciCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	nodePools := ocp.poolManager.GetInstancePools()
	result := make([]cloudprovider.NodeGroup, 0, len(nodePools))
	for _, nodePool := range nodePools {
		result = append(result, nodePool)
	}
	return result
}

// NodeGroupForNode returns the node group for the given node, nil if the node
// should not be processed by cluster autoscaler, or non-nil error if such
// occurred. Must be implemented.
func (ocp *OciCloudProvider) NodeGroupForNode(n *apiv1.Node) (cloudprovider.NodeGroup, error) {

	ociRef, err := nodeToOciRef(n)
	if err != nil {
		return nil, err
	}

	ng, err := ocp.poolManager.GetInstancePoolForInstance(ociRef)

	// this instance may not be a part of an instance pool, or it may be part of a instance pool that the autoscaler does not manage
	if errors.Cause(err) == errInstanceInstancePoolNotFound {
		// should not be processed by cluster autoscaler
		return nil, nil
	}

	return ng, err
}

// HasInstance returns whether a given node has a corresponding instance in this cloud provider
func (ocp *OciCloudProvider) HasInstance(n *apiv1.Node) (bool, error) {
	return true, cloudprovider.ErrNotImplemented
}

// Pricing returns pricing model for this cloud provider or error if not available.
// Implementation optional.
func (ocp *OciCloudProvider) Pricing() (cloudprovider.PricingModel, caerrors.AutoscalerError) {
	klog.Info("Pricing called")
	return nil, cloudprovider.ErrNotImplemented
}

// GetAvailableMachineTypes getInstancePool all machine types that can be requested from the cloud provider.
// Implementation optional.
func (ocp *OciCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	klog.Info("GetAvailableMachineTypes called")
	return nil, cloudprovider.ErrNotImplemented
}

// NewNodeGroup builds a theoretical node group based on the node definition provided. The node group is not automatically
// created on the cloud provider side. The node group is not returned by NodeGroups() until it is created.
// Implementation optional.
func (ocp *OciCloudProvider) NewNodeGroup(machineType string,
	labels map[string]string,
	systemLabels map[string]string,
	taints []apiv1.Taint,
	extraResources map[string]resource.Quantity,
) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns struct containing limits (max, min) for resources (cores, memory etc.).
func (ocp *OciCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return ocp.rl, nil
}

// GPULabel returns the label added to nodes with GPU resource.
func (ocp *OciCloudProvider) GPULabel() string {
	// No labels, only taint: nvidia.com/gpu:NoSchedule
	return ""
}

// GetAvailableGPUTypes return all available GPU types cloud provider supports.
func (ocp *OciCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	return map[string]struct{}{}
}

// GetNodeGpuConfig returns the label, type and resource name for the GPU added to node. If node doesn't have
// any GPUs, it returns nil.
func (ocp *OciCloudProvider) GetNodeGpuConfig(node *apiv1.Node) *cloudprovider.GpuConfig {
	return gpu.GetNodeGPUFromCloudProvider(ocp, node)
}

// Cleanup cleans up open resources before the cloud provider is destroyed, i.e. go routines etc.
func (ocp *OciCloudProvider) Cleanup() error {
	return ocp.poolManager.Cleanup()
}

// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
// In particular the list of node groups returned by NodeGroups can change as a result of CloudProvider.Refresh().
func (ocp *OciCloudProvider) Refresh() error {
	return ocp.poolManager.Refresh()
}

// BuildOCI constructs the OciCloudProvider object that implements the could provider interface (InstancePoolManager).
func BuildOCI(opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) *OciCloudProvider {

	ipManager, err := CreateInstancePoolManager(opts.CloudConfig, do, createKubeClient(opts))
	if err != nil {
		klog.Fatalf("Could not create OCI cloud provider: %v", err)
	}
	return &OciCloudProvider{
		poolManager: ipManager,
		rl:          rl,
	}
}

func getKubeConfig(opts config.AutoscalingOptions) *rest.Config {
	klog.V(1).Infof("Using kubeconfig file: %s", opts.KubeConfigPath)
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", opts.KubeConfigPath)
	if err != nil {
		klog.Fatalf("Failed to build kubeConfig: %v", err)
	}

	return kubeConfig
}

func createKubeClient(opts config.AutoscalingOptions) kubernetes.Interface {
	return kubernetes.NewForConfigOrDie(getKubeConfig(opts))
}
