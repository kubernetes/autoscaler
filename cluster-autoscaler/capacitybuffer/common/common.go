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
	"time"

	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Constants to use in Capacity Buffers objects
const (
	ActiveProvisioningStrategy    = "buffer.x-k8s.io/active-capacity"
	CapacityBufferKind            = "CapacityBuffer"
	CapacityBufferApiVersion      = "autoscaling.x-k8s.io/v1alpha1"
	ReadyForProvisioningCondition = "ReadyForProvisioning"
	ProvisioningCondition         = "Provisioning"
	ConditionTrue                 = "True"
	ConditionFalse                = "False"
)

// SetBufferAsReadyForProvisioning updates the passed buffer object with the rest of the attributes and sets its condition to ready
func SetBufferAsReadyForProvisioning(buffer *v1.CapacityBuffer, PodTemplateRef *v1.LocalObjectRef, podTemplateGeneration *int64, replicas *int32, provStrategy *string) {
	buffer.Status.PodTemplateRef = PodTemplateRef
	buffer.Status.Replicas = replicas
	buffer.Status.PodTemplateGeneration = podTemplateGeneration
	buffer.Status.ProvisioningStrategy = mapEmptyProvStrategyToDefault(provStrategy)
	readyCondition := metav1.Condition{
		Type:               ReadyForProvisioningCondition,
		Status:             ConditionTrue,
		Message:            "ready",
		Reason:             "atrtibutesSetSuccessfully",
		LastTransitionTime: metav1.Time{Time: time.Now()},
	}
	buffer.Status.Conditions = []metav1.Condition{readyCondition}
}

// SetBufferAsNotReadyForProvisioning updates the passed buffer object with the rest of the attributes and sets its condition to not ready with the passed error
func SetBufferAsNotReadyForProvisioning(buffer *v1.CapacityBuffer, PodTemplateRef *v1.LocalObjectRef, podTemplateGeneration *int64, replicas *int32, provStrategy *string, err error) {
	errorMessage := "Buffer not ready for provisioing"
	if err != nil {
		errorMessage = err.Error()
	}

	buffer.Status.PodTemplateRef = PodTemplateRef
	buffer.Status.Replicas = replicas
	buffer.Status.PodTemplateGeneration = podTemplateGeneration
	buffer.Status.ProvisioningStrategy = mapEmptyProvStrategyToDefault(provStrategy)
	notReadyCondition := metav1.Condition{
		Type:               ReadyForProvisioningCondition,
		Status:             ConditionFalse,
		Message:            errorMessage,
		Reason:             "error",
		LastTransitionTime: metav1.Time{Time: time.Now()},
	}
	buffer.Status.Conditions = []metav1.Condition{notReadyCondition}
}

func mapEmptyProvStrategyToDefault(ps *string) *string {
	if ps != nil && *ps == "" {
		defaultProvStrategy := ActiveProvisioningStrategy
		ps = &defaultProvStrategy
	}
	return ps
}

// UpdateBufferStatusToFailedProvisioing updates the status of the passed buffer and set Provisioning to false with the passes reason and message
func UpdateBufferStatusToFailedProvisioing(buffer *v1.CapacityBuffer, reason, errorMessage string) {
	buffer.Status.Conditions = []metav1.Condition{{
		Type:               ProvisioningCondition,
		Status:             ConditionFalse,
		Message:            errorMessage,
		Reason:             reason,
		LastTransitionTime: metav1.Time{Time: time.Now()},
	}}
}

// UpdateBufferStatusToSuccessfullyProvisioing updates the status of the passed buffer and set Provisioning to true with the passes reason
func UpdateBufferStatusToSuccessfullyProvisioing(buffer *v1.CapacityBuffer, reason string) {
	buffer.Status.Conditions = []metav1.Condition{{
		Type:               ProvisioningCondition,
		Status:             ConditionTrue,
		Reason:             reason,
		LastTransitionTime: metav1.Time{Time: time.Now()},
	}}
}
