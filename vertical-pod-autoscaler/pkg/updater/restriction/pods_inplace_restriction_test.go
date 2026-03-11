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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/version"
	"k8s.io/client-go/tools/record"
	featuregatetesting "k8s.io/component-base/featuregate/testing"
	baseclocktest "k8s.io/utils/clock/testing"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/features"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/updater/utils"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
)

type CanInPlaceUpdateTestParams struct {
	name                    string
	pods                    []*corev1.Pod
	replicas                int32
	evictionTolerance       float64
	lastInPlaceAttempt      time.Time
	expectedInPlaceDecision utils.InPlaceDecision
	vpa                     *vpa_types.VerticalPodAutoscaler
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
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rc.Spec = corev1.ReplicationControllerSpec{
				Replicas: &tc.replicas,
			}

			selectedPod := tc.pods[whichPodIdxForCanInPlaceUpdate]

			updateMode := vpa_api_util.GetUpdateMode(tc.vpa)

			clock := baseclocktest.NewFakeClock(time.UnixMilli(3600001)) // 1 hour from epoch + 1 millis
			lipatm := map[string]time.Time{getPodID(selectedPod): tc.lastInPlaceAttempt}

			factory, err := getRestrictionFactory(&rc, nil, nil, nil, 2, tc.evictionTolerance, clock, lipatm, GetFakeCalculatorsWithFakeResourceCalc(), false)
			assert.NoError(t, err)
			creatorToSingleGroupStatsMap, podToReplicaCreatorMap, err := factory.GetCreatorMaps(tc.pods, getIPORVpa())
			assert.NoError(t, err)
			inPlace := factory.NewPodsInPlaceRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)

			result := inPlace.CanInPlaceUpdate(selectedPod, updateMode)
			assert.Equal(t, tc.expectedInPlaceDecision, result)
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
	updateMode := vpa_api_util.GetUpdateMode(basicVpa)
	factory, err := getRestrictionFactory(&rc, nil, nil, nil, 2, tolerance, nil, nil, GetFakeCalculatorsWithFakeResourceCalc(), false)
	assert.NoError(t, err)
	creatorToSingleGroupStatsMap, podToReplicaCreatorMap, err := factory.GetCreatorMaps(pods, basicVpa)
	assert.NoError(t, err)
	inplace := factory.NewPodsInPlaceRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)

	for _, pod := range pods {
		assert.Equal(t, utils.InPlaceEvict, inplace.CanInPlaceUpdate(pod, updateMode))
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
	updateMode := vpa_api_util.GetUpdateMode(basicVpa)
	factory, err := getRestrictionFactory(&rc, nil, nil, nil, 10 /* minReplicas */, tolerance, clock, lipatm, GetFakeCalculatorsWithFakeResourceCalc(), false)
	assert.NoError(t, err)
	creatorToSingleGroupStatsMap, podToReplicaCreatorMap, err := factory.GetCreatorMaps(pods, basicVpa)
	assert.NoError(t, err)
	inplace := factory.NewPodsInPlaceRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)

	for _, pod := range pods {
		assert.Equal(t, utils.InPlaceDeferred, inplace.CanInPlaceUpdate(pod, updateMode))
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
	updateMode := vpa_api_util.GetUpdateMode(basicVpa)
	factory, err := getRestrictionFactory(&rc, nil, nil, nil, 2 /* minReplicas */, tolerance, clock, lipatm, GetFakeCalculatorsWithFakeResourceCalc(), false)
	assert.NoError(t, err)
	creatorToSingleGroupStatsMap, podToReplicaCreatorMap, err := factory.GetCreatorMaps(pods, basicVpa)
	assert.NoError(t, err)
	inplace := factory.NewPodsInPlaceRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)

	for _, pod := range pods {
		assert.Equal(t, utils.InPlaceApproved, inplace.CanInPlaceUpdate(pod, updateMode))
	}

	for _, pod := range pods[:4] {
		err := inplace.InPlaceUpdate(pod, basicVpa, test.FakeEventRecorder())
		assert.Nil(t, err, "Should evict with no error")
	}
	for _, pod := range pods[4:] {
		err := inplace.InPlaceUpdate(pod, basicVpa, test.FakeEventRecorder())
		assert.Error(t, err, "Error expected")
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
	updateMode := vpa_api_util.GetUpdateMode(basicVpa)

	inplace := factory.NewPodsInPlaceRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)

	// All in-place updates should be approved
	for _, pod := range pods {
		assert.Equal(t, utils.InPlaceApproved, inplace.CanInPlaceUpdate(pod, updateMode))
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
		assert.Equal(t, utils.InPlaceApproved, inplace.CanInPlaceUpdate(pod, *basicVpa.Spec.UpdatePolicy.UpdateMode))
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
	updateMode := vpa_api_util.GetUpdateMode(inPlaceOrRecreateVPA)
	factory, err := getRestrictionFactory(&rc, nil, nil, nil, 2, tolerance, nil, nil, GetFakeCalculatorsWithFakeResourceCalc(), false)
	assert.NoError(t, err)
	creatorToSingleGroupStatsMap, podToReplicaCreatorMap, err := factory.GetCreatorMaps(pods, inPlaceOrRecreateVPA)
	assert.NoError(t, err)
	inplace := factory.NewPodsInPlaceRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)

	for _, pod := range pods {
		assert.Equal(t, utils.InPlaceApproved, inplace.CanInPlaceUpdate(pod, updateMode))
	}

	for _, pod := range pods[:1] {
		err := inplace.InPlaceUpdate(pod, inPlaceOrRecreateVPA, test.FakeEventRecorder())
		assert.Nil(t, err, "Should in-place update with no error")
	}
	for _, pod := range pods[1:] {
		err := inplace.InPlaceUpdate(pod, inPlaceOrRecreateVPA, test.FakeEventRecorder())
		assert.Error(t, err, "Error expected")
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
	updateMode := vpa_api_util.GetUpdateMode(inPlaceVpa)
	factory, err := getRestrictionFactory(&rc, nil, nil, nil, 2, tolerance, nil, nil, GetFakeCalculatorsWithFakeResourceCalc(), false)
	assert.NoError(t, err)
	creatorToSingleGroupStatsMap, podToReplicaCreatorMap, err := factory.GetCreatorMaps(pods, inPlaceVpa)
	assert.NoError(t, err)
	inplace := factory.NewPodsInPlaceRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)

	for _, pod := range pods {
		assert.Equal(t, utils.InPlaceDeferred, inplace.CanInPlaceUpdate(pod, updateMode),
			"InPlace mode should return InPlaceDeferred when feature gate is disabled")
	}
}

