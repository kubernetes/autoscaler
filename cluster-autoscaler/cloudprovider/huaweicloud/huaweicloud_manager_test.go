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

package huaweicloud

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huawei-cloud-sdk-go"
	th "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huawei-cloud-sdk-go/testhelper"
	"net/http"
	"testing"
)

const (
	clusterUUID = "25b884ab-623f-11ea-9981-0255ac101546"
)

var (
	isTestingDeleteNodes = false
	getRequestCount      = 0
	getClusterResponse   = `
	{
		"kind": "Cluster",
		"apiVersion": "v3",
		"metadata": {
			"name": "crc-test",
			"uid": "%s",
			"creationTimestamp": "2020-03-09 19:49:54.353283 +0000 UTC",
			"updateTimestamp": "2020-03-23 18:33:42.788547 +0000 UTC"
		},
		"status": {
			"phase": "%s"
		}
	}`

	getClusterResponseSuccess = fmt.Sprintf(getClusterResponse, clusterUUID, clusterStatusAvailable)

	getNodePoolsResponse = `
	{
		"kind": "List",
		"apiVersion": "v3",
		"items": [
			{
				"kind": "NodePool",
				"apiVersion": "v3",
				"metadata": {
					"name": "%s",
					"uid": "deb25e14-6558-11ea-97c2-0255ac101d4a"
				},
				"spec": {
					"initialNodeCount": 1,
					"autoscaling": {
						"enable": true,
						"maxNodeCount": 50
					}
				},
				"status": {
					"currentNode": %d,
					"phase": ""
				}
			},
			{
				"kind": "NodePool",
				"apiVersion": "v3",
				"metadata": {
					"name": "DefaultPool",
					"uid": "DefaultPool"
				},
				"spec": {
					"initialNodeCount": 1,
					"autoscaling": {}
				},
				"status": {
					"currentNode": 1,
					"phase": ""
				}
			}
		]
	}`

	getNodePoolsResponseSuccess = fmt.Sprintf(getNodePoolsResponse, nodePoolName, nodePoolNodeCount)

	getNodePoolResponseSuccess = fmt.Sprintf(`
	{
		"kind": "NodePool",
		"apiVersion": "v3",
		"metadata": {
			"name": "%s",
			"uid": "%s"
		},
		"spec": {
			"initialNodeCount": 1,
			"autoscaling": {
				"enable": true,
				"maxNodeCount": 50
			},
			"nodeManagement": {}
		},
		"status": {
			"currentNode": %d,
			"phase": ""
		}
	}`, nodePoolName, nodePoolUID, nodePoolNodeCount)

	deleteNodesResponseSuccess = `
	{
		"kind": "Node",
		"apiVersion": "v3",
		"metadata": {
			"name": "crc-nodepool-tc3ei",
			"uid": "c1b6ff0c-6ee1-11ea-befa-0255ac101d4c"
		},
		"status": {
			"phase": "Active",
			"jobID": "f5fb1835-6ee1-11ea-befa-0255ac101d4c",
			"serverId": "4a2a7527-7d0d-4bc5-ba1b-379efed8da3c",
			"privateIP": "192.168.0.15"
		}
	}`
)

// create fake service client
func createTestServiceClient() *huaweicloudsdk.ServiceClient {
	return &huaweicloudsdk.ServiceClient{
		ProviderClient: &huaweicloudsdk.ProviderClient{},
		Endpoint:       th.GetEndpoint() + "/huaweitest/",
	}
}

// create fake manager
func createTestHuaweicloudManager() *huaweicloudCloudManager {
	sc := createTestServiceClient()
	return &huaweicloudCloudManager{
		clusterClient: sc,
		clusterName:   clusterUUID,
		timeIncrement: waitTimeStep,
	}
}

