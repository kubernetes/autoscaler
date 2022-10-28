package oci

import (
	"context"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/oci-go-sdk/v43/core"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/oci-go-sdk/v43/workrequests"
	"reflect"
	"testing"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/oci-go-sdk/v43/common"
)

type mockComputeManagementClient struct {
	err                                error
	getInstancePoolResponse            core.GetInstancePoolResponse
	getInstancePoolInstanceResponse    core.GetInstancePoolInstanceResponse
	listInstancePoolInstancesResponse  core.ListInstancePoolInstancesResponse
	updateInstancePoolResponse         core.UpdateInstancePoolResponse
	detachInstancePoolInstanceResponse core.DetachInstancePoolInstanceResponse
}

type mockVirtualNetworkClient struct {
	err             error
	getVnicResponse core.GetVnicResponse
}

type mockComputeClient struct {
	err                         error
	listVnicAttachmentsResponse core.ListVnicAttachmentsResponse
}

type mockWorkRequestClient struct {
	err error
}

func (m *mockWorkRequestClient) GetWorkRequest(ctx context.Context, request workrequests.GetWorkRequestRequest) (workrequests.GetWorkRequestResponse, error) {
	return workrequests.GetWorkRequestResponse{}, m.err
}

func (m *mockWorkRequestClient) ListWorkRequests(ctx context.Context, request workrequests.ListWorkRequestsRequest) (workrequests.ListWorkRequestsResponse, error) {
	return workrequests.ListWorkRequestsResponse{}, m.err
}

func (m *mockWorkRequestClient) ListWorkRequestErrors(ctx context.Context, request workrequests.ListWorkRequestErrorsRequest) (workrequests.ListWorkRequestErrorsResponse, error) {
	return workrequests.ListWorkRequestErrorsResponse{}, m.err
}

func (m *mockComputeClient) ListVnicAttachments(ctx context.Context, request core.ListVnicAttachmentsRequest) (core.ListVnicAttachmentsResponse, error) {
	return m.listVnicAttachmentsResponse, m.err
}

func (m *mockVirtualNetworkClient) GetVnic(context.Context, core.GetVnicRequest) (core.GetVnicResponse, error) {
	return m.getVnicResponse, m.err
}

func (m *mockComputeManagementClient) ListInstancePoolInstances(_ context.Context, _ core.ListInstancePoolInstancesRequest) (core.ListInstancePoolInstancesResponse, error) {
	return m.listInstancePoolInstancesResponse, m.err
}

func (m *mockComputeManagementClient) GetInstancePool(context.Context, core.GetInstancePoolRequest) (core.GetInstancePoolResponse, error) {
	return m.getInstancePoolResponse, m.err
}

func (m *mockComputeManagementClient) UpdateInstancePool(context.Context, core.UpdateInstancePoolRequest) (core.UpdateInstancePoolResponse, error) {
	return m.updateInstancePoolResponse, m.err
}

func (m *mockComputeManagementClient) GetInstancePoolInstance(context.Context, core.GetInstancePoolInstanceRequest) (core.GetInstancePoolInstanceResponse, error) {
	return m.getInstancePoolInstanceResponse, m.err
}

func (m *mockComputeManagementClient) DetachInstancePoolInstance(context.Context, core.DetachInstancePoolInstanceRequest) (core.DetachInstancePoolInstanceResponse, error) {
	return m.detachInstancePoolInstanceResponse, m.err
}

var computeClient = &mockComputeClient{
	err: nil,
	listVnicAttachmentsResponse: core.ListVnicAttachmentsResponse{
		RawResponse: nil,
		Items: []core.VnicAttachment{{
			Id:             common.String("ocid1.vnic.oc1.phx.abc"),
			LifecycleState: core.VnicAttachmentLifecycleStateAttached,
		}},
	},
}