func TestInPlaceModeWaitsIndefinitely(t *testing.T) {
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

	// Create a pod that's been updating for a very long time
	pods := []*corev1.Pod{
		test.Pod().WithName("test-1").WithCreator(&rc.ObjectMeta, &rc.TypeMeta).
			WithPodConditions([]corev1.PodCondition{
				{
					Type:   corev1.PodResizeInProgress,
					Status: corev1.ConditionTrue,
				},
			}).Get(),
		test.Pod().WithName("test-2").WithCreator(&rc.ObjectMeta, &rc.TypeMeta).Get(),
		test.Pod().WithName("test-3").WithCreator(&rc.ObjectMeta, &rc.TypeMeta).Get(),
	}

	inPlaceVpa := getIPVpa()
	updateMode := vpa_api_util.GetUpdateMode(inPlaceVpa)

	// Set last attempt to a very old time (would timeout in InPlaceOrRecreate mode)
	clock := baseclocktest.NewFakeClock(time.UnixMilli(7200000))         // 2 hours from epoch
	lipatm := map[string]time.Time{getPodID(pods[0]): time.UnixMilli(0)} // epoch (very old)

	factory, err := getRestrictionFactory(&rc, nil, nil, nil, 2, tolerance, clock, lipatm, GetFakeCalculatorsWithFakeResourceCalc(), false)
	assert.NoError(t, err)
	creatorToSingleGroupStatsMap, podToReplicaCreatorMap, err := factory.GetCreatorMaps(pods, inPlaceVpa)
	assert.NoError(t, err)
	inplace := factory.NewPodsInPlaceRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)

	// InPlace mode should return Deferred even after a very long time
	result := inplace.CanInPlaceUpdate(pods[0], updateMode)
	assert.Equal(t, utils.InPlaceDeferred, result,
		"InPlace mode should wait indefinitely for in-progress updates, not timeout")
}

