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

package conditions

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/apis/provisioningrequest/autoscaling.x-k8s.io/v1"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqwrapper"
)

func TestBookCapacity(t *testing.T) {
	tests := []struct {
		name         string
		prConditions []metav1.Condition
		want         bool
	}{
		{
			name: "BookingExpired",
			prConditions: []metav1.Condition{
				{
					Type:   v1.Provisioned,
					Status: metav1.ConditionTrue,
				},
				{
					Type:   v1.BookingExpired,
					Status: metav1.ConditionTrue,
				},
			},
			want: false,
		},
		{
			name: "Failed",
			prConditions: []metav1.Condition{
				{
					Type:   v1.Provisioned,
					Status: metav1.ConditionTrue,
				},
				{
					Type:   v1.Failed,
					Status: metav1.ConditionTrue,
				},
			},
			want: false,
		},
		{
			name: "empty conditions",
			want: false,
		},
		{
			name: "Capacity found and provisioned",
			prConditions: []metav1.Condition{
				{
					Type:   v1.Provisioned,
					Status: metav1.ConditionTrue,
				},
				{
					Type:   v1.Provisioned,
					Status: metav1.ConditionTrue,
				},
			},
			want: true,
		},
		{
			name: "Capacity is not found",
			prConditions: []metav1.Condition{
				{
					Type:   v1.Provisioned,
					Status: metav1.ConditionFalse,
				},
			},
			want: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for class := range provisioningrequest.SupportedProvisioningClasses {
				pr := provreqwrapper.NewProvisioningRequest(
					&v1.ProvisioningRequest{
						Spec: v1.ProvisioningRequestSpec{
							ProvisioningClassName: class,
						},
						Status: v1.ProvisioningRequestStatus{
							Conditions: test.prConditions,
						},
					}, nil)
				got := ShouldCapacityBeBooked(pr)
				if got != test.want {
					t.Errorf("Want: %v, got: %v", test.want, got)
				}
			}
		})
	}
}

func TestSetCondition(t *testing.T) {
	tests := []struct {
		name          string
		oldConditions []metav1.Condition
		newType       string
		newStatus     metav1.ConditionStatus
		want          []metav1.Condition
	}{
		{
			name:      "Accepted added, empty conditions before",
			newType:   v1.Accepted,
			newStatus: metav1.ConditionTrue,
			want: []metav1.Condition{
				{
					Type:   v1.Accepted,
					Status: metav1.ConditionTrue,
				},
			},
		},
		{
			name:      "Provisioned added, empty conditions before",
			newType:   v1.Provisioned,
			newStatus: metav1.ConditionTrue,
			want: []metav1.Condition{
				{
					Type:   v1.Provisioned,
					Status: metav1.ConditionTrue,
				},
			},
		},
		{
			name: "Provisioned updated",
			oldConditions: []metav1.Condition{
				{
					Type:   v1.Provisioned,
					Status: metav1.ConditionFalse,
				},
			},
			newType:   v1.Provisioned,
			newStatus: metav1.ConditionTrue,
			want: []metav1.Condition{
				{
					Type:   v1.Provisioned,
					Status: metav1.ConditionTrue,
				},
			},
		},
		{
			name: "Failed added, non-empty conditions before",
			oldConditions: []metav1.Condition{
				{
					Type:   v1.Provisioned,
					Status: metav1.ConditionTrue,
				},
			},
			newType:   v1.Failed,
			newStatus: metav1.ConditionTrue,
			want: []metav1.Condition{
				{
					Type:   v1.Provisioned,
					Status: metav1.ConditionTrue,
				},
				{
					Type:   v1.Failed,
					Status: metav1.ConditionTrue,
				},
			},
		},
		{
			name: "Unknown condition status, conditions are updated",
			oldConditions: []metav1.Condition{
				{
					Type:   v1.Provisioned,
					Status: metav1.ConditionTrue,
				},
			},
			newType:   v1.Failed,
			newStatus: metav1.ConditionUnknown,
			want: []metav1.Condition{
				{
					Type:   v1.Provisioned,
					Status: metav1.ConditionTrue,
				},
				{
					Type:   v1.Failed,
					Status: metav1.ConditionUnknown,
				},
			},
		},
		{
			name: "Unknown condition type, conditions are not updated",
			oldConditions: []metav1.Condition{
				{
					Type:   v1.Provisioned,
					Status: metav1.ConditionTrue,
				},
			},
			newType:   "Unknown",
			newStatus: metav1.ConditionTrue,
			want: []metav1.Condition{
				{
					Type:   v1.Provisioned,
					Status: metav1.ConditionTrue,
				},
			},
		},
		{
			name:      "BookingExpired, empty conditions before",
			newType:   v1.BookingExpired,
			newStatus: metav1.ConditionFalse,
			want: []metav1.Condition{
				{
					Type:   v1.BookingExpired,
					Status: metav1.ConditionFalse,
				},
			},
		},
		{
			name: "Capacity found with unknown condition before",
			oldConditions: []metav1.Condition{
				{
					Type:   "unknown",
					Status: metav1.ConditionTrue,
				},
			},
			newType:   v1.Provisioned,
			newStatus: metav1.ConditionTrue,
			want: []metav1.Condition{
				{
					Type:   "unknown",
					Status: metav1.ConditionTrue,
				},
				{
					Type:   v1.Provisioned,
					Status: metav1.ConditionTrue,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pr := provreqwrapper.NewProvisioningRequest(
				&v1.ProvisioningRequest{
					Status: v1.ProvisioningRequestStatus{
						Conditions: test.oldConditions,
					},
				}, nil)
			AddOrUpdateCondition(pr, test.newType, test.newStatus, "", "", metav1.Now())
			got := pr.Status.Conditions
			if len(got) > 2 || len(got) != len(test.want) || got[0].Type != test.want[0].Type || got[0].Status != test.want[0].Status {
				t.Errorf("want %v, got: %v", test.want, got)
			}
			if len(got) == 2 {
				if got[1].Type != test.want[1].Type || got[1].Status != test.want[1].Status {
					t.Errorf("want %v, got: %v", test.want, got)
				}
			}
		})
	}
}
