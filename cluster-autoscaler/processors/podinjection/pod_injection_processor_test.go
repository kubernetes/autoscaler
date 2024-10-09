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

package podinjection

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	podinjectionbackoff "k8s.io/autoscaler/cluster-autoscaler/processors/podinjection/backoff"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

// TODO(DRA): Add DRA-specific test cases.

func TestTargetCountInjectionPodListProcessor(t *testing.T) {
	node := BuildTestNode("node1", 100, 0)

	replicaSet1 := createTestReplicaSet("rep-set-1", "default", 5)
	scheduledPodRep1Copy1 := buildTestPod("default", "-scheduled-pod-rep1-1", withControllerOwnerRef(replicaSet1.Name, "ReplicaSet", replicaSet1.UID), withNodeName(node.Name))
	podRep1Copy1 := buildTestPod("default", "pod-rep1-1", withControllerOwnerRef(replicaSet1.Name, "ReplicaSet", replicaSet1.UID))
	podRep1Copy2 := buildTestPod("default", "pod-rep1-2", withControllerOwnerRef(replicaSet1.Name, "ReplicaSet", replicaSet1.UID))

	job1 := createTestJob("job-1", "default", 10, 10, 0)
	scheduledPodJob1Copy1 := buildTestPod("default", "scheduled-pod-job1-1", withControllerOwnerRef(job1.Name, "Job", job1.UID), withNodeName(node.Name))
	podJob1Copy1 := buildTestPod("default", "pod-job1-1", withControllerOwnerRef(job1.Name, "Job", job1.UID))
	podJob1Copy2 := buildTestPod("default", "pod-job1-2", withControllerOwnerRef(job1.Name, "Job", job1.UID))

	parallelStatefulset := createTestStatefulset("parallel-statefulset-1", "default", appsv1.ParallelPodManagement, 10)
	scheduledParallelStatefulsetPod := buildTestPod("default", "parallel-scheduled-pod-statefulset-1", withControllerOwnerRef(parallelStatefulset.Name, "StatefulSet", parallelStatefulset.UID), withNodeName(node.Name))
	parallelStatefulsetPodCopy1 := buildTestPod("default", "parallel-pod-statefulset1-1", withControllerOwnerRef(parallelStatefulset.Name, "StatefulSet", parallelStatefulset.UID))
	parallelStatefulsetPodCopy2 := buildTestPod("default", "parallel-pod-statefulset1-2", withControllerOwnerRef(parallelStatefulset.Name, "StatefulSet", parallelStatefulset.UID))

	sequentialStatefulset := createTestStatefulset("sequential-statefulset-1", "default", appsv1.OrderedReadyPodManagement, 10)
	scheduledSequentialStatefulsetPod := buildTestPod("default", "sequential-scheduled-pod-statefulset-1", withControllerOwnerRef(sequentialStatefulset.Name, "StatefulSet", sequentialStatefulset.UID), withNodeName(node.Name))
	sequentialStatefulsetPodCopy1 := buildTestPod("default", "sequential-pod-statefulset1-1", withControllerOwnerRef(sequentialStatefulset.Name, "StatefulSet", sequentialStatefulset.UID))
	sequentialStatefulsetPodCopy2 := buildTestPod("default", "sequential-pod-statefulset1-2", withControllerOwnerRef(sequentialStatefulset.Name, "StatefulSet", sequentialStatefulset.UID))

	replicaSetLister, err := kubernetes.NewTestReplicaSetLister([]*appsv1.ReplicaSet{&replicaSet1})
	assert.NoError(t, err)
	jobLister, err := kubernetes.NewTestJobLister([]*batchv1.Job{&job1})
	assert.NoError(t, err)
	statefulsetLister, err := kubernetes.NewTestStatefulSetLister([]*appsv1.StatefulSet{&parallelStatefulset, &sequentialStatefulset})
	assert.NoError(t, err)

	testCases := []struct {
		name             string
		scheduledPods    []*apiv1.Pod
		unschedulabePods []*apiv1.Pod
		wantPods         []*apiv1.Pod
	}{
		{
			name:             "ReplicaSet",
			scheduledPods:    []*apiv1.Pod{scheduledPodRep1Copy1},
			unschedulabePods: []*apiv1.Pod{podRep1Copy1, podRep1Copy2},
			wantPods:         append([]*apiv1.Pod{podRep1Copy1, podRep1Copy2}, makeFakePodsIgnoreClaims(replicaSet1.UID, podRep1Copy1, 2)...),
		},
		{
			name:             "Job",
			scheduledPods:    []*apiv1.Pod{scheduledPodJob1Copy1},
			unschedulabePods: []*apiv1.Pod{podJob1Copy1, podJob1Copy2},
			wantPods:         append([]*apiv1.Pod{podJob1Copy1, podJob1Copy2}, makeFakePodsIgnoreClaims(job1.UID, podJob1Copy1, 7)...),
		},
		{
			name:             "Statefulset - Parallel pod management policy",
			scheduledPods:    []*apiv1.Pod{scheduledParallelStatefulsetPod},
			unschedulabePods: []*apiv1.Pod{parallelStatefulsetPodCopy1, parallelStatefulsetPodCopy2},
			wantPods:         append([]*apiv1.Pod{parallelStatefulsetPodCopy1, parallelStatefulsetPodCopy2}, makeFakePodsIgnoreClaims(parallelStatefulset.UID, parallelStatefulsetPodCopy1, 7)...),
		},
		{
			name:             "Statefulset - sequential pod management policy",
			scheduledPods:    []*apiv1.Pod{scheduledSequentialStatefulsetPod},
			unschedulabePods: []*apiv1.Pod{sequentialStatefulsetPodCopy1, sequentialStatefulsetPodCopy2},
			wantPods:         []*apiv1.Pod{sequentialStatefulsetPodCopy1, sequentialStatefulsetPodCopy2},
		},
		{
			name:             "Mix of controllers",
			scheduledPods:    []*apiv1.Pod{scheduledPodRep1Copy1, scheduledPodJob1Copy1, scheduledParallelStatefulsetPod},
			unschedulabePods: []*apiv1.Pod{podRep1Copy1, podRep1Copy2, podJob1Copy1, podJob1Copy2, parallelStatefulsetPodCopy1, parallelStatefulsetPodCopy2},
			wantPods: append(
				append(
					append(
						[]*apiv1.Pod{podRep1Copy1, podRep1Copy2, podJob1Copy1, podJob1Copy2, parallelStatefulsetPodCopy1, parallelStatefulsetPodCopy2},
						makeFakePodsIgnoreClaims(replicaSet1.UID, podRep1Copy1, 2)...),
					makeFakePodsIgnoreClaims(job1.UID, podJob1Copy1, 7)...),
				makeFakePodsIgnoreClaims(parallelStatefulset.UID, parallelStatefulsetPodCopy1, 7)...,
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewPodInjectionPodListProcessor(podinjectionbackoff.NewFakePodControllerRegistry())
			clusterSnapshot := clustersnapshot.NewDeltaClusterSnapshot(framework.TestFrameworkHandleOrDie(t), true)
			err := clusterSnapshot.AddNodeInfo(framework.NewTestNodeInfo(node, tc.scheduledPods...))
			assert.NoError(t, err)
			ctx := context.AutoscalingContext{
				AutoscalingKubeClients: context.AutoscalingKubeClients{
					ListerRegistry: kubernetes.NewListerRegistry(nil, nil, nil, nil, nil, nil, jobLister, replicaSetLister, statefulsetLister),
				},
				ClusterSnapshot: clusterSnapshot,
			}
			pods, err := p.Process(&ctx, tc.unschedulabePods)
			assert.NoError(t, err)
			assert.ElementsMatch(t, tc.wantPods, pods)
		})
	}
}

