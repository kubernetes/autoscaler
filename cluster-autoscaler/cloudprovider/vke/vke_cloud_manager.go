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

package vke

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/vke/sdk"
	"k8s.io/klog/v2"
)

const flavorCacheDuration = time.Hour

// ClientInterface defines all mandatory methods to be exposed as a client (mock or API)
type ClientInterface interface {
	// ListNodePools lists all the node pools found in a Kubernetes cluster.
	ListNodePools(ctx context.Context, clusterID string) ([]sdk.NodePool, error)

	// ListNodePoolNodes lists all the nodes contained in a node pool.
	ListNodePoolNodes(ctx context.Context, clusterID string, poolID string) ([]sdk.Node, error)

	// DeleteNode deletes a specific node.
	DeleteNode(ctx context.Context, clusterID string, nodeGroupID string, id string) error

	// ListClusterFlavors lists all available flavors usable in a Kubernetes cluster.
	ListClusterFlavors(ctx context.Context, clusterID string) ([]sdk.Flavor, error)

	// AddNode adds a node to a node pool.
	AddNode(ctx context.Context, clusterID string, nodeGroupID string) (*sdk.Node, error)
}

// VKEManager handles VKE API state for the cloud provider.
type VKEManager struct {
	Client            ClientInterface
	OpenStackProvider *sdk.OpenStackProvider

	ClusterID string
	ProjectID string

	authURL                     string
	applicationCredentialID     string
	applicationCredentialSecret string

	NodePools                  []sdk.NodePool
	NodeGroupPerProviderID     map[string]*NodeGroup
	NodeGroupPerProviderIDLock sync.RWMutex

	FlavorsCache               map[string]sdk.Flavor
	FlavorsCacheExpirationTime time.Time
}

// Config is the JSON cloud-config payload for the VKE provider.
type Config struct {
	// ClusterID is the id associated with the cluster where CA is running.
	ClusterID string `json:"cluster_id"`
	// OpenStack keystone credentials if CA is run without API consumer.
	// By default, this is used as it on cluster control plane.
	OpenStackAuthUrl            string `json:"openstack_auth_url"`
	OpenStackDomain             string `json:"openstack_domain"`
	ApplicationCredentialID     string `json:"application_key"`
	ApplicationCredentialSecret string `json:"application_secret"`
	TenantID                    string `json:"tenant_id"`
}

