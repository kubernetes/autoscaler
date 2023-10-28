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
	"k8s.io/autoscaler/cluster-autoscaler/utils/drain"
	"k8s.io/autoscaler/cluster-autoscaler/utils/test"

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
	)

	for desc, test := range map[string]struct {
		pod *apiv1.Pod
		rcs []*apiv1.ReplicationController
		rss []*appsv1.ReplicaSet

		wantReason drain.BlockingPodReason
		wantError  bool
	}{
		"pod with EmptyDir": {
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "bar",
					Namespace:       "default",
					OwnerReferences: test.GenerateOwnerReferences(rc.Name, "ReplicationController", "core/v1", ""),
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
			},
			rcs:        []*apiv1.ReplicationController{&rc},
			wantReason: drain.LocalStorageRequested,
			wantError:  true,
		},
		"pod with EmptyDir and SafeToEvictLocalVolumesKey annotation": {
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "bar",
					Namespace:       "default",
					OwnerReferences: test.GenerateOwnerReferences(rc.Name, "ReplicationController", "core/v1", ""),
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
			},
			rcs: []*apiv1.ReplicationController{&rc},
		},
		"pod with EmptyDir and empty value for SafeToEvictLocalVolumesKey annotation": {
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "bar",
					Namespace:       "default",
					OwnerReferences: test.GenerateOwnerReferences(rc.Name, "ReplicationController", "core/v1", ""),
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
			},
			rcs:        []*apiv1.ReplicationController{&rc},
			wantReason: drain.LocalStorageRequested,
			wantError:  true,
		},
		"pod with EmptyDir and non-matching value for SafeToEvictLocalVolumesKey annotation": {
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "bar",
					Namespace:       "default",
					OwnerReferences: test.GenerateOwnerReferences(rc.Name, "ReplicationController", "core/v1", ""),
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
			},
			rcs:        []*apiv1.ReplicationController{&rc},
			wantReason: drain.LocalStorageRequested,
			wantError:  true,
		},
		"pod with EmptyDir and SafeToEvictLocalVolumesKey annotation with matching values": {
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "bar",
					Namespace:       "default",
					OwnerReferences: test.GenerateOwnerReferences(rc.Name, "ReplicationController", "core/v1", ""),
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
			},
			rcs: []*apiv1.ReplicationController{&rc},
		},
		"pod with EmptyDir and SafeToEvictLocalVolumesKey annotation with non-matching values": {
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "bar",
					Namespace:       "default",
					OwnerReferences: test.GenerateOwnerReferences(rc.Name, "ReplicationController", "core/v1", ""),
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
			},
			rcs:        []*apiv1.ReplicationController{&rc},
			wantReason: drain.LocalStorageRequested,
			wantError:  true,
		},
		"pod with EmptyDir and SafeToEvictLocalVolumesKey annotation with some matching values": {
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "bar",
					Namespace:       "default",
					OwnerReferences: test.GenerateOwnerReferences(rc.Name, "ReplicationController", "core/v1", ""),
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
			},
			rcs:        []*apiv1.ReplicationController{&rc},
			wantReason: drain.LocalStorageRequested,
			wantError:  true,
		},
		"pod with EmptyDir and SafeToEvictLocalVolumesKey annotation empty values": {
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "bar",
					Namespace:       "default",
					OwnerReferences: test.GenerateOwnerReferences(rc.Name, "ReplicationController", "core/v1", ""),
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
			},
			rcs:        []*apiv1.ReplicationController{&rc},
			wantReason: drain.LocalStorageRequested,
			wantError:  true,
		},
	} {
		t.Run(desc, func(t *testing.T) {
			drainCtx := &drainability.DrainContext{
				Timestamp: testTime,
			}
			status := New().Drainable(drainCtx, test.pod)
			assert.Equal(t, test.wantReason, status.BlockingReason)
			assert.Equal(t, test.wantError, status.Error != nil)
		})
	}
}
