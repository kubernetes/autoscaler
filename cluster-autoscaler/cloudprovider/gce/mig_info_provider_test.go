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

package gce

import (
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	gce "google.golang.org/api/compute/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/utils/units"
)

var (
	errFetchMig             = errors.New("fetch migs error")
	errFetchMigInstances    = errors.New("fetch mig instances error")
	errFetchMigTargetSize   = errors.New("fetch mig target size error")
	errFetchMigBaseName     = errors.New("fetch mig basename error")
	errFetchMigTemplateName = errors.New("fetch mig template name error")
	errFetchMigTemplate     = errors.New("fetch mig template error")
	errFetchMachineType     = errors.New("fetch machine type error")

	mig = &gceMig{
		gceRef: GceRef{
			Project: "project",
			Zone:    "us-test1",
			Name:    "mig",
		},
	}
)

type mockAutoscalingGceClient struct {
	fetchMigs            func(string) ([]*gce.InstanceGroupManager, error)
	fetchMigTargetSize   func(GceRef) (int64, error)
	fetchMigBasename     func(GceRef) (string, error)
	fetchMigInstances    func(GceRef) ([]cloudprovider.Instance, error)
	fetchMigTemplateName func(GceRef) (string, error)
	fetchMigTemplate     func(GceRef, string) (*gce.InstanceTemplate, error)
	fetchMachineType     func(string, string) (*gce.MachineType, error)
}

func (client *mockAutoscalingGceClient) FetchMachineType(zone, machineName string) (*gce.MachineType, error) {
	return client.fetchMachineType(zone, machineName)
}

func (client *mockAutoscalingGceClient) FetchMachineTypes(_ string) ([]*gce.MachineType, error) {
	return nil, nil
}

func (client *mockAutoscalingGceClient) FetchAllMigs(zone string) ([]*gce.InstanceGroupManager, error) {
	return client.fetchMigs(zone)
}

func (client *mockAutoscalingGceClient) FetchMigTargetSize(migRef GceRef) (int64, error) {
	return client.fetchMigTargetSize(migRef)
}

func (client *mockAutoscalingGceClient) FetchMigBasename(migRef GceRef) (string, error) {
	return client.fetchMigBasename(migRef)
}

func (client *mockAutoscalingGceClient) FetchMigInstances(migRef GceRef) ([]cloudprovider.Instance, error) {
	return client.fetchMigInstances(migRef)
}

func (client *mockAutoscalingGceClient) FetchMigTemplateName(migRef GceRef) (string, error) {
	return client.fetchMigTemplateName(migRef)
}

func (client *mockAutoscalingGceClient) FetchMigTemplate(migRef GceRef, templateName string) (*gce.InstanceTemplate, error) {
	return client.fetchMigTemplate(migRef, templateName)
}

func (client *mockAutoscalingGceClient) FetchMigsWithName(_ string, _ *regexp.Regexp) ([]string, error) {
	return nil, nil
}

func (client *mockAutoscalingGceClient) FetchZones(_ string) ([]string, error) {
	return nil, nil
}

func (client *mockAutoscalingGceClient) FetchAvailableCpuPlatforms() (map[string][]string, error) {
	return nil, nil
}

func (client *mockAutoscalingGceClient) FetchReservations() ([]*gce.Reservation, error) {
	return nil, nil
}

func (client *mockAutoscalingGceClient) ResizeMig(_ GceRef, _ int64) error {
	return nil
}

func (client *mockAutoscalingGceClient) DeleteInstances(_ GceRef, _ []GceRef) error {
	return nil
}

func (client *mockAutoscalingGceClient) CreateInstances(_ GceRef, _ string, _ int64, _ []string) error {
	return nil
}

