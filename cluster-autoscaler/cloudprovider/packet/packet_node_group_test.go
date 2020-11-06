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

const createPacketDeviceResponsePool2 = ``
const deletePacketDeviceResponsePool2 = ``

const createPacketDeviceResponsePool3 = ``
const deletePacketDeviceResponsePool3 = ``

func TestIncreaseDecreaseSize(t *testing.T) {
	var m *packetManagerRest
	server := NewHttpServerMockWithContentType()
	defer server.Close()
	assert.Equal(t, true, true)
	if len(os.Getenv("PACKET_AUTH_TOKEN")) > 0 {
		// If auth token set in env, hit the actual Packet API
		m = newTestPacketManagerRest(t, "https://api.packet.net")
	} else {
		// Set up a mock Packet API
		m = newTestPacketManagerRest(t, server.URL)
		server.On("handleWithContentType", "/projects/"+m.packetManagerNodePools["default"].projectID+"/devices").Return("application/json", listPacketDevicesResponse).Times(3)
		server.On("handleWithContentType", "/projects/"+m.packetManagerNodePools["default"].projectID+"/devices").Return("application/json", createPacketDeviceResponsePool3).Times(1)
		server.On("handleWithContentType", "/projects/"+m.packetManagerNodePools["default"].projectID+"/devices").Return("application/json", listPacketDevicesResponseAfterIncreasePool3).Times(2)
		server.On("handleWithContentType", "/projects/"+m.packetManagerNodePools["default"].projectID+"/devices").Return("application/json", createPacketDeviceResponsePool2).Times(1)
		server.On("handleWithContentType", "/projects/"+m.packetManagerNodePools["default"].projectID+"/devices").Return("application/json", listPacketDevicesResponseAfterIncreasePool2).Times(3)
		server.On("handleWithContentType", "/devices/0f5609af-1c27-451b-8edd-a1283f2c9440").Return("application/json", deletePacketDeviceResponsePool2).Times(1)
		server.On("handleWithContentType", "/projects/"+m.packetManagerNodePools["default"].projectID+"/devices").Return("application/json", listPacketDevicesResponseAfterIncreasePool3).Times(3)
		server.On("handleWithContentType", "/devices/8fa90049-e715-4794-ba31-81c1c78cee84").Return("application/json", deletePacketDeviceResponsePool3).Times(1)
		server.On("handleWithContentType", "/projects/"+m.packetManagerNodePools["default"].projectID+"/devices").Return("application/json", listPacketDevicesResponse).Times(3)
	}
	clusterUpdateLock := sync.Mutex{}
	ngPool2 := &packetNodeGroup{
		packetManager:       m,
		id:                  "pool2",
		clusterUpdateMutex:  &clusterUpdateLock,
		minSize:             0,
		maxSize:             10,
		targetSize:          new(int),
		waitTimeStep:        30 * time.Second,
		deleteBatchingDelay: 2 * time.Second,
	}

	ngPool3 := &packetNodeGroup{
		packetManager:       m,
		id:                  "pool3",
		clusterUpdateMutex:  &clusterUpdateLock,
		minSize:             0,
		maxSize:             10,
		targetSize:          new(int),
		waitTimeStep:        30 * time.Second,
		deleteBatchingDelay: 2 * time.Second,
	}

	n1Pool2, err := ngPool2.packetManager.getNodeNames(ngPool2.id)
	assert.NoError(t, err)
	assert.Equal(t, int(0), len(n1Pool2))

	n1Pool3, err := ngPool3.packetManager.getNodeNames(ngPool3.id)
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
	err = ngPool3.IncreaseSize(-1)
	assert.Error(t, err)

	// Now try to increase the pool3 size by 1, that should work
	err = ngPool3.IncreaseSize(1)
	assert.NoError(t, err)

	if len(os.Getenv("PACKET_AUTH_TOKEN")) > 0 {
		// If testing with actual API give it some time until the nodes bootstrap
		time.Sleep(420 * time.Second)
	}

	n2Pool3, err := ngPool3.packetManager.getNodeNames(ngPool3.id)
	assert.NoError(t, err)
	// Assert that the nodepool3 size is now 2
	assert.Equal(t, int(2), len(n2Pool3))

	// Now try to increase the pool2 size by 1, that should work
	err = ngPool2.IncreaseSize(1)
	assert.NoError(t, err)

	if len(os.Getenv("PACKET_AUTH_TOKEN")) > 0 {
		// If testing with actual API give it some time until the nodes bootstrap
		time.Sleep(420 * time.Second)
	}

	n2Pool2, err := ngPool2.packetManager.getNodeNames(ngPool2.id)
	assert.NoError(t, err)
	// Assert that the nodepool2 size is now 1
	assert.Equal(t, int(1), len(n2Pool2))

	// Let's try to delete the new nodes
	nodesPool2 := []*apiv1.Node{}
	nodesPool3 := []*apiv1.Node{}
	for _, node := range n2Pool2 {
		if _, ok := existingNodesPool2[node]; !ok {
			nodesPool2 = append(nodesPool2, BuildTestNode(node, 1000, 1000))
		}
	}
	for _, node := range n2Pool3 {
		if _, ok := existingNodesPool3[node]; !ok {
			nodesPool3 = append(nodesPool3, BuildTestNode(node, 1000, 1000))
		}
	}

	err = ngPool2.DeleteNodes(nodesPool2)
	assert.NoError(t, err)

	err = ngPool3.DeleteNodes(nodesPool3)
	assert.NoError(t, err)

	// Wait a few seconds if talking to the actual Packet API
	if len(os.Getenv("PACKET_AUTH_TOKEN")) > 0 {
		time.Sleep(10 * time.Second)
	}

	// Make sure that there were no errors and the nodepool2 size is once again 0
	n3Pool2, err := ngPool2.packetManager.getNodeNames(ngPool2.id)
	assert.NoError(t, err)
	assert.Equal(t, int(0), len(n3Pool2))

	// Make sure that there were no errors and the nodepool3 size is once again 1
	n3Pool3, err := ngPool3.packetManager.getNodeNames(ngPool3.id)
	assert.NoError(t, err)
	assert.Equal(t, int(1), len(n3Pool3))
	mock.AssertExpectationsForObjects(t, server)
}
