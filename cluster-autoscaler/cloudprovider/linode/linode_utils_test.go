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

	"github.com/stretchr/testify/mock"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/linode/linodego"
)

type linodeClientMock struct {
	mock.Mock
}

func (l *linodeClientMock) ListLKEClusterPools(ctx context.Context, clusterID int, opts *linodego.ListOptions) ([]linodego.LKEClusterPool, error) {
	args := l.Called(ctx, clusterID, nil)
	return args.Get(0).([]linodego.LKEClusterPool), args.Error(1)
}

func (l *linodeClientMock) CreateLKEClusterPool(ctx context.Context, clusterID int, createOpts linodego.LKEClusterPoolCreateOptions) (*linodego.LKEClusterPool, error) {
	args := l.Called(ctx, clusterID, createOpts)
	return args.Get(0).(*linodego.LKEClusterPool), args.Error(1)
}

func (l *linodeClientMock) DeleteLKEClusterPool(ctx context.Context, clusterID int, id int) error {
	args := l.Called(ctx, clusterID, id)
	return args.Error(0)
}
