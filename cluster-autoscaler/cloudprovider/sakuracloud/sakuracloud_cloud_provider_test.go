/*
Copyright The Kubernetes Authors.

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

package sakuracloud

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func testNode(name, providerID string) *apiv1.Node {
	return &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec:       apiv1.NodeSpec{ProviderID: providerID},
	}
}

func TestProviderIDHelpers(t *testing.T) {
	assert.Equal(t, "sakuracloud://is1a/pool-abc12", providerIDForServer("is1a", "pool-abc12"))

	name, err := serverNameFromProviderID("sakuracloud://is1a/pool-abc12")
	assert.NoError(t, err)
	assert.Equal(t, "pool-abc12", name)

	for _, invalid := range []string{"", "k3s://node-1", "aws:///us-east-1a/i-abc", "sakuracloud://zoneonly"} {
		_, err := serverNameFromProviderID(invalid)
		assert.Error(t, err, invalid)
	}
}

func TestNodeGroupForNode(t *testing.T) {
	m := &sakuracloudManager{
		zone: "is1a",
		servers: []sakuraServer{
			{ID: "101", Name: "pool-abc12", Tags: []string{groupTagPrefix + "pool"}},
			{ID: "102", Name: "untagged-1", Tags: []string{}},
		},
		nodeGroups: map[string]*sakuracloudNodeGroup{},
	}
	ng := &sakuracloudNodeGroup{id: "pool", manager: m, config: &nodeGroupConfig{MaxSize: 2}}
	m.nodeGroups["pool"] = ng
	provider := &sakuracloudCloudProvider{manager: m}

	// Managed node resolves to its node group.
	group, err := provider.NodeGroupForNode(testNode("pool-abc12", "sakuracloud://is1a/pool-abc12"))
	assert.NoError(t, err)
	assert.Equal(t, ng, group)

	// Nodes from other providers are unmanaged: nil, nil per the contract.
	for _, providerID := range []string{"k3s://control-plane-1", "gce://p/z/i", ""} {
		group, err := provider.NodeGroupForNode(testNode("foreign", providerID))
		assert.NoError(t, err)
		assert.Nil(t, group)
	}

	// A sakuracloud node without a group tag is unmanaged too.
	group, err = provider.NodeGroupForNode(testNode("untagged-1", "sakuracloud://is1a/untagged-1"))
	assert.NoError(t, err)
	assert.Nil(t, group)

	// An unknown server name is unmanaged.
	group, err = provider.NodeGroupForNode(testNode("gone", "sakuracloud://is1a/gone"))
	assert.NoError(t, err)
	assert.Nil(t, group)
}

func TestServerGroupName(t *testing.T) {
	s := sakuraServer{Tags: []string{"other-tag", groupTagPrefix + "pool"}}
	name, ok := s.groupName()
	assert.True(t, ok)
	assert.Equal(t, "pool", name)

	s = sakuraServer{Tags: []string{"other-tag"}}
	_, ok = s.groupName()
	assert.False(t, ok)
}

func TestNewManagerConfigValidation(t *testing.T) {
	t.Setenv("SAKURACLOUD_ACCESS_TOKEN", "token")
	t.Setenv("SAKURACLOUD_ACCESS_TOKEN_SECRET", "secret")

	for name, cfg := range map[string]string{
		"invalid json": "{",
		"missing zone": `{"nodeGroups":{"a":{"minSize":0,"maxSize":1,"core":1,"memoryGB":1,"diskGB":20,"sourceArchiveID":"1","startupNoteID":"2"}}}`,
		"no groups":    `{"zone":"is1a","nodeGroups":{}}`,
		"bad sizes":    `{"zone":"is1a","nodeGroups":{"a":{"minSize":0,"maxSize":1,"core":0,"memoryGB":1,"diskGB":20,"sourceArchiveID":"1","startupNoteID":"2"}}}`,
		"missing ids":  `{"zone":"is1a","nodeGroups":{"a":{"minSize":0,"maxSize":1,"core":1,"memoryGB":1,"diskGB":20}}}`,
	} {
		t.Setenv("SAKURACLOUD_CLUSTER_CONFIG", cfg)
		_, err := newManager()
		assert.Error(t, err, name)
	}
}

func TestDoRequestAuthAndErrors(t *testing.T) {
	srv := newFakeAPI(t)
	defer srv.close()

	m := srv.manager()
	var out struct {
		Servers []sakuraServer `json:"Servers"`
	}
	err := m.doRequest(http.MethodGet, "/server", map[string]interface{}{"Count": 500}, &out)
	assert.NoError(t, err)
	assert.Equal(t, "token", srv.lastUser)
	assert.Equal(t, "secret", srv.lastPass)

	err = m.doRequest(http.MethodGet, "/missing", nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "status 404")
}
