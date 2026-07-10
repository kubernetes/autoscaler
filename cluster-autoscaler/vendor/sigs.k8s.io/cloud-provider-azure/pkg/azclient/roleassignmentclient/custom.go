/*
Copyright 2024 The Kubernetes Authors.

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

// +azure:enableclientgen:=true
package roleassignmentclient

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	armauthorization "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/metrics"
)

// Get gets a role assignment.
const GetOperationName = "RoleAssignmentsClient.Get"

func (client *Client) Get(ctx context.Context, scope string, roleAssignmentName string, options *armauthorization.RoleAssignmentsClientGetOptions) (result *armauthorization.RoleAssignment, err error) {
	metricsCtx := metrics.BeginARMRequest(client.subscriptionID, scope, "RoleAssignment", "get")
	defer func() { metricsCtx.Observe(ctx, err) }()
	ctx, endSpan := runtime.StartSpan(ctx, GetOperationName, client.tracer, nil)
	defer endSpan(err)
	resp, err := client.RoleAssignmentsClient.Get(ctx, scope, roleAssignmentName, options)
	if err != nil {
		return nil, err
	}
	return &resp.RoleAssignment, nil
}

const DeleteOperationName = "RoleAssignmentsClient.Delete"

// Delete deletes a Subnet by name.
func (client *Client) Delete(ctx context.Context, scope string, roleAssignmentName string, options *armauthorization.RoleAssignmentsClientDeleteOptions) (result *armauthorization.RoleAssignment, err error) {
	metricsCtx := metrics.BeginARMRequest(client.subscriptionID, scope, "RoleAssignment", "delete")
	defer func() { metricsCtx.Observe(ctx, err) }()
	ctx, endSpan := runtime.StartSpan(ctx, DeleteOperationName, client.tracer, nil)
	defer endSpan(err)
	resp, err := client.RoleAssignmentsClient.Delete(ctx, scope, roleAssignmentName, options)
	if err != nil {
		return nil, err
	}
	return &resp.RoleAssignment, nil
}

const CreateOrUpdateOperationName = "RoleAssignmentsClient.Create"

// CreateOrUpdate creates or updates a Subnet.
func (client *Client) Create(ctx context.Context, scope string, roleAssignmentName string, parameters armauthorization.RoleAssignmentCreateParameters) (result *armauthorization.RoleAssignment, err error) {
	metricsCtx := metrics.BeginARMRequest(client.subscriptionID, scope, "RoleAssignment", "create")
	defer func() { metricsCtx.Observe(ctx, err) }()
	ctx, endSpan := runtime.StartSpan(ctx, CreateOrUpdateOperationName, client.tracer, nil)
	defer endSpan(err)
	resp, err := client.RoleAssignmentsClient.Create(ctx, scope, roleAssignmentName, parameters, nil)
	if err != nil {
		return nil, err
	}
	return &resp.RoleAssignment, nil
}
