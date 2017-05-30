/*
Copyright 2017 The Kubernetes Authors.

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

package price

import (
	"testing"

	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	apiv1 "k8s.io/kubernetes/pkg/api/v1"

	"github.com/stretchr/testify/assert"
)

type testNodeLister struct {
	list []*apiv1.Node
}

func (n *testNodeLister) List() ([]*apiv1.Node, error) {
	return n.list, nil
}

func TestPreferred(t *testing.T) {
	n1 := BuildTestNode("n1", 1000, 1000)
	n2 := BuildTestNode("n2", 2000, 2000)
	n3 := BuildTestNode("n2", 2000, 2000)

	provider := simplePreferredNodeProvider{
		nodeLister: &testNodeLister{
			list: []*apiv1.Node{n1, n2, n3},
		},
	}
	node, err := provider.Node()
	assert.NoError(t, err)
	cpu := node.Status.Capacity[apiv1.ResourceCPU]
	assert.Equal(t, int64(2), cpu.Value())
	assert.Equal(t, 2.0, simpleNodeUnfitness(n1, n2))
	assert.Equal(t, 2.0, simpleNodeUnfitness(n2, n1))
	assert.Equal(t, 1.0, simpleNodeUnfitness(n1, n1))
}
