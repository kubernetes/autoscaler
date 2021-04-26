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

	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/rancher/rancher"
)

func TestRancherProvider_NodeGroupForNode(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var cli clientMock
		nodePool := NodePool{id: "pool1"}
		cli.nodeByProviderIDFn = func(providerID string) (*rancher.Node, error) {
			return &rancher.Node{Name: "worker1", NodePoolID: nodePool.id}, nil
		}

		manager := manager{client: &cli}
		u := rancherProvider{manager: &manager, nodePools: []*NodePool{&nodePool}}
		np, err := u.NodeGroupForNode(&apiv1.Node{ObjectMeta: v1.ObjectMeta{Name: "worker1"}})
		if err != nil {
			t.Errorf("unexpected error %v", err)
		}

		if np.Id() != nodePool.id {
			t.Errorf("got %s expected %s", np.Id(), nodePool.id)
		}
	})

	t.Run("node does not exist - failed", func(t *testing.T) {
		var cli clientMock
		cli.nodeByProviderIDFn = func(providerID string) (*rancher.Node, error) {
			return nil, fmt.Errorf("node %q does not exist", providerID)
		}

		manager := manager{client: &cli}
		u := rancherProvider{manager: &manager}
		_, err := u.NodeGroupForNode(&apiv1.Node{ObjectMeta: v1.ObjectMeta{Name: "worker1"}})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("node belongs to nodePool without auto-scale", func(t *testing.T) {
		var cli clientMock
		nodePool := NodePool{id: "pool1"}
		cli.nodeByProviderIDFn = func(providerID string) (*rancher.Node, error) {
			return &rancher.Node{Name: "worker1", NodePoolID: "pool2"}, nil
		}

		manager := manager{client: &cli}
		u := rancherProvider{manager: &manager, nodePools: []*NodePool{&nodePool}}
		np, err := u.NodeGroupForNode(&apiv1.Node{ObjectMeta: v1.ObjectMeta{Name: "worker1"}})
		if err != nil {
			t.Errorf("unexpected error %v", err)
		}

		if np != nil {
			t.Errorf("unexpected value, %v", np)
		}
	})

}
