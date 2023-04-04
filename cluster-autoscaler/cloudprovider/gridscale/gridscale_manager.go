/*
Copyright 2019 The Kubernetes Authors.

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

package gridscale

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/gridscale/gsclient-go/v3"
)

const (
	defaultGridscaleAPIURL        = "https://api.gridscale.io"
	defaultDelayIntervalMilliSecs = 5000
	defaultMaxNumberOfRetries     = 5
)

type nodeGroupClient interface {
	GetPaaSService(ctx context.Context, id string) (gsclient.PaaSService, error)

	UpdatePaaSService(ctx context.Context, id string, body gsclient.PaaSServiceUpdateRequest) error

	GetServerList(ctx context.Context) ([]gsclient.Server, error)
}

// Manager handles gridscale communication and data caching of
// node groups (node pools in DOKS)
type Manager struct {
	client       nodeGroupClient
	clusterUUID  string
	nodeGroups   []*NodeGroup
	maxNodeCount int
	minNodeCount int
}

// Config is the configuration of the gridscale cloud provider
type Config struct {
	// ClusterUUID is the uuid associated with the gsk cluster where gridscale
	// Cluster Autoscaler is running.
	ClusterUUID string `json:"cluster_uuid"`

	// UserUUID is the uuid associated with the user who is running the
	UserUUID string `json:"user_uuid"`

	// APIToken is the API token used to authenticate with the gridscale API.
	APIToken string `json:"api_token"`

	// URL points to gridscale API. If empty, defaults to
	// https://api.gridscale.io
	APIURL string `json:"api_url"`

	// Debug enables debug logging if set to true.
	Debug bool `json:"debug"`

	MaxNodeCount int `json:"max_node_count"`

	MinNodeCount int `json:"min_node_count"`
}

func newManager(configReader io.Reader) (*Manager, error) {
	cfg := &Config{}
	if configReader != nil {
		body, err := ioutil.ReadAll(configReader)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(body, cfg)
		if err != nil {
			return nil, err
		}
	}

	if cfg.UserUUID == "" {
		return nil, errors.New("user UUID is not provided")
	}
	if cfg.APIToken == "" {
		return nil, errors.New("API token is not provided")
	}
	if cfg.ClusterUUID == "" {
		return nil, errors.New("cluster UUID is not provided")
	}
	apiURL := defaultGridscaleAPIURL
	if cfg.APIURL != "" {
		apiURL = cfg.APIURL
	}
	gsConfig := gsclient.NewConfiguration(apiURL, cfg.UserUUID, cfg.APIToken, cfg.Debug, true, defaultDelayIntervalMilliSecs, defaultMaxNumberOfRetries)
	client := gsclient.NewClient(gsConfig)
	m := &Manager{
		client:       client,
		clusterUUID:  cfg.ClusterUUID,
		nodeGroups:   make([]*NodeGroup, 0),
		maxNodeCount: cfg.MaxNodeCount,
		minNodeCount: cfg.MinNodeCount,
	}

	return m, nil
}

// Refresh refreshes the cache holding the nodegroups. This is called by the CA
// based on the `--scan-interval`. By default it's 10 seconds.
func (m *Manager) Refresh() error {
	ctx := context.Background()

	k8sCluster, err := m.client.GetPaaSService(ctx, m.clusterUUID)
	if err != nil {
		return err
	}
	nodeCount, ok := k8sCluster.Properties.Parameters["k8s_worker_node_count"].(float64)
	if !ok {
		return errors.New("k8s_worker_node_count is not found in cluster properties")
	}

	m.nodeGroups = []*NodeGroup{
		{
			id:          fmt.Sprintf("%s-nodepool0", m.clusterUUID),
			clusterUUID: m.clusterUUID,
			client:      m.client,
			nodeCount:   int(nodeCount),
			minSize:     m.minNodeCount,
			maxSize:     m.maxNodeCount,
		},
	}
	return nil
}
