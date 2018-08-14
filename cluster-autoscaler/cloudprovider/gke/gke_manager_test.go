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

package gke

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/gce"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"

	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	gce_api "google.golang.org/api/compute/v1"
	apiv1 "k8s.io/api/core/v1"
)

const (
	projectId              = "project1"
	zoneB                  = "us-central1-b"
	zoneC                  = "us-central1-c"
	zoneF                  = "us-central1-f"
	region                 = "us-central1"
	defaultPoolMig         = "gke-cluster-1-default-pool"
	defaultPool            = "default-pool"
	autoprovisionedPoolMig = "gke-cluster-1-nodeautoprovisioning-323233232"
	autoprovisionedPool    = "nodeautoprovisioning-323233232"
	clusterName            = "cluster1"

	gkeMigA = "gce-mig-a"
	gkeMigB = "gce-mig-b"
)

const allNodePools1 = `{
  "nodePools": [
    {
      "name": "default-pool",
      "config": {
        "machineType": "n1-standard-1",
        "diskSizeGb": 100,
        "oauthScopes": [
          "https://www.googleapis.com/auth/compute",
          "https://www.googleapis.com/auth/devstorage.read_only",
          "https://www.googleapis.com/auth/logging.write",
          "https://www.googleapis.com/auth/monitoring.write",
          "https://www.googleapis.com/auth/servicecontrol",
          "https://www.googleapis.com/auth/service.management.readonly",
          "https://www.googleapis.com/auth/trace.append"
        ],
        "imageType": "COS",
        "serviceAccount": "default"
      },
      "initialNodeCount": 3,
      "autoscaling": {
         "Enabled": true,
         "MinNodeCount": 1,
         "MaxNodeCount": 11
      },
      "management": {},
      "selfLink": "https://container.googleapis.com/v1/projects/project1/locations/us-central1-b/clusters/cluster-1/nodePools/default-pool",
      "version": "1.6.9",
      "instanceGroupUrls": [
        "https://www.googleapis.com/compute/v1/projects/project1/zones/us-central1-b/instanceGroupManagers/gke-cluster-1-default-pool"
      ],
      "status": "RUNNING"
    }
  ]
}`

const allNodePoolsRegional = `{
  "nodePools": [
    {
      "name": "default-pool",
      "config": {
        "machineType": "n1-standard-1",
        "diskSizeGb": 100,
        "oauthScopes": [
          "https://www.googleapis.com/auth/compute",
          "https://www.googleapis.com/auth/devstorage.read_only",
          "https://www.googleapis.com/auth/logging.write",
          "https://www.googleapis.com/auth/monitoring.write",
          "https://www.googleapis.com/auth/servicecontrol",
          "https://www.googleapis.com/auth/service.management.readonly",
          "https://www.googleapis.com/auth/trace.append"
        ],
        "imageType": "COS",
        "serviceAccount": "default"
      },
      "initialNodeCount": 3,
      "autoscaling": {
         "Enabled": true,
         "MinNodeCount": 1,
         "MaxNodeCount": 11
      },
      "management": {},
      "selfLink": "https://container.googleapis.com/v1/projects/project1/locations/us-central1-b/clusters/cluster-1/nodePools/default-pool",
      "version": "1.6.9",
      "instanceGroupUrls": [
        "https://www.googleapis.com/compute/v1/projects/project1/zones/us-central1-b/instanceGroupManagers/gke-cluster-1-default-pool",
        "https://www.googleapis.com/compute/v1/projects/project1/zones/us-central1-c/instanceGroupManagers/gke-cluster-1-default-pool",
        "https://www.googleapis.com/compute/v1/projects/project1/zones/us-central1-f/instanceGroupManagers/gke-cluster-1-default-pool"
      ],
      "status": "RUNNING"
    }
  ]
}`

const allNodePools2 = `{
  "nodePools": [
    {
      "name": "default-pool",
      "config": {
        "machineType": "n1-standard-1",
        "diskSizeGb": 100,
        "oauthScopes": [
          "https://www.googleapis.com/auth/compute",
          "https://www.googleapis.com/auth/devstorage.read_only",
          "https://www.googleapis.com/auth/logging.write",
          "https://www.googleapis.com/auth/monitoring.write",
          "https://www.googleapis.com/auth/servicecontrol",
          "https://www.googleapis.com/auth/service.management.readonly",
          "https://www.googleapis.com/auth/trace.append"
        ],
        "imageType": "COS",
        "serviceAccount": "default"
      },
      "initialNodeCount": 3,
      "autoscaling": {
         "Enabled": true,
         "MinNodeCount": 1,
         "MaxNodeCount": 11},
      "management": {},
      "selfLink": "https://container.googleapis.com/v1/projects/project1/locations/us-central1-b/clusters/cluster-1/nodePools/default-pool",
      "version": "1.6.9",
      "instanceGroupUrls": [
        "https://www.googleapis.com/compute/v1/projects/project1/zones/us-central1-b/instanceGroupManagers/gke-cluster-1-default-pool"
      ],
      "status": "RUNNING"
    },
    {
      "name": "nodeautoprovisioning-323233232",
      "config": {
        "machineType": "n1-standard-1",
        "diskSizeGb": 100,
        "oauthScopes": [
          "https://www.googleapis.com/auth/compute",
          "https://www.googleapis.com/auth/devstorage.read_only",
          "https://www.googleapis.com/auth/logging.write",
          "https://www.googleapis.com/auth/monitoring.write",
          "https://www.googleapis.com/auth/servicecontrol",
          "https://www.googleapis.com/auth/service.management.readonly",
          "https://www.googleapis.com/auth/trace.append"
        ],
        "imageType": "COS",
        "serviceAccount": "default"
      },
      "initialNodeCount": 3,
      "autoscaling": {
         "Enabled": true,
         "MinNodeCount": 0,
         "MaxNodeCount": 1000
      },
      "management": {},
      "selfLink": "https://container.googleapis.com/v1/projects/project1/locations/us-central1-b/clusters/cluster-1/nodePools/nodeautoprovisioning-323233232",
      "version": "1.6.9",
      "instanceGroupUrls": [
        "https://www.googleapis.com/compute/v1/projects/project1/zones/us-central1-b/instanceGroupManagers/gke-cluster-1-nodeautoprovisioning-323233232"
      ],
      "status": "RUNNING"
    }
  ]
}`

