/*
Copyright 2019 The Kubernetes Authors.

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

package openstack

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"

	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	apiv1 "k8s.io/api/core/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type OSASGManagerMock struct {
	mock.Mock
}

// Cleanup closes the channel to stop the goroutine refreshing cache.
func (m *OSASGManagerMock) Cleanup() error {
	args := m.Called()
	return args.Error(0)
}

// registerASG registers asg in OpenStackManager. Returns true if the node group didn't exist before or its config has changed.
func (m *OSASGManagerMock) registerASG(asg ASG) bool {
	args := m.Called(asg)
	return args.Get(0).(bool)
}

// GetASGSize gets ASG size.
func (m *OSASGManagerMock) GetASGSize(asg ASG) (int64, error) {
	args := m.Called(asg)
	return args.Get(0).(int64), args.Error(1)
}

// SetASGSize sets ASG size.
func (m *OSASGManagerMock) SetASGSize(asg ASG, size int64) error {
	args := m.Called(asg, size)
	return args.Error(0)
}

// DeleteInstances deletes the given instances. All instances must be controlled by the same ASG.
func (m *OSASGManagerMock) DeleteInstances(instances []*OpenStackRef) error {
	args := m.Called(instances)
	return args.Error(0)
}

// GetASGs returns list of registered ASGs.
func (m *OSASGManagerMock) GetASGs() []*ASGInformation {
	args := m.Called()
	return args.Get(0).([]*ASGInformation)
}

// GetASGForInstance returns ASG to which the given instance belongs.
func (m *OSASGManagerMock) GetASGForInstance(instance *OpenStackRef) (ASG, error) {
	args := m.Called(instance)
	return args.Get(0).(ASG), args.Error(1)
}

// GetASGNodes returns asg nodes.
func (m *OSASGManagerMock) GetASGNodes(asg ASG) ([]string, error) {
	args := m.Called(asg)
	return args.Get(0).([]string), args.Error(1)
}

// Refresh triggers refresh of cached resources.
func (m *OSASGManagerMock) Refresh() error {
	args := m.Called()
	return args.Error(0)
}

func (m *OSASGManagerMock) forceRefresh() error {
	args := m.Called()
	return args.Error(0)
}

// Fetch explicitly configured ASGs. These ASGs should never be unregistered
// during refreshes, even if they no longer exist in OpenStack.
func (m *OSASGManagerMock) fetchExplicitASGs(specs []string) error {
	args := m.Called(specs)
	return args.Error(0)
}

func (m *OSASGManagerMock) buildASGFromFlag(flag string) (ASG, error) {
	args := m.Called(flag)
	return args.Get(0).(ASG), args.Error(1)
}

func (m *OSASGManagerMock) buildASGFromAutoCfg(link string, cfg cloudprovider.OSASGAutoDiscoveryConfig) (ASG, error) {
	args := m.Called(link, cfg)
	return args.Get(0).(ASG), args.Error(1)
}

func (m *OSASGManagerMock) buildASGFromSpec(s *dynamic.NodeGroupSpec) (ASG, error) {
	args := m.Called(s)
	return args.Get(0).(ASG), args.Error(1)
}

// Fetch automatically discovered ASGs. These ASGs should be unregistered if
// they no longer exist in OpenStack.
func (m *OSASGManagerMock) fetchAutoASGs() error {
	args := m.Called()
	return args.Error(0)
}

// GetResourceLimiter returns resource limiter from cache.
func (m *OSASGManagerMock) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	args := m.Called()
	return args.Get(0).(*cloudprovider.ResourceLimiter), args.Error(1)
}

func (m *OSASGManagerMock) findASGsNamed(name *regexp.Regexp) ([]string, error) {
	args := m.Called(name)
	return args.Get(0).([]string), args.Error(1)
}

func TestBuildOpenStackCloudProvider(t *testing.T) {
	osASGManagerMock := &OSASGManagerMock{}

	resourceLimiter := cloudprovider.NewResourceLimiter(
		map[string]int64{cloudprovider.ResourceNameCores: 1, cloudprovider.ResourceNameMemory: 10000000},
		map[string]int64{cloudprovider.ResourceNameCores: 10, cloudprovider.ResourceNameMemory: 100000000})

	provider, err := BuildOpenStackCloudProvider(osASGManagerMock, resourceLimiter)
	assert.NoError(t, err)
	assert.NotNil(t, provider)
}

func TestNodeGroups(t *testing.T) {
	osASGManagerMock := &OSASGManagerMock{}

	os := &OpenStackCloudProvider{
		openstackManager: osASGManagerMock,
	}
	asgi := &ASGInformation{Config: &openstackASG{openstackRef: OpenStackRef{Name: "asg1"}}}
	osASGManagerMock.On("GetASGs").Return([]*ASGInformation{asgi}).Once()
	result := os.NodeGroups()
	assert.Equal(t, []cloudprovider.NodeGroup{asgi.Config}, result)
	mock.AssertExpectationsForObjects(t, osASGManagerMock)
}

func TestNodeGroupForNode(t *testing.T) {
	osASGManagerMock := &OSASGManagerMock{}

	os := &OpenStackCloudProvider{
		openstackManager: osASGManagerMock,
	}
	n := BuildTestNode("n1", 1000, 1000)
	n.Spec.ProviderID = "project1/asgroot/asg/n1"
	asg := openstackASG{openstackRef: OpenStackRef{Name: "asg1"}}
	osASGManagerMock.On("GetASGForInstance", mock.AnythingOfType("*openstack.OpenStackRef")).Return(&asg, nil).Once()

	nodeGroup, err := os.NodeGroupForNode(n)
	assert.NoError(t, err)
	assert.Equal(t, asg, *reflect.ValueOf(nodeGroup).Interface().(*openstackASG))
	mock.AssertExpectationsForObjects(t, osASGManagerMock)
}

func TestGetResourceLimiter(t *testing.T) {
	osASGManagerMock := &OSASGManagerMock{}

	resourceLimiter := cloudprovider.NewResourceLimiter(
		map[string]int64{cloudprovider.ResourceNameCores: 1, cloudprovider.ResourceNameMemory: 10000000},
		map[string]int64{cloudprovider.ResourceNameCores: 10, cloudprovider.ResourceNameMemory: 100000000})
	os := &OpenStackCloudProvider{
		openstackManager: osASGManagerMock,
		resourceLimiterFromFlags: resourceLimiter,
	}

	// Return default.
	osASGManagerMock.On("GetResourceLimiter").Return((*cloudprovider.ResourceLimiter)(nil), nil).Once()
	returnedResourceLimiter, err := os.GetResourceLimiter()
	assert.NoError(t, err)
	assert.Equal(t, resourceLimiter, returnedResourceLimiter)

	// Return for OS.
	resourceLimiterOS := cloudprovider.NewResourceLimiter(
		map[string]int64{cloudprovider.ResourceNameCores: 2, cloudprovider.ResourceNameMemory: 20000000},
		map[string]int64{cloudprovider.ResourceNameCores: 5, cloudprovider.ResourceNameMemory: 200000000})
	osASGManagerMock.On("GetResourceLimiter").Return(resourceLimiterOS, nil).Once()
	returnedResourceLimiterOS, err := os.GetResourceLimiter()
	assert.NoError(t, err)
	assert.Equal(t, returnedResourceLimiterOS, resourceLimiterOS)

	// Error in osASGManager.
	osASGManagerMock.On("GetResourceLimiter").Return((*cloudprovider.ResourceLimiter)(nil), fmt.Errorf("Some error")).Once()
	returnedResourceLimiter, err = os.GetResourceLimiter()
	assert.Error(t, err)
}

func TestASG(t *testing.T) {
	osASGManagerMock := &OSASGManagerMock{}

	asg1 := &openstackASG{
		openstackManager: osASGManagerMock,
		minSize:    0,
		maxSize:    1000,
	}

	// Test TargetSize.
	osASGManagerMock.On("GetASGSize", mock.AnythingOfType("*openstack.openstackASG")).Return(int64(2), nil).Once()
	targetSize, err := asg1.TargetSize()
	assert.NoError(t, err)
	assert.Equal(t, 2, targetSize)
	mock.AssertExpectationsForObjects(t, osASGManagerMock)

	// Test IncreaseSize.
	osASGManagerMock.On("GetASGSize", mock.AnythingOfType("*openstack.openstackASG")).Return(int64(2), nil).Once()
	osASGManagerMock.On("SetASGSize", mock.AnythingOfType("*openstack.openstackASG"), int64(3)).Return(nil).Once()
	err = asg1.IncreaseSize(1)
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, osASGManagerMock)

	// Test IncreaseSize - fail on wrong size.
	err = asg1.IncreaseSize(0)
	assert.Error(t, err)
	assert.Equal(t, "size increase must be positive", err.Error())

	// Test IncreaseSize - fail on too big delta.
	osASGManagerMock.On("GetASGSize", mock.AnythingOfType("*openstack.openstackASG")).Return(int64(2), nil).Once()
	err = asg1.IncreaseSize(1000)
	assert.Error(t, err)
	assert.Equal(t, "size increase too large - desired:1002 max:1000", err.Error())
	mock.AssertExpectationsForObjects(t, osASGManagerMock)

	// Test DecreaseTargetSize.
	osASGManagerMock.On("GetASGSize", mock.AnythingOfType("*openstack.openstackASG")).Return(int64(3), nil).Once()
	osASGManagerMock.On("GetASGNodes", mock.AnythingOfType("*openstack.openstackASG")).Return(
		[]string{"project1/asgroot/asg/n1",
			"project1/asgroot/asg2/n2"}, nil).Once()
	osASGManagerMock.On("SetASGSize", mock.AnythingOfType("*openstack.openstackASG"), int64(2)).Return(nil).Once()
	err = asg1.DecreaseTargetSize(-1)
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, osASGManagerMock)

	// Test DecreaseTargetSize - fail on positive delta.
	err = asg1.DecreaseTargetSize(1)
	assert.Error(t, err)
	assert.Equal(t, "size decrease must be negative", err.Error())

	// Test DecreaseTargetSize - fail on deleting existing nodes.
	osASGManagerMock.On("GetASGSize", mock.AnythingOfType("*openstack.openstackASG")).Return(int64(3), nil).Once()
	osASGManagerMock.On("GetASGNodes", mock.AnythingOfType("*openstack.openstackASG")).Return(
		[]string{"project1/asgroot/asg/n1", "project1/asgroot/asg2/n2"}, nil).Once()

	err = asg1.DecreaseTargetSize(-2)
	assert.Error(t, err)
	assert.Equal(t, "attempt to delete existing nodes targetSize:3 delta:-2 existingNodes: 2", err.Error())
	mock.AssertExpectationsForObjects(t, osASGManagerMock)

	// Test Belongs - true.
	osASGManagerMock.On("GetASGForInstance", mock.AnythingOfType("*openstack.OpenStackRef")).Return(asg1, nil).Once()
	node := BuildTestNode("n1", 1000, 1000)
	node.Spec.ProviderID = "project1/asgroot/asg/n1"

	belongs, err := asg1.Belongs(node)
	assert.NoError(t, err)
	assert.True(t, belongs)
	mock.AssertExpectationsForObjects(t, osASGManagerMock)

	// Test Belongs - false.
	asg2 := &openstackASG{
		openstackRef: OpenStackRef{
            Project:        "project2",
            RootResource:   "rootres2",
            Resource:       "resid",
            Name:           "resname",
		},
		openstackManager: osASGManagerMock,
		minSize:    0,
		maxSize:    1000,
	}
	osASGManagerMock.On("GetASGForInstance", mock.AnythingOfType("*openstack.OpenStackRef")).Return(asg2, nil).Once()

	belongs, err = asg1.Belongs(node)
	assert.NoError(t, err)
	assert.False(t, belongs)
	mock.AssertExpectationsForObjects(t, osASGManagerMock)

	// Test DeleteNodes.
	n1 := BuildTestNode("resname", 1000, 1000)
	n1.Spec.ProviderID = "project2/rootres2/resid/resname"
	n1ref := &OpenStackRef{"project2", "rootres2", "resid", "resname"}
	n2 := BuildTestNode("resname2", 1000, 1000)
	n2.Spec.ProviderID = "project2/rootres3/resid2/resname2"
	n2ref := &OpenStackRef{"project2", "rootres3", "resid2", "resname2"}
	osASGManagerMock.On("GetASGSize", mock.AnythingOfType("*openstack.openstackASG")).Return(int64(2), nil).Once()
	osASGManagerMock.On("GetASGForInstance", n1ref).Return(asg1, nil).Once()
	osASGManagerMock.On("GetASGForInstance", n2ref).Return(asg1, nil).Once()
	osASGManagerMock.On("DeleteInstances", []*OpenStackRef{n1ref, n2ref}).Return(nil).Once()
	err = asg1.DeleteNodes([]*apiv1.Node{n1, n2})
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, osASGManagerMock)

	// Test DeleteNodes - fail on reaching min size.
	osASGManagerMock.On("GetASGSize", mock.AnythingOfType("*openstack.openstackASG")).Return(int64(0), nil).Once()
	err = asg1.DeleteNodes([]*apiv1.Node{n1, n2})
	assert.Error(t, err)
	assert.Equal(t, "min size reached, nodes will not be deleted", err.Error())
	mock.AssertExpectationsForObjects(t, osASGManagerMock)

	// Test Nodes.
	osASGManagerMock.On("GetASGNodes", mock.AnythingOfType("*openstack.openstackASG")).Return(
		[]string{"project2/rootres2/resid/resname",
			"project2/rootres3/resid2/resname2"}, nil).Once()
	nodes, err := asg1.Nodes()
	assert.NoError(t, err)
    instanceNames := []string{"project2/rootres2/resid/resname", "project2/rootres3/resid2/resname2"}
	instances := make([]cloudprovider.Instance, 0, len(instanceNames))
	for _, instanceName := range instanceNames {
		instances = append(instances, cloudprovider.Instance{Id: instanceName})
	}
	assert.Equal(t, instances, nodes)
	mock.AssertExpectationsForObjects(t, osASGManagerMock)
}

func TestOpenStackRefFromProviderId(t *testing.T) {
	ref, err := OpenStackRefFromProviderId("project2/rootres2/resid/resname")
	assert.NoError(t, err)
	assert.Equal(t, OpenStackRef{"project2", "rootres2", "resid", "resname"}, *ref)
}
