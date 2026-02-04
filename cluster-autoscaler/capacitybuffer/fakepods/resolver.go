/*
Copyright 2025 The Kubernetes Authors.

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

package fakepods

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	podutils "k8s.io/autoscaler/cluster-autoscaler/utils/pod"
)

// Resolver creates a fully defaulted and validated pod using server-side dry-run.
type Resolver interface {
	// Resolve returns a fake pod based on the provided template.
	// It performs a dry-run create request to the API server.
	Resolve(ctx context.Context, namespace string, template *corev1.PodTemplateSpec) (*corev1.Pod, error)
}

type fakePodResolver struct {
	client kubernetes.Interface
}

// NewResolver returns a new instance of Resolver.
func NewResolver(client kubernetes.Interface) Resolver {
	return &fakePodResolver{
		client: client,
	}
}

func (r *fakePodResolver) Resolve(ctx context.Context, namespace string, template *corev1.PodTemplateSpec) (*corev1.Pod, error) {
	pod := podutils.GetPodFromTemplate(template)
	pod.GenerateName = "fake-pod-"

	createdPod, err := r.client.CoreV1().Pods(namespace).Create(ctx, pod, metav1.CreateOptions{
		DryRun: []string{metav1.DryRunAll},
	})
	if err != nil {
		klog.Errorf("Failed to create dry-run pod for template %s/%s: %v", template.Namespace, template.Name, err)
		return nil, err
	}

	return createdPod, nil
}
