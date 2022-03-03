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

package virtualnetworklinksclient

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/privatedns/mgmt/2018-09-01/privatedns"
	"k8s.io/klog/v2"
	azclients "sigs.k8s.io/cloud-provider-azure/pkg/azureclients"
)

var _ Interface = &Client{}

// Client implements virtualnetworklinksclient Interface.
type Client struct {
	virtualNetworkLinksClient privatedns.VirtualNetworkLinksClient
}

// New creates a new virtualnetworklinks client.
func New(config *azclients.ClientConfig) *Client {
	virtualNetworkLinksClient := privatedns.NewVirtualNetworkLinksClient(config.SubscriptionID)
	virtualNetworkLinksClient.Authorizer = config.Authorizer

	client := &Client{
		virtualNetworkLinksClient: virtualNetworkLinksClient,
	}
	return client
}

// CreateOrUpdate creates or updates a virtual network link
func (c *Client) CreateOrUpdate(ctx context.Context, resourceGroupName string, privateZoneName string, virtualNetworkLinkName string, parameters privatedns.VirtualNetworkLink, waitForCompletion bool) error {
	createOrUpdateFuture, err := c.virtualNetworkLinksClient.CreateOrUpdate(ctx, resourceGroupName, privateZoneName, virtualNetworkLinkName, parameters, "", "*")
	if err != nil {
		klog.V(5).Infof("Received error for %s, resourceGroup: %s, privateZoneName: %s, error: %s", "virtualnetworklinks.put.request", resourceGroupName, privateZoneName, err)
		return err
	}
	if waitForCompletion {
		err := createOrUpdateFuture.WaitForCompletionRef(ctx, c.virtualNetworkLinksClient.Client)
		if err != nil {
			klog.V(5).Infof("Received error while waiting for completion for %s, resourceGroup: %s, privateZoneName: %s, error: %s", "virtualnetworklinks.put.request", resourceGroupName, privateZoneName, err)
			return err
		}
	}
	return nil
}

// Get gets a virtual network link
func (c *Client) Get(ctx context.Context, resourceGroupName string, privateZoneName string, virtualNetworkLinkName string) (result privatedns.VirtualNetworkLink, err error) {
	return c.virtualNetworkLinksClient.Get(ctx, resourceGroupName, privateZoneName, virtualNetworkLinkName)
}
