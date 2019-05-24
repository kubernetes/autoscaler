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
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"

	v1_listers "k8s.io/client-go/listers/core/v1"
)

// LimitsRangeCalculator checks for LimitRange and if container needs limits to be set
type LimitsRangeCalculator interface {
	// GetContainerLimitRangeItem returns LimitRangeItem that describes limitation on container limits in the given namespace.
	GetContainerLimitRangeItem(namespace string) (*v1.LimitRangeItem, error)
}

type noopLimitsRangeCalculator struct{}

func (lc *noopLimitsRangeCalculator) GetContainerLimitRangeItem(namespace string) (*v1.LimitRangeItem, error) {
	return nil, nil
}

type limitsChecker struct {
	limitRangeLister v1_listers.LimitRangeLister
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
			if f.Core().V1().LimitRanges().Informer().HasSynced() {
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

func (lc *limitsChecker) GetContainerLimitRangeItem(namespace string) (*v1.LimitRangeItem, error) {
	limitRanges, err := lc.limitRangeLister.LimitRanges(namespace).List(labels.Everything())
	if err != nil {
		return nil, fmt.Errorf("error loading limit ranges: %s", err)
	}

	for _, lr := range limitRanges {
		for _, lri := range lr.Spec.Limits {
			if lri.Type == v1.LimitTypeContainer && (lri.Max != nil || lri.Default != nil) {
				// TODO: handle multiple limit ranges matching a pod.
				return &v1.LimitRangeItem{
					Max:     lri.Max.DeepCopy(),
					Default: lri.Default.DeepCopy(),
				}, nil
			}
		}
	}
	return nil, nil
}
