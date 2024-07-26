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

package fileshareclient

import (
	"context"

	armstorage "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
)

func (client *Client) Create(ctx context.Context, resourceGroupName string, resourceName string, parentResourceName string, resource armstorage.FileShare) (*armstorage.FileShare, error) {
	resp, err := client.FileSharesClient.Create(ctx, resourceGroupName, resourceName, parentResourceName, resource, nil)
	if err != nil {
		return nil, err
	}
	return &resp.FileShare, nil
}

func (client *Client) Update(ctx context.Context, resourceGroupName string, resourceName string, parentResourceName string, resource armstorage.FileShare) (*armstorage.FileShare, error) {
	resp, err := client.FileSharesClient.Update(ctx, resourceGroupName, resourceName, parentResourceName, resource, nil)
	if err != nil {
		return nil, err
	}
	return &resp.FileShare, nil
}

// Delete deletes a FileShare by name.
func (client *Client) Delete(ctx context.Context, resourceGroupName string, parentResourceName string, resourceName string) error {
	_, err := client.FileSharesClient.Delete(ctx, resourceGroupName, parentResourceName, resourceName, nil)
	return err
}
