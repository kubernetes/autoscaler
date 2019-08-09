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

package digitalocean

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"

	"golang.org/x/oauth2"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/digitalocean/godo"
	"k8s.io/klog"
)

const (
	tagPrefix  = "k8s-cluster-autoscaler-"
	tagEnabled = tagPrefix + "enabled:"
	tagMin     = tagPrefix + "min:"
	tagMax     = tagPrefix + "max:"
)

var (
	version = "dev"
)

type nodeGroupClient interface {
	// ListNodePools lists all the node pools found in a Kubernetes cluster.
	ListNodePools(ctx context.Context, clusterID string, opts *godo.ListOptions) ([]*godo.KubernetesNodePool, *godo.Response, error)

	// UpdateNodePool updates the details of an existing node pool.
	UpdateNodePool(ctx context.Context, clusterID, poolID string, req *godo.KubernetesNodePoolUpdateRequest) (*godo.KubernetesNodePool, *godo.Response, error)

	// DeleteNode deletes a specific node in a node pool.
	DeleteNode(ctx context.Context, clusterID, poolID, nodeID string, req *godo.KubernetesNodeDeleteRequest) (*godo.Response, error)
}

// Manager handles DigitalOcean communication and data caching of
// node groups (node pools in DOKS)
type Manager struct {
	client     nodeGroupClient
	clusterID  string
	nodeGroups []*NodeGroup
}

// Config is the configuration of the DigitalOcean cloud provider
type Config struct {
	// ClusterID is the id associated with the cluster where DigitalOcean
	// Cluster Autoscaler is running.
	ClusterID string `json:"cluster_id"`

	// Token is the User's Access Token associated with the cluster where
	// DigitalOcean Cluster Autoscaler is running.
	Token string `json:"token"`

	// URL points to DigitalOcean API. If empty, defaults to
	// https://api.digitalocean.com/
	URL string `json:"url"`
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

	if cfg.Token == "" {
		return nil, errors.New("access token is not provided")
	}
	if cfg.ClusterID == "" {
		return nil, errors.New("cluster ID is not provided")
	}

	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: cfg.Token,
	})
	oauthClient := oauth2.NewClient(context.Background(), tokenSource)

	opts := []godo.ClientOpt{}
	if cfg.URL != "" {
		opts = append(opts, godo.SetBaseURL(cfg.URL))
	}

	opts = append(opts, godo.SetUserAgent("cluster-autoscaler-digitalocean/"+version))

	doClient, err := godo.New(oauthClient, opts...)
	if err != nil {
		return nil, fmt.Errorf("couldn't initialize DigitalOcean client: %s", err)
	}

	m := &Manager{
		client:     doClient.Kubernetes,
		clusterID:  cfg.ClusterID,
		nodeGroups: make([]*NodeGroup, 0),
	}

	return m, nil
}

// Refresh refreshes the cache holding the nodegroups. This is called by the CA
// based on the `--scan-interval`. By default it's 10 seconds.
func (m *Manager) Refresh() error {
	ctx := context.Background()
	nodePools, _, err := m.client.ListNodePools(ctx, m.clusterID, nil)
	if err != nil {
		return err
	}

	var group []*NodeGroup
	for _, nodePool := range nodePools {
		spec, err := parseTags(nodePool.Tags)
		if err != nil {
			// we should not return an error here, because one misconfigured
			// node pool shouldn't bring down the whole cluster-autoscaler
			klog.V(4).Infof("skipping misconfigured node pool: %q name: %s tags: %+v err: %s",
				nodePool.ID, nodePool.Name, nodePool.Tags, err)
			continue
		}

		if !spec.enabled {
			continue
		}

		minSize := minNodePoolSize
		if spec.min != 0 {
			minSize = spec.min
		}

		maxSize := maxNodePoolSize
		if spec.max != 0 {
			maxSize = spec.max
		}

		klog.V(4).Infof("adding node pool: %q name: %s min: %d max: %d",
			nodePool.ID, nodePool.Name, minSize, maxSize)

		group = append(group, &NodeGroup{
			id:        nodePool.ID,
			clusterID: m.clusterID,
			client:    m.client,
			nodePool:  nodePool,
			minSize:   minSize,
			maxSize:   maxSize,
		})
	}

	if len(group) == 0 {
		klog.V(4).Info("cluster-autoscaler is disabled. no node pools are configured")
	}

	m.nodeGroups = group
	return nil
}

// nodeSpec defines a custom specification for a given node
type nodeSpec struct {
	min     int
	max     int
	enabled bool
}

// parseTags parses a list of tags from a DigitalOcean node pool
func parseTags(tags []string) (*nodeSpec, error) {
	spec := &nodeSpec{}

	for _, tag := range tags {
		if !strings.HasPrefix(tag, tagPrefix) {
			continue
		}

		splitted := strings.Split(strings.TrimPrefix(tag, tagPrefix), ":")
		if len(splitted) != 2 {
			return nil, fmt.Errorf("malformed tag: %q", tag)
		}

		key, value := splitted[0], splitted[1]

		switch key {
		case "enabled":
			if value == "true" {
				spec.enabled = true
			}
		case "min":
			min, err := strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("invalid minimum nodes: %q", value)
			}

			if min <= 0 {
				return nil, fmt.Errorf("minimum nodes: %d can't be set to zero or less", min)
			}

			spec.min = min
		case "max":
			max, err := strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("invalid maximum nodes: %q", value)
			}

			if max <= 0 {
				return nil, fmt.Errorf("maximum nodes: %d can't be set to zero or less", max)
			}

			spec.max = max
		}
	}

	if spec.min != 0 && spec.max != 0 && spec.min > spec.max {
		return nil, fmt.Errorf("minimum nodes: %d can't be higher than maximum nodes: %d",
			spec.min, spec.max)
	}

	return spec, nil
}
