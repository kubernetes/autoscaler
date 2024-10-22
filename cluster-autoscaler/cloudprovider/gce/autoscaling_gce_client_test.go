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
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"testing"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/utils/klogx"
	test_util "k8s.io/autoscaler/cluster-autoscaler/utils/test"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	gce_api "google.golang.org/api/compute/v1"
)

func newTestAutoscalingGceClient(t *testing.T, projectId, url, userAgent string) *autoscalingGceClientV1 {
	return newTestAutoscalingGceClientWithTimeout(t, projectId, url, userAgent, time.Duration(0))
}

func newTestAutoscalingGceClientWithTimeout(t *testing.T, projectId, url, userAgent string, timeout time.Duration) *autoscalingGceClientV1 {
	client := &http.Client{}
	client.Timeout = timeout
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

	// default polling interval is too big for testing purposes
	g.operationPollInterval = 1 * time.Millisecond

	server.On("handle", "/projects/project1/zones/us-central1-b/operations/operation-1505728466148-d16f5197/wait").Return(operationRunningResponse).Times(3)
	server.On("handle", "/projects/project1/zones/us-central1-b/operations/operation-1505728466148-d16f5197/wait").Return(operationDoneResponse).Once()

	err := g.WaitForOperation("operation-1505728466148-d16f5197", "TestWaitForOp", projectId, zoneB)
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, server)
}

func TestWaitForOpError(t *testing.T) {
	server := test_util.NewHttpServerMock()
	defer server.Close()
	g := newTestAutoscalingGceClient(t, "project1", server.URL, "")

	server.On("handle", "/projects/project1/zones/us-central1-b/operations/operation-1505728466148-d16f5197/wait").Return(operationDoneResponseError).Once()

	err := g.WaitForOperation("operation-1505728466148-d16f5197", "TestWaitForOpError", projectId, zoneB)
	assert.Error(t, err)
	mock.AssertExpectationsForObjects(t, server)
}

func TestWaitForOpTimeout(t *testing.T) {
	server := test_util.NewHttpServerMock()
	defer server.Close()
	g := newTestAutoscalingGceClient(t, "project1", server.URL, "")

	// default polling interval and wait time are too big for the test
	g.operationWaitTimeout = 10 * time.Millisecond
	g.operationPollInterval = 20 * time.Millisecond

	server.On("handle", "/projects/project1/zones/us-central1-b/operations/operation-1505728466148-d16f5197/wait").Return(operationRunningResponse).Once()

	err := g.WaitForOperation("operation-1505728466148-d16f5197", "TestWaitForOpTimeout", projectId, zoneB)
	assert.Error(t, err)
	mock.AssertExpectationsForObjects(t, server)
}

