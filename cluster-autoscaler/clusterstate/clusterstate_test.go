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
	"fmt"
	"testing"
	"time"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate/api"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate/utils"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroupconfig"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroups/asyncnodegroups"

	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/client-go/kubernetes/fake"
	kube_record "k8s.io/client-go/tools/record"

	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/cluster-autoscaler/utils/backoff"
)

// GetCloudProviderDeletedNodeNames returns a list of the names of nodes removed
// from cloud provider but registered in Kubernetes.
func GetCloudProviderDeletedNodeNames(csr *ClusterStateRegistry) []string {
	csr.Lock()
	defer csr.Unlock()

	result := make([]string, 0, len(csr.deletedNodes))
	for nodeName := range csr.deletedNodes {
		result = append(result, nodeName)
	}
	return result
}

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
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false, "my-cool-configmap")
	clusterstate := NewClusterStateRegistry(provider, ClusterStateRegistryConfig{
		MaxTotalUnreadyPercentage: 10,
		OkTotalUnreadyCount:       1,
	}, fakeLogRecorder, newBackoff(), nodegroupconfig.NewDefaultNodeGroupConfigProcessor(config.NodeGroupAutoscalingOptions{MaxNodeProvisionTime: time.Minute}), asyncnodegroups.NewDefaultAsyncNodeGroupStateChecker())
	clusterstate.RegisterScaleUp(provider.GetNodeGroup("ng1"), 4, time.Now())
	err := clusterstate.UpdateNodes([]*apiv1.Node{ng1_1, ng2_1}, nil, now)
	assert.NoError(t, err)
	assert.True(t, clusterstate.IsClusterHealthy())
	assert.Empty(t, clusterstate.GetScaleUpFailures())

	status := clusterstate.GetStatus(now)
	assert.Equal(t, api.ClusterAutoscalerInProgress, status.ClusterWide.ScaleUp.Status)
	assert.Equal(t, 2, len(status.NodeGroups))
	ng1Checked := false
	ng2Checked := true
	for _, nodeGroupStatus := range status.NodeGroups {
		if nodeGroupStatus.Name == "ng1" {
			assert.Equal(t, api.ClusterAutoscalerInProgress, nodeGroupStatus.ScaleUp.Status)
			ng1Checked = true
		}
		if nodeGroupStatus.Name == "ng2" {
			assert.Equal(t, api.ClusterAutoscalerNoActivity, nodeGroupStatus.ScaleUp.Status)
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
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false, "my-cool-configmap")
	clusterstate := NewClusterStateRegistry(provider, ClusterStateRegistryConfig{
		MaxTotalUnreadyPercentage: 10,
		OkTotalUnreadyCount:       1,
	}, fakeLogRecorder, newBackoff(), nodegroupconfig.NewDefaultNodeGroupConfigProcessor(config.NodeGroupAutoscalingOptions{MaxNodeProvisionTime: time.Minute}), asyncnodegroups.NewDefaultAsyncNodeGroupStateChecker())
	err := clusterstate.UpdateNodes([]*apiv1.Node{}, nil, now.Add(-5*time.Second))
	assert.NoError(t, err)
	assert.True(t, clusterstate.IsClusterHealthy())
	assert.Empty(t, clusterstate.GetScaleUpFailures())
	assert.True(t, clusterstate.IsNodeGroupHealthy("ng1"))
	assert.False(t, clusterstate.IsNodeGroupScalingUp("ng1"))
	assert.False(t, clusterstate.HasNodeGroupStartedScaleUp("ng1"))

	provider.AddNodeGroup("ng1", 0, 10, 3)
	clusterstate.RegisterScaleUp(provider.GetNodeGroup("ng1"), 3, now.Add(-3*time.Second))
	//	clusterstate.scaleUpRequests["ng1"].Time = now.Add(-3 * time.Second)
	//	clusterstate.scaleUpRequests["ng1"].ExpectedAddTime = now.Add(1 * time.Minute)
	err = clusterstate.UpdateNodes([]*apiv1.Node{}, nil, now)

	assert.NoError(t, err)
	assert.True(t, clusterstate.IsClusterHealthy())
	assert.True(t, clusterstate.IsNodeGroupHealthy("ng1"))
	assert.True(t, clusterstate.IsNodeGroupScalingUp("ng1"))
	assert.True(t, clusterstate.HasNodeGroupStartedScaleUp("ng1"))
}

func TestHasNodeGroupStartedScaleUp(t *testing.T) {
	tests := map[string]struct {
		initialSize int
		delta       int
	}{
		"Target size reverts back to zero": {
			initialSize: 0,
			delta:       3,
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			now := time.Now()
			provider := testprovider.NewTestCloudProvider(nil, nil)
			provider.AddNodeGroup("ng1", 0, 5, tc.initialSize)
			fakeClient := &fake.Clientset{}
			fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false, "my-cool-configmap")
			clusterstate := NewClusterStateRegistry(provider, ClusterStateRegistryConfig{
				MaxTotalUnreadyPercentage: 10,
				OkTotalUnreadyCount:       1,
			}, fakeLogRecorder, newBackoff(), nodegroupconfig.NewDefaultNodeGroupConfigProcessor(config.NodeGroupAutoscalingOptions{MaxNodeProvisionTime: time.Minute}), asyncnodegroups.NewDefaultAsyncNodeGroupStateChecker())
			err := clusterstate.UpdateNodes([]*apiv1.Node{}, nil, now.Add(-5*time.Second))
			assert.NoError(t, err)
			assert.False(t, clusterstate.IsNodeGroupScalingUp("ng1"))
			assert.False(t, clusterstate.HasNodeGroupStartedScaleUp("ng1"))

			provider.AddNodeGroup("ng1", 0, 5, tc.initialSize+tc.delta)
			clusterstate.RegisterScaleUp(provider.GetNodeGroup("ng1"), tc.delta, now.Add(-3*time.Second))
			err = clusterstate.UpdateNodes([]*apiv1.Node{}, nil, now)
			assert.NoError(t, err)
			assert.True(t, clusterstate.IsNodeGroupScalingUp("ng1"))
			assert.True(t, clusterstate.HasNodeGroupStartedScaleUp("ng1"))

			provider.AddNodeGroup("ng1", 0, 5, tc.initialSize)
			clusterstate.Recalculate()
			assert.False(t, clusterstate.IsNodeGroupScalingUp("ng1"))
			assert.True(t, clusterstate.HasNodeGroupStartedScaleUp("ng1"))
		})
	}
}