func TestMigInfoProviderGetMigForInstance(t *testing.T) {
	instance := cloudprovider.Instance{
		Id: "gce://project/us-test1/base-instance-name-abcd",
	}
	instanceRef, err := GceRefFromProviderId(instance.Id)
	assert.Nil(t, err)

	testCases := []struct {
		name                     string
		instanceRef              GceRef
		cache                    *GceCache
		fetchMigInstances        func(GceRef) ([]cloudprovider.Instance, error)
		fetchMigBasename         func(GceRef) (string, error)
		expectedMig              Mig
		expectedErr              error
		expectedCachedMigRef     GceRef
		expectedCached           bool
		expectedCachedMigUnknown bool
	}{
		{
			name: "instance mig ref and mig in cache",
			cache: &GceCache{
				migs:           map[GceRef]Mig{mig.GceRef(): mig},
				instancesToMig: map[GceRef]GceRef{instanceRef: mig.GceRef()},
			},
			expectedMig:          mig,
			expectedCachedMigRef: mig.GceRef(),
			expectedCached:       true,
		},
		{
			name: "only instance mig ref in cache",
			cache: &GceCache{
				instancesToMig: map[GceRef]GceRef{instanceRef: mig.GceRef()},
			},
			expectedMig:          nil,
			expectedErr:          fmt.Errorf("instance %v belongs to unregistered mig %v", instanceRef, mig.GceRef()),
			expectedCachedMigRef: mig.GceRef(),
			expectedCached:       true,
		},
		{
			name: "instance mig unknown in cache",
			cache: &GceCache{
				instancesFromUnknownMig: map[GceRef]bool{instanceRef: true},
			},
			expectedCachedMigUnknown: true,
		},
		{
			name: "mig from cache fill",
			cache: &GceCache{
				migs:             map[GceRef]Mig{mig.GceRef(): mig},
				instances:        map[GceRef][]cloudprovider.Instance{},
				instancesToMig:   map[GceRef]GceRef{},
				migBaseNameCache: map[GceRef]string{mig.GceRef(): "base-instance-name"},
			},
			fetchMigInstances:    fetchMigInstancesConst([]cloudprovider.Instance{instance}),
			expectedMig:          mig,
			expectedCachedMigRef: mig.GceRef(),
			expectedCached:       true,
		},
		{
			name: "mig and basename from cache fill",
			cache: &GceCache{
				migs:             map[GceRef]Mig{mig.GceRef(): mig},
				instances:        map[GceRef][]cloudprovider.Instance{},
				instancesToMig:   map[GceRef]GceRef{},
				migBaseNameCache: map[GceRef]string{},
			},
			fetchMigInstances:    fetchMigInstancesConst([]cloudprovider.Instance{instance}),
			fetchMigBasename:     fetchMigBasenameConst("base-instance-name"),
			expectedMig:          mig,
			expectedCachedMigRef: mig.GceRef(),
			expectedCached:       true,
		},
		{
			name: "unknown mig from cache fill",
			cache: &GceCache{
				migs:                    map[GceRef]Mig{mig.GceRef(): mig},
				instances:               map[GceRef][]cloudprovider.Instance{},
				instancesFromUnknownMig: map[GceRef]bool{},
				migBaseNameCache:        map[GceRef]string{mig.GceRef(): "base-instance-name"},
			},
			fetchMigInstances:        fetchMigInstancesConst([]cloudprovider.Instance{}),
			expectedCachedMigUnknown: true,
		},
		{
			name: "no candidate mig during cache fill",
			cache: &GceCache{
				migs:             map[GceRef]Mig{mig.GceRef(): mig},
				migBaseNameCache: map[GceRef]string{mig.GceRef(): "different-base-instance-name"},
				instancesToMig:   map[GceRef]GceRef{},
			},
			fetchMigInstances: fetchMigInstancesConst([]cloudprovider.Instance{instance}),
		},
		{
			name: "fetch instances error during cache fill",
			cache: &GceCache{
				migs:             map[GceRef]Mig{mig.GceRef(): mig},
				instancesToMig:   map[GceRef]GceRef{},
				migBaseNameCache: map[GceRef]string{mig.GceRef(): "base-instance-name"},
			},
			fetchMigInstances: fetchMigInstancesFail,
			expectedErr:       errFetchMigInstances,
		},
		{
			name: "fetch basename error during cache fill",
			cache: &GceCache{
				migs:             map[GceRef]Mig{mig.GceRef(): mig},
				instancesToMig:   map[GceRef]GceRef{},
				migBaseNameCache: map[GceRef]string{},
			},
			fetchMigBasename: fetchMigBasenameFail,
			expectedErr:      nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := &mockAutoscalingGceClient{
				fetchMigBasename:  tc.fetchMigBasename,
				fetchMigInstances: tc.fetchMigInstances,
				fetchMigs:         fetchMigsConst(nil),
			}
			migLister := NewMigLister(tc.cache)
			provider := NewCachingMigInfoProvider(tc.cache, migLister, client, mig.GceRef().Project, 1, 0*time.Second)

			mig, err := provider.GetMigForInstance(instanceRef)

			cachedMigRef, cached := tc.cache.GetMigForInstance(instanceRef)
			cachedMigUnknown := tc.cache.IsMigUnknownForInstance(instanceRef)

			assert.Equal(t, tc.expectedMig, mig)
			assert.Equal(t, tc.expectedErr, err)

			assert.Equal(t, tc.expectedCachedMigRef, cachedMigRef)
			assert.Equal(t, tc.expectedCached, cached)
			assert.Equal(t, tc.expectedCachedMigUnknown, cachedMigUnknown)
		})
	}
}

