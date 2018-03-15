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

package clusterstate

import (
	"testing"
	"time"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate/api"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate/utils"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/client-go/kubernetes/fake"
	kube_record "k8s.io/client-go/tools/record"

	"github.com/stretchr/testify/assert"
)

func TestOKWithScaleUp(t *testing.T) {
	now := time.Now()

	ng1_1 := BuildTestNode("ng1-1", 1000, 1000)
	SetNodeReadyState(ng1_1, true, now.Add(-time.Minute))
	ng2_1 := BuildTestNode("ng2-1", 1000, 1000)
	SetNodeReadyState(ng2_1, true, now.Add(-time.Minute))

	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 1, 10, 5)
	provider.AddNodeGroup("ng2", 1, 10, 1)

	provider.AddNode("ng1", ng1_1)
	provider.AddNode("ng2", ng2_1)
	assert.NotNil(t, provider)

	fakeClient := &fake.Clientset{}
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false)
	clusterstate := NewClusterStateRegistry(provider, ClusterStateRegistryConfig{
		MaxTotalUnreadyPercentage: 10,
		OkTotalUnreadyCount:       1,
	}, fakeLogRecorder)
	clusterstate.RegisterScaleUp(&ScaleUpRequest{
		NodeGroupName:   "ng1",
		Increase:        4,
		Time:            now,
		ExpectedAddTime: now.Add(time.Minute),
	})
	err := clusterstate.UpdateNodes([]*apiv1.Node{ng1_1, ng2_1}, now)
	assert.NoError(t, err)
	assert.True(t, clusterstate.IsClusterHealthy())

	status := clusterstate.GetStatus(now)
	assert.Equal(t, api.ClusterAutoscalerInProgress,
		api.GetConditionByType(api.ClusterAutoscalerScaleUp, status.ClusterwideConditions).Status)
	assert.Equal(t, 2, len(status.NodeGroupStatuses))
	ng1Checked := false
	ng2Checked := true
	for _, nodeStatus := range status.NodeGroupStatuses {
		if nodeStatus.ProviderID == "ng1" {
			assert.Equal(t, api.ClusterAutoscalerInProgress,
				api.GetConditionByType(api.ClusterAutoscalerScaleUp, nodeStatus.Conditions).Status)
			ng1Checked = true
		}
		if nodeStatus.ProviderID == "ng2" {
			assert.Equal(t, api.ClusterAutoscalerNoActivity,
				api.GetConditionByType(api.ClusterAutoscalerScaleUp, nodeStatus.Conditions).Status)
			ng2Checked = true
		}
	}
	assert.True(t, ng1Checked)
	assert.True(t, ng2Checked)
}

func TestEmptyOK(t *testing.T) {
	now := time.Now()

	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 0, 10, 0)
	assert.NotNil(t, provider)

	fakeClient := &fake.Clientset{}
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false)
	clusterstate := NewClusterStateRegistry(provider, ClusterStateRegistryConfig{
		MaxTotalUnreadyPercentage: 10,
		OkTotalUnreadyCount:       1,
	}, fakeLogRecorder)
	err := clusterstate.UpdateNodes([]*apiv1.Node{}, now.Add(-5*time.Second))
	assert.NoError(t, err)
	assert.True(t, clusterstate.IsClusterHealthy())
	assert.True(t, clusterstate.IsNodeGroupHealthy("ng1"))
	assert.False(t, clusterstate.IsNodeGroupScalingUp("ng1"))

	provider.AddNodeGroup("ng1", 0, 10, 3)
	clusterstate.RegisterScaleUp(&ScaleUpRequest{
		NodeGroupName:   "ng1",
		Increase:        3,
		Time:            now.Add(-3 * time.Second),
		ExpectedAddTime: now.Add(1 * time.Minute),
	})
	err = clusterstate.UpdateNodes([]*apiv1.Node{}, now)

	assert.NoError(t, err)
	assert.True(t, clusterstate.IsClusterHealthy())
	assert.True(t, clusterstate.IsNodeGroupHealthy("ng1"))
	assert.True(t, clusterstate.IsNodeGroupScalingUp("ng1"))
}

