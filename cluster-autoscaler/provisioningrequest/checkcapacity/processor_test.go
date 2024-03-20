/*
Copyright 2024 The Kubernetes Authors.

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

package checkcapacity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/autoscaler/cluster-autoscaler/apis/provisioningrequest/autoscaling.x-k8s.io/v1beta1"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/conditions"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqclient"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqwrapper"
)

func TestProcess(t *testing.T) {
	now := time.Now()
	dayAgo := now.Add(-1 * 24 * time.Hour)
	weekAgo := now.Add(-1 * defaultExpirationTime).Add(-1 * 5 * time.Minute)

	testCases := []struct {
		name           string
		creationTime   time.Time
		conditions     []metav1.Condition
		wantConditions []metav1.Condition
	}{
		{
			name:         "New ProvisioningRequest, empty conditions",
			creationTime: now,
		},
		{
			name:         "ProvisioningRequest with empty conditions, expired",
			creationTime: weekAgo,
			wantConditions: []metav1.Condition{
				{
					Type:               v1beta1.Failed,
					Status:             metav1.ConditionTrue,
					LastTransitionTime: metav1.NewTime(now),
					Reason:             conditions.ExpiredReason,
					Message:            conditions.ExpiredMsg,
				},
			},
		},
		{
			name:         "ProvisioningRequest wasn't provisioned, expired",
			creationTime: weekAgo,
			conditions: []metav1.Condition{
				{
					Type:               v1beta1.Provisioned,
					Status:             metav1.ConditionFalse,
					LastTransitionTime: metav1.NewTime(dayAgo),
					Reason:             conditions.ExpiredReason,
					Message:            conditions.ExpiredMsg,
				},
			},
			wantConditions: []metav1.Condition{
				{
					Type:               v1beta1.Provisioned,
					Status:             metav1.ConditionFalse,
					LastTransitionTime: metav1.NewTime(dayAgo),
					Reason:             conditions.ExpiredReason,
					Message:            conditions.ExpiredMsg,
				},
				{
					Type:               v1beta1.Failed,
					Status:             metav1.ConditionTrue,
					LastTransitionTime: metav1.NewTime(now),
					Reason:             conditions.ExpiredReason,
					Message:            conditions.ExpiredMsg,
				},
			},
		},
		{
			name:         "BookingCapacity time is expired ",
			creationTime: dayAgo,
			conditions: []metav1.Condition{
				{
					Type:               v1beta1.Provisioned,
					Status:             metav1.ConditionTrue,
					LastTransitionTime: metav1.NewTime(dayAgo),
					Reason:             conditions.ExpiredReason,
					Message:            conditions.ExpiredMsg,
				},
			},
			wantConditions: []metav1.Condition{
				{
					Type:               v1beta1.Provisioned,
					Status:             metav1.ConditionTrue,
					LastTransitionTime: metav1.NewTime(dayAgo),
					Reason:             conditions.ExpiredReason,
					Message:            conditions.ExpiredMsg,
				},
				{
					Type:               v1beta1.BookingExpired,
					Status:             metav1.ConditionTrue,
					LastTransitionTime: metav1.NewTime(now),
					Reason:             conditions.CapacityReservationTimeExpiredReason,
					Message:            conditions.CapacityReservationTimeExpiredMsg,
				},
			},
		},
		{
			name:         "Failed ProvisioningRequest",
			creationTime: dayAgo,
			conditions: []metav1.Condition{
				{
					Type:               v1beta1.Failed,
					Status:             metav1.ConditionTrue,
					LastTransitionTime: metav1.NewTime(dayAgo),
					Reason:             "Failed",
					Message:            "Failed",
				},
			},
			wantConditions: []metav1.Condition{
				{
					Type:               v1beta1.Failed,
					Status:             metav1.ConditionTrue,
					LastTransitionTime: metav1.NewTime(dayAgo),
					Reason:             "Failed",
					Message:            "Failed",
				},
			},
		},
	}
	for _, test := range testCases {
		pr := provreqclient.ProvisioningRequestWrapperForTesting("namespace", "name-1")
		pr.V1Beta1().Status.Conditions = test.conditions
		pr.V1Beta1().CreationTimestamp = metav1.NewTime(test.creationTime)
		pr.V1Beta1().Spec.ProvisioningClassName = v1beta1.ProvisioningClassCheckCapacity
		additionalPr := provreqclient.ProvisioningRequestWrapperForTesting("namespace", "additional")
		additionalPr.V1Beta1().CreationTimestamp = metav1.NewTime(weekAgo)
		additionalPr.V1Beta1().Spec.ProvisioningClassName = v1beta1.ProvisioningClassCheckCapacity
		processor := checkCapacityProcessor{func() time.Time { return now }, 1}
		processor.Process([]*provreqwrapper.ProvisioningRequest{pr, additionalPr})
		assert.ElementsMatch(t, test.wantConditions, pr.Conditions())
		if len(test.conditions) == len(test.wantConditions) {
			assert.ElementsMatch(t, []metav1.Condition{
				{
					Type:               v1beta1.Failed,
					Status:             metav1.ConditionTrue,
					LastTransitionTime: metav1.NewTime(now),
					Reason:             conditions.ExpiredReason,
					Message:            conditions.ExpiredMsg,
				},
			}, additionalPr.Conditions())
		} else {
			assert.ElementsMatch(t, []metav1.Condition{}, additionalPr.Conditions())
		}
	}
}
