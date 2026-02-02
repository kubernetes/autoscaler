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
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	featuregatetesting "k8s.io/component-base/featuregate/testing"
	baseclocktest "k8s.io/utils/clock/testing"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/features"
	utils "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/updater/utils"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
)

type CanInPlaceUpdateTestParams struct {
	name                    string
	pods                    []*apiv1.Pod
	replicas                int32
	evictionTolerance       float64
	lastInPlaceAttempt      time.Time
	expectedInPlaceDecision utils.InPlaceDecision
}

func TestCanInPlaceUpdate(t *testing.T) {
	featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.InPlaceOrRecreate, true)

	rc := apiv1.ReplicationController{
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
			pods: []*apiv1.Pod{
				generatePod().Get(),
				generatePod().Get(),
				generatePod().Get(),
			},
			replicas:                3,
			evictionTolerance:       0.5,
			lastInPlaceAttempt:      time.Time{},
			expectedInPlaceDecision: utils.InPlaceApproved,
		},
		{
			name: "CanInPlaceUpdate=InPlaceDeferred - no pods can be in-placed, one missing",
			pods: []*apiv1.Pod{
				generatePod().Get(),
				generatePod().Get(),
			},
			replicas:                3,
			evictionTolerance:       0.5,
			lastInPlaceAttempt:      time.Time{},
			expectedInPlaceDecision: utils.InPlaceDeferred,
		},
		{
			name: "CanInPlaceUpdate=InPlaceApproved - small tolerance, all running",
			pods: []*apiv1.Pod{
				generatePod().Get(),
				generatePod().Get(),
				generatePod().Get(),
			},
			replicas:                3,
			evictionTolerance:       0.1,
			lastInPlaceAttempt:      time.Time{},
			expectedInPlaceDecision: utils.InPlaceApproved,
		},
		{
			name: "CanInPlaceUpdate=InPlaceApproved - small tolerance, one missing",
			pods: []*apiv1.Pod{
				generatePod().Get(),
				generatePod().Get(),
			},
			replicas:                3,
			evictionTolerance:       0.5,
			lastInPlaceAttempt:      time.Time{},
			expectedInPlaceDecision: utils.InPlaceDeferred,
		},
		{
			name: "CanInPlaceUpdate=InPlaceDeferred - resize Deferred, conditions not met to fallback",
			pods: []*apiv1.Pod{
				generatePod().WithPodConditions([]apiv1.PodCondition{
					{
						Type:   apiv1.PodResizePending,
						Status: apiv1.ConditionTrue,
						Reason: apiv1.PodReasonDeferred,
					},
				}).Get(),
				generatePod().Get(),
				generatePod().Get(),
			},
			replicas:                3,
			evictionTolerance:       0.5,
			lastInPlaceAttempt:      time.UnixMilli(3600000), // 1 hour from epoch
			expectedInPlaceDecision: utils.InPlaceDeferred,
		},
		{
			name: ("CanInPlaceUpdate=InPlaceEvict - resize inProgress for more too long"),
			pods: []*apiv1.Pod{
				generatePod().WithPodConditions([]apiv1.PodCondition{
					{
						Type:   apiv1.PodResizeInProgress,
						Status: apiv1.ConditionTrue,
					},
				}).Get(),
				generatePod().Get(),
				generatePod().Get(),
			},
			replicas:                3,
			evictionTolerance:       0.5,
			lastInPlaceAttempt:      time.UnixMilli(0), // epoch (too long ago...)
			expectedInPlaceDecision: utils.InPlaceEvict,
		},
		{
			name: "CanInPlaceUpdate=InPlaceDeferred - resize InProgress, conditions not met to fallback",
			pods: []*apiv1.Pod{
				generatePod().WithPodConditions([]apiv1.PodCondition{
					{
						Type:   apiv1.PodResizeInProgress,
						Status: apiv1.ConditionTrue,
					},
				}).Get(),
				generatePod().Get(),
				generatePod().Get(),
			},
			replicas:                3,
			evictionTolerance:       0.5,
			lastInPlaceAttempt:      time.UnixMilli(3600000), // 1 hour from epoch
			expectedInPlaceDecision: utils.InPlaceDeferred,
		},
		{
			name: "CanInPlaceUpdate=InPlaceEvict - infeasible",
			pods: []*apiv1.Pod{
				generatePod().WithPodConditions([]apiv1.PodCondition{
					{
						Type:   apiv1.PodResizePending,
						Status: apiv1.ConditionTrue,
						Reason: apiv1.PodReasonInfeasible,
					},
				}).Get(),
				generatePod().Get(),
				generatePod().Get(),
			},
			replicas:                3,
			evictionTolerance:       0.5,
			lastInPlaceAttempt:      time.Time{},
			expectedInPlaceDecision: utils.InPlaceEvict,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rc.Spec = apiv1.ReplicationControllerSpec{
				Replicas: &tc.replicas,
			}

			selectedPod := tc.pods[whichPodIdxForCanInPlaceUpdate]

			clock := baseclocktest.NewFakeClock(time.UnixMilli(3600001)) // 1 hour from epoch + 1 millis
			lipatm := map[string]time.Time{getPodID(selectedPod): tc.lastInPlaceAttempt}

			factory, err := getRestrictionFactory(&rc, nil, nil, nil, 2, tc.evictionTolerance, clock, lipatm, GetFakeCalculatorsWithFakeResourceCalc())
			assert.NoError(t, err)
			creatorToSingleGroupStatsMap, podToReplicaCreatorMap, err := factory.GetCreatorMaps(tc.pods, getIPORVpa())
			assert.NoError(t, err)
			inPlace := factory.NewPodsInPlaceRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)

			result := inPlace.CanInPlaceUpdate(selectedPod)
			assert.Equal(t, tc.expectedInPlaceDecision, result)
		})
	}
}

