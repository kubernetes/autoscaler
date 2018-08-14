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

package gke

import (
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/gce"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"

	apiv1 "k8s.io/api/core/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	gcev1 "google.golang.org/api/compute/v1"
)

type gkeManagerMock struct {
	mock.Mock
}

func (m *gkeManagerMock) GetMigSize(mig gce.Mig) (int64, error) {
	args := m.Called(mig)
	return args.Get(0).(int64), args.Error(1)
}

func (m *gkeManagerMock) SetMigSize(mig gce.Mig, size int64) error {
	args := m.Called(mig, size)
	return args.Error(0)
}

func (m *gkeManagerMock) DeleteInstances(instances []*gce.GceRef) error {
	args := m.Called(instances)
	return args.Error(0)
}

func (m *gkeManagerMock) GetMigForInstance(instance *gce.GceRef) (gce.Mig, error) {
	args := m.Called(instance)
	return args.Get(0).(*gkeMig), args.Error(1)
}

func (m *gkeManagerMock) GetMigNodes(mig gce.Mig) ([]string, error) {
	args := m.Called(mig)
	return args.Get(0).([]string), args.Error(1)
}

func (m *gkeManagerMock) Refresh() error {
	args := m.Called()
	return args.Error(0)
}

func (m *gkeManagerMock) Cleanup() error {
	args := m.Called()
	return args.Error(0)
}

func (m *gkeManagerMock) GetMigs() []*gce.MigInformation {
	args := m.Called()
	return args.Get(0).([]*gce.MigInformation)
}

func (m *gkeManagerMock) CreateNodePool(mig *gkeMig) (*gkeMig, error) {
	args := m.Called(mig)
	return mig, args.Error(0)
}

func (m *gkeManagerMock) DeleteNodePool(toBeRemoved *gkeMig) error {
	args := m.Called(toBeRemoved)
	return args.Error(0)
}

func (m *gkeManagerMock) GetLocation() string {
	args := m.Called()
	return args.String(0)
}

func (m *gkeManagerMock) GetProjectId() string {
	args := m.Called()
	return args.String(0)
}

func (m *gkeManagerMock) GetClusterName() string {
	args := m.Called()
	return args.String(0)
}

func (m *gkeManagerMock) getMode() GcpCloudProviderMode {
	args := m.Called()
	return args.Get(0).(GcpCloudProviderMode)
}

func (m *gkeManagerMock) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	args := m.Called()
	return args.Get(0).(*cloudprovider.ResourceLimiter), args.Error(1)
}

func (m *gkeManagerMock) findMigsNamed(name *regexp.Regexp) ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
}

func (m *gkeManagerMock) GetMigTemplateNode(mig *gkeMig) (*apiv1.Node, error) {
	args := m.Called(mig)
	return args.Get(0).(*apiv1.Node), args.Error(1)
}

func (m *gkeManagerMock) getCpuAndMemoryForMachineType(machineType string, zone string) (cpu int64, mem int64, err error) {
	args := m.Called(machineType, zone)
	return args.Get(0).(int64), args.Get(1).(int64), args.Error(2)
}

func TestBuildGkeCloudProvider(t *testing.T) {
	gkeManagerMock := &gkeManagerMock{}

	resourceLimiter := cloudprovider.NewResourceLimiter(
		map[string]int64{cloudprovider.ResourceNameCores: 1, cloudprovider.ResourceNameMemory: 10000000},
		map[string]int64{cloudprovider.ResourceNameCores: 10, cloudprovider.ResourceNameMemory: 100000000})

	provider, err := BuildGkeCloudProvider(gkeManagerMock, resourceLimiter)
	assert.NoError(t, err)
	assert.NotNil(t, provider)
}

func TestNodeGroups(t *testing.T) {
	gkeManagerMock := &gkeManagerMock{}
	gke := &GkeCloudProvider{
		gkeManager: gkeManagerMock,
	}
	mig := &gce.MigInformation{Config: &gkeMig{gceRef: gce.GceRef{Name: "ng1"}}}
	gkeManagerMock.On("GetMigs").Return([]*gce.MigInformation{mig}).Once()
	result := gke.NodeGroups()
	assert.Equal(t, []cloudprovider.NodeGroup{mig.Config}, result)
	mock.AssertExpectationsForObjects(t, gkeManagerMock)
}

