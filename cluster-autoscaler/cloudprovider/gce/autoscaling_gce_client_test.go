/*
Copyright 2018 The Kubernetes Authors.

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

package gce

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	test_util "k8s.io/autoscaler/cluster-autoscaler/utils/test"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	gce_api "google.golang.org/api/compute/v1"
)

func newTestAutoscalingGceClient(t *testing.T, projectId, url, userAgent string) *autoscalingGceClientV1 {
	client := &http.Client{}
	gceClient, err := NewAutoscalingGceClientV1(client, projectId, userAgent)
	if !assert.NoError(t, err) {
		t.Fatalf("fatal error: %v", err)
	}
	gceClient.gceService.BasePath = url
	return gceClient
}

const operationRunningResponse = `{
  "name": "operation-1505728466148-d16f5197",
  "zone": "us-central1-a",
  "operationType": "CREATE_NODE_POOL",
  "status": "RUNNING",
  "selfLink": "https://container.googleapis.com/v1/projects/601024681890/locations/us-central1-a/operations/operation-1505728466148-d16f5197",
  "targetLink": "https://container.googleapis.com/v1/projects/601024681890/locations/us-central1-a/clusters/cluster-1/nodePools/nodeautoprovisioning-323233232",
  "startTime": "2017-09-18T09:54:26.148507311Z",
  "endTime": "2017-09-18T09:54:35.124878859Z"
}`

const operationDoneResponse = `{
  "name": "operation-1505728466148-d16f5197",
  "zone": "us-central1-a",
  "operationType": "CREATE_NODE_POOL",
  "status": "DONE",
  "selfLink": "https://container.googleapis.com/v1/projects/601024681890/locations/us-central1-a/operations/operation-1505728466148-d16f5197",
  "targetLink": "https://container.googleapis.com/v1/projects/601024681890/locations/us-central1-a/clusters/cluster-1/nodePools/nodeautoprovisioning-323233232",
  "startTime": "2017-09-18T09:54:26.148507311Z",
  "endTime": "2017-09-18T09:54:35.124878859Z"
}`

const operationDoneResponseError = `{
  "endTime": "2021-12-08T11:42:45.071-08:00",
  "error": {
    "errors": [
    {
		"code": "CONDITION_NOT_MET",
		"message": "CreateInstances cannot be used when UpdatePolicy type is set to PROACTIVE and replacementMethod to SUBSTITUTE. Set replacementMethod to RECREATE or disable rolling update by setting type to OPPORTUNISTIC."
	}
	]
},
"httpErrorMessage": "PRECONDITION FAILED",
"httpErrorStatusCode": 412,
"name": "operation-1505728466148-d16f5197",
"operationType": "compute.instanceGroupManagers.createInstances",
"progress": 100,
"selfLink": "https://container.googleapis.com/v1/projects/601024681890/locations/us-central1-a/operations/operation-1505728466148-d16f5197",
"startTime": "2021-12-08T11:42:41.543-08:00",
"status": "DONE",
"targetLink": "https://container.googleapis.com/v1/projects/601024681890/locations/us-central1-a/instanceGroupManagers/workspace-ws-us21-pool",
"zone": "us-central1-a"
}`

func TestWaitForOp(t *testing.T) {
	server := test_util.NewHttpServerMock()
	defer server.Close()
	g := newTestAutoscalingGceClient(t, "project1", server.URL, "")

	g.operationPollInterval = 1 * time.Millisecond
	g.operationWaitTimeout = 500 * time.Millisecond

	server.On("handle", "/projects/project1/zones/us-central1-b/operations/operation-1505728466148-d16f5197").Return(operationRunningResponse).Times(3)
	server.On("handle", "/projects/project1/zones/us-central1-b/operations/operation-1505728466148-d16f5197").Return(operationDoneResponse).Once()

	operation := &gce_api.Operation{Name: "operation-1505728466148-d16f5197"}

	err := g.waitForOp(operation, projectId, zoneB, false)
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, server)
}

func TestWaitForOpError(t *testing.T) {
	server := test_util.NewHttpServerMock()
	defer server.Close()
	g := newTestAutoscalingGceClient(t, "project1", server.URL, "")

	server.On("handle", "/projects/project1/zones/us-central1-b/operations/operation-1505728466148-d16f5197").Return(operationDoneResponseError).Once()

	operation := &gce_api.Operation{Name: "operation-1505728466148-d16f5197"}

	err := g.waitForOp(operation, projectId, zoneB, false)
	assert.Error(t, err)
}

func TestWaitForOpTimeout(t *testing.T) {
	server := test_util.NewHttpServerMock()
	defer server.Close()
	g := newTestAutoscalingGceClient(t, "project1", server.URL, "")

	// The values here are higher than in other tests since we're aiming for timeout.
	// Lower values make this fragile and flakey.
	g.operationPollInterval = 10 * time.Millisecond
	g.operationWaitTimeout = 49 * time.Millisecond

	// Sometimes, only 3 calls are made, but it doesn't really matter,
	// so let's not assert expectations for this mock, just check for timeout error.
	server.On("handle", "/projects/project1/zones/us-central1-b/operations/operation-1505728466148-d16f5197").Return(operationRunningResponse).Times(5)

	operation := &gce_api.Operation{Name: "operation-1505728466148-d16f5197"}

	err := g.waitForOp(operation, projectId, zoneB, false)
	assert.Error(t, err)
}

func TestErrors(t *testing.T) {
	const instanceUrl = "https://content.googleapis.com/compute/v1/projects/myprojid/zones/myzone/instances/myinst"
	server := test_util.NewHttpServerMock()
	defer server.Close()
	g := newTestAutoscalingGceClient(t, "project1", server.URL, "")

	testCases := []struct {
		errorCodes         []string
		errorMessage       string
		expectedErrorCode  string
		expectedErrorClass cloudprovider.InstanceErrorClass
	}{
		{
			errorCodes:         []string{"IP_SPACE_EXHAUSTED"},
			expectedErrorCode:  "IP_SPACE_EXHAUSTED",
			expectedErrorClass: cloudprovider.OtherErrorClass,
		},
		{
			errorCodes:         []string{"RESOURCE_POOL_EXHAUSTED", "ZONE_RESOURCE_POOL_EXHAUSTED", "ZONE_RESOURCE_POOL_EXHAUSTED_WITH_DETAILS"},
			expectedErrorCode:  "RESOURCE_POOL_EXHAUSTED",
			expectedErrorClass: cloudprovider.OutOfResourcesErrorClass,
		},
		{
			errorCodes:         []string{"QUOTA"},
			expectedErrorCode:  "QUOTA_EXCEEDED",
			expectedErrorClass: cloudprovider.OutOfResourcesErrorClass,
		},
		{
			errorCodes:         []string{"PERMISSIONS_ERROR"},
			expectedErrorCode:  "PERMISSIONS_ERROR",
			expectedErrorClass: cloudprovider.OtherErrorClass,
		},
		{
			errorCodes:         []string{"CONDITION_NOT_MET"},
			errorMessage:       "Instance 'myinst' creation failed: Constraint constraints/compute.vmExternalIpAccess violated for project 1234567890.",
			expectedErrorCode:  "VM_EXTERNAL_IP_ACCESS_POLICY_CONSTRAINT",
			expectedErrorClass: cloudprovider.OtherErrorClass,
		},
		{
			errorCodes:         []string{"xyz", "abc"},
			expectedErrorCode:  "OTHER",
			expectedErrorClass: cloudprovider.OtherErrorClass,
		},
	}
	for _, tc := range testCases {
		for _, errorCode := range tc.errorCodes {
			lmiResponse := gce_api.InstanceGroupManagersListManagedInstancesResponse{
				ManagedInstances: []*gce_api.ManagedInstance{
					{
						Instance:      instanceUrl,
						CurrentAction: "CREATING",
						LastAttempt: &gce_api.ManagedInstanceLastAttempt{
							Errors: &gce_api.ManagedInstanceLastAttemptErrors{
								Errors: []*gce_api.ManagedInstanceLastAttemptErrorsErrors{
									{
										Code:    errorCode,
										Message: tc.errorMessage,
									},
								},
							},
						},
					},
				},
			}
			b, err := json.Marshal(lmiResponse)
			assert.NoError(t, err)
			server.On("handle", "/projects/zones/instanceGroupManagers/listManagedInstances").Return(string(b)).Times(1)
			instances, err := g.FetchMigInstances(GceRef{})
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedErrorCode, instances[0].Status.ErrorInfo.ErrorCode)
			assert.Equal(t, tc.expectedErrorClass, instances[0].Status.ErrorInfo.ErrorClass)
		}
	}
	mock.AssertExpectationsForObjects(t, server)
}

func TestFetchMigInstancesInstanceUrlHandling(t *testing.T) {
	const goodInstanceUrlTempl = "https://content.googleapis.com/compute/v1/projects/myprojid/zones/myzone/instances/myinst_%d"
	const badInstanceUrl = "https://badurl.com/compute/v1/projects/myprojid/zones/myzone/instances/myinst"
	server := test_util.NewHttpServerMock()
	defer server.Close()
	g := newTestAutoscalingGceClient(t, "project1", server.URL, "")

	testCases := []struct {
		name          string
		lmiResponse   gce_api.InstanceGroupManagersListManagedInstancesResponse
		wantInstances []cloudprovider.Instance
	}{
		{
			name: "all instances good",
			lmiResponse: gce_api.InstanceGroupManagersListManagedInstancesResponse{
				ManagedInstances: []*gce_api.ManagedInstance{
					{
						Instance:      fmt.Sprintf(goodInstanceUrlTempl, 2),
						CurrentAction: "CREATING",
						LastAttempt: &gce_api.ManagedInstanceLastAttempt{
							Errors: &gce_api.ManagedInstanceLastAttemptErrors{},
						},
					},
					{
						Instance:      fmt.Sprintf(goodInstanceUrlTempl, 42),
						CurrentAction: "CREATING",
						LastAttempt: &gce_api.ManagedInstanceLastAttempt{
							Errors: &gce_api.ManagedInstanceLastAttemptErrors{},
						},
					},
				},
			},
			wantInstances: []cloudprovider.Instance{
				{
					Id:     "gce://myprojid/myzone/myinst_2",
					Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceCreating},
				},
				{
					Id:     "gce://myprojid/myzone/myinst_42",
					Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceCreating},
				},
			},
		},
		{
			name: "instances with bad url",
			lmiResponse: gce_api.InstanceGroupManagersListManagedInstancesResponse{
				ManagedInstances: []*gce_api.ManagedInstance{
					{
						Instance:      badInstanceUrl,
						CurrentAction: "CREATING",
						LastAttempt: &gce_api.ManagedInstanceLastAttempt{
							Errors: &gce_api.ManagedInstanceLastAttemptErrors{},
						},
					},
					{
						Instance:      fmt.Sprintf(goodInstanceUrlTempl, 42),
						CurrentAction: "CREATING",
						LastAttempt: &gce_api.ManagedInstanceLastAttempt{
							Errors: &gce_api.ManagedInstanceLastAttemptErrors{},
						},
					},
				},
			},
			wantInstances: []cloudprovider.Instance{
				{
					Id:     "gce://myprojid/myzone/myinst_42",
					Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceCreating},
				},
			},
		},
		{
			name: "instance with empty url",
			lmiResponse: gce_api.InstanceGroupManagersListManagedInstancesResponse{
				ManagedInstances: []*gce_api.ManagedInstance{
					{
						Instance:      "",
						CurrentAction: "CREATING",
						LastAttempt: &gce_api.ManagedInstanceLastAttempt{
							Errors: &gce_api.ManagedInstanceLastAttemptErrors{},
						},
					},
					{
						Instance:      fmt.Sprintf(goodInstanceUrlTempl, 42),
						CurrentAction: "CREATING",
						LastAttempt: &gce_api.ManagedInstanceLastAttempt{
							Errors: &gce_api.ManagedInstanceLastAttemptErrors{},
						},
					},
				},
			},
			wantInstances: []cloudprovider.Instance{
				{
					Id:     "gce://myprojid/myzone/myinst_42",
					Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceCreating},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := json.Marshal(tc.lmiResponse)
			assert.NoError(t, err)
			server.On("handle", "/projects/zones/instanceGroupManagers/listManagedInstances").Return(string(b)).Times(1)
			gotInstances, err := g.FetchMigInstances(GceRef{})
			assert.NoError(t, err)
			if diff := cmp.Diff(tc.wantInstances, gotInstances, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("FetchMigInstances(...): err diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestUserAgent(t *testing.T) {
	server := test_util.NewHttpServerMock(test_util.MockFieldUserAgent, test_util.MockFieldResponse)
	defer server.Close()
	g := newTestAutoscalingGceClient(t, "project1", server.URL, "testuseragent")

	g.operationPollInterval = 10 * time.Millisecond
	g.operationWaitTimeout = 49 * time.Millisecond

	server.On("handle", "/projects/project1/zones/us-central1-b/operations/operation-1505728466148-d16f5197").Return("testuseragent", operationRunningResponse).Maybe()

	operation := &gce_api.Operation{Name: "operation-1505728466148-d16f5197"}

	g.waitForOp(operation, projectId, zoneB, false)
}
