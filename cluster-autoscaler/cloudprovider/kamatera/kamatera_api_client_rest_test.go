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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"testing"
)

const (
	mockKamateraClientId = "mock-client-id"
	mockKamateraSecret   = "mock-secret"
)

func NewMockKamateraApiClientRest(url string, maxRetries int, expSecondsBetweenRetries int) (client KamateraApiClientRest) {
	return KamateraApiClientRest{
		userAgent:                userAgent,
		clientId:                 mockKamateraClientId,
		secret:                   mockKamateraSecret,
		url:                      url,
		maxRetries:               maxRetries,
		expSecondsBetweenRetries: expSecondsBetweenRetries,
	}
}

func TestApiClientRest_ListServers_NoServers(t *testing.T) {
	server := NewHttpServerMock(MockFieldContentType, MockFieldResponse)
	defer server.Close()
	ctx := context.Background()
	client := NewMockKamateraApiClientRest(server.URL, 5, 0)
	server.On("handle", "/service/servers").Return(
		"application/json",
		`[]`,
	).Once()
	servers, err := client.ListServers(ctx, map[string]*Instance{}, "")
	assert.NoError(t, err)
	assert.Equal(t, 0, len(servers))
	mock.AssertExpectationsForObjects(t, server)
}

func TestApiClientRest_ListServers(t *testing.T) {
	server := NewHttpServerMock(MockFieldContentType, MockFieldResponse)
	defer server.Close()
	ctx := context.Background()
	client := NewMockKamateraApiClientRest(server.URL, 5, 0)
	newServerName1 := mockKamateraServerName()
	cachedServerName2 := mockKamateraServerName()
	cachedServerName3 := mockKamateraServerName()
	cachedServerName4 := mockKamateraServerName()
	server.On("handle", "/service/servers").Return(
		"application/json",
		fmt.Sprintf(`[
	{"name": "%s", "power": "on"},
	{"name": "%s", "power": "on"},
	{"name": "%s", "power": "off"},
	{"name": "%s", "power": "on"}
]`, newServerName1, cachedServerName2, cachedServerName3, cachedServerName4),
	).On("handle", "/server/tags").Return(
		"application/json",
		`[{"tag name": "test-tag"}, {"tag name": "other-test-tag"}]`,
	)
	servers, err := client.ListServers(ctx, map[string]*Instance{
		cachedServerName2: {
			Id:      cachedServerName2,
			Status:  &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning},
			PowerOn: true,
			Tags:    []string{"my-tag", "my-other-tag"},
		},
		cachedServerName3: {
			Id:      cachedServerName3,
			Status:  &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning},
			PowerOn: true,
			Tags:    []string{"another-tag", "my-other-tag"},
		},
		cachedServerName4: {
			Id:      cachedServerName4,
			Status:  &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning},
			PowerOn: true,
			Tags:    []string{},
		},
	}, "")
	assert.NoError(t, err)
	assert.Equal(t, 4, len(servers))
	assert.Equal(t, servers, []Server{
		{
			Name:    newServerName1,
			Tags:    []string{"test-tag", "other-test-tag"},
			PowerOn: true,
		},
		{
			Name:    cachedServerName2,
			Tags:    []string{"my-tag", "my-other-tag"},
			PowerOn: true,
		},
		{
			Name:    cachedServerName3,
			Tags:    []string{"another-tag", "my-other-tag"},
			PowerOn: false,
		},
		{
			Name:    cachedServerName4,
			Tags:    []string{},
			PowerOn: true,
		},
	})
	mock.AssertExpectationsForObjects(t, server)
}

func TestApiClientRest_ListServersNamePrefix(t *testing.T) {
	server := NewHttpServerMock(MockFieldContentType, MockFieldResponse)
	defer server.Close()
	ctx := context.Background()
	client := NewMockKamateraApiClientRest(server.URL, 5, 0)
	newServerName1 := "prefixa" + mockKamateraServerName()
	newServerName2 := "prefixb" + mockKamateraServerName()
	server.On("handle", "/service/servers").Return(
		"application/json",
		fmt.Sprintf(`[
	{"name": "%s", "power": "on"},
	{"name": "%s", "power": "on"}
]`, newServerName1, newServerName2),
	).On("handle", "/server/tags").Return(
		"application/json",
		`[{"tag name": "test-tag"}, {"tag name": "other-test-tag"}]`,
	)
	servers, err := client.ListServers(ctx, map[string]*Instance{}, "prefixb")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(servers))
	assert.Equal(t, servers, []Server{
		{
			Name:    newServerName2,
			Tags:    []string{"test-tag", "other-test-tag"},
			PowerOn: true,
		},
	})
	mock.AssertExpectationsForObjects(t, server)
}