func TestNodeGroupForNode(t *testing.T) {
	gkeManagerMock := &gkeManagerMock{}
	gke := &GkeCloudProvider{
		gkeManager: gkeManagerMock,
	}
	n := BuildTestNode("n1", 1000, 1000)
	n.Spec.ProviderID = "gce://project1/us-central1-b/n1"
	mig := gkeMig{gceRef: gce.GceRef{Name: "ng1"}}
	gkeManagerMock.On("GetMigForInstance", mock.AnythingOfType("*gce.GceRef")).Return(&mig, nil).Once()

	nodeGroup, err := gke.NodeGroupForNode(n)
	assert.NoError(t, err)
	assert.Equal(t, mig, *reflect.ValueOf(nodeGroup).Interface().(*gkeMig))
	mock.AssertExpectationsForObjects(t, gkeManagerMock)
}

func TestGetResourceLimiter(t *testing.T) {
	gkeManagerMock := &gkeManagerMock{}
	resourceLimiter := cloudprovider.NewResourceLimiter(
		map[string]int64{cloudprovider.ResourceNameCores: 1, cloudprovider.ResourceNameMemory: 10000000},
		map[string]int64{cloudprovider.ResourceNameCores: 10, cloudprovider.ResourceNameMemory: 100000000})
	gke := &GkeCloudProvider{
		gkeManager:               gkeManagerMock,
		resourceLimiterFromFlags: resourceLimiter,
	}

	// Return default.
	gkeManagerMock.On("GetResourceLimiter").Return((*cloudprovider.ResourceLimiter)(nil), nil).Once()
	returnedResourceLimiter, err := gke.GetResourceLimiter()
	assert.NoError(t, err)
	assert.Equal(t, resourceLimiter, returnedResourceLimiter)

	// Return for GKE.
	resourceLimiterGKE := cloudprovider.NewResourceLimiter(
		map[string]int64{cloudprovider.ResourceNameCores: 2, cloudprovider.ResourceNameMemory: 20000000},
		map[string]int64{cloudprovider.ResourceNameCores: 5, cloudprovider.ResourceNameMemory: 200000000})
	gkeManagerMock.On("GetResourceLimiter").Return(resourceLimiterGKE, nil).Once()
	returnedResourceLimiterGKE, err := gke.GetResourceLimiter()
	assert.NoError(t, err)
	assert.Equal(t, returnedResourceLimiterGKE, resourceLimiterGKE)

	// Error in GceManager.
	gkeManagerMock.On("GetResourceLimiter").Return((*cloudprovider.ResourceLimiter)(nil), fmt.Errorf("Some error")).Once()
	returnedResourceLimiter, err = gke.GetResourceLimiter()
	assert.Error(t, err)
}

const getInstanceGroupManagerResponse = `{
  "kind": "compute#instanceGroupManager",
  "id": "3213213219",
  "creationTimestamp": "2017-09-15T04:47:24.687-07:00",
  "name": "gke-cluster-1-default-pool",
  "zone": "https://www.googleapis.com/compute/v1/projects/project1/zones/us-central1-b",
  "instanceTemplate": "https://www.googleapis.com/compute/v1/projects/project1/global/instanceTemplates/gke-cluster-1-default-pool",
  "instanceGroup": "https://www.googleapis.com/compute/v1/projects/project1/zones/us-central1-b/instanceGroups/gke-cluster-1-default-pool",
  "baseInstanceName": "gke-cluster-1-default-pool-f23aac-grp",
  "fingerprint": "kfdsuH",
  "currentActions": {
    "none": 3,
    "creating": 0,
    "creatingWithoutRetries": 0,
    "recreating": 0,
    "deleting": 0,
    "abandoning": 0,
    "restarting": 0,
    "refreshing": 0
  },
  "targetSize": 3,
  "selfLink": "https://www.googleapis.com/compute/v1/projects/project1/zones/us-central1-b/instanceGroupManagers/gke-cluster-1-default-pool"
}`

var gceInstanceTemplate = &gcev1.InstanceTemplate{
	Kind:              "compute#instanceTemplate",
	Id:                28701103232323232,
	CreationTimestamp: "2017-09-15T04:47:21.577-07:00",
	Name:              "gke-cluster-1-default-pool",
	Properties: &gcev1.InstanceProperties{
		Tags: &gcev1.Tags{
			Items: []string{"gke-cluster-1-000-node"},
		},
		MachineType:  "n1-standard-1",
		CanIpForward: true,
		NetworkInterfaces: []*gcev1.NetworkInterface{
			{
				Kind:    "compute#networkInterface",
				Network: "https://www.googleapis.com/compute/v1/projects/project1/global/networks/default",
			},
		},
		Metadata: &gcev1.Metadata{
			Kind:        "compute#metadata",
			Fingerprint: "F7n_RsHD3ng=",
			Items: []*gcev1.MetadataItems{
				{
					Key:   "kube-env",
					Value: createString("ALLOCATE_NODE_CIDRS: \"true\"\n"),
				},
				{
					Key:   "user-data",
					Value: createString("#cloud-config"),
				},
				{
					Key:   "gci-update-strategy",
					Value: createString("update_disabled"),
				},
				{
					Key:   "gci-ensure-gke-docker",
					Value: createString("true"),
				},
				{
					Key:   "configure-sh",
					Value: createString("#!/bin/bash\n\n<#"),
				},
				{
					Key:   "cluster-name",
					Value: createString("cluster-1"),
				},
			},
		},
	},
}

