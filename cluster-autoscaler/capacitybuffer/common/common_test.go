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

package common

import (
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1beta1"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/testutil"
	"k8s.io/utils/ptr"
)

func TestConditionUpdates(t *testing.T) {
	existingCondition := metav1.Condition{
		Type:               "ExistingCondition",
		Status:             metav1.ConditionTrue,
		Reason:             "Existing",
		Message:            "Existing message",
		LastTransitionTime: metav1.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	tests := []struct {
		name           string
		initialBuffer  *v1.CapacityBuffer
		updateFunc     func(*v1.CapacityBuffer)
		wantConditions []metav1.Condition
	}{
		{
			name: "SetBufferAsReadyForProvisioning preserves existing conditions",
			initialBuffer: testutil.NewBuffer(func(b *v1.CapacityBuffer) {
				b.Status.Conditions = []metav1.Condition{existingCondition}
			}),
			updateFunc: func(b *v1.CapacityBuffer) {
				SetBufferAsReadyForProvisioning(b, &v1.LocalObjectRef{Name: "pt"}, ptr.To(int64(1)), ptr.To(int32(1)), ptr.To(capacitybuffer.ActiveProvisioningStrategy))
			},
			wantConditions: []metav1.Condition{
				existingCondition,
				{
					Type:    capacitybuffer.ReadyForProvisioningCondition,
					Status:  metav1.ConditionTrue,
					Reason:  capacitybuffer.AttributesSetSuccessfullyReason,
					Message: "ready",
				},
			},
		},
		{
			name: "SetBufferAsNotReadyForProvisioning preserves existing conditions",
			initialBuffer: testutil.NewBuffer(func(b *v1.CapacityBuffer) {
				b.Status.Conditions = []metav1.Condition{existingCondition}
			}),
			updateFunc: func(b *v1.CapacityBuffer) {
				SetBufferAsNotReadyForProvisioning(b, &v1.LocalObjectRef{Name: "pt"}, ptr.To(int64(1)), ptr.To(int32(1)), ptr.To(capacitybuffer.ActiveProvisioningStrategy), errors.New("some error"))
			},
			wantConditions: []metav1.Condition{
				existingCondition,
				{
					Type:    capacitybuffer.ReadyForProvisioningCondition,
					Status:  metav1.ConditionFalse,
					Reason:  "error",
					Message: "Buffer not ready for provisioning: some error",
				},
			},
		},
		{
			name: "UpdateBufferStatusToFailedProvisioning preserves existing conditions",
			initialBuffer: testutil.NewBuffer(func(b *v1.CapacityBuffer) {
				b.Status.Conditions = []metav1.Condition{existingCondition}
			}),
			updateFunc: func(b *v1.CapacityBuffer) {
				UpdateBufferStatusToFailedProvisioning(b, "FailedReason", "failed message")
			},
			wantConditions: []metav1.Condition{
				existingCondition,
				{
					Type:    capacitybuffer.ProvisioningCondition,
					Status:  metav1.ConditionFalse,
					Reason:  "FailedReason",
					Message: "failed message",
				},
			},
		},
		{
			name: "UpdateBufferStatusToSuccessfullyProvisioning preserves existing conditions",
			initialBuffer: testutil.NewBuffer(func(b *v1.CapacityBuffer) {
				b.Status.Conditions = []metav1.Condition{existingCondition}
			}),
			updateFunc: func(b *v1.CapacityBuffer) {
				UpdateBufferStatusToSuccessfullyProvisioning(b, "SuccessReason")
			},
			wantConditions: []metav1.Condition{
				existingCondition,
				{
					Type:    capacitybuffer.ProvisioningCondition,
					Status:  metav1.ConditionTrue,
					Reason:  "SuccessReason",
					Message: "",
				},
			},
		},
		{
			name: "MarkBufferAsLimitedByQuota preserves existing conditions",
			initialBuffer: testutil.NewBuffer(func(b *v1.CapacityBuffer) {
				b.Status.Conditions = []metav1.Condition{existingCondition}
			}),
			updateFunc: func(b *v1.CapacityBuffer) {
				MarkBufferAsLimitedByQuota(b, 10, 5, []string{"quota-1", "quota-2"})
			},
			wantConditions: []metav1.Condition{
				existingCondition,
				{
					Type:    capacitybuffer.LimitedByQuotasCondition,
					Status:  metav1.ConditionTrue,
					Reason:  capacitybuffer.LimitedByQuotasReason,
					Message: "Buffer replicas limited from 10 to 5 due to quotas: quota-1, quota-2",
				},
			},
		},
		{
			name: "UpdateBufferStatusLimitedByQuotas preserves existing conditions",
			initialBuffer: testutil.NewBuffer(func(b *v1.CapacityBuffer) {
				b.Status.Conditions = []metav1.Condition{existingCondition}
			}),
			updateFunc: func(b *v1.CapacityBuffer) {
				UpdateBufferStatusLimitedByQuotas(b, false, "")
			},
			wantConditions: []metav1.Condition{
				existingCondition,
				{
					Type:   capacitybuffer.LimitedByQuotasCondition,
					Status: metav1.ConditionFalse,
					Reason: capacitybuffer.LimitedByQuotasReason,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.updateFunc(tc.initialBuffer)

			sortOpt := cmpopts.SortSlices(func(a, b metav1.Condition) bool {
				return a.Type < b.Type
			})
			ignoreTime := cmpopts.IgnoreFields(metav1.Condition{}, "LastTransitionTime")

			if diff := cmp.Diff(tc.wantConditions, tc.initialBuffer.Status.Conditions, sortOpt, ignoreTime); diff != "" {
				t.Errorf("Conditions mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
