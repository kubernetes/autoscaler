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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1beta1"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	ca_context "k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroupset"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	kube_record "k8s.io/client-go/tools/record"
)

type testReason struct {
	message string
}

func (tr *testReason) Reasons() []string {
	return []string{tr.message}
}

func TestProcess(t *testing.T) {
	testCases := map[string]struct {
		podsRemainUnschedulable         []*apiv1.Pod
		podsAwaitEvaluation             []*apiv1.Pod
		podsTriggeredScaleUp            []*apiv1.Pod
		expectedPodsRemainUnschedulable []*apiv1.Pod
		expectedPodsAwaitEvaluation     []*apiv1.Pod
		expectedPodsTriggeredScaleUp    []*apiv1.Pod
	}{
		"Fake pods are removed from PodsRemainUnschedulable": {
			podsRemainUnschedulable:         []*apiv1.Pod{createPod("pod-1", false, nil), createPod("fake-pod-1", true, nil)},
			expectedPodsRemainUnschedulable: []*apiv1.Pod{createPod("pod-1", false, nil)},
		},
		"Fake pods are removed from PodsTriggerScaleup": {
			podsTriggeredScaleUp:         []*apiv1.Pod{createPod("pod-1", false, nil), createPod("fake-pod-1", true, nil)},
			expectedPodsTriggeredScaleUp: []*apiv1.Pod{createPod("pod-1", false, nil)},
		},
		"Fake pods are removed from PodsAwaitEvaluation": {
			podsAwaitEvaluation:         []*apiv1.Pod{createPod("pod-1", false, nil), createPod("fake-pod-1", true, nil)},
			expectedPodsAwaitEvaluation: []*apiv1.Pod{createPod("pod-1", false, nil)},
		},
		"Fake pods are removed from all pod related lists in scaleup status": {
			podsTriggeredScaleUp:            []*apiv1.Pod{createPod("pod-1", false, nil), createPod("fake-pod-1", true, nil)},
			expectedPodsTriggeredScaleUp:    []*apiv1.Pod{createPod("pod-1", false, nil)},
			podsRemainUnschedulable:         []*apiv1.Pod{createPod("pod-2", false, nil), createPod("fake-pod-2", true, nil)},
			expectedPodsRemainUnschedulable: []*apiv1.Pod{createPod("pod-2", false, nil)},
			podsAwaitEvaluation:             []*apiv1.Pod{createPod("pod-3", false, nil), createPod("fake-pod-3", true, nil)},
			expectedPodsAwaitEvaluation:     []*apiv1.Pod{createPod("pod-3", false, nil)},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			scaleUpStatus := &status.ScaleUpStatus{
				PodsTriggeredScaleUp:    tc.podsTriggeredScaleUp,
				PodsAwaitEvaluation:     tc.podsAwaitEvaluation,
				PodsRemainUnschedulable: makeNoScaleUpInfoFromPods(tc.podsRemainUnschedulable),
			}
			autoscalingCtx := &ca_context.AutoscalingContext{}

			p := NewFakePodsScaleUpStatusProcessor()
			p.Process(autoscalingCtx, scaleUpStatus)

			assert.ElementsMatch(t, tc.expectedPodsRemainUnschedulable, extractPodsFromNoScaleUpInfo(scaleUpStatus.PodsRemainUnschedulable))
			assert.ElementsMatch(t, tc.expectedPodsAwaitEvaluation, scaleUpStatus.PodsAwaitEvaluation)
			assert.ElementsMatch(t, tc.expectedPodsTriggeredScaleUp, scaleUpStatus.PodsTriggeredScaleUp)
		})
	}
}