func TestOKOneUnreadyNode(t *testing.T) {
	now := time.Now()

	ng1_1 := BuildTestNode("ng1-1", 1000, 1000)
	SetNodeReadyState(ng1_1, true, now.Add(-time.Minute))
	ng2_1 := BuildTestNode("ng2-1", 1000, 1000)
	SetNodeReadyState(ng2_1, false, now.Add(-time.Minute))

	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 1, 10, 1)
	provider.AddNodeGroup("ng2", 1, 10, 1)
	provider.AddNode("ng1", ng1_1)
	provider.AddNode("ng2", ng2_1)
	assert.NotNil(t, provider)

	fakeClient := &fake.Clientset{}
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false)
	clusterstate := NewClusterStateRegistry(provider, ClusterStateRegistryConfig{
		MaxTotalUnreadyPercentage: 10,
		OkTotalUnreadyCount:       1,
	}, fakeLogRecorder)
	err := clusterstate.UpdateNodes([]*apiv1.Node{ng1_1, ng2_1}, now)
	assert.NoError(t, err)
	assert.True(t, clusterstate.IsClusterHealthy())
	assert.True(t, clusterstate.IsNodeGroupHealthy("ng1"))

	status := clusterstate.GetStatus(now)
	assert.Equal(t, api.ClusterAutoscalerHealthy,
		api.GetConditionByType(api.ClusterAutoscalerHealth, status.ClusterwideConditions).Status)
	assert.Equal(t, api.ClusterAutoscalerNoActivity,
		api.GetConditionByType(api.ClusterAutoscalerScaleUp, status.ClusterwideConditions).Status)

	assert.Equal(t, 2, len(status.NodeGroupStatuses))
	ng1Checked := false
	for _, nodeStatus := range status.NodeGroupStatuses {
		if nodeStatus.ProviderID == "ng1" {
			assert.Equal(t, api.ClusterAutoscalerHealthy,
				api.GetConditionByType(api.ClusterAutoscalerHealth, nodeStatus.Conditions).Status)
			ng1Checked = true
		}
	}
	assert.True(t, ng1Checked)
}

func TestNodeWithoutNodeGroupDontCrash(t *testing.T) {
	now := time.Now()

	noNgNode := BuildTestNode("no_ng", 1000, 1000)
	SetNodeReadyState(noNgNode, true, now.Add(-time.Minute))
	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNode("no_ng", noNgNode)

	fakeClient := &fake.Clientset{}
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false)
	clusterstate := NewClusterStateRegistry(provider, ClusterStateRegistryConfig{
		MaxTotalUnreadyPercentage: 10,
		OkTotalUnreadyCount:       1,
	}, fakeLogRecorder)
	err := clusterstate.UpdateNodes([]*apiv1.Node{noNgNode}, now)
	assert.NoError(t, err)
	clusterstate.UpdateScaleDownCandidates([]*apiv1.Node{noNgNode}, now)
}

func TestOKOneUnreadyNodeWithScaleDownCandidate(t *testing.T) {
	now := time.Now()

	ng1_1 := BuildTestNode("ng1-1", 1000, 1000)
	SetNodeReadyState(ng1_1, true, now.Add(-time.Minute))
	ng2_1 := BuildTestNode("ng2-1", 1000, 1000)
	SetNodeReadyState(ng2_1, false, now.Add(-time.Minute))

	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 1, 10, 1)
	provider.AddNodeGroup("ng2", 1, 10, 1)
	provider.AddNode("ng1", ng1_1)
	provider.AddNode("ng2", ng2_1)
	assert.NotNil(t, provider)

	fakeClient := &fake.Clientset{}
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false)
	clusterstate := NewClusterStateRegistry(provider, ClusterStateRegistryConfig{
		MaxTotalUnreadyPercentage: 10,
		OkTotalUnreadyCount:       1,
	}, fakeLogRecorder)
	err := clusterstate.UpdateNodes([]*apiv1.Node{ng1_1, ng2_1}, now)
	clusterstate.UpdateScaleDownCandidates([]*apiv1.Node{ng1_1}, now)

	assert.NoError(t, err)
	assert.True(t, clusterstate.IsClusterHealthy())
	assert.True(t, clusterstate.IsNodeGroupHealthy("ng1"))

	status := clusterstate.GetStatus(now)
	assert.Equal(t, api.ClusterAutoscalerHealthy,
		api.GetConditionByType(api.ClusterAutoscalerHealth, status.ClusterwideConditions).Status)
	assert.Equal(t, api.ClusterAutoscalerNoActivity,
		api.GetConditionByType(api.ClusterAutoscalerScaleUp, status.ClusterwideConditions).Status)
	assert.Equal(t, api.ClusterAutoscalerCandidatesPresent,
		api.GetConditionByType(api.ClusterAutoscalerScaleDown, status.ClusterwideConditions).Status)

	assert.Equal(t, 2, len(status.NodeGroupStatuses))
	ng1Checked := false
	ng2Checked := false
	for _, nodeStatus := range status.NodeGroupStatuses {
		if nodeStatus.ProviderID == "ng1" {
			assert.Equal(t, api.ClusterAutoscalerHealthy,
				api.GetConditionByType(api.ClusterAutoscalerHealth, nodeStatus.Conditions).Status)

			assert.Equal(t, api.ClusterAutoscalerCandidatesPresent,
				api.GetConditionByType(api.ClusterAutoscalerScaleDown, nodeStatus.Conditions).Status)

			ng1Checked = true
		}
		if nodeStatus.ProviderID == "ng2" {
			assert.Equal(t, api.ClusterAutoscalerHealthy,
				api.GetConditionByType(api.ClusterAutoscalerHealth, nodeStatus.Conditions).Status)

			assert.Equal(t, api.ClusterAutoscalerNoCandidates,
				api.GetConditionByType(api.ClusterAutoscalerScaleDown, nodeStatus.Conditions).Status)

			ng2Checked = true
		}
	}
	assert.True(t, ng1Checked)
	assert.True(t, ng2Checked)
}