func TestInPlaceModeInfeasible(t *testing.T) {
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

	// Create a pod with infeasible resize condition
	pods := []*corev1.Pod{
		test.Pod().WithName("test-1").WithCreator(&rc.ObjectMeta, &rc.TypeMeta).
			WithPodConditions([]corev1.PodCondition{
				{
					Type:   corev1.PodResizePending,
					Status: corev1.ConditionTrue,
					Reason: corev1.PodReasonInfeasible,
				},
			}).Get(),
		test.Pod().WithName("test-2").WithCreator(&rc.ObjectMeta, &rc.TypeMeta).Get(),
		test.Pod().WithName("test-3").WithCreator(&rc.ObjectMeta, &rc.TypeMeta).Get(),
	}

	inPlaceVpa := getIPVpa()
	updateMode := vpa_api_util.GetUpdateMode(inPlaceVpa)

	clock := baseclocktest.NewFakeClock(time.Time{})
	lipatm := map[string]time.Time{}

	factory, err := getRestrictionFactory(&rc, nil, nil, nil, 2, tolerance, clock, lipatm, GetFakeCalculatorsWithFakeResourceCalc(), false)
	assert.NoError(t, err)
	creatorToSingleGroupStatsMap, podToReplicaCreatorMap, err := factory.GetCreatorMaps(pods, inPlaceVpa)
	assert.NoError(t, err)
	inplace := factory.NewPodsInPlaceRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)

	// Infeasible updates should still return InPlaceDeferred in InPlace mode
	result := inplace.CanInPlaceUpdate(pods[0], updateMode)
	assert.Equal(t, utils.InPlaceInfeasible, result,
		"InPlace mode should return InPlaceInfeasible for infeasible updates")
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
	updateMode := vpa_api_util.GetUpdateMode(inPlaceVpa)
	factory, err := getRestrictionFactory(&rc, nil, nil, nil, 2, tolerance, nil, nil, GetFakeCalculatorsWithFakeResourceCalc(), false)
	assert.NoError(t, err)
	creatorToSingleGroupStatsMap, podToReplicaCreatorMap, err := factory.GetCreatorMaps(pods, inPlaceVpa)
	assert.NoError(t, err)
	inplace := factory.NewPodsInPlaceRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)

	for _, pod := range pods {
		assert.Equal(t, utils.InPlaceApproved, inplace.CanInPlaceUpdate(pod, updateMode))
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

			result := inplace.CanInPlaceUpdate(pod, tc.updateMode)
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
	decision := inplace.CanInPlaceUpdate(pod, vpa_types.UpdateModeInPlace)
	assert.Equal(t, utils.InPlaceInfeasible, decision,
		"InPlace mode should return InPlaceInfeasible for infeasible pods")

	// InPlaceUpdate should succeed for InPlaceInfeasible decision in InPlace mode
	err = inplace.InPlaceUpdate(pod, vpa, test.FakeEventRecorder())
	assert.NoError(t, err, "InPlace mode should allow retry for infeasible pods")
}

