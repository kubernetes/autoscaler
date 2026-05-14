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

package restriction

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/version"
	"k8s.io/client-go/tools/record"
	featuregatetesting "k8s.io/component-base/featuregate/testing"
	baseclocktest "k8s.io/utils/clock/testing"
	"k8s.io/utils/ptr"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/features"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/updater/utils"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
)

type CanInPlaceUpdateTestParams struct {
	name                    string
	pods                    []*corev1.Pod
	replicas                int32
	evictionTolerance       float64
	lastInPlaceAttempt      time.Time
	expectedInPlaceDecision utils.InPlaceDecision
	vpa                     *vpa_types.VerticalPodAutoscaler
	clockTime               *time.Time
	minReplicas             int
	infeasibleAttempts      map[types.UID]*vpa_types.RecommendedPodResources
	vpaForCreatorMaps       *vpa_types.VerticalPodAutoscaler
	alsoTestInPlaceUpdate   bool
}

func TestCanInPlaceUpdate(t *testing.T) {
	featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.InPlaceOrRecreate, true)
	featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.InPlace, true)

	rc := corev1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rc",
			Namespace: "default",
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "ReplicationController",
		},
	}

	index := 0
	generatePod := func() test.PodBuilder {
		index++
		return test.Pod().WithName(fmt.Sprintf("test-%v", index)).WithCreator(&rc.ObjectMeta, &rc.TypeMeta)
	}
	// NOTE: the pod we are checking for CanInPlaceUpdate will always be the first one for these tests
	whichPodIdxForCanInPlaceUpdate := 0

	// Recommendations for infeasible caching tests
	rec1000m1Gi := &vpa_types.RecommendedPodResources{
		ContainerRecommendations: []vpa_types.RecommendedContainerResources{
			{
				ContainerName: "test-container",
				Target: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("1000m"),
					corev1.ResourceMemory: resource.MustParse("1Gi"),
				},
			},
		},
	}
	rec2000m2Gi := &vpa_types.RecommendedPodResources{
		ContainerRecommendations: []vpa_types.RecommendedContainerResources{
			{
				ContainerName: "test-container",
				Target: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("2000m"),
					corev1.ResourceMemory: resource.MustParse("2Gi"),
				},
			},
		},
	}
	rec1000m2Gi := &vpa_types.RecommendedPodResources{
		ContainerRecommendations: []vpa_types.RecommendedContainerResources{
			{
				ContainerName: "test-container",
				Target: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("1000m"),
					corev1.ResourceMemory: resource.MustParse("2Gi"),
				},
			},
		},
	}

	ipVpaWithRec := func(rec *vpa_types.RecommendedPodResources) *vpa_types.VerticalPodAutoscaler {
		v := getIPVpa()
		v.Status.Recommendation = rec
		return v
	}
	iporVpaWithRec := func(rec *vpa_types.RecommendedPodResources) *vpa_types.VerticalPodAutoscaler {
		v := getIPORVpa()
		v.Status.Recommendation = rec
		return v
	}

	podWithUID := func(uid string) *corev1.Pod {
		return test.Pod().
			WithName("test-pod").
			WithCreator(&rc.ObjectMeta, &rc.TypeMeta).
			WithUID(types.UID(uid)).
			Get()
	}

	// Pods for infeasible caching test cases (each needs a unique UID)
	icPodSameRec := podWithUID("test-uid-same-rec")
	icPodLowerRec := podWithUID("test-uid-lower-rec")
	icPodHigherRec := podWithUID("test-uid-higher-rec")
	icPodNoHistory := podWithUID("test-uid-no-history")
	icPodIPOR := podWithUID("test-uid-ipor")
	icPodNoCurrentRec := podWithUID("test-uid-no-current-rec")

	testCases := []CanInPlaceUpdateTestParams{
		{
			name: "CanInPlaceUpdate=InPlaceApproved - (half of 3)",
			pods: []*corev1.Pod{
				generatePod().Get(),
				generatePod().Get(),
				generatePod().Get(),
			},
			replicas:                3,
			evictionTolerance:       0.5,
			lastInPlaceAttempt:      time.Time{},
			expectedInPlaceDecision: utils.InPlaceApproved,
			vpa:                     getIPORVpa(),
		},
		{
			name: "CanInPlaceUpdate=InPlaceDeferred - no pods can be in-placed, one missing",
			pods: []*corev1.Pod{
				generatePod().Get(),
				generatePod().Get(),
			},
			replicas:                3,
			evictionTolerance:       0.5,
			lastInPlaceAttempt:      time.Time{},
			expectedInPlaceDecision: utils.InPlaceDeferred,
			vpa:                     getIPORVpa(),
		},
		{
			name: "CanInPlaceUpdate=InPlaceApproved - small tolerance, all running",
			pods: []*corev1.Pod{
				generatePod().Get(),
				generatePod().Get(),
				generatePod().Get(),
			},
			replicas:                3,
			evictionTolerance:       0.1,
			lastInPlaceAttempt:      time.Time{},
			expectedInPlaceDecision: utils.InPlaceApproved,
			vpa:                     getIPORVpa(),
		},
		{
			name: "CanInPlaceUpdate=InPlaceApproved - small tolerance, one missing",
			pods: []*corev1.Pod{
				generatePod().Get(),
				generatePod().Get(),
			},
			replicas:                3,
			evictionTolerance:       0.5,
			lastInPlaceAttempt:      time.Time{},
			expectedInPlaceDecision: utils.InPlaceDeferred,
			vpa:                     getIPORVpa(),
		},
		{
			name: "CanInPlaceUpdate=InPlaceDeferred - resize Deferred, conditions not met to fallback",
			pods: []*corev1.Pod{
				generatePod().WithPodConditions([]corev1.PodCondition{
					{
						Type:   corev1.PodResizePending,
						Status: corev1.ConditionTrue,
						Reason: corev1.PodReasonDeferred,
					},
				}).Get(),
				generatePod().Get(),
				generatePod().Get(),
			},
			replicas:                3,
			evictionTolerance:       0.5,
			lastInPlaceAttempt:      time.UnixMilli(3600000), // 1 hour from epoch
			expectedInPlaceDecision: utils.InPlaceDeferred,
			vpa:                     getIPORVpa(),
		},
		{
			name: ("CanInPlaceUpdate=InPlaceEvict - resize inProgress for more too long"),
			pods: []*corev1.Pod{
				generatePod().WithPodConditions([]corev1.PodCondition{
					{
						Type:   corev1.PodResizeInProgress,
						Status: corev1.ConditionTrue,
					},
				}).Get(),
				generatePod().Get(),
				generatePod().Get(),
			},
			replicas:                3,
			evictionTolerance:       0.5,
			lastInPlaceAttempt:      time.UnixMilli(0), // epoch (too long ago...)
			expectedInPlaceDecision: utils.InPlaceEvict,
			vpa:                     getIPORVpa(),
		},
		{
			name: "CanInPlaceUpdate=InPlaceDeferred - resize InProgress, conditions not met to fallback",
			pods: []*corev1.Pod{
				generatePod().WithPodConditions([]corev1.PodCondition{
					{
						Type:   corev1.PodResizeInProgress,
						Status: corev1.ConditionTrue,
					},
				}).Get(),
				generatePod().Get(),
				generatePod().Get(),
			},
			replicas:                3,
			evictionTolerance:       0.5,
			lastInPlaceAttempt:      time.UnixMilli(3600000), // 1 hour from epoch
			expectedInPlaceDecision: utils.InPlaceDeferred,
			vpa:                     getIPORVpa(),
		},
		{
			name: "CanInPlaceUpdate=InPlaceEvict - infeasible",
			pods: []*corev1.Pod{
				generatePod().WithPodConditions([]corev1.PodCondition{
					{
						Type:   corev1.PodResizePending,
						Status: corev1.ConditionTrue,
						Reason: corev1.PodReasonInfeasible,
					},
				}).Get(),
				generatePod().Get(),
				generatePod().Get(),
			},
			replicas:                3,
			evictionTolerance:       0.5,
			lastInPlaceAttempt:      time.Time{},
			expectedInPlaceDecision: utils.InPlaceEvict,
			vpa:                     getIPORVpa(),
		},
		{
			name: "InPlace - CanInPlaceUpdate=InPlaceApproved - all pods running",
			pods: []*corev1.Pod{
				generatePod().Get(),
				generatePod().Get(),
				generatePod().Get(),
			},
			replicas:                3,
			evictionTolerance:       0.5,
			lastInPlaceAttempt:      time.Time{},
			expectedInPlaceDecision: utils.InPlaceApproved,
			vpa:                     getIPVpa(),
		},
		{
			name: "InPlace - CanInPlaceUpdate=InPlaceDeferred - resize InProgress, waits indefinitely (no timeout)",
			pods: []*corev1.Pod{
				generatePod().WithPodConditions([]corev1.PodCondition{
					{
						Type:   corev1.PodResizeInProgress,
						Status: corev1.ConditionTrue,
					},
				}).Get(),
				generatePod().Get(),
				generatePod().Get(),
			},
			replicas:                3,
			evictionTolerance:       0.5,
			lastInPlaceAttempt:      time.UnixMilli(0), // epoch (too long ago...)
			expectedInPlaceDecision: utils.InPlaceDeferred,
			vpa:                     getIPVpa(),
		},
		{
			name: "InPlace - CanInPlaceUpdate=InPlaceDeferred - no update policy for vpa",
			pods: []*corev1.Pod{
				generatePod().WithPodConditions([]corev1.PodCondition{
					{
						Type:   corev1.PodResizeInProgress,
						Status: corev1.ConditionTrue,
					},
				}).Get(),
				generatePod().Get(),
				generatePod().Get(),
			},
			replicas:                3,
			evictionTolerance:       0.5,
			lastInPlaceAttempt:      time.UnixMilli(0), // epoch (too long ago...)
			expectedInPlaceDecision: utils.InPlaceEvict,
			vpa:                     getBasicVpa(),
		},
		{
			name: "InPlace mode - waits indefinitely for in-progress updates",
			pods: []*corev1.Pod{
				test.Pod().WithName("wait-1").WithCreator(&rc.ObjectMeta, &rc.TypeMeta).
					WithPodConditions([]corev1.PodCondition{
						{Type: corev1.PodResizeInProgress, Status: corev1.ConditionTrue},
					}).Get(),
				test.Pod().WithName("wait-2").WithCreator(&rc.ObjectMeta, &rc.TypeMeta).Get(),
				test.Pod().WithName("wait-3").WithCreator(&rc.ObjectMeta, &rc.TypeMeta).Get(),
			},
			replicas:                3,
			evictionTolerance:       0.5,
			lastInPlaceAttempt:      time.UnixMilli(0),
			expectedInPlaceDecision: utils.InPlaceDeferred,
			vpa:                     getIPVpa(),
			clockTime:               ptr.To(time.UnixMilli(7200000)),
			vpaForCreatorMaps:       getIPVpa(),
		},
		{
			name: "InPlace mode - infeasible returns InPlaceInfeasible",
			pods: []*corev1.Pod{
				test.Pod().WithName("infeasible-1").WithCreator(&rc.ObjectMeta, &rc.TypeMeta).
					WithPodConditions([]corev1.PodCondition{
						{Type: corev1.PodResizePending, Status: corev1.ConditionTrue, Reason: corev1.PodReasonInfeasible},
					}).Get(),
				test.Pod().WithName("infeasible-2").WithCreator(&rc.ObjectMeta, &rc.TypeMeta).Get(),
				test.Pod().WithName("infeasible-3").WithCreator(&rc.ObjectMeta, &rc.TypeMeta).Get(),
			},
			replicas:                3,
			evictionTolerance:       0.5,
			lastInPlaceAttempt:      time.Time{},
			expectedInPlaceDecision: utils.InPlaceInfeasible,
			vpa:                     getIPVpa(),
			clockTime:               ptr.To(time.Time{}),
			vpaForCreatorMaps:       getIPVpa(),
		},
		{
			name: "InPlace mode - deferred pod with changed recommendations retries",
			pods: []*corev1.Pod{
				test.Pod().WithName("deferred-retry-1").WithCreator(&rc.ObjectMeta, &rc.TypeMeta).
					WithPodConditions([]corev1.PodCondition{
						{Type: corev1.PodResizePending, Status: corev1.ConditionTrue, Reason: corev1.PodReasonDeferred},
					}).Get(),
				test.Pod().WithName("deferred-retry-2").WithCreator(&rc.ObjectMeta, &rc.TypeMeta).Get(),
				test.Pod().WithName("deferred-retry-3").WithCreator(&rc.ObjectMeta, &rc.TypeMeta).Get(),
			},
			replicas:                3,
			evictionTolerance:       0.5,
			lastInPlaceAttempt:      time.Time{},
			expectedInPlaceDecision: utils.InPlaceDeferred,
			vpa:                     getIPVpa(),
			clockTime:               ptr.To(time.Time{}),
			vpaForCreatorMaps:       getIPVpa(),
			alsoTestInPlaceUpdate:   true,
		},
		{
			name: "InPlaceOrRecreate mode - deferred pod with changed recommendations retries",
			pods: []*corev1.Pod{
				test.Pod().WithName("deferred-ipor-1").WithCreator(&rc.ObjectMeta, &rc.TypeMeta).
					WithPodConditions([]corev1.PodCondition{
						{Type: corev1.PodResizePending, Status: corev1.ConditionTrue, Reason: corev1.PodReasonDeferred},
					}).Get(),
				test.Pod().WithName("deferred-ipor-2").WithCreator(&rc.ObjectMeta, &rc.TypeMeta).Get(),
				test.Pod().WithName("deferred-ipor-3").WithCreator(&rc.ObjectMeta, &rc.TypeMeta).Get(),
			},
			replicas:                3,
			evictionTolerance:       0.5,
			lastInPlaceAttempt:      time.Time{},
			expectedInPlaceDecision: utils.InPlaceDeferred,
			vpa:                     getIPORVpa(),
			clockTime:               ptr.To(time.Time{}),
			alsoTestInPlaceUpdate:   true,
		},
		{
			name:                    "InfeasibleCaching - same recommendation returns cached",
			pods:                    []*corev1.Pod{icPodSameRec},
			replicas:                1,
			evictionTolerance:       0.5,
			lastInPlaceAttempt:      time.Time{},
			expectedInPlaceDecision: utils.InPlaceInfeasibleCached,
			vpa:                     ipVpaWithRec(rec1000m1Gi),
			clockTime:               ptr.To(time.Time{}),
			minReplicas:             1,
			vpaForCreatorMaps:       getIPVpa(),
			infeasibleAttempts: map[types.UID]*vpa_types.RecommendedPodResources{
				icPodSameRec.UID: rec1000m1Gi,
			},
		},
		{
			name:                    "InfeasibleCaching - lower recommendation allows retry",
			pods:                    []*corev1.Pod{icPodLowerRec},
			replicas:                1,
			evictionTolerance:       0.5,
			lastInPlaceAttempt:      time.Time{},
			expectedInPlaceDecision: utils.InPlaceApproved,
			vpa:                     ipVpaWithRec(rec1000m2Gi),
			clockTime:               ptr.To(time.Time{}),
			minReplicas:             1,
			vpaForCreatorMaps:       getIPVpa(),
			infeasibleAttempts: map[types.UID]*vpa_types.RecommendedPodResources{
				icPodLowerRec.UID: rec2000m2Gi,
			},
		},
		{
			name:                    "InfeasibleCaching - higher recommendation returns cached",
			pods:                    []*corev1.Pod{icPodHigherRec},
			replicas:                1,
			evictionTolerance:       0.5,
			lastInPlaceAttempt:      time.Time{},
			expectedInPlaceDecision: utils.InPlaceInfeasibleCached,
			vpa:                     ipVpaWithRec(rec2000m2Gi),
			clockTime:               ptr.To(time.Time{}),
			minReplicas:             1,
			vpaForCreatorMaps:       getIPVpa(),
			infeasibleAttempts: map[types.UID]*vpa_types.RecommendedPodResources{
				icPodHigherRec.UID: rec1000m1Gi,
			},
		},
		{
			name:                    "InfeasibleCaching - no history proceeds normally",
			pods:                    []*corev1.Pod{icPodNoHistory},
			replicas:                1,
			evictionTolerance:       0.5,
			lastInPlaceAttempt:      time.Time{},
			expectedInPlaceDecision: utils.InPlaceApproved,
			vpa:                     ipVpaWithRec(rec1000m1Gi),
			clockTime:               ptr.To(time.Time{}),
			minReplicas:             1,
			vpaForCreatorMaps:       getIPVpa(),
		},
		{
			name:                    "InfeasibleCaching - InPlaceOrRecreate ignores cache",
			pods:                    []*corev1.Pod{icPodIPOR},
			replicas:                1,
			evictionTolerance:       0.5,
			lastInPlaceAttempt:      time.Time{},
			expectedInPlaceDecision: utils.InPlaceApproved,
			vpa:                     iporVpaWithRec(rec1000m1Gi),
			clockTime:               ptr.To(time.Time{}),
			minReplicas:             1,
			vpaForCreatorMaps:       getIPORVpa(),
			infeasibleAttempts: map[types.UID]*vpa_types.RecommendedPodResources{
				icPodIPOR.UID: rec1000m1Gi,
			},
		},
		{
			name:                    "InfeasibleCaching - no current recommendation defers",
			pods:                    []*corev1.Pod{icPodNoCurrentRec},
			replicas:                1,
			evictionTolerance:       0.5,
			lastInPlaceAttempt:      time.Time{},
			expectedInPlaceDecision: utils.InPlaceDeferred,
			vpa:                     ipVpaWithRec(nil),
			clockTime:               ptr.To(time.Time{}),
			minReplicas:             1,
			vpaForCreatorMaps:       getIPVpa(),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rc.Spec = corev1.ReplicationControllerSpec{
				Replicas: &tc.replicas,
			}

			selectedPod := tc.pods[whichPodIdxForCanInPlaceUpdate]

			clockTime := time.UnixMilli(3600001) // default: 1 hour from epoch + 1 millis
			if tc.clockTime != nil {
				clockTime = *tc.clockTime
			}
			clock := baseclocktest.NewFakeClock(clockTime)
			lipatm := map[string]time.Time{getPodID(selectedPod): tc.lastInPlaceAttempt}

			minReplicas := 2
			if tc.minReplicas > 0 {
				minReplicas = tc.minReplicas
			}

			factory, err := getRestrictionFactory(&rc, nil, nil, nil, minReplicas, tc.evictionTolerance, clock, lipatm, GetFakeCalculatorsWithFakeResourceCalc(), false)
			assert.NoError(t, err)

			vpaForCreatorMaps := getIPORVpa()
			if tc.vpaForCreatorMaps != nil {
				vpaForCreatorMaps = tc.vpaForCreatorMaps
			}

			creatorToSingleGroupStatsMap, podToReplicaCreatorMap, err := factory.GetCreatorMaps(tc.pods, vpaForCreatorMaps)
			assert.NoError(t, err)
			inPlace := factory.NewPodsInPlaceRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)

			result := inPlace.CanInPlaceUpdate(selectedPod, tc.vpa, tc.infeasibleAttempts)
			assert.Equal(t, tc.expectedInPlaceDecision, result)

			if tc.alsoTestInPlaceUpdate {
				err := inPlace.InPlaceUpdate(selectedPod, tc.vpa, test.FakeEventRecorder())
				assert.NoError(t, err)
			}
		})
	}
}

