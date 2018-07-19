/*
Copyright 2016 The Kubernetes Authors.

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

package estimator

import (
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	schedulercache "k8s.io/kubernetes/pkg/scheduler/cache"
)

const (
	//BasicEstimatorName is the name of basic estimator.
	BasicEstimatorName = "basic"
	// BinpackingEstimatorName is the name of binpacking estimator.
	BinpackingEstimatorName = "binpacking"
)

// AvailableEstimators is a list of available estimators.
var AvailableEstimators = []string{BasicEstimatorName, BinpackingEstimatorName}

type Estimator interface {
	Estimate([]*apiv1.Pod, *schedulercache.NodeInfo, []*schedulercache.NodeInfo) int
}

type EstimatorBuilder func(*simulator.PredicateChecker) Estimator

func NewEstimatorBuilder(name string) (EstimatorBuilder, error) {
	switch name {
	case BasicEstimatorName:
		return func(_ *simulator.PredicateChecker) Estimator {
			return NewBasicNodeEstimator()
		}, nil
	case BinpackingEstimatorName:
		return func(predicateChecker *simulator.PredicateChecker) Estimator {
			return NewBinpackingNodeEstimator(predicateChecker)
		}, nil
	}
	return nil, fmt.Errorf("Unknown estimator: %s", name)
}
