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
	errFetchMig                         = errors.New("fetch migs error")
	errFetchMigInstances                = errors.New("fetch mig instances error")
	errFetchMigTargetSize               = errors.New("fetch mig target size error")
	errFetchMigBaseName                 = errors.New("fetch mig basename error")
	errFetchMigTemplateName             = errors.New("fetch mig template name error")
	errFetchMigTemplate                 = errors.New("fetch mig template error")
	errFetchListManagedInstancesResults = errors.New("fetch ListManagedInstancesResults error")
	errFetchMachineType                 = errors.New("fetch machine type error")

	mig = &gceMig{
		gceRef: GceRef{
			Project: "project",
			Zone:    "us-test1",
			Name:    "mig",
		},
	}
)

type mockAutoscalingGceClient struct {
	fetchMigs                        func(string) ([]*gce.InstanceGroupManager, error)
	fetchMigTargetSize               func(GceRef) (int64, error)
	fetchMigBasename                 func(GceRef) (string, error)
	fetchMigInstances                func(GceRef) ([]GceInstance, error)
	fetchMigTemplateName             func(GceRef) (InstanceTemplateName, error)
	fetchMigTemplate                 func(GceRef, string, bool) (*gce.InstanceTemplate, error)
	fetchMachineType                 func(string, string) (*gce.MachineType, error)
	fetchListManagedInstancesResults func(GceRef) (string, error)
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

func (client *mockAutoscalingGceClient) FetchListManagedInstancesResults(migRef GceRef) (string, error) {
	return client.fetchListManagedInstancesResults(migRef)
}

func (client *mockAutoscalingGceClient) FetchMigInstances(migRef GceRef) ([]GceInstance, error) {
	return client.fetchMigInstances(migRef)
}

func (client *mockAutoscalingGceClient) FetchMigTemplateName(migRef GceRef) (InstanceTemplateName, error) {
	return client.fetchMigTemplateName(migRef)
}

func (client *mockAutoscalingGceClient) FetchMigTemplate(migRef GceRef, templateName string, regional bool) (*gce.InstanceTemplate, error) {
	return client.fetchMigTemplate(migRef, templateName, regional)
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

func (client *mockAutoscalingGceClient) FetchAvailableDiskTypes() (map[string][]string, error) {
	return nil, nil
}

func (client *mockAutoscalingGceClient) FetchReservations() ([]*gce.Reservation, error) {
	return nil, nil
}

func (client *mockAutoscalingGceClient) FetchReservationsInProject(_ string) ([]*gce.Reservation, error) {
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

func (client *mockAutoscalingGceClient) WaitForOperation(_, _, _, _ string) error {
	return nil
}

func TestFillMigInstances(t *testing.T) {
	migRef := GceRef{Project: "test", Zone: "zone-A", Name: "some-mig"}
	oldInstances := []GceInstance{
		{Instance: cloudprovider.Instance{Id: "gce://test/zone-A/some-mig-old-instance-1"}, NumericId: 1},
		{Instance: cloudprovider.Instance{Id: "gce://test/zone-A/some-mig-old-instance-2"}, NumericId: 2},
	}
	newInstances := []GceInstance{
		{Instance: cloudprovider.Instance{Id: "gce://test/zone-A/some-mig-new-instance-1"}, NumericId: 3},
		{Instance: cloudprovider.Instance{Id: "gce://test/zone-A/some-mig-new-instance-2"}, NumericId: 4},
	}

	timeNow := time.Now()
	timeRecent := timeNow.Add(-30 * time.Minute)
	timeOld := timeNow.Add(-90 * time.Minute)

	testCases := []struct {
		name            string
		cache           *GceCache
		wantClientCalls int
		wantInstances   []GceInstance
		wantUpdateTime  time.Time
	}{
		{
			name: "No instances in cache",
			cache: &GceCache{
				instances:           map[GceRef][]GceInstance{},
				instancesUpdateTime: map[GceRef]time.Time{},
				instancesToMig:      map[GceRef]GceRef{},
			},
			wantClientCalls: 1,
			wantInstances:   newInstances,
			wantUpdateTime:  timeNow,
		},
		{
			name: "Old instances in cache",
			cache: &GceCache{
				instances:           map[GceRef][]GceInstance{migRef: oldInstances},
				instancesUpdateTime: map[GceRef]time.Time{migRef: timeOld},
				instancesToMig:      map[GceRef]GceRef{},
			},
			wantClientCalls: 1,
			wantInstances:   newInstances,
			wantUpdateTime:  timeNow,
		},
		{
			name: "Recently updated instances in cache",
			cache: &GceCache{
				instances:           map[GceRef][]GceInstance{migRef: oldInstances},
				instancesUpdateTime: map[GceRef]time.Time{migRef: timeRecent},
				instancesToMig:      map[GceRef]GceRef{},
			},
			wantClientCalls: 0,
			wantInstances:   oldInstances,
			wantUpdateTime:  timeRecent,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			callCounter := make(map[GceRef]int)
			client := &mockAutoscalingGceClient{
				fetchMigInstances: fetchMigInstancesWithCounter(newInstances, callCounter),
			}

			provider, ok := NewCachingMigInfoProvider(tc.cache, NewMigLister(tc.cache), client, mig.GceRef().Project, 1, time.Hour).(*cachingMigInfoProvider)
			assert.True(t, ok)
			provider.timeProvider = &fakeTime{now: timeNow}

			assert.NoError(t, provider.fillMigInstances(migRef))
			assert.Equal(t, tc.wantClientCalls, callCounter[migRef])

			updateTime, updateTimeFound := tc.cache.GetMigInstancesUpdateTime(migRef)
			assert.True(t, updateTimeFound)
			assert.Equal(t, tc.wantUpdateTime, updateTime)

			instances, instancesFound := tc.cache.GetMigInstances(migRef)
			assert.True(t, instancesFound)
			assert.ElementsMatch(t, tc.wantInstances, instances)
		})
	}
}

func TestMigInfoProviderGetMigForInstance(t *testing.T) {
	instance := GceInstance{
		Instance:  cloudprovider.Instance{Id: "gce://project/us-test1/base-instance-name-abcd"},
		NumericId: 777,
	}
	instanceRef, err := GceRefFromProviderId(instance.Id)
	assert.Nil(t, err)

	testCases := []struct {
		name                     string
		instanceRef              GceRef
		cache                    *GceCache
		fetchMigInstances        func(GceRef) ([]GceInstance, error)
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
				migs:                map[GceRef]Mig{mig.GceRef(): mig},
				instances:           map[GceRef][]GceInstance{},
				instancesUpdateTime: map[GceRef]time.Time{},
				instancesToMig:      map[GceRef]GceRef{},
				migBaseNameCache:    map[GceRef]string{mig.GceRef(): "base-instance-name"},
			},
			fetchMigInstances:    fetchMigInstancesConst([]GceInstance{instance}),
			expectedMig:          mig,
			expectedCachedMigRef: mig.GceRef(),
			expectedCached:       true,
		},
		{
			name: "mig and basename from cache fill",
			cache: &GceCache{
				migs:                map[GceRef]Mig{mig.GceRef(): mig},
				instances:           map[GceRef][]GceInstance{},
				instancesUpdateTime: map[GceRef]time.Time{},
				instancesToMig:      map[GceRef]GceRef{},
				migBaseNameCache:    map[GceRef]string{},
			},
			fetchMigInstances:    fetchMigInstancesConst([]GceInstance{instance}),
			fetchMigBasename:     fetchMigBasenameConst("base-instance-name"),
			expectedMig:          mig,
			expectedCachedMigRef: mig.GceRef(),
			expectedCached:       true,
		},
		{
			name: "unknown mig from cache fill",
			cache: &GceCache{
				migs:                    map[GceRef]Mig{mig.GceRef(): mig},
				instances:               map[GceRef][]GceInstance{},
				instancesUpdateTime:     map[GceRef]time.Time{},
				instancesFromUnknownMig: map[GceRef]bool{},
				migBaseNameCache:        map[GceRef]string{mig.GceRef(): "base-instance-name"},
			},
			fetchMigInstances:        fetchMigInstancesConst([]GceInstance{}),
			expectedCachedMigUnknown: true,
		},
		{
			name: "no candidate mig during cache fill",
			cache: &GceCache{
				migs:             map[GceRef]Mig{mig.GceRef(): mig},
				migBaseNameCache: map[GceRef]string{mig.GceRef(): "different-base-instance-name"},
				instancesToMig:   map[GceRef]GceRef{},
			},
			fetchMigInstances: fetchMigInstancesConst([]GceInstance{instance}),
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
	oldRefreshTime := time.Now().Add(-time.Hour)
	newRefreshTime := time.Now()
	instances := []GceInstance{
		{Instance: cloudprovider.Instance{Id: "gce://project/us-test1/base-instance-name-abcd"}, NumericId: 7},
		{Instance: cloudprovider.Instance{Id: "gce://project/us-test1/base-instance-name-efgh"}, NumericId: 88},
	}

	testCases := []struct {
		name                    string
		cache                   *GceCache
		fetchMigInstances       func(GceRef) ([]GceInstance, error)
		expectedInstances       []GceInstance
		expectedErr             error
		expectedCachedInstances []GceInstance
		expectedCached          bool
		expectedRefreshTime     time.Time
		expectedRefreshed       bool
	}{
		{
			name: "instances in cache",
			cache: &GceCache{
				migs:                map[GceRef]Mig{mig.GceRef(): mig},
				instances:           map[GceRef][]GceInstance{mig.GceRef(): instances},
				instancesUpdateTime: map[GceRef]time.Time{mig.GceRef(): oldRefreshTime},
			},
			expectedInstances:       instances,
			expectedCachedInstances: instances,
			expectedCached:          true,
			expectedRefreshTime:     oldRefreshTime,
			expectedRefreshed:       true,
		},
		{
			name: "instances cache fill",
			cache: &GceCache{
				migs:                map[GceRef]Mig{mig.GceRef(): mig},
				instances:           map[GceRef][]GceInstance{},
				instancesUpdateTime: map[GceRef]time.Time{},
				instancesToMig:      map[GceRef]GceRef{},
			},
			fetchMigInstances:       fetchMigInstancesConst(instances),
			expectedInstances:       instances,
			expectedCachedInstances: instances,
			expectedCached:          true,
			expectedRefreshTime:     newRefreshTime,
			expectedRefreshed:       true,
		},
		{
			name: "error during instances cache fill",
			cache: &GceCache{
				migs:                map[GceRef]Mig{mig.GceRef(): mig},
				instances:           map[GceRef][]GceInstance{},
				instancesUpdateTime: map[GceRef]time.Time{},
			},
			fetchMigInstances:       fetchMigInstancesFail,
			expectedErr:             errFetchMigInstances,
			expectedCachedInstances: []GceInstance{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := &mockAutoscalingGceClient{
				fetchMigInstances: tc.fetchMigInstances,
			}
			migLister := NewMigLister(tc.cache)
			provider, ok := NewCachingMigInfoProvider(tc.cache, migLister, client, mig.GceRef().Project, 1, 0*time.Second).(*cachingMigInfoProvider)
			assert.True(t, ok)
			provider.timeProvider = &fakeTime{now: newRefreshTime}

			instances, err := provider.GetMigInstances(mig.GceRef())
			cachedInstances, cached := tc.cache.GetMigInstances(mig.GceRef())
			refreshTime, refreshed := tc.cache.GetMigInstancesUpdateTime(mig.GceRef())

			assert.Equal(t, tc.expectedInstances, instances)
			assert.Equal(t, tc.expectedErr, err)

			assert.Equal(t, tc.expectedCachedInstances, cachedInstances)
			assert.Equal(t, tc.expectedCached, cached)

			assert.Equal(t, tc.expectedRefreshTime, refreshTime)
			assert.Equal(t, tc.expectedRefreshed, refreshed)
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

	instances := []GceInstance{
		{Instance: cloudprovider.Instance{Id: "gce://project/us-test1/base-instance-name-abcd"}, NumericId: 1},
		{Instance: cloudprovider.Instance{Id: "gce://project/us-test1/base-instance-name-efgh"}, NumericId: 2},
	}
	otherInstances := []GceInstance{
		{Instance: cloudprovider.Instance{Id: "gce://project/us-test1/other-base-instance-name-abcd"}},
		{Instance: cloudprovider.Instance{Id: "gce://project/us-test1/other-base-instance-name-efgh"}},
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
		fetchMigInstances      func(GceRef) ([]GceInstance, error)
		expectedErr            error
		expectedMigInstances   map[GceRef][]GceInstance
		expectedInstancesToMig map[GceRef]GceRef
	}{
		{
			name: "fill empty cache for one mig",
			cache: &GceCache{
				migs:           map[GceRef]Mig{mig.GceRef(): mig},
				instances:      map[GceRef][]GceInstance{},
				instancesToMig: map[GceRef]GceRef{},
			},
			fetchMigInstances: fetchMigInstancesConst(instances),
			expectedMigInstances: map[GceRef][]GceInstance{
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
				instances:      map[GceRef][]GceInstance{},
				instancesToMig: map[GceRef]GceRef{},
			},
			fetchMigInstances: fetchMigInstancesMapping(map[GceRef][]GceInstance{
				mig.GceRef():      instances,
				otherMig.GceRef(): otherInstances,
			}),
			expectedMigInstances: map[GceRef][]GceInstance{
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
				instances: map[GceRef][]GceInstance{
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
			fetchMigInstances: fetchMigInstancesMapping(map[GceRef][]GceInstance{
				mig.GceRef():      instances,
				otherMig.GceRef(): otherInstances,
			}),
			expectedMigInstances: map[GceRef][]GceInstance{
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
				instances: map[GceRef][]GceInstance{
					mig.GceRef(): instances,
				},
				instancesToMig: map[GceRef]GceRef{
					instancesRefs[0]: mig.GceRef(),
					instancesRefs[1]: mig.GceRef(),
				},
			},
			fetchMigInstances: fetchMigInstancesConst(otherInstances),
			expectedMigInstances: map[GceRef][]GceInstance{
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
				instances:      map[GceRef][]GceInstance{},
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
			name:             "basename from cache fill",
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

func TestGetListManagedInstancesResults(t *testing.T) {
	results := "PAGELESS"
	instanceGroupManager := &gce.InstanceGroupManager{
		Zone:                        mig.GceRef().Zone,
		Name:                        mig.GceRef().Name,
		ListManagedInstancesResults: results,
	}
	testCases := []struct {
		name            string
		cache           *GceCache
		fetchMigs       func(string) ([]*gce.InstanceGroupManager, error)
		fetchResults    func(GceRef) (string, error)
		expectedResults string
		expectedErr     error
	}{
		{
			name: "results in cache",
			cache: &GceCache{
				migs:                             map[GceRef]Mig{mig.GceRef(): mig},
				listManagedInstancesResultsCache: map[GceRef]string{mig.GceRef(): results},
			},
			expectedResults: results,
		},
		{
			name:            "results from cache fill",
			cache:           emptyCache(),
			fetchMigs:       fetchMigsConst([]*gce.InstanceGroupManager{instanceGroupManager}),
			expectedResults: results,
		},
		{
			name:            "cache fill without mig, fallback success",
			cache:           emptyCache(),
			fetchMigs:       fetchMigsConst([]*gce.InstanceGroupManager{}),
			fetchResults:    fetchListManagedInstancesResultsConst(results),
			expectedResults: results,
		},
		{
			name:            "cache fill failure, fallback success",
			cache:           emptyCache(),
			fetchMigs:       fetchMigsFail,
			fetchResults:    fetchListManagedInstancesResultsConst(results),
			expectedResults: results,
		},
		{
			name:         "cache fill without mig, fallback failure",
			cache:        emptyCache(),
			fetchMigs:    fetchMigsConst([]*gce.InstanceGroupManager{}),
			fetchResults: fetchListManagedInstancesResultsFail,
			expectedErr:  errFetchListManagedInstancesResults,
		},
		{
			name:         "cache fill failure, fallback failure",
			cache:        emptyCache(),
			fetchMigs:    fetchMigsFail,
			fetchResults: fetchListManagedInstancesResultsFail,
			expectedErr:  errFetchListManagedInstancesResults,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := &mockAutoscalingGceClient{
				fetchMigs:                        tc.fetchMigs,
				fetchListManagedInstancesResults: tc.fetchResults,
			}
			migLister := NewMigLister(tc.cache)
			provider := NewCachingMigInfoProvider(tc.cache, migLister, client, mig.GceRef().Project, 1, 0*time.Second)

			results, err := provider.GetListManagedInstancesResults(mig.GceRef())
			cachedResults, found := tc.cache.GetListManagedInstancesResults(mig.GceRef())

			assert.Equal(t, tc.expectedErr, err)
			assert.Equal(t, tc.expectedErr == nil, found)
			if tc.expectedErr == nil {
				assert.Equal(t, tc.expectedResults, results)
				assert.Equal(t, tc.expectedResults, cachedResults)
			}
		})
	}
}

func TestGetMigInstanceTemplateName(t *testing.T) {
	templateName := "template-name"
	instanceGroupManager := &gce.InstanceGroupManager{
		Zone:             mig.GceRef().Zone,
		Name:             mig.GceRef().Name,
		InstanceTemplate: "https://www.googleapis.com/compute/v1/projects/test-project/global/instanceTemplates/template-name",
	}

	instanceGroupManagerRegional := &gce.InstanceGroupManager{
		Zone:             mig.GceRef().Zone,
		Name:             mig.GceRef().Name,
		InstanceTemplate: "https://www.googleapis.com/compute/v1/projects/test-project/regions/us-central1/instanceTemplates/template-name",
	}

	testCases := []struct {
		name                 string
		cache                *GceCache
		fetchMigs            func(string) ([]*gce.InstanceGroupManager, error)
		fetchMigTemplateName func(GceRef) (InstanceTemplateName, error)
		expectedTemplateName string
		expectedRegion       bool
		expectedErr          error
	}{
		{
			name: "template name in cache",
			cache: &GceCache{
				migs:                      map[GceRef]Mig{mig.GceRef(): mig},
				instanceTemplateNameCache: map[GceRef]InstanceTemplateName{mig.GceRef(): {templateName, false}},
			},
			expectedTemplateName: templateName,
		},
		{
			name:                 "template name from cache fill",
			cache:                emptyCache(),
			fetchMigs:            fetchMigsConst([]*gce.InstanceGroupManager{instanceGroupManager}),
			expectedTemplateName: templateName,
		},
		{
			name:                 "target size from cache fill, regional",
			cache:                emptyCache(),
			fetchMigs:            fetchMigsConst([]*gce.InstanceGroupManager{instanceGroupManagerRegional}),
			expectedTemplateName: templateName,
		},
		{
			name:                 "cache fill without mig, fallback success",
			cache:                emptyCache(),
			fetchMigs:            fetchMigsConst([]*gce.InstanceGroupManager{}),
			fetchMigTemplateName: fetchMigTemplateNameConst(InstanceTemplateName{templateName, false}),
			expectedTemplateName: templateName,
		},
		{
			name:                 "cache fill failure, fallback success",
			cache:                emptyCache(),
			fetchMigs:            fetchMigsFail,
			fetchMigTemplateName: fetchMigTemplateNameConst(InstanceTemplateName{templateName, false}),
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

			instanceTemplateName, err := provider.GetMigInstanceTemplateName(mig.GceRef())
			cachedInstanceTemplateName, found := tc.cache.GetMigInstanceTemplateName(mig.GceRef())

			assert.Equal(t, tc.expectedErr, err)
			assert.Equal(t, tc.expectedErr == nil, found)
			if tc.expectedErr == nil {
				assert.Equal(t, tc.expectedTemplateName, instanceTemplateName.Name)
				assert.Equal(t, tc.expectedTemplateName, cachedInstanceTemplateName.Name)
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
		fetchMigTemplateName   func(GceRef) (InstanceTemplateName, error)
		fetchMigTemplate       func(GceRef, string, bool) (*gce.InstanceTemplate, error)
		expectedTemplate       *gce.InstanceTemplate
		expectedCachedTemplate *gce.InstanceTemplate
		expectedErr            error
	}{
		{
			name: "template in cache",
			cache: &GceCache{
				migs:                      map[GceRef]Mig{mig.GceRef(): mig},
				instanceTemplateNameCache: map[GceRef]InstanceTemplateName{mig.GceRef(): {templateName, false}},
				instanceTemplatesCache:    map[GceRef]*gce.InstanceTemplate{mig.GceRef(): template},
			},
			expectedTemplate:       template,
			expectedCachedTemplate: template,
		},
		{
			name: "cache without template, fetch success",
			cache: &GceCache{
				migs:                      map[GceRef]Mig{mig.GceRef(): mig},
				instanceTemplateNameCache: map[GceRef]InstanceTemplateName{mig.GceRef(): {templateName, false}},
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
				instanceTemplateNameCache: map[GceRef]InstanceTemplateName{mig.GceRef(): {templateName, false}},
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
				instanceTemplateNameCache: map[GceRef]InstanceTemplateName{mig.GceRef(): {templateName, false}},
				instanceTemplatesCache:    make(map[GceRef]*gce.InstanceTemplate),
			},
			fetchMigTemplate: fetchMigTemplateFail,
			expectedErr:      errFetchMigTemplate,
		},
		{
			name: "cache with old template, fetch failure",
			cache: &GceCache{
				migs:                      map[GceRef]Mig{mig.GceRef(): mig},
				instanceTemplateNameCache: map[GceRef]InstanceTemplateName{mig.GceRef(): {templateName, false}},
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

func TestGetMigInstanceKubeEnv(t *testing.T) {
	templateName := "template-name"
	kubeEnvValue := "VAR1: VALUE1\nVAR2: VALUE2"
	kubeEnv, err := ParseKubeEnv(templateName, kubeEnvValue)
	assert.NoError(t, err)
	template := &gce.InstanceTemplate{
		Name:        templateName,
		Description: "instance template",
		Properties: &gce.InstanceProperties{
			Metadata: &gce.Metadata{
				Items: []*gce.MetadataItems{
					{Key: "kube-env", Value: &kubeEnvValue},
				},
			},
		},
	}

	oldTemplateName := "old-template-name"
	oldKubeEnvValue := "VAR3: VALUE3\nVAR4: VALUE4"
	oldKubeEnv, err := ParseKubeEnv(oldTemplateName, oldKubeEnvValue)
	assert.NoError(t, err)
	oldTemplate := &gce.InstanceTemplate{
		Name:        oldTemplateName,
		Description: "old instance template",
		Properties: &gce.InstanceProperties{
			Metadata: &gce.Metadata{
				Items: []*gce.MetadataItems{
					{Key: "kube-env", Value: &oldKubeEnvValue},
				},
			},
		},
	}

	testCases := []struct {
		name                  string
		cache                 *GceCache
		fetchMigs             func(string) ([]*gce.InstanceGroupManager, error)
		fetchMigTemplateName  func(GceRef) (InstanceTemplateName, error)
		fetchMigTemplate      func(GceRef, string, bool) (*gce.InstanceTemplate, error)
		expectedKubeEnv       KubeEnv
		expectedCachedKubeEnv KubeEnv
		expectedErr           error
	}{
		{
			name: "kube-env in cache",
			cache: &GceCache{
				migs:                      map[GceRef]Mig{mig.GceRef(): mig},
				instanceTemplateNameCache: map[GceRef]InstanceTemplateName{mig.GceRef(): {templateName, false}},
				kubeEnvCache:              map[GceRef]KubeEnv{mig.GceRef(): kubeEnv},
			},
			expectedKubeEnv:       kubeEnv,
			expectedCachedKubeEnv: kubeEnv,
		},
		{
			name: "cache without kube-env, template in cache",
			cache: &GceCache{
				migs:                      map[GceRef]Mig{mig.GceRef(): mig},
				instanceTemplateNameCache: map[GceRef]InstanceTemplateName{mig.GceRef(): {templateName, false}},
				instanceTemplatesCache:    map[GceRef]*gce.InstanceTemplate{mig.GceRef(): template},
				kubeEnvCache:              make(map[GceRef]KubeEnv),
			},
			expectedKubeEnv:       kubeEnv,
			expectedCachedKubeEnv: kubeEnv,
		},
		{
			name: "cache without kube-env, fetch success",
			cache: &GceCache{
				migs:                      map[GceRef]Mig{mig.GceRef(): mig},
				instanceTemplateNameCache: map[GceRef]InstanceTemplateName{mig.GceRef(): {templateName, false}},
				instanceTemplatesCache:    make(map[GceRef]*gce.InstanceTemplate),
				kubeEnvCache:              make(map[GceRef]KubeEnv),
			},
			fetchMigTemplate:      fetchMigTemplateConst(template),
			expectedKubeEnv:       kubeEnv,
			expectedCachedKubeEnv: kubeEnv,
		},
		{
			name: "cache with old kube-env, new template cached",
			cache: &GceCache{
				migs:                      map[GceRef]Mig{mig.GceRef(): mig},
				instanceTemplateNameCache: map[GceRef]InstanceTemplateName{mig.GceRef(): {templateName, false}},
				instanceTemplatesCache:    map[GceRef]*gce.InstanceTemplate{mig.GceRef(): template},
				kubeEnvCache:              map[GceRef]KubeEnv{mig.GceRef(): oldKubeEnv},
			},
			expectedKubeEnv:       kubeEnv,
			expectedCachedKubeEnv: kubeEnv,
		},
		{
			name: "cache with old kube-env, fetch success",
			cache: &GceCache{
				migs:                      map[GceRef]Mig{mig.GceRef(): mig},
				instanceTemplateNameCache: map[GceRef]InstanceTemplateName{mig.GceRef(): {templateName, false}},
				instanceTemplatesCache:    map[GceRef]*gce.InstanceTemplate{mig.GceRef(): oldTemplate},
				kubeEnvCache:              map[GceRef]KubeEnv{mig.GceRef(): oldKubeEnv},
			},
			fetchMigTemplate:      fetchMigTemplateConst(template),
			expectedKubeEnv:       kubeEnv,
			expectedCachedKubeEnv: kubeEnv,
		},
		{
			name: "cache without kube-env, fetch failure",
			cache: &GceCache{
				migs:                      map[GceRef]Mig{mig.GceRef(): mig},
				instanceTemplateNameCache: map[GceRef]InstanceTemplateName{mig.GceRef(): {templateName, false}},
				instanceTemplatesCache:    make(map[GceRef]*gce.InstanceTemplate),
				kubeEnvCache:              make(map[GceRef]KubeEnv),
			},
			fetchMigTemplate: fetchMigTemplateFail,
			expectedErr:      errFetchMigTemplate,
		},
		{
			name: "cache with old kube-env, fetch failure",
			cache: &GceCache{
				migs:                      map[GceRef]Mig{mig.GceRef(): mig},
				instanceTemplateNameCache: map[GceRef]InstanceTemplateName{mig.GceRef(): {templateName, false}},
				instanceTemplatesCache:    map[GceRef]*gce.InstanceTemplate{mig.GceRef(): oldTemplate},
				kubeEnvCache:              map[GceRef]KubeEnv{mig.GceRef(): oldKubeEnv},
			},
			fetchMigTemplate:      fetchMigTemplateFail,
			expectedCachedKubeEnv: oldKubeEnv,
			expectedErr:           errFetchMigTemplate,
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

			kubeEnv, err := provider.GetMigKubeEnv(mig.GceRef())
			cachedKubeEnv, found := tc.cache.GetMigKubeEnv(mig.GceRef())

			assert.Equal(t, tc.expectedErr, err)
			if tc.expectedErr == nil {
				assert.Equal(t, tc.expectedKubeEnv, kubeEnv)
			}

			assert.Equal(t, tc.expectedCachedKubeEnv.env != nil, found)
			if tc.expectedCachedKubeEnv.env != nil {
				assert.Equal(t, tc.expectedCachedKubeEnv, cachedKubeEnv)
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
				instanceTemplateNameCache: map[GceRef]InstanceTemplateName{mig.GceRef(): {"template", false}},
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
	instance := GceInstance{
		Instance: cloudprovider.Instance{Id: "gce://project/zone/base-instance-name-abcd"}, NumericId: 1111,
	}
	instanceRef, err := GceRefFromProviderId(instance.Id)
	assert.Nil(t, err)
	instance2 := GceInstance{
		Instance: cloudprovider.Instance{Id: "gce://project/zone/base-instance-name-abcd2"}, NumericId: 222,
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
		migs:                             map[GceRef]Mig{mig.GceRef(): mig},
		instances:                        make(map[GceRef][]GceInstance),
		instancesUpdateTime:              make(map[GceRef]time.Time),
		migTargetSizeCache:               make(map[GceRef]int64),
		migBaseNameCache:                 make(map[GceRef]string),
		listManagedInstancesResultsCache: make(map[GceRef]string),
		instanceTemplateNameCache:        make(map[GceRef]InstanceTemplateName),
		instanceTemplatesCache:           make(map[GceRef]*gce.InstanceTemplate),
		instancesFromUnknownMig:          make(map[GceRef]bool),
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

func fetchMigInstancesFail(_ GceRef) ([]GceInstance, error) {
	return nil, errFetchMigInstances
}

func fetchMigInstancesConst(instances []GceInstance) func(GceRef) ([]GceInstance, error) {
	return func(GceRef) ([]GceInstance, error) {
		return instances, nil
	}
}

func fetchMigInstancesWithCounter(instances []GceInstance, migCounter map[GceRef]int) func(GceRef) ([]GceInstance, error) {
	return func(ref GceRef) ([]GceInstance, error) {
		migCounter[ref] = migCounter[ref] + 1
		return instances, nil
	}
}

func fetchMigInstancesMapping(instancesMapping map[GceRef][]GceInstance) func(GceRef) ([]GceInstance, error) {
	return func(migRef GceRef) ([]GceInstance, error) {
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

func fetchMigTemplateNameFail(_ GceRef) (InstanceTemplateName, error) {
	return InstanceTemplateName{}, errFetchMigTemplateName
}

func fetchMigTemplateNameConst(instanceTemplateName InstanceTemplateName) func(GceRef) (InstanceTemplateName, error) {
	return func(GceRef) (InstanceTemplateName, error) {
		return instanceTemplateName, nil
	}
}

func fetchMigTemplateFail(_ GceRef, _ string, _ bool) (*gce.InstanceTemplate, error) {
	return nil, errFetchMigTemplate
}

func fetchMigTemplateConst(template *gce.InstanceTemplate) func(GceRef, string, bool) (*gce.InstanceTemplate, error) {
	return func(GceRef, string, bool) (*gce.InstanceTemplate, error) {
		return template, nil
	}
}

func fetchListManagedInstancesResultsFail(_ GceRef) (string, error) {
	return "", errFetchListManagedInstancesResults
}

func fetchListManagedInstancesResultsConst(listManagedInstancesResults string) func(GceRef) (string, error) {
	return func(GceRef) (string, error) {
		return listManagedInstancesResults, nil
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
