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
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LimitRange returns an object that helps build a LimitRangeItem object for tests.
func LimitRange() *limitRangeBuilder {
	return &limitRangeBuilder{}
}

type limitRangeBuilder struct {
	namespace     string
	name          string
	rangeType     core.LimitType
	defaultValues []*core.ResourceList
	maxValues     []*core.ResourceList
	minValues     []*core.ResourceList
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

func (lrb *limitRangeBuilder) WithType(rangeType core.LimitType) *limitRangeBuilder {
	result := *lrb
	result.rangeType = rangeType
	return &result
}

func (lrb *limitRangeBuilder) WithDefault(defaultValues core.ResourceList) *limitRangeBuilder {
	result := *lrb
	result.defaultValues = append(result.defaultValues, &defaultValues)
	return &result
}

func (lrb *limitRangeBuilder) WithMax(max core.ResourceList) *limitRangeBuilder {
	result := *lrb
	result.maxValues = append(result.maxValues, &max)
	return &result
}

func (lrb *limitRangeBuilder) WithMin(min core.ResourceList) *limitRangeBuilder {
	result := *lrb
	result.minValues = append(result.minValues, &min)
	return &result
}

func (lrb *limitRangeBuilder) Get() *core.LimitRange {
	result := core.LimitRange{
		ObjectMeta: meta.ObjectMeta{
			Namespace: lrb.namespace,
			Name:      lrb.name,
		},
	}
	if len(lrb.defaultValues) > 0 || len(lrb.maxValues) > 0 || len(lrb.minValues) > 0 {
		result.Spec = core.LimitRangeSpec{
			Limits: []core.LimitRangeItem{},
		}
	}
	for _, v := range lrb.defaultValues {
		result.Spec.Limits = append(result.Spec.Limits, core.LimitRangeItem{
			Type:    lrb.rangeType,
			Default: *v,
		})
	}
	for _, v := range lrb.maxValues {
		result.Spec.Limits = append(result.Spec.Limits, core.LimitRangeItem{
			Type: lrb.rangeType,
			Max:  *v,
		})
	}
	for _, v := range lrb.minValues {
		result.Spec.Limits = append(result.Spec.Limits, core.LimitRangeItem{
			Type: lrb.rangeType,
			Min:  *v,
		})
	}
	return &result
}
