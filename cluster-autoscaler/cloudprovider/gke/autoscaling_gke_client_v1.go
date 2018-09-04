/*
Copyright 2018 The Kubernetes Authors.

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

package gke

import (
	"fmt"
	"net/http"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"

	gke_api "google.golang.org/api/container/v1"
)

type autoscalingGkeClientV1 struct {
	gkeService *gke_api.Service

	clusterPath   string
	operationPath string
}

// NewAutoscalingGkeClientV1 creates a new client for communicating with GKE v1 API.
func NewAutoscalingGkeClientV1(client *http.Client, projectId, location, clusterName string) (*autoscalingGkeClientV1, error) {
	autoscalingGkeClient := &autoscalingGkeClientV1{
		clusterPath:   fmt.Sprintf(clusterPathPrefix, projectId, location, clusterName),
		operationPath: fmt.Sprintf(operationPathPrefix, projectId, location),
	}

	gkeService, err := gke_api.New(client)
	if err != nil {
		return nil, err
	}
	if *gkeAPIEndpoint != "" {
		gkeService.BasePath = *gkeAPIEndpoint
	}
	autoscalingGkeClient.gkeService = gkeService

	return autoscalingGkeClient, nil
}

func (m *autoscalingGkeClientV1) GetCluster() (Cluster, error) {
	clusterResponse, err := m.gkeService.Projects.Locations.Clusters.Get(m.clusterPath).Do()
	if err != nil {
		return Cluster{}, err
	}
	nodePools := []NodePool{}
	for _, pool := range clusterResponse.NodePools {
		if pool.Autoscaling != nil && pool.Autoscaling.Enabled {
			nodePools = append(nodePools, NodePool{
				Name:              pool.Name,
				InstanceGroupUrls: pool.InstanceGroupUrls,
				Autoscaled:        pool.Autoscaling.Enabled,
				MinNodeCount:      pool.Autoscaling.MinNodeCount,
				MaxNodeCount:      pool.Autoscaling.MaxNodeCount,
			})
		}
	}
	return Cluster{
		Locations: clusterResponse.Locations,
		NodePools: nodePools,
	}, nil
}

func (m *autoscalingGkeClientV1) DeleteNodePool(toBeRemoved string) error {
	return cloudprovider.ErrNotImplemented
}

func (m *autoscalingGkeClientV1) CreateNodePool(mig *GkeMig) error {
	return cloudprovider.ErrNotImplemented
}
