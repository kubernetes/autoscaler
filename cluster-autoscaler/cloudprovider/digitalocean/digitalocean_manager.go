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
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/oauth2"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/digitalocean/godo"
	"k8s.io/klog"
)

const tagName = "k8s-cluster-autoscaler:"

var (
	version = "dev"
)

type nodeGroupClient interface {
	// GetNodePool retrieves an existing node pool in a Kubernetes cluster.
	GetNodePool(ctx context.Context, clusterID, poolID string) (*godo.KubernetesNodePool, *godo.Response, error)

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
	client         nodeGroupClient
	clusterID      string
	nodeGroups     []*NodeGroup
	nodeGroupsSpec map[string]*nodeSpec
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

func newManager(configReader io.Reader, specs []string) (*Manager, error) {
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

	nodeSpecs, err := parseNodeSpec(specs)
	if err != nil {
		return nil, err
	}

	m := &Manager{
		client:         doClient.Kubernetes,
		clusterID:      cfg.ClusterID,
		nodeGroups:     make([]*NodeGroup, 0),
		nodeGroupsSpec: nodeSpecs,
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

	group := make([]*NodeGroup, len(nodePools))
	for i, np := range nodePools {
		nodePool, resp, err := m.client.GetNodePool(ctx, m.clusterID, np.ID)
		if err != nil {
			if resp != nil && resp.StatusCode == http.StatusNotFound {
				return ErrNodePoolNotExist
			}

			return err
		}

		// default values
		minSize := minNodePoolSize
		maxSize := maxNodePoolSize

		for _, tag := range nodePool.Tags {
			v := strings.TrimPrefix(tag, tagName)
			spec, ok := m.nodeGroupsSpec[v]
			if ok {
				minSize = spec.minSize
				maxSize = spec.maxSize

				klog.V(4).Infof("Found custom spec for node pool: %q (%s) with min: %d and max: %d",
					nodePool.Name, nodePool.ID, minSize, maxSize)
			}
		}

		group[i] = &NodeGroup{
			id:        np.ID,
			clusterID: m.clusterID,
			client:    m.client,
			nodePool:  nodePool,
			minSize:   minSize,
			maxSize:   maxSize,
		}
	}

	m.nodeGroups = group
	return nil
}

// nodeSpec defines a custom specification for a given node
type nodeSpec struct {
	tagValue string
	minSize  int
	maxSize  int
}

// parseNodeSpecs parses a list of specs in the format of 'min,max,tagValue'
func parseNodeSpec(specs []string) (map[string]*nodeSpec, error) {
	nodeSpecs := map[string]*nodeSpec{}

	for _, spec := range specs {
		// format should be "min,max,tagValue"
		// e.g: "3,10,foo"
		splitted := strings.Split(spec, ",")
		if len(splitted) != 3 {
			return nil, fmt.Errorf("spec %q should be in format: 'min,max,tagValue'", spec)
		}

		min, err := strconv.Atoi(splitted[0])
		if err != nil {
			return nil, fmt.Errorf("invalid minimum nodes: %q", splitted[0])
		}

		max, err := strconv.Atoi(splitted[1])
		if err != nil {
			return nil, fmt.Errorf("invalid maximum nodes: %q", splitted[1])
		}

		tagValue := splitted[2]
		if tagValue == "" {
			return nil, errors.New("tag value should be not empty")
		}

		if max == 0 {
			return nil, fmt.Errorf("maximum nodes: %d can't be set to zero", max)
		}

		if min > max {
			return nil, fmt.Errorf("minimum nodes: %d can't be higher than maximum nodes: %d", min, max)
		}

		nodeSpecs[tagValue] = &nodeSpec{
			tagValue: tagValue,
			minSize:  min,
			maxSize:  max,
		}
	}

	return nodeSpecs, nil
}
