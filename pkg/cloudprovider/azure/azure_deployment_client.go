/*
Copyright The Kubernetes Authors.

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

package azure

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/deploymentclient"
)

// DeploymentClient extends the azclient's deployment interface with ExportTemplate.
type DeploymentClient interface {
	deploymentclient.Interface
	ExportTemplate(ctx context.Context, resourceGroupName string, deploymentName string) (*armresources.DeploymentExportResult, error)
}

// deploymentClientWrapper wraps the azclient's deployment client.
type deploymentClientWrapper struct {
	deploymentclient.Interface
	client *deploymentclient.Client
}

// NewDeploymentClient creates a wrapper around the azclient's deployment client.
func NewDeploymentClient(client deploymentclient.Interface) DeploymentClient {
	c, ok := client.(*deploymentclient.Client)
	if !ok {
		return nil
	}
	return &deploymentClientWrapper{Interface: client, client: c}
}

// ExportTemplate exports the template used for specified deployment.
func (w *deploymentClientWrapper) ExportTemplate(ctx context.Context, resourceGroupName string, deploymentName string) (*armresources.DeploymentExportResult, error) {
	resp, err := w.client.DeploymentsClient.ExportTemplate(ctx, resourceGroupName, deploymentName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.DeploymentExportResult, nil
}
