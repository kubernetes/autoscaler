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
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqwrapper"
)

// ProvisioningRequestCondition is a type of Condition that ClusterAutoscaler appends to ProvisioningRequest.
type ProvisioningRequestCondition string

const (
	// BookCapacityCondition is appended if capacity for ProvisioningRequest was found in the cluster.
	BookCapacityCondition = ProvisioningRequestCondition("BookCapacity")

	// ExpiredCondition is append if the ProvisioningRequest has BookCapacity condition before
	// and the reservation time is expired or the ProvisioningRequest has Pending condition before
	// and expiration time is expired.
	ExpiredCondition = ProvisioningRequestCondition("Expired")

	// PendingCondition is append if no capacity for ProvisioningRequest was found in the cluster
	// and ClusterAutoscaler will try to find capacity later.
	PendingCondition = ProvisioningRequestCondition("Pending")

	// RejectedCondition is append if ProvisioningRequest is invalid.
	RejectedCondition = ProvisioningRequestCondition("Rejected")

	CheckCapacityClass      = "check-capacity.kubernetes.io"
	DefaultReservationTime = 10 * time.Minute
	DefaultExpirationTime  = 7 * 24 * time.Hour // 7 days
)

// HasBookCapacityCondition return if PR has BookCapacity condition
func HasBookCapacityCondition(pr *provreqwrapper.ProvisioningRequest) bool {
	if pr.V1Beta1().Spec.ProvisioningClassName != CheckCapacityClass {
		return false
	}
	if pr.Conditions() == nil || len(pr.Conditions()) == 0 {
		return false
	}
	condition := pr.Conditions()[len(pr.Conditions())-1]
	if condition.Type == string(BookCapacityCondition) && condition.Status == v1.ConditionTrue {
		return true
	}
	return false
}

func SetCondition(pr *provreqwrapper.ProvisioningRequest, conditionType ProvisioningRequestCondition, reason, message string) {
	conditions := pr.Conditions()
	conditions = append(conditions, v1.Condition{
		Type:               string(conditionType),
		Status:             v1.ConditionTrue,
		ObservedGeneration: pr.V1Beta1().GetObjectMeta().GetGeneration(),
		LastTransitionTime: v1.Now(),
		Reason:             reason,
		Message:            message,
	})
	pr.SetConditions(conditions)
}
