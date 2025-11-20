/*
Copyright 2020-2023 Oracle and/or its affiliates.
*/

package nodepools

import (
	"context"
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/nodepools/consts"
	"net/http"
	"reflect"
	"testing"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	kubeletapis "k8s.io/kubelet/pkg/apis"

	ocicommon "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/common"
	oke "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/containerengine"
)

const (
	autoDiscoveryCompartment = "ocid1.compartment.oc1.test-region.test"
)

func TestNodePoolFromArgs(t *testing.T) {
	value := `1:5:ocid`
	nodePool, err := nodePoolFromArg(value)
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	if nodePool.minSize != 1 {
		t.Errorf("got minSize %d ; wanted minSize 1", nodePool.minSize)
	}

	if nodePool.maxSize != 5 {
		t.Errorf("got maxSize %d ; wanted maxSize 1", nodePool.maxSize)
	}

	if nodePool.id != "ocid" {
		t.Errorf("got ocid %q ; wanted id \"ocid\"", nodePool.id)
	}
}

func TestGetNodePoolSize(t *testing.T) {
	nodePoolCache := newNodePoolCache(nil)
	nodePoolCache.targetSize["id"] = 5

	manager := &ociManagerImpl{nodePoolCache: nodePoolCache}
	size, err := manager.GetNodePoolSize(&nodePool{id: "id"})
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	if size != 5 {
		t.Errorf("got size %d ; wanted size 5", size)
	}
}

func TestGetNodePoolForInstance(t *testing.T) {
	nodePoolCache := newNodePoolCache(nil)
	nodePoolCache.cache["ocid2"] = &oke.NodePool{
		Id: common.String("ocid2"),
		Nodes: []oke.Node{
			{Id: common.String("node1")},
		},
	}

	manager := &ociManagerImpl{
		staticNodePools: map[string]NodePool{
			"ocid1": &nodePool{id: "ocid1"},
			"ocid2": &nodePool{id: "ocid2"},
		},
		nodePoolCache: nodePoolCache,
	}

	// first verify node pool can be found when node pool id is specified.
	np, err := manager.GetNodePoolForInstance(ocicommon.OciRef{NodePoolID: "ocid1"})
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	if np.Id() != "ocid1" {
		t.Fatalf("got unexpected ocid %q ; wanted \"ocid1\"", np.Id())
	}

	// now verify node pool can be found via lookup up by instance id in cache
	np, err = manager.GetNodePoolForInstance(ocicommon.OciRef{InstanceID: "node1"})
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	if np.Id() != "ocid2" {
		t.Fatalf("got unexpected ocid %q ; wanted \"ocid2\"", np.Id())
	}
}

func TestGetNodePoolAvailabilityDomain(t *testing.T) {
	testCases := map[string]struct {
		np          *oke.NodePool
		result      string
		expectedErr bool
	}{
		"single ad": {
			np: &oke.NodePool{
				Id: common.String("id"),
				NodeConfigDetails: &oke.NodePoolNodeConfigDetails{
					PlacementConfigs: []oke.NodePoolPlacementConfigDetails{
						{AvailabilityDomain: common.String("hash:US-ASHBURN-1")},
					},
				},
			},
			result: "US-ASHBURN-1",
		},
		"multi-ad": {
			np: &oke.NodePool{
				Id: common.String("id"),
				NodeConfigDetails: &oke.NodePoolNodeConfigDetails{
					PlacementConfigs: []oke.NodePoolPlacementConfigDetails{
						{AvailabilityDomain: common.String("hash:US-ASHBURN-2")},
						{AvailabilityDomain: common.String("hash:US-ASHBURN-1")},
					},
				},
			},
			result: "US-ASHBURN-2",
		},
		"no placement configs": {
			np: &oke.NodePool{
				Id: common.String("id"),
				NodeConfigDetails: &oke.NodePoolNodeConfigDetails{
					PlacementConfigs: []oke.NodePoolPlacementConfigDetails{},
				},
			},
			expectedErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ad, err := getNodePoolAvailabilityDomain(tc.np)
			if tc.expectedErr {
				if err == nil {
					t.Fatalf("expected err but not nil")
				}
				return
			}

			if ad != tc.result {
				t.Errorf("got %q ; wanted %q", ad, tc.result)
			}
		})
	}
}