func TestMissingNodes(t *testing.T) {
	now := time.Now()

	ng1_1 := BuildTestNode("ng1-1", 1000, 1000)
	SetNodeReadyState(ng1_1, true, now.Add(-time.Minute))
	ng2_1 := BuildTestNode("ng2-1", 1000, 1000)
	SetNodeReadyState(ng2_1, true, now.Add(-time.Minute))

	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 1, 10, 5)
	provider.AddNodeGroup("ng2", 1, 10, 1)

	provider.AddNode("ng1", ng1_1)
	provider.AddNode("ng2", ng2_1)
	assert.NotNil(t, provider)

	fakeClient := &fake.Clientset{}
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false)
	clusterstate := NewClusterStateRegistry(provider, ClusterStateRegistryConfig{
		MaxTotalUnreadyPercentage: 10,
		OkTotalUnreadyCount:       1,
	}, fakeLogRecorder)
	err := clusterstate.UpdateNodes([]*apiv1.Node{ng1_1, ng2_1}, now)
	assert.NoError(t, err)
	assert.True(t, clusterstate.IsClusterHealthy())
	assert.False(t, clusterstate.IsNodeGroupHealthy("ng1"))

	status := clusterstate.GetStatus(now)
	assert.Equal(t, api.ClusterAutoscalerHealthy,
		api.GetConditionByType(api.ClusterAutoscalerHealth, status.ClusterwideConditions).Status)
	assert.Equal(t, 2, len(status.NodeGroupStatuses))
	ng1Checked := false
	for _, nodeStatus := range status.NodeGroupStatuses {
		if nodeStatus.ProviderID == "ng1" {
			assert.Equal(t, api.ClusterAutoscalerUnhealthy,
				api.GetConditionByType(api.ClusterAutoscalerHealth, nodeStatus.Conditions).Status)
			ng1Checked = true
		}
	}
	assert.True(t, ng1Checked)
}

func TestToManyUnready(t *testing.T) {
	now := time.Now()

	ng1_1 := BuildTestNode("ng1-1", 1000, 1000)
	SetNodeReadyState(ng1_1, false, now.Add(-time.Minute))
	ng2_1 := BuildTestNode("ng2-1", 1000, 1000)
	SetNodeReadyState(ng2_1, false, now.Add(-time.Minute))

	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 1, 10, 1)
	provider.AddNodeGroup("ng2", 1, 10, 1)
	provider.AddNode("ng1", ng1_1)
	provider.AddNode("ng2", ng2_1)

	assert.NotNil(t, provider)
	fakeClient := &fake.Clientset{}
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false)
	clusterstate := NewClusterStateRegistry(provider, ClusterStateRegistryConfig{
		MaxTotalUnreadyPercentage: 10,
		OkTotalUnreadyCount:       1,
	}, fakeLogRecorder)
	err := clusterstate.UpdateNodes([]*apiv1.Node{ng1_1, ng2_1}, now)
	assert.NoError(t, err)
	assert.False(t, clusterstate.IsClusterHealthy())
	assert.True(t, clusterstate.IsNodeGroupHealthy("ng1"))
}

