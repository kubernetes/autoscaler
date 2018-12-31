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

package drain

import (
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	v1appslister "k8s.io/client-go/listers/apps/v1"
	v1lister "k8s.io/client-go/listers/core/v1"
	"k8s.io/kubernetes/pkg/api/testapi"

	"github.com/stretchr/testify/assert"
)

func TestDrain(t *testing.T) {
	replicas := int32(5)

	rc := apiv1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rc",
			Namespace: "default",
			SelfLink:  testapi.Default.SelfLink("replicationcontrollers", "rc"),
		},
		Spec: apiv1.ReplicationControllerSpec{
			Replicas: &replicas,
		},
	}

	rcPod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "bar",
			Namespace:       "default",
			OwnerReferences: GenerateOwnerReferences(rc.Name, "ReplicationController", "core/v1", ""),
		},
		Spec: apiv1.PodSpec{
			NodeName: "node",
		},
	}

	kubeSystemRc := apiv1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rc",
			Namespace: "kube-system",
			SelfLink:  testapi.Default.SelfLink("replicationcontrollers", "rc"),
		},
		Spec: apiv1.ReplicationControllerSpec{
			Replicas: &replicas,
		},
	}

	kubeSystemRcPod := &apiv1.Pod{
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

	ds := appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ds",
			Namespace: "default",
			SelfLink:  "/apiv1s/apps/v1/namespaces/default/daemonsets/ds",
		},
	}

	dsPod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "bar",
			Namespace:       "default",
			OwnerReferences: GenerateOwnerReferences(ds.Name, "DaemonSet", "apps/v1", ""),
		},
		Spec: apiv1.PodSpec{
			NodeName: "node",
		},
	}

	job := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "job",
			Namespace: "default",
			SelfLink:  "/apiv1s/batch/v1/namespaces/default/jobs/job",
		},
	}

	jobPod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "bar",
			Namespace:       "default",
			OwnerReferences: GenerateOwnerReferences(job.Name, "Job", "batch/v1", ""),
		},
	}

	statefulset := appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ss",
			Namespace: "default",
			SelfLink:  "/apiv1s/apps/v1/namespaces/default/statefulsets/ss",
		},
	}

	ssPod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "bar",
			Namespace:       "default",
			OwnerReferences: GenerateOwnerReferences(statefulset.Name, "StatefulSet", "apps/v1", ""),
		},
	}

	rs := appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rs",
			Namespace: "default",
			SelfLink:  testapi.Default.SelfLink("replicasets", "rs"),
		},
		Spec: appsv1.ReplicaSetSpec{
			Replicas: &replicas,
		},
	}

	rsPod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "bar",
			Namespace:       "default",
			OwnerReferences: GenerateOwnerReferences(rs.Name, "ReplicaSet", "apps/v1", ""),
		},
		Spec: apiv1.PodSpec{
			NodeName: "node",
		},
	}

	rsPodDeleted := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "bar",
			Namespace:         "default",
			OwnerReferences:   GenerateOwnerReferences(rs.Name, "ReplicaSet", "apps/v1", ""),
			DeletionTimestamp: &metav1.Time{Time: time.Now().Add(-time.Hour)},
		},
		Spec: apiv1.PodSpec{
			NodeName: "node",
		},
	}

	nakedPod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bar",
			Namespace: "default",
		},
		Spec: apiv1.PodSpec{
			NodeName: "node",
		},
	}

	emptydirPod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bar",
			Namespace: "default",
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

	terminalPod := &apiv1.Pod{
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

	failedPod := &apiv1.Pod{
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

	evictedPod := &apiv1.Pod{
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

	safePod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bar",
			Namespace: "default",
			Annotations: map[string]string{
				PodSafeToEvictKey: "true",
			},
		},
		Spec: apiv1.PodSpec{
			NodeName: "node",
		},
	}

	unsafeRcPod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "bar",
			Namespace:       "default",
			OwnerReferences: GenerateOwnerReferences(rc.Name, "ReplicationController", "core/v1", ""),
			Annotations: map[string]string{
				PodSafeToEvictKey: "false",
			},
		},
		Spec: apiv1.PodSpec{
			NodeName: "node",
		},
	}

	unsafeJobPod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "bar",
			Namespace:       "default",
			OwnerReferences: GenerateOwnerReferences(job.Name, "Job", "batch/v1", ""),
			Annotations: map[string]string{
				PodSafeToEvictKey: "false",
			},
		},
	}

	kubeSystemSafePod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bar",
			Namespace: "kube-system",
			Annotations: map[string]string{
				PodSafeToEvictKey: "true",
			},
		},
		Spec: apiv1.PodSpec{
			NodeName: "node",
		},
	}

	emptydirSafePod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bar",
			Namespace: "default",
			Annotations: map[string]string{
				PodSafeToEvictKey: "true",
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

	emptyPDB := &policyv1.PodDisruptionBudget{}

	kubeSystemPDB := &policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "kube-system",
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"k8s-app": "bar",
				},
			},
		},
	}

	kubeSystemFakePDB := &policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "kube-system",
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"k8s-app": "foo",
				},
			},
		},
	}

	defaultNamespacePDB := &policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"k8s-app": "PDB-managed pod",
				},
			},
		},
	}

	tests := []struct {
		description string
		pods        []*apiv1.Pod
		pdbs        []*policyv1.PodDisruptionBudget
		rcs         []*apiv1.ReplicationController
		replicaSets []*appsv1.ReplicaSet
		expectFatal bool
		expectPods  []*apiv1.Pod
	}{
		{
			description: "RC-managed pod",
			pods:        []*apiv1.Pod{rcPod},
			pdbs:        []*policyv1.PodDisruptionBudget{},
			rcs:         []*apiv1.ReplicationController{&rc},
			expectFatal: false,
			expectPods:  []*apiv1.Pod{rcPod},
		},
		{
			description: "DS-managed pod",
			pods:        []*apiv1.Pod{dsPod},
			pdbs:        []*policyv1.PodDisruptionBudget{},
			expectFatal: false,
			expectPods:  []*apiv1.Pod{},
		},
		{
			description: "Job-managed pod",
			pods:        []*apiv1.Pod{jobPod},
			pdbs:        []*policyv1.PodDisruptionBudget{},
			rcs:         []*apiv1.ReplicationController{&rc},
			expectFatal: false,
			expectPods:  []*apiv1.Pod{jobPod},
		},
		{
			description: "SS-managed pod",
			pods:        []*apiv1.Pod{ssPod},
			pdbs:        []*policyv1.PodDisruptionBudget{},
			rcs:         []*apiv1.ReplicationController{&rc},
			expectFatal: false,
			expectPods:  []*apiv1.Pod{ssPod},
		},
		{
			description: "RS-managed pod",
			pods:        []*apiv1.Pod{rsPod},
			pdbs:        []*policyv1.PodDisruptionBudget{},
			replicaSets: []*appsv1.ReplicaSet{&rs},
			expectFatal: false,
			expectPods:  []*apiv1.Pod{rsPod},
		},
		{
			description: "RS-managed pod that is being deleted",
			pods:        []*apiv1.Pod{rsPodDeleted},
			pdbs:        []*policyv1.PodDisruptionBudget{},
			replicaSets: []*appsv1.ReplicaSet{&rs},
			expectFatal: false,
			expectPods:  []*apiv1.Pod{},
		},
		{
			description: "naked pod",
			pods:        []*apiv1.Pod{nakedPod},
			pdbs:        []*policyv1.PodDisruptionBudget{},
			expectFatal: true,
			expectPods:  []*apiv1.Pod{},
		},
		{
			description: "pod with EmptyDir",
			pods:        []*apiv1.Pod{emptydirPod},
			pdbs:        []*policyv1.PodDisruptionBudget{},
			expectFatal: true,
			expectPods:  []*apiv1.Pod{},
		},
		{
			description: "failed pod",
			pods:        []*apiv1.Pod{failedPod},
			pdbs:        []*policyv1.PodDisruptionBudget{},
			expectFatal: false,
			expectPods:  []*apiv1.Pod{failedPod},
		},
		{
			description: "evicted pod",
			pods:        []*apiv1.Pod{evictedPod},
			pdbs:        []*policyv1.PodDisruptionBudget{},
			expectFatal: false,
			expectPods:  []*apiv1.Pod{evictedPod},
		},
		{
			description: "pod in terminal state",
			pods:        []*apiv1.Pod{terminalPod},
			pdbs:        []*policyv1.PodDisruptionBudget{},
			expectFatal: false,
			expectPods:  []*apiv1.Pod{terminalPod},
		},
		{
			description: "pod with PodSafeToEvict annotation",
			pods:        []*apiv1.Pod{safePod},
			pdbs:        []*policyv1.PodDisruptionBudget{},
			expectFatal: false,
			expectPods:  []*apiv1.Pod{safePod},
		},
		{
			description: "kube-system pod with PodSafeToEvict annotation",
			pods:        []*apiv1.Pod{kubeSystemSafePod},
			pdbs:        []*policyv1.PodDisruptionBudget{},
			expectFatal: false,
			expectPods:  []*apiv1.Pod{kubeSystemSafePod},
		},
		{
			description: "pod with EmptyDir and PodSafeToEvict annotation",
			pods:        []*apiv1.Pod{emptydirSafePod},
			pdbs:        []*policyv1.PodDisruptionBudget{},
			expectFatal: false,
			expectPods:  []*apiv1.Pod{emptydirSafePod},
		},
		{
			description: "RC-managed pod with PodSafeToEvict=false annotation",
			pods:        []*apiv1.Pod{unsafeRcPod},
			rcs:         []*apiv1.ReplicationController{&rc},
			pdbs:        []*policyv1.PodDisruptionBudget{},
			expectFatal: true,
			expectPods:  []*apiv1.Pod{},
		},
		{
			description: "Job-managed pod with PodSafeToEvict=false annotation",
			pods:        []*apiv1.Pod{unsafeJobPod},
			pdbs:        []*policyv1.PodDisruptionBudget{},
			rcs:         []*apiv1.ReplicationController{&rc},
			expectFatal: true,
			expectPods:  []*apiv1.Pod{},
		},
		{
			description: "empty PDB with RC-managed pod",
			pods:        []*apiv1.Pod{rcPod},
			pdbs:        []*policyv1.PodDisruptionBudget{emptyPDB},
			rcs:         []*apiv1.ReplicationController{&rc},
			expectFatal: false,
			expectPods:  []*apiv1.Pod{rcPod},
		},
		{
			description: "kube-system PDB with matching kube-system pod",
			pods:        []*apiv1.Pod{kubeSystemRcPod},
			pdbs:        []*policyv1.PodDisruptionBudget{kubeSystemPDB},
			rcs:         []*apiv1.ReplicationController{&kubeSystemRc},
			expectFatal: false,
			expectPods:  []*apiv1.Pod{kubeSystemRcPod},
		},
		{
			description: "kube-system PDB with non-matching kube-system pod",
			pods:        []*apiv1.Pod{kubeSystemRcPod},
			pdbs:        []*policyv1.PodDisruptionBudget{kubeSystemFakePDB},
			rcs:         []*apiv1.ReplicationController{&kubeSystemRc},
			expectFatal: true,
			expectPods:  []*apiv1.Pod{},
		},
		{
			description: "kube-system PDB with default namespace pod",
			pods:        []*apiv1.Pod{rcPod},
			pdbs:        []*policyv1.PodDisruptionBudget{kubeSystemPDB},
			rcs:         []*apiv1.ReplicationController{&rc},
			expectFatal: false,
			expectPods:  []*apiv1.Pod{rcPod},
		},
		{
			description: "default namespace PDB with matching labels kube-system pod",
			pods:        []*apiv1.Pod{kubeSystemRcPod},
			pdbs:        []*policyv1.PodDisruptionBudget{defaultNamespacePDB},
			rcs:         []*apiv1.ReplicationController{&kubeSystemRc},
			expectFatal: true,
			expectPods:  []*apiv1.Pod{},
		},
	}

	for _, test := range tests {
		var err error
		var rcLister v1lister.ReplicationControllerLister
		if len(test.rcs) > 0 {
			rcLister, err = kube_util.NewTestReplicationControllerLister(test.rcs)
			assert.NoError(t, err)
		}
		var rsLister v1appslister.ReplicaSetLister
		if len(test.replicaSets) > 0 {
			rsLister, err = kube_util.NewTestReplicaSetLister(test.replicaSets)
			assert.NoError(t, err)
		}

		dsLister, err := kube_util.NewTestDaemonSetLister([]*appsv1.DaemonSet{&ds})
		assert.NoError(t, err)
		jobLister, err := kube_util.NewTestJobLister([]*batchv1.Job{&job})
		assert.NoError(t, err)
		ssLister, err := kube_util.NewTestStatefulSetLister([]*appsv1.StatefulSet{&statefulset})
		assert.NoError(t, err)

		registry := kube_util.NewListerRegistry(nil, nil, nil, nil, nil, dsLister, rcLister, jobLister, rsLister, ssLister)

		pods, err := GetPodsForDeletionOnNodeDrain(test.pods, test.pdbs,
			false, true, true, true, registry, 0, time.Now())

		if test.expectFatal {
			if err == nil {
				t.Fatalf("%s: unexpected non-error", test.description)
			}
		}

		if !test.expectFatal {
			if err != nil {
				t.Fatalf("%s: error occurred: %v", test.description, err)
			}
		}

		if len(pods) != len(test.expectPods) {
			t.Fatalf("Wrong pod list content: %v", test.description)
		}
	}
}
