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

package exoscale

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

var testMockInstancePool1ID = "baca3aee-e609-4287-833f-573f6459ffe1"

var testMockInstancePool1 = fmt.Sprintf(`
{"getinstancepoolresponse": {
	"count": 1,
	"instancepool": [
	  {
		"id": %q,
		"keypair": "test",
		"name": "k8s-nodes1",
		"rootdisksize": 50,
		"securitygroupids": [
		  "5cbdfbb8-31ea-4791-962f-8a9719da8758"
		],
		"serviceofferingid": "5e5fb3c6-e076-429d-9b6c-b71f7b27760b",
		"size": 1,
		"state": "running",
		"templateid": "d860ceb8-684d-47e0-a6ad-970a0eec95d3",
		"virtualmachines": [
		  {
			"affinitygroup": [],
			"cpunumber": 2,
			"cpuspeed": 2198,
			"cpuused": "0.04%%",
			"created": "2020-08-25T10:04:51+0200",
			"diskioread": 843,
			"diskiowrite": 2113,
			"diskkbsread": 96120,
			"diskkbswrite": 673840,
			"displayname": "pool-1a11c-dbmaa",
			"id": "10e48003-3ac5-4b90-b9fb-c1c7c5a597ff",
			"keypair": "pierre",
			"lock": {
			  "calls": [
				"scaleVirtualMachine",
				"updateDefaultNicForVirtualMachine",
				"expungeVirtualMachine",
				"restoreVirtualMachine",
				"recoverVirtualMachine",
				"updateVirtualMachine",
				"changeServiceForVirtualMachine"
			  ]
			},
			"manager": "instancepool",
			"managerid": "1a11c398-cab1-6c91-3b94-a0561c92ce3c",
			"memory": 4096,
			"name": "pool-1a11c-dbmaa",
			"networkkbsread": 13,
			"networkkbswrite": 8,
			"nic": [
			  {
				"broadcasturi": "vlan://untagged",
				"gateway": "89.145.160.1",
				"id": "353054ab-83fe-45c6-b515-47cacacde7e6",
				"ipaddress": "89.145.160.58",
				"isdefault": true,
				"macaddress": "06:5a:b2:00:00:3f",
				"netmask": "255.255.252.0",
				"networkid": "71d5d5a8-f8b8-4331-82f5-d6f1d18ffbca",
				"networkname": "defaultGuestNetwork",
				"traffictype": "Guest",
				"type": "Shared"
			  }
			],
			"oscategoryid": "9594477e-ea0e-4c63-a642-25cbd6747493",
			"oscategoryname": "Ubuntu",
			"ostypeid": "bf3c2b62-1b0d-4432-8160-19ac837a777a",
			"passwordenabled": true,
			"rootdeviceid": 0,
			"rootdevicetype": "ROOT",
			"securitygroup": [
			  {
				"account": "exoscale-2",
				"description": "Default Security Group",
				"id": "5cbdfbb8-31ea-4791-962f-8a9719da8758",
				"name": "default"
			  }
			],
			"serviceofferingid": "5e5fb3c6-e076-429d-9b6c-b71f7b27760b",
			"serviceofferingname": "Medium",
			"state": "Running",
			"tags": [],
			"templatedisplaytext": "Linux Ubuntu 20.04 LTS 64-bit 2020-08-11-e15f6a",
			"templateid": "d860ceb8-684d-47e0-a6ad-970a0eec95d3",
			"templatename": "Linux Ubuntu 20.04 LTS 64-bit",
			"zoneid": "de88c980-78f6-467c-a431-71bcc88e437f",
			"zonename": "de-fra-1"
		  }
		],
		"zoneid": "de88c980-78f6-467c-a431-71bcc88e437f"
	  }
	]
  }}`, testMockInstancePool1ID)

var testMockInstancePool2ID = "b0520c25-66c6-440d-a533-43881a15a679"

