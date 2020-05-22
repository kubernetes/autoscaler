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

import "fmt"

// Node groups and stack UUIDs.
const (
	TestNodeGroupUUID                 = "e9927ee4-0ca7-4ac1-ab46-4a7230c76dab"
	TestNodeGroupStackUUID            = "3773d4d9-d6c8-4493-8ea2-2a7215dfc046"
	TestNodeGroupStackName            = "kube-test-ng-j5g7osyr44tb"
	TestNodeGroupKubeMinionsStackUUID = "0c8d8826-7333-4705-bfe9-f73d35477341"
	TestNodeGroupKubeMinionsStackName = "kube-test-ng-j5g7osyr44tb-kube_minions-3vb5diwdaggh"
)

// GetTestNGNodeGroupResponse is a response for a Get request for this node group.
var GetTestNGNodeGroupResponse = fmt.Sprintf(`
{
  "links":[],
  "labels":{},
  "updated_at":"2020-05-05T14:33:19+00:00",
  "cluster_id":"91444514-b8db-4314-849c-650df15e8e83",
  "min_node_count":1,
  "id":11836,
  "version":null,
  "role":"autoscaling",
  "node_count":2,
  "project_id":"57bf1ed0-52b8-4236-b1a7-dc58eceb8ff2",
  "status":"UPDATE_COMPLETE",
  "docker_volume_size":null,
  "max_node_count":4,
  "is_default":false,
  "image_id":"5b338766-0fbf-47fa-9d9a-8f9543be9729",
  "node_addresses":[
    "10.100.100.102",
    "10.100.100.103"
  ],
  "status_reason":"Stack UPDATE completed successfully",
  "name":"test-ng",
  "stack_id":"3773d4d9-d6c8-4493-8ea2-2a7215dfc046",
  "created_at":"2020-04-20T10:20:18+00:00",
  "flavor_id":"m2.medium",
  "uuid":"%s"
}
`, TestNodeGroupUUID)

// GetTestNGStackResponse is a response for a Get request for the node group stack.
var GetTestNGStackResponse = fmt.Sprintf(`
{
  "stack":{
    "parent":null,
    "disable_rollback":true,
    "description":"This template will boot a Kubernetes cluster with one or more minions (as specified by the number_of_minions parameter, which defaults to 1).\n",
    "parameters":{},
    "deletion_time":null,
    "stack_user_project_id":"bd2594b6b0a44ff0a14e13c04a1ed031",
    "stack_status_reason":"Stack UPDATE completed successfully",
    "creation_time":"2020-04-20T10:20:23Z",
    "links":[],
    "capabilities":[],
    "notification_topics":[],
    "tags":null,
    "timeout_mins":60,
    "stack_status":"UPDATE_COMPLETE",
    "stack_owner":null,
    "updated_time":"2020-05-05T14:30:18Z",
    "outputs":[
      {
        "output_value":null,
        "output_key":"kube_masters_private",
        "description":"This is a list of the \"private\" IP addresses of all the Kubernetes masters.\n"
      },
      {
        "output_value":null,
        "output_key":"kube_masters",
        "description":"This is a list of the \"public\" IP addresses of all the Kubernetes masters. Use these IP addresses to log in to the Kubernetes masters via ssh.\n"
      },
      {
        "output_value":null,
        "output_key":"api_address",
        "description":"This is the API endpoint of the Kubernetes cluster. Use this to access the Kubernetes API.\n"
      },
      {
        "output_value":[
		  "10.100.100.102",
		  "10.100.100.103"
        ],
        "output_key":"kube_minions_private",
        "description":"This is a list of the \"private\" IP addresses of all the Kubernetes minions.\n"
      },
      {
        "output_value":[
		  "10.100.100.102",
		  "10.100.100.103"
        ],
        "output_key":"kube_minions",
        "description":"This is a list of the \"public\" IP addresses of all the Kubernetes minions. Use these IP addresses to log in to the Kubernetes minions via ssh."
      },
      {
        "output_value":null,
        "output_key":"registry_address",
        "description":"This is the url of docker registry server where you can store docker images."
      }
    ],
    "template_description":"This template will boot a Kubernetes cluster with one or more minions (as specified by the number_of_minions parameter, which defaults to 1).\n",
    "id":"%s",
    "stack_name":"%s"
  }
}`, TestNodeGroupStackUUID, TestNodeGroupStackName)

// GetTestNGKubeMinionsResourceResponse is a response for a Get request for the kube_minions resource of the node group stack.
var GetTestNGKubeMinionsResourceResponse = fmt.Sprintf(`
{
  "resource":{
    "resource_name":"kube_minions",
    "description":"",
    "links":[],
    "logical_resource_id":"kube_minions",
    "creation_time":"2020-04-20T10:20:24Z",
    "resource_status_reason":"state changed",
    "updated_time":"2020-05-05T14:30:18Z",
    "required_by":[],
    "resource_status":"UPDATE_COMPLETE",
    "attributes":{
      "attributes":null,
      "refs":null,
      "refs_map":null,
      "removed_rsrc_list":[]
    },
    "resource_type":"OS::Heat::ResourceGroup",
    "physical_resource_id":"%s"
  }
}
`, TestNodeGroupKubeMinionsStackUUID)

// GetTestNGKubeMinionsStackResponse is a response for a Get request for the kube_minions stack.
var GetTestNGKubeMinionsStackResponse = fmt.Sprintf(`
{
  "stack":{
    "parent":"3773d4d9-d6c8-4493-8ea2-2a7215dfc046",
    "disable_rollback":true,
    "description":"No description",
    "parameters":{
      "OS::project_id":"57bf1ed0-52b8-4236-b1a7-dc58eceb8ff2",
      "OS::stack_id":"0c8d8826-7333-4705-bfe9-f73d35477341",
      "OS::stack_name":"kube-test-ng-j5g7osyr44tb-kube_minions-3vb5diwdaggh"
    },
    "deletion_time":null,
    "stack_name":"kube-test-ng-j5g7osyr44tb-kube_minions-3vb5diwdaggh",
    "stack_user_project_id":"bd2594b6b0a44ff0a14e13c04a1ed031",
    "stack_status_reason":"Stack UPDATE completed successfully",
    "creation_time":"2020-04-20T10:20:24Z",
    "links":[],
    "capabilities":[],
    "notification_topics":[],
    "tags":null,
    "timeout_mins":60,
    "stack_status":"UPDATE_COMPLETE",
    "stack_owner":null,
    "updated_time":"2020-05-05T14:30:18Z",
    "id":"0c8d8826-7333-4705-bfe9-f73d35477341",
    "outputs":[
      {
        "output_value":[
		  "10.100.100.102",
		  "10.100.100.103"
        ],
        "output_key":"kube_minion_external_ip",
        "description":"No description given"
      },
      {
        "output_value":{
          "0":"fee26145-994a-4776-bcb6-3b04c332c0f5",
          "1":"d28f8e0f-da6e-434d-99aa-aa8baf2b328f"
        },
        "output_key":"refs_map",
        "description":"No description given"
      },
      {
        "output_value":[
		  "10.100.100.102",
		  "10.100.100.103"
        ],
        "output_key":"kube_minion_ip",
        "description":"No description given"
      }
    ],
    "template_description":"No description"
  }
}
`)
