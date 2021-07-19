/*
Copyright 2021 The Kubernetes Authors.

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

package rancher

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/rancher/rancher"
)

func TestNewManager(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cfg := `
				[Global]
				cluster-id=c-jha58
				secret=my-token
				access=my-access
				url=https://auks/v3
		`
		_, err := newManager(strings.NewReader(cfg))
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("failed: missing url", func(t *testing.T) {
		cfg := `
				[Global]
				cluster-id=c-jha58
				secret=my-secret
				access=my-access
		`
		_, err := newManager(strings.NewReader(cfg))
		if !errors.Is(err, errURLRequired) {
			t.Errorf("expected error: %v got %v", errURLRequired, err)
		}
	})

	t.Run("failed: missing secret", func(t *testing.T) {
		cfg := `
				[Global]
				cluster-id=c-jha58
				url=https://auks/v3
				access=my-access
		`
		_, err := newManager(strings.NewReader(cfg))
		if !errors.Is(err, errSecretRequired) {
			t.Errorf("expected error: %v got %v", errSecretRequired, err)
		}
	})

	t.Run("failed: missing access", func(t *testing.T) {
		cfg := `
				[Global]
				cluster-id=c-jha58
				url=https://auks/v3
				secret=my-secret
		`
		_, err := newManager(strings.NewReader(cfg))
		if !errors.Is(err, errAccessRequired) {
			t.Errorf("expected error: %v got %v", errAccessRequired, err)
		}
	})

	t.Run("failed: missing clusterID", func(t *testing.T) {
		cfg := `
				[Global]
				secret=my-secret
				access=my-access
				url=https://auks/v3
		`
		_, err := newManager(strings.NewReader(cfg))
		if !errors.Is(err, errClusterIDRequired) {
			t.Errorf("expected error: %v got %v", errClusterIDRequired, err)
		}
	})
}

func TestManager_getNodePools(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var cli clientMock
		nodePools := []rancher.NodePool{
			{ID: "1", Name: "worker-1"},
			{ID: "2", Name: "worker-2"},
			{ID: "3", Name: "worker-3"},
		}

		cli.nodePoolsByClusterFn = func(clusterID string) ([]rancher.NodePool, error) {
			return nodePools, nil
		}

		manager := manager{client: &cli}
		nps, err := manager.getNodePools()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(nps) != len(nodePools) {
			t.Errorf("got %d expected %d", len(nps), len(nodePools))
		}
	})

	t.Run("failed", func(t *testing.T) {
		var cli clientMock
		cli.nodePoolsByClusterFn = func(clusterID string) ([]rancher.NodePool, error) {
			return nil, fmt.Errorf("cluster does not have nodePools")
		}

		manager := manager{client: &cli}
		_, err := manager.getNodePools()
		if err == nil {
			t.Errorf("expected error, got %v", err)
		}
	})
}

func TestManager_getNode(t *testing.T) {
	t.Run("get node by providerID - success", func(t *testing.T) {
		var cli clientMock
		cli.nodeByProviderIDFn = func(providerID string) (*rancher.Node, error) {
			if providerID == "openstack:///gerg-fergerg-ergreg" {
				return &rancher.Node{Name: providerID}, nil
			}
			return nil, fmt.Errorf("node %q does not exist", providerID)
		}

		manager := manager{client: &cli}
		node, err := manager.getNode(&apiv1.Node{Spec: apiv1.NodeSpec{ProviderID: "openstack:///gerg-fergerg-ergreg"}})
		if err != nil {
			t.Errorf("unexpected error, got %v", err)
		}

		if node == nil {
			t.Errorf("expected a node, got %v", node)
		}
	})

	t.Run("get node by providerID - failed ", func(t *testing.T) {
		var cli clientMock
		cli.nodeByProviderIDFn = func(providerID string) (*rancher.Node, error) {
			return nil, fmt.Errorf("node %q does not exist", providerID)
		}

		manager := manager{client: &cli}
		_, err := manager.getNode(&apiv1.Node{Spec: apiv1.NodeSpec{ProviderID: "openstack:///gerg-fergerg-ergreg"}})
		if err == nil {
			t.Error("expected error")
		}
	})
}

type clientMock struct {
	clusterByIDFn          func(id string) (*rancher.Cluster, error)
	resizeNodePoolFn       func(id string, size int) (*rancher.NodePool, error)
	nodePoolsByClusterFn   func(clusterID string) ([]rancher.NodePool, error)
	nodePoolByIDFN         func(id string) (*rancher.NodePool, error)
	nodesByNodePoolFn      func(nodePoolID string) ([]rancher.Node, error)
	nodeByProviderIDFn     func(providerID string) (*rancher.Node, error)
	deleteNodeFn           func(id string) error
	nodeByNameAndClusterFn func(name, cluster string) (*rancher.Node, error)
	ScaleDownNodeFn        func(nodeID string) error
}

func (s *clientMock) ResizeNodePool(id string, size int) (*rancher.NodePool, error) {
	return s.resizeNodePoolFn(id, size)
}

func (s *clientMock) NodePoolsByCluster(clusterID string) ([]rancher.NodePool, error) {
	return s.nodePoolsByClusterFn(clusterID)
}

func (s *clientMock) NodePoolByID(id string) (*rancher.NodePool, error) {
	return s.nodePoolByIDFN(id)
}

func (s *clientMock) NodesByNodePool(nodePoolID string) ([]rancher.Node, error) {
	return s.nodesByNodePoolFn(nodePoolID)
}

func (s clientMock) NodeByProviderID(providerID string) (*rancher.Node, error) {
	return s.nodeByProviderIDFn(providerID)
}

func (s clientMock) DeleteNode(id string) error {
	return s.deleteNodeFn(id)
}

func (s clientMock) NodeByNameAndCluster(name, cluster string) (*rancher.Node, error) {
	return s.nodeByNameAndClusterFn(name, cluster)
}

func (s clientMock) ClusterByID(id string) (*rancher.Cluster, error) {
	return s.clusterByIDFn(id)
}

func (s clientMock) ScaleDownNode(nodeID string) error {
	return s.ScaleDownNodeFn(nodeID)
}
