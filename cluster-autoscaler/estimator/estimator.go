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
	"k8s.io/klog"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"
)

const (
	// BinpackingEstimatorName is the name of binpacking estimator.
	BinpackingEstimatorName = "binpacking"
	// OldBinpackingEstimatorName is the name of the older binpacking estimator.
	OldBinpackingEstimatorName = "oldbinpacking"
)

// AvailableEstimators is a list of available estimators.
var AvailableEstimators = []string{BinpackingEstimatorName, OldBinpackingEstimatorName}

// Estimator calculates the number of nodes of given type needed to schedule pods.
type Estimator interface {
	Estimate([]*apiv1.Pod, *schedulernodeinfo.NodeInfo, []*schedulernodeinfo.NodeInfo) int
}

// EstimatorBuilder creates a new estimator object.
type EstimatorBuilder func(*simulator.PredicateChecker) Estimator

// NewEstimatorBuilder creates a new estimator object from flag.
func NewEstimatorBuilder(name string) (EstimatorBuilder, error) {
	switch name {
	case BinpackingEstimatorName:
		return func(predicateChecker *simulator.PredicateChecker) Estimator {
			return NewBinpackingNodeEstimator(predicateChecker)
		}, nil
	// Deprecated.
	case OldBinpackingEstimatorName:
		klog.Warning(oldBinPackingEstimatorDeprecationMessage)
		return func(predicateChecker *simulator.PredicateChecker) Estimator {
			return NewOldBinpackingNodeEstimator(predicateChecker)
		}, nil
	}
	return nil, fmt.Errorf("unknown estimator: %s", name)
}
