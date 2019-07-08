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

package gce

import (
	"fmt"
	"io"
	"os"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/klog"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"
)

const (
	// ProviderNameGCE is the name of GCE cloud provider.
	ProviderNameGCE = "gce"

	// GPULabel is the label added to nodes with GPU resource.
	GPULabel = "cloud.google.com/gke-accelerator"
)

var (
	availableGPUTypes = map[string]struct{}{
		"nvidia-tesla-k80":  {},
		"nvidia-tesla-p100": {},
		"nvidia-tesla-v100": {},
		"nvidia-tesla-p4":   {},
		"nvidia-tesla-t4":   {},
	}
)

// GceCloudProvider implements CloudProvider interface.
type GceCloudProvider struct {
	gceManager GceManager
	// This resource limiter is used if resource limits are not defined through cloud API.
	resourceLimiterFromFlags *cloudprovider.ResourceLimiter
}

// BuildGceCloudProvider builds CloudProvider implementation for GCE.
func BuildGceCloudProvider(gceManager GceManager, resourceLimiter *cloudprovider.ResourceLimiter) (*GceCloudProvider, error) {
	return &GceCloudProvider{gceManager: gceManager, resourceLimiterFromFlags: resourceLimiter}, nil
}

// Cleanup cleans up all resources before the cloud provider is removed
func (gce *GceCloudProvider) Cleanup() error {
	gce.gceManager.Cleanup()
	return nil
}

// Name returns name of the cloud provider.
func (gce *GceCloudProvider) Name() string {
	return ProviderNameGCE
}

// GPULabel returns the label added to nodes with GPU resource.
func (gce *GceCloudProvider) GPULabel() string {
	return GPULabel
}

// GetAvailableGPUTypes return all available GPU types cloud provider supports
func (gce *GceCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	return availableGPUTypes
}

// NodeGroups returns all node groups configured for this cloud provider.
func (gce *GceCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	migs := gce.gceManager.GetMigs()
	result := make([]cloudprovider.NodeGroup, 0, len(migs))
	for _, mig := range migs {
		result = append(result, mig)
	}
	return result
}

// NodeGroupForNode returns the node group for the given node.
func (gce *GceCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	ref, err := GceRefFromProviderId(node.Spec.ProviderID)
	if err != nil {
		klog.Errorf("Error extracting node.Spec.ProviderID for node %v: %v", node.Name, err)
		return nil, err
	}
	mig, err := gce.gceManager.GetMigForInstance(ref)
	return mig, err
}

// Pricing returns pricing model for this cloud provider or error if not available.
func (gce *GceCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	return &GcePriceModel{}, nil
}

// GetAvailableMachineTypes get all machine types that can be requested from the cloud provider.
func (gce *GceCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return []string{}, nil
}

