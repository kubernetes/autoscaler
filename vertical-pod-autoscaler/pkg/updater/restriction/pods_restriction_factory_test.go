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
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsinformer "k8s.io/client-go/informers/apps/v1"
	coreinformer "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
	featuregatetesting "k8s.io/component-base/featuregate/testing"
	"k8s.io/utils/clock"
	baseclocktest "k8s.io/utils/clock/testing"
	"k8s.io/utils/ptr"

	resource_admission "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource/pod/patch"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/features"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/updater/utils"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
)

type podWithExpectations struct {
	pod                  *apiv1.Pod
	canEvict             bool
	evictionSuccess      bool
	canInPlaceUpdate     utils.InPlaceDecision
	inPlaceUpdateSuccess bool
}

func getBasicVpa() *vpa_types.VerticalPodAutoscaler {
	return test.VerticalPodAutoscaler().WithContainer("any").Get()
}

func getIPORVpa() *vpa_types.VerticalPodAutoscaler {
	vpa := getBasicVpa()
	vpa.Spec.UpdatePolicy = &vpa_types.PodUpdatePolicy{
		UpdateMode: ptr.To(vpa_types.UpdateModeInPlaceOrRecreate),
	}
	return vpa
}

func TestDisruptReplicatedByController(t *testing.T) {
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

	vpaSingleReplica := getBasicVpa()
	minReplicas := int32(1)
	vpaSingleReplica.Spec.UpdatePolicy = &vpa_types.PodUpdatePolicy{MinReplicas: &minReplicas}

	index := 0
	generatePod := func() test.PodBuilder {
		index++
		return test.Pod().WithName(fmt.Sprintf("test-%v", index)).WithCreator(&rc.ObjectMeta, &rc.TypeMeta)
	}

	testCases := []struct {
		name              string
		replicas          int32
		evictionTolerance float64
		vpa               *vpa_types.VerticalPodAutoscaler
		pods              []podWithExpectations
	}{
		{
			name:              "Evict only first pod (half of 3).",
			replicas:          3,
			evictionTolerance: 0.5,
			vpa:               getBasicVpa(),
			pods: []podWithExpectations{
				{
					pod:             generatePod().Get(),
					canEvict:        true,
					evictionSuccess: true,
				},
				{
					pod:             generatePod().Get(),
					canEvict:        true,
					evictionSuccess: false,
				},
				{
					pod:             generatePod().Get(),
					canEvict:        true,
					evictionSuccess: false,
				},
			},
		},
		{
			name:              "Evict two pods (half of 4).",
			replicas:          4,
			evictionTolerance: 0.5,
			vpa:               getBasicVpa(),
			pods: []podWithExpectations{
				{

					pod:             generatePod().Get(),
					canEvict:        true,
					evictionSuccess: true,
				},
				{
					pod:             generatePod().Get(),
					canEvict:        true,
					evictionSuccess: true,
				},
				{
					pod:             generatePod().Get(),
					canEvict:        true,
					evictionSuccess: false,
				},
				{
					pod:             generatePod().Get(),
					canEvict:        true,
					evictionSuccess: false,
				},
			},
		},
		{
			name:              "Half of the population can be evicted. One pod is missing already.",
			replicas:          4,
			evictionTolerance: 0.5,
			vpa:               getBasicVpa(),
			pods: []podWithExpectations{
				{
					pod:             generatePod().Get(),
					canEvict:        true,
					evictionSuccess: true,
				},
				{
					pod:             generatePod().Get(),
					canEvict:        true,
					evictionSuccess: false,
				},
				{
					pod:             generatePod().Get(),
					canEvict:        true,
					evictionSuccess: false,
				},
			},
		},
		{
			name:              "For small eviction tolerance at least one pod is evicted.",
			replicas:          3,
			evictionTolerance: 0.1,
			vpa:               getBasicVpa(),
			pods: []podWithExpectations{
				{
					pod:             generatePod().Get(),
					canEvict:        true,
					evictionSuccess: true,
				},
				{
					pod:             generatePod().Get(),
					canEvict:        true,
					evictionSuccess: false,
				},
				{
					pod:             generatePod().Get(),
					canEvict:        true,
					evictionSuccess: false,
				},
			},
		},
		{
			name:              "Only 2 pods in replica of 3 and tolerance is 0. None of pods can be evicted.",
			replicas:          3,
			evictionTolerance: 0.1,
			vpa:               getBasicVpa(),
			pods: []podWithExpectations{
				{
					pod:             generatePod().Get(),
					canEvict:        false,
					evictionSuccess: false,
				},
				{
					pod:             generatePod().Get(),
					canEvict:        false,
					evictionSuccess: false,
				},
			},
		},
		{
			name:              "Only pending pod can be evicted without violation of tolerance.",
			replicas:          3,
			evictionTolerance: 0.5,
			vpa:               getBasicVpa(),
			pods: []podWithExpectations{
				{
					pod:             generatePod().Get(),
					canEvict:        false,
					evictionSuccess: false,
				},
				{
					pod:             generatePod().WithPhase(apiv1.PodPending).Get(),
					canEvict:        true,
					evictionSuccess: true,
				},
				{
					pod:             generatePod().Get(),
					canEvict:        false,
					evictionSuccess: false,
				},
			},
		},
		{
			name:              "Pending pods are always evictable.",
			replicas:          4,
			evictionTolerance: 0.5,
			vpa:               getBasicVpa(),
			pods: []podWithExpectations{
				{
					pod:             generatePod().Get(),
					canEvict:        false,
					evictionSuccess: false,
				},
				{
					pod:             generatePod().WithPhase(apiv1.PodPending).Get(),
					canEvict:        true,
					evictionSuccess: true,
				},
				{
					pod:             generatePod().WithPhase(apiv1.PodPending).Get(),
					canEvict:        true,
					evictionSuccess: true,
				},
				{
					pod:             generatePod().WithPhase(apiv1.PodPending).Get(),
					canEvict:        true,
					evictionSuccess: true,
				},
			},
		},
		{
			name:              "Cannot evict a single Pod under default settings.",
			replicas:          1,
			evictionTolerance: 0.5,
			vpa:               getBasicVpa(),
			pods: []podWithExpectations{
				{
					pod:             generatePod().Get(),
					canEvict:        false,
					evictionSuccess: false,
				},
			},
		},
		{
			name:              "Can evict even a single Pod using PodUpdatePolicy.MinReplicas.",
			replicas:          1,
			evictionTolerance: 0.5,
			vpa:               vpaSingleReplica,
			pods: []podWithExpectations{
				{
					pod:             generatePod().Get(),
					canEvict:        true,
					evictionSuccess: true,
				},
			},
		},
		{
			name:              "In-place update only first pod (half of 3).",
			replicas:          3,
			evictionTolerance: 0.5,
			vpa:               getIPORVpa(),
			pods: []podWithExpectations{
				{
					pod:                  generatePod().Get(),
					canInPlaceUpdate:     utils.InPlaceApproved,
					inPlaceUpdateSuccess: true,
				},
				{
					pod:                  generatePod().Get(),
					canInPlaceUpdate:     utils.InPlaceApproved,
					inPlaceUpdateSuccess: false,
				},
				{
					pod:                  generatePod().Get(),
					canInPlaceUpdate:     utils.InPlaceApproved,
					inPlaceUpdateSuccess: false,
				},
			},
		},
		{
			name:              "For small eviction tolerance at least one pod is in-place resized.",
			replicas:          3,
			evictionTolerance: 0.1,
			vpa:               getIPORVpa(),
			pods: []podWithExpectations{
				{
					pod:                  generatePod().Get(),
					canInPlaceUpdate:     utils.InPlaceApproved,
					inPlaceUpdateSuccess: true,
				},
				{
					pod:                  generatePod().Get(),
					canInPlaceUpdate:     utils.InPlaceApproved,
					inPlaceUpdateSuccess: false,
				},
				{
					pod:                  generatePod().Get(),
					canInPlaceUpdate:     utils.InPlaceApproved,
					inPlaceUpdateSuccess: false,
				},
			},
		},
		{
			name:              "Ongoing in-placing pods will not get resized again, but may be considered for eviction or deferred.",
			replicas:          3,
			evictionTolerance: 0.1,
			vpa:               getIPORVpa(),
			pods: []podWithExpectations{
				{
					pod: generatePod().WithPodConditions([]apiv1.PodCondition{
						{
							Type:   apiv1.PodResizePending,
							Status: apiv1.ConditionTrue,
							Reason: apiv1.PodReasonInfeasible,
						},
					}).Get(),
					canInPlaceUpdate:     utils.InPlaceEvict,
					inPlaceUpdateSuccess: false,
				},
				{
					pod: generatePod().WithPodConditions([]apiv1.PodCondition{
						{
							Type:   apiv1.PodResizeInProgress,
							Status: apiv1.ConditionTrue,
						},
					}).Get(),
					canInPlaceUpdate:     utils.InPlaceDeferred,
					inPlaceUpdateSuccess: false,
				},
				{
					pod:                  generatePod().Get(),
					canInPlaceUpdate:     utils.InPlaceApproved,
					inPlaceUpdateSuccess: true,
				},
			},
		},
		{
			name:              "Can in-place a single Pod under default settings (belowMinReplicas allows in-place, blocks eviction).",
			replicas:          1,
			evictionTolerance: 0.5,
			vpa:               getIPORVpa(),
			pods: []podWithExpectations{
				{
					pod:                  generatePod().Get(),
					canInPlaceUpdate:     utils.InPlaceApproved,
					inPlaceUpdateSuccess: true,
				},
			},
		},
		{
			name:              "Can in-place even a single Pod using PodUpdatePolicy.MinReplicas.",
			replicas:          1,
			evictionTolerance: 0.5,
			vpa: func() *vpa_types.VerticalPodAutoscaler {
				vpa := getIPORVpa()
				vpa.Spec.UpdatePolicy.MinReplicas = ptr.To(int32(1))
				return vpa
			}(),
			pods: []podWithExpectations{
				{
					pod:                  generatePod().Get(),
					canInPlaceUpdate:     utils.InPlaceApproved,
					inPlaceUpdateSuccess: true,
				},
			},
		},
		{
			name:              "First pod can be evicted without violation of tolerance, even if other evictable pods have ongoing resizes.",
			replicas:          3,
			evictionTolerance: 0.5,
			vpa:               getBasicVpa(),
			pods: []podWithExpectations{
				{
					pod:             generatePod().Get(),
					canEvict:        true,
					evictionSuccess: true,
				},
				{
					pod: generatePod().WithPodConditions([]apiv1.PodCondition{
						{
							Type:   apiv1.PodResizePending,
							Status: apiv1.ConditionTrue,
							Reason: apiv1.PodReasonInfeasible,
						},
					}).Get(),
					canEvict:        true,
					evictionSuccess: false,
				},
				{
					pod: generatePod().WithPodConditions([]apiv1.PodCondition{
						{
							Type:   apiv1.PodResizePending,
							Status: apiv1.ConditionTrue,
							Reason: apiv1.PodReasonInfeasible,
						},
					}).Get(),
					canEvict:        true,
					evictionSuccess: false,
				},
			},
		},
		{
			name:              "No pods are evictable even if some pods are stuck resizing, but some are missing and eviction tolerance is small.",
			replicas:          4,
			evictionTolerance: 0.1,
			vpa:               getBasicVpa(),
			pods: []podWithExpectations{
				{
					pod:             generatePod().Get(),
					canEvict:        false,
					evictionSuccess: false,
				},
				{
					pod: generatePod().WithPodConditions([]apiv1.PodCondition{
						{
							Type:   apiv1.PodResizePending,
							Status: apiv1.ConditionTrue,
							Reason: apiv1.PodReasonInfeasible,
						},
					}).Get(),
					canEvict:        false,
					evictionSuccess: false,
				},
				{
					pod:             generatePod().Get(),
					canEvict:        false,
					evictionSuccess: false,
				},
			},
		},
		{
			name:              "All pods, including resizing pods, are evictable due to large tolerance.",
			replicas:          3,
			evictionTolerance: 1,
			vpa:               getBasicVpa(),
			pods: []podWithExpectations{
				{
					pod:             generatePod().Get(),
					canEvict:        true,
					evictionSuccess: true,
				},
				{
					pod: generatePod().WithPodConditions([]apiv1.PodCondition{
						{
							Type:   apiv1.PodResizePending,
							Status: apiv1.ConditionTrue,
							Reason: apiv1.PodReasonInfeasible,
						},
					}).Get(),
					canEvict:        true,
					evictionSuccess: true,
				},
				{
					pod:             generatePod().Get(),
					canEvict:        true,
					evictionSuccess: true,
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			rc.Spec = apiv1.ReplicationControllerSpec{
				Replicas: &testCase.replicas,
			}
			pods := make([]*apiv1.Pod, 0, len(testCase.pods))
			for _, p := range testCase.pods {
				pods = append(pods, p.pod)
			}
			factory, err := getRestrictionFactory(&rc, nil, nil, nil, 2, testCase.evictionTolerance, baseclocktest.NewFakeClock(time.Time{}), make(map[string]time.Time), GetFakeCalculatorsWithFakeResourceCalc())
			assert.NoError(t, err)
			creatorToSingleGroupStatsMap, podToReplicaCreatorMap, err := factory.GetCreatorMaps(pods, testCase.vpa)
			assert.NoError(t, err)
			eviction := factory.NewPodsEvictionRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)
			inplace := factory.NewPodsInPlaceRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)
			updateMode := vpa_api_util.GetUpdateMode(testCase.vpa)
			for i, p := range testCase.pods {
				if updateMode == vpa_types.UpdateModeInPlaceOrRecreate {
					assert.Equalf(t, p.canInPlaceUpdate, inplace.CanInPlaceUpdate(p.pod), "unexpected CanInPlaceUpdate result for pod-%v %#v", testCase.name, i, p.pod)
				} else {
					assert.Equalf(t, p.canEvict, eviction.CanEvict(p.pod), "unexpected CanEvict result for pod-%v %#v", i, p.pod)
				}
			}
			for i, p := range testCase.pods {
				if updateMode == vpa_types.UpdateModeInPlaceOrRecreate {
					err := inplace.InPlaceUpdate(p.pod, testCase.vpa, test.FakeEventRecorder())
					if p.inPlaceUpdateSuccess {
						assert.NoErrorf(t, err, "unexpected InPlaceUpdate result for pod-%v %#v", i, p.pod)
					} else {
						assert.Errorf(t, err, "unexpected InPlaceUpdate result for pod-%v %#v", i, p.pod)
					}
				} else {
					err := eviction.Evict(p.pod, testCase.vpa, test.FakeEventRecorder())
					if p.evictionSuccess {
						assert.NoErrorf(t, err, "unexpected Evict result for pod-%v %#v", i, p.pod)
					} else {
						assert.Errorf(t, err, "unexpected Evict result for pod-%v %#v", i, p.pod)
					}
				}
			}
		})
	}
}

