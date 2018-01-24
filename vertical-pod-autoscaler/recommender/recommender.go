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
	"fmt"
	"net/http"
	"time"

	"github.com/golang/glog"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	vpa_api "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned/typed/poc.autoscaling.k8s.io/v1alpha1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/recommender/clients"
	"k8s.io/autoscaler/vertical-pod-autoscaler/recommender/logic"
	"k8s.io/autoscaler/vertical-pod-autoscaler/recommender/model"
	"k8s.io/autoscaler/vertical-pod-autoscaler/recommender/signals"
	kube_client "k8s.io/client-go/kubernetes"
	v1lister "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	resourceclient "k8s.io/metrics/pkg/client/clientset_generated/clientset/typed/metrics/v1beta1"
)

// Recommender recommend resources for certain containers, based on utilization periodically got from metrics api.
type Recommender interface {
	Run()
}

type recommender struct {
	clusterState            *model.ClusterState
	clusterFeeder           clients.ClusterStateFeeder
	metricsFetchingInterval time.Duration
	prometheusClient        signals.PrometheusClient
	vpaClient               vpa_api.VerticalPodAutoscalerInterface
	podResourceRecommender  logic.PodResourceRecommender
}

func getContainerIDFromLabels(labels map[string]string) (*model.ContainerID, error) {
	namespace, ok := labels["namespace"]
	if !ok {
		return nil, fmt.Errorf("no namespace label")
	}
	podName, ok := labels["pod_name"]
	if !ok {
		return nil, fmt.Errorf("no pod_name label")
	}
	containerName, ok := labels["name"]
	if !ok {
		return nil, fmt.Errorf("no name label on container data")
	}
	return &model.ContainerID{
		PodID: model.PodID{
			Namespace: namespace,
			PodName:   podName},
		ContainerName: containerName}, nil
}

func (r *recommender) readHistory() {
	// TODO: Add one more layer of abstraction so that recommender does not know it's
	// talking to Prometheus and does not have to hardcode queries.
	// TODO: This should also read memory data.
	tss, err := r.prometheusClient.GetTimeseries("container_cpu_usage_seconds_total[1d]")
	if err != nil {
		glog.Errorf("Cannot get timeseries: %v", err)
	}
	for _, ts := range tss {
		containerID, err := getContainerIDFromLabels(ts.Labels)
		if err != nil {
			glog.Errorf("Cannot get container ID from labels: %v", ts.Labels)
			continue
		}
		for _, sample := range ts.Samples {
			r.clusterState.AddSample(
				&model.ContainerUsageSampleWithKey{
					ContainerUsageSample: model.ContainerUsageSample{
						MeasureStart: sample.Timestamp,
						Usage:        sample.Value,
						Resource:     model.ResourceCPU},
					Container: *containerID})
		}
	}
}

// Currently it just prints out current utilization to the console.
// It will be soon replaced by something more useful.

func (r *recommender) runOnce() {
	glog.V(3).Infof("Recommender Run")

	vpas, err := r.vpaClient.List(metav1.ListOptions{})
	if err != nil {
		glog.Errorf("Cannot list VPAs. Reason: %+v", err)
	} else {
		glog.V(3).Infof("Fetched VPAs.")
	}
	for n, vpa := range vpas.Items {
		glog.V(3).Infof("VPA #%v: %+v", n, vpa)
	}

	r.clusterFeeder.Feed()
	glog.V(3).Infof("Current ClusterState:  %+v", r.clusterState)
}

func (r *recommender) Run() {
	r.readHistory()
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
	// Create a fake recommender that returns a hard-coded recommendation.
	// TODO: Replace with a real recommender based on past usage.
	var MiB float64 = 1024 * 1024
	target := model.Resources{
		model.ResourceCPU:    model.CPUAmountFromCores(0.5),
		model.ResourceMemory: model.MemoryAmountFromBytes(200. * MiB),
	}
	lowerBound := model.Resources{
		model.ResourceCPU:    model.CPUAmountFromCores(0.4),
		model.ResourceMemory: model.MemoryAmountFromBytes(150. * MiB),
	}
	upperBound := model.Resources{
		model.ResourceCPU:    model.CPUAmountFromCores(0.6),
		model.ResourceMemory: model.MemoryAmountFromBytes(250. * MiB),
	}
	return logic.NewPodResourceRecommender(
		logic.NewConstEstimator(target),
		logic.NewConstEstimator(lowerBound),
		logic.NewConstEstimator(upperBound))
}

// NewRecommender creates a new recommender instance,
// which can be run in order to provide continuous resource recommendations for containers.
// It requires cluster configuration object and duration between recommender intervals.
func NewRecommender(namespace string, config *rest.Config, metricsFetcherInterval time.Duration, prometheusAddress string) Recommender {
	clusterState := model.NewClusterState()
	feeder := clients.NewClusterStateFeeder(clusterState, newSpecClient(config), newMetricsClient(config))

	recommender := &recommender{
		clusterState:            clusterState,
		clusterFeeder:           feeder,
		metricsFetchingInterval: metricsFetcherInterval,
		prometheusClient:        signals.NewPrometheusClient(&http.Client{}, prometheusAddress),
		vpaClient:               vpa_clientset.NewForConfigOrDie(config).PocV1alpha1().VerticalPodAutoscalers(namespace),
		podResourceRecommender:  createPodResourceRecommender(),
	}
	glog.V(3).Infof("New Recommender created %+v", recommender)

	return recommender
}

func newSpecClient(config *rest.Config) clients.SpecClient {
	kubeClient := kube_client.NewForConfigOrDie(config)
	podLister := newPodLister(kubeClient)
	return clients.NewSpecClient(podLister)
}

func newMetricsClient(config *rest.Config) clients.MetricsClient {
	metricsGetter := resourceclient.NewForConfigOrDie(config)
	return clients.NewMetricsClient(metricsGetter)
}

// Creates PodLister, listing only not terminated pods.
func newPodLister(kubeClient kube_client.Interface) v1lister.PodLister {
	selector := fields.ParseSelectorOrDie("status.phase!=" + string(apiv1.PodPending) + ",status.phase!=" + string(apiv1.PodUnknown))
	podListWatch := cache.NewListWatchFromClient(kubeClient.CoreV1().RESTClient(), "pods", apiv1.NamespaceAll, selector)
	store := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	podLister := v1lister.NewPodLister(store)
	podReflector := cache.NewReflector(podListWatch, &apiv1.Pod{}, store, time.Hour)
	stopCh := make(chan struct{})
	go podReflector.Run(stopCh)
	return podLister
}
