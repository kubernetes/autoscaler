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

package ovhcloud

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"slices"
	"sync"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/ovhcloud/sdk"
	"k8s.io/klog/v2"
)

const flavorCacheDuration = time.Hour

// ClientInterface defines all mandatory methods to be exposed as a client (mock or API)
type ClientInterface interface {
	// ListNodePools lists all the node pools found in a Kubernetes cluster.
	ListNodePools(ctx context.Context, projectID string, clusterID string) ([]sdk.NodePool, error)

	// ListNodePoolNodes lists all the nodes contained in a node pool.
	ListNodePoolNodes(ctx context.Context, projectID string, clusterID string, poolID string) ([]sdk.Node, error)

	// CreateNodePool fills and installs a new pool in a Kubernetes cluster.
	CreateNodePool(ctx context.Context, projectID string, clusterID string, opts *sdk.CreateNodePoolOpts) (*sdk.NodePool, error)

	// UpdateNodePool updates the details of an existing node pool.
	UpdateNodePool(ctx context.Context, projectID string, clusterID string, poolID string, opts *sdk.UpdateNodePoolOpts) (*sdk.NodePool, error)

	// DeleteNodePool deletes a specific pool.
	DeleteNodePool(ctx context.Context, projectID string, clusterID string, poolID string) (*sdk.NodePool, error)

	// ListClusterFlavors list all available flavors usable in a Kubernetes cluster.
	ListClusterFlavors(ctx context.Context, projectID string, clusterID string) ([]sdk.Flavor, error)
}

// OvhCloudManager defines current application context manager to interact
// with resources and API (or mock)
type OvhCloudManager struct {
	Client            ClientInterface
	OpenStackProvider *sdk.OpenStackProvider

	ClusterID string
	ProjectID string

	NodePoolsPerID             map[string]*sdk.NodePool
	NodeGroupPerProviderID     map[string]*NodeGroup
	NodeGroupPerProviderIDLock sync.RWMutex

	FlavorsCache               map[string]sdk.Flavor
	FlavorsCacheExpirationTime time.Time
}

// Config is the configuration file content of OVHcloud provider
type Config struct {
	// ProjectID is the id associated with the cluster project tenant.
	ProjectID string `json:"project_id"`

	// ClusterID is the id associated with the cluster where CA is running.
	ClusterID string `json:"cluster_id"`

	// AuthenticationType is the authentication method used to call the API (should be openstack or consumer)
	AuthenticationType string `json:"authentication_type"`

	// OpenStack keystone credentials if CA is run without API consumer.
	// By default, this is used as it on cluster control plane.
	OpenStackAuthUrl  string `json:"openstack_auth_url"`
	OpenStackUsername string `json:"openstack_username"`
	OpenStackPassword string `json:"openstack_password"`
	OpenStackDomain   string `json:"openstack_domain"`

	// Application credentials if CA is run as API consumer without using OpenStack keystone.
	// Tokens can be created here: https://api.ovh.com/createToken/
	ApplicationEndpoint    string `json:"application_endpoint"`
	ApplicationKey         string `json:"application_key"`
	ApplicationSecret      string `json:"application_secret"`
	ApplicationConsumerKey string `json:"application_consumer_key"`
}

// Authentication methods defines the way to interact with API.
const (
	// OpenStackAuthenticationType to request a keystone token credentials.
	OpenStackAuthenticationType = "openstack"

	// ApplicationConsumerAuthenticationType to consume an application key credentials.
	ApplicationConsumerAuthenticationType = "consumer"
)

