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
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang/glog"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/poc.autoscaling.k8s.io/v1alpha1"
	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	vpa_api "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned/typed/poc.autoscaling.k8s.io/v1alpha1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/recommender/cluster"
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
	specClient              cluster.SpecClient
	metricsClient           cluster.MetricsClient
	metricsFetchingInterval time.Duration
	prometheusClient        signals.PrometheusClient
	vpaClient               vpa_api.VerticalPodAutoscalerInterface
	podResourceRecommender  logic.PodResourceRecommender
}

type patchRecord struct {
	Op    string      `json:"op,inline"`
	Path  string      `json:"path,inline"`
	Value interface{} `json:"value"`
}

type verticalPodAutoscalerStatusPatch struct {
	Status vpa_types.VerticalPodAutoscalerStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
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

func initVPAStatus(vpaClient vpa_api.VerticalPodAutoscalerInterface, vpaName string) {
	patchVPA(vpaClient, vpaName, []patchRecord{{
		Op:    "add",
		Path:  "/status",
		Value: vpa_types.VerticalPodAutoscalerStatus{},
	},
	})
}

func patchVPA(vpaClient vpa_api.VerticalPodAutoscalerInterface, vpaName string, patches []patchRecord) {
	bytes, err := json.Marshal(patches)
	if err != nil {
		glog.Errorf("Cannot marshal VPA status patches %+v. Reason: %+v", patches, err)
		return
	}

	_, err = vpaClient.Patch(vpaName, types.JSONPatchType, bytes)
	if err != nil {
		glog.Errorf("Cannot patch VPA %v. Reason: %+v", vpaName, err)
	} else {
		glog.V(3).Infof("VPA %v patched", vpaName)
	}

}

// Fetch VPA objects and load them into the cluster state.
func (r *recommender) loadVPAs() {
	vpaCRDs, err := r.vpaClient.List(metav1.ListOptions{})
	if err != nil {
		glog.Errorf("Cannot list VPAs. Reason: %+v", err)
	} else {
		glog.V(3).Infof("Fetched VPAs.")
	}

	vpas := make(map[model.VpaID]labels.Selector)
	for n, vpaCRD := range vpaCRDs.Items {
		glog.V(3).Infof("VPA CRD #%v: %+v", n, vpaCRD)
		selector, err := metav1.LabelSelectorAsSelector(vpaCRD.Spec.Selector)
		if err != nil {
			glog.Errorf(
				"Cannot convert VPA Selector %+v to internal representation. Reason: %+v",
				vpaCRD.Spec.Selector, err)
			continue
		}
		vpaName := vpaCRD.ObjectMeta.Name
		vpas[model.VpaID{vpaName}] = selector
		if vpaCRD.Status.LastUpdateTime.IsZero() {
			glog.V(3).Infof("Empty status in %v, initializing", vpaName)
			initVPAStatus(r.vpaClient, vpaName)

		}
	}
	for key := range r.clusterState.Vpas {
		if _, exists := vpas[key]; !exists {
			glog.V(3).Infof("Deleting VPA %v", key)
			r.clusterState.DeleteVpa(key)
		}
	}
	for key, selector := range vpas {
		r.clusterState.AddOrUpdateVpa(key, selector.String())
	}
}

// Load pod into the cluster state.
func (r *recommender) loadPods() {
	podSpecs, err := r.specClient.GetPodSpecs()
	if err != nil {
		glog.Errorf("Cannot get SimplePodSpecs. Reason: %+v", err)
	}
	pods := make(map[model.PodID]*model.BasicPodSpec)
	for n, spec := range podSpecs {
		glog.V(3).Infof("SimplePodSpec #%v: %+v", n, spec)
		pods[spec.ID] = spec
	}
	for key := range r.clusterState.Pods {
		if _, exists := pods[key]; !exists {
			glog.V(3).Infof("Deleting Pod %v", key)
			r.clusterState.DeletePod(key)
		}
	}
	for _, pod := range pods {
		r.clusterState.AddOrUpdatePod(pod.ID, pod.PodLabels)
		for _, container := range pod.Containers {
			r.clusterState.AddOrUpdateContainer(container.ID)
		}
	}
}

// Converts internal Resources representation to ResourcesList.
func resoucesAsResourceList(resources model.Resources) apiv1.ResourceList {
	result := make(apiv1.ResourceList)
	for key, value := range resources {
		var newKey apiv1.ResourceName
		switch key {
		case model.ResourceCPU:
			newKey = apiv1.ResourceCPU
		case model.ResourceMemory:
			newKey = apiv1.ResourceMemory
		default:
			glog.Errorf("Cannot translate %v resource name", key)
			continue
		}
		result[newKey] = *resource.NewScaledQuantity(int64(value), 0)
	}

	return result
}

// Updates VPA CRD objects' statuses.
func (r *recommender) updateVPAs() {
	for key, vpa := range r.clusterState.Vpas {
		glog.V(3).Infof("VPA to update #%v: %+v", key, vpa)

		resources := r.podResourceRecommender.GetRecommendedPodResources(vpa)
		vpaName := vpa.ID.VpaName

		contanerResources := make([]vpa_types.RecommendedContainerResources, 0, len(resources))
		for containerId, res := range resources {
			contanerResources = append(contanerResources, vpa_types.RecommendedContainerResources{
				Name:           containerId,
				Target:         resoucesAsResourceList(res.Target),
				MinRecommended: resoucesAsResourceList(res.MinRecommended),
				MaxRecommended: resoucesAsResourceList(res.MaxRecommended),
			})

		}

		recommendation := vpa_types.RecommendedPodResources{contanerResources}

		patches := []patchRecord{
			{
				Op:    "add",
				Path:  "/status/lastUpdateTime",
				Value: metav1.Time{time.Now()},
			},
			{
				Op:    "add",
				Path:  "/status/recommendation",
				Value: recommendation,
			},
		}
		patchVPA(r.vpaClient, vpaName, patches)

	}

}

// Currently it just prints out current utilization to the console.
// It will be soon replaced by something more useful.
func (r *recommender) runOnce() {
	glog.V(3).Infof("Recommender Run")
	r.loadVPAs()
	r.loadPods()

	metricsSnapshots, err := r.metricsClient.GetContainersMetrics()
	if err != nil {
		glog.Errorf("Cannot get containers metrics. Reason: %+v", err)
	}
	for n, snap := range metricsSnapshots {
		glog.V(3).Infof("ContainerMetricsSnapshot #%v: %+v", n, snap)
	}
	r.updateVPAs()
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
	recommender := &recommender{
		clusterState:            model.NewClusterState(),
		specClient:              newSpecClient(config),
		metricsClient:           newMetricsClient(config),
		metricsFetchingInterval: metricsFetcherInterval,
		prometheusClient:        signals.NewPrometheusClient(&http.Client{}, prometheusAddress),
		vpaClient:               vpa_clientset.NewForConfigOrDie(config).PocV1alpha1().VerticalPodAutoscalers(namespace),
		podResourceRecommender:  createPodResourceRecommender(),
	}
	glog.V(3).Infof("New Recommender created %+v", recommender)

	return recommender
}

func newSpecClient(config *rest.Config) cluster.SpecClient {
	kubeClient := kube_client.NewForConfigOrDie(config)
	podLister := newPodLister(kubeClient)
	return cluster.NewSpecClient(podLister)
}

func newMetricsClient(config *rest.Config) cluster.MetricsClient {
	metricsGetter := resourceclient.NewForConfigOrDie(config)
	return cluster.NewMetricsClient(metricsGetter)
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
