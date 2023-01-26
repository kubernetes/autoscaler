package oci

import (
	"context"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"reflect"
	"strings"
	"testing"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/oci-go-sdk/v43/common"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/oci-go-sdk/v43/core"
)

type mockShapeClient struct {
	err                   error
	listShapeResp         core.ListShapesResponse
	getInstanceConfigResp core.GetInstanceConfigurationResponse
}

func (m *mockShapeClient) ListShapes(_ context.Context, _ core.ListShapesRequest) (core.ListShapesResponse, error) {
	return m.listShapeResp, m.err
}

func (m *mockShapeClient) GetInstanceConfiguration(context.Context, core.GetInstanceConfigurationRequest) (core.GetInstanceConfigurationResponse, error) {
	return m.getInstanceConfigResp, m.err
}

var launchDetails = core.InstanceConfigurationLaunchInstanceDetails{
	CompartmentId:     nil,
	DisplayName:       nil,
	CreateVnicDetails: nil,
	Shape:             common.String("VM.Standard.E3.Flex"),
	ShapeConfig: &core.InstanceConfigurationLaunchInstanceShapeConfigDetails{
		Ocpus:       common.Float32(8),
		MemoryInGBs: common.Float32(128),
	},
	SourceDetails: nil,
}
var instanceDetails = core.ComputeInstanceDetails{
	LaunchDetails: &launchDetails,
}

var shapeClient = &mockShapeClient{
	err: nil,
	listShapeResp: core.ListShapesResponse{
		Items: []core.Shape{
			{
				Shape:       common.String("VM.Standard2.8"),
				Ocpus:       common.Float32(8),
				MemoryInGBs: common.Float32(120),
			},
		},
	},
	getInstanceConfigResp: core.GetInstanceConfigurationResponse{
		RawResponse: nil,
		InstanceConfiguration: core.InstanceConfiguration{
			CompartmentId:   nil,
			Id:              common.String("ocid1.instanceconfiguration.oc1.phx.aaaaaaaa1"),
			TimeCreated:     nil,
			DefinedTags:     nil,
			DisplayName:     nil,
			FreeformTags:    nil,
			InstanceDetails: instanceDetails,
			DeferredFields:  nil,
		},
		Etag:         nil,
		OpcRequestId: nil,
	},
}

func TestGetShape(t *testing.T) {

	testCases := map[string]struct {
		shape    string
		expected *Shape
	}{
		"flex shape": {
			shape: "VM.Standard.E3.Flex",
			expected: &Shape{
				Name:          "VM.Standard.E3.Flex",
				CPU:           8,
				MemoryInBytes: float32(128) * 1024 * 1024 * 1024,
				GPU:           0,
			},
		},
	}

	for name, tc := range testCases {
		shapeGetter := createShapeGetter(shapeClient)

		t.Run(name, func(t *testing.T) {
			shape, err := shapeGetter.GetInstancePoolShape(&core.InstancePool{Id: &tc.shape, InstanceConfigurationId: common.String("ocid1.instanceconfiguration.oc1.phx.aaaaaaaa1")})
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(shape, tc.expected) {
				t.Errorf("wanted %+v ; got %+v", tc.expected, shape)
			}

			if !strings.Contains(tc.shape, "Flex") {
				// we can't poolCache flex shapes so only check poolCache on non flex shapes
				cacheShape, ok := shapeGetter.(*shapeGetterImpl).cache[tc.shape]
				if !ok {
					t.Error("shape not found in poolCache")
				}

				if !reflect.DeepEqual(cacheShape, tc.expected) {
					t.Errorf("wanted %+v ; got %+v", tc.expected, shape)
				}
			}

		})
	}
}

