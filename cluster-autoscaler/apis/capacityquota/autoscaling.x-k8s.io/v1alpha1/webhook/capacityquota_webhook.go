/*
Copyright 2025 The Kubernetes Authors.

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

package v1alpha1

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	cqv1alpha1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacityquota/autoscaling.x-k8s.io/v1alpha1"
)

type CapacityQuotaValidator interface {
	Validate(ctx context.Context, cq *cqv1alpha1.CapacityQuota) *field.Error
}

// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-autoscaling-x-k8s-io-v1alpha1-capacityquota,mutating=false,failurePolicy=fail,sideEffects=None,groups=autoscaling.x-k8s.io,resources=capacityquotas,verbs=create;update,versions=v1alpha1,name=vcapacityquota-v1alpha1.kb.io,admissionReviewVersions=v1

// CapacityQuotaValidatingWebhook struct is responsible for validating the CapacityQuota resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type CapacityQuotaValidatingWebhook struct {
	validators []CapacityQuotaValidator
}

func NewCapacityQuotaValidatingWebhook(customValidators []CapacityQuotaValidator) *CapacityQuotaValidatingWebhook {
	validators := []CapacityQuotaValidator{&labelSelectorValidator{}}
	validators = append(validators, customValidators...)
	return &CapacityQuotaValidatingWebhook{
		validators: validators,
	}
}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type CapacityQuota.
func (w *CapacityQuotaValidatingWebhook) ValidateCreate(ctx context.Context, cq *cqv1alpha1.CapacityQuota) (admission.Warnings, error) {
	return nil, w.validateCapacityQuota(ctx, cq)
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type CapacityQuota.
func (w *CapacityQuotaValidatingWebhook) ValidateUpdate(ctx context.Context, _, newCQ *cqv1alpha1.CapacityQuota) (admission.Warnings, error) {
	return nil, w.validateCapacityQuota(ctx, newCQ)
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type CapacityQuota.
func (w *CapacityQuotaValidatingWebhook) ValidateDelete(context.Context, *cqv1alpha1.CapacityQuota) (admission.Warnings, error) {
	return nil, nil
}

func (w *CapacityQuotaValidatingWebhook) validateCapacityQuota(ctx context.Context, cq *cqv1alpha1.CapacityQuota) error {
	var allErrs field.ErrorList
	for _, v := range w.validators {
		if err := v.Validate(ctx, cq); err != nil {
			allErrs = append(allErrs, err)
		}
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{Group: cqv1alpha1.APIGroup, Kind: "CapacityQuota"}, cq.Name, allErrs)
}

type labelSelectorValidator struct{}

func (v *labelSelectorValidator) Validate(_ context.Context, cq *cqv1alpha1.CapacityQuota) *field.Error {
	if _, err := metav1.LabelSelectorAsSelector(cq.Spec.Selector); err != nil {
		return field.Invalid(field.NewPath("spec").Child("selector"), cq.Spec.Selector, err.Error())
	}
	return nil
}

func (w *CapacityQuotaValidatingWebhook) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &cqv1alpha1.CapacityQuota{}).
		WithValidator(w).
		Complete()
}