func TestWaitForOpContextTimeout(t *testing.T) {
	server := test_util.NewHttpServerMock()
	defer server.Close()
	g := newTestAutoscalingGceClient(t, "project1", server.URL, "")

	g.operationWaitTimeout = 10 * time.Millisecond

	server.On("handle", "/projects/project1/zones/us-central1-b/operations/operation-1505728466148-d16f5197/wait").After(time.Minute).Return(operationDoneResponse).Once()

	err := g.WaitForOperation("operation-1505728466148-d16f5197", "TestWaitForOpContextTimeout", projectId, zoneB)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	mock.AssertExpectationsForObjects(t, server)
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
			errorCodes:         []string{"CONDITION_NOT_MET"},
			errorMessage:       "Instance 'myinst' creation failed: The reservation must exist in the same project as the instance.",
			expectedErrorCode:  "INVALID_RESERVATION",
			expectedErrorClass: cloudprovider.OtherErrorClass,
		},
		{
			errorCodes:         []string{"CONDITION_NOT_MET"},
			errorMessage:       "Specified reservation 'rsv-name' does not exist.",
			expectedErrorCode:  "RESERVATION_NOT_FOUND",
			expectedErrorClass: cloudprovider.OtherErrorClass,
		},
		{
			errorCodes:         []string{"CONDITION_NOT_MET"},
			errorMessage:       "Specified reservations [this-reservation-does-not-exist] do not exist. (when acting as",
			expectedErrorCode:  "RESERVATION_NOT_FOUND",
			expectedErrorClass: cloudprovider.OtherErrorClass,
		},
		{
			errorCodes:         []string{"CONDITION_NOT_MET"},
			errorMessage:       "Cannot insert instance to a reservation with status: CREATING, as it requires reservation to be in READY state.",
			expectedErrorCode:  "RESERVATION_NOT_READY",
			expectedErrorClass: cloudprovider.OtherErrorClass,
		},
		{
			errorCodes:         []string{"xyz", "abc"},
			expectedErrorCode:  "OTHER",
			expectedErrorClass: cloudprovider.OtherErrorClass,
		},
		{
			errorCodes:         []string{"CONDITION_NOT_MET"},
			errorMessage:       "Specified reservation 'rsv-name' does not have available resources for the request.",
			expectedErrorCode:  "RESERVATION_CAPACITY_EXCEEDED",
			expectedErrorClass: cloudprovider.OtherErrorClass,
		},
		{
			errorCodes:         []string{"CONDITION_NOT_MET"},
			errorMessage:       "No available resources in specified reservations 'rsv-name'",
			expectedErrorCode:  "RESERVATION_INCOMPATIBLE",
			expectedErrorClass: cloudprovider.OtherErrorClass,
		},
		{
			errorCodes:         []string{"CONDITION_NOT_MET"},
			errorMessage:       "Unsupported TPU configuration",
			expectedErrorCode:  ErrorUnsupportedTpuConfiguration,
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
	const badInstanceUrl = "https://badurl.com/compute/v1/projects3/myprojid/zones/myzone/instances/myinst"

	const instanceTemplateNameTempl = "my_inst_templ%d"
	const instanceTemplateUrlTempl = "https://content.googleapis.com/compute/v1/projects/myprojid/global/instanceTemplates/my_inst_templ%d"

	server := test_util.NewHttpServerMock()
	defer server.Close()
	g := newTestAutoscalingGceClient(t, "project1", server.URL, "")

	testCases := []struct {
		name             string
		lmiResponse      gce_api.InstanceGroupManagersListManagedInstancesResponse
		lmiPageResponses map[string]gce_api.InstanceGroupManagersListManagedInstancesResponse
		wantInstances    []GceInstance
	}{
		{
			name: "all instances good",
			lmiResponse: gce_api.InstanceGroupManagersListManagedInstancesResponse{
				ManagedInstances: []*gce_api.ManagedInstance{
					{
						Id:            2,
						Instance:      fmt.Sprintf(goodInstanceUrlTempl, 2),
						CurrentAction: "CREATING",
						LastAttempt: &gce_api.ManagedInstanceLastAttempt{
							Errors: &gce_api.ManagedInstanceLastAttemptErrors{},
						},
						Version: &gce_api.ManagedInstanceVersion{
							InstanceTemplate: fmt.Sprintf(instanceTemplateUrlTempl, 2),
						},
					},
					{
						Id:            42,
						Instance:      fmt.Sprintf(goodInstanceUrlTempl, 42),
						CurrentAction: "CREATING",
						LastAttempt: &gce_api.ManagedInstanceLastAttempt{
							Errors: &gce_api.ManagedInstanceLastAttemptErrors{},
						},
						Version: &gce_api.ManagedInstanceVersion{
							InstanceTemplate: fmt.Sprintf(instanceTemplateUrlTempl, 42),
						},
					},
				},
			},
			wantInstances: []GceInstance{
				{
					Instance: cloudprovider.Instance{
						Id:     "gce://myprojid/myzone/myinst_2",
						Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceCreating},
					},
					NumericId:            2,
					InstanceTemplateName: fmt.Sprintf(instanceTemplateNameTempl, 2),
				},
				{
					Instance: cloudprovider.Instance{
						Id:     "gce://myprojid/myzone/myinst_42",
						Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceCreating},
					},
					NumericId:            42,
					InstanceTemplateName: fmt.Sprintf(instanceTemplateNameTempl, 42),
				},
			},
		},
		{
			name: "paginated response",
			lmiResponse: gce_api.InstanceGroupManagersListManagedInstancesResponse{
				ManagedInstances: []*gce_api.ManagedInstance{
					{
						Id:            2,
						Instance:      fmt.Sprintf(goodInstanceUrlTempl, 2),
						CurrentAction: "CREATING",
						LastAttempt: &gce_api.ManagedInstanceLastAttempt{
							Errors: &gce_api.ManagedInstanceLastAttemptErrors{},
						},
						Version: &gce_api.ManagedInstanceVersion{
							InstanceTemplate: fmt.Sprintf(instanceTemplateUrlTempl, 2),
						},
					},
					{
						Id:            42,
						Instance:      fmt.Sprintf(goodInstanceUrlTempl, 42),
						CurrentAction: "CREATING",
						LastAttempt: &gce_api.ManagedInstanceLastAttempt{
							Errors: &gce_api.ManagedInstanceLastAttemptErrors{},
						},
						Version: &gce_api.ManagedInstanceVersion{
							InstanceTemplate: fmt.Sprintf(instanceTemplateUrlTempl, 42),
						},
					},
				},
				NextPageToken: "foo",
			},
			lmiPageResponses: map[string]gce_api.InstanceGroupManagersListManagedInstancesResponse{
				"foo": {
					ManagedInstances: []*gce_api.ManagedInstance{
						{
							Id:            123,
							Instance:      fmt.Sprintf(goodInstanceUrlTempl, 123),
							CurrentAction: "CREATING",
							LastAttempt: &gce_api.ManagedInstanceLastAttempt{
								Errors: &gce_api.ManagedInstanceLastAttemptErrors{},
							},
							Version: &gce_api.ManagedInstanceVersion{
								InstanceTemplate: fmt.Sprintf(instanceTemplateUrlTempl, 127),
							},
						},
						{
							Id:            456,
							Instance:      fmt.Sprintf(goodInstanceUrlTempl, 456),
							CurrentAction: "CREATING",
							LastAttempt: &gce_api.ManagedInstanceLastAttempt{
								Errors: &gce_api.ManagedInstanceLastAttemptErrors{},
							},
							Version: &gce_api.ManagedInstanceVersion{
								InstanceTemplate: fmt.Sprintf(instanceTemplateUrlTempl, 17),
							},
						},
					},
				},
			},
			wantInstances: []GceInstance{
				{
					Instance: cloudprovider.Instance{
						Id:     "gce://myprojid/myzone/myinst_2",
						Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceCreating},
					},
					NumericId:            2,
					InstanceTemplateName: fmt.Sprintf(instanceTemplateNameTempl, 2),
				},
				{
					Instance: cloudprovider.Instance{
						Id:     "gce://myprojid/myzone/myinst_42",
						Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceCreating},
					},
					NumericId:            42,
					InstanceTemplateName: fmt.Sprintf(instanceTemplateNameTempl, 42),
				},
				{
					Instance: cloudprovider.Instance{
						Id:     "gce://myprojid/myzone/myinst_123",
						Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceCreating},
					},
					NumericId:            123,
					InstanceTemplateName: fmt.Sprintf(instanceTemplateNameTempl, 127),
				},
				{
					Instance: cloudprovider.Instance{
						Id:     "gce://myprojid/myzone/myinst_456",
						Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceCreating},
					},
					NumericId:            456,
					InstanceTemplateName: fmt.Sprintf(instanceTemplateNameTempl, 17),
				},
			},
		},
		{
			name: "paginated response, more pages",
			lmiResponse: gce_api.InstanceGroupManagersListManagedInstancesResponse{
				ManagedInstances: []*gce_api.ManagedInstance{
					{
						Id:            2,
						Instance:      fmt.Sprintf(goodInstanceUrlTempl, 2),
						CurrentAction: "CREATING",
						LastAttempt: &gce_api.ManagedInstanceLastAttempt{
							Errors: &gce_api.ManagedInstanceLastAttemptErrors{},
						},
						Version: &gce_api.ManagedInstanceVersion{
							InstanceTemplate: fmt.Sprintf(instanceTemplateUrlTempl, 17),
						},
					},
					{
						Id:            42,
						Instance:      fmt.Sprintf(goodInstanceUrlTempl, 42),
						CurrentAction: "CREATING",
						LastAttempt: &gce_api.ManagedInstanceLastAttempt{
							Errors: &gce_api.ManagedInstanceLastAttemptErrors{},
						},
						Version: &gce_api.ManagedInstanceVersion{
							InstanceTemplate: fmt.Sprintf(instanceTemplateUrlTempl, 17),
						},
					},
				},
				NextPageToken: "foo",
			},
			lmiPageResponses: map[string]gce_api.InstanceGroupManagersListManagedInstancesResponse{
				"foo": {
					ManagedInstances: []*gce_api.ManagedInstance{
						{
							Id:            123,
							Instance:      fmt.Sprintf(goodInstanceUrlTempl, 123),
							CurrentAction: "CREATING",
							LastAttempt: &gce_api.ManagedInstanceLastAttempt{
								Errors: &gce_api.ManagedInstanceLastAttemptErrors{},
							},
							Version: &gce_api.ManagedInstanceVersion{
								InstanceTemplate: fmt.Sprintf(instanceTemplateUrlTempl, 17),
							},
						},
						{
							Id:            456,
							Instance:      fmt.Sprintf(goodInstanceUrlTempl, 456),
							CurrentAction: "CREATING",
							LastAttempt: &gce_api.ManagedInstanceLastAttempt{
								Errors: &gce_api.ManagedInstanceLastAttemptErrors{},
							},
							Version: &gce_api.ManagedInstanceVersion{
								InstanceTemplate: fmt.Sprintf(instanceTemplateUrlTempl, 17),
							},
						},
					},
					NextPageToken: "bar",
				},
				"bar": {
					ManagedInstances: []*gce_api.ManagedInstance{
						{
							Id:            789,
							Instance:      fmt.Sprintf(goodInstanceUrlTempl, 789),
							CurrentAction: "CREATING",
							LastAttempt: &gce_api.ManagedInstanceLastAttempt{
								Errors: &gce_api.ManagedInstanceLastAttemptErrors{},
							},
							Version: &gce_api.ManagedInstanceVersion{
								InstanceTemplate: fmt.Sprintf(instanceTemplateUrlTempl, 17),
							},
						},
						{
							Id:            666,
							Instance:      fmt.Sprintf(goodInstanceUrlTempl, 666),
							CurrentAction: "CREATING",
							LastAttempt: &gce_api.ManagedInstanceLastAttempt{
								Errors: &gce_api.ManagedInstanceLastAttemptErrors{},
							},
							Version: &gce_api.ManagedInstanceVersion{
								InstanceTemplate: fmt.Sprintf(instanceTemplateUrlTempl, 127),
							},
						},
					},
				},
			},
			wantInstances: []GceInstance{
				{
					Instance: cloudprovider.Instance{
						Id:     "gce://myprojid/myzone/myinst_2",
						Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceCreating},
					},
					NumericId:            2,
					InstanceTemplateName: fmt.Sprintf(instanceTemplateNameTempl, 17),
				},
				{
					Instance: cloudprovider.Instance{
						Id:     "gce://myprojid/myzone/myinst_42",
						Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceCreating},
					},
					NumericId:            42,
					InstanceTemplateName: fmt.Sprintf(instanceTemplateNameTempl, 17),
				},
				{
					Instance: cloudprovider.Instance{
						Id:     "gce://myprojid/myzone/myinst_123",
						Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceCreating},
					},
					NumericId:            123,
					InstanceTemplateName: fmt.Sprintf(instanceTemplateNameTempl, 17),
				},
				{
					Instance: cloudprovider.Instance{
						Id:     "gce://myprojid/myzone/myinst_456",
						Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceCreating},
					},
					NumericId:            456,
					InstanceTemplateName: fmt.Sprintf(instanceTemplateNameTempl, 17),
				},
				{
					Instance: cloudprovider.Instance{
						Id:     "gce://myprojid/myzone/myinst_789",
						Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceCreating},
					},
					NumericId:            789,
					InstanceTemplateName: fmt.Sprintf(instanceTemplateNameTempl, 17),
				},
				{
					Instance: cloudprovider.Instance{
						Id:     "gce://myprojid/myzone/myinst_666",
						Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceCreating},
					},
					NumericId:            666,
					InstanceTemplateName: fmt.Sprintf(instanceTemplateNameTempl, 127),
				},
			},
		},
		{
			name: "instances with bad url",
			lmiResponse: gce_api.InstanceGroupManagersListManagedInstancesResponse{
				ManagedInstances: []*gce_api.ManagedInstance{
					{
						Id:            99999,
						Instance:      badInstanceUrl,
						CurrentAction: "CREATING",
						LastAttempt: &gce_api.ManagedInstanceLastAttempt{
							Errors: &gce_api.ManagedInstanceLastAttemptErrors{},
						},
					},
					{
						Id:            42,
						Instance:      fmt.Sprintf(goodInstanceUrlTempl, 42),
						CurrentAction: "CREATING",
						LastAttempt: &gce_api.ManagedInstanceLastAttempt{
							Errors: &gce_api.ManagedInstanceLastAttemptErrors{},
						},
					},
				},
			},
			wantInstances: []GceInstance{
				{
					Instance: cloudprovider.Instance{
						Id:     "gce://myprojid/myzone/myinst_42",
						Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceCreating},
					},
					NumericId:            42,
					InstanceTemplateName: "",
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
						Id:            42,
						Instance:      fmt.Sprintf(goodInstanceUrlTempl, 42),
						CurrentAction: "CREATING",
						LastAttempt: &gce_api.ManagedInstanceLastAttempt{
							Errors: &gce_api.ManagedInstanceLastAttemptErrors{},
						},
					},
				},
			},
			wantInstances: []GceInstance{
				{
					Instance: cloudprovider.Instance{
						Id:     "gce://myprojid/myzone/myinst_42",
						Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceCreating},
					},
					NumericId:            42,
					InstanceTemplateName: "",
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := json.Marshal(tc.lmiResponse)
			assert.NoError(t, err)
			server.On("handle", "/projects/zones/instanceGroupManagers/listManagedInstances").Return(string(b)).Times(1)
			for token, response := range tc.lmiPageResponses {
				b, err := json.Marshal(response)
				assert.NoError(t, err)
				server.On("handle", "/projects/zones/instanceGroupManagers/listManagedInstances", token).Return(string(b)).Times(1)
			}
			gotInstances, err := g.FetchMigInstances(GceRef{})
			assert.NoError(t, err)
			if diff := cmp.Diff(tc.wantInstances, gotInstances, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("FetchMigInstances(...): err diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestFetchAvailableDiskTypes(t *testing.T) {
	server := test_util.NewHttpServerMock()
	defer server.Close()
	g := newTestAutoscalingGceClient(t, "project-id", server.URL, "")

	// ref: https://cloud.google.com/compute/docs/reference/rest/v1/diskTypes/aggregatedList
	getDiskTypesAggregatedListOKResponse, _ := os.ReadFile("fixtures/diskTypes_aggregatedList.json")
	server.On("handle", "/projects/project-id/aggregated/diskTypes").Return(string(getDiskTypesAggregatedListOKResponse)).Times(1)

	t.Run("correctly parse a response", func(t *testing.T) {
		want := map[string][]string{
			// "us-central1" region should be skipped
			"us-central1-a": {"local-ssd", "pd-balanced", "pd-ssd", "pd-standard"},
			"us-central1-b": {"hyperdisk-balanced", "hyperdisk-extreme", "hyperdisk-throughput", "local-ssd", "pd-balanced", "pd-extreme", "pd-ssd", "pd-standard"},
		}

		got, err := g.FetchAvailableDiskTypes()

		assert.NoError(t, err)
		if diff := cmp.Diff(want, got, cmpopts.EquateErrors()); diff != "" {
			t.Errorf("FetchAvailableDiskTypes(): err diff (-want +got):\n%s", diff)
		}
	})
}

func TestUserAgent(t *testing.T) {
	server := test_util.NewHttpServerMock(test_util.MockFieldUserAgent, test_util.MockFieldResponse)
	defer server.Close()
	g := newTestAutoscalingGceClient(t, "project1", server.URL, "testuseragent")

	server.On("handle", "/projects/project1/zones/us-central1-b/operations/operation-1505728466148-d16f5197/wait").Return("testuseragent", operationDoneResponse).Maybe()

	err := g.WaitForOperation("operation-1505728466148-d16f5197", "TestUserAgent", projectId, zoneB)

	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, server)
}

// NOTE: pagination operations can't be tested with context timeouts as it's not possible
// to control per call timeouts as context is global per operation
func TestAutoscalingClientTimeouts(t *testing.T) {
	// non zero timeout to indicate that timeout should be respected for http client
	instantTimeout := 1 * time.Nanosecond
	tests := map[string]struct {
		clientFunc              func(*autoscalingGceClientV1) error
		httpTimeout             time.Duration
		operationPerCallTimeout *time.Duration
	}{
		"CreateInstances_ContextTimeout": {
			clientFunc: func(client *autoscalingGceClientV1) error {
				return client.CreateInstances(GceRef{}, "", 0, nil)
			},
			operationPerCallTimeout: &instantTimeout,
		},
		"DeleteInstances_ContextTimeout": {
			clientFunc: func(client *autoscalingGceClientV1) error {
				return client.DeleteInstances(GceRef{}, nil)
			},
			operationPerCallTimeout: &instantTimeout,
		},
		"ResizeMig_ContextTimeout": {
			clientFunc: func(client *autoscalingGceClientV1) error {
				return client.ResizeMig(GceRef{}, 0)
			},
			operationPerCallTimeout: &instantTimeout,
		},
		"FetchMachineType_ContextTimeout": {
			clientFunc: func(client *autoscalingGceClientV1) error {
				_, err := client.FetchMachineType("", "")
				return err
			},
			operationPerCallTimeout: &instantTimeout,
		},
		"FetchMigBasename_ContextTimeout": {
			clientFunc: func(client *autoscalingGceClientV1) error {
				_, err := client.FetchMigBasename(GceRef{})
				return err
			},
			operationPerCallTimeout: &instantTimeout,
		},
		"FetchMigTargetSize_ContextTimeout": {
			clientFunc: func(client *autoscalingGceClientV1) error {
				_, err := client.FetchMigTargetSize(GceRef{})
				return err
			},
			operationPerCallTimeout: &instantTimeout,
		},
		"FetchMigTemplate_ContextTimeout": {
			clientFunc: func(client *autoscalingGceClientV1) error {
				_, err := client.FetchMigTemplate(GceRef{}, "", false)
				return err
			},
			operationPerCallTimeout: &instantTimeout,
		},
		"FetchMigTemplateName_ContextTimeout": {
			clientFunc: func(client *autoscalingGceClientV1) error {
				_, err := client.FetchMigTemplateName(GceRef{})
				return err
			},
			operationPerCallTimeout: &instantTimeout,
		},
		"FetchListManagedInstancesResults_ContextTimeout": {
			clientFunc: func(client *autoscalingGceClientV1) error {
				_, err := client.FetchListManagedInstancesResults(GceRef{})
				return err
			},
			operationPerCallTimeout: &instantTimeout,
		},
		"FetchZones_ContextTimeout": {
			clientFunc: func(client *autoscalingGceClientV1) error {
				_, err := client.FetchZones("")
				return err
			},
			operationPerCallTimeout: &instantTimeout,
		},
		"CreateInstances_HttpClientTimeout": {
			clientFunc: func(client *autoscalingGceClientV1) error {
				return client.CreateInstances(GceRef{}, "", 0, nil)
			},
			httpTimeout: instantTimeout,
		},
		"DeleteInstances_HttpClientTimeout": {
			clientFunc: func(client *autoscalingGceClientV1) error {
				return client.DeleteInstances(GceRef{}, nil)
			},
			httpTimeout: instantTimeout,
		},
		"ResizeMig_HttpClientTimeout": {
			clientFunc: func(client *autoscalingGceClientV1) error {
				return client.ResizeMig(GceRef{}, 0)
			},
			httpTimeout: instantTimeout,
		},
		"FetchMachineType_HttpClientTimeout": {
			clientFunc: func(client *autoscalingGceClientV1) error {
				_, err := client.FetchMachineType("", "")
				return err
			},
			httpTimeout: instantTimeout,
		},
		"FetchMigBasename_HttpClientTimeout": {
			clientFunc: func(client *autoscalingGceClientV1) error {
				_, err := client.FetchMigBasename(GceRef{})
				return err
			},
			httpTimeout: instantTimeout,
		},
		"FetchMigTargetSize_HttpClientTimeout": {
			clientFunc: func(client *autoscalingGceClientV1) error {
				_, err := client.FetchMigTargetSize(GceRef{})
				return err
			},
			httpTimeout: instantTimeout,
		},
		"FetchMigTemplate_HttpClientTimeout": {
			clientFunc: func(client *autoscalingGceClientV1) error {
				_, err := client.FetchMigTemplate(GceRef{}, "", false)
				return err
			},
			httpTimeout: instantTimeout,
		},
		"FetchMigTemplateName_HttpClientTimeout": {
			clientFunc: func(client *autoscalingGceClientV1) error {
				_, err := client.FetchMigTemplateName(GceRef{})
				return err
			},
			httpTimeout: instantTimeout,
		},
		"FetchListManagedInstancesResults_HttpClientTimeout": {
			clientFunc: func(client *autoscalingGceClientV1) error {
				_, err := client.FetchListManagedInstancesResults(GceRef{})
				return err
			},
			httpTimeout: instantTimeout,
		},
		"FetchZones_HttpClientTimeout": {
			clientFunc: func(client *autoscalingGceClientV1) error {
				_, err := client.FetchZones("")
				return err
			},
			httpTimeout: instantTimeout,
		},
		"FetchMachineTypes_HttpClientTimeout": {
			clientFunc: func(client *autoscalingGceClientV1) error {
				_, err := client.FetchMachineTypes("")
				return err
			},
			httpTimeout: instantTimeout,
		},
		"FetchAllMigs_HttpClientTimeout": {
			clientFunc: func(client *autoscalingGceClientV1) error {
				_, err := client.FetchAllMigs("")
				return err
			},
			httpTimeout: instantTimeout,
		},
		"FetchMigInstances_HttpClientTimeout": {
			clientFunc: func(client *autoscalingGceClientV1) error {
				_, err := client.FetchMigInstances(GceRef{})
				return err
			},
			httpTimeout: instantTimeout,
		},
		"FetchAvailableCpuPlatforms_HttpClientTimeout": {
			clientFunc: func(client *autoscalingGceClientV1) error {
				_, err := client.FetchAvailableCpuPlatforms()
				return err
			},
			httpTimeout: instantTimeout,
		},
		"FetchAvailableDiskTypes_HttpClientTimeout": {
			clientFunc: func(client *autoscalingGceClientV1) error {
				_, err := client.FetchAvailableDiskTypes()
				return err
			},
			httpTimeout: instantTimeout,
		},
		"FetchMigsWithName_HttpClientTimeout": {
			clientFunc: func(client *autoscalingGceClientV1) error {
				_, err := client.FetchMigsWithName("", &regexp.Regexp{})
				return err
			},
			httpTimeout: instantTimeout,
		},
		"FetchReservationsInProject_HttpClientTimeout": {
			clientFunc: func(client *autoscalingGceClientV1) error {
				_, err := client.FetchReservationsInProject("")
				return err
			},
			httpTimeout: instantTimeout,
		},
	}

	server := test_util.NewHttpServerMock()
	defer server.Close()
	server.On("handle", mock.Anything).Return(`{"status": "unreachable"}`).After(50 * time.Millisecond)
	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			client := newTestAutoscalingGceClientWithTimeout(t, "project", server.URL, "", test.httpTimeout)
			if test.operationPerCallTimeout != nil {
				client.operationPerCallTimeout = *test.operationPerCallTimeout
			}
			err := test.clientFunc(client)
			// NOTE: unable to test with ErrorIs as http errors are not wrapping an err, but overwriting it
			assert.ErrorContains(t, err, context.DeadlineExceeded.Error())
		})
	}
}

