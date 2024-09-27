/*
Copyright 2024 The Kubernetes Authors.

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

package nodegroupset

import (
	"testing"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	testutils "k8s.io/autoscaler/cluster-autoscaler/utils/test"

	"github.com/stretchr/testify/assert"
)

// CheckNodesSimilar is a helper func for tests to validate node comparison outcomes
func CheckNodesSimilar(t *testing.T, n1, n2 *apiv1.Node, comparator NodeInfoComparator, shouldEqual bool) {
	CheckNodesSimilarWithPods(t, n1, n2, []*apiv1.Pod{}, []*apiv1.Pod{}, comparator, shouldEqual)
}

// CheckNodesSimilarWithPods is a helper func for tests to validate nodes with pods comparison outcomes
func CheckNodesSimilarWithPods(t *testing.T, n1, n2 *apiv1.Node, pods1, pods2 []*apiv1.Pod, comparator NodeInfoComparator, shouldEqual bool) {
	ni1 := framework.NewTestNodeInfo(n1, pods1...)
	ni2 := framework.NewTestNodeInfo(n2, pods2...)
	assert.Equal(t, shouldEqual, comparator(ni1, ni2))
}

// BuildBasicNodeGroups is a helper func for tests to get a set of NodeInfo objects
func BuildBasicNodeGroups(context *context.AutoscalingContext) (*framework.NodeInfo, *framework.NodeInfo, *framework.NodeInfo) {
	n1 := testutils.BuildTestNode("n1", 1000, 1000)
	n2 := testutils.BuildTestNode("n2", 1000, 1000)
	n3 := testutils.BuildTestNode("n3", 2000, 2000)
	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 1, 10, 1)
	provider.AddNodeGroup("ng2", 1, 10, 1)
	provider.AddNodeGroup("ng3", 1, 10, 1)
	provider.AddNode("ng1", n1)
	provider.AddNode("ng2", n2)
	provider.AddNode("ng3", n3)

	ni1 := framework.NewNodeInfo(n1, nil)
	ni2 := framework.NewNodeInfo(n2, nil)
	ni3 := framework.NewNodeInfo(n3, nil)

	context.CloudProvider = provider
	return ni1, ni2, ni3
}

// BasicSimilarNodeGroupsTest is a helper func for tests to assert node group similarity
func BasicSimilarNodeGroupsTest(
	t *testing.T,
	context *context.AutoscalingContext,
	processor NodeGroupSetProcessor,
	ni1 *framework.NodeInfo,
	ni2 *framework.NodeInfo,
	ni3 *framework.NodeInfo,
) {
	nodeInfosForGroups := map[string]*framework.NodeInfo{
		"ng1": ni1, "ng2": ni2, "ng3": ni3,
	}

	ng1, _ := context.CloudProvider.NodeGroupForNode(ni1.Node())
	ng2, _ := context.CloudProvider.NodeGroupForNode(ni2.Node())
	ng3, _ := context.CloudProvider.NodeGroupForNode(ni3.Node())

	similar, err := processor.FindSimilarNodeGroups(context, ng1, nodeInfosForGroups)
	assert.NoError(t, err)
	assert.Equal(t, []cloudprovider.NodeGroup{ng2}, similar)

	similar, err = processor.FindSimilarNodeGroups(context, ng2, nodeInfosForGroups)
	assert.NoError(t, err)
	assert.Equal(t, []cloudprovider.NodeGroup{ng1}, similar)

	similar, err = processor.FindSimilarNodeGroups(context, ng3, nodeInfosForGroups)
	assert.NoError(t, err)
	assert.Equal(t, []cloudprovider.NodeGroup{}, similar)
}
