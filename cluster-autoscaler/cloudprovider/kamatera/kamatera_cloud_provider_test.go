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
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/client-go/kubernetes/fake"
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
	m, _ := newManager(
		cfg,
		fake.NewSimpleClientset(),
	)
	kamateraServerName1 := mockKamateraServerName()
	kamateraCloudProviderID1 := formatKamateraProviderID(defaultKamateraProviderIDPrefix, kamateraServerName1)
	kamateraServerName2 := mockKamateraServerName()
	kamateraCloudProviderID2 := formatKamateraProviderID(defaultKamateraProviderIDPrefix, kamateraServerName2)
	kamateraServerName3 := mockKamateraServerName()
	kamateraCloudProviderID3 := formatKamateraProviderID(defaultKamateraProviderIDPrefix, kamateraServerName3)
	kamateraServerName4 := mockKamateraServerName()
	kamateraCloudProviderID4 := formatKamateraProviderID(defaultKamateraProviderIDPrefix, kamateraServerName4)
	ng1 := &NodeGroup{
		id: "ng1",
		instances: map[string]*Instance{
			kamateraCloudProviderID1: {Id: kamateraCloudProviderID1, Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning}},
			kamateraCloudProviderID2: {Id: kamateraCloudProviderID2, Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning}},
		},
		manager: m,
	}
	ng2 := &NodeGroup{
		id: "ng2",
		instances: map[string]*Instance{
			kamateraCloudProviderID3: {Id: kamateraCloudProviderID3, Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning}},
			kamateraCloudProviderID4: {Id: kamateraCloudProviderID4, Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning}},
		},
		manager: m,
	}
	m.nodeGroups = map[string]*NodeGroup{
		"ng1": ng1,
		"ng2": ng2,
	}
	kcp := &kamateraCloudProvider{manager: m}

	// test ok on getting the right node group for an apiv1.Node based on its ProviderID
	node := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: formatKamateraProviderID(defaultKamateraProviderIDPrefix, kamateraServerName1),
		},
	}
	ng, err := kcp.NodeGroupForNode(node)
	assert.NoError(t, err)
	assert.Equal(t, ng1, ng)

	// test ok on getting the right node group for an apiv1.Node based on name
	node = &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: kamateraServerName4,
		},
	}
	ng, err = kcp.NodeGroupForNode(node)
	assert.NoError(t, err)
	assert.Equal(t, ng2, ng)

	// test ok on getting nil when looking for a apiv1.Node we do not manage
	node = &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "non-existing-node",
		},
		Spec: apiv1.NodeSpec{
			ProviderID: "kamatera://---",
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
	kubeClient := fake.NewClientset()
	m.kubeClient = kubeClient

	kcp := &kamateraCloudProvider{manager: m}

	serverName1 := "myprefix" + mockKamateraServerName()
	serverName2 := "myprefix" + mockKamateraServerName()
	client.On(
		"ListServers", mock.Anything, m.instances, "myprefix", defaultKamateraProviderIDPrefix,
	).Return(
		[]Server{
			{
				Name: serverName1,
				Tags: []string{
					fmt.Sprintf("%s%s", clusterServerTagPrefix, "aaabbb"),
					fmt.Sprintf("%s%s", nodeGroupTagPrefix, "ng1"),
				},
				PowerOn: true,
			},
			{
				Name: serverName2,
				Tags: []string{
					fmt.Sprintf("%s%s", clusterServerTagPrefix, "aaabbb"),
					fmt.Sprintf("%s%s", nodeGroupTagPrefix, "ng1"),
				},
				PowerOn: false,
			},
		},
		nil,
	).Once()
	assert.NoError(t, kubeClient.Tracker().Add(&apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: serverName1},
		Spec:       apiv1.NodeSpec{ProviderID: formatKamateraProviderID(defaultKamateraProviderIDPrefix, serverName1)},
	}))
	err = kcp.Refresh()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(m.nodeGroups))
	assert.Equal(t, 2, len(m.nodeGroups["ng1"].instances))
	// TargetSize is updated only after NodeGroupForNode is called with a related node
	ng, err := kcp.NodeGroupForNode(&apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: serverName1,
		},
	})
	assert.Equal(t, m.nodeGroups["ng1"], ng)
	targetSize, _ := m.nodeGroups["ng1"].TargetSize()
	assert.Equal(t, 1, targetSize)
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
	kamateraCloudProviderID1 := formatKamateraProviderID(defaultKamateraProviderIDPrefix, kamateraServerName1)
	kamateraServerName2 := mockKamateraServerName()
	kamateraCloudProviderID2 := formatKamateraProviderID(defaultKamateraProviderIDPrefix, kamateraServerName2)
	ng1 := &NodeGroup{
		id: "ng1",
		instances: map[string]*Instance{
			kamateraCloudProviderID1: {Id: kamateraCloudProviderID1, Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning}},
		},
		manager: m,
	}
	ng2 := &NodeGroup{
		id: "ng2",
		instances: map[string]*Instance{
			kamateraCloudProviderID2: {Id: kamateraCloudProviderID2, Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning}},
		},
		manager: m,
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
					ProviderID: kamateraCloudProviderID1,
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
					ProviderID: kamateraCloudProviderID2,
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

