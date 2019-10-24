/*
Copyright 2016 The Kubernetes Authors.

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

package simulator

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	apiv1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/kubelet/types"

	pod_util "k8s.io/autoscaler/cluster-autoscaler/utils/pod"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

func TestRequiredPodsForNode(t *testing.T) {
	nodeName1 := "node1"
	nodeName2 := "node2"
	pod1 := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "pod1",
			SelfLink:  "pod1",
		},
		Spec: apiv1.PodSpec{
			NodeName: nodeName1,
		},
	}
	// Manifest pod.
	pod2 := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod2",
			Namespace: "kube-system",
			SelfLink:  "pod2",
			Annotations: map[string]string{
				types.ConfigMirrorAnnotationKey: "something",
			},
		},
		Spec: apiv1.PodSpec{
			NodeName: nodeName1,
		},
	}
	pod3 := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod2",
			Namespace: "kube-system",
			SelfLink:  "pod2",
			Annotations: map[string]string{
				types.ConfigMirrorAnnotationKey: "something",
			},
		},
		Spec: apiv1.PodSpec{
			NodeName: nodeName2,
		},
	}

	podsForNodes := map[string][]*apiv1.Pod{nodeName1: {pod1, pod2}, nodeName2: {pod3}}
	pods, err := getRequiredPodsForNode(nodeName1, podsForNodes)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(pods))
	assert.Equal(t, "pod2", pods[0].Name)
}

func Test_filterRequiredPodsForNode(t *testing.T) {
	nodeName := "node1"
	pod1 := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "pod1",
			SelfLink:  "pod1",
		},
		Spec: apiv1.PodSpec{
			NodeName: nodeName,
		},
	}
	// Manifest pod.
	mirrorPod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mirrorPod",
			Namespace: "kube-system",
			SelfLink:  "mirrorPod",
			Annotations: map[string]string{
				types.ConfigMirrorAnnotationKey: "something",
			},
		},
		Spec: apiv1.PodSpec{
			NodeName: nodeName,
		},
	}
	now := metav1.NewTime(time.Now())
	podDeleted := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "podDeleted",
			SelfLink:  "podDeleted",
			Annotations: map[string]string{
				types.ConfigMirrorAnnotationKey: "something",
			},
			DeletionTimestamp: &now,
		},
		Spec: apiv1.PodSpec{
			NodeName: nodeName,
		},
	}

	podDaemonset := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       "default",
			Name:            "podDaemonset",
			SelfLink:        "podDaemonset",
			OwnerReferences: GenerateOwnerReferences("ds", "DaemonSet", "apps/v1", ""),
			Annotations: map[string]string{
				types.ConfigSourceAnnotationKey: types.FileSource,
			},
		},
		Spec: apiv1.PodSpec{
			NodeName: nodeName,
		},
	}

	podDaemonsetAnnotation := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       "default",
			Name:            "podDaemonset2",
			SelfLink:        "podDaemonset2",
			OwnerReferences: GenerateOwnerReferences("ds2", "CustomDaemonset", "crd/v1", ""),
			Annotations: map[string]string{
				pod_util.DaemonSetPodAnnotationKey: "true",
			},
		},
		Spec: apiv1.PodSpec{
			NodeName: nodeName,
		},
	}

	tests := []struct {
		name      string
		inputPods []*apiv1.Pod
		want      []*apiv1.Pod
	}{
		{
			name:      "nil input pod list",
			inputPods: nil,
			want:      []*apiv1.Pod{},
		},
		{
			name:      "should return only mirrorPod",
			inputPods: []*apiv1.Pod{pod1, mirrorPod},
			want:      []*apiv1.Pod{mirrorPod},
		},
		{
			name:      "should ignore podDeleted",
			inputPods: []*apiv1.Pod{pod1, mirrorPod, podDeleted},
			want:      []*apiv1.Pod{mirrorPod},
		},
		{
			name:      "should return daemonset pod",
			inputPods: []*apiv1.Pod{pod1, podDaemonset},
			want:      []*apiv1.Pod{podDaemonset},
		},
		{
			name:      "should return daemonset pods with",
			inputPods: []*apiv1.Pod{pod1, podDaemonset, podDaemonsetAnnotation},
			want:      []*apiv1.Pod{podDaemonset, podDaemonsetAnnotation},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filterRequiredPodsForNode(tt.inputPods); !apiequality.Semantic.DeepEqual(got, tt.want) {
				t.Errorf("filterRequiredPodsForNode() = %v, want %v", got, tt.want)
			}
		})
	}
}