func TestGetMigInstances(t *testing.T) {
	instances := []cloudprovider.Instance{
		{Id: "gce://project/us-test1/base-instance-name-abcd"},
		{Id: "gce://project/us-test1/base-instance-name-efgh"},
	}

	testCases := []struct {
		name                    string
		cache                   *GceCache
		fetchMigInstances       func(GceRef) ([]cloudprovider.Instance, error)
		expectedInstances       []cloudprovider.Instance
		expectedErr             error
		expectedCachedInstances []cloudprovider.Instance
		expectedCached          bool
	}{
		{
			name: "instances in cache",
			cache: &GceCache{
				migs:      map[GceRef]Mig{mig.GceRef(): mig},
				instances: map[GceRef][]cloudprovider.Instance{mig.GceRef(): instances},
			},
			expectedInstances:       instances,
			expectedCachedInstances: instances,
			expectedCached:          true,
		},
		{
			name: "instances cache fill",
			cache: &GceCache{
				migs:           map[GceRef]Mig{mig.GceRef(): mig},
				instances:      map[GceRef][]cloudprovider.Instance{},
				instancesToMig: map[GceRef]GceRef{},
			},
			fetchMigInstances:       fetchMigInstancesConst(instances),
			expectedInstances:       instances,
			expectedCachedInstances: instances,
			expectedCached:          true,
		},
		{
			name: "error during instances cache fill",
			cache: &GceCache{
				migs:      map[GceRef]Mig{mig.GceRef(): mig},
				instances: map[GceRef][]cloudprovider.Instance{},
			},
			fetchMigInstances:       fetchMigInstancesFail,
			expectedErr:             errFetchMigInstances,
			expectedCachedInstances: []cloudprovider.Instance{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := &mockAutoscalingGceClient{
				fetchMigInstances: tc.fetchMigInstances,
			}
			migLister := NewMigLister(tc.cache)
			provider := NewCachingMigInfoProvider(tc.cache, migLister, client, mig.GceRef().Project, 1, 0*time.Second)
			instances, err := provider.GetMigInstances(mig.GceRef())
			cachedInstances, cached := tc.cache.GetMigInstances(mig.GceRef())

			assert.Equal(t, tc.expectedInstances, instances)
			assert.Equal(t, tc.expectedErr, err)

			assert.Equal(t, tc.expectedCachedInstances, cachedInstances)
			assert.Equal(t, tc.expectedCached, cached)
		})
	}
}

func TestRegenerateMigInstancesCache(t *testing.T) {
	otherMig := &gceMig{
		gceRef: GceRef{
			Project: "project",
			Zone:    "us-test1",
			Name:    "other-mig",
		},
	}

	instances := []cloudprovider.Instance{
		{Id: "gce://project/us-test1/base-instance-name-abcd"},
		{Id: "gce://project/us-test1/base-instance-name-efgh"},
	}
	otherInstances := []cloudprovider.Instance{
		{Id: "gce://project/us-test1/other-base-instance-name-abcd"},
		{Id: "gce://project/us-test1/other-base-instance-name-efgh"},
	}

	var instancesRefs, otherInstancesRefs []GceRef
	for _, instance := range instances {
		instanceRef, err := GceRefFromProviderId(instance.Id)
		assert.Nil(t, err)
		instancesRefs = append(instancesRefs, instanceRef)
	}
	for _, instance := range otherInstances {
		instanceRef, err := GceRefFromProviderId(instance.Id)
		assert.Nil(t, err)
		otherInstancesRefs = append(otherInstancesRefs, instanceRef)
	}

	testCases := []struct {
		name                   string
		cache                  *GceCache
		fetchMigInstances      func(GceRef) ([]cloudprovider.Instance, error)
		expectedErr            error
		expectedMigInstances   map[GceRef][]cloudprovider.Instance
		expectedInstancesToMig map[GceRef]GceRef
	}{
		{
			name: "fill empty cache for one mig",
			cache: &GceCache{
				migs:           map[GceRef]Mig{mig.GceRef(): mig},
				instances:      map[GceRef][]cloudprovider.Instance{},
				instancesToMig: map[GceRef]GceRef{},
			},
			fetchMigInstances: fetchMigInstancesConst(instances),
			expectedMigInstances: map[GceRef][]cloudprovider.Instance{
				mig.GceRef(): instances,
			},
			expectedInstancesToMig: map[GceRef]GceRef{
				instancesRefs[0]: mig.GceRef(),
				instancesRefs[1]: mig.GceRef(),
			},
		},
		{
			name: "fill empty cache for two migs",
			cache: &GceCache{
				migs: map[GceRef]Mig{
					mig.GceRef():      mig,
					otherMig.GceRef(): otherMig,
				},
				instances:      map[GceRef][]cloudprovider.Instance{},
				instancesToMig: map[GceRef]GceRef{},
			},
			fetchMigInstances: fetchMigInstancesMapping(map[GceRef][]cloudprovider.Instance{
				mig.GceRef():      instances,
				otherMig.GceRef(): otherInstances,
			}),
			expectedMigInstances: map[GceRef][]cloudprovider.Instance{
				mig.GceRef():      instances,
				otherMig.GceRef(): otherInstances,
			},
			expectedInstancesToMig: map[GceRef]GceRef{
				instancesRefs[0]:      mig.GceRef(),
				instancesRefs[1]:      mig.GceRef(),
				otherInstancesRefs[0]: otherMig.GceRef(),
				otherInstancesRefs[1]: otherMig.GceRef(),
			},
		},
		{
			name: "clear cache for removed mig",
			cache: &GceCache{
				migs: map[GceRef]Mig{
					mig.GceRef(): mig,
				},
				instances: map[GceRef][]cloudprovider.Instance{
					mig.GceRef():      instances,
					otherMig.GceRef(): otherInstances,
				},
				instancesToMig: map[GceRef]GceRef{
					instancesRefs[0]:      mig.GceRef(),
					instancesRefs[1]:      mig.GceRef(),
					otherInstancesRefs[0]: otherMig.GceRef(),
					otherInstancesRefs[1]: otherMig.GceRef(),
				},
			},
			fetchMigInstances: fetchMigInstancesMapping(map[GceRef][]cloudprovider.Instance{
				mig.GceRef():      instances,
				otherMig.GceRef(): otherInstances,
			}),
			expectedMigInstances: map[GceRef][]cloudprovider.Instance{
				mig.GceRef(): instances,
			},
			expectedInstancesToMig: map[GceRef]GceRef{
				instancesRefs[0]: mig.GceRef(),
				instancesRefs[1]: mig.GceRef(),
			},
		},
		{
			name: "override cache for changed instances",
			cache: &GceCache{
				migs: map[GceRef]Mig{
					mig.GceRef(): mig,
				},
				instances: map[GceRef][]cloudprovider.Instance{
					mig.GceRef(): instances,
				},
				instancesToMig: map[GceRef]GceRef{
					instancesRefs[0]: mig.GceRef(),
					instancesRefs[1]: mig.GceRef(),
				},
			},
			fetchMigInstances: fetchMigInstancesConst(otherInstances),
			expectedMigInstances: map[GceRef][]cloudprovider.Instance{
				mig.GceRef(): otherInstances,
			},
			expectedInstancesToMig: map[GceRef]GceRef{
				otherInstancesRefs[0]: mig.GceRef(),
				otherInstancesRefs[1]: mig.GceRef(),
			},
		},
		{
			name: "refill mig instances error",
			cache: &GceCache{
				migs: map[GceRef]Mig{
					mig.GceRef(): mig,
				},
				instances:      map[GceRef][]cloudprovider.Instance{},
				instancesToMig: map[GceRef]GceRef{},
			},
			fetchMigInstances: fetchMigInstancesFail,
			expectedErr:       errFetchMigInstances,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := &mockAutoscalingGceClient{
				fetchMigInstances: tc.fetchMigInstances,
			}
			migLister := NewMigLister(tc.cache)
			provider := NewCachingMigInfoProvider(tc.cache, migLister, client, mig.GceRef().Project, 1, 0*time.Second)
			err := provider.RegenerateMigInstancesCache()

			assert.Equal(t, tc.expectedErr, err)
			if tc.expectedErr == nil {
				assert.Equal(t, tc.expectedMigInstances, tc.cache.instances)
				assert.Equal(t, tc.expectedInstancesToMig, tc.cache.instancesToMig)
			}
		})
	}
}

