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
	DefaultWorkerNodeGroupUUID        = "f4ecc247-8732-4614-8265-716ab281c7c7"
	DefaultWorkerStackUUID            = "5d48650b-6707-4565-ad5a-bd9f482093b2"
	DefaultWorkerStackName            = "kube-w5axvlfrz5lr"
	DefaultWorkerKubeMinionsStackUUID = "f692ba59-598a-4c6d-8b69-0b78090114a8"
	DefaultWorkerKubeMinionsStackName = "kube-w5axvlfrz5lr-kube_minions-vv7bsojktpi7"

	DefaultWorkerMinion0ID         = "fa405ca9-4486-4159-9763-0f82fbc2e4ac"
	DefaultWorkerMinion0ProviderID = "openstack:///fa405ca9-4486-4159-9763-0f82fbc2e4ac"
)

// GetDefaultWorkerNodeGroupResponse is a response for a Get request for this node group.
var GetDefaultWorkerNodeGroupResponse = fmt.Sprintf(`
{
  "links":[],
  "labels":{},
  "updated_at":"2020-04-21T09:42:29+00:00",
  "cluster_id":"91444514-b8db-4314-849c-650df15e8e83",
  "min_node_count":1,
  "id":11831,
  "version":null,
  "role":"worker",
  "node_count":1,
  "project_id":"57bf1ed0-52b8-4236-b1a7-dc58eceb8ff2",
  "status":"UPDATE_COMPLETE",
  "docker_volume_size":null,
  "max_node_count":4,
  "is_default":true,
  "image_id":"5b338766-0fbf-47fa-9d9a-8f9543be9729",
  "node_addresses":[
    ""
  ],
  "status_reason":"Stack UPDATE completed successfully",
  "name":"default-worker",
  "stack_id":"5d48650b-6707-4565-ad5a-bd9f482093b2",
  "created_at":"2020-04-20T09:55:47+00:00",
  "flavor_id":"m2.medium",
  "uuid":"%s"
}
`, DefaultWorkerNodeGroupUUID)

// GetDefaultWorkerStackResponse is a response for a Get request for the node group stack.
var GetDefaultWorkerStackResponse = fmt.Sprintf(`
{
  "stack":{
    "parent":null,
    "disable_rollback":true,
    "description":"This template will boot a Kubernetes cluster with one or more minions (as specified by the number_of_minions parameter, which defaults to 1).\n",
    "parameters":{},
    "deletion_time":null,
    "stack_name":"kube-w5axvlfrz5lr",
    "stack_user_project_id":"e602edb34b394d34b6a0f50c1ac46370",
    "stack_status_reason":"Stack UPDATE completed successfully",
    "creation_time":"2020-04-20T09:55:59Z",
    "links":[],
    "capabilities":[],
    "notification_topics":[],
    "tags":null,
    "timeout_mins":60,
    "stack_status":"UPDATE_COMPLETE",
    "stack_owner":null,
    "updated_time":"2020-04-21T09:42:10Z",
    "id":"5d48650b-6707-4565-ad5a-bd9f482093b2",
    "outputs":[
      {
        "output_value":[
          "10.100.100.100"
        ],
        "output_key":"kube_masters_private",
        "description":"This is a list of the \"private\" IP addresses of all the Kubernetes masters.\n"
      },
      {
        "output_value":[
          "10.100.100.100"
        ],
        "output_key":"kube_masters",
        "description":"This is a list of the \"public\" IP addresses of all the Kubernetes masters. Use these IP addresses to log in to the Kubernetes masters via ssh.\n"
      },
      {
        "output_value":"10.100.100.100",
        "output_key":"api_address",
        "description":"This is the API endpoint of the Kubernetes cluster. Use this to access the Kubernetes API.\n"
      },
      {
        "output_value":[
          "10.100.100.101"
        ],
        "output_key":"kube_minions_private",
        "description":"This is a list of the \"private\" IP addresses of all the Kubernetes minions.\n"
      },
      {
        "output_value":[
          "10.100.100.101"
        ],
        "output_key":"kube_minions",
        "description":"This is a list of the \"public\" IP addresses of all the Kubernetes minions. Use these IP addresses to log in to the Kubernetes minions via ssh."
      },
      {
        "output_value":"localhost:5000",
        "output_key":"registry_address",
        "description":"This is the url of docker registry server where you can store docker images."
      }
    ],
    "template_description":"This template will boot a Kubernetes cluster with one or more minions (as specified by the number_of_minions parameter, which defaults to 1).\n"
  }
}
`)

