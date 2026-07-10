/*
Copyright The Kubernetes Authors.

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

package scheduling

import (
	"context"
	"fmt"
	"unique"

	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	resourcev1 "k8s.io/api/resource/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"

	v1 "sigs.k8s.io/karpenter/pkg/apis/v1"
	"sigs.k8s.io/karpenter/pkg/cloudprovider"
	"sigs.k8s.io/karpenter/pkg/controllers/state"
	"sigs.k8s.io/karpenter/pkg/scheduling"
	"sigs.k8s.io/karpenter/pkg/scheduling/dynamicresources"
)

// draNodeClaim adapts a scheduling *NodeClaim (in-flight or new) to the allocator's NodeClaim interface. An in-flight
// NodeClaim is a superposition of candidate instance types; ResourceSlices() therefore surfaces the template devices
// for every candidate instance type.
type draNodeClaim struct {
	nc *NodeClaim
}

func (d *draNodeClaim) ID() dynamicresources.NodeClaimID {
	// The hostname placeholder is unique per NodeClaim within a scheduling loop and stable across CanAdd/Add/Release.
	return unique.Make(d.nc.hostname)
}

func (d *draNodeClaim) NodeName() string {
	// In-flight and new NodeClaims have no concrete node, so node-name-pinned published slices never apply.
	return ""
}

func (d *draNodeClaim) NodePoolID() dynamicresources.NodePoolID {
	return unique.Make(d.nc.NodePoolName)
}

func (d *draNodeClaim) Requirements() scheduling.Requirements {
	return d.nc.Requirements
}

func (d *draNodeClaim) InstanceTypes() []dynamicresources.InstanceTypeID {
	return lo.Map(d.nc.InstanceTypeOptions, func(it *cloudprovider.InstanceType, _ int) dynamicresources.InstanceTypeID {
		return unique.Make(it.Name)
	})
}

func (d *draNodeClaim) ResourceSlices() map[dynamicresources.InstanceTypeID][]dynamicresources.ResourceSlice {
	slices := map[dynamicresources.InstanceTypeID][]dynamicresources.ResourceSlice{}
	for _, it := range d.nc.InstanceTypeOptions {
		slices[unique.Make(it.Name)] = templateSlicesForInstanceType(it)
	}
	return slices
}

// draExistingNode adapts a scheduling *ExistingNode to the allocator's NodeClaim interface. An existing node has a
// single known instance type. Once initialized, its devices are published as in-cluster ResourceSlices, so
// ResourceSlices() is empty. While uninitialized (pre-initialized), template devices are the source of truth, so
// ResourceSlices() returns the full template set for the node's instance type.
type draExistingNode struct {
	en *ExistingNode
	// instanceType is the resolved cloud provider instance type for the node, or nil if it could not be resolved
	// (e.g. an unmanaged node, or an instance type no longer offered by the NodePool).
	instanceType *cloudprovider.InstanceType
}

func (d *draExistingNode) ID() dynamicresources.NodeClaimID {
	return unique.Make(d.en.ProviderID())
}

func (d *draExistingNode) NodeName() string {
	// Published ResourceSlices pinned via spec.nodeName are accessible only from the node with this name.
	if d.en.Node != nil {
		return d.en.Node.Name
	}
	return ""
}

func (d *draExistingNode) NodePoolID() dynamicresources.NodePoolID {
	return unique.Make(d.en.Labels()[v1.NodePoolLabelKey])
}

func (d *draExistingNode) Requirements() scheduling.Requirements {
	return d.en.requirements
}

func (d *draExistingNode) InstanceTypes() []dynamicresources.InstanceTypeID {
	return []dynamicresources.InstanceTypeID{unique.Make(d.en.Labels()[corev1.LabelInstanceTypeStable])}
}

func (d *draExistingNode) ResourceSlices() map[dynamicresources.InstanceTypeID][]dynamicresources.ResourceSlice {
	// Initialized nodes have their devices published in-cluster; uninitialized nodes are represented by templates.
	if d.en.Initialized() || d.instanceType == nil {
		return map[dynamicresources.InstanceTypeID][]dynamicresources.ResourceSlice{}
	}
	return map[dynamicresources.InstanceTypeID][]dynamicresources.ResourceSlice{
		unique.Make(d.instanceType.Name): templateSlicesForInstanceType(d.instanceType),
	}
}

func templateSlicesForInstanceType(it *cloudprovider.InstanceType) []dynamicresources.ResourceSlice {
	return lo.Map(it.DynamicResources.ResourceSliceTemplates, func(t *cloudprovider.ResourceSliceTemplate, _ int) dynamicresources.ResourceSlice {
		return dynamicresources.NewTemplateSlice(t)
	})
}

// instanceTypeForNode resolves the cloud provider instance type backing a node from the scheduler's per-NodePool
// instance type set. Returns nil when the node is unmanaged or its instance type is not in the current set.
func (s *Scheduler) instanceTypeForNode(n *state.StateNode) *cloudprovider.InstanceType {
	itName := n.Labels()[corev1.LabelInstanceTypeStable]
	nodePoolName := n.Labels()[v1.NodePoolLabelKey]
	if itName == "" || nodePoolName == "" {
		return nil
	}
	return lo.FindOrElse(s.instanceTypes[nodePoolName], nil, func(it *cloudprovider.InstanceType) bool {
		return it.Name == itName
	})
}

// draDriversForNodeClaim returns the sorted set of DRA driver names whose devices were allocated to pods scheduled to
// the given NodeClaim. It reads the allocator's per-claim allocation metadata, which records the allocated devices
// (and therefore their drivers) keyed by the source NodeClaim. Returns nil when DRA is disabled or no devices were
// allocated for the NodeClaim.
func (s *Scheduler) draDriversForNodeClaim(nc *NodeClaim) []string {
	if s.allocator == nil {
		return nil
	}
	nodeClaimID := unique.Make(nc.hostname)
	drivers := sets.New[string]()
	for _, meta := range s.allocator.ResourceClaimAllocationMetadata() {
		if meta.NodeClaimID != nodeClaimID {
			continue
		}
		for _, deviceResults := range meta.Devices {
			for _, deviceResult := range deviceResults {
				drivers.Insert(deviceResult.DeviceID.Driver.Value())
			}
		}
	}
	return sets.List(drivers)
}

// resolvePodClaims resolves the ResourceClaim objects referenced by a pod into concrete *resourcev1.ResourceClaim
// objects, memoizing lookups for the duration of the scheduling loop. Claims that don't need to be generated (a
// ResourceClaimTemplate whose status entry has a nil ResourceClaimName) are skipped. Returns an error if a referenced
// claim has not yet been created, so the pod is deferred to a subsequent loop.
func (s *Scheduler) resolvePodClaims(ctx context.Context, pod *corev1.Pod) ([]*resourcev1.ResourceClaim, error) {
	claims := make([]*resourcev1.ResourceClaim, 0, len(pod.Spec.ResourceClaims))
	for i := range pod.Spec.ResourceClaims {
		pc := &pod.Spec.ResourceClaims[i]
		claimName, ok := resourceClaimName(pod, pc)
		if !ok {
			// The claim was not generated (e.g. a template whose status name is nil); nothing to allocate for it.
			continue
		}
		key := types.NamespacedName{Namespace: pod.Namespace, Name: claimName}
		claim, ok := s.cachedResourceClaims[key]
		if !ok {
			claim = &resourcev1.ResourceClaim{}
			if err := s.kubeClient.Get(ctx, key, claim); err != nil {
				if errors.IsNotFound(err) {
					return nil, fmt.Errorf("resourceclaim %q not found", key)
				}
				return nil, fmt.Errorf("getting resourceclaim %q, %w", key, err)
			}
			s.cachedResourceClaims[key] = claim
		}
		claims = append(claims, claim)
	}
	return claims, nil
}

// resourceClaimName resolves the name of the ResourceClaim backing a pod's claim reference. A direct
// ResourceClaimName is used as-is; otherwise the generated name is looked up from the pod's
// status.resourceClaimStatuses. The second return is false when no claim needs to be allocated.
func resourceClaimName(pod *corev1.Pod, pc *corev1.PodResourceClaim) (string, bool) {
	if pc.ResourceClaimName != nil {
		return *pc.ResourceClaimName, true
	}
	for i := range pod.Status.ResourceClaimStatuses {
		status := &pod.Status.ResourceClaimStatuses[i]
		if status.Name == pc.Name {
			if status.ResourceClaimName == nil {
				return "", false
			}
			return *status.ResourceClaimName, true
		}
	}
	return "", false
}
