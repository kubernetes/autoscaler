/*
Copyright 2020-2023 Oracle and/or its affiliates.
*/

package nodepools

import (
	"context"
	"testing"

	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	ocicommon "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/common"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

func TestDeletePastMinSize(t *testing.T) {
	client := fake.NewSimpleClientset()

	nodeNames := []string{
		"nodeA",
	}
	manager := &mockManager{
		err:        nil,
		timeOutErr: apierrors.NewTimeoutError("timeout error", 5),
		size:       1,
	}
	np := &nodePool{
		kubeClient: client,
		manager:    manager,
		minSize:    1,
		maxSize:    10,
		id:         "abc",
	}
	manager.nodePool = np

	var nodesToDelete []*apiv1.Node
	for _, name := range nodeNames {
		node, err := client.CoreV1().Nodes().Create(context.Background(), &apiv1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		}, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("unexpected create error: %+v", err)
		}
		nodesToDelete = append(nodesToDelete, node)
	}
	err := np.DeleteNodes(nodesToDelete)
	if err == nil {
		t.Fatalf("expected to have an error because node pool is at the min size")
	}
}

func TestDeleteCreateErrorNodeWithoutInstanceIDDecreasesTargetSize(t *testing.T) {
	client := fake.NewSimpleClientset()

	manager := &mockManager{
		size:                           1,
		existingNodePoolSizeViaCompute: 0,
		nodes: []cloudprovider.Instance{
			{
				Status: &cloudprovider.InstanceStatus{
					State: cloudprovider.InstanceCreating,
					ErrorInfo: &cloudprovider.InstanceErrorInfo{
						ErrorClass:   cloudprovider.OutOfResourcesErrorClass,
						ErrorCode:    "QuotaExceeded",
						ErrorMessage: "quota exceeded",
					},
				},
			},
		},
	}
	np := &nodePool{
		kubeClient: client,
		manager:    manager,
		minSize:    0,
		maxSize:    10,
		id:         "abc",
	}
	manager.nodePool = np

	nodeWithoutInstanceID := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "",
			Annotations: map[string]string{
				cloudprovider.FakeNodeReasonAnnotation: cloudprovider.FakeNodeCreateError,
			},
		},
	}

	if err := np.DeleteNodes([]*apiv1.Node{nodeWithoutInstanceID}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if manager.deleteInstancesCalled != 0 {
		t.Fatalf("expected DeleteInstances not to be called, got %d calls", manager.deleteInstancesCalled)
	}
	if !manager.invalidateAndRefreshCacheCalled {
		t.Fatalf("expected cache invalidation before decreasing target size")
	}
	if !manager.setSizeCalled {
		t.Fatalf("expected SetNodePoolSize to be called")
	}
	if manager.setSize != 0 {
		t.Fatalf("expected target size to be decreased to 0, got %d", manager.setSize)
	}
}

func TestHasInstanceForUnmanagedNodePool(t *testing.T) {
	provider := NewOciCloudProvider(&mockManager{err: errInstanceNodePoolNotFound}, nil)
	node := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "unmanaged-node",
			Annotations: map[string]string{
				"oci.oraclecloud.com/node-pool-id": "unmanaged-pool",
			},
		},
	}

	hasInstance, err := provider.HasInstance(node)
	if err != nil {
		t.Fatalf("HasInstance() returned an error for an unmanaged node pool: %v", err)
	}
	if !hasInstance {
		t.Fatal("HasInstance() = false for an unmanaged node pool, want true")
	}
}

type mockManager struct {
	called                          []string
	nodePools                       []NodePool
	nodes                           []cloudprovider.Instance
	size                            int
	setSize                         int
	setSizeCalled                   bool
	deleteInstancesCalled           int
	existingNodePoolSizeViaCompute  int
	invalidateAndRefreshCacheCalled bool

	// used for GetNodePoolForInstance
	nodePool     NodePool
	templateNode *apiv1.Node
	NodePoolManager
	err        error
	timeOutErr error
}

