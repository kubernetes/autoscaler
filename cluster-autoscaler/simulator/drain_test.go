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

	apiv1 "k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/kubelet/types"
	"k8s.io/kubernetes/plugin/pkg/scheduler/schedulercache"

	"github.com/stretchr/testify/assert"
)

func TestFastGetPodsToMove(t *testing.T) {

	// Unreplicated pod
	pod1 := &apiv1.Pod{
		ObjectMeta: apiv1.ObjectMeta{
			Name:      "pod1",
			Namespace: "ns",
		},
	}
	_, err := FastGetPodsToMove(schedulercache.NewNodeInfo(pod1), true, true)
	assert.Error(t, err)

	// Replicated pod
	pod2 := &apiv1.Pod{
		ObjectMeta: apiv1.ObjectMeta{
			Name:      "pod2",
			Namespace: "ns",
			Annotations: map[string]string{
				"kubernetes.io/created-by": "{\"kind\":\"SerializedReference\",\"apiVersion\":\"v1\",\"reference\":{\"kind\":\"ReplicaSet\"}}",
			},
		},
	}
	r2, err := FastGetPodsToMove(schedulercache.NewNodeInfo(pod2), true, true)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(r2))
	assert.Equal(t, pod2, r2[0])

	// Manifest pod
	pod3 := &apiv1.Pod{
		ObjectMeta: apiv1.ObjectMeta{
			Name:      "pod3",
			Namespace: "kube-system",
			Annotations: map[string]string{
				types.ConfigMirrorAnnotationKey: "something",
			},
		},
	}
	r3, err := FastGetPodsToMove(schedulercache.NewNodeInfo(pod3), true, true)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(r3))

	// DeamonSet pod
	pod4 := &apiv1.Pod{
		ObjectMeta: apiv1.ObjectMeta{
			Name:      "pod4",
			Namespace: "ns",
			Annotations: map[string]string{
				"kubernetes.io/created-by": "{\"kind\":\"SerializedReference\",\"apiVersion\":\"v1\",\"reference\":{\"kind\":\"DaemonSet\"}}",
			},
		},
	}
	r4, err := FastGetPodsToMove(schedulercache.NewNodeInfo(pod2, pod3, pod4), true, true)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(r4))
	assert.Equal(t, pod2, r4[0])

	// Kube-system
	pod5 := &apiv1.Pod{
		ObjectMeta: apiv1.ObjectMeta{
			Name:      "pod5",
			Namespace: "kube-system",
			Annotations: map[string]string{
				"kubernetes.io/created-by": "{\"kind\":\"SerializedReference\",\"apiVersion\":\"v1\",\"reference\":{\"kind\":\"ReplicaSet\"}}",
			},
		},
	}
	_, err = FastGetPodsToMove(schedulercache.NewNodeInfo(pod5), true, true)
	assert.Error(t, err)

	// Local storage
	pod6 := &apiv1.Pod{
		ObjectMeta: apiv1.ObjectMeta{
			Name:      "pod6",
			Namespace: "ns",
			Annotations: map[string]string{
				"kubernetes.io/created-by": "{\"kind\":\"SerializedReference\",\"apiVersion\":\"v1\",\"reference\":{\"kind\":\"ReplicaSet\"}}",
			},
		},
		Spec: apiv1.PodSpec{
			Volumes: []apiv1.Volume{
				{
					VolumeSource: apiv1.VolumeSource{
						EmptyDir: &apiv1.EmptyDirVolumeSource{},
					},
				},
			},
		},
	}
	_, err = FastGetPodsToMove(schedulercache.NewNodeInfo(pod6), true, true)
	assert.Error(t, err)

	// Non-local storage
	pod7 := &apiv1.Pod{
		ObjectMeta: apiv1.ObjectMeta{
			Name:      "pod7",
			Namespace: "ns",
			Annotations: map[string]string{
				"kubernetes.io/created-by": "{\"kind\":\"SerializedReference\",\"apiVersion\":\"v1\",\"reference\":{\"kind\":\"ReplicaSet\"}}",
			},
		},
		Spec: apiv1.PodSpec{
			Volumes: []apiv1.Volume{
				{
					VolumeSource: apiv1.VolumeSource{
						GitRepo: &apiv1.GitRepoVolumeSource{
							Repository: "my-repo",
						},
					},
				},
			},
		},
	}
	r7, err := FastGetPodsToMove(schedulercache.NewNodeInfo(pod7), true, true)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(r7))
}
