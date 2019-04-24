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

package magnum

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum/gophercloud"
	th "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum/gophercloud/testhelper"
)

var clusterUUID = "732851e1-f792-4194-b966-4cbfa5f30093"
var clusterNodeCount = 5
var clusterStatus = clusterStatusUpdateComplete
var clusterStackID = "2e35472f-d3c1-40b1-93fe-a421db19cc89"

var clusterGetResponseSuccess = fmt.Sprintf(`
{
    "create_timeout":60,
    "links":[
        {
            "href":"http://10.100.101.102:9511/v1/clusters/732851e1-f792-4194-b966-4cbfa5f30093",
            "rel":"self"
        },
        {
            "href":"http://10.100.101.102:9511/clusters/732851e1-f792-4194-b966-4cbfa5f30093",
            "rel":"bookmark"
        }
    ],
    "labels":{
    },
    "updated_at":"2019-02-07T14:07:33+00:00",
    "keypair":"test-key",
    "master_flavor_id":"m2.medium",
    "user_id":"tester",
    "uuid":"%s",
    "api_address":"https://172.24.4.6:6443",
    "master_addresses":[
        "172.24.4.6"
    ],
    "node_count":%d,
    "project_id":"d656bd69-82a2-4efe-bbee-abd0b0a83252",
    "status":"%s",
    "docker_volume_size":null,
    "master_count":1,
    "node_addresses":[
        "172.24.4.10"
    ],
    "status_reason":"Stack UPDATE completed successfully",
    "coe_version":"v1.9.3",
    "cluster_template_id":"af067ca7-d0a2-43ee-9529-40e538072d86",
    "name":"cluster-01",
    "stack_id":"%s",
    "created_at":"2019-02-01T13:13:48+00:00",
    "discovery_url":"https://discovery.etcd.io/efa82529a12040a19deec2e7addd3183",
    "flavor_id":"m2.medium",
    "container_version":"1.12.6"
}`, clusterUUID, clusterNodeCount, clusterStatus, clusterStackID)

var badClusterUUID = "fa9887b8-e8d8-46d1-a82f-3c28618f8db1"

var clusterGetResponseFail = fmt.Sprintf(`
{
    "errors":[
        {
            "status":404,
            "code":"client",
            "links":[

            ],
            "title":"Cluster %s could not be found",
            "detail":"Cluster %s could not be found.",
            "request_id":""
        }
    ]
}`, badClusterUUID, badClusterUUID)

var stackName = "cluster-01-aftjnjwdczjr"
var stackID = "2e35472f-d3c1-40b1-93fe-a421db19cc89"
var stackStatus = "UPDATE_COMPLETE"

var badStackID = "708389ff-10fa-4008-9e0f-a3d5f979e88c"

var stackGetResponse = `
{
    "stack":{
        "parent":null,
        "disable_rollback":true,
        "description":"This template will boot a Kubernetes cluster with one or more minions.\n",
        "parameters":{
            "OS::stack_id":"2e35472f-d3c1-40b1-93fe-a421db19cc89",
            "OS::project_id":"d656bd69-82a2-4efe-bbee-abd0b0a83252",
            "OS::stack_name":"cluster-01-aftjnjwdczjr"
        },
        "deletion_time":null,
        "stack_name":"%s",
        "stack_user_project_id":"0aa1d684229d4b5985b090e5b2c651d6",
        "stack_status_reason":"Stack UPDATE completed successfully",
        "creation_time":"2019-02-01T13:13:55Z",
        "links":[
            {
                "href":"https://10.100.101.102:8004/v1/d656bd69-82a2-4efe-bbee-abd0b0a83252/stacks/cluster-01-aftjnjwdczjr/2e35472f-d3c1-40b1-93fe-a421db19cc89",
                "rel":"self"
            }
        ],
        "capabilities":[

        ],
        "notification_topics":[

        ],
        "tags":null,
        "timeout_mins":60,
        "stack_status":"%s",
        "stack_owner":null,
        "updated_time":"2019-02-07T14:06:29Z",
        "id":"%s",
        "outputs":[
        ]
    }
}`

