/*
Copyright 2017 The Kubernetes Authors.

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

package priority

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	featuregatetesting "k8s.io/component-base/featuregate/testing"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/features"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
)

func TestQuickPressure(t *testing.T) {
	const quickWindow = 10 * time.Minute
	config := updateconfig
	config.PressureQuickUpdateWindow = quickWindow
	config.PressureQuickUpdateFraction = 0.1

	testCases := []struct {
		name             string
		gateDisabled     bool
		optOut           bool
		updateMode       vpa_types.UpdateMode
		cpuTarget        string
		memTarget        string
		stampAge         time.Duration
		podAge           time.Duration
		pods             int
		resizePending    bool
		wantAdmitted     int
		wantPressureOnly int
	}{
		{
			name:             "young pod below the memory target is admitted",
			wantAdmitted:     1,
			wantPressureOnly: 1,
		},
		{
			name:         "gate disabled",
			gateDisabled: true,
		},
		{
			name:   "container opted out",
			optOut: true,
		},
		{
			name:       "Recreate mode is not eligible",
			updateMode: vpa_types.UpdateModeRecreate,
		},
		{
			name:     "stamp older than the window",
			stampAge: quickWindow + time.Second,
		},
		{
			name:      "memory request already at target",
			memTarget: "100M",
		},
		{
			name:          "resize already pending",
			resizePending: true,
		},
		{
			name:             "admissions capped at 10% of the pods",
			pods:             20,
			wantAdmitted:     2,
			wantPressureOnly: 2,
		},
		{
			// The updater applies it clamped to the current request.
			name:             "recommendation that lowers CPU is admitted",
			cpuTarget:        "2",
			wantAdmitted:     1,
			wantPressureOnly: 1,
		},
		{
			name:             "pod past the lifetime gate is admitted normally, not as pressure-only",
			podAge:           13 * time.Hour,
			wantAdmitted:     1,
			wantPressureOnly: 0,
		},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("test case: %s", tc.name), func(t *testing.T) {
			featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.ReactiveMemoryPressureDetection, !tc.gateDisabled)
			cpuTarget, memTarget, updateMode, pods := "4", "200M", vpa_types.UpdateModeInPlace, 1
			if tc.cpuTarget != "" {
				cpuTarget = tc.cpuTarget
			}
			if tc.memTarget != "" {
				memTarget = tc.memTarget
			}
			if tc.updateMode != "" {
				updateMode = tc.updateMode
			}
			if tc.pods != 0 {
				pods = tc.pods
			}

			rp := &test.RecommendationProcessorMock{}
			rp.On("Apply").Return(test.Recommendation().WithContainer(containerName).WithTarget(cpuTarget, memTarget).Get(), nil)
			vpa := test.VerticalPodAutoscaler().WithContainer(containerName).WithTarget(cpuTarget, memTarget).WithUpdateMode(updateMode).Get()
			if !tc.optOut {
				enabled := vpa_types.PressureDetectionEnabled
				vpa.Spec.ResourcePolicy.ContainerPolicies[0].PressureDetection = &enabled
			}

			var podList []*corev1.Pod
			prio := map[string]PodPriority{}
			for i := range pods {
				pod := test.Pod().WithName(fmt.Sprintf("POD%d", i)).AddContainer(test.Container().WithName(containerName).
					WithCPURequest(resource.MustParse("4")).WithMemRequest(resource.MustParse("100M")).Get()).Get()
				pod.UID = types.UID(pod.Name)
				if tc.resizePending {
					pod.Status.Conditions = append(pod.Status.Conditions, corev1.PodCondition{Type: corev1.PodResizePending, Status: corev1.ConditionTrue})
				}
				podList = append(podList, pod)
				prio[pod.Name] = PodPriority{ResourceDiff: 1.0}
			}
			podAge := time.Minute
			if tc.podAge != 0 {
				podAge = tc.podAge
			}
			now := podList[0].Status.StartTime.Add(podAge)
			stampAge := time.Minute
			if tc.stampAge != 0 {
				stampAge = tc.stampAge
			}
			stamp := metav1.NewTime(now.Add(-stampAge))
			vpa.Status.Recommendation.ContainerRecommendations[0].LastPressureTime = &stamp

			calc := NewUpdatePriorityCalculator(vpa, config, rp, NewFakeProcessor(prio))
			calc.SetQuickPressureBudget(pods)
			for _, pod := range podList {
				calc.AddPod(pod, now, map[types.UID]*vpa_types.RecommendedPodResources{})
			}
			assert.Len(t, calc.GetSortedPods(NewDefaultPodEvictionAdmission()), tc.wantAdmitted)
			assert.Len(t, calc.PressureOnly(), tc.wantPressureOnly)
		})
	}
}