func TestInPlaceDisabledFeatureGate(t *testing.T) {
	featuregatetesting.SetFeatureGateEmulationVersionDuringTest(t, features.MutableFeatureGate, version.MustParse("1.5"))
	featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.InPlaceOrRecreate, false)

	replicas := int32(5)
	livePods := 5
	tolerance := 1.0

	rc := corev1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rc",
			Namespace: "default",
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "ReplicationController",
		},
		Spec: corev1.ReplicationControllerSpec{
			Replicas: &replicas,
		},
	}

	pods := make([]*corev1.Pod, livePods)
	for i := range pods {
		pods[i] = test.Pod().WithName(getTestPodName(i)).WithCreator(&rc.ObjectMeta, &rc.TypeMeta).Get()
	}

	basicVpa := getBasicVpa()
	factory, err := getRestrictionFactory(&rc, nil, nil, nil, 2, tolerance, nil, nil, GetFakeCalculatorsWithFakeResourceCalc(), false)
	assert.NoError(t, err)
	creatorToSingleGroupStatsMap, podToReplicaCreatorMap, err := factory.GetCreatorMaps(pods, basicVpa)
	assert.NoError(t, err)
	inplace := factory.NewPodsInPlaceRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)

	for _, pod := range pods {
		assert.Equal(t, utils.InPlaceEvict, inplace.CanInPlaceUpdate(pod, basicVpa, nil))
	}
}