var stackGetResponseSuccess = fmt.Sprintf(stackGetResponse, stackName, stackStatus, stackID)

var stackGetResponseNotFound = fmt.Sprintf(`
{
    "explanation":"The resource could not be found.",
    "code":404,
    "error":{
        "message":"The Stack (%s) could not be found.",
        "traceback":null,
        "type":"EntityNotFound"
    },
    "title":"Not Found"
}`, badStackID)

var deleteNodeRefs = []NodeRef{
	{Name: stackName + "-minion-1", MachineID: "3ae2e15807bd48ccbb26f9bb5f2996d6", ProviderID: "openstack:///3ae2e158-07bd-48cc-bb26-f9bb5f2996d6", IPs: []string{"10.0.0.52"}},
}
var patchResponseSuccess = `{"uuid": "732851e1-f792-4194-b966-4cbfa5f30093"}`

var stackUpdateResponseFail = fmt.Sprintf(`
{
    "errors":[
        {
            "status":404,
            "code":"client",
            "links":[

            ],
            "title":"Cluster %s could not be found",
            "detail":"Cluster %s could not be found.",
            "request_id":""
        }
    ]
}`, badClusterUUID, badClusterUUID)

var stackBadPatchResponse = `
{
    "explanation":"The server could not comply with the request since it is either malformed or otherwise incorrect.",
    "code":400,
    "error":{
        "message":"Property error: : resources.kube_minions.properties.count: : -1 is out of range (min: 0, max: None)",
        "traceback":null,
        "type":"StackValidationFailed"
    },
    "title":"Bad Request"
}`

var kubeMinionsStackID = "be32904e-021b-4069-a670-67bbf4650dca"
var kubeMinionsStackName = fmt.Sprintf("%s-kube_minions-aidbol4lxwlf", stackName)

var stackresourceGetKubeMinionsResponse = fmt.Sprintf(`
{
    "resource":{
        "resource_name":"kube_minions",
        "description":"",
        "logical_resource_id":"kube_minions",
        "creation_time":"2019-02-21T10:59:33Z",
        "resource_status_reason":"state changed",
        "updated_time":"2019-02-22T15:33:30Z",
        "required_by":[

        ],
        "resource_status":"UPDATE_COMPLETE",
        "physical_resource_id":"%s",
        "attributes":{
            "attributes":null,
            "refs":null,
            "refs_map":null,
            "removed_rsrc_list":[

            ]
        },
        "resource_type":"OS::Heat::ResourceGroup"
    }
}`, kubeMinionsStackID)

var stackresourceGetKubeMinionsNotFound = fmt.Sprintf(`
{
    "explanation":"The resource could not be found.",
    "code":404,
    "error":{
        "message":"The Stack (%s) could not be found.",
        "traceback":null,
        "type":"EntityNotFound"
    },
    "title":"Not Found"
}`, badStackID)

var stackGetKubeMinionsStackResponse = fmt.Sprintf(`
{
    "stack":{
        "parent":"afa343f6-17f5-4cb6-82ea-779a7142f014",
        "disable_rollback":true,
        "description":"No description",
        "parameters":{
            "OS::project_id":"d656bd69-82a2-4efe-bbee-abd0b0a83252",
            "OS::stack_id":"%s",
            "OS::stack_name":"%s"
        },
        "deletion_time":null,
        "stack_name":"%s",
        "stack_user_project_id":"bc47c8c25c084b22914e0491bd9c5fab",
        "stack_status_reason":"Stack UPDATE completed successfully",
        "creation_time":"2019-02-21T11:06:28Z",
        "links":[

        ],
        "capabilities":[

        ],
        "notification_topics":[

        ],
        "tags":null,
        "timeout_mins":60,
        "stack_status":"UPDATE_COMPLETE",
        "stack_owner":null,
        "updated_time":"2019-02-22T15:33:30Z",
        "id":"be32904e-021b-4069-a670-67bbf4650dca",
        "outputs":[
            {
                "output_value":{
                    "0":"ffbc651d-a661-462d-9435-2e01f3020688",
                    "1":"3ae2e158-07bd-48cc-bb26-f9bb5f2996d6"
                },
                "output_key":"refs_map",
                "description":"No description given"
            }
        ],
        "template_description":"No description"
    }
}`, kubeMinionsStackID, kubeMinionsStackName, kubeMinionsStackName)

