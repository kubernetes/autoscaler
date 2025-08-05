/*
Copyright 2025 The Kubernetes Authors.

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

package utho

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strconv"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/utho/utho-go"
	"k8s.io/klog/v2"
)

type nodeGroupClient interface {
	// ListNodePools lists all the node pools found in a Kubernetes cluster.
	ReadNodePool(ctx context.Context, clusterId int, nodePoolId string) (*utho.NodepoolDetails, error)
	// ListNodePools lists all the node pools found in a Kubernetes cluster.
	ListNodePools(ctx context.Context, clusterID string) ([]utho.NodepoolDetails, error)
	// UpdateNodePool updates the details of an existing node pool.
	UpdateNodePool(ctx context.Context, params utho.UpdateKubernetesAutoscaleNodepool) (*utho.UpdateKubernetesAutoscaleNodepoolResponse, error)
	// DeleteNode deletes a specific node in a node pool.
	DeleteNode(ctx context.Context, params utho.DeleteNodeParams) (*utho.DeleteResponse, error)
}

// Manager manages Utho cloud provider node groups.
type Manager struct {
	clusterID  string
	client     nodeGroupClient
	nodeGroups []*NodeGroup
}

// Config is the configuration of the Utho cloud provider
type Config struct {
	// ClusterID is the id associated with the cluster where Utho
	// Cluster Autoscaler is running.
	ClusterID string `json:"cluster_id"`

	// Token is the User's Access Token associated with the cluster where
	// Utho Cluster Autoscaler is running.
	Token string `json:"token"`
}

func newManager(configReader io.Reader) (*Manager, error) {
	cfg := &Config{}

	if configReader != nil {
		body, err := io.ReadAll(configReader)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(body, cfg)
		if err != nil {
			return nil, err
		}
	}

	if cfg.Token == "" {
		return nil, errors.New("access token is not provided")
	}

	if cfg.ClusterID == "" {
		clusterID, err := getNodeLabel("cluster_id")
		if err != nil {
			klog.Warningf("failed to get cluster ID from nodes: %v", err)
			return nil, errors.New("cluster ID is not provided and couldn't be retrieved from nodes")
		}
		cfg.ClusterID = clusterID
	}

	uthoClient, err := utho.NewClient(cfg.Token)
	if err != nil {
		return nil, errors.New("unable to create utho client instance")
	}
	m := &Manager{
		client:     uthoClient.Kubernetes(),
		clusterID:  cfg.ClusterID,
		nodeGroups: make([]*NodeGroup, 0),
	}

	return m, nil
}

// Refresh updates the state of the Utho Manager
func (m *Manager) Refresh() error {
	ctx := context.Background()

	nodePools, err := m.client.ListNodePools(ctx, m.clusterID)
	if err != nil {
		return err
	}

	klog.V(4).Infof("fetched %d node pools from the cluster", len(nodePools))

	var group []*NodeGroup
	for _, nodePool := range nodePools {
		klog.V(4).Infof("node-pool %s: autoscaling=%t (min=%d max=%d)",
			nodePool.ID, nodePool.AutoScale, nodePool.MinNodes, nodePool.MaxNodes)

		if !nodePool.AutoScale {
			klog.V(4).Infof("skipping non-autoscaling node pool: %s", nodePool.ID)
			continue
		}

		clusterID, err := strconv.Atoi(m.clusterID)
		if err != nil {
			klog.Errorf("failed to convert clusterID to integer: %v", err)
			continue
		}
		group = append(group, &NodeGroup{
			id:        nodePool.ID,
			clusterID: clusterID,
			client:    m.client,
			nodePool:  &nodePool,
			minSize:   nodePool.MinNodes,
			maxSize:   nodePool.MaxNodes,
		})
	}

	klog.V(4).Infof("configured %d autoscaling node pools", len(group))

	if len(group) == 0 {
		klog.V(4).Info("cluster-autoscaler is disabled. no node pools are configured")
	}

	m.nodeGroups = group
	return nil
}
