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
	errFetchMigs         = errors.New("fetch migs error")
	errFetchMig          = errors.New("fetch single mig error")
	errFetchMigInstances = errors.New("fetch mig instances error")
	errFetchMigBaseName  = errors.New("fetch mig basename error")
	errFetchMigTemplate  = errors.New("fetch mig template error")
	errFetchMachineType  = errors.New("fetch machine type error")

	mig = &gceMig{
		gceRef: GceRef{
			Project: "project",
			Zone:    "us-test1",
			Name:    "mig",
		},
	}
	mig1 = &gceMig{
		gceRef: GceRef{
			Project: "myprojid",
			Zone:    "myzone1",
			Name:    "mig1",
		},
	}
	mig2 = &gceMig{
		gceRef: GceRef{
			Project: "myprojid",
			Zone:    "myzone2",
			Name:    "mig2",
		},
	}

	instance1 = GceInstance{
		Instance: cloudprovider.Instance{
			Id:     "gce://myprojid/myzone1/test-instance-1",
			Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning},
		},
		Igm: mig1.GceRef(),
	}
	instance2 = GceInstance{
		Instance: cloudprovider.Instance{
			Id:     "gce://myprojid/myzone1/test-instance-2",
			Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceCreating},
		},
		Igm: mig1.GceRef(),
	}
	instance3 = GceInstance{
		Instance: cloudprovider.Instance{
			Id:     "gce://myprojid/myzone2/test-instance-3",
			Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning},
		},
		Igm: mig2.GceRef(),
	}

	instance4 = GceInstance{
		Instance: cloudprovider.Instance{
			Id:     "gce://myprojid/myzone2/test-instance-4",
			Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning},
		},
		Igm: GceRef{},
	}

	instance5 = GceInstance{
		Instance: cloudprovider.Instance{
			Id:     "gce://myprojid/myzone2/test-instance-5",
			Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning},
		},
		Igm: GceRef{},
	}

	instance6 = GceInstance{
		Instance: cloudprovider.Instance{
			Id:     "gce://myprojid/myzone2/test-instance-6",
			Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning},
		},
		Igm: mig2.GceRef(),
	}
)

