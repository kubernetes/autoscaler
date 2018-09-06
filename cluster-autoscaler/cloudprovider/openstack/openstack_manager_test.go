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
    "errors"
	"fmt"
	"regexp"
	"testing"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var Clients = []string{"Orchestration"}

const (
	projectId      = "project1"
)


type AutoscalingOSClientMock struct {
	mock.Mock
}

func (client *AutoscalingOSClientMock) FetchASGTargetSize(asgRef OpenStackRef) (int64, error) {
	m := client.Called(asgRef)
	return m.Get(0).(int64), m.Error(1)
}

func (client *AutoscalingOSClientMock) ResizeASG(asgRef OpenStackRef, size int64) error {
	m := client.Called(asgRef, size)
	return m.Error(0)
}

func (client *AutoscalingOSClientMock) DeleteInstances(asgRef OpenStackRef, instances []*OpenStackRef) error {
	m := client.Called(asgRef, instances)
	return m.Error(0)
}

func (client *AutoscalingOSClientMock) FetchASGInstances(asgRef OpenStackRef) ([]OpenStackRef, error) {
	m := client.Called(asgRef)
	return m.Get(0).([]OpenStackRef), m.Error(1)
}

func (client *AutoscalingOSClientMock) FetchASGBasename(asgRef OpenStackRef) (string, error) {
	m := client.Called(asgRef)
	return m.Get(0).(string), m.Error(1)
}


func (client *AutoscalingOSClientMock) FetchASGsWithName(name *regexp.Regexp) ([]string, error) {
	m := client.Called(name)
	return m.Get(0).([]string), m.Error(1)
}

func newTestOSManager(t *testing.T) (*AutoscalingOSClientMock, *openstackManagerImpl) {
	osService := &AutoscalingOSClientMock{}

	manager := &openstackManagerImpl{
		cache: OpenStackCache{
			asgs:           make([]*ASGInformation, 0),
			OpenStackService:     osService,
			instancesCache: make(map[OpenStackRef]ASG),
		},
		OpenStackService:   osService,
		projectId:          projectId,
		explicitlyConfigured: make(map[OpenStackRef]bool),
	}
	return osService, manager
}

func TestGetASGSizeWithError(t *testing.T) {
    osService, manager := newTestOSManager(t)
	osService.On("FetchASGTargetSize",  mock.AnythingOfType("openstack.OpenStackRef")).Return(int64(1), errors.New("failed")).Once()
	asg := &openstackASG{openstackRef: OpenStackRef{Name: "asg1"}}
    target_size, err := manager.GetASGSize(asg)
	assert.Error(t, err)
	assert.Equal(t, int64(-1), target_size)
	mock.AssertExpectationsForObjects(t, osService)
}

func TestDeleteInstances(t *testing.T) {
	osService := &AutoscalingOSClientMock{}
	asg := &openstackASG{openstackRef: OpenStackRef{Name: "asg1"}}

	manager := &openstackManagerImpl{
		cache: OpenStackCache{
			asgs:           make([]*ASGInformation, 0),
			OpenStackService:     osService,
			instancesCache: map[OpenStackRef]ASG{
				{"project1", "rootres1", "resid1", "resname1"}: asg,
				{"project1", "rootres1", "resid2", "resname2"}: asg,
				{"project1", "rootres1", "resid3", "resname3"}: asg,
			},
		},
		OpenStackService:   osService,
		projectId:          projectId,
		explicitlyConfigured: make(map[OpenStackRef]bool),
	}

	osService.On("DeleteInstances", mock.AnythingOfType("openstack.OpenStackRef"), mock.AnythingOfType("[]*openstack.OpenStackRef")).Return(nil).Once()
    
    osref := []*OpenStackRef{
        {
            Project:        "project1",
            RootResource:   "rootres1",
            Resource:       "resid1",
            Name:           "resname1",
        },
    }
    err := manager.DeleteInstances(osref)
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, osService)
}

