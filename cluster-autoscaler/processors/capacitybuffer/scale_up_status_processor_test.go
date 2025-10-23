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
	v1alpha1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1alpha1"
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
			podsRemainUnschedulable:         []*apiv1.Pod{createPod("pod-1", false), createPod("fake-pod-1", true)},
			expectedPodsRemainUnschedulable: []*apiv1.Pod{createPod("pod-1", false)},
		},
		"Fake pods are removed from PodsTriggerScaleup": {
			podsTriggeredScaleUp:         []*apiv1.Pod{createPod("pod-1", false), createPod("fake-pod-1", true)},
			expectedPodsTriggeredScaleUp: []*apiv1.Pod{createPod("pod-1", false)},
		},
		"Fake pods are removed from PodsAwaitEvaluation": {
			podsAwaitEvaluation:         []*apiv1.Pod{createPod("pod-1", false), createPod("fake-pod-1", true)},
			expectedPodsAwaitEvaluation: []*apiv1.Pod{createPod("pod-1", false)},
		},
		"Fake pods are removed from all pod related lists in scaleup status": {
			podsTriggeredScaleUp:            []*apiv1.Pod{createPod("pod-1", false), createPod("fake-pod-1", true)},
			expectedPodsTriggeredScaleUp:    []*apiv1.Pod{createPod("pod-1", false)},
			podsRemainUnschedulable:         []*apiv1.Pod{createPod("pod-2", false), createPod("fake-pod-2", true)},
			expectedPodsRemainUnschedulable: []*apiv1.Pod{createPod("pod-2", false)},
			podsAwaitEvaluation:             []*apiv1.Pod{createPod("pod-3", false), createPod("fake-pod-3", true)},
			expectedPodsAwaitEvaluation:     []*apiv1.Pod{createPod("pod-3", false)},
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

			p := NewFakePodsScaleUpStatusProcessor(NewDefaultCapacityBuffersFakePodsRegistry())
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
	buffer1 := &v1alpha1.CapacityBuffer{
		ObjectMeta: metav1.ObjectMeta{
			Name: "buffer_1",
			UID:  "buffer_1",
		},
	}
	buffer2 := &v1alpha1.CapacityBuffer{
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
		buffersRegistry             *capacityBuffersFakePodsRegistry
		expectedTriggeredScaleUp    int
		expectedNotTriggeredScaleUp int
	}{
		"One fake pod, successful scale up": {
			buffersRegistry: NewCapacityBuffersFakePodsRegistry(map[string]*v1alpha1.CapacityBuffer{"fake-pod-1": buffer1}),
			state: &status.ScaleUpStatus{
				Result:                  status.ScaleUpSuccessful,
				ConsideredNodeGroups:    consideredNodeGroups,
				ScaleUpInfos:            []nodegroupset.ScaleUpInfo{scaleUpInfo1},
				PodsTriggeredScaleUp:    []*apiv1.Pod{createPod("pod-1", false), createPod("fake-pod-1", true)},
				PodsRemainUnschedulable: []status.NoScaleUpInfo{},
				PodsAwaitEvaluation:     []*apiv1.Pod{},
			},
			expectedTriggeredScaleUp:    1,
			expectedNotTriggeredScaleUp: 0,
		},
		"One fake pod, error scale up": {
			buffersRegistry: NewCapacityBuffersFakePodsRegistry(map[string]*v1alpha1.CapacityBuffer{"fake-pod-1": buffer1}),
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
			buffersRegistry: NewCapacityBuffersFakePodsRegistry(map[string]*v1alpha1.CapacityBuffer{"fake-pod-1": buffer1}),
			state: &status.ScaleUpStatus{
				Result:                  status.ScaleUpError,
				ScaleUpInfos:            []nodegroupset.ScaleUpInfo{},
				PodsTriggeredScaleUp:    []*apiv1.Pod{createPod("pod-1", false), createPod("fake-pod-1", true)},
				PodsRemainUnschedulable: []status.NoScaleUpInfo{},
				PodsAwaitEvaluation:     []*apiv1.Pod{},
			},
			expectedTriggeredScaleUp:    0,
			expectedNotTriggeredScaleUp: 0,
		},
		"One fake pod, with no node in Registry": {
			buffersRegistry: NewCapacityBuffersFakePodsRegistry(map[string]*v1alpha1.CapacityBuffer{}),
			state: &status.ScaleUpStatus{
				Result:                  status.ScaleUpError,
				ConsideredNodeGroups:    consideredNodeGroups,
				ScaleUpInfos:            []nodegroupset.ScaleUpInfo{},
				PodsTriggeredScaleUp:    []*apiv1.Pod{createPod("pod-1", false), createPod("fake-pod-1", true)},
				PodsRemainUnschedulable: []status.NoScaleUpInfo{},
				PodsAwaitEvaluation:     []*apiv1.Pod{},
			},
			expectedTriggeredScaleUp:    0,
			expectedNotTriggeredScaleUp: 0,
		},
		"One fake pod, unschedulalble": {
			buffersRegistry: NewCapacityBuffersFakePodsRegistry(map[string]*v1alpha1.CapacityBuffer{"fake-pod-1": buffer1}),
			state: &status.ScaleUpStatus{
				Result:               status.ScaleUpNoOptionsAvailable,
				ConsideredNodeGroups: consideredNodeGroups,
				ScaleUpInfos:         []nodegroupset.ScaleUpInfo{},
				PodsTriggeredScaleUp: []*apiv1.Pod{createPod("pod-1", false)},
				PodsRemainUnschedulable: []status.NoScaleUpInfo{
					{
						Pod:                createPod("fake-pod-1", true),
						RejectedNodeGroups: reasons,
					},
				},
				PodsAwaitEvaluation: []*apiv1.Pod{},
			},
			expectedTriggeredScaleUp:    0,
			expectedNotTriggeredScaleUp: 1,
		},
		"One fake pod, unschedulalble with error": {
			buffersRegistry: NewCapacityBuffersFakePodsRegistry(map[string]*v1alpha1.CapacityBuffer{"fake-pod-1": buffer1}),
			state: &status.ScaleUpStatus{
				Result:               status.ScaleUpError,
				ConsideredNodeGroups: consideredNodeGroups,
				ScaleUpInfos:         []nodegroupset.ScaleUpInfo{},
				PodsTriggeredScaleUp: []*apiv1.Pod{createPod("pod-1", false)},
				PodsRemainUnschedulable: []status.NoScaleUpInfo{
					{
						Pod:                createPod("fake-pod-1", true),
						RejectedNodeGroups: reasons,
					},
				},
				PodsAwaitEvaluation: []*apiv1.Pod{},
			},
			expectedTriggeredScaleUp:    0,
			expectedNotTriggeredScaleUp: 0,
		},
		"two fake pods for same buffer, one triggers scale up and the other doesn't": {
			buffersRegistry: NewCapacityBuffersFakePodsRegistry(map[string]*v1alpha1.CapacityBuffer{"fake-pod-1": buffer1, "fake-pod-2": buffer1}),
			state: &status.ScaleUpStatus{
				Result:               status.ScaleUpNoOptionsAvailable,
				ConsideredNodeGroups: consideredNodeGroups,
				ScaleUpInfos:         []nodegroupset.ScaleUpInfo{scaleUpInfo1},
				PodsTriggeredScaleUp: []*apiv1.Pod{createPod("pod-1", false), createPod("fake-pod-1", true)},
				PodsRemainUnschedulable: []status.NoScaleUpInfo{
					{
						Pod:                createPod("fake-pod-2", true),
						RejectedNodeGroups: reasons,
					},
				},
				PodsAwaitEvaluation: []*apiv1.Pod{},
			},
			expectedTriggeredScaleUp:    1,
			expectedNotTriggeredScaleUp: 1,
		},
		"multiple pods for multiple buffers with mixed conditions": {
			buffersRegistry: NewCapacityBuffersFakePodsRegistry(map[string]*v1alpha1.CapacityBuffer{
				"fake-pod-1": buffer1,
				"fake-pod-2": buffer1,
				"fake-pod-3": buffer2,
				"fake-pod-4": buffer2,
				"fake-pod-5": buffer2,
			}),
			state: &status.ScaleUpStatus{
				Result:               status.ScaleUpNoOptionsAvailable,
				ScaleUpInfos:         []nodegroupset.ScaleUpInfo{scaleUpInfo1, scaleUpInfo2},
				ConsideredNodeGroups: consideredNodeGroups,
				PodsTriggeredScaleUp: []*apiv1.Pod{
					createPod("pod-1", false),
					createPod("fake-pod-2", true),
					createPod("fake-pod-4", true),
					createPod("fake-pod-5", true),
				},
				PodsRemainUnschedulable: []status.NoScaleUpInfo{
					{
						Pod:                createPod("fake-pod-1", true),
						RejectedNodeGroups: reasons,
					},
					{
						Pod:                createPod("fake-pod-3", true),
						RejectedNodeGroups: reasons,
					},
					{
						Pod:                createPod("pod-2", false),
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
			ctx := &context.AutoscalingContext{
				AutoscalingKubeClients: context.AutoscalingKubeClients{
					Recorder: fakeRecorder,
				},
			}
			p := NewFakePodsScaleUpStatusProcessor(tc.buffersRegistry)
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

func createPod(name string, isFake bool) *apiv1.Pod {
	return BuildTestPod(name, 10, 10, func(p *apiv1.Pod) {
		if !isFake {
			return
		}
		*p = *withCapacityBufferFakePodAnnotation(p)
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