func TestInPlaceDisabledFeatureGate(t *testing.T) {
	featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.InPlaceOrRecreate, false)

	replicas := int32(5)
	livePods := 5
	tolerance := 1.0

	rc := apiv1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rc",
			Namespace: "default",
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "ReplicationController",
		},
		Spec: apiv1.ReplicationControllerSpec{
			Replicas: &replicas,
		},
	}

	pods := make([]*apiv1.Pod, livePods)
	for i := range pods {
		pods[i] = test.Pod().WithName(getTestPodName(i)).WithCreator(&rc.ObjectMeta, &rc.TypeMeta).Get()
	}

	basicVpa := getBasicVpa()
	factory, err := getRestrictionFactory(&rc, nil, nil, nil, 2, tolerance, nil, nil, GetFakeCalculatorsWithFakeResourceCalc())
	assert.NoError(t, err)
	creatorToSingleGroupStatsMap, podToReplicaCreatorMap, err := factory.GetCreatorMaps(pods, basicVpa)
	assert.NoError(t, err)
	inplace := factory.NewPodsInPlaceRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)

	for _, pod := range pods {
		assert.Equal(t, utils.InPlaceEvict, inplace.CanInPlaceUpdate(pod))
	}
}

// TestInPlaceTooFewReplicas: with minReplicas=10 and only 5 live pods the group is belowMinReplicas.
// Eviction is blocked; in-place is still allowed. All pods get InPlaceApproved initially.
// In one "loop" we can only do as many in-place updates as evictionTolerance allows
// (isPodDisruptable() becomes false after that). With tolerance 0.5, evictionTolerance=2 → first 2 succeed, rest fail.
func TestInPlaceTooFewReplicas(t *testing.T) {
	featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.InPlaceOrRecreate, true)

	replicas := int32(5)
	livePods := 5
	tolerance := 0.5

	rc := apiv1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rc",
			Namespace: "default",
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "ReplicationController",
		},
		Spec: apiv1.ReplicationControllerSpec{
			Replicas: &replicas,
		},
	}

	pods := make([]*apiv1.Pod, livePods)
	for i := range pods {
		pods[i] = test.Pod().WithName(getTestPodName(i)).WithCreator(&rc.ObjectMeta, &rc.TypeMeta).Get()
	}

	clock := baseclocktest.NewFakeClock(time.Time{})
	lipatm := map[string]time.Time{}

	basicVpa := getIPORVpa()
	factory, err := getRestrictionFactory(&rc, nil, nil, nil, 10 /*minReplicas*/, tolerance, clock, lipatm, GetFakeCalculatorsWithFakeResourceCalc())
	assert.NoError(t, err)
	creatorToSingleGroupStatsMap, podToReplicaCreatorMap, err := factory.GetCreatorMaps(pods, basicVpa)
	assert.NoError(t, err)
	inplace := factory.NewPodsInPlaceRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)

	// belowMinReplicas does not block in-place: all pods are initially InPlaceApproved.
	for _, pod := range pods {
		assert.Equal(t, utils.InPlaceApproved, inplace.CanInPlaceUpdate(pod))
	}

	// In one loop, inPlaceUpdateInitiated is incremented after each successful update.
	// isPodDisruptable() requires actuallyAlive > shouldBeAlive; with tolerance 0.5, configured=5,
	// evictionTolerance=2 → after 2 updates actuallyAlive=3, shouldBeAlive=3 → no more allowed.
	allowedInOneLoop := 2
	for i, pod := range pods {
		err := inplace.InPlaceUpdate(pod, basicVpa, test.FakeEventRecorder())
		if i < allowedInOneLoop {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err, "Expected error after %d in-place updates in one loop (tolerance limit)", allowedInOneLoop)
		}
	}
}