var testMockInstancePool2 = fmt.Sprintf(`
  {"getinstancepoolresponse": {
	  "count": 1,
	  "instancepool": [
		{
		  "id": %q,
		  "keypair": "test",
		  "name": "k8s-nodes2",
		  "rootdisksize": 50,
		  "securitygroupids": [
			"5cbdfbb8-31ea-4791-962f-8a9719da8758"
		  ],
		  "serviceofferingid": "5e5fb3c6-e076-429d-9b6c-b71f7b27760b",
		  "size": 1,
		  "state": "running",
		  "templateid": "d860ceb8-684d-47e0-a6ad-970a0eec95d3",
		  "virtualmachines": [
			{
			  "affinitygroup": [],
			  "cpunumber": 2,
			  "cpuspeed": 2198,
			  "cpuused": "0.04%%",
			  "created": "2020-08-25T10:04:51+0200",
			  "diskioread": 843,
			  "diskiowrite": 2113,
			  "diskkbsread": 96120,
			  "diskkbswrite": 673840,
			  "displayname": "pool-1a11c-dbmaa",
			  "id": "10e48003-3ac5-4b90-b9fb-c1c7c5a597ff",
			  "keypair": "pierre",
			  "lock": {
				"calls": [
				  "scaleVirtualMachine",
				  "updateDefaultNicForVirtualMachine",
				  "expungeVirtualMachine",
				  "restoreVirtualMachine",
				  "recoverVirtualMachine",
				  "updateVirtualMachine",
				  "changeServiceForVirtualMachine"
				]
			  },
			  "manager": "instancepool",
			  "managerid": "1a11c398-cab1-6c91-3b94-a0561c92ce3c",
			  "memory": 4096,
			  "name": "pool-1a11c-dbmaa",
			  "networkkbsread": 13,
			  "networkkbswrite": 8,
			  "nic": [
				{
				  "broadcasturi": "vlan://untagged",
				  "gateway": "89.145.160.1",
				  "id": "353054ab-83fe-45c6-b515-47cacacde7e6",
				  "ipaddress": "89.145.160.58",
				  "isdefault": true,
				  "macaddress": "06:5a:b2:00:00:3f",
				  "netmask": "255.255.252.0",
				  "networkid": "71d5d5a8-f8b8-4331-82f5-d6f1d18ffbca",
				  "networkname": "defaultGuestNetwork",
				  "traffictype": "Guest",
				  "type": "Shared"
				}
			  ],
			  "oscategoryid": "9594477e-ea0e-4c63-a642-25cbd6747493",
			  "oscategoryname": "Ubuntu",
			  "ostypeid": "bf3c2b62-1b0d-4432-8160-19ac837a777a",
			  "passwordenabled": true,
			  "rootdeviceid": 0,
			  "rootdevicetype": "ROOT",
			  "securitygroup": [
				{
				  "account": "exoscale-2",
				  "description": "Default Security Group",
				  "id": "5cbdfbb8-31ea-4791-962f-8a9719da8758",
				  "name": "default"
				}
			  ],
			  "serviceofferingid": "5e5fb3c6-e076-429d-9b6c-b71f7b27760b",
			  "serviceofferingname": "Medium",
			  "state": "Running",
			  "tags": [],
			  "templatedisplaytext": "Linux Ubuntu 20.04 LTS 64-bit 2020-08-11-e15f6a",
			  "templateid": "d860ceb8-684d-47e0-a6ad-970a0eec95d3",
			  "templatename": "Linux Ubuntu 20.04 LTS 64-bit",
			  "zoneid": "de88c980-78f6-467c-a431-71bcc88e437f",
			  "zonename": "de-fra-1"
			}
		  ],
		  "zoneid": "de88c980-78f6-467c-a431-71bcc88e437f"
		}
	  ]
	}}`, testMockInstancePool2ID)

var testMockGetZoneID = "de88c980-78f6-467c-a431-71bcc88e437f"
var testMockGetZoneName = "de-fra-1"

var testMockGetZone = fmt.Sprintf(`
{"listzonesresponse": {
	"count": 1,
	"zone": [
	  {
		"allocationstate": "Enabled",
		"id": %q,
		"localstorageenabled": true,
		"name": %q,
		"networktype": "Basic",
		"securitygroupsenabled": true,
		"tags": [],
		"zonetoken": "c4bdb9f2-c28d-36a3-bbc5-f91fc69527e6"
	  }
	]
  }}`, testMockGetZoneID, testMockGetZoneName)

var testMockResourceLimitMax = 50

var testMockResourceLimit = fmt.Sprintf(`
{"listresourcelimitsresponse": {
	"count": 1,
	"resourcelimit": [
	  {
		"max": %d,
		"resourcetype": "0",
		"resourcetypename": "user_vm"
	  }
	]
  }}`, testMockResourceLimitMax)

var testMockInstance1ID = "7ce1c7a6-d9ca-45b5-91bd-2688dbce7ab0"

var testMockInstance1 = fmt.Sprintf(`
{"listvirtualmachinesresponse": {
	"count": 1,
	"virtualmachine": [
	  {
		"affinitygroup": [],
		"cpunumber": 2,
		"cpuspeed": 2198,
		"created": "2020-08-25T10:04:51+0200",
		"displayname": "pool-1a11c-dbmaa",
		"hypervisor": "KVM",
		"id": %q,
		"keypair": "pierre",
		"manager": "instancepool",
		"managerid": "baca3aee-e609-4287-833f-573f6459ffe1",
		"memory": 4096,
		"name": "pool-1a11c-dbmaa",
		"nic": [
		  {
			"broadcasturi": "vlan://untagged",
			"gateway": "89.145.160.1",
			"id": "353054ab-83fe-45c6-b515-47cacacde7e6",
			"ipaddress": "89.145.160.58",
			"isdefault": true,
			"macaddress": "06:5a:b2:00:00:3f",
			"netmask": "255.255.252.0",
			"networkid": "71d5d5a8-f8b8-4331-82f5-d6f1d18ffbca",
			"networkname": "defaultGuestNetwork",
			"traffictype": "Guest",
			"type": "Shared"
		  }
		],
		"oscategoryid": "9594477e-ea0e-4c63-a642-25cbd6747493",
		"oscategoryname": "Ubuntu",
		"ostypeid": "bf3c2b62-1b0d-4432-8160-19ac837a777a",
		"passwordenabled": true,
		"rootdeviceid": 0,
		"rootdevicetype": "ROOT",
		"securitygroup": [
		  {
			"account": "exoscale-2",
			"description": "Default Security Group",
			"id": "5cbdfbb8-31ea-4791-962f-8a9719da8758",
			"name": "default"
		  }
		],
		"serviceofferingid": "5e5fb3c6-e076-429d-9b6c-b71f7b27760b",
		"serviceofferingname": "Medium",
		"state": "Running",
		"tags": [],
		"templatedisplaytext": "Linux Ubuntu 20.04 LTS 64-bit 2020-08-11-e15f6a",
		"templateid": "d860ceb8-684d-47e0-a6ad-970a0eec95d3",
		"templatename": "Linux Ubuntu 20.04 LTS 64-bit",
		"zoneid": "de88c980-78f6-467c-a431-71bcc88e437f",
		"zonename": "de-fra-1"
	  }
	]
  }}`, testMockInstance1ID)

