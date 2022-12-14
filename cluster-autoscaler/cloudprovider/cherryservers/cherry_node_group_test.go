/*
Copyright 2022 The Kubernetes Authors.

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

package cherryservers

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	apiv1 "k8s.io/api/core/v1"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"

	"github.com/stretchr/testify/assert"
)

const (
	createCherryServerResponsePool2 = ``
	deleteCherryServerResponsePool2 = ``
	createCherryServerResponsePool3 = ``
	deleteCherryServerResponsePool3 = ``
)

func TestIncreaseDecreaseSize(t *testing.T) {
	var m *cherryManagerRest
	memServers := []Server{
		{ID: 1000, Name: "server-1000", Hostname: "k8s-cluster2-pool3-gndxdmmw", State: "active", Tags: map[string]string{"k8s-cluster": "cluster2", "k8s-nodepool": "pool3"}},
		{ID: 1001, Name: "server-1001", Hostname: "k8s-cluster2-master", State: "active", Tags: map[string]string{"k8s-cluster": "cluster2"}},
	}
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()
	assert.Equal(t, true, true)
	if useRealEndpoint {
		// If auth token set in env, hit the actual Cherry API
		m = newTestCherryManagerRest(t, "")
	} else {
		// Set up a mock Cherry API
		m = newTestCherryManagerRest(t, server.URL)
		// the flow needs to match our actual calls below
		mux.HandleFunc(fmt.Sprintf("/projects/%d/servers", m.nodePools["default"].projectID), func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case "GET":
				b, _ := json.Marshal(memServers)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(200)
				w.Write(b)
				return
			case "POST":
				b, err := io.ReadAll(r.Body)
				if err != nil {
					w.WriteHeader(500)
					w.Write([]byte("could not read request body"))
					return
				}
				var createRequest CreateServer
				if err := json.Unmarshal(b, &createRequest); err != nil {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(400)
					w.Write([]byte(`{"error": "invalid body"}`))
					return
				}
				planSlug := createRequest.Plan
				if err != nil {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(400)
					w.Write([]byte(`{"error": "invalid plan slug"}`))
					return
				}
				if createRequest.ProjectID != m.nodePools["default"].projectID {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(400)
					w.Write([]byte(`{"error": "mismatched project ID in body and path"}`))
					return
				}
				projectID := createRequest.ProjectID
				if err != nil {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(400)
					w.Write([]byte(`{"error": "invalid project ID"}`))
					return
				}
				server := Server{
					ID:       rand.Intn(10000),
					Name:     createRequest.Hostname,
					Hostname: createRequest.Hostname,
					Plan:     Plan{Slug: planSlug},
					Project:  Project{ID: projectID},
					Image:    createRequest.Image,
					Tags:     *createRequest.Tags,
					//UserData: createRequest.UserData,
				}
				memServers = append(memServers, server)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(201)
				b, _ = json.Marshal(server)
				w.Write(b)
				return
			}
		})
		mux.HandleFunc("/servers/", func(w http.ResponseWriter, r *http.Request) {
			// extract the ID
			serverID := strings.Replace(r.URL.Path, "/servers/", "", 1)
			id32, err := strconv.ParseInt(serverID, 10, 32)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(400)
				w.Write([]byte(`{"error": "invalid server ID"}`))
				return
			}
			var (
				index int = -1
			)
			for i, s := range memServers {
				if s.ID == int(id32) {
					index = i
				}
			}

			switch r.Method {
			case "GET":
				if index >= 0 {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(200)
					b, _ := json.Marshal(memServers[index])
					w.Write(b)
					return
				}
				w.WriteHeader(404)
			case "DELETE":
				memServers = append(memServers[:index], memServers[index+1:]...)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(204)
				w.Write([]byte("{}"))
			}
		})
	}
	ngPool2 := newCherryNodeGroup(m, "pool2", 0, 10, 0, 30*time.Second, 2*time.Second)
	ngPool3 := newCherryNodeGroup(m, "pool3", 0, 10, 0, 30*time.Second, 2*time.Second)

	// calls: listServers
	n1Pool2, err := ngPool2.cherryManager.getNodeNames(ngPool2.id)
	assert.NoError(t, err)
	assert.Equal(t, int(0), len(n1Pool2))

	// calls: listServers
	n1Pool3, err := ngPool3.cherryManager.getNodeNames(ngPool3.id)
	assert.NoError(t, err)
	assert.Equal(t, int(1), len(n1Pool3))

	existingNodesPool2 := make(map[string]bool)
	existingNodesPool3 := make(map[string]bool)

	for _, node := range n1Pool2 {
		existingNodesPool2[node] = true
	}

	for _, node := range n1Pool3 {
		existingNodesPool3[node] = true
	}

	// Try to increase pool3 with negative size, this should return an error
	// calls: (should error before any calls)
	err = ngPool3.IncreaseSize(-1)
	assert.Error(t, err)

	// Now try to increase the pool3 size by 1, that should work
	// calls: listServers, createServer
	err = ngPool3.IncreaseSize(1)
	assert.NoError(t, err)

	if useRealEndpoint {
		// If testing with actual API give it some time until the nodes bootstrap
		time.Sleep(420 * time.Second)
	}

	// calls: listServers
	n2Pool3, err := ngPool3.cherryManager.getNodeNames(ngPool3.id)
	assert.NoError(t, err)
	// Assert that the nodepool3 size is now 2
	assert.Equal(t, int(2), len(n2Pool3))
	// calls: listServers
	n2Pool3providers, err := ngPool3.cherryManager.getNodes(ngPool3.id)
	assert.NoError(t, err)
	// Asset that provider ID lengths matches names length
	assert.Equal(t, len(n2Pool3providers), len(n2Pool3))

	// Now try to increase the pool2 size by 1, that should work
	// calls: listServers, createServer
	err = ngPool2.IncreaseSize(1)
	assert.NoError(t, err)

	if useRealEndpoint {
		// If testing with actual API give it some time until the nodes bootstrap
		time.Sleep(420 * time.Second)
	}

	// calls: listServers
	n2Pool2, err := ngPool2.cherryManager.getNodeNames(ngPool2.id)
	assert.NoError(t, err)
	// Assert that the nodepool2 size is now 1
	assert.Equal(t, int(1), len(n2Pool2))
	// calls: listServers
	n2Pool2providers, err := ngPool2.cherryManager.getNodes(ngPool2.id)
	assert.NoError(t, err)
	// Asset that provider ID lengths matches names length
	assert.Equal(t, len(n2Pool2providers), len(n2Pool2))

	// Let's try to delete the new nodes
	nodesPool2 := []*apiv1.Node{}
	nodesPool3 := []*apiv1.Node{}
	for i, node := range n2Pool2 {
		if _, ok := existingNodesPool2[node]; !ok {
			testNode := BuildTestNode(node, 1000, 1000)
			testNode.Spec.ProviderID = n2Pool2providers[i]
			nodesPool2 = append(nodesPool2, testNode)
		}
	}
	for i, node := range n2Pool3 {
		if _, ok := existingNodesPool3[node]; !ok {
			testNode := BuildTestNode(node, 1000, 1000)
			testNode.Spec.ProviderID = n2Pool3providers[i]
			nodesPool3 = append(nodesPool3, testNode)
		}
	}

	err = ngPool2.DeleteNodes(nodesPool2)
	assert.NoError(t, err)

	err = ngPool3.DeleteNodes(nodesPool3)
	assert.NoError(t, err)

	// Wait a few seconds if talking to the actual Cherry API
	if useRealEndpoint {
		time.Sleep(10 * time.Second)
	}

	// Make sure that there were no errors and the nodepool2 size is once again 0
	n3Pool2, err := ngPool2.cherryManager.getNodeNames(ngPool2.id)
	assert.NoError(t, err)
	assert.Equal(t, int(0), len(n3Pool2))

	// Make sure that there were no errors and the nodepool3 size is once again 1
	n3Pool3, err := ngPool3.cherryManager.getNodeNames(ngPool3.id)
	assert.NoError(t, err)
	assert.Equal(t, int(1), len(n3Pool3))
}