func TestInPlaceTooFewReplicas(t *testing.T) {
	featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.InPlaceOrRecreate, true)

	replicas := int32(5)
	livePods := 5
	tolerance := 0.5

	rc := corev1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rc",
			Namespace: "default",
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "ReplicationController",
		},
		Spec: corev1.ReplicationControllerSpec{
			Replicas: &replicas,
		},
	}

	pods := make([]*corev1.Pod, livePods)
	for i := range pods {
		pods[i] = test.Pod().WithName(getTestPodName(i)).WithCreator(&rc.ObjectMeta, &rc.TypeMeta).Get()
	}

	clock := baseclocktest.NewFakeClock(time.Time{})
	lipatm := map[string]time.Time{}

	basicVpa := getIPORVpa()
	factory, err := getRestrictionFactory(&rc, nil, nil, nil, 10 /* minReplicas */, tolerance, clock, lipatm, GetFakeCalculatorsWithFakeResourceCalc(), false)
	assert.NoError(t, err)
	creatorToSingleGroupStatsMap, podToReplicaCreatorMap, err := factory.GetCreatorMaps(pods, basicVpa)
	assert.NoError(t, err)
	inplace := factory.NewPodsInPlaceRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)

	for _, pod := range pods {
		assert.Equal(t, utils.InPlaceDeferred, inplace.CanInPlaceUpdate(pod, basicVpa, nil))
	}

	for _, pod := range pods {
		err := inplace.InPlaceUpdate(pod, basicVpa, test.FakeEventRecorder())
		assert.Error(t, err, "Error expected")
	}
}

