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

package controller

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1beta1"
	fakebuffers "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/client/clientset/versioned/fake"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer"
	cbclient "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/client"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/testutil"
	fakek8s "k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
	"k8s.io/utils/ptr"
)

func TestControllerIntegration_ResourceQuotas(t *testing.T) {
	// TODO: refactor to ginkgo and envtest
	// Setup
	podTemp := testutil.NewPodTemplate(
		testutil.WithPodTemplateName("podTemp"),
		testutil.WithPodTemplateResources(corev1.ResourceList{"cpu": resource.MustParse("1")}, nil),
	)

	rq := testutil.NewResourceQuota(
		testutil.WithResourceQuotaName("quota"),
		testutil.WithResourceQuotaHard(corev1.ResourceList{"cpu": resource.MustParse("5")}),
		testutil.WithResourceQuotaUsed(corev1.ResourceList{"cpu": resource.MustParse("0")}),
	)

	k8sClient := fakek8s.NewSimpleClientset(podTemp, rq)
	buffersClient := fakebuffers.NewSimpleClientset()

	// Fake k8s client does not work well with .UpdateStatus(), which BufferController uses.
	// Instead of updating only the status as the real client does, it updates the entire object, which makes this test flaky.
	// This is a workaround mimicking the behavior of the real client.
	// Can be removed once we switch to controller-runtime and envtest (and to real clients).
	buffersClient.PrependReactor("update", "capacitybuffers/status", func(action k8stesting.Action) (bool, runtime.Object, error) {
		updateAction := action.(k8stesting.UpdateAction)
		buffer := updateAction.GetObject().(*v1.CapacityBuffer)
		tracker := buffersClient.Tracker()

		existingObj, err := tracker.Get(v1.SchemeGroupVersion.WithResource("capacitybuffers"), buffer.Namespace, buffer.Name)
		if err != nil {
			if errors.IsNotFound(err) {
				return false, nil, nil
			}
			return true, nil, err
		}
		existingBuf := existingObj.(*v1.CapacityBuffer)
		newObj := existingBuf.DeepCopy()
		newObj.ObjectMeta = buffer.ObjectMeta
		newObj.Status = buffer.Status
		err = tracker.Update(v1.SchemeGroupVersion.WithResource("capacitybuffers"), newObj, buffer.Namespace)
		return true, newObj, err
	})

	ensureResourceVersionUpdates(k8sClient)
	ensureResourceVersionUpdates(buffersClient)

	// We need to use NewCapacityBufferClientFromClients to initialize informers correctly
	client, err := cbclient.NewCapacityBufferClientFromClients(buffersClient, k8sClient, nil, nil)
	assert.NoError(t, err)

	resolver := testutil.NewFakeResolver()
	controller := NewDefaultBufferController(client, resolver).(*bufferController)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run controller in background
	go controller.Run(ctx.Done())

	// Helper to wait for buffer status
	waitForStatus := func(name string, expectedReplicas int32, checkLimited bool) {
		err := wait.PollUntilContextTimeout(ctx, 100*time.Millisecond, 5*time.Second, true, func(ctx context.Context) (bool, error) {
			b, err := buffersClient.AutoscalingV1beta1().CapacityBuffers("default").Get(ctx, name, metav1.GetOptions{})
			if err != nil {
				return false, nil
			}
			if b.Status.Replicas == nil {
				return false, nil
			}
			if *b.Status.Replicas != expectedReplicas {
				return false, nil
			}

			if checkLimited {
				hasLimit := false
				for _, c := range b.Status.Conditions {
					if c.Type == capacitybuffer.LimitedByQuotasCondition && c.Status == metav1.ConditionTrue {
						hasLimit = true
					}
				}
				// If we expect limited, verify the condition is present
				if !hasLimit {
					return false, nil
				}
			}
			return true, nil
		})
		if err != nil {
			b, _ := buffersClient.AutoscalingV1beta1().CapacityBuffers("default").Get(ctx, name, metav1.GetOptions{})
			var gotLimited bool
			for _, c := range b.Status.Conditions {
				if c.Type == capacitybuffer.LimitedByQuotasCondition && c.Status == metav1.ConditionTrue {
					gotLimited = true
				}
			}
			if b.Status.Replicas != nil {
				gotReplicas := *b.Status.Replicas
				t.Errorf("%s reconciliation failed, got replicas: %d, limited: %t, want replicas: %d, limited: %t", name, gotReplicas, gotLimited, expectedReplicas, checkLimited)
			} else {
				t.Errorf("%s reconciliation failed, got replicas: nil, want replicas: %d", name, expectedReplicas)
			}
		}
	}

	// 1. Create a buffer that fits
	b1 := testutil.NewBuffer(
		testutil.WithName("b1"),
		testutil.WithPodTemplateRef("podTemp"),
		testutil.WithReplicas(2),
		testutil.WithActiveProvisioningStrategy(),
	)
	_, err = buffersClient.AutoscalingV1beta1().CapacityBuffers("default").Create(ctx, b1, metav1.CreateOptions{})
	assert.NoError(t, err)

	// Wait for reconciliation: Should get 2 replicas
	waitForStatus("b1", 2, false)

	// 2. Create a buffer that exceeds remaining quota (Quota=5, Used=2, Remaining=3. Request=4)
	b2 := testutil.NewBuffer(
		testutil.WithName("b2"),
		testutil.WithPodTemplateRef("podTemp"),
		testutil.WithReplicas(4),
		testutil.WithActiveProvisioningStrategy(),
	)
	_, err = buffersClient.AutoscalingV1beta1().CapacityBuffers("default").Create(ctx, b2, metav1.CreateOptions{})
	assert.NoError(t, err)

	// Wait for reconciliation: Should get 3 replicas (limited) and have condition
	waitForStatus("b2", 3, true)

	// 3. Update b1 to use more resources (2 -> 4).
	// New usage: b1=4. Remaining=1. b2 (req 4) should shrink to 1.
	b1, _ = buffersClient.AutoscalingV1beta1().CapacityBuffers("default").Get(ctx, "b1", metav1.GetOptions{})
	b1.Spec.Replicas = ptr.To[int32](4)
	_, err = updateBuffer(ctx, buffersClient, b1)
	assert.NoError(t, err)

	waitForStatus("b1", 4, false)
	waitForStatus("b2", 1, true)

	// 4. Update Quota to be larger (5 -> 10).
	// b1 (4) + b2 (4) = 8. Both should fit.
	rq, _ = k8sClient.CoreV1().ResourceQuotas("default").Get(ctx, "quota", metav1.GetOptions{})
	rq.Spec.Hard["cpu"] = resource.MustParse("10")
	// The Allocator looks at `quota.Status.Hard` which we must update manually in fake client
	rq.Status.Hard["cpu"] = resource.MustParse("10")
	_, err = k8sClient.CoreV1().ResourceQuotas("default").UpdateStatus(ctx, rq, metav1.UpdateOptions{})
	assert.NoError(t, err)

	// Since we updated Status, the Informer should catch it and trigger reconcile.
	waitForStatus("b1", 4, false)
	waitForStatus("b2", 4, false) // Should no longer be limited
}

type withReactors interface {
	PrependReactor(verb, resource string, reaction k8stesting.ReactionFunc)
}

// ensureResourceVersionUpdates prepends reactor to the clients to ensure that
// resourceVersion is bumped on every object update. This is needed because fake client
// do not do this properly. This function can be removed once we migrate to controller-runtime and envtest.
func ensureResourceVersionUpdates(client withReactors) {
	client.PrependReactor("update", "*", func(action k8stesting.Action) (bool, runtime.Object, error) {
		updateAction := action.(k8stesting.UpdateAction)
		obj := updateAction.GetObject()

		acc, _ := meta.Accessor(obj)
		acc.SetResourceVersion(uuid.New().String())

		return false, obj, nil
	})
}

func updateBuffer(ctx context.Context, client *fakebuffers.Clientset, buffer *v1.CapacityBuffer) (*v1.CapacityBuffer, error) {
	// fake clients do not bump Generation when spec is changed. We need to do it manually.
	// can be removed once we migrate to controller-runtime and envtest.
	buffer.Generation += 1
	return client.AutoscalingV1beta1().CapacityBuffers("default").Update(ctx, buffer, metav1.UpdateOptions{})
}