func createTestServiceClient() *gophercloud.ServiceClient {
	return &gophercloud.ServiceClient{
		ProviderClient: &gophercloud.ProviderClient{TokenID: "cbc36478b0bd8e67e89469c7749d4127"},
		Endpoint:       th.Endpoint() + "v1/",
	}
}

func createTestMagnumManagerHeat(client *gophercloud.ServiceClient) *magnumManagerHeat {
	return &magnumManagerHeat{
		clusterClient: client,
		heatClient:    client,
		waitTimeStep:  100 * time.Millisecond,
	}
}

func createManagerGetClusterSuccess() *magnumManagerHeat {
	th.Mux.HandleFunc("/v1/clusters/"+clusterUUID, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, clusterGetResponseSuccess)
	})

	sc := createTestServiceClient()

	manager := createTestMagnumManagerHeat(sc)
	manager.clusterName = clusterUUID

	return manager
}

func createManagerGetClusterFail() *magnumManagerHeat {
	th.Mux.HandleFunc("/v1/clusters/"+clusterUUID, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)

		fmt.Fprint(w, clusterGetResponseFail)
	})

	sc := createTestServiceClient()

	manager := createTestMagnumManagerHeat(sc)
	manager.clusterName = badClusterUUID

	return manager
}

func TestNodeGroupSizeSuccess(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	manager := createManagerGetClusterSuccess()

	nodeCount, err := manager.nodeGroupSize("default")
	assert.NoError(t, err)
	assert.Equal(t, clusterNodeCount, nodeCount)
}

func TestNodeGroupSizeFail(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	manager := createManagerGetClusterFail()

	_, err := manager.nodeGroupSize("default")
	assert.Error(t, err)
	assert.Equal(t, "could not get cluster: Resource not found", err.Error())
}

func TestGetClusterStatusSuccess(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	manager := createManagerGetClusterSuccess()

	gotStatus, err := manager.getClusterStatus()
	assert.NoError(t, err)
	assert.Equal(t, clusterStatus, gotStatus)
}

func TestGetClusterStatusFail(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	manager := createManagerGetClusterFail()

	_, err := manager.getClusterStatus()
	assert.Error(t, err)
	assert.Equal(t, "could not get cluster: Resource not found", err.Error())
}

func TestCanUpdateSuccess(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	manager := createManagerGetClusterSuccess()

	can, status, err := manager.canUpdate()
	assert.NoError(t, err)
	assert.Equal(t, true, can)
	assert.Equal(t, clusterStatus, status)
}

func TestCanUpdateFail(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	manager := createManagerGetClusterFail()

	_, _, err := manager.canUpdate()
	assert.Error(t, err)
	assert.Equal(t, "could not get cluster status: could not get cluster: Resource not found", err.Error())
}

func TestGetStackNameSuccess(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	th.Mux.HandleFunc("/v1/stacks/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprintf(w, stackGetResponse, stackName, stackStatusUpdateComplete, stackID)
	})

	sc := createTestServiceClient()

	manager := createTestMagnumManagerHeat(sc)
	manager.clusterName = clusterUUID
	manager.stackID = stackID

	gotStackName, err := manager.getStackName(manager.stackID)
	assert.NoError(t, err)
	assert.Equal(t, stackName, gotStackName)
}

