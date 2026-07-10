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

package nodeclaim

import (
	"context"
	"errors"
	"fmt"

	"github.com/awslabs/operatorpkg/object"
	"github.com/awslabs/operatorpkg/status"
	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	resourcev1 "k8s.io/api/resource/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	v1 "sigs.k8s.io/karpenter/pkg/apis/v1"
	"sigs.k8s.io/karpenter/pkg/cloudprovider"
)

func IsManaged(nodeClaim *v1.NodeClaim, cp cloudprovider.CloudProvider) bool {
	return lo.ContainsBy(cp.GetSupportedNodeClasses(), func(nodeClass status.Object) bool {
		return object.GVK(nodeClass).GroupKind() == nodeClaim.Spec.NodeClassRef.GroupKind()
	})
}

// IsManagedPredicateFuncs is used to filter controller-runtime NodeClaim watches to NodeClaims managed by the given cloudprovider.
func IsManagedPredicateFuncs(cp cloudprovider.CloudProvider) predicate.Funcs {
	return predicate.NewPredicateFuncs(func(o client.Object) bool {
		return IsManaged(o.(*v1.NodeClaim), cp)
	})
}

func ForProviderID(providerID string) client.ListOption {
	return client.MatchingFields{"status.providerID": providerID}
}

func ForNodePool(nodePoolName string) client.ListOption {
	return client.MatchingLabels(map[string]string{v1.NodePoolLabelKey: nodePoolName})
}

func ForNodeClass(nodeClass status.Object) client.ListOption {
	return client.MatchingFields{
		"spec.nodeClassRef.group": object.GVK(nodeClass).Group,
		"spec.nodeClassRef.kind":  object.GVK(nodeClass).Kind,
		"spec.nodeClassRef.name":  nodeClass.GetName(),
	}
}

func ListManaged(ctx context.Context, c client.Client, cloudProvider cloudprovider.CloudProvider, opts ...client.ListOption) ([]*v1.NodeClaim, error) {
	nodeClaimList := &v1.NodeClaimList{}
	if err := c.List(ctx, nodeClaimList, opts...); err != nil {
		return nil, err
	}
	return lo.FilterMap(nodeClaimList.Items, func(nc v1.NodeClaim, _ int) (*v1.NodeClaim, bool) {
		return &nc, IsManaged(&nc, cloudProvider)
	}), nil
}

// PodEventHandler is a watcher on corev1.Pods that maps Pods to NodeClaim based on the node names
// and enqueues reconcile.Requests for the NodeClaims
func PodEventHandler(c client.Client, cloudProvider cloudprovider.CloudProvider) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, o client.Object) []reconcile.Request {
		nodeName := o.(*corev1.Pod).Spec.NodeName
		if nodeName == "" {
			return nil
		}
		node := &corev1.Node{}
		if err := c.Get(ctx, types.NamespacedName{Name: nodeName}, node); err != nil {
			return nil
		}
		// Because we get so many NodeClaims from this response, we are not DeepCopying the cached data here
		// DO NOT MUTATE NodeClaims in this function as this will affect the underlying cached NodeClaim
		ncs, err := ListManaged(ctx, c, cloudProvider, ForProviderID(node.Spec.ProviderID), client.UnsafeDisableDeepCopy)
		if err != nil {
			return nil
		}
		return lo.Map(ncs, func(nc *v1.NodeClaim, _ int) reconcile.Request {
			return reconcile.Request{NamespacedName: client.ObjectKeyFromObject(nc)}
		})
	})
}

