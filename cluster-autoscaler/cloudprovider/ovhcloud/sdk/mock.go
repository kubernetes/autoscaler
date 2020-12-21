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

// ListNodePools mocks API call for listing node pool in cluster
func (m *ClientMock) ListNodePools(ctx context.Context, projectID string, clusterID string) ([]NodePool, error) {
	args := m.Called(ctx, projectID, clusterID)

	return args.Get(0).([]NodePool), args.Error(1)
}

// ListNodePoolNodes mocks API call for listing node in a pool
func (m *ClientMock) ListNodePoolNodes(ctx context.Context, projectID string, clusterID string, poolID string) ([]Node, error) {
	args := m.Called(ctx, projectID, clusterID, poolID)

	return args.Get(0).([]Node), args.Error(1)
}

// CreateNodePool mocks API call for creating a new pool
func (m *ClientMock) CreateNodePool(ctx context.Context, projectID string, clusterID string, opts *CreateNodePoolOpts) (*NodePool, error) {
	args := m.Called(ctx, projectID, clusterID, opts)

	return args.Get(0).(*NodePool), args.Error(1)
}

// UpdateNodePool mocks API call to update size of a pool
func (m *ClientMock) UpdateNodePool(ctx context.Context, projectID string, clusterID string, poolID string, opts *UpdateNodePoolOpts) (*NodePool, error) {
	args := m.Called(ctx, projectID, clusterID, poolID, opts)

	return args.Get(0).(*NodePool), args.Error(1)
}

// DeleteNodePool mocks API call to delete an existing pool
func (m *ClientMock) DeleteNodePool(ctx context.Context, projectID string, clusterID string, poolID string) (*NodePool, error) {
	args := m.Called(ctx, projectID, clusterID, poolID)

	return args.Get(0).(*NodePool), args.Error(1)
}

// ListFlavors mocks API call for listing available flavors in cluster
func (m *ClientMock) ListFlavors(ctx context.Context, projectID string, clusterID string) ([]Flavor, error) {
	args := m.Called(ctx, projectID, clusterID)

	return args.Get(0).([]Flavor), args.Error(1)
}
