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

package testutil

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1alpha1"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/common"
)

// To use their pointers in creating testing capacity buffer objects
var (
	ProvisioningStrategy      = common.ActiveProvisioningStrategy
	SomeNumberOfReplicas      = int32(3)
	AnotherNumberOfReplicas   = int32(5)
	SomePodTemplateRefName    = "some-pod-template"
	AnotherPodTemplateRefName = "another-pod-template"
)

// SanitizeBuffersStatus returns a list of the status objects of the passed buffers after sanitizing them for testing comparison
func SanitizeBuffersStatus(buffers []*v1.CapacityBuffer) []*v1.CapacityBufferStatus {
	resultedStatus := []*v1.CapacityBufferStatus{}
	for _, buffer := range buffers {
		for i := range buffer.Status.Conditions {
			buffer.Status.Conditions[i].LastTransitionTime = metav1.Time{}
			buffer.Status.Conditions[i].Message = ""
		}
		resultedStatus = append(resultedStatus, &buffer.Status)
	}
	return resultedStatus
}

// GetPodTemplateRefBuffer returns a buffer with podTemplateRef with the passed attributes and empty status, should be used for testing purposes only
func GetPodTemplateRefBuffer(podTemplateRef *v1.LocalObjectRef, replicas *int32) *v1.CapacityBuffer {
	return &v1.CapacityBuffer{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
		},
		Spec: v1.CapacityBufferSpec{
			ProvisioningStrategy: &ProvisioningStrategy,
			PodTemplateRef:       podTemplateRef,
			ScalableRef:          nil,
			Replicas:             replicas,
			Percentage:           nil,
			Limits:               nil,
		},
		Status: *GetBufferStatus(nil, nil, nil, nil, nil),
	}
}

// GetBuffer returns a capacity buffer with the passed attributes, should be used for testing purposes only
func GetBuffer(strategy *string, podTemplateRef *v1.LocalObjectRef, replicas *int32, statusPodTempRef *v1.LocalObjectRef,
	statusReplicas *int32, podTemplateGeneration *int64, conditions []metav1.Condition,
	limits *v1.ResourceList) *v1.CapacityBuffer {
	return &v1.CapacityBuffer{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
		},
		Spec: v1.CapacityBufferSpec{
			ProvisioningStrategy: strategy,
			PodTemplateRef:       podTemplateRef,
			ScalableRef:          nil,
			Replicas:             replicas,
			Percentage:           nil,
			Limits:               limits,
		},
		Status: *GetBufferStatus(statusPodTempRef, statusReplicas, podTemplateGeneration, strategy, conditions),
	}
}

// GetBufferStatus returns a buffer status with the passed attributes, should be used for testing purposes only
func GetBufferStatus(podTempRef *v1.LocalObjectRef, replicas *int32, podTemplateGeneration *int64, provisioningStrategy *string, conditions []metav1.Condition) *v1.CapacityBufferStatus {
	return &v1.CapacityBufferStatus{
		PodTemplateRef:        podTempRef,
		Replicas:              replicas,
		PodTemplateGeneration: podTemplateGeneration,
		Conditions:            conditions,
		ProvisioningStrategy:  provisioningStrategy,
	}
}

// GetConditionReady returns a list of conditions with a condition ready and empty message, should be used for testing purposes only
func GetConditionReady() []metav1.Condition {
	readyCondition := metav1.Condition{
		Type:               common.ReadyForProvisioningCondition,
		Status:             common.ConditionTrue,
		Message:            "",
		Reason:             "atrtibutesSetSuccessfully",
		LastTransitionTime: metav1.Time{},
	}
	return []metav1.Condition{readyCondition}
}

// GetConditionNotReady returns a list of conditions with a condition not ready and empty message, should be used for testing purposes only
func GetConditionNotReady() []metav1.Condition {
	notReadyCondition := metav1.Condition{
		Type:               common.ReadyForProvisioningCondition,
		Status:             common.ConditionFalse,
		Message:            "",
		Reason:             "error",
		LastTransitionTime: metav1.Time{},
	}
	return []metav1.Condition{notReadyCondition}
}
