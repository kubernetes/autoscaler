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

package localstorage

import (
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
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

		emptydirPod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "bar",
				Namespace:       "default",
				OwnerReferences: GenerateOwnerReferences(rc.Name, "ReplicationController", "core/v1", ""),
			},
			Spec: apiv1.PodSpec{
				NodeName: "node",
				Volumes: []apiv1.Volume{
					{
						Name:         "scratch",
						VolumeSource: apiv1.VolumeSource{EmptyDir: &apiv1.EmptyDirVolumeSource{Medium: ""}},
					},
				},
			},
		}

		emptyDirSafeToEvictVolumeSingleVal = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "bar",
				Namespace:       "default",
				OwnerReferences: GenerateOwnerReferences(rc.Name, "ReplicationController", "core/v1", ""),
				Annotations: map[string]string{
					drain.SafeToEvictLocalVolumesKey: "scratch",
				},
			},
			Spec: apiv1.PodSpec{
				NodeName: "node",
				Volumes: []apiv1.Volume{
					{
						Name:         "scratch",
						VolumeSource: apiv1.VolumeSource{EmptyDir: &apiv1.EmptyDirVolumeSource{Medium: ""}},
					},
				},
			},
		}

		emptyDirSafeToEvictLocalVolumeSingleValEmpty = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "bar",
				Namespace:       "default",
				OwnerReferences: GenerateOwnerReferences(rc.Name, "ReplicationController", "core/v1", ""),
				Annotations: map[string]string{
					drain.SafeToEvictLocalVolumesKey: "",
				},
			},
			Spec: apiv1.PodSpec{
				NodeName: "node",
				Volumes: []apiv1.Volume{
					{
						Name:         "scratch",
						VolumeSource: apiv1.VolumeSource{EmptyDir: &apiv1.EmptyDirVolumeSource{Medium: ""}},
					},
				},
			},
		}

		emptyDirSafeToEvictLocalVolumeSingleValNonMatching = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "bar",
				Namespace:       "default",
				OwnerReferences: GenerateOwnerReferences(rc.Name, "ReplicationController", "core/v1", ""),
				Annotations: map[string]string{
					drain.SafeToEvictLocalVolumesKey: "scratch-2",
				},
			},
			Spec: apiv1.PodSpec{
				NodeName: "node",
				Volumes: []apiv1.Volume{
					{
						Name:         "scratch-1",
						VolumeSource: apiv1.VolumeSource{EmptyDir: &apiv1.EmptyDirVolumeSource{Medium: ""}},
					},
				},
			},
		}

		emptyDirSafeToEvictLocalVolumeMultiValAllMatching = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "bar",
				Namespace:       "default",
				OwnerReferences: GenerateOwnerReferences(rc.Name, "ReplicationController", "core/v1", ""),
				Annotations: map[string]string{
					drain.SafeToEvictLocalVolumesKey: "scratch-1,scratch-2,scratch-3",
				},
			},
			Spec: apiv1.PodSpec{
				NodeName: "node",
				Volumes: []apiv1.Volume{
					{
						Name:         "scratch-1",
						VolumeSource: apiv1.VolumeSource{EmptyDir: &apiv1.EmptyDirVolumeSource{Medium: ""}},
					},
					{
						Name:         "scratch-2",
						VolumeSource: apiv1.VolumeSource{EmptyDir: &apiv1.EmptyDirVolumeSource{Medium: ""}},
					},
					{
						Name:         "scratch-3",
						VolumeSource: apiv1.VolumeSource{EmptyDir: &apiv1.EmptyDirVolumeSource{Medium: ""}},
					},
				},
			},
		}

		emptyDirSafeToEvictLocalVolumeMultiValNonMatching = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "bar",
				Namespace:       "default",
				OwnerReferences: GenerateOwnerReferences(rc.Name, "ReplicationController", "core/v1", ""),
				Annotations: map[string]string{
					drain.SafeToEvictLocalVolumesKey: "scratch-1,scratch-2,scratch-5",
				},
			},
			Spec: apiv1.PodSpec{
				NodeName: "node",
				Volumes: []apiv1.Volume{
					{
						Name:         "scratch-1",
						VolumeSource: apiv1.VolumeSource{EmptyDir: &apiv1.EmptyDirVolumeSource{Medium: ""}},
					},
					{
						Name:         "scratch-2",
						VolumeSource: apiv1.VolumeSource{EmptyDir: &apiv1.EmptyDirVolumeSource{Medium: ""}},
					},
					{
						Name:         "scratch-3",
						VolumeSource: apiv1.VolumeSource{EmptyDir: &apiv1.EmptyDirVolumeSource{Medium: ""}},
					},
				},
			},
		}

		emptyDirSafeToEvictLocalVolumeMultiValSomeMatchingVals = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "bar",
				Namespace:       "default",
				OwnerReferences: GenerateOwnerReferences(rc.Name, "ReplicationController", "core/v1", ""),
				Annotations: map[string]string{
					drain.SafeToEvictLocalVolumesKey: "scratch-1,scratch-2",
				},
			},
			Spec: apiv1.PodSpec{
				NodeName: "node",
				Volumes: []apiv1.Volume{
					{
						Name:         "scratch-1",
						VolumeSource: apiv1.VolumeSource{EmptyDir: &apiv1.EmptyDirVolumeSource{Medium: ""}},
					},
					{
						Name:         "scratch-2",
						VolumeSource: apiv1.VolumeSource{EmptyDir: &apiv1.EmptyDirVolumeSource{Medium: ""}},
					},
					{
						Name:         "scratch-3",
						VolumeSource: apiv1.VolumeSource{EmptyDir: &apiv1.EmptyDirVolumeSource{Medium: ""}},
					},
				},
			},
		}

		emptyDirSafeToEvictLocalVolumeMultiValEmpty = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "bar",
				Namespace:       "default",
				OwnerReferences: GenerateOwnerReferences(rc.Name, "ReplicationController", "core/v1", ""),
				Annotations: map[string]string{
					drain.SafeToEvictLocalVolumesKey: ",",
				},
			},
			Spec: apiv1.PodSpec{
				NodeName: "node",
				Volumes: []apiv1.Volume{
					{
						Name:         "scratch-1",
						VolumeSource: apiv1.VolumeSource{EmptyDir: &apiv1.EmptyDirVolumeSource{Medium: ""}},
					},
					{
						Name:         "scratch-2",
						VolumeSource: apiv1.VolumeSource{EmptyDir: &apiv1.EmptyDirVolumeSource{Medium: ""}},
					},
					{
						Name:         "scratch-3",
						VolumeSource: apiv1.VolumeSource{EmptyDir: &apiv1.EmptyDirVolumeSource{Medium: ""}},
					},
				},
			},
		}

		emptyDirFailedPod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "bar",
				Namespace: "default",
			},
			Spec: apiv1.PodSpec{
				NodeName:      "node",
				RestartPolicy: apiv1.RestartPolicyNever,
				Volumes: []apiv1.Volume{
					{
						Name:         "scratch",
						VolumeSource: apiv1.VolumeSource{EmptyDir: &apiv1.EmptyDirVolumeSource{Medium: ""}},
					},
				},
			},
			Status: apiv1.PodStatus{
				Phase: apiv1.PodFailed,
			},
		}

		emptyDirTerminalPod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "bar",
				Namespace: "default",
			},
			Spec: apiv1.PodSpec{
				NodeName:      "node",
				RestartPolicy: apiv1.RestartPolicyOnFailure,
				Volumes: []apiv1.Volume{
					{
						Name:         "scratch",
						VolumeSource: apiv1.VolumeSource{EmptyDir: &apiv1.EmptyDirVolumeSource{Medium: ""}},
					},
				},
			},
			Status: apiv1.PodStatus{
				Phase: apiv1.PodSucceeded,
			},
		}

		emptyDirEvictedPod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "bar",
				Namespace: "default",
			},
			Spec: apiv1.PodSpec{
				NodeName:      "node",
				RestartPolicy: apiv1.RestartPolicyAlways,
				Volumes: []apiv1.Volume{
					{
						Name:         "scratch",
						VolumeSource: apiv1.VolumeSource{EmptyDir: &apiv1.EmptyDirVolumeSource{Medium: ""}},
					},
				},
			},
			Status: apiv1.PodStatus{
				Phase: apiv1.PodFailed,
			},
		}

		emptyDirSafePod = &apiv1.Pod{
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

		zeroGracePeriod            = int64(0)
		emptyDirLongTerminatingPod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "bar",
				Namespace:         "default",
				DeletionTimestamp: &metav1.Time{Time: testTime.Add(-2 * drain.PodLongTerminatingExtraThreshold)},
			},
			Spec: apiv1.PodSpec{
				NodeName:                      "node",
				RestartPolicy:                 apiv1.RestartPolicyOnFailure,
				TerminationGracePeriodSeconds: &zeroGracePeriod,
				Volumes: []apiv1.Volume{
					{
						Name:         "scratch",
						VolumeSource: apiv1.VolumeSource{EmptyDir: &apiv1.EmptyDirVolumeSource{Medium: ""}},
					},
				},
			},
			Status: apiv1.PodStatus{
				Phase: apiv1.PodUnknown,
			},
		}

		extendedGracePeriod                               = int64(6 * 60) // 6 minutes
		emptyDirLongTerminatingPodWithExtendedGracePeriod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "bar",
				Namespace:         "default",
				DeletionTimestamp: &metav1.Time{Time: testTime.Add(-2 * time.Duration(extendedGracePeriod) * time.Second)},
			},
			Spec: apiv1.PodSpec{
				NodeName:                      "node",
				RestartPolicy:                 apiv1.RestartPolicyOnFailure,
				TerminationGracePeriodSeconds: &extendedGracePeriod,
				Volumes: []apiv1.Volume{
					{
						Name:         "scratch",
						VolumeSource: apiv1.VolumeSource{EmptyDir: &apiv1.EmptyDirVolumeSource{Medium: ""}},
					},
				},
			},
			Status: apiv1.PodStatus{
				Phase: apiv1.PodUnknown,
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
			desc:       "pod with EmptyDir",
			pod:        emptydirPod,
			rcs:        []*apiv1.ReplicationController{&rc},
			wantReason: drain.LocalStorageRequested,
			wantError:  true,
		},
		{
			desc: "pod with EmptyDir and SafeToEvictLocalVolumesKey annotation",
			pod:  emptyDirSafeToEvictVolumeSingleVal,
			rcs:  []*apiv1.ReplicationController{&rc},
		},
		{
			desc:       "pod with EmptyDir and empty value for SafeToEvictLocalVolumesKey annotation",
			pod:        emptyDirSafeToEvictLocalVolumeSingleValEmpty,
			rcs:        []*apiv1.ReplicationController{&rc},
			wantReason: drain.LocalStorageRequested,
			wantError:  true,
		},
		{
			desc:       "pod with EmptyDir and non-matching value for SafeToEvictLocalVolumesKey annotation",
			pod:        emptyDirSafeToEvictLocalVolumeSingleValNonMatching,
			rcs:        []*apiv1.ReplicationController{&rc},
			wantReason: drain.LocalStorageRequested,
			wantError:  true,
		},
		{
			desc: "pod with EmptyDir and SafeToEvictLocalVolumesKey annotation with matching values",
			pod:  emptyDirSafeToEvictLocalVolumeMultiValAllMatching,
			rcs:  []*apiv1.ReplicationController{&rc},
		},
		{
			desc:       "pod with EmptyDir and SafeToEvictLocalVolumesKey annotation with non-matching values",
			pod:        emptyDirSafeToEvictLocalVolumeMultiValNonMatching,
			rcs:        []*apiv1.ReplicationController{&rc},
			wantReason: drain.LocalStorageRequested,
			wantError:  true,
		},
		{
			desc:       "pod with EmptyDir and SafeToEvictLocalVolumesKey annotation with some matching values",
			pod:        emptyDirSafeToEvictLocalVolumeMultiValSomeMatchingVals,
			rcs:        []*apiv1.ReplicationController{&rc},
			wantReason: drain.LocalStorageRequested,
			wantError:  true,
		},
		{
			desc:       "pod with EmptyDir and SafeToEvictLocalVolumesKey annotation empty values",
			pod:        emptyDirSafeToEvictLocalVolumeMultiValEmpty,
			rcs:        []*apiv1.ReplicationController{&rc},
			wantReason: drain.LocalStorageRequested,
			wantError:  true,
		},

		{
			desc: "EmptyDir failed pod",
			pod:  emptyDirFailedPod,
		},
		{
			desc: "EmptyDir terminal pod",
			pod:  emptyDirTerminalPod,
		},
		{
			desc: "EmptyDir evicted pod",
			pod:  emptyDirEvictedPod,
		},
		{
			desc: "EmptyDir pod with PodSafeToEvict annotation",
			pod:  emptyDirSafePod,
		},
		{
			desc: "EmptyDir long terminating pod with 0 grace period",
			pod:  emptyDirLongTerminatingPod,
		},
		{
			desc: "EmptyDir long terminating pod with extended grace period",
			pod:  emptyDirLongTerminatingPodWithExtendedGracePeriod,
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			drainCtx := &drainability.DrainContext{
				DeleteOptions: options.NodeDeleteOptions{
					SkipNodesWithLocalStorage: true,
				},
				Timestamp: testTime,
			}
			status := New().Drainable(drainCtx, test.pod)
			assert.Equal(t, test.wantReason, status.BlockingReason)
			assert.Equal(t, test.wantError, status.Error != nil)
		})
	}
}
