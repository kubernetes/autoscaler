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

package sdk

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// ClientMock mocks the API client
type ClientMock struct {
	mock.Mock
}

// ListNodePools mocks API call for listing node pools in a cluster
func (m *ClientMock) ListNodePools(ctx context.Context, clusterID string) ([]NodePool, error) {
	args := m.Called(ctx, clusterID)
	return args.Get(0).([]NodePool), args.Error(1)
}

// ListNodePoolNodes mocks API call for listing nodes in a pool
func (m *ClientMock) ListNodePoolNodes(ctx context.Context, clusterID string, poolID string) ([]Node, error) {
	args := m.Called(ctx, clusterID, poolID)
	return args.Get(0).([]Node), args.Error(1)
}

// DeleteNode mocks API call to delete a node
func (m *ClientMock) DeleteNode(ctx context.Context, clusterID string, nodeGroupID string, id string) error {
	args := m.Called(ctx, clusterID, nodeGroupID, id)
	return args.Error(0)
}

// ListClusterFlavors mocks API call for listing available flavors in a cluster
func (m *ClientMock) ListClusterFlavors(ctx context.Context, clusterID string) ([]Flavor, error) {
	args := m.Called(ctx, clusterID)
	return args.Get(0).([]Flavor), args.Error(1)
}

// AddNode mocks API call for adding a node to a node pool
func (m *ClientMock) AddNode(ctx context.Context, clusterID string, nodeGroupID string) (*Node, error) {
	args := m.Called(ctx, clusterID, nodeGroupID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Node), args.Error(1)
}
