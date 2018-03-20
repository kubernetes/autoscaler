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

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/poc.autoscaling.k8s.io/v1alpha1"
)

// VerticalPodAutoscalerBuilder helps building test instances of VerticalPodAutoscaler.
type VerticalPodAutoscalerBuilder interface {
	WithContainer(containerName string) VerticalPodAutoscalerBuilder
	WithNamespace(namespace string) VerticalPodAutoscalerBuilder
	WithSelector(labelSelector string) VerticalPodAutoscalerBuilder
	WithUpdateMode(updateMode vpa_types.UpdateMode) VerticalPodAutoscalerBuilder
	WithCreationTimestamp(timestamp time.Time) VerticalPodAutoscalerBuilder
	WithMinAllowed(cpu, memory string) VerticalPodAutoscalerBuilder
	WithMaxAllowed(cpu, memory string) VerticalPodAutoscalerBuilder
	WithTarget(cpu, memory string) VerticalPodAutoscalerBuilder
	WithMinRecommended(cpu, memory string) VerticalPodAutoscalerBuilder
	WithMaxRecommended(cpu, memory string) VerticalPodAutoscalerBuilder
	Get() *vpa_types.VerticalPodAutoscaler
}

// VerticalPodAutoscaler returns a new VerticalPodAutoscalerBuilder.
func VerticalPodAutoscaler() VerticalPodAutoscalerBuilder {
	return &verticalPodAutoscalerBuilder{
		recommendation: Recommendation(),
		namespace:      "default",
	}
}

type verticalPodAutoscalerBuilder struct {
	containerName     string
	namespace         string
	labelSelector     *metav1.LabelSelector
	updatePolicy      vpa_types.PodUpdatePolicy
	creationTimestamp time.Time
	minAllowed        apiv1.ResourceList
	maxAllowed        apiv1.ResourceList
	recommendation    RecommendationBuilder
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

func (b *verticalPodAutoscalerBuilder) WithSelector(labelSelector string) VerticalPodAutoscalerBuilder {
	c := *b
	if labelSelector, err := metav1.ParseToLabelSelector(labelSelector); err != nil {
		panic(err)
	} else {
		c.labelSelector = labelSelector
	}
	return &c
}

func (b *verticalPodAutoscalerBuilder) WithUpdateMode(updateMode vpa_types.UpdateMode) VerticalPodAutoscalerBuilder {
	c := *b
	c.updatePolicy.UpdateMode = updateMode
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

func (b *verticalPodAutoscalerBuilder) WithMinRecommended(cpu, memory string) VerticalPodAutoscalerBuilder {
	c := *b
	c.recommendation = c.recommendation.WithMinRecommended(cpu, memory)
	return &c
}

func (b *verticalPodAutoscalerBuilder) WithMaxRecommended(cpu, memory string) VerticalPodAutoscalerBuilder {
	c := *b
	c.recommendation = c.recommendation.WithMaxRecommended(cpu, memory)
	return &c
}

func (b *verticalPodAutoscalerBuilder) Get() *vpa_types.VerticalPodAutoscaler {
	if b.containerName == "" {
		panic("Must call WithContainer() before Get()")
	}
	resourcePolicy := vpa_types.PodResourcePolicy{ContainerPolicies: []vpa_types.ContainerResourcePolicy{{
		Name:       b.containerName,
		MinAllowed: b.minAllowed,
		MaxAllowed: b.maxAllowed,
	}}}

	return &vpa_types.VerticalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         b.namespace,
			CreationTimestamp: metav1.NewTime(b.creationTimestamp),
		},
		Spec: vpa_types.VerticalPodAutoscalerSpec{
			Selector:       b.labelSelector,
			UpdatePolicy:   b.updatePolicy,
			ResourcePolicy: resourcePolicy,
		},
		Status: vpa_types.VerticalPodAutoscalerStatus{
			Recommendation: *b.recommendation.WithContainer(b.containerName).Get(),
		},
	}
}