func TestEvictionToleranceForInPlace(t *testing.T) {
	featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.InPlaceOrRecreate, true)

	replicas := int32(5)
	livePods := 5
	tolerance := 0.8

	rc := apiv1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rc",
			Namespace: "default",
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "ReplicationController",
		},
		Spec: apiv1.ReplicationControllerSpec{
			Replicas: &replicas,
		},
	}

	pods := make([]*apiv1.Pod, livePods)
	for i := range pods {
		pods[i] = test.Pod().WithName(getTestPodName(i)).WithCreator(&rc.ObjectMeta, &rc.TypeMeta).Get()
	}

	clock := baseclocktest.NewFakeClock(time.Time{})
	lipatm := map[string]time.Time{}

	basicVpa := getIPORVpa()
	factory, err := getRestrictionFactory(&rc, nil, nil, nil, 2 /*minReplicas*/, tolerance, clock, lipatm, GetFakeCalculatorsWithFakeResourceCalc())
	assert.NoError(t, err)
	creatorToSingleGroupStatsMap, podToReplicaCreatorMap, err := factory.GetCreatorMaps(pods, basicVpa)
	assert.NoError(t, err)
	inplace := factory.NewPodsInPlaceRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)

	for _, pod := range pods {
		assert.Equal(t, utils.InPlaceApproved, inplace.CanInPlaceUpdate(pod))
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

func TestInPlaceAtLeastOne(t *testing.T) {
	featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.InPlaceOrRecreate, true)

	replicas := int32(5)
	livePods := 5
	tolerance := 0.1

	rc := apiv1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rc",
			Namespace: "default",
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "ReplicationController",
		},
		Spec: apiv1.ReplicationControllerSpec{
			Replicas: &replicas,
		},
	}

	pods := make([]*apiv1.Pod, livePods)
	for i := range pods {
		pods[i] = test.Pod().WithName(getTestPodName(i)).WithCreator(&rc.ObjectMeta, &rc.TypeMeta).Get()
	}

	basicVpa := getBasicVpa()
	factory, err := getRestrictionFactory(&rc, nil, nil, nil, 2, tolerance, nil, nil, GetFakeCalculatorsWithFakeResourceCalc())
	assert.NoError(t, err)
	creatorToSingleGroupStatsMap, podToReplicaCreatorMap, err := factory.GetCreatorMaps(pods, basicVpa)
	assert.NoError(t, err)
	inplace := factory.NewPodsInPlaceRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)

	for _, pod := range pods {
		assert.Equal(t, utils.InPlaceApproved, inplace.CanInPlaceUpdate(pod))
	}

	for _, pod := range pods[:1] {
		err := inplace.InPlaceUpdate(pod, basicVpa, test.FakeEventRecorder())
		assert.Nil(t, err, "Should in-place update with no error")
	}
	for _, pod := range pods[1:] {
		err := inplace.InPlaceUpdate(pod, basicVpa, test.FakeEventRecorder())
		assert.Error(t, err, "Error expected")
	}
}

func TestMinReplicasAllowsInPlaceForSingletonButBlocksEviction(t *testing.T) {
	featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.InPlaceOrRecreate, true)

	// replica controller reports replicas = 1
	replicas := int32(1)
	livePods := 1
	tolerance := 0.5

	rc := apiv1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rc-minrep-singleton",
			Namespace: "default",
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "ReplicationController",
		},
		Spec: apiv1.ReplicationControllerSpec{
			Replicas: &replicas,
		},
	}

	// create exactly one pod (livePods = 1)
	pods := make([]*apiv1.Pod, livePods)
	for i := range pods {
		pods[i] = test.Pod().WithName(getTestPodName(i)).WithCreator(&rc.ObjectMeta, &rc.TypeMeta).Get()
	}

	// fake clock / no prior in-place attempts
	clock := baseclocktest.NewFakeClock(time.Time{})
	lipatm := map[string]time.Time{}

	// VPA configured with InPlaceOrRecreate (helper provides this)
	iporVpa := getIPORVpa()

	// minReplicas = 2: group is below minReplicas, but we still include it with belowMinReplicas=true.
	// In-place is allowed; eviction is blocked.
	factory, err := getRestrictionFactory(&rc, nil, nil, nil, 2 /*minReplicas*/, tolerance, clock, lipatm, GetFakeCalculatorsWithFakeResourceCalc())
	assert.NoError(t, err)

	creatorToSingleGroupStatsMap, podToReplicaCreatorMap, err := factory.GetCreatorMaps(pods, iporVpa)
	assert.NoError(t, err)

	inplace := factory.NewPodsInPlaceRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)
	eviction := factory.NewPodsEvictionRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)

	// In-place is allowed even when livePods < minReplicas (belowMinReplicas only blocks eviction).
	for _, pod := range pods {
		assert.Equal(t, utils.InPlaceApproved, inplace.CanInPlaceUpdate(pod))
	}
	// Eviction is blocked when below minReplicas (fallback to Recreate would do nothing).
	for _, pod := range pods {
		assert.False(t, eviction.CanEvict(pod))
	}

	// In-place update succeeds.
	for _, pod := range pods {
		err := inplace.InPlaceUpdate(pod, iporVpa, test.FakeEventRecorder())
		assert.NoError(t, err)
	}
}
