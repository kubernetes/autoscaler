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
	"errors"
	"fmt"
	"os"
	"strconv"

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

func newManager() (*Manager, error) {
	gridscaleUUID := os.Getenv("GRIDSCALE_UUID")
	if gridscaleUUID == "" {
		return nil, errors.New("env var GRIDSCALE_UUID is not provided")
	}
	gridscaleToken := os.Getenv("GRIDSCALE_TOKEN")
	if gridscaleToken == "" {
		return nil, errors.New("env var GRIDSCALE_TOKEN is not provided")
	}
	gskClusterUUID := os.Getenv("GRIDSCALE_GSK_UUID")
	if gskClusterUUID == "" {
		return nil, errors.New("env var GRIDSCALE_GSK_UUID is not provided")
	}
	minNodeCountStr := os.Getenv("GRIDSCALE_GSK_MIN_NODE_COUNT")
	if minNodeCountStr == "" {
		return nil, errors.New("env var GRIDSCALE_GSK_MIN_NODE_COUNT is not provided")
	}
	// convert minNodeCount to int
	minNodeCount, err := strconv.Atoi(minNodeCountStr)
	if err != nil {
		return nil, fmt.Errorf("env var GRIDSCALE_GSK_MIN_NODE_COUNT is not a valid integer: %v", err)
	}
	maxNodeCountStr := os.Getenv("GRIDSCALE_GSK_MAX_NODE_COUNT")
	if maxNodeCountStr == "" {
		return nil, errors.New("env var GRIDSCALE_GSK_MAX_NODE_COUNT is not provided")
	}
	// convert maxNodeCount to int
	maxNodeCount, err := strconv.Atoi(maxNodeCountStr)
	if err != nil {
		return nil, fmt.Errorf("env var GRIDSCALE_GSK_MAX_NODE_COUNT is not a valid integer: %v", err)
	}

	apiURL := defaultGridscaleAPIURL
	envVarApiURL := os.Getenv("GRIDSCALE_API_URL")
	if envVarApiURL != "" {
		apiURL = envVarApiURL
	}
	gsConfig := gsclient.NewConfiguration(apiURL, gridscaleUUID, gridscaleToken, false, true, defaultDelayIntervalMilliSecs, defaultMaxNumberOfRetries)
	client := gsclient.NewClient(gsConfig)
	m := &Manager{
		client:       client,
		clusterUUID:  gskClusterUUID,
		nodeGroups:   make([]*NodeGroup, 0),
		maxNodeCount: maxNodeCount,
		minNodeCount: minNodeCount,
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
