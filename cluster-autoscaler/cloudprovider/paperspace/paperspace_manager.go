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

package paperspace

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"

	psgo "github.com/Paperspace/paperspace-go"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/klog"
)

type nodeGroupClient interface {
	// GetNodePools lists all the node pools found in a Kubernetes cluster.
	GetAutoscalingGroups(params psgo.AutoscalingGroupListParams) ([]psgo.AutoscalingGroup, error)

	// UpdateNodePool updates the details of an existing node pool.
	UpdateAutoscalingGroup(id string, params psgo.AutoscalingGroupUpdateParams) error

	// DeleteNode deletes a specific node in a node pool.
	DeleteMachine(id string, params psgo.MachineDeleteParams) error
}

var _ nodeGroupClient = (*psgo.Client)(nil)

// Manager handles Paperspace communication and data caching of
// node groups (node pools in DOKS)
type Manager struct {
	client     nodeGroupClient
	clusterID  string
	nodeGroups []*NodeGroup
}

// Config is the configuration of the Paperspace cloud provider
type Config struct {
	// ClusterID is the id associated with the cluster where Paperspace
	// Cluster Autoscaler is running.
	ClusterID string `json:"clusterId"`

	// Token is the User's Access Token associated with the cluster where
	// Paperspace Cluster Autoscaler is running.
	APIKey string `json:"apiToken"`

	// URL points to Paperspace API. If empty, defaults to
	// https://api.paperspace.com/
	URL string `json:"url"`
}

func newManager(configReader io.Reader, do cloudprovider.NodeGroupDiscoveryOptions, instanceTypes map[string]string) (*Manager, error) {
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

	// todo process config
	if cfg.APIKey == "" {
		return nil, errors.New("access token is not provided")
	}
	if cfg.ClusterID == "" {
		return nil, errors.New("cluster ID is not provided")
	}

	apiBackend := psgo.NewAPIBackend()
	if cfg.URL != "" {
		apiBackend.BaseURL = cfg.URL
	}

	client := psgo.NewClientWithBackend(apiBackend)
	client.APIKey = cfg.APIKey

	//specs, err := do.ParseASGAutoDiscoverySpecs()
	//if err != nil {

	m := &Manager{
		client:     client,
		clusterID:  cfg.ClusterID,
		nodeGroups: make([]*NodeGroup, 0),
	}

	return m, nil
}

// Refresh refreshes the cache holding the nodegroups. This is called by the CA
// based on the `--scan-interval`. By default it's 10 seconds.
func (m *Manager) Refresh() error {
	ctx := context.Background()
	req := psgo.AutoscalingGroupListParams{
		RequestParams: psgo.RequestParams{Context: ctx},
		Filter:        nil,
	}
	autoscalingGroups, err := m.client.GetAutoscalingGroups(req)
	if err != nil {
		return err
	}

	var groups []*NodeGroup
	for _, asg := range autoscalingGroups {
		klog.V(4).Infof("adding node pool: %q name: %s min: %d max: %d",
			asg.ID, asg.Name, asg.Min, asg.Max)

		groups = append(groups, &NodeGroup{
			id:        asg.ID,
			clusterID: m.clusterID,
			client:    m.client,
			asg:       asg,
			minSize:   asg.Min,
			maxSize:   asg.Max,
		})
	}

	if len(groups) == 0 {
		klog.V(4).Info("cluster-autoscaler is disabled. no node pools are configured")
	}

	m.nodeGroups = groups
	return nil
}
