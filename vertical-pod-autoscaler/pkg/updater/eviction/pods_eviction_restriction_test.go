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

	"github.com/stretchr/testify/assert"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
	kube_client "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"
	"k8s.io/kubernetes/pkg/api/testapi"
)

type podWithExpectations struct {
	pod             *apiv1.Pod
	canEvict        bool
	evictionSuccess bool
}

func TestEvictReplicatedByController(t *testing.T) {

	rc := apiv1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rc",
			Namespace: "default",
			SelfLink:  testapi.Default.SelfLink("replicationcontrollers", "rc"),
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

	testCases := []struct {
		replicas           int32
		evictionTollerance float64
		pods               []podWithExpectations
	}{
		{
			replicas:           3,
			evictionTollerance: 0.5,
			pods: []podWithExpectations{
				{
					// Only first pod will be evicted.
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
			// Half of the population can be evicted.
			replicas:           4,
			evictionTollerance: 0.5,
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
			// Half of the population can be evicted. One pod is missing already.
			replicas:           4,
			evictionTollerance: 0.5,
			pods: []podWithExpectations{
				{
					// Half of the population can be evicted.
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
			replicas:           3,
			evictionTollerance: 0.1,
			pods: []podWithExpectations{
				{
					// For small eviction tollerance at least one pod is evicted.
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
			replicas:           3,
			evictionTollerance: 0.1,
			pods: []podWithExpectations{
				{
					// only 2 pods in replica of 3 and tollerance is 0. None of pods can be evicted
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
			replicas:           3,
			evictionTollerance: 0.5,
			pods: []podWithExpectations{
				{
					pod:             generatePod().Get(),
					canEvict:        false,
					evictionSuccess: false,
				},
				{
					// Only pending pod can be evicted without violation of tollerance
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
			// Pending pods are always evictable
			replicas:           4,
			evictionTollerance: 0.5,
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
	}

	for tcIndex, testCase := range testCases {
		rc.Spec = apiv1.ReplicationControllerSpec{
			Replicas: &testCase.replicas,
		}
		pods := make([]*apiv1.Pod, 0, len(testCase.pods))
		for _, p := range testCase.pods {
			pods = append(pods, p.pod)
		}
		eviction := NewPodsEvictionRestrictionFactory(fakeClient(&rc, nil, nil, nil, pods), 2, testCase.evictionTollerance).NewPodsEvictionRestriction(pods)
		for i, p := range testCase.pods {
			assert.Equalf(t, p.canEvict, eviction.CanEvict(p.pod), "TC %v - unexpected CanEvict result for pod-%v %#v", tcIndex, i, p.pod)
		}
		for i, p := range testCase.pods {
			err := eviction.Evict(p.pod, test.FakeEventRecorder())
			if p.evictionSuccess {
				assert.NoErrorf(t, err, "TC %v - unexpected Evict result for pod-%v %#v", tcIndex, i, p.pod)
			} else {
				assert.Errorf(t, err, "TC %v - unexpected Evict result for pod-%v %#v", tcIndex, i, p.pod)
			}
		}
	}

}

func TestEvictReplicatedByReplicaSet(t *testing.T) {
	replicas := int32(5)
	livePods := 5

	rs := extensions.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rs",
			Namespace: "default",
			SelfLink:  testapi.Default.SelfLink("replicasets", "rs"),
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "ReplicaSet",
		},
		Spec: extensions.ReplicaSetSpec{
			Replicas: &replicas,
		},
	}

	pods := make([]*apiv1.Pod, livePods)
	for i := range pods {
		pods[i] = test.Pod().WithName("test"+string(i)).WithCreator(&rs.ObjectMeta, &rs.TypeMeta).Get()
	}

	eviction := NewPodsEvictionRestrictionFactory(fakeClient(nil, &rs, nil, nil, pods), 2, 0.5).NewPodsEvictionRestriction(pods)

	for _, pod := range pods {
		assert.True(t, eviction.CanEvict(pod))
	}

	for _, pod := range pods[:2] {
		err := eviction.Evict(pod, test.FakeEventRecorder())
		assert.Nil(t, err, "Should evict with no error")
	}
	for _, pod := range pods[2:] {
		err := eviction.Evict(pod, test.FakeEventRecorder())
		assert.Error(t, err, "Error expected")
	}
}

func TestEvictReplicatedByStatefulSet(t *testing.T) {
	replicas := int32(5)
	livePods := 5

	ss := appsv1beta1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ss",
			Namespace: "default",
			SelfLink:  "/apiv1s/extensions/v1beta1/namespaces/default/statefulsets/ss",
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "StatefulSet",
		},
		Spec: appsv1beta1.StatefulSetSpec{
			Replicas: &replicas,
		},
	}

	pods := make([]*apiv1.Pod, livePods)
	for i := range pods {
		pods[i] = test.Pod().WithName("test"+string(i)).WithCreator(&ss.ObjectMeta, &ss.TypeMeta).Get()
	}

	eviction := NewPodsEvictionRestrictionFactory(fakeClient(nil, nil, &ss, nil, pods), 2, 0.5).NewPodsEvictionRestriction(pods)

	for _, pod := range pods {
		assert.True(t, eviction.CanEvict(pod))
	}

	for _, pod := range pods[:2] {
		err := eviction.Evict(pod, test.FakeEventRecorder())
		assert.Nil(t, err, "Should evict with no error")
	}
	for _, pod := range pods[2:] {
		err := eviction.Evict(pod, test.FakeEventRecorder())
		assert.Error(t, err, "Error expected")
	}
}

func TestEvictReplicatedByJob(t *testing.T) {
	job := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "job",
			Namespace: "default",
			SelfLink:  "/apiv1s/extensions/v1beta1/namespaces/default/jobs/job",
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "Job",
		},
	}

	livePods := 5

	pods := make([]*apiv1.Pod, livePods)
	for i := range pods {
		pods[i] = test.Pod().WithName("test"+string(i)).WithCreator(&job.ObjectMeta, &job.TypeMeta).Get()
	}

	eviction := NewPodsEvictionRestrictionFactory(fakeClient(nil, nil, nil, &job, pods), 2, 0.5).NewPodsEvictionRestriction(pods)

	for _, pod := range pods {
		assert.True(t, eviction.CanEvict(pod))
	}

	for _, pod := range pods[:2] {
		err := eviction.Evict(pod, test.FakeEventRecorder())
		assert.Nil(t, err, "Should evict with no error")
	}
	for _, pod := range pods[2:] {
		err := eviction.Evict(pod, test.FakeEventRecorder())
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
			SelfLink:  testapi.Default.SelfLink("replicationcontrollers", "rc"),
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
		pods[i] = test.Pod().WithName("test"+string(i)).WithCreator(&rc.ObjectMeta, &rc.TypeMeta).Get()
	}

	eviction := NewPodsEvictionRestrictionFactory(fakeClient(&rc, nil, nil, nil, pods), 10, 0.5).NewPodsEvictionRestriction(pods)

	for _, pod := range pods {
		assert.False(t, eviction.CanEvict(pod))
	}

	for _, pod := range pods {
		err := eviction.Evict(pod, test.FakeEventRecorder())
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
			SelfLink:  testapi.Default.SelfLink("replicationcontrollers", "rc"),
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
		pods[i] = test.Pod().WithName("test"+string(i)).WithCreator(&rc.ObjectMeta, &rc.TypeMeta).Get()
	}

	eviction := NewPodsEvictionRestrictionFactory(fakeClient(&rc, nil, nil, nil, pods), 2 /*minReplicas*/, tolerance).NewPodsEvictionRestriction(pods)

	for _, pod := range pods {
		assert.True(t, eviction.CanEvict(pod))
	}

	for _, pod := range pods[:4] {
		err := eviction.Evict(pod, test.FakeEventRecorder())
		assert.Nil(t, err, "Should evict with no error")
	}
	for _, pod := range pods[4:] {
		err := eviction.Evict(pod, test.FakeEventRecorder())
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
			SelfLink:  testapi.Default.SelfLink("replicationcontrollers", "rc"),
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
		pods[i] = test.Pod().WithName("test"+string(i)).WithCreator(&rc.ObjectMeta, &rc.TypeMeta).Get()
	}

	eviction := NewPodsEvictionRestrictionFactory(fakeClient(&rc, nil, nil, nil, pods), 2, tolerance).NewPodsEvictionRestriction(pods)

	for _, pod := range pods {
		assert.True(t, eviction.CanEvict(pod))
	}

	for _, pod := range pods[:1] {
		err := eviction.Evict(pod, test.FakeEventRecorder())
		assert.Nil(t, err, "Should evict with no error")
	}
	for _, pod := range pods[1:] {
		err := eviction.Evict(pod, test.FakeEventRecorder())
		assert.Error(t, err, "Error expected")
	}
}

func fakeClient(rc *apiv1.ReplicationController, rs *extensions.ReplicaSet, ss *appsv1beta1.StatefulSet, job *batchv1.Job, pods []*apiv1.Pod) kube_client.Interface {
	fakeClient := &fake.Clientset{}
	register := func(resource string, obj runtime.Object, meta metav1.ObjectMeta) {
		fakeClient.Fake.AddReactor("get", resource, func(action core.Action) (bool, runtime.Object, error) {
			getAction := action.(core.GetAction)
			if getAction.GetName() == meta.GetName() && getAction.GetNamespace() == meta.GetNamespace() {
				return true, obj, nil
			}
			return false, nil, fmt.Errorf("Not found")
		})
	}
	if rc != nil {
		register("replicationcontrollers", rc, rc.ObjectMeta)
	}
	if rs != nil {
		register("replicasets", rs, rs.ObjectMeta)
	}
	if ss != nil {
		register("statefulsets", ss, ss.ObjectMeta)
	}
	if job != nil {
		register("jobs", job, job.ObjectMeta)
	}
	return fakeClient
}
