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

package verda

import "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/verdacloud/verdacloud-sdk-go/verda/testutil"

// NewTestClient creates a test client using the testutil configuration approach
// This is the standard way to create test clients for unit tests
//
// Example usage:
//
//	mockServer := testutil.NewMockServer()
//	defer mockServer.Close()
//	client := NewTestClient(mockServer)
//	// Use client in tests...
func NewTestClient(mockServer *testutil.MockServer) *Client {
	config := testutil.NewTestClientConfig(mockServer)
	client, _ := NewClient(
		WithBaseURL(config.BaseURL),
		WithClientID(config.ClientID),
		WithClientSecret(config.ClientSecret),
	)
	return client
}
