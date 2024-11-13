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
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"

	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/utils"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

func TestPreFilteringScaleDownNodeProcessor_GetPodDestinationCandidates(t *testing.T) {
	n1 := BuildTestNode("n1", 100, 1000)
	n2 := BuildTestNode("n2", 100, 1000)
	n3 := BuildTestNode("n3", 100, 1000)
	n3.Annotations = map[string]string{utils.NodeUpcomingAnnotation: "true"}
	ctx := &context.AutoscalingContext{}
	defaultProcessor := NewPreFilteringScaleDownNodeProcessor()
	expectedNodes := []*apiv1.Node{n1, n2}
	nodes := []*apiv1.Node{n1, n2, n3}
	nodes, err := defaultProcessor.GetPodDestinationCandidates(ctx, nodes)

	assert.NoError(t, err)
	assert.Equal(t, expectedNodes, nodes)
}

func TestPreFilteringScaleDownNodeProcessor_GetScaleDownCandidateNodes(t *testing.T) {
	ng1_1 := BuildTestNode("ng1-1", 1000, 1000)
	ng1_2 := BuildTestNode("ng1-2", 1000, 1000)
	ng2_1 := BuildTestNode("ng2-1", 1000, 1000)
	noNg := BuildTestNode("no-ng", 1000, 1000)
	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 1, 10, 2)
	provider.AddNodeGroup("ng2", 1, 10, 1)
	provider.AddNode("ng1", ng1_1)
	provider.AddNode("ng1", ng1_2)
	provider.AddNode("ng2", ng2_1)

	ctx := &context.AutoscalingContext{
		CloudProvider: provider,
	}

	expectedNodes := []*apiv1.Node{ng1_1, ng1_2}
	defaultProcessor := NewPreFilteringScaleDownNodeProcessor()
	inputNodes := []*apiv1.Node{ng1_1, ng1_2, ng2_1, noNg}
	result, err := defaultProcessor.GetScaleDownCandidates(ctx, inputNodes)

	assert.NoError(t, err)
	assert.Equal(t, result, expectedNodes)
}
