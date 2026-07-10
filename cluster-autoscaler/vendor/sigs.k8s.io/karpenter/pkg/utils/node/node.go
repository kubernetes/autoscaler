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

package node

import (
	"context"
	"errors"
	"fmt"

	"github.com/awslabs/operatorpkg/object"
	"github.com/awslabs/operatorpkg/serrors"
	"github.com/awslabs/operatorpkg/status"
	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/klog/v2"

	"k8s.io/utils/clock"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"sigs.k8s.io/karpenter/pkg/events"

	v1 "sigs.k8s.io/karpenter/pkg/apis/v1"
	"sigs.k8s.io/karpenter/pkg/cloudprovider"
	nodeclaimutils "sigs.k8s.io/karpenter/pkg/utils/nodeclaim"
	"sigs.k8s.io/karpenter/pkg/utils/pdb"
	"sigs.k8s.io/karpenter/pkg/utils/pod"
)

// NodeClaimNotFoundError is an error returned when no v1.NodeClaims are found matching the passed providerID
type NodeClaimNotFoundError struct {
	error
}

func NewNodeClaimNotFoundError(providerID string) NodeClaimNotFoundError {
	return NodeClaimNotFoundError{
		error: serrors.Wrap(
			fmt.Errorf("no nodeclaims found for provider-id"),
			"provider-id", providerID,
		),
	}
}

func IsNodeClaimNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	nnfErr := NodeClaimNotFoundError{}
	return errors.As(err, &nnfErr)
}

func IgnoreNodeClaimNotFoundError(err error) error {
	if !IsNodeClaimNotFoundError(err) {
		return err
	}
	return nil
}

// DuplicateNodeClaimError is an error returned when multiple v1.NodeClaims are found matching the passed providerID
type DuplicateNodeClaimError struct {
	error
}

func NewDuplicateNodeClaimError(providerID string, nodeClaims ...*v1.NodeClaim) DuplicateNodeClaimError {
	return DuplicateNodeClaimError{
		error: serrors.Wrap(
			fmt.Errorf("found duplicate nodeclaims for provider-id"),
			"provider-id", providerID,
			"NodeClaims", lo.Map(nodeClaims, func(nc *v1.NodeClaim, _ int) klog.ObjectRef { return klog.KObj(nc) }),
		),
	}
}

func IsDuplicateNodeClaimError(err error) bool {
	if err == nil {
		return false
	}
	dnErr := DuplicateNodeClaimError{}
	return errors.As(err, &dnErr)
}

func IgnoreDuplicateNodeClaimError(err error) error {
	if !IsDuplicateNodeClaimError(err) {
		return err
	}
	return nil
}

// GetPods grabs all pods that are currently bound to the passed nodes
func GetPods(ctx context.Context, kubeClient client.Client, nodes ...*corev1.Node) ([]*corev1.Pod, error) {
	var pods []*corev1.Pod
	for _, node := range nodes {
		var podList corev1.PodList
		if err := kubeClient.List(ctx, &podList, client.MatchingFields{"spec.nodeName": node.Name}); err != nil {
			return nil, fmt.Errorf("listing pods, %w", err)
		}
		for i := range podList.Items {
			pods = append(pods, &podList.Items[i])
		}
	}
	return pods, nil
}

// GetNodeClaims grabs all NodeClaims with a providerID that matches the provided Node
func GetNodeClaims(ctx context.Context, kubeClient client.Client, node *corev1.Node) ([]*v1.NodeClaim, error) {
	// Nodes without providerID should not match any NodeClaims to prevent false positives
	// with NodeClaims that also have empty providerIDs (e.g., during NodeClaim creation)
	if node.Spec.ProviderID == "" {
		return nil, nil
	}
	ncs := &v1.NodeClaimList{}
	if err := kubeClient.List(ctx, ncs, nodeclaimutils.ForProviderID(node.Spec.ProviderID)); err != nil {
		return nil, fmt.Errorf("listing nodeclaims, %w", err)
	}
	return lo.ToSlicePtr(ncs.Items), nil
}

