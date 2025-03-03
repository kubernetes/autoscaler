/*
Copyright 2023 The Kubernetes Authors.

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

package metrics

import (
	"context"
	"os"
	"time"

	k8sapiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	resourceclient "k8s.io/metrics/pkg/client/clientset/versioned/typed/metrics/v1beta1"
	"k8s.io/metrics/pkg/client/external_metrics"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
)

// PodMetricsLister wraps both metrics-client and External Metrics
type PodMetricsLister interface {
	List(ctx context.Context, namespace string, opts v1.ListOptions) (*v1beta1.PodMetricsList, error)
}

// podMetricsSource is the metrics-client source of metrics.
type podMetricsSource struct {
	metricsGetter resourceclient.PodMetricsesGetter
}

// NewPodMetricsesSource Returns a Source-wrapper around PodMetricsesGetter.
func NewPodMetricsesSource(source resourceclient.PodMetricsesGetter) PodMetricsLister {
	return podMetricsSource{metricsGetter: source}
}

func (s podMetricsSource) List(ctx context.Context, namespace string, opts v1.ListOptions) (*v1beta1.PodMetricsList, error) {
	podMetricsInterface := s.metricsGetter.PodMetricses(namespace)
	return podMetricsInterface.List(ctx, opts)
}

// externalMetricsClient is the External Metrics source of metrics.
type externalMetricsClient struct {
	externalClient external_metrics.ExternalMetricsClient
	options        ExternalClientOptions
	clusterState   *model.ClusterState
}

// ExternalClientOptions specifies parameters for using an External Metrics Client.
type ExternalClientOptions struct {
	ResourceMetrics map[k8sapiv1.ResourceName]string
	// Label to use for the container name.
	ContainerNameLabel string
}

// NewExternalClient returns a Source for an External Metrics Client.
func NewExternalClient(c *rest.Config, clusterState *model.ClusterState, options ExternalClientOptions) PodMetricsLister {
	extClient, err := external_metrics.NewForConfig(c)
	if err != nil {
		klog.ErrorS(err, "Failed initializing external metrics client")
		os.Exit(255)
	}
	return &externalMetricsClient{
		externalClient: extClient,
		options:        options,
		clusterState:   clusterState,
	}
}

func (s *externalMetricsClient) List(ctx context.Context, namespace string, opts v1.ListOptions) (*v1beta1.PodMetricsList, error) {
	result := v1beta1.PodMetricsList{}

	for _, vpa := range s.clusterState.Vpas {
		if vpa.PodCount == 0 {
			continue
		}

		if namespace != "" && vpa.ID.Namespace != namespace {
			continue
		}

		nsClient := s.externalClient.NamespacedMetrics(vpa.ID.Namespace)
		pods := s.clusterState.GetMatchingPods(vpa)

		for _, pod := range pods {
			podNameReq, err := labels.NewRequirement("pod", selection.Equals, []string{pod.PodName})
			if err != nil {
				return nil, err
			}
			selector := vpa.PodSelector.Add(*podNameReq)
			podMets := v1beta1.PodMetrics{
				TypeMeta:   v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{Namespace: vpa.ID.Namespace, Name: pod.PodName},
				Window:     v1.Duration{},
				Containers: make([]v1beta1.ContainerMetrics, 0),
			}
			// Query each resource in turn, then assemble back to a single []ContainerMetrics.
			containerMetrics := make(map[string]k8sapiv1.ResourceList)
			for resourceName, metricName := range s.options.ResourceMetrics {
				m, err := nsClient.List(metricName, selector)
				if err != nil {
					return nil, err
				}
				if m == nil || len(m.Items) == 0 {
					klog.V(4).InfoS("External Metrics Query for VPA: No items", "vpa", klog.KRef(vpa.ID.Namespace, vpa.ID.VpaName), "resource", resourceName, "metric", metricName)
					continue
				}
				klog.V(4).InfoS("External Metrics Query for VPA", "vpa", klog.KRef(vpa.ID.Namespace, vpa.ID.VpaName), "resource", resourceName, "metric", metricName, "itemCount", len(m.Items), "firstItem", m.Items[0])
				podMets.Timestamp = m.Items[0].Timestamp
				if m.Items[0].WindowSeconds != nil {
					podMets.Window = v1.Duration{Duration: time.Duration(*m.Items[0].WindowSeconds) * time.Second}
				}
				for _, val := range m.Items {
					ctrName, hasCtrName := val.MetricLabels[s.options.ContainerNameLabel]
					if !hasCtrName {
						continue
					}
					if containerMetrics[ctrName] == nil {
						containerMetrics[ctrName] = make(k8sapiv1.ResourceList)
					}
					containerMetrics[ctrName][resourceName] = val.Value
				}

			}
			for cname, res := range containerMetrics {
				podMets.Containers = append(podMets.Containers, v1beta1.ContainerMetrics{Name: cname, Usage: res})
			}
			result.Items = append(result.Items, podMets)

		}
	}
	return &result, nil
}