func TestGetASGNodesWithError(t *testing.T) {
    osService, manager := newTestOSManager(t)
	osService.On("FetchASGInstances",  mock.AnythingOfType("openstack.OpenStackRef")).Return([]OpenStackRef{}, errors.New("failed")).Once()
	asg := &openstackASG{openstackRef: OpenStackRef{Name: "asg1"}}
    result, err := manager.GetASGNodes(asg)
	assert.Error(t, err)
	assert.Equal(t, []string{}, result)
	mock.AssertExpectationsForObjects(t, osService)
}

func TestNoRefresh(t *testing.T) {
    time_last_refresh := time.Now()
    _, manager := newTestOSManager(t)
    manager.lastRefresh = time_last_refresh
    err := manager.Refresh()
	assert.NoError(t, err)
	assert.Equal(t, time_last_refresh, manager.lastRefresh)
}

func TestForceRefresh(t *testing.T) {
    time_last_refresh := time.Now().Add(-2*time.Minute)
    _, manager := newTestOSManager(t)
    manager.lastRefresh = time_last_refresh
    err := manager.Refresh()
	assert.NoError(t, err)
	assert.NotEqual(t, time_last_refresh, manager.lastRefresh)
}

func TestFetchExplicitASGs(t *testing.T) {
    url := "project_id/root_resource_id/resource_id/name"
	s := &dynamic.NodeGroupSpec{
		Name:               url,
		MinSize:            2,
		MaxSize:            3,
		SupportScaleToZero: false,
	}
    _, manager := newTestOSManager(t)
    asg, err := manager.buildASGFromSpec(s)
	assert.NoError(t, err)
	assert.Equal(t, 2, asg.MinSize())
	assert.Equal(t, 3, asg.MaxSize())
}

func TestBuildASGFromSpecFailedToParse(t *testing.T) {
    url := "project_id/root_resource_id/resource_id/name/extra_info"
    expected_format := "<project-id>/<root_resource_id>/<resource_id>/<name>"
	s := &dynamic.NodeGroupSpec{
		Name:               url,
		MinSize:            2,
		MaxSize:            3,
		SupportScaleToZero: false,
	}
    _, manager := newTestOSManager(t)
    _, err := manager.buildASGFromSpec(s)
	assert.Error(t, err)
	assert.Equal(t, fmt.Sprintf(
        "failed to parse asg url: %s got error: Wrong id: expected format %s, got %s",
        url, expected_format, url), err.Error())
}

func TestBuildASGFromSpecWithInvalidSpec(t *testing.T) {
	s := &dynamic.NodeGroupSpec{
		Name:               "Name",
		MinSize:            3,
		MaxSize:            2,
		SupportScaleToZero: false,
	}
    _, manager := newTestOSManager(t)
    _, err := manager.buildASGFromSpec(s)
	assert.Error(t, err)
	assert.Equal(t, "invalid node group spec: max size must be greater or equal to min size", err.Error())
}

func TestParseASGUrl(t *testing.T) {
    url := "project_id/root_resource_id/resource_id/name"
    osref, err := ParseASGUrl(url)
	assert.NoError(t, err)
	assert.Equal(t, "project_id", osref.Project)
	assert.Equal(t, "root_resource_id", osref.RootResource)
	assert.Equal(t, "resource_id", osref.Resource)
	assert.Equal(t, "name", osref.Name)
}

func TestParseWrongASGUrl(t *testing.T) {
    url := "project_id/root_resource_id/resource_id/name/extra_info"
    _, err := ParseASGUrl(url)
	assert.Error(t, err)
	assert.Equal(t, "Wrong id: expected format <project-id>/<root_resource_id>/<resource_id>/<name>, got project_id/root_resource_id/resource_id/name/extra_info", err.Error())
}

