/*
Copyright 2020 The Kubernetes Authors.

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

package exoscale

import (
	"context"
	"errors"
	"fmt"
	"os"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/exoscale/internal/github.com/exoscale/egoscale"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/exoscale/internal/k8s.io/klog"
)

// Manager handles Exoscale communication and data caching of
// node groups (Instance Pools).
type Manager struct {
	client     *egoscale.Client
	nodeGroups []*NodeGroup
}

func newManager() (*Manager, error) {
	var exoscaleAPIKey, exoscaleAPISecret, exoscaleAPIEndpoint string

	if exoscaleAPIKey == "" {
		exoscaleAPIKey = os.Getenv("EXOSCALE_API_KEY")
	}
	if exoscaleAPISecret == "" {
		exoscaleAPISecret = os.Getenv("EXOSCALE_API_SECRET")
	}
	if exoscaleAPIEndpoint == "" {
		exoscaleAPIEndpoint = os.Getenv("EXOSCALE_API_ENDPOINT")
	}

	if exoscaleAPIKey == "" {
		return nil, errors.New("Exoscale API key is not specified")
	}
	if exoscaleAPISecret == "" {
		return nil, errors.New("Exoscale API secret is not specified")
	}
	if exoscaleAPIEndpoint == "" {
		return nil, errors.New("Exoscale API endpoint is not specified")
	}

	client := egoscale.NewClient(
		exoscaleAPIEndpoint,
		exoscaleAPIKey,
		exoscaleAPISecret,
	)

	m := &Manager{
		client:     client,
		nodeGroups: []*NodeGroup{},
	}

	return m, nil
}

// Refresh refreshes the cache holding the node groups. This is called by the CA
// based on the `--scan-interval`. By default it's 10 seconds.
func (m *Manager) Refresh() error {
	var nodeGroups []*NodeGroup

	for _, ng := range m.nodeGroups {
		_, err := m.client.Request(egoscale.GetInstancePool{
			ID:     ng.instancePool.ID,
			ZoneID: ng.instancePool.ZoneID,
		})
		if csError, ok := err.(*egoscale.ErrorResponse); ok && csError.ErrorCode == egoscale.NotFound {
			klog.V(4).Infof("Removing node group %q", ng.id)
			continue
		} else if err != nil {
			return err
		}

		nodeGroups = append(nodeGroups, ng)
	}

	m.nodeGroups = nodeGroups

	if len(m.nodeGroups) == 0 {
		klog.V(4).Info("cluster-autoscaler is disabled: no node groups found")
	}

	return nil
}

func (m *Manager) computeInstanceLimit() (int, error) {
	limits, err := m.client.ListWithContext(
		context.Background(),
		&egoscale.ResourceLimit{},
	)
	if err != nil {
		return 0, err
	}

	for _, key := range limits {
		limit := key.(*egoscale.ResourceLimit)

		if limit.ResourceTypeName == "user_vm" {
			return int(limit.Max), nil
		}
	}

	return 0, fmt.Errorf(`resource limit "user_vm" not found`)
}
