/*
Copyright 2021-2023 Oracle and/or its affiliates.
*/

package common

import (
	"context"
	oke "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/containerengine"
	"reflect"
	"strings"
	"testing"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/core"
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

func TestNodePoolGetShape(t *testing.T) {

	shapeClient := &mockShapeClient{
		err: nil,
		listShapeResp: core.ListShapesResponse{
			Items: []core.Shape{
				{
					Shape:       common.String("VM.Standard1.2"),
					Ocpus:       common.Float32(2),
					MemoryInGBs: common.Float32(16),
				},
			},
		},
	}

	testCases := map[string]struct {
		shape       string
		shapeConfig *oke.NodeShapeConfig
		expected    *Shape
	}{
		"basic shape": {
			shape: "VM.Standard1.2",
			expected: &Shape{
				CPU:                     4,
				MemoryInBytes:           16 * 1024 * 1024 * 1024,
				GPU:                     0,
				EphemeralStorageInBytes: -1,
			},
		},
		"flex shape": {
			shape: "VM.Standard.E3.Flex",
			shapeConfig: &oke.NodeShapeConfig{
				Ocpus:       common.Float32(4),
				MemoryInGBs: common.Float32(64),
			},
			expected: &Shape{
				CPU:                     8,
				MemoryInBytes:           4 * 16 * 1024 * 1024 * 1024,
				GPU:                     0,
				EphemeralStorageInBytes: -1,
			},
		},
	}

	for name, tc := range testCases {
		shapeGetter := CreateShapeGetter(shapeClient)

		t.Run(name, func(t *testing.T) {
			shape, err := shapeGetter.GetNodePoolShape(&oke.NodePool{NodeShape: &tc.shape, NodeShapeConfig: tc.shapeConfig}, -1)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(shape, tc.expected) {
				t.Errorf("wanted %+v ; got %+v", tc.expected, shape)
			}

			if !strings.Contains(tc.shape, "Flex") {
				// we can't cache flex shapes so only check cache on non flex shapes
				cacheShape, ok := shapeGetter.(*shapeGetterImpl).cache[tc.shape]
				if !ok {
					t.Error("shape not found in cache")
				}

				if !reflect.DeepEqual(cacheShape, tc.expected) {
					t.Errorf("wanted %+v ; got %+v", tc.expected, shape)
				}
			}

		})
	}
}

func TestGetInstancePoolShape(t *testing.T) {

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
		shapeGetter := CreateShapeGetter(shapeClient)

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
