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
	WithMinAllowed(containerName, cpu, memory string) MultidimPodAutoscalerBuilder
	WithMaxAllowed(containerName, cpu, memory string) MultidimPodAutoscalerBuilder
	WithControlledValues(containerName string, mode vpa_types.ContainerControlledValues) MultidimPodAutoscalerBuilder
	WithScalingMode(containerName string, scalingMode vpa_types.ContainerScalingMode) MultidimPodAutoscalerBuilder
	WithTarget(cpu, memory string) MultidimPodAutoscalerBuilder
	WithTargetResource(resource core.ResourceName, value string) MultidimPodAutoscalerBuilder
	WithLowerBound(cpu, memory string) MultidimPodAutoscalerBuilder
	WithScaleTargetRef(targetRef *autoscaling.CrossVersionObjectReference) MultidimPodAutoscalerBuilder
	WithUpperBound(cpu, memory string) MultidimPodAutoscalerBuilder
	WithAnnotations(map[string]string) MultidimPodAutoscalerBuilder
	WithRecommender(string2 string) MultidimPodAutoscalerBuilder
	WithGroupVersion(gv meta.GroupVersion) MultidimPodAutoscalerBuilder
	// WithEvictionRequirements([]*vpa_types.EvictionRequirement) MultidimPodAutoscalerBuilder // TODO: add this functionality
	WithScalingConstraints(constraints mpa_types.ScalingConstraints) MultidimPodAutoscalerBuilder
	WithHorizontalScalingConstraints(hconstraints mpa_types.HorizontalScalingConstraints) MultidimPodAutoscalerBuilder
	AppendCondition(conditionType mpa_types.MultidimPodAutoscalerConditionType,
		status core.ConditionStatus, reason, message string, lastTransitionTime time.Time) MultidimPodAutoscalerBuilder
	AppendRecommendation(vpa_types.RecommendedContainerResources) MultidimPodAutoscalerBuilder
	Get() *mpa_types.MultidimPodAutoscaler
}

// MultidimPodAutoscaler returns a new MultidimPodAutoscalerBuilder.
func MultidimPodAutoscaler() MultidimPodAutoscalerBuilder {
	return &multidimPodAutoscalerBuilder{
		groupVersion:            meta.GroupVersion(mpa_types.SchemeGroupVersion),
		recommendation:          Recommendation(),
		appendedRecommendations: []vpa_types.RecommendedContainerResources{},
		namespace:               "default",
		conditions:              []mpa_types.MultidimPodAutoscalerCondition{},
		minAllowed:              map[string]core.ResourceList{},
		maxAllowed:              map[string]core.ResourceList{},
		controlledValues:        map[string]*vpa_types.ContainerControlledValues{},
		scalingMode:             map[string]*vpa_types.ContainerScalingMode{},
		scalingConstraints:      mpa_types.ScalingConstraints{},
	}
}

type multidimPodAutoscalerBuilder struct {
	groupVersion            meta.GroupVersion
	mpaName                 string
	containerNames          []string
	namespace               string
	updatePolicy            *mpa_types.PodUpdatePolicy
	creationTimestamp       time.Time
	minAllowed              map[string]core.ResourceList
	maxAllowed              map[string]core.ResourceList
	controlledValues        map[string]*vpa_types.ContainerControlledValues
	scalingMode             map[string]*vpa_types.ContainerScalingMode
	recommendation          RecommendationBuilder
	conditions              []mpa_types.MultidimPodAutoscalerCondition
	annotations             map[string]string
	scaleTargetRef          *autoscaling.CrossVersionObjectReference
	appendedRecommendations []vpa_types.RecommendedContainerResources
	recommender             string
	scalingConstraints      mpa_types.ScalingConstraints
}

func (b *multidimPodAutoscalerBuilder) WithName(mpaName string) MultidimPodAutoscalerBuilder {
	c := *b
	c.mpaName = mpaName
	return &c
}

