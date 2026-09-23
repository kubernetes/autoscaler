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

package limitrange

import (
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	listersv1 "k8s.io/client-go/listers/core/v1"
)

// LimitRangeCalculator calculates limit range items that has the same effect as all limit range items present in the cluster.
type LimitRangeCalculator interface {
	// GetContainerLimitRangeItem returns LimitRangeItem that describes limitation on container limits in the given namespace.
	GetContainerLimitRangeItem(namespace string) (*corev1.LimitRangeItem, error)
	// GetPodLimitRangeItem returns LimitRangeItem that describes limitation on pod limits in the given namespace.
	GetPodLimitRangeItem(namespace string) (*corev1.LimitRangeItem, error)
}

type noopLimitsRangeCalculator struct{}

func (*noopLimitsRangeCalculator) GetContainerLimitRangeItem(namespace string) (*corev1.LimitRangeItem, error) {
	return nil, nil
}

func (*noopLimitsRangeCalculator) GetPodLimitRangeItem(namespace string) (*corev1.LimitRangeItem, error) {
	return nil, nil
}

type limitsChecker struct {
	limitRangeLister listersv1.LimitRangeLister
}

// NewLimitsRangeCalculator returns a limitsChecker or an error it encountered when attempting to create it.
func NewLimitsRangeCalculator(f informers.SharedInformerFactory) (*limitsChecker, error) {
	if f == nil {
		return nil, errors.New("NewLimitsRangeCalculator requires a SharedInformerFactory but got nil")
	}
	limitRangeLister := f.Core().V1().LimitRanges().Lister()
	return &limitsChecker{limitRangeLister}, nil
}

// NewNoopLimitsCalculator returns a limit calculator that instantly returns no limits.
func NewNoopLimitsCalculator() *noopLimitsRangeCalculator {
	return &noopLimitsRangeCalculator{}
}

func (lc *limitsChecker) GetContainerLimitRangeItem(namespace string) (*corev1.LimitRangeItem, error) {
	return lc.getLimitRangeItem(namespace, corev1.LimitTypeContainer)
}

func (lc *limitsChecker) GetPodLimitRangeItem(namespace string) (*corev1.LimitRangeItem, error) {
	return lc.getLimitRangeItem(namespace, corev1.LimitTypePod)
}

func (lc *limitsChecker) getLimitRangeItem(namespace string, limitType corev1.LimitType) (*corev1.LimitRangeItem, error) {
	limitRanges, err := lc.limitRangeLister.LimitRanges(namespace).List(labels.Everything())
	if err != nil {
		return nil, fmt.Errorf("error loading limit ranges: %s", err)
	}

	updatedResult := func(result corev1.ResourceList, lrItem corev1.ResourceList,
		resourceName corev1.ResourceName, picker func(q1, q2 resource.Quantity) resource.Quantity) corev1.ResourceList {
		if lrItem == nil {
			return result
		}
		if result == nil {
			return lrItem.DeepCopy()
		}
		if lrResource, lrHas := lrItem[resourceName]; lrHas {
			resultResource, resultHas := result[resourceName]
			if !resultHas {
				result[resourceName] = lrResource.DeepCopy()
			} else {
				result[resourceName] = picker(resultResource, lrResource)
			}
		}
		return result
	}
	pickLowerMax := func(q1, q2 resource.Quantity) resource.Quantity {
		if q1.Cmp(q2) < 0 {
			return q1
		}
		return q2
	}
	chooseHigherMin := func(q1, q2 resource.Quantity) resource.Quantity {
		if q1.Cmp(q2) > 0 {
			return q1
		}
		return q2
	}

	result := &corev1.LimitRangeItem{Type: limitType}
	for _, lr := range limitRanges {
		for _, lri := range lr.Spec.Limits {
			if lri.Type == limitType && (lri.Max != nil || lri.Default != nil || lri.Min != nil) {
				if lri.Default != nil {
					result.Default = lri.Default
				}
				result.Max = updatedResult(result.Max, lri.Max, corev1.ResourceCPU, pickLowerMax)
				result.Max = updatedResult(result.Max, lri.Max, corev1.ResourceMemory, pickLowerMax)
				result.Min = updatedResult(result.Min, lri.Min, corev1.ResourceCPU, chooseHigherMin)
				result.Min = updatedResult(result.Min, lri.Min, corev1.ResourceMemory, chooseHigherMin)
			}
		}
	}
	if result.Min != nil || result.Max != nil || result.Default != nil {
		return result, nil
	}
	return nil, nil
}
