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

package brightbox

import (
	"encoding/json"
	"flag"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	brightbox "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/brightbox/gobrightbox"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/brightbox/k8ssdk"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/brightbox/k8ssdk/mocks"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	klog "k8s.io/klog/v2"
)

const (
	fakeServer      = "srv-testy"
	fakeGroup       = "grp-testy"
	missingServer   = "srv-notty"
	fakeClusterName = "k8s-fake.cluster.local"
)

var (
	fakeNodeMap = map[string]string{
		fakeServer: fakeGroup,
	}
	fakeNodeGroup = &brightboxNodeGroup{
		id: fakeGroup,
	}
	fakeNodeGroups = []cloudprovider.NodeGroup{
		fakeNodeGroup,
	}
)

func init() {
	klog.InitFlags(nil)
	flag.Set("alsologtostderr", "true")
	flag.Set("v", "4")
}

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

func TestName(t *testing.T) {
	assert.Equal(t, makeFakeCloudProvider(nil).Name(), cloudprovider.BrightboxProviderName)
}

func TestGPULabel(t *testing.T) {
	assert.Equal(t, makeFakeCloudProvider(nil).GPULabel(), GPULabel)
}

func TestGetAvailableGPUTypes(t *testing.T) {
	assert.Equal(t, makeFakeCloudProvider(nil).GetAvailableGPUTypes(), availableGPUTypes)
}

func TestPricing(t *testing.T) {
	obj, err := makeFakeCloudProvider(nil).Pricing()
	assert.Equal(t, err, cloudprovider.ErrNotImplemented)
	assert.Nil(t, obj)
}

func TestGetAvailableMachineTypes(t *testing.T) {
	obj, err := makeFakeCloudProvider(nil).GetAvailableMachineTypes()
	assert.Equal(t, err, cloudprovider.ErrNotImplemented)
	assert.Nil(t, obj)
}

func TestNewNodeGroup(t *testing.T) {
	obj, err := makeFakeCloudProvider(nil).NewNodeGroup("", nil, nil, nil, nil)
	assert.Equal(t, err, cloudprovider.ErrNotImplemented)
	assert.Nil(t, obj)
}

func TestCleanUp(t *testing.T) {
	assert.Nil(t, makeFakeCloudProvider(nil).Cleanup())
}

func TestResourceLimiter(t *testing.T) {
	client := makeFakeCloudProvider(nil)
	obj, err := client.GetResourceLimiter()
	assert.Equal(t, obj, client.resourceLimiter)
	assert.NoError(t, err)
}

func TestNodeGroups(t *testing.T) {
	client := makeFakeCloudProvider(nil)
	assert.Zero(t, client.NodeGroups())
	client.nodeGroups = make([]cloudprovider.NodeGroup, 0)
	assert.NotZero(t, client.NodeGroups())
	assert.Empty(t, client.NodeGroups())
	nodeGroup := &brightboxNodeGroup{}
	client.nodeGroups = append(client.nodeGroups, nodeGroup)
	newGroups := client.NodeGroups()
	assert.Len(t, newGroups, 1)
	assert.Same(t, newGroups[0], client.nodeGroups[0])
}

func TestNodeGroupForNode(t *testing.T) {
	client := makeFakeCloudProvider(nil)
	client.nodeGroups = fakeNodeGroups
	client.nodeMap = fakeNodeMap
	nodeGroup, err := client.NodeGroupForNode(makeNode(fakeServer))
	assert.Equal(t, fakeNodeGroup, nodeGroup)
	assert.NoError(t, err)
	nodeGroup, err = client.NodeGroupForNode(makeNode(missingServer))
	assert.Nil(t, nodeGroup)
	assert.NoError(t, err)
}

func TestBuildBrightBox(t *testing.T) {
	ts := k8ssdk.GetAuthEnvTokenHandler(t)
	defer k8ssdk.ResetAuthEnvironment()
	defer ts.Close()
	rl := cloudprovider.NewResourceLimiter(nil, nil)
	do := cloudprovider.NodeGroupDiscoveryOptions{}
	opts := config.AutoscalingOptions{
		CloudProviderName: cloudprovider.BrightboxProviderName,
		ClusterName:       fakeClusterName,
	}
	cloud := BuildBrightbox(opts, do, rl)
	assert.Equal(t, cloud.Name(), cloudprovider.BrightboxProviderName)
	obj, err := cloud.GetResourceLimiter()
	assert.Equal(t, rl, obj)
	assert.NoError(t, err)
}

