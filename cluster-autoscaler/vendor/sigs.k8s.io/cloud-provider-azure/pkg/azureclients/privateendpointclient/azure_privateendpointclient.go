/*
Copyright 2021 The Kubernetes Authors.

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

package privateendpointclient

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2021-02-01/network"
	"k8s.io/klog/v2"
	azclients "sigs.k8s.io/cloud-provider-azure/pkg/azureclients"
)

var _ Interface = &Client{}

// Client implements privateendpointclient Interface.
type Client struct {
	privateEndpointClient network.PrivateEndpointsClient
}

// New creates a new private endpoint client.
func New(config *azclients.ClientConfig) *Client {
	privateEndpointClient := network.NewPrivateEndpointsClientWithBaseURI(config.ResourceManagerEndpoint, config.SubscriptionID)
	privateEndpointClient.Authorizer = config.Authorizer

	client := &Client{
		privateEndpointClient: privateEndpointClient,
	}
	return client
}

// CreateOrUpdate creates or updates a private endpoint.
func (c *Client) CreateOrUpdate(ctx context.Context, resourceGroupName string, endpointName string, privateEndpoint network.PrivateEndpoint, waitForCompletion bool) error {
	createOrUpdateFuture, err := c.privateEndpointClient.CreateOrUpdate(ctx, resourceGroupName, endpointName, privateEndpoint)
	if err != nil {
		klog.V(5).Infof("Received error for %s, resourceGroup: %s, error: %s", "privateendpoint.put.request", resourceGroupName, err)
		return err
	}
	if waitForCompletion {
		err = createOrUpdateFuture.WaitForCompletionRef(ctx, c.privateEndpointClient.Client)
		if err != nil {
			klog.V(5).Infof("Received error while waiting for completion for %s, resourceGroup: %s, error: %s", "privateendpoint.put.request", resourceGroupName, err)
			return err
		}

	}
	return nil
}

// Get gets the private endpoint
func (c *Client) Get(ctx context.Context, resourceGroupName string, privateEndpointName string, expand string) (result network.PrivateEndpoint, err error) {
	return c.privateEndpointClient.Get(ctx, resourceGroupName, privateEndpointName, expand)
}