// TestRecalculateStateAfterNodeGroupSizeChanged checks that Recalculate updates state correctly after
// some node group size changed. We verify that acceptable ranges are updated accordingly
// and that the UpcomingNodes reflect the node group size change (important for recalculating state after
// deleting scale-up nodes that failed to create).
func TestRecalculateStateAfterNodeGroupSizeChanged(t *testing.T) {
	ngName := "ng1"
	testCases := []struct {
		name                string
		acceptableRange     AcceptableRange
		readiness           Readiness
		newTarget           int
		scaleUpRequest      *ScaleUpRequest
		wantAcceptableRange AcceptableRange
		wantUpcoming        int
	}{
		{
			name:                "failed scale up by 3 nodes",
			acceptableRange:     AcceptableRange{MinNodes: 1, CurrentTarget: 4, MaxNodes: 4},
			readiness:           Readiness{Ready: make([]string, 1)},
			newTarget:           1,
			wantAcceptableRange: AcceptableRange{MinNodes: 1, CurrentTarget: 1, MaxNodes: 1},
			wantUpcoming:        0,
		}, {
			name:                "partially failed scale up",
			acceptableRange:     AcceptableRange{MinNodes: 5, CurrentTarget: 7, MaxNodes: 8},
			readiness:           Readiness{Ready: make([]string, 5)},
			newTarget:           6,
			wantAcceptableRange: AcceptableRange{MinNodes: 5, CurrentTarget: 6, MaxNodes: 6},
			scaleUpRequest:      &ScaleUpRequest{Increase: 1},
			wantUpcoming:        1,
		}, {
			name:                "scale up ongoing, no change",
			acceptableRange:     AcceptableRange{MinNodes: 1, CurrentTarget: 4, MaxNodes: 4},
			readiness:           Readiness{Ready: make([]string, 1)},
			newTarget:           4,
			wantAcceptableRange: AcceptableRange{MinNodes: 1, CurrentTarget: 4, MaxNodes: 4},
			scaleUpRequest:      &ScaleUpRequest{Increase: 3},
			wantUpcoming:        3,
		}, {
			name:                "no scale up, no change",
			acceptableRange:     AcceptableRange{MinNodes: 4, CurrentTarget: 4, MaxNodes: 4},
			readiness:           Readiness{Ready: make([]string, 4)},
			newTarget:           4,
			wantAcceptableRange: AcceptableRange{MinNodes: 4, CurrentTarget: 4, MaxNodes: 4},
			wantUpcoming:        0,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := testprovider.NewTestCloudProvider(nil, nil)
			provider.AddNodeGroup(ngName, 0, 1000, tc.newTarget)

			fakeLogRecorder, _ := utils.NewStatusMapRecorder(&fake.Clientset{}, "kube-system", kube_record.NewFakeRecorder(5), false, "my-cool-configmap")
			clusterState := NewClusterStateRegistry(provider, ClusterStateRegistryConfig{}, fakeLogRecorder,
				newBackoff(), nodegroupconfig.NewDefaultNodeGroupConfigProcessor(config.NodeGroupAutoscalingOptions{}), asyncnodegroups.NewDefaultAsyncNodeGroupStateChecker())
			clusterState.acceptableRanges = map[string]AcceptableRange{ngName: tc.acceptableRange}
			clusterState.perNodeGroupReadiness = map[string]Readiness{ngName: tc.readiness}
			if tc.scaleUpRequest != nil {
				clusterState.scaleUpRequests = map[string]*ScaleUpRequest{ngName: tc.scaleUpRequest}
			}

			clusterState.Recalculate()
			assert.Equal(t, tc.wantAcceptableRange, clusterState.acceptableRanges[ngName])
			upcomingCounts, _ := clusterState.GetUpcomingNodes()
			if upcoming, found := upcomingCounts[ngName]; found {
				assert.Equal(t, tc.wantUpcoming, upcoming, "Unexpected upcoming nodes count, want: %d got: %d", tc.wantUpcoming, upcomingCounts[ngName])
			}
		})
	}
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
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false, "my-cool-configmap")
	clusterstate := NewClusterStateRegistry(provider, ClusterStateRegistryConfig{
		MaxTotalUnreadyPercentage: 10,
		OkTotalUnreadyCount:       1,
	}, fakeLogRecorder, newBackoff(), nodegroupconfig.NewDefaultNodeGroupConfigProcessor(config.NodeGroupAutoscalingOptions{MaxNodeProvisionTime: 15 * time.Minute}), asyncnodegroups.NewDefaultAsyncNodeGroupStateChecker())
	err := clusterstate.UpdateNodes([]*apiv1.Node{ng1_1, ng2_1}, nil, now)
	assert.NoError(t, err)
	assert.True(t, clusterstate.IsClusterHealthy())
	assert.Empty(t, clusterstate.GetScaleUpFailures())
	assert.True(t, clusterstate.IsNodeGroupHealthy("ng1"))

	status := clusterstate.GetStatus(now)
	assert.Equal(t, api.ClusterAutoscalerHealthy, status.ClusterWide.Health.Status)
	assert.Equal(t, api.ClusterAutoscalerNoActivity, status.ClusterWide.ScaleUp.Status)

	assert.Equal(t, 2, len(status.NodeGroups))
	ng1Checked := false
	for _, nodeGroupStatus := range status.NodeGroups {
		if nodeGroupStatus.Name == "ng1" {
			assert.Equal(t, api.ClusterAutoscalerHealthy, nodeGroupStatus.Health.Status)
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
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false, "my-cool-configmap")
	clusterstate := NewClusterStateRegistry(provider, ClusterStateRegistryConfig{
		MaxTotalUnreadyPercentage: 10,
		OkTotalUnreadyCount:       1,
	}, fakeLogRecorder, newBackoff(), nodegroupconfig.NewDefaultNodeGroupConfigProcessor(config.NodeGroupAutoscalingOptions{MaxNodeProvisionTime: 15 * time.Minute}), asyncnodegroups.NewDefaultAsyncNodeGroupStateChecker())
	err := clusterstate.UpdateNodes([]*apiv1.Node{noNgNode}, nil, now)
	assert.NoError(t, err)
	assert.Empty(t, clusterstate.GetScaleUpFailures())
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
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false, "my-cool-configmap")
	clusterstate := NewClusterStateRegistry(provider, ClusterStateRegistryConfig{
		MaxTotalUnreadyPercentage: 10,
		OkTotalUnreadyCount:       1,
	}, fakeLogRecorder, newBackoff(), nodegroupconfig.NewDefaultNodeGroupConfigProcessor(config.NodeGroupAutoscalingOptions{MaxNodeProvisionTime: 15 * time.Minute}), asyncnodegroups.NewDefaultAsyncNodeGroupStateChecker())
	err := clusterstate.UpdateNodes([]*apiv1.Node{ng1_1, ng2_1}, nil, now)
	clusterstate.UpdateScaleDownCandidates([]*apiv1.Node{ng1_1}, now)

	assert.NoError(t, err)
	assert.True(t, clusterstate.IsClusterHealthy())
	assert.Empty(t, clusterstate.GetScaleUpFailures())
	assert.True(t, clusterstate.IsNodeGroupHealthy("ng1"))

	status := clusterstate.GetStatus(now)
	assert.Equal(t, api.ClusterAutoscalerHealthy, status.ClusterWide.Health.Status)
	assert.Equal(t, api.ClusterAutoscalerNoActivity, status.ClusterWide.ScaleUp.Status)
	assert.Equal(t, api.ClusterAutoscalerCandidatesPresent, status.ClusterWide.ScaleDown.Status)

	assert.Equal(t, 2, len(status.NodeGroups))
	ng1Checked := false
	ng2Checked := false
	for _, nodeGroupStatus := range status.NodeGroups {
		if nodeGroupStatus.Name == "ng1" {
			assert.Equal(t, api.ClusterAutoscalerHealthy, nodeGroupStatus.Health.Status)
			assert.Equal(t, api.ClusterAutoscalerCandidatesPresent, nodeGroupStatus.ScaleDown.Status)
			ng1Checked = true
		}
		if nodeGroupStatus.Name == "ng2" {
			assert.Equal(t, api.ClusterAutoscalerHealthy, nodeGroupStatus.Health.Status)
			assert.Equal(t, api.ClusterAutoscalerNoCandidates, nodeGroupStatus.ScaleDown.Status)
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
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false, "my-cool-configmap")
	clusterstate := NewClusterStateRegistry(provider, ClusterStateRegistryConfig{
		MaxTotalUnreadyPercentage: 10,
		OkTotalUnreadyCount:       1,
	}, fakeLogRecorder, newBackoff(), nodegroupconfig.NewDefaultNodeGroupConfigProcessor(config.NodeGroupAutoscalingOptions{MaxNodeProvisionTime: 15 * time.Minute}), asyncnodegroups.NewDefaultAsyncNodeGroupStateChecker())
	err := clusterstate.UpdateNodes([]*apiv1.Node{ng1_1, ng2_1}, nil, now)
	assert.NoError(t, err)
	assert.True(t, clusterstate.IsClusterHealthy())
	assert.Empty(t, clusterstate.GetScaleUpFailures())
	assert.False(t, clusterstate.IsNodeGroupHealthy("ng1"))

	status := clusterstate.GetStatus(now)
	assert.Equal(t, api.ClusterAutoscalerHealthy, status.ClusterWide.Health.Status)
	assert.Equal(t, 2, len(status.NodeGroups))
	ng1Checked := false
	for _, nodeGroupStatus := range status.NodeGroups {
		if nodeGroupStatus.Name == "ng1" {
			assert.Equal(t, api.ClusterAutoscalerUnhealthy, nodeGroupStatus.Health.Status)
			ng1Checked = true
		}
	}
	assert.True(t, ng1Checked)
}

func TestTooManyUnready(t *testing.T) {
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
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false, "my-cool-configmap")
	clusterstate := NewClusterStateRegistry(provider, ClusterStateRegistryConfig{
		MaxTotalUnreadyPercentage: 10,
		OkTotalUnreadyCount:       1,
	}, fakeLogRecorder, newBackoff(), nodegroupconfig.NewDefaultNodeGroupConfigProcessor(config.NodeGroupAutoscalingOptions{MaxNodeProvisionTime: 15 * time.Minute}), asyncnodegroups.NewDefaultAsyncNodeGroupStateChecker())
	err := clusterstate.UpdateNodes([]*apiv1.Node{ng1_1, ng2_1}, nil, now)
	assert.NoError(t, err)
	assert.False(t, clusterstate.IsClusterHealthy())
	assert.Empty(t, clusterstate.GetScaleUpFailures())
	assert.True(t, clusterstate.IsNodeGroupHealthy("ng1"))
}