func TestBuildGenericLabels(t *testing.T) {

	shape := "VM.Standard1.2"
	id1 := "ocid1.nodepool.oc1.iad.aaaaaa1"
	np := &oke.NodePool{
		NodeShape: common.String(shape),
		Id:        &id1,
	}

	nodeName := "node1"
	availabilityDomain := "US-ASHBURN-1"
	region := "iad"

	expected := map[string]string{
		kubeletapis.LabelArch:              cloudprovider.DefaultArch,
		apiv1.LabelArchStable:              cloudprovider.DefaultArch,
		kubeletapis.LabelOS:                cloudprovider.DefaultOS,
		apiv1.LabelOSStable:                cloudprovider.DefaultOS,
		apiv1.LabelInstanceType:            shape,
		apiv1.LabelInstanceTypeStable:      shape,
		apiv1.LabelZoneFailureDomain:       availabilityDomain,
		apiv1.LabelZoneFailureDomainStable: availabilityDomain,
		apiv1.LabelHostname:                nodeName,
		apiv1.LabelZoneRegion:              region,
		apiv1.LabelZoneRegionStable:        region,
	}

	output := ocicommon.BuildGenericLabels(*np.Id, nodeName, shape, availabilityDomain)
	if !reflect.DeepEqual(output, expected) {
		t.Fatalf("got %+v\nwanted %+v", output, expected)
	}

	// Make sure labels are set properly for ARM node pool
	armShape := "VM.Standard.A1.Flex"
	id2 := "ocid1.nodepool.oc1.iad.aaaaaa"
	armNp := &oke.NodePool{
		NodeShape: common.String(armShape),
		Id:        &id2,
	}

	armNodeName := "node2"
	availabilityDomain = "US-ASHBURN-1"
	region = "iad"

	armExpected := map[string]string{
		kubeletapis.LabelArch:              consts.ArmArch,
		apiv1.LabelArchStable:              consts.ArmArch,
		kubeletapis.LabelOS:                cloudprovider.DefaultOS,
		apiv1.LabelOSStable:                cloudprovider.DefaultOS,
		apiv1.LabelInstanceType:            armShape,
		apiv1.LabelInstanceTypeStable:      armShape,
		apiv1.LabelZoneFailureDomain:       availabilityDomain,
		apiv1.LabelZoneFailureDomainStable: availabilityDomain,
		apiv1.LabelHostname:                armNodeName,
		apiv1.LabelZoneRegion:              region,
		apiv1.LabelZoneRegionStable:        region,
	}

	armOutput := ocicommon.BuildGenericLabels(*armNp.Id, armNodeName, armShape, availabilityDomain)
	if !reflect.DeepEqual(armOutput, armExpected) {
		t.Fatalf("got %+v\nwanted %+v", armOutput, armExpected)
	}

}

type mockOKEClient struct{}

func (c mockOKEClient) GetNodePool(ctx context.Context, req oke.GetNodePoolRequest) (oke.GetNodePoolResponse, error) {
	return oke.GetNodePoolResponse{
		NodePool: oke.NodePool{
			Id: req.NodePoolId,
			NodeConfigDetails: &oke.NodePoolNodeConfigDetails{
				Size: common.Int(1),
			},
		},
	}, nil
}
func (c mockOKEClient) UpdateNodePool(context.Context, oke.UpdateNodePoolRequest) (oke.UpdateNodePoolResponse, error) {
	return oke.UpdateNodePoolResponse{}, nil
}
func (c mockOKEClient) DeleteNode(context.Context, oke.DeleteNodeRequest) (oke.DeleteNodeResponse, error) {
	return oke.DeleteNodeResponse{
		RawResponse: &http.Response{
			Status:     "200 OK",
			StatusCode: 200,
		},
	}, nil
}