func TestGetMigTargetSize(t *testing.T) {
	targetSize := int64(42)
	instanceGroupManager := &gce.InstanceGroupManager{
		Zone:       mig.GceRef().Zone,
		Name:       mig.GceRef().Name,
		TargetSize: targetSize,
	}

	testCases := []struct {
		name               string
		cache              *GceCache
		fetchMigs          func(string) ([]*gce.InstanceGroupManager, error)
		fetchMigTargetSize func(GceRef) (int64, error)
		expectedTargetSize int64
		expectedErr        error
	}{
		{
			name: "target size in cache",
			cache: &GceCache{
				migs:               map[GceRef]Mig{mig.GceRef(): mig},
				migTargetSizeCache: map[GceRef]int64{mig.GceRef(): targetSize},
			},
			expectedTargetSize: targetSize,
		},
		{
			name:               "target size from cache fill",
			cache:              emptyCache(),
			fetchMigs:          fetchMigsConst([]*gce.InstanceGroupManager{instanceGroupManager}),
			expectedTargetSize: targetSize,
		},
		{
			name:               "cache fill without mig, fallback success",
			cache:              emptyCache(),
			fetchMigs:          fetchMigsConst([]*gce.InstanceGroupManager{}),
			fetchMigTargetSize: fetchMigTargetSizeConst(targetSize),
			expectedTargetSize: targetSize,
		},
		{
			name:               "cache fill failure, fallback success",
			cache:              emptyCache(),
			fetchMigs:          fetchMigsFail,
			fetchMigTargetSize: fetchMigTargetSizeConst(targetSize),
			expectedTargetSize: targetSize,
		},
		{
			name:               "cache fill without mig, fallback failure",
			cache:              emptyCache(),
			fetchMigs:          fetchMigsConst([]*gce.InstanceGroupManager{}),
			fetchMigTargetSize: fetchMigTargetSizeFail,
			expectedErr:        errFetchMigTargetSize,
		},
		{
			name:               "cache fill failure, fallback failure",
			cache:              emptyCache(),
			fetchMigs:          fetchMigsFail,
			fetchMigTargetSize: fetchMigTargetSizeFail,
			expectedErr:        errFetchMigTargetSize,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := &mockAutoscalingGceClient{
				fetchMigs:          tc.fetchMigs,
				fetchMigTargetSize: tc.fetchMigTargetSize,
			}
			migLister := NewMigLister(tc.cache)
			provider := NewCachingMigInfoProvider(tc.cache, migLister, client, mig.GceRef().Project, 1, 0*time.Second)

			targetSize, err := provider.GetMigTargetSize(mig.GceRef())
			cachedTargetSize, found := tc.cache.GetMigTargetSize(mig.GceRef())

			assert.Equal(t, tc.expectedErr, err)
			assert.Equal(t, tc.expectedErr == nil, found)
			if tc.expectedErr == nil {
				assert.Equal(t, tc.expectedTargetSize, targetSize)
				assert.Equal(t, tc.expectedTargetSize, cachedTargetSize)
			}
		})
	}
}

