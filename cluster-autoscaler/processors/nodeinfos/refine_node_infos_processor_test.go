/*
Copyright 2020 The Kubernetes Authors.

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

package nodeinfos

import (
	"testing"

	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/utils"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroupset"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"

	"github.com/stretchr/testify/assert"
)

type testNodeGroup struct {
	name         string
	cpu          int64
	mem          int64
	fromRealNode bool
}

type refinetest struct {
	name       string
	nodegroups []testNodeGroup
	templates  []testNodeGroup
	expected   []testNodeGroup
}

var refineCases = []refinetest{
	{
		"Generated templates should leverage NodeInfo from similar nodegroup when available",
		[]testNodeGroup{{"n1", 1000, 1000, true}, {"n2", 3000, 3000, false}},
		[]testNodeGroup{{"n1", 1000, 1000, false}, {"n2", 1000, 1000, false}},
		[]testNodeGroup{{"n1", 1000, 1000, true}, {"n2", 1000, 1000, true}},
	},
	{
		"NodeInfos obtained from real-world nodes should remain untouched",
		[]testNodeGroup{{"n1", 1000, 1000, true}, {"n2", 1000, 1000, true}},
		[]testNodeGroup{{"n1", 3000, 3000, false}, {"n2", 3000, 3000, false}},
		[]testNodeGroup{{"n1", 1000, 1000, true}, {"n2", 1000, 1000, true}},
	},
	{
		"Should be no-op on cloudproviders not implementing TemplateInfos",
		[]testNodeGroup{{"n1", 1000, 1000, true}, {"n2", 2000, 2000, false}},
		nil, /* TemplateNodeInfo() returns cloudprovider.ErrNotImplemented */
		[]testNodeGroup{{"n1", 1000, 1000, true}, {"n2", 2000, 2000, false}},
	},
	{
		"Should not use NodeInfo templates that aren't built from real nodes",
		[]testNodeGroup{{"n1", 1000, 1000, false}, {"n2", 7000, 7000, false}},
		[]testNodeGroup{{"n1", 1000, 1000, false}, {"n2", 1000, 1000, false}},
		[]testNodeGroup{{"n1", 1000, 1000, false}, {"n2", 7000, 7000, false}},
	},
	{
		"Should not use NodeInfo templates from dissimilar nodegroups",
		[]testNodeGroup{{"n1", 1000, 1000, true}, {"n2", 7000, 7000, false}},
		[]testNodeGroup{{"n1", 1000, 1000, false}, {"n2", 7000, 7000, false}},
		[]testNodeGroup{{"n1", 1000, 1000, true}, {"n2", 7000, 7000, false}},
	},
}

func buildRefineNodeInfosProcessorTest(r refinetest) (*context.AutoscalingContext, map[string]*schedulerframework.NodeInfo) {
	templates := toNodeInfoMap(r.templates)
	nodeinfos := toNodeInfoMap(r.nodegroups)
	provider := testprovider.NewTestAutoprovisioningCloudProvider(nil, nil, nil, nil, []string{}, templates)
	if r.templates == nil {
		provider = testprovider.NewTestCloudProvider(nil, nil)
	}
	for _, ng := range r.nodegroups {
		provider.AddNodeGroup(ng.name, 1, 10, 1)
	}

	ctx := &context.AutoscalingContext{
		CloudProvider: provider,
	}

	return ctx, nodeinfos
}

func toNodeInfoMap(nodeGroups []testNodeGroup) map[string]*schedulerframework.NodeInfo {
	nodeinfos := make(map[string]*schedulerframework.NodeInfo)
	for _, ng := range nodeGroups {
		node := BuildTestNode(ng.name, ng.cpu, ng.mem)
		ni := schedulerframework.NewNodeInfo()
		ni.SetNode(node)
		if !ng.fromRealNode {
			utils.SetNodeInfoBuiltFromTemplate(ni)
		}
		nodeinfos[ng.name] = ni
	}
	return nodeinfos
}

func TestRefineNodeInfosProcessor(t *testing.T) {
	for _, tt := range refineCases {
		proc := RefineNodeInfosProcessor{
			NodeGroupSetProcessor:        nodegroupset.NewDefaultNodeGroupSetProcessor([]string{}),
			RefineUsingSimilarNodeGroups: true,
		}
		ctx, nodeInfosForNodeGroups := buildRefineNodeInfosProcessorTest(tt)
		out, err := proc.Process(ctx, nodeInfosForNodeGroups)

		assert.NoError(t, err)
		expected := toNodeInfoMap(tt.expected)
		comparator := nodegroupset.CreateGenericNodeInfoComparator([]string{})

		for nodeGroupName, nodeGroup := range expected {
			assert.Contains(t, out, nodeGroupName, tt.name)
			assert.True(t, comparator(out[nodeGroupName], nodeGroup), tt.name)
		}
	}
}
