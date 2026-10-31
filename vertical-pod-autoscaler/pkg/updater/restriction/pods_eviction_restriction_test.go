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

package restriction

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	baseclocktest "k8s.io/utils/clock/testing"
	"k8s.io/utils/set"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
)

func TestEvictTooFewReplicas(t *testing.T) {
	replicas := int32(5)
	livePods := 5

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
	factory, err := getRestrictionFactory(&rc, nil, nil, nil, 10, 0.5, nil, nil, nil, false)
	assert.NoError(t, err)
	creatorToSingleGroupStatsMap, podToReplicaCreatorMap, err := factory.GetCreatorMaps(pods, basicVpa)
	assert.NoError(t, err)
	eviction := factory.NewPodsEvictionRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)

	for _, pod := range pods {
		assert.False(t, eviction.CanEvict(pod))
	}

	for _, pod := range pods {
		err := eviction.Evict(pod, basicVpa, test.FakeEventRecorder())
		assert.Error(t, err, "Error expected")
	}
}

func TestEvictionTolerance(t *testing.T) {
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

	basicVpa := getBasicVpa()
	factory, err := getRestrictionFactory(&rc, nil, nil, nil, 2 /* minReplicas */, tolerance, nil, nil, nil, false)
	assert.NoError(t, err)
	creatorToSingleGroupStatsMap, podToReplicaCreatorMap, err := factory.GetCreatorMaps(pods, basicVpa)
	assert.NoError(t, err)
	eviction := factory.NewPodsEvictionRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)

	for _, pod := range pods {
		assert.True(t, eviction.CanEvict(pod))
	}

	for _, pod := range pods[:4] {
		err := eviction.Evict(pod, basicVpa, test.FakeEventRecorder())
		assert.Nil(t, err, "Should evict with no error")
	}
	for _, pod := range pods[4:] {
		err := eviction.Evict(pod, basicVpa, test.FakeEventRecorder())
		assert.Error(t, err, "Error expected")
	}
}

func TestEvictAtLeastOne(t *testing.T) {
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

	basicVpa := getBasicVpa()
	factory, err := getRestrictionFactory(&rc, nil, nil, nil, 2, tolerance, nil, nil, nil, false)
	assert.NoError(t, err)
	creatorToSingleGroupStatsMap, podToReplicaCreatorMap, err := factory.GetCreatorMaps(pods, basicVpa)
	assert.NoError(t, err)
	eviction := factory.NewPodsEvictionRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)

	for _, pod := range pods {
		assert.True(t, eviction.CanEvict(pod))
	}

	for _, pod := range pods[:1] {
		err := eviction.Evict(pod, basicVpa, test.FakeEventRecorder())
		assert.Nil(t, err, "Should evict with no error")
	}
	for _, pod := range pods[1:] {
		err := eviction.Evict(pod, basicVpa, test.FakeEventRecorder())
		assert.Error(t, err, "Error expected")
	}
}

func TestEvictEmitEvent(t *testing.T) {
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

	basicVpa := getBasicVpa()

	testCases := []struct {
		name              string
		replicas          int32
		evictionTolerance float64
		vpa               *vpa_types.VerticalPodAutoscaler
		pods              []podWithExpectations
		errorExpected     bool
	}{
		{
			name:              "Pods that can be evicted",
			replicas:          4,
			evictionTolerance: 0.5,
			vpa:               basicVpa,
			pods: []podWithExpectations{
				{
					pod:             generatePod().WithPhase(corev1.PodPending).Get(),
					canEvict:        true,
					evictionSuccess: true,
				},
				{
					pod:             generatePod().WithPhase(corev1.PodPending).Get(),
					canEvict:        true,
					evictionSuccess: true,
				},
			},
			errorExpected: false,
		},
		{
			name:              "Pod that can not be evicted",
			replicas:          4,
			evictionTolerance: 0.5,
			vpa:               basicVpa,
			pods: []podWithExpectations{

				{
					pod:             generatePod().Get(),
					canEvict:        false,
					evictionSuccess: false,
				},
			},
			errorExpected: true,
		},
	}

	for _, testCase := range testCases {
		rc.Spec = corev1.ReplicationControllerSpec{
			Replicas: &testCase.replicas,
		}
		pods := make([]*corev1.Pod, 0, len(testCase.pods))
		for _, p := range testCase.pods {
			pods = append(pods, p.pod)
		}
		clock := baseclocktest.NewFakeClock(time.Time{})
		factory, err := getRestrictionFactory(&rc, nil, nil, nil, 2, testCase.evictionTolerance, clock, map[string]time.Time{}, nil, false)
		assert.NoError(t, err)
		creatorToSingleGroupStatsMap, podToReplicaCreatorMap, err := factory.GetCreatorMaps(pods, testCase.vpa)
		assert.NoError(t, err)
		eviction := factory.NewPodsEvictionRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)

		for _, p := range testCase.pods {
			mockRecorder := test.MockEventRecorder()
			mockRecorder.On("Event", mock.Anything, corev1.EventTypeNormal, "EvictedByVPA", mock.Anything).Return()
			mockRecorder.On("Event", mock.Anything, corev1.EventTypeNormal, "EvictedPod", mock.Anything).Return()

			errGot := eviction.Evict(p.pod, testCase.vpa, mockRecorder)
			if testCase.errorExpected {
				assert.Error(t, errGot)
			} else {
				assert.NoError(t, errGot)
			}

			if p.canEvict {
				mockRecorder.AssertNumberOfCalls(t, "Event", 2)
			} else {
				mockRecorder.AssertNumberOfCalls(t, "Event", 0)
			}
		}
	}
}