// NodeEventHandler is a watcher on corev1.Node that maps Nodes to NodeClaims based on provider ids
// and enqueues reconcile.Requests for the NodeClaims
func NodeEventHandler(c client.Client, cloudProvider cloudprovider.CloudProvider) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, o client.Object) []reconcile.Request {
		// Because we get so many NodeClaims from this response, we are not DeepCopying the cached data here
		// DO NOT MUTATE NodeClaims in this function as this will affect the underlying cached NodeClaim
		ncs, err := ListManaged(ctx, c, cloudProvider, ForProviderID(o.(*corev1.Node).Spec.ProviderID), client.UnsafeDisableDeepCopy)
		if err != nil {
			return nil
		}
		return lo.Map(ncs, func(nc *v1.NodeClaim, _ int) reconcile.Request {
			return reconcile.Request{NamespacedName: client.ObjectKeyFromObject(nc)}
		})
	})
}

// ResourceSliceEventHandler is a watcher on resourcev1.ResourceSlice that maps a slice to the NodeClaim(s) backing the
// node the slice is local to (via spec.nodeName or a Node owner reference) and enqueues reconcile.Requests for them.
// It lets the lifecycle controller re-evaluate initialization when a DRA driver publishes its slices.
func ResourceSliceEventHandler(c client.Client, cloudProvider cloudprovider.CloudProvider) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, o client.Object) []reconcile.Request {
		slice := o.(*resourcev1.ResourceSlice)
		nodeName := lo.FromPtr(slice.Spec.NodeName)
		if nodeName == "" {
			for _, ref := range slice.OwnerReferences {
				if ref.Kind == "Node" {
					nodeName = ref.Name
					break
				}
			}
		}
		if nodeName == "" {
			return nil
		}
		node := &corev1.Node{}
		if err := c.Get(ctx, types.NamespacedName{Name: nodeName}, node); err != nil {
			return nil
		}
		// Because we get so many NodeClaims from this response, we are not DeepCopying the cached data here
		// DO NOT MUTATE NodeClaims in this function as this will affect the underlying cached NodeClaim
		ncs, err := ListManaged(ctx, c, cloudProvider, ForProviderID(node.Spec.ProviderID), client.UnsafeDisableDeepCopy)
		if err != nil {
			return nil
		}
		return lo.Map(ncs, func(nc *v1.NodeClaim, _ int) reconcile.Request {
			return reconcile.Request{NamespacedName: client.ObjectKeyFromObject(nc)}
		})
	})
}

// NodePoolEventHandler is a watcher on v1.NodeClaim that maps NodePool to NodeClaims based
// on the v1.NodePoolLabelKey and enqueues reconcile.Requests for the NodeClaim
func NodePoolEventHandler(c client.Client, cloudProvider cloudprovider.CloudProvider) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, o client.Object) (requests []reconcile.Request) {
		// Because we get so many NodeClaims from this response, we are not DeepCopying the cached data here
		// DO NOT MUTATE NodeClaims in this function as this will affect the underlying cached NodeClaim
		ncs, err := ListManaged(ctx, c, cloudProvider, ForNodePool(o.GetName()), client.UnsafeDisableDeepCopy)
		if err != nil {
			return nil
		}
		return lo.Map(ncs, func(nc *v1.NodeClaim, _ int) reconcile.Request {
			return reconcile.Request{NamespacedName: client.ObjectKeyFromObject(nc)}
		})
	})
}

// NodeClassEventHandler is a watcher on v1.NodeClaim that maps NodeClass to NodeClaims based
// on the nodeClassRef and enqueues reconcile.Requests for the NodeClaim
func NodeClassEventHandler(c client.Client) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, o client.Object) (requests []reconcile.Request) {
		nodeClaimList := &v1.NodeClaimList{}
		if err := c.List(ctx, nodeClaimList, client.MatchingFields{
			"spec.nodeClassRef.group": object.GVK(o).Group,
			"spec.nodeClassRef.kind":  object.GVK(o).Kind,
			"spec.nodeClassRef.name":  o.GetName(),
			// Because we get so many NodeClaims from this response, we are not DeepCopying the cached data here
			// DO NOT MUTATE NodeClaims in this function as this will affect the underlying cached NodeClaim
		}, client.UnsafeDisableDeepCopy); err != nil {
			return requests
		}
		return lo.Map(nodeClaimList.Items, func(n v1.NodeClaim, _ int) reconcile.Request {
			return reconcile.Request{
				NamespacedName: client.ObjectKeyFromObject(&n),
			}
		})
	})
}