func TestCreateKubeClient(t *testing.T) {
	// Create a temporary kubeconfig file
	tmpDir := t.TempDir()
	kubeConfigPath := filepath.Join(tmpDir, "kubeconfig")

	kubeConfigContent := `apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://localhost:6443
    insecure-skip-tls-verify: true
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
current-context: test-context
users:
- name: test-user
  user:
    token: test-token
`
	err := os.WriteFile(kubeConfigPath, []byte(kubeConfigContent), 0600)
	assert.NoError(t, err)

	opts := config.AutoscalingOptions{
		KubeClientOpts: config.KubeClientOptions{
			KubeConfigPath: kubeConfigPath,
		},
	}

	client := createKubeClient(opts)
	assert.NotNil(t, client)
}

func kcpKClientOnListServers(kclient *kamateraClientMock, kcp *kamateraCloudProvider, servers []Server, err error) *mock.Call {
	return kclient.On(
		"ListServers", mock.Anything, kcp.manager.snapshotInstances(), "", kcp.manager.config.providerIDPrefix,
	).Return(
		servers, err,
	)
}

func kcpGetNodeGroup(kcp *kamateraCloudProvider, ngID string) *NodeGroup {
	for _, candidate := range kcp.manager.nodeGroups {
		if candidate.Id() == ngID {
			return candidate
		}
	}
	return nil
}

type kcpExpectedInstance struct {
	State        *cloudprovider.InstanceState
	CommandId    string
	HasErrorInfo bool
}

func kcpAssertNodeGroup(
	kcp *kamateraCloudProvider, t *testing.T, ngID string,
	expectedTargetSize int,
	expectedNumNodes int,
	expectedInstances map[string]kcpExpectedInstance,
) *NodeGroup {
	ng := kcpGetNodeGroup(kcp, ngID)
	assert.NotNil(t, ng, "node group %s not found", ngID)
	targetSize, err := ng.TargetSize()
	assert.NoError(t, err, "failed to get target size for node group %s", ngID)
	assert.Equal(t, expectedTargetSize, targetSize, "target size for node group %s does not match", ngID)
	nodes, err := ng.Nodes()
	assert.NoError(t, err, "failed to get nodes for node group %s", ngID)
	assert.Equal(t, expectedNumNodes, len(nodes), "number of nodes in node group %s does not match", ngID)
	assert.Equal(t, len(expectedInstances), len(ng.instances), "number of instances in node group %s does not match", ngID)
	for serverName, expectedInstance := range expectedInstances {
		instanceId := formatKamateraProviderID(kcp.manager.config.providerIDPrefix, serverName)
		actualInstance, exists := ng.instances[instanceId]
		assert.True(t, exists, "instance %s not found in node group %s", instanceId, ngID)
		if expectedInstance.State == nil {
			assert.Nil(t, actualInstance.Status, "expected instance %s in node group %s to have nil status", instanceId, ngID)
			assert.False(t, expectedInstance.HasErrorInfo, "expected instance %s in node group %s to not have error info when state is nil", instanceId, ngID)
		} else {
			assert.NotNil(t, actualInstance.Status, "expected instance %s in node group %s to have non-nil status", instanceId, ngID)
			assert.Equal(t, *expectedInstance.State, actualInstance.Status.State, "unexpected state for instance %s in node group %s", instanceId, ngID)
			assert.Equal(t, expectedInstance.HasErrorInfo, actualInstance.Status.ErrorInfo != nil, "unexpected error info presence for instance %s in node group %s", instanceId, ngID)
		}
		assert.Equal(t, expectedInstance.CommandId, actualInstance.StatusCommandId, "unexpected command for instance %s in node group %s", instanceId, ngID)
	}
	return ng
}

