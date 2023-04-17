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
	"fmt"
	"testing"
	"time"

	apiv1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/drainability"
	"k8s.io/autoscaler/cluster-autoscaler/utils/drain"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/kubernetes/pkg/kubelet/types"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"

	"github.com/stretchr/testify/assert"
)

func TestGetPodsToMove(t *testing.T) {
	testTime := time.Date(2020, time.December, 18, 17, 0, 0, 0, time.UTC)
	unreplicatedPod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "unreplicatedPod",
			Namespace: "ns",
		},
	}
	rsPod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "rsPod",
			Namespace:       "ns",
			OwnerReferences: GenerateOwnerReferences("rs", "ReplicaSet", "extensions/v1beta1", ""),
		},
	}
	manifestPod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "manifestPod",
			Namespace: "kube-system",
			Annotations: map[string]string{
				types.ConfigMirrorAnnotationKey: "something",
			},
		},
	}
	dsPod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "dsPod",
			Namespace:       "ns",
			OwnerReferences: GenerateOwnerReferences("ds", "DaemonSet", "extensions/v1beta1", ""),
		},
	}
	systemPod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "systemPod",
			Namespace:       "kube-system",
			OwnerReferences: GenerateOwnerReferences("rs", "ReplicaSet", "extensions/v1beta1", ""),
		},
	}
	localStoragePod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "localStoragePod",
			Namespace:       "ns",
			OwnerReferences: GenerateOwnerReferences("rs", "ReplicaSet", "extensions/v1beta1", ""),
		},
		Spec: apiv1.PodSpec{
			Volumes: []apiv1.Volume{
				{
					Name: "empty-vol",
					VolumeSource: apiv1.VolumeSource{
						EmptyDir: &apiv1.EmptyDirVolumeSource{},
					},
				},
			},
		},
	}
	nonLocalStoragePod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "nonLocalStoragePod",
			Namespace:       "ns",
			OwnerReferences: GenerateOwnerReferences("rs", "ReplicaSet", "extensions/v1beta1", ""),
		},
		Spec: apiv1.PodSpec{
			Volumes: []apiv1.Volume{
				{
					Name: "my-repo",
					VolumeSource: apiv1.VolumeSource{
						GitRepo: &apiv1.GitRepoVolumeSource{
							Repository: "my-repo",
						},
					},
				},
			},
		},
	}
	pdbPod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "pdbPod",
			Namespace:       "ns",
			OwnerReferences: GenerateOwnerReferences("rs", "ReplicaSet", "extensions/v1beta1", ""),
			Labels: map[string]string{
				"critical": "true",
			},
		},
		Spec: apiv1.PodSpec{},
	}
	one := intstr.FromInt(1)
	restrictivePdb := &policyv1.PodDisruptionBudget{
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
	permissivePdb := &policyv1.PodDisruptionBudget{
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
	terminatedPod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "terminatedPod",
			Namespace:       "ns",
			OwnerReferences: GenerateOwnerReferences("rs", "ReplicaSet", "extensions/v1beta1", ""),
			DeletionTimestamp: &metav1.Time{
				Time: testTime.Add(-1*drain.PodLongTerminatingExtraThreshold - time.Minute), // more than PodLongTerminatingExtraThreshold
			},
		},
	}
	terminatingPod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "terminatingPod",
			Namespace:       "ns",
			OwnerReferences: GenerateOwnerReferences("rs", "ReplicaSet", "extensions/v1beta1", ""),
			DeletionTimestamp: &metav1.Time{
				Time: testTime.Add(-1*drain.PodLongTerminatingExtraThreshold + time.Minute), // still terminating, below the default TerminatingGracePeriode
			},
		},
	}

	testCases := []struct {
		desc         string
		pods         []*apiv1.Pod
		pdbs         []*policyv1.PodDisruptionBudget
		rules        []drainability.Rule
		wantPods     []*apiv1.Pod
		wantDs       []*apiv1.Pod
		wantBlocking *drain.BlockingPod
		wantErr      bool
	}{
		{
			desc:    "Unreplicated pod",
			pods:    []*apiv1.Pod{unreplicatedPod},
			wantErr: true,
			wantBlocking: &drain.BlockingPod{
				Pod:    unreplicatedPod,
				Reason: drain.NotReplicated,
			},
		},
		{
			desc:     "Replicated pod",
			pods:     []*apiv1.Pod{rsPod},
			wantPods: []*apiv1.Pod{rsPod},
		},
		{
			desc: "Manifest pod",
			pods: []*apiv1.Pod{manifestPod},
		},
		{
			desc:     "DaemonSet pod",
			pods:     []*apiv1.Pod{rsPod, manifestPod, dsPod},
			wantPods: []*apiv1.Pod{rsPod},
			wantDs:   []*apiv1.Pod{dsPod},
		},
		{
			desc:    "Kube-system",
			pods:    []*apiv1.Pod{systemPod},
			wantErr: true,
			wantBlocking: &drain.BlockingPod{
				Pod:    systemPod,
				Reason: drain.UnmovableKubeSystemPod,
			},
		},
		{
			desc:    "Local storage",
			pods:    []*apiv1.Pod{localStoragePod},
			wantErr: true,
			wantBlocking: &drain.BlockingPod{
				Pod:    localStoragePod,
				Reason: drain.LocalStorageRequested,
			},
		},
		{
			desc:     "Non-local storage",
			pods:     []*apiv1.Pod{nonLocalStoragePod},
			wantPods: []*apiv1.Pod{nonLocalStoragePod},
		},
		{
			desc:    "Pdb blocking",
			pods:    []*apiv1.Pod{pdbPod},
			pdbs:    []*policyv1.PodDisruptionBudget{restrictivePdb},
			wantErr: true,
			wantBlocking: &drain.BlockingPod{
				Pod:    pdbPod,
				Reason: drain.NotEnoughPdb,
			},
		},
		{
			desc:     "Pdb allowing",
			pods:     []*apiv1.Pod{pdbPod},
			pdbs:     []*policyv1.PodDisruptionBudget{permissivePdb},
			wantPods: []*apiv1.Pod{pdbPod},
		},
		{
			desc:     "Pod termination",
			pods:     []*apiv1.Pod{rsPod, terminatedPod, terminatingPod},
			wantPods: []*apiv1.Pod{rsPod, terminatingPod},
		},
		{
			desc:     "Rule allows",
			pods:     []*apiv1.Pod{unreplicatedPod},
			rules:    []drainability.Rule{alwaysDrain{}},
			wantPods: []*apiv1.Pod{unreplicatedPod},
		},
		{
			desc:     "Second rule allows",
			pods:     []*apiv1.Pod{unreplicatedPod},
			rules:    []drainability.Rule{cantDecide{}, alwaysDrain{}},
			wantPods: []*apiv1.Pod{unreplicatedPod},
		},
		{
			desc:    "Rule blocks",
			pods:    []*apiv1.Pod{rsPod},
			rules:   []drainability.Rule{neverDrain{}},
			wantErr: true,
			wantBlocking: &drain.BlockingPod{
				Pod:    rsPod,
				Reason: drain.UnexpectedError,
			},
		},
		{
			desc:    "Second rule blocks",
			pods:    []*apiv1.Pod{rsPod},
			rules:   []drainability.Rule{cantDecide{}, neverDrain{}},
			wantErr: true,
			wantBlocking: &drain.BlockingPod{
				Pod:    rsPod,
				Reason: drain.UnexpectedError,
			},
		},
		{
			desc:    "Undecisive rule fallback to default logic: Unreplicated pod",
			pods:    []*apiv1.Pod{unreplicatedPod},
			rules:   []drainability.Rule{cantDecide{}},
			wantErr: true,
			wantBlocking: &drain.BlockingPod{
				Pod:    unreplicatedPod,
				Reason: drain.NotReplicated,
			},
		},
		{
			desc:     "Undecisive rule fallback to default logic: Replicated pod",
			pods:     []*apiv1.Pod{rsPod},
			rules:    []drainability.Rule{cantDecide{}},
			wantPods: []*apiv1.Pod{rsPod},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			deleteOptions := NodeDeleteOptions{
				SkipNodesWithSystemPods:           true,
				SkipNodesWithLocalStorage:         true,
				MinReplicaCount:                   0,
				SkipNodesWithCustomControllerPods: true,
				DrainabilityRules:                 tc.rules,
			}
			p, d, b, err := GetPodsToMove(schedulerframework.NewNodeInfo(tc.pods...), deleteOptions, nil, tc.pdbs, testTime)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.ElementsMatch(t, tc.wantPods, p)
			assert.ElementsMatch(t, tc.wantDs, d)
			assert.Equal(t, tc.wantBlocking, b)
		})
	}
}

type alwaysDrain struct{}

func (a alwaysDrain) Drainable(*apiv1.Pod) drainability.Status {
	return drainability.NewDrainableStatus()
}

type neverDrain struct{}

func (n neverDrain) Drainable(*apiv1.Pod) drainability.Status {
	return drainability.NewBlockedStatus(drain.UnexpectedError, fmt.Errorf("nope"))
}

type cantDecide struct{}

func (c cantDecide) Drainable(*apiv1.Pod) drainability.Status {
	return drainability.NewUnmatchedStatus()
}