func TestExpiredScaleUp(t *testing.T) {
	now := time.Now()

	ng1_1 := BuildTestNode("ng1-1", 1000, 1000)
	SetNodeReadyState(ng1_1, true, now.Add(-time.Minute))

	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 1, 10, 5)
	provider.AddNode("ng1", ng1_1)
	assert.NotNil(t, provider)

	fakeClient := &fake.Clientset{}
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false)
	clusterstate := NewClusterStateRegistry(provider, ClusterStateRegistryConfig{
		MaxTotalUnreadyPercentage: 10,
		OkTotalUnreadyCount:       1,
	}, fakeLogRecorder)
	clusterstate.RegisterScaleUp(&ScaleUpRequest{
		NodeGroupName:   "ng1",
		Increase:        4,
		Time:            now.Add(-3 * time.Minute),
		ExpectedAddTime: now.Add(-1 * time.Minute),
	})
	err := clusterstate.UpdateNodes([]*apiv1.Node{ng1_1}, now)
	assert.NoError(t, err)
	assert.True(t, clusterstate.IsClusterHealthy())
	assert.False(t, clusterstate.IsNodeGroupHealthy("ng1"))
}

func TestRegisterScaleDown(t *testing.T) {
	ng1_1 := BuildTestNode("ng1-1", 1000, 1000)
	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 1, 10, 1)
	provider.AddNode("ng1", ng1_1)
	assert.NotNil(t, provider)

	fakeClient := &fake.Clientset{}
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false)
	clusterstate := NewClusterStateRegistry(provider, ClusterStateRegistryConfig{
		MaxTotalUnreadyPercentage: 10,
		OkTotalUnreadyCount:       1,
	}, fakeLogRecorder)

	now := time.Now()

	clusterstate.RegisterScaleDown(&ScaleDownRequest{
		NodeGroupName:      "ng1",
		NodeName:           "ng1-1",
		ExpectedDeleteTime: now.Add(time.Minute),
		Time:               now,
	})
	assert.Equal(t, 1, len(clusterstate.scaleDownRequests))
	clusterstate.updateScaleRequests(now.Add(5 * time.Minute))
	assert.Equal(t, 0, len(clusterstate.scaleDownRequests))
}

func TestUpcomingNodes(t *testing.T) {
	provider := testprovider.NewTestCloudProvider(nil, nil)
	now := time.Now()

	// 6 nodes are expected to come.
	ng1_1 := BuildTestNode("ng1-1", 1000, 1000)
	SetNodeReadyState(ng1_1, true, now.Add(-time.Minute))
	provider.AddNodeGroup("ng1", 1, 10, 7)
	provider.AddNode("ng1", ng1_1)

	// One node is expected to come. One node is unready for the long time
	// but this should not make any difference.
	ng2_1 := BuildTestNode("ng2-1", 1000, 1000)
	SetNodeReadyState(ng2_1, false, now.Add(-time.Minute))
	provider.AddNodeGroup("ng2", 1, 10, 2)
	provider.AddNode("ng2", ng2_1)

	// Two nodes are expected to come. One is just being started for the first time,
	// the other one is not there yet.
	ng3_1 := BuildTestNode("ng3-1", 1000, 1000)
	SetNodeReadyState(ng3_1, false, now.Add(-time.Minute))
	ng3_1.CreationTimestamp = metav1.Time{Time: now.Add(-time.Minute)}
	provider.AddNodeGroup("ng3", 1, 10, 2)
	provider.AddNode("ng3", ng3_1)

	// Nothing should be added here.
	ng4_1 := BuildTestNode("ng4-1", 1000, 1000)
	SetNodeReadyState(ng4_1, false, now.Add(-time.Minute))
	provider.AddNodeGroup("ng4", 1, 10, 1)
	provider.AddNode("ng4", ng4_1)

	assert.NotNil(t, provider)
	fakeClient := &fake.Clientset{}
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false)
	clusterstate := NewClusterStateRegistry(provider, ClusterStateRegistryConfig{
		MaxTotalUnreadyPercentage: 10,
		OkTotalUnreadyCount:       1,
	}, fakeLogRecorder)
	err := clusterstate.UpdateNodes([]*apiv1.Node{ng1_1, ng2_1, ng3_1, ng4_1}, now)
	assert.NoError(t, err)

	upcomingNodes := clusterstate.GetUpcomingNodes()
	assert.Equal(t, 6, upcomingNodes["ng1"])
	assert.Equal(t, 1, upcomingNodes["ng2"])
	assert.Equal(t, 2, upcomingNodes["ng3"])
	assert.NotContains(t, upcomingNodes, "ng4")
}