func TestUnreadyLongAfterCreation(t *testing.T) {
	now := time.Now()

	ng1_1 := BuildTestNode("ng1-1", 1000, 1000)
	SetNodeReadyState(ng1_1, true, now.Add(-time.Minute))
	ng2_1 := BuildTestNode("ng2-1", 1000, 1000)
	SetNodeReadyState(ng2_1, false, now.Add(-time.Minute))
	ng2_1.CreationTimestamp = metav1.Time{Time: now.Add(-30 * time.Minute)}

	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 1, 10, 1)
	provider.AddNodeGroup("ng2", 1, 10, 1)
	provider.AddNode("ng1", ng1_1)
	provider.AddNode("ng2", ng2_1)

	assert.NotNil(t, provider)
	fakeClient := &fake.Clientset{}
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false, "some-map")
	clusterstate := NewClusterStateRegistry(provider, ClusterStateRegistryConfig{
		MaxTotalUnreadyPercentage: 10,
		OkTotalUnreadyCount:       1,
	}, fakeLogRecorder, newBackoff(), nodegroupconfig.NewDefaultNodeGroupConfigProcessor(config.NodeGroupAutoscalingOptions{MaxNodeProvisionTime: 15 * time.Minute}), asyncnodegroups.NewDefaultAsyncNodeGroupStateChecker())
	err := clusterstate.UpdateNodes([]*apiv1.Node{ng1_1, ng2_1}, nil, now)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(clusterstate.GetClusterReadiness().Unready))
	assert.Equal(t, 0, len(clusterstate.GetClusterReadiness().NotStarted))
	upcoming, upcomingRegistered := clusterstate.GetUpcomingNodes()
	assert.Equal(t, 0, upcoming["ng1"])
	assert.Empty(t, upcomingRegistered["ng1"])
}

func TestNotStarted(t *testing.T) {
	now := time.Now()

	ng1_1 := BuildTestNode("ng1-1", 1000, 1000)
	SetNodeReadyState(ng1_1, true, now.Add(-time.Minute))
	ng2_1 := BuildTestNode("ng2-1", 1000, 1000)
	SetNodeReadyState(ng2_1, false, now.Add(-4*time.Minute))
	SetNodeNotReadyTaint(ng2_1)
	ng2_1.CreationTimestamp = metav1.Time{Time: now.Add(-10 * time.Minute)}

	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 1, 10, 1)
	provider.AddNodeGroup("ng2", 1, 10, 1)
	provider.AddNode("ng1", ng1_1)
	provider.AddNode("ng2", ng2_1)

	assert.NotNil(t, provider)
	fakeClient := &fake.Clientset{}
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false, "some-map")
	clusterstate := NewClusterStateRegistry(provider, ClusterStateRegistryConfig{
		MaxTotalUnreadyPercentage: 10,
		OkTotalUnreadyCount:       1,
	}, fakeLogRecorder, newBackoff(), nodegroupconfig.NewDefaultNodeGroupConfigProcessor(config.NodeGroupAutoscalingOptions{MaxNodeProvisionTime: 15 * time.Minute}), asyncnodegroups.NewDefaultAsyncNodeGroupStateChecker())
	err := clusterstate.UpdateNodes([]*apiv1.Node{ng1_1, ng2_1}, nil, now)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(clusterstate.GetClusterReadiness().NotStarted))
	assert.Equal(t, 1, len(clusterstate.GetClusterReadiness().Ready))

	// node ng2_1 moves condition to ready
	SetNodeReadyState(ng2_1, true, now.Add(-4*time.Minute))
	err = clusterstate.UpdateNodes([]*apiv1.Node{ng1_1, ng2_1}, nil, now)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(clusterstate.GetClusterReadiness().NotStarted))
	assert.Equal(t, 1, len(clusterstate.GetClusterReadiness().Ready))

	// node ng2_1 no longer has the taint
	RemoveNodeNotReadyTaint(ng2_1)
	err = clusterstate.UpdateNodes([]*apiv1.Node{ng1_1, ng2_1}, nil, now)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(clusterstate.GetClusterReadiness().NotStarted))
	assert.Equal(t, 2, len(clusterstate.GetClusterReadiness().Ready))
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
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false, "my-cool-configmap")
	clusterstate := NewClusterStateRegistry(provider, ClusterStateRegistryConfig{
		MaxTotalUnreadyPercentage: 10,
		OkTotalUnreadyCount:       1,
	}, fakeLogRecorder, newBackoff(), nodegroupconfig.NewDefaultNodeGroupConfigProcessor(config.NodeGroupAutoscalingOptions{MaxNodeProvisionTime: 2 * time.Minute}), asyncnodegroups.NewDefaultAsyncNodeGroupStateChecker())
	clusterstate.RegisterScaleUp(provider.GetNodeGroup("ng1"), 4, now.Add(-3*time.Minute))
	err := clusterstate.UpdateNodes([]*apiv1.Node{ng1_1}, nil, now)
	assert.NoError(t, err)
	assert.True(t, clusterstate.IsClusterHealthy())
	assert.False(t, clusterstate.IsNodeGroupHealthy("ng1"))
	assert.Equal(t, clusterstate.GetScaleUpFailures(), map[string][]ScaleUpFailure{
		"ng1": {
			{NodeGroup: provider.GetNodeGroup("ng1"), Time: now, Reason: metrics.Timeout},
		},
	})
}

func TestRegisterScaleDown(t *testing.T) {
	ng1_1 := BuildTestNode("ng1-1", 1000, 1000)
	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 1, 10, 1)
	provider.AddNode("ng1", ng1_1)
	assert.NotNil(t, provider)

	fakeClient := &fake.Clientset{}
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false, "my-cool-configmap")
	clusterstate := NewClusterStateRegistry(provider, ClusterStateRegistryConfig{
		MaxTotalUnreadyPercentage: 10,
		OkTotalUnreadyCount:       1,
	}, fakeLogRecorder, newBackoff(), nodegroupconfig.NewDefaultNodeGroupConfigProcessor(config.NodeGroupAutoscalingOptions{MaxNodeProvisionTime: 15 * time.Minute}), asyncnodegroups.NewDefaultAsyncNodeGroupStateChecker())
	now := time.Now()
	clusterstate.RegisterScaleDown(provider.GetNodeGroup("ng1"), "ng1-1", now.Add(time.Minute), now)
	assert.Equal(t, 1, len(clusterstate.scaleDownRequests))
	clusterstate.updateScaleRequests(now.Add(5 * time.Minute))
	assert.Equal(t, 0, len(clusterstate.scaleDownRequests))
	assert.Empty(t, clusterstate.GetScaleUpFailures())
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

	// One node is already there, for a second nde deletion / draining was already started.
	ng5_1 := BuildTestNode("ng5-1", 1000, 1000)
	SetNodeReadyState(ng5_1, true, now.Add(-time.Minute))
	ng5_2 := BuildTestNode("ng5-2", 1000, 1000)
	SetNodeReadyState(ng5_2, true, now.Add(-time.Minute))
	ng5_2.Spec.Taints = []apiv1.Taint{
		{
			Key:    taints.ToBeDeletedTaint,
			Value:  fmt.Sprint(time.Now().Unix()),
			Effect: apiv1.TaintEffectNoSchedule,
		},
	}
	provider.AddNodeGroup("ng5", 1, 10, 2)
	provider.AddNode("ng5", ng5_1)
	provider.AddNode("ng5", ng5_2)

	assert.NotNil(t, provider)
	fakeClient := &fake.Clientset{}
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false, "my-cool-configmap")
	clusterstate := NewClusterStateRegistry(provider, ClusterStateRegistryConfig{
		MaxTotalUnreadyPercentage: 10,
		OkTotalUnreadyCount:       1,
	}, fakeLogRecorder, newBackoff(), nodegroupconfig.NewDefaultNodeGroupConfigProcessor(config.NodeGroupAutoscalingOptions{MaxNodeProvisionTime: 15 * time.Minute}), asyncnodegroups.NewDefaultAsyncNodeGroupStateChecker())
	err := clusterstate.UpdateNodes([]*apiv1.Node{ng1_1, ng2_1, ng3_1, ng4_1, ng5_1, ng5_2}, nil, now)
	assert.NoError(t, err)
	assert.Empty(t, clusterstate.GetScaleUpFailures())

	upcomingNodes, upcomingRegistered := clusterstate.GetUpcomingNodes()
	assert.Equal(t, 6, upcomingNodes["ng1"])
	assert.Empty(t, upcomingRegistered["ng1"]) // Only unregistered.
	assert.Equal(t, 1, upcomingNodes["ng2"])
	assert.Empty(t, upcomingRegistered["ng2"]) // Only unregistered.
	assert.Equal(t, 2, upcomingNodes["ng3"])
	assert.Equal(t, []string{"ng3-1"}, upcomingRegistered["ng3"]) // 1 registered, 1 unregistered.
	assert.NotContains(t, upcomingNodes, "ng4")
	assert.NotContains(t, upcomingRegistered, "ng4")
	assert.Equal(t, 0, upcomingNodes["ng5"])
	assert.Empty(t, upcomingRegistered["ng5"])
}

