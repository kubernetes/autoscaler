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
	"github.com/stretchr/testify/mock"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	egoscale "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/exoscale/internal/github.com/exoscale/egoscale/v2"
)

func (ts *cloudProviderTestSuite) TestSKSNodepoolNodeGroup_MaxSize() {
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

	nodeGroup := &sksNodepoolNodeGroup{
		sksNodepool: &egoscale.SKSNodepool{
			ID:   &testSKSNodepoolID,
			Name: &testSKSNodepoolName,
		},
		sksCluster: &egoscale.SKSCluster{
			ID:   &testSKSClusterID,
			Name: &testSKSClusterName,
		},
		m:       ts.p.manager,
		minSize: int(testSKSNodepoolSize),
		maxSize: int(testComputeInstanceQuotaLimit),
	}

	ts.Require().Equal(int(testComputeInstanceQuotaLimit), nodeGroup.MaxSize())
}

func (ts *cloudProviderTestSuite) TestSKSNodepoolNodeGroup_MinSize() {
	nodeGroup := &sksNodepoolNodeGroup{
		sksNodepool: &egoscale.SKSNodepool{
			ID:   &testSKSNodepoolID,
			Name: &testSKSNodepoolName,
		},
		sksCluster: &egoscale.SKSCluster{
			ID:   &testSKSClusterID,
			Name: &testSKSClusterName,
		},
		m:       ts.p.manager,
		minSize: int(testSKSNodepoolSize),
		maxSize: int(testComputeInstanceQuotaLimit),
	}

	ts.Require().Equal(1, nodeGroup.MinSize())
}

func (ts *cloudProviderTestSuite) TestSKSNodepoolNodeGroup_TargetSize() {
	nodeGroup := &sksNodepoolNodeGroup{
		sksNodepool: &egoscale.SKSNodepool{
			ID:   &testSKSNodepoolID,
			Name: &testSKSNodepoolName,
			Size: &testSKSNodepoolSize,
		},
		sksCluster: &egoscale.SKSCluster{
			ID:   &testSKSClusterID,
			Name: &testSKSClusterName,
		},
		m: ts.p.manager,
	}

	actual, err := nodeGroup.TargetSize()
	ts.Require().NoError(err)
	ts.Require().Equal(int(testInstancePoolSize), actual)
}

func (ts *cloudProviderTestSuite) TestSKSNodepoolNodeGroup_IncreaseSize() {
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

	ts.p.manager.client.(*exoscaleClientMock).
		On(
			"ScaleSKSNodepool",
			ts.p.manager.ctx,
			ts.p.manager.zone,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).
		Return(nil)

	ts.p.manager.client.(*exoscaleClientMock).
		On("GetInstancePool", ts.p.manager.ctx, ts.p.manager.zone, testInstancePoolID).
		Return(&egoscale.InstancePool{
			ID:    &testInstancePoolID,
			Name:  &testInstancePoolName,
			Size:  &testInstancePoolSize,
			State: &testInstancePoolState,
		}, nil)

	nodeGroup := &sksNodepoolNodeGroup{
		sksNodepool: &egoscale.SKSNodepool{
			ID:             &testSKSNodepoolID,
			InstancePoolID: &testInstancePoolID,
			Name:           &testSKSNodepoolName,
			Size:           &testSKSNodepoolSize,
		},
		sksCluster: &egoscale.SKSCluster{
			ID:   &testSKSClusterID,
			Name: &testSKSClusterName,
		},
		m:       ts.p.manager,
		minSize: int(testSKSNodepoolSize),
		maxSize: int(testComputeInstanceQuotaLimit),
	}

	ts.Require().NoError(nodeGroup.IncreaseSize(int(testInstancePoolSize + 1)))

	// Test size increase failure if beyond current limits:
	ts.Require().Error(nodeGroup.IncreaseSize(1000))
}

