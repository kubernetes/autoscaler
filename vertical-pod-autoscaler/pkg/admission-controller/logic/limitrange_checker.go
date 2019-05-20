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

package logic

import (
	"fmt"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"

	v1_listers "k8s.io/client-go/listers/core/v1"
)

// LimitsHints provides hinted limits that respect limit range max ratio
type LimitsHints interface {
	RequestsExceedsRatio(indexOfContainer int, resourceName v1.ResourceName) bool
	HintedLimit(indexOfContainer int, resourceName v1.ResourceName) resource.Quantity
}

// LimitRangeHints implements LimitsHints interface
type LimitRangeHints struct {
	limitsRespectingRatio []v1.ResourceList
}

var _ LimitsHints = &LimitRangeHints{}

// LimitsChecker checks for LimitRange and if container needs limits to be set
type LimitsChecker interface {
	NeedsLimits(*v1.Pod, []ContainerResources) (LimitsHints, error)
}

// RequestsExceedsRatio return true if limits have to be set to respect limit range max ratio
func (lrh *LimitRangeHints) RequestsExceedsRatio(indexOfContainer int, resourceName v1.ResourceName) bool {
	if lrh != nil && indexOfContainer < len(lrh.limitsRespectingRatio) {
		_, present := lrh.limitsRespectingRatio[indexOfContainer][resourceName]
		return present
	}
	return false
}

// HintedLimit return the limit Quantity that respect the limit range max ration
func (lrh *LimitRangeHints) HintedLimit(indexOfContainer int, resourceName v1.ResourceName) resource.Quantity {
	if lrh != nil && indexOfContainer < len(lrh.limitsRespectingRatio) {
		limit, ok := lrh.limitsRespectingRatio[indexOfContainer][resourceName]
		if ok {
			return limit
		}
		return resource.Quantity{}
	}
	return resource.Quantity{}
}

type neverNeedsLimitsChecker struct{}

var _ LimitsChecker = &neverNeedsLimitsChecker{}

func (lc *neverNeedsLimitsChecker) NeedsLimits(pod *v1.Pod, containersResources []ContainerResources) (LimitsHints, error) {
	return LimitsHints((*LimitRangeHints)(nil)), nil
}

type limitsChecker struct {
	limitrangeLister v1_listers.LimitRangeLister
}

var _ LimitsChecker = &limitsChecker{}

// NewLimitsChecker returns a limitsChecker or an error it encountered when attempting to create it.
func NewLimitsChecker(f informers.SharedInformerFactory) (*limitsChecker, error) {
	if f == nil {
		return nil, fmt.Errorf("NewLimitsChecker requires a SharedInformerFactory but got nil")
	}
	limitrangeLister := f.Core().V1().LimitRanges().Lister()
	stopCh := make(chan struct{})
	f.Start(stopCh)
	for _, ok := range f.WaitForCacheSync(stopCh) {
		if !ok {
			if f.Core().V1().LimitRanges().Informer().HasSynced() {
				return nil, fmt.Errorf("Informer did not sync")
			}
		}
	}
	return &limitsChecker{limitrangeLister}, nil
}

// NewNoopLimitsChecker returns a limit checker that
func NewNoopLimitsChecker() *neverNeedsLimitsChecker {
	return &neverNeedsLimitsChecker{}
}

type interestingData struct {
	MaxLimitRequestRatio v1.ResourceList
	Default              v1.ResourceList
}