func TestTaintBasedNodeDeletion(t *testing.T) {
	// Create a new Cloud Provider that does not implement the HasInstance check
	// it will return the ErrNotImplemented error instead.
	provider := testprovider.NewTestNodeDeletionDetectionCloudProvider(nil, nil,
		func(string) (bool, error) { return false, cloudprovider.ErrNotImplemented })
	now := time.Now()

	// One node is already there, for a second nde deletion / draining was already started.
	ng1_1 := BuildTestNode("ng1-1", 1000, 1000)
	SetNodeReadyState(ng1_1, true, now.Add(-time.Minute))
	ng1_2 := BuildTestNode("ng1-2", 1000, 1000)
	SetNodeReadyState(ng1_2, true, now.Add(-time.Minute))
	ng1_2.Spec.Taints = []apiv1.Taint{
		{
			Key:    taints.ToBeDeletedTaint,
			Value:  fmt.Sprint(time.Now().Unix()),
			Effect: apiv1.TaintEffectNoSchedule,
		},
	}
	provider.AddNodeGroup("ng1", 1, 10, 2)
	provider.AddNode("ng1", ng1_1)
	provider.AddNode("ng1", ng1_2)

	assert.NotNil(t, provider)
	fakeClient := &fake.Clientset{}
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false, "my-cool-configmap")
	clusterstate := NewClusterStateRegistry(provider, ClusterStateRegistryConfig{
		MaxTotalUnreadyPercentage: 10,
		OkTotalUnreadyCount:       1,
	}, fakeLogRecorder, newBackoff(), nodegroupconfig.NewDefaultNodeGroupConfigProcessor(config.NodeGroupAutoscalingOptions{MaxNodeProvisionTime: 15 * time.Minute}), asyncnodegroups.NewDefaultAsyncNodeGroupStateChecker())
	err := clusterstate.UpdateNodes([]*apiv1.Node{ng1_1, ng1_2}, nil, now)
	assert.NoError(t, err)
	assert.Empty(t, clusterstate.GetScaleUpFailures())

	upcomingNodes, upcomingRegistered := clusterstate.GetUpcomingNodes()
	assert.Equal(t, 1, upcomingNodes["ng1"])
	assert.Empty(t, upcomingRegistered["ng1"]) // Only unregistered.
}

func TestIncorrectSize(t *testing.T) {
	ng1_1 := BuildTestNode("ng1-1", 1000, 1000)
	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 1, 10, 5)
	provider.AddNode("ng1", ng1_1)
	assert.NotNil(t, provider)
	fakeClient := &fake.Clientset{}
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false, "my-cool-configmap")
	clusterstate := NewClusterStateRegistry(provider, ClusterStateRegistryConfig{
		MaxTotalUnreadyPercentage: 10,
		OkTotalUnreadyCount:       1,
	}, fakeLogRecorder, newBackoff(), nodegroupconfig.NewDefaultNodeGroupConfigProcessor(config.NodeGroupAutoscalingOptions{MaxNodeProvisionTime: 15 * time.Minute}), asyncnodegroups.NewDefaultAsyncNodeGroupStateChecker())
	now := time.Now()
	clusterstate.UpdateNodes([]*apiv1.Node{ng1_1}, nil, now.Add(-5*time.Minute))
	incorrect := clusterstate.incorrectNodeGroupSizes["ng1"]
	assert.Equal(t, 5, incorrect.ExpectedSize)
	assert.Equal(t, 1, incorrect.CurrentSize)
	assert.Equal(t, now.Add(-5*time.Minute), incorrect.FirstObserved)

	clusterstate.UpdateNodes([]*apiv1.Node{ng1_1}, nil, now.Add(-4*time.Minute))
	incorrect = clusterstate.incorrectNodeGroupSizes["ng1"]
	assert.Equal(t, 5, incorrect.ExpectedSize)
	assert.Equal(t, 1, incorrect.CurrentSize)
	assert.Equal(t, now.Add(-5*time.Minute), incorrect.FirstObserved)

	clusterstate.UpdateNodes([]*apiv1.Node{ng1_1, ng1_1}, nil, now.Add(-3*time.Minute))
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
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false, "my-cool-configmap")
	clusterstate := NewClusterStateRegistry(provider, ClusterStateRegistryConfig{
		MaxTotalUnreadyPercentage: 10,
		OkTotalUnreadyCount:       1,
	}, fakeLogRecorder, newBackoff(), nodegroupconfig.NewDefaultNodeGroupConfigProcessor(config.NodeGroupAutoscalingOptions{MaxNodeProvisionTime: 10 * time.Second}), asyncnodegroups.NewDefaultAsyncNodeGroupStateChecker())
	err := clusterstate.UpdateNodes([]*apiv1.Node{ng1_1}, nil, time.Now().Add(-time.Minute))

	assert.NoError(t, err)
	assert.Equal(t, 1, len(clusterstate.GetUnregisteredNodes()))
	assert.Equal(t, "ng1-2", clusterstate.GetUnregisteredNodes()[0].Node.Name)
	upcomingNodes, upcomingRegistered := clusterstate.GetUpcomingNodes()
	assert.Equal(t, 1, upcomingNodes["ng1"])
	assert.Empty(t, upcomingRegistered["ng1"]) // Unregistered only.

	// The node didn't come up in MaxNodeProvisionTime, it should no longer be
	// counted as upcoming (but it is still an unregistered node)
	err = clusterstate.UpdateNodes([]*apiv1.Node{ng1_1}, nil, time.Now().Add(time.Minute))
	assert.NoError(t, err)
	assert.Equal(t, 1, len(clusterstate.GetUnregisteredNodes()))
	assert.Equal(t, "ng1-2", clusterstate.GetUnregisteredNodes()[0].Node.Name)
	upcomingNodes, upcomingRegistered = clusterstate.GetUpcomingNodes()
	assert.Equal(t, 0, len(upcomingNodes))
	assert.Empty(t, upcomingRegistered["ng1"])

	err = clusterstate.UpdateNodes([]*apiv1.Node{ng1_1, ng1_2}, nil, time.Now().Add(time.Minute))
	assert.NoError(t, err)
	assert.Equal(t, 0, len(clusterstate.GetUnregisteredNodes()))
}

