/*
Copyright 2018 The Kubernetes Authors.

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

package test

import (
	"time"

	autoscaling "k8s.io/api/autoscaling/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
)

// VerticalPodAutoscalerBuilder helps building test instances of VerticalPodAutoscaler.
type VerticalPodAutoscalerBuilder interface {
	WithName(vpaName string) VerticalPodAutoscalerBuilder
	WithNamespace(namespace string) VerticalPodAutoscalerBuilder
	WithUpdateMode(updateMode vpa_types.UpdateMode) VerticalPodAutoscalerBuilder
	WithCreationTimestamp(timestamp time.Time) VerticalPodAutoscalerBuilder
	WithTargetRef(targetRef *autoscaling.CrossVersionObjectReference) VerticalPodAutoscalerBuilder
	WithAnnotations(map[string]string) VerticalPodAutoscalerBuilder

	AppendCondition(conditionType vpa_types.VerticalPodAutoscalerConditionType,
		status core.ConditionStatus, reason, message string, lastTransitionTime time.Time) VerticalPodAutoscalerBuilder

	AppendContainerResourcePolicy(vpa_types.ContainerResourcePolicy) VerticalPodAutoscalerBuilder
	AppendRecommendation(vpa_types.RecommendedContainerResources) VerticalPodAutoscalerBuilder

	Get() *vpa_types.VerticalPodAutoscaler
}

// VerticalPodAutoscaler returns a new VerticalPodAutoscalerBuilder.
func VerticalPodAutoscaler() VerticalPodAutoscalerBuilder {
	return &verticalPodAutoscalerBuilder{
		namespace:                     "default",
		conditions:                    []vpa_types.VerticalPodAutoscalerCondition{},
		containerResourcePolicies:     []vpa_types.ContainerResourcePolicy{},
		recommendedContainerResources: []vpa_types.RecommendedContainerResources{},
	}
}

type verticalPodAutoscalerBuilder struct {
	vpaName                       string
	namespace                     string
	updatePolicy                  *vpa_types.PodUpdatePolicy
	creationTimestamp             time.Time
	annotations                   map[string]string
	targetRef                     *autoscaling.CrossVersionObjectReference
	conditions                    []vpa_types.VerticalPodAutoscalerCondition
	containerResourcePolicies     []vpa_types.ContainerResourcePolicy
	recommendedContainerResources []vpa_types.RecommendedContainerResources
}

func (b *verticalPodAutoscalerBuilder) WithName(vpaName string) VerticalPodAutoscalerBuilder {
	c := *b
	c.vpaName = vpaName
	return &c
}

func (b *verticalPodAutoscalerBuilder) WithNamespace(namespace string) VerticalPodAutoscalerBuilder {
	c := *b
	c.namespace = namespace
	return &c
}

func (b *verticalPodAutoscalerBuilder) WithUpdateMode(updateMode vpa_types.UpdateMode) VerticalPodAutoscalerBuilder {
	c := *b
	if c.updatePolicy == nil {
		c.updatePolicy = &vpa_types.PodUpdatePolicy{}
	}
	c.updatePolicy.UpdateMode = &updateMode
	return &c
}

func (b *verticalPodAutoscalerBuilder) WithCreationTimestamp(timestamp time.Time) VerticalPodAutoscalerBuilder {
	c := *b
	c.creationTimestamp = timestamp
	return &c
}

func (b *verticalPodAutoscalerBuilder) WithTargetRef(targetRef *autoscaling.CrossVersionObjectReference) VerticalPodAutoscalerBuilder {
	c := *b
	c.targetRef = targetRef
	return &c
}

func (b *verticalPodAutoscalerBuilder) WithAnnotations(annotations map[string]string) VerticalPodAutoscalerBuilder {
	c := *b
	c.annotations = annotations
	return &c
}

func (b *verticalPodAutoscalerBuilder) AppendCondition(conditionType vpa_types.VerticalPodAutoscalerConditionType,
	status core.ConditionStatus, reason, message string, lastTransitionTime time.Time) VerticalPodAutoscalerBuilder {
	c := *b
	c.conditions = append(c.conditions, vpa_types.VerticalPodAutoscalerCondition{
		Type:               conditionType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: meta.NewTime(lastTransitionTime)})
	return &c
}

func (b *verticalPodAutoscalerBuilder) AppendContainerResourcePolicy(containerResourcePolicy vpa_types.ContainerResourcePolicy) VerticalPodAutoscalerBuilder {
	c := *b
	c.containerResourcePolicies = append(c.containerResourcePolicies, containerResourcePolicy)
	return &c
}

func (b *verticalPodAutoscalerBuilder) AppendRecommendation(recommendation vpa_types.RecommendedContainerResources) VerticalPodAutoscalerBuilder {
	c := *b
	c.recommendedContainerResources = append(c.recommendedContainerResources, recommendation)
	return &c
}

func (b *verticalPodAutoscalerBuilder) Get() *vpa_types.VerticalPodAutoscaler {
	resourcePolicy := vpa_types.PodResourcePolicy{ContainerPolicies: b.containerResourcePolicies}

	return &vpa_types.VerticalPodAutoscaler{
		ObjectMeta: meta.ObjectMeta{
			Name:              b.vpaName,
			Namespace:         b.namespace,
			Annotations:       b.annotations,
			CreationTimestamp: meta.NewTime(b.creationTimestamp),
		},
		Spec: vpa_types.VerticalPodAutoscalerSpec{
			UpdatePolicy:   b.updatePolicy,
			ResourcePolicy: &resourcePolicy,
			TargetRef:      b.targetRef,
		},
		Status: vpa_types.VerticalPodAutoscalerStatus{
			Recommendation: &vpa_types.RecommendedPodResources{
				ContainerRecommendations: b.recommendedContainerResources,
			},
			Conditions:     b.conditions,
		},
	}
}