// NewManager initializes manager state from a cloud provider configuration file.
// OpenStack authentication is deferred until the first Refresh/ReAuthenticate call.
func NewManager(configFile io.Reader) (*VKEManager, error) {
	cfg, err := readConfig(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	err = validatePayload(cfg)
	if err != nil {
		return nil, fmt.Errorf("config content validation failed: %w", err)
	}

	return &VKEManager{
		ProjectID:                   cfg.TenantID,
		ClusterID:                   cfg.ClusterID,
		authURL:                     cfg.OpenStackAuthUrl,
		applicationCredentialID:     cfg.ApplicationCredentialID,
		applicationCredentialSecret: cfg.ApplicationCredentialSecret,

		NodePools:                  make([]sdk.NodePool, 0),
		NodeGroupPerProviderID:     make(map[string]*NodeGroup),
		NodeGroupPerProviderIDLock: sync.RWMutex{},

		FlavorsCache:               make(map[string]sdk.Flavor),
		FlavorsCacheExpirationTime: time.Time{},
	}, nil
}

func (m *VKEManager) getFlavorsByID() (map[string]sdk.Flavor, error) {
	// Update the flavors cache if expired
	if m.FlavorsCacheExpirationTime.Before(time.Now()) {
		newFlavorCacheExpirationTime := time.Now().Add(flavorCacheDuration)
		klog.V(4).Infof("Listing flavors to update flavors cache (will expire at %s)", newFlavorCacheExpirationTime)

		// Fetch all flavors in API
		flavors, err := m.Client.ListClusterFlavors(context.Background(), m.ClusterID)
		if err != nil {
			return nil, fmt.Errorf("failed to list available flavors: %w", err)
		}

		// Update the flavors cache
		m.FlavorsCache = make(map[string]sdk.Flavor)
		for _, flavor := range flavors {
			m.FlavorsCache[flavor.Id] = flavor
			m.FlavorsCacheExpirationTime = newFlavorCacheExpirationTime
		}
	}

	return m.FlavorsCache, nil
}

// getFlavorByName returns the given flavor from cache or API
func (m *VKEManager) getFlavorByName(flavorName string) (sdk.Flavor, error) {
	flavorsByName, err := m.getFlavorsByID()
	if err != nil {
		return sdk.Flavor{}, err
	}

	if flavor, ok := flavorsByName[flavorName]; ok {
		return flavor, nil
	}

	return sdk.Flavor{}, fmt.Errorf("flavor %s not found in available flavors", flavorName)
}

// setNodeGroupPerProviderID stores the association provider ID => node group in cache for future reference
func (m *VKEManager) setNodeGroupPerProviderID(providerID string, nodeGroup *NodeGroup) {
	m.NodeGroupPerProviderIDLock.Lock()
	defer m.NodeGroupPerProviderIDLock.Unlock()

	m.NodeGroupPerProviderID[providerID] = nodeGroup
}

// getNodeGroupPerProviderID gets from cache the node group associated to the given provider ID
func (m *VKEManager) getNodeGroupPerProviderID(providerID string) *NodeGroup {
	m.NodeGroupPerProviderIDLock.RLock()
	defer m.NodeGroupPerProviderIDLock.RUnlock()

	return m.NodeGroupPerProviderID[providerID]
}

// ReAuthenticate allows OpenStack keystone token to be revoked and re-created to call API
func (m *VKEManager) ReAuthenticate() error {
	if m.OpenStackProvider == nil {
		openStackProvider, err := sdk.NewOpenStackProvider(m.authURL, m.applicationCredentialID, m.applicationCredentialSecret, m.ProjectID)
		if err != nil {
			return fmt.Errorf("failed to create OpenStack provider: %w", err)
		}
		m.OpenStackProvider = openStackProvider
	} else if m.OpenStackProvider.IsTokenExpired() {
		klog.V(4).Infof("OpenStack token expired, re-authenticating")
		err := m.OpenStackProvider.ReauthenticateToken()
		if err != nil {
			return fmt.Errorf("failed to re-authenticate OpenStack token: %w", err)
		}
	}

	client, err := sdk.NewDefaultClientWithToken(m.OpenStackProvider.AuthUrl, m.OpenStackProvider.Token)
	if err != nil {
		return fmt.Errorf("failed to re-create client: %w", err)
	}

	m.Client = client

	return nil
}

// readConfig read cloud provider configuration file into a struct
func readConfig(configFile io.Reader) (*Config, error) {
	cfg := &Config{}
	if configFile != nil {
		body, err := io.ReadAll(configFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read content: %w", err)
		}

		err = json.Unmarshal(body, cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal body: %w", err)
		}
	}

	return cfg, nil
}

// validatePayload check that cloud provider configuration file is correctly formatted
func validatePayload(cfg *Config) error {
	if cfg.ClusterID == "" {
		return fmt.Errorf("`cluster_id` not found in config file")
	}

	if cfg.TenantID == "" {
		return fmt.Errorf("`tenant_id` not found in config file")
	}
	if cfg.ApplicationCredentialID == "" {
		return fmt.Errorf("`application_key` not found in config file")
	}
	if cfg.ApplicationCredentialSecret == "" {
		return fmt.Errorf("`application_secret` not found in config file")
	}
	if cfg.OpenStackAuthUrl == "" {
		return fmt.Errorf("`openstack_auth_url` not found in config file")
	}
	return nil
}
