/*
Copyright 2020-2023 Oracle and/or its affiliates.
*/

package nodepools

import (
	"context"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/nodepools/consts"
	"net/http"
	"reflect"
	"testing"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	kubeletapis "k8s.io/kubelet/pkg/apis"

	ocicommon "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/common"
	oke "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/containerengine"
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

func TestGetNodePoolNodes(t *testing.T) {
	nodePoolCache := newNodePoolCache(nil)
	nodePoolCache.cache["id"] = &oke.NodePool{
		Nodes: []oke.Node{
			{Id: common.String("node1"), LifecycleState: oke.NodeLifecycleStateDeleted},
			{Id: common.String("node2"), LifecycleState: oke.NodeLifecycleStateDeleting},
			{Id: common.String("node3"), LifecycleState: oke.NodeLifecycleStateActive},
			{Id: common.String("node4"), LifecycleState: oke.NodeLifecycleStateCreating},
			{Id: common.String("node5"), LifecycleState: oke.NodeLifecycleStateUpdating},
			{
				Id: common.String("node6"),
				NodeError: &oke.NodeError{
					Code:    common.String("unknown"),
					Message: common.String("message"),
				},
			},
			{
				Id: common.String("node7"),
				NodeError: &oke.NodeError{
					Code:    common.String("LimitExceeded"),
					Message: common.String("message"),
				},
			},
			{
				Id: common.String("node8"),
				NodeError: &oke.NodeError{
					Code:    common.String("InternalServerError"),
					Message: common.String("blah blah quota exceeded blah blah"),
				},
			},
		},
	}

	expected := []cloudprovider.Instance{
		{
			Id: "node2",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceDeleting,
			},
		},
		{
			Id: "node3",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceRunning,
			},
		},
		{
			Id: "node4",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceCreating,
			},
		},
		{
			Id: "node5",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceCreating,
			},
		},
		{
			Id: "node6",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceCreating,
				ErrorInfo: &cloudprovider.InstanceErrorInfo{
					ErrorClass:   cloudprovider.OtherErrorClass,
					ErrorCode:    "unknown",
					ErrorMessage: "message",
				},
			},
		},
		{
			Id: "node7",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceCreating,
				ErrorInfo: &cloudprovider.InstanceErrorInfo{
					ErrorClass:   cloudprovider.OutOfResourcesErrorClass,
					ErrorCode:    "LimitExceeded",
					ErrorMessage: "message",
				},
			},
		},
		{
			Id: "node8",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceCreating,
				ErrorInfo: &cloudprovider.InstanceErrorInfo{
					ErrorClass:   cloudprovider.OutOfResourcesErrorClass,
					ErrorCode:    "InternalServerError",
					ErrorMessage: "blah blah quota exceeded blah blah",
				},
			},
		},
	}

	manager := &ociManagerImpl{nodePoolCache: nodePoolCache}
	instances, err := manager.GetNodePoolNodes(&nodePool{id: "id"})
	if err != nil {
		t.Fatalf("received unexpected error; %+v", err)
	}

	if !reflect.DeepEqual(instances, expected) {
		t.Errorf("got %+v\nwanted %+v", instances, expected)
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

func (c mockOKEClient) GetNodePool(context.Context, oke.GetNodePoolRequest) (oke.GetNodePoolResponse, error) {
	return oke.GetNodePoolResponse{}, nil
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

func (c mockOKEClient) ListNodePools(context.Context, oke.ListNodePoolsRequest) (oke.ListNodePoolsResponse, error) {
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

func TestNodeGroupFromArg(t *testing.T) {
	var nodeGroupArg = "clusterId:ocid1.cluster.oc1.test-region.test,compartmentId:ocid1.compartment.oc1.test-region.test,nodepoolTags:ca-managed=true&namespace.foo=bar,min:1,max:5"
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
