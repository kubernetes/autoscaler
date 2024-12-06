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

package eviction

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
	appsinformer "k8s.io/client-go/informers/apps/v1"
	coreinformer "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
)

type podWithExpectations struct {
	pod             *apiv1.Pod
	canEvict        bool
	evictionSuccess bool
}

func getBasicVpa() *vpa_types.VerticalPodAutoscaler {
	return test.VerticalPodAutoscaler().WithContainer("any").Get()
}

func TestEvictReplicatedByController(t *testing.T) {
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
	}

	for _, testCase := range testCases {
		rc.Spec = apiv1.ReplicationControllerSpec{
			Replicas: &testCase.replicas,
		}
		pods := make([]*apiv1.Pod, 0, len(testCase.pods))
		for _, p := range testCase.pods {
			pods = append(pods, p.pod)
		}
		factory, _ := getEvictionRestrictionFactory(&rc, nil, nil, nil, 2, testCase.evictionTolerance)
		eviction := factory.NewPodsEvictionRestriction(pods, testCase.vpa)
		for i, p := range testCase.pods {
			assert.Equalf(t, p.canEvict, eviction.CanEvict(p.pod), "TC %v - unexpected CanEvict result for pod-%v %#v", testCase.name, i, p.pod)
		}
		for i, p := range testCase.pods {
			err := eviction.Evict(p.pod, testCase.vpa, test.FakeEventRecorder())
			if p.evictionSuccess {
				assert.NoErrorf(t, err, "TC %v - unexpected Evict result for pod-%v %#v", testCase.name, i, p.pod)
			} else {
				assert.Errorf(t, err, "TC %v - unexpected Evict result for pod-%v %#v", testCase.name, i, p.pod)
			}
		}
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
	factory, _ := getEvictionRestrictionFactory(nil, &rs, nil, nil, 2, 0.5)
	eviction := factory.NewPodsEvictionRestriction(pods, basicVpa)

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
	factory, _ := getEvictionRestrictionFactory(nil, nil, &ss, nil, 2, 0.5)
	eviction := factory.NewPodsEvictionRestriction(pods, basicVpa)

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
	factory, _ := getEvictionRestrictionFactory(nil, nil, nil, &ds, 2, 0.5)
	eviction := factory.NewPodsEvictionRestriction(pods, basicVpa)

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
	factory, _ := getEvictionRestrictionFactory(nil, nil, nil, nil, 2, 0.5)
	eviction := factory.NewPodsEvictionRestriction(pods, basicVpa)

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

func TestEvictTooFewReplicas(t *testing.T) {
	replicas := int32(5)
	livePods := 5

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
	factory, _ := getEvictionRestrictionFactory(&rc, nil, nil, nil, 10, 0.5)
	eviction := factory.NewPodsEvictionRestriction(pods, basicVpa)

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
	factory, _ := getEvictionRestrictionFactory(&rc, nil, nil, nil, 2 /*minReplicas*/, tolerance)
	eviction := factory.NewPodsEvictionRestriction(pods, basicVpa)

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
	factory, _ := getEvictionRestrictionFactory(&rc, nil, nil, nil, 2, tolerance)
	eviction := factory.NewPodsEvictionRestriction(pods, basicVpa)

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
		rc.Spec = apiv1.ReplicationControllerSpec{
			Replicas: &testCase.replicas,
		}
		pods := make([]*apiv1.Pod, 0, len(testCase.pods))
		for _, p := range testCase.pods {
			pods = append(pods, p.pod)
		}
		factory, _ := getEvictionRestrictionFactory(&rc, nil, nil, nil, 2, testCase.evictionTolerance)
		eviction := factory.NewPodsEvictionRestriction(pods, testCase.vpa)

		for _, p := range testCase.pods {
			mockRecorder := test.MockEventRecorder()
			mockRecorder.On("Event", mock.Anything, apiv1.EventTypeNormal, "EvictedByVPA", mock.Anything).Return()
			mockRecorder.On("Event", mock.Anything, apiv1.EventTypeNormal, "EvictedPod", mock.Anything).Return()

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

func getEvictionRestrictionFactory(rc *apiv1.ReplicationController, rs *appsv1.ReplicaSet,
	ss *appsv1.StatefulSet, ds *appsv1.DaemonSet, minReplicas int,
	evictionToleranceFraction float64) (PodsEvictionRestrictionFactory, error) {
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

	return &podsEvictionRestrictionFactoryImpl{
		client:                    kubeClient,
		rsInformer:                rsInformer,
		rcInformer:                rcInformer,
		ssInformer:                ssInformer,
		dsInformer:                dsInformer,
		minReplicas:               minReplicas,
		evictionToleranceFraction: evictionToleranceFraction,
	}, nil
}

func getTestPodName(index int) string {
	return fmt.Sprintf("test-%v", index)
}
