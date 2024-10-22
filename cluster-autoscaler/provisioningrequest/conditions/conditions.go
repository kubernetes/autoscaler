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
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/apis/provisioningrequest/autoscaling.x-k8s.io/v1"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqwrapper"
	"k8s.io/klog/v2"
)

const (
	// AcceptedReason is added when ProvisioningRequest is accepted by ClusterAutoscaler
	AcceptedReason = "Accepted"
	// AcceptedMsg is added when ProvisioningRequest is accepted by ClusterAutoscaler
	AcceptedMsg = "ProvisioningRequest is accepted by ClusterAutoscaler"
	// CapacityIsNotFoundReason is added when capacity was not found in the cluster.
	CapacityIsNotFoundReason = "CapacityIsNotFound"
	// CapacityIsFoundReason is added when capacity was found in the cluster.
	CapacityIsFoundReason = "CapacityIsFound"
	// CapacityIsFoundMsg is added when capacity was found in the cluster.
	CapacityIsFoundMsg = "Capacity is found in the cluster"
	// CapacityIsProvisionedReason is added when capacity was requested successfully.
	CapacityIsProvisionedReason = "CapacityIsProvisioned"
	// CapacityIsProvisionedMsg is added when capacity was requested successfully.
	CapacityIsProvisionedMsg = "Capacity is found in the cluster"
	// FailedToCheckCapacityReason is added when CA failed to check pre-existing capacity.
	FailedToCheckCapacityReason = "FailedToCheckCapacity"
	// FailedToCheckCapacityMsg is added when CA failed to check pre-existing capacity.
	FailedToCheckCapacityMsg = "Failed to check pre-existing capacity in the cluster"
	// FailedToCreatePodsReason is added when CA failed to create pods for ProvisioningRequest.
	FailedToCreatePodsReason = "FailedToCreatePods"
	// FailedToBookCapacityReason is added when Cluster Autoscaler failed to book capacity in the cluster.
	FailedToBookCapacityReason = "FailedToBookCapacity"
	// CapacityReservationTimeExpiredReason is added whed capacity reservation time is expired.
	CapacityReservationTimeExpiredReason = "CapacityReservationTimeExpired"
	// CapacityReservationTimeExpiredMsg is added if capacity reservation time is expired.
	CapacityReservationTimeExpiredMsg = "Capacity reservation time is expired"
	// ExpiredReason is added if ProvisioningRequest is expired.
	ExpiredReason = "Expired"
	// ExpiredMsg is added if ProvisioningRequest is expired.
	ExpiredMsg = "ProvisioningRequest is expired"
)

// ShouldCapacityBeBooked returns whether capacity should be booked.
func ShouldCapacityBeBooked(pr *provreqwrapper.ProvisioningRequest) bool {
	if ok, found := provisioningrequest.SupportedProvisioningClasses[pr.Spec.ProvisioningClassName]; !ok || !found {
		return false
	}
	conditions := pr.Status.Conditions
	if apimeta.IsStatusConditionTrue(conditions, v1.Failed) || apimeta.IsStatusConditionTrue(conditions, v1.BookingExpired) {
		return false
	} else if apimeta.IsStatusConditionTrue(conditions, v1.Provisioned) {
		return true
	}
	return false
}

// AddOrUpdateCondition adds a Condition if the condition is not present amond ProvisioningRequest conditions or updte it otherwise.
func AddOrUpdateCondition(pr *provreqwrapper.ProvisioningRequest, conditionType string, conditionStatus metav1.ConditionStatus, reason, message string, now metav1.Time) {
	var newConditions []metav1.Condition
	newCondition := metav1.Condition{
		Type:               conditionType,
		Status:             conditionStatus,
		ObservedGeneration: pr.GetObjectMeta().GetGeneration(),
		LastTransitionTime: now,
		Reason:             reason,
		Message:            message,
	}
	prevConditions := pr.Status.Conditions
	switch conditionType {
	case v1.Provisioned, v1.BookingExpired, v1.Failed, v1.Accepted:
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
