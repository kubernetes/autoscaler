/*
Copyright 2020 The Kubernetes Authors.

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

package fileclient

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2021-09-01/storage"
)

// Interface is the client interface for creating file shares, interface for test injection.
// Don't forget to run "hack/update-mock-clients.sh" command to generate the mock client.
type Interface interface {
	CreateFileShare(ctx context.Context, resourceGroupName, accountName string, shareOptions *ShareOptions, expand string) (storage.FileShare, error)
	DeleteFileShare(ctx context.Context, resourceGroupName, accountName, name, xMsSnapshot string) error
	ResizeFileShare(ctx context.Context, resourceGroupName, accountName, name string, sizeGiB int) error
	GetFileShare(ctx context.Context, resourceGroupName, accountName, name, xMsSnapshot string) (storage.FileShare, error)
	ListFileShare(ctx context.Context, resourceGroupName, accountName, filter, expand string) ([]storage.FileShareItem, error)
	GetServiceProperties(ctx context.Context, resourceGroupName, accountName string) (storage.FileServiceProperties, error)
	SetServiceProperties(ctx context.Context, resourceGroupName, accountName string, parameters storage.FileServiceProperties) (storage.FileServiceProperties, error)
	WithSubscriptionID(subscriptionID string) Interface
}