func TestFetchAllInstances(t *testing.T) {
	igm1 := "projects/893226960234/zones/zones/instanceGroupManagers/test-igm1-grp"
	igm2 := "projects/893226960234/zones/zones/instanceGroupManagers/test-igm2-grp"
	malformedIgm := "projects/893226960234/zones/zones/miss-formed/test-igm1-grp"
	tests := []struct {
		name            string
		liResponse      gce_api.InstanceList
		liPageResponses map[string]gce_api.InstanceList
		want            []GceInstance
	}{
		{
			name: "empty response",
			liResponse: gce_api.InstanceList{
				Items: []*gce_api.Instance{},
			},
			want: []GceInstance{},
		},
		{
			name: "response with malformed created-by field",
			liResponse: gce_api.InstanceList{
				Items: []*gce_api.Instance{
					{
						Id: 10,
						Metadata: &gce_api.Metadata{
							Items: []*gce_api.MetadataItems{
								{Key: "created-by", Value: &malformedIgm},
							},
						},
						SelfLink: "https://www.googleapis.com/compute/v1/projects/myprojid/zones/myzone/instances/test-instance-1",
						Status:   "RUNNING",
					},
					{
						Id: 11,
						Metadata: &gce_api.Metadata{
							Items: []*gce_api.MetadataItems{
								{Key: "created-by", Value: &igm1},
							},
						},
						SelfLink: "https://www.googleapis.com/compute/v1/projects/myprojid/zones/myzone/instances/test-instance-2",
						Status:   "PROVISIONING",
					},
				},
			},
			want: []GceInstance{
				{
					Instance: cloudprovider.Instance{
						Id:     "gce://myprojid/myzone/test-instance-1",
						Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning},
					},
					NumericId: 10,
					Igm:       GceRef{},
				},
				{
					Instance: cloudprovider.Instance{
						Id:     "gce://myprojid/myzone/test-instance-2",
						Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceCreating},
					},
					NumericId: 11,
					Igm:       GceRef{"myprojid", "zones", "test-igm1-grp"},
				},
			},
		},
		{
			name: "response without created-by field",
			liResponse: gce_api.InstanceList{
				Items: []*gce_api.Instance{
					{
						Id: 10,
						Metadata: &gce_api.Metadata{
							Items: []*gce_api.MetadataItems{},
						},
						SelfLink: "https://www.googleapis.com/compute/v1/projects/myprojid/zones/myzone/instances/test-instance-1",
						Status:   "STOPPING",
					},
				},
			},
			want: []GceInstance{
				{
					Instance: cloudprovider.Instance{
						Id:     "gce://myprojid/myzone/test-instance-1",
						Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceDeleting},
					},
					NumericId: 10,
					Igm:       GceRef{},
				},
			},
		},
		{
			name: "successfully fetch multiple instances",
			liResponse: gce_api.InstanceList{
				Items: []*gce_api.Instance{
					{
						Id: 10,
						Metadata: &gce_api.Metadata{
							Items: []*gce_api.MetadataItems{
								{Key: "created-by", Value: &igm1},
							},
						},
						SelfLink: "https://www.googleapis.com/compute/v1/projects/myprojid/zones/myzone/instances/test-instance-1",
						Status:   "RUNNING",
					},
					{
						Id: 11,
						Metadata: &gce_api.Metadata{
							Items: []*gce_api.MetadataItems{
								{Key: "created-by", Value: &igm1},
							},
						},
						SelfLink: "https://www.googleapis.com/compute/v1/projects/myprojid/zones/myzone/instances/test-instance-2",
						Status:   "RUNNING",
					},
				},
			},
			want: []GceInstance{
				{
					Instance: cloudprovider.Instance{
						Id:     "gce://myprojid/myzone/test-instance-1",
						Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning},
					},
					NumericId: 10,
					Igm:       GceRef{"myprojid", "zones", "test-igm1-grp"},
				},
				{
					Instance: cloudprovider.Instance{
						Id:     "gce://myprojid/myzone/test-instance-2",
						Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning},
					},
					NumericId: 11,
					Igm:       GceRef{"myprojid", "zones", "test-igm1-grp"},
				},
			},
		},
		{
			name: "on multiple pages",
			liResponse: gce_api.InstanceList{
				Items: []*gce_api.Instance{
					{
						Id: 10,
						Metadata: &gce_api.Metadata{
							Items: []*gce_api.MetadataItems{
								{Key: "created-by", Value: &igm1},
							},
						},
						SelfLink: "https://www.googleapis.com/compute/v1/projects/myprojid/zones/myzone/instances/test-instance-1",
						Status:   "RUNNING",
					},
					{
						Id: 11,
						Metadata: &gce_api.Metadata{
							Items: []*gce_api.MetadataItems{
								{Key: "created-by", Value: &igm2},
							},
						},
						SelfLink: "https://www.googleapis.com/compute/v1/projects/myprojid/zones/myzone/instances/test-instance-2",
						Status:   "RUNNING",
					},
				},
				NextPageToken: "foo",
			},
			liPageResponses: map[string]gce_api.InstanceList{
				"foo": {
					Items: []*gce_api.Instance{
						{
							Id: 12,
							Metadata: &gce_api.Metadata{
								Items: []*gce_api.MetadataItems{
									{Key: "created-by", Value: &igm1},
								},
							},
							SelfLink: "https://www.googleapis.com/compute/v1/projects/myprojid/zones/myzone/instances/test-instance-3",
							Status:   "RUNNING",
						},
						{
							Id: 13,
							Metadata: &gce_api.Metadata{
								Items: []*gce_api.MetadataItems{
									{Key: "created-by", Value: &igm1},
								},
							},
							SelfLink: "https://www.googleapis.com/compute/v1/projects/myprojid/zones/myzone/instances/test-instance-4",
							Status:   "RUNNING",
						},
					},
					NextPageToken: "bar",
				},
				"bar": {
					Items: []*gce_api.Instance{
						{
							Id: 14,
							Metadata: &gce_api.Metadata{
								Items: []*gce_api.MetadataItems{
									{Key: "created-by", Value: &igm2},
								},
							},
							SelfLink: "https://www.googleapis.com/compute/v1/projects/myprojid/zones/myzone/instances/test-instance-5",
							Status:   "RUNNING",
						},
						{
							Id: 15,
							Metadata: &gce_api.Metadata{
								Items: []*gce_api.MetadataItems{
									{Key: "created-by", Value: &igm1},
								},
							},
							SelfLink: "https://www.googleapis.com/compute/v1/projects/myprojid/zones/myzone/instances/test-instance-6",
							Status:   "RUNNING",
						},
					},
				},
			},
			want: []GceInstance{
				{
					Instance: cloudprovider.Instance{
						Id:     "gce://myprojid/myzone/test-instance-1",
						Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning},
					},
					NumericId: 10,
					Igm:       GceRef{"myprojid", "zones", "test-igm1-grp"},
				},
				{
					Instance: cloudprovider.Instance{
						Id:     "gce://myprojid/myzone/test-instance-2",
						Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning},
					},
					NumericId: 11,
					Igm:       GceRef{"myprojid", "zones", "test-igm2-grp"},
				},
				{
					Instance: cloudprovider.Instance{
						Id:     "gce://myprojid/myzone/test-instance-3",
						Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning},
					},
					NumericId: 12,
					Igm:       GceRef{"myprojid", "zones", "test-igm1-grp"},
				},
				{
					Instance: cloudprovider.Instance{
						Id:     "gce://myprojid/myzone/test-instance-4",
						Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning},
					},
					NumericId: 13,
					Igm:       GceRef{"myprojid", "zones", "test-igm1-grp"},
				},
				{
					Instance: cloudprovider.Instance{
						Id:     "gce://myprojid/myzone/test-instance-5",
						Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning},
					},
					NumericId: 14,
					Igm:       GceRef{"myprojid", "zones", "test-igm2-grp"},
				},
				{
					Instance: cloudprovider.Instance{
						Id:     "gce://myprojid/myzone/test-instance-6",
						Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning},
					},
					NumericId: 15,
					Igm:       GceRef{"myprojid", "zones", "test-igm1-grp"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := test_util.NewHttpServerMock()
			defer server.Close()
			gceInternalService := newTestAutoscalingGceClient(t, "project1", server.URL, "")

			b, err := json.Marshal(tt.liResponse)
			assert.NoError(t, err)
			server.On("handle", "/projects/myprojid/zones/myzone/instances").Return(string(b)).Times(1)
			for token, response := range tt.liPageResponses {
				b, err := json.Marshal(response)
				assert.NoError(t, err)
				server.On("handle", "/projects/myprojid/zones/myzone/instances", token).Return(string(b)).Times(1)
			}

			got, err := gceInternalService.FetchAllInstances("myprojid", "myzone", "test-cluster")
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("autoscalingInternalGceClient.FetchAllInstances() diff (-want +got): %s", diff)
			}
			mock.AssertExpectationsForObjects(t, server)
		})
	}
}

