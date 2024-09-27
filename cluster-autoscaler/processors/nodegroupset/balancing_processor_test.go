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

package nodegroupset

import (
	"testing"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"

	"github.com/stretchr/testify/assert"
)

func buildBasicNodeGroups(context *context.AutoscalingContext) (*framework.NodeInfo, *framework.NodeInfo, *framework.NodeInfo) {
	n1 := BuildTestNode("n1", 1000, 1000)
	n2 := BuildTestNode("n2", 1000, 1000)
	n3 := BuildTestNode("n3", 2000, 2000)
	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 1, 10, 1)
	provider.AddNodeGroup("ng2", 1, 10, 1)
	provider.AddNodeGroup("ng3", 1, 10, 1)
	provider.AddNode("ng1", n1)
	provider.AddNode("ng2", n2)
	provider.AddNode("ng3", n3)

	ni1 := framework.NewTestNodeInfo(n1)
	ni2 := framework.NewTestNodeInfo(n2)
	ni3 := framework.NewTestNodeInfo(n3)

	context.CloudProvider = provider
	return ni1, ni2, ni3
}

func basicSimilarNodeGroupsTest(
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

func TestFindSimilarNodeGroups(t *testing.T) {
	context := &context.AutoscalingContext{}
	ni1, ni2, ni3 := buildBasicNodeGroups(context)
	processor := NewDefaultNodeGroupSetProcessor([]string{}, config.NodeGroupDifferenceRatios{})
	basicSimilarNodeGroupsTest(t, context, processor, ni1, ni2, ni3)
}

func TestFindSimilarNodeGroupsCustomLabels(t *testing.T) {
	context := &context.AutoscalingContext{}
	ni1, ni2, ni3 := buildBasicNodeGroups(context)
	ni1.Node().Labels["example.com/ready"] = "true"
	ni2.Node().Labels["example.com/ready"] = "false"

	processor := NewDefaultNodeGroupSetProcessor([]string{"example.com/ready"}, config.NodeGroupDifferenceRatios{})
	basicSimilarNodeGroupsTest(t, context, processor, ni1, ni2, ni3)
}

func TestFindSimilarNodeGroupsCustomComparator(t *testing.T) {
	context := &context.AutoscalingContext{}
	ni1, ni2, ni3 := buildBasicNodeGroups(context)

	processor := &BalancingNodeGroupSetProcessor{
		Comparator: func(n1, n2 *framework.NodeInfo) bool {
			return (n1.Node().Name == "n1" && n2.Node().Name == "n2") ||
				(n1.Node().Name == "n2" && n2.Node().Name == "n1")
		},
	}
	basicSimilarNodeGroupsTest(t, context, processor, ni1, ni2, ni3)
}

func TestBalanceSingleGroup(t *testing.T) {
	processor := NewDefaultNodeGroupSetProcessor([]string{}, config.NodeGroupDifferenceRatios{})
	context := &context.AutoscalingContext{}

	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 1, 10, 1)

	// just one node
	scaleUpInfo, err := processor.BalanceScaleUpBetweenGroups(context, provider.NodeGroups(), 1)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(scaleUpInfo))
	assert.Equal(t, 2, scaleUpInfo[0].NewSize)

	// multiple nodes
	scaleUpInfo, err = processor.BalanceScaleUpBetweenGroups(context, provider.NodeGroups(), 4)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(scaleUpInfo))
	assert.Equal(t, 5, scaleUpInfo[0].NewSize)
}

func TestBalanceUnderMaxSize(t *testing.T) {
	processor := NewDefaultNodeGroupSetProcessor([]string{}, config.NodeGroupDifferenceRatios{})
	context := &context.AutoscalingContext{}

	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 1, 10, 1)
	provider.AddNodeGroup("ng2", 1, 10, 3)
	provider.AddNodeGroup("ng3", 1, 10, 5)
	provider.AddNodeGroup("ng4", 1, 10, 5)

	// add a single node
	scaleUpInfo, err := processor.BalanceScaleUpBetweenGroups(context, provider.NodeGroups(), 1)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(scaleUpInfo))
	assert.Equal(t, 2, scaleUpInfo[0].NewSize)

	// add multiple nodes to single group
	scaleUpInfo, err = processor.BalanceScaleUpBetweenGroups(context, provider.NodeGroups(), 2)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(scaleUpInfo))
	assert.Equal(t, 3, scaleUpInfo[0].NewSize)

	// add nodes to groups of different sizes, divisible
	scaleUpInfo, err = processor.BalanceScaleUpBetweenGroups(context, provider.NodeGroups(), 4)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(scaleUpInfo))
	assert.Equal(t, 4, scaleUpInfo[0].NewSize)
	assert.Equal(t, 4, scaleUpInfo[1].NewSize)
	assert.True(t, scaleUpInfo[0].Group.Id() == "ng1" || scaleUpInfo[1].Group.Id() == "ng1")
	assert.True(t, scaleUpInfo[0].Group.Id() == "ng2" || scaleUpInfo[1].Group.Id() == "ng2")

	// add nodes to groups of different sizes, non-divisible
	// we expect new sizes to be 4 and 5, doesn't matter which group gets how many
	scaleUpInfo, err = processor.BalanceScaleUpBetweenGroups(context, provider.NodeGroups(), 5)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(scaleUpInfo))
	assert.Equal(t, 9, scaleUpInfo[0].NewSize+scaleUpInfo[1].NewSize)
	assert.True(t, scaleUpInfo[0].NewSize == 4 || scaleUpInfo[0].NewSize == 5)
	assert.True(t, scaleUpInfo[0].Group.Id() == "ng1" || scaleUpInfo[1].Group.Id() == "ng1")
	assert.True(t, scaleUpInfo[0].Group.Id() == "ng2" || scaleUpInfo[1].Group.Id() == "ng2")

	// add nodes to all groups, divisible
	scaleUpInfo, err = processor.BalanceScaleUpBetweenGroups(context, provider.NodeGroups(), 10)
	assert.NoError(t, err)
	assert.Equal(t, 4, len(scaleUpInfo))
	for _, info := range scaleUpInfo {
		assert.Equal(t, 6, info.NewSize)
	}
}

