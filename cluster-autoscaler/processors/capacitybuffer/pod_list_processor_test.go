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
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1beta1"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/client"
	"k8s.io/autoscaler/cluster-autoscaler/utils/drain"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

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
		forceSafeToEvict             bool
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
			expectedBuffersProvCondition: map[string]metav1.Condition{"buffer": {Type: capacitybuffer.ProvisioningCondition, Status: metav1.ConditionFalse}},
			expectError:                  false,
		},
		{
			name:                         "Buffer ready for provisiong with valid reference",
			objectsInKubernetesClient:    []runtime.Object{getTestingPodTemplate("ref", 1)},
			objectsInBuffersClient:       []runtime.Object{getTestingBuffer("buffer", "ref", 1, 1, true, 1, testProvStrategyAllowed)},
			unschedulablePods:            []*corev1.Pod{getTestingPod("Pod")},
			forceSafeToEvict:             true,
			expectedUnschedPodsCount:     2,
			expectedUnschedFakePodsCount: 1,
			expectedBuffersProvCondition: map[string]metav1.Condition{"buffer": {Type: capacitybuffer.ProvisioningCondition, Status: metav1.ConditionTrue}},
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
			forceSafeToEvict:             false,
			expectedUnschedPodsCount:     11,
			expectedUnschedFakePodsCount: 8,
			expectedBuffersProvCondition: map[string]metav1.Condition{
				"buffer1": {Type: capacitybuffer.ProvisioningCondition, Status: metav1.ConditionTrue},
				"buffer2": {Type: capacitybuffer.ProvisioningCondition, Status: metav1.ConditionTrue},
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
			expectedBuffersProvCondition: map[string]metav1.Condition{"buffer2": {Type: capacitybuffer.ProvisioningCondition, Status: metav1.ConditionTrue}},
			expectError:                  false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fakeKubernetesClient := fakeclient.NewSimpleClientset(test.objectsInKubernetesClient...)
			fakeBuffersClient := buffersfake.NewSimpleClientset(test.objectsInBuffersClient...)
			fakeCapacityBuffersClient, _ := client.NewCapacityBufferClientFromClients(fakeBuffersClient, fakeKubernetesClient, nil, nil)

			processor := NewCapacityBufferPodListProcessor(fakeCapacityBuffersClient, []string{testProvStrategyAllowed}, NewDefaultCapacityBuffersFakePodsRegistry(), test.forceSafeToEvict)
			resUnschedulablePods, err := processor.Process(nil, test.unschedulablePods)
			assert.Equal(t, err != nil, test.expectError)

			numberOfFakePods := 0
			fakePodsNames := map[string]bool{}
			for _, pod := range resUnschedulablePods {
				if IsFakeCapacityBuffersPod(pod) {
					numberOfFakePods += 1
					assert.False(t, fakePodsNames[pod.Name])
					safeToEvict, err := strconv.ParseBool(pod.Annotations[drain.PodSafeToEvictKey])
					assert.Equal(t, err == nil && safeToEvict, test.forceSafeToEvict)
					fakePodsNames[pod.Name] = true
				}
			}
			assert.Equal(t, test.expectedUnschedFakePodsCount, numberOfFakePods)

			for bufferName, expectedCondition := range test.expectedBuffersProvCondition {
				buffer, err := fakeBuffersClient.AutoscalingV1beta1().CapacityBuffers(corev1.NamespaceDefault).Get(context.TODO(), bufferName, metav1.GetOptions{})
				assert.Equal(t, err, nil)
				found := false
				for _, cond := range buffer.Status.Conditions {
					if cond.Type == expectedCondition.Type {
						assert.Equal(t, string(expectedCondition.Status), string(cond.Status))
						found = true
						break
					}
				}
				assert.True(t, found, "Condition %s not found", expectedCondition.Type)
			}
		})
	}
}