func TestGetStackNameNotFound(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	th.Mux.HandleFunc("/v1/stacks/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)

		fmt.Fprint(w, stackGetResponseNotFound)
	})

	sc := createTestServiceClient()

	manager := createTestMagnumManagerHeat(sc)
	manager.clusterName = clusterUUID
	manager.stackID = badStackID

	_, err := manager.getStackName(manager.stackID)
	assert.Error(t, err)
	assert.Equal(t, fmt.Sprintf("could not find stack with ID %s: Resource not found", badStackID), err.Error())
}

func TestGetStackStatusSuccess(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	th.Mux.HandleFunc("/v1/stacks/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, stackGetResponseSuccess)
	})

	sc := createTestServiceClient()

	manager := createTestMagnumManagerHeat(sc)
	manager.clusterName = clusterUUID
	manager.stackID = stackID
	manager.stackName = stackName

	status, err := manager.getStackStatus()
	assert.NoError(t, err)
	assert.Equal(t, stackStatus, status)
}

func TestGetStackStatusFail(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	th.Mux.HandleFunc("/v1/stacks/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)

		fmt.Fprint(w, stackGetResponseNotFound)
	})

	sc := createTestServiceClient()

	manager := createTestMagnumManagerHeat(sc)
	manager.clusterName = clusterUUID
	manager.stackID = badStackID
	manager.stackName = stackName

	_, err := manager.getStackStatus()
	assert.Error(t, err)
	assert.Equal(t, "could not get stack from heat: Resource not found", err.Error())
}

func TestUpdateNodeCountSuccess(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	th.Mux.HandleFunc("/v1/clusters/"+clusterUUID, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)

		fmt.Fprint(w, patchResponseSuccess)
	})

	sc := createTestServiceClient()

	manager := createTestMagnumManagerHeat(sc)
	manager.clusterName = clusterUUID

	err := manager.updateNodeCount("default", 2)
	assert.NoError(t, err)
}

func TestUpdateNodeCountFail(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	th.Mux.HandleFunc("/v1/clusters/"+badClusterUUID, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)

		fmt.Fprint(w, stackUpdateResponseFail)
	})

	sc := createTestServiceClient()

	manager := createTestMagnumManagerHeat(sc)
	manager.clusterName = badClusterUUID

	err := manager.updateNodeCount("default", 2)
	assert.Error(t, err)
	assert.Equal(t, "could not update cluster: Resource not found", err.Error())
}

func TestDeleteNodesSuccess(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	GETRequestCount := 0

	// Handle stack PATCH, then two stack GETS (for checking UPDATE_IN_PROGRESS -> UPDATE_COMPLETE)
	th.Mux.HandleFunc("/v1/stacks/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPatch {
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			return
		}
		if r.Method == http.MethodGet {
			GETRequestCount += 1
			if GETRequestCount == 1 {
				// findStackIndices
				w.Header().Add("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, stackGetKubeMinionsStackResponse)
			} else if GETRequestCount == 2 {
				w.Header().Add("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, stackGetResponse, stackName, stackStatusUpdateInProgress, stackID)
			} else if GETRequestCount == 3 {
				w.Header().Add("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, stackGetResponse, stackName, stackStatusUpdateComplete, stackID)
			}
		}
	})

	// Handle cluster node_count PATCH
	th.Mux.HandleFunc("/v1/clusters/"+clusterUUID, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprint(w, patchResponseSuccess)
	})

	sc := createTestServiceClient()

	manager := createTestMagnumManagerHeat(sc)
	manager.clusterName = clusterUUID
	manager.stackID = stackID
	manager.stackName = stackName

	err := manager.deleteNodes("default", deleteNodeRefs, 1)
	assert.NoError(t, err)
}

