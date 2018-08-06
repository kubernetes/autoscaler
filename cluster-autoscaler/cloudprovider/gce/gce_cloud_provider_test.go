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

package gce

import (
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"

	apiv1 "k8s.io/api/core/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	gcev1 "google.golang.org/api/compute/v1"
)

type gceManagerMock struct {
	mock.Mock
}

func (m *gceManagerMock) RegisterMig(mig *Mig) bool {
	args := m.Called(mig)
	return args.Bool(0)
}

func (m *gceManagerMock) UnregisterMig(toBeRemoved *Mig) bool {
	args := m.Called(toBeRemoved)
	return args.Bool(0)
}

func (m *gceManagerMock) GetMigSize(mig *Mig) (int64, error) {
	args := m.Called(mig)
	return args.Get(0).(int64), args.Error(1)
}

func (m *gceManagerMock) SetMigSize(mig *Mig, size int64) error {
	args := m.Called(mig, size)
	return args.Error(0)
}

func (m *gceManagerMock) DeleteInstances(instances []*GceRef) error {
	args := m.Called(instances)
	return args.Error(0)
}

func (m *gceManagerMock) GetMigForInstance(instance *GceRef) (*Mig, error) {
	args := m.Called(instance)
	return args.Get(0).(*Mig), args.Error(1)
}

func (m *gceManagerMock) GetMigNodes(mig *Mig) ([]string, error) {
	args := m.Called(mig)
	return args.Get(0).([]string), args.Error(1)
}

func (m *gceManagerMock) Refresh() error {
	args := m.Called()
	return args.Error(0)
}

func (m *gceManagerMock) Cleanup() error {
	args := m.Called()
	return args.Error(0)
}

func (m *gceManagerMock) getMigs() []*migInformation {
	args := m.Called()
	return args.Get(0).([]*migInformation)
}

func (m *gceManagerMock) createNodePool(mig *Mig) error {
	args := m.Called(mig)
	return args.Error(0)
}

func (m *gceManagerMock) deleteNodePool(toBeRemoved *Mig) error {
	args := m.Called(toBeRemoved)
	return args.Error(0)
}

func (m *gceManagerMock) getLocation() string {
	args := m.Called()
	return args.String(0)
}

func (m *gceManagerMock) getProjectId() string {
	args := m.Called()
	return args.String(0)
}

func (m *gceManagerMock) getMode() GcpCloudProviderMode {
	args := m.Called()
	return args.Get(0).(GcpCloudProviderMode)
}

func (m *gceManagerMock) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	args := m.Called()
	return args.Get(0).(*cloudprovider.ResourceLimiter), args.Error(1)
}

func (m *gceManagerMock) findMigsNamed(name *regexp.Regexp) ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
}

func (m *gceManagerMock) getMigTemplateNode(mig *Mig) (*apiv1.Node, error) {
	args := m.Called(mig)
	return args.Get(0).(*apiv1.Node), args.Error(1)
}

func (m *gceManagerMock) getCpuAndMemoryForMachineType(machineType string, zone string) (cpu int64, mem int64, err error) {
	args := m.Called(machineType, zone)
	return args.Get(0).(int64), args.Get(1).(int64), args.Error(2)
}

func TestBuildGceCloudProvider(t *testing.T) {
	gceManagerMock := &gceManagerMock{}

	resourceLimiter := cloudprovider.NewResourceLimiter(
		map[string]int64{cloudprovider.ResourceNameCores: 1, cloudprovider.ResourceNameMemory: 10000000},
		map[string]int64{cloudprovider.ResourceNameCores: 10, cloudprovider.ResourceNameMemory: 100000000})

	provider, err := BuildGceCloudProvider(gceManagerMock, resourceLimiter)
	assert.NoError(t, err)
	assert.NotNil(t, provider)
}

func TestNodeGroups(t *testing.T) {
	gceManagerMock := &gceManagerMock{}
	gce := &GceCloudProvider{
		gceManager: gceManagerMock,
	}
	mig := &migInformation{config: &Mig{GceRef: GceRef{Name: "ng1"}}}
	gceManagerMock.On("getMigs").Return([]*migInformation{mig}).Once()
	result := gce.NodeGroups()
	assert.Equal(t, []cloudprovider.NodeGroup{mig.config}, result)
	mock.AssertExpectationsForObjects(t, gceManagerMock)
}