func (m *mockManager) Refresh() error {
	m.called = append(m.called, "refresh")
	return nil
}

func (m *mockManager) Cleanup() error {
	m.called = append(m.called, "cleanup")
	return nil
}

func (m *mockManager) GetNodePools() []NodePool {
	m.called = append(m.called, "get-node-pools")
	return m.nodePools
}

func (m *mockManager) GetNodePoolNodes(np NodePool) ([]cloudprovider.Instance, error) {
	m.called = append(m.called, "get-node-pool-nodes")
	return m.nodes, nil
}

func (m *mockManager) GetNodePoolNodePrivateIP(np NodePool) (string, error) {
	m.called = append(m.called, "get-node-pool-node-private-ip")
	return "10.0.0.10", m.err
}

func (m *mockManager) GetExistingNodePoolSizeViaCompute(np NodePool) (int, error) {
	m.called = append(m.called, "get-existing-node-pool-size-via-compute")
	return m.existingNodePoolSizeViaCompute, nil
}

func (m *mockManager) GetNodePoolForInstance(instance ocicommon.OciRef) (NodePool, error) {
	m.called = append(m.called, "get-node-pool-for-instance")
	return m.nodePool, m.err
}

func (m *mockManager) GetNodePoolTemplateNode(np NodePool) (*apiv1.Node, error) {
	m.called = append(m.called, "get-node-pool-template-node")
	if m.templateNode != nil {
		return m.templateNode.DeepCopy(), m.err
	}
	panic("implement me")
}

func (m *mockManager) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	m.called = append(m.called, "get-resource-limiter")
	panic("implement me")
}

func (m *mockManager) GetNodePoolSize(np NodePool) (int, error) {
	m.called = append(m.called, "get-node-pool-size")
	if m.size != 0 {
		return m.size, nil
	}
	return np.MinSize() + 1, nil
}

func (m *mockManager) SetNodePoolSize(np NodePool, size int) error {
	m.called = append(m.called, "set-node-pool-size")
	m.setSize = size
	m.setSizeCalled = true
	return nil
}

func (m *mockManager) DeleteInstances(np NodePool, instances []ocicommon.OciRef) error {
	m.called = append(m.called, "delete-instances")
	m.deleteInstancesCalled++
	return m.timeOutErr
}

func (m *mockManager) InvalidateAndRefreshCache() error {
	m.called = append(m.called, "invalidate-and-refresh-cache")
	m.invalidateAndRefreshCacheCalled = true
	return nil
}

func (m *mockManager) TaintToPreventFurtherSchedulingOnRestart(nodes []*apiv1.Node, client kubernetes.Interface) error {
	m.called = append(m.called, "taint-to-prevent-further-scheduling-on-restart")
	return nil
}

func TestSetEphemeralStorageFromRegisteredNodeCopiesResources(t *testing.T) {
	ctx := context.Background()
	registeredNode := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "registered-node",
			Labels: map[string]string{
				"internal_addr": "10.0.0.10",
			},
			Annotations: map[string]string{
				"oci.oraclecloud.com/node-pool-id": "pool-1",
			},
		},
		Status: apiv1.NodeStatus{
			Capacity: apiv1.ResourceList{
				apiv1.ResourceEphemeralStorage: resource.MustParse("30Gi"),
			},
			Allocatable: apiv1.ResourceList{
				apiv1.ResourceEphemeralStorage: resource.MustParse("25Gi"),
			},
		},
	}
	client := fake.NewSimpleClientset(registeredNode)
	np := &nodePool{
		id:         "pool-1",
		manager:    &mockManager{},
		kubeClient: client,
	}
	template := &apiv1.Node{
		Status: apiv1.NodeStatus{
			Capacity: apiv1.ResourceList{
				apiv1.ResourceEphemeralStorage: resource.MustParse("10Gi"),
			},
			Allocatable: apiv1.ResourceList{
				apiv1.ResourceEphemeralStorage: resource.MustParse("8Gi"),
			},
		},
	}

	if err := np.setEphemeralStorageFromRegisteredNode(ctx, template); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := template.Status.Capacity[apiv1.ResourceEphemeralStorage]; got.Cmp(resource.MustParse("30Gi")) != 0 {
		t.Errorf("capacity ephemeral-storage = %v, want 30Gi", got.String())
	}
	if got := template.Status.Allocatable[apiv1.ResourceEphemeralStorage]; got.Cmp(resource.MustParse("25Gi")) != 0 {
		t.Errorf("allocatable ephemeral-storage = %v, want 25Gi", got.String())
	}
}

