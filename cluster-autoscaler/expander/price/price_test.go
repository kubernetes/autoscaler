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
	"strings"
	"testing"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/expander"
	"k8s.io/autoscaler/cluster-autoscaler/utils/units"

	apiv1 "k8s.io/api/core/v1"
	cloudprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"

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

type testPreferredNodeProvider struct {
	preferred *apiv1.Node
}

func (tpnp *testPreferredNodeProvider) Node() (*apiv1.Node, error) {
	return tpnp.preferred, nil
}

func optionsToDebug(options []expander.Option) []string {
	var ret []string
	for _, option := range options {
		s := strings.Split(option.Debug, " ")
		if len(s) == 0 {
			s = append(s, "")
		}
		ret = append(ret, s[0])
	}
	return ret
}

func TestPriceExpander(t *testing.T) {
	n1 := BuildTestNode("n1", 1000, 1000)
	n2 := BuildTestNode("n2", 4000, 1000)
	n3 := BuildTestNode("n3", 4000, 1000)

	p1 := BuildTestPod("p1", 1000, 0)
	p2 := BuildTestPod("p2", 500, 0)

	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 1, 10, 1)
	provider.AddNodeGroup("ng2", 1, 10, 1)
	provider.AddNode("ng1", n1)
	provider.AddNode("ng2", n2)
	ng1, _ := provider.NodeGroupForNode(n1)
	ng2, _ := provider.NodeGroupForNode(n2)
	ng3, _ := provider.NewNodeGroup("MT1", nil, nil, nil, nil)

	ni1 := framework.NewNodeInfo(n1, nil)
	ni2 := framework.NewNodeInfo(n2, nil)
	ni3 := framework.NewNodeInfo(n3, nil)
	nodeInfosForGroups := map[string]*framework.NodeInfo{
		"ng1": ni1, "ng2": ni2,
	}
	var pricingModel cloudprovider.PricingModel

	// All node groups accept the same set of pods.
	options := []expander.Option{
		{
			NodeGroup: ng1,
			NodeCount: 2,
			Pods:      []*apiv1.Pod{p1, p2},
			Debug:     "ng1",
		},
		{
			NodeGroup: ng2,
			NodeCount: 1,
			Pods:      []*apiv1.Pod{p1, p2},
			Debug:     "ng2",
		},
	}

	// First node group is cheaper.
	pricingModel = &testPricingModel{
		podPrice: map[string]float64{
			"p1":        20.0,
			"p2":        10.0,
			"stabilize": 10,
		},
		nodePrice: map[string]float64{
			"n1": 20.0,
			"n2": 200.0,
		},
	}
	provider.SetPricingModel(pricingModel)
	assert.Equal(t, optionsToDebug(NewFilter(
		provider,
		&testPreferredNodeProvider{
			preferred: buildNode(2000, units.GiB),
		},
		SimpleNodeUnfitness,
	).BestOptions(options, nodeInfosForGroups)), []string{"ng1"})

	// First node group is cheaper, however, the second one is preferred.
	pricingModel = &testPricingModel{
		podPrice: map[string]float64{
			"p1":        20.0,
			"p2":        10.0,
			"stabilize": 10,
		},
		nodePrice: map[string]float64{
			"n1": 50.0,
			"n2": 200.0,
		},
	}
	provider.SetPricingModel(pricingModel)
	assert.Equal(t, optionsToDebug(NewFilter(
		provider,
		&testPreferredNodeProvider{
			preferred: buildNode(4000, units.GiB),
		},
		SimpleNodeUnfitness,
	).BestOptions(options, nodeInfosForGroups)), []string{"ng2"})

	// All node groups accept the same set of pods. Lots of nodes.
	options1b := []expander.Option{
		{
			NodeGroup: ng1,
			NodeCount: 80,
			Pods:      []*apiv1.Pod{p1, p2},
			Debug:     "ng1",
		},
		{
			NodeGroup: ng2,
			NodeCount: 40,
			Pods:      []*apiv1.Pod{p1, p2},
			Debug:     "ng2",
		},
	}
	// First node group is cheaper, the second is preferred
	// but there is lots of nodes to be created.
	pricingModel = &testPricingModel{
		podPrice: map[string]float64{
			"p1":        20.0,
			"p2":        10.0,
			"stabilize": 10,
		},
		nodePrice: map[string]float64{
			"n1": 20.0,
			"n2": 200.0,
		},
	}
	provider.SetPricingModel(pricingModel)
	assert.Equal(t, optionsToDebug(NewFilter(
		provider,

		&testPreferredNodeProvider{
			preferred: buildNode(4000, units.GiB),
		},
		SimpleNodeUnfitness,
	).BestOptions(options1b, nodeInfosForGroups)), []string{"ng1"})

	// Second node group is cheaper
	pricingModel = &testPricingModel{
		podPrice: map[string]float64{
			"p1":        20.0,
			"p2":        10.0,
			"stabilize": 10,
		},
		nodePrice: map[string]float64{
			"n1": 200.0,
			"n2": 100.0,
		},
	}
	provider.SetPricingModel(pricingModel)
	assert.Equal(t, optionsToDebug(NewFilter(
		provider,
		&testPreferredNodeProvider{
			preferred: buildNode(2000, units.GiB),
		},
		SimpleNodeUnfitness,
	).BestOptions(options, nodeInfosForGroups)), []string{"ng2"})

	// First group accept 1 pod and second accepts 2.
	options2 := []expander.Option{
		{
			NodeGroup: ng1,
			NodeCount: 2,
			Pods:      []*apiv1.Pod{p1},
			Debug:     "ng1",
		},
		{
			NodeGroup: ng2,
			NodeCount: 1,
			Pods:      []*apiv1.Pod{p1, p2},
			Debug:     "ng2",
		},
	}
	pricingModel = &testPricingModel{
		podPrice: map[string]float64{
			"p1":        20.0,
			"p2":        10.0,
			"stabilize": 10,
		},
		nodePrice: map[string]float64{
			"n1": 200.0,
			"n2": 200.0,
		},
	}
	provider.SetPricingModel(pricingModel)
	// Both node groups are equally expensive. However 2
	// accept two pods.
	assert.Equal(t, optionsToDebug(NewFilter(
		provider,
		&testPreferredNodeProvider{
			preferred: buildNode(2000, units.GiB),
		},
		SimpleNodeUnfitness,
	).BestOptions(options2, nodeInfosForGroups)), []string{"ng2"})

	// Errors are expected
	pricingModel = &testPricingModel{
		podPrice:  map[string]float64{},
		nodePrice: map[string]float64{},
	}
	provider.SetPricingModel(pricingModel)
	assert.Empty(t, NewFilter(
		provider,
		&testPreferredNodeProvider{
			preferred: buildNode(2000, units.GiB),
		},
		SimpleNodeUnfitness,
	).BestOptions(options2, nodeInfosForGroups))

	// Add node info for autoprovisioned group.
	nodeInfosForGroups["autoprovisioned-MT1"] = ni3
	// First group accept 1 pod, second accepts 2 and third accepts 2 (non-existent autoprovisioned)
	options3 := []expander.Option{
		{
			NodeGroup: ng1,
			NodeCount: 2,
			Pods:      []*apiv1.Pod{p1},
			Debug:     "ng1",
		},
		{
			NodeGroup: ng2,
			NodeCount: 1,
			Pods:      []*apiv1.Pod{p1, p2},
			Debug:     "ng2",
		},
		{
			NodeGroup: ng3,
			NodeCount: 1,
			Pods:      []*apiv1.Pod{p1, p2},
			Debug:     "ng3",
		},
	}
	// Choose existing group when non-existing has the same price.
	pricingModel = &testPricingModel{
		podPrice: map[string]float64{
			"p1":        20.0,
			"p2":        10.0,
			"stabilize": 10,
		},
		nodePrice: map[string]float64{
			"n1": 200.0,
			"n2": 200.0,
			"n3": 200.0,
		},
	}
	provider.SetPricingModel(pricingModel)
	assert.Equal(t, optionsToDebug(NewFilter(
		provider,
		&testPreferredNodeProvider{
			preferred: buildNode(2000, units.GiB),
		},
		SimpleNodeUnfitness,
	).BestOptions(options3, nodeInfosForGroups)), []string{"ng2"})

	// Choose non-existing group when non-existing is cheaper.
	pricingModel = &testPricingModel{
		podPrice: map[string]float64{
			"p1":        20.0,
			"p2":        10.0,
			"stabilize": 10,
		},
		nodePrice: map[string]float64{
			"n1": 200.0,
			"n2": 200.0,
			"n3": 90.0,
		},
	}
	provider.SetPricingModel(pricingModel)
	assert.Equal(t, optionsToDebug(NewFilter(
		provider,
		&testPreferredNodeProvider{
			preferred: buildNode(2000, units.GiB),
		},
		SimpleNodeUnfitness,
	).BestOptions(options3, nodeInfosForGroups)), []string{"ng3"})
}
