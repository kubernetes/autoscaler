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
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestScaleDownUnneededNode(t *testing.T) {
	t.Parallel()

	newPod := func(namespace string) *corev1.Pod {
		return NewTestPod("scaledown-test-pod", namespace, "kind-worker2", "500m", "500Mi")
	}

	feature := features.New("Scale Down Unneeded Node").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			client, err := cfg.NewClient()
			if err != nil {
				t.Fatal(err)
			}
			ns := cfg.Namespace()
			_ = client.Resources().Delete(ctx, newPod(ns))
			_ = CleanUpNodeGroup(ctx, client, "kind-worker2")
			return ctx
		}).
		Assess("scale down empty node after pod deletion", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			client, err := cfg.NewClient()
			if err != nil {
				t.Fatal(err)
			}

			ns := cfg.Namespace()
			pod := newPod(ns)

			// Step 1: Create pod to trigger scale up
			err = client.Resources().Create(ctx, pod)
			if err != nil {
				t.Fatalf("failed to create pod: %v", err)
			}

			err = WaitForPodScheduled(ctx, client, pod, defaultPodTimeout)
			if err != nil {
				t.Fatalf("pod not scheduled: %v", err)
			}

			// Wait for node count to increase to 1 and be Ready
			err = WaitForNodesReady(ctx, client, "kind-worker2", 1, defaultNodeTimeout)
			if err != nil {
				t.Fatalf("node did not become ready: %v", err)
			}

			// Step 2: Delete pod to make node unneeded
			err = client.Resources().Delete(ctx, pod)
			if err != nil {
				t.Fatalf("failed to delete pod: %v", err)
			}
			_ = WaitForPodDeleted(ctx, client, pod, defaultPodTimeout)

			// Step 3: Wait for scale down to delete the unneeded node back to 0
			err = WaitForNodeCount(ctx, client, "kind-worker2", 0, defaultScaleDownTimeout)
			if err != nil {
				t.Fatalf("node was not scaled down after pod deletion: %v", err)
			}

			return ctx
		}).
		Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			client, err := cfg.NewClient()
			if err != nil {
				t.Fatal(err)
			}
			ns := cfg.Namespace()
			TeardownPodAndNodeGroup(ctx, client, []*corev1.Pod{newPod(ns)}, "kind-worker2")
			return ctx
		}).
		Feature()

	testEnv.Test(t, feature)
}

func TestScaleDownWithPDB(t *testing.T) {
	newPod := func(namespace string) *corev1.Pod {
		pod := NewTestPod("pdb-test-pod", namespace, "kind-worker", "500m", "500Mi")
		pod.Labels["app"] = "pdb-test"
		return pod
	}

	newPDB := func(namespace string) *policyv1.PodDisruptionBudget {
		minAvailable := intstr.FromInt(1)
		return NewTestPDB("pdb-test-budget", namespace, "app", "pdb-test", &minAvailable, nil)
	}

	feature := features.New("Scale Down With PDB Protection").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			client, err := cfg.NewClient()
			if err != nil {
				t.Fatal(err)
			}
			ns := cfg.Namespace()
			_ = client.Resources().Delete(ctx, newPDB(ns))
			_ = client.Resources().Delete(ctx, newPod(ns))
			_ = CleanUpNodeGroup(ctx, client, "kind-worker")
			return ctx
		}).
		Assess("scale down respects PDB rules", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			client, err := cfg.NewClient()
			if err != nil {
				t.Fatal(err)
			}

			ns := cfg.Namespace()
			pod := newPod(ns)
			pdb := newPDB(ns)

			// Create PDB with minAvailable=1
			err = client.Resources().Create(ctx, pdb)
			if err != nil {
				t.Fatalf("failed to create PDB: %v", err)
			}

			// Create pod matching PDB
			err = client.Resources().Create(ctx, pod)
			if err != nil {
				t.Fatalf("failed to create pod: %v", err)
			}

			err = WaitForPodScheduled(ctx, client, pod, defaultPodTimeout)
			if err != nil {
				t.Fatalf("pod not scheduled: %v", err)
			}

			err = WaitForNodesAtLeast(ctx, client, "kind-worker", 1, defaultNodeTimeout)
			if err != nil {
				t.Fatalf("node count did not increase: %v", err)
			}

			// Verify that while pod is active and protected by PDB, node is retained
			time.Sleep(15 * time.Second)

			count, err := CountNodegroupNodes(ctx, client, "kind-worker")
			if err != nil {
				t.Fatalf("failed to count nodes: %v", err)
			}
			if count < 1 {
				t.Fatalf("node unexpectedly scaled down despite active PDB: %d", count)
			}

			return ctx
		}).
		Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			client, err := cfg.NewClient()
			if err != nil {
				t.Fatal(err)
			}
			ns := cfg.Namespace()
			_ = client.Resources().Delete(ctx, newPDB(ns))
			TeardownPodAndNodeGroup(ctx, client, []*corev1.Pod{newPod(ns)}, "kind-worker")
			return ctx
		}).
		Feature()

	testEnv.Test(t, feature)
}
