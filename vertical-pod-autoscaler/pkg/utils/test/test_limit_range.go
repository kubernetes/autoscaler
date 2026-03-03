/*
Copyright 2019 The Kubernetes Authors.

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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LimitRange returns an object that helps build a LimitRangeItem object for tests.
func LimitRange() *limitRangeBuilder {
	return &limitRangeBuilder{}
}

type limitRangeBuilder struct {
	namespace     string
	name          string
	rangeType     corev1.LimitType
	defaultValues []*corev1.ResourceList
	maxValues     []*corev1.ResourceList
	minValues     []*corev1.ResourceList
}

func (lrb *limitRangeBuilder) WithName(name string) *limitRangeBuilder {
	result := *lrb
	result.name = name
	return &result
}

func (lrb *limitRangeBuilder) WithNamespace(namespace string) *limitRangeBuilder {
	result := *lrb
	result.namespace = namespace
	return &result
}

func (lrb *limitRangeBuilder) WithType(rangeType corev1.LimitType) *limitRangeBuilder {
	result := *lrb
	result.rangeType = rangeType
	return &result
}

func (lrb *limitRangeBuilder) WithDefault(defaultValues corev1.ResourceList) *limitRangeBuilder {
	result := *lrb
	result.defaultValues = append(result.defaultValues, &defaultValues)
	return &result
}

func (lrb *limitRangeBuilder) WithMax(maxValues corev1.ResourceList) *limitRangeBuilder {
	result := *lrb
	result.maxValues = append(result.maxValues, &maxValues)
	return &result
}

func (lrb *limitRangeBuilder) WithMin(minValues corev1.ResourceList) *limitRangeBuilder {
	result := *lrb
	result.minValues = append(result.minValues, &minValues)
	return &result
}

func (lrb *limitRangeBuilder) Get() *corev1.LimitRange {
	result := corev1.LimitRange{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: lrb.namespace,
			Name:      lrb.name,
		},
	}
	if len(lrb.defaultValues) > 0 || len(lrb.maxValues) > 0 || len(lrb.minValues) > 0 {
		result.Spec = corev1.LimitRangeSpec{
			Limits: []corev1.LimitRangeItem{},
		}
	}
	for _, v := range lrb.defaultValues {
		result.Spec.Limits = append(result.Spec.Limits, corev1.LimitRangeItem{
			Type:    lrb.rangeType,
			Default: *v,
		})
	}
	for _, v := range lrb.maxValues {
		result.Spec.Limits = append(result.Spec.Limits, corev1.LimitRangeItem{
			Type: lrb.rangeType,
			Max:  *v,
		})
	}
	for _, v := range lrb.minValues {
		result.Spec.Limits = append(result.Spec.Limits, corev1.LimitRangeItem{
			Type: lrb.rangeType,
			Min:  *v,
		})
	}
	return &result
}