func TestCloudProviderDeletedNodes(t *testing.T) {
	now := time.Now()
	ng1_1 := BuildTestNode("ng1-1", 1000, 1000)
	SetNodeReadyState(ng1_1, true, now.Add(-time.Minute))
	ng1_1.Spec.ProviderID = "ng1-1"
	ng1_2 := BuildTestNode("ng1-2", 1000, 1000)
	SetNodeReadyState(ng1_2, true, now.Add(-time.Minute))
	ng1_2.Spec.ProviderID = "ng1-2"
	// No Node Group - Not Autoscaled Node
	noNgNode := BuildTestNode("no-ng", 1000, 1000)
	SetNodeReadyState(noNgNode, true, now.Add(-time.Minute))

	noNgNode.Spec.ProviderID = "no-ng"
	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 1, 10, 2)
	provider.AddNode("ng1", ng1_1)
	provider.AddNode("ng1", ng1_2)
	provider.AddNode("no_ng", noNgNode)

	fakeClient := &fake.Clientset{}
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false, "my-cool-configmap")
	clusterstate := NewClusterStateRegistry(provider, ClusterStateRegistryConfig{
		MaxTotalUnreadyPercentage: 10,
		OkTotalUnreadyCount:       1,
	}, fakeLogRecorder, newBackoff(), nodegroupconfig.NewDefaultNodeGroupConfigProcessor(config.NodeGroupAutoscalingOptions{MaxNodeProvisionTime: 10 * time.Second}), asyncnodegroups.NewDefaultAsyncNodeGroupStateChecker())
	now.Add(time.Minute)
	err := clusterstate.UpdateNodes([]*apiv1.Node{ng1_1, ng1_2, noNgNode}, nil, now)

	// Nodes are registered correctly between Kubernetes and cloud provider.
	assert.NoError(t, err)
	assert.Equal(t, 0, len(GetCloudProviderDeletedNodeNames(clusterstate)))

	// The node was removed from Cloud Provider
	// should be counted as Deleted by cluster state
	nodeGroup, err := provider.NodeGroupForNode(ng1_2)
	assert.NoError(t, err)
	provider.DeleteNode(ng1_2)
	clusterstate.InvalidateNodeInstancesCacheEntry(nodeGroup)
	now.Add(time.Minute)

	err = clusterstate.UpdateNodes([]*apiv1.Node{ng1_1, ng1_2, noNgNode}, nil, now)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(GetCloudProviderDeletedNodeNames(clusterstate)))
	assert.Equal(t, "ng1-2", GetCloudProviderDeletedNodeNames(clusterstate)[0])
	assert.Equal(t, 1, len(clusterstate.GetClusterReadiness().Deleted))

	// The node is removed from Kubernetes
	now.Add(time.Minute)

	err = clusterstate.UpdateNodes([]*apiv1.Node{ng1_1, noNgNode}, nil, now)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(GetCloudProviderDeletedNodeNames(clusterstate)))

	// New Node is added afterwards
	ng1_3 := BuildTestNode("ng1-3", 1000, 1000)
	SetNodeReadyState(ng1_3, true, now.Add(-time.Minute))
	ng1_3.Spec.ProviderID = "ng1-3"
	provider.AddNode("ng1", ng1_3)
	clusterstate.InvalidateNodeInstancesCacheEntry(nodeGroup)
	now.Add(time.Minute)

	err = clusterstate.UpdateNodes([]*apiv1.Node{ng1_1, ng1_3, noNgNode}, nil, now)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(GetCloudProviderDeletedNodeNames(clusterstate)))

	// Newly added node is removed from Cloud Provider
	// should be counted as Deleted by cluster state
	nodeGroup, err = provider.NodeGroupForNode(ng1_3)
	assert.NoError(t, err)
	provider.DeleteNode(ng1_3)
	clusterstate.InvalidateNodeInstancesCacheEntry(nodeGroup)
	now.Add(time.Minute)

	err = clusterstate.UpdateNodes([]*apiv1.Node{ng1_1, noNgNode, ng1_3}, nil, now)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(GetCloudProviderDeletedNodeNames(clusterstate)))
	assert.Equal(t, "ng1-3", GetCloudProviderDeletedNodeNames(clusterstate)[0])
	assert.Equal(t, 1, len(clusterstate.GetClusterReadiness().Deleted))

	// Confirm that previously identified deleted Cloud Provider nodes are still included
	// until it is removed from Kubernetes
	now.Add(time.Minute)

	err = clusterstate.UpdateNodes([]*apiv1.Node{ng1_1, noNgNode, ng1_3}, nil, now)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(GetCloudProviderDeletedNodeNames(clusterstate)))
	assert.Equal(t, "ng1-3", GetCloudProviderDeletedNodeNames(clusterstate)[0])
	assert.Equal(t, 1, len(clusterstate.GetClusterReadiness().Deleted))

	// The node is removed from Kubernetes
	now.Add(time.Minute)

	err = clusterstate.UpdateNodes([]*apiv1.Node{ng1_1, noNgNode}, nil, now)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(GetCloudProviderDeletedNodeNames(clusterstate)))
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
	ng1 := provider.GetNodeGroup("ng1")
	provider.AddNode("ng1", ng1_1)
	provider.AddNode("ng1", ng1_2)
	provider.AddNode("ng1", ng1_3)
	assert.NotNil(t, provider)

	fakeClient := &fake.Clientset{}
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false, "my-cool-configmap")
	clusterstate := NewClusterStateRegistry(
		provider, ClusterStateRegistryConfig{
			MaxTotalUnreadyPercentage: 10,
			OkTotalUnreadyCount:       1,
		}, fakeLogRecorder, newBackoff(), nodegroupconfig.NewDefaultNodeGroupConfigProcessor(config.NodeGroupAutoscalingOptions{MaxNodeProvisionTime: 120 * time.Second}),
		asyncnodegroups.NewDefaultAsyncNodeGroupStateChecker())

	// After failed scale-up, node group should be still healthy, but should backoff from scale-ups
	clusterstate.RegisterScaleUp(provider.GetNodeGroup("ng1"), 1, now.Add(-180*time.Second))
	err := clusterstate.UpdateNodes([]*apiv1.Node{ng1_1, ng1_2, ng1_3}, nil, now)
	assert.NoError(t, err)
	assert.True(t, clusterstate.IsClusterHealthy())
	assert.True(t, clusterstate.IsNodeGroupHealthy("ng1"))
	assert.Equal(t, NodeGroupScalingSafety{
		SafeToScale: false,
		Healthy:     true,
		BackoffStatus: backoff.Status{
			IsBackedOff: true,
			ErrorInfo: cloudprovider.InstanceErrorInfo{
				ErrorClass:   cloudprovider.OtherErrorClass,
				ErrorCode:    "timeout",
				ErrorMessage: "Scale-up timed out for node group ng1 after 3m0s",
			},
		},
	}, clusterstate.NodeGroupScaleUpSafety(ng1, now))
	assert.Equal(t, backoff.Status{
		IsBackedOff: true,
		ErrorInfo: cloudprovider.InstanceErrorInfo{
			ErrorClass:   cloudprovider.OtherErrorClass,
			ErrorCode:    "timeout",
			ErrorMessage: "Scale-up timed out for node group ng1 after 3m0s",
		}}, clusterstate.backoff.BackoffStatus(ng1, nil, now))

	// Backoff should expire after timeout
	now = now.Add(5 * time.Minute /*InitialNodeGroupBackoffDuration*/).Add(time.Second)
	assert.True(t, clusterstate.IsClusterHealthy())
	assert.True(t, clusterstate.IsNodeGroupHealthy("ng1"))
	assert.Equal(t, NodeGroupScalingSafety{SafeToScale: true, Healthy: true}, clusterstate.NodeGroupScaleUpSafety(ng1, now))

	// Another failed scale up should cause longer backoff
	clusterstate.RegisterScaleUp(provider.GetNodeGroup("ng1"), 1, now.Add(-121*time.Second))

	err = clusterstate.UpdateNodes([]*apiv1.Node{ng1_1, ng1_2, ng1_3}, nil, now)
	assert.NoError(t, err)
	assert.True(t, clusterstate.IsClusterHealthy())
	assert.True(t, clusterstate.IsNodeGroupHealthy("ng1"))
	assert.Equal(t, NodeGroupScalingSafety{
		SafeToScale: false,
		Healthy:     true,
		BackoffStatus: backoff.Status{
			IsBackedOff: true,
			ErrorInfo: cloudprovider.InstanceErrorInfo{
				ErrorClass:   cloudprovider.OtherErrorClass,
				ErrorCode:    "timeout",
				ErrorMessage: "Scale-up timed out for node group ng1 after 2m1s",
			},
		},
	}, clusterstate.NodeGroupScaleUpSafety(ng1, now))

	now = now.Add(5 * time.Minute /*InitialNodeGroupBackoffDuration*/).Add(time.Second)
	assert.Equal(t, NodeGroupScalingSafety{
		SafeToScale: false,
		Healthy:     true,
		BackoffStatus: backoff.Status{
			IsBackedOff: true,
			ErrorInfo: cloudprovider.InstanceErrorInfo{
				ErrorClass:   cloudprovider.OtherErrorClass,
				ErrorCode:    "timeout",
				ErrorMessage: "Scale-up timed out for node group ng1 after 2m1s",
			},
		},
	}, clusterstate.NodeGroupScaleUpSafety(ng1, now))

	// After successful scale-up, node group should still be backed off
	clusterstate.RegisterScaleUp(provider.GetNodeGroup("ng1"), 1, now)
	ng1_4 := BuildTestNode("ng1-4", 1000, 1000)
	SetNodeReadyState(ng1_4, true, now.Add(-1*time.Minute))
	provider.AddNode("ng1", ng1_4)
	err = clusterstate.UpdateNodes([]*apiv1.Node{ng1_1, ng1_2, ng1_3, ng1_4}, nil, now)
	assert.NoError(t, err)
	assert.True(t, clusterstate.IsClusterHealthy())
	assert.True(t, clusterstate.IsNodeGroupHealthy("ng1"))
	assert.Equal(t, NodeGroupScalingSafety{
		SafeToScale: false,
		Healthy:     true,
		BackoffStatus: backoff.Status{
			IsBackedOff: true,
			ErrorInfo: cloudprovider.InstanceErrorInfo{
				ErrorClass:   cloudprovider.OtherErrorClass,
				ErrorCode:    "timeout",
				ErrorMessage: "Scale-up timed out for node group ng1 after 2m1s",
			},
		},
	}, clusterstate.NodeGroupScaleUpSafety(ng1, now))
	assert.Equal(t, backoff.Status{
		IsBackedOff: true,
		ErrorInfo: cloudprovider.InstanceErrorInfo{
			ErrorClass:   cloudprovider.OtherErrorClass,
			ErrorCode:    "timeout",
			ErrorMessage: "Scale-up timed out for node group ng1 after 2m1s",
		}}, clusterstate.backoff.BackoffStatus(ng1, nil, now))
}

