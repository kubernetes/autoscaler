/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

	kube_api "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/kubelet/types"
	"k8s.io/kubernetes/plugin/pkg/scheduler/schedulercache"

	"github.com/stretchr/testify/assert"
)

func TestFastGetPodsToMove(t *testing.T) {

	// Unreplicated pod
	pod1 := &kube_api.Pod{
		ObjectMeta: kube_api.ObjectMeta{
			Name:      "pod1",
			Namespace: "ns",
		},
	}
	_, err := FastGetPodsToMove(schedulercache.NewNodeInfo(pod1), false, true, true)
	assert.Error(t, err)

	// Replicated pod
	pod2 := &kube_api.Pod{
		ObjectMeta: kube_api.ObjectMeta{
			Name:      "pod2",
			Namespace: "ns",
			Annotations: map[string]string{
				"kubernetes.io/created-by": "{\"kind\":\"SerializedReference\",\"apiVersion\":\"v1\",\"reference\":{\"kind\":\"ReplicaSet\"}}",
			},
		},
	}
	r2, err := FastGetPodsToMove(schedulercache.NewNodeInfo(pod2), false, true, true)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(r2))
	assert.Equal(t, pod2, r2[0])

	// Manifest pod
	pod3 := &kube_api.Pod{
		ObjectMeta: kube_api.ObjectMeta{
			Name:      "pod3",
			Namespace: "kube-system",
			Annotations: map[string]string{
				types.ConfigMirrorAnnotationKey: "something",
			},
		},
	}
	r3, err := FastGetPodsToMove(schedulercache.NewNodeInfo(pod3), false, true, true)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(r3))

	// DeamonSet pod
	pod4 := &kube_api.Pod{
		ObjectMeta: kube_api.ObjectMeta{
			Name:      "pod4",
			Namespace: "ns",
			Annotations: map[string]string{
				"kubernetes.io/created-by": "{\"kind\":\"SerializedReference\",\"apiVersion\":\"v1\",\"reference\":{\"kind\":\"DaemonSet\"}}",
			},
		},
	}
	r4, err := FastGetPodsToMove(schedulercache.NewNodeInfo(pod2, pod3, pod4), false, true, true)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(r4))
	assert.Equal(t, pod2, r4[0])

	// Kube-system
	pod5 := &kube_api.Pod{
		ObjectMeta: kube_api.ObjectMeta{
			Name:      "pod5",
			Namespace: "kube-system",
			Annotations: map[string]string{
				"kubernetes.io/created-by": "{\"kind\":\"SerializedReference\",\"apiVersion\":\"v1\",\"reference\":{\"kind\":\"ReplicaSet\"}}",
			},
		},
	}
	_, err = FastGetPodsToMove(schedulercache.NewNodeInfo(pod5), false, true, true)
	assert.Error(t, err)

	// Local storage
	pod6 := &kube_api.Pod{
		ObjectMeta: kube_api.ObjectMeta{
			Name:      "pod6",
			Namespace: "ns",
			Annotations: map[string]string{
				"kubernetes.io/created-by": "{\"kind\":\"SerializedReference\",\"apiVersion\":\"v1\",\"reference\":{\"kind\":\"ReplicaSet\"}}",
			},
		},
		Spec: kube_api.PodSpec{
			Volumes: []kube_api.Volume{
				{
					VolumeSource: kube_api.VolumeSource{
						EmptyDir: &kube_api.EmptyDirVolumeSource{},
					},
				},
			},
		},
	}
	_, err = FastGetPodsToMove(schedulercache.NewNodeInfo(pod6), false, true, true)
	assert.Error(t, err)

	// Non-local storage
	pod7 := &kube_api.Pod{
		ObjectMeta: kube_api.ObjectMeta{
			Name:      "pod7",
			Namespace: "ns",
			Annotations: map[string]string{
				"kubernetes.io/created-by": "{\"kind\":\"SerializedReference\",\"apiVersion\":\"v1\",\"reference\":{\"kind\":\"ReplicaSet\"}}",
			},
		},
		Spec: kube_api.PodSpec{
			Volumes: []kube_api.Volume{
				{
					VolumeSource: kube_api.VolumeSource{
						GitRepo: &kube_api.GitRepoVolumeSource{
							Repository: "my-repo",
						},
					},
				},
			},
		},
	}
	r7, err := FastGetPodsToMove(schedulercache.NewNodeInfo(pod7), false, true, true)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(r7))
}