func (ts *cloudProviderTestSuite) TestSKSNodepoolNodeGroup_DeleteNodes() {
	ts.p.manager.client.(*exoscaleClientMock).
		On(
			"EvictSKSNodepoolMembers",
			ts.p.manager.ctx,
			ts.p.manager.zone,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).
		Return(nil)

	ts.p.manager.client.(*exoscaleClientMock).
		On("GetInstancePool", ts.p.manager.ctx, ts.p.manager.zone, testInstancePoolID).
		Return(&egoscale.InstancePool{
			ID:    &testInstancePoolID,
			Name:  &testInstancePoolName,
			Size:  &testInstancePoolSize,
			State: &testInstancePoolState,
		}, nil)

	node := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: toProviderID(testInstanceID),
		},
	}

	nodeGroup := &sksNodepoolNodeGroup{
		sksNodepool: &egoscale.SKSNodepool{
			ID:             &testSKSNodepoolID,
			InstancePoolID: &testInstancePoolID,
			Name:           &testSKSNodepoolName,
			Size:           &testSKSNodepoolSize,
		},
		sksCluster: &egoscale.SKSCluster{
			ID:   &testSKSClusterID,
			Name: &testSKSClusterName,
		},
		m:       ts.p.manager,
		minSize: int(testSKSNodepoolSize),
		maxSize: int(testComputeInstanceQuotaLimit),
	}

	ts.Require().NoError(nodeGroup.DeleteNodes([]*apiv1.Node{node}))
}

func (ts *cloudProviderTestSuite) TestSKSNodepoolNodeGroup_Id() {
	nodeGroup := &sksNodepoolNodeGroup{
		sksNodepool: &egoscale.SKSNodepool{
			ID:             &testSKSNodepoolID,
			InstancePoolID: &testInstancePoolID,
			Name:           &testSKSNodepoolName,
		},
		sksCluster: &egoscale.SKSCluster{
			ID:   &testSKSClusterID,
			Name: &testSKSClusterName,
		},
		m:       ts.p.manager,
		minSize: int(testSKSNodepoolSize),
		maxSize: int(testComputeInstanceQuotaLimit),
	}

	ts.Require().Equal(testInstancePoolID, nodeGroup.Id())
}

func (ts *cloudProviderTestSuite) TestSKSNodepoolNodeGroup_Nodes() {
	ts.p.manager.client.(*exoscaleClientMock).
		On("GetInstancePool", ts.p.manager.ctx, ts.p.manager.zone, testInstancePoolID).
		Return(&egoscale.InstancePool{
			ID:          &testInstancePoolID,
			InstanceIDs: &[]string{testInstanceID},
			Name:        &testInstancePoolName,
			Size:        &testInstancePoolSize,
			State:       &testInstancePoolState,
		}, nil)

	ts.p.manager.client.(*exoscaleClientMock).
		On("GetInstance", ts.p.manager.ctx, ts.p.manager.zone, testInstanceID).
		Return(&egoscale.Instance{
			ID:    &testInstanceID,
			State: &testInstanceState,
		}, nil)

	nodeGroup := &sksNodepoolNodeGroup{
		sksNodepool: &egoscale.SKSNodepool{
			ID:             &testSKSNodepoolID,
			InstancePoolID: &testInstancePoolID,
			Name:           &testSKSNodepoolName,
		},
		sksCluster: &egoscale.SKSCluster{
			ID:   &testSKSClusterID,
			Name: &testSKSClusterName,
		},
		m:       ts.p.manager,
		minSize: int(testSKSNodepoolSize),
		maxSize: int(testComputeInstanceQuotaLimit),
	}

	instances, err := nodeGroup.Nodes()
	ts.Require().NoError(err)
	ts.Require().Len(instances, 1)
	ts.Require().Equal(testInstanceID, toNodeID(instances[0].Id))
	ts.Require().Equal(cloudprovider.InstanceRunning, instances[0].Status.State)
}

func (ts *cloudProviderTestSuite) TestSKSNodepoolNodeGroup_Exist() {
	nodeGroup := &sksNodepoolNodeGroup{
		sksNodepool: &egoscale.SKSNodepool{
			ID:   &testSKSNodepoolID,
			Name: &testSKSNodepoolName,
		},
		sksCluster: &egoscale.SKSCluster{
			ID:   &testSKSClusterID,
			Name: &testSKSClusterName,
		},
		m:       ts.p.manager,
		minSize: int(testSKSNodepoolSize),
		maxSize: int(testComputeInstanceQuotaLimit),
	}

	ts.Require().True(nodeGroup.Exist())
}