const instanceGroupManager = `{
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
  "targetSize": 3,
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

const managedInstancesResponse1 = `{
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

const managedInstancesResponse2 = `{
  "managedInstances": [
    {
      "instance": "https://www.googleapis.com/compute/v1/projects/project1/zones/%s/instances/%s-gdf607aac-9j4g",
      "id": "1974815323221473983",
      "instanceStatus": "RUNNING",
      "currentAction": "NONE"
    }
  ]
}`

const getClusterResponse = `{
  "name": "usertest",
  "nodeConfig": {
    "machineType": "n1-standard-1",
    "diskSizeGb": 100,
    "oauthScopes": [
      "https://www.googleapis.com/auth/compute",
      "https://www.googleapis.com/auth/devstorage.read_only",
      "https://www.googleapis.com/auth/service.management.readonly",
      "https://www.googleapis.com/auth/servicecontrol",
      "https://www.googleapis.com/auth/logging.write",
      "https://www.googleapis.com/auth/monitoring"
    ],
    "imageType": "COS",
    "serviceAccount": "default",
    "diskType": "pd-standard"
  },
  "masterAuth": {
    "username": "admin",
    "password": "pass",
    "clusterCaCertificate": "cer1",
    "clientCertificate": "cer1",
    "clientKey": "cer1=="
  },
  "loggingService": "logging.googleapis.com",
  "monitoringService": "monitoring.googleapis.com",
  "network": "default",
  "clusterIpv4Cidr": "10.32.0.0/14",
  "addonsConfig": {
    "networkPolicyConfig": {
      "disabled": true
    }
  },
  "nodePools": [
    {
      "name": "default-pool",
      "config": {
        "machineType": "n1-standard-1",
        "diskSizeGb": 100,
        "oauthScopes": [
          "https://www.googleapis.com/auth/compute",
          "https://www.googleapis.com/auth/devstorage.read_only",
          "https://www.googleapis.com/auth/service.management.readonly",
          "https://www.googleapis.com/auth/servicecontrol",
          "https://www.googleapis.com/auth/logging.write",
          "https://www.googleapis.com/auth/monitoring"
        ],
        "imageType": "COS",
        "serviceAccount": "default",
        "diskType": "pd-standard"
      },
      "initialNodeCount": 1,
      "autoscaling": {
        "enabled": true,
        "maxNodeCount": 5
      },
      "management": {},
      "selfLink": "https:///v1beta1/projects/user-gke-dev/locations/us-central1-c/clusters/usertest/nodePools/default-pool",
      "version": "1.8.0-gke.1",
      "instanceGroupUrls": [
        "https://www.googleapis.com/compute/v1/projects/user-gke-dev/zones/us-central1-c/instanceGroupManagers/gke-usertest-default-pool-fdsafds2d5-grp"
      ],
      "status": "RUNNING"
    }
  ],
  "locations": [
    "us-central1-b"
  ],
  "labelFingerprint": "fasdfds",
  "legacyAbac": {},
  "autoscaling": {
    "resourceLimits": [
      {
        "resourceType": "cpu",
        "minimum": "2",
        "maximum": "3"
      },
      {
        "resourceType": "memory",
        "minimum": "2000000000",
        "maximum": "3000000000"
      }
    ]
  },
  "networkConfig": {
    "network": "https://www.googleapis.com/compute/v1/projects/user-gke-dev/global/networks/default"
  },
  "selfLink": "https:///v1beta1/projects/user-gke-dev/locations/us-central1-c/clusters/usertest",
  "zone": "us-central1-c",
  "endpoint": "xxx",
  "initialClusterVersion": "1.sdafsa",
  "currentMasterVersion": "1fdsfdsfsauser",
  "currentNodeVersion": "xxx",
  "createTime": "2017-10-24T12:20:00+00:00",
  "status": "RUNNING",
  "nodeIpv4CidrSize": 24,
  "servicesIpv4Cidr": "10.35.240.0/20",
  "instanceGroupUrls": [
    "https://www.googleapis.com/compute/v1/projects/user-gke-dev/zones/us-central1-c/instanceGroupManagers/gke-usertest-default-pool-323-grp"
  ],
  "currentNodeCount": 1
}`

func getInstanceGroupManager(zone string) string {
	return getInstanceGroupManagerNamed(defaultPoolMig, zone)
}

func getInstanceGroupManagerNamed(name, zone string) string {
	return fmt.Sprintf(instanceGroupManager, name, zone, name, zone, name, zone, name)
}

func getManagedInstancesResponse1(zone string) string {
	return getManagedInstancesResponse1Named(defaultPoolMig, zone)
}

func getManagedInstancesResponse1Named(name, zone string) string {
	return fmt.Sprintf(managedInstancesResponse1, zone, name, zone, name, zone, name, zone, name)
}

func getManagedInstancesResponse2(zone string) string {
	return getManagedInstancesResponse2Named(autoprovisionedPoolMig, zone)
}

func getManagedInstancesResponse2Named(name, zone string) string {
	return fmt.Sprintf(managedInstancesResponse2, zone, name)
}

func newTestAutoscalingGceClient(t *testing.T, projectId, url string, waitTimeout, pollInterval time.Duration) gce.AutoscalingGceClient {
	client := &http.Client{}
	gceClient, err := gce.NewCustomAutoscalingGceClientV1(client, projectId, url, waitTimeout, pollInterval)
	if !assert.NoError(t, err) {
		t.Fatalf("fatal error: %v", err)
	}
	return gceClient
}

func newTestGkeManager(t *testing.T, testServerURL string, mode GcpCloudProviderMode, regional bool) *gkeManagerImpl {
	// Override wait for op timeouts.
	waitTimeout := 50 * time.Millisecond
	pollInterval := 1 * time.Millisecond

	gceService := newTestAutoscalingGceClient(t, projectId, testServerURL, waitTimeout, pollInterval)

	manager := &gkeManagerImpl{
		cache:       gce.NewGceCache(gceService),
		GceService:  gceService,
		projectId:   projectId,
		clusterName: clusterName,
		mode:        mode,
		regional:    regional,
		templates:   &GkeTemplateBuilder{},
	}
	if regional {
		manager.location = region
	} else {
		manager.location = zoneB
	}
	machinesCache := map[gce.MachineTypeKey]*gce_api.MachineType{
		{"us-central1-b", "n1-standard-1"}: {GuestCpus: 1, MemoryMb: 1},
		{"us-central1-c", "n1-standard-1"}: {GuestCpus: 1, MemoryMb: 1},
		{"us-central1-f", "n1-standard-1"}: {GuestCpus: 1, MemoryMb: 1},
	}
	manager.cache.SetMachinesCache(machinesCache)

	client := &http.Client{}
	gkeAPIEndpoint = &testServerURL
	var err error
	if mode == ModeGKE {
		manager.GkeService, err = NewAutoscalingGkeClientV1(client, projectId, manager.location, clusterName)
		assert.NoError(t, err)
	}

	if mode == ModeGKENAP {
		manager.GkeService, err = NewAutoscalingGkeClientV1beta1(client, projectId, manager.location, clusterName)
		assert.NoError(t, err)
	}

	return manager
}

func validateMig(t *testing.T, mig gce.Mig, zone string, name string, minSize int, maxSize int) {
	assert.Equal(t, name, mig.GceRef().Name)
	assert.Equal(t, zone, mig.GceRef().Zone)
	assert.Equal(t, projectId, mig.GceRef().Project)
	assert.Equal(t, minSize, mig.MinSize())
	assert.Equal(t, maxSize, mig.MaxSize())
}

func TestRefreshNodePools(t *testing.T) {
	server := NewHttpServerMock()
	g := newTestGkeManager(t, server.URL, ModeGKE, false)

	// Fetch one node pool.
	server.On("handle", "/v1/projects/project1/locations/us-central1-b/clusters/cluster1/nodePools").Return(allNodePools1).Once()
	server.On("handle", "/project1/zones/us-central1-b/instanceGroupManagers/gke-cluster-1-default-pool").Return(getInstanceGroupManager(zoneB)).Once()
	server.On("handle", "/project1/global/instanceTemplates/gke-cluster-1-default-pool").Return(instanceTemplate).Once()
	server.On("handle", "/project1/zones/us-central1-b/instanceGroupManagers/gke-cluster-1-default-pool").Return(getInstanceGroupManager(zoneB)).Once()
	server.On("handle", "/project1/zones/us-central1-b/instanceGroupManagers/gke-cluster-1-default-pool/listManagedInstances").Return(getManagedInstancesResponse1(zoneB)).Once()

	err := g.refreshNodePools()
	assert.NoError(t, err)
	migs := g.GetMigs()
	assert.Equal(t, 1, len(migs))
	validateMig(t, migs[0].Config, zoneB, "gke-cluster-1-default-pool", 1, 11)
	mock.AssertExpectationsForObjects(t, server)

	// Fetch three node pools, skip one.

	server.On("handle", "/v1/projects/project1/locations/us-central1-b/clusters/cluster1/nodePools").Return(allNodePools2).Once()
	server.On("handle", "/project1/zones/us-central1-b/instanceGroupManagers/gke-cluster-1-nodeautoprovisioning-323233232").Return(getInstanceGroupManager(zoneB)).Once()
	server.On("handle", "/project1/global/instanceTemplates/gke-cluster-1-default-pool").Return(instanceTemplate).Once()
	server.On("handle", "/project1/zones/us-central1-b/instanceGroupManagers/gke-cluster-1-default-pool").Return(instanceGroupManager).Once()
	server.On("handle", "/project1/zones/us-central1-b/instanceGroupManagers/gke-cluster-1-default-pool/listManagedInstances").Return(getManagedInstancesResponse1(zoneB)).Once()
	server.On("handle", "/project1/zones/us-central1-b/instanceGroupManagers/gke-cluster-1-nodeautoprovisioning-323233232").Return(getInstanceGroupManager(zoneB)).Once()
	server.On("handle", "/project1/zones/us-central1-b/instanceGroupManagers/gke-cluster-1-nodeautoprovisioning-323233232/listManagedInstances").Return(getManagedInstancesResponse2(zoneB)).Once()

	err = g.refreshNodePools()
	assert.NoError(t, err)
	migs = g.GetMigs()
	assert.Equal(t, 2, len(migs))
	validateMig(t, migs[0].Config, zoneB, "gke-cluster-1-default-pool", 1, 11)
	validateMig(t, migs[1].Config, zoneB, "gke-cluster-1-nodeautoprovisioning-323233232", 0, 1000)
	mock.AssertExpectationsForObjects(t, server)

	// Fetch one node pool, remove node pool registered in previous step.

	server.On("handle", "/v1/projects/project1/locations/us-central1-b/clusters/cluster1/nodePools").Return(allNodePools1).Once()
	server.On("handle", "/project1/zones/us-central1-b/instanceGroupManagers/gke-cluster-1-default-pool").Return(getInstanceGroupManager(zoneB)).Once()
	server.On("handle", "/project1/zones/us-central1-b/instanceGroupManagers/gke-cluster-1-default-pool/listManagedInstances").Return(getManagedInstancesResponse1(zoneB)).Once()

	err = g.refreshNodePools()
	assert.NoError(t, err)
	migs = g.GetMigs()
	assert.Equal(t, 1, len(migs))
	validateMig(t, migs[0].Config, zoneB, "gke-cluster-1-default-pool", 1, 11)
	mock.AssertExpectationsForObjects(t, server)
}

func TestFetchAllNodePoolsRegional(t *testing.T) {
	server := NewHttpServerMock()
	g := newTestGkeManager(t, server.URL, ModeGKE, true)

	// Fetch one node pool.
	server.On("handle", "/v1/projects/project1/locations/us-central1/clusters/cluster1/nodePools").Return(allNodePoolsRegional).Once()
	server.On("handle", "/project1/zones/us-central1-b/instanceGroupManagers/gke-cluster-1-default-pool").Return(getInstanceGroupManager(zoneB)).Once()
	server.On("handle", "/project1/zones/us-central1-c/instanceGroupManagers/gke-cluster-1-default-pool").Return(getInstanceGroupManager(zoneC)).Once()
	server.On("handle", "/project1/zones/us-central1-f/instanceGroupManagers/gke-cluster-1-default-pool").Return(getInstanceGroupManager(zoneF)).Once()
	server.On("handle", "/project1/global/instanceTemplates/gke-cluster-1-default-pool").Return(instanceTemplate).Times(3)
	server.On("handle", "/project1/zones/us-central1-b/instanceGroupManagers/gke-cluster-1-default-pool").Return(getInstanceGroupManager(zoneB)).Once()
	server.On("handle", "/project1/zones/us-central1-c/instanceGroupManagers/gke-cluster-1-default-pool").Return(getInstanceGroupManager(zoneC)).Once()
	server.On("handle", "/project1/zones/us-central1-f/instanceGroupManagers/gke-cluster-1-default-pool").Return(getInstanceGroupManager(zoneF)).Once()
	server.On("handle", "/project1/zones/us-central1-b/instanceGroupManagers/gke-cluster-1-default-pool/listManagedInstances").Return(getManagedInstancesResponse1(zoneB)).Once()
	server.On("handle", "/project1/zones/us-central1-c/instanceGroupManagers/gke-cluster-1-default-pool/listManagedInstances").Return(getManagedInstancesResponse1(zoneC)).Once()
	server.On("handle", "/project1/zones/us-central1-f/instanceGroupManagers/gke-cluster-1-default-pool/listManagedInstances").Return(getManagedInstancesResponse1(zoneF)).Once()

	err := g.refreshNodePools()
	assert.NoError(t, err)
	migs := g.GetMigs()
	assert.Equal(t, 3, len(migs))
	validateMig(t, migs[0].Config, zoneB, "gke-cluster-1-default-pool", 1, 11)
	validateMig(t, migs[1].Config, zoneC, "gke-cluster-1-default-pool", 1, 11)
	validateMig(t, migs[2].Config, zoneF, "gke-cluster-1-default-pool", 1, 11)
	mock.AssertExpectationsForObjects(t, server)
}

const deleteNodePoolResponse = `{
  "name": "operation-1505732351373-819ed94e",
  "zone": "us-central1-a",
  "operationType": "DELETE_NODE_POOL",
  "status": "RUNNING",
  "selfLink": "https://container.googleapis.com/v1/projects/601024681890/locations/us-central1-a/operations/operation-1505732351373-819ed94e",
  "targetLink": "https://container.googleapis.com/v1/projects/601024681890/locations/us-central1-a/clusters/cluster-1/nodePools/nodeautoprovisioning-323233232",
  "startTime": "2017-09-18T10:59:11.373456931Z"
}`

const deleteNodePoolOperationResponse = `{
  "name": "operation-1505732351373-819ed94e",
  "zone": "us-central1-a",
  "operationType": "DELETE_NODE_POOL",
  "status": "DONE",
  "selfLink": "https://container.googleapis.com/v1/projects/601024681890/locations/us-central1-a/operations/operation-1505732351373-819ed94e",
  "targetLink": "https://container.googleapis.com/v1/projects/601024681890/locations/us-central1-a/clusters/cluster-1/nodePools/nodeautoprovisioning-323233232",
  "startTime": "2017-09-18T10:59:11.373456931Z"
}`

func TestDeleteNodePool(t *testing.T) {
	server := NewHttpServerMock()
	defer server.Close()
	g := newTestGkeManager(t, server.URL, ModeGKENAP, false)

	server.On("handle", "/v1beta1/projects/project1/locations/us-central1-b/clusters/cluster1/nodePools/nodeautoprovisioning-323233232").Return(deleteNodePoolResponse).Once()
	server.On("handle", "/v1beta1/projects/project1/locations/us-central1-b/operations/operation-1505732351373-819ed94e").Return(deleteNodePoolOperationResponse).Once()
	server.On("handle", "/v1beta1/projects/project1/locations/us-central1-b/clusters/cluster1/nodePools").Return(allNodePools2).Once()
	server.On("handle", "/project1/zones/us-central1-b/instanceGroupManagers/gke-cluster-1-default-pool").Return(getInstanceGroupManager(zoneB)).Once()
	server.On("handle", "/project1/zones/us-central1-b/instanceGroupManagers/gke-cluster-1-nodeautoprovisioning-323233232").Return(getInstanceGroupManager(zoneB)).Once()
	server.On("handle", "/project1/global/instanceTemplates/gke-cluster-1-default-pool").Return(instanceTemplate).Once()
	server.On("handle", "/project1/global/instanceTemplates/gke-cluster-1-default-pool").Return(instanceTemplate).Once()
	server.On("handle", "/project1/zones/us-central1-b/instanceGroupManagers/gke-cluster-1-default-pool").Return(instanceGroupManager).Once()
	server.On("handle", "/project1/zones/us-central1-b/instanceGroupManagers/gke-cluster-1-default-pool/listManagedInstances").Return(getManagedInstancesResponse1(zoneB)).Once()
	server.On("handle", "/project1/zones/us-central1-b/instanceGroupManagers/gke-cluster-1-nodeautoprovisioning-323233232").Return(getInstanceGroupManager(zoneB)).Once()
	server.On("handle", "/project1/zones/us-central1-b/instanceGroupManagers/gke-cluster-1-nodeautoprovisioning-323233232/listManagedInstances").Return(getManagedInstancesResponse2(zoneB)).Once()

	mig := &gkeMig{
		gceRef: gce.GceRef{
			Project: projectId,
			Zone:    zoneB,
			Name:    "nodeautoprovisioning-323233232",
		},
		gkeManager:      g,
		minSize:         0,
		maxSize:         1000,
		autoprovisioned: true,
		exist:           true,
		nodePoolName:    "nodeautoprovisioning-323233232",
		spec:            nil}

	err := g.DeleteNodePool(mig)
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, server)
}