func makeFakePodsIgnoreClaims(ownerUid types.UID, samplePod *apiv1.Pod, podCount int) []*apiv1.Pod {
	podInfos, _ := makeFakePods(ownerUid, &framework.PodInfo{Pod: samplePod}, podCount)
	var result []*apiv1.Pod
	for _, podInfo := range podInfos {
		result = append(result, podInfo.Pod)
	}
	return result
}

func TestGroupPods(t *testing.T) {
	noControllerPod := buildTestPod("default", "pod-no-podGroup")

	replicaSet1 := createTestReplicaSet("rep-set-1", "default", 10)
	podRep1Copy1 := buildTestPod("default", "pod-rep1-1", withControllerOwnerRef(replicaSet1.Name, "ReplicaSet", replicaSet1.UID))
	podRep1Copy2 := buildTestPod("default", "pod-rep1-2", withControllerOwnerRef(replicaSet1.Name, "ReplicaSet", replicaSet1.UID))
	podRep1ScheduledCopy1 := buildTestPod("default", "pod-rep1-3", withControllerOwnerRef(replicaSet1.Name, "ReplicaSet", replicaSet1.UID), withNodeName("n1"))
	podRep1ScheduledCopy2 := buildTestPod("default", "pod-rep1-4", withControllerOwnerRef(replicaSet1.Name, "ReplicaSet", replicaSet1.UID), withNodeName("n1"))

	replicaSet2 := createTestReplicaSet("rep-set-2", "default", 10)
	podRep2Copy1 := buildTestPod("default", "pod-rep2-1", withControllerOwnerRef(replicaSet2.Name, "ReplicaSet", replicaSet2.UID))
	podRep2ScheduledCopy1 := buildTestPod("default", "pod-rep2-1", withControllerOwnerRef(replicaSet2.Name, "ReplicaSet", replicaSet2.UID), withNodeName("n1"))

	replicaSet3 := createTestReplicaSet("rep-set-3", "default", 10)
	podRep3Copy1 := buildTestPod("default", "pod-rep3-1", withControllerOwnerRef(replicaSet3.Name, "ReplicaSet", replicaSet3.UID))

	job1 := createTestJob("job-1", "default", 10, 10, 0)
	podJob1Copy1 := buildTestPod("default", "pod-job1-1", withControllerOwnerRef(job1.Name, "Job", job1.UID))
	podJob1Copy2 := buildTestPod("default", "pod-job1-2", withControllerOwnerRef(job1.Name, "Job", job1.UID))

	job2 := createTestJob("job-2", "default", 10, 10, 0)
	podJob2Copy1 := buildTestPod("default", "pod-job-2", withControllerOwnerRef(job2.Name, "Job", job2.UID))

	statefulset1 := createTestStatefulset("statefulset-1", "default", appsv1.ParallelPodManagement, 10)
	statefulset1Copy1 := buildTestPod("default", "pod-statefulset1-1", withControllerOwnerRef(statefulset1.Name, "StatefulSet", statefulset1.UID))
	statefulset1Copy2 := buildTestPod("default", "pod-statefulset1-2", withControllerOwnerRef(statefulset1.Name, "StatefulSet", statefulset1.UID))

	statefulset2 := createTestStatefulset("statefulset-2", "default", appsv1.ParallelPodManagement, 10)
	statefulset2Copy1 := buildTestPod("default", "pod-statefulset2-1", withControllerOwnerRef(statefulset2.Name, "StatefulSet", statefulset2.UID))

	testCases := []struct {
		name            string
		unscheduledPods []*apiv1.Pod
		scheduledPods   []*apiv1.Pod
		replicaSets     []*appsv1.ReplicaSet
		jobs            []*batchv1.Job
		statefulsets    []*appsv1.StatefulSet
		wantGroupedPods map[types.UID]podGroup
	}{
		{
			name:        "no pods",
			replicaSets: []*appsv1.ReplicaSet{&replicaSet1, &replicaSet2},
			wantGroupedPods: map[types.UID]podGroup{
				replicaSet1.UID: {podCount: 0, desiredReplicas: 10, sample: nil},
				replicaSet2.UID: {podCount: 0, desiredReplicas: 10, sample: nil},
			},
		},
		{
			name:          "no unschedulable pods",
			scheduledPods: []*apiv1.Pod{podRep1ScheduledCopy1, podRep1ScheduledCopy2, podRep2ScheduledCopy1},
			replicaSets:   []*appsv1.ReplicaSet{&replicaSet1, &replicaSet2},
			wantGroupedPods: map[types.UID]podGroup{
				replicaSet1.UID: {podCount: 2, desiredReplicas: 10, sample: nil},
				replicaSet2.UID: {podCount: 1, desiredReplicas: 10, sample: nil},
			},
		},
		{
			name:            "scheduled and unschedulable pods",
			scheduledPods:   []*apiv1.Pod{podRep1ScheduledCopy2},
			unscheduledPods: []*apiv1.Pod{podRep1Copy1, podRep2Copy1},
			replicaSets:     []*appsv1.ReplicaSet{&replicaSet1, &replicaSet2},
			wantGroupedPods: map[types.UID]podGroup{
				replicaSet1.UID: {podCount: 2, desiredReplicas: 10, sample: &framework.PodInfo{Pod: podRep1Copy1}, ownerUid: replicaSet1.UID},
				replicaSet2.UID: {podCount: 1, desiredReplicas: 10, sample: &framework.PodInfo{Pod: podRep2Copy1}, ownerUid: replicaSet2.UID},
			},
		},
		{
			name:            "pods without a controller are ignored",
			unscheduledPods: []*apiv1.Pod{noControllerPod},
			wantGroupedPods: map[types.UID]podGroup{},
		},
		{
			name:            "unable to retrieve a controller - pods are ignored",
			unscheduledPods: []*apiv1.Pod{podRep3Copy1},
			wantGroupedPods: map[types.UID]podGroup{},
		},
		{
			name:            "pods form multiple replicaSets",
			unscheduledPods: []*apiv1.Pod{podRep1Copy1, podRep1Copy2, podRep2Copy1},
			replicaSets:     []*appsv1.ReplicaSet{&replicaSet1, &replicaSet2},
			wantGroupedPods: map[types.UID]podGroup{
				replicaSet1.UID: {podCount: 2, desiredReplicas: 10, sample: &framework.PodInfo{Pod: podRep1Copy1}, ownerUid: replicaSet1.UID},
				replicaSet2.UID: {podCount: 1, desiredReplicas: 10, sample: &framework.PodInfo{Pod: podRep2Copy1}, ownerUid: replicaSet2.UID},
			},
		},
		{
			name:            "pods form multiple jobs",
			unscheduledPods: []*apiv1.Pod{podJob1Copy1, podJob1Copy2, podJob2Copy1},
			jobs:            []*batchv1.Job{&job1, &job2},
			wantGroupedPods: map[types.UID]podGroup{
				job1.UID: {podCount: 2, desiredReplicas: 10, sample: &framework.PodInfo{Pod: podJob1Copy1}, ownerUid: job1.UID},
				job2.UID: {podCount: 1, desiredReplicas: 10, sample: &framework.PodInfo{Pod: podJob2Copy1}, ownerUid: job2.UID},
			},
		},
		{
			name:            "pods form multiple statefulsets",
			unscheduledPods: []*apiv1.Pod{statefulset1Copy1, statefulset1Copy2, statefulset2Copy1},
			statefulsets:    []*appsv1.StatefulSet{&statefulset1, &statefulset2},
			wantGroupedPods: map[types.UID]podGroup{
				statefulset1.UID: {podCount: 2, desiredReplicas: 10, sample: &framework.PodInfo{Pod: statefulset1Copy1}, ownerUid: statefulset1.UID},
				statefulset2.UID: {podCount: 1, desiredReplicas: 10, sample: &framework.PodInfo{Pod: statefulset2Copy1}, ownerUid: statefulset2.UID},
			},
		},
		{
			name:            "unscheduledPods from multiple different controllers",
			unscheduledPods: []*apiv1.Pod{podRep1Copy1, podRep1Copy2, podRep2Copy1, podJob1Copy1, statefulset1Copy1},
			replicaSets:     []*appsv1.ReplicaSet{&replicaSet1, &replicaSet2},
			jobs:            []*batchv1.Job{&job1},
			statefulsets:    []*appsv1.StatefulSet{&statefulset1},
			wantGroupedPods: map[types.UID]podGroup{
				replicaSet1.UID:  {podCount: 2, desiredReplicas: 10, sample: &framework.PodInfo{Pod: podRep1Copy1}, ownerUid: replicaSet1.UID},
				replicaSet2.UID:  {podCount: 1, desiredReplicas: 10, sample: &framework.PodInfo{Pod: podRep2Copy1}, ownerUid: replicaSet2.UID},
				job1.UID:         {podCount: 1, desiredReplicas: 10, sample: &framework.PodInfo{Pod: podJob1Copy1}, ownerUid: job1.UID},
				statefulset1.UID: {podCount: 1, desiredReplicas: 10, sample: &framework.PodInfo{Pod: statefulset1Copy1}, ownerUid: statefulset1.UID},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			replicaSetLister, err := kubernetes.NewTestReplicaSetLister(tc.replicaSets)
			assert.NoError(t, err)
			jobLister, err := kubernetes.NewTestJobLister(tc.jobs)
			assert.NoError(t, err)
			statefulsetLister, err := kubernetes.NewTestStatefulSetLister(tc.statefulsets)
			assert.NoError(t, err)

			ctx := context.AutoscalingContext{
				ClusterSnapshot: clustersnapshot.NewBasicClusterSnapshot(framework.TestFrameworkHandleOrDie(t), true),
				AutoscalingKubeClients: context.AutoscalingKubeClients{
					ListerRegistry: kubernetes.NewListerRegistry(nil, nil, nil, nil, nil, nil, jobLister, replicaSetLister, statefulsetLister),
				},
			}
			controllers := listControllers(&ctx)
			groupedPods := groupPods(append(tc.scheduledPods, tc.unscheduledPods...), controllers, ctx.ClusterSnapshot, true)
			assert.Equal(t, tc.wantGroupedPods, groupedPods)
		})
	}
}