func TestSetEphemeralStorageFromRegisteredNodeWrongPoolPreservesResources(t *testing.T) {
	ctx := context.Background()
	registeredNode := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "registered-node",
			Labels: map[string]string{
				"internal_addr": "10.0.0.10",
			},
			Annotations: map[string]string{
				"oci.oraclecloud.com/node-pool-id": "different-pool",
			},
		},
		Status: apiv1.NodeStatus{
			Capacity: apiv1.ResourceList{
				apiv1.ResourceEphemeralStorage: resource.MustParse("30Gi"),
			},
			Allocatable: apiv1.ResourceList{
				apiv1.ResourceEphemeralStorage: resource.MustParse("25Gi"),
			},
		},
	}
	client := fake.NewSimpleClientset(registeredNode)
	np := &nodePool{
		id:         "pool-1",
		manager:    &mockManager{},
		kubeClient: client,
	}
	template := &apiv1.Node{
		Status: apiv1.NodeStatus{
			Capacity: apiv1.ResourceList{
				apiv1.ResourceEphemeralStorage: resource.MustParse("10Gi"),
			},
			Allocatable: apiv1.ResourceList{
				apiv1.ResourceEphemeralStorage: resource.MustParse("8Gi"),
			},
		},
	}

	if err := np.setEphemeralStorageFromRegisteredNode(ctx, template); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := template.Status.Capacity[apiv1.ResourceEphemeralStorage]; got.Cmp(resource.MustParse("10Gi")) != 0 {
		t.Errorf("capacity ephemeral-storage = %v, want unchanged 10Gi", got.String())
	}
	if got := template.Status.Allocatable[apiv1.ResourceEphemeralStorage]; got.Cmp(resource.MustParse("8Gi")) != 0 {
		t.Errorf("allocatable ephemeral-storage = %v, want unchanged 8Gi", got.String())
	}
}

func TestSetEphemeralStorageFromRegisteredNodeLookupErrorPreservesResources(t *testing.T) {
	ctx := context.Background()
	np := &nodePool{
		id:         "pool-1",
		manager:    &mockManager{err: apierrors.NewBadRequest("test lookup failure")},
		kubeClient: fake.NewSimpleClientset(),
	}
	template := &apiv1.Node{
		Status: apiv1.NodeStatus{
			Capacity: apiv1.ResourceList{
				apiv1.ResourceEphemeralStorage: resource.MustParse("10Gi"),
			},
			Allocatable: apiv1.ResourceList{
				apiv1.ResourceEphemeralStorage: resource.MustParse("8Gi"),
			},
		},
	}

	if err := np.setEphemeralStorageFromRegisteredNode(ctx, template); err == nil {
		t.Fatal("expected private IP lookup error")
	}

	if got := template.Status.Capacity[apiv1.ResourceEphemeralStorage]; got.Cmp(resource.MustParse("10Gi")) != 0 {
		t.Errorf("capacity ephemeral-storage = %v, want unchanged 10Gi", got.String())
	}
	if got := template.Status.Allocatable[apiv1.ResourceEphemeralStorage]; got.Cmp(resource.MustParse("8Gi")) != 0 {
		t.Errorf("allocatable ephemeral-storage = %v, want unchanged 8Gi", got.String())
	}
}
