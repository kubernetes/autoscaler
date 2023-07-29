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
	mpa_types "k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1alpha1"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
)

// MultidimPodAutoscalerBuilder helps building test instances of MultidimPodAutoscaler.
type MultidimPodAutoscalerBuilder interface {
	WithName(vpaName string) MultidimPodAutoscalerBuilder
	WithContainer(containerName string) MultidimPodAutoscalerBuilder
	WithNamespace(namespace string) MultidimPodAutoscalerBuilder
	WithUpdateMode(updateMode vpa_types.UpdateMode) MultidimPodAutoscalerBuilder
	WithCreationTimestamp(timestamp time.Time) MultidimPodAutoscalerBuilder
	WithMinAllowed(cpu, memory string) MultidimPodAutoscalerBuilder
	WithMaxAllowed(cpu, memory string) MultidimPodAutoscalerBuilder
	WithControlledValues(mode vpa_types.ContainerControlledValues) MultidimPodAutoscalerBuilder
	WithTarget(cpu, memory string) MultidimPodAutoscalerBuilder
	WithLowerBound(cpu, memory string) MultidimPodAutoscalerBuilder
	WithScaleTargetRef(targetRef *autoscaling.CrossVersionObjectReference) MultidimPodAutoscalerBuilder
	WithUpperBound(cpu, memory string) MultidimPodAutoscalerBuilder
	WithAnnotations(map[string]string) MultidimPodAutoscalerBuilder
	WithRecommender(string2 string) MultidimPodAutoscalerBuilder
	AppendCondition(conditionType mpa_types.MultidimPodAutoscalerConditionType,
		status core.ConditionStatus, reason, message string, lastTransitionTime time.Time) MultidimPodAutoscalerBuilder
	AppendRecommendation(vpa_types.RecommendedContainerResources) MultidimPodAutoscalerBuilder
	Get() *mpa_types.MultidimPodAutoscaler
}

// MultidimPodAutoscaler returns a new MultidimPodAutoscalerBuilder.
func MultidimPodAutoscaler() MultidimPodAutoscalerBuilder {
	return &multidimPodAutoscalerBuilder{
		recommendation:          Recommendation(),
		appendedRecommendations: []vpa_types.RecommendedContainerResources{},
		namespace:               "default",
		conditions:              []mpa_types.MultidimPodAutoscalerCondition{},
	}
}

type multidimPodAutoscalerBuilder struct {
	mpaName                 string
	containerName           string
	namespace               string
	updatePolicy            *mpa_types.PodUpdatePolicy
	creationTimestamp       time.Time
	minAllowed              core.ResourceList
	maxAllowed              core.ResourceList
	ControlledValues        *vpa_types.ContainerControlledValues
	recommendation          RecommendationBuilder
	conditions              []mpa_types.MultidimPodAutoscalerCondition
	annotations             map[string]string
	scaleTargetRef          *autoscaling.CrossVersionObjectReference
	appendedRecommendations []vpa_types.RecommendedContainerResources
	recommender             string
}

func (b *multidimPodAutoscalerBuilder) WithName(mpaName string) MultidimPodAutoscalerBuilder {
	c := *b
	c.mpaName = mpaName
	return &c
}

func (b *multidimPodAutoscalerBuilder) WithContainer(containerName string) MultidimPodAutoscalerBuilder {
	c := *b
	c.containerName = containerName
	return &c
}

func (b *multidimPodAutoscalerBuilder) WithNamespace(namespace string) MultidimPodAutoscalerBuilder {
	c := *b
	c.namespace = namespace
	return &c
}

func (b *multidimPodAutoscalerBuilder) WithUpdateMode(updateMode vpa_types.UpdateMode) MultidimPodAutoscalerBuilder {
	c := *b
	if c.updatePolicy == nil {
		c.updatePolicy = &mpa_types.PodUpdatePolicy{}
	}
	c.updatePolicy.UpdateMode = &updateMode
	return &c
}

func (b *multidimPodAutoscalerBuilder) WithCreationTimestamp(timestamp time.Time) MultidimPodAutoscalerBuilder {
	c := *b
	c.creationTimestamp = timestamp
	return &c
}

func (b *multidimPodAutoscalerBuilder) WithMinAllowed(cpu, memory string) MultidimPodAutoscalerBuilder {
	c := *b
	c.minAllowed = Resources(cpu, memory)
	return &c
}