var testMockInstance2ID = "25775367-fac5-451f-b14d-7eb1869abe2c"

var testMockInstance2 = fmt.Sprintf(`
{"listvirtualmachinesresponse": {
	"count": 1,
	"virtualmachine": [
	  {
		"affinitygroup": [],
		"cpunumber": 2,
		"cpuspeed": 2198,
		"created": "2020-08-25T10:04:51+0200",
		"displayname": "pool-1a11c-dbmaa",
		"hypervisor": "KVM",
		"id": %q,
		"keypair": "pierre",
		"manager": "instancepool",
		"managerid": "b0520c25-66c6-440d-a533-43881a15a679",
		"memory": 4096,
		"name": "pool-1a11c-dbmaa",
		"nic": [
		  {
			"broadcasturi": "vlan://untagged",
			"gateway": "89.145.160.1",
			"id": "353054ab-83fe-45c6-b515-47cacacde7e6",
			"ipaddress": "89.145.160.58",
			"isdefault": true,
			"macaddress": "06:5a:b2:00:00:3f",
			"netmask": "255.255.252.0",
			"networkid": "71d5d5a8-f8b8-4331-82f5-d6f1d18ffbca",
			"networkname": "defaultGuestNetwork",
			"traffictype": "Guest",
			"type": "Shared"
		  }
		],
		"oscategoryid": "9594477e-ea0e-4c63-a642-25cbd6747493",
		"oscategoryname": "Ubuntu",
		"ostypeid": "bf3c2b62-1b0d-4432-8160-19ac837a777a",
		"passwordenabled": true,
		"rootdeviceid": 0,
		"rootdevicetype": "ROOT",
		"securitygroup": [
		  {
			"account": "exoscale-2",
			"description": "Default Security Group",
			"id": "5cbdfbb8-31ea-4791-962f-8a9719da8758",
			"name": "default"
		  }
		],
		"serviceofferingid": "5e5fb3c6-e076-429d-9b6c-b71f7b27760b",
		"serviceofferingname": "Medium",
		"state": "Running",
		"tags": [],
		"templatedisplaytext": "Linux Ubuntu 20.04 LTS 64-bit 2020-08-11-e15f6a",
		"templateid": "d860ceb8-684d-47e0-a6ad-970a0eec95d3",
		"templatename": "Linux Ubuntu 20.04 LTS 64-bit",
		"zoneid": "de88c980-78f6-467c-a431-71bcc88e437f",
		"zonename": "de-fra-1"
	  }
	]
  }}`, testMockInstance2ID)

func testMockBooleanResponse(cmd string) string {
	return fmt.Sprintf(`
{%q: {
  "success": true
}}`, cmd)
}

func testMockAPICloudProviderTest() string {
	ts := newTestServer(
		testHTTPResponse{200, testMockInstance1},
		testHTTPResponse{200, testMockGetZone},
		testHTTPResponse{200, testMockInstancePool1},
		testHTTPResponse{200, testMockInstancePool1},
		testHTTPResponse{200, testMockInstance2},
		testHTTPResponse{200, testMockGetZone},
		testHTTPResponse{200, testMockInstancePool2},
		testHTTPResponse{200, testMockInstancePool1},
		testHTTPResponse{200, testMockInstancePool2},
	)

	return ts.URL
}

type testHTTPResponse struct {
	code int
	body string
}

type testServer struct {
	*httptest.Server
	lastResponse int
	responses    []testHTTPResponse
}

func newTestServer(responses ...testHTTPResponse) *testServer {
	mux := http.NewServeMux()

	ts := &testServer{
		httptest.NewServer(mux),
		0,
		responses,
	}

	mux.Handle("/", ts)

	return ts
}

func (ts *testServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	i := ts.lastResponse
	if i >= len(ts.responses) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte("{}")) // nolint: errcheck
		return
	}
	response := ts.responses[i]
	ts.lastResponse++

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.code)
	w.Write([]byte(response.body)) // nolint: errcheck
}
