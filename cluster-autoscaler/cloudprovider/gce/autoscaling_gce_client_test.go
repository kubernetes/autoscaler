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
	"net/http"
	"testing"

	test_util "k8s.io/autoscaler/cluster-autoscaler/utils/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	gce_api "google.golang.org/api/compute/v1"
)

func newTestAutoscalingGceClient(t *testing.T, projectId, url string) *autoscalingGceClientV1 {
	client := &http.Client{}
	gceClient, err := NewAutoscalingGceClientV1(client, projectId)
	if !assert.NoError(t, err) {
		t.Fatalf("fatal error: %v", err)
	}
	gceClient.gceService.BasePath = url
	return gceClient
}

func TestWaitForOp(t *testing.T) {
	server := test_util.NewHttpServerMock()
	defer server.Close()
	g := newTestAutoscalingGceClient(t, "project1", server.URL)
	server.On("handle", "/project1/zones/us-central1-b/operations/operation-1505728466148-d16f5197").Return(operationRunningResponse).Times(3)
	server.On("handle", "/project1/zones/us-central1-b/operations/operation-1505728466148-d16f5197").Return(operationDoneResponse).Once()

	operation := &gce_api.Operation{Name: "operation-1505728466148-d16f5197"}

	err := g.waitForOp(operation, projectId, zoneB)
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, server)
}