// NewManager initializes an API client given a cloud provider configuration file
func NewManager(configFile io.Reader) (*OvhCloudManager, error) {
	var client ClientInterface
	var openStackProvider *sdk.OpenStackProvider

	// First, read configuration file to properly boot API client
	cfg, err := readConfig(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Then, validate payload
	err = validatePayload(cfg)
	if err != nil {
		return nil, fmt.Errorf("config content validation failed: %w", err)
	}

	// Eventually, create API client given its authentication method
	switch cfg.AuthenticationType {
	case OpenStackAuthenticationType:
		openStackProvider, err = sdk.NewOpenStackProvider(cfg.OpenStackAuthUrl, cfg.OpenStackUsername, cfg.OpenStackPassword, cfg.OpenStackDomain, cfg.ProjectID)
		if err != nil {
			return nil, fmt.Errorf("failed to create OpenStack provider: %w", err)
		}

		client, err = sdk.NewDefaultClientWithToken(openStackProvider.AuthUrl, openStackProvider.Token)
	case ApplicationConsumerAuthenticationType:
		client, err = sdk.NewClient(cfg.ApplicationEndpoint, cfg.ApplicationKey, cfg.ApplicationSecret, cfg.ApplicationConsumerKey)
	default:
		err = errors.New("authentication method unknown")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create API client: %w", err)
	}

	return &OvhCloudManager{
		Client:            client,
		OpenStackProvider: openStackProvider,

		ProjectID: cfg.ProjectID,
		ClusterID: cfg.ClusterID,

		NodePoolsPerID:             make(map[string]*sdk.NodePool),
		NodeGroupPerProviderID:     make(map[string]*NodeGroup),
		NodeGroupPerProviderIDLock: sync.RWMutex{},

		FlavorsCache:               make(map[string]sdk.Flavor),
		FlavorsCacheExpirationTime: time.Time{},
	}, nil
}

// getFlavorsByName lists available flavors from cache or from OVHCloud APIs if the cache is outdated
func (m *OvhCloudManager) getFlavorsByName() (map[string]sdk.Flavor, error) {
	// Update the flavors cache if expired
	if m.FlavorsCacheExpirationTime.Before(time.Now()) {
		newFlavorCacheExpirationTime := time.Now().Add(flavorCacheDuration)
		klog.V(4).Infof("Listing flavors to update flavors cache (will expire at %s)", newFlavorCacheExpirationTime)

		// Fetch all flavors in API
		flavors, err := m.Client.ListClusterFlavors(context.Background(), m.ProjectID, m.ClusterID)
		if err != nil {
			return nil, fmt.Errorf("failed to list available flavors: %w", err)
		}

		// Update the flavors cache
		m.FlavorsCache = make(map[string]sdk.Flavor)
		for _, flavor := range flavors {
			m.FlavorsCache[flavor.Name] = flavor
			m.FlavorsCacheExpirationTime = newFlavorCacheExpirationTime
		}
	}

	return m.FlavorsCache, nil
}

// getFlavorByName returns the given flavor from cache or API
func (m *OvhCloudManager) getFlavorByName(flavorName string) (sdk.Flavor, error) {
	flavorsByName, err := m.getFlavorsByName()
	if err != nil {
		return sdk.Flavor{}, err
	}

	if flavor, ok := flavorsByName[flavorName]; ok {
		return flavor, nil
	}

	return sdk.Flavor{}, fmt.Errorf("flavor %s not found in available flavors", flavorName)
}

// setNodeGroupPerProviderID stores the association provider ID => node group in cache for future reference
func (m *OvhCloudManager) setNodeGroupPerProviderID(providerID string, nodeGroup *NodeGroup) {
	m.NodeGroupPerProviderIDLock.Lock()
	defer m.NodeGroupPerProviderIDLock.Unlock()

	m.NodeGroupPerProviderID[providerID] = nodeGroup
}

// getNodeGroupPerProviderID gets from cache the node group associated to the given provider ID
func (m *OvhCloudManager) getNodeGroupPerProviderID(providerID string) *NodeGroup {
	m.NodeGroupPerProviderIDLock.RLock()
	defer m.NodeGroupPerProviderIDLock.RUnlock()

	return m.NodeGroupPerProviderID[providerID]
}

// ReAuthenticate allows OpenStack keystone token to be revoked and re-created to call API
func (m *OvhCloudManager) ReAuthenticate() error {
	if m.OpenStackProvider != nil {
		if m.OpenStackProvider.IsTokenExpired() {
			err := m.OpenStackProvider.ReauthenticateToken()
			if err != nil {
				return fmt.Errorf("failed to re-authenticate OpenStack token: %w", err)
			}

			client, err := sdk.NewDefaultClientWithToken(m.OpenStackProvider.AuthUrl, m.OpenStackProvider.Token)
			if err != nil {
				return fmt.Errorf("failed to re-create client: %w", err)
			}

			m.Client = client
		}
	}

	return nil
}

// setNodePoolsState updates nodepool local informations based on given list
// Updates NodePoolsPerID by modifying data so the reference in NodeGroupPerProviderID can access refreshed data
//
// - Updates fields on already referenced nodepool
// - Adds nodepool if not referenced yet
// - Deletes from map if nodepool is not in the given list (it doesn't exist anymore)
func (m *OvhCloudManager) setNodePoolsState(pools []sdk.NodePool) {
	m.NodeGroupPerProviderIDLock.Lock()
	defer m.NodeGroupPerProviderIDLock.Unlock()

	poolIDsToKeep := []string{}
	for _, pool := range pools {
		poolIDsToKeep = append(poolIDsToKeep, pool.ID)
	}

	// Update nodepools state
	for _, pool := range pools {
		poolRef, ok := m.NodePoolsPerID[pool.ID]
		if ok {
			*poolRef = pool // Update existing value
		} else {
			poolCopy := pool
			m.NodePoolsPerID[pool.ID] = &poolCopy
		}
	}

	// Remove nodepools that doesn't exist anymore
	for poolID := range m.NodePoolsPerID {
		if !slices.Contains(poolIDsToKeep, poolID) {
			delete(m.NodePoolsPerID, poolID)
		}
	}
}

// readConfig read cloud provider configuration file into a struct
func readConfig(configFile io.Reader) (*Config, error) {
	cfg := &Config{}
	if configFile != nil {
		body, err := ioutil.ReadAll(configFile)
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

	if cfg.ProjectID == "" {
		return fmt.Errorf("`project_id` not found in config file")
	}

	if cfg.AuthenticationType != OpenStackAuthenticationType && cfg.AuthenticationType != ApplicationConsumerAuthenticationType {
		return fmt.Errorf("`authentication_type` should only be `openstack` or `consumer`")
	}

	if cfg.AuthenticationType == OpenStackAuthenticationType {
		if cfg.OpenStackAuthUrl == "" {
			return fmt.Errorf("`openstack_auth_url` not found in config file")
		}

		if cfg.OpenStackUsername == "" {
			return fmt.Errorf("`openstack_username` not found in config file")
		}

		if cfg.OpenStackPassword == "" {
			return fmt.Errorf("`openstack_password` not found in config file")
		}

		if cfg.OpenStackDomain == "" {
			return fmt.Errorf("`openstack_domain` not found in config file")
		}
	}

	if cfg.AuthenticationType == ApplicationConsumerAuthenticationType {
		if cfg.ApplicationEndpoint == "" {
			return fmt.Errorf("`application_endpoint` not found in config file")
		}

		if cfg.ApplicationKey == "" {
			return fmt.Errorf("`application_key` not found in config file")
		}

		if cfg.ApplicationSecret == "" {
			return fmt.Errorf("`application_secret` not found in config file")
		}

		if cfg.ApplicationConsumerKey == "" {
			return fmt.Errorf("`application_consumer_key` not found in config file")
		}
	}

	return nil
}