func TestIncorrectSize(t *testing.T) {
	ng1_1 := BuildTestNode("ng1-1", 1000, 1000)
	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 1, 10, 5)
	provider.AddNode("ng1", ng1_1)
	assert.NotNil(t, provider)
	fakeClient := &fake.Clientset{}
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false)
	clusterstate := NewClusterStateRegistry(provider, ClusterStateRegistryConfig{
		MaxTotalUnreadyPercentage: 10,
		OkTotalUnreadyCount:       1,
	}, fakeLogRecorder)
	now := time.Now()
	clusterstate.UpdateNodes([]*apiv1.Node{ng1_1}, now.Add(-5*time.Minute))
	incorrect := clusterstate.incorrectNodeGroupSizes["ng1"]
	assert.Equal(t, 5, incorrect.ExpectedSize)
	assert.Equal(t, 1, incorrect.CurrentSize)
	assert.Equal(t, now.Add(-5*time.Minute), incorrect.FirstObserved)

	clusterstate.UpdateNodes([]*apiv1.Node{ng1_1}, now.Add(-4*time.Minute))
	incorrect = clusterstate.incorrectNodeGroupSizes["ng1"]
	assert.Equal(t, 5, incorrect.ExpectedSize)
	assert.Equal(t, 1, incorrect.CurrentSize)
	assert.Equal(t, now.Add(-5*time.Minute), incorrect.FirstObserved)

	clusterstate.UpdateNodes([]*apiv1.Node{ng1_1, ng1_1}, now.Add(-3*time.Minute))
	incorrect = clusterstate.incorrectNodeGroupSizes["ng1"]
	assert.Equal(t, 5, incorrect.ExpectedSize)
	assert.Equal(t, 2, incorrect.CurrentSize)
	assert.Equal(t, now.Add(-3*time.Minute), incorrect.FirstObserved)
}

func TestUnregisteredNodes(t *testing.T) {
	ng1_1 := BuildTestNode("ng1-1", 1000, 1000)
	ng1_1.Spec.ProviderID = "ng1-1"
	ng1_2 := BuildTestNode("ng1-2", 1000, 1000)
	ng1_2.Spec.ProviderID = "ng1-2"
	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 1, 10, 2)
	provider.AddNode("ng1", ng1_1)
	provider.AddNode("ng1", ng1_2)

	fakeClient := &fake.Clientset{}
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false)
	clusterstate := NewClusterStateRegistry(provider, ClusterStateRegistryConfig{
		MaxTotalUnreadyPercentage: 10,
		OkTotalUnreadyCount:       1,
		MaxNodeProvisionTime:      10 * time.Second,
	}, fakeLogRecorder)
	err := clusterstate.UpdateNodes([]*apiv1.Node{ng1_1}, time.Now().Add(-time.Minute))

	assert.NoError(t, err)
	assert.Equal(t, 1, len(clusterstate.GetUnregisteredNodes()))
	assert.Equal(t, "ng1-2", clusterstate.GetUnregisteredNodes()[0].Node.Name)
	upcomingNodes := clusterstate.GetUpcomingNodes()
	assert.Equal(t, 1, upcomingNodes["ng1"])

	// The node didn't come up in MaxNodeProvisionTime, it should no longer be
	// counted as upcoming (but it is still an unregistered node)
	err = clusterstate.UpdateNodes([]*apiv1.Node{ng1_1}, time.Now().Add(time.Minute))
	assert.NoError(t, err)
	assert.Equal(t, 1, len(clusterstate.GetUnregisteredNodes()))
	assert.Equal(t, "ng1-2", clusterstate.GetUnregisteredNodes()[0].Node.Name)
	upcomingNodes = clusterstate.GetUpcomingNodes()
	assert.Equal(t, 0, len(upcomingNodes))

	err = clusterstate.UpdateNodes([]*apiv1.Node{ng1_1, ng1_2}, time.Now().Add(time.Minute))
	assert.NoError(t, err)
	assert.Equal(t, 0, len(clusterstate.GetUnregisteredNodes()))
}

