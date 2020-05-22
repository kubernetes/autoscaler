/*
Copyright 2020 The Kubernetes Authors.

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

package magnum

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum/fixtures"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum/gophercloud"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum/gophercloud/openstack/containerinfra/v1/nodegroups"
	th "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum/gophercloud/testhelper"
)

func createTestServiceClient() *gophercloud.ServiceClient {
	return &gophercloud.ServiceClient{
		ProviderClient: &gophercloud.ProviderClient{TokenID: "cbc36478b0bd8e67e89469c7749d4127"},
		Endpoint:       th.Endpoint() + "v1/",
	}
}

func createTestMagnumManager(client *gophercloud.ServiceClient) *magnumManagerImpl {
	return &magnumManagerImpl{
		clusterName:                fixtures.ClusterUUID,
		clusterClient:              client,
		heatClient:                 client,
		stackInfo:                  make(map[string]nodeGroupStacks),
		providerIDToNodeGroupCache: make(map[string]string),
	}
}

// setupFetchNodeGroupStackIDs sets up handlers for the default-worker and
// test-ng node groups, and their stacks.
func setupFetchNodeGroupStackIDs() {
	// Get default worker node group
	path := fmt.Sprintf("/v1/clusters/%s/nodegroups/%s", fixtures.ClusterUUID, fixtures.DefaultWorkerNodeGroupUUID)
	th.Mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, fixtures.GetDefaultWorkerNodeGroupResponse)
	})

	// Get the stack of the default-worker node group
	path = fmt.Sprintf("/v1/stacks/%s", fixtures.DefaultWorkerStackUUID)
	th.Mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, fixtures.GetDefaultWorkerStackResponse)
	})

	// Get the kube_minions resource of the default-worker node group
	path = fmt.Sprintf("/v1/stacks/%s/%s/resources/kube_minions", fixtures.DefaultWorkerStackName, fixtures.DefaultWorkerStackUUID)
	th.Mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, fixtures.GetDefaultWorkerKubeMinionsResourceResponse)
	})

	// Get the physical stack of the kube_minions resource of the default-worker node group
	path = fmt.Sprintf("/v1/stacks/%s", fixtures.DefaultWorkerKubeMinionsStackUUID)
	th.Mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, fixtures.GetDefaultWorkerKubeMinionsStackResponse)
	})

	// Get test-ng node group
	path = fmt.Sprintf("/v1/clusters/%s/nodegroups/%s", fixtures.ClusterUUID, fixtures.TestNodeGroupUUID)
	th.Mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, fixtures.GetTestNGNodeGroupResponse)
	})

	// Get the stack of the test-ng node group
	path = fmt.Sprintf("/v1/stacks/%s", fixtures.TestNodeGroupStackUUID)
	th.Mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, fixtures.GetTestNGStackResponse)
	})

	// Get the kube_minions resource of the test-ng node group
	path = fmt.Sprintf("/v1/stacks/%s/%s/resources/kube_minions", fixtures.TestNodeGroupStackName, fixtures.TestNodeGroupStackUUID)
	th.Mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, fixtures.GetTestNGKubeMinionsResourceResponse)
	})

	// Get the physical stack of the kube_minions resource of the test-ng node group
	path = fmt.Sprintf("/v1/stacks/%s", fixtures.TestNodeGroupKubeMinionsStackUUID)
	th.Mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, fixtures.GetTestNGKubeMinionsStackResponse)
	})
}

// TestFetchNodeGroupStackIDs checks that all stack IDs and stack names
// for a node group are correctly found.
func TestFetchNodeGroupStackIDs(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	setupFetchNodeGroupStackIDs()

	sc := createTestServiceClient()
	manager := createTestMagnumManager(sc)

	stackInfo, err := manager.fetchNodeGroupStackIDs(fixtures.DefaultWorkerNodeGroupUUID)
	assert.NoError(t, err)
	assert.Equal(t, fixtures.DefaultWorkerStackUUID, stackInfo.stackID)
	assert.Equal(t, fixtures.DefaultWorkerKubeMinionsStackUUID, stackInfo.kubeMinionsStackID)

	stackInfo, err = manager.fetchNodeGroupStackIDs(fixtures.TestNodeGroupUUID)
	assert.NoError(t, err)
	assert.Equal(t, fixtures.TestNodeGroupStackUUID, stackInfo.stackID)
	assert.Equal(t, fixtures.TestNodeGroupKubeMinionsStackUUID, stackInfo.kubeMinionsStackID)
}

// TestFetchNodeGroupStackIDsCaching checks that the stack IDs for a
// node group are only queried from the API once, and are cached after that.
func TestFetchNodeGroupStackIDsCaching(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	apiCalls := 0

	// Get default worker node group
	path := fmt.Sprintf("/v1/clusters/%s/nodegroups/%s", fixtures.ClusterUUID, fixtures.DefaultWorkerNodeGroupUUID)
	th.Mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		apiCalls += 1

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, fixtures.GetDefaultWorkerNodeGroupResponse)
	})

	// Get the stack of the default-worker node group
	path = fmt.Sprintf("/v1/stacks/%s", fixtures.DefaultWorkerStackUUID)
	th.Mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, fixtures.GetDefaultWorkerStackResponse)
	})

	// Get the kube_minions resource of the default-worker node group
	path = fmt.Sprintf("/v1/stacks/%s/%s/resources/kube_minions", fixtures.DefaultWorkerStackName, fixtures.DefaultWorkerStackUUID)
	th.Mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, fixtures.GetDefaultWorkerKubeMinionsResourceResponse)
	})

	// Get the physical stack of the kube_minions resource of the default-worker node group
	path = fmt.Sprintf("/v1/stacks/%s", fixtures.DefaultWorkerKubeMinionsStackUUID)
	th.Mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, fixtures.GetDefaultWorkerKubeMinionsStackResponse)
	})

	sc := createTestServiceClient()
	manager := createTestMagnumManager(sc)

	stackInfo, err := manager.fetchNodeGroupStackIDs(fixtures.DefaultWorkerNodeGroupUUID)
	assert.NoError(t, err)
	assert.Equal(t, fixtures.DefaultWorkerStackUUID, stackInfo.stackID)
	assert.Equal(t, fixtures.DefaultWorkerKubeMinionsStackUUID, stackInfo.kubeMinionsStackID)
	assert.Equal(t, 1, apiCalls)

	stackInfo, err = manager.fetchNodeGroupStackIDs(fixtures.DefaultWorkerNodeGroupUUID)
	assert.NoError(t, err)
	assert.Equal(t, fixtures.DefaultWorkerStackUUID, stackInfo.stackID)
	assert.Equal(t, fixtures.DefaultWorkerKubeMinionsStackUUID, stackInfo.kubeMinionsStackID)
	assert.Equalf(t, 1, apiCalls, "should only fetch the node group from Magnum API once")
}

// TestNodeGroupForNode tests finding the node groups
// of two nodes belonging to different groups.
func TestNodeGroupForNode(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	setupFetchNodeGroupStackIDs()

	path := fmt.Sprintf("/v1/clusters/%s/nodegroups", fixtures.ClusterUUID)
	th.Mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, fixtures.ListNodeGroupsResponse)
	})

	// Get the physical stack of the kube_minions resource of the default-worker node group
	path = fmt.Sprintf("/v1/stacks/%s/%s", fixtures.DefaultWorkerKubeMinionsStackName, fixtures.DefaultWorkerKubeMinionsStackUUID)
	th.Mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, fixtures.GetDefaultWorkerKubeMinionsStackResponse)
	})

	// Get the physical stack of the kube_minions resource of the test-ng node group
	path = fmt.Sprintf("/v1/stacks/%s/%s", fixtures.TestNodeGroupKubeMinionsStackName, fixtures.TestNodeGroupKubeMinionsStackUUID)
	th.Mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, fixtures.GetTestNGKubeMinionsStackResponse)
	})

	sc := createTestServiceClient()

	manager := createTestMagnumManager(sc)

	node := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kube-w5axvlfrz5lr-node-0",
		},
		Spec: apiv1.NodeSpec{
			ProviderID: "openstack:///fa405ca9-4486-4159-9763-0f82fbc2e4ac",
		},
	}

	group, err := manager.nodeGroupForNode(node)
	assert.NoError(t, err)
	assert.Equal(t, fixtures.DefaultWorkerNodeGroupUUID, group)

	node2 := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kube-test-ng-j5g7osyr44tb-node-1",
		},
		Spec: apiv1.NodeSpec{
			ProviderID: "openstack:///d28f8e0f-da6e-434d-99aa-aa8baf2b328f",
		},
	}

	group2, err := manager.nodeGroupForNode(node2)
	assert.NoError(t, err)
	assert.Equal(t, fixtures.TestNodeGroupUUID, group2)
}

// TestNodeGroupForNodeCaching checks that the minimal amount
// of API calls are made when calling nodeGroupForNode multiple times.
func TestNodeGroupForNodeCaching(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	setupFetchNodeGroupStackIDs()

	timesListCalled := 0

	path := fmt.Sprintf("/v1/clusters/%s/nodegroups", fixtures.ClusterUUID)
	th.Mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		timesListCalled += 1
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, fixtures.ListNodeGroupsResponse)
	})

	// Get the physical stack of the kube_minions resource of the default-worker node group
	path = fmt.Sprintf("/v1/stacks/%s/%s", fixtures.DefaultWorkerKubeMinionsStackName, fixtures.DefaultWorkerKubeMinionsStackUUID)
	th.Mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, fixtures.GetDefaultWorkerKubeMinionsStackResponse)
	})

	// Get the physical stack of the kube_minions resource of the test-ng node group
	path = fmt.Sprintf("/v1/stacks/%s/%s", fixtures.TestNodeGroupKubeMinionsStackName, fixtures.TestNodeGroupKubeMinionsStackUUID)
	th.Mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, fixtures.GetTestNGKubeMinionsStackResponse)
	})

	sc := createTestServiceClient()

	manager := createTestMagnumManager(sc)

	for i := 0; i < 5; i++ {
		node := &apiv1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "kube-w5axvlfrz5lr-node-0",
			},
			Spec: apiv1.NodeSpec{
				ProviderID: "openstack:///fa405ca9-4486-4159-9763-0f82fbc2e4ac",
			},
		}

		group, err := manager.nodeGroupForNode(node)
		assert.NoError(t, err)
		assert.Equal(t, fixtures.DefaultWorkerNodeGroupUUID, group)
	}

	assert.Equal(t, 1, timesListCalled, "API should only be contacted once, the result should be cached")
}

// TestNodeGroupForNodeError checks that an unknown node ID will
// not be found in any node group, and that an error will be returned
// instead.
func TestNodeGroupForNodeError(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	setupFetchNodeGroupStackIDs()

	path := fmt.Sprintf("/v1/clusters/%s/nodegroups", fixtures.ClusterUUID)
	th.Mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, fixtures.ListNodeGroupsResponse)
	})

	// Get the physical stack of the kube_minions resource of the default-worker node group
	path = fmt.Sprintf("/v1/stacks/%s/%s", fixtures.DefaultWorkerKubeMinionsStackName, fixtures.DefaultWorkerKubeMinionsStackUUID)
	th.Mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, fixtures.GetDefaultWorkerKubeMinionsStackResponse)
	})

	// Get the physical stack of the kube_minions resource of the test-ng node group
	path = fmt.Sprintf("/v1/stacks/%s/%s", fixtures.TestNodeGroupKubeMinionsStackName, fixtures.TestNodeGroupKubeMinionsStackUUID)
	th.Mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, fixtures.GetTestNGKubeMinionsStackResponse)
	})

	sc := createTestServiceClient()

	manager := createTestMagnumManager(sc)

	node := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kube-w5axvlfrz5lr-node-0",
		},
		Spec: apiv1.NodeSpec{
			ProviderID: "openstack:///not-a-node",
		},
	}

	_, err := manager.nodeGroupForNode(node)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not find node group for node")
}

// TestNodeGroupForNodeFake passes a node with a fake provider ID
// to nodeGroupForNode and asserts that it returns the node group
// UUID that is contained in the provider ID.
func TestNodeGroupForNodeFake(t *testing.T) {
	sc := createTestServiceClient()

	manager := createTestMagnumManager(sc)

	node := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: fmt.Sprintf("fake:///%s/5", fixtures.DefaultWorkerNodeGroupUUID),
		},
	}

	group, err := manager.nodeGroupForNode(node)
	assert.NoError(t, err)
	assert.Equal(t, fixtures.DefaultWorkerNodeGroupUUID, group)
}

// TestGetNodes tests calling getNodes for a node group
// with a single running node, and checks that the single
// instance is returned as expected.
func TestGetNodes(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	setupFetchNodeGroupStackIDs()

	// Get the physical stack of the kube_minions resource of the default-worker node group
	path := fmt.Sprintf("/v1/stacks/%s/%s", fixtures.DefaultWorkerKubeMinionsStackName, fixtures.DefaultWorkerKubeMinionsStackUUID)
	th.Mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, fixtures.GetDefaultWorkerKubeMinionsStackResponse)
	})

	// List the resources belong to the kube_minions stack
	path = fmt.Sprintf("/v1/stacks/%s/%s/resources", fixtures.DefaultWorkerKubeMinionsStackName, fixtures.DefaultWorkerKubeMinionsStackUUID)
	th.Mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, fixtures.ListDefaultWorkerKubeMinionsResources)
	})

	sc := createTestServiceClient()

	manager := createTestMagnumManager(sc)

	nodes, err := manager.getNodes(fixtures.DefaultWorkerNodeGroupUUID)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(nodes))
	assert.Equal(t, fixtures.DefaultWorkerMinion0ProviderID, nodes[0].Id)
	assert.Equal(t, cloudprovider.InstanceRunning, nodes[0].Status.State)
}

// TestGetNodesAllStates tests calling getNodes for a node group
// with a node in every possible state.
func TestGetNodesAllStates(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	// Get the all-states node group
	path := fmt.Sprintf("/v1/clusters/%s/nodegroups/%s", fixtures.ClusterUUID, fixtures.AllStatesNodeGroupUUID)
	th.Mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, fixtures.GetAllStatesNodeGroupResponse)
	})

	// Get the stack of the default-worker node group
	path = fmt.Sprintf("/v1/stacks/%s", fixtures.AllStatesStackUUID)
	th.Mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, fixtures.GetAllStatesStackResponse)
	})

	// Get the kube_minions resource of the default-worker node group
	path = fmt.Sprintf("/v1/stacks/%s/%s/resources/kube_minions", fixtures.AllStatesStackName, fixtures.AllStatesStackUUID)
	th.Mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, fixtures.GetAllStatesKubeMinionsResourceResponse)
	})

	// Get the physical stack of the kube_minions resource of the default-worker node group
	path = fmt.Sprintf("/v1/stacks/%s", fixtures.AllStatesKubeMinionsStackUUID)
	th.Mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, fixtures.GetAllStatesKubeMinionsStackResponse)
	})

	// Get the physical stack of the kube_minions resource of the default-worker node group
	path = fmt.Sprintf("/v1/stacks/%s/%s", fixtures.AllStatesKubeMinionsStackName, fixtures.AllStatesKubeMinionsStackUUID)
	th.Mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, fixtures.GetAllStatesKubeMinionsStackResponse)
	})

	// List the resources belong to the kube_minions stack
	path = fmt.Sprintf("/v1/stacks/%s/%s/resources", fixtures.AllStatesKubeMinionsStackName, fixtures.AllStatesKubeMinionsStackUUID)
	th.Mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, fixtures.ListAllStatesKubeMinionsResources)
	})

	for _, node := range fixtures.AllNodes {
		// Handle the response for listing the resources of a specific minion
		if node.UUID == "kube-minion" {
			continue
		}
		path := fmt.Sprintf("/v1/stacks/%s/resources", node.UUID)
		response := fixtures.BuildAllStatesKubeMinionPhysicalResources(node.Index)
		th.Mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			fmt.Fprint(w, response)
		})

	}

	sc := createTestServiceClient()

	manager := createTestMagnumManager(sc)

	nodes, err := manager.getNodes(fixtures.AllStatesNodeGroupUUID)
	assert.NoError(t, err)
	assert.ElementsMatch(t, nodes, fixtures.ExpectedInstances)
}

// TestAutoDiscoverNodeGroups checks that only the node groups
// which meet all requirements for autoscaling are returned by
// node group auto discovery.
//
// That means:
// * node groups which match an auto discovery config (node group role)
// * node groups which have a max node count set in Magnum
//
// Both of the above points must be satisfied for a node group to be
// auto discovered.
func TestAutoDiscoverNodeGroups(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	sc := createTestServiceClient()
	manager := createTestMagnumManager(sc)

	five := 5
	autoScalingNodeGroups := []*nodegroups.NodeGroup{
		{
			UUID:         "435d8e97-e4cc-4c6a-9d04-ed64523c4f10",
			Name:         "test-ng-1",
			Role:         "autoscaling",
			NodeCount:    2,
			MinNodeCount: 1,
			MaxNodeCount: &five,
			IsDefault:    false,
			Status:       "UPDATE_COMPLETE",
		},
		{
			UUID:         "d31a8cc1-6b4b-4e94-bd0d-9020ebbc033e",
			Name:         "test-ng-2",
			Role:         "autoscaling",
			NodeCount:    1,
			MinNodeCount: 1,
			MaxNodeCount: &five,
			IsDefault:    false,
			Status:       "UPDATE_COMPLETE",
		},
	}

	ignoredNodeGroups := []*nodegroups.NodeGroup{
		{
			// Not autoscaled - master node group
			UUID:         "8b46b5f8-52df-4d40-a2a3-6b8b2f094c1a",
			Name:         "default-master",
			Role:         "master",
			NodeCount:    1,
			MinNodeCount: 1,
			MaxNodeCount: nil,
			IsDefault:    true,
			Status:       "UPDATE_COMPLETE",
		},
		{
			// Not autoscaled - wrong role
			UUID:         "d905084e-ac2c-4db8-be98-f6b58c075a41",
			Name:         "default-worker",
			Role:         "worker",
			NodeCount:    1,
			MinNodeCount: 1,
			MaxNodeCount: nil,
			IsDefault:    true,
			Status:       "UPDATE_COMPLETE",
		},
		{
			// Not autoscaled - right role but max node count not set
			UUID:         "f718ad99-1473-4fa6-b5da-4f78fcc2aef8",
			Name:         "test-ng-3",
			Role:         "autoscaling",
			NodeCount:    1,
			MinNodeCount: 1,
			MaxNodeCount: nil,
			IsDefault:    false,
			Status:       "UPDATE_COMPLETE",
		},
		{
			// Not autoscaled - Max node count set but wrong role
			UUID:         "4abd7f6c-086d-42ba-a439-99b22a0db5f9",
			Name:         "test-ng-4",
			Role:         "other",
			NodeCount:    1,
			MinNodeCount: 1,
			MaxNodeCount: &five,
			IsDefault:    false,
			Status:       "UPDATE_COMPLETE",
		},
	}

	allNodeGroups := append(autoScalingNodeGroups, ignoredNodeGroups...)

	// Build a Magnum response for listing all nodegroups.
	response := fixtures.BuildTemplatedNodeGroupsListResponse(allNodeGroups)

	path := fmt.Sprintf("/v1/clusters/%s/nodegroups", fixtures.ClusterUUID)
	th.Mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, response)
	})

	for _, ng := range allNodeGroups {
		// Build a Magnum response for getting this node group.
		response := fixtures.BuildTemplatedNodeGroupsGetResponse(ng)

		path := fmt.Sprintf("/v1/clusters/%s/nodegroups/%s", fixtures.ClusterUUID, ng.UUID)
		th.Mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			fmt.Fprint(w, response)
		})
	}

	// Create a discovery config for selecting the role "autoscaling".
	options := cloudprovider.NodeGroupDiscoveryOptions{
		NodeGroupAutoDiscoverySpecs: []string{"magnum:role=autoscaling"},
	}
	configs, err := parseMagnumAutoDiscoverySpecs(options)
	require.NoError(t, err)

	// Discover node groups and check that only the node groups which should be autoscaled are returned.
	discoveredNodeGroups, err := manager.autoDiscoverNodeGroups(configs)
	assert.NoError(t, err)
	assert.ElementsMatchf(t, autoScalingNodeGroups, discoveredNodeGroups, "autoDiscoverNodeGroups did not return the expected groups")
}

// TestUniqueNameAndIDForNodeGroup checks that the method
// uniqueNameAndIDForNodeGroup returns the expected output
// for a given node group, which is the name + first part of ID.
func TestUniqueNameAndIDForNodeGroup(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	sc := createTestServiceClient()
	manager := createTestMagnumManager(sc)

	path := fmt.Sprintf("/v1/clusters/%s/nodegroups/%s", fixtures.ClusterUUID, fixtures.DefaultWorkerNodeGroupUUID)
	th.Mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, fixtures.GetDefaultWorkerNodeGroupResponse)
	})

	uniqueName, ID, err := manager.uniqueNameAndIDForNodeGroup(fixtures.DefaultWorkerNodeGroupUUID)
	assert.NoError(t, err)
	assert.Equal(t, "default-worker-f4ecc247", uniqueName)
	assert.Equal(t, fixtures.DefaultWorkerNodeGroupUUID, ID)
}

func TestNodeGroupSize(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	sc := createTestServiceClient()
	manager := createTestMagnumManager(sc)

	path := fmt.Sprintf("/v1/clusters/%s/nodegroups/%s", fixtures.ClusterUUID, fixtures.DefaultWorkerNodeGroupUUID)
	th.Mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, fixtures.GetDefaultWorkerNodeGroupResponse)
	})

	n, err := manager.nodeGroupSize(fixtures.DefaultWorkerNodeGroupUUID)
	assert.NoError(t, err)
	assert.Equal(t, 1, n)
}

// TestUpdateNodeCountSuccess checks that updateNodeCount
// correctly makes the resize request to Magnum and returns
// no error if the resize was accepted.
func TestUpdateNodeCountSuccess(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	sc := createTestServiceClient()
	manager := createTestMagnumManager(sc)

	// Expected values to receive in the resize request
	ngToResize := fixtures.DefaultWorkerNodeGroupUUID
	nodeCount := 2

	path := fmt.Sprintf("/v1/clusters/%s/actions/resize", fixtures.ClusterUUID)
	th.Mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		th.TestMethod(t, r, "POST")
		th.TestJSONRequest(t, r, fmt.Sprintf(`{"node_count": %d, "nodegroup": "%s"}`, nodeCount, ngToResize))
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, fixtures.ResizeDefaultWorkerNodeGroupResponse)
	})

	err := manager.updateNodeCount(ngToResize, nodeCount)
	assert.NoError(t, err)
}

// TestUpdateNodeCountError checks that an error is returned from
// updateNodeCount if the request is not valid, for example if
// the requested node count exceeds the maximum set in Magnum.
func TestUpdateNodeCountError(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	sc := createTestServiceClient()
	manager := createTestMagnumManager(sc)

	// Expected values to receive in the resize request
	ngToResize := fixtures.DefaultWorkerNodeGroupUUID
	nodeCount := 10

	path := fmt.Sprintf("/v1/clusters/%s/actions/resize", fixtures.ClusterUUID)
	th.Mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		th.TestMethod(t, r, "POST")
		th.TestJSONRequest(t, r, fmt.Sprintf(`{"node_count": %d, "nodegroup": "%s"}`, nodeCount, ngToResize))
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		fmt.Fprint(w, fixtures.ResizeDefaultWorkerNodeGroupError)
	})

	err := manager.updateNodeCount(ngToResize, nodeCount)
	assert.Error(t, err)
}

func TestManagerDeleteNodes(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	sc := createTestServiceClient()
	manager := createTestMagnumManager(sc)

	// Expected values to receive in the resize request
	ngToResize := fixtures.DefaultWorkerNodeGroupUUID
	nodeCount := 1
	nodesToRemoveList := []string{"47a5d403-109e-481f-8c01-42f72e08d6a7", "db015c35-c975-43d3-b64e-a4134c94a8f1", "3"}

	var remove []string
	for _, node := range nodesToRemoveList {
		remove = append(remove, fmt.Sprintf(`"%s"`, node))
	}
	nodesToRemove := strings.Join(remove, ", ")

	path := fmt.Sprintf("/v1/clusters/%s/actions/resize", fixtures.ClusterUUID)
	th.Mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		th.TestMethod(t, r, "POST")

		th.TestJSONRequest(t, r, fmt.Sprintf(`{"node_count": %d, "nodegroup": "%s", "nodes_to_remove": [%s]}`, nodeCount, ngToResize, nodesToRemove))
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, fixtures.ResizeDefaultWorkerNodeGroupResponse)
	})

	fakeProviderID := fmt.Sprintf("fake:///%s/3", fixtures.DefaultWorkerNodeGroupUUID)

	instances := []NodeRef{
		{Name: "openstack:///47a5d403-109e-481f-8c01-42f72e08d6a7", SystemUUID: "47a5d403-109e-481f-8c01-42f72e08d6a7"},
		{Name: "openstack:///db015c35-c975-43d3-b64e-a4134c94a8f1", SystemUUID: "db015c35-c975-43d3-b64e-a4134c94a8f1"},
		{Name: fakeProviderID, ProviderID: fakeProviderID, IsFake: true},
	}

	err := manager.deleteNodes(fixtures.DefaultWorkerNodeGroupUUID, instances, 1)
	assert.NoError(t, err)
}
