/*
Copyright 2019 The Kubernetes Authors.

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

package nodes

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"

	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

func TestPreFilteringScaleDownNodeProcessor_GetPodDestinationCandidates(t *testing.T) {
	n1 := BuildTestNode("n1", 100, 1000)
	n2 := BuildTestNode("n2", 100, 1000)
	ctx := &context.AutoscalingContext{}
	defaultProcessor := NewPreFilteringScaleDownNodeProcessor()
	expectedNodes := []*apiv1.Node{n1, n2}
	nodes := []*apiv1.Node{n1, n2}
	nodes, err := defaultProcessor.GetPodDestinationCandidates(ctx, nodes)

	assert.NoError(t, err)
	assert.Equal(t, nodes, expectedNodes)
}

func TestPreFilteringScaleDownNodeProcessor_GetScaleDownCandidateNodes(t *testing.T) {
	ng1_1 := BuildTestNode("ng1-1", 1000, 1000)
	ng1_2 := BuildTestNode("ng1-2", 1000, 1000)
	ng2_1 := BuildTestNode("ng2-1", 1000, 1000)
	ng2_2 := BuildTestNode("ng2-2", 1000, 1000)
	noNg := BuildTestNode("no-ng", 1000, 1000)

	testCases := map[string]struct {
		buildProvider     func() *testprovider.TestCloudProvider
		configureProvider func(p *testprovider.TestCloudProvider)
		expectedNodes     []*apiv1.Node
		inputNodes        []*apiv1.Node
	}{
		// Expectation: only node groups not at minimum size should be candidates.
		"1 scale down candidate, 1 node group at minimum size, 1 node with no node group, 1 node group above minimum size.": {
			configureProvider: func(p *testprovider.TestCloudProvider) {
				p.AddNodeGroup("ng1", 1, 10, 2)
				p.AddNodeGroup("ng2", 1, 10, 1)
				p.AddNode("ng1", ng1_1)
				p.AddNode("ng1", ng1_2)
				p.AddNode("ng2", ng2_1)
			},
			expectedNodes: []*apiv1.Node{ng1_1, ng1_2},
			inputNodes:    []*apiv1.Node{ng1_1, ng1_2, ng2_1, noNg},
		},
		// Expectation: only node groups that contain nodes the cloud provider considers candidates for deletion should be candidates.
		"1 scale down candidate, 1 node group with nodes that are not candidates for deletion, 1 node group above minimum size.": {
			buildProvider: func() *testprovider.TestCloudProvider {
				provider := testprovider.
					NewTestCloudProviderBuilder().
					WithIsNodeCandidateForScaleDown(func(n *apiv1.Node) (bool, error) {
						if strings.HasPrefix(n.Name, "ng2") {
							return false, nil
						}
						return true, nil
					}).
					Build()
				return provider
			},
			configureProvider: func(p *testprovider.TestCloudProvider) {
				p.AddNodeGroup("ng1", 1, 10, 2)
				p.AddNodeGroup("ng2", 1, 10, 2)
				p.AddNode("ng1", ng1_1)
				p.AddNode("ng1", ng1_2)
				p.AddNode("ng2", ng2_1)
				p.AddNode("ng2", ng2_2)
			},
			expectedNodes: []*apiv1.Node{ng1_1, ng1_2},
			inputNodes:    []*apiv1.Node{ng1_1, ng1_2, ng2_1, ng2_2},
		},
	}

	for description, testCase := range testCases {
		t.Run(description, func(t *testing.T) {
			var provider *testprovider.TestCloudProvider
			if testCase.buildProvider == nil {
				provider = testprovider.NewTestCloudProviderBuilder().Build()
			} else {
				provider = testCase.buildProvider()
			}
			assert.NotNil(t, provider)

			testCase.configureProvider(provider)

			ctx := &context.AutoscalingContext{
				CloudProvider: provider,
			}

			defaultProcessor := NewPreFilteringScaleDownNodeProcessor()
			result, err := defaultProcessor.GetScaleDownCandidates(ctx, testCase.inputNodes)

			assert.NoError(t, err)
			assert.Equal(t, result, testCase.expectedNodes)
		})
	}
}
