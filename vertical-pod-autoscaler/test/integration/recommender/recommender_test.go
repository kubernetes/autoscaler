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

package recommender

import (
	"context"
	"testing"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/kubernetes/test/integration/framework"

	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	recommender_config "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/config"
)

/*
Tests in this file are for illustrative purposes only.
Once we start writing integreation tests, we can build out the nessesary scaffolding to support them, possibly even reusing the e2e scaffolding
*/

func TestRemovingGC(t *testing.T) {
	t.Parallel()
	recommenderConfig := recommender_config.DefaultRecommenderConfig()
	recommenderConfig.CheckpointsGCInterval = 1 * time.Second // Short interval for testing

	tCtx, closeFn, rm, informers, c, vpaClient := recommenderSetup(t, recommenderConfig)
	defer closeFn()
	ns := framework.CreateNamespaceOrDie(c, "cleanup-orphaned-gc", t)
	defer framework.DeleteNamespaceOrDie(c, ns, t)
	stopControllers := runControllerAndInformers(tCtx, rm, informers)
	defer stopControllers()

	checkpointName := "test-checkpoint"
	checkpointObj := &vpav1.VerticalPodAutoscalerCheckpoint{
		TypeMeta: metav1.TypeMeta{
			Kind:       "VerticalPodAutoscalerCheckpoint",
			APIVersion: "autoscaling.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      checkpointName,
			Namespace: ns.Name,
		},
		Spec: vpav1.VerticalPodAutoscalerCheckpointSpec{
			VPAObjectName: "dummy-vpa",
			ContainerName: "dummy-container",
		},
	}

	// Create the checkpoint
	created, err := vpaClient.AutoscalingV1().VerticalPodAutoscalerCheckpoints(ns.Name).Create(tCtx, checkpointObj, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create VPA Checkpoint: %v", err)
	}
	t.Logf("Created VPA Checkpoint: %s", created.Name)

	// Verify it exists
	fetched, err := vpaClient.AutoscalingV1().VerticalPodAutoscalerCheckpoints(ns.Name).Get(tCtx, checkpointName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get VPA Checkpoint: %v", err)
	}
	if fetched.Name != checkpointName {
		t.Fatalf("Fetched checkpoint name mismatch: got %s, want %s", fetched.Name, checkpointName)
	}
	t.Logf("Verified VPA Checkpoint exists: %s", fetched.Name)

	// // Wait for deletion
	err = wait.PollUntilContextTimeout(tCtx, 1*time.Second, 30*time.Second, true, func(ctx context.Context) (done bool, err error) {
		cp, err := vpaClient.AutoscalingV1().VerticalPodAutoscalerCheckpoints(ns.Name).Get(ctx, checkpointName, metav1.GetOptions{})
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
	t.Logf("Confirmed VPA Checkpoint deletion: %s", checkpointName)

	list, err := vpaClient.AutoscalingV1().VerticalPodAutoscalerCheckpoints(ns.Name).List(tCtx, metav1.ListOptions{})
	if err != nil {
		t.Fatalf("Failed to list VPA Checkpoints: %v", err)
	}

	if len(list.Items) > 0 {
		t.Fatalf("Expected no remaining VPA Checkpoints, but found %d", len(list.Items))
	}
}
