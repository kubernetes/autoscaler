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

package exoscale

import (
	"context"
	"errors"
	"fmt"
	"os"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	egoscale "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/exoscale/internal/github.com/exoscale/egoscale/v2"
	exoapi "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/exoscale/internal/github.com/exoscale/egoscale/v2/api"
)

type exoscaleClient interface {
	EvictInstancePoolMembers(context.Context, string, *egoscale.InstancePool, []string) error
	EvictSKSNodepoolMembers(context.Context, string, *egoscale.SKSCluster, *egoscale.SKSNodepool, []string) error
	GetInstance(context.Context, string, string) (*egoscale.Instance, error)
	GetInstancePool(context.Context, string, string) (*egoscale.InstancePool, error)
	GetQuota(context.Context, string, string) (*egoscale.Quota, error)
	ListSKSClusters(context.Context, string) ([]*egoscale.SKSCluster, error)
	ScaleInstancePool(context.Context, string, *egoscale.InstancePool, int64) error
	ScaleSKSNodepool(context.Context, string, *egoscale.SKSCluster, *egoscale.SKSNodepool, int64) error
}

const defaultAPIEnvironment = "api"

// Manager handles Exoscale communication and data caching of
// node groups (Instance Pools).
type Manager struct {
	ctx           context.Context
	client        exoscaleClient
	zone          string
	nodeGroups    []cloudprovider.NodeGroup
	discoveryOpts cloudprovider.NodeGroupDiscoveryOptions
}

func newManager(discoveryOpts cloudprovider.NodeGroupDiscoveryOptions) (*Manager, error) {
	var (
		zone           string
		apiKey         string
		apiSecret      string
		apiEnvironment string
		err            error
	)

	if zone = os.Getenv("EXOSCALE_ZONE"); zone == "" {
		return nil, errors.New("no Exoscale zone specified")
	}

	if apiKey = os.Getenv("EXOSCALE_API_KEY"); apiKey == "" {
		return nil, errors.New("no Exoscale API key specified")
	}

	if apiSecret = os.Getenv("EXOSCALE_API_SECRET"); apiSecret == "" {
		return nil, errors.New("no Exoscale API secret specified")
	}

	if apiEnvironment = os.Getenv("EXOSCALE_API_ENVIRONMENT"); apiEnvironment == "" {
		apiEnvironment = defaultAPIEnvironment
	}

	client, err := egoscale.NewClient(apiKey, apiSecret)
	if err != nil {
		return nil, err
	}

	debugf("initializing manager with zone=%s environment=%s", zone, apiEnvironment)

	m := &Manager{
		ctx:           exoapi.WithEndpoint(context.Background(), exoapi.NewReqEndpoint(apiEnvironment, zone)),
		client:        client,
		zone:          zone,
		discoveryOpts: discoveryOpts,
	}

	return m, nil
}

// Refresh refreshes the cache holding the node groups. This is called by the CA
// based on the `--scan-interval`. By default it's 10 seconds.
func (m *Manager) Refresh() error {
	var nodeGroups []cloudprovider.NodeGroup

	// load clusters, it's required for SKS node groups check
	sksClusters, err := m.client.ListSKSClusters(m.ctx, m.zone)
	if err != nil {
		errorf("unable to list SKS clusters: %v", err)
		return err
	}

	for _, ng := range m.nodeGroups {
		// Check SKS Nodepool existence first
		found := false
		for _, c := range sksClusters {
			for _, np := range c.Nodepools {
				if *np.ID == ng.Id() {
					if _, err := m.client.GetInstancePool(m.ctx, m.zone, *np.InstancePoolID); err != nil {
						if !errors.Is(err, exoapi.ErrNotFound) {
							errorf("unable to retrieve SKS Instance Pool %s: %v", ng.Id(), err)
							return err
						}
					} else {
						found = true
						break
					}
				}
			}
		}

		if !found {
			// If SKS Nodepool is not found, check the Instance Pool
			// it was the previous behavior which was less convenient for end user UX
			if _, err := m.client.GetInstancePool(m.ctx, m.zone, ng.Id()); err != nil {
				if !errors.Is(err, exoapi.ErrNotFound) {
					errorf("unable to retrieve SKS Instance Pool %s: %v", ng.Id(), err)
					return err
				}

				// Neither SKS Nodepool nor Instance Pool found, remove it from cache
				debugf("removing node group %s from manager cache", ng.Id())
				continue
			}
		}

		nodeGroups = append(nodeGroups, ng)
	}
	m.nodeGroups = nodeGroups

	if len(m.nodeGroups) == 0 {
		infof("cluster-autoscaler is disabled: no node groups found")
	}

	return nil
}

func (m *Manager) computeInstanceQuota() (int, error) {
	instanceQuota, err := m.client.GetQuota(m.ctx, m.zone, "instance")
	if err != nil {
		return 0, fmt.Errorf("unable to retrieve Compute instances quota: %v", err)
	}

	return int(*instanceQuota.Limit), nil
}
