/*
Copyright 2024 The Kubernetes Authors.

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

package podinjection

import (
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	podinjectionbackoff "k8s.io/autoscaler/cluster-autoscaler/processors/podinjection/backoff"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

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
			ctx := &context.AutoscalingContext{}

			p := NewFakePodsScaleUpStatusProcessor(podinjectionbackoff.NewFakePodControllerRegistry())
			p.Process(ctx, scaleUpStatus)

			assert.ElementsMatch(t, tc.expectedPodsRemainUnschedulable, extractPodsFromNoScaleUpInfo(scaleUpStatus.PodsRemainUnschedulable))
			assert.ElementsMatch(t, tc.expectedPodsAwaitEvaluation, scaleUpStatus.PodsAwaitEvaluation)
			assert.ElementsMatch(t, tc.expectedPodsTriggeredScaleUp, scaleUpStatus.PodsTriggeredScaleUp)
		})
	}
}

func createPod(name string, isFake bool) *apiv1.Pod {
	return BuildTestPod(name, 10, 10, func(p *apiv1.Pod) {
		if !isFake {
			return
		}
		*p = *withFakePodAnnotation(p)
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