func (b *multidimPodAutoscalerBuilder) WithMaxAllowed(cpu, memory string) MultidimPodAutoscalerBuilder {
	c := *b
	c.maxAllowed = Resources(cpu, memory)
	return &c
}

func (b *multidimPodAutoscalerBuilder) WithControlledValues(mode vpa_types.ContainerControlledValues) MultidimPodAutoscalerBuilder {
	c := *b
	c.ControlledValues = &mode
	return &c
}

func (b *multidimPodAutoscalerBuilder) WithTarget(cpu, memory string) MultidimPodAutoscalerBuilder {
	c := *b
	c.recommendation = c.recommendation.WithTarget(cpu, memory)
	return &c
}

func (b *multidimPodAutoscalerBuilder) WithLowerBound(cpu, memory string) MultidimPodAutoscalerBuilder {
	c := *b
	c.recommendation = c.recommendation.WithLowerBound(cpu, memory)
	return &c
}

func (b *multidimPodAutoscalerBuilder) WithUpperBound(cpu, memory string) MultidimPodAutoscalerBuilder {
	c := *b
	c.recommendation = c.recommendation.WithUpperBound(cpu, memory)
	return &c
}

func (b *multidimPodAutoscalerBuilder) WithScaleTargetRef(scaleTargetRef *autoscaling.CrossVersionObjectReference) MultidimPodAutoscalerBuilder {
	c := *b
	c.scaleTargetRef = scaleTargetRef
	return &c
}

func (b *multidimPodAutoscalerBuilder) WithAnnotations(annotations map[string]string) MultidimPodAutoscalerBuilder {
	c := *b
	c.annotations = annotations
	return &c
}

func (b *multidimPodAutoscalerBuilder) WithRecommender(recommender string) MultidimPodAutoscalerBuilder {
	c := *b
	c.recommender = recommender
	return &c
}

func (b *multidimPodAutoscalerBuilder) AppendCondition(conditionType mpa_types.MultidimPodAutoscalerConditionType,
	status core.ConditionStatus, reason, message string, lastTransitionTime time.Time) MultidimPodAutoscalerBuilder {
	c := *b
	c.conditions = append(c.conditions, mpa_types.MultidimPodAutoscalerCondition{
		Type:               conditionType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: meta.NewTime(lastTransitionTime)})
	return &c
}

func (b *multidimPodAutoscalerBuilder) AppendRecommendation(recommendation vpa_types.RecommendedContainerResources) MultidimPodAutoscalerBuilder {
	c := *b
	c.appendedRecommendations = append(c.appendedRecommendations, recommendation)
	return &c
}

func (b *multidimPodAutoscalerBuilder) Get() *mpa_types.MultidimPodAutoscaler {
	if b.containerName == "" {
		panic("Must call WithContainer() before Get()")
	}
	var recommenders []*mpa_types.MultidimPodAutoscalerRecommenderSelector
	if b.recommender != "" {
		recommenders = []*mpa_types.MultidimPodAutoscalerRecommenderSelector{{Name: b.recommender}}
	}
	resourcePolicy := vpa_types.PodResourcePolicy{ContainerPolicies: []vpa_types.ContainerResourcePolicy{{
		ContainerName:    b.containerName,
		MinAllowed:       b.minAllowed,
		MaxAllowed:       b.maxAllowed,
		ControlledValues: b.ControlledValues,
	}}}

	recommendation := b.recommendation.WithContainer(b.containerName).Get()
	recommendation.ContainerRecommendations = append(recommendation.ContainerRecommendations, b.appendedRecommendations...)

	return &mpa_types.MultidimPodAutoscaler{
		ObjectMeta: meta.ObjectMeta{
			Name:              b.mpaName,
			Namespace:         b.namespace,
			Annotations:       b.annotations,
			CreationTimestamp: meta.NewTime(b.creationTimestamp),
		},
		Spec: mpa_types.MultidimPodAutoscalerSpec{
			UpdatePolicy:   b.updatePolicy,
			ResourcePolicy: &resourcePolicy,
			ScaleTargetRef: b.scaleTargetRef,
			Recommenders:   recommenders,
		},
		Status: mpa_types.MultidimPodAutoscalerStatus{
			Recommendation: recommendation,
			Conditions:     b.conditions,
		},
	}
}
