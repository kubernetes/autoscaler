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

package linode

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/linode/linodego"
)

func TestCloudProvider_newLinodeCloudProvider(t *testing.T) {
	// test ok on correctly creating a linode provider
	rl := &cloudprovider.ResourceLimiter{}
	cfg := strings.NewReader(`
[global]
linode-token=123123123
lke-cluster-id=456456
`)
	_, err := newLinodeCloudProvider(cfg, rl)
	assert.NoError(t, err)

	// test error on creating a linode provider when config is bad
	cfg = strings.NewReader(`
[globalxxx]
linode-token=123123123
lke-cluster-id=456456
`)
	_, err = newLinodeCloudProvider(cfg, rl)
	assert.Error(t, err)
}

func TestCloudProvider_NodeGroups(t *testing.T) {
	// test ok on getting the correct nodes when calling NodeGroups()
	cfg := strings.NewReader(`
[global]
linode-token=123123123
lke-cluster-id=456456
`)
	m, _ := newManager(cfg)
	m.nodeGroups = map[string]*NodeGroup{
		"g6-standard-1": {id: "g6-standard-1"},
		"g6-standard-2": {id: "g6-standard-2"},
	}
	lcp := &linodeCloudProvider{manager: m}
	ng := lcp.NodeGroups()
	assert.Equal(t, 2, len(ng))
	assert.Contains(t, ng, m.nodeGroups["g6-standard-1"])
	assert.Contains(t, ng, m.nodeGroups["g6-standard-2"])
}

func TestCloudProvider_NodeGroupForNode(t *testing.T) {
	cfg := strings.NewReader(`
[global]
linode-token=123123123
lke-cluster-id=456456
`)
	m, _ := newManager(cfg)
	ng1 := &NodeGroup{
		id: "g6-standard-1",
		lkePools: map[int]*linodego.LKEClusterPool{
			1: {ID: 1,
				Count: 2,
				Type:  "g6-standard-1",
				Linodes: []linodego.LKEClusterPoolLinode{
					{InstanceID: 111},
					{InstanceID: 222},
				},
			},
			2: {ID: 2,
				Count: 2,
				Type:  "g6-standard-1",
				Linodes: []linodego.LKEClusterPoolLinode{
					{InstanceID: 333},
				},
			},
			3: {ID: 3,
				Count: 2,
				Type:  "g6-standard-1",
				Linodes: []linodego.LKEClusterPoolLinode{
					{InstanceID: 444},
					{InstanceID: 555},
				},
			},
		},
	}
	ng2 := &NodeGroup{
		id: "g6-standard-2",
		lkePools: map[int]*linodego.LKEClusterPool{
			4: {ID: 4,
				Count: 2,
				Type:  "g6-standard-2",
				Linodes: []linodego.LKEClusterPoolLinode{
					{InstanceID: 666},
					{InstanceID: 777},
				},
			},
		},
	}
	m.nodeGroups = map[string]*NodeGroup{
		"g6-standard-1": ng1,
		"g6-standard-2": ng2,
	}
	lcp := &linodeCloudProvider{manager: m}

	// test ok on getting the right node group for an apiv1.Node
	node := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "linode://555",
		},
	}
	ng, err := lcp.NodeGroupForNode(node)
	assert.NoError(t, err)
	assert.Equal(t, ng1, ng)

	// test ok on getting the right node group for an apiv1.Node
	node = &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "linode://666",
		},
	}
	ng, err = lcp.NodeGroupForNode(node)
	assert.NoError(t, err)
	assert.Equal(t, ng2, ng)

	// test ok on getting nil when looking for a apiv1.Node we do not manage
	node = &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "linode://999",
		},
	}
	ng, err = lcp.NodeGroupForNode(node)
	assert.NoError(t, err)
	assert.Nil(t, ng)

	// test error on looking for a apiv1.Node with a bad providerID
	node = &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "linode://aaa",
		},
	}
	ng, err = lcp.NodeGroupForNode(node)
	assert.Error(t, err)

}

func TestCloudProvider_others(t *testing.T) {
	cfg := strings.NewReader(`
[global]
linode-token=123123123
lke-cluster-id=456456
`)
	m, _ := newManager(cfg)
	resourceLimiter := &cloudprovider.ResourceLimiter{}
	lcp := &linodeCloudProvider{
		manager:         m,
		resourceLimiter: resourceLimiter,
	}
	assert.Equal(t, cloudprovider.LinodeProviderName, lcp.Name())
	_, err := lcp.GetAvailableMachineTypes()
	assert.Error(t, err)
	_, err = lcp.NewNodeGroup("", nil, nil, nil, nil)
	assert.Error(t, err)
	rl, err := lcp.GetResourceLimiter()
	assert.Equal(t, resourceLimiter, rl)
	assert.Equal(t, "", lcp.GPULabel())
	assert.Nil(t, lcp.GetAvailableGPUTypes())
	assert.Nil(t, lcp.Cleanup())
	_, err2 := lcp.Pricing()
	assert.Error(t, err2)
}
