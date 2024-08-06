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
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	egoscale "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/exoscale/internal/github.com/exoscale/egoscale/v2"
)

var (
	testComputeInstanceQuotaLimit int64 = 20
	testComputeInstanceQuotaName        = "instance"
	testComputeInstanceQuotaUsage int64 = 4
	testInstanceID                      = new(cloudProviderTestSuite).randomID()
	testInstanceName                    = new(cloudProviderTestSuite).randomString(10)
	testInstancePoolID                  = new(cloudProviderTestSuite).randomID()
	testInstancePoolName                = new(cloudProviderTestSuite).randomString(10)
	testInstancePoolSize          int64 = 1
	testInstancePoolState               = "running"
	testInstanceState                   = "running"
	testSKSClusterID                    = new(cloudProviderTestSuite).randomID()
	testSKSClusterName                  = new(cloudProviderTestSuite).randomString(10)
	testSKSNodepoolID                   = new(cloudProviderTestSuite).randomID()
	testSKSNodepoolName                 = new(cloudProviderTestSuite).randomString(10)
	testSKSNodepoolSize           int64 = 1
	testSeededRand                      = rand.New(rand.NewSource(time.Now().UnixNano()))
	testZone                            = "ch-gva-2"
)

type exoscaleClientMock struct {
	mock.Mock
}

func (m *exoscaleClientMock) EvictInstancePoolMembers(
	ctx context.Context,
	zone string,
	instancePool *egoscale.InstancePool,
	members []string,
) error {
	args := m.Called(ctx, zone, instancePool, members)
	return args.Error(0)
}

func (m *exoscaleClientMock) EvictSKSNodepoolMembers(
	ctx context.Context,
	zone string,
	cluster *egoscale.SKSCluster,
	nodepool *egoscale.SKSNodepool,
	members []string,
) error {
	args := m.Called(ctx, zone, cluster, nodepool, members)
	return args.Error(0)
}

func (m *exoscaleClientMock) GetInstance(ctx context.Context, zone, id string) (*egoscale.Instance, error) {
	args := m.Called(ctx, zone, id)
	return args.Get(0).(*egoscale.Instance), args.Error(1)
}

func (m *exoscaleClientMock) GetInstancePool(ctx context.Context, zone, id string) (*egoscale.InstancePool, error) {
	args := m.Called(ctx, zone, id)
	return args.Get(0).(*egoscale.InstancePool), args.Error(1)
}

func (m *exoscaleClientMock) GetQuota(ctx context.Context, zone string, resource string) (*egoscale.Quota, error) {
	args := m.Called(ctx, zone, resource)
	return args.Get(0).(*egoscale.Quota), args.Error(1)
}

func (m *exoscaleClientMock) ListSKSClusters(ctx context.Context, zone string) ([]*egoscale.SKSCluster, error) {
	args := m.Called(ctx, zone)
	return args.Get(0).([]*egoscale.SKSCluster), args.Error(1)
}

func (m *exoscaleClientMock) ScaleInstancePool(
	ctx context.Context,
	zone string,
	instancePool *egoscale.InstancePool,
	size int64,
) error {
	args := m.Called(ctx, zone, instancePool, size)
	return args.Error(0)
}

func (m *exoscaleClientMock) ScaleSKSNodepool(
	ctx context.Context,
	zone string,
	cluster *egoscale.SKSCluster,
	nodepool *egoscale.SKSNodepool,
	size int64,
) error {
	args := m.Called(ctx, zone, cluster, nodepool, size)
	return args.Error(0)
}

type cloudProviderTestSuite struct {
	p *exoscaleCloudProvider

	suite.Suite
}

func (ts *cloudProviderTestSuite) SetupTest() {
	ts.T().Setenv("EXOSCALE_ZONE", testZone)
	ts.T().Setenv("EXOSCALE_API_KEY", "x")
	ts.T().Setenv("EXOSCALE_API_SECRET", "x")

	manager, err := newManager(cloudprovider.NodeGroupDiscoveryOptions{})
	if err != nil {
		ts.T().Fatalf("error initializing cloud provider manager: %v", err)
	}
	manager.client = new(exoscaleClientMock)

	provider, err := newExoscaleCloudProvider(manager, &cloudprovider.ResourceLimiter{})
	if err != nil {
		ts.T().Fatalf("error initializing cloud provider: %v", err)
	}

	ts.p = provider
}

func (ts *cloudProviderTestSuite) TearDownTest() {
}

func (ts *cloudProviderTestSuite) randomID() string {
	id, err := uuid.NewV4()
	if err != nil {
		ts.T().Fatalf("unable to generate a new UUID: %s", err)
	}
	return id.String()
}

func (ts *cloudProviderTestSuite) randomStringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[testSeededRand.Intn(len(charset))]
	}
	return string(b)
}

