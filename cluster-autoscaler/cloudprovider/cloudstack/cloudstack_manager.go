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

package cloudstack

import (
	"sync"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/cloudstack/service"

	v1 "k8s.io/api/core/v1"
	klog "k8s.io/klog/v2"
)

type manager struct {
	asg           *asg
	mux           sync.Mutex
	service       service.CKSService
	clusterConfig *clusterConfig
}

type clusterConfig struct {
	clusterID string
	minSize   int
	maxSize   int
}

// CSConfig wraps the config for the CloudStack cloud provider.
type CSConfig struct {
	Global struct {
		APIURL      string `gcfg:"api-url"`
		APIKey      string `gcfg:"api-key"`
		SecretKey   string `gcfg:"secret-key"`
		SSLNoVerify bool   `gcfg:"ssl-no-verify"`
		ProjectID   string `gcfg:"project-id"`
		Zone        string `gcfg:"zone"`
	}
}

func (manager *manager) clusterForNode(node *v1.Node) (*asg, error) {
	_, err := manager.asg.Belongs(node)
	if err != nil {
		return nil, err
	}
	return manager.asg, nil
}

func (manager *manager) refresh() error {
	return manager.fetchCluster()
}

func (manager *manager) cleanup() error {
	manager.service.Close()
	manager.mux.Lock()
	defer manager.mux.Unlock()
	return nil
}

func (manager *manager) setMinMaxIfNotPresent(cluster *service.Cluster) {
	if cluster.Minsize == 0 || cluster.Maxsize == 0 {
		cluster.Minsize = manager.clusterConfig.minSize
		cluster.Maxsize = manager.clusterConfig.maxSize
	}
}

func (manager *manager) fetchCluster() error {
	manager.mux.Lock()
	defer manager.mux.Unlock()

	cluster, err := manager.service.GetClusterDetails(manager.clusterConfig.clusterID)
	if err != nil {
		return err
	}

	klog.Info("Got cluster : ", cluster)
	manager.setMinMaxIfNotPresent(cluster)
	manager.asg.Copy(cluster)
	return nil
}

func (manager *manager) scaleCluster(clusterID string, workerCount int) (*service.Cluster, error) {
	manager.mux.Lock()
	defer manager.mux.Unlock()

	cluster, err := manager.service.ScaleCluster(clusterID, workerCount)
	if err != nil {
		return nil, err
	}

	klog.Info("Scaled up cluster : ", cluster)
	manager.setMinMaxIfNotPresent(cluster)
	return cluster, nil
}

func (manager *manager) removeNodesFromCluster(clusterID string, nodeIDs ...string) (*service.Cluster, error) {
	manager.mux.Lock()
	defer manager.mux.Unlock()

	cluster, err := manager.service.RemoveNodesFromCluster(clusterID, nodeIDs...)
	if err != nil {
		return nil, err
	}

	klog.Info("Scaled down cluster : ", cluster)
	manager.setMinMaxIfNotPresent(cluster)
	return cluster, nil
}

func newManager(clusterConfig *clusterConfig, opts ...option) (*manager, error) {
	cfg, err := createConfig(opts...)
	if err != nil {
		return nil, err
	}

	cfg.asg.cluster = &service.Cluster{
		ID:      clusterConfig.clusterID,
		Minsize: clusterConfig.minSize,
		Maxsize: clusterConfig.maxSize,
	}

	manager := &manager{
		asg:           cfg.asg,
		service:       cfg.service,
		clusterConfig: clusterConfig,
	}
	cfg.asg.manager = manager
	manager.refresh()
	return manager, nil
}
