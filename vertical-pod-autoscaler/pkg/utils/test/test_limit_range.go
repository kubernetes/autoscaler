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
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LimitRange returns an object that helps build a LimitRangeItem object for tests.
func LimitRange() *limitRangeBuilder {
	return &limitRangeBuilder{}
}

type limitRangeBuilder struct {
	namespace     string
	name          string
	rangeType     v1.LimitType
	defaultValues *v1.ResourceList
	max           *v1.ResourceList
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

func (lrb *limitRangeBuilder) WithType(rangeType v1.LimitType) *limitRangeBuilder {
	result := *lrb
	result.rangeType = rangeType
	return &result
}

func (lrb *limitRangeBuilder) WithDefault(defaultValues v1.ResourceList) *limitRangeBuilder {
	result := *lrb
	result.defaultValues = &defaultValues
	return &result
}

func (lrb *limitRangeBuilder) WithMax(max v1.ResourceList) *limitRangeBuilder {
	result := *lrb
	result.max = &max
	return &result
}

func (lrb *limitRangeBuilder) Get() *v1.LimitRange {
	result := v1.LimitRange{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: lrb.namespace,
			Name:      lrb.name,
		},
	}
	if lrb.defaultValues != nil || lrb.max != nil {
		result.Spec = v1.LimitRangeSpec{
			Limits: []v1.LimitRangeItem{},
		}
	}
	if lrb.defaultValues != nil {
		result.Spec.Limits = append(result.Spec.Limits, v1.LimitRangeItem{
			Type:    lrb.rangeType,
			Default: *lrb.defaultValues,
		})
	}
	if lrb.max != nil {
		result.Spec.Limits = append(result.Spec.Limits, v1.LimitRangeItem{
			Type: lrb.rangeType,
			Max:  *lrb.max,
		})
	}
	return &result
}