func TestEvictReplicatedByReplicaSet(t *testing.T) {
	replicas := int32(5)
	livePods := 5

	rs := appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rs",
			Namespace: "default",
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "ReplicaSet",
		},
		Spec: appsv1.ReplicaSetSpec{
			Replicas: &replicas,
		},
	}

	pods := make([]*apiv1.Pod, livePods)
	for i := range pods {
		pods[i] = test.Pod().WithName(getTestPodName(i)).WithCreator(&rs.ObjectMeta, &rs.TypeMeta).Get()
	}

	basicVpa := getBasicVpa()
	factory, err := getRestrictionFactory(nil, &rs, nil, nil, 2, 0.5, nil, nil, nil)
	assert.NoError(t, err)
	creatorToSingleGroupStatsMap, podToReplicaCreatorMap, err := factory.GetCreatorMaps(pods, basicVpa)
	assert.NoError(t, err)
	eviction := factory.NewPodsEvictionRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)

	for _, pod := range pods {
		assert.True(t, eviction.CanEvict(pod))
	}

	for _, pod := range pods[:2] {
		err := eviction.Evict(pod, basicVpa, test.FakeEventRecorder())
		assert.Nil(t, err, "Should evict with no error")
	}
	for _, pod := range pods[2:] {
		err := eviction.Evict(pod, basicVpa, test.FakeEventRecorder())
		assert.Error(t, err, "Error expected")
	}
}

