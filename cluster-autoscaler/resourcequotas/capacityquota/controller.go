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

package capacityquota

import (
	"context"
	"fmt"
	"maps"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/autoscaler/cluster-autoscaler/resourcequotas"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"k8s.io/autoscaler/cluster-autoscaler/apis/capacityquota/autoscaling.x-k8s.io/v1alpha1"
)

// Reconciler reconciles a CapacityQuota object
type Reconciler struct {
	client     client.Client
	nodeFilter resourcequotas.NodeFilter
	validators []Validator
}

// ReconcilerOptions are configuration options for CapacityQuota reconciler.
type ReconcilerOptions struct {
	NodeFilter       resourcequotas.NodeFilter
	CustomValidators []Validator
}

// NewCapacityQuotaReconciler returns a new CapacityQuotaReconciler
func NewCapacityQuotaReconciler(client client.Client, opts ReconcilerOptions) *Reconciler {
	validators := []Validator{&labelSelectorValidator{}}
	validators = append(validators, opts.CustomValidators...)
	return &Reconciler{
		client:     client,
		nodeFilter: opts.NodeFilter,
		validators: validators,
	}
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var cq v1alpha1.CapacityQuota
	if err := r.client.Get(ctx, req.NamespacedName, &cq); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	originalCQ := cq.DeepCopy()

	validationErrs := r.validateCapacityQuota(ctx, &cq)
	if len(validationErrs) > 0 {
		errMsg := formatValidationErrors(validationErrs)
		setCondition(&cq, v1alpha1.ValidCondition, metav1.ConditionFalse, v1alpha1.ValidationFailed, errMsg)
		setCondition(&cq, v1alpha1.ReconciledCondition, metav1.ConditionFalse, v1alpha1.ReconciliationFailed, "Validation failed")

		if patchErr := r.patchStatus(ctx, originalCQ, &cq); patchErr != nil {
			return ctrl.Result{}, patchErr
		}
		return ctrl.Result{}, nil
	}
	setCondition(&cq, v1alpha1.ValidCondition, metav1.ConditionTrue, v1alpha1.ValidationSucceeded, "CapacityQuota is valid")

	result, err := r.reconcileCapacityQuota(ctx, &cq)

	if err != nil {
		setCondition(&cq, v1alpha1.ReconciledCondition, metav1.ConditionFalse, v1alpha1.ReconciliationFailed, err.Error())
	} else {
		setCondition(&cq, v1alpha1.ReconciledCondition, metav1.ConditionTrue, v1alpha1.ReconciliationSucceeded, "CapacityQuota successfully reconciled")
	}

	if patchErr := r.patchStatus(ctx, originalCQ, &cq); patchErr != nil {
		return ctrl.Result{}, patchErr
	}

	return result, err
}

func formatValidationErrors(errs []error) string {
	errStrings := make([]string, len(errs))
	for i, err := range errs {
		errStrings[i] = err.Error()
	}
	return strings.Join(errStrings, "; ")
}

func (r *Reconciler) patchStatus(ctx context.Context, originalCQ, currentCQ *v1alpha1.CapacityQuota) error {
	if !apiequality.Semantic.DeepEqual(originalCQ.Status, currentCQ.Status) {
		if patchErr := r.client.Status().Patch(ctx, currentCQ, client.MergeFrom(originalCQ)); patchErr != nil {
			return fmt.Errorf("failed to patch status: %w", patchErr)
		}
	}
	return nil
}

func (r *Reconciler) validateCapacityQuota(ctx context.Context, cq *v1alpha1.CapacityQuota) []error {
	var allErrs []error
	for _, v := range r.validators {
		if err := v.Validate(ctx, cq); err != nil {
			allErrs = append(allErrs, err)
		}
	}

	if len(allErrs) == 0 {
		return nil
	}

	return allErrs
}