// This test ensures that in-place-skip-disruption-budget only affects in-place
// updates and does not bypass eviction tolerance when performing pod evictions.
func TestEvictTooFewReplicasWithInPlaceSkipDisruptionBudget(t *testing.T) {
	replicas := int32(5)
	livePods := 5

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
	// factory with inPlaceSkipDisruptionBudget on
	factory, err := getRestrictionFactory(&rc, nil, nil, nil, 10, 0.5, nil, nil, nil, true)
	assert.NoError(t, err)
	creatorToSingleGroupStatsMap, podToReplicaCreatorMap, err := factory.GetCreatorMaps(pods, basicVpa)
	assert.NoError(t, err)
	eviction := factory.NewPodsEvictionRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)

	for _, pod := range pods {
		assert.False(t, eviction.CanEvict(pod))
	}

	for _, pod := range pods {
		err := eviction.Evict(pod, basicVpa, test.FakeEventRecorder())
		assert.Error(t, err, "Error expected")
	}
}

func TestEvictionRequestCreation(t *testing.T) {
	replicas := int32(5)
	livePods := 5

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
		pods[i].UID = types.UID(fmt.Sprintf("pod-uid-%d", i))
	}

	basicVpa := getBasicVpa()
	factory, err := getRestrictionFactory(&rc, nil, nil, nil, 2, 0.5, nil, nil, nil, false)
	assert.NoError(t, err)

	factory.SetEvictionRequestOptions(true, "vpa.autoscaling.k8s.io/updater")
	creatorToSingleGroupStatsMap, podToReplicaCreatorMap, err := factory.GetCreatorMaps(pods, basicVpa)
	assert.NoError(t, err)

	eviction := factory.NewPodsEvictionRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)

	// CanEvict should be true initially
	assert.True(t, eviction.CanEvict(pods[0]))

	// Evict should create an EvictionRequest
	err = eviction.Evict(pods[0], basicVpa, test.FakeEventRecorder())
	assert.NoError(t, err)

	// Pod 0 now has an active EvictionRequest, so CanEvict should return false for Pod 0
	assert.False(t, eviction.CanEvict(pods[0]))
}

func TestEvictionRequestExcludeInFlightPod(t *testing.T) {
	replicas := int32(5)
	livePods := 5

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
		pods[i].UID = types.UID(fmt.Sprintf("pod-uid-%d", i))
	}

	basicVpa := getBasicVpa()
	factory, err := getRestrictionFactory(&rc, nil, nil, nil, 2, 0.5, nil, nil, nil, false)
	assert.NoError(t, err)

	// Pre-populate active EvictionRequest for pod 0
	activeUIDs := set.New[types.UID](pods[0].UID)
	factory.SetOutstandingEvictionRequests(activeUIDs)
	factory.SetEvictionRequestOptions(true, "vpa.autoscaling.k8s.io/updater")

	creatorToSingleGroupStatsMap, podToReplicaCreatorMap, err := factory.GetCreatorMaps(pods, basicVpa)
	assert.NoError(t, err)

	eviction := factory.NewPodsEvictionRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)

	// Pod 0 already has an active EvictionRequest: cannot evict
	assert.False(t, eviction.CanEvict(pods[0]))
	// Other pods without EvictionRequests can still be evaluated
	assert.True(t, eviction.CanEvict(pods[1]))
}

func TestAsynchronousEvictionBudgetAccounting(t *testing.T) {
	// Replicas = 4, tolerance = 0.5 (configured: 4, tolerance: 2, shouldBeAlive: 2)
	// If 2 pods already have outstanding EvictionRequests from previous loops:
	// actuallyAlive = 4 - 2 = 2.
	// Since 2 is not > 2, further evictions must be blocked!
	replicas := int32(4)
	livePods := 4

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
		pods[i].UID = types.UID(fmt.Sprintf("pod-uid-%d", i))
	}

	basicVpa := getBasicVpa()
	factory, err := getRestrictionFactory(&rc, nil, nil, nil, 2, 0.5, nil, nil, nil, false)
	assert.NoError(t, err)

	// Mark Pods 0 and 1 as having outstanding EvictionRequests
	activeUIDs := set.New[types.UID](pods[0].UID, pods[1].UID)
	factory.SetOutstandingEvictionRequests(activeUIDs)
	factory.SetEvictionRequestOptions(true, "vpa.autoscaling.k8s.io/updater")

	creatorToSingleGroupStatsMap, podToReplicaCreatorMap, err := factory.GetCreatorMaps(pods, basicVpa)
	assert.NoError(t, err)

	// Verify stats accounted for 2 outstanding requests
	cr := podToReplicaCreatorMap[getPodID(pods[0])]
	stats := creatorToSingleGroupStatsMap[cr]
	assert.Equal(t, 2, stats.outstandingEvictionRequests)

	eviction := factory.NewPodsEvictionRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)

	// Pods 0 and 1 cannot be evicted because they are already in flight
	assert.False(t, eviction.CanEvict(pods[0]))
	assert.False(t, eviction.CanEvict(pods[1]))

	// Pods 2 and 3 must also be blocked from eviction because the disruption budget is exhausted
	assert.False(t, eviction.CanEvict(pods[2]))
	assert.False(t, eviction.CanEvict(pods[3]))
}