func TestEvictReplicatedByStatefulSet(t *testing.T) {
	replicas := int32(5)
	livePods := 5

	ss := appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ss",
			Namespace: "default",
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "StatefulSet",
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &replicas,
		},
	}

	pods := make([]*apiv1.Pod, livePods)
	for i := range pods {
		pods[i] = test.Pod().WithName(getTestPodName(i)).WithCreator(&ss.ObjectMeta, &ss.TypeMeta).Get()
	}

	basicVpa := getBasicVpa()
	factory, err := getRestrictionFactory(nil, nil, &ss, nil, 2, 0.5, nil, nil, nil)
	assert.NoError(t, err)
	creatorToSingleGroupStatsMap, podToReplicaCreatorMap, err := factory.GetCreatorMaps(pods, basicVpa)
	assert.NoError(t, err)
	eviction := factory.NewPodsEvictionRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)

	for _, pod := range pods {
		assert.True(t, eviction.CanEvict(pod))
	}

	for _, pod := range pods[:2] {
		err := eviction.Evict(pod, basicVpa, test.FakeEventRecorder())
		assert.Nil(t, err, "Should evict with no error")
	}
	for _, pod := range pods[2:] {
		err := eviction.Evict(pod, basicVpa, test.FakeEventRecorder())
		assert.Error(t, err, "Error expected")
	}
}