func TestUpdatePodGroups(t *testing.T) {
	replicaSet1 := createTestReplicaSet("rep-set-1", "default", 10)
	podRep1Copy1 := buildTestPod("default", "pod-rep1-1", withControllerOwnerRef(replicaSet1.Name, "ReplicaSet", replicaSet1.UID))
	podRep1Copy2 := buildTestPod("default", "pod-rep1-2", withControllerOwnerRef(replicaSet1.Name, "ReplicaSet", replicaSet1.UID))
	samplePodGroups := map[types.UID]podGroup{replicaSet1.UID: makePodGroup(10)}
	sampleFalse := false
	sampleTrue := true

	testCases := []struct {
		name         string
		pod          *apiv1.Pod
		ownerRef     metav1.OwnerReference
		podGroups    map[types.UID]podGroup
		wantPodGroup map[types.UID]podGroup
	}{
		{
			name:         "owner ref nil controller",
			pod:          podRep1Copy1,
			ownerRef:     metav1.OwnerReference{},
			podGroups:    samplePodGroups,
			wantPodGroup: samplePodGroups,
		},
		{
			name:         "owner ref controller set to false",
			pod:          podRep1Copy1,
			ownerRef:     metav1.OwnerReference{Controller: &sampleFalse},
			podGroups:    samplePodGroups,
			wantPodGroup: samplePodGroups,
		},
		{
			name:         "owner ref controller not found",
			pod:          podRep1Copy1,
			ownerRef:     metav1.OwnerReference{Controller: &sampleTrue, UID: types.UID("not found uid")},
			podGroups:    samplePodGroups,
			wantPodGroup: samplePodGroups,
		},
		{
			name:      "sample pod added and count updated",
			pod:       podRep1Copy1,
			ownerRef:  podRep1Copy1.OwnerReferences[0],
			podGroups: samplePodGroups,
			wantPodGroup: map[types.UID]podGroup{replicaSet1.UID: {
				podCount:        1,
				desiredReplicas: 10,
				sample:          &framework.PodInfo{Pod: podRep1Copy1},
				ownerUid:        replicaSet1.UID,
			},
			},
		},
		{
			name:     "only count updated",
			pod:      podRep1Copy2,
			ownerRef: podRep1Copy1.OwnerReferences[0],
			podGroups: map[types.UID]podGroup{replicaSet1.UID: {
				podCount:        1,
				desiredReplicas: 10,
				sample:          &framework.PodInfo{Pod: podRep1Copy1},
				ownerUid:        replicaSet1.UID,
			},
			},
			wantPodGroup: map[types.UID]podGroup{replicaSet1.UID: {
				podCount:        2,
				desiredReplicas: 10,
				sample:          &framework.PodInfo{Pod: podRep1Copy1},
				ownerUid:        replicaSet1.UID,
			},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			podGroups := updatePodGroups(tc.pod, tc.ownerRef, tc.podGroups, clustersnapshot.NewBasicClusterSnapshot(framework.TestFrameworkHandleOrDie(t), true), true)
			assert.Equal(t, tc.wantPodGroup, podGroups)
		})
	}
}

