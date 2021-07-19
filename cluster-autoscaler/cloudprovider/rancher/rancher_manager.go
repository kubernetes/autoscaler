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

package rancher

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"gopkg.in/gcfg.v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/rancher/rancher"
	"k8s.io/klog/v2"
)

type service interface {
	ResizeNodePool(id string, size int) (*rancher.NodePool, error)
	NodePoolsByCluster(clusterID string) ([]rancher.NodePool, error)
	NodePoolByID(id string) (*rancher.NodePool, error)
	NodesByNodePool(nodePoolID string) ([]rancher.Node, error)
	NodeByProviderID(providerID string) (*rancher.Node, error)
	NodeByNameAndCluster(name, cluster string) (*rancher.Node, error)
	ClusterByID(id string) (*rancher.Cluster, error)
	ScaleDownNode(nodeID string) error
}

var (
	errSecretRequired         = errors.New("secret is required")
	errAccessRequired         = errors.New("access is required")
	errClusterIDRequired      = errors.New("clusterID is required")
	errURLRequired            = errors.New("url is required")
	errClusterIDInvalidFormat = errors.New("clusterID invalid format")
)

type manager struct {
	client    service
	clusterID string
}

// rancherConfig is the configuration of the rancher provider
type rancherConfig struct {
	Global struct {
		// ClusterID is the id associated with the Cluster where rancher
		// Cluster Autoscaler is running.
		ClusterID string `gcfg:"cluster-id"`

		// Secret is the User's Secret  associated with the cluster where
		// rancher Cluster Autoscaler is running.
		Secret string `gcfg:"secret"`

		// Access is the User's Access  associated with the cluster where
		// rancher Cluster Autoscaler is running.
		Access string `gcfg:"access"`

		// URL points to Rancher API.
		URL string `gcfg:"url"`
	}
}

func (c *rancherConfig) validate() error {
	if c.Global.Access == "" {
		return errAccessRequired
	}

	if c.Global.Secret == "" {
		return errSecretRequired
	}

	if c.Global.ClusterID == "" {
		return errClusterIDRequired
	}

	if !strings.HasPrefix(c.Global.ClusterID, "c-") || len(c.Global.ClusterID) != 7 {
		return errClusterIDInvalidFormat
	}

	if c.Global.URL == "" {
		return errURLRequired
	}

	return nil
}

func newManager(configReader io.Reader) (*manager, error) {
	var cfg rancherConfig
	if err := gcfg.ReadInto(&cfg, configReader); err != nil {
		return nil, fmt.Errorf("couldn't read rancher cloud config: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("rancher autoscaler cloudconfig error: %w", err)
	}

	return &manager{
		client:    rancher.New(cfg.Global.URL, fmt.Sprintf("%s:%s", cfg.Global.Access, cfg.Global.Secret)),
		clusterID: cfg.Global.ClusterID,
	}, nil
}

func (m *manager) getNodePools() (map[string]rancher.NodePool, error) {
	nps, err := m.client.NodePoolsByCluster(m.clusterID)
	if err != nil {
		return nil, err
	}

	out := make(map[string]rancher.NodePool, len(nps))
	for _, np := range nps {
		out[np.ID] = np
	}

	return out, nil
}

func (m *manager) getNode(node *apiv1.Node) (*rancher.Node, error) {
	klog.Infof("Trying to get Node by ProviderID %q", node.Spec.ProviderID)
	return m.client.NodeByProviderID(node.Spec.ProviderID)
}
