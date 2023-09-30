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

package notsafetoevict

import (
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/drainability"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/options"
	"k8s.io/autoscaler/cluster-autoscaler/utils/drain"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"

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

		safePod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "bar",
				Namespace: "kube-system",
				Annotations: map[string]string{
					drain.PodSafeToEvictKey: "true",
				},
			},
			Spec: apiv1.PodSpec{
				NodeName: "node",
			},
		}

		unsafeSystemFailedPod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "bar",
				Namespace: "kube-system",
				Annotations: map[string]string{
					drain.PodSafeToEvictKey: "false",
				},
			},
			Spec: apiv1.PodSpec{
				NodeName:      "node",
				RestartPolicy: apiv1.RestartPolicyNever,
			},
			Status: apiv1.PodStatus{
				Phase: apiv1.PodFailed,
			},
		}

		unsafeSystemTerminalPod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "bar",
				Namespace: "kube-system",
				Annotations: map[string]string{
					drain.PodSafeToEvictKey: "false",
				},
			},
			Spec: apiv1.PodSpec{
				NodeName:      "node",
				RestartPolicy: apiv1.RestartPolicyOnFailure,
			},
			Status: apiv1.PodStatus{
				Phase: apiv1.PodSucceeded,
			},
		}

		unsafeSystemEvictedPod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "bar",
				Namespace: "kube-system",
				Annotations: map[string]string{
					drain.PodSafeToEvictKey: "false",
				},
			},
			Spec: apiv1.PodSpec{
				NodeName:      "node",
				RestartPolicy: apiv1.RestartPolicyAlways,
			},
			Status: apiv1.PodStatus{
				Phase: apiv1.PodFailed,
			},
		}

		zeroGracePeriod          = int64(0)
		unsafeLongTerminatingPod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "bar",
				Namespace:         "kube-system",
				DeletionTimestamp: &metav1.Time{Time: testTime.Add(-2 * drain.PodLongTerminatingExtraThreshold)},
				Annotations: map[string]string{
					drain.PodSafeToEvictKey: "false",
				},
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

		extendedGracePeriod                             = int64(6 * 60) // 6 minutes
		unsafeLongTerminatingPodWithExtendedGracePeriod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "bar",
				Namespace:         "kube-system",
				DeletionTimestamp: &metav1.Time{Time: testTime.Add(-2 * time.Duration(extendedGracePeriod) * time.Second)},
				Annotations: map[string]string{
					drain.PodSafeToEvictKey: "false",
				},
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

		unsafeRcPod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "bar",
				Namespace:       "default",
				OwnerReferences: GenerateOwnerReferences(rc.Name, "ReplicationController", "core/v1", ""),
				Annotations: map[string]string{
					drain.PodSafeToEvictKey: "false",
				},
			},
			Spec: apiv1.PodSpec{
				NodeName: "node",
			},
		}

		unsafeJobPod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "bar",
				Namespace:       "default",
				OwnerReferences: GenerateOwnerReferences(job.Name, "Job", "batch/v1", ""),
				Annotations: map[string]string{
					drain.PodSafeToEvictKey: "false",
				},
			},
		}
	)

	for _, test := range []struct {
		desc string
		pod  *apiv1.Pod
		rcs  []*apiv1.ReplicationController
		rss  []*appsv1.ReplicaSet

		wantReason drain.BlockingPodReason
		wantError  bool
	}{
		{
			desc: "pod with PodSafeToEvict annotation",
			pod:  safePod,
		},
		{
			desc: "RC-managed pod with no annotation",
			pod:  rcPod,
			rcs:  []*apiv1.ReplicationController{&rc},
		},
		{
			desc:       "RC-managed pod with PodSafeToEvict=false annotation",
			pod:        unsafeRcPod,
			rcs:        []*apiv1.ReplicationController{&rc},
			wantReason: drain.NotSafeToEvictAnnotation,
			wantError:  true,
		},
		{
			desc: "Job-managed pod with no annotation",
			pod:  jobPod,
			rcs:  []*apiv1.ReplicationController{&rc},
		},
		{
			desc:       "Job-managed pod with PodSafeToEvict=false annotation",
			pod:        unsafeJobPod,
			rcs:        []*apiv1.ReplicationController{&rc},
			wantReason: drain.NotSafeToEvictAnnotation,
			wantError:  true,
		},

		{
			desc: "unsafe failed pod",
			pod:  unsafeSystemFailedPod,
		},
		{
			desc: "unsafe terminal pod",
			pod:  unsafeSystemTerminalPod,
		},
		{
			desc: "unsafe evicted pod",
			pod:  unsafeSystemEvictedPod,
		},
		{
			desc: "unsafe long terminating pod with 0 grace period",
			pod:  unsafeLongTerminatingPod,
		},
		{
			desc: "unsafe long terminating pod with extended grace period",
			pod:  unsafeLongTerminatingPodWithExtendedGracePeriod,
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			drainCtx := &drainability.DrainContext{
				DeleteOptions: options.NodeDeleteOptions{
					SkipNodesWithSystemPods: true,
				},
				Timestamp: testTime,
			}
			status := New().Drainable(drainCtx, test.pod)
			assert.Equal(t, test.wantReason, status.BlockingReason)
			assert.Equal(t, test.wantError, status.Error != nil)
		})
	}
}