func TestDeferredPodWithChangedRecommendationsRetries(t *testing.T) {
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

	// Create a pod in deferred state - this simulates a pod where kubelet deferred the resize
	// but the VPA has new recommendations that differ from the current spec
	deferredPod := test.Pod().WithName("deferred-pod").WithCreator(&rc.ObjectMeta, &rc.TypeMeta).
		WithPodConditions([]corev1.PodCondition{
			{
				Type:   corev1.PodResizePending,
				Status: corev1.ConditionTrue,
				Reason: corev1.PodReasonDeferred,
			},
		}).Get()

	pods := []*corev1.Pod{
		deferredPod,
		test.Pod().WithName("test-pod-2").WithCreator(&rc.ObjectMeta, &rc.TypeMeta).Get(),
		test.Pod().WithName("test-pod-3").WithCreator(&rc.ObjectMeta, &rc.TypeMeta).Get(),
	}

	vpa := getIPVpa()
	updateMode := vpa_api_util.GetUpdateMode(vpa)

	clock := baseclocktest.NewFakeClock(time.Time{})
	lipatm := map[string]time.Time{}

	factory, err := getRestrictionFactory(&rc, nil, nil, nil, 2, tolerance, clock, lipatm, GetFakeCalculatorsWithFakeResourceCalc(), false)
	assert.NoError(t, err)
	creatorToSingleGroupStatsMap, podToReplicaCreatorMap, err := factory.GetCreatorMaps(pods, vpa)
	assert.NoError(t, err)
	inplace := factory.NewPodsInPlaceRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)

	// Verify CanInPlaceUpdate returns InPlaceDeferred for the deferred pod
	decision := inplace.CanInPlaceUpdate(deferredPod, updateMode)
	assert.Equal(t, utils.InPlaceDeferred, decision,
		"Deferred pod should return InPlaceDeferred decision")

	// When recommendations have changed (which is determined by the priority calculator
	// including the pod in podsForInPlace), InPlaceUpdate should succeed as a retry.
	// The InPlaceUpdate method itself doesn't check the deferred status - it just
	// applies the patch if allowed by the restriction limits.
	err = inplace.InPlaceUpdate(deferredPod, vpa, test.FakeEventRecorder())
	assert.NoError(t, err,
		"InPlaceUpdate should succeed for deferred pod when called (simulating recommendations changed)")
}

func TestDeferredPodWithChangedRecommendationsInPlaceOrRecreate(t *testing.T) {
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

	// Create a pod in deferred state - this simulates a pod where kubelet deferred the resize
	// but the VPA has new recommendations that differ from the current spec
	deferredPod := test.Pod().WithName("deferred-pod").WithCreator(&rc.ObjectMeta, &rc.TypeMeta).
		WithPodConditions([]corev1.PodCondition{
			{
				Type:   corev1.PodResizePending,
				Status: corev1.ConditionTrue,
				Reason: corev1.PodReasonDeferred,
			},
		}).Get()

	pods := []*corev1.Pod{
		deferredPod,
		test.Pod().WithName("test-pod-2").WithCreator(&rc.ObjectMeta, &rc.TypeMeta).Get(),
		test.Pod().WithName("test-pod-3").WithCreator(&rc.ObjectMeta, &rc.TypeMeta).Get(),
	}

	vpa := getIPORVpa()
	updateMode := vpa_api_util.GetUpdateMode(vpa)

	clock := baseclocktest.NewFakeClock(time.Time{})
	lipatm := map[string]time.Time{}

	factory, err := getRestrictionFactory(&rc, nil, nil, nil, 2, tolerance, clock, lipatm, GetFakeCalculatorsWithFakeResourceCalc(), false)
	assert.NoError(t, err)
	creatorToSingleGroupStatsMap, podToReplicaCreatorMap, err := factory.GetCreatorMaps(pods, vpa)
	assert.NoError(t, err)
	inplace := factory.NewPodsInPlaceRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)

	// Verify CanInPlaceUpdate returns InPlaceDeferred for the deferred pod
	decision := inplace.CanInPlaceUpdate(deferredPod, updateMode)
	assert.Equal(t, utils.InPlaceDeferred, decision,
		"Deferred pod should return InPlaceDeferred decision")

	// When recommendations have changed (which is determined by the priority calculator
	// including the pod in podsForInPlace), InPlaceUpdate should succeed as a retry.
	// The InPlaceUpdate method itself doesn't check the deferred status - it just
	// applies the patch if allowed by the restriction limits.
	err = inplace.InPlaceUpdate(deferredPod, vpa, test.FakeEventRecorder())
	assert.Error(t, err, "cannot in-place update pod default/deferred-pod, decision: InPlaceDeferred")
}