// NewNodeGroup builds a theoretical node group based on the node definition provided. The node group is not automatically
// created on the cloud provider side. The node group is not returned by NodeGroups() until it is created.
func (gce *GceCloudProvider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string,
	taints []apiv1.Taint, extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns struct containing limits (max, min) for resources (cores, memory etc.).
func (gce *GceCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	resourceLimiter, err := gce.gceManager.GetResourceLimiter()
	if err != nil {
		return nil, err
	}
	if resourceLimiter != nil {
		return resourceLimiter, nil
	}
	return gce.resourceLimiterFromFlags, nil
}

// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
// In particular the list of node groups returned by NodeGroups can change as a result of CloudProvider.Refresh().
func (gce *GceCloudProvider) Refresh() error {
	return gce.gceManager.Refresh()
}

// GceRef contains s reference to some entity in GCE world.
type GceRef struct {
	Project string
	Zone    string
	Name    string
}

func (ref GceRef) String() string {
	return fmt.Sprintf("%s/%s/%s", ref.Project, ref.Zone, ref.Name)
}

// ToProviderId converts GceRef to string in format used as ProviderId in Node object.
func (ref GceRef) ToProviderId() string {
	return fmt.Sprintf("gce://%s/%s/%s", ref.Project, ref.Zone, ref.Name)
}

// GceRefFromProviderId creates InstanceConfig object
// from provider id which must be in format:
// gce://<project-id>/<zone>/<name>
// TODO(piosz): add better check whether the id is correct
func GceRefFromProviderId(id string) (GceRef, error) {
	if len(id) == 0 {
		return GceRef{}, fmt.Errorf("wrong id: expected format gce://<project-id>/<zone>/<name>, got nil")
	}

	splitted := strings.Split(id[6:], "/")
	if len(splitted) != 3 {
		return GceRef{}, fmt.Errorf("wrong id: expected format gce://<project-id>/<zone>/<name>, got %v", id)
	}
	return GceRef{
		Project: splitted[0],
		Zone:    splitted[1],
		Name:    splitted[2],
	}, nil
}

// Mig implements NodeGroup interface.
type Mig interface {
	cloudprovider.NodeGroup

	GceRef() GceRef
}

type gceMig struct {
	gceRef GceRef

	gceManager GceManager
	minSize    int
	maxSize    int
}

// GceRef returns Mig's GceRef
func (mig *gceMig) GceRef() GceRef {
	return mig.gceRef
}

// MaxSize returns maximum size of the node group.
func (mig *gceMig) MaxSize() int {
	return mig.maxSize
}

// MinSize returns minimum size of the node group.
func (mig *gceMig) MinSize() int {
	return mig.minSize
}

// TargetSize returns the current TARGET size of the node group. It is possible that the
// number is different from the number of nodes registered in Kubernetes.
func (mig *gceMig) TargetSize() (int, error) {
	size, err := mig.gceManager.GetMigSize(mig)
	return int(size), err
}

// IncreaseSize increases Mig size
func (mig *gceMig) IncreaseSize(delta int) error {
	if delta <= 0 {
		return fmt.Errorf("size increase must be positive")
	}
	size, err := mig.gceManager.GetMigSize(mig)
	if err != nil {
		return err
	}
	if int(size)+delta > mig.MaxSize() {
		return fmt.Errorf("size increase too large - desired:%d max:%d", int(size)+delta, mig.MaxSize())
	}
	return mig.gceManager.SetMigSize(mig, size+int64(delta))
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
func (mig *gceMig) DecreaseTargetSize(delta int) error {
	if delta >= 0 {
		return fmt.Errorf("size decrease must be negative")
	}
	size, err := mig.gceManager.GetMigSize(mig)
	if err != nil {
		return err
	}
	nodes, err := mig.gceManager.GetMigNodes(mig)
	if err != nil {
		return err
	}
	if int(size)+delta < len(nodes) {
		return fmt.Errorf("attempt to delete existing nodes targetSize:%d delta:%d existingNodes: %d",
			size, delta, len(nodes))
	}
	return mig.gceManager.SetMigSize(mig, size+int64(delta))
}

// Belongs returns true if the given node belongs to the NodeGroup.
func (mig *gceMig) Belongs(node *apiv1.Node) (bool, error) {
	ref, err := GceRefFromProviderId(node.Spec.ProviderID)
	if err != nil {
		return false, err
	}
	targetMig, err := mig.gceManager.GetMigForInstance(ref)
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
func (mig *gceMig) DeleteNodes(nodes []*apiv1.Node) error {
	size, err := mig.gceManager.GetMigSize(mig)
	if err != nil {
		return err
	}
	if int(size) <= mig.MinSize() {
		return fmt.Errorf("min size reached, nodes will not be deleted")
	}
	refs := make([]GceRef, 0, len(nodes))
	for _, node := range nodes {

		belongs, err := mig.Belongs(node)
		if err != nil {
			return err
		}
		if !belongs {
			return fmt.Errorf("%s belong to a different mig than %s", node.Name, mig.Id())
		}
		gceref, err := GceRefFromProviderId(node.Spec.ProviderID)
		if err != nil {
			return err
		}
		refs = append(refs, gceref)
	}
	return mig.gceManager.DeleteInstances(refs)
}

// Id returns mig url.
func (mig *gceMig) Id() string {
	return GenerateMigUrl(mig.gceRef)
}

// Debug returns a debug string for the Mig.
func (mig *gceMig) Debug() string {
	return fmt.Sprintf("%s (%d:%d)", mig.Id(), mig.MinSize(), mig.MaxSize())
}

// Nodes returns a list of all nodes that belong to this node group.
func (mig *gceMig) Nodes() ([]cloudprovider.Instance, error) {
	return mig.gceManager.GetMigNodes(mig)
}

// Exist checks if the node group really exists on the cloud provider side.
func (mig *gceMig) Exist() bool {
	return true
}

// Create creates the node group on the cloud provider side.
func (mig *gceMig) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Delete deletes the node group on the cloud provider side.
func (mig *gceMig) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// Autoprovisioned returns true if the node group is autoprovisioned.
func (mig *gceMig) Autoprovisioned() bool {
	return false
}

// TemplateNodeInfo returns a node template for this node group.
func (mig *gceMig) TemplateNodeInfo() (*schedulernodeinfo.NodeInfo, error) {
	node, err := mig.gceManager.GetMigTemplateNode(mig)
	if err != nil {
		return nil, err
	}
	nodeInfo := schedulernodeinfo.NewNodeInfo(cloudprovider.BuildKubeProxy(mig.Id()))
	nodeInfo.SetNode(node)
	return nodeInfo, nil
}

// BuildGCE builds GCE cloud provider, manager etc.
func BuildGCE(opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
	var config io.ReadCloser
	if opts.CloudConfig != "" {
		var err error
		config, err = os.Open(opts.CloudConfig)
		if err != nil {
			klog.Fatalf("Couldn't open cloud provider configuration %s: %#v", opts.CloudConfig, err)
		}
		defer config.Close()
	}

	manager, err := CreateGceManager(config, do, opts.Regional)
	if err != nil {
		klog.Fatalf("Failed to create GCE Manager: %v", err)
	}

	provider, err := BuildGceCloudProvider(manager, rl)
	if err != nil {
		klog.Fatalf("Failed to create GCE cloud provider: %v", err)
	}
	// Register GCE API usage metrics.
	RegisterMetrics()
	return provider
}