func TestBuffersEvent(t *testing.T) {
	nodeGroup1 := testprovider.NewTestNodeGroup("ng_1", 10, 0, 5, true, true, "some_type", map[string]string{}, []apiv1.Taint{})
	nodeGroup2 := testprovider.NewTestNodeGroup("ng_2", 20, 3, 8, true, true, "some_other_type", map[string]string{}, []apiv1.Taint{})
	consideredNodeGroups := []cloudprovider.NodeGroup{nodeGroup1, nodeGroup2}
	scaleUpInfo1 := nodegroupset.ScaleUpInfo{
		Group:       nodeGroup1,
		CurrentSize: 5,
		NewSize:     6,
		MaxSize:     nodeGroup1.MaxSize(),
	}
	scaleUpInfo2 := nodegroupset.ScaleUpInfo{
		Group:       nodeGroup2,
		CurrentSize: 8,
		NewSize:     9,
		MaxSize:     nodeGroup1.MaxSize(),
	}
	buffer1 := &v1beta1.CapacityBuffer{
		TypeMeta: metav1.TypeMeta{
			Kind:       capacitybuffer.CapacityBufferKind,
			APIVersion: capacitybuffer.CapacityBufferApiVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "buffer_1",
			UID:  "buffer_1",
		},
	}
	buffer2 := &v1beta1.CapacityBuffer{
		TypeMeta: metav1.TypeMeta{
			Kind:       capacitybuffer.CapacityBufferKind,
			APIVersion: capacitybuffer.CapacityBufferApiVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "buffer_2",
			UID:  "buffer_2",
		},
	}

	notSchedulableReason := &testReason{"not schedulable for testing"}
	alsoNotSchedulableReason := &testReason{"also not schedulable for testing"}
	reasons := map[string]status.Reasons{
		nodeGroup1.Id(): notSchedulableReason,
		nodeGroup2.Id(): alsoNotSchedulableReason,
	}
	testCases := map[string]struct {
		state                       *status.ScaleUpStatus
		expectedTriggeredScaleUp    int
		expectedNotTriggeredScaleUp int
	}{
		"One fake pod, successful scale up": {
			state: &status.ScaleUpStatus{
				Result:                  status.ScaleUpSuccessful,
				ConsideredNodeGroups:    consideredNodeGroups,
				ScaleUpInfos:            []nodegroupset.ScaleUpInfo{scaleUpInfo1},
				PodsTriggeredScaleUp:    []*apiv1.Pod{createPod("pod-1", false, nil), createPod("fake-pod-1", true, buffer1)},
				PodsRemainUnschedulable: []status.NoScaleUpInfo{},
				PodsAwaitEvaluation:     []*apiv1.Pod{},
			},
			expectedTriggeredScaleUp:    1,
			expectedNotTriggeredScaleUp: 0,
		},
		"One fake pod, error scale up": {
			state: &status.ScaleUpStatus{
				Result:                  status.ScaleUpError,
				ConsideredNodeGroups:    consideredNodeGroups,
				ScaleUpInfos:            []nodegroupset.ScaleUpInfo{{}},
				PodsTriggeredScaleUp:    []*apiv1.Pod{},
				PodsRemainUnschedulable: []status.NoScaleUpInfo{},
				PodsAwaitEvaluation:     []*apiv1.Pod{},
			},
			expectedTriggeredScaleUp:    0,
			expectedNotTriggeredScaleUp: 0,
		},
		"One fake pod, empty scale up infos": {
			state: &status.ScaleUpStatus{
				Result:                  status.ScaleUpError,
				ScaleUpInfos:            []nodegroupset.ScaleUpInfo{},
				PodsTriggeredScaleUp:    []*apiv1.Pod{createPod("pod-1", false, nil), createPod("fake-pod-1", true, buffer1)},
				PodsRemainUnschedulable: []status.NoScaleUpInfo{},
				PodsAwaitEvaluation:     []*apiv1.Pod{},
			},
			expectedTriggeredScaleUp:    0,
			expectedNotTriggeredScaleUp: 0,
		},
		"One fake pod, with no ownerReference": {
			state: &status.ScaleUpStatus{
				Result:                  status.ScaleUpError,
				ConsideredNodeGroups:    consideredNodeGroups,
				ScaleUpInfos:            []nodegroupset.ScaleUpInfo{},
				PodsTriggeredScaleUp:    []*apiv1.Pod{createPod("pod-1", false, nil), createPod("fake-pod-1", true, nil)},
				PodsRemainUnschedulable: []status.NoScaleUpInfo{},
				PodsAwaitEvaluation:     []*apiv1.Pod{},
			},
			expectedTriggeredScaleUp:    0,
			expectedNotTriggeredScaleUp: 0,
		},
		"One fake pod, unschedulalble": {
			state: &status.ScaleUpStatus{
				Result:               status.ScaleUpNoOptionsAvailable,
				ConsideredNodeGroups: consideredNodeGroups,
				ScaleUpInfos:         []nodegroupset.ScaleUpInfo{},
				PodsTriggeredScaleUp: []*apiv1.Pod{createPod("pod-1", false, nil)},
				PodsRemainUnschedulable: []status.NoScaleUpInfo{
					{
						Pod:                createPod("fake-pod-1", true, buffer1),
						RejectedNodeGroups: reasons,
					},
				},
				PodsAwaitEvaluation: []*apiv1.Pod{},
			},
			expectedTriggeredScaleUp:    0,
			expectedNotTriggeredScaleUp: 1,
		},
		"One fake pod, unschedulalble with error": {
			state: &status.ScaleUpStatus{
				Result:               status.ScaleUpError,
				ConsideredNodeGroups: consideredNodeGroups,
				ScaleUpInfos:         []nodegroupset.ScaleUpInfo{},
				PodsTriggeredScaleUp: []*apiv1.Pod{createPod("pod-1", false, nil)},
				PodsRemainUnschedulable: []status.NoScaleUpInfo{
					{
						Pod:                createPod("fake-pod-1", true, buffer1),
						RejectedNodeGroups: reasons,
					},
				},
				PodsAwaitEvaluation: []*apiv1.Pod{},
			},
			expectedTriggeredScaleUp:    0,
			expectedNotTriggeredScaleUp: 0,
		},
		"two fake pods for same buffer, one triggers scale up and the other doesn't": {
			state: &status.ScaleUpStatus{
				Result:               status.ScaleUpNoOptionsAvailable,
				ConsideredNodeGroups: consideredNodeGroups,
				ScaleUpInfos:         []nodegroupset.ScaleUpInfo{scaleUpInfo1},
				PodsTriggeredScaleUp: []*apiv1.Pod{createPod("pod-1", false, nil), createPod("fake-pod-1", true, buffer1)},
				PodsRemainUnschedulable: []status.NoScaleUpInfo{
					{
						Pod:                createPod("fake-pod-2", true, buffer1),
						RejectedNodeGroups: reasons,
					},
				},
				PodsAwaitEvaluation: []*apiv1.Pod{},
			},
			expectedTriggeredScaleUp:    1,
			expectedNotTriggeredScaleUp: 1,
		},
		"multiple pods for multiple buffers with mixed conditions": {
			state: &status.ScaleUpStatus{
				Result:               status.ScaleUpNoOptionsAvailable,
				ScaleUpInfos:         []nodegroupset.ScaleUpInfo{scaleUpInfo1, scaleUpInfo2},
				ConsideredNodeGroups: consideredNodeGroups,
				PodsTriggeredScaleUp: []*apiv1.Pod{
					createPod("pod-1", false, nil),
					createPod("fake-pod-2", true, buffer1),
					createPod("fake-pod-4", true, buffer2),
					createPod("fake-pod-5", true, buffer2),
				},
				PodsRemainUnschedulable: []status.NoScaleUpInfo{
					{
						Pod:                createPod("fake-pod-1", true, buffer1),
						RejectedNodeGroups: reasons,
					},
					{
						Pod:                createPod("fake-pod-3", true, buffer2),
						RejectedNodeGroups: reasons,
					},
					{
						Pod:                createPod("pod-2", false, nil),
						RejectedNodeGroups: reasons,
					},
				},
				PodsAwaitEvaluation: []*apiv1.Pod{},
			},
			expectedTriggeredScaleUp:    2,
			expectedNotTriggeredScaleUp: 2,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			fakeRecorder := kube_record.NewFakeRecorder(5)
			ctx := &ca_context.AutoscalingContext{
				AutoscalingKubeClients: context.AutoscalingKubeClients{
					Recorder: fakeRecorder,
				},
			}
			p := NewFakePodsScaleUpStatusProcessor()
			p.Process(ctx, tc.state)

			triggeredScaleUp := 0
			notTriggerScaleUp := 0
			for eventsLeft := true; eventsLeft; {
				select {
				case event := <-fakeRecorder.Events:
					if strings.Contains(event, "TriggeredScaleUp") {
						triggeredScaleUp += 1
					} else if strings.Contains(event, "NotTriggerScaleUp") {
						notTriggerScaleUp += 1
					} else {
						t.Fatalf("Test case '%v' failed. Unexpected event %v", name, event)
					}
				default:
					eventsLeft = false
				}
			}
			assert.Equal(t, tc.expectedTriggeredScaleUp, triggeredScaleUp)
			assert.Equal(t, tc.expectedNotTriggeredScaleUp, notTriggerScaleUp)

		})
	}
}

