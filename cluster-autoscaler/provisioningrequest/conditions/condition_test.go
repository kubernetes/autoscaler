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

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/apis/provisioningrequest/autoscaling.x-k8s.io/v1beta1"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqwrapper"
)

func TestBookCapacity(t *testing.T) {
	tests := []struct {
		name         string
		prConditions []v1.Condition
		want         bool
	}{
		{
			name: "BookingExpired",
			prConditions: []v1.Condition{
				{
					Type:   v1beta1.Provisioned,
					Status: v1.ConditionTrue,
				},
				{
					Type:   v1beta1.BookingExpired,
					Status: v1.ConditionTrue,
				},
			},
			want: false,
		},
		{
			name: "Failed",
			prConditions: []v1.Condition{
				{
					Type:   v1beta1.Provisioned,
					Status: v1.ConditionTrue,
				},
				{
					Type:   v1beta1.Failed,
					Status: v1.ConditionTrue,
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
			prConditions: []v1.Condition{
				{
					Type:   v1beta1.Provisioned,
					Status: v1.ConditionTrue,
				},
				{
					Type:   v1beta1.Provisioned,
					Status: v1.ConditionTrue,
				},
			},
			want: true,
		},
		{
			name: "Capacity is not found",
			prConditions: []v1.Condition{
				{
					Type:   v1beta1.Provisioned,
					Status: v1.ConditionFalse,
				},
			},
			want: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pr := provreqwrapper.NewV1Beta1ProvisioningRequest(
				&v1beta1.ProvisioningRequest{
					Spec: v1beta1.ProvisioningRequestSpec{
						ProvisioningClassName: v1beta1.ProvisioningClassCheckCapacity,
					},
					Status: v1beta1.ProvisioningRequestStatus{
						Conditions: test.prConditions,
					},
				}, nil)
			got := ShouldCapacityBeBooked(pr)
			if got != test.want {
				t.Errorf("Want: %v, got: %v", test.want, got)
			}
		})
	}
}

func TestSetCondition(t *testing.T) {
	tests := []struct {
		name          string
		oldConditions []v1.Condition
		newType       string
		newStatus     v1.ConditionStatus
		want          []v1.Condition
	}{
		{
			name:      "Accepted added, empty conditions before",
			newType:   v1beta1.Accepted,
			newStatus: v1.ConditionTrue,
			want: []v1.Condition{
				{
					Type:   v1beta1.Accepted,
					Status: v1.ConditionTrue,
				},
			},
		},
		{
			name:      "Provisioned added, empty conditions before",
			newType:   v1beta1.Provisioned,
			newStatus: v1.ConditionTrue,
			want: []v1.Condition{
				{
					Type:   v1beta1.Provisioned,
					Status: v1.ConditionTrue,
				},
			},
		},
		{
			name: "Provisioned updated",
			oldConditions: []v1.Condition{
				{
					Type:   v1beta1.Provisioned,
					Status: v1.ConditionFalse,
				},
			},
			newType:   v1beta1.Provisioned,
			newStatus: v1.ConditionTrue,
			want: []v1.Condition{
				{
					Type:   v1beta1.Provisioned,
					Status: v1.ConditionTrue,
				},
			},
		},
		{
			name: "Failed added, non-empty conditions before",
			oldConditions: []v1.Condition{
				{
					Type:   v1beta1.Provisioned,
					Status: v1.ConditionTrue,
				},
			},
			newType:   v1beta1.Failed,
			newStatus: v1.ConditionTrue,
			want: []v1.Condition{
				{
					Type:   v1beta1.Provisioned,
					Status: v1.ConditionTrue,
				},
				{
					Type:   v1beta1.Failed,
					Status: v1.ConditionTrue,
				},
			},
		},
		{
			name: "Unknown condition status, conditions are updated",
			oldConditions: []v1.Condition{
				{
					Type:   v1beta1.Provisioned,
					Status: v1.ConditionTrue,
				},
			},
			newType:   v1beta1.Failed,
			newStatus: v1.ConditionUnknown,
			want: []v1.Condition{
				{
					Type:   v1beta1.Provisioned,
					Status: v1.ConditionTrue,
				},
				{
					Type:   v1beta1.Failed,
					Status: v1.ConditionUnknown,
				},
			},
		},
		{
			name: "Unknown condition type, conditions are not updated",
			oldConditions: []v1.Condition{
				{
					Type:   v1beta1.Provisioned,
					Status: v1.ConditionTrue,
				},
			},
			newType:   "Unknown",
			newStatus: v1.ConditionTrue,
			want: []v1.Condition{
				{
					Type:   v1beta1.Provisioned,
					Status: v1.ConditionTrue,
				},
			},
		},
		{
			name:      "BookingExpired, empty conditions before",
			newType:   v1beta1.BookingExpired,
			newStatus: v1.ConditionFalse,
			want: []v1.Condition{
				{
					Type:   v1beta1.BookingExpired,
					Status: v1.ConditionFalse,
				},
			},
		},
		{
			name: "Capacity found with unknown condition before",
			oldConditions: []v1.Condition{
				{
					Type:   "unknown",
					Status: v1.ConditionTrue,
				},
			},
			newType:   v1beta1.Provisioned,
			newStatus: v1.ConditionTrue,
			want: []v1.Condition{
				{
					Type:   "unknown",
					Status: v1.ConditionTrue,
				},
				{
					Type:   v1beta1.Provisioned,
					Status: v1.ConditionTrue,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pr := provreqwrapper.NewV1Beta1ProvisioningRequest(
				&v1beta1.ProvisioningRequest{
					Status: v1beta1.ProvisioningRequestStatus{
						Conditions: test.oldConditions,
					},
				}, nil)
			AddOrUpdateCondition(pr, test.newType, test.newStatus, "", "", v1.Now())
			got := pr.Conditions()
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