func (id *interestingData) parse(lri *v1.LimitRangeItem) {
	if value, hasCPU := lri.MaxLimitRequestRatio[v1.ResourceCPU]; hasCPU {
		if id.MaxLimitRequestRatio == nil {
			id.MaxLimitRequestRatio = make(v1.ResourceList)
		}
		if maxRatio, exists := id.MaxLimitRequestRatio[v1.ResourceCPU]; !exists || maxRatio.Cmp(value) > 0 {
			id.MaxLimitRequestRatio[v1.ResourceCPU] = *value.Copy()
		}
	}
	if value, hasMemory := lri.MaxLimitRequestRatio[v1.ResourceMemory]; hasMemory {
		if id.MaxLimitRequestRatio == nil {
			id.MaxLimitRequestRatio = make(v1.ResourceList)
		}
		if maxRatio, exists := id.MaxLimitRequestRatio[v1.ResourceMemory]; !exists || maxRatio.Cmp(value) > 0 {
			id.MaxLimitRequestRatio[v1.ResourceMemory] = *value.Copy()
		}
	}
	if value, hasCPU := lri.Default[v1.ResourceCPU]; hasCPU {
		if id.Default == nil {
			id.Default = make(v1.ResourceList)
		}
		if _, exists := id.Default[v1.ResourceCPU]; !exists {
			id.Default[v1.ResourceCPU] = *value.Copy()
		}
	}
	if value, hasMemory := lri.Default[v1.ResourceMemory]; hasMemory {
		if id.Default == nil {
			id.Default = make(v1.ResourceList)
		}
		if _, exists := id.Default[v1.ResourceMemory]; !exists {
			id.Default[v1.ResourceMemory] = *value.Copy()
		}
	}
}

func (lc *limitsChecker) getLimitRangeItem(pod *v1.Pod) (*v1.LimitRangeItem, error) {
	limitranges, err := lc.limitrangeLister.
		LimitRanges(pod.GetNamespace()).
		List(labels.Everything())

	if err != nil {
		return nil, fmt.Errorf("error loading limit ranges: %s", err)
	}

	id := &interestingData{}
	foundInterstingData := false
	for _, lr := range limitranges {
		for _, lri := range lr.Spec.Limits {
			if lri.Type != v1.LimitTypeContainer && lri.Type != v1.LimitTypePod {
				continue
			}
			if lri.MaxLimitRequestRatio == nil &&
				lri.Default == nil {
				continue
			}
			// TODO: handle multiple limit ranges matching a pod.
			foundInterstingData = true
			id.parse(&lri)
		}
	}
	if foundInterstingData {
		return &v1.LimitRangeItem{
			MaxLimitRequestRatio: id.MaxLimitRequestRatio,
			Default:              id.Default,
		}, nil
	}
	return nil, nil
}

func (lc *limitsChecker) NeedsLimits(pod *v1.Pod, containersResources []ContainerResources) (LimitsHints, error) {
	lri, err := lc.getLimitRangeItem(pod)
	if err != nil {
		return nil, fmt.Errorf("error getting limit range for pod: %s", err)
	}

	if lri == (*v1.LimitRangeItem)(nil) {
		return &LimitRangeHints{}, nil
	}

	lrh := &LimitRangeHints{
		limitsRespectingRatio: make([]v1.ResourceList, len(containersResources)),
	}
	needsLimits := false

	for i, cr := range containersResources {
		lrh.limitsRespectingRatio[i] = make(v1.ResourceList)
		for name, value := range cr.Requests {
			var ctrLimit *resource.Quantity
			if pod.Spec.Containers[i].Resources.Limits != nil {
				if q, hasLimit := pod.Spec.Containers[i].Resources.Limits[name]; hasLimit {
					ctrLimit = &q
				}
			}
			if q, hasDefault := lri.Default[name]; hasDefault && ctrLimit == nil {
				ctrLimit = &q
			}
			if ctrLimit == nil {
				// no limits for this container, neither default will be set
				continue
			}

			if ratio, hasRatio := lri.MaxLimitRequestRatio[name]; hasRatio {
				dl := *ctrLimit
				dlv := dl.Value()
				vv := value.Value()
				useMilli := false
				if dlv <= resource.MaxMilliValue &&
					vv <= resource.MaxMilliValue &&
					name == v1.ResourceCPU {
					dlv = dl.MilliValue()
					vv = value.MilliValue()
					useMilli = true
				}

				futureRatio := float64(dlv) / float64(vv)
				maxRatio := float64(ratio.Value())

				if futureRatio > maxRatio {
					needsLimits = true
					l := int64(float64(vv) * maxRatio)
					if useMilli {
						if l > resource.MaxMilliValue {
							l = resource.MaxMilliValue
						}
						lrh.limitsRespectingRatio[i][name] = *resource.NewMilliQuantity(l, value.Format)
					} else {
						lrh.limitsRespectingRatio[i][name] = *resource.NewQuantity(l, value.Format)
					}
				}
			}
		}
	}

	if !needsLimits {
		lrh = nil
	}
	return LimitsHints(lrh), nil
}