func TestMig(t *testing.T) {
	server := NewHttpServerMock()
	defer server.Close()
	gkeManagerMock := &gkeManagerMock{}
	client := &http.Client{}
	gceService, err := gcev1.New(client)
	assert.NoError(t, err)
	gceService.BasePath = server.URL
	gke := &GkeCloudProvider{
		gkeManager: gkeManagerMock,
	}

	// Test NewNodeGroup.
	gkeManagerMock.On("GetProjectId").Return("project1").Once()
	gkeManagerMock.On("GetLocation").Return("us-central1-b").Once()
	gkeManagerMock.On("GetMigTemplateNode", mock.AnythingOfType("*gke.gkeMig")).Return(&apiv1.Node{}, nil).Once()
	nodeGroup, err := gke.NewNodeGroup("n1-standard-1", nil, nil, nil, nil)
	assert.NoError(t, err)
	assert.NotNil(t, nodeGroup)
	mig1 := reflect.ValueOf(nodeGroup).Interface().(*gkeMig)
	assert.Equal(t, true, mig1.Autoprovisioned())
	mig1.exist = true
	assert.True(t, strings.HasPrefix(mig1.Id(), "https://content.googleapis.com/compute/v1/projects/project1/zones/us-central1-b/instanceGroups/"+nodeAutoprovisioningPrefix+"-n1-standard-1"))
	assert.Equal(t, true, mig1.Autoprovisioned())
	assert.Equal(t, 0, mig1.MinSize())
	assert.Equal(t, 1000, mig1.MaxSize())
	mock.AssertExpectationsForObjects(t, gkeManagerMock)

	// Test TargetSize.
	gkeManagerMock.On("GetMigSize", mock.AnythingOfType("*gke.gkeMig")).Return(int64(2), nil).Once()
	targetSize, err := mig1.TargetSize()
	assert.NoError(t, err)
	assert.Equal(t, 2, targetSize)
	mock.AssertExpectationsForObjects(t, gkeManagerMock)

	// Test IncreaseSize.
	gkeManagerMock.On("GetMigSize", mock.AnythingOfType("*gke.gkeMig")).Return(int64(2), nil).Once()
	gkeManagerMock.On("SetMigSize", mock.AnythingOfType("*gke.gkeMig"), int64(3)).Return(nil).Once()
	err = mig1.IncreaseSize(1)
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, gkeManagerMock)

	// Test IncreaseSize - fail on wrong size.
	err = mig1.IncreaseSize(0)
	assert.Error(t, err)
	assert.Equal(t, "size increase must be positive", err.Error())

	// Test IncreaseSize - fail on too big delta.
	gkeManagerMock.On("GetMigSize", mock.AnythingOfType("*gke.gkeMig")).Return(int64(2), nil).Once()
	err = mig1.IncreaseSize(1000)
	assert.Error(t, err)
	assert.Equal(t, "size increase too large - desired:1002 max:1000", err.Error())
	mock.AssertExpectationsForObjects(t, gkeManagerMock)

	// Test DecreaseTargetSize.
	gkeManagerMock.On("GetMigSize", mock.AnythingOfType("*gke.gkeMig")).Return(int64(3), nil).Once()
	gkeManagerMock.On("GetMigNodes", mock.AnythingOfType("*gke.gkeMig")).Return(
		[]string{"gce://project1/us-central1-b/gke-cluster-1-default-pool-f7607aac-9j4g",
			"gce://project1/us-central1-b/gke-cluster-1-default-pool-f7607aac-dck1"}, nil).Once()
	gkeManagerMock.On("SetMigSize", mock.AnythingOfType("*gke.gkeMig"), int64(2)).Return(nil).Once()
	err = mig1.DecreaseTargetSize(-1)
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, gkeManagerMock)

	// Test DecreaseTargetSize - fail on positive delta.
	err = mig1.DecreaseTargetSize(1)
	assert.Error(t, err)
	assert.Equal(t, "size decrease must be negative", err.Error())

	// Test DecreaseTargetSize - fail on deleting existing nodes.
	gkeManagerMock.On("GetMigSize", mock.AnythingOfType("*gke.gkeMig")).Return(int64(3), nil).Once()
	gkeManagerMock.On("GetMigNodes", mock.AnythingOfType("*gke.gkeMig")).Return(
		[]string{"gce://project1/us-central1-b/gke-cluster-1-default-pool-f7607aac-9j4g",
			"gce://project1/us-central1-b/gke-cluster-1-default-pool-f7607aac-dck1"}, nil).Once()

	err = mig1.DecreaseTargetSize(-2)
	assert.Error(t, err)
	assert.Equal(t, "attempt to delete existing nodes targetSize:3 delta:-2 existingNodes: 2", err.Error())
	mock.AssertExpectationsForObjects(t, gkeManagerMock)

	// Test Belongs - true.
	gkeManagerMock.On("GetMigForInstance", mock.AnythingOfType("*gce.GceRef")).Return(mig1, nil).Once()
	node := BuildTestNode("gke-cluster-1-default-pool-f7607aac-dck1", 1000, 1000)
	node.Spec.ProviderID = "gce://project1/us-central1-b/gke-cluster-1-default-pool-f7607aac-dck1"

	belongs, err := mig1.Belongs(node)
	assert.NoError(t, err)
	assert.True(t, belongs)
	mock.AssertExpectationsForObjects(t, gkeManagerMock)

	// Test Belongs - false.
	mig2 := &gkeMig{
		gceRef: gce.GceRef{
			Project: "project1",
			Zone:    "us-central1-b",
			Name:    "default-pool",
		},
		gkeManager:      gkeManagerMock,
		minSize:         0,
		maxSize:         1000,
		autoprovisioned: true,
		exist:           true,
		nodePoolName:    "default-pool",
		spec:            nil}
	gkeManagerMock.On("GetMigForInstance", mock.AnythingOfType("*gce.GceRef")).Return(mig2, nil).Once()

	belongs, err = mig1.Belongs(node)
	assert.NoError(t, err)
	assert.False(t, belongs)
	mock.AssertExpectationsForObjects(t, gkeManagerMock)

	// Test DeleteNodes.
	n1 := BuildTestNode("gke-cluster-1-default-pool-f7607aac-9j4g", 1000, 1000)
	n1.Spec.ProviderID = "gce://project1/us-central1-b/gke-cluster-1-default-pool-f7607aac-9j4g"
	n1ref := &gce.GceRef{"project1", "us-central1-b", "gke-cluster-1-default-pool-f7607aac-9j4g"}
	n2 := BuildTestNode("gke-cluster-1-default-pool-f7607aac-dck1", 1000, 1000)
	n2.Spec.ProviderID = "gce://project1/us-central1-b/gke-cluster-1-default-pool-f7607aac-dck1"
	n2ref := &gce.GceRef{"project1", "us-central1-b", "gke-cluster-1-default-pool-f7607aac-dck1"}
	gkeManagerMock.On("GetMigSize", mock.AnythingOfType("*gke.gkeMig")).Return(int64(2), nil).Once()
	gkeManagerMock.On("GetMigForInstance", n1ref).Return(mig1, nil).Once()
	gkeManagerMock.On("GetMigForInstance", n2ref).Return(mig1, nil).Once()
	gkeManagerMock.On("DeleteInstances", []*gce.GceRef{n1ref, n2ref}).Return(nil).Once()
	err = mig1.DeleteNodes([]*apiv1.Node{n1, n2})
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, gkeManagerMock)

	// Test DeleteNodes - fail on reaching min size.
	gkeManagerMock.On("GetMigSize", mock.AnythingOfType("*gke.gkeMig")).Return(int64(0), nil).Once()
	err = mig1.DeleteNodes([]*apiv1.Node{n1, n2})
	assert.Error(t, err)
	assert.Equal(t, "min size reached, nodes will not be deleted", err.Error())
	mock.AssertExpectationsForObjects(t, gkeManagerMock)

	// Test Nodes.
	gkeManagerMock.On("GetMigNodes", mock.AnythingOfType("*gke.gkeMig")).Return(
		[]string{"gce://project1/us-central1-b/gke-cluster-1-default-pool-f7607aac-9j4g",
			"gce://project1/us-central1-b/gke-cluster-1-default-pool-f7607aac-dck1"}, nil).Once()
	nodes, err := mig1.Nodes()
	assert.NoError(t, err)
	assert.Equal(t, "gce://project1/us-central1-b/gke-cluster-1-default-pool-f7607aac-9j4g", nodes[0])
	assert.Equal(t, "gce://project1/us-central1-b/gke-cluster-1-default-pool-f7607aac-dck1", nodes[1])
	mock.AssertExpectationsForObjects(t, gkeManagerMock)

	// Test Create.
	mig1.exist = false
	gkeManagerMock.On("CreateNodePool", mock.AnythingOfType("*gke.gkeMig")).Return(nil, nil).Once()
	_, err = mig1.Create()
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, gkeManagerMock)

	gkeManagerMock.On("DeleteNodePool", mock.AnythingOfType("*gke.gkeMig")).Return(nil).Once()
	mig1.exist = true
	err = mig1.Delete()
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, gkeManagerMock)

	// Test TemplateNodeInfo.
	gkeManagerMock.On("GetMigTemplateNode", mock.AnythingOfType("*gke.gkeMig")).Return(&apiv1.Node{}, nil).Once()
	templateNodeInfo, err := mig2.TemplateNodeInfo()
	assert.NoError(t, err)
	assert.NotNil(t, templateNodeInfo)
	assert.NotNil(t, templateNodeInfo.Node())
	mock.AssertExpectationsForObjects(t, gkeManagerMock)
}

