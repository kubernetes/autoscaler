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

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1alpha1"
	fakebuffers "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/client/clientset/versioned/fake"
	cbclient "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/client"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/common"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/testutil"
	fakek8s "k8s.io/client-go/kubernetes/fake"
	testing2 "k8s.io/client-go/testing"
)

func TestControllerIntegration(t *testing.T) {
	// Setup
	podTemp := &corev1.PodTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "podTemp",
			Namespace: "default",
		},
		Template: corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								"cpu": resource.MustParse("1"),
							},
						},
					},
				},
			},
		},
	}

	rq := &corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "quota",
			Namespace: "default",
		},
		Spec: corev1.ResourceQuotaSpec{
			Hard: corev1.ResourceList{
				"cpu": resource.MustParse("5"),
			},
		},
		Status: corev1.ResourceQuotaStatus{
			Hard: corev1.ResourceList{
				"cpu": resource.MustParse("5"),
			},
			Used: corev1.ResourceList{
				"cpu": resource.MustParse("0"),
			},
		},
	}

	k8sClient := fakek8s.NewSimpleClientset(podTemp, rq)
	buffersClient := fakebuffers.NewSimpleClientset()

	buffersClient.PrependReactor("update", "capacitybuffers/status", func(action testing2.Action) (bool, runtime.Object, error) {
		updateAction := action.(testing2.UpdateAction)
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
		newObj.Status = buffer.Status
		err = tracker.Update(v1.SchemeGroupVersion.WithResource("capacitybuffers"), newObj, buffer.Namespace)
		return true, newObj, err
	})

	// We need to use NewCapacityBufferClientFromClients to initialize informers correctly
	client, err := cbclient.NewCapacityBufferClientFromClients(buffersClient, k8sClient, nil, nil)
	assert.NoError(t, err)

	controller := NewDefaultBufferController(client).(*bufferController)
	// Override loop interval for faster tests
	controller.loopInterval = 100 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run controller in background
	go controller.Run(ctx.Done())

	// Helper to wait for buffer status
	waitForStatus := func(name string, expectedReplicas int32, checkLimited bool) {
		err := wait.PollImmediate(100*time.Millisecond, 5*time.Second, func() (bool, error) {
			b, err := buffersClient.AutoscalingV1alpha1().CapacityBuffers("default").Get(context.TODO(), name, metav1.GetOptions{})
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
					if c.Type == common.LimitedByQuotasCondition && c.Status == common.ConditionTrue {
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
			b, _ := buffersClient.AutoscalingV1alpha1().CapacityBuffers("default").Get(context.TODO(), name, metav1.GetOptions{})
			var gotLimited bool
			for _, c := range b.Status.Conditions {
				if c.Type == common.LimitedByQuotasCondition && c.Status == common.ConditionTrue {
					gotLimited = true
				}
			}
			t.Errorf("%s reconciliation failed, got replicas: %d, limited: %t, want replicas: %d, limited: %t", name, *b.Status.Replicas, gotLimited, expectedReplicas, checkLimited)
		}
	}

	// 1. Create a buffer that fits
	activeStrategy := common.ActiveProvisioningStrategy
	b1 := testutil.NewBuffer(
		testutil.WithStatusPodTemplateRef("podTemp"), // Spec reference actually
		func(b *v1.CapacityBuffer) {
			b.Name = "b1"
			b.Spec.PodTemplateRef = &v1.LocalObjectRef{Name: "podTemp"}
			b.Spec.Replicas = int32Ptr(2)
			b.Spec.ProvisioningStrategy = &activeStrategy
		},
	)
	_, err = buffersClient.AutoscalingV1alpha1().CapacityBuffers("default").Create(context.TODO(), b1, metav1.CreateOptions{})
	assert.NoError(t, err)

	// Wait for reconciliation: Should get 2 replicas
	waitForStatus("b1", 2, false)

	// 2. Create a buffer that exceeds remaining quota (Quota=5, Used=2, Remaining=3. Request=4)
	b2 := testutil.NewBuffer(
		func(b *v1.CapacityBuffer) {
			b.Name = "b2"
			b.Spec.PodTemplateRef = &v1.LocalObjectRef{Name: "podTemp"}
			b.Spec.Replicas = int32Ptr(4)
			b.Spec.ProvisioningStrategy = &activeStrategy
		},
	)
	_, err = buffersClient.AutoscalingV1alpha1().CapacityBuffers("default").Create(context.TODO(), b2, metav1.CreateOptions{})
	assert.NoError(t, err)

	// Wait for reconciliation: Should get 3 replicas (limited) and have condition
	waitForStatus("b2", 3, true)

	// 3. Update b1 to use more resources (2 -> 4).
	// New usage: b1=4. Remaining=1. b2 (req 4) should shrink to 1.
	// This verifies the ripple effect.
	b1, _ = buffersClient.AutoscalingV1alpha1().CapacityBuffers("default").Get(context.TODO(), "b1", metav1.GetOptions{})
	b1.Spec.Replicas = int32Ptr(4)
	t.Logf("Updating buffer %s to %d replicas", b1.Name, *b1.Spec.Replicas)
	_, err = buffersClient.AutoscalingV1alpha1().CapacityBuffers("default").Update(context.TODO(), b1, metav1.UpdateOptions{})
	assert.NoError(t, err)

	waitForStatus("b1", 4, false)
	waitForStatus("b2", 1, true)

	// 4. Update Quota to be larger (5 -> 10).
	// b1 (4) + b2 (4) = 8. Both should fit.
	rq, _ = k8sClient.CoreV1().ResourceQuotas("default").Get(context.TODO(), "quota", metav1.GetOptions{})
	rq.Spec.Hard["cpu"] = resource.MustParse("10")
	// The Translator looks at `quota.Status.Hard` which we must update manually in fake client
	rq.Status.Hard["cpu"] = resource.MustParse("10")
	_, err = k8sClient.CoreV1().ResourceQuotas("default").UpdateStatus(context.TODO(), rq, metav1.UpdateOptions{})
	assert.NoError(t, err)

	// Since we updated Status, the Informer should catch it and trigger reconcile.
	waitForStatus("b1", 4, false)
	waitForStatus("b2", 4, false) // Should no longer be limited
}

func int32Ptr(i int32) *int32 {
	return &i
}
