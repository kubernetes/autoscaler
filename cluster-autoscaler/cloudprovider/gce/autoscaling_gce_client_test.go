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
	"net/http"
	"testing"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	test_util "k8s.io/autoscaler/cluster-autoscaler/utils/test"

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

func TestWaitForOp(t *testing.T) {
	server := test_util.NewHttpServerMock()
	defer server.Close()
	g := newTestAutoscalingGceClient(t, "project1", server.URL, "")

	g.operationPollInterval = 1 * time.Millisecond
	g.operationWaitTimeout = 500 * time.Millisecond

	server.On("handle", "/project1/zones/us-central1-b/operations/operation-1505728466148-d16f5197").Return(operationRunningResponse).Times(3)
	server.On("handle", "/project1/zones/us-central1-b/operations/operation-1505728466148-d16f5197").Return(operationDoneResponse).Once()

	operation := &gce_api.Operation{Name: "operation-1505728466148-d16f5197"}

	err := g.waitForOp(operation, projectId, zoneB, false)
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, server)
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
	server.On("handle", "/project1/zones/us-central1-b/operations/operation-1505728466148-d16f5197").Return(operationRunningResponse).Times(4)

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
										Code: errorCode,
									},
								},
							},
						},
					},
				},
			}
			b, err := json.Marshal(lmiResponse)
			assert.NoError(t, err)
			server.On("handle", "/zones/instanceGroupManagers/listManagedInstances").Return(string(b)).Times(1)
			instances, err := g.FetchMigInstances(GceRef{})
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedErrorCode, instances[0].Status.ErrorInfo.ErrorCode)
			assert.Equal(t, tc.expectedErrorClass, instances[0].Status.ErrorInfo.ErrorClass)
		}
	}
	mock.AssertExpectationsForObjects(t, server)
}

func TestUserAgent(t *testing.T) {
	server := test_util.NewHttpServerMock(test_util.MockFieldUserAgent, test_util.MockFieldResponse)
	defer server.Close()
	g := newTestAutoscalingGceClient(t, "project1", server.URL, "testuseragent")

	g.operationPollInterval = 10 * time.Millisecond
	g.operationWaitTimeout = 49 * time.Millisecond

	server.On("handle", "/project1/zones/us-central1-b/operations/operation-1505728466148-d16f5197").Return("testuseragent", operationRunningResponse).Maybe()

	operation := &gce_api.Operation{Name: "operation-1505728466148-d16f5197"}

	g.waitForOp(operation, projectId, zoneB, false)
}
