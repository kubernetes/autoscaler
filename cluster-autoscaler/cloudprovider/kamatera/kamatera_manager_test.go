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
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
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

[nodegroup "ng2"]
min-size=4
max-size=5
`))
	m, err := newManager(cfg, nil)
	assert.NoError(t, err)

	client := kamateraClientMock{}
	m.client = &client
	ctx := context.Background()

	serverName1 := "myprefix" + mockKamateraServerName()
	serverName2 := "myprefix" + mockKamateraServerName()
	serverName3 := "myprefix" + mockKamateraServerName()
	serverName4 := "myprefix" + mockKamateraServerName()
	client.On(
		"ListServers", ctx, m.instances, "myprefix",
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
		"ListServers", ctx, m.instances, "myprefix",
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
		"ListServers", ctx, m.instances, "",
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

func TestManager_addInstance(t *testing.T) {
	cfg := strings.NewReader(`
[global]
kamatera-api-client-id=1a222bbb3ccc44d5555e6ff77g88hh9i
kamatera-api-secret=9ii88h7g6f55555ee4444444dd33eee2
cluster-name=aaabbb
`)
	m, err := newManager(cfg, nil)
	assert.NoError(t, err)

	serverName1 := mockKamateraServerName()
	server1 := Server{
		Name:    serverName1,
		PowerOn: true,
		Tags:    []string{"tag1", "tag2"},
	}

	// Test adding a new instance
	instance, err := m.addInstance(server1, cloudprovider.InstanceCreating)
	assert.NoError(t, err)
	assert.NotNil(t, instance)
	assert.Equal(t, serverName1, instance.Id)
	assert.Equal(t, cloudprovider.InstanceCreating, instance.Status.State)
	assert.True(t, instance.PowerOn)
	assert.Equal(t, []string{"tag1", "tag2"}, instance.Tags)

	// Verify instance was added to manager's instances map
	assert.NotNil(t, m.instances[serverName1])
	assert.Equal(t, serverName1, m.instances[serverName1].Id)

	// Test updating an existing instance
	server1Updated := Server{
		Name:    serverName1,
		PowerOn: false,
		Tags:    []string{"tag3"},
	}
	instanceUpdated, err := m.addInstance(server1Updated, cloudprovider.InstanceRunning)
	assert.NoError(t, err)
	assert.NotNil(t, instanceUpdated)
	assert.Equal(t, serverName1, instanceUpdated.Id)
	assert.Equal(t, cloudprovider.InstanceRunning, instanceUpdated.Status.State)
	assert.False(t, instanceUpdated.PowerOn)
	assert.Equal(t, []string{"tag3"}, instanceUpdated.Tags)

	// Verify the updated instance in the map
	assert.Equal(t, cloudprovider.InstanceRunning, m.instances[serverName1].Status.State)
	assert.False(t, m.instances[serverName1].PowerOn)
}

func TestManager_snapshotInstances(t *testing.T) {
	cfg := strings.NewReader(`
[global]
kamatera-api-client-id=1a222bbb3ccc44d5555e6ff77g88hh9i
kamatera-api-secret=9ii88h7g6f55555ee4444444dd33eee2
cluster-name=aaabbb
`)
	m, err := newManager(cfg, nil)
	assert.NoError(t, err)

	serverName1 := mockKamateraServerName()
	serverName2 := mockKamateraServerName()
	m.instances[serverName1] = &Instance{Id: serverName1, PowerOn: true}
	m.instances[serverName2] = &Instance{Id: serverName2, PowerOn: false}

	// Get snapshot
	snapshot := m.snapshotInstances()

	// Verify snapshot has same content
	assert.Equal(t, 2, len(snapshot))
	assert.Equal(t, serverName1, snapshot[serverName1].Id)
	assert.Equal(t, serverName2, snapshot[serverName2].Id)

	// Verify snapshot is independent copy (modifying snapshot doesn't affect original)
	delete(snapshot, serverName1)
	assert.Equal(t, 1, len(snapshot))
	assert.Equal(t, 2, len(m.instances))

	// Verify original still has both instances
	assert.NotNil(t, m.instances[serverName1])
	assert.NotNil(t, m.instances[serverName2])
}

func TestManager_addInstanceConcurrent(t *testing.T) {
	cfg := strings.NewReader(`
[global]
kamatera-api-client-id=1a222bbb3ccc44d5555e6ff77g88hh9i
kamatera-api-secret=9ii88h7g6f55555ee4444444dd33eee2
cluster-name=aaabbb
`)
	m, err := newManager(cfg, nil)
	assert.NoError(t, err)

	const numGoroutines = 50
	done := make(chan bool, numGoroutines)

	// Concurrently add instances
	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			serverName := fmt.Sprintf("server-%d", idx)
			server := Server{
				Name:    serverName,
				PowerOn: true,
				Tags:    []string{"tag1"},
			}
			_, err := m.addInstance(server, cloudprovider.InstanceRunning)
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify all instances were added
	assert.Equal(t, numGoroutines, len(m.instances))
}

func TestManager_snapshotInstancesConcurrent(t *testing.T) {
	cfg := strings.NewReader(`
[global]
kamatera-api-client-id=1a222bbb3ccc44d5555e6ff77g88hh9i
kamatera-api-secret=9ii88h7g6f55555ee4444444dd33eee2
cluster-name=aaabbb
`)
	m, err := newManager(cfg, nil)
	assert.NoError(t, err)

	// Pre-populate some instances
	for i := 0; i < 10; i++ {
		serverName := fmt.Sprintf("server-%d", i)
		m.instances[serverName] = &Instance{Id: serverName, PowerOn: true}
	}

	const numGoroutines = 50
	done := make(chan bool, numGoroutines*2)

	// Concurrently snapshot and add instances
	for i := 0; i < numGoroutines; i++ {
		// Snapshot goroutine
		go func() {
			snapshot := m.snapshotInstances()
			assert.NotNil(t, snapshot)
			done <- true
		}()
		// Add instance goroutine
		go func(idx int) {
			serverName := fmt.Sprintf("new-server-%d", idx)
			server := Server{
				Name:    serverName,
				PowerOn: true,
				Tags:    []string{"tag1"},
			}
			_, err := m.addInstance(server, cloudprovider.InstanceRunning)
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines*2; i++ {
		<-done
	}
}
