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

package sshpublickeyresourceclient

import (
	"context"

	armcompute "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"
)

// Delete deletes a SSHPublicKeyResource by name.
func (client *Client) Delete(ctx context.Context, resourceGroupName string, resourceName string) error {
	_, err := client.SSHPublicKeysClient.Delete(ctx, resourceGroupName, resourceName, nil)
	return err
}

func (client *Client) GenerateKeyPair(ctx context.Context, resourceGroupName string, sshPublicKeyName string) (*armcompute.SSHPublicKeyGenerateKeyPairResult, error) {
	resp, err := client.SSHPublicKeysClient.GenerateKeyPair(ctx, resourceGroupName, sshPublicKeyName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.SSHPublicKeyGenerateKeyPairResult, nil
}

func (client *Client) Create(ctx context.Context, resourceGroupName string, sshPublicKeyName string, parameters armcompute.SSHPublicKeyResource) (*armcompute.SSHPublicKeyResource, error) {
	resp, err := client.SSHPublicKeysClient.Create(ctx, resourceGroupName, sshPublicKeyName, parameters, nil)
	if err != nil {
		return nil, err
	}
	return &resp.SSHPublicKeyResource, nil
}

func (client *Client) Update(ctx context.Context, resourceGroupName string, sshPublicKeyName string, parameters armcompute.SSHPublicKeyUpdateResource) (*armcompute.SSHPublicKeyResource, error) {
	resp, err := client.SSHPublicKeysClient.Update(ctx, resourceGroupName, sshPublicKeyName, parameters, nil)
	if err != nil {
		return nil, err
	}
	return &resp.SSHPublicKeyResource, nil
}
