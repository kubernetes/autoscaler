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

package nanny

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	scheme "k8s.io/client-go/kubernetes/scheme"
)

const (
	// objectCountMetricName is the preferred metric to be used to get number of nodes (present in Kubernetes 1.21 and higher)
	objectCountMetricName = "apiserver_storage_objects"
	// objectCountFallbackMetricName is the metric to be used to get number of nodes if objectCountMetricName metric is missing
	objectCountFallbackMetricName = "etcd_object_counts"
	// nodeResourceName is the label value for Nodes in objectCountFallbackMetricName and objectCountMetricName metrics.
	nodeResourceName = "nodes"
	// resourceLabel is the label name for resource.
	resourceLabel = "resource"
)

type kubernetesClient struct {
	namespace  string
	deployment string
	pod        string
	container  string
	clientset  *kubernetes.Clientset
	useMetrics bool
}

// CountNodes returns the number of nodes in the cluster:
// 1) by listing Nodes using API (default)
// 2) using etcd_object_count metric exposed by kube-apiserver
func (k *kubernetesClient) CountNodes() (uint64, error) {
	if k.useMetrics {
		return k.countNodesThroughMetrics()
	}
	return k.countNodesThroughAPI()
}

func (k *kubernetesClient) countNodesThroughAPI() (uint64, error) {
	// Set ResourceVersion = 0 to use cached versions.
	options := metav1.ListOptions{
		ResourceVersion: "0",
	}
	result := &metav1beta1.PartialObjectMetadataList{}
	err := k.clientset.
		CoreV1().
		RESTClient().
		Get().
		Resource("nodes").
		// Set as=PartialObjectMetadataList to fetch only nodes metadata.
		SetHeader("Accept", "application/vnd.kubernetes.protobuf;as=PartialObjectMetadataList;g=meta.k8s.io;v=v1beta1").
		VersionedParams(&options, scheme.ParameterCodec).
		Do(context.Background()).
		Into(result)
	return uint64(len(result.Items)), err
}

func hasEqualValues(a string, b *string) bool {
	return b != nil && a == *b
}

func extractMetricValueForNodeCount(mf dto.MetricFamily) (uint64, error) {
	for _, metric := range mf.Metric {
		hasLabel := false
		for _, label := range metric.Label {
			if hasEqualValues(resourceLabel, label.Name) && hasEqualValues(nodeResourceName, label.Value) {
				hasLabel = true
				break
			}
		}
		if !hasLabel {
			continue
		}
		if metric.Gauge == nil || metric.Gauge.Value == nil {
			continue
		}
		if *metric.Gauge.Value < 0 {
			return 0, fmt.Errorf("metric unknown")
		}
		value := uint64(*metric.Gauge.Value)
		return value, nil
	}
	return 0, fmt.Errorf("no valid metric values")
}

func getNodeCountFromDecoder(decoder expfmt.Decoder) (uint64, error) {
	var mf dto.MetricFamily
	var fallbackNodeCountMetricValue uint64
	var fallbackMetricError error
	useNodeCountFromFallbackMetric := false

	for {
		if err := decoder.Decode(&mf); err != nil {
			if err == io.EOF {
				break
			}
			return 0, fmt.Errorf("decoding error: %v", err)
		}
		if hasEqualValues(objectCountMetricName, mf.Name) {
			// Preferred metric is present - return immediately
			return extractMetricValueForNodeCount(mf)
		}
		if hasEqualValues(objectCountFallbackMetricName, mf.Name) {
			// Fallback metric is present - store values to return later (in case the preferred metric is missing)
			fallbackNodeCountMetricValue, fallbackMetricError = extractMetricValueForNodeCount(mf)
			useNodeCountFromFallbackMetric = true
		}
	}

	if useNodeCountFromFallbackMetric {
		return fallbackNodeCountMetricValue, fallbackMetricError
	}
	return 0, fmt.Errorf("metric unset")
}

func (k *kubernetesClient) countNodesThroughMetrics() (uint64, error) {
	// Similarly as for listing nodes, permissions for /metrics endpoint are needed.
	// Other than that, endpoint is visible from everywhere.
	reader, err := k.clientset.CoreV1().RESTClient().Get().RequestURI("/metrics").Stream(context.Background())
	if err != nil {
		return 0, err
	}
	return getNodeCountFromDecoder(expfmt.NewDecoder(reader, expfmt.FmtText))
}

func (k *kubernetesClient) ContainerResources() (*corev1.ResourceRequirements, error) {
	pod, err := k.clientset.CoreV1().Pods(k.namespace).Get(context.Background(), k.pod, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	for _, container := range pod.Spec.Containers {
		if container.Name == k.container {
			return &container.Resources, nil
		}
	}
	return nil, fmt.Errorf("Container %s was not found in deployment %s in namespace %s.", k.container, k.deployment, k.namespace)
}

func (k *kubernetesClient) UpdateDeployment(resources *corev1.ResourceRequirements) error {
	// First, get the Deployment.
	dep, err := k.clientset.AppsV1().Deployments(k.namespace).Get(context.Background(), k.deployment, metav1.GetOptions{})
	if err != nil {
		return err
	}

	// Modify the Deployment object with our ResourceRequirements.
	for i, container := range dep.Spec.Template.Spec.Containers {
		if container.Name == k.container {
			return k.patchDeployment(getContainerResourcesPatch(i, mergeResources(&container.Resources, resources)))
		}
	}

	return fmt.Errorf("Container %s was not found in the deployment %s in namespace %s.", k.container, k.deployment, k.namespace)
}

func (k *kubernetesClient) patchDeployment(patch patchRecord) error {
	bytes, err := json.Marshal([]patchRecord{patch})
	if err != nil {
		return fmt.Errorf("Cannot marshal deployment patch %+v. Reason: %+v", patch, err)
	}

	_, err = k.clientset.AppsV1().Deployments(k.namespace).Patch(context.Background(), k.deployment, types.JSONPatchType, bytes, metav1.PatchOptions{})
	return err
}

type patchRecord struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

func getContainerResourcesPatch(index int, resources *corev1.ResourceRequirements) patchRecord {
	return patchRecord{
		Op:    "add",
		Path:  fmt.Sprintf("/spec/template/spec/containers/%d/resources", index),
		Value: *resources,
	}
}

func mergeResources(current, new *corev1.ResourceRequirements) *corev1.ResourceRequirements {
	res := current.DeepCopy()
	if res.Limits == nil {
		res.Limits = corev1.ResourceList{}
	}
	if res.Requests == nil {
		res.Requests = corev1.ResourceList{}
	}
	for resource, value := range new.Limits {
		res.Limits[resource] = value
	}
	for resource, value := range new.Requests {
		res.Requests[resource] = value
	}
	return res
}

// NewKubernetesClient gives a KubernetesClient with the given dependencies.
func NewKubernetesClient(namespace, deployment, pod, container string, clientset *kubernetes.Clientset, useMetrics bool) KubernetesClient {
	result := &kubernetesClient{
		namespace:  namespace,
		deployment: deployment,
		pod:        pod,
		container:  container,
		clientset:  clientset,
		useMetrics: useMetrics,
	}
	return result
}
