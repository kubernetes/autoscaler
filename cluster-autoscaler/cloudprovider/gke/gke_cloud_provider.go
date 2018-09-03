/*
Copyright 2016 The Kubernetes Authors.

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

package gke

import (
	"fmt"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/gce"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	schedulercache "k8s.io/kubernetes/pkg/scheduler/cache"
)

const (
	// ProviderNameGKE is the name of GKE cloud provider.
	ProviderNameGKE = "gke"
)

const (
	maxAutoprovisionedSize = 1000
	minAutoprovisionedSize = 0
)

// Big machines are temporarily commented out.
// TODO(mwielgus): get this list programmatically
var autoprovisionedMachineTypes = []string{
	"n1-standard-1",
	"n1-standard-2",
	"n1-standard-4",
	"n1-standard-8",
	"n1-standard-16",
	//"n1-standard-32",
	//"n1-standard-64",
	"n1-highcpu-2",
	"n1-highcpu-4",
	"n1-highcpu-8",
	"n1-highcpu-16",
	//"n1-highcpu-32",
	// "n1-highcpu-64",
	"n1-highmem-2",
	"n1-highmem-4",
	"n1-highmem-8",
	"n1-highmem-16",
	//"n1-highmem-32",
	//"n1-highmem-64",
}

// GkeCloudProvider implements CloudProvider interface.
type GkeCloudProvider struct {
	gkeManager GkeManager
	// This resource limiter is used if resource limits are not defined through cloud API.
	resourceLimiterFromFlags *cloudprovider.ResourceLimiter
}

// BuildGkeCloudProvider builds CloudProvider implementation for GKE.
func BuildGkeCloudProvider(gkeManager GkeManager, resourceLimiter *cloudprovider.ResourceLimiter) (*GkeCloudProvider, error) {
	return &GkeCloudProvider{gkeManager: gkeManager, resourceLimiterFromFlags: resourceLimiter}, nil
}

// Cleanup cleans up all resources before the cloud provider is removed
func (gke *GkeCloudProvider) Cleanup() error {
	gke.gkeManager.Cleanup()
	return nil
}

// Name returns name of the cloud provider.
func (gke *GkeCloudProvider) Name() string {
	return ProviderNameGKE
}

// NodeGroups returns all node groups configured for this cloud provider.
func (gke *GkeCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	migs := gke.gkeManager.GetMigs()
	result := make([]cloudprovider.NodeGroup, 0, len(migs))
	for _, mig := range migs {
		result = append(result, mig.Config)
	}
	return result
}

// NodeGroupForNode returns the node group for the given node.
func (gke *GkeCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	ref, err := gce.GceRefFromProviderId(node.Spec.ProviderID)
	if err != nil {
		return nil, err
	}
	mig, err := gke.gkeManager.GetMigForInstance(ref)
	return mig, err
}

// Pricing returns pricing model for this cloud provider or error if not available.
func (gke *GkeCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	return &gce.GcePriceModel{}, nil
}

// GetAvailableMachineTypes get all machine types that can be requested from the cloud provider.
func (gke *GkeCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return autoprovisionedMachineTypes, nil
}

// NewNodeGroup builds a theoretical node group based on the node definition provided. The node group is not automatically
// created on the cloud provider side. The node group is not returned by NodeGroups() until it is created.
func (gke *GkeCloudProvider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string,
	taints []apiv1.Taint, extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	nodePoolName := fmt.Sprintf("%s-%s-%d", nodeAutoprovisioningPrefix, machineType, time.Now().Unix())
	// TODO(aleksandra-malinowska): GkeManager's location will be a region
	// for regional clusters. We should support regional clusters by looking at
	// node locations instead.
	zone := gke.gkeManager.GetLocation()

	if gpuRequest, found := extraResources[gpu.ResourceNvidiaGPU]; found {
		gpuType, found := systemLabels[gpu.GPULabel]
		if !found {
			return nil, cloudprovider.ErrIllegalConfiguration
		}
		gpuCount, err := getNormalizedGpuCount(gpuRequest.Value())
		if err != nil {
			return nil, err
		}
		extraResources[gpu.ResourceNvidiaGPU] = *resource.NewQuantity(gpuCount, resource.DecimalSI)
		err = validateGpuConfig(gpuType, gpuCount, zone, machineType)
		if err != nil {
			return nil, err
		}
		nodePoolName = fmt.Sprintf("%s-%s-gpu-%d", nodeAutoprovisioningPrefix, machineType, time.Now().Unix())
		labels[gpu.GPULabel] = gpuType

		taint := apiv1.Taint{
			Effect: apiv1.TaintEffectNoSchedule,
			Key:    gpu.ResourceNvidiaGPU,
			Value:  "present",
		}
		taints = append(taints, taint)
	}

	mig := &GkeMig{
		gceRef: gce.GceRef{
			Project: gke.gkeManager.GetProjectId(),
			Zone:    zone,
			Name:    nodePoolName + "-temporary-mig",
		},
		gkeManager:      gke.gkeManager,
		autoprovisioned: true,
		exist:           false,
		nodePoolName:    nodePoolName,
		minSize:         minAutoprovisionedSize,
		maxSize:         maxAutoprovisionedSize,
		spec: &MigSpec{
			MachineType:    machineType,
			Labels:         labels,
			Taints:         taints,
			ExtraResources: extraResources,
		},
	}

	// Try to build a node from autoprovisioning spec. We don't need one right now,
	// but if it fails later, we'd end up with a node group we can't scale anyway,
	// so there's no point creating it.
	if _, err := gke.gkeManager.GetMigTemplateNode(mig); err != nil {
		return nil, fmt.Errorf("Failed to build node from spec: %v", err)
	}

	return mig, nil
}

// GetResourceLimiter returns struct containing limits (max, min) for resources (cores, memory etc.).
func (gke *GkeCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	resourceLimiter, err := gke.gkeManager.GetResourceLimiter()
	if err != nil {
		return nil, err
	}
	if resourceLimiter != nil {
		return resourceLimiter, nil
	}
	return gke.resourceLimiterFromFlags, nil
}

// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
// In particular the list of node groups returned by NodeGroups can change as a result of CloudProvider.Refresh().
func (gke *GkeCloudProvider) Refresh() error {
	return gke.gkeManager.Refresh()
}

// GetClusterInfo returns the project id, location and cluster name.
func (gke *GkeCloudProvider) GetClusterInfo() (projectId, location, clusterName string) {
	return gke.gkeManager.GetProjectId(), gke.gkeManager.GetLocation(), gke.gkeManager.GetClusterName()
}

// GetNodeLocations returns the list of zones in which the cluster has nodes.
func (gke *GkeCloudProvider) GetNodeLocations() []string {
	return gke.gkeManager.GetNodeLocations()
}

// MigSpec contains information about what machines in a MIG look like.
type MigSpec struct {
	MachineType    string
	Labels         map[string]string
	Taints         []apiv1.Taint
	ExtraResources map[string]resource.Quantity
}

// GkeMig represents the GKE Managed Instance Group implementation of a NodeGroup.
type GkeMig struct {
	gceRef gce.GceRef

	gkeManager      GkeManager
	minSize         int
	maxSize         int
	autoprovisioned bool
	exist           bool
	nodePoolName    string
	spec            *MigSpec
}

// GceRef returns Mig's GceRef
func (mig *GkeMig) GceRef() gce.GceRef {
	return mig.gceRef
}

// NodePoolName returns the name of the GKE node pool this Mig belongs to.
func (mig *GkeMig) NodePoolName() string {
	return mig.nodePoolName
}

// Spec returns specification of the Mig.
func (mig *GkeMig) Spec() *MigSpec {
	return mig.spec
}

// MaxSize returns maximum size of the node group.
func (mig *GkeMig) MaxSize() int {
	return mig.maxSize
}

// MinSize returns minimum size of the node group.
func (mig *GkeMig) MinSize() int {
	return mig.minSize
}

// TargetSize returns the current TARGET size of the node group. It is possible that the
// number is different from the number of nodes registered in Kubernetes.
func (mig *GkeMig) TargetSize() (int, error) {
	if !mig.exist {
		return 0, nil
	}
	size, err := mig.gkeManager.GetMigSize(mig)
	return int(size), err
}

// IncreaseSize increases Mig size
func (mig *GkeMig) IncreaseSize(delta int) error {
	if delta <= 0 {
		return fmt.Errorf("size increase must be positive")
	}
	size, err := mig.gkeManager.GetMigSize(mig)
	if err != nil {
		return err
	}
	if int(size)+delta > mig.MaxSize() {
		return fmt.Errorf("size increase too large - desired:%d max:%d", int(size)+delta, mig.MaxSize())
	}
	return mig.gkeManager.SetMigSize(mig, size+int64(delta))
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
func (mig *GkeMig) DecreaseTargetSize(delta int) error {
	if delta >= 0 {
		return fmt.Errorf("size decrease must be negative")
	}
	size, err := mig.gkeManager.GetMigSize(mig)
	if err != nil {
		return err
	}
	nodes, err := mig.gkeManager.GetMigNodes(mig)
	if err != nil {
		return err
	}
	if int(size)+delta < len(nodes) {
		return fmt.Errorf("attempt to delete existing nodes targetSize:%d delta:%d existingNodes: %d",
			size, delta, len(nodes))
	}
	return mig.gkeManager.SetMigSize(mig, size+int64(delta))
}

// Belongs returns true if the given node belongs to the NodeGroup.
func (mig *GkeMig) Belongs(node *apiv1.Node) (bool, error) {
	ref, err := gce.GceRefFromProviderId(node.Spec.ProviderID)
	if err != nil {
		return false, err
	}
	targetMig, err := mig.gkeManager.GetMigForInstance(ref)
	if err != nil {
		return false, err
	}
	if targetMig == nil {
		return false, fmt.Errorf("%s doesn't belong to a known mig", node.Name)
	}
	if targetMig.Id() != mig.Id() {
		return false, nil
	}
	return true, nil
}

// DeleteNodes deletes the nodes from the group.
func (mig *GkeMig) DeleteNodes(nodes []*apiv1.Node) error {
	size, err := mig.gkeManager.GetMigSize(mig)
	if err != nil {
		return err
	}
	if int(size) <= mig.MinSize() {
		return fmt.Errorf("min size reached, nodes will not be deleted")
	}
	refs := make([]*gce.GceRef, 0, len(nodes))
	for _, node := range nodes {

		belongs, err := mig.Belongs(node)
		if err != nil {
			return err
		}
		if !belongs {
			return fmt.Errorf("%s belong to a different mig than %s", node.Name, mig.Id())
		}
		gceref, err := gce.GceRefFromProviderId(node.Spec.ProviderID)
		if err != nil {
			return err
		}
		refs = append(refs, gceref)
	}
	return mig.gkeManager.DeleteInstances(refs)
}

// Id returns mig url.
func (mig *GkeMig) Id() string {
	return gce.GenerateMigUrl(mig.gceRef)
}

// Debug returns a debug string for the Mig.
func (mig *GkeMig) Debug() string {
	return fmt.Sprintf("%s (%d:%d)", mig.Id(), mig.MinSize(), mig.MaxSize())
}

// Nodes returns a list of all nodes that belong to this node group.
func (mig *GkeMig) Nodes() ([]string, error) {
	return mig.gkeManager.GetMigNodes(mig)
}

// Exist checks if the node group really exists on the cloud provider side. Allows to tell the
// theoretical node group from the real one.
func (mig *GkeMig) Exist() bool {
	return mig.exist
}

// Create creates the node group on the cloud provider side.
func (mig *GkeMig) Create() (cloudprovider.NodeGroup, error) {
	if !mig.exist && mig.autoprovisioned {
		return mig.gkeManager.CreateNodePool(mig)
	}
	return nil, fmt.Errorf("Cannot create non-autoprovisioned node group")
}

// Delete deletes the node group on the cloud provider side.
// This will be executed only for autoprovisioned node groups, once their size drops to 0.
func (mig *GkeMig) Delete() error {
	if mig.exist && mig.autoprovisioned {
		return mig.gkeManager.DeleteNodePool(mig)
	}
	return fmt.Errorf("Cannot delete non-autoprovisioned node group")
}

// Autoprovisioned returns true if the node group is autoprovisioned.
func (mig *GkeMig) Autoprovisioned() bool {
	return mig.autoprovisioned
}

// TemplateNodeInfo returns a node template for this node group.
func (mig *GkeMig) TemplateNodeInfo() (*schedulercache.NodeInfo, error) {
	node, err := mig.gkeManager.GetMigTemplateNode(mig)
	if err != nil {
		return nil, err
	}
	nodeInfo := schedulercache.NewNodeInfo(cloudprovider.BuildKubeProxy(mig.Id()))
	nodeInfo.SetNode(node)
	return nodeInfo, nil
}
