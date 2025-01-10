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

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/pdb"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/drainability"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/drainability/rules"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/options"
	"k8s.io/autoscaler/cluster-autoscaler/utils/drain"
	"k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/kubernetes/pkg/kubelet/types"

	"github.com/stretchr/testify/assert"
)

func TestGetPodsToMove(t *testing.T) {
	var (
		testTime = time.Date(2020, time.December, 18, 17, 0, 0, 0, time.UTC)
		replicas = int32(5)

		unreplicatedPod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "unreplicatedPod",
				Namespace: "ns",
			},
		}
		manifestPod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "manifestPod",
				Namespace: "kube-system",
				Annotations: map[string]string{
					types.ConfigMirrorAnnotationKey: "something",
				},
			},
		}
		systemPod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "systemPod",
				Namespace:       "kube-system",
				OwnerReferences: GenerateOwnerReferences("rs", "ReplicaSet", "extensions/v1beta1", ""),
			},
		}
		localStoragePod = &apiv1.Pod{
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
		nonLocalStoragePod = &apiv1.Pod{
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
		pdbPod = &apiv1.Pod{
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
		one            = intstr.FromInt(1)
		restrictivePdb = &policyv1.PodDisruptionBudget{
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
		permissivePdb = &policyv1.PodDisruptionBudget{
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
		terminatedPod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "terminatedPod",
				Namespace:       "ns",
				OwnerReferences: GenerateOwnerReferences("rs", "ReplicaSet", "extensions/v1beta1", ""),
				DeletionTimestamp: &metav1.Time{
					Time: testTime.Add(-1*drain.PodLongTerminatingExtraThreshold - time.Minute), // more than PodLongTerminatingExtraThreshold
				},
			},
		}
		terminatingPod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "terminatingPod",
				Namespace:       "ns",
				OwnerReferences: GenerateOwnerReferences("rs", "ReplicaSet", "extensions/v1beta1", ""),
				DeletionTimestamp: &metav1.Time{
					Time: testTime.Add(-1*drain.PodLongTerminatingExtraThreshold + time.Minute), // still terminating, below the default TerminatingGracePeriod
				},
			},
		}

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
		kubeSystemRc = apiv1.ReplicationController{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "rc",
				Namespace: "kube-system",
				SelfLink:  "api/v1/namespaces/kube-system/replicationcontrollers/rc",
			},
			Spec: apiv1.ReplicationControllerSpec{
				Replicas: &replicas,
			},
		}
		kubeSystemRcPod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "bar",
				Namespace:       "kube-system",
				OwnerReferences: GenerateOwnerReferences(kubeSystemRc.Name, "ReplicationController", "core/v1", ""),
				Labels: map[string]string{
					"k8s-app": "bar",
				},
			},
			Spec: apiv1.PodSpec{
				NodeName: "node",
			},
		}
		ds = appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ds",
				Namespace: "default",
				SelfLink:  "/apiv1s/apps/v1/namespaces/default/daemonsets/ds",
			},
		}
		dsPod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "bar",
				Namespace:       "default",
				OwnerReferences: GenerateOwnerReferences(ds.Name, "DaemonSet", "apps/v1", ""),
			},
			Spec: apiv1.PodSpec{
				NodeName: "node",
			},
		}
		cdsPod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "bar",
				Namespace:       "default",
				OwnerReferences: GenerateOwnerReferences(ds.Name, "CustomDaemonSet", "crd/v1", ""),
				Annotations: map[string]string{
					"cluster-autoscaler.kubernetes.io/daemonset-pod": "true",
				},
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
		statefulset = appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ss",
				Namespace: "default",
				SelfLink:  "/apiv1s/apps/v1/namespaces/default/statefulsets/ss",
			},
		}
		ssPod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "bar",
				Namespace:       "default",
				OwnerReferences: GenerateOwnerReferences(statefulset.Name, "StatefulSet", "apps/v1", ""),
			},
		}
		rs = appsv1.ReplicaSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "rs",
				Namespace: "default",
				SelfLink:  "api/v1/namespaces/default/replicasets/rs",
			},
			Spec: appsv1.ReplicaSetSpec{
				Replicas: &replicas,
			},
		}
		rsPod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "bar",
				Namespace:       "default",
				OwnerReferences: GenerateOwnerReferences(rs.Name, "ReplicaSet", "apps/v1", ""),
			},
			Spec: apiv1.PodSpec{
				NodeName: "node",
			},
		}
		rsPodDeleted = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "bar",
				Namespace:         "default",
				OwnerReferences:   GenerateOwnerReferences(rs.Name, "ReplicaSet", "apps/v1", ""),
				DeletionTimestamp: &metav1.Time{Time: testTime.Add(-time.Hour)},
			},
			Spec: apiv1.PodSpec{
				NodeName: "node",
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
		terminalPod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "bar",
				Namespace: "default",
			},
			Spec: apiv1.PodSpec{
				NodeName:      "node",
				RestartPolicy: apiv1.RestartPolicyOnFailure,
			},
			Status: apiv1.PodStatus{
				Phase: apiv1.PodSucceeded,
			},
		}
		zeroGracePeriod    = int64(0)
		longTerminatingPod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "bar",
				Namespace:         "default",
				DeletionTimestamp: &metav1.Time{Time: testTime.Add(-2 * drain.PodLongTerminatingExtraThreshold)},
				OwnerReferences:   GenerateOwnerReferences(rc.Name, "ReplicationController", "core/v1", ""),
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
		extendedGracePeriod                       = int64(6 * 60) // 6 minutes
		longTerminatingPodWithExtendedGracePeriod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "bar",
				Namespace:         "default",
				DeletionTimestamp: &metav1.Time{Time: testTime.Add(-time.Duration(extendedGracePeriod/2) * time.Second)},
				OwnerReferences:   GenerateOwnerReferences(rc.Name, "ReplicationController", "core/v1", ""),
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
		failedPod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "bar",
				Namespace: "default",
			},
			Spec: apiv1.PodSpec{
				NodeName:      "node",
				RestartPolicy: apiv1.RestartPolicyNever,
			},
			Status: apiv1.PodStatus{
				Phase: apiv1.PodFailed,
			},
		}
		evictedPod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "bar",
				Namespace: "default",
			},
			Spec: apiv1.PodSpec{
				NodeName:      "node",
				RestartPolicy: apiv1.RestartPolicyAlways,
			},
			Status: apiv1.PodStatus{
				Phase: apiv1.PodFailed,
			},
		}
		safePod = &apiv1.Pod{
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
		kubeSystemSafePod = &apiv1.Pod{
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
		emptydirSafePod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "bar",
				Namespace: "default",
				Annotations: map[string]string{
					drain.PodSafeToEvictKey: "true",
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
		emptyPDB      = &policyv1.PodDisruptionBudget{}
		kubeSystemPDB = &policyv1.PodDisruptionBudget{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "kube-system",
			},
			Spec: policyv1.PodDisruptionBudgetSpec{
				MinAvailable: &one,
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"k8s-app": "bar",
					},
				},
			},
			Status: policyv1.PodDisruptionBudgetStatus{
				DisruptionsAllowed: 1,
			},
		}
		kubeSystemFakePDB = &policyv1.PodDisruptionBudget{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "kube-system",
			},
			Spec: policyv1.PodDisruptionBudgetSpec{
				MinAvailable: &one,
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"k8s-app": "foo",
					},
				},
			},
			Status: policyv1.PodDisruptionBudgetStatus{
				DisruptionsAllowed: 1,
			},
		}
		defaultNamespacePDB = &policyv1.PodDisruptionBudget{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
			},
			Spec: policyv1.PodDisruptionBudgetSpec{
				MinAvailable: &one,
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"k8s-app": "PDB-managed pod",
					},
				},
			},
			Status: policyv1.PodDisruptionBudgetStatus{
				DisruptionsAllowed: 1,
			},
		}
	)

	testCases := []struct {
		desc         string
		pods         []*apiv1.Pod
		pdbs         []*policyv1.PodDisruptionBudget
		rcs          []*apiv1.ReplicationController
		replicaSets  []*appsv1.ReplicaSet
		rules        rules.Rules
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
			rules:    []rules.Rule{alwaysDrain{}},
			wantPods: []*apiv1.Pod{unreplicatedPod},
		},
		{
			desc:     "Second rule allows",
			pods:     []*apiv1.Pod{unreplicatedPod},
			rules:    []rules.Rule{cantDecide{}, alwaysDrain{}},
			wantPods: []*apiv1.Pod{unreplicatedPod},
		},
		{
			desc:    "Rule blocks",
			pods:    []*apiv1.Pod{rsPod},
			rules:   []rules.Rule{neverDrain{}},
			wantErr: true,
			wantBlocking: &drain.BlockingPod{
				Pod:    rsPod,
				Reason: drain.UnexpectedError,
			},
		},
		{
			desc:    "Second rule blocks",
			pods:    []*apiv1.Pod{rsPod},
			rules:   []rules.Rule{cantDecide{}, neverDrain{}},
			wantErr: true,
			wantBlocking: &drain.BlockingPod{
				Pod:    rsPod,
				Reason: drain.UnexpectedError,
			},
		},
		{
			desc:    "Undecisive rule fallback to default logic: Unreplicated pod",
			pods:    []*apiv1.Pod{unreplicatedPod},
			rules:   []rules.Rule{cantDecide{}},
			wantErr: true,
			wantBlocking: &drain.BlockingPod{
				Pod:    unreplicatedPod,
				Reason: drain.NotReplicated,
			},
		},
		{
			desc:     "Undecisive rule fallback to default logic: Replicated pod",
			pods:     []*apiv1.Pod{rsPod},
			rules:    []rules.Rule{cantDecide{}},
			wantPods: []*apiv1.Pod{rsPod},
		},

		{
			desc:     "RC-managed pod",
			pods:     []*apiv1.Pod{rcPod},
			rcs:      []*apiv1.ReplicationController{&rc},
			wantPods: []*apiv1.Pod{rcPod},
		},
		{
			desc:   "DS-managed pod",
			pods:   []*apiv1.Pod{dsPod},
			wantDs: []*apiv1.Pod{dsPod},
		},
		{
			desc:   "DS-managed pod by a custom Daemonset",
			pods:   []*apiv1.Pod{cdsPod},
			wantDs: []*apiv1.Pod{cdsPod},
		},
		{
			desc:     "Job-managed pod",
			pods:     []*apiv1.Pod{jobPod},
			rcs:      []*apiv1.ReplicationController{&rc},
			wantPods: []*apiv1.Pod{jobPod},
		},
		{
			desc:     "SS-managed pod",
			pods:     []*apiv1.Pod{ssPod},
			rcs:      []*apiv1.ReplicationController{&rc},
			wantPods: []*apiv1.Pod{ssPod},
		},
		{
			desc:        "RS-managed pod",
			pods:        []*apiv1.Pod{rsPod},
			replicaSets: []*appsv1.ReplicaSet{&rs},
			wantPods:    []*apiv1.Pod{rsPod},
		},
		{
			desc:        "RS-managed pod that is being deleted",
			pods:        []*apiv1.Pod{rsPodDeleted},
			replicaSets: []*appsv1.ReplicaSet{&rs},
		},
		{
			desc:     "pod with EmptyDir and SafeToEvictLocalVolumesKey annotation with matching values",
			pods:     []*apiv1.Pod{emptyDirSafeToEvictLocalVolumeMultiValAllMatching},
			rcs:      []*apiv1.ReplicationController{&rc},
			wantPods: []*apiv1.Pod{emptyDirSafeToEvictLocalVolumeMultiValAllMatching},
		},
		{
			desc:     "failed pod",
			pods:     []*apiv1.Pod{failedPod},
			wantPods: []*apiv1.Pod{failedPod},
		},
		{
			desc: "long terminating pod with 0 grace period",
			pods: []*apiv1.Pod{longTerminatingPod},
			rcs:  []*apiv1.ReplicationController{&rc},
		},
		{
			desc:     "long terminating pod with extended grace period",
			pods:     []*apiv1.Pod{longTerminatingPodWithExtendedGracePeriod},
			rcs:      []*apiv1.ReplicationController{&rc},
			wantPods: []*apiv1.Pod{longTerminatingPodWithExtendedGracePeriod},
		},
		{
			desc:     "evicted pod",
			pods:     []*apiv1.Pod{evictedPod},
			wantPods: []*apiv1.Pod{evictedPod},
		},
		{
			desc:     "pod in terminal state",
			pods:     []*apiv1.Pod{terminalPod},
			wantPods: []*apiv1.Pod{terminalPod},
		},
		{
			desc:     "pod with PodSafeToEvict annotation",
			pods:     []*apiv1.Pod{safePod},
			wantPods: []*apiv1.Pod{safePod},
		},
		{
			desc:     "kube-system pod with PodSafeToEvict annotation",
			pods:     []*apiv1.Pod{kubeSystemSafePod},
			wantPods: []*apiv1.Pod{kubeSystemSafePod},
		},
		{
			desc:     "pod with EmptyDir and PodSafeToEvict annotation",
			pods:     []*apiv1.Pod{emptydirSafePod},
			wantPods: []*apiv1.Pod{emptydirSafePod},
		},
		{
			desc:     "empty PDB with RC-managed pod",
			pods:     []*apiv1.Pod{rcPod},
			pdbs:     []*policyv1.PodDisruptionBudget{emptyPDB},
			rcs:      []*apiv1.ReplicationController{&rc},
			wantPods: []*apiv1.Pod{rcPod},
		},
		{
			desc:     "kube-system PDB with matching kube-system pod",
			pods:     []*apiv1.Pod{kubeSystemRcPod},
			pdbs:     []*policyv1.PodDisruptionBudget{kubeSystemPDB},
			rcs:      []*apiv1.ReplicationController{&kubeSystemRc},
			wantPods: []*apiv1.Pod{kubeSystemRcPod},
		},
		{
			desc:         "kube-system PDB with non-matching kube-system pod",
			pods:         []*apiv1.Pod{kubeSystemRcPod},
			pdbs:         []*policyv1.PodDisruptionBudget{kubeSystemFakePDB},
			rcs:          []*apiv1.ReplicationController{&kubeSystemRc},
			wantErr:      true,
			wantBlocking: &drain.BlockingPod{Pod: kubeSystemRcPod, Reason: drain.UnmovableKubeSystemPod},
		},
		{
			desc:     "kube-system PDB with default namespace pod",
			pods:     []*apiv1.Pod{rcPod},
			pdbs:     []*policyv1.PodDisruptionBudget{kubeSystemPDB},
			rcs:      []*apiv1.ReplicationController{&rc},
			wantPods: []*apiv1.Pod{rcPod},
		},
		{
			desc:         "default namespace PDB with matching labels kube-system pod",
			pods:         []*apiv1.Pod{kubeSystemRcPod},
			pdbs:         []*policyv1.PodDisruptionBudget{defaultNamespacePDB},
			rcs:          []*apiv1.ReplicationController{&kubeSystemRc},
			wantErr:      true,
			wantBlocking: &drain.BlockingPod{Pod: kubeSystemRcPod, Reason: drain.UnmovableKubeSystemPod},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			var registry kubernetes.ListerRegistry
			if tc.rcs != nil || tc.replicaSets != nil {
				rcLister, err := kube_util.NewTestReplicationControllerLister(tc.rcs)
				assert.NoError(t, err)
				rsLister, err := kube_util.NewTestReplicaSetLister(tc.replicaSets)
				assert.NoError(t, err)
				dsLister, err := kube_util.NewTestDaemonSetLister([]*appsv1.DaemonSet{&ds})
				assert.NoError(t, err)
				jobLister, err := kube_util.NewTestJobLister([]*batchv1.Job{&job})
				assert.NoError(t, err)
				ssLister, err := kube_util.NewTestStatefulSetLister([]*appsv1.StatefulSet{&statefulset})
				assert.NoError(t, err)

				registry = kube_util.NewListerRegistry(nil, nil, nil, nil, dsLister, rcLister, jobLister, rsLister, ssLister)
			}

			deleteOptions := options.NodeDeleteOptions{
				SkipNodesWithSystemPods:           true,
				SkipNodesWithLocalStorage:         true,
				SkipNodesWithCustomControllerPods: true,
			}
			rules := append(tc.rules, rules.Default(deleteOptions)...)
			tracker := pdb.NewBasicRemainingPdbTracker()
			tracker.SetPdbs(tc.pdbs)
			ni := framework.NewTestNodeInfo(nil, tc.pods...)
			p, d, b, err := GetPodsToMove(ni, deleteOptions, rules, registry, tracker, testTime)
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

func (a alwaysDrain) Name() string {
	return "AlwaysDrain"
}

func (a alwaysDrain) Drainable(*drainability.DrainContext, *apiv1.Pod, *framework.NodeInfo) drainability.Status {
	return drainability.NewDrainableStatus()
}

type neverDrain struct{}

func (n neverDrain) Name() string {
	return "NeverDrain"
}

func (n neverDrain) Drainable(*drainability.DrainContext, *apiv1.Pod, *framework.NodeInfo) drainability.Status {
	return drainability.NewBlockedStatus(drain.UnexpectedError, fmt.Errorf("nope"))
}

type cantDecide struct{}

func (c cantDecide) Name() string {
	return "CantDecide"
}

func (c cantDecide) Drainable(*drainability.DrainContext, *apiv1.Pod, *framework.NodeInfo) drainability.Status {
	return drainability.NewUndefinedStatus()
}