func TestCloudProviderScalingFlow(t *testing.T) {
	tests := []struct {
		name                string
		poweronOnScaleUp    bool
		poweroffOnScaleDown bool
	}{
		{
			name:                "default",
			poweronOnScaleUp:    false,
			poweroffOnScaleDown: false,
		},
		{
			name:                "poweron on scale up",
			poweronOnScaleUp:    true,
			poweroffOnScaleDown: false,
		},
		{
			name:                "poweroff on scale down",
			poweronOnScaleUp:    false,
			poweroffOnScaleDown: true,
		},
		{
			name:                "poweron on scale up and poweroff on scale down",
			poweronOnScaleUp:    true,
			poweroffOnScaleDown: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kamateraClient := &kamateraClientMock{}
			kubeClient := fake.NewClientset()
			configString := strings.NewReader(`
[global]
kamatera-api-client-id=1a222bbb3ccc44d5555e6ff77g88hh9i
kamatera-api-secret=9ii88h7g6f55555ee4444444dd33eee2
cluster-name=kamatera
poweron-on-scale-up=` + fmt.Sprintf("%t", tt.poweronOnScaleUp) + `
poweroff-on-scale-down=` + fmt.Sprintf("%t", tt.poweroffOnScaleDown) + `
default-script-base64=SGVsbG8sIHdvcmxkIQo=
default-datacenter=US-NY2
default-image=ubuntu
default-cpu=2D
default-ram=2048
default-disk=size=20
default-network=name=wan,ip=auto

[nodegroup "ng1"]
min-size=0
max-size=3
`)
			cfg, err := buildCloudConfig(configString)
			assert.NoError(t, err)
			m := &manager{
				client:     kamateraClient,
				config:     cfg,
				nodeGroups: make(map[string]*NodeGroup),
				instances:  make(map[string]*Instance),
				kubeClient: kubeClient,
			}
			kcp := kamateraCloudProvider{
				manager: m,
			}
			assert.Equal(t, kcp.manager.config.PoweronOnScaleUp, tt.poweronOnScaleUp)
			assert.Equal(t, kcp.manager.config.PoweroffOnScaleDown, tt.poweroffOnScaleDown)
			// initial refresh - no servers, no instances in node group
			kcpKClientOnListServers(kamateraClient, &kcp, []Server{}, nil).Once()
			assert.NoError(t, kcp.Refresh())
			ng := kcpAssertNodeGroup(&kcp, t, "ng1", 0, 0, map[string]kcpExpectedInstance{})
			// scale up to 3 instances
			// 2 servers started creation, 1 server failed to start creation
			kamateraClient.On("StartCreateServers", mock.Anything, 3, mock.Anything).Return(
				map[string]string{
					"server1": "create-server-command-id-1",
					"server2": "",
					"server3": "create-server-command-id-3",
				},
				nil,
			).Once()
			assert.NoError(t, ng.IncreaseSize(3))
			instanceCreating := cloudprovider.InstanceCreating
			kcpAssertNodeGroup(&kcp, t, "ng1", 3, 3, map[string]kcpExpectedInstance{
				"server1": {State: &instanceCreating, CommandId: "create-server-command-id-1", HasErrorInfo: false},
				"server2": {State: &instanceCreating, CommandId: "", HasErrorInfo: false},
				"server3": {State: &instanceCreating, CommandId: "create-server-command-id-3", HasErrorInfo: false},
			})
			// server 3 finished creation with error, server 1 still creating
			kcpKClientOnListServers(kamateraClient, &kcp, []Server{}, nil).Once()
			kamateraClient.On("getCommandStatus", mock.Anything, "create-server-command-id-1").Return(
				CommandStatusPending, nil,
			).Once()
			kamateraClient.On("getCommandStatus", mock.Anything, "create-server-command-id-3").Return(
				CommandStatusError, nil,
			).Once()
			assert.NoError(t, kcp.Refresh())
			kcpAssertNodeGroup(&kcp, t, "ng1", 3, 3, map[string]kcpExpectedInstance{
				"server1": {State: &instanceCreating, CommandId: "create-server-command-id-1", HasErrorInfo: false},
				"server2": {State: &instanceCreating, CommandId: "", HasErrorInfo: false},
				"server3": {State: &instanceCreating, CommandId: "", HasErrorInfo: true},
			})
			// server 1 create command completed successfully, but still server is not listed in Kamatera API
			kcpKClientOnListServers(kamateraClient, &kcp, []Server{}, nil).Once()
			kamateraClient.On("getCommandStatus", mock.Anything, "create-server-command-id-1").Return(
				CommandStatusComplete, nil,
			).Once()
			assert.NoError(t, kcp.Refresh())
			kcpAssertNodeGroup(&kcp, t, "ng1", 3, 3, map[string]kcpExpectedInstance{
				"server1": {State: &instanceCreating, CommandId: "", HasErrorInfo: false},
				"server2": {State: &instanceCreating, CommandId: "", HasErrorInfo: false},
				"server3": {State: &instanceCreating, CommandId: "", HasErrorInfo: true},
			})
			// now server 1 is listed but not powered on yet - doesn't change the instances state
			kcpKClientOnListServers(kamateraClient, &kcp, []Server{
				{Name: "server1", PowerOn: false, Tags: ng.serverConfig.Tags},
			}, nil).Once()
			assert.NoError(t, kcp.Refresh())
			kcpAssertNodeGroup(&kcp, t, "ng1", 3, 3, map[string]kcpExpectedInstance{
				"server1": {State: &instanceCreating, CommandId: "", HasErrorInfo: false},
				"server2": {State: &instanceCreating, CommandId: "", HasErrorInfo: false},
				"server3": {State: &instanceCreating, CommandId: "", HasErrorInfo: true},
			})
			// now server 1 is powered on - state set to running
			// server 2 unexpectedly appears - but powered off, so state remains creating
			instanceRunning := cloudprovider.InstanceRunning
			kcpKClientOnListServers(kamateraClient, &kcp, []Server{
				{Name: "server1", PowerOn: true, Tags: ng.serverConfig.Tags},
				{Name: "server2", PowerOn: false, Tags: ng.serverConfig.Tags},
			}, nil).Once()
			assert.NoError(t, kcp.Refresh())
			kcpAssertNodeGroup(&kcp, t, "ng1", 3, 3, map[string]kcpExpectedInstance{
				"server1": {State: &instanceRunning, CommandId: "", HasErrorInfo: false},
				"server2": {State: &instanceCreating, CommandId: "", HasErrorInfo: false},
				"server3": {State: &instanceCreating, CommandId: "", HasErrorInfo: true},
			})
			// unexpectedly all 3 servers suddenly appear as powered on and registered in kubernetes
			kubeClient.Tracker().Add(
				&apiv1.Node{
					ObjectMeta: metav1.ObjectMeta{Name: "server2"},
					Spec:       apiv1.NodeSpec{ProviderID: formatKamateraProviderID(kcp.manager.config.providerIDPrefix, "server2")},
				},
			)
			kubeClient.Tracker().Add(
				&apiv1.Node{
					ObjectMeta: metav1.ObjectMeta{Name: "server3"},
					Spec:       apiv1.NodeSpec{ProviderID: formatKamateraProviderID(kcp.manager.config.providerIDPrefix, "server3")},
				},
			)
			kcpKClientOnListServers(kamateraClient, &kcp, []Server{
				{Name: "server1", PowerOn: true, Tags: ng.serverConfig.Tags},
				{Name: "server2", PowerOn: true, Tags: ng.serverConfig.Tags},
				{Name: "server3", PowerOn: true, Tags: ng.serverConfig.Tags},
			}, nil).Once()
			assert.NoError(t, kcp.Refresh())
			kcpAssertNodeGroup(&kcp, t, "ng1", 3, 3, map[string]kcpExpectedInstance{
				"server1": {State: &instanceRunning, CommandId: "", HasErrorInfo: false},
				"server2": {State: &instanceRunning, CommandId: "", HasErrorInfo: false},
				"server3": {State: &instanceRunning, CommandId: "", HasErrorInfo: false},
			})
			// scale down by 2 instances: server1 and server3
			kamateraClient.On("StartServerRequest", mock.Anything, ServerRequestPoweroff, "server1").Return(
				"server1-power-off", nil,
			).Once()
			kamateraClient.On("StartServerRequest", mock.Anything, ServerRequestPoweroff, "server3").Return(
				"server3-power-off", nil,
			).Once()
			assert.NoError(t, ng.DeleteNodes([]*apiv1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "server1"},
					Spec:       apiv1.NodeSpec{ProviderID: formatKamateraProviderID(kcp.manager.config.providerIDPrefix, "server1")},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "server3"},
					Spec:       apiv1.NodeSpec{ProviderID: formatKamateraProviderID(kcp.manager.config.providerIDPrefix, "server3")},
				},
			}))
			instanceDeleting := cloudprovider.InstanceDeleting
			kcpAssertNodeGroup(&kcp, t, "ng1", 1, 3, map[string]kcpExpectedInstance{
				"server1": {State: &instanceDeleting, CommandId: "server1-power-off", HasErrorInfo: false},
				"server2": {State: &instanceRunning, CommandId: "", HasErrorInfo: false},
				"server3": {State: &instanceDeleting, CommandId: "server3-power-off", HasErrorInfo: false},
			})
			// refresh - no change
			kcpKClientOnListServers(kamateraClient, &kcp, []Server{
				{Name: "server1", PowerOn: true, Tags: ng.serverConfig.Tags},
				{Name: "server2", PowerOn: true, Tags: ng.serverConfig.Tags},
				{Name: "server3", PowerOn: true, Tags: ng.serverConfig.Tags},
			}, nil).Once().On(
				"getCommandStatus", mock.Anything, "server1-power-off",
			).Return(CommandStatusPending, nil).Once().On(
				"getCommandStatus", mock.Anything, "server3-power-off",
			).Return(CommandStatusPending, nil).Once()
			assert.NoError(t, kcp.Refresh())
			kcpAssertNodeGroup(&kcp, t, "ng1", 1, 3, map[string]kcpExpectedInstance{
				"server1": {State: &instanceDeleting, CommandId: "server1-power-off", HasErrorInfo: false},
				"server2": {State: &instanceRunning, CommandId: "", HasErrorInfo: false},
				"server3": {State: &instanceDeleting, CommandId: "server3-power-off", HasErrorInfo: false},
			})
			// refresh - server1 disappeared - no change
			kcpKClientOnListServers(kamateraClient, &kcp, []Server{
				{Name: "server2", PowerOn: true, Tags: ng.serverConfig.Tags},
				{Name: "server3", PowerOn: true, Tags: ng.serverConfig.Tags},
			}, nil).Once().On(
				"getCommandStatus", mock.Anything, "server1-power-off",
			).Return(CommandStatusPending, nil).Once().On(
				"getCommandStatus", mock.Anything, "server3-power-off",
			).Return(CommandStatusPending, nil).Once()
			assert.NoError(t, kcp.Refresh())
			kcpAssertNodeGroup(&kcp, t, "ng1", 1, 3, map[string]kcpExpectedInstance{
				"server1": {State: &instanceDeleting, CommandId: "server1-power-off", HasErrorInfo: false},
				"server2": {State: &instanceRunning, CommandId: "", HasErrorInfo: false},
				"server3": {State: &instanceDeleting, CommandId: "server3-power-off", HasErrorInfo: false},
			})
			// refresh - server 1 power off completed
			// server 3 power off failed
			kcpKClientOnListServers(kamateraClient, &kcp, []Server{
				{Name: "server1", PowerOn: false, Tags: ng.serverConfig.Tags},
				{Name: "server2", PowerOn: false, Tags: ng.serverConfig.Tags},
				{Name: "server3", PowerOn: true, Tags: ng.serverConfig.Tags},
			}, nil).Once().On(
				"getCommandStatus", mock.Anything, "server1-power-off",
			).Return(CommandStatusComplete, nil).Once().On(
				"getCommandStatus", mock.Anything, "server3-power-off",
			).Return(CommandStatusError, nil).Once()
			if !tt.poweroffOnScaleDown {
				// continue to terminate
				kamateraClient.On("StartServerTerminate", mock.Anything, "server1", true).Return(
					"server1-terminate", nil,
				).Once()
			}
			assert.NoError(t, kcp.Refresh())
			if tt.poweroffOnScaleDown {
				kcpAssertNodeGroup(&kcp, t, "ng1", 1, 2, map[string]kcpExpectedInstance{
					"server1": {State: nil, CommandId: "", HasErrorInfo: false},
					"server2": {State: &instanceRunning, CommandId: "", HasErrorInfo: false},
					"server3": {State: &instanceDeleting, CommandId: "", HasErrorInfo: true},
				})
			} else {
				kcpAssertNodeGroup(&kcp, t, "ng1", 1, 3, map[string]kcpExpectedInstance{
					"server1": {State: &instanceDeleting, CommandId: "server1-terminate", HasErrorInfo: false},
					"server2": {State: &instanceRunning, CommandId: "", HasErrorInfo: false},
					"server3": {State: &instanceDeleting, CommandId: "", HasErrorInfo: true},
				})
			}
			if tt.poweroffOnScaleDown {
				// servers 1 and 3 powered off
				kcpKClientOnListServers(kamateraClient, &kcp, []Server{
					{Name: "server1", PowerOn: false, Tags: ng.serverConfig.Tags},
					{Name: "server2", PowerOn: true, Tags: ng.serverConfig.Tags},
					{Name: "server3", PowerOn: false, Tags: ng.serverConfig.Tags},
				}, nil).Once()
			} else {
				// servers 1 and 3 disappear
				kcpKClientOnListServers(kamateraClient, &kcp, []Server{
					{Name: "server2", PowerOn: true, Tags: ng.serverConfig.Tags},
				}, nil).Once().On(
					"getCommandStatus", mock.Anything, "server1-terminate",
				).Return(CommandStatusComplete, nil).Once()
			}
			assert.NoError(t, kcp.Refresh())
			if tt.poweroffOnScaleDown {
				kcpAssertNodeGroup(&kcp, t, "ng1", 1, 1, map[string]kcpExpectedInstance{
					"server1": {State: nil, CommandId: "", HasErrorInfo: false},
					"server2": {State: &instanceRunning, CommandId: "", HasErrorInfo: false},
					"server3": {State: nil, CommandId: "", HasErrorInfo: false},
				})
			} else {
				kcpAssertNodeGroup(&kcp, t, "ng1", 1, 1, map[string]kcpExpectedInstance{
					"server2": {State: &instanceRunning, CommandId: "", HasErrorInfo: false},
				})
			}
			// scale up by 2 instances
			if tt.poweronOnScaleUp && tt.poweroffOnScaleDown {
				kamateraClient.On("StartServerRequest", mock.Anything, ServerRequestPoweron, "server1").Return(
					"server1-power-on", nil,
				).Once().On("StartServerRequest", mock.Anything, ServerRequestPoweron, "server3").Return(
					"server3-power-on", nil,
				).Once()
			} else {
				kamateraClient.On("StartCreateServers", mock.Anything, 2, mock.Anything).Return(
					map[string]string{
						"server4": "create-server-command-id-4",
						"server5": "create-server-command-id-5",
					},
					nil,
				).Once()
			}
			assert.NoError(t, ng.IncreaseSize(2))
			if tt.poweronOnScaleUp && tt.poweroffOnScaleDown {
				kcpAssertNodeGroup(&kcp, t, "ng1", 3, 3, map[string]kcpExpectedInstance{
					"server2": {State: &instanceRunning, CommandId: "", HasErrorInfo: false},
					"server1": {State: &instanceCreating, CommandId: "server1-power-on", HasErrorInfo: false},
					"server3": {State: &instanceCreating, CommandId: "server3-power-on", HasErrorInfo: false},
				})
			} else if tt.poweroffOnScaleDown {
				kcpAssertNodeGroup(&kcp, t, "ng1", 3, 3, map[string]kcpExpectedInstance{
					"server1": {State: nil, CommandId: "", HasErrorInfo: false},
					"server2": {State: &instanceRunning, CommandId: "", HasErrorInfo: false},
					"server3": {State: nil, CommandId: "", HasErrorInfo: false},
					"server4": {State: &instanceCreating, CommandId: "create-server-command-id-4", HasErrorInfo: false},
					"server5": {State: &instanceCreating, CommandId: "create-server-command-id-5", HasErrorInfo: false},
				})
			} else {
				kcpAssertNodeGroup(&kcp, t, "ng1", 3, 3, map[string]kcpExpectedInstance{
					"server2": {State: &instanceRunning, CommandId: "", HasErrorInfo: false},
					"server4": {State: &instanceCreating, CommandId: "create-server-command-id-4", HasErrorInfo: false},
					"server5": {State: &instanceCreating, CommandId: "create-server-command-id-5", HasErrorInfo: false},
				})
			}
			// scale up succeeded for one of the instances, other failed
			// but none are yet registered in kubernetes - so no state changes
			if tt.poweronOnScaleUp && tt.poweroffOnScaleDown {
				kcpKClientOnListServers(kamateraClient, &kcp, []Server{
					{Name: "server1", PowerOn: false, Tags: ng.serverConfig.Tags},
					{Name: "server2", PowerOn: true, Tags: ng.serverConfig.Tags},
					{Name: "server3", PowerOn: false, Tags: ng.serverConfig.Tags},
				}, nil).Once().On("getCommandStatus", mock.Anything, "server1-power-on").Return(
					CommandStatusComplete, nil,
				).Once().On("getCommandStatus", mock.Anything, "server3-power-on").Return(
					CommandStatusError, nil,
				).Once()
			} else {
				if tt.poweroffOnScaleDown {
					kcpKClientOnListServers(kamateraClient, &kcp, []Server{
						{Name: "server1", PowerOn: false, Tags: ng.serverConfig.Tags},
						{Name: "server2", PowerOn: true, Tags: ng.serverConfig.Tags},
						{Name: "server3", PowerOn: false, Tags: ng.serverConfig.Tags},
					}, nil).Once()
				} else {
					kcpKClientOnListServers(kamateraClient, &kcp, []Server{
						{Name: "server2", PowerOn: true, Tags: ng.serverConfig.Tags},
					}, nil).Once()
				}
				kamateraClient.On("getCommandStatus", mock.Anything, "create-server-command-id-4").Return(
					CommandStatusComplete, nil,
				).Once().On("getCommandStatus", mock.Anything, "create-server-command-id-5").Return(
					CommandStatusError, nil,
				).Once()
			}
			assert.NoError(t, kcp.Refresh())
			if tt.poweronOnScaleUp && tt.poweroffOnScaleDown {
				kcpAssertNodeGroup(&kcp, t, "ng1", 3, 3, map[string]kcpExpectedInstance{
					"server1": {State: &instanceCreating, CommandId: "", HasErrorInfo: false},
					"server2": {State: &instanceRunning, CommandId: "", HasErrorInfo: false},
					"server3": {State: &instanceCreating, CommandId: "", HasErrorInfo: true},
				})
			} else if tt.poweroffOnScaleDown {
				kcpAssertNodeGroup(&kcp, t, "ng1", 3, 3, map[string]kcpExpectedInstance{
					"server1": {State: nil, CommandId: "", HasErrorInfo: false},
					"server2": {State: &instanceRunning, CommandId: "", HasErrorInfo: false},
					"server3": {State: nil, CommandId: "", HasErrorInfo: false},
					"server4": {State: &instanceCreating, CommandId: "", HasErrorInfo: false},
					"server5": {State: &instanceCreating, CommandId: "", HasErrorInfo: true},
				})
			} else {
				kcpAssertNodeGroup(&kcp, t, "ng1", 3, 3, map[string]kcpExpectedInstance{
					"server2": {State: &instanceRunning, CommandId: "", HasErrorInfo: false},
					"server4": {State: &instanceCreating, CommandId: "", HasErrorInfo: false},
					"server5": {State: &instanceCreating, CommandId: "", HasErrorInfo: true},
				})
			}
			// all servers powered on - behavior depends on poweron/poweroff settings
			if tt.poweronOnScaleUp && tt.poweroffOnScaleDown {
				kcpKClientOnListServers(kamateraClient, &kcp, []Server{
					{Name: "server1", PowerOn: true, Tags: ng.serverConfig.Tags},
					{Name: "server2", PowerOn: true, Tags: ng.serverConfig.Tags},
					{Name: "server3", PowerOn: true, Tags: ng.serverConfig.Tags},
				}, nil).Once()
			} else if tt.poweroffOnScaleDown {
				kcpKClientOnListServers(kamateraClient, &kcp, []Server{
					{Name: "server1", PowerOn: false, Tags: ng.serverConfig.Tags},
					{Name: "server2", PowerOn: true, Tags: ng.serverConfig.Tags},
					{Name: "server3", PowerOn: false, Tags: ng.serverConfig.Tags},
					{Name: "server4", PowerOn: true, Tags: ng.serverConfig.Tags},
					{Name: "server5", PowerOn: true, Tags: ng.serverConfig.Tags},
				}, nil).Once()
			} else {
				kcpKClientOnListServers(kamateraClient, &kcp, []Server{
					{Name: "server2", PowerOn: true, Tags: ng.serverConfig.Tags},
					{Name: "server4", PowerOn: true, Tags: ng.serverConfig.Tags},
					{Name: "server5", PowerOn: true, Tags: ng.serverConfig.Tags},
				}, nil).Once()
			}
			assert.NoError(t, kcp.Refresh())
			if tt.poweronOnScaleUp && tt.poweroffOnScaleDown {
				// instances already registered in Kubernetes - once powered on, state changes to running
				kcpAssertNodeGroup(&kcp, t, "ng1", 3, 3, map[string]kcpExpectedInstance{
					"server1": {State: &instanceRunning, CommandId: "", HasErrorInfo: false},
					"server2": {State: &instanceRunning, CommandId: "", HasErrorInfo: false},
					"server3": {State: &instanceRunning, CommandId: "", HasErrorInfo: false},
				})
			} else if tt.poweroffOnScaleDown {
				kcpAssertNodeGroup(&kcp, t, "ng1", 3, 3, map[string]kcpExpectedInstance{
					"server1": {State: nil, CommandId: "", HasErrorInfo: false},
					"server2": {State: &instanceRunning, CommandId: "", HasErrorInfo: false},
					"server3": {State: nil, CommandId: "", HasErrorInfo: false},
					"server4": {State: &instanceRunning, CommandId: "", HasErrorInfo: false},
					"server5": {State: &instanceRunning, CommandId: "", HasErrorInfo: false},
				})
			} else {
				kcpAssertNodeGroup(&kcp, t, "ng1", 3, 3, map[string]kcpExpectedInstance{
					"server2": {State: &instanceRunning, CommandId: "", HasErrorInfo: false},
					"server4": {State: &instanceRunning, CommandId: "", HasErrorInfo: false},
					"server5": {State: &instanceRunning, CommandId: "", HasErrorInfo: false},
				})
			}
		})
	}
}