// GetDefaultWorkerKubeMinionsResourceResponse is a response for a Get request for the kube_minions resource of the node group stack.
var GetDefaultWorkerKubeMinionsResourceResponse = fmt.Sprintf(`
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
    "resource_status":"UPDATE_COMPLETE",
    "physical_resource_id":"f692ba59-598a-4c6d-8b69-0b78090114a8",
    "attributes":{
      "attributes":null,
      "refs":null,
      "refs_map":null,
      "removed_rsrc_list":[]
    },
    "resource_type":"OS::Heat::ResourceGroup"
  }
}`)

// GetDefaultWorkerKubeMinionsStackResponse is a response for a Get request for the kube_minions stack.
var GetDefaultWorkerKubeMinionsStackResponse = fmt.Sprintf(`
{
  "stack":{
    "parent":"5d48650b-6707-4565-ad5a-bd9f482093b2",
    "disable_rollback":true,
    "description":"No description",
    "parameters":{
      "OS::project_id":"57bf1ed0-52b8-4236-b1a7-dc58eceb8ff2",
      "OS::stack_id":"f692ba59-598a-4c6d-8b69-0b78090114a8",
      "OS::stack_name":"kube-w5axvlfrz5lr-kube_minions-vv7bsojktpi7"
    },
    "deletion_time":null,
    "stack_name":"kube-w5axvlfrz5lr-kube_minions-vv7bsojktpi7",
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
    "id":"f692ba59-598a-4c6d-8b69-0b78090114a8",
    "outputs":[
      {
        "output_value":[
          "10.100.100.101"
        ],
        "output_key":"kube_minion_external_ip",
        "description":"No description given"
      },
      {
        "output_value":{
          "0":"%s"
        },
        "output_key":"refs_map",
        "description":"No description given"
      },
      {
        "output_value":[
          "10.100.100.101"
        ],
        "output_key":"kube_minion_ip",
        "description":"No description given"
      }
    ],
    "template_description":"No description"
  }
}`, DefaultWorkerMinion0ID)

// ListDefaultWorkerKubeMinionsResources is a response for listing the resources belonging to the kube_minions stack.
var ListDefaultWorkerKubeMinionsResources = `
{
  "resources":[
    {
      "parent_resource":"kube_minions",
      "resource_name":"0",
      "links":[],
      "logical_resource_id":"0",
      "creation_time":"2020-04-20T09:58:56Z",
      "resource_status_reason":"state changed",
      "updated_time":"2020-04-21T09:42:21Z",
      "required_by":[],
      "resource_status":"UPDATE_COMPLETE",
      "physical_resource_id":"86e1a27e-4953-4ccc-8f7c-498786b7d0c0",
      "resource_type":"file:///usr/lib/python2.7/site-packages/magnum/drivers/k8s_fedora_coreos_v1/templates/kubeminion.yaml"
    }
  ]
}
`

// ResizeDefaultWorkerNodeGroupResponse is a response for resizing a node group.
var ResizeDefaultWorkerNodeGroupResponse = `{"uuid": "add2739b-7af1-4d8c-a293-dad16c6fe789"}`

// ResizeDefaultWorkerNodeGroupError is a response for resizing a node group where the
// requested number of nodes did not fit the min/max constraints.
var ResizeDefaultWorkerNodeGroupError = `
{
  "errors":[
    {
      "status":400,
      "code":"client",
      "links":[

      ],
      "title":"Resizing default-worker outside the allowed range: min_node_count = 1, max_node_count = 4",
      "detail":"Resizing default-worker outside the allowed range: min_node_count = 1, max_node_count = 4",
      "request_id":""
    }
  ]
}
`
