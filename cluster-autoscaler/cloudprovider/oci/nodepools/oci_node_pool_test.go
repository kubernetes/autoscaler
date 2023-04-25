/*
Copyright 2020-2023 Oracle and/or its affiliates.
*/

package nodepools

import (
	"context"
	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	ocicommon "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/common"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
)

func TestDeletePastMinSize(t *testing.T) {
	client := fake.NewSimpleClientset()

	nodeNames := []string{
		"nodeA",
	}
	manager := &mockManager{
		err:        nil,
		timeOutErr: apierrors.NewTimeoutError("timeout error", 5),
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

type mockManager struct {
	called    []string
	nodePools []NodePool
	nodes     []cloudprovider.Instance

	// used for GetNodePoolForInstance
	nodePool NodePool
	NodePoolManager
	err        error
	timeOutErr error
}

func (m mockManager) Refresh() error {
	m.called = append(m.called, "refresh")
	return nil
}

func (m mockManager) Cleanup() error {
	m.called = append(m.called, "cleanup")
	return nil
}

func (m mockManager) GetNodePools() []NodePool {
	m.called = append(m.called, "get-node-pools")
	return m.nodePools
}

func (m mockManager) GetNodePoolNodes(np NodePool) ([]cloudprovider.Instance, error) {
	m.called = append(m.called, "get-node-pool-nodes")
	return m.nodes, nil
}

func (m mockManager) GetNodePoolForInstance(instance ocicommon.OciRef) (NodePool, error) {
	m.called = append(m.called, "get-node-pool-for-instance")
	return m.nodePool, m.err
}

func (m mockManager) GetNodePoolTemplateNode(np NodePool) (*apiv1.Node, error) {
	m.called = append(m.called, "get-node-pool-template-node")
	panic("implement me")
}

func (m mockManager) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	m.called = append(m.called, "get-resource-limiter")
	panic("implement me")
}

func (m mockManager) GetNodePoolSize(np NodePool) (int, error) {
	m.called = append(m.called, "get-node-pool-size")
	return np.MinSize() + 1, nil
}

func (m mockManager) SetNodePoolSize(np NodePool, size int) error {
	m.called = append(m.called, "set-node-pool-size")
	panic("implement me")
}

func (m mockManager) DeleteInstances(np NodePool, instances []ocicommon.OciRef) error {
	m.called = append(m.called, "delete-instances")
	return m.timeOutErr
}