const createNodePoolResponse = `{
  "name": "operation-1505728466148-d16f5197",
  "zone": "us-central1-a",
  "operationType": "CREATE_NODE_POOL",
  "status": "RUNNING",
  "selfLink": "https://container.googleapis.com/v1/projects/601024681890/locations/us-central1-a/operations/operation-1505728466148-d16f5197",
  "targetLink": "https://container.googleapis.com/v1/projects/601024681890/locations/us-central1-a/clusters/cluster-1/nodePools/nodeautoprovisioning-323233232",
  "startTime": "2017-09-18T09:54:26.148507311Z"
}`

const createNodePoolOperationResponse = `{
  "name": "operation-1505728466148-d16f5197",
  "zone": "us-central1-a",
  "operationType": "CREATE_NODE_POOL",
  "status": "DONE",
  "selfLink": "https://container.googleapis.com/v1/projects/601024681890/locations/us-central1-a/operations/operation-1505728466148-d16f5197",
  "targetLink": "https://container.googleapis.com/v1/projects/601024681890/locations/us-central1-a/clusters/cluster-1/nodePools/nodeautoprovisioning-323233232",
  "startTime": "2017-09-18T09:54:26.148507311Z",
  "endTime": "2017-09-18T09:54:35.124878859Z"
}`

