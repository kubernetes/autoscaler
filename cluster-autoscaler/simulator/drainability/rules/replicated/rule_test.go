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

package replicated

import (
	"fmt"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/drainability"
	"k8s.io/autoscaler/cluster-autoscaler/utils/drain"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	v1appslister "k8s.io/client-go/listers/apps/v1"
	v1lister "k8s.io/client-go/listers/core/v1"

	"github.com/stretchr/testify/assert"
)

func TestDrain(t *testing.T) {
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

		rcPod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "bar",
				Namespace:       "default",
				OwnerReferences: GenerateOwnerReferences(rc.Name, "ReplicationController", "core/v1", ""),
			},
			Spec: apiv1.PodSpec{
				NodeName: "node",
			},
		}

		ds = appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ds",
				Namespace: "default",
				SelfLink:  "/apiv1s/apps/v1/namespaces/default/daemonsets/ds",
			},
		}

		dsPod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "bar",
				Namespace:       "default",
				OwnerReferences: GenerateOwnerReferences(ds.Name, "DaemonSet", "apps/v1", ""),
			},
			Spec: apiv1.PodSpec{
				NodeName: "node",
			},
		}

		cdsPod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "bar",
				Namespace:       "default",
				OwnerReferences: GenerateOwnerReferences(ds.Name, "CustomDaemonSet", "crd/v1", ""),
				Annotations: map[string]string{
					"cluster-autoscaler.kubernetes.io/daemonset-pod": "true",
				},
			},
			Spec: apiv1.PodSpec{
				NodeName: "node",
			},
		}

		job = batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "job",
				Namespace: "default",
				SelfLink:  "/apiv1s/batch/v1/namespaces/default/jobs/job",
			},
		}

		jobPod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "bar",
				Namespace:       "default",
				OwnerReferences: GenerateOwnerReferences(job.Name, "Job", "batch/v1", ""),
			},
		}

		statefulset = appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ss",
				Namespace: "default",
				SelfLink:  "/apiv1s/apps/v1/namespaces/default/statefulsets/ss",
			},
		}

		ssPod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "bar",
				Namespace:       "default",
				OwnerReferences: GenerateOwnerReferences(statefulset.Name, "StatefulSet", "apps/v1", ""),
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

		rsPod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "bar",
				Namespace:       "default",
				OwnerReferences: GenerateOwnerReferences(rs.Name, "ReplicaSet", "apps/v1", ""),
			},
			Spec: apiv1.PodSpec{
				NodeName: "node",
			},
		}

		rsPodDeleted = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "bar",
				Namespace:         "default",
				OwnerReferences:   GenerateOwnerReferences(rs.Name, "ReplicaSet", "apps/v1", ""),
				DeletionTimestamp: &metav1.Time{Time: testTime.Add(-time.Hour)},
			},
			Spec: apiv1.PodSpec{
				NodeName: "node",
			},
		}

		customControllerPod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "bar",
				Namespace: "default",
				// Using names like FooController is discouraged
				// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#naming-conventions
				// vadasambar: I am using it here just because `FooController``
				// is easier to understand than say `FooSet`
				OwnerReferences: GenerateOwnerReferences("Foo", "FooController", "apps/v1", ""),
			},
			Spec: apiv1.PodSpec{
				NodeName: "node",
			},
		}

		nakedPod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "bar",
				Namespace: "default",
			},
			Spec: apiv1.PodSpec{
				NodeName: "node",
			},
		}

		nakedFailedPod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "bar",
				Namespace: "default",
			},
			Spec: apiv1.PodSpec{
				NodeName:      "node",
				RestartPolicy: apiv1.RestartPolicyNever,
			},
			Status: apiv1.PodStatus{
				Phase: apiv1.PodFailed,
			},
		}

		nakedTerminalPod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "bar",
				Namespace: "default",
			},
			Spec: apiv1.PodSpec{
				NodeName:      "node",
				RestartPolicy: apiv1.RestartPolicyOnFailure,
			},
			Status: apiv1.PodStatus{
				Phase: apiv1.PodSucceeded,
			},
		}

		nakedEvictedPod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "bar",
				Namespace: "default",
			},
			Spec: apiv1.PodSpec{
				NodeName:      "node",
				RestartPolicy: apiv1.RestartPolicyAlways,
			},
			Status: apiv1.PodStatus{
				Phase: apiv1.PodFailed,
			},
		}

		nakedSafePod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "bar",
				Namespace: "default",
				Annotations: map[string]string{
					drain.PodSafeToEvictKey: "true",
				},
			},
			Spec: apiv1.PodSpec{
				NodeName: "node",
			},
		}

		zeroGracePeriod         = int64(0)
		nakedLongTerminatingPod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "bar",
				Namespace:         "default",
				DeletionTimestamp: &metav1.Time{Time: testTime.Add(-2 * drain.PodLongTerminatingExtraThreshold)},
			},
			Spec: apiv1.PodSpec{
				NodeName:                      "node",
				RestartPolicy:                 apiv1.RestartPolicyOnFailure,
				TerminationGracePeriodSeconds: &zeroGracePeriod,
			},
			Status: apiv1.PodStatus{
				Phase: apiv1.PodUnknown,
			},
		}

		extendedGracePeriod                            = int64(6 * 60) // 6 minutes
		nakedLongTerminatingPodWithExtendedGracePeriod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "bar",
				Namespace:         "default",
				DeletionTimestamp: &metav1.Time{Time: testTime.Add(-2 * time.Duration(extendedGracePeriod) * time.Second)},
			},
			Spec: apiv1.PodSpec{
				NodeName:                      "node",
				RestartPolicy:                 apiv1.RestartPolicyOnFailure,
				TerminationGracePeriodSeconds: &extendedGracePeriod,
			},
			Status: apiv1.PodStatus{
				Phase: apiv1.PodUnknown,
			},
		}
	)

	type testCase struct {
		desc string
		pod  *apiv1.Pod
		rcs  []*apiv1.ReplicationController
		rss  []*appsv1.ReplicaSet

		// TODO(vadasambar): remove this when we get rid of scaleDownNodesWithCustomControllerPods
		skipNodesWithCustomControllerPods bool

		wantReason drain.BlockingPodReason
		wantError  bool
	}

	sharedTests := []testCase{
		{
			desc: "RC-managed pod",
			pod:  rcPod,
			rcs:  []*apiv1.ReplicationController{&rc},
		},
		{
			desc: "DS-managed pod",
			pod:  dsPod,
		},
		{
			desc: "DS-managed pod by a custom Daemonset",
			pod:  cdsPod,
		},
		{
			desc: "Job-managed pod",
			pod:  jobPod,
			rcs:  []*apiv1.ReplicationController{&rc},
		},
		{
			desc: "SS-managed pod",
			pod:  ssPod,
			rcs:  []*apiv1.ReplicationController{&rc},
		},
		{
			desc: "RS-managed pod",
			pod:  rsPod,
			rss:  []*appsv1.ReplicaSet{&rs},
		},
		{
			desc: "RS-managed pod that is being deleted",
			pod:  rsPodDeleted,
			rss:  []*appsv1.ReplicaSet{&rs},
		},
		{
			desc:       "naked pod",
			pod:        nakedPod,
			wantReason: drain.NotReplicated,
			wantError:  true,
		},
		{
			desc: "naked failed pod",
			pod:  nakedFailedPod,
		},
		{
			desc: "naked terminal pod",
			pod:  nakedTerminalPod,
		},
		{
			desc: "naked evicted pod",
			pod:  nakedEvictedPod,
		},
		{
			desc: "naked pod with PodSafeToEvict annotation",
			pod:  nakedSafePod,
		},
		{
			desc: "naked long terminating pod with 0 grace period",
			pod:  nakedLongTerminatingPod,
		},
		{
			desc: "naked long terminating pod with extended grace period",
			pod:  nakedLongTerminatingPodWithExtendedGracePeriod,
		},
	}

	var tests []testCase

	// Note: do not modify the underlying reference values for sharedTests.
	for _, test := range sharedTests {
		for _, skipNodesWithCustomControllerPods := range []bool{true, false} {
			// Copy test to prevent side effects.
			test := test
			test.skipNodesWithCustomControllerPods = skipNodesWithCustomControllerPods
			test.desc = fmt.Sprintf("%s with skipNodesWithCustomControllerPods:%t", test.desc, skipNodesWithCustomControllerPods)
			tests = append(tests, test)
		}
	}

	customControllerTests := []testCase{
		{
			desc:                              "Custom-controller-managed blocking pod",
			pod:                               customControllerPod,
			skipNodesWithCustomControllerPods: true,
			wantReason:                        drain.NotReplicated,
			wantError:                         true,
		},
		{
			desc: "Custom-controller-managed non-blocking pod",
			pod:  customControllerPod,
		},
	}
	tests = append(tests, customControllerTests...)

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
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
			status := New(test.skipNodesWithCustomControllerPods, 0).Drainable(drainCtx, test.pod)
			assert.Equal(t, test.wantReason, status.BlockingReason)
			assert.Equal(t, test.wantError, status.Error != nil)
		})
	}
}