func TestBalanceHittingMaxSize(t *testing.T) {
	processor := NewDefaultNodeGroupSetProcessor([]string{}, config.NodeGroupDifferenceRatios{})
	context := &context.AutoscalingContext{}

	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 1, 1, 1)
	provider.AddNodeGroup("ng2", 1, 3, 1)
	provider.AddNodeGroup("ng3", 1, 10, 3)
	provider.AddNodeGroup("ng4", 1, 7, 5)
	provider.AddNodeGroup("ng5", 1, 3, 6)
	groupsMap := make(map[string]cloudprovider.NodeGroup)
	for _, group := range provider.NodeGroups() {
		groupsMap[group.Id()] = group
	}

	getGroups := func(names ...string) []cloudprovider.NodeGroup {
		result := make([]cloudprovider.NodeGroup, 0)
		for _, n := range names {
			result = append(result, groupsMap[n])
		}
		return result
	}

	toMap := func(suiList []ScaleUpInfo) map[string]ScaleUpInfo {
		result := make(map[string]ScaleUpInfo, 0)
		for _, sui := range suiList {
			result[sui.Group.Id()] = sui
		}
		return result
	}

	// Just one maxed out group
	scaleUpInfo, err := processor.BalanceScaleUpBetweenGroups(context, getGroups("ng1"), 1)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(scaleUpInfo))

	// Smallest group already maxed out, add one node
	scaleUpInfo, err = processor.BalanceScaleUpBetweenGroups(context, getGroups("ng1", "ng2"), 1)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(scaleUpInfo))
	assert.Equal(t, "ng2", scaleUpInfo[0].Group.Id())
	assert.Equal(t, 2, scaleUpInfo[0].NewSize)

	// Smallest group already maxed out, too many nodes (should cap to max capacity)
	scaleUpInfo, err = processor.BalanceScaleUpBetweenGroups(context, getGroups("ng1", "ng2"), 5)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(scaleUpInfo))
	assert.Equal(t, "ng2", scaleUpInfo[0].Group.Id())
	assert.Equal(t, 3, scaleUpInfo[0].NewSize)

	// First group maxes out before proceeding to next one
	scaleUpInfo, _ = processor.BalanceScaleUpBetweenGroups(context, getGroups("ng2", "ng3"), 4)
	assert.Equal(t, 2, len(scaleUpInfo))
	scaleUpMap := toMap(scaleUpInfo)
	assert.Equal(t, 3, scaleUpMap["ng2"].NewSize)
	assert.Equal(t, 5, scaleUpMap["ng3"].NewSize)

	// Last group maxes out before previous one
	scaleUpInfo, _ = processor.BalanceScaleUpBetweenGroups(context, getGroups("ng2", "ng3", "ng4"), 9)
	assert.Equal(t, 3, len(scaleUpInfo))
	scaleUpMap = toMap(scaleUpInfo)
	assert.Equal(t, 3, scaleUpMap["ng2"].NewSize)
	assert.Equal(t, 8, scaleUpMap["ng3"].NewSize)
	assert.Equal(t, 7, scaleUpMap["ng4"].NewSize)

	// Use all capacity, cap to max
	scaleUpInfo, _ = processor.BalanceScaleUpBetweenGroups(context, getGroups("ng2", "ng3", "ng4"), 900)
	assert.Equal(t, 3, len(scaleUpInfo))
	scaleUpMap = toMap(scaleUpInfo)
	assert.Equal(t, 3, scaleUpMap["ng2"].NewSize)
	assert.Equal(t, 10, scaleUpMap["ng3"].NewSize)
	assert.Equal(t, 7, scaleUpMap["ng4"].NewSize)

	// One node group exceeds max.
	scaleUpInfo, _ = processor.BalanceScaleUpBetweenGroups(context, getGroups("ng2", "ng5"), 1)
	assert.Equal(t, 1, len(scaleUpInfo))
	scaleUpMap = toMap(scaleUpInfo)
	assert.Equal(t, 2, scaleUpMap["ng2"].NewSize)
}
