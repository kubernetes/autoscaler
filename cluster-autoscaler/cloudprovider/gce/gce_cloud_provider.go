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
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/opentracing/opentracing-go"
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
func (gce *GceCloudProvider) Cleanup(ctx context.Context) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "GceCloudProvider.Cleanup")
	defer span.Finish()

	gce.gceManager.Cleanup(ctx)
	return nil
}

// Name returns name of the cloud provider.
func (gce *GceCloudProvider) Name(ctx context.Context) string {
	span, ctx := opentracing.StartSpanFromContext(ctx, "GceCloudProvider.Name")
	defer span.Finish()

	return ProviderNameGCE
}

// NodeGroups returns all node groups configured for this cloud provider.
func (gce *GceCloudProvider) NodeGroups(ctx context.Context) []cloudprovider.NodeGroup {
	span, ctx := opentracing.StartSpanFromContext(ctx, "GceCloudProvider.NodeGroups")
	defer span.Finish()

	migs := gce.gceManager.GetMigs(ctx)
	result := make([]cloudprovider.NodeGroup, 0, len(migs))
	for _, mig := range migs {
		result = append(result, mig.Config)
	}
	return result
}

// NodeGroupForNode returns the node group for the given node.
func (gce *GceCloudProvider) NodeGroupForNode(ctx context.Context, node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "GceCloudProvider.NodeGroupForNode")
	defer span.Finish()

	ref, err := GceRefFromProviderId(node.Spec.ProviderID)
	if err != nil {
		return nil, err
	}
	mig, err := gce.gceManager.GetMigForInstance(ctx, ref)
	return mig, err
}

// Pricing returns pricing model for this cloud provider or error if not available.
func (gce *GceCloudProvider) Pricing(ctx context.Context) (cloudprovider.PricingModel, errors.AutoscalerError) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "GceCloudProvider.Pricing")
	defer span.Finish()

	return &GcePriceModel{}, nil
}

// GetAvailableMachineTypes get all machine types that can be requested from the cloud provider.
func (gce *GceCloudProvider) GetAvailableMachineTypes(ctx context.Context) ([]string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "GceCloudProvider.GetAvailableMachineTypes")
	defer span.Finish()

	return []string{}, nil
}

// NewNodeGroup builds a theoretical node group based on the node definition provided. The node group is not automatically
// created on the cloud provider side. The node group is not returned by NodeGroups() until it is created.
func (gce *GceCloudProvider) NewNodeGroup(ctx context.Context, machineType string, labels map[string]string, systemLabels map[string]string, taints []apiv1.Taint, extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "GceCloudProvider.NewNodeGroup")
	defer span.Finish()

	return nil, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns struct containing limits (max, min) for resources (cores, memory etc.).
func (gce *GceCloudProvider) GetResourceLimiter(ctx context.Context) (*cloudprovider.ResourceLimiter, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "GceCloudProvider.GetResourceLimiter")
	defer span.Finish()

	resourceLimiter, err := gce.gceManager.GetResourceLimiter(ctx)
	if err != nil {
		return nil, err
	}
	if resourceLimiter != nil {
		return resourceLimiter, nil
	}
	return gce.resourceLimiterFromFlags, nil
}

// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
// In particular the list of node groups returned by NodeGroups can change as a result of CloudProvider.Refresh(ctx).
func (gce *GceCloudProvider) Refresh(ctx context.Context) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "GceCloudProvider.Refresh")
	defer span.Finish()

	return gce.gceManager.Refresh(ctx)
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
func GceRefFromProviderId(id string) (*GceRef, error) {
	splitted := strings.Split(id[6:], "/")
	if len(splitted) != 3 {
		return nil, fmt.Errorf("wrong id: expected format gce://<project-id>/<zone>/<name>, got %v", id)
	}
	return &GceRef{
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
func (mig *gceMig) TargetSize(ctx context.Context) (int, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "gceMig.TargetSize")
	defer span.Finish()

	size, err := mig.gceManager.GetMigSize(ctx, mig)
	return int(size), err
}

// IncreaseSize increases Mig size
func (mig *gceMig) IncreaseSize(ctx context.Context, delta int) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "gceMig.IncreaseSize")
	defer span.Finish()

	if delta <= 0 {
		return fmt.Errorf("size increase must be positive")
	}
	size, err := mig.gceManager.GetMigSize(ctx, mig)
	if err != nil {
		return err
	}
	if int(size)+delta > mig.MaxSize() {
		return fmt.Errorf("size increase too large - desired:%d max:%d", int(size)+delta, mig.MaxSize())
	}
	return mig.gceManager.SetMigSize(ctx, mig, size+int64(delta))
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
func (mig *gceMig) DecreaseTargetSize(ctx context.Context, delta int) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "gceMig.DecreaseTargetSize")
	defer span.Finish()

	if delta >= 0 {
		return fmt.Errorf("size decrease must be negative")
	}
	size, err := mig.gceManager.GetMigSize(ctx, mig)
	if err != nil {
		return err
	}
	nodes, err := mig.gceManager.GetMigNodes(ctx, mig)
	if err != nil {
		return err
	}
	if int(size)+delta < len(nodes) {
		return fmt.Errorf("attempt to delete existing nodes targetSize:%d delta:%d existingNodes: %d",
			size, delta, len(nodes))
	}
	return mig.gceManager.SetMigSize(ctx, mig, size+int64(delta))
}

// Belongs returns true if the given node belongs to the NodeGroup.
func (mig *gceMig) Belongs(ctx context.Context, node *apiv1.Node) (bool, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "gceMig.Belongs")
	defer span.Finish()

	ref, err := GceRefFromProviderId(node.Spec.ProviderID)
	if err != nil {
		return false, err
	}
	targetMig, err := mig.gceManager.GetMigForInstance(ctx, ref)
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
func (mig *gceMig) DeleteNodes(ctx context.Context, nodes []*apiv1.Node) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "gceMig.DeleteNodes")
	defer span.Finish()

	size, err := mig.gceManager.GetMigSize(ctx, mig)
	if err != nil {
		return err
	}
	if int(size) <= mig.MinSize() {
		return fmt.Errorf("min size reached, nodes will not be deleted")
	}
	refs := make([]*GceRef, 0, len(nodes))
	for _, node := range nodes {

		belongs, err := mig.Belongs(ctx, node)
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
	return mig.gceManager.DeleteInstances(ctx, refs)
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
func (mig *gceMig) Nodes(ctx context.Context) ([]cloudprovider.Instance, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "gceMig.Nodes")
	defer span.Finish()

	return mig.gceManager.GetMigNodes(ctx, mig)
}

// Exist checks if the node group really exists on the cloud provider side.
func (mig *gceMig) Exist() bool {
	return true
}

// Create creates the node group on the cloud provider side.
func (mig *gceMig) Create(ctx context.Context) (cloudprovider.NodeGroup, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "gceMig.Create")
	defer span.Finish()

	return nil, cloudprovider.ErrNotImplemented
}

// Delete deletes the node group on the cloud provider side.
func (mig *gceMig) Delete(ctx context.Context) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "gceMig.Delete")
	defer span.Finish()

	return cloudprovider.ErrNotImplemented
}

// Autoprovisioned returns true if the node group is autoprovisioned.
func (mig *gceMig) Autoprovisioned() bool {
	return false
}

// TemplateNodeInfo returns a node template for this node group.
func (mig *gceMig) TemplateNodeInfo(ctx context.Context) (*schedulernodeinfo.NodeInfo, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "gceMig.TemplateNodeInfo")
	defer span.Finish()

	node, err := mig.gceManager.GetMigTemplateNode(ctx, mig)
	if err != nil {
		return nil, err
	}
	nodeInfo := schedulernodeinfo.NewNodeInfo(cloudprovider.BuildKubeProxy(mig.Id()))
	nodeInfo.SetNode(node)
	return nodeInfo, nil
}

// BuildGCE builds GCE cloud provider, manager etc.
func BuildGCE(ctx context.Context, opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
	var config io.ReadCloser
	if opts.CloudConfig != "" {
		var err error
		config, err = os.Open(opts.CloudConfig)
		if err != nil {
			klog.Fatalf("Couldn't open cloud provider configuration %s: %#v", opts.CloudConfig, err)
		}
		defer config.Close()
	}

	manager, err := CreateGceManager(ctx, config, do, opts.Regional)
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
