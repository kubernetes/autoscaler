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
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/predicatechecker"
)

const (
	// BinpackingEstimatorName is the name of binpacking estimator.
	BinpackingEstimatorName = "binpacking"
)

// AvailableEstimators is a list of available estimators.
var AvailableEstimators = []string{BinpackingEstimatorName}

// PodEquivalenceGroup represents a group of pods, which have the same scheduling
// requirements and are managed by the same controller.
type PodEquivalenceGroup struct {
	Pods []*apiv1.Pod
}

// Exemplar returns an example pod from the group.
func (p *PodEquivalenceGroup) Exemplar() *apiv1.Pod {
	if len(p.Pods) == 0 {
		return nil
	}
	return p.Pods[0]
}

// Estimator calculates the number of nodes of given type needed to schedule pods.
// It returns the number of new nodes needed as well as the list of pods it managed
// to schedule on those nodes.
type Estimator interface {
	// Estimate estimates how many nodes are needed to provision pods coming from the given equivalence groups.
	Estimate([]PodEquivalenceGroup, *framework.NodeInfo, cloudprovider.NodeGroup) (int, []*apiv1.Pod)
}

// EstimatorBuilder creates a new estimator object.
type EstimatorBuilder func(predicatechecker.PredicateChecker, clustersnapshot.ClusterSnapshot, EstimationContext) Estimator

// EstimationAnalyserFunc to be run at the end of the estimation logic.
type EstimationAnalyserFunc func(clustersnapshot.ClusterSnapshot, cloudprovider.NodeGroup, map[string]bool)

// NewEstimatorBuilder creates a new estimator object from flag.
func NewEstimatorBuilder(name string, limiter EstimationLimiter, orderer EstimationPodOrderer, estimationAnalyserFunc EstimationAnalyserFunc) (EstimatorBuilder, error) {
	switch name {
	case BinpackingEstimatorName:
		return func(
			predicateChecker predicatechecker.PredicateChecker,
			clusterSnapshot clustersnapshot.ClusterSnapshot,
			context EstimationContext) Estimator {
			return NewBinpackingNodeEstimator(predicateChecker, clusterSnapshot, limiter, orderer, context, estimationAnalyserFunc)
		}, nil
	}
	return nil, fmt.Errorf("unknown estimator: %s", name)
}

// EstimationLimiter controls how many nodes can be added by Estimator.
// A limiter can be used to prevent costly estimation if an actual ability to
// scale-up is limited by external factors.
type EstimationLimiter interface {
	// StartEstimation is called at the start of estimation.
	StartEstimation([]PodEquivalenceGroup, cloudprovider.NodeGroup, EstimationContext)
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
	Order(podsEquivalentGroups []PodEquivalenceGroup, nodeTemplate *framework.NodeInfo, nodeGroup cloudprovider.NodeGroup) []PodEquivalenceGroup
}
