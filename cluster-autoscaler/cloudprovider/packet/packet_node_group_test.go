/*
Copyright 2019 The Kubernetes Authors.

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

package packet

import (
	"os"
	"sync"
	"testing"
	"time"

	apiv1 "k8s.io/api/core/v1"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestIncreaseDecreaseSize(t *testing.T) {
	var m *packetManagerRest
	server := NewHttpServerMock()
	defer server.Close()
	assert.Equal(t, true, true)
	if len(os.Getenv("PACKET_AUTH_TOKEN")) > 0 {
		// If auth token set in env, hit the actual Packet API
		m = newTestPacketManagerRest(t, "https://api.packet.net")
	} else {
		// Set up a mock Packet API
		m = newTestPacketManagerRest(t, server.URL)
		server.On("handle", "/projects/"+m.projectID+"/devices").Return(listPacketDevicesResponse).Times(4)
		server.On("handle", "/projects/"+m.projectID+"/devices").Return(listPacketDevicesResponseAfterCreate).Times(2)
		server.On("handle", "/projects/"+m.projectID+"/devices").Return(listPacketDevicesResponse)
	}
	clusterUpdateLock := sync.Mutex{}
	ng := &packetNodeGroup{
		packetManager:       m,
		id:                  "pool1",
		clusterUpdateMutex:  &clusterUpdateLock,
		minSize:             1,
		maxSize:             10,
		targetSize:          new(int),
		waitTimeStep:        30 * time.Second,
		deleteBatchingDelay: 2 * time.Second,
	}

	n1, err := ng.Nodes()
	assert.NoError(t, err)
	assert.Equal(t, int(1), len(n1))

	// Try to increase pool with negative size, this should return an error
	err = ng.IncreaseSize(-1)
	assert.Error(t, err)

	// Now try to increase the pool size by 2, that should work
	err = ng.IncreaseSize(2)
	assert.NoError(t, err)

	if len(os.Getenv("PACKET_AUTH_TOKEN")) > 0 {
		// If testing with actual API give it some time until the nodes bootstrap
		time.Sleep(420 * time.Second)
	}
	n2, err := ng.packetManager.getNodeNames(ng.id)
	assert.NoError(t, err)
	// Assert that the nodepool size is now 3
	assert.Equal(t, int(3), len(n2))

	// Let's try to delete the new nodes
	nodes := []*apiv1.Node{}
	for _, node := range n2 {
		if node != "k8s-worker-1" {
			nodes = append(nodes, BuildTestNode(node, 1000, 1000))
		}
	}
	err = ng.DeleteNodes(nodes)
	assert.NoError(t, err)

	// Wait a few seconds if talking to the actual Packet API
	if len(os.Getenv("PACKET_AUTH_TOKEN")) > 0 {
		time.Sleep(10 * time.Second)
	}

	// Make sure that there were no errors and the nodepool size is once again 1
	n3, err := ng.Nodes()
	assert.NoError(t, err)
	assert.Equal(t, int(1), len(n3))
	mock.AssertExpectationsForObjects(t, server)
}
