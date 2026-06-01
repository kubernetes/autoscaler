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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	cqv1alpha1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacityquota/autoscaling.x-k8s.io/v1alpha1"
)

// Validator checks whether CapacityQuota is valid.
type Validator interface {
	Validate(ctx context.Context, cq *cqv1alpha1.CapacityQuota) error
}

type labelSelectorValidator struct{}

func (v *labelSelectorValidator) Validate(_ context.Context, cq *cqv1alpha1.CapacityQuota) error {
	if _, err := metav1.LabelSelectorAsSelector(cq.Spec.Selector); err != nil {
		return field.Invalid(field.NewPath("spec").Child("selector"), cq.Spec.Selector, err.Error())
	}
	return nil
}