var computeManagementClient = &mockComputeManagementClient{
	err: nil,
	getInstancePoolResponse: core.GetInstancePoolResponse{
		InstancePool: core.InstancePool{
			Id:                      common.String("ocid1.instancepool.oc1.phx.aaaaaaaa1"),
			CompartmentId:           common.String("ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
			InstanceConfigurationId: common.String("ocid1.instanceconfiguration.oc1.phx.aaaaaaaa1"),
			LifecycleState:          core.InstancePoolLifecycleStateRunning,
			Size:                    common.Int(2),
		},
	},
	listInstancePoolInstancesResponse: core.ListInstancePoolInstancesResponse{
		RawResponse: nil,
		Items: []core.InstanceSummary{{
			Id:                 common.String("ocid1.instance.oc1.phx.aaa1"),
			AvailabilityDomain: common.String("Uocm:PHX-AD-2"),
			CompartmentId:      common.String("ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
			DisplayName:        common.String("inst-1ncvn-ociinstancepool"),
			Shape:              common.String("VM.Standard2.8"),
			State:              common.String(string(core.InstanceLifecycleStateRunning)),
		}, {
			Id:                 common.String("ocid1.instance.oc1.phx.aaacachemiss"),
			AvailabilityDomain: common.String("Uocm:PHX-AD-2"),
			CompartmentId:      common.String("ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
			DisplayName:        common.String("inst-2ncvn-ociinstancepool"),
			Shape:              common.String("VM.Standard2.8"),
			State:              common.String(string(core.InstanceLifecycleStateRunning)),
		}},
	},
}

var virtualNetworkClient = &mockVirtualNetworkClient{
	err: nil,
	getVnicResponse: core.GetVnicResponse{
		RawResponse: nil,
		Vnic: core.Vnic{
			Id:        common.String("ocid1.vnic.oc1.phx.abyhqljsxigued23s7ywgcqlbpqfiysgnhxj672awzjluhoopzf7l7wvm6rq"),
			PrivateIp: common.String("10.0.20.59"),
			PublicIp:  common.String("129.146.58.250"),
		},
	},
}

var workRequestsClient = &mockWorkRequestClient{
	err: nil,
}

func TestInstancePoolFromArgs(t *testing.T) {

	value := `1:5:ocid1.instancepool.oc1.phx.aaaaaaaah`
	instanceNodePool, err := instancePoolFromArg(value)
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	if instanceNodePool.minSize != 1 {
		t.Errorf("got minSize %d ; wanted minSize 1", instanceNodePool.minSize)
	}

	if instanceNodePool.maxSize != 5 {
		t.Errorf("got maxSize %d ; wanted maxSize 1", instanceNodePool.maxSize)
	}

	if instanceNodePool.id != "ocid1.instancepool.oc1.phx.aaaaaaaah" {
		t.Errorf("got ocid %q ; wanted id \"ocid1.instancepool.oc1.phx.aaaaaaaah\"", instanceNodePool.id)
	}

	value = `1:5:ocid1.nodepool.oc1.phx.aaaaaaaah`
	_, err = instancePoolFromArg(value)
	if err == nil {
		t.Fatal("expected error processing an oke based node-pool")
	}

	value = `1:5:incorrect:ocid1.instancepool.oc1.phx.aaaaaaaah`
	_, err = instancePoolFromArg(value)
	if err == nil {
		t.Fatal("expected error of an invalid instance pool")
	}
}

func TestGetSetInstancePoolSize(t *testing.T) {

	nodePoolCache := newInstancePoolCache(computeManagementClient, computeClient, virtualNetworkClient, workRequestsClient)
	nodePoolCache.poolCache["ocid1.instancepool.oc1.phx.aaaaaaaai"] = &core.InstancePool{Size: common.Int(2)}

	manager := &InstancePoolManagerImpl{instancePoolCache: nodePoolCache}
	size, err := manager.GetInstancePoolSize(InstancePoolNodeGroup{id: "ocid1.instancepool.oc1.phx.aaaaaaaai"})
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}
	if size != 2 {
		t.Errorf("got size %d ; wanted size 5", size)
	}

	computeManagementClient.listInstancePoolInstancesResponse.Items = append(computeManagementClient.listInstancePoolInstancesResponse.Items, core.InstanceSummary{
		Id:                 common.String("ocid1.instance.oc1.phx.newInstance"),
		AvailabilityDomain: common.String("Uocm:PHX-AD-2"),
		CompartmentId:      common.String("ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
		DisplayName:        common.String("inst-1ncvn-ociinstancepool"),
		Shape:              common.String("VM.Standard2.8"),
		State:              common.String(string(core.InstanceLifecycleStateRunning)),
	})

	err = manager.SetInstancePoolSize(InstancePoolNodeGroup{id: "ocid1.instancepool.oc1.phx.aaaaaaaai"}, 3)
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}
	size, err = manager.GetInstancePoolSize(InstancePoolNodeGroup{id: "ocid1.instancepool.oc1.phx.aaaaaaaai"})
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}
	if size != 3 {
		t.Errorf("got size %d ; wanted size 6", size)
	}

}

func TestGetInstancePoolForInstance(t *testing.T) {

	nodePoolCache := newInstancePoolCache(computeManagementClient, computeClient, virtualNetworkClient, workRequestsClient)
	nodePoolCache.poolCache["ocid1.instancepool.oc1.phx.aaaaaaaa1"] = &core.InstancePool{
		Id:   common.String("ocid1.instancepool.oc1.phx.aaaaaaaa1"),
		Size: common.Int(1),
	}

	var cloudConfig = CloudConfig{}
	cloudConfig.Global.CompartmentID = "compartment.oc1..aaaaaaaa7ey4sg3a6b5wnv5hlkjlkjadslkfjalskfjalsadfadsf"

	manager := &InstancePoolManagerImpl{
		cfg: &cloudConfig,
		staticInstancePools: map[string]*InstancePoolNodeGroup{
			"ocid1.instancepool.oc1.phx.aaaaaaaa1": {id: "ocid1.instancepool.oc1.phx.aaaaaaaa1"},
			"ocid1.instancepool.oc1.phx.aaaaaaaa2": {id: "ocid1.instancepool.oc1.phx.aaaaaaaa2"},
		},
		instancePoolCache: nodePoolCache,
	}

	// first verify instance pool can be found when only the instance id is specified.
	np, err := manager.GetInstancePoolForInstance(OciRef{InstanceID: "ocid1.instance.oc1.phx.aaa1"})
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	if np.Id() != "ocid1.instancepool.oc1.phx.aaaaaaaa1" {
		t.Fatalf("got unexpected ocid %q ; wanted \"ocid1.instancepool.oc1.phx.aaaaaaaa1\"", np.Id())
	}

	// next, verify a valid instance can be found if it is currently missing from the cache.
	np, err = manager.GetInstancePoolForInstance(OciRef{InstanceID: "ocid1.instance.oc1.phx.aaacachemiss"})
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	if np.Id() != "ocid1.instancepool.oc1.phx.aaaaaaaa1" {
		t.Fatalf("got unexpected ocid %q ; wanted \"ocid1.instancepool.oc1.phx.aaaaaaaa1s\"", np.Id())
	}

	// next, verify an invalid instance cant be found if it is missing from the cache and pool.
	_, err = manager.GetInstancePoolForInstance(OciRef{InstanceID: "ocid1.instance.oc1.phx.aaadne"})
	if err != errInstanceInstancePoolNotFound {
		t.Fatalf("epected error looking for an invalid instance")
	}

	// verify an invalid instance pool produces an error.
	ip, err := manager.GetInstancePoolForInstance(OciRef{InstanceID: "ocid1.instance.oc1.phx.aaadne", PoolID: "ocid1.instancepool.oc1.phx.aaaaaaaadne"})
	if err == nil {
		t.Fatalf("expected error looking for an instance with invalid instance & pool ids")
	}
	if ip != nil {
		t.Fatalf("expected nil looking for an instance with invalid instance & pool ids")
	}

	// next verify instance pool can be found when the instance pool id is specified directly.
	_, err = manager.GetInstancePoolForInstance(OciRef{PoolID: "ocid1.instancepool.oc1.phx.aaaaaaaa1", PrivateIPAddress: "10.0.20.59"})
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	// next verify instance pool can be found when only the private IP is specified.
	np, err = manager.GetInstancePoolForInstance(OciRef{
		PrivateIPAddress: "10.0.20.59"})
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	if np.Id() != "ocid1.instancepool.oc1.phx.aaaaaaaa1" {
		t.Fatalf("got unexpected ocid %q ; wanted \"ocid1.instancepool.oc1.phx.aaaaaaaa1\"", np.Id())
	}

	// now verify node pool can be found via lookup up by instance id in poolCache
	np, err = manager.GetInstancePoolForInstance(OciRef{InstanceID: "ocid1.instance.oc1.phx.aaa1"})
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	if np.Id() != "ocid1.instancepool.oc1.phx.aaaaaaaa1" {
		t.Fatalf("got unexpected ocid %q ; wanted \"ocid1.instancepool.oc1.phx.aaaaaaaa1\"", np.Id())
	}

}

func TestGetInstancePoolNodes(t *testing.T) {

	nodePoolCache := newInstancePoolCache(computeManagementClient, computeClient, virtualNetworkClient, workRequestsClient)
	nodePoolCache.poolCache["ocid1.instancepool.oc1.phx.aaaaaaaa1"] = &core.InstancePool{
		Id:             common.String("ocid1.instancepool.oc1.phx.aaaaaaaa1"),
		CompartmentId:  common.String("ocid1.compartment.oc1..aaaaaaaa1"),
		LifecycleState: core.InstancePoolLifecycleStateRunning,
	}
	nodePoolCache.instanceSummaryCache["ocid1.instancepool.oc1.phx.aaaaaaaa1"] = &[]core.InstanceSummary{{
		Id:                 common.String("ocid1.instance.oc1.phx.aaa1"),
		AvailabilityDomain: common.String("PHX-AD-2"),
		State:              common.String(string(core.InstanceLifecycleStateRunning)),
	}, {
		Id:                 common.String("ocid1.instance.oc1.phx.aaa2"),
		AvailabilityDomain: common.String("PHX-AD-1"),
		State:              common.String(string(core.InstanceLifecycleStateTerminating)),
	},
	}

	expected := []cloudprovider.Instance{
		{
			Id: "ocid1.instance.oc1.phx.aaa1",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceRunning,
			},
		},
		{
			Id: "ocid1.instance.oc1.phx.aaa2",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceDeleting,
			},
		},
	}

	manager := &InstancePoolManagerImpl{instancePoolCache: nodePoolCache, cfg: &CloudConfig{}}
	manager.shapeGetter = createShapeGetter(shapeClient)
	instances, err := manager.GetInstancePoolNodes(InstancePoolNodeGroup{id: "ocid1.instancepool.oc1.phx.aaaaaaaa1"})
	if err != nil {
		t.Fatalf("received unexpected error; %+v", err)
	}

	if !reflect.DeepEqual(instances, expected) {
		t.Errorf("got %+v\nwanted %+v", instances, expected)
	}

	err = manager.forceRefresh()
	if err != nil {
		t.Fatalf("received unexpected error refreshing cache; %+v", err)
	}
}

