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
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"text/template"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

// This file contains fixtures for a hypothetical node group
// that has a node in every possible state, to test the Magnum
// manager's getNodes method.

// Node groups and stack UUIDs.
const (
	AllStatesNodeGroupUUID        = "b78e8587-785e-430c-a179-6f2c995d2993"
	AllStatesStackUUID            = "ff87db0d-fdfc-4cca-9489-cc1494064349"
	AllStatesStackName            = "kube-all-states-w5axvlfrz5lr"
	AllStatesKubeMinionsStackUUID = "2c6ef157-3d53-46b2-834e-9d1144e7f470"
	AllStatesKubeMinionsStackName = "kube-all-states-dd1a8b074cb6-kube_minions-8e1410ada96e"

	AllStatesMinion0ID         = "43e67ac6-0d22-4161-9727-4d624b3e6649"
	AllStatesMinion0ProviderID = "openstack:///43e67ac6-0d22-4161-9727-4d624b3e6649"
)

type node struct {
	Index        int
	UUID         string
	Status       string
	StatusReason string
}

// AllNodes is a list of indices, IDs and statuses of the nodes in this node group.
// There is a node in every distinct state.
var AllNodes = []node{
	// Running nodes
	{0, "888c414b-17e2-41d6-8f3d-c0b4cb296e1f", "CREATE_COMPLETE", ""},
	{1, "31c58d44-ce8d-458d-84b1-b1cdb601df81", "UPDATE_COMPLETE", ""},
	{2, "200ec335-1617-46ac-85c3-68cf083bbd08", "UPDATE_IN_PROGRESS", ""},

	// Creating nodes
	{3, "ca10499b-82c3-499b-b77f-a5a8c07f42db", "INIT_COMPLETE", ""},
	{4, "2ecad77f-a6b7-43b6-a765-9ac9d3bdd24d", "CREATE_IN_PROGRESS", ""},
	{5, "509ec187-7ef4-4753-a1b9-29d236e8d73e", "UPDATE_IN_PROGRESS", ""}, // Is UPDATE_IN_PROGRESS but still actually creating
	{6, "kube-minion", "CREATE_IN_PROGRESS", ""},                          // has not yet been assigned a server ID
	{7, "kube-minion", "UPDATE_IN_PROGRESS", ""},                          // has not yet been assigned a server ID

	// Deleting nodes
	{8, "1f09a01b-9459-459a-9c4c-cc732704b98a", "DELETE_IN_PROGRESS", ""},
	{9, "65fbfddc-67fd-4fa5-a906-dd1204ea677b", "DELETE_COMPLETE", ""},

	// Failed nodes
	{10, "58ee5f00-991e-4875-bf60-3c3b51773e5c", "CREATE_FAILED", "out of quota"},
	{11, "78ae755f-a649-40d3-889a-ad5196adbed9", "UPDATE_FAILED", "other error"},
	{12, "af4fb9d7-6aa2-4f39-b2be-e20affa87c79", "DELETE_FAILED", "delete error"},
}

// Create a provider ID from a server ID
func p(id string) string {
	return fmt.Sprintf("openstack:///%s", id)
}

// Create a fake provider ID from the minion index
func f(index int) string {
	return fmt.Sprintf("fake:///%s/%d", AllStatesNodeGroupUUID, index)
}

// ExpectedInstances are the instances that should be returned for a node group with nodes defined by AllNodes.
var ExpectedInstances = []cloudprovider.Instance{
	// Running nodes
	{p(AllNodes[0].UUID), &cloudprovider.InstanceStatus{cloudprovider.InstanceRunning, nil}},
	{p(AllNodes[1].UUID), &cloudprovider.InstanceStatus{cloudprovider.InstanceRunning, nil}},
	{p(AllNodes[2].UUID), &cloudprovider.InstanceStatus{cloudprovider.InstanceRunning, nil}},

	// Creating nodes
	{f(AllNodes[3].Index), &cloudprovider.InstanceStatus{cloudprovider.InstanceCreating, nil}},
	{f(AllNodes[4].Index), &cloudprovider.InstanceStatus{cloudprovider.InstanceCreating, nil}},
	{p(AllNodes[5].UUID), &cloudprovider.InstanceStatus{cloudprovider.InstanceCreating, nil}},
	{f(AllNodes[6].Index), &cloudprovider.InstanceStatus{cloudprovider.InstanceCreating, nil}},
	{f(AllNodes[7].Index), &cloudprovider.InstanceStatus{cloudprovider.InstanceCreating, nil}},

	// Deleting nodes
	{p(AllNodes[8].UUID), &cloudprovider.InstanceStatus{cloudprovider.InstanceDeleting, nil}},
	// node 9 is deleted so it is not reported

	// Failed nodes
	{f(AllNodes[10].Index), &cloudprovider.InstanceStatus{
		cloudprovider.InstanceCreating, &cloudprovider.InstanceErrorInfo{
			cloudprovider.OutOfResourcesErrorClass, "", "out of quota"}},
	},
	{f(AllNodes[11].Index), &cloudprovider.InstanceStatus{
		cloudprovider.InstanceCreating, &cloudprovider.InstanceErrorInfo{
			cloudprovider.OtherErrorClass, "", "other error"}},
	},
	// node 12 is not reported
}

