/*
Copyright 2023 The Kubernetes Authors.

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

package mocks

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/upcloud/pkg/github.com/upcloudltd/upcloud-go-api/v8/upcloud"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/upcloud/pkg/github.com/upcloudltd/upcloud-go-api/v8/upcloud/request"
)

// UpCloudService is mock that implements UpCloudService
type UpCloudService struct {
	Clusters map[string]upcloud.KubernetesCluster
	Plans    []upcloud.KubernetesPlan
	nodes    map[string][]upcloud.KubernetesNode
	mu       sync.Mutex
}

// GetKubernetesNodeGroups list node groups
func (s *UpCloudService) GetKubernetesNodeGroups(ctx context.Context, r *request.GetKubernetesNodeGroupsRequest) ([]upcloud.KubernetesNodeGroup, error) {
	cluster, err := s.GetKubernetesCluster(ctx, &request.GetKubernetesClusterRequest{UUID: r.ClusterUUID})
	if err != nil {
		return nil, err
	}
	return cluster.NodeGroups, nil
}

// ModifyKubernetesNodeGroup modifies the node group
func (s *UpCloudService) ModifyKubernetesNodeGroup(ctx context.Context, r *request.ModifyKubernetesNodeGroupRequest) (*upcloud.KubernetesNodeGroup, error) {
	cluster, err := s.GetKubernetesCluster(ctx, &request.GetKubernetesClusterRequest{UUID: r.ClusterUUID})
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range cluster.NodeGroups {
		if cluster.NodeGroups[i].Name == r.Name {
			cluster.NodeGroups[i].Count = r.NodeGroup.Count
			return &cluster.NodeGroups[i], nil
		}
	}
	return nil, fmt.Errorf("node group not found %+v", r)
}

// DeleteKubernetesNodeGroupNode deletes the node group
func (s *UpCloudService) DeleteKubernetesNodeGroupNode(ctx context.Context, r *request.DeleteKubernetesNodeGroupNodeRequest) error {
	_, err := s.GetKubernetesNodeGroup(ctx, &request.GetKubernetesNodeGroupRequest{ClusterUUID: r.ClusterUUID, Name: r.Name})
	if err != nil {
		return err
	}
	cluster, err := s.GetKubernetesCluster(ctx, &request.GetKubernetesClusterRequest{UUID: r.ClusterUUID})
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	nodes, ok := s.nodes[r.ClusterUUID]
	if !ok {
		return fmt.Errorf("node %s not found", r.NodeName)
	}
	n := make([]upcloud.KubernetesNode, 0)
	for i := range nodes {
		if nodes[i].Name != r.NodeName {
			n = append(n, nodes[i])
		}
	}

	for i := range cluster.NodeGroups {
		if cluster.NodeGroups[i].Name == r.Name {
			cluster.NodeGroups[i].Count--
			break
		}
	}
	s.nodes[r.ClusterUUID] = n
	return nil
}

// GetKubernetesNodeGroup returns node group details
func (s *UpCloudService) GetKubernetesNodeGroup(ctx context.Context, r *request.GetKubernetesNodeGroupRequest) (*upcloud.KubernetesNodeGroupDetails, error) {
	cluster, err := s.GetKubernetesCluster(ctx, &request.GetKubernetesClusterRequest{UUID: r.ClusterUUID})
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.nodes == nil {
		s.nodes = make(map[string][]upcloud.KubernetesNode)
	}
	for i := range cluster.NodeGroups {
		if cluster.NodeGroups[i].Name == r.Name {
			s.nodes[r.ClusterUUID] = s.initNodeGroupNodes(&cluster.NodeGroups[i])
			return &upcloud.KubernetesNodeGroupDetails{
				KubernetesNodeGroup: cluster.NodeGroups[i],
				Nodes:               s.nodes[r.ClusterUUID],
			}, nil
		}
	}
	return nil, fmt.Errorf("node group details not found %+v", r)
}

func (s *UpCloudService) initNodeGroupNodes(nodeGroup *upcloud.KubernetesNodeGroup) []upcloud.KubernetesNode {
	nodes := make([]upcloud.KubernetesNode, nodeGroup.Count)
	for i := 0; i < nodeGroup.Count; i++ {
		nodes[i] = upcloud.KubernetesNode{
			UUID:  fmt.Sprintf("%s-%d", nodeGroup.Name, i),
			Name:  fmt.Sprintf("%s-node-%d", nodeGroup.Name, i),
			State: upcloud.KubernetesNodeStateRunning,
		}
	}
	return nodes
}

// GetKubernetesCluster return UKS cluster object
func (s *UpCloudService) GetKubernetesCluster(_ context.Context, r *request.GetKubernetesClusterRequest) (*upcloud.KubernetesCluster, error) {
	if c, ok := s.Clusters[r.UUID]; ok {
		return &c, nil
	}
	return nil, &upcloud.Problem{Status: http.StatusNotFound}
}

// GetKubernetesPlans list UKS plans
func (s *UpCloudService) GetKubernetesPlans(_ context.Context, _ *request.GetKubernetesPlansRequest) ([]upcloud.KubernetesPlan, error) {
	return s.Plans, nil
}

// AppendNodeGroup is mock helper function to add new node groups during tests
func (s *UpCloudService) AppendNodeGroup(ctx context.Context, clusterID uuid.UUID, group upcloud.KubernetesNodeGroup) error {
	cluster, err := s.GetKubernetesCluster(ctx, &request.GetKubernetesClusterRequest{UUID: clusterID.String()})
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	cluster.NodeGroups = append(cluster.NodeGroups, group)
	s.Clusters[clusterID.String()] = *cluster
	return nil
}
