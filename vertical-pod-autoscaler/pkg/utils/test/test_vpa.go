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
	WithMinAllowed(containerName, cpu, memory string) VerticalPodAutoscalerBuilder
	WithMaxAllowed(containerName, cpu, memory string) VerticalPodAutoscalerBuilder
	WithControlledValues(containerName string, mode vpa_types.ContainerControlledValues) VerticalPodAutoscalerBuilder
	WithScalingMode(containerName string, scalingMode vpa_types.ContainerScalingMode) VerticalPodAutoscalerBuilder
	WithTarget(cpu, memory string) VerticalPodAutoscalerBuilder
	WithTargetResource(resource core.ResourceName, value string) VerticalPodAutoscalerBuilder
	WithLowerBound(cpu, memory string) VerticalPodAutoscalerBuilder
	WithTargetRef(targetRef *autoscaling.CrossVersionObjectReference) VerticalPodAutoscalerBuilder
	WithPodLabelSelector(podLabelSelector *meta.LabelSelector) VerticalPodAutoscalerBuilder
	WithUpperBound(cpu, memory string) VerticalPodAutoscalerBuilder
	WithAnnotations(map[string]string) VerticalPodAutoscalerBuilder
	WithRecommender(string2 string) VerticalPodAutoscalerBuilder
	WithGroupVersion(gv meta.GroupVersion) VerticalPodAutoscalerBuilder
	WithEvictionRequirements([]*vpa_types.EvictionRequirement) VerticalPodAutoscalerBuilder
	WithMinReplicas(minReplicas *int32) VerticalPodAutoscalerBuilder
	AppendCondition(conditionType vpa_types.VerticalPodAutoscalerConditionType,
		status core.ConditionStatus, reason, message string, lastTransitionTime time.Time) VerticalPodAutoscalerBuilder
	AppendRecommendation(vpa_types.RecommendedContainerResources) VerticalPodAutoscalerBuilder
	Get() *vpa_types.VerticalPodAutoscaler
}

// TODO part of this interface is a repetition of RecommendationBuilder, we can probably factorize some code

// VerticalPodAutoscaler returns a new VerticalPodAutoscalerBuilder.
func VerticalPodAutoscaler() VerticalPodAutoscalerBuilder {
	return &verticalPodAutoscalerBuilder{
		groupVersion:            meta.GroupVersion(vpa_types.SchemeGroupVersion),
		recommendation:          Recommendation(),
		appendedRecommendations: []vpa_types.RecommendedContainerResources{},
		namespace:               "default",
		conditions:              []vpa_types.VerticalPodAutoscalerCondition{},
		minAllowed:              map[string]core.ResourceList{},
		maxAllowed:              map[string]core.ResourceList{},
		controlledValues:        map[string]*vpa_types.ContainerControlledValues{},
		scalingMode:             map[string]*vpa_types.ContainerScalingMode{},
	}
}

type verticalPodAutoscalerBuilder struct {
	groupVersion            meta.GroupVersion
	vpaName                 string
	containerNames          []string
	namespace               string
	updatePolicy            *vpa_types.PodUpdatePolicy
	creationTimestamp       time.Time
	minAllowed              map[string]core.ResourceList
	maxAllowed              map[string]core.ResourceList
	controlledValues        map[string]*vpa_types.ContainerControlledValues
	scalingMode             map[string]*vpa_types.ContainerScalingMode
	recommendation          RecommendationBuilder
	conditions              []vpa_types.VerticalPodAutoscalerCondition
	annotations             map[string]string
	targetRef               *autoscaling.CrossVersionObjectReference
	podLabelSelector        *meta.LabelSelector
	appendedRecommendations []vpa_types.RecommendedContainerResources
	recommender             string
}

func (b *verticalPodAutoscalerBuilder) WithName(vpaName string) VerticalPodAutoscalerBuilder {
	c := *b
	c.vpaName = vpaName
	return &c
}

