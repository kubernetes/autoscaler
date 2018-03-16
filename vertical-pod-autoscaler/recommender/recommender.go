/*
Copyright 2017 The Kubernetes Authors.

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

package main

import (
	"time"

	"github.com/golang/glog"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/poc.autoscaling.k8s.io/v1alpha1"
	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	vpa_api "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned/typed/poc.autoscaling.k8s.io/v1alpha1"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
	"k8s.io/autoscaler/vertical-pod-autoscaler/recommender/input"
	"k8s.io/autoscaler/vertical-pod-autoscaler/recommender/input/history"
	"k8s.io/autoscaler/vertical-pod-autoscaler/recommender/logic"
	"k8s.io/autoscaler/vertical-pod-autoscaler/recommender/model"
	"k8s.io/client-go/rest"
)

// Recommender recommend resources for certain containers, based on utilization periodically got from metrics api.
type Recommender interface {
	Run()
}

type recommender struct {
	clusterState            *model.ClusterState
	clusterStateFeeder      input.ClusterStateFeeder
	metricsFetchingInterval time.Duration
	vpaClient               vpa_api.VerticalPodAutoscalersGetter
	podResourceRecommender  logic.PodResourceRecommender
}

// Updates VPA CRD objects' statuses.
func (r *recommender) updateVPAs() {
	for key, vpa := range r.clusterState.Vpas {
		glog.V(3).Infof("VPA to update #%v: %+v", key, vpa)
		vpa.Conditions.Set(vpa_types.Configured, true, "", "")
		resources := r.podResourceRecommender.GetRecommendedPodResources(vpa)
		containerResources := make([]vpa_types.RecommendedContainerResources, 0, len(resources))
		for containerID, res := range resources {
			containerResources = append(containerResources, vpa_types.RecommendedContainerResources{
				Name:           containerID,
				Target:         model.ResourcesAsResourceList(res.Target),
				MinRecommended: model.ResourcesAsResourceList(res.MinRecommended),
				MaxRecommended: model.ResourcesAsResourceList(res.MaxRecommended),
			})

		}
		vpa.Recommendation = &vpa_types.RecommendedPodResources{containerResources}
		vpa.Conditions.Set(vpa_types.RecommendationProvided, true, "", "")

		_, err := vpa_api_util.UpdateVpaStatus(
			r.vpaClient.VerticalPodAutoscalers(vpa.ID.Namespace), vpa)
		if err != nil {
			glog.Errorf(
				"Cannot update VPA %v object. Reason: %+v", vpa.ID.VpaName, err)
		}
	}

}

// Currently it just prints out current utilization to the console.
// It will be soon replaced by something more useful.
func (r *recommender) runOnce() {
	glog.V(3).Infof("Recommender Run")
	r.clusterStateFeeder.LoadVPAs()
	r.clusterStateFeeder.LoadPods()
	r.clusterStateFeeder.LoadRealTimeMetrics()
	r.updateVPAs()
	glog.V(3).Infof("ClusterState is tracking  %v PodStates and %v VPAs", len(r.clusterState.Pods), len(r.clusterState.Vpas))
}

func (r *recommender) Run() {
	r.clusterStateFeeder.LoadHistory()
	for {
		select {
		case <-time.After(r.metricsFetchingInterval):
			{
				r.runOnce()
			}
		}
	}
}

func createPodResourceRecommender() logic.PodResourceRecommender {
	targetCPUPercentile := 0.9
	lowerBoundCPUPercentile := 0.5
	upperBoundCPUPercentile := 0.95

	targetMemoryPeaksPercentile := 0.9
	lowerBoundMemoryPeaksPercentile := 0.5
	upperBoundMemoryPeaksPercentile := 0.95

	targetEstimator := logic.NewPercentileEstimator(targetCPUPercentile, targetMemoryPeaksPercentile)
	lowerBoundEstimator := logic.NewPercentileEstimator(lowerBoundCPUPercentile, lowerBoundMemoryPeaksPercentile)
	upperBoundEstimator := logic.NewPercentileEstimator(upperBoundCPUPercentile, upperBoundMemoryPeaksPercentile)

	// Apply confidence multiplier to the upper bound estimator. This means
	// that the updater will be less eager to evict pods with short history
	// in order to reclaim unused resources.
	// Using the confidence multiplier 1 with exponent +1 means that
	// the upper bound is multiplied by (1 + 1/history-length-in-days).
	// See estimator.go to see how the history length and the confidence
	// multiplier are determined. The formula yields the following multipliers:
	// No history     : *INF  (do not force pod eviction)
	// 12h history    : *3    (force pod eviction if the request is > 3 * upper bound)
	// 24h history    : *2
	// 1 week history : *1.14	
	upperBoundEstimator = logic.WithConfidenceMultiplier(1.0, 1.0, upperBoundEstimator)

	// Apply confidence multiplier to the lower bound estimator. This means
	// that the updater will be less eager to evict pods with short history
	// in order to provision them with more resources.
	// Using the confidence multiplier 0.001 with exponent -2 means that
	// the lower bound is multiplied by the factor (1 + 0.001/history-length-in-days)^-2
	// (which is very rapidly converging to 1.0).
	// See estimator.go to see how the history length and the confidence
	// multiplier are determined. The formula yields the following multipliers:
	// No history   : *0   (do not force pod eviction)
	// 5m history   : *0.6 (force pod eviction if the request is < 0.6 * lower bound)
	// 30m history  : *0.9
	// 60m history  : *0.95
	lowerBoundEstimator = logic.WithConfidenceMultiplier(0.001, -2.0, lowerBoundEstimator)

	return logic.NewPodResourceRecommender(
		targetEstimator,
		lowerBoundEstimator,
		upperBoundEstimator)
}

// NewRecommender creates a new recommender instance,
// which can be run in order to provide continuous resource recommendations for containers.
// It requires cluster configuration object and duration between recommender intervals.
func NewRecommender(config *rest.Config, metricsFetcherInterval time.Duration, historyProvider history.HistoryProvider) Recommender {
	clusterState := model.NewClusterState()
	recommender := &recommender{
		clusterState:            clusterState,
		clusterStateFeeder:      input.NewClusterStateFeeder(config, historyProvider, clusterState),
		metricsFetchingInterval: metricsFetcherInterval,
		vpaClient:               vpa_clientset.NewForConfigOrDie(config).PocV1alpha1(),
		podResourceRecommender:  createPodResourceRecommender(),
	}
	glog.V(3).Infof("New Recommender created %+v", recommender)
	return recommender
}
