/*
Copyright 2023 The Kubernetes Authors.

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
package deploymentclient

import (
	"context"

	resources "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/utils"
)

// +azure:client:verbs=delete,resource=Deployment,packageName=github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources,packageAlias=resources,clientName=DeploymentsClient,expand=false,rateLimitKey=deploymentRateLimit
type Interface interface {
	Get(ctx context.Context, resourceGroupName string, resourceName string) (result *resources.DeploymentExtended, rerr error)
	List(ctx context.Context, resourceGroupName string) (result []*resources.DeploymentExtended, rerr error)
	CreateOrUpdate(ctx context.Context, resourceGroupName string, resourceName string, resourceParam resources.Deployment) (*resources.DeploymentExtended, error)
	utils.DeleteFunc[resources.Deployment]
}