func TestCapacityBufferFakePodsRegistry(t *testing.T) {
	tests := []struct {
		name                      string
		objectsInKubernetesClient []runtime.Object
		objectsInBuffersClient    []runtime.Object
		unschedulablePods         []*corev1.Pod
		expectedUnschedPodsCount  int
		expectedBuffersPodsNum    map[string]int
	}{
		{
			name:                      "1 ready buffer and 1 not ready buffer",
			objectsInKubernetesClient: []runtime.Object{getTestingPodTemplate("ref1", 1), getTestingPodTemplate("ref2", 1)},
			objectsInBuffersClient: []runtime.Object{
				getTestingBuffer("buffer1", "ref1", 2, 1, false, 1, testProvStrategyAllowed),
				getTestingBuffer("buffer2", "ref2", 3, 1, true, 1, testProvStrategyAllowed),
			},
			unschedulablePods:        []*corev1.Pod{getTestingPod("Pod1"), getTestingPod("Pod2"), getTestingPod("Pod3")},
			expectedUnschedPodsCount: 6,
			expectedBuffersPodsNum:   map[string]int{"buffer2": 3},
		},
		{
			name:                      "2 ready buffers",
			objectsInKubernetesClient: []runtime.Object{getTestingPodTemplate("ref1", 1), getTestingPodTemplate("ref2", 1)},
			objectsInBuffersClient: []runtime.Object{
				getTestingBuffer("buffer1", "ref1", 2, 1, true, 1, testProvStrategyAllowed),
				getTestingBuffer("buffer2", "ref2", 3, 1, true, 1, testProvStrategyAllowed),
			},
			unschedulablePods:        []*corev1.Pod{getTestingPod("Pod1"), getTestingPod("Pod2"), getTestingPod("Pod3")},
			expectedUnschedPodsCount: 8,
			expectedBuffersPodsNum:   map[string]int{"buffer1": 2, "buffer2": 3},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fakeKubernetesClient := fakeclient.NewSimpleClientset(test.objectsInKubernetesClient...)
			fakeBuffersClient := buffersfake.NewSimpleClientset(test.objectsInBuffersClient...)
			fakeCapacityBuffersClient, _ := client.NewCapacityBufferClientFromClients(fakeBuffersClient, fakeKubernetesClient, nil, nil)

			registry := NewDefaultCapacityBuffersFakePodsRegistry()
			processor := NewCapacityBufferPodListProcessor(fakeCapacityBuffersClient, []string{testProvStrategyAllowed}, registry, false)
			resUnschedulablePods, err := processor.Process(nil, test.unschedulablePods)
			assert.Equal(t, nil, err)
			assert.Equal(t, test.expectedUnschedPodsCount, len(resUnschedulablePods))
			for _, pod := range resUnschedulablePods {
				if IsFakeCapacityBuffersPod(pod) {
					podBufferObj, found := registry.FakePodsUIDToBuffer[string(pod.UID)]
					assert.True(t, found)
					expectedPodsNum, found := test.expectedBuffersPodsNum[podBufferObj.Name]
					assert.True(t, found)
					test.expectedBuffersPodsNum[podBufferObj.Name] = expectedPodsNum - 1
				}
			}
			for bufferName := range test.expectedBuffersPodsNum {
				assert.Equal(t, 0, test.expectedBuffersPodsNum[bufferName])
			}
		})
	}
}

func getTestingPod(name string) *corev1.Pod {
	return &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: name}}
}

func getTestingPodTemplate(name string, generation int64) *corev1.PodTemplate {
	return &corev1.PodTemplate{
		TypeMeta:   metav1.TypeMeta{Kind: "PodTemplate", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "default", Generation: generation},
	}
}

func getTestingBuffer(bufferName, refName string, replicas int32, generation int64, ready bool, bufferGeneration int64, provStrategy string) *apiv1.CapacityBuffer {
	buffer := &apiv1.CapacityBuffer{
		TypeMeta:   metav1.TypeMeta{Kind: "CapacityBuffer", APIVersion: "autoscaling.x-k8s.io/v1beta1"},
		ObjectMeta: metav1.ObjectMeta{Name: bufferName, Namespace: "default", Generation: bufferGeneration, UID: types.UID(fmt.Sprintf("%s-uid", bufferName))},
	}
	buffer.Status = *testutil.GetBufferStatus(&apiv1.LocalObjectRef{Name: refName}, &replicas, &generation, &provStrategy, nil)
	if ready {
		buffer.Status.Conditions = testutil.GetConditionReady()
	} else {
		buffer.Status.Conditions = testutil.GetConditionNotReady()
	}
	return buffer
}