func TestEvictionToleranceForInPlace(t *testing.T) {
	featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.InPlaceOrRecreate, true)

	replicas := int32(5)
	livePods := 5
	tolerance := 0.8

	rc := corev1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rc",
			Namespace: "default",
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "ReplicationController",
		},
		Spec: corev1.ReplicationControllerSpec{
			Replicas: &replicas,
		},
	}

	pods := make([]*corev1.Pod, livePods)
	for i := range pods {
		pods[i] = test.Pod().WithName(getTestPodName(i)).WithCreator(&rc.ObjectMeta, &rc.TypeMeta).Get()
	}

	clock := baseclocktest.NewFakeClock(time.Time{})
	lipatm := map[string]time.Time{}

	basicVpa := getIPORVpa()
	factory, err := getRestrictionFactory(&rc, nil, nil, nil, 2 /* minReplicas */, tolerance, clock, lipatm, GetFakeCalculatorsWithFakeResourceCalc(), false)
	assert.NoError(t, err)
	creatorToSingleGroupStatsMap, podToReplicaCreatorMap, err := factory.GetCreatorMaps(pods, basicVpa)
	assert.NoError(t, err)
	inplace := factory.NewPodsInPlaceRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)

	for _, pod := range pods {
		assert.Equal(t, utils.InPlaceApproved, inplace.CanInPlaceUpdate(pod, basicVpa, nil))
	}

	for _, pod := range pods[:4] {
		err := inplace.InPlaceUpdate(pod, basicVpa, test.FakeEventRecorder())
		assert.Nil(t, err, "Should evict with no error")
	}
}

func TestEvictionToleranceForInPlaceWithSkipDisruptionBudget(t *testing.T) {
	featuregatetesting.SetFeatureGateDuringTest(
		t,
		features.MutableFeatureGate,
		features.InPlaceOrRecreate,
		true,
	)

	replicas := int32(5)
	livePods := 5
	tolerance := 0.8

	rc := corev1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rc",
			Namespace: "default",
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "ReplicationController",
		},
		Spec: corev1.ReplicationControllerSpec{
			Replicas: &replicas,
		},
	}

	pods := make([]*corev1.Pod, livePods)
	for i := range pods {
		pods[i] = test.Pod().
			WithName(getTestPodName(i)).
			WithCreator(&rc.ObjectMeta, &rc.TypeMeta).
			Get()
	}

	clock := baseclocktest.NewFakeClock(time.Time{})
	lipatm := map[string]time.Time{}

	basicVpa := getIPORVpa()
	// inPlaceSkipDisruptionBudget = true
	factory, err := getRestrictionFactory(&rc, nil, nil, nil, 2 /* minReplicas */, tolerance, clock, lipatm, GetFakeCalculatorsWithFakeResourceCalc(), true /* inPlaceSkipDisruptionBudget */)
	assert.NoError(t, err)

	creatorToSingleGroupStatsMap, podToReplicaCreatorMap, err := factory.GetCreatorMaps(pods, basicVpa)
	assert.NoError(t, err)
	inplace := factory.NewPodsInPlaceRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)

	// All in-place updates should be approved
	for _, pod := range pods {
		assert.Equal(t, utils.InPlaceApproved, inplace.CanInPlaceUpdate(pod, basicVpa, nil))
	}

	// And all updates should succeed without being blocked by eviction tolerance
	for _, pod := range pods {
		err := inplace.InPlaceUpdate(pod, basicVpa, test.FakeEventRecorder())
		assert.NoError(t, err)
	}
}