func TestCreateNodePool(t *testing.T) {
	server := NewHttpServerMock()
	defer server.Close()
	g := newTestGkeManager(t, server.URL, ModeGKENAP, false)

	server.On("handle", "/v1beta1/projects/project1/locations/us-central1-b/operations/operation-1505728466148-d16f5197").Return(createNodePoolOperationResponse).Once()
	server.On("handle", "/v1beta1/projects/project1/locations/us-central1-b/clusters/cluster1/nodePools").Return(createNodePoolResponse).Once()

	server.On("handle", "/v1beta1/projects/project1/locations/us-central1-b/clusters/cluster1/nodePools").Return(allNodePools2).Once()
	server.On("handle", "/project1/zones/us-central1-b/instanceGroupManagers/gke-cluster-1-default-pool").Return(getInstanceGroupManager(zoneB)).Once()
	server.On("handle", "/project1/zones/us-central1-b/instanceGroupManagers/gke-cluster-1-nodeautoprovisioning-323233232").Return(getInstanceGroupManager(zoneB)).Once()
	server.On("handle", "/project1/global/instanceTemplates/gke-cluster-1-default-pool").Return(instanceTemplate).Once()
	server.On("handle", "/project1/global/instanceTemplates/gke-cluster-1-default-pool").Return(instanceTemplate).Once()
	server.On("handle", "/project1/zones/us-central1-b/instanceGroupManagers/gke-cluster-1-default-pool").Return(getInstanceGroupManager(zoneB)).Once()
	server.On("handle", "/project1/zones/us-central1-b/instanceGroupManagers/gke-cluster-1-default-pool/listManagedInstances").Return(getManagedInstancesResponse1(zoneB)).Once()
	server.On("handle", "/project1/zones/us-central1-b/instanceGroupManagers/gke-cluster-1-nodeautoprovisioning-323233232").Return(getInstanceGroupManager(zoneB)).Once()
	server.On("handle", "/project1/zones/us-central1-b/instanceGroupManagers/gke-cluster-1-nodeautoprovisioning-323233232/listManagedInstances").Return(getManagedInstancesResponse2(zoneB)).Once()

	mig := &gkeMig{
		gceRef: gce.GceRef{
			Project: projectId,
			Zone:    zoneB,
			Name:    "nodeautoprovisioning-323233232",
		},
		gkeManager:      g,
		minSize:         0,
		maxSize:         1000,
		autoprovisioned: true,
		exist:           true,
		nodePoolName:    "nodeautoprovisioning-323233232",
		spec: &MigSpec{
			MachineType: "n1-standard-1",
			Taints: []apiv1.Taint{
				{
					Key:   gpu.ResourceNvidiaGPU,
					Value: "present",
				},
				{
					Key:   "taint1",
					Value: "value",
				},
			},
		},
	}

	newMig, err := g.CreateNodePool(mig)
	assert.NoError(t, err)
	assert.True(t, newMig.Exist())
	migs := g.GetMigs()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(migs))
	mock.AssertExpectationsForObjects(t, server)
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

