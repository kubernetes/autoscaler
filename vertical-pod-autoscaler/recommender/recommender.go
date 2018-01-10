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
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/autoscaler/vertical-pod-autoscaler/recommender/cluster"
	"k8s.io/autoscaler/vertical-pod-autoscaler/recommender/model"
	"k8s.io/autoscaler/vertical-pod-autoscaler/recommender/signals"
	kube_client "k8s.io/client-go/kubernetes"
	v1lister "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

// Recommender recommend resources for certain containers, based on utilization periodically got from metrics api.
type Recommender interface {
	Run()
}

type recommender struct {
	clusterState            model.ClusterState
	specClient              cluster.SpecClient
	metricsFetchingInterval time.Duration
	prometheusClient        signals.PrometheusClient
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
	// TODO: This should also read memory data and merge those two.
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
						CPUUsage:     sample.Value,
						MemoryUsage:  0},
					Container: *containerID})
		}
	}
}

// Currently it just prints out current utilization to the console.
// It will be soon replaced by something more useful.
func (r *recommender) runOnce() {
	glog.V(3).Infof("Recommender Run")

	podSpecs, err := r.specClient.GetPodSpecs()
	if err != nil {
		glog.Errorf("Cannot get SimplePodSpecs. Reason: %+v", err)
	}
	for n, spec := range podSpecs {
		glog.V(3).Infof("SimplePodSpec #%v: %+v", n, spec)
	}
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

// NewRecommender creates a new recommender instance,
// which can be run in order to provide continuous resource recommendations for containers.
// It requires cluster configuration object and duration between recommender intervals.
func NewRecommender(config *rest.Config, metricsFetcherInterval time.Duration, prometheusAddress string) Recommender {
	recommender := &recommender{
		specClient:              newSpecClient(config),
		metricsFetchingInterval: metricsFetcherInterval,
		prometheusClient:        signals.NewPrometheusClient(&http.Client{}, prometheusAddress),
	}
	glog.V(3).Infof("New Recommender created %+v", recommender)

	return recommender
}

func newSpecClient(config *rest.Config) cluster.SpecClient {
	kubeClient := kube_client.NewForConfigOrDie(config)
	podLister := newPodLister(kubeClient)
	return cluster.NewSpecClient(podLister)
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
