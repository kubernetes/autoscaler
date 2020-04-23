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

	apiv1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/autoscaler/cluster-autoscaler/utils/drain"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/kubernetes/pkg/kubelet/types"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"

	"github.com/stretchr/testify/assert"
)

func TestFastGetPodsToMove(t *testing.T) {

	// Unreplicated pod
	pod1 := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod1",
			Namespace: "ns",
		},
	}
	_, blockingPod, err := FastGetPodsToMove(schedulerframework.NewNodeInfo(pod1), true, true, nil)
	assert.Error(t, err)
	assert.Equal(t, &drain.BlockingPod{Pod: pod1, Reason: drain.NotReplicated}, blockingPod)

	// Replicated pod
	pod2 := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "pod2",
			Namespace:       "ns",
			OwnerReferences: GenerateOwnerReferences("rs", "ReplicaSet", "extensions/v1beta1", ""),
		},
	}
	r2, blockingPod, err := FastGetPodsToMove(schedulerframework.NewNodeInfo(pod2), true, true, nil)
	assert.NoError(t, err)
	assert.Nil(t, blockingPod)
	assert.Equal(t, 1, len(r2))
	assert.Equal(t, pod2, r2[0])

	// Manifest pod
	pod3 := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod3",
			Namespace: "kube-system",
			Annotations: map[string]string{
				types.ConfigMirrorAnnotationKey: "something",
			},
		},
	}
	r3, blockingPod, err := FastGetPodsToMove(schedulerframework.NewNodeInfo(pod3), true, true, nil)
	assert.NoError(t, err)
	assert.Nil(t, blockingPod)
	assert.Equal(t, 0, len(r3))

	// DaemonSet pod
	pod4 := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "pod4",
			Namespace:       "ns",
			OwnerReferences: GenerateOwnerReferences("ds", "DaemonSet", "extensions/v1beta1", ""),
		},
	}
	r4, blockingPod, err := FastGetPodsToMove(schedulerframework.NewNodeInfo(pod2, pod3, pod4), true, true, nil)
	assert.NoError(t, err)
	assert.Nil(t, blockingPod)
	assert.Equal(t, 1, len(r4))
	assert.Equal(t, pod2, r4[0])

	// Kube-system
	pod5 := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "pod5",
			Namespace:       "kube-system",
			OwnerReferences: GenerateOwnerReferences("rs", "ReplicaSet", "extensions/v1beta1", ""),
		},
	}
	_, blockingPod, err = FastGetPodsToMove(schedulerframework.NewNodeInfo(pod5), true, true, nil)
	assert.Error(t, err)
	assert.Equal(t, &drain.BlockingPod{Pod: pod5, Reason: drain.UnmovableKubeSystemPod}, blockingPod)

	// Local storage
	pod6 := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "pod6",
			Namespace:       "ns",
			OwnerReferences: GenerateOwnerReferences("rs", "ReplicaSet", "extensions/v1beta1", ""),
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
	_, blockingPod, err = FastGetPodsToMove(schedulerframework.NewNodeInfo(pod6), true, true, nil)
	assert.Error(t, err)
	assert.Equal(t, &drain.BlockingPod{Pod: pod6, Reason: drain.LocalStorageRequested}, blockingPod)

	// Non-local storage
	pod7 := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "pod7",
			Namespace:       "ns",
			OwnerReferences: GenerateOwnerReferences("rs", "ReplicaSet", "extensions/v1beta1", ""),
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
	r7, blockingPod, err := FastGetPodsToMove(schedulerframework.NewNodeInfo(pod7), true, true, nil)
	assert.NoError(t, err)
	assert.Nil(t, blockingPod)
	assert.Equal(t, 1, len(r7))

	// Pdb blocking
	pod8 := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "pod8",
			Namespace:       "ns",
			OwnerReferences: GenerateOwnerReferences("rs", "ReplicaSet", "extensions/v1beta1", ""),
			Labels: map[string]string{
				"critical": "true",
			},
		},
		Spec: apiv1.PodSpec{},
	}
	one := intstr.FromInt(1)
	pdb8 := &policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foobar",
			Namespace: "ns",
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			MinAvailable: &one,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"critical": "true",
				},
			},
		},
		Status: policyv1.PodDisruptionBudgetStatus{
			DisruptionsAllowed: 0,
		},
	}

	_, blockingPod, err = FastGetPodsToMove(schedulerframework.NewNodeInfo(pod8), true, true, []*policyv1.PodDisruptionBudget{pdb8})
	assert.Error(t, err)
	assert.Equal(t, &drain.BlockingPod{Pod: pod8, Reason: drain.NotEnoughPdb}, blockingPod)

	// Pdb allowing
	pod9 := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "pod9",
			Namespace:       "ns",
			OwnerReferences: GenerateOwnerReferences("rs", "ReplicaSet", "extensions/v1beta1", ""),
			Labels: map[string]string{
				"critical": "true",
			},
		},
		Spec: apiv1.PodSpec{},
	}
	pdb9 := &policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foobar",
			Namespace: "ns",
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			MinAvailable: &one,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"critical": "true",
				},
			},
		},
		Status: policyv1.PodDisruptionBudgetStatus{
			DisruptionsAllowed: 1,
		},
	}

	r9, blockingPod, err := FastGetPodsToMove(schedulerframework.NewNodeInfo(pod9), true, true, []*policyv1.PodDisruptionBudget{pdb9})
	assert.NoError(t, err)
	assert.Nil(t, blockingPod)
	assert.Equal(t, 1, len(r9))
}
