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
	"fmt"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	listers "k8s.io/client-go/listers/core/v1"
)

// LimitRangeCalculator calculates limit range items that has the same effect as all limit range items present in the cluster.
type LimitRangeCalculator interface {
	// GetContainerLimitRangeItem returns LimitRangeItem that describes limitation on container limits in the given namespace.
	GetContainerLimitRangeItem(namespace string) (*core.LimitRangeItem, error)
	// GetPodLimitRangeItem returns LimitRangeItem that describes limitation on pod limits in the given namespace.
	GetPodLimitRangeItem(namespace string) (*core.LimitRangeItem, error)
}

type noopLimitsRangeCalculator struct{}

func (lc *noopLimitsRangeCalculator) GetContainerLimitRangeItem(namespace string) (*core.LimitRangeItem, error) {
	return nil, nil
}

func (lc *noopLimitsRangeCalculator) GetPodLimitRangeItem(namespace string) (*core.LimitRangeItem, error) {
	return nil, nil
}

type limitsChecker struct {
	limitRangeLister listers.LimitRangeLister
}

// NewLimitsRangeCalculator returns a limitsChecker or an error it encountered when attempting to create it.
func NewLimitsRangeCalculator(f informers.SharedInformerFactory) (*limitsChecker, error) {
	if f == nil {
		return nil, fmt.Errorf("NewLimitsRangeCalculator requires a SharedInformerFactory but got nil")
	}
	limitRangeLister := f.Core().V1().LimitRanges().Lister()
	stopCh := make(chan struct{})
	f.Start(stopCh)
	for _, ok := range f.WaitForCacheSync(stopCh) {
		if !ok {
			if !f.Core().V1().LimitRanges().Informer().HasSynced() {
				return nil, fmt.Errorf("informer did not sync")
			}
		}
	}
	return &limitsChecker{limitRangeLister}, nil
}

// NewNoopLimitsCalculator returns a limit calculator that instantly returns no limits.
func NewNoopLimitsCalculator() *noopLimitsRangeCalculator {
	return &noopLimitsRangeCalculator{}
}

func (lc *limitsChecker) GetContainerLimitRangeItem(namespace string) (*core.LimitRangeItem, error) {
	return lc.getLimitRangeItem(namespace, core.LimitTypeContainer)
}

func (lc *limitsChecker) GetPodLimitRangeItem(namespace string) (*core.LimitRangeItem, error) {
	return lc.getLimitRangeItem(namespace, core.LimitTypePod)
}

func (lc *limitsChecker) getLimitRangeItem(namespace string, limitType core.LimitType) (*core.LimitRangeItem, error) {
	limitRanges, err := lc.limitRangeLister.LimitRanges(namespace).List(labels.Everything())
	if err != nil {
		return nil, fmt.Errorf("error loading limit ranges: %s", err)
	}

	updatedResult := func(result core.ResourceList, lrItem core.ResourceList,
		resourceName core.ResourceName, picker func(q1, q2 resource.Quantity) resource.Quantity) core.ResourceList {
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

	result := &core.LimitRangeItem{Type: limitType}
	for _, lr := range limitRanges {
		for _, lri := range lr.Spec.Limits {
			if lri.Type == limitType && (lri.Max != nil || lri.Default != nil || lri.Min != nil) {
				if lri.Default != nil {
					result.Default = lri.Default
				}
				result.Max = updatedResult(result.Max, lri.Max, core.ResourceCPU, pickLowerMax)
				result.Max = updatedResult(result.Max, lri.Max, core.ResourceMemory, pickLowerMax)
				result.Min = updatedResult(result.Min, lri.Min, core.ResourceCPU, chooseHigherMin)
				result.Min = updatedResult(result.Min, lri.Min, core.ResourceMemory, chooseHigherMin)
			}
		}
	}
	if result.Min != nil || result.Max != nil || result.Default != nil {
		return result, nil
	}
	return nil, nil
}