func TestApiClientRest_ListServersNoTags(t *testing.T) {
	server := NewHttpServerMock(MockFieldContentType, MockFieldResponse)
	defer server.Close()
	ctx := context.Background()
	client := NewMockKamateraApiClientRest(server.URL, 5, 0)
	newServerName1 := mockKamateraServerName()
	server.On("handle", "/service/servers").Return(
		"application/json", fmt.Sprintf(`[{"name": "%s", "power": "on"}]`, newServerName1),
	).On("handle", "/server/tags").Return(
		"application/json", `[]`,
	)
	servers, err := client.ListServers(ctx, map[string]*Instance{}, "")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(servers))
	assert.Equal(t, servers, []Server{
		{
			Name:    newServerName1,
			Tags:    []string{},
			PowerOn: true,
		},
	})
	mock.AssertExpectationsForObjects(t, server)
}

func TestApiClientRest_ListServersTagsError(t *testing.T) {
	server := NewHttpServerMock(MockFieldContentType, MockFieldResponse, MockFieldStatusCode)
	defer server.Close()
	ctx := context.Background()
	client := NewMockKamateraApiClientRest(server.URL, 5, 0)
	newServerName1 := mockKamateraServerName()
	server.On("handle", "/service/servers").Return(
		"application/json", fmt.Sprintf(`[{"name": "%s", "power": "on"}]`, newServerName1), 200,
	).On("handle", "/server/tags").Return(
		"application/json", `{"message":"Failed to run command method (serverTags)"}`, 500,
	)
	servers, err := client.ListServers(ctx, map[string]*Instance{}, "")
	assert.Error(t, err)
	assert.Nil(t, servers)
}

func TestApiClientRest_DeleteServer(t *testing.T) {
	server := NewHttpServerMock(MockFieldContentType, MockFieldResponse)
	defer server.Close()
	ctx := context.Background()
	client := NewMockKamateraApiClientRest(server.URL, 5, 0)
	serverName := mockKamateraServerName()
	commandId := "mock-command-id"
	server.On("handle", "/service/server/poweroff").Return(
		"application/json",
		fmt.Sprintf(`["%s"]`, commandId),
	).Once().On("handle", "/service/queue").Return(
		"application/json",
		`[{"status": "complete"}]`,
	).Once().On("handle", "/service/server/terminate").Return(
		"application/json",
		"{}",
	).Once()
	err := client.DeleteServer(ctx, serverName)
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, server)
}

func TestApiClientRest_DeleteServer_TerminateError(t *testing.T) {
	server := NewHttpServerMock(MockFieldContentType, MockFieldResponse, MockFieldStatusCode)
	defer server.Close()
	ctx := context.Background()
	client := NewMockKamateraApiClientRest(server.URL, 5, 0)
	serverName := mockKamateraServerName()
	commandId := "mock-command-id"
	server.On("handle", "/service/server/poweroff").Return(
		"application/json", fmt.Sprintf(`["%s"]`, commandId), 200,
	).Once().On("handle", "/service/queue").Return(
		"application/json", `[{"status": "complete"}]`, 200,
	).Once().On("handle", "/service/server/terminate").Return(
		"application/json",
		"Gateway Timeout",
		504,
	).Times(5)
	err := client.DeleteServer(ctx, serverName)
	assert.Error(t, err)
}

func TestApiClientRest_CreateServers(t *testing.T) {
	server := NewHttpServerMock(MockFieldContentType, MockFieldResponse)
	defer server.Close()
	ctx := context.Background()
	client := NewMockKamateraApiClientRest(server.URL, 5, 0)
	commandId := "command"
	server.On("handle", "/service/server").Return(
		"application/json",
		fmt.Sprintf(`["%s"]`, commandId),
	).Twice().On("handle", "/service/queue").Return(
		"application/json",
		`[{"status": "complete"}]`,
	).Twice()
	servers, err := client.CreateServers(ctx, 2, mockServerConfig("test", []string{"foo", "bar"}))
	assert.NoError(t, err)
	assert.Equal(t, 2, len(servers))
	assert.Less(t, 10, len(servers[0].Name))
	assert.Less(t, 10, len(servers[1].Name))
	assert.Equal(t, servers[0].Tags, []string{"foo", "bar"})
	assert.Equal(t, servers[1].Tags, []string{"foo", "bar"})
	mock.AssertExpectationsForObjects(t, server)
}
