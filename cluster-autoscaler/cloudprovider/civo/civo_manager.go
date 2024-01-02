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

package civo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"

	civocloud "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/civo/civo-cloud-sdk-go"
	"k8s.io/klog/v2"
)

var (
	// Region is the region where the cluster is located.
	Region string
)

type nodeGroupClient interface {
	// ListKubernetesClusterPools lists all node pools in the Kubernetes cluster.
	ListKubernetesClusterPools(clusterID string) ([]civocloud.KubernetesPool, error)
	// UpdateKubernetesClusterPool updates an existing Kubernetes cluster pool with the Civo API.
	UpdateKubernetesClusterPool(cid, pid string, config *civocloud.KubernetesClusterPoolUpdateConfig) (*civocloud.KubernetesPool, error)
	// DeleteKubernetesClusterPoolInstance deletes a instance from pool
	DeleteKubernetesClusterPoolInstance(clusterID, poolID, instanceID string) (*civocloud.SimpleResponse, error)
	// FindInstanceSizes find instance size
	FindInstanceSizes(size string) (*civocloud.InstanceSize, error)
}

// Manager handles Civo communication and data caching of
// node groups
type Manager struct {
	client        nodeGroupClient
	clusterID     string
	nodeGroups    []*NodeGroup
	discoveryOpts cloudprovider.NodeGroupDiscoveryOptions
}

// Config is the configuration of the Civo cloud provider
type Config struct {
	// ClusterID is the id associated with the cluster where Civo
	// Cluster Autoscaler is running.
	ClusterID string `json:"cluster_id" yaml:"cluster_id"`
	// ApiKey is the Civo User's API Key associated with the cluster where
	// Civo Cluster Autoscaler is running.
	ApiKey string `json:"api_key" yaml:"api_key"`
	// ApiURL is the Civo API URL
	ApiURL string `json:"api_url" yaml:"api_url"`
	// Region is the Civo region
	Region string `json:"region" yaml:"region"`
}

func newManager(configReader io.Reader, discoveryOpts cloudprovider.NodeGroupDiscoveryOptions) (*Manager, error) {
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
	} else {
		cfg.ApiURL = os.Getenv("CIVO_API_URL")
		cfg.ApiKey = os.Getenv("CIVO_API_KEY")
		cfg.ClusterID = os.Getenv("CIVO_CLUSTER_ID")
		cfg.Region = os.Getenv("CIVO_REGION")
	}

	if cfg.ApiURL == "" {
		return nil, errors.New("civo API URL was not provided")
	}
	if cfg.ApiKey == "" {
		return nil, errors.New("civo API Key was not provided")
	}
	if cfg.ClusterID == "" {
		return nil, errors.New("cluster ID was not provided")
	}
	if cfg.Region == "" {
		return nil, errors.New("region was not provided")
	}

	Region = cfg.Region

	civoClient, err := civocloud.NewClientWithURL(cfg.ApiKey, cfg.ApiURL, cfg.Region)
	if err != nil {
		return nil, fmt.Errorf("couldn't initialize Civo client: %s", err)
	}

	m := &Manager{
		client:        civoClient,
		clusterID:     cfg.ClusterID,
		nodeGroups:    make([]*NodeGroup, 0),
		discoveryOpts: discoveryOpts,
	}

	return m, nil
}

// Refresh refreshes the cache holding the nodegroups. This is called by the CA
// based on the `--scan-interval`. By default it's 10 seconds.
func (m *Manager) Refresh() error {
	var (
		minSize           int
		maxSize           int
		workerConfigFound = false
		poolConfigFound   = false
		poolGroups        []*NodeGroup
		workerGroups      []*NodeGroup
	)

	pools, err := m.client.ListKubernetesClusterPools(m.clusterID)
	if err != nil {
		return fmt.Errorf("couldn't list Kubernetes cluster pools: %s", err)
	}

	klog.V(4).Infof("refreshing workers node group kubernetes cluster: %q", m.clusterID)

	for _, specString := range m.discoveryOpts.NodeGroupSpecs {
		spec, err := dynamic.SpecFromString(specString, true)
		if err != nil {
			return fmt.Errorf("failed to parse node group spec: %v", err)
		}

		if spec.Name == "workers" {
			minSize = spec.MinSize
			maxSize = spec.MaxSize
			workerConfigFound = true
			klog.V(4).Infof("found configuration for workers node group: min: %d max: %d", minSize, maxSize)
		} else {
			poolConfigFound = true
			nodeGroup := m.getNodeGroupConfig(spec, pools)
			if nodeGroup != nil {
				poolGroups = append(poolGroups, nodeGroup)
				klog.V(4).Infof("found configuration for pool node group: min: %d max: %d", nodeGroup.minSize, nodeGroup.maxSize)
			}
		}
	}

	if poolConfigFound {
		m.nodeGroups = poolGroups
	} else if workerConfigFound {
		for _, nodePool := range pools {
			np := nodePool
			klog.V(4).Infof("adding node pool: %q", nodePool.ID)

			workerGroups = append(workerGroups, &NodeGroup{
				id:           nodePool.ID,
				clusterID:    m.clusterID,
				client:       m.client,
				nodePool:     &np,
				minSize:      minSize,
				maxSize:      maxSize,
				nodeTemplate: getCivoNodeTemplate(nodePool, m.client),
			})
		}
		m.nodeGroups = workerGroups
	} else {
		return fmt.Errorf("no workers node group configuration found")
	}

	// If both config found, pool config get precedence
	if poolConfigFound && workerConfigFound {
		m.nodeGroups = poolGroups
	}

	if len(m.nodeGroups) == 0 {
		klog.V(4).Info("cluster-autoscaler is disabled. no node pools are configured")
	}

	return nil
}

// getNodeGroupConfig get the node group configuration from the cluster pool configuration
func (m *Manager) getNodeGroupConfig(spec *dynamic.NodeGroupSpec, pools []civocloud.KubernetesPool) *NodeGroup {
	for _, nodePool := range pools {
		if spec.Name == nodePool.ID {
			np := nodePool
			klog.V(4).Infof("adding node pool: %q min: %d max: %d", nodePool.ID, spec.MinSize, spec.MaxSize)

			return &NodeGroup{
				id:           nodePool.ID,
				clusterID:    m.clusterID,
				client:       m.client,
				nodePool:     &np,
				minSize:      spec.MinSize,
				maxSize:      spec.MaxSize,
				nodeTemplate: getCivoNodeTemplate(nodePool, m.client),
			}
		}
	}
	return nil
}

// getCivoNodeTemplate returns the CivoNodeTemplate for the given node pool
func getCivoNodeTemplate(pool civocloud.KubernetesPool, client nodeGroupClient) *CivoNodeTemplate {
	template := &CivoNodeTemplate{}
	size, err := client.FindInstanceSizes(pool.Size)
	if err != nil {
		klog.V(4).ErrorS(err, "Failed to get size")
		template.Size = pool.Size
		template.CPUCores = 2
		template.RAMMegabytes = 1024 * 1024
		template.DiskGigabytes = 1024 * 1024
		template.Labels["kubernetes.civo.com/civo-node-pool"] = pool.ID
		return template
	}

	template.Size = pool.Size
	template.CPUCores = size.CPUCores
	template.RAMMegabytes = size.RAMMegabytes
	template.DiskGigabytes = size.DiskGigabytes
	template.Labels = pool.Labels
	template.Region = pool.Region
	template.Taints = pool.Taints
	template.GpuCount = size.GPUCount

	return template
}