// NodeNotFoundError is an error returned when no corev1.Nodes are found matching the passed providerID
type NodeNotFoundError struct {
	ProviderID string
}

func (e *NodeNotFoundError) Error() string {
	return fmt.Sprintf("no nodes found for provider id '%s'", e.ProviderID)
}

func IsNodeNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	nnfErr := &NodeNotFoundError{}
	return errors.As(err, &nnfErr)
}

func IgnoreNodeNotFoundError(err error) error {
	if !IsNodeNotFoundError(err) {
		return err
	}
	return nil
}

// DuplicateNodeError is an error returned when multiple corev1.Nodes are found matching the passed providerID
type DuplicateNodeError struct {
	ProviderID string
}

func (e *DuplicateNodeError) Error() string {
	return fmt.Sprintf("multiple found for provider id '%s'", e.ProviderID)
}

func IsDuplicateNodeError(err error) bool {
	if err == nil {
		return false
	}
	dnErr := &DuplicateNodeError{}
	return errors.As(err, &dnErr)
}

func IgnoreDuplicateNodeError(err error) error {
	if !IsDuplicateNodeError(err) {
		return err
	}
	return nil
}

// NodeForNodeClaim is a helper function that takes a v1.NodeClaim and attempts to find the matching corev1.Node by its providerID
// This function will return errors if:
//  1. No corev1.Nodes match the v1.NodeClaim providerID
//  2. Multiple corev1.Nodes match the v1.NodeClaim providerID
func NodeForNodeClaim(ctx context.Context, c client.Client, nodeClaim *v1.NodeClaim) (*corev1.Node, error) {
	nodes, err := AllNodesForNodeClaim(ctx, c, nodeClaim)
	if err != nil {
		return nil, err
	}
	if len(nodes) > 1 {
		return nil, &DuplicateNodeError{ProviderID: nodeClaim.Status.ProviderID}
	}
	if len(nodes) == 0 {
		return nil, &NodeNotFoundError{ProviderID: nodeClaim.Status.ProviderID}
	}
	return nodes[0], nil
}

// AllNodesForNodeClaim is a helper function that takes a v1.NodeClaim and finds ALL matching corev1.Nodes by their providerID
// If the providerID is not resolved for a NodeClaim, then no Nodes will map to it
func AllNodesForNodeClaim(ctx context.Context, c client.Client, nodeClaim *v1.NodeClaim) ([]*corev1.Node, error) {
	// NodeClaims that have no resolved providerID have no nodes mapped to them
	if nodeClaim.Status.ProviderID == "" {
		return nil, nil
	}
	nodeList := corev1.NodeList{}
	if err := c.List(ctx, &nodeList, client.MatchingFields{"spec.providerID": nodeClaim.Status.ProviderID}); err != nil {
		return nil, fmt.Errorf("listing nodes, %w", err)
	}
	return lo.ToSlicePtr(nodeList.Items), nil
}

func UpdateNodeOwnerReferences(nodeClaim *v1.NodeClaim, node *corev1.Node) *corev1.Node {
	gvk := object.GVK(nodeClaim)
	if lo.ContainsBy(node.OwnerReferences, func(o metav1.OwnerReference) bool {
		return o.APIVersion == gvk.GroupVersion().String() && o.Kind == gvk.Kind && o.UID == nodeClaim.UID
	}) {
		return node
	}
	node.OwnerReferences = append(node.OwnerReferences, metav1.OwnerReference{
		APIVersion:         gvk.GroupVersion().String(),
		Kind:               gvk.Kind,
		Name:               nodeClaim.Name,
		UID:                nodeClaim.UID,
		BlockOwnerDeletion: new(true),
	})
	return node
}
