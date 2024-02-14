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
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/apis/autoscaling.x-k8s.io/v1beta1"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqwrapper"
	"k8s.io/klog/v2"
)

const (
	//CapacityIsNotFoundReason is added when capacity was not found in the cluster.
	CapacityIsNotFoundReason = "CapacityIsNotFound"
	//CapacityIsFoundReason is added when capacity was found in the cluster.
	CapacityIsFoundReason = "CapacityIsFound"
	//FailedToBookCapacityReason is added when Cluster Autoscaler failed to book capacity in the cluster.
	FailedToBookCapacityReason = "FailedToBookCapacity"
	//CapacityReservationTimeExpiredReason is added whed capacity reservation time is expired.
	CapacityReservationTimeExpiredReason = "CapacityReservationTimeExpired"
	//CapacityReservationTimeExpiredMsg is added if capacity reservation time is expired.
	CapacityReservationTimeExpiredMsg = "Capacity reservation time is expired"
	//ExpiredReason is added if ProvisioningRequest is expired.
	ExpiredReason = "Expired"
	//ExpiredMsg is added if ProvisioningRequest is expired.
	ExpiredMsg = "ProvisioningRequest is expired"
)

func shouldCapacityBeBooked(pr *provreqwrapper.ProvisioningRequest) bool {
	if pr.V1Beta1().Spec.ProvisioningClassName != v1beta1.ProvisioningClassCheckCapacity {
		return false
	}
	if pr.Conditions() == nil || len(pr.Conditions()) == 0 {
		return false
	}
	book := false
	for _, condition := range pr.Conditions() {
		if checkConditionType(condition, v1beta1.BookingExpired) || checkConditionType(condition, v1beta1.Failed) {
			return false
		} else if checkConditionType(condition, v1beta1.Provisioned) {
			book = true
		}
	}
	return book
}

func setCondition(pr *provreqwrapper.ProvisioningRequest, conditionType string, conditionStatus v1.ConditionStatus, reason, message string, now v1.Time) {
	var newConditions []v1.Condition
	newCondition := v1.Condition{
		Type:               conditionType,
		Status:             conditionStatus,
		ObservedGeneration: pr.V1Beta1().GetObjectMeta().GetGeneration(),
		LastTransitionTime: now,
		Reason:             reason,
		Message:            message,
	}
	prevConditions := pr.Conditions()
	switch conditionType {
	case v1beta1.Provisioned, v1beta1.BookingExpired, v1beta1.Failed:
		conditionFound := false
		for _, condition := range prevConditions {
			if condition.Type == conditionType {
				conditionFound = true
				newConditions = append(newConditions, newCondition)
			} else {
				newConditions = append(newConditions, condition)
			}
		}
		if !conditionFound {
			newConditions = append(prevConditions, newCondition)
		}
	default:
		klog.Errorf("Unknown (conditionType; conditionStatus) pair: (%s; %s) ", conditionType, conditionStatus)
		newConditions = prevConditions
	}
	pr.SetConditions(newConditions)
}

func checkConditionType(condition v1.Condition, conditionType string) bool {
	return condition.Type == conditionType && condition.Status == v1.ConditionTrue
}