func (ts *cloudProviderTestSuite) randomString(length int) string {
	const defaultCharset = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	return ts.randomStringWithCharset(length, defaultCharset)
}

func (ts *cloudProviderTestSuite) TestExoscaleCloudProvider_Name() {
	ts.Require().Equal(cloudprovider.ExoscaleProviderName, ts.p.Name())
}

func (ts *cloudProviderTestSuite) TestExoscaleCloudProvider_NodeGroupForNode_InstancePool() {
	ts.p.manager.client.(*exoscaleClientMock).
		On("GetInstancePool", ts.p.manager.ctx, ts.p.manager.zone, testInstancePoolID).
		Return(
			&egoscale.InstancePool{
				ID:   &testInstancePoolID,
				Name: &testInstancePoolName,
			},
			nil,
		)

	ts.p.manager.client.(*exoscaleClientMock).
		On("GetInstance", ts.p.manager.ctx, ts.p.manager.zone, testInstanceID).
		Return(
			&egoscale.Instance{
				ID:   &testInstanceID,
				Name: &testInstanceName,
				Manager: &egoscale.InstanceManager{
					ID:   testInstancePoolID,
					Type: "instance-pool",
				},
			},
			nil,
		)

	nodeGroup, err := ts.p.NodeGroupForNode(&apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: toProviderID(testInstanceID),
		},
		ObjectMeta: v1.ObjectMeta{
			Labels: map[string]string{
				"topology.kubernetes.io/region": testZone,
			},
		},
	})
	ts.Require().NoError(err)
	ts.Require().NotNil(nodeGroup)
	ts.Require().Equal(testInstancePoolID, nodeGroup.Id())
	ts.Require().IsType(&instancePoolNodeGroup{}, nodeGroup)
}

func (ts *cloudProviderTestSuite) TestExoscaleCloudProvider_NodeGroupForNode_SKSNodepool() {
	ts.p.manager.client.(*exoscaleClientMock).
		On("GetQuota", ts.p.manager.ctx, ts.p.manager.zone, testComputeInstanceQuotaName).
		Return(
			&egoscale.Quota{
				Resource: &testComputeInstanceQuotaName,
				Usage:    &testComputeInstanceQuotaUsage,
				Limit:    &testComputeInstanceQuotaLimit,
			},
			nil,
		)

	ts.p.manager.client.(*exoscaleClientMock).
		On("ListSKSClusters", ts.p.manager.ctx, ts.p.manager.zone).
		Return(
			[]*egoscale.SKSCluster{{
				ID:   &testSKSClusterID,
				Name: &testSKSClusterName,
				Nodepools: []*egoscale.SKSNodepool{{
					ID:             &testSKSNodepoolID,
					InstancePoolID: &testInstancePoolID,
					Name:           &testSKSNodepoolName,
				}},
			}},
			nil,
		)

	ts.p.manager.client.(*exoscaleClientMock).
		On("GetInstancePool", ts.p.manager.ctx, ts.p.manager.zone, testInstancePoolID).
		Return(
			&egoscale.InstancePool{
				ID: &testInstancePoolID,
				Manager: &egoscale.InstancePoolManager{
					ID:   testSKSNodepoolID,
					Type: "sks-nodepool",
				},
				Name: &testInstancePoolName,
			},
			nil,
		)

	ts.p.manager.client.(*exoscaleClientMock).
		On("GetInstance", ts.p.manager.ctx, ts.p.manager.zone, testInstanceID).
		Return(
			&egoscale.Instance{
				ID:   &testInstanceID,
				Name: &testInstanceName,
				Manager: &egoscale.InstanceManager{
					ID:   testInstancePoolID,
					Type: "instance-pool",
				},
			},
			nil,
		)

	nodeGroup, err := ts.p.NodeGroupForNode(&apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: toProviderID(testInstanceID),
		},
		ObjectMeta: v1.ObjectMeta{
			Labels: map[string]string{
				"topology.kubernetes.io/region": testZone,
			},
		},
	})
	ts.Require().NoError(err)
	ts.Require().NotNil(nodeGroup)
	ts.Require().Equal(testInstancePoolID, nodeGroup.Id())
	ts.Require().IsType(&sksNodepoolNodeGroup{}, nodeGroup)
}

func (ts *cloudProviderTestSuite) TestExoscaleCloudProvider_NodeGroupForNode_Standalone() {
	ts.p.manager.client.(*exoscaleClientMock).
		On("GetInstance", ts.p.manager.ctx, ts.p.manager.zone, testInstanceID).
		Return(
			&egoscale.Instance{
				ID:   &testInstanceID,
				Name: &testInstanceName,
			},
			nil,
		)

	nodeGroup, err := ts.p.NodeGroupForNode(&apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: toProviderID(testInstanceID),
		},
		ObjectMeta: v1.ObjectMeta{
			Labels: map[string]string{
				"topology.kubernetes.io/region": testZone,
			},
		},
	})
	ts.Require().NoError(err)
	ts.Require().Nil(nodeGroup)
}

