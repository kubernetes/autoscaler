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

package testutil

import (
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/apis/capacityquota/autoscaling.x-k8s.io/v1alpha1"
)

// QuotaOption is a functional option for configuring a CapacityQuota.
type QuotaOption func(*v1alpha1.CapacityQuota)

// NewCapacityQuota creates a new CapacityQuota with the given name and options.
func NewCapacityQuota(name string, opts ...QuotaOption) *v1alpha1.CapacityQuota {
	cq := &v1alpha1.CapacityQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	for _, opt := range opts {
		opt(cq)
	}
	return cq
}

// WithLabelSelector configures the CapacityQuota's LabelSelector to match the given labels.
func WithLabelSelector(labels map[string]string) QuotaOption {
	return func(cq *v1alpha1.CapacityQuota) {
		cq.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: labels,
		}
	}
}

// WithLimits configures the CapacityQuota's resource limits.
func WithLimits(limits v1alpha1.ResourceList) QuotaOption {
	return func(cq *v1alpha1.CapacityQuota) {
		cq.Spec.Limits = v1alpha1.CapacityQuotaLimits{
			Resources: limits,
		}
	}
}

// WithValidCondition sets Valid condition on the CapacityQuota to true.
func WithValidCondition() QuotaOption {
	return func(cq *v1alpha1.CapacityQuota) {
		c := metav1.Condition{
			Type:   v1alpha1.ValidCondition,
			Status: metav1.ConditionTrue,
			Reason: v1alpha1.ValidationSucceeded,
		}
		meta.SetStatusCondition(&cq.Status.Conditions, c)
	}
}

// WithInvalidCondition sets Valid condition on the CapacityQuota to false.
func WithInvalidCondition() QuotaOption {
	return func(cq *v1alpha1.CapacityQuota) {
		c := metav1.Condition{
			Type:   v1alpha1.ValidCondition,
			Status: metav1.ConditionFalse,
			Reason: v1alpha1.ValidationFailed,
		}
		meta.SetStatusCondition(&cq.Status.Conditions, c)
	}
}
