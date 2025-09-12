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

package capacitybufferpodlister

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1alpha1"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/client"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/common"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	buffersfake "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/client/clientset/versioned/fake"
	testutil "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/testutil"
	fakeclient "k8s.io/client-go/kubernetes/fake"
)

var (
	testProvStrategyAllowed    = "strategy_1"
	testProvStrategyNotAllowed = "strategy_2"
)

func TestPodListProcessor(t *testing.T) {
	tests := []struct {
		name                         string
		objectsInKubernetesClient    []runtime.Object
		objectsInBuffersClient       []runtime.Object
		unschedulablePods            []*corev1.Pod
		expectedUnschedPodsCount     int
		expectedUnschedFakePodsCount int
		expectedBuffersProvCondition map[string]metav1.Condition
		expectError                  bool
	}{
		{
			name:                         "No buffers to process",
			objectsInKubernetesClient:    []runtime.Object{},
			objectsInBuffersClient:       []runtime.Object{},
			unschedulablePods:            []*corev1.Pod{},
			expectedUnschedPodsCount:     0,
			expectedUnschedFakePodsCount: 0,
			expectError:                  false,
		},
		{
			name:                         "No buffers to process, some existing not related unschedulable pods",
			objectsInKubernetesClient:    []runtime.Object{},
			objectsInBuffersClient:       []runtime.Object{},
			unschedulablePods:            []*corev1.Pod{getTestingPod("P1"), getTestingPod("P2")},
			expectedUnschedPodsCount:     1,
			expectedUnschedFakePodsCount: 0,
			expectError:                  false,
		},
		{
			name:                         "Buffer not ready for provisiong to be ignored",
			objectsInKubernetesClient:    []runtime.Object{},
			objectsInBuffersClient:       []runtime.Object{getTestingBuffer("b", "ref", 1, 1, false, 1, testProvStrategyAllowed)},
			unschedulablePods:            []*corev1.Pod{getTestingPod("Pod")},
			expectedUnschedPodsCount:     1,
			expectedUnschedFakePodsCount: 0,
			expectError:                  false,
		},
		{
			name:                         "Buffer ready for provisiong with no valid reference",
			objectsInKubernetesClient:    []runtime.Object{},
			objectsInBuffersClient:       []runtime.Object{getTestingBuffer("buffer", "ref", 1, 1, true, 1, testProvStrategyAllowed)},
			unschedulablePods:            []*corev1.Pod{getTestingPod("Pod")},
			expectedUnschedPodsCount:     1,
			expectedUnschedFakePodsCount: 0,
			expectedBuffersProvCondition: map[string]metav1.Condition{"buffer": {Type: common.ProvisioningCondition, Status: common.ConditionFalse}},
			expectError:                  false,
		},
		{
			name:                         "Buffer ready for provisiong with valid reference",
			objectsInKubernetesClient:    []runtime.Object{getTestingPodTemplate("ref", 1)},
			objectsInBuffersClient:       []runtime.Object{getTestingBuffer("buffer", "ref", 1, 1, true, 1, testProvStrategyAllowed)},
			unschedulablePods:            []*corev1.Pod{getTestingPod("Pod")},
			expectedUnschedPodsCount:     2,
			expectedUnschedFakePodsCount: 1,
			expectedBuffersProvCondition: map[string]metav1.Condition{"buffer": {Type: common.ProvisioningCondition, Status: common.ConditionTrue}},
			expectError:                  false,
		},
		{
			name:                         "Buffer has not allowed prov strategy",
			objectsInKubernetesClient:    []runtime.Object{getTestingPodTemplate("ref", 1)},
			objectsInBuffersClient:       []runtime.Object{getTestingBuffer("buffer", "ref", 1, 1, true, 1, testProvStrategyNotAllowed)},
			unschedulablePods:            []*corev1.Pod{getTestingPod("Pod")},
			expectedUnschedPodsCount:     1,
			expectedUnschedFakePodsCount: 0,
			expectedBuffersProvCondition: map[string]metav1.Condition{},
			expectError:                  false,
		},
		{
			name:                      "Multiple Buffers ready for provisiong with different references",
			objectsInKubernetesClient: []runtime.Object{getTestingPodTemplate("ref1", 1), getTestingPodTemplate("ref2", 1)},
			objectsInBuffersClient: []runtime.Object{
				getTestingBuffer("buffer1", "ref1", 3, 1, true, 1, testProvStrategyAllowed),
				getTestingBuffer("buffer2", "ref2", 5, 1, true, 1, testProvStrategyAllowed),
			},
			unschedulablePods:            []*corev1.Pod{getTestingPod("Pod1"), getTestingPod("Pod2"), getTestingPod("Pod3")},
			expectedUnschedPodsCount:     11,
			expectedUnschedFakePodsCount: 8,
			expectedBuffersProvCondition: map[string]metav1.Condition{
				"buffer1": {Type: common.ProvisioningCondition, Status: common.ConditionTrue},
				"buffer2": {Type: common.ProvisioningCondition, Status: common.ConditionTrue},
			},
			expectError: false,
		},
		{
			name:                      "Mixing cases",
			objectsInKubernetesClient: []runtime.Object{getTestingPodTemplate("ref1", 1), getTestingPodTemplate("ref2", 1)},
			objectsInBuffersClient: []runtime.Object{
				getTestingBuffer("buffer1", "ref1", 3, 1, false, 1, testProvStrategyAllowed),
				getTestingBuffer("buffer2", "ref2", 5, 1, true, 1, testProvStrategyAllowed),
			},
			unschedulablePods:            []*corev1.Pod{getTestingPod("Pod1"), getTestingPod("Pod2"), getTestingPod("Pod3")},
			expectedUnschedPodsCount:     8,
			expectedUnschedFakePodsCount: 5,
			expectedBuffersProvCondition: map[string]metav1.Condition{"buffer2": {Type: common.ProvisioningCondition, Status: common.ConditionTrue}},
			expectError:                  false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fakeKubernetesClient := fakeclient.NewSimpleClientset(test.objectsInKubernetesClient...)
			fakeBuffersClient := buffersfake.NewSimpleClientset(test.objectsInBuffersClient...)
			fakeCapacityBuffersClient, _ := client.NewCapacityBufferClientFromClients(fakeBuffersClient, fakeKubernetesClient, nil, nil)

			processor := NewCapacityBufferPodListProcessor(fakeCapacityBuffersClient, []string{testProvStrategyAllowed})
			resUnschedulablePods, err := processor.Process(nil, test.unschedulablePods)
			assert.Equal(t, err != nil, test.expectError)

			numberOfFakePods := 0
			fakePodsNames := map[string]bool{}
			for _, pod := range resUnschedulablePods {
				if isFakeCapacityBuffersPod(pod) {
					numberOfFakePods += 1
					assert.False(t, fakePodsNames[pod.Name])
					fakePodsNames[pod.Name] = true
				}
			}
			assert.Equal(t, test.expectedUnschedFakePodsCount, numberOfFakePods)

			for buffer, condition := range test.expectedBuffersProvCondition {
				buffer, err := fakeBuffersClient.AutoscalingV1alpha1().CapacityBuffers(corev1.NamespaceDefault).Get(context.TODO(), buffer, metav1.GetOptions{})
				assert.Equal(t, err, nil)
				assert.Equal(t, len(buffer.Status.Conditions), 1)
				assert.Equal(t, string(buffer.Status.Conditions[0].Type), string(condition.Type))
				assert.Equal(t, string(buffer.Status.Conditions[0].Status), string(condition.Status))
			}
		})
	}
}

func getTestingPod(name string) *corev1.Pod {
	return &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: name}}
}

func getTestingPodTemplate(name string, generation int64) *corev1.PodTemplate {
	return &corev1.PodTemplate{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "default", Generation: generation}}
}

func getTestingBuffer(bufferName, refName string, replicas int32, generation int64, ready bool, bufferGeneration int64, provStrategy string) *apiv1.CapacityBuffer {
	buffer := &apiv1.CapacityBuffer{ObjectMeta: metav1.ObjectMeta{Name: bufferName, Namespace: "default", Generation: bufferGeneration}}
	buffer.Status = *testutil.GetBufferStatus(&apiv1.LocalObjectRef{Name: refName}, &replicas, &generation, &provStrategy, nil)
	if ready {
		buffer.Status.Conditions = testutil.GetConditionReady()
	} else {
		buffer.Status.Conditions = testutil.GetConditionNotReady()
	}
	return buffer
}