func TestUpdateLastTransitionTimes(t *testing.T) {
	now := metav1.Time{Time: time.Now()}
	later := metav1.Time{Time: now.Time.Add(10 * time.Second)}
	oldStatus := &api.ClusterAutoscalerStatus{
		ClusterwideConditions: make([]api.ClusterAutoscalerCondition, 0),
		NodeGroupStatuses:     make([]api.NodeGroupStatus, 0),
	}
	oldStatus.ClusterwideConditions = append(
		oldStatus.ClusterwideConditions,
		api.ClusterAutoscalerCondition{
			Type:               api.ClusterAutoscalerHealth,
			Status:             api.ClusterAutoscalerHealthy,
			LastProbeTime:      now,
			LastTransitionTime: now,
		})
	oldStatus.ClusterwideConditions = append(
		oldStatus.ClusterwideConditions,
		api.ClusterAutoscalerCondition{
			Type:               api.ClusterAutoscalerScaleUp,
			Status:             api.ClusterAutoscalerInProgress,
			LastProbeTime:      now,
			LastTransitionTime: now,
		})
	oldStatus.NodeGroupStatuses = append(
		oldStatus.NodeGroupStatuses,
		api.NodeGroupStatus{
			ProviderID: "ng1",
			Conditions: oldStatus.ClusterwideConditions,
		})

	newStatus := &api.ClusterAutoscalerStatus{
		ClusterwideConditions: make([]api.ClusterAutoscalerCondition, 0),
		NodeGroupStatuses:     make([]api.NodeGroupStatus, 0),
	}
	newStatus.ClusterwideConditions = append(
		newStatus.ClusterwideConditions,
		api.ClusterAutoscalerCondition{
			Type:          api.ClusterAutoscalerHealth,
			Status:        api.ClusterAutoscalerHealthy,
			LastProbeTime: later,
		})
	newStatus.ClusterwideConditions = append(
		newStatus.ClusterwideConditions,
		api.ClusterAutoscalerCondition{
			Type:          api.ClusterAutoscalerScaleUp,
			Status:        api.ClusterAutoscalerNotNeeded,
			LastProbeTime: later,
		})
	newStatus.ClusterwideConditions = append(
		newStatus.ClusterwideConditions,
		api.ClusterAutoscalerCondition{
			Type:          api.ClusterAutoscalerScaleDown,
			Status:        api.ClusterAutoscalerNoCandidates,
			LastProbeTime: later,
		})
	newStatus.NodeGroupStatuses = append(
		newStatus.NodeGroupStatuses,
		api.NodeGroupStatus{
			ProviderID: "ng2",
			Conditions: newStatus.ClusterwideConditions,
		})
	newStatus.NodeGroupStatuses = append(
		newStatus.NodeGroupStatuses,
		api.NodeGroupStatus{
			ProviderID: "ng1",
			Conditions: newStatus.ClusterwideConditions,
		})
	updateLastTransition(oldStatus, newStatus)

	for _, cwCondition := range newStatus.ClusterwideConditions {
		switch cwCondition.Type {
		case api.ClusterAutoscalerHealth:
			// Status has not changed
			assert.Equal(t, now, cwCondition.LastTransitionTime)
		case api.ClusterAutoscalerScaleUp:
			// Status has changed
			assert.Equal(t, later, cwCondition.LastTransitionTime)
		case api.ClusterAutoscalerScaleDown:
			// No old status information
			assert.Equal(t, later, cwCondition.LastTransitionTime)
		}
	}

	expectedNgTimestamps := make(map[string](map[api.ClusterAutoscalerConditionType]metav1.Time), 0)
	// Same as cluster-wide
	expectedNgTimestamps["ng1"] = map[api.ClusterAutoscalerConditionType]metav1.Time{
		api.ClusterAutoscalerHealth:    now,
		api.ClusterAutoscalerScaleUp:   later,
		api.ClusterAutoscalerScaleDown: later,
	}
	// New node group - everything should have latest timestamp as last transition time
	expectedNgTimestamps["ng2"] = map[api.ClusterAutoscalerConditionType]metav1.Time{
		api.ClusterAutoscalerHealth:    later,
		api.ClusterAutoscalerScaleUp:   later,
		api.ClusterAutoscalerScaleDown: later,
	}

	for _, ng := range newStatus.NodeGroupStatuses {
		expectations := expectedNgTimestamps[ng.ProviderID]
		for _, ngCondition := range ng.Conditions {
			assert.Equal(t, expectations[ngCondition.Type], ngCondition.LastTransitionTime)
		}
	}
}

