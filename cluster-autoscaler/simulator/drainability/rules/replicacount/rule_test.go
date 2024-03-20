/*
Copyright 2023 The Kubernetes Authors.

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

package replicacount

import (
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/drainability"
	"k8s.io/autoscaler/cluster-autoscaler/utils/drain"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/autoscaler/cluster-autoscaler/utils/test"
	v1appslister "k8s.io/client-go/listers/apps/v1"
	v1lister "k8s.io/client-go/listers/core/v1"

	"github.com/stretchr/testify/assert"
)

func TestDrainable(t *testing.T) {
	var (
		testTime = time.Date(2020, time.December, 18, 17, 0, 0, 0, time.UTC)
		replicas = int32(5)

		rc = apiv1.ReplicationController{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "rc",
				Namespace: "default",
				SelfLink:  "api/v1/namespaces/default/replicationcontrollers/rc",
			},
			Spec: apiv1.ReplicationControllerSpec{
				Replicas: &replicas,
			},
		}
		ds = appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ds",
				Namespace: "default",
				SelfLink:  "/apiv1s/apps/v1/namespaces/default/daemonsets/ds",
			},
		}
		job = batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "job",
				Namespace: "default",
				SelfLink:  "/apiv1s/batch/v1/namespaces/default/jobs/job",
			},
		}
		statefulset = appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ss",
				Namespace: "default",
				SelfLink:  "/apiv1s/apps/v1/namespaces/default/statefulsets/ss",
			},
		}
		rs = appsv1.ReplicaSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "rs",
				Namespace: "default",
				SelfLink:  "api/v1/namespaces/default/replicasets/rs",
			},
			Spec: appsv1.ReplicaSetSpec{
				Replicas: &replicas,
			},
		}
	)

	for desc, test := range map[string]struct {
		desc string
		pod  *apiv1.Pod
		rcs  []*apiv1.ReplicationController
		rss  []*appsv1.ReplicaSet

		wantReason drain.BlockingPodReason
		wantError  bool
	}{
		"RC-managed pod": {
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "bar",
					Namespace:       "default",
					OwnerReferences: test.GenerateOwnerReferences(rc.Name, "ReplicationController", "core/v1", ""),
				},
				Spec: apiv1.PodSpec{
					NodeName: "node",
				},
			},
			rcs: []*apiv1.ReplicationController{&rc},
		},
		"RC-managed pod with missing reference": {
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "bar",
					Namespace:       "default",
					OwnerReferences: test.GenerateOwnerReferences("missing", "ReplicationController", "core/v1", ""),
				},
				Spec: apiv1.PodSpec{
					NodeName: "node",
				},
			},
			rcs:        []*apiv1.ReplicationController{&rc},
			wantReason: drain.ControllerNotFound,
			wantError:  true,
		},
		"DS-managed pod": {
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "bar",
					Namespace:       "default",
					OwnerReferences: test.GenerateOwnerReferences(ds.Name, "DaemonSet", "apps/v1", ""),
				},
				Spec: apiv1.PodSpec{
					NodeName: "node",
				},
			},
		},
		"DS-managed pod with missing reference": {
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "bar",
					Namespace:       "default",
					OwnerReferences: test.GenerateOwnerReferences("missing", "DaemonSet", "apps/v1", ""),
				},
				Spec: apiv1.PodSpec{
					NodeName: "node",
				},
			},
			wantReason: drain.ControllerNotFound,
			wantError:  true,
		},
		"DS-managed pod by a custom API Group": {
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "bar",
					Namespace:       "default",
					OwnerReferences: test.GenerateOwnerReferences(ds.Name, "DaemonSet", "crd/v1", ""),
				},
				Spec: apiv1.PodSpec{
					NodeName: "node",
				},
			},
		},
		"DS-managed pod by a custom API Group with missing reference": {
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "bar",
					Namespace:       "default",
					OwnerReferences: test.GenerateOwnerReferences("missing", "DaemonSet", "crd/v1", ""),
				},
				Spec: apiv1.PodSpec{
					NodeName: "node",
				},
			},
		},
		"DS-managed pod by a custom Daemonset": {
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "bar",
					Namespace:       "default",
					OwnerReferences: test.GenerateOwnerReferences(ds.Name, "CustomDaemonSet", "crd/v1", ""),
					Annotations: map[string]string{
						"cluster-autoscaler.kubernetes.io/daemonset-pod": "true",
					},
				},
				Spec: apiv1.PodSpec{
					NodeName: "node",
				},
			},
		},
		"DS-managed pod by a custom Daemonset with missing reference": {
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "bar",
					Namespace:       "default",
					OwnerReferences: test.GenerateOwnerReferences("missing", "CustomDaemonSet", "crd/v1", ""),
					Annotations: map[string]string{
						"cluster-autoscaler.kubernetes.io/daemonset-pod": "true",
					},
				},
				Spec: apiv1.PodSpec{
					NodeName: "node",
				},
			},
		},
		"Job-managed pod": {
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "bar",
					Namespace:       "default",
					OwnerReferences: test.GenerateOwnerReferences(job.Name, "Job", "batch/v1", ""),
				},
			},
			rcs: []*apiv1.ReplicationController{&rc},
		},
		"Job-managed pod with missing reference": {
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "bar",
					Namespace:       "default",
					OwnerReferences: test.GenerateOwnerReferences("missing", "Job", "batch/v1", ""),
				},
			},
			rcs:        []*apiv1.ReplicationController{&rc},
			wantReason: drain.ControllerNotFound,
			wantError:  true,
		},
		"SS-managed pod": {
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "bar",
					Namespace:       "default",
					OwnerReferences: test.GenerateOwnerReferences(statefulset.Name, "StatefulSet", "apps/v1", ""),
				},
			},
			rcs: []*apiv1.ReplicationController{&rc},
		},
		"SS-managed pod with missing reference": {
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "bar",
					Namespace:       "default",
					OwnerReferences: test.GenerateOwnerReferences("missing", "StatefulSet", "apps/v1", ""),
				},
			},
			rcs:        []*apiv1.ReplicationController{&rc},
			wantReason: drain.ControllerNotFound,
			wantError:  true,
		},
		"RS-managed pod": {
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "bar",
					Namespace:       "default",
					OwnerReferences: test.GenerateOwnerReferences(rs.Name, "ReplicaSet", "apps/v1", ""),
				},
				Spec: apiv1.PodSpec{
					NodeName: "node",
				},
			},
			rss: []*appsv1.ReplicaSet{&rs},
		},
		"RS-managed pod that is being deleted": {
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "bar",
					Namespace:         "default",
					OwnerReferences:   test.GenerateOwnerReferences(rs.Name, "ReplicaSet", "apps/v1", ""),
					DeletionTimestamp: &metav1.Time{Time: testTime.Add(-time.Hour)},
				},
				Spec: apiv1.PodSpec{
					NodeName: "node",
				},
			},
			rss: []*appsv1.ReplicaSet{&rs},
		},
		"RS-managed pod with missing reference": {
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "bar",
					Namespace:       "default",
					OwnerReferences: test.GenerateOwnerReferences("missing", "ReplicaSet", "apps/v1", ""),
				},
				Spec: apiv1.PodSpec{
					NodeName: "node",
				},
			},
			rss:        []*appsv1.ReplicaSet{&rs},
			wantReason: drain.ControllerNotFound,
			wantError:  true,
		},
	} {
		t.Run(desc, func(t *testing.T) {
			var err error
			var rcLister v1lister.ReplicationControllerLister
			if len(test.rcs) > 0 {
				rcLister, err = kube_util.NewTestReplicationControllerLister(test.rcs)
				assert.NoError(t, err)
			}
			var rsLister v1appslister.ReplicaSetLister
			if len(test.rss) > 0 {
				rsLister, err = kube_util.NewTestReplicaSetLister(test.rss)
				assert.NoError(t, err)
			}
			dsLister, err := kube_util.NewTestDaemonSetLister([]*appsv1.DaemonSet{&ds})
			assert.NoError(t, err)
			jobLister, err := kube_util.NewTestJobLister([]*batchv1.Job{&job})
			assert.NoError(t, err)
			ssLister, err := kube_util.NewTestStatefulSetLister([]*appsv1.StatefulSet{&statefulset})
			assert.NoError(t, err)

			registry := kube_util.NewListerRegistry(nil, nil, nil, nil, dsLister, rcLister, jobLister, rsLister, ssLister)

			drainCtx := &drainability.DrainContext{
				Listers:   registry,
				Timestamp: testTime,
			}
			status := New(0).Drainable(drainCtx, test.pod)
			assert.Equal(t, test.wantReason, status.BlockingReason)
			assert.Equal(t, test.wantError, status.Error != nil)
		})
	}
}