func (r *Reconciler) reconcileCapacityQuota(ctx context.Context, cq *v1alpha1.CapacityQuota) (ctrl.Result, error) {
	matchingNodes, err := r.getMatchingNodes(ctx, cq)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get matching nodes: %w", err)
	}

	cq.Status.Used = &v1alpha1.CapacityQuotaUsage{
		Resources: calculateResourceUsage(cq, matchingNodes),
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) getMatchingNodes(ctx context.Context, cq *v1alpha1.CapacityQuota) ([]*corev1.Node, error) {
	var nodeList corev1.NodeList
	var listOpts []client.ListOption

	if cq.Spec.Selector != nil {
		selector, err := metav1.LabelSelectorAsSelector(cq.Spec.Selector)
		if err != nil {
			return nil, err
		}
		listOpts = append(listOpts, client.MatchingLabelsSelector{Selector: selector})
	}

	if err := r.client.List(ctx, &nodeList, listOpts...); err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	matchingNodes := make([]*corev1.Node, 0, len(nodeList.Items))
	for i := range nodeList.Items {
		node := &nodeList.Items[i]
		if r.nodeFilter == nil || !r.nodeFilter.ExcludeFromTracking(node) {
			matchingNodes = append(matchingNodes, node)
		}
	}

	return matchingNodes, nil
}

func calculateResourceUsage(cq *v1alpha1.CapacityQuota, matchingNodes []*corev1.Node) v1alpha1.ResourceList {
	used := make(v1alpha1.ResourceList)

	for resName := range cq.Spec.Limits.Resources {
		usage := resource.NewQuantity(0, resource.DecimalSI)
		for _, node := range matchingNodes {
			if resName == v1alpha1.ResourceNodes {
				usage.Add(*resource.NewQuantity(1, resource.DecimalSI))
			} else if quantity, ok := node.Status.Capacity[corev1.ResourceName(resName)]; ok {
				usage.Add(quantity)
			}
		}
		used[resName] = usage.DeepCopy()
	}
	return used
}

func setCondition(cq *v1alpha1.CapacityQuota, condType string, status metav1.ConditionStatus, reason, message string) {
	newCondition := metav1.Condition{
		Type:               condType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: metav1.Now(),
		ObservedGeneration: cq.Generation,
	}

	meta.SetStatusCondition(&cq.Status.Conditions, newCondition)
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	nodePredicate := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldNode, oldOk := e.ObjectOld.(*corev1.Node)
			newNode, newOk := e.ObjectNew.(*corev1.Node)
			if !oldOk || !newOk {
				return false
			}

			if !maps.Equal(oldNode.Labels, newNode.Labels) {
				return true
			}

			if !maps.EqualFunc(oldNode.Status.Capacity, newNode.Status.Capacity, func(q1, q2 resource.Quantity) bool {
				return q1.Equal(q2)
			}) {
				return true
			}

			return false
		},
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.CapacityQuota{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Watches(
			&corev1.Node{},
			handler.EnqueueRequestsFromMapFunc(r.findCapacityQuotasForNode),
			builder.WithPredicates(nodePredicate),
		).
		Complete(r)
}

func (r *Reconciler) findCapacityQuotasForNode(ctx context.Context, o client.Object) []reconcile.Request {
	node, ok := o.(*corev1.Node)
	if !ok {
		return nil
	}

	var cqList v1alpha1.CapacityQuotaList
	if err := r.client.List(ctx, &cqList); err != nil {
		runtime.HandleError(fmt.Errorf("failed to list capacity quotas in map func: %w", err))
		return nil
	}

	var requests []reconcile.Request
	nodeLabels := labels.Set(node.GetLabels())

	for _, cq := range cqList.Items {
		if cq.Spec.Selector != nil {
			selector, err := metav1.LabelSelectorAsSelector(cq.Spec.Selector)
			if err != nil {
				continue
			}
			// note: this works correctly for node updates both when a label is added and removed from a node.
			// Intuitively, you may think that removing a label from the node so that it no longer matches a CapacityQuota
			// will not trigger the reconciliation for that quota. This is not the case, since `EnqueueRequestsFromMapFunc`
			// runs the provided function both on the old and the new object, and deduplicates the returned requests.
			// Ref: https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.24.1/pkg/handler#EnqueueRequestsFromMapFunc
			if !selector.Matches(nodeLabels) {
				continue
			}
		}
		requests = append(requests, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name: cq.Name,
			},
		})
	}

	return requests
}
