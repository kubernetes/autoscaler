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

package registryclient

import (
	"context"

	armcontainerregistry "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerregistry/armcontainerregistry"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/utils"
)

func (client *Client) Create(ctx context.Context, resourceGroupName string, resourceName string, resourceParam armcontainerregistry.Registry) (*armcontainerregistry.Registry, error) {
	resp, err := utils.NewPollerWrapper(client.RegistriesClient.BeginCreate(ctx, resourceGroupName, resourceName, resourceParam, nil)).WaitforPollerResp(ctx)
	if err != nil {
		return nil, err
	}
	if resp != nil {
		return &resp.Registry, nil
	}
	return nil, nil
}

func (client *Client) ImportImage(ctx context.Context, resourceGroup string, resourceName string, param armcontainerregistry.ImportImageParameters) error {
	_, err := utils.NewPollerWrapper(client.RegistriesClient.BeginImportImage(ctx, resourceGroup, resourceName, param, nil)).WaitforPollerResp(ctx)
	return err
}
