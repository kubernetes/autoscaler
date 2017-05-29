/*
Copyright 2016 The Kubernetes Authors.

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
	"fmt"
	"testing"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/expander"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	apiv1 "k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/plugin/pkg/scheduler/schedulercache"

	"github.com/stretchr/testify/assert"
)

type testPricingModel struct {
	nodePrice map[string]float64
	podPrice  map[string]float64
}

func (tpm *testPricingModel) NodePrice(node *apiv1.Node, startTime time.Time, endTime time.Time) (float64, error) {
	if price, found := tpm.nodePrice[node.Name]; found {
		return price, nil
	}
	return 0.0, fmt.Errorf("price for node %v not found", node.Name)
}

func (tpm *testPricingModel) PodPrice(node *apiv1.Pod, startTime time.Time, endTime time.Time) (float64, error) {
	if price, found := tpm.podPrice[node.Name]; found {
		return price, nil
	}
	return 0.0, fmt.Errorf("price for pod %v not found", node.Name)
}

func TestPriceExpander(t *testing.T) {
	n1 := BuildTestNode("n1", 1000, 1000)
	n2 := BuildTestNode("n2", 10000, 1000)

	p1 := BuildTestPod("p1", 1000, 0)
	p2 := BuildTestPod("p2", 500, 0)

	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 1, 10, 1)
	provider.AddNodeGroup("ng2", 1, 10, 1)
	provider.AddNode("ng1", n1)
	provider.AddNode("ng2", n2)
	ng1, _ := provider.NodeGroupForNode(n1)
	ng2, _ := provider.NodeGroupForNode(n2)

	ni1 := schedulercache.NewNodeInfo()
	ni1.SetNode(n1)
	ni2 := schedulercache.NewNodeInfo()
	ni2.SetNode(n2)
	nodeInfosForGroups := map[string]*schedulercache.NodeInfo{
		"ng1": ni1, "ng2": ni2,
	}

	// All node groups accept the same set of pods
	options := []expander.Option{
		{
			NodeGroup: ng1,
			NodeCount: 2,
			Pods:      []*apiv1.Pod{p1, p2},
			Debug:     "1",
		},
		{
			NodeGroup: ng2,
			NodeCount: 1,
			Pods:      []*apiv1.Pod{p1, p2},
			Debug:     "2",
		},
	}

	// First node group is cheapter
	assert.Equal(t, "1", NewStrategy(&testPricingModel{
		podPrice: map[string]float64{
			"p1": 20.0,
			"p2": 10.0,
		},
		nodePrice: map[string]float64{
			"n1": 20.0,
			"n2": 200.0,
		},
	}).BestOption(options, nodeInfosForGroups).Debug)

	// Second node group is cheapter
	assert.Equal(t, "2", NewStrategy(&testPricingModel{
		podPrice: map[string]float64{
			"p1": 20.0,
			"p2": 10.0,
		},
		nodePrice: map[string]float64{
			"n1": 200.0,
			"n2": 100.0,
		},
	}).BestOption(options, nodeInfosForGroups).Debug)

	// First group accept 1 pod and second accepts 2.
	options2 := []expander.Option{
		{
			NodeGroup: ng1,
			NodeCount: 2,
			Pods:      []*apiv1.Pod{p1},
			Debug:     "1",
		},
		{
			NodeGroup: ng2,
			NodeCount: 1,
			Pods:      []*apiv1.Pod{p1, p2},
			Debug:     "2",
		},
	}

	// Both node groups are equally expensive. However 2
	// accept two pods.
	assert.Equal(t, "2", NewStrategy(&testPricingModel{
		podPrice: map[string]float64{
			"p1": 20.0,
			"p2": 10.0,
		},
		nodePrice: map[string]float64{
			"n1": 200.0,
			"n2": 200.0,
		},
	}).BestOption(options2, nodeInfosForGroups).Debug)

	// Errors are expected
	assert.Nil(t, NewStrategy(&testPricingModel{
		podPrice:  map[string]float64{},
		nodePrice: map[string]float64{},
	}).BestOption(options2, nodeInfosForGroups))
}
