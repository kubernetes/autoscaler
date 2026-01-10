//go:build integration
// +build integration

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

package integration

import (
	"context"
	"testing"
	"time"

	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/autoscaler/vertical-pod-autoscaler/common"
	"k8s.io/autoscaler/vertical-pod-autoscaler/e2e/utils"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/app"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
)

func TestRecommenderWithNamespaceFiltering(t *testing.T) {
	ctx := t.Context()

	// Create test namespaces
	watchedNS := "ns-filtering-watched"
	ignoredNS := "ns-filtering-ignored"

	for _, ns := range []string{watchedNS, ignoredNS} {
		_, err := kubeClient.CoreV1().Namespaces().Create(ctx, &apiv1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: ns},
		}, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("Failed to create namespace %s: %v", ns, err)
		}
		defer func(ns string) {
			_ = kubeClient.CoreV1().Namespaces().Delete(ctx, ns, metav1.DeleteOptions{})
		}(ns)
	}

	// Create VPA objects in both namespaces
	for _, ns := range []string{watchedNS, ignoredNS} {
		vpa := test.VerticalPodAutoscaler().
			WithName("test-vpa").
			WithContainer("hamster").
			WithNamespace(ns).
			WithTargetRef(utils.HamsterTargetRef).
			Get()

		_, err := vpaClient.AutoscalingV1().VerticalPodAutoscalers(ns).Create(ctx, vpa, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("Failed to create VPA in namespace %s: %v", ns, err)
		}
	}

	// Configure the recommender to watch only the watched namespace
	config := app.DefaultRecommenderConfig()
	config.CommonFlags = &common.CommonFlags{
		KubeConfig:                 kubeconfig,
		VpaObjectNamespace:         watchedNS, // Only watch the watched namespace
		IgnoredVpaObjectNamespaces: "",
	}
	config.MetricsFetcherInterval = 1 * time.Second // Short interval for testing

	_, cancel := startRecommender(t, config)
	defer cancel()

	// Wait for the recommender to process the VPA in the watched namespace.
	// The recommender should add status conditions to VPAs it manages.
	err := wait.PollUntilContextTimeout(ctx, 1*time.Second, 50*time.Second, true, func(ctx context.Context) (done bool, err error) {
		watchedVPA, err := vpaClient.AutoscalingV1().VerticalPodAutoscalers(watchedNS).Get(ctx, "test-vpa", metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		// watched namespace should have status updates
		if len(watchedVPA.Status.Conditions) > 0 {
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		t.Fatalf("VPA in watched namespace should have status conditions: %v", err)
	}

	// Fetch VPA in the ignored namespace.
	// The recommender should NOT have added a status conditions to this VPA.
	ignoredVPA, err := vpaClient.AutoscalingV1().VerticalPodAutoscalers(ignoredNS).Get(ctx, "test-vpa", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Unable to get VPA in ignored namespace: %v", err)
	}

	if len(ignoredVPA.Status.Conditions) != 0 {
		t.Fatal("VPA in ignored namespace should NOT have status conditions")
	}
}

func TestRecommenderWithNamespaceExclusions(t *testing.T) {
	ctx := t.Context()

	// Create test namespaces
	watchedNS := "ns-exclusions-watched"
	ignoredNS := "ns-exclusions-ignored"

	for _, ns := range []string{watchedNS, ignoredNS} {
		_, err := kubeClient.CoreV1().Namespaces().Create(ctx, &apiv1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: ns},
		}, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("Failed to create namespace %s: %v", ns, err)
		}
		defer func(ns string) {
			_ = kubeClient.CoreV1().Namespaces().Delete(ctx, ns, metav1.DeleteOptions{})
		}(ns)
	}

	// Create VPA objects in both namespaces
	for _, ns := range []string{watchedNS, ignoredNS} {
		vpa := test.VerticalPodAutoscaler().
			WithName("test-vpa").
			WithContainer("hamster").
			WithNamespace(ns).
			WithTargetRef(utils.HamsterTargetRef).
			Get()

		_, err := vpaClient.AutoscalingV1().VerticalPodAutoscalers(ns).Create(ctx, vpa, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("Failed to create VPA in namespace %s: %v", ns, err)
		}
	}

	// Configure the recommender to exclude the ignored namespace
	config := app.DefaultRecommenderConfig()
	config.CommonFlags = &common.CommonFlags{
		KubeConfig:                 kubeconfig,
		VpaObjectNamespace:         "", // Watch all namespaces
		IgnoredVpaObjectNamespaces: ignoredNS,
	}
	config.MetricsFetcherInterval = 1 * time.Second // Short interval for testing

	_, cancel := startRecommender(t, config)
	defer cancel()

	// Wait for the recommender to process the VPA in the watched namespace.
	// The recommender should add status conditions to VPAs it manages.
	err := wait.PollUntilContextTimeout(ctx, 1*time.Second, 50*time.Second, true, func(ctx context.Context) (done bool, err error) {
		watchedVPA, err := vpaClient.AutoscalingV1().VerticalPodAutoscalers(watchedNS).Get(ctx, "test-vpa", metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		// watched namespace should have status updates
		if len(watchedVPA.Status.Conditions) > 0 {
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		t.Fatalf("VPA in watched namespace should have status conditions: %v", err)
	}

	// Fetch VPA in the ignored namespace.
	// The recommender should NOT have added a status conditions to this VPA.
	ignoredVPA, err := vpaClient.AutoscalingV1().VerticalPodAutoscalers(ignoredNS).Get(ctx, "test-vpa", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Unable to get VPA in ignored namespace: %v", err)
	}

	if len(ignoredVPA.Status.Conditions) != 0 {
		t.Fatal("VPA in ignored namespace should NOT have status conditions")
	}
}

func TestCRDCheckpointGC(t *testing.T) {
	ctx := t.Context()

	ns := "checkpoint-gc-test"

	// Create test namespace
	_, err := kubeClient.CoreV1().Namespaces().Create(ctx, &apiv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: ns},
	}, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create namespace %s: %v", ns, err)
	}
	defer func(ns string) {
		_ = kubeClient.CoreV1().Namespaces().Delete(ctx, ns, metav1.DeleteOptions{})
	}(ns)

	// Create a Deployment that the VPA will target
	deploymentLabel := map[string]string{"app": "hamster"}
	deployment := newHamsterDeployment(ns, 1, deploymentLabel)

	_, err = kubeClient.AppsV1().Deployments(ns).Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create Deployment in namespace %s: %v", ns, err)
	}

	// Create Pods matching the deployment (simulating what the deployment controller would do)
	pod := test.Pod().
		WithName("hamster-pod-0").
		WithLabels(deploymentLabel).
		WithPhase(apiv1.PodRunning).
		AddContainer(test.Container().
			WithName("hamster").
			WithImage("busybox").
			WithCPURequest(resource.MustParse("100m")).
			WithMemRequest(resource.MustParse("50Mi")).
			Get()).
		AddContainerStatus(apiv1.ContainerStatus{
			Name:  "hamster",
			Ready: true,
			State: apiv1.ContainerState{
				Running: &apiv1.ContainerStateRunning{
					StartedAt: metav1.Now(),
				},
			},
		}).
		Get()
	pod.Namespace = ns

	createdPod, err := kubeClient.CoreV1().Pods(ns).Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create Pod in namespace %s: %v", ns, err)
	}

	// Since the API server ignores status on Create, we need to update it separately
	createdPod.Status = pod.Status
	_, err = kubeClient.CoreV1().Pods(ns).UpdateStatus(ctx, createdPod, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("Failed to update Pod status in namespace %s: %v", ns, err)
	}

	// Create the VPA targeting the deployment
	vpa := test.VerticalPodAutoscaler().
		WithName("test-vpa").
		WithNamespace(ns).
		WithContainer("hamster").
		WithTargetRef(utils.HamsterTargetRef).
		Get()

	_, err = vpaClient.AutoscalingV1().VerticalPodAutoscalers(ns).Create(ctx, vpa, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create VPA in namespace %s: %v", ns, err)
	}

	// Configure the recommender to watch only the watched namespace
	config := app.DefaultRecommenderConfig()
	config.CommonFlags = &common.CommonFlags{
		KubeConfig: kubeconfig,
	}
	config.MetricsFetcherInterval = 1 * time.Second // Short interval for testing
	config.CheckpointsGCInterval = 1 * time.Second  // Short interval for testing

	_, cancel := startRecommender(t, config)
	defer cancel()

	err = wait.PollUntilContextTimeout(ctx, 1*time.Second, 50*time.Second, true, func(ctx context.Context) (done bool, err error) {
		_, err = vpaClient.AutoscalingV1().VerticalPodAutoscalerCheckpoints(ns).Get(ctx, "test-vpa-hamster", metav1.GetOptions{})
		if err == nil {
			return true, nil // Checkpoint found
		}
		if apierrors.IsNotFound(err) {
			return false, nil // Not found yet, keep polling
		}
		return false, err // Real error, stop and fail
	})

	if err != nil {
		t.Fatalf("Timed out waiting for checkpoint to be created: %v", err)
	}

	err = vpaClient.AutoscalingV1().VerticalPodAutoscalers(ns).Delete(ctx, "test-vpa", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Failed to delete VPA: %v", err)
	}

	err = wait.PollUntilContextTimeout(ctx, 1*time.Second, 50*time.Second, true, func(ctx context.Context) (done bool, err error) {
		cp, err := vpaClient.AutoscalingV1().VerticalPodAutoscalerCheckpoints(ns).Get(ctx, "test-vpa-hamster", metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return true, nil // Checkpoint was deleted by GC - this is what we're testing for
			}
			return false, err // Some other error (e.g., server shutdown) - fail the test
		}
		t.Log("Found Checkpoint, still waiting for GC", cp.Name)
		return false, nil
	})

	if err != nil {
		t.Fatalf("Timed out waiting for VPA Checkpoint to be garbage collected: %v", err)
	}
}