func (b *multidimPodAutoscalerBuilder) WithContainer(containerName string) MultidimPodAutoscalerBuilder {
	c := *b
	c.containerNames = append(c.containerNames, containerName)
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

func (b *multidimPodAutoscalerBuilder) WithMinAllowed(containerName, cpu, memory string) MultidimPodAutoscalerBuilder {
	c := *b
	c.minAllowed[containerName] = Resources(cpu, memory)
	return &c
}

func (b *multidimPodAutoscalerBuilder) WithMaxAllowed(containerName, cpu, memory string) MultidimPodAutoscalerBuilder {
	c := *b
	c.maxAllowed[containerName] = Resources(cpu, memory)
	return &c
}

func (b *multidimPodAutoscalerBuilder) WithControlledValues(containerName string, mode vpa_types.ContainerControlledValues) MultidimPodAutoscalerBuilder {
	c := *b
	c.controlledValues[containerName] = &mode
	return &c
}

func (b *multidimPodAutoscalerBuilder) WithScalingMode(containerName string, scalingMode vpa_types.ContainerScalingMode) MultidimPodAutoscalerBuilder {
	c := *b
	c.scalingMode[containerName] = &scalingMode
	return &c
}

func (b *multidimPodAutoscalerBuilder) WithTarget(cpu, memory string) MultidimPodAutoscalerBuilder {
	c := *b
	c.recommendation = c.recommendation.WithTarget(cpu, memory)
	return &c
}

func (b *multidimPodAutoscalerBuilder) WithTargetResource(resource core.ResourceName, value string) MultidimPodAutoscalerBuilder {
	c := *b
	c.recommendation = c.recommendation.WithTargetResource(resource, value)
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

func (b *multidimPodAutoscalerBuilder) WithGroupVersion(gv meta.GroupVersion) MultidimPodAutoscalerBuilder {
	c := *b
	c.groupVersion = gv
	return &c
}

func (b *multidimPodAutoscalerBuilder) WithScalingConstraints(constraints mpa_types.ScalingConstraints) MultidimPodAutoscalerBuilder {
	c := *b
	c.scalingConstraints = constraints
	return &c
}

func (b *multidimPodAutoscalerBuilder) WithHorizontalScalingConstraints(constraints mpa_types.HorizontalScalingConstraints) MultidimPodAutoscalerBuilder {
	c := *b
	c.scalingConstraints.Global = &constraints
	return &c
}

// TODO: add this functionality
// func (b *multidimPodAutoscalerBuilder) WithEvictionRequirements(evictionRequirements []*vpa_types.EvictionRequirement) MultidimPodAutoscalerBuilder {
// 	updateModeAuto := vpa_types.UpdateModeAuto
// 	c := *b
// 	if c.updatePolicy == nil {
// 		c.updatePolicy = &mpa_types.PodUpdatePolicy{UpdateMode: &updateModeAuto}
// 	}
// 	c.updatePolicy.EvictionRequirements = evictionRequirements
// 	return &c
// }

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
	if len(b.containerNames) == 0 {
		panic("Must call WithContainer() before Get()")
	}
	var recommenders []*mpa_types.MultidimPodAutoscalerRecommenderSelector
	if b.recommender != "" {
		recommenders = []*mpa_types.MultidimPodAutoscalerRecommenderSelector{{Name: b.recommender}}
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
	// MPAs with a single container may still use the old/implicit way of adding recommendations
	r := b.recommendation.WithContainer(b.containerNames[0]).Get()
	if r.ContainerRecommendations[0].Target != nil {
		recommendation = r
	}
	recommendation.ContainerRecommendations = append(recommendation.ContainerRecommendations, b.appendedRecommendations...)

	return &mpa_types.MultidimPodAutoscaler{
		ObjectMeta: meta.ObjectMeta{
			Name:              b.mpaName,
			Namespace:         b.namespace,
			Annotations:       b.annotations,
			CreationTimestamp: meta.NewTime(b.creationTimestamp),
		},
		Spec: mpa_types.MultidimPodAutoscalerSpec{
			Policy:         b.updatePolicy,
			ResourcePolicy: &resourcePolicy,
			ScaleTargetRef: b.scaleTargetRef,
			Recommenders:   recommenders,
			Constraints:    &b.scalingConstraints,
		},
		Status: mpa_types.MultidimPodAutoscalerStatus{
			Recommendation: recommendation,
			Conditions:     b.conditions,
		},
	}
}
