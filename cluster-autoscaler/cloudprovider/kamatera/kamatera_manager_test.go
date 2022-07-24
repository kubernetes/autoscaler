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
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestManager_newManager(t *testing.T) {
	cfg := strings.NewReader(`
[globalxxx]
`)
	_, err := newManager(cfg, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "can't store data at section \"globalxxx\"")

	cfg = strings.NewReader(`
[global]
kamatera-api-client-id=1a222bbb3ccc44d5555e6ff77g88hh9i
kamatera-api-secret=9ii88h7g6f55555ee4444444dd33eee2
cluster-name=aaabbb
`)
	_, err = newManager(cfg, nil)
	assert.NoError(t, err)
}

func TestManager_refresh(t *testing.T) {
	cfg := strings.NewReader(fmt.Sprintf(`
[global]
kamatera-api-client-id=1a222bbb3ccc44d5555e6ff77g88hh9i
kamatera-api-secret=9ii88h7g6f55555ee4444444dd33eee2
cluster-name=aaabbb
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

[nodegroup "ng2"]
min-size=4
max-size=5
`))
	m, err := newManager(cfg, nil)
	assert.NoError(t, err)

	client := kamateraClientMock{}
	m.client = &client
	ctx := context.Background()

	serverName1 := mockKamateraServerName()
	serverName2 := mockKamateraServerName()
	serverName3 := mockKamateraServerName()
	serverName4 := mockKamateraServerName()
	client.On(
		"ListServers", ctx, m.instances,
	).Return(
		[]Server{
			{Name: serverName1, Tags: []string{fmt.Sprintf("%s%s", clusterServerTagPrefix, "aaabbb"), fmt.Sprintf("%s%s", nodeGroupTagPrefix, "ng1")}},
			{Name: serverName2, Tags: []string{fmt.Sprintf("%s%s", nodeGroupTagPrefix, "ng1"), fmt.Sprintf("%s%s", clusterServerTagPrefix, "aaabbb")}},
			{Name: serverName3, Tags: []string{fmt.Sprintf("%s%s", nodeGroupTagPrefix, "ng1"), fmt.Sprintf("%s%s", clusterServerTagPrefix, "aaabbb")}},
			{Name: serverName4, Tags: []string{fmt.Sprintf("%s%s", nodeGroupTagPrefix, "ng2"), fmt.Sprintf("%s%s", clusterServerTagPrefix, "aaabbb")}},
		},
		nil,
	).Once()
	err = m.refresh()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(m.nodeGroups))
	assert.Equal(t, 3, len(m.nodeGroups["ng1"].instances))
	assert.Equal(t, 1, len(m.nodeGroups["ng2"].instances))

	// test api error
	client.On(
		"ListServers", ctx, m.instances,
	).Return(
		[]Server{},
		fmt.Errorf("error on API call"),
	).Once()
	err = m.refresh()
	assert.Error(t, err)
	assert.Equal(t, "failed to get list of Kamatera servers from Kamatera API: error on API call", err.Error())
}

func TestManager_refreshInvalidServerConfiguration(t *testing.T) {
	cfgString := ""
	assertRefreshServerConfigError(t, cfgString, "script for node group ng1 is empty")
	cfgString = "default-script-base64=invalid"
	assertRefreshServerConfigError(t, cfgString, "failed to decode script for node group ng1: illegal base64 data")
	cfgString = "default-script-base64=ZGVmYXVsdAo="
	assertRefreshServerConfigError(t, cfgString, "datacenter for node group ng1 is empty")
	cfgString += "\ndefault-datacenter=IL"
	assertRefreshServerConfigError(t, cfgString, "image for node group ng1 is empty")
	cfgString += "\ndefault-image=ubuntu"
	assertRefreshServerConfigError(t, cfgString, "cpu for node group ng1 is empty")
	cfgString += "\ndefault-cpu=1a"
	assertRefreshServerConfigError(t, cfgString, "ram for node group ng1 is empty")
	cfgString += "\ndefault-ram=1024"
	assertRefreshServerConfigError(t, cfgString, "no disks for node group ng1")
	cfgString += "\ndefault-disk=size=10"
	assertRefreshServerConfigError(t, cfgString, "no networks for node group ng1")
	cfgString += "\ndefault-network=name=wan,ip=auto"
	assertRefreshServerConfigError(t, cfgString, "")
}

func assertRefreshServerConfigError(t *testing.T, cfgString string, expectedError string) {
	cfg := strings.NewReader(fmt.Sprintf(`
[global]
kamatera-api-client-id=1a222bbb3ccc44d5555e6ff77g88hh9i
kamatera-api-secret=9ii88h7g6f55555ee4444444dd33eee2
cluster-name=aaabbb
%s

[nodegroup "ng1"]
`, cfgString))
	m, err := newManager(cfg, nil)
	assert.NoError(t, err)
	client := kamateraClientMock{}
	m.client = &client
	ctx := context.Background()
	serverName1 := mockKamateraServerName()
	client.On(
		"ListServers", ctx, m.instances,
	).Return(
		[]Server{
			{Name: serverName1, Tags: []string{fmt.Sprintf("%s%s", clusterServerTagPrefix, "aaabbb"), fmt.Sprintf("%s%s", nodeGroupTagPrefix, "ng1")}},
		},
		nil,
	).Once()
	err = m.refresh()
	if expectedError == "" {
		assert.NoError(t, err)
	} else {
		assert.Error(t, err)
		assert.Contains(t, err.Error(), expectedError)
	}
}