func TestMakeFakePods(t *testing.T) {
	samplePod := buildTestPod("default", "test-pod")
	// Test case: Positive fake pod count
	fakePodCount := 5
	ownerUid := types.UID("sample uid")
	fakePods, err := makeFakePods(ownerUid, &framework.PodInfo{Pod: samplePod}, fakePodCount)
	assert.NoError(t, err)
	assert.Equal(t, fakePodCount, len(fakePods))
	for idx, fakePod := range fakePods {
		assert.Equal(t, fakePod.Name, fmt.Sprintf("%s-copy-%d", samplePod.Name, idx+1))
		assert.Equal(t, fakePod.UID, types.UID(fmt.Sprintf("%s-%d", string(ownerUid), idx+1)))
		assert.NotNil(t, fakePod.Annotations)
		assert.Equal(t, fakePod.Annotations[FakePodAnnotationKey], FakePodAnnotationValue)
	}

	// Test case: Zero fake pod count
	fakePodCount = 0
	fakePods, err = makeFakePods(ownerUid, &framework.PodInfo{Pod: samplePod}, fakePodCount)
	assert.NoError(t, err)
	assert.Nil(t, fakePods)
}

func createTestReplicaSet(uid, namespace string, targetReplicaCount int32) appsv1.ReplicaSet {
	return appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{UID: types.UID(uid), Name: uid, Namespace: namespace},
		Spec: appsv1.ReplicaSetSpec{
			Replicas: &targetReplicaCount,
		},
	}
}