func TestScaleUpBackoff(t *testing.T) {
	now := time.Now()

	ng1_1 := BuildTestNode("ng1-1", 1000, 1000)
	SetNodeReadyState(ng1_1, true, now.Add(-time.Minute))
	ng1_2 := BuildTestNode("ng1-2", 1000, 1000)
	SetNodeReadyState(ng1_2, true, now.Add(-time.Minute))
	ng1_3 := BuildTestNode("ng1-3", 1000, 1000)
	SetNodeReadyState(ng1_3, true, now.Add(-time.Minute))

	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 1, 10, 4)
	provider.AddNode("ng1", ng1_1)
	provider.AddNode("ng1", ng1_2)
	provider.AddNode("ng1", ng1_3)
	assert.NotNil(t, provider)

	fakeClient := &fake.Clientset{}
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false)
	clusterstate := NewClusterStateRegistry(provider, ClusterStateRegistryConfig{
		MaxTotalUnreadyPercentage: 10,
		OkTotalUnreadyCount:       1,
	}, fakeLogRecorder)

	// After failed scale-up, node group should be still healthy, but should backoff from scale-ups
	clusterstate.RegisterScaleUp(&ScaleUpRequest{
		NodeGroupName:   "ng1",
		Increase:        1,
		Time:            now.Add(-3 * time.Minute),
		ExpectedAddTime: now.Add(-1 * time.Minute),
	})
	err := clusterstate.UpdateNodes([]*apiv1.Node{ng1_1, ng1_2, ng1_3}, now)
	assert.NoError(t, err)
	assert.True(t, clusterstate.IsClusterHealthy())
	assert.True(t, clusterstate.IsNodeGroupHealthy("ng1"))
	assert.False(t, clusterstate.IsNodeGroupSafeToScaleUp("ng1", now))

	// Backoff should expire after timeout
	now = now.Add(InitialNodeGroupBackoffDuration).Add(time.Second)
	assert.True(t, clusterstate.IsClusterHealthy())
	assert.True(t, clusterstate.IsNodeGroupHealthy("ng1"))
	assert.True(t, clusterstate.IsNodeGroupSafeToScaleUp("ng1", now))

	// Another failed scale up should cause longer backoff
	clusterstate.RegisterScaleUp(&ScaleUpRequest{
		NodeGroupName:   "ng1",
		Increase:        1,
		Time:            now.Add(-2 * time.Minute),
		ExpectedAddTime: now.Add(-1 * time.Second),
	})
	err = clusterstate.UpdateNodes([]*apiv1.Node{ng1_1, ng1_2, ng1_3}, now)
	assert.NoError(t, err)
	assert.True(t, clusterstate.IsClusterHealthy())
	assert.True(t, clusterstate.IsNodeGroupHealthy("ng1"))
	assert.False(t, clusterstate.IsNodeGroupSafeToScaleUp("ng1", now))

	now = now.Add(InitialNodeGroupBackoffDuration).Add(time.Second)
	assert.False(t, clusterstate.IsNodeGroupSafeToScaleUp("ng1", now))

	// The backoff should be cleared after a successful scale-up
	clusterstate.RegisterScaleUp(&ScaleUpRequest{
		NodeGroupName:   "ng1",
		Increase:        1,
		Time:            now,
		ExpectedAddTime: now.Add(time.Second),
	})
	ng1_4 := BuildTestNode("ng1-4", 1000, 1000)
	SetNodeReadyState(ng1_4, true, now.Add(-1*time.Minute))
	provider.AddNode("ng1", ng1_4)
	err = clusterstate.UpdateNodes([]*apiv1.Node{ng1_1, ng1_2, ng1_3, ng1_4}, now)
	assert.NoError(t, err)
	assert.True(t, clusterstate.IsClusterHealthy())
	assert.True(t, clusterstate.IsNodeGroupHealthy("ng1"))
	assert.True(t, clusterstate.IsNodeGroupSafeToScaleUp("ng1", now))
	_, found := clusterstate.nodeGroupBackoffInfo["ng1"]
	assert.False(t, found)
}
