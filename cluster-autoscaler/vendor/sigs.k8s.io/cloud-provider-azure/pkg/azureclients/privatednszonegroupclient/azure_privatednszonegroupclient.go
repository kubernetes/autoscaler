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

package privatednszonegroupclient

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2021-02-01/network"
	"k8s.io/klog/v2"
	azclients "sigs.k8s.io/cloud-provider-azure/pkg/azureclients"
)

var _ Interface = &Client{}

// Client implements privatednszonegroupclient client Interface.
type Client struct {
	privateDNSZoneGroupClient network.PrivateDNSZoneGroupsClient
}

// New creates a new private dns zone group client.
func New(config *azclients.ClientConfig) *Client {
	privateDNSZoneGroupClient := network.NewPrivateDNSZoneGroupsClientWithBaseURI(config.ResourceManagerEndpoint, config.SubscriptionID)
	privateDNSZoneGroupClient.Authorizer = config.Authorizer
	client := &Client{
		privateDNSZoneGroupClient: privateDNSZoneGroupClient,
	}
	return client
}

// CreateOrUpdate creates or updates a private dns zone group
func (c *Client) CreateOrUpdate(ctx context.Context, resourceGroupName string, privateEndpointName string, privateDNSZoneGroupName string, parameters network.PrivateDNSZoneGroup, waitForCompletion bool) error {
	createOrUpdateFuture, err := c.privateDNSZoneGroupClient.CreateOrUpdate(ctx, resourceGroupName, privateEndpointName, privateDNSZoneGroupName, parameters)
	if err != nil {
		klog.V(5).Infof("Received error for %s, resourceGroup: %s, privateEndpointName: %s, error: %s", "privatednszonegroup.put.request", resourceGroupName, privateEndpointName, err)
		return err
	}
	if waitForCompletion {
		err = createOrUpdateFuture.WaitForCompletionRef(ctx, c.privateDNSZoneGroupClient.Client)
		if err != nil {
			klog.V(5).Infof("Received error while waiting for completion for %s, resourceGroup: %s, privateEndpointName: %s, error: %s", "privatednszonegroup.put.request", resourceGroupName, privateEndpointName, err)
			return err
		}
	}
	return nil
}

// Get gets the private dns zone group
func (c *Client) Get(ctx context.Context, resourceGroupName string, privateEndpointName string, privateDNSZoneGroupName string) (result network.PrivateDNSZoneGroup, err error) {
	return c.privateDNSZoneGroupClient.Get(ctx, resourceGroupName, privateEndpointName, privateDNSZoneGroupName)
}