func setupTestNodePool(manager *gkeManagerImpl) {
	mig := &gkeMig{
		gceRef: gce.GceRef{
			Name:    defaultPoolMig,
			Zone:    zoneB,
			Project: projectId,
		},
		gkeManager:      manager,
		exist:           true,
		autoprovisioned: false,
		nodePoolName:    defaultPool,
		minSize:         1,
		maxSize:         11,
	}
	manager.cache.RegisterMig(mig)
}

func setupTestAutoprovisionedPool(manager *gkeManagerImpl) {
	mig := &gkeMig{
		gceRef: gce.GceRef{
			Name:    autoprovisionedPoolMig,
			Zone:    zoneB,
			Project: projectId,
		},
		gkeManager:      manager,
		exist:           true,
		autoprovisioned: true,
		nodePoolName:    autoprovisionedPool,
		minSize:         minAutoprovisionedSize,
		maxSize:         maxAutoprovisionedSize,
	}
	manager.cache.RegisterMig(mig)
}

func TestDeleteInstances(t *testing.T) {
	server := NewHttpServerMock()
	defer server.Close()
	g := newTestGkeManager(t, server.URL, ModeGKE, false)

	setupTestNodePool(g)
	setupTestAutoprovisionedPool(g)

	// Test DeleteInstance function.
	server.On("handle", "/project1/zones/us-central1-b/instanceGroupManagers/gke-cluster-1-default-pool").Return(getInstanceGroupManager(zoneB)).Once()
	server.On("handle", "/project1/zones/us-central1-b/instanceGroupManagers/gke-cluster-1-default-pool/listManagedInstances").Return(getManagedInstancesResponse1(zoneB)).Once()
	server.On("handle", "/project1/zones/us-central1-b/instanceGroupManagers/gke-cluster-1-nodeautoprovisioning-323233232").Return(getInstanceGroupManager(zoneB)).Once()
	server.On("handle", "/project1/zones/us-central1-b/instanceGroupManagers/gke-cluster-1-nodeautoprovisioning-323233232/listManagedInstances").Return(getManagedInstancesResponse2(zoneB)).Once()
	server.On("handle", "/project1/zones/us-central1-b/instanceGroupManagers/gke-cluster-1-default-pool/deleteInstances").Return(deleteInstancesResponse).Once()
	server.On("handle", "/project1/zones/us-central1-b/operations/operation-1505802641136-55984ff86d980-a99e8c2b-0c8aaaaa").Return(deleteInstancesOperationResponse).Once()

	instances := []*gce.GceRef{
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

	// Fail on deleting instances from different MIGs.
	instances = []*gce.GceRef{
		{
			Project: projectId,
			Zone:    zoneB,
			Name:    "gke-cluster-1-default-pool-f7607aac-f1hm",
		},
		{
			Project: projectId,
			Zone:    zoneB,
			Name:    "gke-cluster-1-nodeautoprovisioning-323233232-gdf607aac-9j4g",
		},
	}

	err = g.DeleteInstances(instances)
	assert.Error(t, err)
	assert.Equal(t, "Cannot delete instances which don't belong to the same MIG.", err.Error())
	mock.AssertExpectationsForObjects(t, server)
}

func TestGetMigSize(t *testing.T) {
	server := NewHttpServerMock()
	defer server.Close()
	g := newTestGkeManager(t, server.URL, ModeGKE, false)

	server.On("handle", "/project1/zones/us-central1-b/instanceGroupManagers/nodeautoprovisioning-323233232").Return(instanceGroupManager).Once()

	mig := &gkeMig{
		gceRef: gce.GceRef{
			Project: projectId,
			Zone:    zoneB,
			Name:    "nodeautoprovisioning-323233232",
		},
		gkeManager:      g,
		minSize:         0,
		maxSize:         1000,
		autoprovisioned: true,
		exist:           true,
		nodePoolName:    "nodeautoprovisioning-323233232",
		spec:            nil}

	size, err := g.GetMigSize(mig)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), size)
	mock.AssertExpectationsForObjects(t, server)
}

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