func TestGetBufferReference(t *testing.T) {
	buffer := &v1beta1.CapacityBuffer{
		TypeMeta: metav1.TypeMeta{
			Kind:       capacitybuffer.CapacityBufferKind,
			APIVersion: capacitybuffer.CapacityBufferApiVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-buffer",
			Namespace: "test-ns",
			UID:       "test-uid",
		},
	}
	expectedRef := &apiv1.ObjectReference{
		Kind:       capacitybuffer.CapacityBufferKind,
		APIVersion: capacitybuffer.CapacityBufferApiVersion,
		Name:       "test-buffer",
		UID:        "test-uid",
		Namespace:  "test-ns",
	}

	tests := []struct {
		name          string
		pod           *apiv1.Pod
		expectedFound bool
		expectedOwner *apiv1.ObjectReference
	}{
		{
			name:          "Pod with CapacityBuffer owner",
			pod:           createPod("pod-1", true, buffer),
			expectedFound: true,
			expectedOwner: expectedRef,
		},
		{
			name:          "Pod with no owner",
			pod:           createPod("pod-2", true, nil),
			expectedFound: false,
		},
		{
			name: "Pod with different owner type",
			pod: BuildTestPod("pod-3", 10, 10, func(p *apiv1.Pod) {
				p.OwnerReferences = []metav1.OwnerReference{
					{
						Kind:       "ReplicaSet",
						Name:       "rs-1",
						APIVersion: "apps/v1",
						UID:        "rs-uid",
					},
				}
			}),
			expectedFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner := getBufferReference(tt.pod)
			if tt.expectedFound {
				assert.NotNil(t, owner)
				assert.Equal(t, tt.expectedOwner, owner)
			} else {
				assert.Nil(t, owner)
			}
		})
	}
}

