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

package linodego

import (
	"context"
	"net/http"
	"strconv"
	"testing"

	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const createLKEClusterPoolResponse1 = `
{"id": 19933, "type": "g6-standard-1", "count": 1, "nodes": [{"id": "19933-5ff4aabd3176", "instance_id": null, "status": "not_ready"}], "disks": []}
`

const deleteLKEClusterPoolResponse1 = `
{}
`

const listLKEClusterPoolsResponse1 = `
{"data": [{"id": 19930, "type": "g6-standard-1", "count": 2, "nodes": [{"id": "19930-5ff4a5cdc29d", "instance_id": 23810703, "status": "not_ready"}, {"id": "19930-5ff4a5ce2b64", "instance_id": 23810705, "status": "not_ready"}], "disks": []}, {"id": 19931, "type": "g6-standard-2", "count": 2, "nodes": [{"id": "19931-5ff4a5ce8f24", "instance_id": 23810707, "status": "not_ready"}, {"id": "19931-5ff4a5cef13e", "instance_id": 23810704, "status": "not_ready"}], "disks": []}], "page": 1, "pages": 1, "results": 2}
`

const listLKEClusterPoolsResponse2 = `
{"data": [{"id": 19930, "type": "g6-standard-1", "count": 2, "nodes": [{"id": "19930-5ff4a5cdc29d", "instance_id": 23810703, "status": "not_ready"}, {"id": "19930-5ff4a5ce2b64", "instance_id": 23810705, "status": "not_ready"}], "disks": []}, {"id": 19931, "type": "g6-standard-2", "count": 2, "nodes": [{"id": "19931-5ff4a5ce8f24", "instance_id": 23810707, "status": "not_ready"}, {"id": "19931-5ff4a5cef13e", "instance_id": 23810704, "status": "not_ready"}], "disks": []}], "page": 1, "pages": 3, "results": 4}
`

const listLKEClusterPoolsResponse3 = `
{"data": [{"id": 19932, "type": "g6-standard-1", "count": 1, "nodes": [{"id": "19932-5ff4a5cdc29f", "instance_id": 23810705, "status": "not_ready"}], "disks": []}], "page": 2, "pages": 3, "results": 4}
`

const listLKEClusterPoolsResponse4 = `
{"data": [{"id": 19933, "type": "g6-standard-1", "count": 1, "nodes": [{"id": "19932-5ff4a5cdc29a", "instance_id": 23810706, "status": "not_ready"}], "disks": []}], "page": 3, "pages": 3, "results": 4}
`

func TestApiClientRest_CreateLKEClusterPool(t *testing.T) {
	server := NewHttpServerMock(MockFieldContentType, MockFieldResponse)
	defer server.Close()

	client := NewClient(&http.Client{})
	client.SetBaseURL(server.URL)

	clusterID := 16293
	ctx := context.Background()
	createOpts := LKEClusterPoolCreateOptions{
		Count: 1,
		Type:  "g6-standard-1",
		Disks: []LKEClusterPoolDisk{},
	}
	requestPath := "/lke/clusters/" + strconv.Itoa(clusterID) + "/pools"
	server.On("handle", requestPath).Return("application/json", createLKEClusterPoolResponse1).Once()
	pool, err := client.CreateLKEClusterPool(ctx, clusterID, createOpts)

	assert.NoError(t, err)
	assert.Equal(t, 1, pool.Count)
	assert.Equal(t, "g6-standard-1", pool.Type)
	assert.Equal(t, 1, len(pool.Linodes))

	mock.AssertExpectationsForObjects(t, server)
}

func TestApiClientRest_DeleteLKEClusterPool(t *testing.T) {
	server := NewHttpServerMock(MockFieldContentType, MockFieldResponse)
	defer server.Close()

	client := NewClient(&http.Client{})
	client.SetBaseURL(server.URL)

	clusterID := 111
	poolID := 222
	ctx := context.Background()
	requestPath := "/lke/clusters/" + strconv.Itoa(clusterID) + "/pools/" + strconv.Itoa(poolID)
	server.On("handle", requestPath).Return("application/json", deleteLKEClusterPoolResponse1).Once()
	err := client.DeleteLKEClusterPool(ctx, clusterID, poolID)
	assert.NoError(t, err)

	mock.AssertExpectationsForObjects(t, server)
}

func TestApiClientRest_ListLKEClusterPools(t *testing.T) {
	server := NewHttpServerMock(MockFieldContentType, MockFieldResponse)
	defer server.Close()

	client := NewClient(&http.Client{})
	client.SetBaseURL(server.URL)

	clusterID := 16293
	ctx := context.Background()
	requestPath := "/lke/clusters/" + strconv.Itoa(clusterID) + "/pools"
	server.On("handle", requestPath).Return("application/json", listLKEClusterPoolsResponse1).Once().On("handle", requestPath).Return("application/json", listLKEClusterPoolsResponse2).Once().On("handle", requestPath).Return("application/json", listLKEClusterPoolsResponse3).Once().On("handle", requestPath).Return("application/json", listLKEClusterPoolsResponse4).Once()

	pools, err := client.ListLKEClusterPools(ctx, clusterID, nil)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(pools))
	assert.Equal(t, 19930, pools[0].ID)
	assert.Equal(t, 19931, pools[1].ID)

	pools, err = client.ListLKEClusterPools(ctx, clusterID, nil)
	assert.NoError(t, err)
	assert.Equal(t, 4, len(pools))
	assert.Equal(t, 19930, pools[0].ID)
	assert.Equal(t, 19931, pools[1].ID)
	assert.Equal(t, 19932, pools[2].ID)
	assert.Equal(t, 19933, pools[3].ID)

	mock.AssertExpectationsForObjects(t, server)
}
