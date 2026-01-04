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
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	apiv1 "k8s.io/api/core/v1"
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
			ProviderID: formatKamateraProviderID(kamateraServerName1),
		},
	}
	ng, err := kcp.NodeGroupForNode(node)
	assert.NoError(t, err)
	assert.Equal(t, ng1, ng)

	// test ok on getting the right node group for an apiv1.Node
	node = &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: formatKamateraProviderID(kamateraServerName4),
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

func TestCloudProvider_HasInstance(t *testing.T) {
	cfg := strings.NewReader(`
[global]
kamatera-api-client-id=1a222bbb3ccc44d5555e6ff77g88hh9i
kamatera-api-secret=9ii88h7g6f55555ee4444444dd33eee2
cluster-name=aaabbb
`)
	m, _ := newManager(cfg, nil)
	kcp := &kamateraCloudProvider{manager: m}

	node := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: mockKamateraServerName(),
		},
	}
	hasInstance, err := kcp.HasInstance(node)
	assert.True(t, hasInstance)
	assert.Error(t, err)
	assert.Equal(t, cloudprovider.ErrNotImplemented, err)
}

func TestCloudProvider_GetNodeGpuConfig(t *testing.T) {
	cfg := strings.NewReader(`
[global]
kamatera-api-client-id=1a222bbb3ccc44d5555e6ff77g88hh9i
kamatera-api-secret=9ii88h7g6f55555ee4444444dd33eee2
cluster-name=aaabbb
`)
	m, _ := newManager(cfg, nil)
	kcp := &kamateraCloudProvider{manager: m}

	node := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: mockKamateraServerName(),
		},
	}
	gpuConfig := kcp.GetNodeGpuConfig(node)
	assert.Nil(t, gpuConfig)
}

func TestCloudProvider_Refresh(t *testing.T) {
	cfg := strings.NewReader(fmt.Sprintf(`
[global]
kamatera-api-client-id=1a222bbb3ccc44d5555e6ff77g88hh9i
kamatera-api-secret=9ii88h7g6f55555ee4444444dd33eee2
cluster-name=aaabbb
filter-name-prefix=myprefix
default-datacenter=IL
default-image=ubuntu
default-cpu=1a
default-ram=1024
default-disk=size=10
default-network=name=wan,ip=auto
default-script-base64=ZGVmYXVsdAo=

[nodegroup "ng1"]
min-size=1
max-size=2
`))
	m, err := newManager(cfg, nil)
	assert.NoError(t, err)

	client := kamateraClientMock{}
	m.client = &client

	kcp := &kamateraCloudProvider{manager: m}

	serverName1 := "myprefix" + mockKamateraServerName()
	client.On(
		"ListServers", mock.Anything, m.instances, "myprefix",
	).Return(
		[]Server{
			{Name: serverName1, Tags: []string{fmt.Sprintf("%s%s", clusterServerTagPrefix, "aaabbb"), fmt.Sprintf("%s%s", nodeGroupTagPrefix, "ng1")}},
		},
		nil,
	).Once()

	err = kcp.Refresh()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(m.nodeGroups))
	assert.Equal(t, 1, len(m.nodeGroups["ng1"].instances))
}

func TestCloudProvider_NodeGroupsConcurrent(t *testing.T) {
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

	const numGoroutines = 50
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Concurrently call NodeGroups
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			ng := kcp.NodeGroups()
			assert.Equal(t, 2, len(ng))
		}()
	}

	wg.Wait()
}