func TestGetMigBasename(t *testing.T) {
	basename := "base-instance-name"
	instanceGroupManager := &gce.InstanceGroupManager{
		Zone:             mig.GceRef().Zone,
		Name:             mig.GceRef().Name,
		BaseInstanceName: basename,
	}

	testCases := []struct {
		name             string
		cache            *GceCache
		fetchMigs        func(string) ([]*gce.InstanceGroupManager, error)
		fetchMigBasename func(GceRef) (string, error)
		expectedBasename string
		expectedErr      error
	}{
		{
			name: "basename in cache",
			cache: &GceCache{
				migs:             map[GceRef]Mig{mig.GceRef(): mig},
				migBaseNameCache: map[GceRef]string{mig.GceRef(): basename},
			},
			expectedBasename: basename,
		},
		{
			name:             "target size from cache fill",
			cache:            emptyCache(),
			fetchMigs:        fetchMigsConst([]*gce.InstanceGroupManager{instanceGroupManager}),
			expectedBasename: basename,
		},
		{
			name:             "cache fill without mig, fallback success",
			cache:            emptyCache(),
			fetchMigs:        fetchMigsConst([]*gce.InstanceGroupManager{}),
			fetchMigBasename: fetchMigBasenameConst(basename),
			expectedBasename: basename,
		},
		{
			name:             "cache fill failure, fallback success",
			cache:            emptyCache(),
			fetchMigs:        fetchMigsFail,
			fetchMigBasename: fetchMigBasenameConst(basename),
			expectedBasename: basename,
		},
		{
			name:             "cache fill without mig, fallback failure",
			cache:            emptyCache(),
			fetchMigs:        fetchMigsConst([]*gce.InstanceGroupManager{}),
			fetchMigBasename: fetchMigBasenameFail,
			expectedErr:      errFetchMigBaseName,
		},
		{
			name:             "cache fill failure, fallback failure",
			cache:            emptyCache(),
			fetchMigs:        fetchMigsFail,
			fetchMigBasename: fetchMigBasenameFail,
			expectedErr:      errFetchMigBaseName,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := &mockAutoscalingGceClient{
				fetchMigs:        tc.fetchMigs,
				fetchMigBasename: tc.fetchMigBasename,
			}
			migLister := NewMigLister(tc.cache)
			provider := NewCachingMigInfoProvider(tc.cache, migLister, client, mig.GceRef().Project, 1, 0*time.Second)

			basename, err := provider.GetMigBasename(mig.GceRef())
			cachedBasename, found := tc.cache.GetMigBasename(mig.GceRef())

			assert.Equal(t, tc.expectedErr, err)
			assert.Equal(t, tc.expectedErr == nil, found)
			if tc.expectedErr == nil {
				assert.Equal(t, tc.expectedBasename, basename)
				assert.Equal(t, tc.expectedBasename, cachedBasename)
			}
		})
	}
}