func TestSetMigSize(t *testing.T) {
	server := NewHttpServerMock()
	defer server.Close()
	g := newTestGkeManager(t, server.URL, ModeGKE, false)

	server.On("handle", "/project1/zones/us-central1-b/instanceGroupManagers/nodeautoprovisioning-323233232/resize").Return(setMigSizeResponse).Once()
	server.On("handle", "/project1/zones/us-central1-b/operations/operation-1505739408819-5597646964339-eb839c88-28805931").Return(setMigSizeOperationResponse).Once()

	mig := &gkeMig{
		gceRef: gce.GceRef{
			Project: projectId,
			Zone:    zoneB,
			Name:    "nodeautoprovisioning-323233232",
		},
		gkeManager:      g,
		minSize:         0,
		maxSize:         1000,
		autoprovisioned: true,
		exist:           true,
		nodePoolName:    "nodeautoprovisioning-323233232",
		spec:            nil}

	err := g.SetMigSize(mig, 3)
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, server)
}

func TestGetMigForInstance(t *testing.T) {
	server := NewHttpServerMock()
	defer server.Close()
	g := newTestGkeManager(t, server.URL, ModeGKE, false)

	setupTestNodePool(g)

	server.On("handle", "/project1/zones/us-central1-b/instanceGroupManagers/gke-cluster-1-default-pool").Return(getInstanceGroupManager(zoneB)).Once()
	server.On("handle", "/project1/zones/us-central1-b/instanceGroupManagers/gke-cluster-1-default-pool/listManagedInstances").Return(getManagedInstancesResponse1(zoneB)).Once()
	gceRef := &gce.GceRef{
		Project: projectId,
		Zone:    zoneB,
		Name:    "gke-cluster-1-default-pool-f7607aac-f1hm",
	}

	mig, err := g.GetMigForInstance(gceRef)
	assert.NoError(t, err)
	assert.NotNil(t, mig)
	assert.Equal(t, "gke-cluster-1-default-pool", mig.GceRef().Name)
	mock.AssertExpectationsForObjects(t, server)
}