func TestEvictReplicatedByDaemonSet(t *testing.T) {
	livePods := int32(5)

	ds := appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ds",
			Namespace: "default",
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "DaemonSet",
		},
		Status: appsv1.DaemonSetStatus{
			NumberReady: livePods,
		},
	}

	pods := make([]*apiv1.Pod, livePods)
	for i := range pods {
		pods[i] = test.Pod().WithName(getTestPodName(i)).WithCreator(&ds.ObjectMeta, &ds.TypeMeta).Get()
	}

	basicVpa := getBasicVpa()
	factory, err := getRestrictionFactory(nil, nil, nil, &ds, 2, 0.5, nil, nil, nil)
	assert.NoError(t, err)
	creatorToSingleGroupStatsMap, podToReplicaCreatorMap, err := factory.GetCreatorMaps(pods, basicVpa)
	assert.NoError(t, err)
	eviction := factory.NewPodsEvictionRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)

	for _, pod := range pods {
		assert.True(t, eviction.CanEvict(pod))
	}

	for _, pod := range pods[:2] {
		err := eviction.Evict(pod, basicVpa, test.FakeEventRecorder())
		assert.Nil(t, err, "Should evict with no error")
	}
	for _, pod := range pods[2:] {
		err := eviction.Evict(pod, basicVpa, test.FakeEventRecorder())
		assert.Error(t, err, "Error expected")
	}
}