func TestDeleteNodesStackPatchFail(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	GETRequestCount := 0

	// Handle bad stack PATCH, then two stack GETS (for checking UPDATE_IN_PROGRESS -> UPDATE_COMPLETE)
	th.Mux.HandleFunc("/v1/stacks/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPatch {
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, stackBadPatchResponse)
			return
		}
		if r.Method == http.MethodGet {
			GETRequestCount += 1
			if GETRequestCount == 1 {
				// findStackIndices
				w.Header().Add("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, stackGetKubeMinionsStackResponse)
			} else if GETRequestCount == 2 {
				w.Header().Add("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, stackGetResponse, stackName, stackStatusUpdateInProgress, stackID)
			} else if GETRequestCount == 3 {
				w.Header().Add("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, stackGetResponse, stackName, stackStatusUpdateComplete, stackID)
			}
		}
	})

	// Handle cluster node_count PATCH
	th.Mux.HandleFunc("/v1/clusters/"+clusterUUID, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
	})

	sc := createTestServiceClient()

	manager := createTestMagnumManagerHeat(sc)
	manager.clusterName = clusterUUID
	manager.stackID = stackID
	manager.stackName = stackName

	err := manager.deleteNodes("default", deleteNodeRefs, -1)
	assert.Error(t, err)

	expectedErrString := fmt.Sprintf("stack patch failed: Bad request with: [PATCH %s/v1/stacks/%s/%s], error message: %s",
		th.Server.URL, stackName, stackID, stackBadPatchResponse)

	assert.Equal(t, expectedErrString, err.Error())
}

func TestFindStackIndicesMissing(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	th.Mux.HandleFunc("/v1/stacks/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, stackGetKubeMinionsStackResponse)
	})

	sc := createTestServiceClient()

	manager := createTestMagnumManagerHeat(sc)
	manager.clusterName = clusterUUID
	manager.stackID = stackID
	manager.stackName = stackName

	nodes := []NodeRef{
		{Name: fmt.Sprintf("%s-minion-0", stackName), MachineID: "ffbc651da661462d94352e01f3020688"},
		{Name: fmt.Sprintf("%s-minion-1", stackName)},
	}

	_, err := manager.findStackIndices(nodes)
	assert.Error(t, err)
	assert.Equal(t, "1 nodes could not be resolved to stack indices", err.Error())
}

func TestWaitForStackStatusSuccess(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	GETRequestCount := 0

	th.Mux.HandleFunc("/v1/stacks/", func(w http.ResponseWriter, r *http.Request) {
		if GETRequestCount == 0 {
			GETRequestCount += 1
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, stackGetResponse, stackName, stackStatusUpdateInProgress, stackID)
		} else {
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, stackGetResponseSuccess)
		}
	})

	sc := createTestServiceClient()

	manager := createTestMagnumManagerHeat(sc)
	manager.clusterName = clusterUUID
	manager.stackID = stackID
	manager.stackName = stackName

	err := manager.waitForStackStatus(stackStatusUpdateComplete, 200*time.Millisecond)
	assert.NoError(t, err)
}

func TestWaitForStackStatusTimeout(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	th.Mux.HandleFunc("/v1/stacks/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, stackGetResponse, stackName, stackStatusUpdateInProgress, stackID)
	})

	sc := createTestServiceClient()

	manager := createTestMagnumManagerHeat(sc)
	manager.clusterName = clusterUUID
	manager.stackID = stackID
	manager.stackName = stackName

	err := manager.waitForStackStatus(stackStatusUpdateComplete, 200*time.Millisecond)
	assert.Error(t, err)
	assert.Equal(t, "timeout (200ms) waiting for stack status UPDATE_COMPLETE", err.Error())
}

func TestWaitForStackStatusError(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	th.Mux.HandleFunc("/v1/stacks/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, stackGetResponseNotFound)
	})

	sc := createTestServiceClient()

	manager := createTestMagnumManagerHeat(sc)
	manager.clusterName = clusterUUID
	manager.stackID = stackID
	manager.stackName = stackName

	err := manager.waitForStackStatus(stackStatusUpdateComplete, 200*time.Millisecond)
	assert.Error(t, err)
	assert.Equal(t, "error waiting for stack status: could not get stack from heat: Resource not found", err.Error())
}

