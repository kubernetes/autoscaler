/*
Copyright 2021 The Kubernetes Authors.

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

package exoscale

import (
	"os"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	egoscale "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/exoscale/internal/github.com/exoscale/egoscale/v2"
)

func (ts *cloudProviderTestSuite) TestNewManager() {
	manager, err := newManager(cloudprovider.NodeGroupDiscoveryOptions{})
	ts.Require().NoError(err)
	ts.Require().NotNil(manager)

	os.Unsetenv("EXOSCALE_API_KEY")
	os.Unsetenv("EXOSCALE_API_SECRET")

	manager, err = newManager(cloudprovider.NodeGroupDiscoveryOptions{})
	ts.Require().Error(err)
	ts.Require().Nil(manager)
}

func (ts *cloudProviderTestSuite) TestComputeInstanceQuota() {
	ts.p.manager.client.(*exoscaleClientMock).
		On("GetQuota", ts.p.manager.ctx, ts.p.manager.zone, "instance").
		Return(
			&egoscale.Quota{
				Resource: &testComputeInstanceQuotaName,
				Usage:    &testComputeInstanceQuotaUsage,
				Limit:    &testComputeInstanceQuotaLimit,
			},
			nil,
		)

	actual, err := ts.p.manager.computeInstanceQuota()
	ts.Require().NoError(err)
	ts.Require().Equal(int(testComputeInstanceQuotaLimit), actual)
}