func testOsExit(t *testing.T, funcName string, testFunc func(*testing.T)) {
	if os.Getenv(funcName) == "1" {
		testFunc(t)
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run="+funcName)
	cmd.Env = append(os.Environ(), funcName+"=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("%s subprocess ran successfully, want non-zero exit status", funcName)
}

func TestBuildBrightboxMissingClusterName(t *testing.T) {
	testOsExit(t, "TestBuildBrightboxMissingClusterName", func(t *testing.T) {
		ts := k8ssdk.GetAuthEnvTokenHandler(t)
		defer k8ssdk.ResetAuthEnvironment()
		defer ts.Close()
		rl := cloudprovider.NewResourceLimiter(nil, nil)
		do := cloudprovider.NodeGroupDiscoveryOptions{}
		opts := config.AutoscalingOptions{
			CloudProviderName: cloudprovider.BrightboxProviderName,
		}
		BuildBrightbox(opts, do, rl)
	})
}

func TestRefresh(t *testing.T) {
	mockclient := new(mocks.CloudAccess)
	testclient := k8ssdk.MakeTestClient(mockclient, nil)
	provider := makeFakeCloudProvider(testclient)
	groups := fakeGroups()
	mockclient.On("ServerGroup", "grp-sda44").Return(fakeServerGroupsda44(), nil)
	mockclient.On("ConfigMaps").Return(fakeConfigMaps(), nil)
	mockclient.On("ConfigMap", "cfg-502vh").Return(fakeConfigMap502vh(), nil)
	err := provider.Refresh()
	require.NoError(t, err)
	assert.Len(t, provider.nodeGroups, 1)
	assert.NotEmpty(t, provider.nodeMap)
	node, err := provider.NodeGroupForNode(makeNode("srv-lv426"))
	assert.NoError(t, err)
	require.NotNil(t, node)
	assert.Equal(t, node.Id(), groups[0].Id)
	node, err = provider.NodeGroupForNode(makeNode("srv-rp897"))
	assert.NoError(t, err)
	require.NotNil(t, node)
	assert.Equal(t, node.Id(), groups[0].Id)
	mockclient.AssertExpectations(t)
}

func TestFetchDefaultGroup(t *testing.T) {
	groups := fakeGroups()
	groupID := fetchDefaultGroup(groups, "fred")
	assert.Empty(t, groupID)
	groupID = fetchDefaultGroup(groups, groups[0].Name)
	assert.Equal(t, groups[0].Id, groupID)
}

func makeNode(serverID string) *v1.Node {
	return &v1.Node{
		Spec: v1.NodeSpec{
			ProviderID: k8ssdk.MapServerIDToProviderID(serverID),
		},
	}
}

func makeFakeCloudProvider(brightboxCloudClient *k8ssdk.Cloud) *brightboxCloudProvider {
	return &brightboxCloudProvider{
		resourceLimiter: &cloudprovider.ResourceLimiter{},
		ClusterName:     fakeClusterName,
		Cloud:           brightboxCloudClient,
	}
}

func fakeConfigMaps() []brightbox.ConfigMap {
	const groupjson = `
[{
  "id": "cfg-502vh",
  "resource_type": "config_map",
  "url": "https://api.gb1.brightbox.com/1.0/config_maps/cfg-502vh",
  "name": "storage.k8s-fake.cluster.local",
  "data": {
   "image": "img-svqx9",
   "max": "4",
   "min": "1",
   "region": "gb1",
   "server_group": "grp-sda44",
   "default_group": "grp-vnr33",
   "type": "2gb.ssd",
   "user_data": "fake_userdata",
   "zone": ""
  }
 }]
     `
	var result []brightbox.ConfigMap
	_ = json.NewDecoder(strings.NewReader(groupjson)).Decode(&result)
	return result
}

func fakeConfigMap502vh() *brightbox.ConfigMap {
	const groupjson = `
{
  "id": "cfg-502vh",
  "resource_type": "config_map",
  "url": "https://api.gb1.brightbox.com/1.0/config_maps/cfg-502vh",
  "name": "storage.k8s-fake.cluster.local",
  "data": {
   "image": "img-svqx9",
   "max": "4",
   "min": "1",
   "region": "gb1",
   "server_group": "grp-sda44",
   "default_group": "grp-vnr33",
   "type": "2gb.ssd",
   "user_data": "fake_userdata",
   "zone": ""
  }
 }
     `
	var result brightbox.ConfigMap
	_ = json.NewDecoder(strings.NewReader(groupjson)).Decode(&result)
	return &result
}

func fakeServerGroupsda44() *brightbox.ServerGroup {
	const groupjson = `
{"id": "grp-sda44",
  "resource_type": "server_group",
  "url": "https://api.gb1.brightbox.com/1.0/server_groups/grp-sda44",
  "name": "storage.k8s-fake.cluster.local",
  "description": "1:4",
  "created_at": "2011-10-01T00:00:00Z",
  "default": true,
  "account":
   {"id": "acc-43ks4",
    "resource_type": "account",
    "url": "https://api.gb1.brightbox.com/1.0/accounts/acc-43ks4",
    "name": "Brightbox",
    "status": "active"},
  "firewall_policy":
   {"id": "fwp-j3654",
    "resource_type": "firewall_policy",
    "url": "https://api.gb1.brightbox.com/1.0/firewall_policies/fwp-j3654",
    "default": true,
    "name": "default",
    "created_at": "2011-10-01T00:00:00Z",
    "description": null},
  "servers":
   [
   {"id": "srv-lv426",
     "resource_type": "server",
     "url": "https://api.gb1.brightbox.com/1.0/servers/srv-lv426",
     "name": "",
     "status": "active",
     "locked": false,
     "hostname": "srv-lv426",
     "fqdn": "srv-lv426.gb1.brightbox.com",
     "created_at": "2011-10-01T01:00:00Z",
     "started_at": "2011-10-01T01:01:00Z",
     "deleted_at": null},
   {"id": "srv-rp897",
     "resource_type": "server",
     "url": "https://api.gb1.brightbox.com/1.0/servers/srv-rp897",
     "name": "",
     "status": "active",
     "locked": false,
     "hostname": "srv-rp897",
     "fqdn": "srv-rp897.gb1.brightbox.com",
     "created_at": "2011-10-01T01:00:00Z",
     "started_at": "2011-10-01T01:01:00Z",
     "deleted_at": null}
     ]}
     `
	var result brightbox.ServerGroup
	_ = json.NewDecoder(strings.NewReader(groupjson)).Decode(&result)
	return &result
}

func fakeGroups() []brightbox.ServerGroup {
	const groupjson = `
[{"id": "grp-sda44",
  "resource_type": "server_group",
  "url": "https://api.gb1.brightbox.com/1.0/server_groups/grp-sda44",
  "name": "storage.k8s-fake.cluster.local",
  "description": "1:4",
  "created_at": "2011-10-01T00:00:00Z",
  "default": true,
  "account":
   {"id": "acc-43ks4",
    "resource_type": "account",
    "url": "https://api.gb1.brightbox.com/1.0/accounts/acc-43ks4",
    "name": "Brightbox",
    "status": "active"},
  "firewall_policy":
   {"id": "fwp-j3654",
    "resource_type": "firewall_policy",
    "url": "https://api.gb1.brightbox.com/1.0/firewall_policies/fwp-j3654",
    "default": true,
    "name": "default",
    "created_at": "2011-10-01T00:00:00Z",
    "description": null},
  "servers":
   [
   {"id": "srv-lv426",
     "resource_type": "server",
     "url": "https://api.gb1.brightbox.com/1.0/servers/srv-lv426",
     "name": "",
     "status": "active",
     "locked": false,
     "hostname": "srv-lv426",
     "fqdn": "srv-lv426.gb1.brightbox.com",
     "created_at": "2011-10-01T01:00:00Z",
     "started_at": "2011-10-01T01:01:00Z",
     "deleted_at": null},
   {"id": "srv-rp897",
     "resource_type": "server",
     "url": "https://api.gb1.brightbox.com/1.0/servers/srv-rp897",
     "name": "",
     "status": "active",
     "locked": false,
     "hostname": "srv-rp897",
     "fqdn": "srv-rp897.gb1.brightbox.com",
     "created_at": "2011-10-01T01:00:00Z",
     "started_at": "2011-10-01T01:01:00Z",
     "deleted_at": null}
     ]}]
     `
	var result []brightbox.ServerGroup
	_ = json.NewDecoder(strings.NewReader(groupjson)).Decode(&result)
	return result
}

func fakeServerlv426() *brightbox.Server {
	const serverjson = `
{"id": "srv-lv426",
 "resource_type": "server",
 "url": "https://api.gb1.brightbox.com/1.0/servers/srv-lv426",
 "name": "storage-0.storage.k8s-fake.cluster.local",
 "status": "active",
 "locked": false,
 "hostname": "srv-lv426",
 "created_at": "2011-10-01T01:00:00Z",
 "started_at": "2011-10-01T01:01:00Z",
 "deleted_at": null,
 "user_data": null,
 "fqdn": "srv-lv426.gb1.brightbox.com",
 "compatibility_mode": false,
 "console_url": null,
 "console_token": null,
 "console_token_expires": null,
 "account":
  {"id": "acc-43ks4",
   "resource_type": "account",
   "url": "https://api.gb1.brightbox.com/1.0/accounts/acc-43ks4",
   "name": "Brightbox",
   "status": "active"},
 "image":
  {"id": "img-3ikco",
   "resource_type": "image",
   "url": "https://api.gb1.brightbox.com/1.0/images/img-3ikco",
   "name": "Ubuntu Lucid 10.04 server",
   "username": "ubuntu",
   "status": "available",
   "locked": false,
   "description": "Expands root partition automatically. login: ubuntu using stored ssh key",
   "source": "ubuntu-lucid-daily-i64-server-20110509",
   "arch": "x86_64",
   "created_at": "2011-05-09T12:00:00Z",
   "official": true,
   "public": true,
   "owner": "acc-43ks4"},
 "server_type":
  {"id": "typ-zx45f",
   "resource_type": "server_type",
   "url": "https://api.gb1.brightbox.com/1.0/server_types/typ-zx45f",
   "name": "Small",
   "status": "available",
   "cores": 2,
   "ram": 2048,
   "disk_size": 81920,
   "handle": "small"},
 "zone":
  {"id": "zon-328ds",
   "resource_type": "zone",
   "url": "https://api.gb1.brightbox.com/1.0/zones/zon-328ds",
   "handle": "gb1"},
 "cloud_ips":
  [{"id": "cip-k4a25",
    "resource_type": "cloud_ip",
    "url": "https://api.gb1.brightbox.com/1.0/cloud_ips/cip-k4a25",
    "status": "mapped",
    "public_ip": "109.107.50.0",
    "public_ipv4": "109.107.50.0",
    "public_ipv6": "2a02:1348:ffff:ffff::6d6b:3200",
    "fqdn": "cip-k4a25.gb1.brightbox.com",
    "reverse_dns": null,
    "name": "product website ip"}],
 "interfaces":
  [{"id": "int-ds42k",
    "resource_type": "interface",
    "url": "https://api.gb1.brightbox.com/1.0/interfaces/int-ds42k",
    "mac_address": "02:24:19:00:00:ee",
    "ipv4_address": "81.15.16.17"}],
 "snapshots":
  [],
 "server_groups":
  [{"id": "grp-sda44",
    "resource_type": "server_group",
    "url": "https://api.gb1.brightbox.com/1.0/server_groups/grp-sda44",
    "name": "",
    "description": null,
    "created_at": "2011-10-01T00:00:00Z",
    "default": true}]}
`
	var result brightbox.Server
	_ = json.NewDecoder(strings.NewReader(serverjson)).Decode(&result)
	return &result
}

func fakeServerTypezx45f() *brightbox.ServerType {
	const serverjson = `
{"id": "typ-zx45f",
  "resource_type": "server_type",
  "url": "https://api.gb1.brightbox.com/1.0/server_types/typ-zx45f",
  "name": "Small",
  "status": "available",
  "cores": 2,
  "ram": 2048,
  "disk_size": 81920,
  "handle": "small"}
`
	var result brightbox.ServerType
	_ = json.NewDecoder(strings.NewReader(serverjson)).Decode(&result)
	return &result
}

func fakeServerrp897() *brightbox.Server {
	const serverjson = `
{"id": "srv-rp897",
 "resource_type": "server",
 "url": "https://api.gb1.brightbox.com/1.0/servers/srv-rp897",
 "name": "storage-0.storage.k8s-fake.cluster.local",
 "status": "active",
 "locked": false,
 "hostname": "srv-rp897",
 "created_at": "2011-10-01T01:00:00Z",
 "started_at": "2011-10-01T01:01:00Z",
 "deleted_at": null,
 "user_data": null,
 "fqdn": "srv-rp897.gb1.brightbox.com",
 "compatibility_mode": false,
 "console_url": null,
 "console_token": null,
 "console_token_expires": null,
 "account":
  {"id": "acc-43ks4",
   "resource_type": "account",
   "url": "https://api.gb1.brightbox.com/1.0/accounts/acc-43ks4",
   "name": "Brightbox",
   "status": "active"},
 "image":
  {"id": "img-3ikco",
   "resource_type": "image",
   "url": "https://api.gb1.brightbox.com/1.0/images/img-3ikco",
   "name": "Ubuntu Lucid 10.04 server",
   "username": "ubuntu",
   "status": "available",
   "locked": false,
   "description": "Expands root partition automatically. login: ubuntu using stored ssh key",
   "source": "ubuntu-lucid-daily-i64-server-20110509",
   "arch": "x86_64",
   "created_at": "2011-05-09T12:00:00Z",
   "official": true,
   "public": true,
   "owner": "acc-43ks4"},
 "server_type":
  {"id": "typ-zx45f",
   "resource_type": "server_type",
   "url": "https://api.gb1.brightbox.com/1.0/server_types/typ-zx45f",
   "name": "Small",
   "status": "available",
   "cores": 2,
   "ram": 2048,
   "disk_size": 81920,
   "handle": "small"},
 "zone":
  {"id": "zon-328ds",
   "resource_type": "zone",
   "url": "https://api.gb1.brightbox.com/1.0/zones/zon-328ds",
   "handle": "gb1"},
 "cloud_ips":
  [{"id": "cip-k4a25",
    "resource_type": "cloud_ip",
    "url": "https://api.gb1.brightbox.com/1.0/cloud_ips/cip-k4a25",
    "status": "mapped",
    "public_ip": "109.107.50.0",
    "public_ipv4": "109.107.50.0",
    "public_ipv6": "2a02:1348:ffff:ffff::6d6b:3200",
    "fqdn": "cip-k4a25.gb1.brightbox.com",
    "reverse_dns": null,
    "name": "product website ip"}],
 "interfaces":
  [{"id": "int-ds42k",
    "resource_type": "interface",
    "url": "https://api.gb1.brightbox.com/1.0/interfaces/int-ds42k",
    "mac_address": "02:24:19:00:00:ee",
    "ipv4_address": "81.15.16.17"}],
 "snapshots":
  [],
 "server_groups":
  [{"id": "grp-sda44",
    "resource_type": "server_group",
    "url": "https://api.gb1.brightbox.com/1.0/server_groups/grp-sda44",
    "name": "",
    "description": null,
    "created_at": "2011-10-01T00:00:00Z",
    "default": true}]}
`
	var result brightbox.Server
	_ = json.NewDecoder(strings.NewReader(serverjson)).Decode(&result)
	return &result
}

func fakeServertesty() *brightbox.Server {
	const serverjson = `
{"id": "srv-testy",
 "resource_type": "server",
 "url": "https://api.gb1.brightbox.com/1.0/servers/srv-testy",
 "name": "storage-0.storage.k8s-fake.cluster.local",
 "status": "active",
 "locked": false,
 "hostname": "srv-testy",
 "created_at": "2011-10-01T01:00:00Z",
 "started_at": "2011-10-01T01:01:00Z",
 "deleted_at": null,
 "user_data": null,
 "fqdn": "srv-testy.gb1.brightbox.com",
 "compatibility_mode": false,
 "console_url": null,
 "console_token": null,
 "console_token_expires": null,
 "account":
  {"id": "acc-43ks4",
   "resource_type": "account",
   "url": "https://api.gb1.brightbox.com/1.0/accounts/acc-43ks4",
   "name": "Brightbox",
   "status": "active"},
 "image":
  {"id": "img-3ikco",
   "resource_type": "image",
   "url": "https://api.gb1.brightbox.com/1.0/images/img-3ikco",
   "name": "Ubuntu Lucid 10.04 server",
   "username": "ubuntu",
   "status": "available",
   "locked": false,
   "description": "Expands root partition automatically. login: ubuntu using stored ssh key",
   "source": "ubuntu-lucid-daily-i64-server-20110509",
   "arch": "x86_64",
   "created_at": "2011-05-09T12:00:00Z",
   "official": true,
   "public": true,
   "owner": "acc-43ks4"},
 "server_type":
  {"id": "typ-zx45f",
   "resource_type": "server_type",
   "url": "https://api.gb1.brightbox.com/1.0/server_types/typ-zx45f",
   "name": "Small",
   "status": "available",
   "cores": 2,
   "ram": 2048,
   "disk_size": 81920,
   "handle": "small"},
 "zone":
  {"id": "zon-328ds",
   "resource_type": "zone",
   "url": "https://api.gb1.brightbox.com/1.0/zones/zon-328ds",
   "handle": "gb1"},
 "cloud_ips":
  [{"id": "cip-k4a25",
    "resource_type": "cloud_ip",
    "url": "https://api.gb1.brightbox.com/1.0/cloud_ips/cip-k4a25",
    "status": "mapped",
    "public_ip": "109.107.50.0",
    "public_ipv4": "109.107.50.0",
    "public_ipv6": "2a02:1348:ffff:ffff::6d6b:3200",
    "fqdn": "cip-k4a25.gb1.brightbox.com",
    "reverse_dns": null,
    "name": "product website ip"}],
 "interfaces":
  [{"id": "int-ds42k",
    "resource_type": "interface",
    "url": "https://api.gb1.brightbox.com/1.0/interfaces/int-ds42k",
    "mac_address": "02:24:19:00:00:ee",
    "ipv4_address": "81.15.16.17"}],
 "snapshots":
  [],
 "server_groups":
  [{"id": "grp-testy",
    "resource_type": "server_group",
    "url": "https://api.gb1.brightbox.com/1.0/server_groups/grp-testy",
    "name": "",
    "description": null,
    "created_at": "2011-10-01T00:00:00Z",
    "default": true}]}
`
	var result brightbox.Server
	_ = json.NewDecoder(strings.NewReader(serverjson)).Decode(&result)
	return &result
}

func fakeServerGroupsPlusOne() []brightbox.ServerGroup {
	const groupjson = `
[{"id": "grp-sda44",
  "resource_type": "server_group",
  "url": "https://api.gb1.brightbox.com/1.0/server_groups/grp-sda44",
  "name": "storage.k8s-fake.cluster.local",
  "description": "1:4",
  "created_at": "2011-10-01T00:00:00Z",
  "default": true,
  "account":
   {"id": "acc-43ks4",
    "resource_type": "account",
    "url": "https://api.gb1.brightbox.com/1.0/accounts/acc-43ks4",
    "name": "Brightbox",
    "status": "active"},
  "firewall_policy":
   {"id": "fwp-j3654",
    "resource_type": "firewall_policy",
    "url": "https://api.gb1.brightbox.com/1.0/firewall_policies/fwp-j3654",
    "default": true,
    "name": "default",
    "created_at": "2011-10-01T00:00:00Z",
    "description": null},
  "servers":
   [
   {"id": "srv-lv426",
     "resource_type": "server",
     "url": "https://api.gb1.brightbox.com/1.0/servers/srv-lv426",
     "name": "",
     "status": "active",
     "locked": false,
     "hostname": "srv-lv426",
     "fqdn": "srv-lv426.gb1.brightbox.com",
     "created_at": "2011-10-01T01:00:00Z",
     "started_at": "2011-10-01T01:01:00Z",
     "deleted_at": null},
   {"id": "srv-testy",
     "resource_type": "server",
     "url": "https://api.gb1.brightbox.com/1.0/servers/srv-testy",
     "name": "",
     "status": "active",
     "locked": false,
     "hostname": "srv-testy",
     "fqdn": "srv-testy.gb1.brightbox.com",
     "created_at": "2011-10-01T01:00:00Z",
     "started_at": "2011-10-01T01:01:00Z",
     "deleted_at": null},
   {"id": "srv-rp897",
     "resource_type": "server",
     "url": "https://api.gb1.brightbox.com/1.0/servers/srv-rp897",
     "name": "",
     "status": "active",
     "locked": false,
     "hostname": "srv-rp897",
     "fqdn": "srv-rp897.gb1.brightbox.com",
     "created_at": "2011-10-01T01:00:00Z",
     "started_at": "2011-10-01T01:01:00Z",
     "deleted_at": null}
     ]}]
     `
	var result []brightbox.ServerGroup
	_ = json.NewDecoder(strings.NewReader(groupjson)).Decode(&result)
	return result
}

func deletedFakeServer(server *brightbox.Server) *brightbox.Server {
	now := time.Now()
	result := *server
	result.DeletedAt = &now
	result.Status = "deleted"
	result.ServerGroups = []brightbox.ServerGroup{}
	return &result
}
