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
	"fmt"
	"io"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/client-go/kubernetes"
	scheme "k8s.io/client-go/kubernetes/scheme"
)

const (
	// objectCountMetricName is the metric to be used to get number of nodes
	objectCountMetricName = "etcd_object_counts"
	// nodeResourceName is the label value for Nodes in objectCountMetricName metric.
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
		Core().
		RESTClient().
		Get().
		Resource("nodes").
		// Set as=PartialObjectMetadataList to fetch only nodes metadata.
		SetHeader("Accept", "application/vnd.kubernetes.protobuf;as=PartialObjectMetadataList;g=meta.k8s.io;v=v1beta1").
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return uint64(len(result.Items)), err
}

func hasEqualValues(a string, b *string) bool {
	return b != nil && a == *b
}

func (k *kubernetesClient) countNodesThroughMetrics() (uint64, error) {
	// Similarly as for listing nodes, permissions for /metrics endpoint are needed.
	// Other than that, endpoint is visible from everywhere.
	reader, err := k.clientset.Core().RESTClient().Get().RequestURI("/metrics").Stream()
	if err != nil {
		return 0, err
	}

	decoder := expfmt.NewDecoder(reader, expfmt.FmtText)

	var mf dto.MetricFamily
	for {
		if err := decoder.Decode(&mf); err != nil {
			if err == io.EOF {
				break
			}
			continue
		}

		if !hasEqualValues(objectCountMetricName, mf.Name) {
			continue
		}
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
			value := uint64(*metric.Gauge.Value)
			if value < 0 {
				return 0, fmt.Errorf("metric unknown")
			}
			return value, nil

		}
	}

	return 0, fmt.Errorf("metric unset")
}

func (k *kubernetesClient) ContainerResources() (*corev1.ResourceRequirements, error) {
	pod, err := k.clientset.Core().Pods(k.namespace).Get(k.pod, metav1.GetOptions{})
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
	dep, err := k.clientset.AppsV1().Deployments(k.namespace).Get(k.deployment, metav1.GetOptions{})
	if err != nil {
		return err
	}

	// Modify the Deployment object with our ResourceRequirements.
	for i, container := range dep.Spec.Template.Spec.Containers {
		if container.Name == k.container {
			// Update the deployment.
			dep.Spec.Template.Spec.Containers[i].Resources = *resources
			_, err = k.clientset.AppsV1().Deployments(k.namespace).Update(dep)
			return err
		}
	}

	return fmt.Errorf("Container %s was not found in the deployment %s in namespace %s.", k.container, k.deployment, k.namespace)
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
