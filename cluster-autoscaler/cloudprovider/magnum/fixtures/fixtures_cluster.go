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

package fixtures

import (
	"encoding/json"
	"fmt"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum/gophercloud/openstack/containerinfra/v1/nodegroups"
)

// Cluster UUIDs.
const (
	ProjectUUID = "57bf1ed0-52b8-4236-b1a7-dc58eceb8ff2"
	ClusterUUID = "91444514-b8db-4314-849c-650df15e8e83"

	DefaultMasterNodeGroupUUID = "8870ab3e-cae8-4971-8198-3b8e37fb1ab7"
	DefaultMasterStackUUID     = "5d48650b-6707-4565-ad5a-bd9f482093b2"
)

// ListNodeGroupsResponse is a response for listing the node groups belonging to this cluster.
var ListNodeGroupsResponse = fmt.Sprintf(`
{
  "nodegroups":[
    {
      "status":"UPDATE_COMPLETE",
      "is_default":true,
      "uuid":"%s",
      "max_node_count":null,
      "stack_id":"%s",
      "min_node_count":1,
      "image_id":"5b338766-0fbf-47fa-9d9a-8f9543be9729",
      "role":"master",
      "flavor_id":"m2.medium",
      "node_count":1,
      "name":"default-master"
    },
    {
      "status":"UPDATE_COMPLETE",
      "is_default":true,
      "uuid":"%s",
      "max_node_count":null,
      "stack_id":"%s",
      "min_node_count":1,
      "image_id":"5b338766-0fbf-47fa-9d9a-8f9543be9729",
      "role":"worker",
      "flavor_id":"m2.medium",
      "node_count":1,
      "name":"default-worker"
    },
    {
      "status":"UPDATE_COMPLETE",
      "is_default":false,
      "uuid":"%s",
      "max_node_count":null,
      "stack_id":"%s",
      "min_node_count":1,
      "image_id":"5b338766-0fbf-47fa-9d9a-8f9543be9729",
      "role":"autoscaling",
      "flavor_id":"m2.medium",
      "node_count":2,
      "name":"test-ng"
    }
  ]
}
`, DefaultMasterNodeGroupUUID, DefaultMasterStackUUID,
	DefaultWorkerNodeGroupUUID, DefaultWorkerStackUUID,
	TestNodeGroupUUID, TestNodeGroupStackUUID)

// BuildTemplatedNodeGroupsListResponse builds a node groups list JSON response
// for the list of node groups that is passed in.
func BuildTemplatedNodeGroupsListResponse(ngs []*nodegroups.NodeGroup) string {
	response := make(map[string][]*nodegroups.NodeGroup)
	response["nodegroups"] = ngs

	respBytes, err := json.Marshal(response)
	if err != nil {
		panic(err)
	}
	return string(respBytes)
}

// BuildTemplatedNodeGroupsGetResponse builds a node group Get JSON response
// for a node group that is passed in.
func BuildTemplatedNodeGroupsGetResponse(ng *nodegroups.NodeGroup) string {
	respBytes, err := json.Marshal(ng)
	if err != nil {
		panic(err)
	}
	return string(respBytes)
}
