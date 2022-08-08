/*
Copyright 2022 The Kubernetes Authors.

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

package vultr

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/vultr/govultr"
	"k8s.io/klog/v2"
)

type vultrClient interface {
	ListNodePools(ctx context.Context, vkeID string, options *govultr.ListOptions) ([]govultr.NodePool, *govultr.Meta, error)
	UpdateNodePool(ctx context.Context, vkeID, nodePoolID string, updateReq *govultr.NodePoolReqUpdate) (*govultr.NodePool, error)
	DeleteNodePoolInstance(ctx context.Context, vkeID, nodePoolID, nodeID string) error
}

type manager struct {
	clusterID  string
	client     vultrClient
	nodeGroups []*NodeGroup
}

// Config is the configuration of the Vultr cloud provider
type Config struct {
	ClusterID string `json:"cluster_id"`
	Token     string `json:"token"`
}

func newManager(config io.Reader) (*manager, error) {
	cfg := &Config{}

	if config != nil {
		body, err := ioutil.ReadAll(config)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(body, cfg); err != nil {
			return nil, err
		}
	}

	//todo smarter checking to see if token is set
	if cfg.Token == "" {
		return nil, errors.New("empty token was supplied")
	}

	if cfg.ClusterID == "" {
		return nil, errors.New("empty cluster ID was supplied")
	}

	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: cfg.Token})
	oauth2Client := &http.Client{
		Timeout: 60 * time.Second,
		Transport: &oauth2.Transport{
			Source: tokenSource,
		},
	}

	m := &manager{
		client:     govultr.NewClient(oauth2Client),
		nodeGroups: make([]*NodeGroup, 0),
		clusterID:  cfg.ClusterID,
	}

	return m, nil
}

func (m *manager) Refresh() error {
	ctx := context.Background()

	//todo do we want to set the paging options here?
	nodePools, _, err := m.client.ListNodePools(ctx, m.clusterID, nil)
	if err != nil {
		return err
	}

	var group []*NodeGroup
	for _, nodePool := range nodePools {

		if !nodePool.AutoScaler {
			continue
		}

		klog.V(3).Infof("adding node pool: %q name with min nodes %d and max nodes %d", nodePool.Label, nodePool.MinNodes, nodePool.MaxNodes)

		np := nodePool
		group = append(group, &NodeGroup{
			id:        nodePool.ID,
			clusterID: m.clusterID,
			client:    m.client,
			nodePool:  &np, // we had to set this as a pointer because we don't return the [] as []*
			minSize:   nodePool.MinNodes,
			maxSize:   nodePool.MaxNodes,
		})
	}

	m.nodeGroups = group
	return nil
}