func TestGetClusterSize(t *testing.T) {
	now := time.Now()

	ng1_1 := BuildTestNode("ng1-1", 1000, 1000)
	SetNodeReadyState(ng1_1, true, now.Add(-time.Minute))
	ng2_1 := BuildTestNode("ng2-1", 1000, 1000)
	SetNodeReadyState(ng2_1, true, now.Add(-time.Minute))
	notAutoscaledNode := BuildTestNode("notAutoscaledNode", 1000, 1000)
	SetNodeReadyState(notAutoscaledNode, true, now.Add(-time.Minute))

	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 1, 10, 5)
	provider.AddNodeGroup("ng2", 1, 10, 1)

	provider.AddNode("ng1", ng1_1)
	provider.AddNode("ng2", ng2_1)

	// Add a node not belonging to any autoscaled node group. This is to make sure that GetAutoscaledNodesCount doesn't
	// take nodes from non-autoscaled node groups into account.
	provider.AddNode("notAutoscaledNode", notAutoscaledNode)

	fakeClient := &fake.Clientset{}
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false, "my-cool-configmap")
	clusterstate := NewClusterStateRegistry(provider, ClusterStateRegistryConfig{
		MaxTotalUnreadyPercentage: 10,
		OkTotalUnreadyCount:       1,
	}, fakeLogRecorder, newBackoff(), nodegroupconfig.NewDefaultNodeGroupConfigProcessor(config.NodeGroupAutoscalingOptions{MaxNodeProvisionTime: 15 * time.Minute}), asyncnodegroups.NewDefaultAsyncNodeGroupStateChecker())

	// There are 2 actual nodes in 2 node groups with target sizes of 5 and 1.
	clusterstate.UpdateNodes([]*apiv1.Node{ng1_1, ng2_1, notAutoscaledNode}, nil, now)
	currentSize, targetSize := clusterstate.GetAutoscaledNodesCount()
	assert.Equal(t, 2, currentSize)
	assert.Equal(t, 6, targetSize)

	// Current size should increase after a new node is added.
	clusterstate.UpdateNodes([]*apiv1.Node{ng1_1, ng1_1, notAutoscaledNode, ng2_1}, nil, now.Add(time.Minute))
	currentSize, targetSize = clusterstate.GetAutoscaledNodesCount()
	assert.Equal(t, 3, currentSize)
	assert.Equal(t, 6, targetSize)

	// Target size should increase after a new node group is added.
	provider.AddNodeGroup("ng3", 1, 10, 1)
	clusterstate.UpdateNodes([]*apiv1.Node{ng1_1, ng1_1, notAutoscaledNode, ng2_1}, nil, now.Add(2*time.Minute))
	currentSize, targetSize = clusterstate.GetAutoscaledNodesCount()
	assert.Equal(t, 3, currentSize)
	assert.Equal(t, 7, targetSize)

	// Target size should change after a node group changes its target size.
	for _, ng := range provider.NodeGroups() {
		ng.(*testprovider.TestNodeGroup).SetTargetSize(10)
	}
	clusterstate.UpdateNodes([]*apiv1.Node{ng1_1, ng1_1, notAutoscaledNode, ng2_1}, nil, now.Add(3*time.Minute))
	currentSize, targetSize = clusterstate.GetAutoscaledNodesCount()
	assert.Equal(t, 3, currentSize)
	assert.Equal(t, 30, targetSize)
}

func TestUpdateScaleUp(t *testing.T) {
	now := time.Now()
	later := now.Add(time.Minute)

	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 1, 10, 5)
	provider.AddNodeGroup("ng2", 1, 10, 5)
	fakeClient := &fake.Clientset{}
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false, "my-cool-configmap")
	clusterstate := NewClusterStateRegistry(
		provider,
		ClusterStateRegistryConfig{
			MaxTotalUnreadyPercentage: 10,
			OkTotalUnreadyCount:       1,
		},
		fakeLogRecorder,
		newBackoff(),
		nodegroupconfig.NewDefaultNodeGroupConfigProcessor(config.NodeGroupAutoscalingOptions{MaxNodeProvisionTime: 10 * time.Second}),
		asyncnodegroups.NewDefaultAsyncNodeGroupStateChecker(),
	)

	// Test cases for `RegisterScaleUp`
	clusterstate.RegisterScaleUp(provider.GetNodeGroup("ng1"), 100, now)
	assert.Equal(t, clusterstate.scaleUpRequests["ng1"].Increase, 100)
	assert.Equal(t, clusterstate.scaleUpRequests["ng1"].Time, now)
	assert.Equal(t, clusterstate.scaleUpRequests["ng1"].ExpectedAddTime, now.Add(10*time.Second))

	// expect no change of times on negative delta
	clusterstate.RegisterScaleUp(provider.GetNodeGroup("ng1"), -20, later)
	assert.Equal(t, clusterstate.scaleUpRequests["ng1"].Increase, 80)
	assert.Equal(t, clusterstate.scaleUpRequests["ng1"].Time, now)
	assert.Equal(t, clusterstate.scaleUpRequests["ng1"].ExpectedAddTime, now.Add(10*time.Second))

	// update times on positive delta
	clusterstate.RegisterScaleUp(provider.GetNodeGroup("ng1"), 30, later)
	assert.Equal(t, clusterstate.scaleUpRequests["ng1"].Increase, 110)
	assert.Equal(t, clusterstate.scaleUpRequests["ng1"].Time, later)
	assert.Equal(t, clusterstate.scaleUpRequests["ng1"].ExpectedAddTime, later.Add(10*time.Second))

	// if we get below 0 scalup is deleted
	clusterstate.RegisterScaleUp(provider.GetNodeGroup("ng1"), -200, now)
	assert.Nil(t, clusterstate.scaleUpRequests["ng1"])

	// If new scalup is registered with negative delta nothing should happen
	clusterstate.RegisterScaleUp(provider.GetNodeGroup("ng1"), -200, now)
	assert.Nil(t, clusterstate.scaleUpRequests["ng1"])
}

func TestScaleUpFailures(t *testing.T) {
	now := time.Now()

	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 0, 10, 0)
	provider.AddNodeGroup("ng2", 0, 10, 0)
	assert.NotNil(t, provider)

	fakeClient := &fake.Clientset{}
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false, "my-cool-configmap")
	clusterstate := NewClusterStateRegistry(provider, ClusterStateRegistryConfig{}, fakeLogRecorder, newBackoff(), nodegroupconfig.NewDefaultNodeGroupConfigProcessor(config.NodeGroupAutoscalingOptions{MaxNodeProvisionTime: 15 * time.Minute}), asyncnodegroups.NewDefaultAsyncNodeGroupStateChecker())

	clusterstate.RegisterFailedScaleUp(provider.GetNodeGroup("ng1"), string(metrics.Timeout), "", "", "", now)
	clusterstate.RegisterFailedScaleUp(provider.GetNodeGroup("ng2"), string(metrics.Timeout), "", "", "", now)
	clusterstate.RegisterFailedScaleUp(provider.GetNodeGroup("ng1"), string(metrics.APIError), "", "", "", now.Add(time.Minute))

	failures := clusterstate.GetScaleUpFailures()
	assert.Equal(t, map[string][]ScaleUpFailure{
		"ng1": {
			{NodeGroup: provider.GetNodeGroup("ng1"), Reason: metrics.Timeout, Time: now},
			{NodeGroup: provider.GetNodeGroup("ng1"), Reason: metrics.APIError, Time: now.Add(time.Minute)},
		},
		"ng2": {
			{NodeGroup: provider.GetNodeGroup("ng2"), Reason: metrics.Timeout, Time: now},
		},
	}, failures)

	clusterstate.clearScaleUpFailures()
	assert.Empty(t, clusterstate.GetScaleUpFailures())
}

func newBackoff() backoff.Backoff {
	return backoff.NewIdBasedExponentialBackoff(5*time.Minute, /*InitialNodeGroupBackoffDuration*/
		30*time.Minute /*MaxNodeGroupBackoffDuration*/, 3*time.Hour /*NodeGroupBackoffResetTimeout*/)
}