func TestNodeGroupForNode(t *testing.T) {
	gceManagerMock := &gceManagerMock{}
	gce := &GceCloudProvider{
		gceManager: gceManagerMock,
	}
	n := BuildTestNode("n1", 1000, 1000)
	n.Spec.ProviderID = "gce://project1/us-central1-b/n1"
	mig := Mig{GceRef: GceRef{Name: "ng1"}}
	gceManagerMock.On("GetMigForInstance", mock.AnythingOfType("*gce.GceRef")).Return(&mig, nil).Once()

	nodeGroup, err := gce.NodeGroupForNode(n)
	assert.NoError(t, err)
	assert.Equal(t, mig, *reflect.ValueOf(nodeGroup).Interface().(*Mig))
	mock.AssertExpectationsForObjects(t, gceManagerMock)
}

func TestGetResourceLimiter(t *testing.T) {
	gceManagerMock := &gceManagerMock{}
	resourceLimiter := cloudprovider.NewResourceLimiter(
		map[string]int64{cloudprovider.ResourceNameCores: 1, cloudprovider.ResourceNameMemory: 10000000},
		map[string]int64{cloudprovider.ResourceNameCores: 10, cloudprovider.ResourceNameMemory: 100000000})
	gce := &GceCloudProvider{
		gceManager:               gceManagerMock,
		resourceLimiterFromFlags: resourceLimiter,
	}

	// Return default.
	gceManagerMock.On("GetResourceLimiter").Return((*cloudprovider.ResourceLimiter)(nil), nil).Once()
	returnedResourceLimiter, err := gce.GetResourceLimiter()
	assert.NoError(t, err)
	assert.Equal(t, resourceLimiter, returnedResourceLimiter)

	// Return for GKE.
	resourceLimiterGKE := cloudprovider.NewResourceLimiter(
		map[string]int64{cloudprovider.ResourceNameCores: 2, cloudprovider.ResourceNameMemory: 20000000},
		map[string]int64{cloudprovider.ResourceNameCores: 5, cloudprovider.ResourceNameMemory: 200000000})
	gceManagerMock.On("GetResourceLimiter").Return(resourceLimiterGKE, nil).Once()
	returnedResourceLimiterGKE, err := gce.GetResourceLimiter()
	assert.NoError(t, err)
	assert.Equal(t, returnedResourceLimiterGKE, resourceLimiterGKE)

	// Error in GceManager.
	gceManagerMock.On("GetResourceLimiter").Return((*cloudprovider.ResourceLimiter)(nil), fmt.Errorf("Some error")).Once()
	returnedResourceLimiter, err = gce.GetResourceLimiter()
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
	gceManagerMock := &gceManagerMock{}
	client := &http.Client{}
	gceService, err := gcev1.New(client)
	assert.NoError(t, err)
	gceService.BasePath = server.URL
	gce := &GceCloudProvider{
		gceManager: gceManagerMock,
	}

	// Test NewNodeGroup.
	gceManagerMock.On("getProjectId").Return("project1").Once()
	gceManagerMock.On("getLocation").Return("us-central1-b").Once()
	gceManagerMock.On("getMigTemplateNode", mock.AnythingOfType("*gce.Mig")).Return(&apiv1.Node{}, nil).Once()
	nodeGroup, err := gce.NewNodeGroup("n1-standard-1", nil, nil, nil, nil)
	assert.NoError(t, err)
	assert.NotNil(t, nodeGroup)
	mig1 := reflect.ValueOf(nodeGroup).Interface().(*Mig)
	mig1.exist = true
	assert.True(t, strings.HasPrefix(mig1.Id(), "https://content.googleapis.com/compute/v1/projects/project1/zones/us-central1-b/instanceGroups/"+nodeAutoprovisioningPrefix+"-n1-standard-1"))
	assert.Equal(t, true, mig1.Autoprovisioned())
	assert.Equal(t, 0, mig1.MinSize())
	assert.Equal(t, 1000, mig1.MaxSize())
	mock.AssertExpectationsForObjects(t, gceManagerMock)

	// Test TargetSize.
	gceManagerMock.On("GetMigSize", mock.AnythingOfType("*gce.Mig")).Return(int64(2), nil).Once()
	targetSize, err := mig1.TargetSize()
	assert.NoError(t, err)
	assert.Equal(t, 2, targetSize)
	mock.AssertExpectationsForObjects(t, gceManagerMock)

	// Test IncreaseSize.
	gceManagerMock.On("GetMigSize", mock.AnythingOfType("*gce.Mig")).Return(int64(2), nil).Once()
	gceManagerMock.On("SetMigSize", mock.AnythingOfType("*gce.Mig"), int64(3)).Return(nil).Once()
	err = mig1.IncreaseSize(1)
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, gceManagerMock)

	// Test IncreaseSize - fail on wrong size.
	err = mig1.IncreaseSize(0)
	assert.Error(t, err)
	assert.Equal(t, "size increase must be positive", err.Error())

	// Test IncreaseSize - fail on too big delta.
	gceManagerMock.On("GetMigSize", mock.AnythingOfType("*gce.Mig")).Return(int64(2), nil).Once()
	err = mig1.IncreaseSize(1000)
	assert.Error(t, err)
	assert.Equal(t, "size increase too large - desired:1002 max:1000", err.Error())
	mock.AssertExpectationsForObjects(t, gceManagerMock)

	// Test DecreaseTargetSize.
	gceManagerMock.On("GetMigSize", mock.AnythingOfType("*gce.Mig")).Return(int64(3), nil).Once()
	gceManagerMock.On("GetMigNodes", mock.AnythingOfType("*gce.Mig")).Return(
		[]string{"gce://project1/us-central1-b/gke-cluster-1-default-pool-f7607aac-9j4g",
			"gce://project1/us-central1-b/gke-cluster-1-default-pool-f7607aac-dck1"}, nil).Once()
	gceManagerMock.On("SetMigSize", mock.AnythingOfType("*gce.Mig"), int64(2)).Return(nil).Once()
	err = mig1.DecreaseTargetSize(-1)
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, gceManagerMock)

	// Test DecreaseTargetSize - fail on positive delta.
	err = mig1.DecreaseTargetSize(1)
	assert.Error(t, err)
	assert.Equal(t, "size decrease must be negative", err.Error())

	// Test DecreaseTargetSize - fail on deleting existing nodes.
	gceManagerMock.On("GetMigSize", mock.AnythingOfType("*gce.Mig")).Return(int64(3), nil).Once()
	gceManagerMock.On("GetMigNodes", mock.AnythingOfType("*gce.Mig")).Return(
		[]string{"gce://project1/us-central1-b/gke-cluster-1-default-pool-f7607aac-9j4g",
			"gce://project1/us-central1-b/gke-cluster-1-default-pool-f7607aac-dck1"}, nil).Once()

	err = mig1.DecreaseTargetSize(-2)
	assert.Error(t, err)
	assert.Equal(t, "attempt to delete existing nodes targetSize:3 delta:-2 existingNodes: 2", err.Error())
	mock.AssertExpectationsForObjects(t, gceManagerMock)

	// Test Belongs - true.
	gceManagerMock.On("GetMigForInstance", mock.AnythingOfType("*gce.GceRef")).Return(mig1, nil).Once()
	node := BuildTestNode("gke-cluster-1-default-pool-f7607aac-dck1", 1000, 1000)
	node.Spec.ProviderID = "gce://project1/us-central1-b/gke-cluster-1-default-pool-f7607aac-dck1"

	belongs, err := mig1.Belongs(node)
	assert.NoError(t, err)
	assert.True(t, belongs)
	mock.AssertExpectationsForObjects(t, gceManagerMock)

	// Test Belongs - false.
	mig2 := &Mig{
		GceRef: GceRef{
			Project: "project1",
			Zone:    "us-central1-b",
			Name:    "default-pool",
		},
		gceManager:      gceManagerMock,
		minSize:         0,
		maxSize:         1000,
		autoprovisioned: true,
		exist:           true,
		nodePoolName:    "default-pool",
		spec:            nil}
	gceManagerMock.On("GetMigForInstance", mock.AnythingOfType("*gce.GceRef")).Return(mig2, nil).Once()

	belongs, err = mig1.Belongs(node)
	assert.NoError(t, err)
	assert.False(t, belongs)
	mock.AssertExpectationsForObjects(t, gceManagerMock)

	// Test DeleteNodes.
	n1 := BuildTestNode("gke-cluster-1-default-pool-f7607aac-9j4g", 1000, 1000)
	n1.Spec.ProviderID = "gce://project1/us-central1-b/gke-cluster-1-default-pool-f7607aac-9j4g"
	n1ref := &GceRef{"project1", "us-central1-b", "gke-cluster-1-default-pool-f7607aac-9j4g"}
	n2 := BuildTestNode("gke-cluster-1-default-pool-f7607aac-dck1", 1000, 1000)
	n2.Spec.ProviderID = "gce://project1/us-central1-b/gke-cluster-1-default-pool-f7607aac-dck1"
	n2ref := &GceRef{"project1", "us-central1-b", "gke-cluster-1-default-pool-f7607aac-dck1"}
	gceManagerMock.On("GetMigSize", mock.AnythingOfType("*gce.Mig")).Return(int64(2), nil).Once()
	gceManagerMock.On("GetMigForInstance", n1ref).Return(mig1, nil).Once()
	gceManagerMock.On("GetMigForInstance", n2ref).Return(mig1, nil).Once()
	gceManagerMock.On("DeleteInstances", []*GceRef{n1ref, n2ref}).Return(nil).Once()
	err = mig1.DeleteNodes([]*apiv1.Node{n1, n2})
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, gceManagerMock)

	// Test DeleteNodes - fail on reaching min size.
	gceManagerMock.On("GetMigSize", mock.AnythingOfType("*gce.Mig")).Return(int64(0), nil).Once()
	err = mig1.DeleteNodes([]*apiv1.Node{n1, n2})
	assert.Error(t, err)
	assert.Equal(t, "min size reached, nodes will not be deleted", err.Error())
	mock.AssertExpectationsForObjects(t, gceManagerMock)

	// Test Nodes.
	gceManagerMock.On("GetMigNodes", mock.AnythingOfType("*gce.Mig")).Return(
		[]string{"gce://project1/us-central1-b/gke-cluster-1-default-pool-f7607aac-9j4g",
			"gce://project1/us-central1-b/gke-cluster-1-default-pool-f7607aac-dck1"}, nil).Once()
	nodes, err := mig1.Nodes()
	assert.NoError(t, err)
	assert.Equal(t, "gce://project1/us-central1-b/gke-cluster-1-default-pool-f7607aac-9j4g", nodes[0])
	assert.Equal(t, "gce://project1/us-central1-b/gke-cluster-1-default-pool-f7607aac-dck1", nodes[1])
	mock.AssertExpectationsForObjects(t, gceManagerMock)

	// Test Create.
	mig1.exist = false
	gceManagerMock.On("createNodePool", mock.AnythingOfType("*gce.Mig")).Return(nil).Once()
	err = mig1.Create()
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, gceManagerMock)

	// Test Delete.
	mig1.exist = true
	gceManagerMock.On("deleteNodePool", mock.AnythingOfType("*gce.Mig")).Return(nil).Once()
	err = mig1.Delete()
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, gceManagerMock)

	// Test TemplateNodeInfo.
	gceManagerMock.On("getMigTemplateNode", mock.AnythingOfType("*gce.Mig")).Return(&apiv1.Node{}, nil).Once()
	templateNodeInfo, err := mig2.TemplateNodeInfo()
	assert.NoError(t, err)
	assert.NotNil(t, templateNodeInfo)
	assert.NotNil(t, templateNodeInfo.Node())
	mock.AssertExpectationsForObjects(t, gceManagerMock)
}