var refsMapJSON string

func init() {
	refsMap := make(map[string]string)
	for _, node := range AllNodes {
		refsMap[strconv.Itoa(node.Index)] = node.UUID
	}

	refsMapJSONBytes, err := json.Marshal(refsMap)
	if err != nil {
		panic(err)
	}
	refsMapJSON = string(refsMapJSONBytes)

	tmpl := template.Must(template.New("minions").Parse(listAllStatesKubeMinionsResourcesTemplate))
	var b bytes.Buffer
	err = tmpl.Execute(&b, AllNodes)
	if err != nil {
		panic(err)
	}
	ListAllStatesKubeMinionsResources = b.String()

	GetAllStatesKubeMinionsStackResponse = fmt.Sprintf(getAllStatesKubeMinionsStackResponseTemplate, refsMapJSON, AllStatesStackUUID, AllStatesKubeMinionsStackUUID, AllStatesKubeMinionsStackName)
}

// GetAllStatesNodeGroupResponse is a response for a Get request for the node group.
var GetAllStatesNodeGroupResponse = fmt.Sprintf(`
{
  "links":[],
  "labels":{},
  "updated_at":"2020-04-21T09:42:29+00:00",
  "cluster_id":"91444514-b8db-4314-849c-650df15e8e83",
  "min_node_count":1,
  "id":11835,
  "version":null,
  "role":"worker",
  "node_count":1,
  "project_id":"57bf1ed0-52b8-4236-b1a7-dc58eceb8ff2",
  "status":"UPDATE_IN_PROGRESS",
  "docker_volume_size":null,
  "max_node_count":10,
  "is_default":true,
  "image_id":"5b338766-0fbf-47fa-9d9a-8f9543be9729",
  "node_addresses":[],
  "status_reason":"Stack UPDATE completed successfully",
  "name":"default-worker",
  "created_at":"2020-04-20T09:55:47+00:00",
  "flavor_id":"m2.medium",
  "uuid":"%s",
  "stack_id":"%s"
}
`, AllStatesNodeGroupUUID, AllStatesStackUUID)

// GetAllStatesStackResponse is a response for a Get request for the node group stack.
var GetAllStatesStackResponse = fmt.Sprintf(`
{
  "stack":{
    "parent":null,
    "disable_rollback":true,
    "description":"This template will boot a Kubernetes cluster with one or more minions (as specified by the number_of_minions parameter, which defaults to 1).\n",
    "parameters":{},
    "deletion_time":null,
    "stack_user_project_id":"e602edb34b394d34b6a0f50c1ac46370",
    "stack_status_reason":"Stack UPDATE completed successfully",
    "creation_time":"2020-04-20T09:55:59Z",
    "links":[],
    "capabilities":[],
    "notification_topics":[],
    "tags":null,
    "timeout_mins":60,
    "stack_status":"UPDATE_IN_PROGRESS",
    "stack_owner":null,
    "updated_time":"2020-04-21T09:42:10Z",
    "outputs":[],
    "template_description":"This template will boot a Kubernetes cluster with one or more minions (as specified by the number_of_minions parameter, which defaults to 1).\n",
    "id":"%s",
    "stack_name":"%s"
  }
}
`, AllStatesStackUUID, AllStatesStackName)

// GetAllStatesKubeMinionsResourceResponse is a response for a Get request for the node group kube_minions resource.
var GetAllStatesKubeMinionsResourceResponse = fmt.Sprintf(`
{
  "resource":{
    "resource_name":"kube_minions",
    "description":"",
    "links":[],
    "logical_resource_id":"kube_minions",
    "creation_time":"2020-04-20T09:55:59Z",
    "resource_status_reason":"state changed",
    "updated_time":"2020-04-21T09:42:18Z",
    "required_by":[],
    "resource_status":"UPDATE_IN_PROGRESS",
    "attributes":{
      "attributes":null,
      "refs":null,
      "refs_map":null,
      "removed_rsrc_list":[]
    },
    "resource_type":"OS::Heat::ResourceGroup",
    "physical_resource_id":"%s"
  }
}`, AllStatesKubeMinionsStackUUID)

