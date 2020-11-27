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

package nodegroupset

import (
	"testing"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"

	"github.com/stretchr/testify/assert"
)

func TestIsAzureNodeInfoSimilar(t *testing.T) {
	comparator := CreateAzureNodeInfoComparator([]string{"example.com/ready"})
	n1 := BuildTestNode("node1", 1000, 2000)
	n1.ObjectMeta.Labels["test-label"] = "test-value"
	n1.ObjectMeta.Labels["character"] = "thing"
	n2 := BuildTestNode("node2", 1000, 2000)
	n2.ObjectMeta.Labels["test-label"] = "test-value"
	// No node-pool labels.
	checkNodesSimilar(t, n1, n2, comparator, false)
	// Empty agentpool labels
	n1.ObjectMeta.Labels["agentpool"] = ""
	n2.ObjectMeta.Labels["agentpool"] = ""
	checkNodesSimilar(t, n1, n2, comparator, false)
	// Only one non empty
	n1.ObjectMeta.Labels["agentpool"] = ""
	n2.ObjectMeta.Labels["agentpool"] = "foo"
	checkNodesSimilar(t, n1, n2, comparator, false)
	// Only one present
	delete(n1.ObjectMeta.Labels, "agentpool")
	n2.ObjectMeta.Labels["agentpool"] = "foo"
	checkNodesSimilar(t, n1, n2, comparator, false)
	// Different vales
	n1.ObjectMeta.Labels["agentpool"] = "foo1"
	n2.ObjectMeta.Labels["agentpool"] = "foo2"
	checkNodesSimilar(t, n1, n2, comparator, false)
	// Same values
	n1.ObjectMeta.Labels["agentpool"] = "foo"
	n2.ObjectMeta.Labels["agentpool"] = "foo"
	checkNodesSimilar(t, n1, n2, comparator, true)
	// Same labels except for agentpool
	delete(n1.ObjectMeta.Labels, "character")
	n1.ObjectMeta.Labels["agentpool"] = "foo"
	n2.ObjectMeta.Labels["agentpool"] = "bar"
	checkNodesSimilar(t, n1, n2, comparator, true)
	// Custom label
	n1.ObjectMeta.Labels["example.com/ready"] = "true"
	n2.ObjectMeta.Labels["example.com/ready"] = "false"
	checkNodesSimilar(t, n1, n2, comparator, true)
}

func TestFindSimilarNodeGroupsAzureBasic(t *testing.T) {
	context := &context.AutoscalingContext{}
	ni1, ni2, ni3 := buildBasicNodeGroups(context)
	processor := &BalancingNodeGroupSetProcessor{Comparator: CreateAzureNodeInfoComparator([]string{})}
	basicSimilarNodeGroupsTest(t, context, processor, ni1, ni2, ni3)
}

func TestFindSimilarNodeGroupsAzureByLabel(t *testing.T) {
	processor := &BalancingNodeGroupSetProcessor{Comparator: CreateAzureNodeInfoComparator([]string{})}
	context := &context.AutoscalingContext{}

	n1 := BuildTestNode("n1", 1000, 1000)
	n2 := BuildTestNode("n2", 2000, 2000)

	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 1, 10, 1)
	provider.AddNodeGroup("ng2", 1, 10, 1)
	provider.AddNode("ng1", n1)
	provider.AddNode("ng2", n2)

	ni1 := schedulerframework.NewNodeInfo()
	ni1.SetNode(n1)
	ni2 := schedulerframework.NewNodeInfo()
	ni2.SetNode(n2)

	nodeInfosForGroups := map[string]*schedulerframework.NodeInfo{
		"ng1": ni1, "ng2": ni2,
	}

	ng1, _ := provider.NodeGroupForNode(n1)
	ng2, _ := provider.NodeGroupForNode(n2)
	context.CloudProvider = provider

	// Groups with different cpu and mem are not similar.
	similar, err := processor.FindSimilarNodeGroups(context, ng1, nodeInfosForGroups)
	assert.NoError(t, err)
	assert.Equal(t, similar, []cloudprovider.NodeGroup{})

	// Unless we give them nodepool label.
	n1.ObjectMeta.Labels["agentpool"] = "foobar"
	n2.ObjectMeta.Labels["agentpool"] = "foobar"
	similar, err = processor.FindSimilarNodeGroups(context, ng1, nodeInfosForGroups)
	assert.NoError(t, err)
	assert.Equal(t, similar, []cloudprovider.NodeGroup{ng2})

	// Groups with the same cpu and mem are similar if they belong to different pools.
	n3 := BuildTestNode("n1", 1000, 1000)
	provider.AddNodeGroup("ng3", 1, 10, 1)
	provider.AddNode("ng3", n3)
	ni3 := schedulerframework.NewNodeInfo()
	ni3.SetNode(n3)
	nodeInfosForGroups["ng3"] = ni3
	ng3, _ := provider.NodeGroupForNode(n3)

	n1.ObjectMeta.Labels["agentpool"] = "foobar1"
	n2.ObjectMeta.Labels["agentpool"] = "foobar2"
	n3.ObjectMeta.Labels["agentpool"] = "foobar3"

	similar, err = processor.FindSimilarNodeGroups(context, ng1, nodeInfosForGroups)
	assert.NoError(t, err)
	assert.Equal(t, similar, []cloudprovider.NodeGroup{ng3})
}