func TestEvictionToleranceForInPlaceWithSkipDisruptionBudgetWithLessThanMinimumPods(t *testing.T) {
	featuregatetesting.SetFeatureGateDuringTest(
		t,
		features.MutableFeatureGate,
		features.InPlaceOrRecreate,
		true,
	)

	replicas := int32(5)
	livePods := 1
	tolerance := 0.8

	rc := corev1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rc",
			Namespace: "default",
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "ReplicationController",
		},
		Spec: corev1.ReplicationControllerSpec{
			Replicas: &replicas,
		},
	}

	pods := make([]*corev1.Pod, livePods)
	for i := range pods {
		pods[i] = test.Pod().
			WithName(getTestPodName(i)).
			WithCreator(&rc.ObjectMeta, &rc.TypeMeta).
			Get()
	}

	clock := baseclocktest.NewFakeClock(time.Time{})
	lipatm := map[string]time.Time{}

	basicVpa := getIPORVpa()
	// inPlaceSkipDisruptionBudget = true
	// minReplicas needs to be greater than the number of live pods
	factory, err := getRestrictionFactory(&rc, nil, nil, nil, 2 /* minReplicas */, tolerance, clock, lipatm, GetFakeCalculatorsWithFakeResourceCalc(), true /* inPlaceSkipDisruptionBudget */)
	assert.NoError(t, err)

	creatorToSingleGroupStatsMap, podToReplicaCreatorMap, err := factory.GetCreatorMaps(pods, basicVpa)
	assert.NoError(t, err)

	inplace := factory.NewPodsInPlaceRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)
	evict := factory.NewPodsEvictionRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)

	// All in-place updates should be approved
	for _, pod := range pods {
		assert.Equal(t, utils.InPlaceApproved, inplace.CanInPlaceUpdate(pod, basicVpa, nil))
	}

	// And all in-place updates should succeed without being blocked by eviction tolerance, but eviction
	// should not be allowed since we are below minimum replicas, even with inPlaceSkipDisruptionBudget=true
	for _, pod := range pods {
		err := inplace.InPlaceUpdate(pod, basicVpa, test.FakeEventRecorder())
		assert.NoError(t, err)

		eviction := evict.CanEvict(pod)
		assert.False(t, eviction, "Pod should not be evictable when below minimum replicas, even with inPlaceSkipDisruptionBudget=true")

		err = evict.Evict(pod, basicVpa, test.FakeEventRecorder())
		assert.Error(t, err)
	}
}

func TestInPlaceSkipDisruptionBudgetWithResizePolicy(t *testing.T) {
	featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.InPlaceOrRecreate, true)

	replicas := int32(5)
	livePods := 5
	tolerance := 0.1 // Very low tolerance - would normally block most updates

	rc := corev1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rc",
			Namespace: "default",
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "ReplicationController",
		},
		Spec: corev1.ReplicationControllerSpec{
			Replicas: &replicas,
		},
	}

	testCases := []struct {
		name                        string
		podBuilder                  test.PodBuilder
		inPlaceSkipDisruptionBudget bool
		expectedUpdateSuccesses     int
	}{
		{
			name: "NotRequired policy with flag enabled - should skip disruption budget",
			podBuilder: test.Pod().WithCreator(&rc.ObjectMeta, &rc.TypeMeta).AddContainer(
				test.Container().WithName("container1").WithContainerResizePolicy([]corev1.ContainerResizePolicy{
					{ResourceName: corev1.ResourceCPU, RestartPolicy: corev1.NotRequired},
					{ResourceName: corev1.ResourceMemory, RestartPolicy: corev1.NotRequired},
				}).Get()),
			inPlaceSkipDisruptionBudget: true,
			expectedUpdateSuccesses:     5,
		},
		{
			name: "RestartContainer policy with flag enabled - should respect disruption budget",
			podBuilder: test.Pod().WithCreator(&rc.ObjectMeta, &rc.TypeMeta).AddContainer(
				test.Container().WithName("container1").WithContainerResizePolicy([]corev1.ContainerResizePolicy{
					{ResourceName: corev1.ResourceCPU, RestartPolicy: corev1.NotRequired},
					{ResourceName: corev1.ResourceMemory, RestartPolicy: corev1.NotRequired},
				}).Get()),
			inPlaceSkipDisruptionBudget: false,
			expectedUpdateSuccesses:     1, // Still respects budget because of RestartContainer
		},
		{
			name: "Only RestartContainer policy with flag enabled - should respect disruption budget",
			podBuilder: test.Pod().WithCreator(&rc.ObjectMeta, &rc.TypeMeta).AddContainer(
				test.Container().WithName("container1").WithContainerResizePolicy([]corev1.ContainerResizePolicy{
					{ResourceName: corev1.ResourceCPU, RestartPolicy: corev1.RestartContainer},
				}).Get()),
			inPlaceSkipDisruptionBudget: true,
			expectedUpdateSuccesses:     1,
		},
		{
			name:                        "No resize policy with flag enabled - should skip disruption budget (K8s default is NotRequired)",
			podBuilder:                  test.Pod().WithCreator(&rc.ObjectMeta, &rc.TypeMeta).AddContainer(test.Container().WithName("container1").Get()),
			inPlaceSkipDisruptionBudget: true,
			expectedUpdateSuccesses:     5,
		},
		{
			name:                        "No resize policy with flag disabled - should respect disruption budget",
			podBuilder:                  test.Pod().WithCreator(&rc.ObjectMeta, &rc.TypeMeta).AddContainer(test.Container().WithName("container1").Get()),
			inPlaceSkipDisruptionBudget: false,
			expectedUpdateSuccesses:     1,
		},
		{
			name: "Multiple containers - all NotRequired - should skip disruption budget",
			podBuilder: test.Pod().WithCreator(&rc.ObjectMeta, &rc.TypeMeta).AddContainer(
				test.Container().WithName("container1").WithContainerResizePolicy([]corev1.ContainerResizePolicy{
					{ResourceName: corev1.ResourceCPU, RestartPolicy: corev1.NotRequired},
					{ResourceName: corev1.ResourceMemory, RestartPolicy: corev1.NotRequired},
				}).Get(),
			).AddContainer(
				test.Container().WithName("container2").WithContainerResizePolicy([]corev1.ContainerResizePolicy{
					{ResourceName: corev1.ResourceCPU, RestartPolicy: corev1.NotRequired},
					{ResourceName: corev1.ResourceMemory, RestartPolicy: corev1.NotRequired},
				}).Get()),
			inPlaceSkipDisruptionBudget: true,
			expectedUpdateSuccesses:     5,
		},
		{
			name: "Multiple containers - one NotRequired, one RestartContainer - should respect disruption budget",
			podBuilder: test.Pod().WithCreator(&rc.ObjectMeta, &rc.TypeMeta).AddContainer(
				test.Container().WithName("container1").WithContainerResizePolicy([]corev1.ContainerResizePolicy{
					{ResourceName: corev1.ResourceCPU, RestartPolicy: corev1.NotRequired},
					{ResourceName: corev1.ResourceMemory, RestartPolicy: corev1.NotRequired},
				}).Get(),
			).AddContainer(
				test.Container().WithName("container2").WithContainerResizePolicy([]corev1.ContainerResizePolicy{
					{ResourceName: corev1.ResourceCPU, RestartPolicy: corev1.RestartContainer},
					{ResourceName: corev1.ResourceMemory, RestartPolicy: corev1.RestartContainer},
				}).Get()),
			inPlaceSkipDisruptionBudget: true,
			expectedUpdateSuccesses:     1, // Respects budget because container2 has RestartContainer
		},
		{
			name: "Multiple containers - one with policy, one without - should skip disruption budget",
			podBuilder: test.Pod().WithCreator(&rc.ObjectMeta, &rc.TypeMeta).AddContainer(
				test.Container().WithName("container1").WithContainerResizePolicy([]corev1.ContainerResizePolicy{
					{ResourceName: corev1.ResourceCPU, RestartPolicy: corev1.NotRequired},
				}).Get(),
			).AddContainer(
				test.Container().WithName("container2").Get()),
			inPlaceSkipDisruptionBudget: true,
			expectedUpdateSuccesses:     5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pods := make([]*corev1.Pod, livePods)
			for index := range pods {
				pods[index] = tc.podBuilder.WithName(getTestPodName(index)).Get()
			}

			clock := baseclocktest.NewFakeClock(time.Time{})
			lipatm := map[string]time.Time{}

			basicVpa := getIPORVpa()
			factory, err := getRestrictionFactory(&rc, nil, nil, nil, 2, tolerance, clock, lipatm, GetFakeCalculatorsWithFakeResourceCalc(), tc.inPlaceSkipDisruptionBudget)
			assert.NoError(t, err)

			creatorToSingleGroupStatsMap, podToReplicaCreatorMap, err := factory.GetCreatorMaps(pods, basicVpa)
			assert.NoError(t, err)

			inplace := factory.NewPodsInPlaceRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)

			successCount := 0
			for _, pod := range pods {
				decision := inplace.CanInPlaceUpdate(pod, basicVpa, nil)
				if decision != utils.InPlaceApproved {
					continue
				}
				err := inplace.InPlaceUpdate(pod, basicVpa, test.FakeEventRecorder())
				if err == nil {
					successCount++
				}
			}
			assert.Equal(t, tc.expectedUpdateSuccesses, successCount,
				"Expected %d successful updates but got %d", tc.expectedUpdateSuccesses, successCount)
		})
	}
}