// GetAllStatesKubeMinionsStackResponse is a response for a Get request for the kube_minions stack.
var GetAllStatesKubeMinionsStackResponse string
var getAllStatesKubeMinionsStackResponseTemplate = `
{
  "stack":{
    "disable_rollback":true,
    "description":"No description",
    "parameters":{},
    "deletion_time":null,
    "stack_user_project_id":"e602edb34b394d34b6a0f50c1ac46370",
    "stack_status_reason":"Stack UPDATE completed successfully",
    "creation_time":"2020-04-20T09:58:56Z",
    "links":[],
    "capabilities":[],
    "notification_topics":[],
    "tags":null,
    "timeout_mins":60,
    "stack_status":"UPDATE_COMPLETE",
    "stack_owner":null,
    "updated_time":"2020-04-21T09:42:18Z",
    "outputs":[
      {
        "output_value":%s,
        "output_key":"refs_map",
        "description":"No description given"
      }
    ],
    "template_description":"No description",
    "parent":"%s",
    "id":"%s",
    "stack_name":"%s"
  }
}`

// ListAllStatesKubeMinionsResources is a response for listing the resources belonging to the kube_minions stack.
var ListAllStatesKubeMinionsResources string
var listAllStatesKubeMinionsResourcesTemplate = `
{
  "resources":[
    {{- range $i, $node := . -}}
	{{ if $i }},{{ end }}{
      "parent_resource":"kube_minions",
      "resource_name":"{{ $node.Index }}",
      "links":[],
      "logical_resource_id":"{{ $node.Index }}",
      "creation_time":"2020-04-20T09:58:56Z",
      "resource_status_reason":"{{ $node.StatusReason }}",
      "updated_time":"2020-04-21T09:42:21Z",
      "required_by":[],
      "resource_status":"{{ $node.Status }}",
      "physical_resource_id":"{{ $node.UUID }}",
      "resource_type":"file:///usr/lib/python2.7/site-packages/magnum/drivers/k8s_fedora_coreos_v1/templates/kubeminion.yaml"
    }
	{{- end -}}
  ]
}
`

// AllStatesKubeMinionPhysicalResources is the list of physical resources
// belonging to a kube_minion. Everything except kube-minion and node_config_deployment
// has been omitted.
var allStatesKubeMinionPhysicalResourcesTemplate = `
{
  "resources":[
    {
      "parent_resource":"%s",
      "resource_name":"kube-minion",
      "links":[],
      "logical_resource_id":"kube-minion",
      "creation_time":"2020-05-05T14:30:21Z",
      "resource_status":"%s",
      "updated_time":"2020-05-05T14:30:21Z",
      "required_by":[
        "docker_volume_attach",
        "upgrade_kubernetes_deployment",
        "node_config_deployment"
      ],
      "resource_status_reason":"state changed",
      "physical_resource_id":"%s",
      "resource_type":"OS::Nova::Server"
    },
    {
      "parent_resource":"%s",
      "resource_name":"node_config_deployment",
      "links":[],
      "logical_resource_id":"node_config_deployment",
      "creation_time":"2020-05-05T14:30:21Z",
      "resource_status":"%s",
      "updated_time":"2020-05-05T14:30:21Z",
      "required_by":[],
      "resource_status_reason":"state changed",
      "physical_resource_id":"f11d7694-8de6-4753-bac9-a055deaccdba",
      "resource_type":"OS::Heat::SoftwareDeployment"
    }
  ]
}
`

// BuildAllStatesKubeMinionPhysicalResources should only be called
// for minions 2 and 5 which are UPDATE_IN_PROGRESS, to determine
// if they are still being created or not.
func BuildAllStatesKubeMinionPhysicalResources(index int) string {
	parentResource := index
	kubeMinionStatus := "CREATE_COMPLETE"
	kubeMinionPhysicalID := AllNodes[index].UUID
	configDeploymentStatus := "CREATE_COMPLETE"
	if index == 5 {
		// Minion 5 is UPDATE_IN_PROGRESS but the node is still being deployed
		configDeploymentStatus = "CREATE_IN_PROGRESS"
	}

	return fmt.Sprintf(allStatesKubeMinionPhysicalResourcesTemplate,
		parentResource, kubeMinionStatus, kubeMinionPhysicalID, parentResource, configDeploymentStatus)

}