func (c mockOKEClient) ListNodePools(ctx context.Context, req oke.ListNodePoolsRequest) (oke.ListNodePoolsResponse, error) {
	// below test data added for auto-discovery tests
	if req.CompartmentId != nil && *req.CompartmentId == autoDiscoveryCompartment {
		freeformTags1 := map[string]string{
			"ca-managed": "true",
		}
		freeformTags2 := map[string]string{
			"ca-managed": "true",
			"minSize":    "4",
			"maxSize":    "10",
		}
		definedTags := map[string]map[string]interface{}{
			"namespace": {
				"foo": "bar",
			},
		}
		resp := oke.ListNodePoolsResponse{
			Items: []oke.NodePoolSummary{
				{
					Id:           common.String("node-pool-1"),
					FreeformTags: freeformTags1,
					DefinedTags:  definedTags,
				},
				{
					Id:           common.String("node-pool-2"),
					FreeformTags: freeformTags2,
					DefinedTags:  definedTags,
				},
			},
		}
		return resp, nil
	}

	return oke.ListNodePoolsResponse{}, nil
}

func TestRemoveInstance(t *testing.T) {
	instanceId1 := "instance1"
	instanceId2 := "instance2"
	instanceId3 := "instance3"
	instanceId4 := "instance4"
	instanceId5 := "instance5"
	instanceId6 := "instance6"
	nodePoolId := "id"

	expectedInstances := map[string]int{instanceId4: 1, instanceId5: 1, instanceId6: 1}

	nodePoolCache := newNodePoolCache(nil)
	nodePoolCache.okeClient = mockOKEClient{}
	nodePoolCache.cache[nodePoolId] = &oke.NodePool{
		Nodes: []oke.Node{
			{Id: common.String(instanceId1), LifecycleState: oke.NodeLifecycleStateDeleting},
			{Id: common.String(instanceId2), LifecycleState: oke.NodeLifecycleStateDeleted},
			{Id: common.String(instanceId3), LifecycleState: oke.NodeLifecycleStateActive},
			{Id: common.String(instanceId4), LifecycleState: oke.NodeLifecycleStateActive},
			{Id: common.String(instanceId5), LifecycleState: oke.NodeLifecycleStateCreating},
			{Id: common.String(instanceId6), LifecycleState: oke.NodeLifecycleStateUpdating},
		},
	}

	if err := nodePoolCache.removeInstance(nodePoolId, instanceId1, instanceId1); err != nil {
		t.Errorf("Remove instance #{instanceId1} incorrectly")
	}

	if err := nodePoolCache.removeInstance(nodePoolId, instanceId2, instanceId2); err != nil {
		t.Errorf("Remove instance #{instanceId2} incorrectly")
	}

	if err := nodePoolCache.removeInstance(nodePoolId, instanceId3, instanceId3); err != nil {
		t.Errorf("Fail to remove instance #{instanceId3}")
	}

	if err := nodePoolCache.removeInstance(nodePoolId, "", "badNode"); err == nil {
		t.Errorf("Bad node should not have been deleted.")
	}

	if len(nodePoolCache.cache[nodePoolId].Nodes) != 3 {
		t.Errorf("Get incorrect nodes size; expected size is 3")
	}

	for _, nodePool := range nodePoolCache.cache {
		for _, node := range nodePool.Nodes {
			if _, ok := expectedInstances[*node.Id]; !ok {
				t.Errorf("Cannot find the instance %q from node pool cache and it shouldn't be deleted", *node.Id)
			}
		}
	}
}

