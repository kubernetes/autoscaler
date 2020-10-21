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

package hetzner

import (
	"context"
	"errors"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"k8s.io/klog/v2"
	"os"
)

var (
	version = "dev"
)

// Manager handles Hetzner communication and data caching of
// node groups
type Manager struct {
	client *hcloud.Client
	nodeGroup *NodeGroup
}

const nodePoolId = "hcloud-node-pool"

func newManager() (*Manager, error) {
	token := os.Getenv("HCLOUD_TOKEN")
	if token == "" {
		return nil, errors.New("`HCLOUD_TOKEN` is not specified")
	}

	client := hcloud.NewClient(hcloud.WithToken(token))

	m := &Manager{
		client:    client,
		nodeGroup: nil,
	}
	m.nodeGroup = &NodeGroup{
		id:  nodePoolId,
		size: 0,
		manager: m,
	}

	return m, nil
}

// Refresh refreshes the cache holding the nodegroups. This is called by the CA
// based on the `--scan-interval`. By default it's 10 seconds.
func (m *Manager) Refresh() error {
	return nil
}

func (m *Manager) allServers() []*hcloud.Server  {
	listOptions := hcloud.ListOpts{
		PerPage: 50,
		LabelSelector: nodeIDLabel,
	}
	ctx := context.Background()
	requestOptions := hcloud.ServerListOpts{ListOpts: listOptions}
	servers, err := m.client.Server.AllWithOpts(ctx, requestOptions)
	if err != nil {
		klog.Fatalf("Failed to get servers for Hcloud: %v", err)
	}

	return servers
}
