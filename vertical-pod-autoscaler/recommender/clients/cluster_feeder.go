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

package clients

import (
	"github.com/golang/glog"
	vpa_api "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned/typed/poc.autoscaling.k8s.io/v1alpha1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/recommender/model"
)

// ClusterStateFeeder can update state of ClusterState object.
type ClusterStateFeeder interface {
	// Subscribe saves a pointer to a ClusterState object which will be updated.
	Subscribe(cluster *model.ClusterState)
	// Feed uptades state of subscribed ClusterState
	Feed()
}

type clusterStateFeeder struct {
	specClient    SpecClient
	metricsClient MetricsClient
	vpaClient     vpa_api.VerticalPodAutoscalerInterface
	subscriber    *model.ClusterState
}

// NewClusterStateFeeder takes SpecClient and MetricsClient and creates new instance of ClusterStateFeeder,
// which can be used to update state of subscribed ClusterState
func NewClusterStateFeeder(specClient SpecClient, metricsClient MetricsClient) ClusterStateFeeder {
	return &clusterStateFeeder{
		specClient:    specClient,
		metricsClient: metricsClient,
	}
}

func (feeder *clusterStateFeeder) Subscribe(cluster *model.ClusterState) {
	if feeder.subscriber != nil {
		glog.Warning("One cluster is already subscribed to the feeder. Replacing with a new one.")
	}
	feeder.subscriber = cluster
}

func (feeder *clusterStateFeeder) Feed() {
	// TODO: Add feeding ClusterState with VPAObjects.
	// Currently it only feeds container specs and metrics.
	if feeder.subscriber != nil {
		feeder.feedSpecs()
		feeder.feedMetrics()
	} else {
		glog.Warning("No cluster state has subscribed. Feeding skipped.")
	}
}

func (feeder *clusterStateFeeder) feedSpecs() {
	// TODO: distinguish remove pods and remove them from our ClusterState
	podSpecs, err := feeder.specClient.GetPodSpecs()
	if err != nil {
		glog.Errorf("Cannot get SimplePodSpecs from SpecClient. Reason: %+v", err)
	}

	containerCount := 0
	for _, podSpec := range podSpecs {
		feeder.subscriber.AddOrUpdatePod(podSpec.ID, podSpec.PodLabels)
		for _, containerSpec := range podSpec.Containers {
			feeder.subscriber.AddOrUpdateContainer(containerSpec.ID)
			containerCount++
		}
	}
	glog.V(3).Infof("ClusterSpec fed with #%v BasicContainerSpecs and #%v BasicPodSpecs", containerCount, len(podSpecs))
}

func (feeder *clusterStateFeeder) feedMetrics() {
	containersMetrics, err := feeder.metricsClient.GetContainersMetrics()
	if err != nil {
		glog.Errorf("Cannot get ContainerMetricsSnapshot from MetricsClient. Reason: %+v", err)
	}

	sampleCount := 0
	for _, containerMetrics := range containersMetrics {
		for _, sample := range newContainerUsageSampleWithKey(containerMetrics) {
			feeder.subscriber.AddSample(sample)
			sampleCount++
		}
	}
	glog.V(3).Infof("ClusterSpec fed with #%v ContainerUsageSamples for #%v containers", sampleCount, len(containersMetrics))
}

func newContainerUsageSampleWithKey(metrics *model.ContainerMetricsSnapshot) []*model.ContainerUsageSampleWithKey {
	var samples []*model.ContainerUsageSampleWithKey

	for metricName, resourceAmmount := range metrics.Usage {
		usage, err := model.UsageFromResourceAmount(metricName, resourceAmmount)
		if err != nil {
			glog.Errorf("Cannot calculate resource usage. Skipping this sample. Reason: %+v", err)
		} else {
			sample := &model.ContainerUsageSampleWithKey{
				Container: metrics.ID,
				ContainerUsageSample: model.ContainerUsageSample{
					MeasureStart: metrics.SnapshotTime,
					Resource:     metricName,
					Usage:        usage,
				},
			}
			samples = append(samples, sample)
		}
	}
	return samples
}
