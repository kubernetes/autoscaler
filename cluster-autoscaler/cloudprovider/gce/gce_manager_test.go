/*
Copyright 2017 The Kubernetes Authors.

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
	"regexp"
	"strings"
	"testing"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/utils/units"

	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	gce "google.golang.org/api/compute/v1"
)

const (
	projectId          = "project1"
	zoneB              = "us-central1-b"
	zoneC              = "us-central1-c"
	zoneF              = "us-central1-f"
	region             = "us-central1"
	defaultPoolMigName = "gke-cluster-1-default-pool"
	defaultPool        = "default-pool"
	extraPoolMigName   = "gke-cluster-1-extra-pool-323233232"
	extraPool2MigName  = "gke-cluster-1-extra-pool2-323233232"
	extraPool          = "extra-pool"
	clusterName        = "cluster1"

	gceMigA = "gce-mig-a"
	gceMigB = "gce-mig-b"
)

const instanceGroupManagerResponseTemplate = `{
  "kind": "compute#instanceGroupManager",
  "id": "3213213219",
  "creationTimestamp": "2017-09-15T04:47:24.687-07:00",
  "name": "%s",
  "zone": "https://www.googleapis.com/compute/v1/projects/project1/zones/%s",
  "instanceTemplate": "https://www.googleapis.com/compute/v1/projects/project1/global/instanceTemplates/%s",
  "instanceGroup": "https://www.googleapis.com/compute/v1/projects/project1/zones/%s/instanceGroups/%s",
  "baseInstanceName": "%s",
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
  "targetSize": %v,
  "selfLink": "https://www.googleapis.com/compute/v1/projects/project1/zones/%s/instanceGroupManagers/%s"
}
`

const instanceGroupManagerTargetSize4ResponseTemplate = `{
  "kind": "compute#instanceGroupManager",
  "id": "3213213219",
  "creationTimestamp": "2017-09-15T04:47:24.687-07:00",
  "name": "%s",
  "zone": "https://www.googleapis.com/compute/v1/projects/project1/zones/%s",
  "instanceTemplate": "https://www.googleapis.com/compute/v1/projects/project1/global/instanceTemplates/%s",
  "instanceGroup": "https://www.googleapis.com/compute/v1/projects/project1/zones/%s/instanceGroups/%s",
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
  "targetSize": 4,
  "selfLink": "https://www.googleapis.com/compute/v1/projects/project1/zones/%s/instanceGroupManagers/%s"
}
`

const instanceTemplate = `
{
 "kind": "compute#instanceTemplate",
 "id": "28701103232323232",
 "creationTimestamp": "2017-09-15T04:47:21.577-07:00",
 "name": "gke-cluster-1-default-pool",
 "description": "",
 "properties": {
  "tags": {
   "items": [
    "gke-cluster-1-fc0afeeb-node"
   ]
  },
  "machineType": "n1-standard-1",
  "canIpForward": true,
  "networkInterfaces": [
   {
    "kind": "compute#networkInterface",
    "network": "https://www.googleapis.com/compute/v1/projects/project1/global/networks/default",
    "subnetwork": "https://www.googleapis.com/compute/v1/projects/project1/regions/us-central1/subnetworks/default",
    "accessConfigs": [
     {
      "kind": "compute#accessConfig",
      "type": "ONE_TO_ONE_NAT",
      "name": "external-nat"
     }
    ]
   }
  ],
  "disks": [
   {
    "kind": "compute#attachedDisk",
    "type": "PERSISTENT",
    "mode": "READ_WRITE",
    "boot": true,
    "initializeParams": {
     "sourceImage": "https://www.googleapis.com/compute/v1/projects/gke-node-images/global/images/cos-stable-60-9592-84-0",
     "diskSizeGb": "100",
     "diskType": "pd-standard"
    },
    "autoDelete": true
   }
  ],
  "metadata": {
   "kind": "compute#metadata",
   "fingerprint": "F7n_RsHD3ng=",
   "items": [
		{
		 "key": "kube-env",
		 "value": "ALLOCATE_NODE_CIDRS: \"true\"\n"
		},
		{
		 "key": "user-data",
		 "value": "#cloud-config\n\nwrite_files:\n  - path: /etc/systemd/system/kube-node-installation.service\n    "
		},
		{
		 "key": "gci-update-strategy",
		 "value": "update_disabled"
		},
		{
		 "key": "gci-ensure-gke-docker",
		 "value": "true"
		},
		{
		 "key": "configure-sh",
		 "value": "#!/bin/bash\n\n# Copyright 2016 The Kubernetes Authors.\n#\n# Licensed under the Apache License, "
		},
		{
		 "key": "cluster-name",
		 "value": "cluster-1"
		}
	   ]
	  },
  "serviceAccounts": [
   {
    "email": "default",
    "scopes": [
     "https://www.googleapis.com/auth/compute",
     "https://www.googleapis.com/auth/devstorage.read_only",
     "https://www.googleapis.com/auth/logging.write",
     "https://www.googleapis.com/auth/monitoring.write",
     "https://www.googleapis.com/auth/servicecontrol",
     "https://www.googleapis.com/auth/service.management.readonly",
     "https://www.googleapis.com/auth/trace.append"
    ]
   }
  ],
  "scheduling": {
   "onHostMaintenance": "MIGRATE",
   "automaticRestart": true,
   "preemptible": false
  }
 },
 "selfLink": "https://www.googleapis.com/compute/v1/projects/project1/global/instanceTemplates/gke-cluster-1-default-pool-f7607aac"
}`

const fourRunningInstancesManagedInstancesResponseTemplate = `{
  "managedInstances": [
    {
      "instance": "https://www.googleapis.com/compute/v1/projects/project1/zones/%s/instances/%s-f7607aac-9j4g",
      "id": "1974815549671473983",
      "instanceStatus": "RUNNING",
      "currentAction": "NONE"
    },
    {
      "instance": "https://www.googleapis.com/compute/v1/projects/project1/zones/%s/instances/%s-f7607aac-c63g",
      "currentAction": "RUNNING",
      "id": "197481554967143333",
      "instanceStatus": "RUNNING",
      "currentAction": "NONE"
    },
    {
      "instance": "https://www.googleapis.com/compute/v1/projects/project1/zones/%s/instances/%s-f7607aac-dck1",
      "id": "4462422841867240255",
      "instanceStatus": "RUNNING",
      "currentAction": "NONE"
    },
    {
      "instance": "https://www.googleapis.com/compute/v1/projects/project1/zones/%s/instances/%s-f7607aac-f1hm",
      "id": "6309299611401323327",
      "instanceStatus": "RUNNING",
      "currentAction": "NONE"
    }
  ]
}`

const oneRunningInstanceManagedInstancesResponseTemplate = `{
  "managedInstances": [
    {
      "instance": "https://www.googleapis.com/compute/v1/projects/project1/zones/%s/instances/%s-gdf607aac-9j4g",
      "id": "1974815323221473983",
      "instanceStatus": "RUNNING",
      "currentAction": "NONE"
    }
  ]
}`

const listInstanceGroupManagerResponsePartTemplate = `
  {
   "kind": "compute#instanceGroupManager",
   "id": "9012769713544464023",
   "creationTimestamp": "2019-03-26T07:34:32.082-07:00",
   "name": "%v",
   "zone": "https://www.googleapis.com/compute/v1/projects/project1/zones/%v",
   "instanceTemplate": "https://www.googleapis.com/compute/v1/projects/project1/global/instanceTemplates/gke-blah-default-pool-67b773a0",
   "versions": [
    {
     "instanceTemplate": "https://www.googleapis.com/compute/v1/projects/project1/global/instanceTemplates/gke-blah-default-pool-67b773a0",
     "targetSize": {
      "calculated": 1
     }
    }
   ],
   "instanceGroup": "https://www.googleapis.com/compute/v1/projects/lukaszos-gke-dev2/zones/%v/instanceGroups/%v",
   "baseInstanceName": "gke-blah-default-pool-67b773a0",
   "fingerprint": "ASJwTpesjDI=",
   "currentActions": {
    "none": 1,
    "creating": 0,
    "creatingWithoutRetries": 0,
    "verifying": 0,
    "recreating": 0,
    "deleting": 0,
    "abandoning": 0,
    "restarting": 0,
    "refreshing": 0
   },
   "status": {
    "isStable": true
   },
   "targetSize": %v,
   "selfLink": "https://www.googleapis.com/compute/v1/projects/lukaszos-gke-dev2/zones/us-west1-b/instanceGroupManagers/gke-blah-default-pool-67b773a0-grp",
   "updatePolicy": {
    "type": "OPPORTUNISTIC",
    "minimalAction": "REPLACE",
    "maxSurge": {
     "fixed": 1,
     "calculated": 1
    },
    "maxUnavailable": {
     "fixed": 1,
     "calculated": 1
    }
   }
  }
`

func buildDefaultInstanceGroupManagerResponse(zone string) string {
	return buildInstanceGroupManagerResponse(zone, defaultPoolMigName, 3)
}

func buildInstanceGroupManagerResponse(zone string, instanceGroup string, targetSize uint64) string {
	return fmt.Sprintf(instanceGroupManagerResponseTemplate, instanceGroup, zone, instanceGroup, zone, instanceGroup, instanceGroup, targetSize, zone, instanceGroup)
}

func buildFourRunningInstancesOnDefaultMigManagedInstancesResponse(zone string) string {
	return buildFourRunningInstancesManagedInstancesResponse(zone, defaultPoolMigName)
}

func buildFourRunningInstancesManagedInstancesResponse(zone string, instanceGroup string) string {
	return fmt.Sprintf(fourRunningInstancesManagedInstancesResponseTemplate, zone, instanceGroup, zone, instanceGroup, zone, instanceGroup, zone, instanceGroup)
}

func buildOneRunningInstanceOnExtraPoolMigManagedInstancesResponse(zone string) string {
	return buildOneRunningInstanceManagedInstancesResponse(zone, extraPoolMigName)
}

func buildOneRunningInstanceManagedInstancesResponse(zone string, instanceGroup string) string {
	return fmt.Sprintf(oneRunningInstanceManagedInstancesResponseTemplate, zone, instanceGroup)
}

func buildListInstanceGroupManagersResponsePart(name, zone string, targetSize uint64) string {
	return fmt.Sprintf(listInstanceGroupManagerResponsePartTemplate, name, zone, zone, name, targetSize)
}

func buildListInstanceGroupManagersResponse(listInstanceGroupManagerResponseParts ...string) string {
	return `{
 "kind": "compute#instanceGroupManagerList",
 "id": "blah",
 "items": [` +
		strings.Join(listInstanceGroupManagerResponseParts, ",") +
		`], "selfLink": "https://blah"}`
}

func newTestGceManager(t *testing.T, testServerURL string, regional bool) *gceManagerImpl {
	gceService := newTestAutoscalingGceClient(t, projectId, testServerURL)

	// Override wait for op timeouts.
	gceService.operationWaitTimeout = 50 * time.Millisecond
	gceService.operationPollInterval = 1 * time.Millisecond

	cache := &GceCache{
		migs:                     make(map[GceRef]Mig),
		GceService:               gceService,
		instanceRefToMigRef:      make(map[GceRef]GceRef),
		instancesFromUnknownMigs: make(map[GceRef]struct{}),
		machinesCache: map[MachineTypeKey]machinesCacheValue{
			{"us-central1-b", "n1-standard-1"}: {&gce.MachineType{GuestCpus: 1, MemoryMb: 1}, nil},
			{"us-central1-c", "n1-standard-1"}: {&gce.MachineType{GuestCpus: 1, MemoryMb: 1}, nil},
			{"us-central1-f", "n1-standard-1"}: {&gce.MachineType{GuestCpus: 1, MemoryMb: 1}, nil},
		},
		migTargetSizeCache:     map[GceRef]int64{},
		instanceTemplatesCache: map[GceRef]*gce.InstanceTemplate{},
		migBaseNameCache:       map[GceRef]string{},
		concurrentGceRefreshes: 1,
	}
	manager := &gceManagerImpl{
		cache:                        cache,
		migTargetSizesProvider:       NewCachingMigTargetSizesProvider(cache, gceService, projectId),
		migInstanceTemplatesProvider: NewCachingMigInstanceTemplatesProvider(cache, gceService),
		GceService:                   gceService,
		projectId:                    projectId,
		regional:                     regional,
		templates:                    &GceTemplateBuilder{},
		explicitlyConfigured:         make(map[GceRef]bool),
		concurrentGceRefreshes:       1,
	}
	if regional {
		manager.location = region
	} else {
		manager.location = zoneB
	}

	return manager
}

const deleteInstancesResponse = `{
  "kind": "compute#operation",
  "id": "8554136016090105726",
  "name": "operation-1505802641136-55984ff86d980-a99e8c2b-0c8aaaaa",
  "zone": "https://www.googleapis.com/compute/v1/projects/project1/zones/us-central1-a",
  "operationType": "compute.instanceGroupManagers.deleteInstances",
  "targetLink": "https://www.googleapis.com/compute/v1/projects/project1/zones/us-central1-a/instanceGroupManagers/gke-cluster-1-default-pool-f7607aac-grp",
  "targetId": "5382990249302819619",
  "status": "DONE",
  "user": "user@example.com",
  "progress": 100,
  "insertTime": "2017-09-18T23:30:41.612-07:00",
  "startTime": "2017-09-18T23:30:41.618-07:00",
  "endTime": "2017-09-18T23:30:41.618-07:00",
  "selfLink": "https://www.googleapis.com/compute/v1/projects/project1/zones/us-central1-a/operations/operation-1505802641136-55984ff86d980-a99e8c2b-0c8aaaaa"
}`

const deleteInstancesOperationResponse = `
{
  "kind": "compute#operation",
  "id": "8554136016090105726",
  "name": "operation-1505802641136-55984ff86d980-a99e8c2b-0c8aaaaa",
  "zone": "https://www.googleapis.com/compute/v1/projects/project1/zones/us-central1-a",
  "operationType": "compute.instanceGroupManagers.deleteInstances",
  "targetLink": "https://www.googleapis.com/compute/v1/projects/project1/zones/us-central1-a/instanceGroupManagers/gke-cluster-1-default-pool-f7607aac-grp",
  "targetId": "5382990249302819619",
  "status": "DONE",
  "user": "user@example.com",
  "progress": 100,
  "insertTime": "2017-09-18T23:30:41.612-07:00",
  "startTime": "2017-09-18T23:30:41.618-07:00",
  "endTime": "2017-09-18T23:30:41.618-07:00",
  "selfLink": "https://www.googleapis.com/compute/v1/projects/project1/zones/us-central1-a/operations/operation-1505802641136-55984ff86d980-a99e8c2b-0c8aaaaa"
}`

func setupTestDefaultPool(manager *gceManagerImpl, setupBaseName bool) *gceMig {
	mig := &gceMig{
		gceRef: GceRef{
			Name:    defaultPoolMigName,
			Zone:    zoneB,
			Project: projectId,
		},
		gceManager: manager,
		minSize:    1,
		maxSize:    11,
	}
	manager.cache.migs[mig.GceRef()] = mig
	if setupBaseName {
		manager.cache.migBaseNameCache[mig.GceRef()] = defaultPoolMigName
	}
	return mig
}

func setupTestExtraPool(manager *gceManagerImpl, setupBaseName bool) *gceMig {
	mig := &gceMig{
		gceRef: GceRef{
			Name:    extraPoolMigName,
			Zone:    zoneB,
			Project: projectId,
		},
		gceManager: manager,
		minSize:    0,
		maxSize:    1000,
	}
	manager.cache.migs[mig.GceRef()] = mig
	if setupBaseName {
		manager.cache.migBaseNameCache[mig.GceRef()] = extraPoolMigName
	}
	return mig
}

func setupTestExtraPool2(manager *gceManagerImpl, setupBaseName bool) *gceMig {
	mig := &gceMig{
		gceRef: GceRef{
			Name:    extraPool2MigName,
			Zone:    zoneC,
			Project: projectId,
		},
		gceManager: manager,
		minSize:    0,
		maxSize:    1000,
	}
	manager.cache.migs[mig.GceRef()] = mig
	if setupBaseName {
		manager.cache.migBaseNameCache[mig.GceRef()] = extraPool2MigName
	}
	return mig
}

func TestDeleteInstances(t *testing.T) {
	server := NewHttpServerMock()
	defer server.Close()
	g := newTestGceManager(t, server.URL, false)

	setupTestDefaultPool(g, false)
	setupTestExtraPool(g, true)

	// Get basename for defaultPool
	server.On("handle", "/project1/zones/us-central1-b/instanceGroupManagers/gke-cluster-1-default-pool").Return(buildDefaultInstanceGroupManagerResponse(zoneB)).Once()

	// Regenerate instances for defaultPool
	server.On("handle", "/project1/zones/us-central1-b/instanceGroupManagers/gke-cluster-1-default-pool/listManagedInstances").Return(buildFourRunningInstancesOnDefaultMigManagedInstancesResponse(zoneB)).Once()

	server.On("handle", "/project1/zones/us-central1-b/instanceGroupManagers/gke-cluster-1-default-pool/deleteInstances").Return(deleteInstancesResponse).Once()
	server.On("handle", "/project1/zones/us-central1-b/operations/operation-1505802641136-55984ff86d980-a99e8c2b-0c8aaaaa").Return(deleteInstancesOperationResponse).Once()

	instances := []GceRef{
		{
			Project: projectId,
			Zone:    zoneB,
			Name:    "gke-cluster-1-default-pool-f7607aac-f1hm",
		},
		{
			Project: projectId,
			Zone:    zoneB,
			Name:    "gke-cluster-1-default-pool-f7607aac-c63g",
		},
	}

	err := g.DeleteInstances(instances)
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, server)

	// Regenerate instances for extraPool (no basename call because it is already in cache)
	server.On("handle", "/project1/zones/us-central1-b/instanceGroupManagers/gke-cluster-1-extra-pool-323233232/listManagedInstances").Return(buildOneRunningInstanceOnExtraPoolMigManagedInstancesResponse(zoneB)).Once()

	// Fail on deleting instances from different MIGs.
	instances = []GceRef{
		{
			Project: projectId,
			Zone:    zoneB,
			Name:    "gke-cluster-1-default-pool-f7607aac-f1hm",
		},
		{
			Project: projectId,
			Zone:    zoneB,
			Name:    "gke-cluster-1-extra-pool-323233232-gdf607aac-9j4g",
		},
	}

	err = g.DeleteInstances(instances)
	assert.Error(t, err)
	assert.Equal(t, "cannot delete instances which don't belong to the same MIG.", err.Error())
	mock.AssertExpectationsForObjects(t, server)
}

// TODO; make Test*MigSize tests use MigTargetSizesProvider mock and move mocking API server to tests of cachingMigTargetSizesProvider

const setMigSizeResponse = `{
  "kind": "compute#operation",
  "id": "7558996788000226430",
  "name": "operation-1505739408819-5597646964339-eb839c88-28805931",
  "zone": "https://www.googleapis.com/compute/v1/projects/project1/zones/us-central1-a",
  "operationType": "compute.instanceGroupManagers.resize",
  "targetLink": "https://www.googleapis.com/compute/v1/projects/project1/zones/us-central1-a/instanceGroupManagers/gke-cluster-1-default-pool-f7607aac-grp",
  "targetId": "5382990249302819619",
  "status": "DONE",
  "user": "user@example.com",
  "progress": 100,
  "insertTime": "2017-09-18T05:56:49.227-07:00",
  "startTime": "2017-09-18T05:56:49.230-07:00",
  "endTime": "2017-09-18T05:56:49.230-07:00",
  "selfLink": "https://www.googleapis.com/compute/v1/projects/project1/zones/us-central1-a/operations/operation-1505739408819-5597646964339-eb839c88-28805931"
}`

const setMigSizeOperationResponse = `{
  "kind": "compute#operation",
  "id": "7558996788000226430",
  "name": "operation-1505739408819-5597646964339-eb839c88-28805931",
  "zone": "https://www.googleapis.com/compute/v1/projects/project1/zones/us-central1-a",
  "operationType": "compute.instanceGroupManagers.resize",
  "targetLink": "https://www.googleapis.com/compute/v1/projects/project1/zones/us-central1-a/instanceGroupManagers/gke-cluster-1-default-pool-f7607aac-grp",
  "targetId": "5382990249302819619",
  "status": "DONE",
  "user": "user@example.com",
  "progress": 100,
  "insertTime": "2017-09-18T05:56:49.227-07:00",
  "startTime": "2017-09-18T05:56:49.230-07:00",
  "endTime": "2017-09-18T05:56:49.230-07:00",
  "selfLink": "https://www.googleapis.com/compute/v1/projects/project1/zones/us-central1-a/operations/operation-1505739408819-5597646964339-eb839c88-28805931"
}`

func TestGetAndSetMigSize(t *testing.T) {
	server := NewHttpServerMock()
	defer server.Close()
	g := newTestGceManager(t, server.URL, false)

	extraPoolMig := setupTestExtraPool(g, true)
	defaultPoolMig := setupTestDefaultPool(g, true)

	server.On("handle", "/project1/zones/us-central1-b/instanceGroupManagers").Return(
		buildListInstanceGroupManagersResponse(
			buildListInstanceGroupManagersResponsePart(defaultPoolMigName, zoneB, 7),
			buildListInstanceGroupManagersResponsePart(extraPoolMigName, zoneB, 8),
		)).Once()

	// getting size for defaultPoolMig should trigger listing all the InstanceGroupManagers
	defaultPoolMigSize, err := g.GetMigSize(defaultPoolMig)
	assert.NoError(t, err)
	assert.Equal(t, int64(7), defaultPoolMigSize)
	mock.AssertExpectationsForObjects(t, server)

	// extra queries for defaultPoolMig and extraPoolMig should not result in any extra API calls
	defaultPoolMigSize, err = g.GetMigSize(defaultPoolMig)
	assert.NoError(t, err)
	assert.Equal(t, int64(7), defaultPoolMigSize)

	extraPoolMigSize, err := g.GetMigSize(extraPoolMig)
	assert.NoError(t, err)
	assert.Equal(t, int64(8), extraPoolMigSize)
	mock.AssertExpectationsForObjects(t, server)

	// set target size for extraPoolMig; will require resize API call and API call for polling for resize operation
	server.On("handle", fmt.Sprintf("/project1/zones/us-central1-b/instanceGroupManagers/%s/resize", extraPoolMigName)).Return(setMigSizeResponse).Once()
	server.On("handle", "/project1/zones/us-central1-b/operations/operation-1505739408819-5597646964339-eb839c88-28805931").Return(setMigSizeOperationResponse).Once()
	err = g.SetMigSize(extraPoolMig, 4)
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, server)

	// query for size of resized extraPoolMig; no extra API calls
	extraPoolMigSize, err = g.GetMigSize(extraPoolMig)
	assert.NoError(t, err)
	assert.Equal(t, int64(4), extraPoolMigSize)
	mock.AssertExpectationsForObjects(t, server)

	// register another pool: extraPool2; pool uses mig in different zone
	extraPool2Mig := setupTestExtraPool2(g, true)

	// query for size of resized extraPool2Mig; execting API call refreshing target sizes
	server.On("handle", "/project1/zones/us-central1-b/instanceGroupManagers").Return(
		buildListInstanceGroupManagersResponse(
			buildListInstanceGroupManagersResponsePart(defaultPoolMigName, zoneB, 7),
			buildListInstanceGroupManagersResponsePart(extraPoolMigName, zoneB, 8),
		)).Once()
	server.On("handle", "/project1/zones/us-central1-c/instanceGroupManagers").Return(
		buildListInstanceGroupManagersResponse(
			buildListInstanceGroupManagersResponsePart(extraPool2MigName, zoneC, 9)),
	).Once()

	extraPool2MigSize, err := g.GetMigSize(extraPool2Mig)
	assert.NoError(t, err)
	assert.Equal(t, int64(9), extraPool2MigSize)
	mock.AssertExpectationsForObjects(t, server)

	// another query for size of extraPool2Mig will not result in any API calls
	extraPool2MigSize, err = g.GetMigSize(extraPool2Mig)
	assert.NoError(t, err)
	assert.Equal(t, int64(9), extraPool2MigSize)
	mock.AssertExpectationsForObjects(t, server)

	// let's invalidate target size cache
	// TODO we should probably call g.Refresh here but that imples more API calls. Leaving just partial cache invalidation for now
	g.cache.InvalidateAllMigTargetSizes()

	// now if w query size of any mig whole cache should be refreshed by listing InstanceGroupManagers; we expect two calls
	// for zoneB and zoneC
	server.On("handle", "/project1/zones/us-central1-b/instanceGroupManagers").Return(
		buildListInstanceGroupManagersResponse(
			buildListInstanceGroupManagersResponsePart(defaultPoolMigName, zoneB, 7),
			buildListInstanceGroupManagersResponsePart(extraPoolMigName, zoneB, 8),
		)).Once()
	server.On("handle", "/project1/zones/us-central1-c/instanceGroupManagers").Return(
		buildListInstanceGroupManagersResponse(
			buildListInstanceGroupManagersResponsePart(extraPool2MigName, zoneC, 9),
		)).Once()

	extraPool2MigSize, err = g.GetMigSize(extraPool2Mig)
	assert.NoError(t, err)
	assert.Equal(t, int64(9), extraPool2MigSize)
	mock.AssertExpectationsForObjects(t, server)
}

func TestGetMigSizeListCallFails(t *testing.T) {
	server := NewHttpServerMock()
	defer server.Close()
	g := newTestGceManager(t, server.URL, false)

	extraPoolMig := setupTestExtraPool(g, true)
	defaultPoolMig := setupTestDefaultPool(g, true)

	server.On("handle", "/project1/zones/us-central1-b/instanceGroupManagers").Return("bad_response").Once()
	server.On("handle", fmt.Sprintf("/project1/zones/us-central1-b/instanceGroupManagers/%s", defaultPoolMigName)).Return(buildInstanceGroupManagerResponse(zoneB, defaultPoolMigName, 7)).Once()

	// getting size for defaultPoolMig should trigger listing all the InstanceGroupManagers
	defaultPoolMigSize, err := g.GetMigSize(defaultPoolMig)
	assert.NoError(t, err)
	assert.Equal(t, int64(7), defaultPoolMigSize)
	mock.AssertExpectationsForObjects(t, server)

	// extra queries for defaultPoolMig and extraPoolMig should not result in any extra API calls
	defaultPoolMigSize, err = g.GetMigSize(defaultPoolMig)
	assert.NoError(t, err)
	assert.Equal(t, int64(7), defaultPoolMigSize)

	// Querying another mig will yet again try to list all migs
	server.On("handle", "/project1/zones/us-central1-b/instanceGroupManagers").Return(
		buildListInstanceGroupManagersResponse(
			buildListInstanceGroupManagersResponsePart(defaultPoolMigName, zoneB, 7),
			buildListInstanceGroupManagersResponsePart(extraPoolMigName, zoneB, 8),
		)).Once()

	// getting size for defaultPoolMig should trigger listing all the InstanceGroupManagers
	extraPoolMigSize, err := g.GetMigSize(extraPoolMig)
	assert.NoError(t, err)
	assert.Equal(t, int64(8), extraPoolMigSize)
	mock.AssertExpectationsForObjects(t, server)
}

func TestGetMigForInstance(t *testing.T) {
	server := NewHttpServerMock()
	defer server.Close()
	g := newTestGceManager(t, server.URL, false)

	setupTestDefaultPool(g, false)
	g.cache.InvalidateAllMigBasenames()

	server.On("handle", "/project1/zones/us-central1-b/instanceGroupManagers/gke-cluster-1-default-pool").Return(buildDefaultInstanceGroupManagerResponse(zoneB)).Once()
	server.On("handle", "/project1/zones/us-central1-b/instanceGroupManagers/gke-cluster-1-default-pool/listManagedInstances").Return(buildFourRunningInstancesOnDefaultMigManagedInstancesResponse(zoneB)).Twice()
	gceRef1 := GceRef{
		Project: projectId,
		Zone:    zoneB,
		Name:    "gke-cluster-1-default-pool-f7607aac-f1hm",
	}

	mig, err := g.GetMigForInstance(gceRef1)
	assert.NoError(t, err)
	assert.NotNil(t, mig)
	assert.Equal(t, "gke-cluster-1-default-pool", mig.GceRef().Name)

	gceRef2 := GceRef{
		Project: projectId,
		Zone:    zoneB,
		Name:    "gke-cluster-1-default-pool-f7607aac-0000", // instance from unknown MIG
	}
	mig, err = g.GetMigForInstance(gceRef2)
	assert.NoError(t, err)
	assert.Nil(t, mig)
	_, found := g.cache.instancesFromUnknownMigs[gceRef2]
	assert.True(t, found)

	mock.AssertExpectationsForObjects(t, server)
}

func TestGetMigNodesBasic(t *testing.T) {
	server := NewHttpServerMock()
	defer server.Close()
	g := newTestGceManager(t, server.URL, false)

	server.On("handle", "/project1/zones/us-central1-b/instanceGroupManagers/extra-pool-323233232/listManagedInstances").Return(buildFourRunningInstancesOnDefaultMigManagedInstancesResponse(zoneB)).Once()

	mig := &gceMig{
		gceRef: GceRef{
			Project: projectId,
			Zone:    zoneB,
			Name:    "extra-pool-323233232",
		},
		gceManager: g,
		minSize:    0,
		maxSize:    1000,
	}

	nodes, err := g.GetMigNodes(mig)
	assert.NoError(t, err)
	assert.Equal(t, 4, len(nodes))
	assert.Equal(t, "gce://project1/us-central1-b/gke-cluster-1-default-pool-f7607aac-9j4g", nodes[0].Id)
	assert.Equal(t, "gce://project1/us-central1-b/gke-cluster-1-default-pool-f7607aac-c63g", nodes[1].Id)
	assert.Equal(t, "gce://project1/us-central1-b/gke-cluster-1-default-pool-f7607aac-dck1", nodes[2].Id)
	assert.Equal(t, "gce://project1/us-central1-b/gke-cluster-1-default-pool-f7607aac-f1hm", nodes[3].Id)

	for i := 0; i < 4; i++ {
		assert.Nil(t, nodes[i].Status.ErrorInfo)
		assert.Equal(t, cloudprovider.InstanceRunning, nodes[i].Status.State)

	}
}

const managedInstancesResponseTemplate = `{"managedInstances": [%s]}`

const managedInstanceWithInstanceStatusResponsePartTemplate = `{
   "instance": "https://www.googleapis.com/compute/v1/projects/project1/zones/%s/instances/%s",
   "id": "1776565833558018907",
   "instanceStatus": "%s",
   "version": {
    "instanceTemplate": "https://www.googleapis.com/compute/beta/projects/project1/global/instanceTemplates/test-1-cpu-1-k80-2"
   },
   "currentAction": "NONE"
  }
`

const managedInstanceWithInstanceStatusAndCurrentActionResponsePartTemplate = `{
   "instance": "https://www.googleapis.com/compute/v1/projects/project1/zones/%s/instances/%s",
   "instanceStatus": "%s",
   "version": {
    "instanceTemplate": "https://www.googleapis.com/compute/beta/projects/project1/global/instanceTemplates/test-1-cpu-1-k80-2"
   },
   "currentAction": "%s",
   "lastAttempt": {}
   }
`

const managedInstanceWithInstanceStatusAndWithCurrentActionAndErrorResponsePartTemplate = `{
   "instance": "https://www.googleapis.com/compute/v1/projects/project1/zones/%s/instances/%s",
   "instanceStatus": "%s",
   "version": {
    "instanceTemplate": "https://www.googleapis.com/compute/beta/projects/project1/global/instanceTemplates/test-1-cpu-1-k80-2"
   },
   "currentAction": "%s",
   "lastAttempt": {
    "errors": {
     "errors": [
      {
       "code": "%s",
       "message": "%s"
      }
     ]
    }
   }
  }
`

const managedInstanceWithCurrentActionResponsePartTemplate = `{
   "instance": "https://www.googleapis.com/compute/v1/projects/project1/zones/%s/instances/%s",
   "version": {
    "instanceTemplate": "https://www.googleapis.com/compute/beta/projects/project1/global/instanceTemplates/test-1-cpu-1-k80-2"
   },
   "currentAction": "%s",
   "lastAttempt": {}
   }
`
const managedInstanceWithCurrentActionAndErrorResponsePartTemplate = `{
   "instance": "https://www.googleapis.com/compute/v1/projects/project1/zones/%s/instances/%s",
   "version": {
    "instanceTemplate": "https://www.googleapis.com/compute/beta/projects/project1/global/instanceTemplates/test-1-cpu-1-k80-2"
   },
   "currentAction": "%s",
   "lastAttempt": {
    "errors": {
     "errors": [
      {
       "code": "%s",
       "message": "%s"
      }
     ]
    }
   }
  }
`

const managedInstanceWithCurrentActionAndTwoErrorsResponsePartTemplate = `{
   "instance": "https://www.googleapis.com/compute/v1/projects/project1/zones/%s/instances/%s",
   "version": {
    "instanceTemplate": "https://www.googleapis.com/compute/beta/projects/project1/global/instanceTemplates/test-1-cpu-1-k80-2"
   },
   "currentAction": "%s",
   "lastAttempt": {
    "errors": {
     "errors": [
      {
       "code": "%s",
       "message": "%s"
      },
      {
       "code": "%s",
       "message": "%s"
      }
     ]
    }
   }
  }
`

func buildManagedInstancesResponse(managedInstanceParts ...string) string {
	partsString := ""
	for _, part := range managedInstanceParts {
		if partsString != "" {
			partsString += ", "
		}
		partsString += part
	}
	return fmt.Sprintf(managedInstancesResponseTemplate, partsString)
}

func buildRunningManagedInstanceResponsePart(zone string, instanceName string) string {
	return buildManagedInstanceWithInstanceStatusResponsePart(zone, instanceName, "RUNNING")
}

func buildRunningManagedInstanceWithCurrentActionResponsePart(zone string, instanceName string, currentAction string) string {
	return buildManagedInstanceWithInstanceStatusAndCurrentActionResponsePart(zone, instanceName, "RUNNING", currentAction)
}

func buildRunningManagedInstanceWithCurrentActionAndErrorResponsePart(zone string, instanceName string, currentAction string, code string, message string) string {
	return buildManagedInstanceWithInstanceStatusAndCurrentActionAndErrorResponsePart(zone, instanceName, "RUNNING", currentAction, code, message)
}

func buildManagedInstanceWithInstanceStatusResponsePart(zone string, instanceName string, instanceStatus string) string {
	return fmt.Sprintf(managedInstanceWithInstanceStatusResponsePartTemplate, zone, instanceName, instanceStatus)
}

func buildManagedInstanceWithInstanceStatusAndCurrentActionResponsePart(zone string, instanceName string, instanceStatus string, currentAction string) string {
	return fmt.Sprintf(managedInstanceWithInstanceStatusAndCurrentActionResponsePartTemplate, zone, instanceName, instanceStatus, currentAction)
}

func buildManagedInstanceWithInstanceStatusAndCurrentActionAndErrorResponsePart(zone string, instanceName string, instanceStatus string, currentAction string, code string, message string) string {
	return fmt.Sprintf(managedInstanceWithInstanceStatusAndWithCurrentActionAndErrorResponsePartTemplate, zone, instanceName, instanceStatus, currentAction, code, message)
}

func buildManagedInstanceWithCurrentActionResponsePart(zone string, instanceName string, currentAction string) string {
	return fmt.Sprintf(managedInstanceWithCurrentActionResponsePartTemplate, zone, instanceName, currentAction)
}

func buildManagedInstanceWithCurrentActionAndErrorResponsePart(zone string, instanceName string, currentAction string, code string, message string) string {
	return fmt.Sprintf(managedInstanceWithCurrentActionAndErrorResponsePartTemplate, zone, instanceName, currentAction, code, message)
}

func buildManagedInstanceWithCurrentActionAndTwoErrorsResponsePart(zone string, instanceName string, currentAction string, code1 string, message1 string, code2 string, message2 string) string {
	return fmt.Sprintf(managedInstanceWithCurrentActionAndTwoErrorsResponsePartTemplate, zone, instanceName, currentAction, code1, message1, code2, message2)
}

func TestGetMigNodesComplex(t *testing.T) {
	server := NewHttpServerMock()
	defer server.Close()
	g := newTestGceManager(t, server.URL, false)

	testCases := []struct {
		instanceName         string
		responsePart         string
		expectedState        cloudprovider.InstanceState
		expectedErrorClass   cloudprovider.InstanceErrorClass
		expectedErrorCode    string
		expectedErrorMessage string
	}{
		{
			"running-creating-no_error",
			buildRunningManagedInstanceResponsePart("europe-west1-b", "running-creating-no_error"),
			cloudprovider.InstanceRunning,
			0,
			"",
			"",
		},
		{
			"none-creating-quota_exceeded",
			buildManagedInstanceWithCurrentActionAndErrorResponsePart("europe-west1-b", "none-creating-quota_exceeded", "CREATING", ErrorCodeQuotaExceeded, "We run out of quota while creating!"),
			cloudprovider.InstanceCreating,
			cloudprovider.OutOfResourcesErrorClass,
			ErrorCodeQuotaExceeded,
			"We run out of quota while creating!",
		},
		{
			"none-recreating-quota_exceeded",
			buildManagedInstanceWithCurrentActionAndErrorResponsePart("europe-west1-b", "none-recreating-quota_exceeded", "RECREATING", ErrorCodeQuotaExceeded, "We run out of quota while recreating!"),
			cloudprovider.InstanceCreating,
			cloudprovider.OutOfResourcesErrorClass,
			ErrorCodeQuotaExceeded,
			"We run out of quota while recreating!",
		},
		{
			"none-creating_no_retries-quota_exceeded",
			buildManagedInstanceWithCurrentActionAndErrorResponsePart("europe-west1-b", "none-creating_no_retries-quota_exceeded", "CREATING_WITHOUT_RETRIES", ErrorCodeQuotaExceeded, "We run out of quota while creating without retries!"),
			cloudprovider.InstanceCreating,
			cloudprovider.OutOfResourcesErrorClass,
			ErrorCodeQuotaExceeded,
			"We run out of quota while creating without retries!",
		},
		{
			"none-creating-other_error",
			buildManagedInstanceWithCurrentActionAndErrorResponsePart("europe-west1-b", "none-creating-other_error", "CREATING", "SOME_ERROR", "Ojojojoj!"),
			cloudprovider.InstanceCreating,
			cloudprovider.OtherErrorClass,
			ErrorCodeOther,
			"Ojojojoj!",
		},
		{
			"none-creating-other_error_and_quota_exceeded",
			buildManagedInstanceWithCurrentActionAndTwoErrorsResponsePart("europe-west1-b", "none-creating-other_error_and_quota_exceeded", "CREATING", "SOME_ERROR", "Ojojojoj!", ErrorCodeQuotaExceeded, "We run out of quota!"),
			cloudprovider.InstanceCreating,
			cloudprovider.OutOfResourcesErrorClass,
			ErrorCodeQuotaExceeded,
			"Ojojojoj!; We run out of quota!",
		},
		{
			"none-creating-quota_exceeded_and_other_error",
			buildManagedInstanceWithCurrentActionAndTwoErrorsResponsePart("europe-west1-b", "none-creating-quota_exceeded_and_other_error", "CREATING", ErrorCodeQuotaExceeded, "We run out of quota!", "SOME_ERROR", "Ojojojoj!"),
			cloudprovider.InstanceCreating,
			cloudprovider.OutOfResourcesErrorClass,
			ErrorCodeQuotaExceeded,
			"We run out of quota!; Ojojojoj!",
		},
		{
			"none-deleting-no_error",
			buildManagedInstanceWithCurrentActionResponsePart("europe-west1-b", "none-deleting-no_error", "DELETING"),
			cloudprovider.InstanceDeleting,
			0,
			"",
			"",
		},
		{
			"running-deleting-no_error",
			buildRunningManagedInstanceWithCurrentActionResponsePart("europe-west1-b", "running-deleting-no_error", "DELETING"),
			cloudprovider.InstanceDeleting,
			0,
			"",
			"",
		},
		{
			"running-deleting-other_error",
			buildRunningManagedInstanceWithCurrentActionAndErrorResponsePart("europe-west1-b", "running-deleting-other_error", "DELETING", "SOME_ERROR", "Error while deleting"),
			cloudprovider.InstanceDeleting,
			0,
			"",
			"",
		},
		{
			"none-creating-resource_pool_exhausted_error",
			buildManagedInstanceWithCurrentActionAndErrorResponsePart("europe-west1-b", "none-creating-resource_pool_exhausted_error", "CREATING", "RESOURCE_POOL_EXHAUSTED", "No resources!"),
			cloudprovider.InstanceCreating,
			cloudprovider.OutOfResourcesErrorClass,
			ErrorCodeResourcePoolExhausted,
			"No resources!",
		},
		{
			"none-creating-zonal_resource_pool_exhausted_error",
			buildManagedInstanceWithCurrentActionAndErrorResponsePart("europe-west1-b", "none-creating-zonal_resource_pool_exhausted_error", "CREATING", "ZONE_RESOURCE_POOL_EXHAUSTED", "No resources!"),
			cloudprovider.InstanceCreating,
			cloudprovider.OutOfResourcesErrorClass,
			ErrorCodeResourcePoolExhausted,
			"No resources!",
		},
		{
			"none-creating-zonal_resource_pool_exhausted_error_with_details",
			buildManagedInstanceWithCurrentActionAndErrorResponsePart("europe-west1-b", "none-creating-zonal_resource_pool_exhausted_error_with_details", "CREATING", "ZONE_RESOURCE_POOL_EXHAUSTED_WITH_DETAILS", "No resources!"),
			cloudprovider.InstanceCreating,
			cloudprovider.OutOfResourcesErrorClass,
			ErrorCodeResourcePoolExhausted,
			"No resources!",
		},
		{
			"running-creating-resource_pool_exhausted_error",
			buildRunningManagedInstanceWithCurrentActionAndErrorResponsePart("europe-west1-b", "running-creating-resource_pool_exhausted_error", "CREATING", "RESOURCE_POOL_EXHAUSTED", "No resources!"),
			cloudprovider.InstanceCreating,
			cloudprovider.OutOfResourcesErrorClass,
			ErrorCodeResourcePoolExhausted,
			"No resources!",
		},
		{
			"running-creating-quota_error",
			buildRunningManagedInstanceWithCurrentActionAndErrorResponsePart("europe-west1-b", "running-creating-quota_error", "CREATING", ErrorCodeQuotaExceeded, "We run out of quota while creating!"),
			cloudprovider.InstanceCreating,
			cloudprovider.OutOfResourcesErrorClass,
			ErrorCodeQuotaExceeded,
			"We run out of quota while creating!",
		},
		{
			"running-creating-other_error",
			buildRunningManagedInstanceWithCurrentActionAndErrorResponsePart("europe-west1-b", "running-creating-other_error", "CREATING", "SOME_ERROR", "Ojojojoj!"),
			cloudprovider.InstanceCreating,
			0,
			"",
			"",
		},
		{
			"repairing-creating-other_error",
			buildManagedInstanceWithInstanceStatusAndCurrentActionAndErrorResponsePart("europe-west1-b", "repairing-creating-other_error", "REPAIRING", "CREATING", "SOME_ERROR", "Ojojojoj!"),
			cloudprovider.InstanceCreating,
			0,
			"",
			"",
		},
		{
			"provisioning-creating-other_error",
			buildManagedInstanceWithInstanceStatusAndCurrentActionAndErrorResponsePart("europe-west1-b", "provisioning-creating-other_error", "PROVISIONING", "CREATING", "SOME_ERROR", "Ojojojoj!"),
			cloudprovider.InstanceCreating,
			cloudprovider.OtherErrorClass,
			ErrorCodeOther,
			"Ojojojoj!",
		},
		{
			"staging-creating-other_error",
			buildManagedInstanceWithInstanceStatusAndCurrentActionAndErrorResponsePart("europe-west1-b", "staging-creating-other_error", "STAGING", "CREATING", "SOME_ERROR", "Ojojojoj!"),
			cloudprovider.InstanceCreating,
			cloudprovider.OtherErrorClass,
			ErrorCodeOther,
			"Ojojojoj!",
		},
	}

	parts := make([]string, 0)
	for _, tc := range testCases {
		parts = append(parts, tc.responsePart)
	}
	response := buildManagedInstancesResponse(parts...)
	server.On("handle", "/project1/zones/europe-west1-b/instanceGroupManagers/some_group/listManagedInstances").Return(response).Once()

	mig := &gceMig{
		gceRef: GceRef{
			Project: projectId,
			Zone:    "europe-west1-b",
			Name:    "some_group",
		},
		gceManager: g,
		minSize:    0,
		maxSize:    1000,
	}
	nodes, err := g.GetMigNodes(mig)

	assert.NoError(t, err)
	assert.Equal(t, len(testCases), len(nodes))

	for i, tc := range testCases {
		instanceInfo := nodes[i]
		assert.Equal(t, fmt.Sprintf("gce://project1/europe-west1-b/%s", tc.instanceName), instanceInfo.Id)
		assert.Equal(t, tc.expectedState, instanceInfo.Status.State)
		if tc.expectedErrorClass == 0 {
			assert.Nil(t, instanceInfo.Status.ErrorInfo)
		} else {
			assert.NotNil(t, instanceInfo.Status.ErrorInfo)
			assert.Equal(t, tc.expectedErrorClass, instanceInfo.Status.ErrorInfo.ErrorClass)
			assert.Equal(t, tc.expectedErrorCode, instanceInfo.Status.ErrorInfo.ErrorCode)
			assert.Equal(t, tc.expectedErrorMessage, instanceInfo.Status.ErrorInfo.ErrorMessage)
		}
	}

	mock.AssertExpectationsForObjects(t, server)
}

const instanceGroupResponsePartTemplate = `{
  "kind": "compute#instanceGroup",
  "id": "1121230570947910218",
  "name": "%s",
  "selfLink": "https://www.googleapis.com/compute/v1/projects/project1/zones/%s/instanceGroups/%s",
  "size": 1
}`

func buildInstanceGroupResponsePart(zone string, instanceGroup string) string {
	return fmt.Sprintf(instanceGroupResponsePartTemplate, instanceGroup, zone, instanceGroup)
}

const listInstanceGroupsResponseTemplate = `{
  "kind": "compute#instanceGroupList",
  "id": "projects/project1a/zones/%s/instanceGroups",
  "items": [%s],
  "selfLink": "https://www.googleapis.com/compute/v1/projects/project1/zones/%s/instanceGroups"
}`

func buildListInstanceGroupsResponse(zone string, instanceGroups ...string) string {

	var items []string
	for _, instanceGroup := range instanceGroups {
		items = append(items, buildInstanceGroupResponsePart(zone, instanceGroup))
	}

	return fmt.Sprintf(listInstanceGroupsResponseTemplate,
		zone,
		strings.Join(items, ", "),
		zone,
	)
}

const getRegionResponse = `{
 "kind": "compute#region",
 "id": "1000",
 "creationTimestamp": "1969-12-31T16:00:00.000-08:00",
 "name": "us-central1",
 "description": "us-central1",
 "status": "UP",
 "zones": [
  "https://www.googleapis.com/compute/v1/projects/project1/zones/us-central1-b"
 ],
 "quotas": [],
 "selfLink": "https://www.googleapis.com/compute/v1/projects/project1/regions/us-central1"
}`

func TestFetchAutoMigsZonal(t *testing.T) {
	server := NewHttpServerMock()
	defer server.Close()

	server.On("handle", "/project1/zones/"+zoneB+"/instanceGroups").Return(buildListInstanceGroupsResponse(zoneB, gceMigA, gceMigB)).Once()
	server.On("handle", "/project1/zones/"+zoneB+"/instanceGroupManagers/"+gceMigA).Return(buildInstanceGroupManagerResponse(zoneB, gceMigA, 3)).Once()
	server.On("handle", "/project1/zones/"+zoneB+"/instanceGroupManagers/"+gceMigB).Return(buildInstanceGroupManagerResponse(zoneB, gceMigB, 3)).Once()

	// Regenerate instance cache
	server.On("handle", "/project1/global/instanceTemplates/"+gceMigA).Return(instanceTemplate).Once()
	server.On("handle", "/project1/global/instanceTemplates/"+gceMigB).Return(instanceTemplate).Once()
	server.On("handle", "/project1/zones/"+zoneB+"/instanceGroupManagers/"+gceMigA+"/listManagedInstances").Return(buildFourRunningInstancesManagedInstancesResponse(zoneB, gceMigA)).Once()
	server.On("handle", "/project1/zones/"+zoneB+"/instanceGroupManagers/"+gceMigB+"/listManagedInstances").Return(buildOneRunningInstanceManagedInstancesResponse(zoneB, gceMigB)).Once()

	regional := false
	g := newTestGceManager(t, server.URL, regional)

	min, max := 0, 100
	g.migAutoDiscoverySpecs = []migAutoDiscoveryConfig{
		{Re: regexp.MustCompile("UNUSED"), MinSize: min, MaxSize: max},
	}

	assert.NoError(t, g.fetchAutoMigs())

	migs := g.GetMigs()
	assert.Equal(t, 2, len(migs))
	validateMigExists(t, migs, zoneB, gceMigA, min, max)
	validateMigExists(t, migs, zoneB, gceMigB, min, max)
	mock.AssertExpectationsForObjects(t, server)
}

func TestFetchAutoMigsUnregistersMissingMigs(t *testing.T) {
	server := NewHttpServerMock()
	defer server.Close()

	// Register explicit instance group
	server.On("handle", "/project1/zones/"+zoneB+"/instanceGroupManagers/"+gceMigA).Return(buildInstanceGroupManagerResponse(zoneB, gceMigA, 3)).Once()
	server.On("handle", "/project1/global/instanceTemplates/"+gceMigA).Return(instanceTemplate).Once()

	// Regenerate cache for explicit instance group
	server.On("handle", "/project1/zones/"+zoneB+"/instanceGroupManagers/"+gceMigA+"/listManagedInstances").Return(buildFourRunningInstancesManagedInstancesResponse(zoneB, gceMigA)).Twice()

	// Register 'previously autodetected' instance group
	server.On("handle", "/project1/zones/"+zoneB+"/instanceGroupManagers/"+gceMigB).Return(buildInstanceGroupManagerResponse(zoneB, gceMigB, 3)).Once()
	server.On("handle", "/project1/global/instanceTemplates/"+gceMigB).Return(instanceTemplate).Once()

	regional := false
	g := newTestGceManager(t, server.URL, regional)

	// This MIG should never be unregistered because it is explicitly configured.
	minA, maxA := 0, 100
	specs := []string{fmt.Sprintf("%d:%d:https://content.googleapis.com/compute/v1/projects/project1/zones/%s/instanceGroups/%s", minA, maxA, zoneB, gceMigA)}
	assert.NoError(t, g.fetchExplicitMigs(specs))

	// This MIG was previously autodetected but is now gone.
	// It should be unregistered.
	unregister := &gceMig{
		gceManager: g,
		gceRef:     GceRef{Project: projectId, Zone: zoneB, Name: gceMigB},
		minSize:    1,
		maxSize:    10,
	}
	assert.True(t, g.registerMig(unregister))

	assert.NoError(t, g.fetchAutoMigs())

	migs := g.GetMigs()
	assert.Equal(t, 1, len(migs))
	validateMigExists(t, migs, zoneB, gceMigA, minA, maxA)
	mock.AssertExpectationsForObjects(t, server)
}

func TestFetchAutoMigsRegional(t *testing.T) {
	server := NewHttpServerMock()
	defer server.Close()

	server.On("handle", "/project1/regions/us-central1").Return(getRegionResponse).Once()
	server.On("handle", "/project1/zones/"+zoneB+"/instanceGroups").Return(buildListInstanceGroupsResponse(zoneB, gceMigA, gceMigB)).Once()
	server.On("handle", "/project1/zones/"+zoneB+"/instanceGroupManagers/"+gceMigA).Return(buildInstanceGroupManagerResponse(zoneB, gceMigA, 3)).Once()
	server.On("handle", "/project1/zones/"+zoneB+"/instanceGroupManagers/"+gceMigB).Return(buildInstanceGroupManagerResponse(zoneB, gceMigB, 3)).Once()

	// Regenerate instance cache
	server.On("handle", "/project1/global/instanceTemplates/"+gceMigA).Return(instanceTemplate).Once()
	server.On("handle", "/project1/global/instanceTemplates/"+gceMigB).Return(instanceTemplate).Once()
	server.On("handle", "/project1/zones/"+zoneB+"/instanceGroupManagers/"+gceMigA+"/listManagedInstances").Return(buildFourRunningInstancesManagedInstancesResponse(zoneB, gceMigA)).Once()
	server.On("handle", "/project1/zones/"+zoneB+"/instanceGroupManagers/"+gceMigB+"/listManagedInstances").Return(buildOneRunningInstanceManagedInstancesResponse(zoneB, gceMigB)).Once()

	regional := true
	g := newTestGceManager(t, server.URL, regional)

	min, max := 0, 100
	g.migAutoDiscoverySpecs = []migAutoDiscoveryConfig{
		{Re: regexp.MustCompile("UNUSED"), MinSize: min, MaxSize: max},
	}

	assert.NoError(t, g.fetchAutoMigs())

	migs := g.GetMigs()
	assert.Equal(t, 2, len(migs))
	validateMigExists(t, migs, zoneB, gceMigA, min, max)
	validateMigExists(t, migs, zoneB, gceMigB, min, max)
	mock.AssertExpectationsForObjects(t, server)
}

func TestFetchExplicitMigs(t *testing.T) {
	server := NewHttpServerMock()
	defer server.Close()

	server.On("handle", "/project1/zones/"+zoneB+"/instanceGroupManagers/"+gceMigA).Return(buildInstanceGroupManagerResponse(zoneB, gceMigA, 3)).Once()
	server.On("handle", "/project1/zones/"+zoneB+"/instanceGroupManagers/"+gceMigB).Return(buildInstanceGroupManagerResponse(zoneB, gceMigB, 3)).Once()

	server.On("handle", "/project1/global/instanceTemplates/"+gceMigA).Return(instanceTemplate).Once()
	server.On("handle", "/project1/global/instanceTemplates/"+gceMigB).Return(instanceTemplate).Once()

	server.On("handle", "/project1/zones/"+zoneB+"/instanceGroupManagers/"+gceMigA+"/listManagedInstances").Return(buildFourRunningInstancesManagedInstancesResponse(zoneB, gceMigA)).Once()
	server.On("handle", "/project1/zones/"+zoneB+"/instanceGroupManagers/"+gceMigB+"/listManagedInstances").Return(buildOneRunningInstanceManagedInstancesResponse(zoneB, gceMigB)).Once()

	regional := false
	g := newTestGceManager(t, server.URL, regional)

	minA, maxA := 0, 100
	minB, maxB := 1, 10
	specs := []string{
		fmt.Sprintf("%d:%d:https://content.googleapis.com/compute/v1/projects/project1/zones/%s/instanceGroups/%s", minA, maxA, zoneB, gceMigA),
		fmt.Sprintf("%d:%d:https://content.googleapis.com/compute/v1/projects/project1/zones/%s/instanceGroups/%s", minB, maxB, zoneB, gceMigB),
	}

	assert.NoError(t, g.fetchExplicitMigs(specs))

	migs := g.GetMigs()
	assert.Equal(t, 2, len(migs))
	validateMigExists(t, migs, zoneB, gceMigA, minA, maxA)
	validateMigExists(t, migs, zoneB, gceMigB, minB, maxB)
	mock.AssertExpectationsForObjects(t, server)
}

const listMachineTypesResponse = `{
 "kind": "compute#machineTypeList",
 "id": "projects/project1/zones/us-central1-c/machineTypes",
 "items": [
  {
   "kind": "compute#machineType",
   "id": "1000",
   "creationTimestamp": "1969-12-31T16:00:00.000-08:00",
   "name": "f1-micro",
   "description": "1 vCPU (shared physical core) and 0.6 GB RAM",
   "guestCpus": 1,
   "memoryMb": 614,
   "imageSpaceGb": 0,
   "maximumPersistentDisks": 16,
   "maximumPersistentDisksSizeGb": "3072",
   "zone": "us-central1-c",
   "selfLink": "https://www.googleapis.com/compute/v1/projects/project1/zones/us-central1-c/machineTypes/f1-micro",
   "isSharedCpu": true
  },
  {
   "kind": "compute#machineType",
   "id": "2000",
   "creationTimestamp": "1969-12-31T16:00:00.000-08:00",
   "name": "g1-small",
   "description": "1 vCPU (shared physical core) and 1.7 GB RAM",
   "guestCpus": 1,
   "memoryMb": 1740,
   "imageSpaceGb": 0,
   "maximumPersistentDisks": 16,
   "maximumPersistentDisksSizeGb": "3072",
   "zone": "us-central1-c",
   "selfLink": "https://www.googleapis.com/compute/v1/projects/project1/zones/us-central1-c/machineTypes/g1-small",
   "isSharedCpu": true
  }
 ],
 "selfLink": "https://www.googleapis.com/compute/v1/projects/project1/zones/us-central1-c/machineTypes"
}`

func TestGetMigTemplateNode(t *testing.T) {
	server := NewHttpServerMock()
	defer server.Close()

	server.On("handle", "/project1/zones/us-central1-b/instanceGroupManagers/default-pool").Return(getInstanceGroupManagerResponse).Once()
	server.On("handle", "/project1/global/instanceTemplates/gke-cluster-1-default-pool").Return(instanceTemplate).Once()

	regional := false
	g := newTestGceManager(t, server.URL, regional)

	mig := &gceMig{
		gceRef: GceRef{
			Project: projectId,
			Zone:    zoneB,
			Name:    "default-pool",
		},
		gceManager: g,
		minSize:    0,
		maxSize:    1000,
	}

	node, err := g.GetMigTemplateNode(mig)
	assert.NoError(t, err)
	assert.NotNil(t, node)
	mock.AssertExpectationsForObjects(t, server)
}

const getMachineTypeResponse = `{
  "kind": "compute#machineType",
  "id": "3001",
  "creationTimestamp": "2015-01-16T09:25:43.314-08:00",
  "name": "n1-standard-2",
  "description": "2 vCPU, 3.75 GB RAM",
  "guestCpus": 2,
  "memoryMb": 3840,
  "maximumPersistentDisks": 32,
  "maximumPersistentDisksSizeGb": "65536",
  "zone": "us-central1-a",
  "selfLink": "https://www.googleapis.com/compute/v1/projects/krzysztof-jastrzebski-dev/zones/us-central1-a/machineTypes/n1-standard-1",
  "isSharedCpu": false
}`

func TestGetCpuAndMemoryForMachineType(t *testing.T) {
	server := NewHttpServerMock()
	defer server.Close()
	regional := false
	g := newTestGceManager(t, server.URL, regional)

	// Custom machine type.
	cpu, mem, err := g.getCpuAndMemoryForMachineType("custom-8-2", zoneB)
	assert.NoError(t, err)
	assert.Equal(t, int64(8), cpu)
	assert.Equal(t, int64(2*units.MiB), mem)
	mock.AssertExpectationsForObjects(t, server)

	// Standard machine type found in cache.
	cpu, mem, err = g.getCpuAndMemoryForMachineType("n1-standard-1", zoneB)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), cpu)
	assert.Equal(t, int64(1*units.MiB), mem)
	mock.AssertExpectationsForObjects(t, server)

	// Standard machine type not found in cache.
	server.On("handle", "/project1/zones/"+zoneB+"/machineTypes/n1-standard-2").Return(getMachineTypeResponse).Once()
	cpu, mem, err = g.getCpuAndMemoryForMachineType("n1-standard-2", zoneB)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), cpu)
	assert.Equal(t, int64(3840*units.MiB), mem)
	mock.AssertExpectationsForObjects(t, server)

	// Standard machine type cached.
	cpu, mem, err = g.getCpuAndMemoryForMachineType("n1-standard-2", zoneB)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), cpu)
	assert.Equal(t, int64(3840*units.MiB), mem)
	mock.AssertExpectationsForObjects(t, server)

	// Standard machine type not found in the zone.
	server.On("handle", "/project1/zones/us-central1-g/machineTypes/n1-standard-1").Return("").Once()
	_, _, err = g.getCpuAndMemoryForMachineType("n1-standard-1", "us-central1-g")
	assert.Error(t, err)
	mock.AssertExpectationsForObjects(t, server)

}

func TestParseCustomMachineType(t *testing.T) {
	cpu, mem, err := parseCustomMachineType("custom-2-2816")
	assert.NoError(t, err)
	assert.Equal(t, int64(2), cpu)
	assert.Equal(t, int64(2816*units.MiB), mem)
	_, _, err = parseCustomMachineType("other-a2-2816")
	assert.Error(t, err)
	_, _, err = parseCustomMachineType("other-2-2816")
	assert.Error(t, err)
}

func validateMigExists(t *testing.T, migs []Mig, zone string, name string, minSize int, maxSize int) {
	ref := GceRef{
		projectId,
		zone,
		name,
	}
	for _, mig := range migs {
		if mig.GceRef() == ref {
			assert.Equal(t, minSize, mig.MinSize())
			assert.Equal(t, maxSize, mig.MaxSize())
			return
		}
	}
	allRefs := []GceRef{}
	for _, mig := range migs {
		allRefs = append(allRefs, mig.GceRef())
	}
	assert.Failf(t, "Mig not found", "Mig %v not found among %v", ref, allRefs)
}

func TestParseMIGAutoDiscoverySpecs(t *testing.T) {
	cases := []struct {
		name    string
		specs   []string
		want    []migAutoDiscoveryConfig
		wantErr bool
	}{
		{
			name: "GoodSpecs",
			specs: []string{
				"mig:namePrefix=pfx,min=0,max=10",
				"mig:namePrefix=anotherpfx,min=1,max=2",
			},
			want: []migAutoDiscoveryConfig{
				{Re: regexp.MustCompile("^pfx.+"), MinSize: 0, MaxSize: 10},
				{Re: regexp.MustCompile("^anotherpfx.+"), MinSize: 1, MaxSize: 2},
			},
		},
		{
			name:    "MissingMIGType",
			specs:   []string{"namePrefix=pfx,min=0,max=10"},
			wantErr: true,
		},
		{
			name:    "WrongType",
			specs:   []string{"asg:namePrefix=pfx,min=0,max=10"},
			wantErr: true,
		},
		{
			name:    "UnknownKey",
			specs:   []string{"mig:namePrefix=pfx,min=0,max=10,unknown=hi"},
			wantErr: true,
		},
		{
			name:    "NonIntegerMin",
			specs:   []string{"mig:namePrefix=pfx,min=a,max=10"},
			wantErr: true,
		},
		{
			name:    "NonIntegerMax",
			specs:   []string{"mig:namePrefix=pfx,min=1,max=donkey"},
			wantErr: true,
		},
		{
			name:    "PrefixDoesNotCompileToRegexp",
			specs:   []string{"mig:namePrefix=a),min=1,max=10"},
			wantErr: true,
		},
		{
			name:    "KeyMissingValue",
			specs:   []string{"mig:namePrefix=prefix,min=,max=10"},
			wantErr: true,
		},
		{
			name:    "ValueMissingKey",
			specs:   []string{"mig:namePrefix=prefix,=0,max=10"},
			wantErr: true,
		},
		{
			name:    "KeyMissingSeparator",
			specs:   []string{"mig:namePrefix=prefix,min,max=10"},
			wantErr: true,
		},
		{
			name:    "TooManySeparators",
			specs:   []string{"mig:namePrefix=prefix,min=0,max=10=20"},
			wantErr: true,
		},
		{
			name:    "PrefixIsEmpty",
			specs:   []string{"mig:namePrefix=,min=0,max=10"},
			wantErr: true,
		},
		{
			name:    "PrefixIsMissing",
			specs:   []string{"mig:min=0,max=10"},
			wantErr: true,
		},
		{
			name:    "MaxBelowMin",
			specs:   []string{"mig:namePrefix=prefix,min=10,max=1"},
			wantErr: true,
		},
		{
			name:    "MaxIsZero",
			specs:   []string{"mig:namePrefix=prefix,min=0,max=0"},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			do := cloudprovider.NodeGroupDiscoveryOptions{NodeGroupAutoDiscoverySpecs: tc.specs}
			got, err := parseMIGAutoDiscoverySpecs(do)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.True(t, assert.ObjectsAreEqualValues(tc.want, got), "\ngot: %#v\nwant: %#v", got, tc.want)
		})
	}
}