func TestGetInstancePoolTemplateNode(t *testing.T) {
	instancePoolCache := newInstancePoolCache(computeManagementClient, computeClient, virtualNetworkClient, workRequestsClient)
	instancePoolCache.poolCache["ocid1.instancepool.oc1.phx.aaaaaaaa1"] = &core.InstancePool{
		Id:             common.String("ocid1.instancepool.oc1.phx.aaaaaaaa1"),
		CompartmentId:  common.String("ocid1.compartment.oc1..aaaaaaaa1"),
		LifecycleState: core.InstancePoolLifecycleStateRunning,
		PlacementConfigurations: []core.InstancePoolPlacementConfiguration{{
			AvailabilityDomain: common.String("hash:US-ASHBURN-1"),
			PrimarySubnetId:    common.String("ocid1.subnet.oc1.phx.aaaaaaaa1"),
		}},
	}
	instancePoolCache.instanceSummaryCache["ocid1.instancepool.oc1.phx.aaaaaaaa1"] = &[]core.InstanceSummary{{
		Id:                 common.String("ocid1.instance.oc1.phx.aaa1"),
		AvailabilityDomain: common.String("PHX-AD-2"),
		State:              common.String(string(core.InstanceLifecycleStateRunning)),
	},
	}

	cloudConfig := &CloudConfig{}
	cloudConfig.Global.CompartmentID = "ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	var manager = &InstancePoolManagerImpl{
		cfg:         cloudConfig,
		shapeGetter: createShapeGetter(shapeClient),
		staticInstancePools: map[string]*InstancePoolNodeGroup{
			"ocid1.instancepool.oc1.phx.aaaaaaaa1": {id: "ocid1.instancepool.oc1.phx.aaaaaaaa1"},
		},
		instancePoolCache: instancePoolCache,
	}

	instancePoolNodeGroups := manager.GetInstancePools()

	if got := len(instancePoolNodeGroups); got != 1 {
		t.Fatalf("expected 1 (static) instance pool, got %d", got)
	}
	nodeTemplate, err := manager.GetInstancePoolTemplateNode(*instancePoolNodeGroups[0])
	if err != nil {
		t.Fatalf("received unexpected error refreshing cache; %+v", err)
	}
	labels := nodeTemplate.GetLabels()
	if labels == nil {
		t.Fatalf("expected labels on node object")
	}
	// Double check the shape label.
	if got := labels[apiv1.LabelInstanceTypeStable]; got != "VM.Standard.E3.Flex" {
		t.Fatalf("expected shape label %s to be set to VM.Standard.E3.Flex: %v", apiv1.LabelInstanceTypeStable, nodeTemplate.Labels)
	}

	// Also check the AD label for good measure.
	if got := labels[apiv1.LabelTopologyZone]; got != "US-ASHBURN-1" {
		t.Fatalf("expected AD zone label %s to be set to US-ASHBURN-1: %v", apiv1.LabelTopologyZone, nodeTemplate.Labels)
	}

}

func TestBuildGenericLabels(t *testing.T) {

	shapeName := "VM.Standard2.8"
	np := &core.InstancePool{
		Id:   common.String("ocid1.instancepool.oc1.phx.aaaaaaaah"),
		Size: common.Int(2),
	}

	nodeName := "node1"
	availabilityDomain := "US-ASHBURN-1"

	expected := map[string]string{
		apiv1.LabelArchStable:              cloudprovider.DefaultArch,
		apiv1.LabelOSStable:                cloudprovider.DefaultOS,
		apiv1.LabelZoneRegion:              "phx",
		apiv1.LabelZoneRegionStable:        "phx",
		apiv1.LabelInstanceType:            shapeName,
		apiv1.LabelInstanceTypeStable:      shapeName,
		apiv1.LabelZoneFailureDomain:       availabilityDomain,
		apiv1.LabelZoneFailureDomainStable: availabilityDomain,
		apiv1.LabelHostname:                nodeName,
	}

	launchDetails := core.InstanceConfigurationLaunchInstanceDetails{
		Shape: common.String("VM.Standard2.8"),
	}

	instanceDetails := core.ComputeInstanceDetails{
		LaunchDetails: &launchDetails,
	}

	// For list shapes
	mockShapeClient := &mockShapeClient{
		err: nil,
		listShapeResp: core.ListShapesResponse{
			Items: []core.Shape{
				{Shape: common.String("VM.Standard2.4"), Ocpus: common.Float32(4), MemoryInGBs: common.Float32(60)},
				{Shape: common.String("VM.Standard2.8"), Ocpus: common.Float32(8), MemoryInGBs: common.Float32(120)}},
		},
		getInstanceConfigResp: core.GetInstanceConfigurationResponse{
			InstanceConfiguration: core.InstanceConfiguration{
				Id:              common.String("ocid1.instanceconfiguration.oc1.phx.aaaaaaaa1"),
				InstanceDetails: instanceDetails,
			},
		},
	}
	shapeGetter := createShapeGetter(mockShapeClient)

	manager := InstancePoolManagerImpl{
		shapeGetter: shapeGetter,
	}

	shape, err := manager.shapeGetter.GetInstancePoolShape(np)
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	output := buildGenericLabelsForInstancePool(np, nodeName, shape.Name, availabilityDomain)
	if !reflect.DeepEqual(output, expected) {
		t.Fatalf("got %+v\nwanted %+v", output, expected)
	}

}
