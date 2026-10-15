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
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	scheme "k8s.io/client-go/kubernetes/scheme"
)

const (
	// resourceObjectsMetricName is the metric with the number of pods and nodes (Kubernetes >=1.34 and higher, with group and resource labels)
	resourceObjectsMetricName = "apiserver_resource_objects"

	// objectCountMetricName is the metric with the number of pods and nodes (Kubernetes 1.21-1.36, with resource label)
	objectCountMetricName = "apiserver_storage_objects"

	// objectCountFallbackMetricName is the metric with the number of pods and nodes (Kubernetes <1.21)
	objectCountFallbackMetricName = "etcd_object_counts"

	fmtText expfmt.Format = `text/plain; version=` + expfmt.TextVersion + `; charset=utf-8`
)

var (
	// nodeResourceLabels are the label values for nodes, that must match if present in the metric.
	nodeResourceLabels = map[string]string{"group": "", "resource": "nodes"}
	// podResourceLabels are the label values for pods, that must match if present in the metric.
	podResourceLabels = map[string]string{"group": "", "resource": "pods"}
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
		return k.countResourcesThroughMetrics(nodeResourceLabels)
	}
	return k.countNodesThroughAPI()
}

// CountContainers returns the number of containers in the cluster:
// 1) by listing Pods (excluding terminated pods) using API (default) and calculating the sum of initContainers, ephemeralContainers, and containers.
// 2) using etcd_object_count metric exposed by kube-apiserver
func (k *kubernetesClient) CountContainers() (uint64, error) {
	if k.useMetrics {
		return k.countResourcesThroughMetrics(podResourceLabels)
	}
	return k.countContainersThroughAPI()
}

func (k *kubernetesClient) countContainersThroughAPI() (uint64, error) {
	// Set ResourceVersion = 0 to use cached versions.
	options := metav1.ListOptions{
		ResourceVersion: "0",
		FieldSelector:   getPodSelectorExcludingDonePodsOrDie(),
	}
	pods, err := k.clientset.CoreV1().Pods(corev1.NamespaceAll).List(context.TODO(), options)
	result := 0
	for _, pod := range pods.Items {
		result += len(pod.Spec.Containers) + len(pod.Spec.InitContainers) + len(pod.Spec.EphemeralContainers)
	}
	return uint64(result), err
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

func extractMetricValueForResourceCount(mf *dto.MetricFamily, matchLabels map[string]string, metricName string) (uint64, error) {
	for _, metric := range mf.Metric {
		matchesExpectedLabel := false
		mismatchesExpectedLabel := false
		for _, label := range metric.Label {
			if expectValue, ok := matchLabels[label.GetName()]; ok {
				if label.GetValue() == expectValue {
					matchesExpectedLabel = true
				} else {
					mismatchesExpectedLabel = true
				}
			}
		}
		if !matchesExpectedLabel || mismatchesExpectedLabel {
			continue
		}
		if metric.Gauge == nil || metric.Gauge.Value == nil {
			continue
		}
		if *metric.Gauge.Value < 0 {
			return 0, fmt.Errorf("%s: metric unknown", metricName)
		}

		value := uint64(*metric.Gauge.Value)
		return value, nil
	}
	return 0, fmt.Errorf("%s: no valid metric values", metricName)
}

func getResourceCountFromDecoder(matchLabels map[string]string, decoder expfmt.Decoder) (uint64, error) {
	var resourceObjectsMetricValue, deprecatedMetricValue, fallbackMetricValue uint64
	var resourceObjectsMetricError, deprecatedMetricError, fallbackMetricError error
	gotResourceObjectsMetric, gotDeprecatedMetric, gotFallbackMetric := false, false, false

	for {
		var mf dto.MetricFamily
		if err := decoder.Decode(&mf); err != nil {
			if err == io.EOF {
				break
			}
			return 0, fmt.Errorf("decoding error: %v", err)
		}
		if hasEqualValues(resourceObjectsMetricName, mf.Name) {
			resourceObjectsMetricValue, resourceObjectsMetricError = extractMetricValueForResourceCount(&mf, matchLabels, mf.GetName())
			gotResourceObjectsMetric = true
		}
		if hasEqualValues(objectCountMetricName, mf.Name) {
			deprecatedMetricValue, deprecatedMetricError = extractMetricValueForResourceCount(&mf, matchLabels, mf.GetName())
			gotDeprecatedMetric = true
		}
		if hasEqualValues(objectCountFallbackMetricName, mf.Name) {
			fallbackMetricValue, fallbackMetricError = extractMetricValueForResourceCount(&mf, matchLabels, mf.GetName())
			gotFallbackMetric = true
		}
	}

	if gotResourceObjectsMetric && resourceObjectsMetricError == nil {
		return resourceObjectsMetricValue, nil
	}
	if gotDeprecatedMetric && deprecatedMetricError == nil {
		return deprecatedMetricValue, nil
	}
	if gotFallbackMetric && fallbackMetricError == nil {
		return fallbackMetricValue, nil
	}
	if gotFallbackMetric || gotDeprecatedMetric || gotResourceObjectsMetric {
		return 0, fmt.Errorf("at least one metric present but all present metrics have errors: %v, %v, %v", resourceObjectsMetricError, deprecatedMetricError, fallbackMetricError)
	}
	return 0, fmt.Errorf("no metric set")
}

func (k *kubernetesClient) countResourcesThroughMetrics(matchLabels map[string]string) (uint64, error) {
	// Similarly as for listing resources, permissions for /metrics endpoint are needed.
	// Other than that, endpoint is visible from everywhere.
	reader, err := k.clientset.CoreV1().RESTClient().Get().RequestURI("/metrics").Stream(context.Background())
	if err != nil {
		return 0, err
	}
	return getResourceCountFromDecoder(matchLabels, expfmt.NewDecoder(reader, fmtText))
}

func (k *kubernetesClient) ContainerResources() (*corev1.ResourceRequirements, error) {
	var containers []corev1.Container
	// When addon resizer runs in a different pod, it cannot get pod name from env,
	// so the pod name will be empty. In that case, get container information
	// from the deployment instead.
	if k.pod == "" {
		dep, err := k.clientset.AppsV1().Deployments(k.namespace).Get(context.Background(), k.deployment, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		containers = dep.Spec.Template.Spec.Containers
	} else {
		pod, err := k.clientset.CoreV1().Pods(k.namespace).Get(context.Background(), k.pod, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		containers = pod.Spec.Containers
	}
	for _, container := range containers {
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

func getPodSelectorExcludingDonePodsOrDie() string {
	stringSelector := "status.phase!=" + string(corev1.PodSucceeded) +
		",status.phase!=" + string(corev1.PodFailed)
	selector := fields.ParseSelectorOrDie(stringSelector)
	return selector.String()
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
