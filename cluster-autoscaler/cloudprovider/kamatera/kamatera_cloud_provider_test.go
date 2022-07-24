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

package kamatera

import (
	"fmt"
	apiv1 "k8s.io/api/core/v1"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

func TestCloudProvider_newKamateraCloudProvider(t *testing.T) {
	// test ok on correctly creating a Kamatera provider
	server := NewHttpServerMock(MockFieldContentType, MockFieldResponse)
	defer server.Close()
	rl := &cloudprovider.ResourceLimiter{}
	cfg := strings.NewReader(fmt.Sprintf(`
[global]
kamatera-api-client-id=1a222bbb3ccc44d5555e6ff77g88hh9i
kamatera-api-secret=9ii88h7g6f55555ee4444444dd33eee2
kamatera-api-url=%s
cluster-name=aaabbb
`, server.URL))
	server.On("handle", "/service/servers").Return(
		"application/json", `[]`).Once()
	_, err := newKamateraCloudProvider(cfg, rl, nil)
	assert.NoError(t, err)

	// test error on creating a Kamatera provider when config is bad
	cfg = strings.NewReader(`
[global]
kamatera-api-client-id=
kamatera-api-secret=
cluster-name=
`)
	_, err = newKamateraCloudProvider(cfg, rl, nil)
	assert.Error(t, err)
	assert.Equal(t, "could not create kamatera manager: failed to parse config: cluster name is not set", err.Error())
}

func TestCloudProvider_NodeGroups(t *testing.T) {
	// test ok on getting the correct nodes when calling NodeGroups()
	cfg := strings.NewReader(`
[global]
kamatera-api-client-id=1a222bbb3ccc44d5555e6ff77g88hh9i
kamatera-api-secret=9ii88h7g6f55555ee4444444dd33eee2
cluster-name=aaabbb
`)
	m, _ := newManager(cfg, nil)
	m.nodeGroups = map[string]*NodeGroup{
		"ng1": {id: "ng1"},
		"ng2": {id: "ng2"},
	}
	kcp := &kamateraCloudProvider{manager: m}
	ng := kcp.NodeGroups()
	assert.Equal(t, 2, len(ng))
	assert.Contains(t, ng, m.nodeGroups["ng1"])
	assert.Contains(t, ng, m.nodeGroups["ng2"])
}

func TestCloudProvider_NodeGroupForNode(t *testing.T) {
	cfg := strings.NewReader(`
[global]
kamatera-api-client-id=1a222bbb3ccc44d5555e6ff77g88hh9i
kamatera-api-secret=9ii88h7g6f55555ee4444444dd33eee2
cluster-name=aaabbb
`)
	m, _ := newManager(cfg, nil)
	kamateraServerName1 := mockKamateraServerName()
	kamateraServerName2 := mockKamateraServerName()
	kamateraServerName3 := mockKamateraServerName()
	kamateraServerName4 := mockKamateraServerName()
	ng1 := &NodeGroup{
		id: "ng1",
		instances: map[string]*Instance{
			kamateraServerName1: {Id: kamateraServerName1},
			kamateraServerName2: {Id: kamateraServerName2},
		},
	}
	ng2 := &NodeGroup{
		id: "ng2",
		instances: map[string]*Instance{
			kamateraServerName3: {Id: kamateraServerName3},
			kamateraServerName4: {Id: kamateraServerName4},
		},
	}
	m.nodeGroups = map[string]*NodeGroup{
		"ng1": ng1,
		"ng2": ng2,
	}
	kcp := &kamateraCloudProvider{manager: m}

	// test ok on getting the right node group for an apiv1.Node
	node := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: kamateraServerName1,
		},
	}
	ng, err := kcp.NodeGroupForNode(node)
	assert.NoError(t, err)
	assert.Equal(t, ng1, ng)

	// test ok on getting the right node group for an apiv1.Node
	node = &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: kamateraServerName4,
		},
	}
	ng, err = kcp.NodeGroupForNode(node)
	assert.NoError(t, err)
	assert.Equal(t, ng2, ng)

	// test ok on getting nil when looking for a apiv1.Node we do not manage
	node = &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: mockKamateraServerName(),
		},
	}
	ng, err = kcp.NodeGroupForNode(node)
	assert.NoError(t, err)
	assert.Nil(t, ng)
}

func TestCloudProvider_others(t *testing.T) {
	cfg := strings.NewReader(`
[global]
kamatera-api-client-id=1a222bbb3ccc44d5555e6ff77g88hh9i
kamatera-api-secret=9ii88h7g6f55555ee4444444dd33eee2
cluster-name=aaabbb
`)
	m, _ := newManager(cfg, nil)
	resourceLimiter := &cloudprovider.ResourceLimiter{}
	kcp := &kamateraCloudProvider{
		manager:         m,
		resourceLimiter: resourceLimiter,
	}
	assert.Equal(t, cloudprovider.KamateraProviderName, kcp.Name())
	_, err := kcp.GetAvailableMachineTypes()
	assert.Error(t, err)
	_, err = kcp.NewNodeGroup("", nil, nil, nil, nil)
	assert.Error(t, err)
	rl, err := kcp.GetResourceLimiter()
	assert.Equal(t, resourceLimiter, rl)
	assert.Equal(t, "", kcp.GPULabel())
	assert.Nil(t, kcp.GetAvailableGPUTypes())
	assert.Nil(t, kcp.Cleanup())
	_, err2 := kcp.Pricing()
	assert.Error(t, err2)
}