func TestNodeGroupAutoDiscovery(t *testing.T) {
	var nodeGroupArg = fmt.Sprintf("clusterId:ocid1.cluster.oc1.test-region.test,compartmentId:%s,nodepoolTags:ca-managed=true&namespace.foo=bar,min:1,max:5", autoDiscoveryCompartment)
	nodeGroup, err := nodeGroupFromArg(nodeGroupArg)
	if err != nil {
		t.Errorf("Error: #{err}")
	}
	nodePoolCache := newNodePoolCache(nil)
	nodePoolCache.okeClient = mockOKEClient{}

	cloudConfig := &ocicommon.CloudConfig{}
	cloudConfig.Global.RefreshInterval = 5 * time.Minute
	cloudConfig.Global.CompartmentID = autoDiscoveryCompartment

	manager := &ociManagerImpl{
		nodePoolCache:   nodePoolCache,
		nodeGroups:      []nodeGroupAutoDiscovery{*nodeGroup},
		okeClient:       mockOKEClient{},
		cfg:             cloudConfig,
		staticNodePools: map[string]NodePool{},
	}
	// test data to use as initial nodepools
	nodepool2 := &nodePool{
		id: "node-pool-2", minSize: 1, maxSize: 5,
	}
	manager.staticNodePools[nodepool2.id] = nodepool2
	nodepool3 := &nodePool{
		id: "node-pool-3", minSize: 2, maxSize: 5,
	}
	manager.staticNodePools[nodepool3.id] = nodepool3

	manager.forceRefresh()
}

func TestNodeGroupFromArg(t *testing.T) {
	var nodeGroupArg = fmt.Sprintf("clusterId:ocid1.cluster.oc1.test-region.test,compartmentId:%s,nodepoolTags:ca-managed=true&namespace.foo=bar,min:1,max:5", autoDiscoveryCompartment)
	nodeGroupAutoDiscovery, err := nodeGroupFromArg(nodeGroupArg)
	if err != nil {
		t.Errorf("Error: #{err}")
	}
	if nodeGroupAutoDiscovery.clusterId != "ocid1.cluster.oc1.test-region.test" {
		t.Errorf("Error: clusterId should be ocid1.cluster.oc1.test-region.test")
	}
	if nodeGroupAutoDiscovery.compartmentId != "ocid1.compartment.oc1.test-region.test" {
		t.Errorf("Error: compartmentId should be ocid1.compartment.oc1.test-region.test")
	}
	if nodeGroupAutoDiscovery.minSize != 1 {
		t.Errorf("Error: minSize should be 1")
	}
	if nodeGroupAutoDiscovery.maxSize != 5 {
		t.Errorf("Error: maxSize should be 5")
	}
	if nodeGroupAutoDiscovery.tags["ca-managed"] != "true" {
		t.Errorf("Error: ca-managed:true is missing in tags.")
	}
	if nodeGroupAutoDiscovery.tags["namespace.foo"] != "bar" {
		t.Errorf("Error: namespace.foo:bar is missing in tags.")
	}
}

func TestValidateNodePoolTags(t *testing.T) {

	testCases := map[string]struct {
		nodeGroupTags  map[string]string
		freeFormTags   map[string]string
		definedTags    map[string]map[string]interface{}
		expectedResult bool
	}{
		"no-tags": {
			nodeGroupTags:  nil,
			freeFormTags:   nil,
			definedTags:    nil,
			expectedResult: true,
		},
		"node-group tags provided but no tags on nodepool": {
			nodeGroupTags: map[string]string{
				"testTag": "testTagValue",
			},
			freeFormTags:   nil,
			definedTags:    nil,
			expectedResult: false,
		},
		"node-group tags and free-form tags do not match": {
			nodeGroupTags: map[string]string{
				"testTag": "testTagValue",
			},
			freeFormTags: map[string]string{
				"foo": "bar",
			},
			definedTags:    nil,
			expectedResult: false,
		},
		"free-form tags have required node-group tags": {
			nodeGroupTags: map[string]string{
				"testTag": "testTagValue",
			},
			freeFormTags: map[string]string{
				"foo":     "bar",
				"testTag": "testTagValue",
			},
			definedTags:    nil,
			expectedResult: true,
		},
		"defined tags have required node-group tags": {
			nodeGroupTags: map[string]string{
				"ns.testTag": "testTagValue",
			},
			freeFormTags: nil,
			definedTags: map[string]map[string]interface{}{
				"ns": {
					"testTag": "testTagValue",
				},
			},
			expectedResult: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result := validateNodepoolTags(tc.nodeGroupTags, tc.freeFormTags, tc.definedTags)
			if result != tc.expectedResult {
				t.Errorf("Testcase '%s' failed: got %t ; expected %t", name, result, tc.expectedResult)
			}
		})
	}
}