func TestInPlaceAtLeastOne(t *testing.T) {
	featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.InPlaceOrRecreate, true)

	replicas := int32(5)
	livePods := 5
	tolerance := 0.1

	rc := corev1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rc",
			Namespace: "default",
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "ReplicationController",
		},
		Spec: corev1.ReplicationControllerSpec{
			Replicas: &replicas,
		},
	}

	pods := make([]*corev1.Pod, livePods)
	for i := range pods {
		pods[i] = test.Pod().WithName(getTestPodName(i)).WithCreator(&rc.ObjectMeta, &rc.TypeMeta).Get()
	}

	inPlaceOrRecreateVPA := getIPORVpa()
	factory, err := getRestrictionFactory(&rc, nil, nil, nil, 2, tolerance, nil, nil, GetFakeCalculatorsWithFakeResourceCalc(), false)
	assert.NoError(t, err)
	creatorToSingleGroupStatsMap, podToReplicaCreatorMap, err := factory.GetCreatorMaps(pods, inPlaceOrRecreateVPA)
	assert.NoError(t, err)
	inplace := factory.NewPodsInPlaceRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)

	for _, pod := range pods[:1] {
		assert.Equal(t, utils.InPlaceApproved, inplace.CanInPlaceUpdate(pod, inPlaceOrRecreateVPA, nil))
		err := inplace.InPlaceUpdate(pod, inPlaceOrRecreateVPA, test.FakeEventRecorder())
		assert.Nil(t, err, "Should in-place update with no error")
	}
	for _, pod := range pods[1:] {
		assert.Equal(t, utils.InPlaceDeferred, inplace.CanInPlaceUpdate(pod, inPlaceOrRecreateVPA, nil))
		err := inplace.InPlaceUpdate(pod, inPlaceOrRecreateVPA, test.FakeEventRecorder())
		assert.Nil(t, err, "Error should not be expected")
	}
}

func TestInPlaceUpdate_EventEmission(t *testing.T) {
	featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.InPlaceOrRecreate, true)

	replicas := int32(5)
	livePods := 5
	tolerance := 0.1

	rc := corev1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rc",
			Namespace: "default",
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "ReplicationController",
		},
		Spec: corev1.ReplicationControllerSpec{
			Replicas: &replicas,
		},
	}

	pods := make([]*corev1.Pod, livePods)
	for i := range pods {
		pods[i] = test.Pod().WithName(getTestPodName(i)).WithCreator(&rc.ObjectMeta, &rc.TypeMeta).Get()
	}

	inPlaceOrRecreateVPA := getIPORVpa()
	factory, err := getRestrictionFactory(&rc, nil, nil, nil, 2, tolerance, nil, nil, GetFakeCalculatorsWithFakeResourceCalc(), false)
	assert.NoError(t, err)
	creatorToSingleGroupStatsMap, podToReplicaCreatorMap, err := factory.GetCreatorMaps(pods, inPlaceOrRecreateVPA)
	assert.NoError(t, err)
	inplace := factory.NewPodsInPlaceRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)

	eventRecorder := record.NewFakeRecorder(10)

	err = inplace.InPlaceUpdate(pods[0], inPlaceOrRecreateVPA, eventRecorder)
	assert.NoError(t, err)

	select {
	case event := <-eventRecorder.Events:
		assert.Contains(t, event, "InPlaceResizedByVPA")
	case <-time.After(1 * time.Second):
		assert.Fail(t, "timeout waiting for event")
	}
}