func createTestJob(uid, namespace string, parallelism, completions, succeeded int32) batchv1.Job {
	return batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{UID: types.UID(uid), Name: uid, Namespace: namespace},
		Spec: batchv1.JobSpec{
			Parallelism: &parallelism,
			Completions: &completions,
		},
		Status: batchv1.JobStatus{
			Succeeded: succeeded,
		},
	}
}
func createTestStatefulset(uid, namespace string, podManagementPolicy appsv1.PodManagementPolicyType, numReplicas int32) appsv1.StatefulSet {
	return appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{UID: types.UID(uid), Name: uid, Namespace: namespace},
		Spec: appsv1.StatefulSetSpec{
			Replicas:            &numReplicas,
			PodManagementPolicy: podManagementPolicy,
		},
	}
}

func buildTestPod(namespace, name string, opts ...podOption) *apiv1.Pod {
	pod := BuildTestPod(name, 10, 10)
	pod.Namespace = namespace
	for _, opt := range opts {
		opt(pod)
	}
	return pod
}

type podOption func(*apiv1.Pod)

func withControllerOwnerRef(name, kind string, uid types.UID) podOption {
	return func(pod *apiv1.Pod) {
		pod.OwnerReferences = GenerateOwnerReferences(name, kind, "apps/v1", uid)
	}
}

func withNodeName(nodeName string) podOption {
	return func(pod *apiv1.Pod) {
		pod.Spec.NodeName = nodeName
	}
}