func TestGetMigInstanceTemplateName(t *testing.T) {
	templateName := "template-name"
	instanceGroupManager := &gce.InstanceGroupManager{
		Zone:             mig.GceRef().Zone,
		Name:             mig.GceRef().Name,
		InstanceTemplate: templateName,
	}

	testCases := []struct {
		name                 string
		cache                *GceCache
		fetchMigs            func(string) ([]*gce.InstanceGroupManager, error)
		fetchMigTemplateName func(GceRef) (string, error)
		expectedTemplateName string
		expectedErr          error
	}{
		{
			name: "template name in cache",
			cache: &GceCache{
				migs:                      map[GceRef]Mig{mig.GceRef(): mig},
				instanceTemplateNameCache: map[GceRef]string{mig.GceRef(): templateName},
			},
			expectedTemplateName: templateName,
		},
		{
			name:                 "target size from cache fill",
			cache:                emptyCache(),
			fetchMigs:            fetchMigsConst([]*gce.InstanceGroupManager{instanceGroupManager}),
			expectedTemplateName: templateName,
		},
		{
			name:                 "cache fill without mig, fallback success",
			cache:                emptyCache(),
			fetchMigs:            fetchMigsConst([]*gce.InstanceGroupManager{}),
			fetchMigTemplateName: fetchMigTemplateNameConst(templateName),
			expectedTemplateName: templateName,
		},
		{
			name:                 "cache fill failure, fallback success",
			cache:                emptyCache(),
			fetchMigs:            fetchMigsFail,
			fetchMigTemplateName: fetchMigTemplateNameConst(templateName),
			expectedTemplateName: templateName,
		},
		{
			name:                 "cache fill without mig, fallback failure",
			cache:                emptyCache(),
			fetchMigs:            fetchMigsConst([]*gce.InstanceGroupManager{}),
			fetchMigTemplateName: fetchMigTemplateNameFail,
			expectedErr:          errFetchMigTemplateName,
		},
		{
			name:                 "cache fill failure, fallback failure",
			cache:                emptyCache(),
			fetchMigs:            fetchMigsFail,
			fetchMigTemplateName: fetchMigTemplateNameFail,
			expectedErr:          errFetchMigTemplateName,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := &mockAutoscalingGceClient{
				fetchMigs:            tc.fetchMigs,
				fetchMigTemplateName: tc.fetchMigTemplateName,
			}
			migLister := NewMigLister(tc.cache)
			provider := NewCachingMigInfoProvider(tc.cache, migLister, client, mig.GceRef().Project, 1, 0*time.Second)

			templateName, err := provider.GetMigInstanceTemplateName(mig.GceRef())
			cachedTemplateName, found := tc.cache.GetMigInstanceTemplateName(mig.GceRef())

			assert.Equal(t, tc.expectedErr, err)
			assert.Equal(t, tc.expectedErr == nil, found)
			if tc.expectedErr == nil {
				assert.Equal(t, tc.expectedTemplateName, templateName)
				assert.Equal(t, tc.expectedTemplateName, cachedTemplateName)
			}
		})
	}
}

func TestGetMigInstanceTemplate(t *testing.T) {
	templateName := "template-name"
	template := &gce.InstanceTemplate{
		Name:        templateName,
		Description: "instance template",
	}
	oldTemplate := &gce.InstanceTemplate{
		Name:        "old-template-name",
		Description: "old instance template",
	}

	testCases := []struct {
		name                   string
		cache                  *GceCache
		fetchMigs              func(string) ([]*gce.InstanceGroupManager, error)
		fetchMigTemplateName   func(GceRef) (string, error)
		fetchMigTemplate       func(GceRef, string) (*gce.InstanceTemplate, error)
		expectedTemplate       *gce.InstanceTemplate
		expectedCachedTemplate *gce.InstanceTemplate
		expectedErr            error
	}{
		{
			name: "template in cache",
			cache: &GceCache{
				migs:                      map[GceRef]Mig{mig.GceRef(): mig},
				instanceTemplateNameCache: map[GceRef]string{mig.GceRef(): templateName},
				instanceTemplatesCache:    map[GceRef]*gce.InstanceTemplate{mig.GceRef(): template},
			},
			expectedTemplate:       template,
			expectedCachedTemplate: template,
		},
		{
			name: "cache without template, fetch success",
			cache: &GceCache{
				migs:                      map[GceRef]Mig{mig.GceRef(): mig},
				instanceTemplateNameCache: map[GceRef]string{mig.GceRef(): templateName},
				instanceTemplatesCache:    make(map[GceRef]*gce.InstanceTemplate),
			},
			fetchMigTemplate:       fetchMigTemplateConst(template),
			expectedTemplate:       template,
			expectedCachedTemplate: template,
		},
		{
			name: "cache with old template, fetch success",
			cache: &GceCache{
				migs:                      map[GceRef]Mig{mig.GceRef(): mig},
				instanceTemplateNameCache: map[GceRef]string{mig.GceRef(): templateName},
				instanceTemplatesCache:    map[GceRef]*gce.InstanceTemplate{mig.GceRef(): oldTemplate},
			},
			fetchMigTemplate:       fetchMigTemplateConst(template),
			expectedTemplate:       template,
			expectedCachedTemplate: template,
		},
		{
			name: "cache without template, fetch failure",
			cache: &GceCache{
				migs:                      map[GceRef]Mig{mig.GceRef(): mig},
				instanceTemplateNameCache: map[GceRef]string{mig.GceRef(): templateName},
				instanceTemplatesCache:    make(map[GceRef]*gce.InstanceTemplate),
			},
			fetchMigTemplate: fetchMigTemplateFail,
			expectedErr:      errFetchMigTemplate,
		},
		{
			name: "cache with old template, fetch failure",
			cache: &GceCache{
				migs:                      map[GceRef]Mig{mig.GceRef(): mig},
				instanceTemplateNameCache: map[GceRef]string{mig.GceRef(): templateName},
				instanceTemplatesCache:    map[GceRef]*gce.InstanceTemplate{mig.GceRef(): oldTemplate},
			},
			fetchMigTemplate:       fetchMigTemplateFail,
			expectedCachedTemplate: oldTemplate,
			expectedErr:            errFetchMigTemplate,
		},
		{
			name:                 "template name fetch failure",
			cache:                emptyCache(),
			fetchMigs:            fetchMigsFail,
			fetchMigTemplateName: fetchMigTemplateNameFail,
			expectedErr:          errFetchMigTemplateName,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := &mockAutoscalingGceClient{
				fetchMigs:            tc.fetchMigs,
				fetchMigTemplateName: tc.fetchMigTemplateName,
				fetchMigTemplate:     tc.fetchMigTemplate,
			}
			migLister := NewMigLister(tc.cache)
			provider := NewCachingMigInfoProvider(tc.cache, migLister, client, mig.GceRef().Project, 1, 0*time.Second)

			template, err := provider.GetMigInstanceTemplate(mig.GceRef())
			cachedTemplate, found := tc.cache.GetMigInstanceTemplate(mig.GceRef())

			assert.Equal(t, tc.expectedErr, err)
			if tc.expectedErr == nil {
				assert.Equal(t, tc.expectedTemplate, template)
			}

			assert.Equal(t, tc.expectedCachedTemplate != nil, found)
			if tc.expectedCachedTemplate != nil {
				assert.Equal(t, tc.expectedCachedTemplate, cachedTemplate)
			}
		})
	}
}

