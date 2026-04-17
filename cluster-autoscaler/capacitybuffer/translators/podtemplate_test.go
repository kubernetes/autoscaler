/*
Copyright The Kubernetes Authors.

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

package translator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1beta1"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer"
	"k8s.io/utils/ptr"
)

func TestGetPodTemplateFromSpec(t *testing.T) {
	testCases := []struct {
		name   string
		pts    *corev1.PodTemplateSpec
		buffer *apiv1.CapacityBuffer
		want   *corev1.PodTemplate
	}{
		{
			name: "basic pod template spec",
			pts: &corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      map[string]string{"app": "foo"},
					Annotations: map[string]string{"anno": "bar"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: "c1"}},
				},
			},
			buffer: &apiv1.CapacityBuffer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-buffer",
					Namespace: "test-ns",
					UID:       "123-uid",
				},
			},
			want: &corev1.PodTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "capacitybuffer-test-buffer-pod-template",
					Namespace: "test-ns",
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: capacitybuffer.CapacityBufferApiVersion,
							Kind:       capacitybuffer.CapacityBufferKind,
							Name:       "test-buffer",
							UID:        "123-uid",
							Controller: ptr.To(true),
						},
					},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels:      map[string]string{"app": "foo"},
						Annotations: map[string]string{"anno": "bar"},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{Name: "c1"}},
					},
				},
			},
		},
		{
			name: "long buffer name",
			pts: &corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "foo"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: "c1"}},
				},
			},
			buffer: &apiv1.CapacityBuffer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "this-is-a-very-long-buffer-name-that-will-exceed-the-maximum-allowed-length-for-a-pod-template-name-which-is-two-hundred-and-fifty-three-characters-in-total-so-it-will-be-truncated-to-fit-the-limit-and-we-need-to-make-sure-it-does-not-exceed-the-limit-whatsoever",
					Namespace: "test-ns",
					UID:       "123-uid",
				},
			},
			want: &corev1.PodTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "capacitybuffer-this-is-a-very-long-buffer-name-that-will-exceed-the-maximum-allowed-length-for-a-pod-template-name-which-is-two-hundred-and-fifty-three-characters-in-total-so-it-will-be-truncated-to-fit-the-limit-and-we-need-to-make-sure-it-pod-template",
					Namespace: "test-ns",
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: capacitybuffer.CapacityBufferApiVersion,
							Kind:       capacitybuffer.CapacityBufferKind,
							Name:       "this-is-a-very-long-buffer-name-that-will-exceed-the-maximum-allowed-length-for-a-pod-template-name-which-is-two-hundred-and-fifty-three-characters-in-total-so-it-will-be-truncated-to-fit-the-limit-and-we-need-to-make-sure-it-does-not-exceed-the-limit-whatsoever",
							UID:        "123-uid",
							Controller: ptr.To(true),
						},
					},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{"app": "foo"},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{Name: "c1"}},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := getPodTemplateFromSpec(tc.pts, tc.buffer)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestGetPodTemplateSpecFromPod(t *testing.T) {
	testCases := []struct {
		name string
		pod  *corev1.Pod
		want *corev1.PodTemplateSpec
	}{
		{
			name: "pod without service account volumes",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      map[string]string{"app": "foo"},
					Annotations: map[string]string{"anno": "bar"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: "c1"}},
				},
			},
			want: &corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      map[string]string{"app": "foo"},
					Annotations: map[string]string{"anno": "bar"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: "c1"}},
				},
			},
		},
		{
			name: "pod with service account volumes",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      map[string]string{"app": "foo"},
					Annotations: map[string]string{"anno": "bar"},
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{Name: "vol-1"},
						{Name: "kube-api-access-abcde"},
					},
					Containers: []corev1.Container{
						{
							Name: "c1",
							VolumeMounts: []corev1.VolumeMount{
								{Name: "vol-1", MountPath: "/vol-1"},
								{Name: "kube-api-access-abcde", MountPath: "/var/run/secrets/kubernetes.io/serviceaccount"},
							},
						},
					},
					InitContainers: []corev1.Container{
						{
							Name: "i1",
							VolumeMounts: []corev1.VolumeMount{
								{Name: "vol-1", MountPath: "/vol-1"},
								{Name: "kube-api-access-abcde", MountPath: "/var/run/secrets/kubernetes.io/serviceaccount"},
							},
						},
					},
				},
			},
			want: &corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      map[string]string{"app": "foo"},
					Annotations: map[string]string{"anno": "bar"},
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{Name: "vol-1"},
					},
					Containers: []corev1.Container{
						{
							Name: "c1",
							VolumeMounts: []corev1.VolumeMount{
								{Name: "vol-1", MountPath: "/vol-1"},
							},
						},
					},
					InitContainers: []corev1.Container{
						{
							Name: "i1",
							VolumeMounts: []corev1.VolumeMount{
								{Name: "vol-1", MountPath: "/vol-1"},
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := getPodTemplateSpecFromPod(tc.pod)
			assert.Equal(t, tc.want, got)
		})
	}
}
