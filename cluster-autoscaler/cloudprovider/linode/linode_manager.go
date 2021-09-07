/*
Copyright 2016 The Kubernetes Authors.

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

package linode

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"sync"
	"time"

	klog "k8s.io/klog/v2"
)

const (
	linodeTokenEnvVar  = "LINODE_API_TOKEN"
	lkeClusterIDEnvVar = "LKE_CLUSTER_ID"

	refreshInterval = 5 * time.Minute
)

// manager handles Linode communication and holds information about
// the node groups (LKE pools with a single linode each)
type manager struct {
	client             linodeAPIClient
	config             *linodeConfig
	nodeGroups         map[int]*NodeGroup // key: NodeGroup.id
	nodeGroupInstances map[int]*NodeGroup // key: InstanceID

	lastRefresh time.Time

	mtx sync.Mutex
}

// linodeConfig represents the configuration of the Linode cloud provider
type linodeConfig struct {
	// Linode API Token to use
	Token string `json:"token"`

	// ClusterID of the cluster for the autoscaler to manage
	ClusterID int `json:"clusterID"`

	// URL is the base URL to use for the Linode client
	// Defaults to: https://api.linode.com
	URL string `json:"url"`

	// APIVersion is the API to use for the Linode client
	// Defaults to: v4
	APIVersion string `json:"apiVersion"`
}

func newManager(config io.Reader) (*manager, error) {
	var conf linodeConfig

	conf.Token = os.Getenv(linodeTokenEnvVar)

	if clusterID, ok := os.LookupEnv(lkeClusterIDEnvVar); ok {
		if id, err := strconv.Atoi(clusterID); err == nil {
			conf.ClusterID = id
		}
	}

	if config != nil {
		configBytes, err := ioutil.ReadAll(config)
		if err != nil {
			return nil, fmt.Errorf("failed to read config: %s", err)
		}

		if err := json.Unmarshal(configBytes, &conf); err != nil {
			return nil, fmt.Errorf("failed to parse config: %s", err)
		}
	}

	if conf.Token == "" {
		return nil, errors.New("config field token is required")
	}

	if conf.ClusterID == 0 {
		return nil, errors.New("config field clusterID is required")
	}

	client := buildLinodeAPIClient(conf.URL, conf.APIVersion, conf.Token)
	m := &manager{
		client:     client,
		config:     &conf,
		nodeGroups: make(map[int]*NodeGroup),
	}
	return m, nil
}

func (m *manager) nodeGroupForNode(id int) *NodeGroup {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	return m.nodeGroupInstances[id]
}

func (m *manager) refreshAfterInterval() error {
	if m.lastRefresh.Add(refreshInterval).After(time.Now()) {
		return nil
	}
	return m.refresh()
}

func (m *manager) refresh() error {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	lkeClusterPools, err := m.client.ListLKEClusterPools(context.Background(), m.config.ClusterID, nil)
	if err != nil {
		return fmt.Errorf("failed to get list of LKE pools from linode API: %v", err)
	}

	m.lastRefresh = time.Now()

	klog.V(2).Info("refreshing LKE node groups")

	m.nodeGroups = make(map[int]*NodeGroup)
	m.nodeGroupInstances = make(map[int]*NodeGroup)
	for i := range lkeClusterPools {
		ng := nodeGroupFromPool(m.client, m.config.ClusterID, &lkeClusterPools[i])
		if ng == nil {
			continue
		}

		m.nodeGroups[ng.pool.ID] = ng
		for _, linode := range ng.pool.Linodes {
			m.nodeGroupInstances[linode.InstanceID] = ng
		}

		klog.V(2).Infof("initialized NodeGroup for pool (%d) of size %d; min = %d, max = %d",
			ng.pool.ID, len(ng.pool.Linodes), ng.pool.Autoscaler.Min, ng.pool.Autoscaler.Max)
	}

	return nil
}