func TestGetMigMachineType(t *testing.T) {
	knownZone := "us-cache1-a"
	unknownZone := "us-nocache42-c"
	testCases := []struct {
		name             string
		machine          string
		zone             string
		fetchMachineType func(string, string) (*gce.MachineType, error)
		cpu              int64
		memory           int64
		expectCpu        int64
		expectMemory     int64
		expectError      bool
	}{
		{
			name:         "custom machine",
			machine:      "custom-8-2",
			zone:         unknownZone,
			expectCpu:    8,
			expectMemory: 2 * units.MiB,
		},
		{
			name:         "machine in cache",
			machine:      "n1-standard-1",
			zone:         knownZone,
			cpu:          1,
			memory:       2,
			expectCpu:    1,
			expectMemory: 2 * units.MiB,
		},
		{
			name:             "machine not in cache",
			machine:          "n1-standard-2",
			zone:             unknownZone,
			fetchMachineType: fetchMachineTypeConst("n1-standard-2", 2, 3840),
			expectCpu:        2,
			expectMemory:     3840 * units.MiB,
		},
		{
			name:             "machine not in cache, request error",
			machine:          "n1-standard-1",
			zone:             unknownZone,
			fetchMachineType: fetchMachineTypeFail,
			expectError:      true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mig := &gceMig{
				gceRef: GceRef{
					Project: "project",
					Zone:    tc.zone,
					Name:    "mig",
				},
			}
			cache := &GceCache{
				instanceTemplateNameCache: map[GceRef]string{mig.GceRef(): "template"},
				instanceTemplatesCache: map[GceRef]*gce.InstanceTemplate{
					mig.GceRef(): {
						Name: "template",
						Properties: &gce.InstanceProperties{
							MachineType: tc.machine,
						},
					},
				},
				machinesCache: map[MachineTypeKey]MachineType{
					{knownZone, tc.machine}: {
						Name:   tc.machine,
						CPU:    tc.cpu,
						Memory: tc.memory * units.MiB,
					},
				},
			}
			client := &mockAutoscalingGceClient{
				fetchMachineType: tc.fetchMachineType,
			}
			migLister := NewMigLister(cache)
			provider := NewCachingMigInfoProvider(cache, migLister, client, mig.GceRef().Project, 1, 0*time.Second)
			machine, err := provider.GetMigMachineType(mig.GceRef())
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectCpu, machine.CPU)
				assert.Equal(t, tc.expectMemory, machine.Memory)
			}
		})
	}
}

func TestMultipleGetMigInstanceCallsLimited(t *testing.T) {
	mig := &gceMig{
		gceRef: GceRef{
			Project: "project",
			Zone:    "zone",
			Name:    "base-instance-name",
		},
	}
	instance := cloudprovider.Instance{
		Id: "gce://project/zone/base-instance-name-abcd",
	}
	instanceRef, err := GceRefFromProviderId(instance.Id)
	assert.Nil(t, err)
	instance2 := cloudprovider.Instance{
		Id: "gce://project/zone/base-instance-name-abcd2",
	}
	instanceRef2, err := GceRefFromProviderId(instance2.Id)
	assert.Nil(t, err)
	now := time.Now()
	for name, tc := range map[string]struct {
		refreshRateDuration              time.Duration
		firstCallTime                    time.Time
		secondCallTime                   time.Time
		expectedCallsToFetchMigInstances int
	}{
		"0s refresh rate duration, refetch expected": {
			refreshRateDuration:              0 * time.Second,
			firstCallTime:                    now,
			secondCallTime:                   now,
			expectedCallsToFetchMigInstances: 2,
		},
		"5s refresh rate duration, 0.01s between calls, no refetch expected": {
			refreshRateDuration:              5 * time.Second,
			firstCallTime:                    now,
			secondCallTime:                   now.Add(10 * time.Millisecond),
			expectedCallsToFetchMigInstances: 1,
		},
		"0.01s refresh rate duration, 0.01s between calls, refetch expected": {
			refreshRateDuration:              10 * time.Millisecond,
			firstCallTime:                    now,
			secondCallTime:                   now.Add(11 * time.Millisecond),
			expectedCallsToFetchMigInstances: 2,
		},
	} {
		t.Run(name, func(t *testing.T) {
			cache := emptyCache()
			cache.migs = map[GceRef]Mig{
				mig.gceRef: mig,
			}
			cache.migBaseNameCache = map[GceRef]string{mig.GceRef(): "base-instance-name"}
			callCounter := make(map[GceRef]int)
			client := &mockAutoscalingGceClient{
				fetchMigInstances: fetchMigInstancesWithCounter(nil, callCounter),
			}
			migLister := NewMigLister(cache)
			ft := &fakeTime{}
			provider := &cachingMigInfoProvider{
				cache:                          cache,
				migLister:                      migLister,
				gceClient:                      client,
				projectId:                      projectId,
				concurrentGceRefreshes:         1,
				migInstancesMinRefreshWaitTime: tc.refreshRateDuration,
				migInstancesLastRefreshedInfo:  make(map[string]time.Time),
				timeProvider:                   ft,
			}
			ft.now = tc.firstCallTime
			_, err = provider.GetMigForInstance(instanceRef)
			assert.NoError(t, err)
			ft.now = tc.secondCallTime
			_, err = provider.GetMigForInstance(instanceRef2)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedCallsToFetchMigInstances, callCounter[mig.GceRef()])
		})
	}
}