func (ts *cloudProviderTestSuite) TestExoscaleCloudProvider_NodeGroups() {
	var (
		instancePoolID              = ts.randomID()
		instancePoolName            = ts.randomString(10)
		instancePoolInstanceID      = ts.randomID()
		sksNodepoolInstanceID       = ts.randomID()
		sksNodepoolInstancePoolID   = ts.randomID()
		sksNodepoolInstancePoolName = ts.randomString(10)
	)

	// In order to test the caching system of the cloud provider manager,
	// we mock 1 Instance Pool based Nodegroup and 1 SKS Nodepool based
	// Nodegroup. If everything works as expected, the
	// cloudprovider.NodeGroups() method should return 2 Nodegroups.

	ts.p.manager.client.(*exoscaleClientMock).
		On("GetQuota", ts.p.manager.ctx, ts.p.manager.zone, testComputeInstanceQuotaName).
		Return(
			&egoscale.Quota{
				Resource: &testComputeInstanceQuotaName,
				Usage:    &testComputeInstanceQuotaUsage,
				Limit:    &testComputeInstanceQuotaLimit,
			},
			nil,
		)

	ts.p.manager.client.(*exoscaleClientMock).
		On("GetInstancePool", ts.p.manager.ctx, ts.p.manager.zone, instancePoolID).
		Return(
			&egoscale.InstancePool{
				ID:   &instancePoolID,
				Name: &instancePoolName,
			},
			nil,
		)

	ts.p.manager.client.(*exoscaleClientMock).
		On("GetInstance", ts.p.manager.ctx, ts.p.manager.zone, instancePoolInstanceID).
		Return(
			&egoscale.Instance{
				ID:   &testInstanceID,
				Name: &testInstanceName,
				Manager: &egoscale.InstanceManager{
					ID:   instancePoolID,
					Type: "instance-pool",
				},
			},
			nil,
		)

	instancePoolNodeGroup, err := ts.p.NodeGroupForNode(&apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: toProviderID(instancePoolInstanceID),
		},
		ObjectMeta: v1.ObjectMeta{
			Labels: map[string]string{
				"topology.kubernetes.io/region": testZone,
			},
		},
	})
	ts.Require().NoError(err)
	ts.Require().NotNil(instancePoolNodeGroup)

	// ---------------------------------------------------------------

	ts.p.manager.client.(*exoscaleClientMock).
		On("ListSKSClusters", ts.p.manager.ctx, ts.p.manager.zone).
		Return(
			[]*egoscale.SKSCluster{{
				ID:   &testSKSClusterID,
				Name: &testSKSClusterName,
				Nodepools: []*egoscale.SKSNodepool{{
					ID:             &testSKSNodepoolID,
					InstancePoolID: &sksNodepoolInstancePoolID,
					Name:           &testSKSNodepoolName,
				}},
			}},
			nil,
		)

	ts.p.manager.client.(*exoscaleClientMock).
		On("GetInstancePool", ts.p.manager.ctx, ts.p.manager.zone, sksNodepoolInstancePoolID).
		Return(
			&egoscale.InstancePool{
				ID: &sksNodepoolInstancePoolID,
				Manager: &egoscale.InstancePoolManager{
					ID:   testSKSNodepoolID,
					Type: "sks-nodepool",
				},
				Name: &sksNodepoolInstancePoolName,
			},
			nil,
		)

	ts.p.manager.client.(*exoscaleClientMock).
		On("GetInstance", ts.p.manager.ctx, ts.p.manager.zone, sksNodepoolInstanceID).
		Return(
			&egoscale.Instance{
				ID:   &testInstanceID,
				Name: &testInstanceName,
				Manager: &egoscale.InstanceManager{
					ID:   sksNodepoolInstancePoolID,
					Type: "instance-pool",
				},
			},
			nil,
		)

	sksNodepoolNodeGroup, err := ts.p.NodeGroupForNode(&apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: toProviderID(sksNodepoolInstanceID),
		},
		ObjectMeta: v1.ObjectMeta{
			Labels: map[string]string{
				"topology.kubernetes.io/region": testZone,
			},
		},
	})
	ts.Require().NoError(err)
	ts.Require().NotNil(sksNodepoolNodeGroup)

	// ---------------------------------------------------------------

	ts.Require().Len(ts.p.NodeGroups(), 2)
}

func TestSuiteExoscaleCloudProvider(t *testing.T) {
	suite.Run(t, new(cloudProviderTestSuite))
}
