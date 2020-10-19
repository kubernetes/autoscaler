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

package cloudstack

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/cloudstack/service"
)

const (
	masterCount = 1
	workerCount = 5
)

var (
	nodeIDs = []string{"vm2", "vm3"}
)

func createClusterDetails() *service.Cluster {
	return &service.Cluster{
		ID:          testConfig.clusterID,
		Maxsize:     testConfig.maxSize,
		Minsize:     testConfig.minSize,
		MasterCount: masterCount,
		WorkerCount: 3,
		VirtualMachines: []*service.VirtualMachine{
			{
				ID: "m1",
			},
			{
				ID: "vm1",
			},
			{
				ID: "vm2",
			},
			{
				ID: "vm3",
			},
		},
	}
}

func createScaleUpClusterDetails() *service.Cluster {
	return &service.Cluster{
		ID:          testConfig.clusterID,
		Maxsize:     testConfig.maxSize,
		Minsize:     testConfig.minSize,
		MasterCount: masterCount,
		WorkerCount: 5,
		VirtualMachines: []*service.VirtualMachine{
			{
				ID: "m1",
			},
			{
				ID: "vm1",
			},
			{
				ID: "vm2",
			},
			{
				ID: "vm3",
			},
			{
				ID: "vm4",
			},
			{
				ID: "vm5",
			},
		},
	}
}

func createScaleDownClusterDetails() *service.Cluster {
	return &service.Cluster{
		ID:          testConfig.clusterID,
		Maxsize:     testConfig.maxSize,
		Minsize:     testConfig.minSize,
		MasterCount: masterCount,
		WorkerCount: 1,
		VirtualMachines: []*service.VirtualMachine{
			{
				ID: "m1",
			},
			{
				ID: "vm1",
			},
		},
	}
}

type mockCKSService struct {
	mock.Mock
}

func (m *mockCKSService) GetClusterDetails(clusterID string) (*service.Cluster, error) {
	a := m.Called(clusterID)
	return a.Get(0).(*service.Cluster), nil
}

func (m *mockCKSService) ScaleCluster(clusterID string, workerCount int) (*service.Cluster, error) {
	a := m.Called(clusterID, workerCount)
	return a.Get(0).(*service.Cluster), nil
}

func (m *mockCKSService) RemoveNodesFromCluster(clusterID string, nodeIDs ...string) (*service.Cluster, error) {
	a := m.Called(clusterID, nodeIDs)
	return a.Get(0).(*service.Cluster), nil
}

func (m *mockCKSService) Close() {
	m.Called()
}

func createMockService() *mockCKSService {
	s := &mockCKSService{}
	s.On("GetClusterDetails",
		testConfig.clusterID).
		Return(createClusterDetails())
	return s
}

func TestFetchCluster(t *testing.T) {
	clusterDetails := createClusterDetails()
	s := &mockCKSService{}
	s.On("GetClusterDetails",
		testConfig.clusterID).After(100 * time.Millisecond).
		Return(clusterDetails)

	asg := &asg{
		cluster: &service.Cluster{
			ID: "123",
		},
	}
	manager, _ := newManager(testConfig,
		withASG(asg),
		withCKSService(s))

	// Ensuring it locks
	start := time.Now()
	c := make(chan struct{})
	go func() {
		manager.fetchCluster()
		close(c)
	}()
	err := manager.fetchCluster()

	<-c
	assert.GreaterOrEqual(t, int64(time.Since(start)), int64(200*time.Millisecond))
	assert.Equal(t, nil, err)
	assert.Equal(t, testConfig.clusterID, asg.Id())
	assert.Equal(t, testConfig.maxSize, asg.MaxSize())
	assert.Equal(t, testConfig.minSize, asg.MinSize())
	s.AssertExpectations(t)
}

func TestScaleCluster(t *testing.T) {
	scaleUpClusterDetails := createScaleUpClusterDetails()
	s := createMockService()
	s.On("ScaleCluster",
		testConfig.clusterID, workerCount).After(100 * time.Millisecond).
		Return(scaleUpClusterDetails)

	manager, _ := newManager(testConfig,
		withCKSService(s))

	// Ensuring it locks
	start := time.Now()
	c := make(chan struct{})
	go func() {
		manager.scaleCluster(testConfig.clusterID, workerCount)
		close(c)
	}()
	cluster, err := manager.scaleCluster(testConfig.clusterID, workerCount)

	<-c
	assert.GreaterOrEqual(t, int64(time.Since(start)), int64(200*time.Millisecond))
	assert.Equal(t, nil, err)
	assert.Equal(t, scaleUpClusterDetails, cluster)
	s.AssertExpectations(t)
}

func TestRemoveNodesFromCluster(t *testing.T) {
	scaleDownClusterDetails := createScaleDownClusterDetails()
	s := createMockService()
	s.On("RemoveNodesFromCluster",
		testConfig.clusterID, nodeIDs).After(100 * time.Millisecond).
		Return(scaleDownClusterDetails)

	asg := &asg{}
	manager, _ := newManager(testConfig,
		withASG(asg),
		withCKSService(s))

	// Ensuring it locks
	start := time.Now()
	c := make(chan struct{})
	go func() {
		manager.removeNodesFromCluster(testConfig.clusterID, nodeIDs...)
		close(c)
	}()
	cluster, err := manager.removeNodesFromCluster(testConfig.clusterID, nodeIDs...)

	<-c
	assert.GreaterOrEqual(t, int64(time.Since(start)), int64(200*time.Millisecond))
	assert.Equal(t, nil, err)
	assert.Equal(t, scaleDownClusterDetails, cluster)
	s.AssertExpectations(t)
}

func TestManagerCleanup(t *testing.T) {
	s := createMockService()
	s.On("Close").Return()

	manager, _ := newManager(testConfig,
		withCKSService(s))
	err := manager.cleanup()

	assert.Equal(t, nil, err)
	s.AssertExpectations(t)
}

func TestCreateACSConfig(t *testing.T) {
	cfg, err := readConfig("testdata/invalid-config")
	assert.NotEqual(t, nil, err)

	cfg, err = readConfig("testdata/valid-config")
	assert.Equal(t, nil, err)
	assert.Equal(t, "api-url", cfg.Global.APIURL)
	assert.Equal(t, "api-key", cfg.Global.APIKey)
	assert.Equal(t, "secret-key", cfg.Global.SecretKey)
}