func TestFetchAutoASGsWithFetchFailed(t *testing.T) {
    osService, manager := newTestOSManager(t)
	osService.On("FetchASGsWithName",
    mock.AnythingOfType("*regexp.Regexp")).Return(
        []string {"<project-id>/<root_stack_id>/<stack_id>/<name>"}, errors.New("failed")).Once()
    manager.asgAutoDiscoverySpecs = []cloudprovider.OSASGAutoDiscoveryConfig{
		{Re: regexp.MustCompile("UNUSED"), MinSize: 2, MaxSize: 3},
	}
    err := manager.fetchAutoASGs()
	assert.Error(t, err)
	assert.Equal(t, "cannot autodiscover managed instance groups: failed", err.Error())
	mock.AssertExpectationsForObjects(t, osService)
}

func TestFetchAutoASGsWithRegenerateInstancesCache(t *testing.T) {
    osService, manager := newTestOSManager(t)
	osService.On("FetchASGsWithName",
    mock.AnythingOfType("*regexp.Regexp")).Return(
        []string {"project-id/root_stack_id/stack_id/name"}, nil).Once()
	osService.On("FetchASGInstances",
    mock.AnythingOfType("openstack.OpenStackRef")).Return(
        []OpenStackRef{
            {Name: "name", Project: "project-id",
            RootResource: "root_resource_id", Resource: "stack_id"}}, nil).Once()
    manager.asgAutoDiscoverySpecs = []cloudprovider.OSASGAutoDiscoveryConfig{
		{Re: regexp.MustCompile("UNUSED"), MinSize: 2, MaxSize: 3},
	}
    manager.explicitlyConfigured[
    OpenStackRef{
        Name: "name", Project: "project-id",
        RootResource: "root_resource_id", Resource: "stack_id"}] = true
    err := manager.fetchAutoASGs()
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, osService)
}

func TestFetchAutoASGsWithUnRegenerateCache(t *testing.T) {
    osService, manager := newTestOSManager(t)
	osService.On("FetchASGsWithName",
    mock.AnythingOfType("*regexp.Regexp")).Return(
        []string {"project-id/root_stack_id/stack_id/name"}, nil).Once()
	osService.On("FetchASGInstances",
    mock.AnythingOfType("openstack.OpenStackRef")).Return(
        []OpenStackRef{
            {Name: "name", Project: "project-id",
            RootResource: "root_resource_id", Resource: "stack_id"}}, nil).Once()
    manager.asgAutoDiscoverySpecs = []cloudprovider.OSASGAutoDiscoveryConfig{
		{Re: regexp.MustCompile("UNUSED"), MinSize: 2, MaxSize: 3},
	}
    asg := &openstackASG{openstackRef: OpenStackRef{Name: "asg_not_exists"}}
    manager.cache.asgs = append(manager.cache.asgs, &ASGInformation{
            Config: asg,
        })
    manager.explicitlyConfigured[
    OpenStackRef{
        Name: "name", Project: "project-id",
        RootResource: "root_resource_id", Resource: "stack_id"}] = true
    err := manager.fetchAutoASGs()

	for i := range manager.cache.asgs {
        assert.NotEqual(t, asg.OpenStackRef(), manager.cache.asgs[i].Config.OpenStackRef())
	}

	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, osService)
}

func TestFetchAutoASGsWithNoFetch(t *testing.T) {
    osService, manager := newTestOSManager(t)
	osService.On("FetchASGsWithName",
    mock.AnythingOfType("*regexp.Regexp")).Return(
        []string {}, nil).Once()
    manager.asgAutoDiscoverySpecs = []cloudprovider.OSASGAutoDiscoveryConfig{
		{Re: regexp.MustCompile("UNUSED"), MinSize: 2, MaxSize: 3},
	}
    err := manager.fetchAutoASGs()
    assert.Equal(t, map[OpenStackRef]ASG{}, manager.cache.instancesCache)

	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, osService)
}