func TestExternalToInternalInstance(t *testing.T) {
	igm1 := "projects/893226960234/zones/zones/instanceGroupManagers/test-igm1-grp"
	tests := []struct {
		name     string
		instance *gce_api.Instance
		want     GceInstance
		wantErr  bool
	}{
		{
			name:    "nil instance argument is rejected",
			wantErr: true,
		},
		{
			name: "no created-by field",
			instance: &gce_api.Instance{
				Id:       10,
				Metadata: &gce_api.Metadata{},
				SelfLink: "https://www.googleapis.com/compute/v1/projects/myprojid/zones/myzone/instances/test-instance-1",
				Status:   "RUNNING",
			},
			want: GceInstance{
				Instance: cloudprovider.Instance{
					Id:     "gce://myprojid/myzone/test-instance-1",
					Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning},
				},
				NumericId: 10,
				Igm:       GceRef{},
			},
		},
		{
			name: "no selfLink field",
			instance: &gce_api.Instance{
				Id: 10,
				Metadata: &gce_api.Metadata{
					Items: []*gce_api.MetadataItems{
						{Key: "created-by", Value: &igm1},
					},
				},
				Status: "RUNNING",
			},
			wantErr: true,
		},
		{
			name: "wrong selfLink format",
			instance: &gce_api.Instance{
				Id: 10,
				Metadata: &gce_api.Metadata{
					Items: []*gce_api.MetadataItems{
						{Key: "created-by", Value: &igm1},
					},
				},
				SelfLink: "https://www.googleapis.com/compute/v1/projects/myprojid/zones/myzone/something/test-instance-1",
				Status:   "RUNNING",
			},
			wantErr: true,
		},
		{
			name: "successful conversion",
			instance: &gce_api.Instance{
				Id: 10,
				Metadata: &gce_api.Metadata{
					Items: []*gce_api.MetadataItems{
						{Key: "created-by", Value: &igm1},
					},
				},
				SelfLink: "https://www.googleapis.com/compute/v1/projects/myprojid/zones/myzone/instances/test-instance-1",
				Status:   "RUNNING",
			},
			want: GceInstance{
				Instance: cloudprovider.Instance{
					Id:     "gce://myprojid/myzone/test-instance-1",
					Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning},
				},
				NumericId: 10,
				Igm:       GceRef{"myprojid", "zones", "test-igm1-grp"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := externalToInternalInstance(tt.instance, klogx.NewLoggingQuota(MaxInstancesLogged))
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
