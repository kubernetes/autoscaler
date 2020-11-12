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
	"strconv"
	"testing"

	"github.com/stretchr/testify/mock"
)

const clusterID = "abcd"

type mockClient struct {
	mock.Mock
}

func (m *mockClient) NewRequest(api string, args map[string]string, out interface{}) (map[string]interface{}, error) {
	m.Called(api, args, out)
	return nil, nil
}

func (m *mockClient) Close() {
	m.Called()
}

func TestGetClusterDetails(t *testing.T) {
	params := map[string]string{
		"id": clusterID,
	}

	s := &mockClient{}
	s.On("NewRequest",
		"listKubernetesClusters", params, &ListClusterResponse{}).Run(func(args mock.Arguments) {
		out := args.Get(2).(*ListClusterResponse)
		out.ClustersResponse = &ClustersResponse{
			Count: 1,
			Clusters: []*Cluster{
				{},
			},
		}
	}).Return().Once()

	service := &cksService{
		client: s,
	}
	service.GetClusterDetails(clusterID)
	s.AssertExpectations(t)
}

func TestScaleCluster(t *testing.T) {
	workerCount := 2
	params := map[string]string{
		"id":   clusterID,
		"size": strconv.Itoa(workerCount),
	}

	s := &mockClient{}
	s.On("NewRequest",
		"scaleKubernetesCluster", params, &ClusterResponse{}).Run(func(args mock.Arguments) {
		out := args.Get(2).(*ClusterResponse)
		out.Cluster = &Cluster{}
	}).Return().Once()

	service := &cksService{
		client: s,
	}
	service.ScaleCluster(clusterID, workerCount)
	s.AssertExpectations(t)
}

func TestRemoveNodesFromCluster(t *testing.T) {
	nodeIDs := "a,b,c"
	params := map[string]string{
		"id":      clusterID,
		"nodeids": nodeIDs,
	}

	s := &mockClient{}
	s.On("NewRequest",
		"scaleKubernetesCluster", params, &ClusterResponse{}).Run(func(args mock.Arguments) {
		out := args.Get(2).(*ClusterResponse)
		out.Cluster = &Cluster{}
	}).Return().Once()

	service := &cksService{
		client: s,
	}
	service.RemoveNodesFromCluster(clusterID, "a", "b", "c")
	s.AssertExpectations(t)
}

func TestCKSClose(t *testing.T) {
	s := &mockClient{}
	s.On("Close").Return().Once()

	service := &cksService{
		client: s,
	}
	service.Close()
	s.AssertExpectations(t)
}
