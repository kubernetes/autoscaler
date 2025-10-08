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

package service

import (
	"fmt"
	"strconv"
	"strings"
)

// CKSService provides all the functionality of CloudStack Kubernetes Service
type CKSService interface {
	// GetClusterDetails returns the details of the CKS Cluster
	GetClusterDetails(clusterID string) (*Cluster, error)

	// ScaleCluster scales up / down the worker nodes in a cluster
	ScaleCluster(clusterID string, workerCount int) (*Cluster, error)

	// RemoveNodesFromCluster removes the given nodes from a cluster. However all masters can not be removed
	RemoveNodesFromCluster(clusterID string, nodeIDs ...string) (*Cluster, error)

	// Close terminates the service
	Close()
}

// ListClusterResponse is the response returned by the listKubernetesClusters API
type ListClusterResponse struct {
	ClustersResponse *ClustersResponse `json:"listkubernetesclustersresponse"`
}

// ClustersResponse contains the CKS Clusters and their total count
type ClustersResponse struct {
	Count    int        `json:"count"`
	Clusters []*Cluster `json:"kubernetescluster"`
}

// ClusterResponse is the response returned by the scaleKubernetesCluster API
type ClusterResponse struct {
	Cluster *Cluster `json:"kubernetescluster"`
}

// Cluster contains the CKS Cluster details
type Cluster struct {
	ID                string            `json:"id"`
	Name              string            `json:"name"`
	Minsize           int               `json:"minsize"`
	Maxsize           int               `json:"maxsize"`
	WorkerCount       int               `json:"size"`
	MasterCount       int               `json:"masternodes"`
	VirtualMachines   []*VirtualMachine `json:"virtualmachines"`
	VirtualMachineMap map[string]*VirtualMachine
}

// VirtualMachine represents a node in a CKS cluster
type VirtualMachine struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	State string `json:"state"`
}

// cksService implements the CKSService interface
type cksService struct {
	client    APIClient
	projectID string
}

func virtaulMachinesToMap(vms []*VirtualMachine) map[string]*VirtualMachine {
	vmMap := make(map[string]*VirtualMachine)
	for _, vm := range vms {
		if vm.Name != "" {
			vmMap[vm.Name] = vm
		}
	}
	return vmMap
}

func (service *cksService) GetClusterDetails(clusterID string) (*Cluster, error) {
	var out ListClusterResponse
	params := map[string]string{
		"id": clusterID,
	}
	
	// Include projectid if configured to support multi-project CloudStack environments
	if service.projectID != "" {
		params["projectid"] = service.projectID
	}
	
	_, err := service.client.NewRequest("listKubernetesClusters", params, &out)

	if err != nil {
		return nil, fmt.Errorf("Unable to fetch cluster details : %v", err)
	}

	clusters := out.ClustersResponse.Clusters
	if len(clusters) == 0 {
		return nil, fmt.Errorf("Unable to fetch cluster with id : %v", clusterID)
	}
	cluster := clusters[0]
	cluster.VirtualMachineMap = virtaulMachinesToMap(cluster.VirtualMachines)
	return cluster, err
}

func (service *cksService) ScaleCluster(clusterID string, workerCount int) (*Cluster, error) {
	var out ClusterResponse
	params := map[string]string{
		"id":   clusterID,
		"size": strconv.Itoa(workerCount),
	}
	
	// Include projectid if configured to support multi-project CloudStack environments
	if service.projectID != "" {
		params["projectid"] = service.projectID
	}
	
	_, err := service.client.NewRequest("scaleKubernetesCluster", params, &out)

	if err != nil {
		return nil, fmt.Errorf("Unable to scale cluster : %v", err)
	}
	cluster := out.Cluster
	cluster.VirtualMachineMap = virtaulMachinesToMap(cluster.VirtualMachines)
	return cluster, err
}

func (service *cksService) RemoveNodesFromCluster(clusterID string, nodeIDs ...string) (*Cluster, error) {
	var out ClusterResponse
	params := map[string]string{
		"id":      clusterID,
		"nodeids": strings.Join(nodeIDs[:], ","),
	}
	
	// Include projectid if configured to support multi-project CloudStack environments
	if service.projectID != "" {
		params["projectid"] = service.projectID
	}
	
	_, err := service.client.NewRequest("scaleKubernetesCluster", params, &out)
	if err != nil {
		return nil, fmt.Errorf("Unable to delete %v from cluster : %v", nodeIDs, err)
	}
	cluster := out.Cluster
	cluster.VirtualMachineMap = virtaulMachinesToMap(cluster.VirtualMachines)
	return cluster, err
}

func (service *cksService) Close() {
	service.client.Close()
}

// NewCKSService returns a new CKS Service
func NewCKSService(config *Config) CKSService {
	client := NewAPIClient(config)
	return &cksService{
		client:    client,
		projectID: config.ProjectID,
	}
}