func TestGetInstancePoolAvailabilityDomain(t *testing.T) {
	testCases := map[string]struct {
		np          *core.InstancePool
		result      string
		expectedErr bool
	}{
		"single ad": {
			np: &core.InstancePool{
				Id:             common.String("id"),
				LifecycleState: "",
				PlacementConfigurations: []core.InstancePoolPlacementConfiguration{{
					AvailabilityDomain: common.String("hash:US-ASHBURN-1"),
					PrimarySubnetId:    common.String("ocid1.subnet.oc1.phx.aaaaaaaa1"),
				}},
				Size: common.Int(2),
			},
			result: "US-ASHBURN-1",
		},
		"multi-ad": {
			np: &core.InstancePool{
				Id:             common.String("id"),
				LifecycleState: "",
				PlacementConfigurations: []core.InstancePoolPlacementConfiguration{{
					AvailabilityDomain: common.String("hash:US-ASHBURN-1"),
					PrimarySubnetId:    common.String("ocid1.subnet.oc1.phx.aaaaaaaa1"),
				}, {
					AvailabilityDomain: common.String("hash:US-ASHBURN-2"),
					PrimarySubnetId:    common.String("ocid1.subnet.oc1.phx.aaaaaaaa2"),
				}},
				Size: common.Int(2),
			},
			result: "US-ASHBURN-1",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ad, err := getInstancePoolAvailabilityDomain(tc.np)
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

func TestGetInstancePoolsAndInstances(t *testing.T) {

	var computeManagementClient = &mockComputeManagementClient{
		getInstancePoolResponse: core.GetInstancePoolResponse{
			InstancePool: core.InstancePool{
				Id:                      common.String("ocid1.instancepool.oc1.phx.aaaaaaaa1"),
				CompartmentId:           common.String("ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
				InstanceConfigurationId: common.String("ocid1.instanceconfiguration.oc1.phx.aaaaaaaa1"),
				PlacementConfigurations: nil,
				Size:                    common.Int(2),
			},
		},
		listInstancePoolInstancesResponse: core.ListInstancePoolInstancesResponse{
			Items: []core.InstanceSummary{{
				Id:                 common.String("ocid1.instance.oc1.phx.aaa1"),
				AvailabilityDomain: common.String("Uocm:PHX-AD-2"),
				CompartmentId:      common.String("ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
				DisplayName:        common.String("inst-1ncvn-ociinstancepool"),
				Shape:              common.String("VM.Standard2.8"),
				State:              common.String(string(core.InstanceLifecycleStateRunning)),
			}, {
				Id:                 common.String("ocid1.instance.oc1.phx.aaaterminal"),
				AvailabilityDomain: common.String("Uocm:PHX-AD-2"),
				CompartmentId:      common.String("ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
				DisplayName:        common.String("inst-2ncvn-ociinstancepool"),
				Shape:              common.String("VM.Standard2.8"),
				State:              common.String(string(core.InstanceLifecycleStateTerminated)),
			}},
		},
	}

	cloudConfig := &CloudConfig{}
	cloudConfig.Global.CompartmentID = "ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	manager := &InstancePoolManagerImpl{
		cfg: cloudConfig,
		staticInstancePools: map[string]*InstancePoolNodeGroup{
			"ocid1.instancepool.oc1.phx.aaaaaaaa1": {id: "ocid1.instancepool.oc1.phx.aaaaaaaa1"},
		},
		instancePoolCache: newInstancePoolCache(computeManagementClient, computeClient, virtualNetworkClient, workRequestsClient),
	}

	// Populate cache(s) (twice to increase code coverage).
	manager.shapeGetter = createShapeGetter(shapeClient)
	_ = manager.Refresh()
	err := manager.Refresh()
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}
	err = manager.forceRefresh()
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	instancePoolNodeGroups := manager.GetInstancePools()
	if got := len(instancePoolNodeGroups); got != 1 {
		t.Fatalf("expected 1 (static) instance pool, got %d", got)
	}
	instances, err := manager.GetInstancePoolNodes(*instancePoolNodeGroups[0])
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}
	// Should not pick up terminated instance.
	if got := len(instances); got != 1 {
		t.Fatalf("expected 1 instance, got %d", got)
	}

	instancePoolNodeGroup, err := manager.GetInstancePoolForInstance(OciRef{InstanceID: instances[0].Id})
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}
	if !reflect.DeepEqual(instancePoolNodeGroup, instancePoolNodeGroups[0]) {
		t.Errorf("got %+v\nwanted %+v", instancePoolNodeGroup, instancePoolNodeGroups[0])
	}
}

func TestDeleteInstances(t *testing.T) {

	var computeManagementClient = &mockComputeManagementClient{
		getInstancePoolResponse: core.GetInstancePoolResponse{
			InstancePool: core.InstancePool{
				Id:                      common.String("ocid1.instancepool.oc1.phx.aaaaaaaa1"),
				CompartmentId:           common.String("ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
				InstanceConfigurationId: common.String("ocid1.instanceconfiguration.oc1.phx.aaaaaaaa1"),
				LifecycleState:          core.InstancePoolLifecycleStateRunning,
				Size:                    common.Int(2),
			},
		},
		listInstancePoolInstancesResponse: core.ListInstancePoolInstancesResponse{
			Items: []core.InstanceSummary{{
				Id:                 common.String("ocid1.instance.oc1.phx.aaa1"),
				AvailabilityDomain: common.String("Uocm:PHX-AD-1"),
				CompartmentId:      common.String("ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
				DisplayName:        common.String("inst-1ncvn-ociinstancepool"),
				Shape:              common.String("VM.Standard2.16"),
				State:              common.String(string(core.InstanceLifecycleStateRunning)),
			}, {
				Id:                 common.String("ocid1.instance.oc1.phx.aaa2"),
				AvailabilityDomain: common.String("Uocm:PHX-AD-1"),
				CompartmentId:      common.String("ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
				DisplayName:        common.String("inst-2ncvn-ociinstancepool"),
				Shape:              common.String("VM.Standard2.16"),
				State:              common.String(string(core.InstanceLifecycleStateRunning)),
			}},
		},
	}

	cloudConfig := &CloudConfig{}
	cloudConfig.Global.CompartmentID = "ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	manager := &InstancePoolManagerImpl{
		cfg: cloudConfig,
		staticInstancePools: map[string]*InstancePoolNodeGroup{
			"ocid1.instancepool.oc1.phx.aaaaaaaa1": {id: "ocid1.instancepool.oc1.phx.aaaaaaaa1"},
		},
		instancePoolCache: newInstancePoolCache(computeManagementClient, computeClient, virtualNetworkClient, workRequestsClient),
	}
	manager.shapeGetter = createShapeGetter(shapeClient)
	// Populate cache(s).
	manager.Refresh()

	instances, err := manager.GetInstancePoolNodes(InstancePoolNodeGroup{id: "ocid1.instancepool.oc1.phx.aaaaaaaa1"})
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}
	// Should not pick up terminated instance.
	if got := len(instances); got != 2 {
		t.Fatalf("expected 2 instance, got %d", got)
	}
	// Check size before and after delete
	size, err := manager.GetInstancePoolSize(InstancePoolNodeGroup{id: "ocid1.instancepool.oc1.phx.aaaaaaaa1"})
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}
	if size != 2 {
		t.Errorf("got size %d ; wanted size 2 before delete", size)
	}

	instanceToDelete := OciRef{
		AvailabilityDomain: "PHX-AD-1",
		Name:               "inst-2ncvn-ociinstancepool",
		InstanceID:         "ocid1.instance.oc1.phx.aaa2",
		PoolID:             "ocid1.instancepool.oc1.phx.aaaaaaaa1",
	}
	err = manager.DeleteInstances(InstancePoolNodeGroup{id: "ocid1.instancepool.oc1.phx.aaaaaaaa1"}, []OciRef{instanceToDelete})
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}
	size, err = manager.GetInstancePoolSize(InstancePoolNodeGroup{id: "ocid1.instancepool.oc1.phx.aaaaaaaa1"})
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}
	if size != 1 {
		t.Errorf("got size %d ; wanted size 1 *after* delete", size)
	}
}