func TestCloudProvider_NodeGroupForNodeConcurrent(t *testing.T) {
	cfg := strings.NewReader(`
[global]
kamatera-api-client-id=1a222bbb3ccc44d5555e6ff77g88hh9i
kamatera-api-secret=9ii88h7g6f55555ee4444444dd33eee2
cluster-name=aaabbb
`)
	m, _ := newManager(cfg, nil)
	kamateraServerName1 := mockKamateraServerName()
	kamateraServerName2 := mockKamateraServerName()
	ng1 := &NodeGroup{
		id: "ng1",
		instances: map[string]*Instance{
			kamateraServerName1: {Id: kamateraServerName1},
		},
	}
	ng2 := &NodeGroup{
		id: "ng2",
		instances: map[string]*Instance{
			kamateraServerName2: {Id: kamateraServerName2},
		},
	}
	m.nodeGroups = map[string]*NodeGroup{
		"ng1": ng1,
		"ng2": ng2,
	}
	kcp := &kamateraCloudProvider{manager: m}

	const numGoroutines = 50
	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2)

	// Concurrently call NodeGroupForNode for different nodes
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			node := &apiv1.Node{
				Spec: apiv1.NodeSpec{
					ProviderID: kamateraServerName1,
				},
			}
			ng, err := kcp.NodeGroupForNode(node)
			assert.NoError(t, err)
			assert.Equal(t, ng1, ng)
		}()
		go func() {
			defer wg.Done()
			node := &apiv1.Node{
				Spec: apiv1.NodeSpec{
					ProviderID: kamateraServerName2,
				},
			}
			ng, err := kcp.NodeGroupForNode(node)
			assert.NoError(t, err)
			assert.Equal(t, ng2, ng)
		}()
	}

	wg.Wait()
}

func TestCloudProvider_NodeGroupsWhileModifyingMap(t *testing.T) {
	cfg := strings.NewReader(`
[global]
kamatera-api-client-id=1a222bbb3ccc44d5555e6ff77g88hh9i
kamatera-api-secret=9ii88h7g6f55555ee4444444dd33eee2
cluster-name=aaabbb
`)
	m, _ := newManager(cfg, nil)
	m.nodeGroups = map[string]*NodeGroup{
		"ng1": {id: "ng1"},
	}
	kcp := &kamateraCloudProvider{manager: m}

	const numGoroutines = 50
	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2)

	// Concurrently call NodeGroups while modifying the nodeGroups map
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			_ = kcp.NodeGroups()
		}()
		go func(idx int) {
			defer wg.Done()
			// Simulate what refresh() does: replace nodeGroups map
			m.nodeGroupsMu.Lock()
			m.nodeGroups = map[string]*NodeGroup{
				"ng1":                    {id: "ng1"},
				fmt.Sprintf("ng%d", idx): {id: fmt.Sprintf("ng%d", idx)},
			}
			m.nodeGroupsMu.Unlock()
		}(i)
	}

	wg.Wait()
}

func TestCloudProvider_NodeGroupForNodeWhileModifyingMap(t *testing.T) {
	cfg := strings.NewReader(`
[global]
kamatera-api-client-id=1a222bbb3ccc44d5555e6ff77g88hh9i
kamatera-api-secret=9ii88h7g6f55555ee4444444dd33eee2
cluster-name=aaabbb
`)
	m, _ := newManager(cfg, nil)
	kamateraServerName1 := mockKamateraServerName()
	ng1 := &NodeGroup{
		id: "ng1",
		instances: map[string]*Instance{
			kamateraServerName1: {Id: kamateraServerName1},
		},
	}
	m.nodeGroups = map[string]*NodeGroup{
		"ng1": ng1,
	}
	kcp := &kamateraCloudProvider{manager: m}

	const numGoroutines = 50
	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2)

	// Concurrently call NodeGroupForNode while modifying the nodeGroups map
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			node := &apiv1.Node{
				Spec: apiv1.NodeSpec{
					ProviderID: kamateraServerName1,
				},
			}
			_, _ = kcp.NodeGroupForNode(node)
		}()
		go func(idx int) {
			defer wg.Done()
			// Simulate what refresh() does: replace nodeGroups map
			m.nodeGroupsMu.Lock()
			m.nodeGroups = map[string]*NodeGroup{
				"ng1":                    ng1,
				fmt.Sprintf("ng%d", idx): {id: fmt.Sprintf("ng%d", idx)},
			}
			m.nodeGroupsMu.Unlock()
		}(i)
	}

	wg.Wait()
}
