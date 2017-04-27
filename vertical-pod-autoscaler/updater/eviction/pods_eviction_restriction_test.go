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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/autoscaler/vertical-pod-autoscaler/updater/test"
	core "k8s.io/client-go/testing"
	"k8s.io/kubernetes/pkg/api/testapi"
	apiv1 "k8s.io/kubernetes/pkg/api/v1"
	appsv1beta1 "k8s.io/kubernetes/pkg/apis/apps/v1beta1"
	batchv1 "k8s.io/kubernetes/pkg/apis/batch/v1"
	extensions "k8s.io/kubernetes/pkg/apis/extensions/v1beta1"
	kube_client "k8s.io/kubernetes/pkg/client/clientset_generated/clientset"
	"k8s.io/kubernetes/pkg/client/clientset_generated/clientset/fake"
)

func TestEvictReplicatedByController(t *testing.T) {
	replicas := int32(5)
	livePods := 5

	rc := apiv1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rc",
			Namespace: "default",
			SelfLink:  testapi.Default.SelfLink("replicationcontrollers", "rc"),
		},
		Spec: apiv1.ReplicationControllerSpec{
			Replicas: &replicas,
		},
	}

	pods := make([]*apiv1.Pod, livePods)
	for i := range pods {
		pods[i] = test.BuildTestPod("test"+string(i), "", "", "", &rc)
	}

	eviction := NewPodsEvictionRestrictionFactory(fakeClient(&rc, nil, nil, nil, pods), 2, 0.5).NewPodsEvictionRestriction(pods)

	for _, pod := range pods {
		assert.True(t, eviction.CanEvict(pod))
	}

	for _, pod := range pods[:2] {
		err := eviction.Evict(pod)
		assert.Nil(t, err, "Should evict with no error")
	}
	for _, pod := range pods[2:] {
		err := eviction.Evict(pod)
		assert.Error(t, err, "Error expected")
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
		Spec: extensions.ReplicaSetSpec{
			Replicas: &replicas,
		},
	}

	pods := make([]*apiv1.Pod, livePods)
	for i := range pods {
		pods[i] = test.BuildTestPod("test"+string(i), "", "", "", &rs)
	}

	eviction := NewPodsEvictionRestrictionFactory(fakeClient(nil, &rs, nil, nil, pods), 2, 0.5).NewPodsEvictionRestriction(pods)

	for _, pod := range pods {
		assert.True(t, eviction.CanEvict(pod))
	}

	for _, pod := range pods[:2] {
		err := eviction.Evict(pod)
		assert.Nil(t, err, "Should evict with no error")
	}
	for _, pod := range pods[2:] {
		err := eviction.Evict(pod)
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
		Spec: appsv1beta1.StatefulSetSpec{
			Replicas: &replicas,
		},
	}

	pods := make([]*apiv1.Pod, livePods)
	for i := range pods {
		pods[i] = test.BuildTestPod("test"+string(i), "", "", "", &ss)
	}

	eviction := NewPodsEvictionRestrictionFactory(fakeClient(nil, nil, &ss, nil, pods), 2, 0.5).NewPodsEvictionRestriction(pods)

	for _, pod := range pods {
		assert.True(t, eviction.CanEvict(pod))
	}

	for _, pod := range pods[:2] {
		err := eviction.Evict(pod)
		assert.Nil(t, err, "Should evict with no error")
	}
	for _, pod := range pods[2:] {
		err := eviction.Evict(pod)
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
	}

	livePods := 5

	pods := make([]*apiv1.Pod, livePods)
	for i := range pods {
		pods[i] = test.BuildTestPod("test"+string(i), "", "", "", &job)
	}

	eviction := NewPodsEvictionRestrictionFactory(fakeClient(nil, nil, nil, &job, pods), 2, 0.5).NewPodsEvictionRestriction(pods)

	for _, pod := range pods {
		assert.True(t, eviction.CanEvict(pod))
	}

	for _, pod := range pods[:2] {
		err := eviction.Evict(pod)
		assert.Nil(t, err, "Should evict with no error")
	}
	for _, pod := range pods[2:] {
		err := eviction.Evict(pod)
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
		Spec: apiv1.ReplicationControllerSpec{
			Replicas: &replicas,
		},
	}

	pods := make([]*apiv1.Pod, livePods)
	for i := range pods {
		pods[i] = test.BuildTestPod("test"+string(i), "", "", "", &rc)
	}

	eviction := NewPodsEvictionRestrictionFactory(fakeClient(&rc, nil, nil, nil, pods), 10, 0.5).NewPodsEvictionRestriction(pods)

	for _, pod := range pods {
		assert.False(t, eviction.CanEvict(pod))
	}

	for _, pod := range pods {
		err := eviction.Evict(pod)
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
		Spec: apiv1.ReplicationControllerSpec{
			Replicas: &replicas,
		},
	}

	pods := make([]*apiv1.Pod, livePods)
	for i := range pods {
		pods[i] = test.BuildTestPod("test"+string(i), "", "", "", &rc)
	}

	eviction := NewPodsEvictionRestrictionFactory(fakeClient(&rc, nil, nil, nil, pods), 2, tolerance).NewPodsEvictionRestriction(pods)

	for _, pod := range pods {
		assert.True(t, eviction.CanEvict(pod))
	}

	for _, pod := range pods[:4] {
		err := eviction.Evict(pod)
		assert.Nil(t, err, "Should evict with no error")
	}
	for _, pod := range pods[4:] {
		err := eviction.Evict(pod)
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
		Spec: apiv1.ReplicationControllerSpec{
			Replicas: &replicas,
		},
	}

	pods := make([]*apiv1.Pod, livePods)
	for i := range pods {
		pods[i] = test.BuildTestPod("test"+string(i), "", "", "", &rc)
	}

	eviction := NewPodsEvictionRestrictionFactory(fakeClient(&rc, nil, nil, nil, pods), 2, tolerance).NewPodsEvictionRestriction(pods)

	for _, pod := range pods {
		assert.True(t, eviction.CanEvict(pod))
	}

	for _, pod := range pods[:1] {
		err := eviction.Evict(pod)
		assert.Nil(t, err, "Should evict with no error")
	}
	for _, pod := range pods[1:] {
		err := eviction.Evict(pod)
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
