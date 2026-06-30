/*
Copyright The Kubernetes Authors.

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

package e2e

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestClusterAutoscaling(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fake-pod",
			Namespace: "default",
			Labels: map[string]string{
				"app": "fake-pod",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "fake-container",
					Image: "fake-image",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("100Mi"),
						},
					},
				},
			},
			NodeSelector: map[string]string{
				"kwok-nodegroup": "kind-worker",
			},
			Tolerations: []corev1.Toleration{
				{
					Key:      "kwok-provider",
					Operator: corev1.TolerationOpExists,
					Effect:   corev1.TaintEffectNoSchedule,
				},
			},
		},
	}

	scaleUpFeature := features.New("Cluster Autoscaler Scale Up").
		Assess("scale up when a pod is pending", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			client, err := cfg.NewClient()
			if err != nil {
				t.Fatal(err)
			}

			// Create the pending pod
			err = client.Resources().Create(ctx, pod)
			if err != nil {
				t.Fatalf("failed to create pod: %v", err)
			}

			// Wait for TriggeredScaleUp event
			err = wait.For(func(ctx context.Context) (done bool, err error) {
				events := &corev1.EventList{}
				err = client.Resources(pod.Namespace).List(ctx, events)
				if err != nil {
					return false, err
				}
				for _, event := range events.Items {
					if event.InvolvedObject.Name == pod.Name && event.Reason == "TriggeredScaleUp" {
						return true, nil
					}
				}
				return false, nil
			}, wait.WithTimeout(2*time.Minute), wait.WithContext(ctx))
			if err != nil {
				t.Fatalf("TriggeredScaleUp event not found: %v", err)
			}

			// Wait for the pod to be scheduled
			err = wait.For(conditions.New(client.Resources()).ResourceMatch(pod, func(object k8s.Object) bool {
				p := object.(*corev1.Pod)
				return p.Spec.NodeName != ""
			}), wait.WithTimeout(2*time.Minute), wait.WithContext(ctx))
			if err != nil {
				t.Fatalf("pod not scheduled: %v", err)
			}

			// Verify new node is created
			nodeList := &corev1.NodeList{}
			err = wait.For(func(ctx context.Context) (done bool, err error) {
				err = client.Resources().List(ctx, nodeList)
				if err != nil {
					return false, err
				}
				for _, node := range nodeList.Items {
					if node.Labels["kwok-nodegroup"] == "kind-worker" {
						return true, nil
					}
				}
				return false, nil
			}, wait.WithTimeout(2*time.Minute), wait.WithContext(ctx))
			if err != nil {
				t.Fatalf("kind-worker node not created: %v", err)
			}

			return ctx
		}).
		Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			client, err := cfg.NewClient()
			if err != nil {
				t.Fatal(err)
			}
			// Delete the pod
			_ = client.Resources().Delete(ctx, pod)
			return ctx
		}).
		Feature()

	testEnv.Test(t, scaleUpFeature)
}
