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
	"testing"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"

	apiv1 "k8s.io/api/core/v1"

	"reflect"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func (m *gceManagerMock) getZone() string {
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

func (m *gceManagerMock) getTemplates() *templateBuilder {
	args := m.Called()
	return args.Get(0).(*templateBuilder)
}

func TestBuildGceCloudProvider(t *testing.T) {
	gceManagerMock := &gceManagerMock{}

	ng1Name := "https://content.googleapis.com/compute/v1/projects/project1/zones/us-central1-b/instanceGroups/ng1"
	ng2Name := "https://content.googleapis.com/compute/v1/projects/project1/zones/us-central1-b/instanceGroups/ng2"

	// GCE mode.
	gceManagerMock.On("getMode").Return(ModeGCE).Once()
	gceManagerMock.On("RegisterMig",
		mock.MatchedBy(func(mig *Mig) bool {
			return mig.Name == "ng1" || mig.Name == "ng2"
		})).Return(true).Times(2)

	provider, err := BuildGceCloudProvider(gceManagerMock,
		[]string{"0:10:" + ng1Name, "0:5:https:" + ng2Name})
	assert.NoError(t, err)
	assert.NotNil(t, provider)

	// GKE mode.
	gceManagerMock.On("getMode").Return(ModeGKE).Once()

	provider, err = BuildGceCloudProvider(gceManagerMock, []string{})
	assert.NoError(t, err)
	assert.NotNil(t, provider)

	// Error on GKE mode with specs.
	gceManagerMock.On("getMode").Return(ModeGKE).Once()
	gceManagerMock.On("RegisterMig",
		mock.MatchedBy(func(mig *Mig) bool {
			return mig.Name == "ng1" || mig.Name == "ng2"
		})).Return(true).Times(2)

	provider, err = BuildGceCloudProvider(gceManagerMock,
		[]string{"0:10:" + ng1Name, "0:5:https:" + ng2Name})
	assert.Error(t, err)
	assert.Equal(t, "GKE gets nodegroup specification via API, command line specs are not allowed", err.Error())
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
}

func TestBuildMig(t *testing.T) {
	_, err := buildMig("a", nil)
	assert.Error(t, err)
	_, err = buildMig("a:b:c", nil)
	assert.Error(t, err)
	_, err = buildMig("1:2:x", nil)
	assert.Error(t, err)
	_, err = buildMig("1:2:", nil)
	assert.Error(t, err)

	mig, err := buildMig("111:222:https://content.googleapis.com/compute/v1/projects/test-project/zones/test-zone/instanceGroups/test-name", nil)
	assert.NoError(t, err)
	assert.Equal(t, 111, mig.MinSize())
	assert.Equal(t, 222, mig.MaxSize())
	assert.Equal(t, "test-zone", mig.Zone)
	assert.Equal(t, "test-name", mig.Name)
}

func TestBuildKubeProxy(t *testing.T) {
	mig, _ := buildMig("1:20:https://content.googleapis.com/compute/v1/projects/test-project/zones/test-zone/instanceGroups/test-name", nil)
	pod := cloudprovider.BuildKubeProxy(mig.Id())
	assert.Equal(t, 1, len(pod.Spec.Containers))
	cpu := pod.Spec.Containers[0].Resources.Requests[apiv1.ResourceCPU]
	assert.Equal(t, int64(100), cpu.MilliValue())
}