func (b *verticalPodAutoscalerBuilder) WithContainer(containerName string) VerticalPodAutoscalerBuilder {
	c := *b
	c.containerNames = append(c.containerNames, containerName)
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

func (b *verticalPodAutoscalerBuilder) WithMinAllowed(containerName, cpu, memory string) VerticalPodAutoscalerBuilder {
	c := *b
	c.minAllowed[containerName] = Resources(cpu, memory)
	return &c
}

func (b *verticalPodAutoscalerBuilder) WithMaxAllowed(containerName, cpu, memory string) VerticalPodAutoscalerBuilder {
	c := *b
	c.maxAllowed[containerName] = Resources(cpu, memory)
	return &c
}

func (b *verticalPodAutoscalerBuilder) WithControlledValues(containerName string, mode vpa_types.ContainerControlledValues) VerticalPodAutoscalerBuilder {
	c := *b
	c.controlledValues[containerName] = &mode
	return &c
}

func (b *verticalPodAutoscalerBuilder) WithScalingMode(containerName string, scalingMode vpa_types.ContainerScalingMode) VerticalPodAutoscalerBuilder {
	c := *b
	c.scalingMode[containerName] = &scalingMode
	return &c
}

func (b *verticalPodAutoscalerBuilder) WithTarget(cpu, memory string) VerticalPodAutoscalerBuilder {
	c := *b
	c.recommendation = c.recommendation.WithTarget(cpu, memory)
	return &c
}

func (b *verticalPodAutoscalerBuilder) WithTargetResource(resource core.ResourceName, value string) VerticalPodAutoscalerBuilder {
	c := *b
	c.recommendation = c.recommendation.WithTargetResource(resource, value)
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

func (b *verticalPodAutoscalerBuilder) WithPodLabelSelector(podLabelSelector *meta.LabelSelector) VerticalPodAutoscalerBuilder {
	c := *b

	c.podLabelSelector = podLabelSelector
	return &c
}

func (b *verticalPodAutoscalerBuilder) WithAnnotations(annotations map[string]string) VerticalPodAutoscalerBuilder {
	c := *b
	c.annotations = annotations
	return &c
}

func (b *verticalPodAutoscalerBuilder) WithRecommender(recommender string) VerticalPodAutoscalerBuilder {
	c := *b
	c.recommender = recommender
	return &c
}

func (b *verticalPodAutoscalerBuilder) WithGroupVersion(gv meta.GroupVersion) VerticalPodAutoscalerBuilder {
	c := *b
	c.groupVersion = gv
	return &c
}

func (b *verticalPodAutoscalerBuilder) WithEvictionRequirements(evictionRequirements []*vpa_types.EvictionRequirement) VerticalPodAutoscalerBuilder {
	updateModeAuto := vpa_types.UpdateModeAuto
	c := *b
	if c.updatePolicy == nil {
		c.updatePolicy = &vpa_types.PodUpdatePolicy{UpdateMode: &updateModeAuto}
	}
	c.updatePolicy.EvictionRequirements = evictionRequirements
	return &c
}

func (b *verticalPodAutoscalerBuilder) WithMinReplicas(minReplicas *int32) VerticalPodAutoscalerBuilder {
	updateModeAuto := vpa_types.UpdateModeAuto
	c := *b
	if c.updatePolicy == nil {
		c.updatePolicy = &vpa_types.PodUpdatePolicy{UpdateMode: &updateModeAuto}
	}
	c.updatePolicy.MinReplicas = minReplicas
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
	if len(b.containerNames) == 0 {
		panic("Must call WithContainer() before Get()")
	}
	var recommenders []*vpa_types.VerticalPodAutoscalerRecommenderSelector
	if b.recommender != "" {
		recommenders = []*vpa_types.VerticalPodAutoscalerRecommenderSelector{{Name: b.recommender}}
	}
	resourcePolicy := vpa_types.PodResourcePolicy{}
	recommendation := &vpa_types.RecommendedPodResources{}
	scalingModeAuto := vpa_types.ContainerScalingModeAuto
	for _, containerName := range b.containerNames {
		containerResourcePolicy := vpa_types.ContainerResourcePolicy{
			ContainerName:    containerName,
			MinAllowed:       b.minAllowed[containerName],
			MaxAllowed:       b.maxAllowed[containerName],
			ControlledValues: b.controlledValues[containerName],
			Mode:             &scalingModeAuto,
		}
		if scalingMode, ok := b.scalingMode[containerName]; ok {
			containerResourcePolicy.Mode = scalingMode
		}
		resourcePolicy.ContainerPolicies = append(resourcePolicy.ContainerPolicies, containerResourcePolicy)
	}
	// VPAs with a single container may still use the old/implicit way of adding recommendations
	r := b.recommendation.WithContainer(b.containerNames[0]).Get()
	if r.ContainerRecommendations[0].Target != nil {
		recommendation = r
	}

	recommendation.ContainerRecommendations = append(recommendation.ContainerRecommendations, b.appendedRecommendations...)

	return &vpa_types.VerticalPodAutoscaler{
		TypeMeta: meta.TypeMeta{
			APIVersion: b.groupVersion.String(),
			Kind:       "VerticalPodAutoscaler",
		},
		ObjectMeta: meta.ObjectMeta{
			Name:              b.vpaName,
			Namespace:         b.namespace,
			Annotations:       b.annotations,
			CreationTimestamp: meta.NewTime(b.creationTimestamp),
		},
		Spec: vpa_types.VerticalPodAutoscalerSpec{
			UpdatePolicy:     b.updatePolicy,
			ResourcePolicy:   &resourcePolicy,
			TargetRef:        b.targetRef,
			PodLabelSelector: b.podLabelSelector,
			Recommenders:     recommenders,
		},
		Status: vpa_types.VerticalPodAutoscalerStatus{
			Recommendation: recommendation,
			Conditions:     b.conditions,
		},
	}
}