func createPod(name string, isFake bool, owner *v1beta1.CapacityBuffer) *apiv1.Pod {
	return BuildTestPod(name, 10, 10, func(p *apiv1.Pod) {
		if !isFake {
			return
		}
		*p = *withCapacityBufferFakePodAnnotation(p)
		if owner != nil {
			p.OwnerReferences = []metav1.OwnerReference{
				{
					APIVersion: owner.APIVersion,
					Kind:       owner.Kind,
					Name:       owner.Name,
					UID:        owner.UID,
				},
			}
			p.Namespace = owner.Namespace
		}
	})
}

func makeNoScaleUpInfoFromPods(pods []*apiv1.Pod) []status.NoScaleUpInfo {
	noScaleUpInfos := make([]status.NoScaleUpInfo, len(pods))
	for idx, pod := range pods {
		noScaleUpInfos[idx] = status.NoScaleUpInfo{
			Pod: pod,
		}
	}
	return noScaleUpInfos
}

func extractPodsFromNoScaleUpInfo(noScaleUpInfos []status.NoScaleUpInfo) []*apiv1.Pod {
	pods := make([]*apiv1.Pod, len(noScaleUpInfos))
	for idx, noScaleUpInfo := range noScaleUpInfos {
		pods[idx] = noScaleUpInfo.Pod
	}
	return pods
}
