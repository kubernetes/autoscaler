/*
Copyright 2025 The Kubernetes Authors.

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

package utho

import (
	"context"

	"github.com/stretchr/testify/mock"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/utho/utho-go"
)

type uthoClientMock struct {
	mock.Mock
}

func (u *uthoClientMock) ReadNodePool(ctx context.Context, clusterId int, nodePoolId string) (*utho.NodepoolDetails, error) {
	args := u.Called(ctx, clusterId, nodePoolId)
	return args.Get(0).(*utho.NodepoolDetails), args.Error(1)
}

func (u *uthoClientMock) ListNodePools(ctx context.Context, clusterID string) ([]utho.NodepoolDetails, error) {
	args := u.Called(ctx, clusterID, nil)
	return args.Get(0).([]utho.NodepoolDetails), args.Error(2)
}

func (u *uthoClientMock) UpdateNodePool(ctx context.Context, params utho.UpdateKubernetesAutoscaleNodepool) (*utho.UpdateKubernetesAutoscaleNodepoolResponse, error) {
	args := u.Called(ctx, params)
	return args.Get(0).(*utho.UpdateKubernetesAutoscaleNodepoolResponse), args.Error(1)
}

func (u *uthoClientMock) DeleteNode(ctx context.Context, params utho.DeleteNodeParams) (*utho.DeleteResponse, error) {
	args := u.Called(ctx, params)
	return args.Get(0).(*utho.DeleteResponse), args.Error(1)
}