func TestGetMigNodes(t *testing.T) {
	server := NewHttpServerMock()
	defer server.Close()
	g := newTestGkeManager(t, server.URL, ModeGKE, false)

	server.On("handle", "/project1/zones/us-central1-b/instanceGroupManagers/nodeautoprovisioning-323233232/listManagedInstances").Return(getManagedInstancesResponse1(zoneB)).Once()

	mig := &gkeMig{
		gceRef: gce.GceRef{
			Project: projectId,
			Zone:    zoneB,
			Name:    "nodeautoprovisioning-323233232",
		},
		gkeManager:      g,
		minSize:         0,
		maxSize:         1000,
		autoprovisioned: true,
		exist:           true,
		nodePoolName:    "nodeautoprovisioning-323233232",
		spec:            nil,
	}

	nodes, err := g.GetMigNodes(mig)
	assert.NoError(t, err)
	assert.Equal(t, 4, len(nodes))
	assert.Equal(t, "gce://project1/us-central1-b/gke-cluster-1-default-pool-f7607aac-9j4g", nodes[0])
	assert.Equal(t, "gce://project1/us-central1-b/gke-cluster-1-default-pool-f7607aac-c63g", nodes[1])
	assert.Equal(t, "gce://project1/us-central1-b/gke-cluster-1-default-pool-f7607aac-dck1", nodes[2])
	assert.Equal(t, "gce://project1/us-central1-b/gke-cluster-1-default-pool-f7607aac-f1hm", nodes[3])
	mock.AssertExpectationsForObjects(t, server)
}