func TestNewNodeGroupForGpu(t *testing.T) {
	server := NewHttpServerMock()
	defer server.Close()
	gceManagerMock := &gceManagerMock{}
	client := &http.Client{}
	gceService, err := gcev1.New(client)
	assert.NoError(t, err)
	gceService.BasePath = server.URL
	gce := &GceCloudProvider{
		gceManager: gceManagerMock,
	}

	// Test NewNodeGroup.
	gceManagerMock.On("getProjectId").Return("project1").Once()
	gceManagerMock.On("getLocation").Return("us-west1-b").Once()
	gceManagerMock.On("getMigTemplateNode", mock.AnythingOfType("*gce.Mig")).Return(&apiv1.Node{}, nil).Once()

	systemLabels := map[string]string{
		gpu.GPULabel: gpu.DefaultGPUType,
	}
	extraResources := map[string]resource.Quantity{
		gpu.ResourceNvidiaGPU: resource.MustParse("1"),
	}
	nodeGroup, err := gce.NewNodeGroup("n1-standard-1", make(map[string]string), systemLabels, nil, extraResources)
	assert.NoError(t, err)
	assert.NotNil(t, nodeGroup)
	mig1 := reflect.ValueOf(nodeGroup).Interface().(*Mig)
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
	assert.Equal(t, expectedTaints, mig1.spec.taints)
	mock.AssertExpectationsForObjects(t, gceManagerMock)
}
func TestGceRefFromProviderId(t *testing.T) {
	ref, err := GceRefFromProviderId("gce://project1/us-central1-b/name1")
	assert.NoError(t, err)
	assert.Equal(t, GceRef{"project1", "us-central1-b", "name1"}, *ref)
}

func createString(s string) *string {
	return &s
}