func TestGetKubeMinionsStackSuccess(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	th.Mux.HandleFunc("/v1/stacks/"+stackName+"/"+stackID+"/resources/kube_minions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, stackresourceGetKubeMinionsResponse)
	})

	th.Mux.HandleFunc("/v1/stacks/"+kubeMinionsStackID, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, stackGetKubeMinionsStackResponse)
	})

	sc := createTestServiceClient()

	manager := createTestMagnumManagerHeat(sc)
	manager.clusterName = clusterUUID
	manager.stackID = stackID
	manager.stackName = stackName

	name, ID, err := manager.getKubeMinionsStack(manager.stackName, manager.stackID)
	assert.NoError(t, err)
	assert.Equal(t, kubeMinionsStackName, name)
	assert.Equal(t, kubeMinionsStackID, ID)
}

func TestGetKubeMinionsStackNotFound(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	th.Mux.HandleFunc("/v1/stacks/"+stackName+"/"+badStackID+"/resources/kube_minions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, stackresourceGetKubeMinionsNotFound)
	})

	sc := createTestServiceClient()

	manager := createTestMagnumManagerHeat(sc)
	manager.clusterName = clusterUUID
	manager.stackID = badStackID
	manager.stackName = stackName

	_, _, err := manager.getKubeMinionsStack(manager.stackName, manager.stackID)
	assert.Error(t, err)
	assert.Equal(t, "could not get kube_minions stack resource: Resource not found", err.Error())

}

func TestStackIndexFromID(t *testing.T) {
	minion0IP := []string{"10.0.0.52"}
	minion0ID := "a390ca6ab24846af8ddd854a52fd5281"
	minion0 := NodeRef{MachineID: minion0ID, IPs: minion0IP}
	minion0StackID := "a390ca6a-b248-46af-8ddd-854a52fd5281"

	minion1IP := []string{"10.0.0.59"}
	minion1ID := "5bc76eeb2b2e4625a1a559560aa90e5a"
	minion1 := NodeRef{MachineID: minion1ID, IPs: minion1IP}
	minion1StackID := "5bc76eeb-2b2e-4625-a1a5-59560aa90e5a"

	minion2IP := []string{"10.0.0.70"}
	minion2ID := "86db2aa4666948cfa2b828eb93cf3fd1"
	minion2 := NodeRef{MachineID: minion2ID, IPs: minion2IP}

	// Empty nodeRef, covers case with MachineID "" and IPs []string{}
	minion3 := NodeRef{}

	mappingWithIPs := map[string]string{
		minion0IP[0]: "0",
		minion1IP[0]: "1",
	}

	mappingWithIDs := map[string]string{
		minion0StackID: "0",
		minion1StackID: "1",
	}

	emptyMapping := map[string]string{}

	tests := []struct {
		name          string
		mapping       map[string]string
		nodeRef       NodeRef
		expectedIndex string
		expectedFound bool
	}{
		{"UUIDs 0", mappingWithIDs, minion0, "0", true},
		{"UUIDs 1", mappingWithIDs, minion1, "1", true},
		{"UUIDs 2", mappingWithIDs, minion2, "", false},

		{"IPs 0", mappingWithIPs, minion0, "0", true},
		{"IPs 1", mappingWithIPs, minion1, "1", true},
		{"IPs 2", mappingWithIPs, minion2, "", false},

		{"empty 0", emptyMapping, minion0, "", false},
		{"empty 1", emptyMapping, minion1, "", false},
		{"empty 2", emptyMapping, minion2, "", false},

		{"empty ref", mappingWithIDs, minion3, "", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			index, found := stackIndexFromID(test.mapping, test.nodeRef)
			assert.Equal(t, test.expectedIndex, index, "failed to find index with mapping %+v and NodeRef %+v", test.mapping, test.nodeRef)
			assert.Equal(t, test.expectedFound, found, "failed to find / not find with mapping %+v and NodeRef %+v", test.mapping, test.nodeRef)
		})
	}
}
