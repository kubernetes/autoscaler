/*
Copyright The Kubernetes Authors.

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

package nebius

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	commonv1 "github.com/nebius/gosdk/proto/nebius/common/v1"
	computev1 "github.com/nebius/gosdk/proto/nebius/compute/v1"
	mk8sv1 "github.com/nebius/gosdk/proto/nebius/mk8s/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// mockNebiusAPI implements nebiusAPI for testing.
type mockNebiusAPI struct {
	listNodeGroupsResponses  []*mk8sv1.ListNodeGroupsResponse
	listNodeGroupsErr        error
	listNodeGroupsPageTokens []string
	listInstancesResponses   []*computev1.ListInstancesResponse
	listInstancesErr         error
	listInstancesPageTokens  []string
	getNodeGroupResponse     *mk8sv1.NodeGroup
	getNodeGroupErr          error
	updateNodeGroupErr       error
	lastUpdateReq            *mk8sv1.UpdateNodeGroupRequest
	deleteInstanceIDs        []string
	deleteInstanceErr        error
	deleteInstanceErrAfter   int // fail after this many successful deletes (0 = fail immediately)
}

func (m *mockNebiusAPI) ListNodeGroups(_ context.Context, req *mk8sv1.ListNodeGroupsRequest) (*mk8sv1.ListNodeGroupsResponse, error) {
	m.listNodeGroupsPageTokens = append(m.listNodeGroupsPageTokens, req.GetPageToken())
	if m.listNodeGroupsErr != nil {
		return nil, m.listNodeGroupsErr
	}
	if len(m.listNodeGroupsResponses) == 0 {
		return &mk8sv1.ListNodeGroupsResponse{}, nil
	}
	resp := m.listNodeGroupsResponses[0]
	m.listNodeGroupsResponses = m.listNodeGroupsResponses[1:]
	return resp, nil
}

func (m *mockNebiusAPI) ListInstances(_ context.Context, req *computev1.ListInstancesRequest) (*computev1.ListInstancesResponse, error) {
	m.listInstancesPageTokens = append(m.listInstancesPageTokens, req.GetPageToken())
	if m.listInstancesErr != nil {
		return nil, m.listInstancesErr
	}
	if len(m.listInstancesResponses) == 0 {
		return &computev1.ListInstancesResponse{}, nil
	}
	resp := m.listInstancesResponses[0]
	m.listInstancesResponses = m.listInstancesResponses[1:]
	return resp, nil
}

func (m *mockNebiusAPI) GetNodeGroup(_ context.Context, _ *mk8sv1.GetNodeGroupRequest) (*mk8sv1.NodeGroup, error) {
	if m.getNodeGroupErr != nil {
		return nil, m.getNodeGroupErr
	}
	return m.getNodeGroupResponse, nil
}

func (m *mockNebiusAPI) UpdateNodeGroup(_ context.Context, req *mk8sv1.UpdateNodeGroupRequest) error {
	m.lastUpdateReq = req
	return m.updateNodeGroupErr
}

func (m *mockNebiusAPI) DeleteInstance(_ context.Context, req *computev1.DeleteInstanceRequest) error {
	m.deleteInstanceIDs = append(m.deleteInstanceIDs, req.GetId())
	if m.deleteInstanceErr != nil && len(m.deleteInstanceIDs) > m.deleteInstanceErrAfter {
		return m.deleteInstanceErr
	}
	return nil
}

// Helper to build a node group proto with autoscaling spec.
func makeNodeGroupProto(id, name string, minCount, maxCount, targetCount int64) *mk8sv1.NodeGroup {
	return &mk8sv1.NodeGroup{
		Metadata: &commonv1.ResourceMetadata{
			Id:   id,
			Name: name,
		},
		Spec: &mk8sv1.NodeGroupSpec{
			Size: &mk8sv1.NodeGroupSpec_Autoscaling{
				Autoscaling: &mk8sv1.NodeGroupAutoscalingSpec{
					MinNodeCount: minCount,
					MaxNodeCount: maxCount,
				},
			},
		},
		Status: &mk8sv1.NodeGroupStatus{
			TargetNodeCount: targetCount,
		},
	}
}

// Helper to build an instance proto.
func makeInstanceProto(id string, labels map[string]string) *computev1.Instance {
	return &computev1.Instance{
		Metadata: &commonv1.ResourceMetadata{
			Id:     id,
			Labels: labels,
		},
	}
}

func newTestManager(api nebiusAPI) *Manager {
	return &Manager{
		client:     api,
		clusterID:  "cluster-1",
		parentID:   "parent-1",
		nodeGroups: make([]*NodeGroup, 0),
	}
}

// --- Config / newManager tests ---

func TestNewManagerMissingToken(t *testing.T) {
	cfg := `{"cluster_id": "cluster-123", "parent_id": "folder-abc"}`
	_, err := newManager(bytes.NewBufferString(cfg))
	assert.ErrorContains(t, err, "IAM token is not provided")
}

func TestNewManagerMissingClusterID(t *testing.T) {
	cfg := `{"iam_token": "my-token", "parent_id": "folder-abc"}`
	_, err := newManager(bytes.NewBufferString(cfg))
	assert.ErrorContains(t, err, "cluster ID is not provided")
}

func TestNewManagerMissingParentID(t *testing.T) {
	cfg := `{"iam_token": "my-token", "cluster_id": "cluster-123"}`
	_, err := newManager(bytes.NewBufferString(cfg))
	assert.ErrorContains(t, err, "parent ID is not provided")
}

func TestNewManagerEnvVarFallback(t *testing.T) {
	// Without env vars, empty config should fail validation.
	_, err := newManager(bytes.NewBufferString(`{}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not provided")

	// With env vars set, config validation should pass (SDK init may still fail).
	t.Setenv("NEBIUS_IAM_TOKEN", "env-token")
	t.Setenv("NEBIUS_CLUSTER_ID", "env-cluster")
	t.Setenv("NEBIUS_PARENT_ID", "env-parent")

	_, err = newManager(bytes.NewBufferString(`{}`))
	if err != nil {
		// SDK init may fail, but config validation should have passed.
		assert.NotContains(t, err.Error(), "not provided")
	}
}

// --- NodeGroup basic tests ---

func TestNodeGroupMinMax(t *testing.T) {
	ng := &NodeGroup{
		id:      "test-ng",
		minSize: 2,
		maxSize: 10,
	}
	assert.Equal(t, 2, ng.MinSize())
	assert.Equal(t, 10, ng.MaxSize())
}

func TestNodeGroupId(t *testing.T) {
	ng := &NodeGroup{
		id: "my-node-group-id",
	}
	assert.Equal(t, "my-node-group-id", ng.Id())
}

func TestNodeGroupDebug(t *testing.T) {
	ng := &NodeGroup{
		id:      "my-node-group",
		minSize: 1,
		maxSize: 5,
	}
	debug := ng.Debug()
	assert.Contains(t, debug, "my-node-group")
	assert.Contains(t, debug, "1")
	assert.Contains(t, debug, "5")
}

func TestNodeGroupHasInstance(t *testing.T) {
	ng := &NodeGroup{
		id: "my-node-group",
		instances: map[string]struct{}{
			"nebius://instance-1": {},
			"nebius://instance-2": {},
		},
	}
	assert.True(t, ng.hasInstance("nebius://instance-1"))
	assert.True(t, ng.hasInstance("nebius://instance-2"))
	assert.False(t, ng.hasInstance("nebius://instance-3"))
}

func TestNodeGroupNodes(t *testing.T) {
	ng := &NodeGroup{
		id: "my-node-group",
		instances: map[string]struct{}{
			"nebius://instance-1": {},
			"nebius://instance-2": {},
		},
	}
	instances, err := ng.Nodes()
	assert.NoError(t, err)
	assert.Len(t, instances, 2)
}

func TestNodeGroupExist(t *testing.T) {
	ng := &NodeGroup{nodeGroup: nil}
	assert.False(t, ng.Exist())

	ng2 := &NodeGroup{nodeGroup: makeNodeGroupProto("ng-1", "group", 1, 5, 3)}
	assert.True(t, ng2.Exist())
}

func TestToProviderID(t *testing.T) {
	id := toProviderID("abc-123")
	assert.Equal(t, "nebius://abc-123", id)
}

// --- Refresh tests ---

func TestRefresh_BasicNodeGroups(t *testing.T) {
	t.Parallel()
	mock := &mockNebiusAPI{
		listNodeGroupsResponses: []*mk8sv1.ListNodeGroupsResponse{
			{
				Items: []*mk8sv1.NodeGroup{
					makeNodeGroupProto("ng-1", "group-one", 1, 5, 3),
					makeNodeGroupProto("ng-2", "group-two", 2, 8, 4),
				},
			},
		},
		listInstancesResponses: []*computev1.ListInstancesResponse{
			{Items: []*computev1.Instance{}},
		},
	}

	m := newTestManager(mock)
	err := m.Refresh()
	require.NoError(t, err)

	groups := m.getNodeGroups()
	require.Len(t, groups, 2)

	assert.Equal(t, "ng-1", groups[0].id)
	assert.Equal(t, 1, groups[0].minSize)
	assert.Equal(t, 5, groups[0].maxSize)
	assert.Equal(t, 3, groups[0].targetSize)

	assert.Equal(t, "ng-2", groups[1].id)
	assert.Equal(t, 2, groups[1].minSize)
	assert.Equal(t, 8, groups[1].maxSize)
	assert.Equal(t, 4, groups[1].targetSize)
}

func TestRefresh_WithPagination(t *testing.T) {
	t.Parallel()
	mock := &mockNebiusAPI{
		listNodeGroupsResponses: []*mk8sv1.ListNodeGroupsResponse{
			{
				Items:         []*mk8sv1.NodeGroup{makeNodeGroupProto("ng-1", "group-one", 1, 5, 2)},
				NextPageToken: "page2",
			},
			{
				Items: []*mk8sv1.NodeGroup{makeNodeGroupProto("ng-2", "group-two", 0, 10, 5)},
			},
		},
		listInstancesResponses: []*computev1.ListInstancesResponse{
			{Items: []*computev1.Instance{}},
		},
	}

	m := newTestManager(mock)
	err := m.Refresh()
	require.NoError(t, err)

	groups := m.getNodeGroups()
	require.Len(t, groups, 2)
	assert.Equal(t, "ng-1", groups[0].id)
	assert.Equal(t, "ng-2", groups[1].id)
	assert.Equal(t, []string{"", "page2"}, mock.listNodeGroupsPageTokens)
}

func TestRefresh_InstancePagination(t *testing.T) {
	t.Parallel()
	mock := &mockNebiusAPI{
		listNodeGroupsResponses: []*mk8sv1.ListNodeGroupsResponse{
			{
				Items: []*mk8sv1.NodeGroup{makeNodeGroupProto("ng-1", "group-one", 0, 10, 3)},
			},
		},
		listInstancesResponses: []*computev1.ListInstancesResponse{
			{
				Items: []*computev1.Instance{
					makeInstanceProto("inst-1", map[string]string{nodeGroupIDLabel: "ng-1"}),
				},
				NextPageToken: "page2",
			},
			{
				Items: []*computev1.Instance{
					makeInstanceProto("inst-2", map[string]string{nodeGroupIDLabel: "ng-1"}),
				},
			},
		},
	}

	m := newTestManager(mock)
	err := m.Refresh()
	require.NoError(t, err)

	groups := m.getNodeGroups()
	require.Len(t, groups, 1)
	assert.Len(t, groups[0].instances, 2)
	assert.True(t, groups[0].hasInstance("nebius://inst-1"))
	assert.True(t, groups[0].hasInstance("nebius://inst-2"))
	assert.Equal(t, []string{"", "page2"}, mock.listInstancesPageTokens)
}

func TestRefresh_InstanceListError(t *testing.T) {
	t.Parallel()
	mock := &mockNebiusAPI{
		listNodeGroupsResponses: []*mk8sv1.ListNodeGroupsResponse{
			{
				Items: []*mk8sv1.NodeGroup{makeNodeGroupProto("ng-1", "group-one", 1, 5, 2)},
			},
		},
		listInstancesErr: fmt.Errorf("instance API unavailable"),
	}

	m := newTestManager(mock)
	err := m.Refresh()
	// Refresh should still succeed even if instance listing fails.
	require.NoError(t, err)

	groups := m.getNodeGroups()
	require.Len(t, groups, 1)
	assert.Empty(t, groups[0].instances)
}

func TestRefresh_NodeGroupListError(t *testing.T) {
	t.Parallel()
	mock := &mockNebiusAPI{
		listNodeGroupsErr: fmt.Errorf("node group API unavailable"),
	}

	m := newTestManager(mock)
	err := m.Refresh()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list node groups")
}

// --- IncreaseSize tests ---

func TestIncreaseSize_Success(t *testing.T) {
	t.Parallel()
	mock := &mockNebiusAPI{
		getNodeGroupResponse: makeNodeGroupProto("ng-1", "group-one", 1, 10, 3),
	}
	m := newTestManager(mock)
	ng := &NodeGroup{
		id:         "ng-1",
		manager:    m,
		targetSize: 3,
		minSize:    1,
		maxSize:    10,
	}

	err := ng.IncreaseSize(2)
	require.NoError(t, err)
	require.NotNil(t, mock.lastUpdateReq)
	assert.Equal(t, int64(5), mock.lastUpdateReq.GetSpec().GetFixedNodeCount())

	// Verify in-memory target size is updated for subsequent calls.
	size, err := ng.TargetSize()
	require.NoError(t, err)
	assert.Equal(t, 5, size)
}

func TestIncreaseSize_ExceedsMax(t *testing.T) {
	t.Parallel()
	mock := &mockNebiusAPI{}
	m := newTestManager(mock)
	ng := &NodeGroup{
		id:         "ng-1",
		manager:    m,
		targetSize: 8,
		minSize:    1,
		maxSize:    10,
	}

	err := ng.IncreaseSize(3)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "size increase is too large")
	assert.Nil(t, mock.lastUpdateReq)
}

func TestIncreaseSize_NegativeDelta(t *testing.T) {
	t.Parallel()
	mock := &mockNebiusAPI{}
	m := newTestManager(mock)
	ng := &NodeGroup{
		id:         "ng-1",
		manager:    m,
		targetSize: 5,
		minSize:    1,
		maxSize:    10,
	}

	err := ng.IncreaseSize(-1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "delta must be positive")
	assert.Nil(t, mock.lastUpdateReq)
}

// --- DeleteNodes tests ---

func TestDeleteNodes_Success(t *testing.T) {
	t.Parallel()
	mock := &mockNebiusAPI{
		getNodeGroupResponse: makeNodeGroupProto("ng-1", "group-one", 1, 10, 5),
	}
	m := newTestManager(mock)
	ng := &NodeGroup{
		id:         "ng-1",
		manager:    m,
		targetSize: 5,
		minSize:    1,
		maxSize:    10,
	}

	nodes := []*apiv1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "node-1"},
			Spec:       apiv1.NodeSpec{ProviderID: "nebius://inst-1"},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "node-2"},
			Spec:       apiv1.NodeSpec{ProviderID: "nebius://inst-2"},
		},
	}
	err := ng.DeleteNodes(nodes)
	require.NoError(t, err)

	// Verify specific instances were deleted.
	assert.Equal(t, []string{"inst-1", "inst-2"}, mock.deleteInstanceIDs)

	// Verify node group size was updated.
	require.NotNil(t, mock.lastUpdateReq)
	assert.Equal(t, int64(3), mock.lastUpdateReq.GetSpec().GetFixedNodeCount())
}

func TestDeleteNodes_BelowMin(t *testing.T) {
	t.Parallel()
	mock := &mockNebiusAPI{}
	m := newTestManager(mock)
	ng := &NodeGroup{
		id:         "ng-1",
		manager:    m,
		targetSize: 3,
		minSize:    2,
		maxSize:    10,
	}

	nodes := []*apiv1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "node-1"},
			Spec:       apiv1.NodeSpec{ProviderID: "nebius://inst-1"},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "node-2"},
			Spec:       apiv1.NodeSpec{ProviderID: "nebius://inst-2"},
		},
	}
	err := ng.DeleteNodes(nodes)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "below minimum size")
	assert.Nil(t, mock.lastUpdateReq)
}

func TestDeleteNodes_MissingProviderID(t *testing.T) {
	t.Parallel()
	mock := &mockNebiusAPI{}
	m := newTestManager(mock)
	ng := &NodeGroup{
		id:         "ng-1",
		manager:    m,
		targetSize: 5,
		minSize:    1,
		maxSize:    10,
	}

	nodes := []*apiv1.Node{
		{ObjectMeta: metav1.ObjectMeta{Name: "node-no-provider-id"}},
	}
	err := ng.DeleteNodes(nodes)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no provider ID")
}

func TestDeleteNodes_DeleteInstanceError(t *testing.T) {
	t.Parallel()
	mock := &mockNebiusAPI{
		deleteInstanceErr: fmt.Errorf("permission denied"),
	}
	m := newTestManager(mock)
	ng := &NodeGroup{
		id:         "ng-1",
		manager:    m,
		targetSize: 5,
		minSize:    1,
		maxSize:    10,
	}

	nodes := []*apiv1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "node-1"},
			Spec:       apiv1.NodeSpec{ProviderID: "nebius://inst-1"},
		},
	}
	err := ng.DeleteNodes(nodes)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete instance")
}

func TestDeleteNodes_PartialFailureAdjustsSize(t *testing.T) {
	t.Parallel()
	mock := &mockNebiusAPI{
		deleteInstanceErr:      fmt.Errorf("instance unavailable"),
		deleteInstanceErrAfter: 1, // first delete succeeds, second fails
		getNodeGroupResponse:   makeNodeGroupProto("ng-1", "group-one", 1, 10, 5),
	}
	m := newTestManager(mock)
	ng := &NodeGroup{
		id:         "ng-1",
		manager:    m,
		targetSize: 5,
		minSize:    1,
		maxSize:    10,
	}

	nodes := []*apiv1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "node-1"},
			Spec:       apiv1.NodeSpec{ProviderID: "nebius://inst-1"},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "node-2"},
			Spec:       apiv1.NodeSpec{ProviderID: "nebius://inst-2"},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "node-3"},
			Spec:       apiv1.NodeSpec{ProviderID: "nebius://inst-3"},
		},
	}
	err := ng.DeleteNodes(nodes)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete instance")

	// One instance was deleted successfully, so target size should be adjusted
	// to account for that deletion (5 - 1 = 4, not the original target of 2).
	require.NotNil(t, mock.lastUpdateReq)
	assert.Equal(t, int64(4), mock.lastUpdateReq.GetSpec().GetFixedNodeCount())

	// In-memory targetSize should also reflect the partial deletion.
	size, sizeErr := ng.TargetSize()
	require.NoError(t, sizeErr)
	assert.Equal(t, 4, size)
}

// --- SetNodeGroupSize tests ---

func TestSetNodeGroupSize_APIError(t *testing.T) {
	t.Parallel()
	mock := &mockNebiusAPI{
		getNodeGroupErr: fmt.Errorf("node group not found"),
	}
	m := newTestManager(mock)
	err := m.setNodeGroupSize("ng-1", 5)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get node group")
}

func TestSetNodeGroupSize_UpdateError(t *testing.T) {
	t.Parallel()
	mock := &mockNebiusAPI{
		getNodeGroupResponse: makeNodeGroupProto("ng-1", "group-one", 1, 10, 3),
		updateNodeGroupErr:   fmt.Errorf("update failed"),
	}
	m := newTestManager(mock)
	err := m.setNodeGroupSize("ng-1", 5)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update node group")
}

// --- DecreaseTargetSize tests ---

func TestDecreaseTargetSize_Success(t *testing.T) {
	t.Parallel()
	mock := &mockNebiusAPI{
		getNodeGroupResponse: makeNodeGroupProto("ng-1", "group-one", 1, 10, 5),
	}
	m := newTestManager(mock)
	ng := &NodeGroup{
		id:         "ng-1",
		manager:    m,
		targetSize: 5,
		minSize:    1,
		maxSize:    10,
	}

	err := ng.DecreaseTargetSize(-2)
	require.NoError(t, err)
	require.NotNil(t, mock.lastUpdateReq)
	assert.Equal(t, int64(3), mock.lastUpdateReq.GetSpec().GetFixedNodeCount())
}

func TestDecreaseTargetSize_PositiveDelta(t *testing.T) {
	t.Parallel()
	mock := &mockNebiusAPI{}
	m := newTestManager(mock)
	ng := &NodeGroup{
		id:         "ng-1",
		manager:    m,
		targetSize: 5,
		minSize:    1,
		maxSize:    10,
	}

	err := ng.DecreaseTargetSize(1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "delta must be negative")
	assert.Nil(t, mock.lastUpdateReq)
}

func TestDecreaseTargetSize_BelowMin(t *testing.T) {
	t.Parallel()
	mock := &mockNebiusAPI{}
	m := newTestManager(mock)
	ng := &NodeGroup{
		id:         "ng-1",
		manager:    m,
		targetSize: 3,
		minSize:    2,
		maxSize:    10,
	}

	err := ng.DecreaseTargetSize(-2)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "decrease target size below minimum")
	assert.Nil(t, mock.lastUpdateReq)
}

// --- NodeGroupForNode tests ---

func TestNodeGroupForNode_ByLabel(t *testing.T) {
	t.Parallel()
	m := newTestManager(&mockNebiusAPI{})
	m.nodeGroups = []*NodeGroup{
		{id: "ng-1", manager: m, nodeGroup: makeNodeGroupProto("ng-1", "group-one", 1, 5, 3)},
		{id: "ng-2", manager: m, nodeGroup: makeNodeGroupProto("ng-2", "group-two", 1, 5, 3)},
	}

	provider := &nebiusCloudProvider{manager: m}
	node := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node",
			Labels: map[string]string{
				nodeGroupIDLabel: "ng-1",
			},
		},
	}

	ng, err := provider.NodeGroupForNode(node)
	require.NoError(t, err)
	require.NotNil(t, ng)
	assert.Equal(t, "ng-1", ng.Id())
}

func TestNodeGroupForNode_ByProviderID(t *testing.T) {
	t.Parallel()
	m := newTestManager(&mockNebiusAPI{})
	m.nodeGroups = []*NodeGroup{
		{
			id:        "ng-1",
			manager:   m,
			nodeGroup: makeNodeGroupProto("ng-1", "group-one", 1, 5, 3),
			instances: map[string]struct{}{
				"nebius://inst-1": {},
				"nebius://inst-2": {},
			},
		},
	}

	provider := &nebiusCloudProvider{manager: m}
	node := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node",
		},
		Spec: apiv1.NodeSpec{
			ProviderID: "nebius://inst-2",
		},
	}

	ng, err := provider.NodeGroupForNode(node)
	require.NoError(t, err)
	require.NotNil(t, ng)
	assert.Equal(t, "ng-1", ng.Id())
}

func TestNodeGroupForNode_NotFound(t *testing.T) {
	t.Parallel()
	m := newTestManager(&mockNebiusAPI{})
	m.nodeGroups = []*NodeGroup{
		{
			id:        "ng-1",
			manager:   m,
			nodeGroup: makeNodeGroupProto("ng-1", "group-one", 1, 5, 3),
			instances: map[string]struct{}{
				"nebius://inst-1": {},
			},
		},
	}

	provider := &nebiusCloudProvider{manager: m}
	node := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node",
		},
		Spec: apiv1.NodeSpec{
			ProviderID: "nebius://unknown-instance",
		},
	}

	ng, err := provider.NodeGroupForNode(node)
	require.NoError(t, err)
	assert.Nil(t, ng)
}