func TestUpdateAcceptableRanges(t *testing.T) {
	testCases := []struct {
		name string

		targetSizes      map[string]int
		readiness        map[string]Readiness
		scaleUpRequests  map[string]*ScaleUpRequest
		scaledDownGroups []string

		wantAcceptableRanges map[string]AcceptableRange
	}{
		{
			name: "No scale-ups/scale-downs",
			targetSizes: map[string]int{
				"ng1": 10,
				"ng2": 20,
			},
			readiness: map[string]Readiness{
				"ng1": {Ready: make([]string, 10)},
				"ng2": {Ready: make([]string, 20)},
			},
			wantAcceptableRanges: map[string]AcceptableRange{
				"ng1": {MinNodes: 10, MaxNodes: 10, CurrentTarget: 10},
				"ng2": {MinNodes: 20, MaxNodes: 20, CurrentTarget: 20},
			},
		},
		{
			name: "Ongoing scale-ups",
			targetSizes: map[string]int{
				"ng1": 10,
				"ng2": 20,
			},
			readiness: map[string]Readiness{
				"ng1": {Ready: make([]string, 10)},
				"ng2": {Ready: make([]string, 20)},
			},
			scaleUpRequests: map[string]*ScaleUpRequest{
				"ng1": {Increase: 3},
				"ng2": {Increase: 5},
			},
			wantAcceptableRanges: map[string]AcceptableRange{
				"ng1": {MinNodes: 7, MaxNodes: 10, CurrentTarget: 10},
				"ng2": {MinNodes: 15, MaxNodes: 20, CurrentTarget: 20},
			},
		},
		{
			name: "Ongoing scale-downs",
			targetSizes: map[string]int{
				"ng1": 10,
				"ng2": 20,
			},
			readiness: map[string]Readiness{
				"ng1": {Ready: make([]string, 10)},
				"ng2": {Ready: make([]string, 20)},
			},
			scaledDownGroups: []string{"ng1", "ng1", "ng2", "ng2", "ng2"},
			wantAcceptableRanges: map[string]AcceptableRange{
				"ng1": {MinNodes: 10, MaxNodes: 12, CurrentTarget: 10},
				"ng2": {MinNodes: 20, MaxNodes: 23, CurrentTarget: 20},
			},
		},
		{
			name: "Some short unregistered nodes",
			targetSizes: map[string]int{
				"ng1": 10,
				"ng2": 20,
			},
			readiness: map[string]Readiness{
				"ng1": {Ready: make([]string, 8), Unregistered: make([]string, 2)},
				"ng2": {Ready: make([]string, 17), Unregistered: make([]string, 3)},
			},
			wantAcceptableRanges: map[string]AcceptableRange{
				"ng1": {MinNodes: 10, MaxNodes: 10, CurrentTarget: 10},
				"ng2": {MinNodes: 20, MaxNodes: 20, CurrentTarget: 20},
			},
		},
		{
			name: "Some long unregistered nodes",
			targetSizes: map[string]int{
				"ng1": 10,
				"ng2": 20,
			},
			readiness: map[string]Readiness{
				"ng1": {Ready: make([]string, 8), LongUnregistered: make([]string, 2)},
				"ng2": {Ready: make([]string, 17), LongUnregistered: make([]string, 3)},
			},
			wantAcceptableRanges: map[string]AcceptableRange{
				"ng1": {MinNodes: 8, MaxNodes: 10, CurrentTarget: 10},
				"ng2": {MinNodes: 17, MaxNodes: 20, CurrentTarget: 20},
			},
		},
		{
			name: "Everything together",
			targetSizes: map[string]int{
				"ng1": 10,
				"ng2": 20,
			},
			readiness: map[string]Readiness{
				"ng1": {Ready: make([]string, 8), Unregistered: make([]string, 1), LongUnregistered: make([]string, 2)},
				"ng2": {Ready: make([]string, 17), Unregistered: make([]string, 3), LongUnregistered: make([]string, 4)},
			},
			scaleUpRequests: map[string]*ScaleUpRequest{
				"ng1": {Increase: 3},
				"ng2": {Increase: 5},
			},
			scaledDownGroups: []string{"ng1", "ng1", "ng2", "ng2", "ng2"},
			wantAcceptableRanges: map[string]AcceptableRange{
				"ng1": {MinNodes: 5, MaxNodes: 12, CurrentTarget: 10},
				"ng2": {MinNodes: 11, MaxNodes: 23, CurrentTarget: 20},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := testprovider.NewTestCloudProvider(nil, nil)
			for nodeGroupName, targetSize := range tc.targetSizes {
				provider.AddNodeGroup(nodeGroupName, 0, 1000, targetSize)
			}
			var scaleDownRequests []*ScaleDownRequest
			for _, nodeGroupName := range tc.scaledDownGroups {
				scaleDownRequests = append(scaleDownRequests, &ScaleDownRequest{
					NodeGroup: provider.GetNodeGroup(nodeGroupName),
				})
			}

			clusterState := &ClusterStateRegistry{
				cloudProvider:         provider,
				perNodeGroupReadiness: tc.readiness,
				scaleUpRequests:       tc.scaleUpRequests,
				scaleDownRequests:     scaleDownRequests,
			}

			clusterState.updateAcceptableRanges(tc.targetSizes)
			assert.Equal(t, tc.wantAcceptableRanges, clusterState.acceptableRanges)
		})
	}
}