func TestGetNodePoolNodes_Combined(t *testing.T) {
	tests := []struct {
		name               string
		nodeID             string
		npID               string
		nodeError          *oke.NodeError
		lifecycleState     oke.NodeLifecycleStateEnum
		maxProvisionTime   time.Duration
		creationTimeOffset time.Duration
		wantState          cloudprovider.InstanceState
		wantErrorInfo      *cloudprovider.InstanceErrorInfo
		appendInstance     bool
	}{
		// Updating timed-out with NodeError
		{
			name:               "UpdatingTimedOutWithNodeError",
			nodeID:             "upd-err",
			npID:               "np1",
			lifecycleState:     oke.NodeLifecycleStateUpdating,
			nodeError:          &oke.NodeError{Code: common.String("SomeCode"), Message: common.String("SomeMessage")},
			maxProvisionTime:   10 * time.Minute,
			creationTimeOffset: -1 * time.Hour,
			wantState:          cloudprovider.InstanceCreating,
			wantErrorInfo: &cloudprovider.InstanceErrorInfo{
				ErrorClass:   cloudprovider.OtherErrorClass,
				ErrorCode:    "SomeCode",
				ErrorMessage: "SomeMessage",
			},
			appendInstance: true,
		},
		// Updating timed-out without NodeError
		{
			name:               "UpdatingTimedOutWithoutNodeError",
			nodeID:             "upd-noerr",
			npID:               "np2",
			lifecycleState:     oke.NodeLifecycleStateUpdating,
			maxProvisionTime:   15 * time.Minute,
			creationTimeOffset: -2 * time.Hour,
			wantState:          cloudprovider.InstanceCreating,
			wantErrorInfo: &cloudprovider.InstanceErrorInfo{
				ErrorClass:   cloudprovider.OtherErrorClass,
				ErrorCode:    "MaxNodeProvisionTimeExceeded",
				ErrorMessage: "MaxNodeProvisionTimeExceeded",
			},
			appendInstance: true,
		},
		// Updating not timed-out → no ErrorInfo
		{
			name:               "UpdatingNotTimedOut_NoErrorInfo",
			nodeID:             "upd-notimeout",
			npID:               "np3",
			lifecycleState:     oke.NodeLifecycleStateUpdating,
			nodeError:          &oke.NodeError{Code: common.String("IgnoredCode"), Message: common.String("IgnoredMessage")},
			maxProvisionTime:   1 * time.Hour,
			creationTimeOffset: -1 * time.Minute,
			wantState:          cloudprovider.InstanceCreating,
			wantErrorInfo:      nil,
			appendInstance:     true,
		},
		// Node Deleting
		{
			name:           "DeletingNode",
			nodeID:         "node2",
			npID:           "np5",
			lifecycleState: oke.NodeLifecycleStateDeleting,
			wantState:      cloudprovider.InstanceDeleting,
			wantErrorInfo:  nil,
			appendInstance: true,
		},
		// Node Active
		{
			name:           "ActiveNode",
			nodeID:         "node3",
			npID:           "np5",
			lifecycleState: oke.NodeLifecycleStateActive,
			wantState:      cloudprovider.InstanceRunning,
			wantErrorInfo:  nil,
			appendInstance: true,
		},
		// Node Deleted → should not append
		{
			name:           "DeletedNode",
			nodeID:         "node4",
			npID:           "np6",
			lifecycleState: oke.NodeLifecycleStateDeleted,
			appendInstance: false,
		},
		// NodeError variations
		{
			name:               "NodeError_Unknown",
			nodeID:             "node6",
			npID:               "np5",
			nodeError:          &oke.NodeError{Code: common.String("unknown"), Message: common.String("message")},
			lifecycleState:     oke.NodeLifecycleStateUpdating,
			maxProvisionTime:   1 * time.Minute,
			creationTimeOffset: -10 * time.Hour,
			wantState:          cloudprovider.InstanceCreating,
			wantErrorInfo: &cloudprovider.InstanceErrorInfo{
				ErrorClass:   cloudprovider.OtherErrorClass,
				ErrorCode:    "unknown",
				ErrorMessage: "message",
			},
			appendInstance: true,
		},
		// Edge case: Out-of-Capacity node with nil ID → skipped
		{
			name:           "OutOfCapacityNode_NilID_Skipped",
			nodeID:         "",
			npID:           "np7",
			lifecycleState: oke.NodeLifecycleStateCreating,
			nodeError:      &oke.NodeError{Code: common.String("OutOfCapacity"), Message: common.String("Out of host capacity")},
			appendInstance: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodePoolCache := newNodePoolCache(nil)
			nodePoolCache.cache = make(map[string]*oke.NodePool)

			nodes := []oke.Node{}
			if tt.nodeID != "" {
				nodes = append(nodes, oke.Node{
					Id:             common.String(tt.nodeID),
					LifecycleState: tt.lifecycleState,
					NodeError:      tt.nodeError,
					Name:           common.String(tt.nodeID),
				})
			} else if tt.nodeError != nil {
				// Include node with nil ID but NodeError
				nodes = append(nodes, oke.Node{
					LifecycleState: tt.lifecycleState,
					NodeError:      tt.nodeError,
					Name:           common.String("noId"),
				})
			}

			nodePoolCache.cache[tt.npID] = &oke.NodePool{
				Id:    common.String(tt.npID),
				Nodes: nodes,
			}

			manager := &ociManagerImpl{
				nodePoolCache:             nodePoolCache,
				instanceCreationTimeCache: make(map[string]time.Time),
				maxNodeProvisionTime:      tt.maxProvisionTime,
			}

			if tt.nodeID != "" && tt.creationTimeOffset != 0 {
				manager.SetInstanceCreationTimeInCache(tt.nodeID, time.Now().Add(tt.creationTimeOffset))
			}

			instances, err := manager.GetNodePoolNodes(&nodePool{id: tt.npID})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.appendInstance {
				if len(instances) != 1 {
					t.Fatalf("expected 1 instance, got %d", len(instances))
				}
				inst := instances[0]
				if inst.Id != tt.nodeID {
					t.Errorf("unexpected instance id %q", inst.Id)
				}
				if inst.Status == nil || inst.Status.State != tt.wantState {
					t.Fatalf("expected state %v, got %+v", tt.wantState, inst.Status)
				}
				if tt.wantErrorInfo == nil && inst.Status.ErrorInfo != nil {
					t.Fatalf("expected ErrorInfo to be nil, got %+v", inst.Status.ErrorInfo)
				}
				if tt.wantErrorInfo != nil {
					if inst.Status.ErrorInfo.ErrorClass != tt.wantErrorInfo.ErrorClass {
						t.Errorf("unexpected ErrorClass %v, want %v", inst.Status.ErrorInfo.ErrorClass, tt.wantErrorInfo.ErrorClass)
					}
					if inst.Status.ErrorInfo.ErrorCode != tt.wantErrorInfo.ErrorCode {
						t.Errorf("unexpected ErrorCode %q, want %q", inst.Status.ErrorInfo.ErrorCode, tt.wantErrorInfo.ErrorCode)
					}
					if inst.Status.ErrorInfo.ErrorMessage != tt.wantErrorInfo.ErrorMessage {
						t.Errorf("unexpected ErrorMessage %q, want %q", inst.Status.ErrorInfo.ErrorMessage, tt.wantErrorInfo.ErrorMessage)
					}
				}
			} else {
				if len(instances) != 0 {
					t.Fatalf("expected 0 instances, got %d", len(instances))
				}
			}
		})
	}
}