// NodeClaimForNode is a helper function that takes a corev1.Node and attempts to find the matching v1.NodeClaim by its providerID
// This function will return errors if:
//  1. No v1.NodeClaims match the corev1.Node's providerID
//  2. Multiple v1.NodeClaims match the corev1.Node's providerID
func NodeClaimForNode(ctx context.Context, c client.Client, node *corev1.Node) (*v1.NodeClaim, error) {
	nodeClaims, err := GetNodeClaims(ctx, c, node)
	if err != nil {
		return nil, err
	}
	if len(nodeClaims) > 1 {
		return nil, NewDuplicateNodeClaimError(node.Spec.ProviderID, nodeClaims...)
	}
	if len(nodeClaims) == 0 {
		return nil, NewNodeClaimNotFoundError(node.Spec.ProviderID)
	}
	return nodeClaims[0], nil
}

// GetCurrentlyReschedulablePods grabs all pods from the passed nodes that satisfy the IsReschedulable criteria
func GetCurrentlyReschedulablePods(ctx context.Context, kubeClient client.Client, clk clock.Clock, recorder events.Recorder, nodes ...*corev1.Node) ([]*corev1.Pod, error) {
	pods, err := GetPods(ctx, kubeClient, nodes...)
	if err != nil {
		return nil, fmt.Errorf("listing pods, %w", err)
	}

	pdbs, err := pdb.NewLimits(ctx, kubeClient)
	if err != nil {
		return nil, fmt.Errorf("tracking PodDisruptionBudgets, %w", err)
	}

	return lo.Filter(pods, func(p *corev1.Pod, _ int) bool {
		return pdbs.IsCurrentlyReschedulable(p, clk, recorder)
	}), nil
}

// GetProvisionablePods grabs all the pods from the passed nodes that satisfy the IsProvisionable criteria
func GetProvisionablePods(ctx context.Context, kubeClient client.Client) ([]*corev1.Pod, error) {
	var podList corev1.PodList
	if err := kubeClient.List(ctx, &podList, client.MatchingFields{"spec.nodeName": ""}); err != nil {
		return nil, fmt.Errorf("listing pods, %w", err)
	}
	return lo.FilterMap(podList.Items, func(p corev1.Pod, _ int) (*corev1.Pod, bool) {
		return &p, pod.IsProvisionable(&p)
	}), nil
}

// GetVolumeAttachments grabs all volumeAttachments associated with the passed node
func GetVolumeAttachments(ctx context.Context, kubeClient client.Client, node *corev1.Node) ([]*storagev1.VolumeAttachment, error) {
	var volumeAttachmentList storagev1.VolumeAttachmentList
	if err := kubeClient.List(ctx, &volumeAttachmentList, client.MatchingFields{"spec.nodeName": node.Name}); err != nil {
		return nil, fmt.Errorf("listing volumeAttachments, %w", err)
	}
	return lo.ToSlicePtr(volumeAttachmentList.Items), nil
}

func GetCondition(n *corev1.Node, match corev1.NodeConditionType) corev1.NodeCondition {
	for _, condition := range n.Status.Conditions {
		if condition.Type == match {
			return condition
		}
	}
	return corev1.NodeCondition{}
}

func IsManaged(node *corev1.Node, cp cloudprovider.CloudProvider) bool {
	return lo.ContainsBy(cp.GetSupportedNodeClasses(), func(nodeClass status.Object) bool {
		_, ok := node.Labels[v1.NodeClassLabelKey(object.GVK(nodeClass).GroupKind())]
		return ok
	})
}

// IsManagedPredicateFuncs is used to filter controller-runtime NodeClaim watches to NodeClaims managed by the given cloudprovider.
func IsManagedPredicateFuncs(cp cloudprovider.CloudProvider) predicate.Funcs {
	return predicate.NewPredicateFuncs(func(o client.Object) bool {
		return IsManaged(o.(*corev1.Node), cp)
	})
}

func NodeClaimEventHandler(c client.Client) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, o client.Object) []reconcile.Request {
		providerID := o.(*v1.NodeClaim).Status.ProviderID
		if providerID == "" {
			return nil
		}
		nodes := &corev1.NodeList{}
		if err := c.List(ctx, nodes, client.MatchingFields{"spec.providerID": providerID}); err != nil {
			return nil
		}
		return lo.Map(nodes.Items, func(n corev1.Node, _ int) reconcile.Request {
			return reconcile.Request{NamespacedName: client.ObjectKeyFromObject(&n)}
		})
	})
}
