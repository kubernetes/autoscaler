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

package privatednsclient

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/privatedns/mgmt/2018-09-01/privatedns"
	"k8s.io/klog/v2"
	azclients "sigs.k8s.io/cloud-provider-azure/pkg/azureclients"
)

var _ Interface = &Client{}

// Client implements privatednsclient Interface.
type Client struct {
	privateDNSClient privatedns.PrivateZonesClient
}

// New creates a new privatedns client.
func New(config *azclients.ClientConfig) *Client {
	privateDNSClient := privatedns.NewPrivateZonesClientWithBaseURI(config.ResourceManagerEndpoint, config.SubscriptionID)
	privateDNSClient.Authorizer = config.Authorizer
	client := &Client{
		privateDNSClient: privateDNSClient,
	}
	return client
}

// CreateOrUpdate creates or updates a private dns zone
func (c *Client) CreateOrUpdate(ctx context.Context, resourceGroupName string, privateZoneName string, parameters privatedns.PrivateZone, waitForCompletion bool) error {
	createOrUpdateFuture, err := c.privateDNSClient.CreateOrUpdate(ctx, resourceGroupName, privateZoneName, parameters, "", "*")

	if err != nil {
		klog.V(5).Infof("Received error for %s, resourceGroup: %s, error: %s", "privatedns.put.request", resourceGroupName, err)
		return err
	}

	if waitForCompletion {
		err := createOrUpdateFuture.WaitForCompletionRef(ctx, c.privateDNSClient.Client)
		if err != nil {
			klog.V(5).Infof("Received error while waiting for completion for %s, resourceGroup: %s, error: %s", "privatedns.put.request", resourceGroupName, err)
			return err
		}
	}
	return nil
}

func (c *Client) Get(ctx context.Context, resourceGroupName string, privateZoneName string) (result privatedns.PrivateZone, err error) {
	return c.privateDNSClient.Get(ctx, resourceGroupName, privateZoneName)
}