// create routers
func register() {
	// create router for getting cluster's status
	th.Mux.HandleFunc("/huaweitest/clusters/"+clusterUUID, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if !isTestingDeleteNodes {
			_, err := fmt.Fprint(w, getClusterResponseSuccess)
			if err != nil {
				fmt.Println("error handling request of getting cluster status available:", err)
			}
		} else {
			getRequestCount += 1
			if getRequestCount == 1 {
				_, err := fmt.Fprintf(w, getClusterResponse, clusterUUID, clusterStatusScalingDown)
				if err != nil {
					fmt.Println("error handling request of getting cluster status scaling down when deleting nodes:", err)
				}
			} else {
				_, err := fmt.Fprintf(w, getClusterResponseSuccess)
				if err != nil {
					fmt.Println("error handling request of getting cluster status available when deleting nodes:", err)
				}
			}
		}

	})

	// create router for getting all nodepools' status associated with a cluster
	th.Mux.HandleFunc("/huaweitest/clusters/"+clusterUUID+"/nodepools", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if !isTestingDeleteNodes {
			_, err := fmt.Fprintf(w, getNodePoolsResponseSuccess)
			if err != nil {
				fmt.Println("error handling request of getting nodepools' status:", err)
			}
		} else {
			_, err := fmt.Fprintf(w, getNodePoolsResponse, nodePoolName, nodePoolNodeCount+decreaseSize)
			if err != nil {
				fmt.Println("error handling request of getting nodepools' status after deleting node:", err)
			}
		}
	})

	// create router for getting status of a certain nodepool
	th.Mux.HandleFunc("/huaweitest/clusters/"+clusterUUID+"/nodepools/"+nodePoolUID, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		_, err := fmt.Fprintf(w, getNodePoolResponseSuccess)
		if err != nil {
			fmt.Println("error handling request of getting/updating a nodepool's status:", err)
		}
	})

	// create router for deleting a node
	th.Mux.HandleFunc("/huaweitest/clusters/"+clusterUUID+"/nodes/"+nodeIdToBeDeleted, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		_, err := fmt.Fprintf(w, deleteNodesResponseSuccess)
		if err != nil {
			fmt.Println("error handling request of deleting a node:", err)
		}
	})
}

func Test_nodeGroupSize(t *testing.T) {
	th.CreateServer()
	defer th.ShutDownServer()
	register()

	manager := createTestHuaweicloudManager()

	nodeCount, err := manager.nodeGroupSize(nodePoolName)
	assert.NoError(t, err)
	assert.Equal(t, nodePoolNodeCount, nodeCount)
}

func Test_updateNodeCount(t *testing.T) {
	th.CreateServer()
	defer th.ShutDownServer()
	register()

	manager := createTestHuaweicloudManager()

	ng := createTestNodeGroup(manager)
	err := manager.updateNodeCount(ng, nodePoolNodeCount)
	assert.NoError(t, err)
}

func Test_deleteNodesHelper(t *testing.T) {
	th.CreateServer()
	defer th.ShutDownServer()
	register()

	manager := createTestHuaweicloudManager()
	nodeIds := []string{nodeIdToBeDeleted}

	err := manager.deleteNodesHelper(manager.clusterClient, clusterUUID, nodeIds)
	assert.NoError(t, err)
}

func Test_deleteNodes(t *testing.T) {
	th.CreateServer()
	defer th.ShutDownServer()
	register()

	manager := createTestHuaweicloudManager()
	ng := createTestNodeGroup(manager)
	nodeIds := []string{nodeIdToBeDeleted}
	isTestingDeleteNodes = true
	getRequestCount = 0

	err := manager.deleteNodes(ng, nodeIds, nodePoolNodeCount)
	assert.NoError(t, err)

	isTestingDeleteNodes = false
	getRequestCount = 0
}

func Test_getClusterStatus(t *testing.T) {
	th.CreateServer()
	defer th.ShutDownServer()
	register()

	manager := createTestHuaweicloudManager()

	status, _ := manager.getClusterStatus()
	assert.Equal(t, clusterStatusAvailable, status)
}

func Test_waitForClusterStatus(t *testing.T) {
	th.CreateServer()
	defer th.ShutDownServer()
	register()

	manager := createTestHuaweicloudManager()

	err := manager.waitForClusterStatus(clusterStatusAvailable, waitForCompleteStatusTimout)
	assert.NoError(t, err)
}

func Test_canUpdate(t *testing.T) {
	th.CreateServer()
	defer th.ShutDownServer()
	register()

	manager := createTestHuaweicloudManager()

	canUpdate, status, _ := manager.canUpdate()
	assert.Equal(t, canUpdate, true)
	assert.Equal(t, status, clusterStatusAvailable)
}