func TestFetchResourceLimiter(t *testing.T) {
	server := NewHttpServerMock()
	defer server.Close()

	g := newTestGkeManager(t, server.URL, ModeGKENAP, false)
	server.On("handle", "/v1beta1/projects/project1/locations/us-central1-b/clusters/cluster1").Return(getClusterResponse).Once()

	err := g.fetchResourceLimiter()
	assert.NoError(t, err)
	resourceLimiter, err := g.GetResourceLimiter()
	assert.NoError(t, err)
	assert.NotNil(t, resourceLimiter)

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

func TestfetchMachinesCache(t *testing.T) {
	server := NewHttpServerMock()
	defer server.Close()

	g := newTestGkeManager(t, server.URL, ModeGKENAP, false)
	server.On("handle", "/v1alpha1/projects/project1/locations/us-central1-b/clusters/cluster1").Return(getClusterResponse).Once()
	server.On("handle", "/project1/zones/us-central1-b/machineTypes").Return(listMachineTypesResponse).Once()

	err := g.fetchMachinesCache()
	assert.NoError(t, err)
	assert.NotNil(t, g.cache.GetMachineFromCache("f1-micro", zoneB))
	assert.NotNil(t, g.cache.GetMachineFromCache("g1-small", zoneB))
	mock.AssertExpectationsForObjects(t, server)

	// Skipped refresh.
	err = g.fetchMachinesCache()
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, server)

	// Skipped refresh.
	server.On("handle", "/v1alpha1/projects/project1/locations/us-central1-b/clusters/cluster1").Return(getClusterResponse).Once()
	server.On("handle", "/project1/zones/us-central1-b/machineTypes").Return(listMachineTypesResponse).Once()
	g.machinesCacheLastRefresh = time.Now().Add(-2 * time.Hour)
	err = g.fetchMachinesCache()
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, server)
}

func TestGetMigTemplateNode(t *testing.T) {
	server := NewHttpServerMock()
	defer server.Close()

	server.On("handle", "/project1/zones/us-central1-b/instanceGroupManagers/default-pool").Return(getInstanceGroupManagerResponse).Once()
	server.On("handle", "/project1/global/instanceTemplates/gke-cluster-1-default-pool").Return(instanceTemplate).Once()

	regional := false
	g := newTestGkeManager(t, server.URL, ModeGKE, regional)

	mig := &gkeMig{
		gceRef: gce.GceRef{
			Project: projectId,
			Zone:    zoneB,
			Name:    "default-pool",
		},
		gkeManager:      g,
		minSize:         0,
		maxSize:         1000,
		autoprovisioned: true,
		exist:           true,
		nodePoolName:    "default-pool",
		spec:            nil,
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
	g := newTestGkeManager(t, server.URL, ModeGKE, regional)

	// Custom machine type.
	cpu, mem, err := g.getCpuAndMemoryForMachineType("custom-8-2", zoneB)
	assert.NoError(t, err)
	assert.Equal(t, int64(8), cpu)
	assert.Equal(t, int64(2*bytesPerMB), mem)
	mock.AssertExpectationsForObjects(t, server)

	// Standard machine type found in cache.
	cpu, mem, err = g.getCpuAndMemoryForMachineType("n1-standard-1", zoneB)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), cpu)
	assert.Equal(t, int64(1*bytesPerMB), mem)
	mock.AssertExpectationsForObjects(t, server)

	// Standard machine type not found in cache.
	server.On("handle", "/project1/zones/"+zoneB+"/machineTypes/n1-standard-2").Return(getMachineTypeResponse).Once()
	cpu, mem, err = g.getCpuAndMemoryForMachineType("n1-standard-2", zoneB)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), cpu)
	assert.Equal(t, int64(3840*bytesPerMB), mem)
	mock.AssertExpectationsForObjects(t, server)

	// Standard machine type cached.
	cpu, mem, err = g.getCpuAndMemoryForMachineType("n1-standard-2", zoneB)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), cpu)
	assert.Equal(t, int64(3840*bytesPerMB), mem)
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
	assert.Equal(t, int64(2816*bytesPerMB), mem)
	cpu, mem, err = parseCustomMachineType("other-a2-2816")
	assert.Error(t, err)
	cpu, mem, err = parseCustomMachineType("other-2-2816")
	assert.Error(t, err)
}
