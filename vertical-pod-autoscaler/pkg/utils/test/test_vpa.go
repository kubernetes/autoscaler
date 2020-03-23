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
	WithContainer(containerName string) VerticalPodAutoscalerBuilder
	WithNamespace(namespace string) VerticalPodAutoscalerBuilder
	WithUpdateMode(updateMode vpa_types.UpdateMode) VerticalPodAutoscalerBuilder
	WithCreationTimestamp(timestamp time.Time) VerticalPodAutoscalerBuilder
	WithMinAllowed(cpu, memory string) VerticalPodAutoscalerBuilder
	WithMaxAllowed(cpu, memory string) VerticalPodAutoscalerBuilder
	WithTarget(cpu, memory string) VerticalPodAutoscalerBuilder
	WithLowerBound(cpu, memory string) VerticalPodAutoscalerBuilder
	WithTargetRef(targetRef *autoscaling.CrossVersionObjectReference) VerticalPodAutoscalerBuilder
	WithUpperBound(cpu, memory string) VerticalPodAutoscalerBuilder
	WithAnnotations(map[string]string) VerticalPodAutoscalerBuilder
	AppendCondition(conditionType vpa_types.VerticalPodAutoscalerConditionType,
		status core.ConditionStatus, reason, message string, lastTransitionTime time.Time) VerticalPodAutoscalerBuilder
	AppendRecommendation(vpa_types.RecommendedContainerResources) VerticalPodAutoscalerBuilder
	Get() *vpa_types.VerticalPodAutoscaler
}

// VerticalPodAutoscaler returns a new VerticalPodAutoscalerBuilder.
func VerticalPodAutoscaler() VerticalPodAutoscalerBuilder {
	return &verticalPodAutoscalerBuilder{
		recommendation:          Recommendation(),
		appendedRecommendations: []vpa_types.RecommendedContainerResources{},
		namespace:               "default",
		conditions:              []vpa_types.VerticalPodAutoscalerCondition{},
	}
}

type verticalPodAutoscalerBuilder struct {
	vpaName                 string
	containerName           string
	namespace               string
	updatePolicy            *vpa_types.PodUpdatePolicy
	creationTimestamp       time.Time
	minAllowed              core.ResourceList
	maxAllowed              core.ResourceList
	recommendation          RecommendationBuilder
	conditions              []vpa_types.VerticalPodAutoscalerCondition
	annotations             map[string]string
	targetRef               *autoscaling.CrossVersionObjectReference
	appendedRecommendations []vpa_types.RecommendedContainerResources
}

func (b *verticalPodAutoscalerBuilder) WithName(vpaName string) VerticalPodAutoscalerBuilder {
	c := *b
	c.vpaName = vpaName
	return &c
}

func (b *verticalPodAutoscalerBuilder) WithContainer(containerName string) VerticalPodAutoscalerBuilder {
	c := *b
	c.containerName = containerName
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

func (b *verticalPodAutoscalerBuilder) WithMinAllowed(cpu, memory string) VerticalPodAutoscalerBuilder {
	c := *b
	c.minAllowed = Resources(cpu, memory)
	return &c
}

func (b *verticalPodAutoscalerBuilder) WithMaxAllowed(cpu, memory string) VerticalPodAutoscalerBuilder {
	c := *b
	c.maxAllowed = Resources(cpu, memory)
	return &c
}

func (b *verticalPodAutoscalerBuilder) WithTarget(cpu, memory string) VerticalPodAutoscalerBuilder {
	c := *b
	c.recommendation = c.recommendation.WithTarget(cpu, memory)
	return &c
}

func (b *verticalPodAutoscalerBuilder) WithLowerBound(cpu, memory string) VerticalPodAutoscalerBuilder {
	c := *b
	c.recommendation = c.recommendation.WithLowerBound(cpu, memory)
	return &c
}

func (b *verticalPodAutoscalerBuilder) WithUpperBound(cpu, memory string) VerticalPodAutoscalerBuilder {
	c := *b
	c.recommendation = c.recommendation.WithUpperBound(cpu, memory)
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

func (b *verticalPodAutoscalerBuilder) AppendRecommendation(recommendation vpa_types.RecommendedContainerResources) VerticalPodAutoscalerBuilder {
	c := *b
	c.appendedRecommendations = append(c.appendedRecommendations, recommendation)
	return &c
}

func (b *verticalPodAutoscalerBuilder) Get() *vpa_types.VerticalPodAutoscaler {
	if b.containerName == "" {
		panic("Must call WithContainer() before Get()")
	}
	resourcePolicy := vpa_types.PodResourcePolicy{ContainerPolicies: []vpa_types.ContainerResourcePolicy{{
		ContainerName: b.containerName,
		MinAllowed:    b.minAllowed,
		MaxAllowed:    b.maxAllowed,
	}}}

	recommendation := b.recommendation.WithContainer(b.containerName).Get()
	for _, rec := range b.appendedRecommendations {
		recommendation.ContainerRecommendations = append(recommendation.ContainerRecommendations, rec)
	}

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
			Recommendation: recommendation,
			Conditions:     b.conditions,
		},
	}
}
