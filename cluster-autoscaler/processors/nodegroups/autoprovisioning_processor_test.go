/*
Copyright 2018 The Kubernetes Authors.

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

package nodegroups

import (
	"testing"

	apiv1 "k8s.io/api/core/v1"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/kubernetes/pkg/scheduler/schedulercache"

	"github.com/stretchr/testify/assert"
)

func TestAutoprovisioningNGLProcessor(t *testing.T) {
	processor := NewAutoprovisioningNodeGroupListProcessor()

	t1 := BuildTestNode("t1", 4000, 1000000)
	ti1 := schedulercache.NewNodeInfo()
	ti1.SetNode(t1)
	p1 := BuildTestPod("p1", 100, 100)

	n1 := BuildTestNode("ng1-xxx", 4000, 1000000)
	ni1 := schedulercache.NewNodeInfo()
	ni1.SetNode(n1)

	provider := testprovider.NewTestAutoprovisioningCloudProvider(nil, nil,
		nil, nil,
		[]string{"T1"}, map[string]*schedulercache.NodeInfo{"T1": ti1})
	provider.AddNodeGroup("ng1", 1, 5, 3)

	context := &context.AutoscalingContext{
		AutoscalingOptions: context.AutoscalingOptions{
			MaxAutoprovisionedNodeGroupCount: 1,
			NodeAutoprovisioningEnabled:      true,
		},
		CloudProvider: provider,
	}
	nodeGroups := provider.NodeGroups()
	nodeInfos := map[string]*schedulercache.NodeInfo{
		"ng1": ni1,
	}
	var err error
	nodeGroups, nodeInfos, err = processor.Process(context, nodeGroups, nodeInfos, []*apiv1.Pod{p1})

	assert.NoError(t, err)
	assert.Equal(t, 2, len(nodeGroups))
	assert.Equal(t, 2, len(nodeInfos))
}

func TestAutoprovisioningNGLProcessorTooMany(t *testing.T) {
	processor := NewAutoprovisioningNodeGroupListProcessor()

	t1 := BuildTestNode("T1-abc", 4000, 1000000)
	ti1 := schedulercache.NewNodeInfo()
	ti1.SetNode(t1)

	x1 := BuildTestNode("X1-cde", 4000, 1000000)
	xi1 := schedulercache.NewNodeInfo()
	xi1.SetNode(x1)

	p1 := BuildTestPod("p1", 100, 100)

	provider := testprovider.NewTestAutoprovisioningCloudProvider(nil, nil,
		nil, nil,
		[]string{"T1", "X1"},
		map[string]*schedulercache.NodeInfo{"T1": ti1, "X1": xi1})
	provider.AddAutoprovisionedNodeGroup("autoprovisioned-X1", 0, 1000, 0, "X1")

	context := &context.AutoscalingContext{
		AutoscalingOptions: context.AutoscalingOptions{
			MaxAutoprovisionedNodeGroupCount: 1,
			NodeAutoprovisioningEnabled:      true,
		},
		CloudProvider: provider,
	}
	nodeGroups := provider.NodeGroups()
	nodeInfos := map[string]*schedulercache.NodeInfo{"X1": xi1}
	var err error
	nodeGroups, nodeInfos, err = processor.Process(context, nodeGroups, nodeInfos, []*apiv1.Pod{p1})

	assert.NoError(t, err)
	assert.Equal(t, 1, len(nodeGroups))
	assert.Equal(t, 1, len(nodeInfos))
}