func TestNewNodeGroupForGpu(t *testing.T) {
	server := NewHttpServerMock()
	defer server.Close()
	gkeManagerMock := &gkeManagerMock{}
	client := &http.Client{}
	gceService, err := gcev1.New(client)
	assert.NoError(t, err)
	gceService.BasePath = server.URL
	gke := &GkeCloudProvider{
		gkeManager: gkeManagerMock,
	}

	// Test NewNodeGroup.
	gkeManagerMock.On("GetProjectId").Return("project1").Once()
	gkeManagerMock.On("GetLocation").Return("us-west1-b").Once()
	gkeManagerMock.On("GetMigTemplateNode", mock.AnythingOfType("*gke.gkeMig")).Return(&apiv1.Node{}, nil).Once()

	systemLabels := map[string]string{
		gpu.GPULabel: gpu.DefaultGPUType,
	}
	extraResources := map[string]resource.Quantity{
		gpu.ResourceNvidiaGPU: resource.MustParse("1"),
	}
	nodeGroup, err := gke.NewNodeGroup("n1-standard-1", make(map[string]string), systemLabels, nil, extraResources)
	assert.NoError(t, err)
	assert.NotNil(t, nodeGroup)
	mig1 := reflect.ValueOf(nodeGroup).Interface().(*gkeMig)
	assert.True(t, strings.HasPrefix(mig1.Id(), "https://content.googleapis.com/compute/v1/projects/project1/zones/us-west1-b/instanceGroups/"+nodeAutoprovisioningPrefix+"-n1-standard-1-gpu"))
	assert.Equal(t, true, mig1.Autoprovisioned())
	assert.Equal(t, 0, mig1.MinSize())
	assert.Equal(t, 1000, mig1.MaxSize())
	expectedTaints := []apiv1.Taint{
		{
			Key:    gpu.ResourceNvidiaGPU,
			Value:  "present",
			Effect: apiv1.TaintEffectNoSchedule,
		},
	}
	assert.Equal(t, expectedTaints, mig1.spec.Taints)
	mock.AssertExpectationsForObjects(t, gkeManagerMock)
}
func TestGceRefFromProviderId(t *testing.T) {
	ref, err := gce.GceRefFromProviderId("gce://project1/us-central1-b/name1")
	assert.NoError(t, err)
	assert.Equal(t, gce.GceRef{"project1", "us-central1-b", "name1"}, *ref)
}

func TestGetClusterInfo(t *testing.T) {
	gkeManagerMock := &gkeManagerMock{}
	gke := &GkeCloudProvider{
		gkeManager: gkeManagerMock,
	}
	gkeManagerMock.On("GetProjectId").Return("project1").Once()
	gkeManagerMock.On("GetLocation").Return("location1").Once()
	gkeManagerMock.On("GetClusterName").Return("cluster1").Once()

	project, location, cluster := gke.GetClusterInfo()
	assert.Equal(t, "project1", project)
	assert.Equal(t, "location1", location)
	assert.Equal(t, "cluster1", cluster)
}

func createString(s string) *string {
	return &s
}
