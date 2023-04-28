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
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/predicatechecker"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

const (
	// BinpackingEstimatorName is the name of binpacking estimator.
	BinpackingEstimatorName = "binpacking"
)

// AvailableEstimators is a list of available estimators.
var AvailableEstimators = []string{BinpackingEstimatorName}

// Estimator calculates the number of nodes of given type needed to schedule pods.
// It returns the number of new nodes needed as well as the list of pods it managed
// to schedule on those nodes.
type Estimator interface {
	Estimate([]*apiv1.Pod, *schedulerframework.NodeInfo, cloudprovider.NodeGroup) (int, []*apiv1.Pod)
}

// EstimatorBuilder creates a new estimator object.
type EstimatorBuilder func(predicatechecker.PredicateChecker, clustersnapshot.ClusterSnapshot) Estimator

// NewEstimatorBuilder creates a new estimator object from flag.
func NewEstimatorBuilder(name string, limiter EstimationLimiter, orderer EstimationPodOrderer) (EstimatorBuilder, error) {
	switch name {
	case BinpackingEstimatorName:
		return func(
			predicateChecker predicatechecker.PredicateChecker,
			clusterSnapshot clustersnapshot.ClusterSnapshot) Estimator {
			return NewBinpackingNodeEstimator(predicateChecker, clusterSnapshot, limiter, orderer)
		}, nil
	}
	return nil, fmt.Errorf("unknown estimator: %s", name)
}

// EstimationLimiter controls how many nodes can be added by Estimator.
// A limiter can be used to prevent costly estimation if an actual ability to
// scale-up is limited by external factors.
type EstimationLimiter interface {
	// StartEstimation is called at the start of estimation.
	StartEstimation([]*apiv1.Pod, cloudprovider.NodeGroup)
	// EndEstimation is called at the end of estimation.
	EndEstimation()
	// PermissionToAddNode is called by an estimator when it wants to add additional
	// nodes to simulation. If permission is not granted the Estimator is expected
	// not to add any more nodes in this simulation.
	// There is no requirement for the Estimator to stop calculations, it's
	// just not expected to add any more nodes.
	PermissionToAddNode() bool
}

// EstimationPodOrderer is an interface used to determine the order of the pods
// used while binpacking during scale up estimation
type EstimationPodOrderer interface {
	Order(pods []*apiv1.Pod, nodeTemplate *framework.NodeInfo, nodeGroup cloudprovider.NodeGroup) []*apiv1.Pod
}