type fakeTime struct {
	now time.Time
}

func (f *fakeTime) Now() time.Time {
	return f.now
}

func emptyCache() *GceCache {
	return &GceCache{
		migs:                      map[GceRef]Mig{mig.GceRef(): mig},
		instances:                 make(map[GceRef][]cloudprovider.Instance),
		migTargetSizeCache:        make(map[GceRef]int64),
		migBaseNameCache:          make(map[GceRef]string),
		instanceTemplateNameCache: make(map[GceRef]string),
		instanceTemplatesCache:    make(map[GceRef]*gce.InstanceTemplate),
		instancesFromUnknownMig:   make(map[GceRef]bool),
	}
}

func fetchMigsFail(_ string) ([]*gce.InstanceGroupManager, error) {
	return nil, errFetchMig
}

func fetchMigsConst(migs []*gce.InstanceGroupManager) func(string) ([]*gce.InstanceGroupManager, error) {
	return func(string) ([]*gce.InstanceGroupManager, error) {
		return migs, nil
	}
}

func fetchMigInstancesFail(_ GceRef) ([]cloudprovider.Instance, error) {
	return nil, errFetchMigInstances
}

func fetchMigInstancesConst(instances []cloudprovider.Instance) func(GceRef) ([]cloudprovider.Instance, error) {
	return func(GceRef) ([]cloudprovider.Instance, error) {
		return instances, nil
	}
}

func fetchMigInstancesWithCounter(instances []cloudprovider.Instance, migCounter map[GceRef]int) func(GceRef) ([]cloudprovider.Instance, error) {
	return func(ref GceRef) ([]cloudprovider.Instance, error) {
		migCounter[ref] = migCounter[ref] + 1
		return instances, nil
	}
}

func fetchMigInstancesMapping(instancesMapping map[GceRef][]cloudprovider.Instance) func(GceRef) ([]cloudprovider.Instance, error) {
	return func(migRef GceRef) ([]cloudprovider.Instance, error) {
		return instancesMapping[migRef], nil
	}
}

func fetchMigTargetSizeFail(_ GceRef) (int64, error) {
	return 0, errFetchMigTargetSize
}

func fetchMigTargetSizeConst(targetSize int64) func(GceRef) (int64, error) {
	return func(GceRef) (int64, error) {
		return targetSize, nil
	}
}

func fetchMigBasenameFail(_ GceRef) (string, error) {
	return "", errFetchMigBaseName
}

func fetchMigBasenameConst(basename string) func(GceRef) (string, error) {
	return func(GceRef) (string, error) {
		return basename, nil
	}
}

func fetchMigTemplateNameFail(_ GceRef) (string, error) {
	return "", errFetchMigTemplateName
}

func fetchMigTemplateNameConst(templateName string) func(GceRef) (string, error) {
	return func(GceRef) (string, error) {
		return templateName, nil
	}
}

func fetchMigTemplateFail(_ GceRef, _ string) (*gce.InstanceTemplate, error) {
	return nil, errFetchMigTemplate
}

func fetchMigTemplateConst(template *gce.InstanceTemplate) func(GceRef, string) (*gce.InstanceTemplate, error) {
	return func(GceRef, string) (*gce.InstanceTemplate, error) {
		return template, nil
	}
}

func fetchMachineTypeFail(_, _ string) (*gce.MachineType, error) {
	return nil, errFetchMachineType
}

func fetchMachineTypeConst(name string, cpu int64, mem int64) func(string, string) (*gce.MachineType, error) {
	return func(string, string) (*gce.MachineType, error) {
		return &gce.MachineType{
			Name:      name,
			GuestCpus: cpu,
			MemoryMb:  mem,
		}, nil
	}
}
