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
	"errors"
	"os"
	"strconv"

	"github.com/linode/linodego"
	"github.com/stretchr/testify/mock"
)

type linodeClientMock struct {
	mock.Mock
}

func (l *linodeClientMock) ListLKEClusterPools(ctx context.Context, clusterID int, opts *linodego.ListOptions) ([]linodego.LKEClusterPool, error) {
	args := l.Called(ctx, clusterID, nil)
	return args.Get(0).([]linodego.LKEClusterPool), args.Error(1)
}

func (l *linodeClientMock) DeleteLKEClusterPoolNode(ctx context.Context, clusterID int, id string) error {
	args := l.Called(ctx, clusterID, id)
	return args.Error(0)
}

func (l *linodeClientMock) UpdateLKEClusterPool(ctx context.Context, clusterID, id int, updateOpts linodego.LKEClusterPoolUpdateOptions) (*linodego.LKEClusterPool, error) {
	args := l.Called(ctx, clusterID, id, updateOpts)
	return args.Get(0).(*linodego.LKEClusterPool), args.Error(1)
}

type readerErrMock struct{}

func (readerErrMock) Read(p []byte) (n int, err error) {
	return 0, errors.New("mock error")
}

func makeMockNodePool(id int, linodes []linodego.LKEClusterPoolLinode, autoscaler linodego.LKEClusterPoolAutoscaler) linodego.LKEClusterPool {
	return linodego.LKEClusterPool{
		ID:         id,
		Type:       "g6-standard-1",
		Count:      len(linodes),
		Linodes:    linodes,
		Autoscaler: autoscaler,
	}
}

func makeTestNodePoolNode(id int) linodego.LKEClusterPoolLinode {
	return linodego.LKEClusterPoolLinode{
		ID:         strconv.Itoa(id),
		InstanceID: id,
	}
}

func makeTestNodePoolNodes(lower, upper int) []linodego.LKEClusterPoolLinode {
	linodes := make([]linodego.LKEClusterPoolLinode, upper-lower+1)
	for i := 0; lower+i <= upper; i++ {
		linodes[i] = makeTestNodePoolNode(lower + i)
	}
	return linodes
}

func testEnvVar(name, value string) func() {
	originalValue := os.Getenv(name)
	os.Setenv(name, value)
	return func() {
		os.Setenv(name, originalValue)
	}
}
