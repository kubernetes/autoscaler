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

package sdk

import (
	"fmt"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum/gophercloud"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum/gophercloud/openstack"
)

// DefaultExpirationTime is the maximum time to be alive of an OpenStack keystone token.
const DefaultExpirationTime = 23 * time.Hour

// OpenStackProvider defines a custom OpenStack provider with a token to re-authenticate
type OpenStackProvider struct {
	provider *gophercloud.ProviderClient

	Token               string
	tokenExpirationTime time.Time
}

// NewOpenStackProvider initializes a client/token pair to interact with OpenStack
func NewOpenStackProvider(authUrl string, username string, password string, domain string, tenant string) (*OpenStackProvider, error) {
	provider, err := openstack.AuthenticatedClient(gophercloud.AuthOptions{
		IdentityEndpoint: authUrl,
		Username:         username,
		Password:         password,
		DomainName:       domain,
		TenantID:         tenant,
		AllowReauth:      true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenStack authenticated client: %w", err)
	}

	return &OpenStackProvider{
		provider:            provider,
		Token:               provider.Token(),
		tokenExpirationTime: time.Now().Add(DefaultExpirationTime),
	}, nil
}

// ReauthenticateToken revoke the current provider token and re-create a new one
func (p *OpenStackProvider) ReauthenticateToken() error {
	err := p.provider.Reauthenticate(p.Token)
	if err != nil {
		return fmt.Errorf("failed to re-auth previous openstack token: %w", err)
	}

	p.Token = p.provider.Token()
	p.tokenExpirationTime = time.Now().Add(DefaultExpirationTime)

	return nil
}

// IsTokenExpired checks if the current token is expired
func (p *OpenStackProvider) IsTokenExpired() bool {
	return p.tokenExpirationTime.Before(time.Now())
}
