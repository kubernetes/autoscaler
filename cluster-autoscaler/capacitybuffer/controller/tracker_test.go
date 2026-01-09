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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1alpha1"
	fakebuffers "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/client/clientset/versioned/fake"
	cbclient "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/client"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/testutil"
	fakek8s "k8s.io/client-go/kubernetes/fake"
)

func TestDirtyNamespaceTracker(t *testing.T) {
	k8sClient := fakek8s.NewSimpleClientset()
	buffersClient := fakebuffers.NewSimpleClientset()
	client, err := cbclient.NewCapacityBufferClientFromClients(buffersClient, k8sClient, nil, nil)
	assert.NoError(t, err)

	tracker := NewDirtyNamespaceTracker(client)

	// Helper to wait for dirty namespace
	waitForDirty := func(namespace string) {
		err := wait.PollImmediate(10*time.Millisecond, 2*time.Second, func() (bool, error) {
			namespaces := tracker.GetAndClearDirtyNamespaces()
			for _, ns := range namespaces {
				if ns == namespace {
					return true, nil
				}
			}
			return false, nil
		})
		assert.NoError(t, err, "Timed out waiting for namespace %s to be dirty", namespace)
	}

	// Helper to ensure NO dirty namespace (for filtered updates)
	ensureNotDirty := func() {
		// Wait a bit to ensure event would have propagated if it wasn't filtered
		time.Sleep(100 * time.Millisecond)
		namespaces := tracker.GetAndClearDirtyNamespaces()
		assert.Empty(t, namespaces, "Expected no dirty namespaces")
	}

	t.Run("Buffer Add/Update/Delete", func(t *testing.T) {
		// Add
		b := testutil.NewBuffer(func(b *v1.CapacityBuffer) {
			b.Name = "b1"
			b.Namespace = "ns1"
		})
		_, err := buffersClient.AutoscalingV1alpha1().CapacityBuffers("ns1").Create(context.TODO(), b, metav1.CreateOptions{})
		assert.NoError(t, err)
		waitForDirty("ns1")

		// Update Status (Should be ignored)
		b.Status.Replicas = int32Ptr(5)
		_, err = buffersClient.AutoscalingV1alpha1().CapacityBuffers("ns1").Update(context.TODO(), b, metav1.UpdateOptions{})
		assert.NoError(t, err)
		ensureNotDirty()

		// Update Spec (Should trigger)
		b.Spec.Replicas = int32Ptr(10)
		_, err = buffersClient.AutoscalingV1alpha1().CapacityBuffers("ns1").Update(context.TODO(), b, metav1.UpdateOptions{})
		assert.NoError(t, err)
		waitForDirty("ns1")

		// Delete
		err = buffersClient.AutoscalingV1alpha1().CapacityBuffers("ns1").Delete(context.TODO(), "b1", metav1.DeleteOptions{})
		assert.NoError(t, err)
		waitForDirty("ns1")
	})

	t.Run("ResourceQuota Add/Update", func(t *testing.T) {
		rq := &corev1.ResourceQuota{
			ObjectMeta: metav1.ObjectMeta{Name: "rq1", Namespace: "ns2"},
			Spec:       corev1.ResourceQuotaSpec{Hard: corev1.ResourceList{"cpu": resource.MustParse("1")}},
		}
		// Add
		_, err := k8sClient.CoreV1().ResourceQuotas("ns2").Create(context.TODO(), rq, metav1.CreateOptions{})
		assert.NoError(t, err)
		waitForDirty("ns2")

		// Update Status (Should trigger, as Quota usage change is relevant)
		rq.Status.Used = corev1.ResourceList{"cpu": resource.MustParse("1")}
		_, err = k8sClient.CoreV1().ResourceQuotas("ns2").Update(context.TODO(), rq, metav1.UpdateOptions{})
		assert.NoError(t, err)
		waitForDirty("ns2")
	})

	t.Run("PodTemplate Add/Update", func(t *testing.T) {
		pt := &corev1.PodTemplate{
			ObjectMeta: metav1.ObjectMeta{Name: "pt1", Namespace: "ns3"},
			Template:   corev1.PodTemplateSpec{},
		}
		// Add
		_, err := k8sClient.CoreV1().PodTemplates("ns3").Create(context.TODO(), pt, metav1.CreateOptions{})
		assert.NoError(t, err)
		waitForDirty("ns3")

		// Update
		pt.Template.ObjectMeta.Labels = map[string]string{"foo": "bar"}
		_, err = k8sClient.CoreV1().PodTemplates("ns3").Update(context.TODO(), pt, metav1.UpdateOptions{})
		assert.NoError(t, err)
		waitForDirty("ns3")
	})
}