func TestUpdateIncorrectNodeGroupSizes(t *testing.T) {
	timeNow := time.Now()
	testCases := []struct {
		name string

		acceptableRanges map[string]AcceptableRange
		readiness        map[string]Readiness
		incorrectSizes   map[string]IncorrectNodeGroupSize

		wantIncorrectSizes map[string]IncorrectNodeGroupSize
	}{
		{
			name: "node groups with correct sizes",
			acceptableRanges: map[string]AcceptableRange{
				"ng1": {CurrentTarget: 10},
				"ng2": {CurrentTarget: 20},
			},
			readiness: map[string]Readiness{
				"ng1": {Registered: make([]string, 10)},
				"ng2": {Registered: make([]string, 20)},
			},
			incorrectSizes:     map[string]IncorrectNodeGroupSize{},
			wantIncorrectSizes: map[string]IncorrectNodeGroupSize{},
		},
		{
			name: "node groups with correct sizes after not being correct sized",
			acceptableRanges: map[string]AcceptableRange{
				"ng1": {CurrentTarget: 10},
				"ng2": {CurrentTarget: 20},
			},
			readiness: map[string]Readiness{
				"ng1": {Registered: make([]string, 10)},
				"ng2": {Registered: make([]string, 20)},
			},
			incorrectSizes: map[string]IncorrectNodeGroupSize{
				"ng1": {CurrentSize: 8, ExpectedSize: 10, FirstObserved: timeNow.Add(-time.Hour)},
				"ng2": {CurrentSize: 15, ExpectedSize: 20, FirstObserved: timeNow.Add(-time.Minute)},
			},
			wantIncorrectSizes: map[string]IncorrectNodeGroupSize{},
		},
		{
			name: "node groups below the target size",
			acceptableRanges: map[string]AcceptableRange{
				"ng1": {CurrentTarget: 10},
				"ng2": {CurrentTarget: 20},
			},
			readiness: map[string]Readiness{
				"ng1": {Registered: make([]string, 8)},
				"ng2": {Registered: make([]string, 15)},
			},
			incorrectSizes: map[string]IncorrectNodeGroupSize{},
			wantIncorrectSizes: map[string]IncorrectNodeGroupSize{
				"ng1": {CurrentSize: 8, ExpectedSize: 10, FirstObserved: timeNow},
				"ng2": {CurrentSize: 15, ExpectedSize: 20, FirstObserved: timeNow},
			},
		},
		{
			name: "node groups above the target size",
			acceptableRanges: map[string]AcceptableRange{
				"ng1": {CurrentTarget: 10},
				"ng2": {CurrentTarget: 20},
			},
			readiness: map[string]Readiness{
				"ng1": {Registered: make([]string, 12)},
				"ng2": {Registered: make([]string, 25)},
			},
			incorrectSizes: map[string]IncorrectNodeGroupSize{},
			wantIncorrectSizes: map[string]IncorrectNodeGroupSize{
				"ng1": {CurrentSize: 12, ExpectedSize: 10, FirstObserved: timeNow},
				"ng2": {CurrentSize: 25, ExpectedSize: 20, FirstObserved: timeNow},
			},
		},
		{
			name: "node groups below the target size with changed delta",
			acceptableRanges: map[string]AcceptableRange{
				"ng1": {CurrentTarget: 10},
				"ng2": {CurrentTarget: 20},
			},
			readiness: map[string]Readiness{
				"ng1": {Registered: make([]string, 8)},
				"ng2": {Registered: make([]string, 15)},
			},
			incorrectSizes: map[string]IncorrectNodeGroupSize{
				"ng1": {CurrentSize: 7, ExpectedSize: 10, FirstObserved: timeNow.Add(-time.Hour)},
				"ng2": {CurrentSize: 14, ExpectedSize: 20, FirstObserved: timeNow.Add(-time.Minute)},
			},
			wantIncorrectSizes: map[string]IncorrectNodeGroupSize{
				"ng1": {CurrentSize: 8, ExpectedSize: 10, FirstObserved: timeNow},
				"ng2": {CurrentSize: 15, ExpectedSize: 20, FirstObserved: timeNow},
			},
		},
		{
			name: "node groups below the target size with the same delta",
			acceptableRanges: map[string]AcceptableRange{
				"ng1": {CurrentTarget: 10},
				"ng2": {CurrentTarget: 20},
			},
			readiness: map[string]Readiness{
				"ng1": {Registered: make([]string, 8)},
				"ng2": {Registered: make([]string, 15)},
			},
			incorrectSizes: map[string]IncorrectNodeGroupSize{
				"ng1": {CurrentSize: 8, ExpectedSize: 10, FirstObserved: timeNow.Add(-time.Hour)},
				"ng2": {CurrentSize: 15, ExpectedSize: 20, FirstObserved: timeNow.Add(-time.Minute)},
			},
			wantIncorrectSizes: map[string]IncorrectNodeGroupSize{
				"ng1": {CurrentSize: 8, ExpectedSize: 10, FirstObserved: timeNow.Add(-time.Hour)},
				"ng2": {CurrentSize: 15, ExpectedSize: 20, FirstObserved: timeNow.Add(-time.Minute)},
			},
		},
		{
			name: "node groups below the target size with short unregistered nodes",
			acceptableRanges: map[string]AcceptableRange{
				"ng1": {CurrentTarget: 10},
				"ng2": {CurrentTarget: 20},
			},
			readiness: map[string]Readiness{
				"ng1": {Registered: make([]string, 8), Unregistered: make([]string, 2)},
				"ng2": {Registered: make([]string, 15), Unregistered: make([]string, 3)},
			},
			incorrectSizes: map[string]IncorrectNodeGroupSize{},
			wantIncorrectSizes: map[string]IncorrectNodeGroupSize{
				"ng2": {CurrentSize: 15, ExpectedSize: 20, FirstObserved: timeNow},
			},
		},
		{
			name: "node groups below the target size with long unregistered nodes",
			acceptableRanges: map[string]AcceptableRange{
				"ng1": {CurrentTarget: 10},
				"ng2": {CurrentTarget: 20},
			},
			readiness: map[string]Readiness{
				"ng1": {Registered: make([]string, 8), LongUnregistered: make([]string, 2)},
				"ng2": {Registered: make([]string, 15), LongUnregistered: make([]string, 3)},
			},
			incorrectSizes: map[string]IncorrectNodeGroupSize{},
			wantIncorrectSizes: map[string]IncorrectNodeGroupSize{
				"ng2": {CurrentSize: 15, ExpectedSize: 20, FirstObserved: timeNow},
			},
		},
		{
			name: "node groups below the target size with various unregistered nodes",
			acceptableRanges: map[string]AcceptableRange{
				"ng1": {CurrentTarget: 10},
				"ng2": {CurrentTarget: 20},
			},
			readiness: map[string]Readiness{
				"ng1": {Registered: make([]string, 8), Unregistered: make([]string, 1), LongUnregistered: make([]string, 1)},
				"ng2": {Registered: make([]string, 15), Unregistered: make([]string, 2), LongUnregistered: make([]string, 2)},
			},
			incorrectSizes: map[string]IncorrectNodeGroupSize{},
			wantIncorrectSizes: map[string]IncorrectNodeGroupSize{
				"ng2": {CurrentSize: 15, ExpectedSize: 20, FirstObserved: timeNow},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := testprovider.NewTestCloudProvider(nil, nil)
			for nodeGroupName, acceptableRange := range tc.acceptableRanges {
				provider.AddNodeGroup(nodeGroupName, 0, 1000, acceptableRange.CurrentTarget)
			}

			clusterState := &ClusterStateRegistry{
				cloudProvider:              provider,
				acceptableRanges:           tc.acceptableRanges,
				perNodeGroupReadiness:      tc.readiness,
				incorrectNodeGroupSizes:    tc.incorrectSizes,
				asyncNodeGroupStateChecker: asyncnodegroups.NewDefaultAsyncNodeGroupStateChecker(),
			}

			clusterState.updateIncorrectNodeGroupSizes(timeNow)
			assert.Equal(t, tc.wantIncorrectSizes, clusterState.incorrectNodeGroupSizes)
		})
	}
}

func TestTruncateIfExceedMaxSize(t *testing.T) {
	testCases := []struct {
		name        string
		message     string
		maxSize     int
		wantMessage string
	}{
		{
			name:        "Message doesn't exceed maxSize",
			message:     "Some message",
			maxSize:     len("Some message"),
			wantMessage: "Some message",
		},
		{
			name:        "Message exceeds maxSize",
			message:     "Some long message",
			maxSize:     len("Some long message") - 1,
			wantMessage: "Some <truncated>",
		},
		{
			name:        "Message doesn't exceed maxSize and maxSize is smaller than truncatedMessageSuffix length",
			message:     "msg",
			maxSize:     len("msg"),
			wantMessage: "msg",
		},
		{
			name:        "Message exceeds maxSize and maxSize is smaller than truncatedMessageSuffix length",
			message:     "msg",
			maxSize:     2,
			wantMessage: "ms",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := truncateIfExceedMaxLength(tc.message, tc.maxSize)
			assert.Equal(t, tc.wantMessage, got)
		})
	}
}

func TestUpcomingNodesFromUpcomingNodeGroups(t *testing.T) {

	testCases := []struct {
		isUpcomingMockMap                 map[string]bool
		nodeGroups                        map[string]int
		expectedGroupsUpcomingNodesNumber map[string]int
		updateNodes                       bool
	}{
		{
			isUpcomingMockMap:                 map[string]bool{"ng": true},
			nodeGroups:                        map[string]int{"ng": 2},
			expectedGroupsUpcomingNodesNumber: map[string]int{"ng": 2},
			updateNodes:                       false,
		},
		{
			isUpcomingMockMap:                 map[string]bool{"ng": true, "ng2": true},
			nodeGroups:                        map[string]int{"ng": 2, "ng2": 3},
			expectedGroupsUpcomingNodesNumber: map[string]int{"ng": 2, "ng2": 3},
			updateNodes:                       false,
		},
		{
			isUpcomingMockMap:                 map[string]bool{},
			nodeGroups:                        map[string]int{"ng": 2},
			expectedGroupsUpcomingNodesNumber: map[string]int{"ng": 2},
			updateNodes:                       true,
		},
		{
			isUpcomingMockMap:                 map[string]bool{"ng": true},
			nodeGroups:                        map[string]int{"ng": 2, "ng2": 1},
			expectedGroupsUpcomingNodesNumber: map[string]int{"ng": 2, "ng2": 1},
			updateNodes:                       true,
		},
		{
			isUpcomingMockMap:                 map[string]bool{"ng": true},
			nodeGroups:                        map[string]int{"ng": 2, "ng2": 1},
			expectedGroupsUpcomingNodesNumber: map[string]int{"ng": 2, "ng2": 0},
			updateNodes:                       false,
		},
	}

	for _, tc := range testCases {

		now := time.Now()
		provider := testprovider.NewTestCloudProvider(nil, nil)
		for groupName, groupSize := range tc.nodeGroups {
			provider.AddUpcomingNodeGroup(groupName, 1, 10, groupSize)
		}

		assert.NotNil(t, provider)
		fakeClient := &fake.Clientset{}
		fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false, "some-map")

		clusterstate := NewClusterStateRegistry(
			provider,
			ClusterStateRegistryConfig{MaxTotalUnreadyPercentage: 10, OkTotalUnreadyCount: 1},
			fakeLogRecorder,
			newBackoff(),
			nodegroupconfig.NewDefaultNodeGroupConfigProcessor(config.NodeGroupAutoscalingOptions{MaxNodeProvisionTime: 15 * time.Minute}),
			&asyncnodegroups.MockAsyncNodeGroupStateChecker{IsUpcomingNodeGroup: tc.isUpcomingMockMap},
		)
		if tc.updateNodes {
			err := clusterstate.UpdateNodes([]*apiv1.Node{}, nil, now)
			assert.NoError(t, err)
		}

		assert.Equal(t, 0, len(clusterstate.GetClusterReadiness().Unready))
		assert.Equal(t, 0, len(clusterstate.GetClusterReadiness().NotStarted))
		upcoming, upcomingRegistered := clusterstate.GetUpcomingNodes()
		for groupName, groupSize := range tc.expectedGroupsUpcomingNodesNumber {
			assert.Equal(t, groupSize, upcoming[groupName])
			assert.Empty(t, upcomingRegistered[groupName])
		}
	}

}
