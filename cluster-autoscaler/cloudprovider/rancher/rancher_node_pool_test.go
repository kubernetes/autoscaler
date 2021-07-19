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
	"fmt"
	"testing"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/rancher/rancher"
)

func TestNodePool_MaxSize(t *testing.T) {
	maxSize := 5
	np := NodePool{maxSize: maxSize}
	if np.MaxSize() != maxSize {
		t.Errorf("got %d expected %d", np.MaxSize(), maxSize)
	}
}

func TestNodePool_MinSize(t *testing.T) {
	minSize := 2
	np := NodePool{minSize: minSize}
	if np.MinSize() != minSize {
		t.Errorf("got %d expected %d", np.MaxSize(), minSize)
	}
}

func TestNodePool_Nodes(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var cli clientMock
		rancherNodes := []rancher.Node{
			{Name: "worker1"},
			{Name: "worker2"},
			{Name: "worker3"},
		}

		cli.nodesByNodePoolFn = func(nodePoolID string) ([]rancher.Node, error) {
			return rancherNodes, nil
		}

		np := NodePool{manager: &manager{client: &cli}}
		nodes, err := np.Nodes()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(nodes) != len(rancherNodes) {
			t.Errorf("got %d expected %d", len(nodes), len(rancherNodes))
		}
	})

	t.Run("failed", func(t *testing.T) {
		var cli clientMock
		cli.nodesByNodePoolFn = func(nodePoolID string) ([]rancher.Node, error) {
			return nil, fmt.Errorf("client error")
		}

		np := NodePool{manager: &manager{client: &cli}}
		nodes, err := np.Nodes()
		if err == nil {
			t.Errorf("expected error")
		}

		if len(nodes) != 0 {
			t.Errorf("got %d expected %d", len(nodes), 0)
		}
	})
}

func TestNodePool_IncreaseSize(t *testing.T) {
	t.Run("increase from 0 to 5 - success", func(t *testing.T) {
		var cli clientMock
		cli.resizeNodePoolFn = func(id string, size int) (*rancher.NodePool, error) {
			return &rancher.NodePool{Quantity: size}, nil
		}

		targetSize := 5
		manager := manager{client: &cli}
		np := NodePool{manager: &manager, minSize: 0, maxSize: 5}
		if err := np.IncreaseSize(targetSize); err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if np.rancherNP.Quantity != targetSize {
			t.Errorf("got %d expected %d", np.rancherNP.Quantity, targetSize)
		}
	})

	t.Run("negative increase - failed", func(t *testing.T) {
		np := NodePool{minSize: 0, maxSize: 5}
		if err := np.IncreaseSize(-3); err == nil {
			t.Error("expected error")
		}
	})

	t.Run("increase above maximum - failed", func(t *testing.T) {
		np := NodePool{minSize: 0, maxSize: 5}
		if err := np.IncreaseSize(10); err == nil {
			t.Error("expected error")
		}
	})
}

func TestNodePool_DecreaseTargetSize(t *testing.T) {
	t.Run("decrease from 6 to 2 - success", func(t *testing.T) {
		var cli clientMock
		cli.resizeNodePoolFn = func(id string, size int) (*rancher.NodePool, error) {
			return &rancher.NodePool{Quantity: size}, nil
		}

		currentQty := 6
		decreaseBy := 4
		manager := manager{client: &cli}
		np := NodePool{manager: &manager, minSize: 2, maxSize: 10, rancherNP: rancher.NodePool{Quantity: currentQty}}
		if err := np.DecreaseTargetSize(decreaseBy); err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if np.rancherNP.Quantity != currentQty-decreaseBy {
			t.Errorf("got %d expected %d", np.rancherNP.Quantity, currentQty-decreaseBy)
		}
	})

	t.Run("decrease over minimum - failed", func(t *testing.T) {
		np := NodePool{minSize: 2, maxSize: 5, rancherNP: rancher.NodePool{Quantity: 4}}
		if err := np.DecreaseTargetSize(3); err == nil {
			t.Error("expected error")
		}
	})
}
