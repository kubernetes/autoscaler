//go:build e2e

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
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	wait2 "k8s.io/apimachinery/pkg/util/wait"
	cqv1alpha1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacityquota/autoscaling.x-k8s.io/v1alpha1"
	cqtest "k8s.io/autoscaler/cluster-autoscaler/resourcequotas/capacityquota/testutil"
	"k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/autoscaler/cluster-autoscaler/utils/units"
	"sigs.k8s.io/e2e-framework/klient"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestCapacityQuotaAutoscaling(t *testing.T) {
	var cq *cqv1alpha1.CapacityQuota

	cqFeature := features.New("CapacityQuota").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			client := cfg.Client()
			cq = cqtest.NewCapacityQuota(
				"test-cq",
				cqtest.WithLabelSelector(map[string]string{"kwok-nodegroup": "kind-worker"}),
				cqtest.WithLimits(cqv1alpha1.ResourceList{
					cqv1alpha1.ResourceNodes: resource.MustParse("3"),
				}),
			)
			if err := client.Resources().Create(ctx, cq); err != nil {
				t.Fatalf("failed to create CapacityQuota: %v", err)
			}
			return ctx
		}).
		Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			namespace := cfg.Namespace()
			client := cfg.Client()

			// Delete pods
			podList := &corev1.PodList{}
			if err := client.Resources(namespace).List(ctx, podList); err == nil {
				for _, pod := range podList.Items {
					_ = client.Resources(namespace).Delete(ctx, &pod)
				}
			}

			// Delete CapacityQuotas
			cqList := &cqv1alpha1.CapacityQuotaList{}
			if err := client.Resources().List(ctx, cqList); err == nil {
				for _, cq := range cqList.Items {
					_ = client.Resources().Delete(ctx, &cq)
				}
			}

			// Delete nodes
			nodeList := &corev1.NodeList{}
			if err := client.Resources().List(ctx, nodeList); err == nil {
				for _, node := range nodeList.Items {
					if node.Labels["kwok-nodegroup"] == "kind-worker" {
						_ = client.Resources().Delete(ctx, &node)
					}
				}
			}

			return ctx
		}).
		Assess("scale up succeeds when quota is not exceeded", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			namespace := cfg.Namespace()
			client := cfg.Client()

			pod1 := buildTestPod("pod-1", namespace)
			pod2 := buildTestPod("pod-2", namespace)
			if err := client.Resources().Create(ctx, pod1); err != nil {
				t.Fatalf("failed to create pod 1: %v", err)
			}
			if err := client.Resources().Create(ctx, pod2); err != nil {
				t.Fatalf("failed to create pod 2: %v", err)
			}

			// Wait for both pods to be scheduled
			err := wait.For(allPodsScheduled(client, namespace), wait.WithTimeout(2*time.Minute), wait.WithContext(ctx))
			if err != nil {
				t.Errorf("pods not scheduled: %v", err)
			}

			// Verify 2 new nodes are created
			wantNodes := 2
			nodeList := &corev1.NodeList{}
			err = client.Resources().List(ctx, nodeList, resources.WithLabelSelector("kwok-nodegroup=kind-worker"))
			if err != nil {
				t.Fatalf("failed to list nodes: %v", err)
			}
			if gotNodes := len(nodeList.Items); gotNodes != wantNodes {
				t.Errorf("got %d kind-worker nodes, want %d", gotNodes, wantNodes)
			}

			// Verify CapacityQuota status used is updated
			wantUsage := cqv1alpha1.ResourceList{
				cqv1alpha1.ResourceNodes: resource.MustParse("2"),
			}
			err = wait.For(capacityQuotaUsageUpdated(client, cq, wantUsage), wait.WithTimeout(1*time.Minute), wait.WithContext(ctx))
			if err != nil {
				diff := cmp.Diff(wantUsage, cq.Status.Used.Resources)
				t.Errorf("CapacityQuota status not updated: %v, diff (-want +got):\n%s", err, diff)
			}

			return ctx
		}).
		Assess("limits scale up to 1 node for the next 2 pods", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			namespace := cfg.Namespace()
			client := cfg.Client()

			pod3 := buildTestPod("pod-3", namespace)
			pod4 := buildTestPod("pod-4", namespace)
			if err := client.Resources().Create(ctx, pod3); err != nil {
				t.Fatalf("failed to create pod 3: %v", err)
			}
			if err := client.Resources().Create(ctx, pod4); err != nil {
				t.Fatalf("failed to create pod 4: %v", err)
			}

			var scheduledCount int
			// Wait until exactly 1 pod is scheduled and 1 node is created
			err := wait.For(func(ctx context.Context) (done bool, err error) {
				p1 := &corev1.Pod{}
				p2 := &corev1.Pod{}
				_ = client.Resources().Get(ctx, pod3.Name, pod3.Namespace, p1)
				_ = client.Resources().Get(ctx, pod4.Name, pod4.Namespace, p2)
				scheduledCount = 0
				if p1.Spec.NodeName != "" {
					scheduledCount++
				}
				if p2.Spec.NodeName != "" {
					scheduledCount++
				}
				return scheduledCount == 1, nil
			}, wait.WithTimeout(2*time.Minute), wait.WithContext(ctx))
			if err != nil {
				t.Errorf("got %d scheduled pods, want 1", scheduledCount)
			}

			// Verify only 1 new node is created
			wantNodes := 3
			nodeList := &corev1.NodeList{}
			err = client.Resources().List(ctx, nodeList, resources.WithLabelSelector("kwok-nodegroup=kind-worker"))
			if err != nil {
				t.Fatalf("failed to list nodes: %v", err)
			}
			if gotNodes := len(nodeList.Items); gotNodes != wantNodes {
				t.Errorf("got %d kind-worker nodes, want %d", gotNodes, wantNodes)
			}

			return ctx
		}).
		Assess("no nodes created and pod emits NotTriggerScaleUp event when quota is exhausted", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			namespace := cfg.Namespace()
			client := cfg.Client()

			pod := buildTestPod("pod-5", namespace)
			if err := client.Resources().Create(ctx, pod); err != nil {
				t.Fatalf("failed to create pod-5: %v", err)
			}

			// Wait for NotTriggerScaleUp event with quota exceeded message
			err := wait.For(func(ctx context.Context) (done bool, err error) {
				events := &corev1.EventList{}
				if err := client.Resources(pod.Namespace).List(ctx, events); err != nil {
					return false, err
				}
				for _, event := range events.Items {
					if event.InvolvedObject.Name == pod.Name && event.Reason == "NotTriggerScaleUp" {
						if strings.Contains(event.Message, `exceeded quota: "CapacityQuota/test-cq"`) {
							return true, nil
						}
					}
				}
				return false, nil
			}, wait.WithTimeout(2*time.Minute), wait.WithContext(ctx))
			if err != nil {
				t.Errorf("NotTriggerScaleUp event with exceeded quota not found: %v", err)
			}

			// Verify 0 new nodes are created
			wantNodes := 3
			nodeList := &corev1.NodeList{}
			err = client.Resources().List(ctx, nodeList, resources.WithLabelSelector("kwok-nodegroup=kind-worker"))
			if err != nil {
				t.Fatalf("failed to list nodes: %v", err)
			}
			if gotNodes := len(nodeList.Items); gotNodes != wantNodes {
				t.Errorf("got %d kind-worker nodes, want %d", gotNodes, wantNodes)
			}

			return ctx
		}).
		Assess("all pods get scheduled once quota is increased", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			namespace := cfg.Namespace()
			client := cfg.Client()
			err := client.Resources().Get(ctx, "test-cq", "", cq)
			if err != nil {
				t.Fatalf("failed to get CapacityQuota: %v", err)
			}

			cq.Spec.Limits.Resources[cqv1alpha1.ResourceNodes] = resource.MustParse("8")
			err = client.Resources().Update(ctx, cq)
			if err != nil {
				t.Fatalf("failed to update CapacityQuota: %v", err)
			}

			// Wait for all pods to be scheduled
			err = wait.For(allPodsScheduled(client, namespace), wait.WithTimeout(2*time.Minute), wait.WithContext(ctx))
			if err != nil {
				t.Errorf("pods not scheduled: %v", err)
			}

			wantUsage := cqv1alpha1.ResourceList{
				cqv1alpha1.ResourceNodes: resource.MustParse("5"),
			}
			err = wait.For(capacityQuotaUsageUpdated(client, cq, wantUsage), wait.WithTimeout(1*time.Minute), wait.WithContext(ctx))
			if err != nil {
				diff := cmp.Diff(wantUsage, cq.Status.Used.Resources)
				t.Errorf("CapacityQuota status not updated: %v, diff (-want +got):\n%s", err, diff)
			}

			return ctx
		}).
		Feature()

	testEnv.Test(t, cqFeature)
}

func buildTestPod(name string, namespace string) *corev1.Pod {
	return test.BuildTestPod(
		name,
		100,
		100*units.MiB,
		test.WithNamespace(namespace),
		test.WithLabels(map[string]string{"app": "test"}),
		test.WithNodeSelector(map[string]string{"kwok-nodegroup": "kind-worker"}),
		test.WithPodHostnameAntiAffinity(map[string]string{"app": "test"}),
		withKwokToleration,
	)
}

func capacityQuotaUsageUpdated(client klient.Client, cq *cqv1alpha1.CapacityQuota, usage cqv1alpha1.ResourceList) wait2.ConditionWithContextFunc {
	return func(ctx context.Context) (done bool, err error) {
		if err := client.Resources().Get(ctx, cq.Name, cq.Namespace, cq); err != nil {
			return false, err
		}
		for res, q := range usage {
			if !cq.Status.Used.Resources[res].Equal(q) {
				return false, nil
			}
		}
		return true, nil
	}
}
