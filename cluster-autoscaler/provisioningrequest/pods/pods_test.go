/*
Copyright 2024 The Kubernetes Authors.

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

package pods

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/proto"

	// "google.golang.org/protobuf/testing/protocmp"
	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/autoscaler/cluster-autoscaler/apis/provisioningrequest/autoscaling.x-k8s.io/v1beta1"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqwrapper"
)

const testProvisioningClassName = "TestProvisioningClass"

func TestPodsForProvisioningRequest(t *testing.T) {
	testPod := func(name, genName, containerName, containerImage, prName string) *v1.Pod {
		return &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:         name,
				GenerateName: genName,
				Namespace:    "test-namespace",
				UID:          types.UID(fmt.Sprintf("test-namespace/%s", name)),
				Annotations: map[string]string{
					ProvisioningRequestPodAnnotationKey: prName,
					ProvisioningClassPodAnnotationKey:   testProvisioningClassName,
				},
				Labels:     map[string]string{},
				Finalizers: []string{},
				OwnerReferences: []metav1.OwnerReference{
					{
						Controller: proto.Bool(true),
						Name:       prName,
					},
				},
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:  containerName,
						Image: containerImage,
					},
				},
			},
		}
	}

	tests := []struct {
		desc         string
		pr           *v1beta1.ProvisioningRequest
		podTemplates []*apiv1.PodTemplate
		want         []*v1.Pod
		wantErr      bool
	}{
		{
			desc: "simple ProvReq",
			pr: &v1beta1.ProvisioningRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pr-name",
					Namespace: "test-namespace",
				},
				Spec: v1beta1.ProvisioningRequestSpec{
					PodSets: []v1beta1.PodSet{
						{
							Count:          1,
							PodTemplateRef: v1beta1.Reference{Name: "template-1"},
						},
					},
					ProvisioningClassName: testProvisioningClassName,
				},
			},
			podTemplates: []*apiv1.PodTemplate{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "template-1",
						Namespace: "test-namespace",
					},
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name:  "test-container",
									Image: "test-image",
								},
							},
						},
					},
				},
			},
			want: []*v1.Pod{
				testPod("test-pr-name-0-0", "test-pr-name-", "test-container", "test-image", "test-pr-name"),
			},
		},
		{
			desc: "ProvReq already having taint",
			pr: &v1beta1.ProvisioningRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pr-name",
					Namespace: "test-namespace",
				},
				Spec: v1beta1.ProvisioningRequestSpec{
					PodSets: []v1beta1.PodSet{
						{
							Count:          1,
							PodTemplateRef: v1beta1.Reference{Name: "template-1"},
						},
					},
					ProvisioningClassName: testProvisioningClassName,
				},
			},
			podTemplates: []*apiv1.PodTemplate{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "template-1",
						Namespace: "test-namespace",
					},
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name:  "test-container",
									Image: "test-image",
								},
							},
						},
					},
				},
			},
			want: []*v1.Pod{
				testPod("test-pr-name-0-0", "test-pr-name-", "test-container", "test-image", "test-pr-name"),
			},
		},
		{
			desc: "ProvReq already having nodeSelector",
			pr: &v1beta1.ProvisioningRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pr-name",
					Namespace: "test-namespace",
				},
				Spec: v1beta1.ProvisioningRequestSpec{
					PodSets: []v1beta1.PodSet{
						{
							Count:          1,
							PodTemplateRef: v1beta1.Reference{Name: "template-1"},
						},
					},
					ProvisioningClassName: testProvisioningClassName,
				},
			},
			podTemplates: []*v1.PodTemplate{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "template-1",
						Namespace: "test-namespace",
					},
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name:  "test-container",
									Image: "test-image",
								},
							},
						},
					},
				},
			},
			want: []*v1.Pod{
				testPod("test-pr-name-0-0", "test-pr-name-", "test-container", "test-image", "test-pr-name"),
			},
		},
		{
			desc: "ProvReq with multiple pod sets",
			pr: &v1beta1.ProvisioningRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pr-name",
					Namespace: "test-namespace",
				},
				Spec: v1beta1.ProvisioningRequestSpec{
					PodSets: []v1beta1.PodSet{
						{
							Count:          2,
							PodTemplateRef: v1beta1.Reference{Name: "template-1"},
						},
						{
							Count:          3,
							PodTemplateRef: v1beta1.Reference{Name: "template-2"},
						},
					},
					ProvisioningClassName: testProvisioningClassName,
				},
			},
			podTemplates: []*v1.PodTemplate{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "template-1",
						Namespace: "test-namespace",
					},
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name:  "test-container",
									Image: "test-image",
								},
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "template-2",
						Namespace: "test-namespace",
					},
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name:  "test-container-2",
									Image: "test-image-2",
								},
							},
						},
					},
				},
			},
			want: []*v1.Pod{
				testPod("test-pr-name-0-0", "test-pr-name-", "test-container", "test-image", "test-pr-name"),
				testPod("test-pr-name-0-1", "test-pr-name-", "test-container", "test-image", "test-pr-name"),
				testPod("test-pr-name-1-0", "test-pr-name-", "test-container-2", "test-image-2", "test-pr-name"),
				testPod("test-pr-name-1-1", "test-pr-name-", "test-container-2", "test-image-2", "test-pr-name"),
				testPod("test-pr-name-1-2", "test-pr-name-", "test-container-2", "test-image-2", "test-pr-name"),
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := PodsForProvisioningRequest(provreqwrapper.NewV1Beta1ProvisioningRequest(tc.pr, tc.podTemplates))
			if (err != nil) != tc.wantErr {
				t.Errorf("PodsForProvisioningRequest() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if diff := cmp.Diff(got, tc.want); diff != "" {
				t.Errorf("unexpected response from PodsForProvisioningRequest(), diff (-want +got): %v", diff)
			}
		})
	}
}