func TestInPlaceModeDisabledFeatureGate(t *testing.T) {
	featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.InPlace, false)

	replicas := int32(5)
	livePods := 5
	tolerance := 1.0

	rc := corev1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rc",
			Namespace: "default",
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "ReplicationController",
		},
		Spec: corev1.ReplicationControllerSpec{
			Replicas: &replicas,
		},
	}

	pods := make([]*corev1.Pod, livePods)
	for i := range pods {
		pods[i] = test.Pod().WithName(getTestPodName(i)).WithCreator(&rc.ObjectMeta, &rc.TypeMeta).Get()
	}

	inPlaceVpa := getIPVpa()
	factory, err := getRestrictionFactory(&rc, nil, nil, nil, 2, tolerance, nil, nil, GetFakeCalculatorsWithFakeResourceCalc(), false)
	assert.NoError(t, err)
	creatorToSingleGroupStatsMap, podToReplicaCreatorMap, err := factory.GetCreatorMaps(pods, inPlaceVpa)
	assert.NoError(t, err)
	inplace := factory.NewPodsInPlaceRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)

	for _, pod := range pods {
		assert.Equal(t, utils.InPlaceDeferred, inplace.CanInPlaceUpdate(pod, inPlaceVpa, nil),
			"InPlace mode should return InPlaceDeferred when feature gate is disabled")
	}
}

func TestInPlaceModeAtLeastOne(t *testing.T) {
	featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.InPlace, true)

	replicas := int32(5)
	livePods := 5
	tolerance := 0.1 // Very small tolerance, but at least one should be allowed

	rc := corev1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rc",
			Namespace: "default",
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "ReplicationController",
		},
		Spec: corev1.ReplicationControllerSpec{
			Replicas: &replicas,
		},
	}

	pods := make([]*corev1.Pod, livePods)
	for i := range pods {
		pods[i] = test.Pod().WithName(getTestPodName(i)).WithCreator(&rc.ObjectMeta, &rc.TypeMeta).Get()
	}

	inPlaceVpa := getIPVpa()
	factory, err := getRestrictionFactory(&rc, nil, nil, nil, 2, tolerance, nil, nil, GetFakeCalculatorsWithFakeResourceCalc(), false)
	assert.NoError(t, err)
	creatorToSingleGroupStatsMap, podToReplicaCreatorMap, err := factory.GetCreatorMaps(pods, inPlaceVpa)
	assert.NoError(t, err)
	inplace := factory.NewPodsInPlaceRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)

	for _, pod := range pods {
		assert.Equal(t, utils.InPlaceApproved, inplace.CanInPlaceUpdate(pod, inPlaceVpa, nil))
	}

	for _, pod := range pods {
		err := inplace.InPlaceUpdate(pod, inPlaceVpa, test.FakeEventRecorder())
		assert.NoError(t, err, "Should in-place update at least one pod even with small tolerance")
	}
}

