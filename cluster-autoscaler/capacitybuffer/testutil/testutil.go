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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1beta1"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer"
)

// To use their pointers in creating testing capacity buffer objects
var (
	ProvisioningStrategy      = capacitybuffer.ActiveProvisioningStrategy
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
	return GetConditionReadyWithMessage("")
}

// GetConditionReadyWithMessage returns a list of conditions with a condition ready and the specified message
func GetConditionReadyWithMessage(message string) []metav1.Condition {
	readyCondition := metav1.Condition{
		Type:               capacitybuffer.ReadyForProvisioningCondition,
		Status:             metav1.ConditionTrue,
		Message:            message,
		Reason:             capacitybuffer.AttributesSetSuccessfullyReason,
		LastTransitionTime: metav1.Time{},
	}
	return []metav1.Condition{readyCondition}
}

// GetConditionNotReady returns a list of conditions with a condition not ready and empty message, should be used for testing purposes only
func GetConditionNotReady() []metav1.Condition {
	return GetConditionNotReadyWithMessage("")
}

// GetConditionNotReadyWithMessage returns a list of condition with a condition not ready and specified message.
func GetConditionNotReadyWithMessage(message string) []metav1.Condition {
	notReadyCondition := metav1.Condition{
		Type:               capacitybuffer.ReadyForProvisioningCondition,
		Status:             metav1.ConditionFalse,
		Message:            message,
		Reason:             "error",
		LastTransitionTime: metav1.Time{},
	}
	return []metav1.Condition{notReadyCondition}
}

// BufferOption is a functional option for creating a CapacityBuffer
type BufferOption func(*v1.CapacityBuffer)

// NewBuffer creates a new CapacityBuffer with the given options
func NewBuffer(opts ...BufferOption) *v1.CapacityBuffer {
	b := &v1.CapacityBuffer{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
		},
	}
	for _, opt := range opts {
		opt(b)
	}
	return b
}

// WithName sets the CapacityBuffer name
func WithName(name string) BufferOption {
	return func(b *v1.CapacityBuffer) {
		b.Name = name
	}
}

// WithPodTemplateRef sets the Spec.PodTemplateRef
func WithPodTemplateRef(name string) BufferOption {
	return func(b *v1.CapacityBuffer) {
		b.Spec.PodTemplateRef = &v1.LocalObjectRef{Name: name}
	}
}

// WithReplicas sets the Spec.Replicas
func WithReplicas(replicas int32) BufferOption {
	return func(b *v1.CapacityBuffer) {
		b.Spec.Replicas = &replicas
	}
}

// WithLimits sets the Spec.Limits
func WithLimits(limits v1.ResourceList) BufferOption {
	return func(b *v1.CapacityBuffer) {
		b.Spec.Limits = &limits
	}
}

// WithStatusPodTemplateRef sets the Status.PodTemplateRef
func WithStatusPodTemplateRef(name string) BufferOption {
	return func(b *v1.CapacityBuffer) {
		b.Status.PodTemplateRef = &v1.LocalObjectRef{Name: name}
	}
}

// WithStatusReplicas sets the Status.Replicas
func WithStatusReplicas(replicas int32) BufferOption {
	return func(b *v1.CapacityBuffer) {
		b.Status.Replicas = &replicas
	}
}

// WithActiveProvisioningStrategy sets the ProvisioningStrategy to ActiveProvisioningStrategy
func WithActiveProvisioningStrategy() BufferOption {
	return func(b *v1.CapacityBuffer) {
		strategy := capacitybuffer.ActiveProvisioningStrategy
		b.Spec.ProvisioningStrategy = &strategy
	}
}

// PodTemplateOption is a functional option for creating a PodTemplate
type PodTemplateOption func(*corev1.PodTemplate)

// NewPodTemplate creates a new PodTemplate with the given options
func NewPodTemplate(opts ...PodTemplateOption) *corev1.PodTemplate {
	pt := &corev1.PodTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-pod-template",
			Namespace: "default",
		},
		Template: corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name: "container",
					},
				},
			},
		},
	}
	for _, opt := range opts {
		opt(pt)
	}
	return pt
}

// WithPodTemplateName sets the PodTemplate name
func WithPodTemplateName(name string) PodTemplateOption {
	return func(pt *corev1.PodTemplate) {
		pt.Name = name
	}
}

// WithPodTemplateResources sets the PodTemplate resources
func WithPodTemplateResources(requests, limits corev1.ResourceList) PodTemplateOption {
	return func(pt *corev1.PodTemplate) {
		if len(pt.Template.Spec.Containers) == 0 {
			pt.Template.Spec.Containers = append(pt.Template.Spec.Containers, corev1.Container{Name: "container"})
		}
		pt.Template.Spec.Containers[0].Resources.Requests = requests
		pt.Template.Spec.Containers[0].Resources.Limits = limits
	}
}

// WithPodTemplatePriorityClassName sets the PriorityClassName
func WithPodTemplatePriorityClassName(name string) PodTemplateOption {
	return func(pt *corev1.PodTemplate) {
		pt.Template.Spec.PriorityClassName = name
	}
}

// WithPodTemplateAffinity sets the Affinity
func WithPodTemplateAffinity(affinity *corev1.Affinity) PodTemplateOption {
	return func(pt *corev1.PodTemplate) {
		pt.Template.Spec.Affinity = affinity
	}
}

// ResourceQuotaOption is a functional option for creating a ResourceQuota
type ResourceQuotaOption func(*corev1.ResourceQuota)

// NewResourceQuota creates a new ResourceQuota with the given options
func NewResourceQuota(opts ...ResourceQuotaOption) *corev1.ResourceQuota {
	rq := &corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-quota",
			Namespace: "default",
		},
	}
	for _, opt := range opts {
		opt(rq)
	}
	return rq
}

// WithResourceQuotaName sets the ResourceQuota name
func WithResourceQuotaName(name string) ResourceQuotaOption {
	return func(rq *corev1.ResourceQuota) {
		rq.Name = name
	}
}

// WithResourceQuotaHard sets the ResourceQuota hard limits
func WithResourceQuotaHard(hard corev1.ResourceList) ResourceQuotaOption {
	return func(rq *corev1.ResourceQuota) {
		rq.Status.Hard = hard
		rq.Spec.Hard = hard
	}
}

// WithResourceQuotaUsed sets the ResourceQuota used resources
func WithResourceQuotaUsed(used corev1.ResourceList) ResourceQuotaOption {
	return func(rq *corev1.ResourceQuota) {
		rq.Status.Used = used
	}
}

// WithResourceQuotaScopes sets the ResourceQuota scopes
func WithResourceQuotaScopes(scopes []corev1.ResourceQuotaScope) ResourceQuotaOption {
	return func(rq *corev1.ResourceQuota) {
		rq.Spec.Scopes = scopes
	}
}

// WithResourceQuotaScopeSelector sets the ResourceQuota scope selector
func WithResourceQuotaScopeSelector(selector *corev1.ScopeSelector) ResourceQuotaOption {
	return func(rq *corev1.ResourceQuota) {
		rq.Spec.ScopeSelector = selector
	}
}