func TestEvictReplicatedByJob(t *testing.T) {
	job := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "job",
			Namespace: "default",
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "Job",
		},
	}

	livePods := 5

	pods := make([]*apiv1.Pod, livePods)
	for i := range pods {
		pods[i] = test.Pod().WithName(getTestPodName(i)).WithCreator(&job.ObjectMeta, &job.TypeMeta).Get()
	}

	basicVpa := getBasicVpa()
	factory, err := getRestrictionFactory(nil, nil, nil, nil, 2, 0.5, nil, nil, nil)
	assert.NoError(t, err)
	creatorToSingleGroupStatsMap, podToReplicaCreatorMap, err := factory.GetCreatorMaps(pods, basicVpa)
	assert.NoError(t, err)
	eviction := factory.NewPodsEvictionRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)

	for _, pod := range pods {
		assert.True(t, eviction.CanEvict(pod))
	}

	for _, pod := range pods[:2] {
		err := eviction.Evict(pod, basicVpa, test.FakeEventRecorder())
		assert.Nil(t, err, "Should evict with no error")
	}
	for _, pod := range pods[2:] {
		err := eviction.Evict(pod, basicVpa, test.FakeEventRecorder())
		assert.Error(t, err, "Error expected")
	}
}

func getRestrictionFactory(rc *apiv1.ReplicationController, rs *appsv1.ReplicaSet,
	ss *appsv1.StatefulSet, ds *appsv1.DaemonSet, minReplicas int,
	evictionToleranceFraction float64, clock clock.Clock, lipuatm map[string]time.Time, patchCalculators []patch.Calculator) (PodsRestrictionFactory, error) {
	kubeClient := &fake.Clientset{}
	rcInformer := coreinformer.NewReplicationControllerInformer(kubeClient, apiv1.NamespaceAll,
		0*time.Second, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	rsInformer := appsinformer.NewReplicaSetInformer(kubeClient, apiv1.NamespaceAll,
		0*time.Second, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	ssInformer := appsinformer.NewStatefulSetInformer(kubeClient, apiv1.NamespaceAll,
		0*time.Second, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	dsInformer := appsinformer.NewDaemonSetInformer(kubeClient, apiv1.NamespaceAll,
		0*time.Second, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	if rc != nil {
		err := rcInformer.GetIndexer().Add(rc)
		if err != nil {
			return nil, fmt.Errorf("Error adding object to cache: %v", err)
		}
	}
	if rs != nil {
		err := rsInformer.GetIndexer().Add(rs)
		if err != nil {
			return nil, fmt.Errorf("Error adding object to cache: %v", err)
		}
	}
	if ss != nil {
		err := ssInformer.GetIndexer().Add(ss)
		if err != nil {
			return nil, fmt.Errorf("Error adding object to cache: %v", err)
		}
	}
	if ds != nil {
		err := dsInformer.GetIndexer().Add(ds)
		if err != nil {
			return nil, fmt.Errorf("Error adding object to cache: %v", err)
		}
	}

	return &PodsRestrictionFactoryImpl{
		client:                    kubeClient,
		rcInformer:                rcInformer,
		ssInformer:                ssInformer,
		rsInformer:                rsInformer,
		dsInformer:                dsInformer,
		minReplicas:               minReplicas,
		evictionToleranceFraction: evictionToleranceFraction,
		clock:                     clock,
		lastInPlaceAttemptTimeMap: lipuatm,
		patchCalculators:          patchCalculators,
	}, nil
}

func getTestPodName(index int) string {
	return fmt.Sprintf("test-%v", index)
}

type fakeResizePatchCalculator struct {
	patches []resource_admission.PatchRecord
	err     error
}

func (c *fakeResizePatchCalculator) CalculatePatches(_ *apiv1.Pod, _ *vpa_types.VerticalPodAutoscaler) (
	[]resource_admission.PatchRecord, error) {
	return c.patches, c.err
}

func (c *fakeResizePatchCalculator) PatchResourceTarget() patch.PatchResourceTarget {
	return patch.Resize
}

func NewFakeCalculatorWithInPlacePatches() patch.Calculator {
	return &fakeResizePatchCalculator{
		patches: []resource_admission.PatchRecord{
			{
				Op:    "fakeop",
				Path:  "fakepath",
				Value: apiv1.ResourceList{},
			},
		},
	}
}

func GetFakeCalculatorsWithFakeResourceCalc() []patch.Calculator {
	return []patch.Calculator{
		NewFakeCalculatorWithInPlacePatches(),
	}
}