func TestInPlaceModeResizeStatuses(t *testing.T) {
	featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.InPlace, true)
	featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.InPlaceOrRecreate, true)

	replicas := int32(3)
	tolerance := 0.5

	rc := corev1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rc",
			Namespace: "default",
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "ReplicationController",
		},
		Spec: corev1.ReplicationControllerSpec{
			Replicas: &replicas,
		},
	}

	testCases := []struct {
		name               string
		podConditions      []corev1.PodCondition
		updateMode         vpa_types.UpdateMode
		expectedDecision   utils.InPlaceDecision
		lastInPlaceAttempt time.Time
		clockTime          time.Time
	}{
		// InPlace mode tests
		{
			name: "InPlace mode - ResizeStatusDeferred returns InPlaceDeferred",
			podConditions: []corev1.PodCondition{
				{
					Type:   corev1.PodResizePending,
					Status: corev1.ConditionTrue,
					Reason: corev1.PodReasonDeferred,
				},
			},
			updateMode:         vpa_types.UpdateModeInPlace,
			expectedDecision:   utils.InPlaceDeferred,
			lastInPlaceAttempt: time.UnixMilli(0),
			clockTime:          time.UnixMilli(3600000),
		},
		{
			name: "InPlace mode - ResizeStatusInfeasible returns InPlaceInfeasible",
			podConditions: []corev1.PodCondition{
				{
					Type:   corev1.PodResizePending,
					Status: corev1.ConditionTrue,
					Reason: corev1.PodReasonInfeasible,
				},
			},
			updateMode:         vpa_types.UpdateModeInPlace,
			expectedDecision:   utils.InPlaceInfeasible,
			lastInPlaceAttempt: time.UnixMilli(0),
			clockTime:          time.UnixMilli(3600000),
		},
		{
			name: "InPlace mode - ResizeStatusInProgress returns InPlaceDeferred",
			podConditions: []corev1.PodCondition{
				{
					Type:   corev1.PodResizeInProgress,
					Status: corev1.ConditionTrue,
				},
			},
			updateMode:         vpa_types.UpdateModeInPlace,
			expectedDecision:   utils.InPlaceDeferred,
			lastInPlaceAttempt: time.UnixMilli(0),
			clockTime:          time.UnixMilli(3600000),
		},
		{
			name: "InPlace mode - ResizeStatusError returns InPlaceInfeasible (retry)",
			podConditions: []corev1.PodCondition{
				{
					Type:    corev1.PodResizeInProgress,
					Status:  corev1.ConditionTrue,
					Reason:  corev1.PodReasonError,
					Message: "some error",
				},
			},
			updateMode:         vpa_types.UpdateModeInPlace,
			expectedDecision:   utils.InPlaceInfeasible,
			lastInPlaceAttempt: time.UnixMilli(0),
			clockTime:          time.UnixMilli(3600000),
		},
		// InPlaceOrRecreate mode tests - same conditions, different behavior
		{
			name: "InPlaceOrRecreate mode - ResizeStatusDeferred within timeout returns InPlaceDeferred",
			podConditions: []corev1.PodCondition{
				{
					Type:   corev1.PodResizePending,
					Status: corev1.ConditionTrue,
					Reason: corev1.PodReasonDeferred,
				},
			},
			updateMode:         vpa_types.UpdateModeInPlaceOrRecreate,
			expectedDecision:   utils.InPlaceDeferred,
			lastInPlaceAttempt: time.UnixMilli(3600000), // 1 hour from epoch
			clockTime:          time.UnixMilli(3600001), // just 1ms later
		},
		{
			name: "InPlaceOrRecreate mode - ResizeStatusDeferred past timeout returns InPlaceEvict",
			podConditions: []corev1.PodCondition{
				{
					Type:   corev1.PodResizePending,
					Status: corev1.ConditionTrue,
					Reason: corev1.PodReasonDeferred,
				},
			},
			updateMode:         vpa_types.UpdateModeInPlaceOrRecreate,
			expectedDecision:   utils.InPlaceEvict,
			lastInPlaceAttempt: time.UnixMilli(0),                                          // epoch
			clockTime:          time.UnixMilli(int64(10 * time.Minute / time.Millisecond)), // 10 minutes later (past 5 min timeout)
		},
		{
			name: "InPlaceOrRecreate mode - ResizeStatusInfeasible returns InPlaceEvict",
			podConditions: []corev1.PodCondition{
				{
					Type:   corev1.PodResizePending,
					Status: corev1.ConditionTrue,
					Reason: corev1.PodReasonInfeasible,
				},
			},
			updateMode:         vpa_types.UpdateModeInPlaceOrRecreate,
			expectedDecision:   utils.InPlaceEvict,
			lastInPlaceAttempt: time.UnixMilli(0),
			clockTime:          time.UnixMilli(3600000),
		},
		{
			name: "InPlaceOrRecreate mode - ResizeStatusInProgress within timeout returns InPlaceDeferred",
			podConditions: []corev1.PodCondition{
				{
					Type:   corev1.PodResizeInProgress,
					Status: corev1.ConditionTrue,
				},
			},
			updateMode:         vpa_types.UpdateModeInPlaceOrRecreate,
			expectedDecision:   utils.InPlaceDeferred,
			lastInPlaceAttempt: time.UnixMilli(3600000),
			clockTime:          time.UnixMilli(3600001),
		},
		{
			name: "InPlaceOrRecreate mode - ResizeStatusInProgress past timeout returns InPlaceEvict",
			podConditions: []corev1.PodCondition{
				{
					Type:   corev1.PodResizeInProgress,
					Status: corev1.ConditionTrue,
				},
			},
			updateMode:         vpa_types.UpdateModeInPlaceOrRecreate,
			expectedDecision:   utils.InPlaceEvict,
			lastInPlaceAttempt: time.UnixMilli(0),                                       // epoch
			clockTime:          time.UnixMilli(int64(2 * time.Hour / time.Millisecond)), // 2 hours later (past 1 hour timeout)
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pod := test.Pod().WithName("test-pod").WithCreator(&rc.ObjectMeta, &rc.TypeMeta).
				WithPodConditions(tc.podConditions).Get()
			pods := []*corev1.Pod{
				pod,
				test.Pod().WithName("test-pod-2").WithCreator(&rc.ObjectMeta, &rc.TypeMeta).Get(),
				test.Pod().WithName("test-pod-3").WithCreator(&rc.ObjectMeta, &rc.TypeMeta).Get(),
			}

			clock := baseclocktest.NewFakeClock(tc.clockTime)
			lipatm := map[string]time.Time{getPodID(pod): tc.lastInPlaceAttempt}

			var vpa *vpa_types.VerticalPodAutoscaler
			if tc.updateMode == vpa_types.UpdateModeInPlace {
				vpa = getIPVpa()
			} else {
				vpa = getIPORVpa()
			}

			factory, err := getRestrictionFactory(&rc, nil, nil, nil, 2, tolerance, clock, lipatm, GetFakeCalculatorsWithFakeResourceCalc(), false)
			assert.NoError(t, err)
			creatorToSingleGroupStatsMap, podToReplicaCreatorMap, err := factory.GetCreatorMaps(pods, vpa)
			assert.NoError(t, err)
			inplace := factory.NewPodsInPlaceRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)

			result := inplace.CanInPlaceUpdate(pod, vpa, nil)
			assert.Equal(t, tc.expectedDecision, result)
		})
	}
}

func TestInPlaceModeAllowsRetryForInfeasible(t *testing.T) {
	featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.InPlace, true)

	replicas := int32(3)
	tolerance := 0.5

	rc := corev1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rc",
			Namespace: "default",
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "ReplicationController",
		},
		Spec: corev1.ReplicationControllerSpec{
			Replicas: &replicas,
		},
	}

	// Pod with infeasible resize status
	pod := test.Pod().WithName("test-pod").WithCreator(&rc.ObjectMeta, &rc.TypeMeta).
		WithPodConditions([]corev1.PodCondition{
			{
				Type:   corev1.PodResizePending,
				Status: corev1.ConditionTrue,
				Reason: corev1.PodReasonInfeasible,
			},
		}).Get()
	pods := []*corev1.Pod{
		pod,
		test.Pod().WithName("test-pod-2").WithCreator(&rc.ObjectMeta, &rc.TypeMeta).Get(),
		test.Pod().WithName("test-pod-3").WithCreator(&rc.ObjectMeta, &rc.TypeMeta).Get(),
	}

	clock := baseclocktest.NewFakeClock(time.Time{})
	lipatm := map[string]time.Time{}

	vpa := getIPVpa()
	factory, err := getRestrictionFactory(&rc, nil, nil, nil, 2, tolerance, clock, lipatm, GetFakeCalculatorsWithFakeResourceCalc(), false)
	assert.NoError(t, err)
	creatorToSingleGroupStatsMap, podToReplicaCreatorMap, err := factory.GetCreatorMaps(pods, vpa)
	assert.NoError(t, err)
	inplace := factory.NewPodsInPlaceRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)

	// CanInPlaceUpdate should return InPlaceInfeasible
	decision := inplace.CanInPlaceUpdate(pod, vpa, nil)
	assert.Equal(t, utils.InPlaceInfeasible, decision,
		"InPlace mode should return InPlaceInfeasible for infeasible pods")

	// InPlaceUpdate should succeed for InPlaceInfeasible decision in InPlace mode
	err = inplace.InPlaceUpdate(pod, vpa, test.FakeEventRecorder())
	assert.NoError(t, err, "InPlace mode should allow retry for infeasible pods")
}