func TestRecommenderName(t *testing.T) {
	ctx := t.Context()

	ns := "recommender-name-test"

	// Create test namespace
	_, err := kubeClient.CoreV1().Namespaces().Create(ctx, &apiv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: ns},
	}, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create namespace %s: %v", ns, err)
	}
	defer func(ns string) {
		_ = kubeClient.CoreV1().Namespaces().Delete(ctx, ns, metav1.DeleteOptions{})
	}(ns)

	// Create two VPAs that target different recommenders:
	// - "vpa-for-custom-recommender" uses recommender "custom-recommender"
	// - "vpa-for-default-recommender" uses recommender "default" (empty string means default)
	vpaCustom := test.VerticalPodAutoscaler().
		WithName("vpa-for-custom-recommender").
		WithContainer("hamster").
		WithNamespace(ns).
		WithRecommender("custom-recommender").
		WithTargetRef(utils.HamsterTargetRef).
		Get()

	vpaDefault := test.VerticalPodAutoscaler().
		WithName("vpa-for-default-recommender").
		WithContainer("hamster").
		WithNamespace(ns).
		WithTargetRef(utils.HamsterTargetRef). // No WithRecommender = uses default recommender
		Get()

	_, err = vpaClient.AutoscalingV1().VerticalPodAutoscalers(ns).Create(ctx, vpaCustom, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create VPA: %v", err)
	}

	_, err = vpaClient.AutoscalingV1().VerticalPodAutoscalers(ns).Create(ctx, vpaDefault, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create VPA: %v", err)
	}

	// Start a recommender named "custom-recommender"
	// It should only process VPAs that specify this recommender name
	config := app.DefaultRecommenderConfig()
	config.CommonFlags = &common.CommonFlags{
		KubeConfig: kubeconfig,
	}
	config.MetricsFetcherInterval = 1 * time.Second
	config.RecommenderName = "custom-recommender"

	_, cancel := startRecommender(t, config)
	defer cancel()

	// The VPA targeting "custom-recommender" should get status updates
	err = wait.PollUntilContextTimeout(ctx, 1*time.Second, 50*time.Second, true, func(ctx context.Context) (done bool, err error) {
		vpa, err := vpaClient.AutoscalingV1().VerticalPodAutoscalers(ns).Get(ctx, "vpa-for-custom-recommender", metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		if len(vpa.Status.Conditions) > 0 {
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		t.Fatalf("VPA targeting custom-recommender should have status conditions: %v", err)
	}

	vpa, err := vpaClient.AutoscalingV1().VerticalPodAutoscalers(ns).Get(ctx, "vpa-for-default-recommender", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Unable to get VPA for the default recommender: %v", err)
	}
	// We expect NO conditions - if we see any, the test should fail
	if len(vpa.Status.Conditions) > 0 {
		t.Fatal("VPA targeting default recommender should NOT have status conditions (custom-recommender should ignore it)")
	}
}
