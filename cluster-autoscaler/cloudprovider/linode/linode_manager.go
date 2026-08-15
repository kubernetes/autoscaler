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
	"fmt"
	"io"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/linode/linodego"
	klog "k8s.io/klog/v2"
)

// manager handles Linode communication and holds information about
// the node groups (LKE pools with a single linode each)
type manager struct {
	client     linodeAPIClient
	config     *linodeConfig
	nodeGroups map[string]*NodeGroup // key: NodeGroup.id
}

func newManager(config io.Reader) (*manager, error) {
	cfg, err := buildCloudConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %v", err)
	}
	client := buildLinodeAPIClient(cfg.token)
	m := &manager{
		client:     client,
		config:     cfg,
		nodeGroups: make(map[string]*NodeGroup),
	}
	return m, nil
}

func (m *manager) refresh() error {
	nodeGroups := make(map[string]*NodeGroup)
	lkeClusterPools, err := m.client.ListLKEClusterPools(context.Background(), m.config.clusterID, nil)
	if err != nil {
		return fmt.Errorf("failed to get list of LKE pools from linode API: %v", err)
	}

	for i, pool := range lkeClusterPools {
		// skip this pool if it is among the ones to be excluded as defined in the config
		_, found := m.config.excludedPoolIDs[pool.ID]
		if found {
			continue
		}
		// check if the nodes in the pool are more than 1, if so skip it
		if pool.Count > 1 {
			klog.V(2).Infof("The LKE pool %d has more than one node (current nodes in pool: %d), will exclude it from the node groups",
				pool.ID, pool.Count)
			continue
		}
		// add the LKE pool to the node groups map
		linodeType := pool.Type
		ng, found := nodeGroups[linodeType]
		if found {
			// if a node group for the node type of this pool already exists, add it to the related node group
			// TODO if node group size is exceeded better to skip it or add it anyway? here we are adding it
			ng.lkePools[pool.ID] = &lkeClusterPools[i]
		} else {
			// create a new node group with this pool in it
			ng := buildNodeGroup(&lkeClusterPools[i], m.config, m.client)
			nodeGroups[linodeType] = ng
		}
	}

	// show some debug info
	klog.V(2).Infof("LKE node group after refresh:")
	for _, ng := range nodeGroups {
		klog.V(2).Infof("%s", ng.extendedDebug())
	}
	for _, ng := range nodeGroups {
		currentSize := len(ng.lkePools)
		if currentSize > ng.maxSize {
			klog.V(2).Infof("imported node pools in node group %q are > maxSize (current size: %d, min size: %d, max size: %d)",
				ng.id, currentSize, ng.minSize, ng.maxSize)
		}
		if currentSize < ng.minSize {
			klog.V(2).Infof("imported node pools in node group %q are < minSize (current size: %d, min size: %d, max size: %d)",
				ng.id, currentSize, ng.minSize, ng.maxSize)
		}
	}

	m.nodeGroups = nodeGroups
	return nil
}

func buildNodeGroup(pool *linodego.LKEClusterPool, cfg *linodeConfig, client linodeAPIClient) *NodeGroup {
	// get specific min and max size for a node group, if defined in the config
	minSize := cfg.defaultMinSize
	maxSize := cfg.defaultMaxSize
	nodeGroupCfg, found := cfg.nodeGroupCfg[pool.Type]
	if found {
		minSize = nodeGroupCfg.minSize
		maxSize = nodeGroupCfg.maxSize
	}
	// create the new node group with this single LKE pool inside
	lkePools := make(map[int]*linodego.LKEClusterPool)
	lkePools[pool.ID] = pool
	poolOpts := linodego.LKEClusterPoolCreateOptions{
		Count: 1,
		Type:  pool.Type,
		Disks: pool.Disks,
	}
	ng := &NodeGroup{
		client:       client,
		lkePools:     lkePools,
		poolOpts:     poolOpts,
		lkeClusterID: cfg.clusterID,
		minSize:      minSize,
		maxSize:      maxSize,
		id:           pool.Type,
	}
	return ng
}