type mockAutoscalingGceClient struct {
	fetchMigs                        func(string) ([]*gce.InstanceGroupManager, error)
	fetchAllInstances                func(project, zone string, filter string) ([]GceInstance, error)
	fetchMig                         func(GceRef) (*gce.InstanceGroupManager, error)
	fetchMigTargetSize               func(GceRef) (int64, error)
	fetchMigBasename                 func(GceRef) (string, error)
	fetchMigTemplateName             func(GceRef) (InstanceTemplateName, error)
	fetchMigInstances                func(GceRef) ([]GceInstance, error)
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

func (client *mockAutoscalingGceClient) FetchAllInstances(project, zone string, filter string) ([]GceInstance, error) {
	return client.fetchAllInstances(project, zone, filter)
}

func (client *mockAutoscalingGceClient) FetchMig(migRef GceRef) (*gce.InstanceGroupManager, error) {
	return client.fetchMig(migRef)
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

func (client *mockAutoscalingGceClient) FetchAvailableDiskTypes(_ string) ([]string, error) {
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
			name:            "No instances in cache",
			cache:           NewGceCache(),
			wantClientCalls: 1,
			wantInstances:   newInstances,
			wantUpdateTime:  timeNow,
		},
		{
			name: "Old instances in cache",
			cache: NewGceCache().
				WithInstances(map[GceRef][]GceInstance{migRef: oldInstances}).
				WithInstancesUpdateTime(map[GceRef]time.Time{migRef: timeOld}),
			wantClientCalls: 1,
			wantInstances:   newInstances,
			wantUpdateTime:  timeNow,
		},
		{
			name: "Recently updated instances in cache",
			cache: NewGceCache().
				WithInstances(map[GceRef][]GceInstance{migRef: oldInstances}).
				WithInstancesUpdateTime(map[GceRef]time.Time{migRef: timeRecent}),
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

			provider, ok := NewCachingMigInfoProvider(tc.cache, NewMigLister(tc.cache), client, mig.GceRef().Project, 1, time.Hour, false, false).(*cachingMigInfoProvider)
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
		fetchMig                 func(GceRef) (*gce.InstanceGroupManager, error)
		expectedMig              Mig
		expectedErr              error
		expectedCachedMigRef     GceRef
		expectedCached           bool
		expectedCachedMigUnknown bool
	}{
		{
			name: "instance mig ref and mig in cache",
			cache: NewGceCache().
				WithMigs(map[GceRef]Mig{mig.GceRef(): mig}).
				WithInstancesToMig(map[GceRef]GceRef{instanceRef: mig.GceRef()}),
			expectedMig:          mig,
			expectedCachedMigRef: mig.GceRef(),
			expectedCached:       true,
		},
		{
			name: "only instance mig ref in cache",
			cache: NewGceCache().
				WithInstancesToMig(map[GceRef]GceRef{instanceRef: mig.GceRef()}),
			expectedMig:          nil,
			expectedErr:          fmt.Errorf("instance %v belongs to unregistered mig %v", instanceRef, mig.GceRef()),
			expectedCachedMigRef: mig.GceRef(),
			expectedCached:       true,
		},
		{
			name: "instance mig unknown in cache",
			cache: NewGceCache().
				WithInstancesFromUnknownMig(map[GceRef]bool{instanceRef: true}),
			expectedCachedMigUnknown: true,
		},
		{
			name: "mig from cache fill",
			cache: NewGceCache().
				WithMigs(map[GceRef]Mig{mig.GceRef(): mig}).
				WithMigBaseName(map[GceRef]string{mig.GceRef(): "base-instance-name"}),
			fetchMigInstances:    fetchMigInstancesConst([]GceInstance{instance}),
			expectedMig:          mig,
			expectedCachedMigRef: mig.GceRef(),
			expectedCached:       true,
		},
		{
			name: "mig and basename from cache fill",
			cache: NewGceCache().
				WithMigs(map[GceRef]Mig{mig.GceRef(): mig}),
			fetchMigInstances: fetchMigInstancesConst([]GceInstance{instance}),
			fetchMig: fetchMigConst(&gce.InstanceGroupManager{
				BaseInstanceName: "base-instance-name",
				Name:             mig.GceRef().Name,
			}),
			expectedMig:          mig,
			expectedCachedMigRef: mig.GceRef(),
			expectedCached:       true,
		},
		{
			name: "unknown mig from cache fill",
			cache: NewGceCache().
				WithMigs(map[GceRef]Mig{mig.GceRef(): mig}).
				WithMigBaseName(map[GceRef]string{mig.GceRef(): "base-instance-name"}),
			fetchMigInstances:        fetchMigInstancesConst([]GceInstance{}),
			expectedCachedMigUnknown: true,
		},
		{
			name: "no candidate mig during cache fill",
			cache: NewGceCache().
				WithMigs(map[GceRef]Mig{mig.GceRef(): mig}).
				WithMigBaseName(map[GceRef]string{mig.GceRef(): "different-base-instance-name"}),
			fetchMigInstances: fetchMigInstancesConst([]GceInstance{instance}),
		},
		{
			name: "fetch instances error during cache fill",
			cache: NewGceCache().
				WithMigs(map[GceRef]Mig{mig.GceRef(): mig}).
				WithMigBaseName(map[GceRef]string{mig.GceRef(): "base-instance-name"}),
			fetchMigInstances: fetchMigInstancesFail,
			expectedErr:       errFetchMigInstances,
		},
		{
			name: "fetch basename error during cache fill",
			cache: NewGceCache().
				WithMigs(map[GceRef]Mig{mig.GceRef(): mig}),
			fetchMig:    fetchMigFail,
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := &mockAutoscalingGceClient{
				fetchMigInstances: tc.fetchMigInstances,
				fetchMig:          tc.fetchMig,
				fetchMigs:         fetchMigsConst(nil),
			}
			migLister := NewMigLister(tc.cache)
			provider := NewCachingMigInfoProvider(tc.cache, migLister, client, mig.GceRef().Project, 1, 0*time.Second, false, false)

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
			cache: NewGceCache().
				WithMigs(map[GceRef]Mig{mig.GceRef(): mig}).
				WithInstances(map[GceRef][]GceInstance{mig.GceRef(): instances}).
				WithInstancesUpdateTime(map[GceRef]time.Time{mig.GceRef(): oldRefreshTime}),
			expectedInstances:       instances,
			expectedCachedInstances: instances,
			expectedCached:          true,
			expectedRefreshTime:     oldRefreshTime,
			expectedRefreshed:       true,
		},
		{
			name: "instances cache fill",
			cache: NewGceCache().
				WithMigs(map[GceRef]Mig{mig.GceRef(): mig}),
			fetchMigInstances:       fetchMigInstancesConst(instances),
			expectedInstances:       instances,
			expectedCachedInstances: instances,
			expectedCached:          true,
			expectedRefreshTime:     newRefreshTime,
			expectedRefreshed:       true,
		},
		{
			name: "error during instances cache fill",
			cache: NewGceCache().
				WithMigs(map[GceRef]Mig{mig.GceRef(): mig}),
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
			provider, ok := NewCachingMigInfoProvider(tc.cache, migLister, client, mig.GceRef().Project, 1, 0*time.Second, false, false).(*cachingMigInfoProvider)
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
	mig1Instances := []GceInstance{instance1, instance2}
	mig2Instances := []GceInstance{instance3, instance6}
	otherInstances := []GceInstance{
		{Instance: cloudprovider.Instance{Id: "gce://project/us-test1/other-base-instance-name-abcd"}},
		{Instance: cloudprovider.Instance{Id: "gce://project/us-test1/other-base-instance-name-efgh"}},
	}

	mig1Igm := &gce.InstanceGroupManager{
		Zone:       mig1.GceRef().Zone,
		Name:       mig1.GceRef().Name,
		TargetSize: 2,
		CurrentActions: &gce.InstanceGroupManagerActionsSummary{
			Creating: 1,
		},
	}
	mig2Igm := &gce.InstanceGroupManager{
		Zone:           mig2.GceRef().Zone,
		Name:           mig2.GceRef().Name,
		TargetSize:     2,
		CurrentActions: &gce.InstanceGroupManagerActionsSummary{},
	}

	instancesRefs := toInstancesRefs(t, instances)
	mig1InstancesRefs := toInstancesRefs(t, mig1Instances)
	mig2InstancesRefs := toInstancesRefs(t, mig2Instances)
	otherInstancesRefs := toInstancesRefs(t, otherInstances)

	testCases := []struct {
		name                              string
		cache                             *GceCache
		fetchMigInstances                 func(GceRef) ([]GceInstance, error)
		fetchMigs                         func(string) ([]*gce.InstanceGroupManager, error)
		fetchAllInstances                 func(string, string, string) ([]GceInstance, error)
		bulkGceMigInstancesListingEnabled bool
		projectId                         string
		expectedErr                       error
		expectedMigInstances              map[GceRef][]GceInstance
		expectedInstancesToMig            map[GceRef]GceRef
	}{
		{
			name: "fill empty cache for one mig",
			cache: NewGceCache().
				WithMigs(map[GceRef]Mig{mig.GceRef(): mig}),
			fetchMigInstances: fetchMigInstancesConst(instances),
			projectId:         mig.GceRef().Project,
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
			cache: NewGceCache().
				WithMigs(map[GceRef]Mig{
					mig.GceRef():      mig,
					otherMig.GceRef(): otherMig,
				}),
			fetchMigInstances: fetchMigInstancesMapping(map[GceRef][]GceInstance{
				mig.GceRef():      instances,
				otherMig.GceRef(): otherInstances,
			}),
			projectId: mig.GceRef().Project,
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
			cache: NewGceCache().
				WithMigs(map[GceRef]Mig{
					mig.GceRef(): mig,
				}).
				WithInstances(map[GceRef][]GceInstance{
					mig.GceRef():      instances,
					otherMig.GceRef(): otherInstances,
				}).
				WithInstancesToMig(map[GceRef]GceRef{
					instancesRefs[0]:      mig.GceRef(),
					instancesRefs[1]:      mig.GceRef(),
					otherInstancesRefs[0]: otherMig.GceRef(),
					otherInstancesRefs[1]: otherMig.GceRef(),
				}),
			fetchMigInstances: fetchMigInstancesMapping(map[GceRef][]GceInstance{
				mig.GceRef():      instances,
				otherMig.GceRef(): otherInstances,
			}),
			projectId: mig.GceRef().Project,
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
			cache: NewGceCache().
				WithMigs(map[GceRef]Mig{
					mig.GceRef(): mig,
				}).
				WithInstances(map[GceRef][]GceInstance{
					mig.GceRef(): instances,
				}).
				WithInstancesToMig(map[GceRef]GceRef{
					instancesRefs[0]: mig.GceRef(),
					instancesRefs[1]: mig.GceRef(),
				}),
			fetchMigInstances: fetchMigInstancesConst(otherInstances),
			projectId:         mig.GceRef().Project,
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
			cache: NewGceCache().
				WithMigs(map[GceRef]Mig{
					mig.GceRef(): mig,
				}),
			fetchMigInstances: fetchMigInstancesFail,
			projectId:         mig.GceRef().Project,
			expectedErr:       errFetchMigInstances,
		},
		{
			name: "bulkGceMigInstancesListingEnabled - fill empty cache for one mig - instances in creating/deleting state",
			cache: NewGceCache().
				WithMigs(map[GceRef]Mig{mig1.GceRef(): mig1}),
			fetchMigInstances:                 fetchMigInstancesConst(mig1Instances),
			fetchMigs:                         fetchMigsConst([]*gce.InstanceGroupManager{mig1Igm}),
			fetchAllInstances:                 fetchAllInstancesInZone(map[string][]GceInstance{"myzone1": {instance1, instance2}}),
			bulkGceMigInstancesListingEnabled: true,
			projectId:                         mig1.GceRef().Project,
			expectedMigInstances: map[GceRef][]GceInstance{
				mig1.GceRef(): mig1Instances,
			},
			expectedInstancesToMig: map[GceRef]GceRef{
				mig1InstancesRefs[0]: mig1.GceRef(),
				mig1InstancesRefs[1]: mig1.GceRef(),
			},
		},
		{
			name: "bulkGceMigInstancesListingEnabled - fill empty cache for one mig - number of instances are inconsistent in bulk listing result",
			cache: NewGceCache().
				WithMigs(map[GceRef]Mig{mig2.GceRef(): mig2}),
			fetchMigInstances: fetchMigInstancesConst(mig2Instances),
			fetchMigs:         fetchMigsConst([]*gce.InstanceGroupManager{mig2Igm}),
			// one instance is missing from the instances of igm2 in myzone2
			fetchAllInstances:                 fetchAllInstancesInZone(map[string][]GceInstance{"myzone2": {instance3}}),
			bulkGceMigInstancesListingEnabled: true,
			projectId:                         mig2.GceRef().Project,
			expectedMigInstances: map[GceRef][]GceInstance{
				mig2.GceRef(): mig2Instances,
			},
			expectedInstancesToMig: map[GceRef]GceRef{
				mig2InstancesRefs[0]: mig2.GceRef(),
				mig2InstancesRefs[1]: mig2.GceRef(),
			},
		},
		{
			name: "bulkGceMigInstancesListingEnabled - fill empty cache for one mig - all instances in running state",
			cache: NewGceCache().
				WithMigs(map[GceRef]Mig{mig2.GceRef(): mig2}),
			fetchMigs:                         fetchMigsConst([]*gce.InstanceGroupManager{mig2Igm}),
			fetchAllInstances:                 fetchAllInstancesInZone(map[string][]GceInstance{"myzone2": {instance3, instance6}}),
			bulkGceMigInstancesListingEnabled: true,
			projectId:                         mig2.GceRef().Project,
			expectedMigInstances: map[GceRef][]GceInstance{
				mig2.GceRef(): mig2Instances,
			},
			expectedInstancesToMig: map[GceRef]GceRef{
				mig2InstancesRefs[0]: mig2.GceRef(),
				mig2InstancesRefs[1]: mig2.GceRef(),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := &mockAutoscalingGceClient{
				fetchMigInstances: tc.fetchMigInstances,
				fetchMigs:         tc.fetchMigs,
				fetchAllInstances: tc.fetchAllInstances,
			}
			migLister := NewMigLister(tc.cache)
			provider := NewCachingMigInfoProvider(tc.cache, migLister, client, tc.projectId, 1, 0*time.Second, tc.bulkGceMigInstancesListingEnabled, false)
			err := provider.RegenerateMigInstancesCache()

			assert.Equal(t, tc.expectedErr, err)
			if tc.expectedErr == nil {
				assert.Equal(t, tc.expectedMigInstances, tc.cache.instances)
				assert.Equal(t, tc.expectedInstancesToMig, tc.cache.instancesToMig)
			}
		})
	}
}

func toInstancesRefs(t *testing.T, instances []GceInstance) []GceRef {
	var refs []GceRef
	for _, instance := range instances {
		instanceRef, err := GceRefFromProviderId(instance.Id)
		assert.Nil(t, err)
		refs = append(refs, instanceRef)
	}
	return refs
}

func TestGetMigTargetSize(t *testing.T) {
	targetSize := int64(42)
	instanceGroupManager := &gce.InstanceGroupManager{
		Zone:       mig.GceRef().Zone,
		Name:       mig.GceRef().Name,
		TargetSize: targetSize,
		SelfLink:   fmt.Sprintf("projects/%s/zones/%s/instanceGroups/%s", mig.GceRef().Project, mig.GceRef().Zone, mig.GceRef().Name),
	}
	instanceGroupManager1 := &gce.InstanceGroupManager{
		Zone:       mig1.GceRef().Zone,
		Name:       mig1.GceRef().Name,
		TargetSize: targetSize,
		SelfLink:   fmt.Sprintf("projects/%s/zones/%s/instanceGroups/%s", mig1.GceRef().Project, mig1.GceRef().Zone, mig1.GceRef().Name),
	}
	targetSuspendedSize := int64(3)
	instanceGroupManagerWithSuspendedNodes := &gce.InstanceGroupManager{
		Zone:                mig.GceRef().Zone,
		Name:                mig.GceRef().Name,
		TargetSize:          targetSize,
		TargetSuspendedSize: targetSuspendedSize,
		SelfLink:            fmt.Sprintf("projects/%s/zones/%s/instanceGroups/%s", mig.GceRef().Project, mig.GceRef().Zone, mig.GceRef().Name),
	}

	testCases := []struct {
		name               string
		cache              *GceCache
		migQuery           *gceMig
		fetchMigs          func(string) ([]*gce.InstanceGroupManager, error)
		fetchMig           func(GceRef) (*gce.InstanceGroupManager, error)
		expectedTargetSize int64
		expectedErr        error
	}{
		{
			name: "target size in cache",
			cache: NewGceCache().
				WithMigs(map[GceRef]Mig{mig.GceRef(): mig}).
				WithMigTargetSize(map[GceRef]int64{mig.GceRef(): targetSize}),
			expectedTargetSize: targetSize,
			migQuery:           mig,
		},
		{
			name:               "target size from cache fill",
			cache:              emptyCache(),
			fetchMigs:          fetchMigsConst([]*gce.InstanceGroupManager{instanceGroupManager}),
			expectedTargetSize: targetSize,
			migQuery:           mig,
		},
		{
			name:               "target size from cache fill with suspended nodes",
			cache:              emptyCache(),
			fetchMigs:          fetchMigsConst([]*gce.InstanceGroupManager{instanceGroupManagerWithSuspendedNodes}),
			expectedTargetSize: targetSize + targetSuspendedSize,
			migQuery:           mig,
		},
		{
			name:               "target size from cache fill, multiple MIG projects",
			cache:              emptyCache(),
			fetchMigs:          fetchMigsConst([]*gce.InstanceGroupManager{instanceGroupManager, instanceGroupManager1}),
			expectedTargetSize: targetSize,
			migQuery:           mig1,
		},
		{
			name:               "cache fill without mig, fallback success",
			cache:              emptyCache(),
			fetchMigs:          fetchMigsConst([]*gce.InstanceGroupManager{}),
			fetchMig:           fetchMigConst(&gce.InstanceGroupManager{TargetSize: targetSize}),
			expectedTargetSize: targetSize,
			migQuery:           mig,
		},
		{
			name:               "cache fill failure, fallback success",
			cache:              emptyCache(),
			fetchMigs:          fetchMigsFail,
			fetchMig:           fetchMigConst(&gce.InstanceGroupManager{TargetSize: targetSize}),
			expectedTargetSize: targetSize,
			migQuery:           mig,
		},
		{
			name:        "cache fill without mig, fallback failure",
			cache:       emptyCache(),
			fetchMigs:   fetchMigsConst([]*gce.InstanceGroupManager{}),
			fetchMig:    fetchMigFail,
			expectedErr: errFetchMig,
			migQuery:    mig,
		},
		{
			name:        "cache fill failure, fallback failure",
			cache:       emptyCache(),
			fetchMigs:   fetchMigsFail,
			fetchMig:    fetchMigFail,
			expectedErr: errFetchMig,
			migQuery:    mig,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := &mockAutoscalingGceClient{
				fetchMigs: tc.fetchMigs,
				fetchMig:  tc.fetchMig,
			}
			migLister := NewMigLister(tc.cache)
			provider := NewCachingMigInfoProvider(tc.cache, migLister, client, mig.GceRef().Project, 1, 0*time.Second, false, true)

			targetSize, err := provider.GetMigTargetSize(tc.migQuery.GceRef())
			cachedTargetSize, found := tc.cache.GetMigTargetSize(tc.migQuery.GceRef())

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
		fetchMig         func(GceRef) (*gce.InstanceGroupManager, error)
		expectedBasename string
		expectedErr      error
	}{
		{
			name: "basename in cache",
			cache: NewGceCache().
				WithMigs(map[GceRef]Mig{mig.GceRef(): mig}).
				WithMigBaseName(map[GceRef]string{mig.GceRef(): basename}),
			expectedBasename: basename,
		},
		{
			name:             "basename from cache fill",
			cache:            emptyCache(),
			fetchMigs:        fetchMigsConst([]*gce.InstanceGroupManager{instanceGroupManager}),
			expectedBasename: basename,
		},
		{
			name:      "cache fill without mig, fallback success",
			cache:     emptyCache(),
			fetchMigs: fetchMigsConst([]*gce.InstanceGroupManager{}),
			fetchMig: fetchMigConst(&gce.InstanceGroupManager{
				BaseInstanceName: basename}),
			expectedBasename: basename,
		},
		{
			name:      "cache fill failure, fallback success",
			cache:     emptyCache(),
			fetchMigs: fetchMigsFail,
			fetchMig: fetchMigConst(&gce.InstanceGroupManager{
				BaseInstanceName: basename}),
			expectedBasename: basename,
		},
		{
			name:        "cache fill without mig, fallback failure",
			cache:       emptyCache(),
			fetchMigs:   fetchMigsConst([]*gce.InstanceGroupManager{}),
			fetchMig:    fetchMigFail,
			expectedErr: errFetchMig,
		},
		{
			name:        "cache fill failure, fallback failure",
			cache:       emptyCache(),
			fetchMigs:   fetchMigsFail,
			fetchMig:    fetchMigFail,
			expectedErr: errFetchMig,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := &mockAutoscalingGceClient{
				fetchMigs: tc.fetchMigs,
				fetchMig:  tc.fetchMig,
			}
			migLister := NewMigLister(tc.cache)
			provider := NewCachingMigInfoProvider(tc.cache, migLister, client, mig.GceRef().Project, 1, 0*time.Second, false, false)

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
		fetchMig        func(GceRef) (*gce.InstanceGroupManager, error)
		expectedResults string
		expectedErr     error
	}{
		{
			name: "results in cache",
			cache: NewGceCache().
				WithMigs(map[GceRef]Mig{mig.GceRef(): mig}).
				WithListManagedInstancesResults(map[GceRef]string{mig.GceRef(): results}),
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
			fetchMig:        fetchMigConst(instanceGroupManager),
			expectedResults: results,
		},
		{
			name:            "cache fill failure, fallback success",
			cache:           emptyCache(),
			fetchMigs:       fetchMigsFail,
			fetchMig:        fetchMigConst(instanceGroupManager),
			expectedResults: results,
		},
		{
			name:        "cache fill without mig, fallback failure",
			cache:       emptyCache(),
			fetchMigs:   fetchMigsConst([]*gce.InstanceGroupManager{}),
			fetchMig:    fetchMigFail,
			expectedErr: errFetchMig,
		},
		{
			name:        "cache fill failure, fallback failure",
			cache:       emptyCache(),
			fetchMigs:   fetchMigsFail,
			fetchMig:    fetchMigFail,
			expectedErr: errFetchMig,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := &mockAutoscalingGceClient{
				fetchMigs: tc.fetchMigs,
				fetchMig:  tc.fetchMig,
			}
			migLister := NewMigLister(tc.cache)
			provider := NewCachingMigInfoProvider(tc.cache, migLister, client, mig.GceRef().Project, 1, 0*time.Second, false, false)

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
		fetchMig             func(GceRef) (*gce.InstanceGroupManager, error)
		expectedTemplateName string
		expectedRegion       bool
		expectedErr          error
	}{
		{
			name: "template name in cache",
			cache: NewGceCache().
				WithMigs(map[GceRef]Mig{mig.GceRef(): mig}).
				WithInstanceTemplateNames(map[GceRef]InstanceTemplateName{mig.GceRef(): {templateName, false}}),
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
			fetchMig:             fetchMigConst(instanceGroupManager),
			expectedTemplateName: templateName,
		},
		{
			name:                 "cache fill failure, fallback success",
			cache:                emptyCache(),
			fetchMigs:            fetchMigsFail,
			fetchMig:             fetchMigConst(instanceGroupManager),
			expectedTemplateName: templateName,
		},
		{
			name:        "cache fill without mig, fallback failure",
			cache:       emptyCache(),
			fetchMigs:   fetchMigsConst([]*gce.InstanceGroupManager{}),
			fetchMig:    fetchMigFail,
			expectedErr: errFetchMig,
		},
		{
			name:        "cache fill failure, fallback failure",
			cache:       emptyCache(),
			fetchMigs:   fetchMigsFail,
			fetchMig:    fetchMigFail,
			expectedErr: errFetchMig,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := &mockAutoscalingGceClient{
				fetchMigs: tc.fetchMigs,
				fetchMig:  tc.fetchMig,
			}
			migLister := NewMigLister(tc.cache)
			provider := NewCachingMigInfoProvider(tc.cache, migLister, client, mig.GceRef().Project, 1, 0*time.Second, false, false)

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
		fetchMig               func(GceRef) (*gce.InstanceGroupManager, error)
		fetchMigTemplate       func(GceRef, string, bool) (*gce.InstanceTemplate, error)
		expectedTemplate       *gce.InstanceTemplate
		expectedCachedTemplate *gce.InstanceTemplate
		expectedErr            error
	}{
		{
			name: "template in cache",
			cache: NewGceCache().
				WithMigs(map[GceRef]Mig{mig.GceRef(): mig}).
				WithInstanceTemplateNames(map[GceRef]InstanceTemplateName{mig.GceRef(): {templateName, false}}).
				WithInstanceTemplates(map[GceRef]*gce.InstanceTemplate{mig.GceRef(): template}),
			expectedTemplate:       template,
			expectedCachedTemplate: template,
		},
		{
			name: "cache without template, fetch success",
			cache: NewGceCache().
				WithMigs(map[GceRef]Mig{mig.GceRef(): mig}).
				WithInstanceTemplateNames(map[GceRef]InstanceTemplateName{mig.GceRef(): {templateName, false}}),
			fetchMigTemplate:       fetchMigTemplateConst(template),
			expectedTemplate:       template,
			expectedCachedTemplate: template,
		},
		{
			name: "cache with old template, fetch success",
			cache: NewGceCache().
				WithMigs(map[GceRef]Mig{mig.GceRef(): mig}).
				WithInstanceTemplateNames(map[GceRef]InstanceTemplateName{mig.GceRef(): {templateName, false}}).
				WithInstanceTemplates(map[GceRef]*gce.InstanceTemplate{mig.GceRef(): oldTemplate}),
			fetchMigTemplate:       fetchMigTemplateConst(template),
			expectedTemplate:       template,
			expectedCachedTemplate: template,
		},
		{
			name: "cache without template, fetch failure",
			cache: NewGceCache().
				WithMigs(map[GceRef]Mig{mig.GceRef(): mig}).
				WithInstanceTemplateNames(map[GceRef]InstanceTemplateName{mig.GceRef(): {templateName, false}}),
			fetchMigTemplate: fetchMigTemplateFail,
			expectedErr:      errFetchMigTemplate,
		},
		{
			name: "cache with old template, fetch failure",
			cache: NewGceCache().
				WithMigs(map[GceRef]Mig{mig.GceRef(): mig}).
				WithInstanceTemplateNames(map[GceRef]InstanceTemplateName{mig.GceRef(): {templateName, false}}).
				WithInstanceTemplates(map[GceRef]*gce.InstanceTemplate{mig.GceRef(): oldTemplate}),
			fetchMigTemplate:       fetchMigTemplateFail,
			expectedCachedTemplate: oldTemplate,
			expectedErr:            errFetchMigTemplate,
		},
		{
			name:        "template name fetch failure",
			cache:       emptyCache(),
			fetchMigs:   fetchMigsFail,
			fetchMig:    fetchMigFail,
			expectedErr: errFetchMig,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := &mockAutoscalingGceClient{
				fetchMigs:        tc.fetchMigs,
				fetchMig:         tc.fetchMig,
				fetchMigTemplate: tc.fetchMigTemplate,
			}
			migLister := NewMigLister(tc.cache)
			provider := NewCachingMigInfoProvider(tc.cache, migLister, client, mig.GceRef().Project, 1, 0*time.Second, false, false)

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

func TestCreateInstancesState(t *testing.T) {
	testCases := []struct {
		name          string
		targetSize    int64
		actionSummary *gce.InstanceGroupManagerActionsSummary
		want          map[cloudprovider.InstanceState]int64
	}{
		{
			name:          "actionSummary is nil",
			targetSize:    10,
			actionSummary: nil,
			want:          nil,
		},
		{
			name:          "actionSummary is empty",
			targetSize:    10,
			actionSummary: &gce.InstanceGroupManagerActionsSummary{},
			want: map[cloudprovider.InstanceState]int64{
				cloudprovider.InstanceCreating: 0,
				cloudprovider.InstanceDeleting: 0,
				cloudprovider.InstanceRunning:  10,
			},
		},
		{
			name:       "actionSummary with data",
			targetSize: 30,
			actionSummary: &gce.InstanceGroupManagerActionsSummary{
				Abandoning:             2,
				Creating:               3,
				CreatingWithoutRetries: 5,
				Deleting:               7,
				Recreating:             11,
			},
			want: map[cloudprovider.InstanceState]int64{
				cloudprovider.InstanceCreating: 19,
				cloudprovider.InstanceDeleting: 9,
				cloudprovider.InstanceRunning:  11,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stateCount := createInstancesStateCount(tc.targetSize, tc.actionSummary)
			assert.Equal(t, tc.want, stateCount)
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
		fetchMig              func(GceRef) (*gce.InstanceGroupManager, error)
		fetchMigTemplate      func(GceRef, string, bool) (*gce.InstanceTemplate, error)
		expectedKubeEnv       KubeEnv
		expectedCachedKubeEnv KubeEnv
		expectedErr           error
	}{
		{
			name: "kube-env in cache",
			cache: NewGceCache().
				WithMigs(map[GceRef]Mig{mig.GceRef(): mig}).
				WithInstanceTemplateNames(map[GceRef]InstanceTemplateName{mig.GceRef(): {templateName, false}}).
				WithKubeEnvs(map[GceRef]KubeEnv{mig.GceRef(): kubeEnv}),
			expectedKubeEnv:       kubeEnv,
			expectedCachedKubeEnv: kubeEnv,
		},
		{
			name: "cache without kube-env, template in cache",
			cache: NewGceCache().
				WithMigs(map[GceRef]Mig{mig.GceRef(): mig}).
				WithInstanceTemplateNames(map[GceRef]InstanceTemplateName{mig.GceRef(): {templateName, false}}).
				WithInstanceTemplates(map[GceRef]*gce.InstanceTemplate{mig.GceRef(): template}),
			expectedKubeEnv:       kubeEnv,
			expectedCachedKubeEnv: kubeEnv,
		},
		{
			name: "cache without kube-env, fetch success",
			cache: NewGceCache().
				WithMigs(map[GceRef]Mig{mig.GceRef(): mig}).
				WithInstanceTemplateNames(map[GceRef]InstanceTemplateName{mig.GceRef(): {templateName, false}}),
			fetchMigTemplate:      fetchMigTemplateConst(template),
			expectedKubeEnv:       kubeEnv,
			expectedCachedKubeEnv: kubeEnv,
		},
		{
			name: "cache with old kube-env, new template cached",
			cache: NewGceCache().
				WithMigs(map[GceRef]Mig{mig.GceRef(): mig}).
				WithInstanceTemplateNames(map[GceRef]InstanceTemplateName{mig.GceRef(): {templateName, false}}).
				WithInstanceTemplates(map[GceRef]*gce.InstanceTemplate{mig.GceRef(): template}).
				WithKubeEnvs(map[GceRef]KubeEnv{mig.GceRef(): oldKubeEnv}),
			expectedKubeEnv:       kubeEnv,
			expectedCachedKubeEnv: kubeEnv,
		},
		{
			name: "cache with old kube-env, fetch success",
			cache: NewGceCache().
				WithMigs(map[GceRef]Mig{mig.GceRef(): mig}).
				WithInstanceTemplateNames(map[GceRef]InstanceTemplateName{mig.GceRef(): {templateName, false}}).
				WithInstanceTemplates(map[GceRef]*gce.InstanceTemplate{mig.GceRef(): oldTemplate}).
				WithKubeEnvs(map[GceRef]KubeEnv{mig.GceRef(): oldKubeEnv}),
			fetchMigTemplate:      fetchMigTemplateConst(template),
			expectedKubeEnv:       kubeEnv,
			expectedCachedKubeEnv: kubeEnv,
		},
		{
			name: "cache without kube-env, fetch failure",
			cache: NewGceCache().
				WithMigs(map[GceRef]Mig{mig.GceRef(): mig}).
				WithInstanceTemplateNames(map[GceRef]InstanceTemplateName{mig.GceRef(): {templateName, false}}),
			fetchMigTemplate: fetchMigTemplateFail,
			expectedErr:      errFetchMigTemplate,
		},
		{
			name: "cache with old kube-env, fetch failure",
			cache: NewGceCache().
				WithMigs(map[GceRef]Mig{mig.GceRef(): mig}).
				WithInstanceTemplateNames(map[GceRef]InstanceTemplateName{mig.GceRef(): {templateName, false}}).
				WithInstanceTemplates(map[GceRef]*gce.InstanceTemplate{mig.GceRef(): oldTemplate}).
				WithKubeEnvs(map[GceRef]KubeEnv{mig.GceRef(): oldKubeEnv}),
			fetchMigTemplate:      fetchMigTemplateFail,
			expectedCachedKubeEnv: oldKubeEnv,
			expectedErr:           errFetchMigTemplate,
		},
		{
			name:        "template name fetch failure",
			cache:       emptyCache(),
			fetchMigs:   fetchMigsFail,
			fetchMig:    fetchMigFail,
			expectedErr: errFetchMig,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := &mockAutoscalingGceClient{
				fetchMigs:        tc.fetchMigs,
				fetchMig:         tc.fetchMig,
				fetchMigTemplate: tc.fetchMigTemplate,
			}
			migLister := NewMigLister(tc.cache)
			provider := NewCachingMigInfoProvider(tc.cache, migLister, client, mig.GceRef().Project, 1, 0*time.Second, false, false)

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
			cache := NewGceCache().
				WithInstanceTemplateNames(map[GceRef]InstanceTemplateName{mig.GceRef(): {"template", false}}).
				WithInstanceTemplates(map[GceRef]*gce.InstanceTemplate{
					mig.GceRef(): {
						Name: "template",
						Properties: &gce.InstanceProperties{
							MachineType: tc.machine,
						},
					},
				}).
				WithMachines(map[MachineTypeKey]MachineType{
					{knownZone, tc.machine}: {
						Name:   tc.machine,
						CPU:    tc.cpu,
						Memory: tc.memory * units.MiB,
					},
				})
			client := &mockAutoscalingGceClient{
				fetchMachineType: tc.fetchMachineType,
			}
			migLister := NewMigLister(cache)
			provider := NewCachingMigInfoProvider(cache, migLister, client, mig.GceRef().Project, 1, 0*time.Second, false, false)
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
			cache := NewGceCache().
				WithMigs(map[GceRef]Mig{mig.gceRef: mig}).
				WithMigBaseName(map[GceRef]string{mig.GceRef(): "base-instance-name"})
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

func TestListInstancesInAllZonesWithMigs(t *testing.T) {
	testCases := []struct {
		name              string
		migs              map[GceRef]Mig
		fetchAllInstances func(string, string, string) ([]GceInstance, error)
		wantInstances     []GceInstance
		wantErr           bool
	}{
		{
			name:              "instance fetching failed",
			migs:              map[GceRef]Mig{mig1.GceRef(): mig1},
			fetchAllInstances: fetchAllInstancesInZone(map[string][]GceInstance{}),
			wantErr:           true,
		},
		{
			name:              "Successfully list mig instances in a single zone",
			migs:              map[GceRef]Mig{mig1.GceRef(): mig1},
			fetchAllInstances: fetchAllInstancesInZone(map[string][]GceInstance{"myzone1": {instance1, instance2}, "myzone2": {instance3}}),
			wantInstances:     []GceInstance{instance1, instance2},
		},
		{
			name:              "Successfully list mig instances in multiple zones",
			migs:              map[GceRef]Mig{mig1.GceRef(): mig1, mig2.GceRef(): mig2},
			fetchAllInstances: fetchAllInstancesInZone(map[string][]GceInstance{"myzone1": {instance1, instance2}, "myzone2": {instance3}}),
			wantInstances:     []GceInstance{instance1, instance2, instance3},
		},
		{
			name:              "Successfully list mig instances in one zones and got errors in another",
			migs:              map[GceRef]Mig{mig1.GceRef(): mig1, mig2.GceRef(): mig2},
			fetchAllInstances: fetchAllInstancesInZone(map[string][]GceInstance{"myzone1": {instance1, instance2}}),
			wantInstances:     []GceInstance{instance1, instance2},
			wantErr:           true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cache := NewGceCache().WithMigs(tc.migs)
			client := &mockAutoscalingGceClient{
				fetchAllInstances: tc.fetchAllInstances,
			}
			migLister := NewMigLister(cache)
			provider := &cachingMigInfoProvider{
				cache:                  cache,
				migLister:              migLister,
				gceClient:              client,
				concurrentGceRefreshes: 1,
			}
			instances, err := provider.listInstancesInAllZonesWithMigs()

			if tc.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.ElementsMatch(t, tc.wantInstances, instances)
		})
	}
}

func TestGroupInstancesToMigs(t *testing.T) {
	testCases := []struct {
		name      string
		instances []GceInstance
		want      map[GceRef][]GceInstance
	}{
		{
			name: "no instances",
			want: map[GceRef][]GceInstance{},
		},
		{
			name:      "instances from multiple migs including unknown migs",
			instances: []GceInstance{instance1, instance2, instance3, instance4, instance5},
			want: map[GceRef][]GceInstance{
				mig1.GceRef(): {instance1, instance2},
				mig2.GceRef(): {instance3},
				{}:            {instance4, instance5},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			groupedInstances := groupInstancesToMigs(tc.instances)
			assert.Equal(t, tc.want, groupedInstances)
		})
	}
}

func TestIsMigInstancesConsistent(t *testing.T) {
	testCases := []struct {
		name                   string
		mig                    Mig
		migToInstances         map[GceRef][]GceInstance
		migInstancesStateCache map[GceRef]map[cloudprovider.InstanceState]int64
		want                   bool
	}{
		{
			name:           "instance not found",
			mig:            mig1,
			migToInstances: map[GceRef][]GceInstance{},
			want:           false,
		},
		{
			name:           "instanceState not found",
			mig:            mig1,
			migToInstances: map[GceRef][]GceInstance{mig1.GceRef(): {instance1, instance2}},
			want:           false,
		},
		{
			name:           "inconsistent number of instances",
			mig:            mig1,
			migToInstances: map[GceRef][]GceInstance{mig1.GceRef(): {instance1, instance2}},
			migInstancesStateCache: map[GceRef]map[cloudprovider.InstanceState]int64{
				mig1.GceRef(): {
					cloudprovider.InstanceCreating: 2,
					cloudprovider.InstanceDeleting: 3,
					cloudprovider.InstanceRunning:  4,
				},
			},
			want: false,
		},
		{
			name:           "consistent number of instances",
			mig:            mig1,
			migToInstances: map[GceRef][]GceInstance{mig1.GceRef(): {instance1, instance2}},
			migInstancesStateCache: map[GceRef]map[cloudprovider.InstanceState]int64{
				mig1.GceRef(): {
					cloudprovider.InstanceCreating: 1,
					cloudprovider.InstanceDeleting: 0,
					cloudprovider.InstanceRunning:  1,
				},
			},
			want: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cache := NewGceCache().WithMigInstancesStateCount(tc.migInstancesStateCache)
			provider := &cachingMigInfoProvider{
				cache: cache,
			}
			got := provider.isMigInstancesConsistent(tc.mig, tc.migToInstances)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestIsMigInCreatingOrDeletingInstanceState(t *testing.T) {
	testCases := []struct {
		name                   string
		mig                    Mig
		migInstancesStateCache map[GceRef]map[cloudprovider.InstanceState]int64
		want                   bool
	}{
		{
			name: "instanceState not found",
			mig:  mig1,
			want: false,
		},
		{
			name: "in creating state",
			mig:  mig1,
			migInstancesStateCache: map[GceRef]map[cloudprovider.InstanceState]int64{
				mig1.GceRef(): {
					cloudprovider.InstanceCreating: 2,
					cloudprovider.InstanceDeleting: 0,
					cloudprovider.InstanceRunning:  1,
				},
			},
			want: true,
		},
		{
			name: "in deleting state",
			mig:  mig1,
			migInstancesStateCache: map[GceRef]map[cloudprovider.InstanceState]int64{
				mig1.GceRef(): {
					cloudprovider.InstanceCreating: 0,
					cloudprovider.InstanceDeleting: 1,
					cloudprovider.InstanceRunning:  0,
				},
			},
			want: true,
		},
		{
			name: "not in creating or deleting states",
			mig:  mig1,
			migInstancesStateCache: map[GceRef]map[cloudprovider.InstanceState]int64{
				mig1.GceRef(): {
					cloudprovider.InstanceCreating: 0,
					cloudprovider.InstanceDeleting: 0,
					cloudprovider.InstanceRunning:  1,
				},
			},
			want: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cache := NewGceCache().WithMigInstancesStateCount(tc.migInstancesStateCache)
			provider := &cachingMigInfoProvider{
				cache: cache,
			}
			got := provider.isMigCreatingOrDeletingInstances(tc.mig)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestUpdateMigInstancesCache(t *testing.T) {
	testCases := []struct {
		name                   string
		migs                   map[GceRef]Mig
		migToInstances         map[GceRef][]GceInstance
		fetchMigInstances      []GceInstance
		wantInstances          map[GceRef][]GceInstance
		migInstancesStateCache map[GceRef]map[cloudprovider.InstanceState]int64
	}{
		{
			name: "inconsistent mig instance state",
			migs: map[GceRef]Mig{mig1.GceRef(): mig1},
			migToInstances: map[GceRef][]GceInstance{
				mig1.GceRef(): {instance1},
			},
			migInstancesStateCache: map[GceRef]map[cloudprovider.InstanceState]int64{
				mig1.GceRef(): {cloudprovider.InstanceRunning: 2, cloudprovider.InstanceDeleting: 0, cloudprovider.InstanceCreating: 0},
			},
			fetchMigInstances: []GceInstance{instance1, instance2},
			wantInstances:     map[GceRef][]GceInstance{mig1.GceRef(): {instance1, instance2}},
		},
		{
			name: "mig with instance in creating or deleting state",
			migs: map[GceRef]Mig{mig1.GceRef(): mig1},
			migInstancesStateCache: map[GceRef]map[cloudprovider.InstanceState]int64{
				mig1.GceRef(): {cloudprovider.InstanceRunning: 0, cloudprovider.InstanceDeleting: 0, cloudprovider.InstanceCreating: 2},
			},
			fetchMigInstances: []GceInstance{instance1, instance2},
			wantInstances:     map[GceRef][]GceInstance{mig1.GceRef(): {instance1, instance2}},
		},
		{
			name: "consistent mig instance state",
			migs: map[GceRef]Mig{mig1.GceRef(): mig1},
			migToInstances: map[GceRef][]GceInstance{
				mig1.GceRef(): {instance1, instance2},
			},
			migInstancesStateCache: map[GceRef]map[cloudprovider.InstanceState]int64{
				mig1.GceRef(): {cloudprovider.InstanceRunning: 2, cloudprovider.InstanceDeleting: 0, cloudprovider.InstanceCreating: 0},
			},
			wantInstances: map[GceRef][]GceInstance{mig1.GceRef(): {instance1, instance2}},
		},
		{
			name: "mix of consistent and inconsistent states",
			migs: map[GceRef]Mig{mig1.GceRef(): mig1, mig2.GceRef(): mig2},
			migToInstances: map[GceRef][]GceInstance{
				mig1.GceRef(): {instance1, instance2},
			},
			migInstancesStateCache: map[GceRef]map[cloudprovider.InstanceState]int64{
				mig1.GceRef(): {cloudprovider.InstanceRunning: 2, cloudprovider.InstanceDeleting: 0, cloudprovider.InstanceCreating: 0},
				mig2.GceRef(): {cloudprovider.InstanceRunning: 1, cloudprovider.InstanceDeleting: 0, cloudprovider.InstanceCreating: 0},
			},
			fetchMigInstances: []GceInstance{instance3},
			wantInstances: map[GceRef][]GceInstance{
				mig1.GceRef(): {instance1, instance2},
				mig2.GceRef(): {instance3},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cache := NewGceCache().
				WithMigs(tc.migs).
				WithMigInstancesStateCount(tc.migInstancesStateCache)
			migLister := NewMigLister(cache)
			client := &mockAutoscalingGceClient{
				fetchMigInstances: fetchMigInstancesConst(tc.fetchMigInstances),
			}
			provider := &cachingMigInfoProvider{
				cache:        cache,
				migLister:    migLister,
				gceClient:    client,
				timeProvider: &realTime{},
			}
			err := provider.updateMigInstancesCache(tc.migToInstances)
			assert.NoError(t, err)
			for migRef, want := range tc.wantInstances {
				instances, found := cache.GetMigInstances(migRef)
				assert.True(t, found)
				assert.Equal(t, want, instances)
			}
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
	return NewGceCache().
		WithMigs(map[GceRef]Mig{mig.GceRef(): mig, mig1.GceRef(): mig1})
}

func fetchMigsFail(_ string) ([]*gce.InstanceGroupManager, error) {
	return nil, errFetchMigs
}
func fetchMigFail(_ GceRef) (*gce.InstanceGroupManager, error) {
	return nil, errFetchMig
}

func fetchMigsConst(migs []*gce.InstanceGroupManager) func(string) ([]*gce.InstanceGroupManager, error) {
	return func(string) ([]*gce.InstanceGroupManager, error) {
		return migs, nil
	}
}

func fetchMigConst(igm *gce.InstanceGroupManager) func(GceRef) (*gce.InstanceGroupManager, error) {
	return func(GceRef) (*gce.InstanceGroupManager, error) {
		return igm, nil
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

func (gc *GceCache) WithMigs(m map[GceRef]Mig) *GceCache {
	gc.migs = m
	return gc
}

func (gc *GceCache) WithInstances(m map[GceRef][]GceInstance) *GceCache {
	gc.instances = m
	return gc
}

func (gc *GceCache) WithInstancesUpdateTime(m map[GceRef]time.Time) *GceCache {
	gc.instancesUpdateTime = m
	return gc
}

func (gc *GceCache) WithInstancesToMig(m map[GceRef]GceRef) *GceCache {
	gc.instancesToMig = m
	return gc
}

func (gc *GceCache) WithInstancesFromUnknownMig(m map[GceRef]bool) *GceCache {
	gc.instancesFromUnknownMig = m
	return gc
}

func (gc *GceCache) WithAutoscalingOptions(m map[GceRef]map[string]string) *GceCache {
	gc.autoscalingOptionsCache = m
	return gc
}

func (gc *GceCache) WithMachines(m map[MachineTypeKey]MachineType) *GceCache {
	gc.machinesCache = m
	return gc
}

func (gc *GceCache) WithMigTargetSize(m map[GceRef]int64) *GceCache {
	gc.migTargetSizeCache = m
	return gc
}

func (gc *GceCache) WithMigIsStable(m map[GceRef]bool) *GceCache {
	gc.migIsStableCache = m
	return gc
}

func (gc *GceCache) WithMigBaseName(m map[GceRef]string) *GceCache {
	gc.migBaseNameCache = m
	return gc
}

func (gc *GceCache) WithMigInstancesStateCount(m map[GceRef]map[cloudprovider.InstanceState]int64) *GceCache {
	gc.migInstancesStateCountCache = m
	return gc
}

func (gc *GceCache) WithListManagedInstancesResults(m map[GceRef]string) *GceCache {
	gc.listManagedInstancesResultsCache = m
	return gc
}

func (gc *GceCache) WithInstanceTemplateNames(m map[GceRef]InstanceTemplateName) *GceCache {
	gc.instanceTemplateNameCache = m
	return gc
}

func (gc *GceCache) WithInstanceTemplates(m map[GceRef]*gce.InstanceTemplate) *GceCache {
	gc.instanceTemplatesCache = m
	return gc
}

func (gc *GceCache) WithKubeEnvs(m map[GceRef]KubeEnv) *GceCache {
	gc.kubeEnvCache = m
	return gc
}

func fetchMigTemplateFail(_ GceRef, _ string, _ bool) (*gce.InstanceTemplate, error) {
	return nil, errFetchMigTemplate
}

func fetchMigTemplateConst(template *gce.InstanceTemplate) func(GceRef, string, bool) (*gce.InstanceTemplate, error) {
	return func(GceRef, string, bool) (*gce.InstanceTemplate, error) {
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

func fetchAllInstancesInZone(allInstances map[string][]GceInstance) func(string, string, string) ([]GceInstance, error) {
	return func(project, zone, filter string) ([]GceInstance, error) {
		instances, found := allInstances[zone]
		if !found {
			return nil, errors.New("")
		}
		return instances, nil
	}
}
