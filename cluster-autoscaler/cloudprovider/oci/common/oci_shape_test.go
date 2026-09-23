/*
Copyright 2021-2023 Oracle and/or its affiliates.
*/

package common

import (
	"context"
	"reflect"
	"strings"
	"testing"

	oke "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/containerengine"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/core"
)

type mockShapeClient struct {
	err                   error
	listShapeResponses    []core.ListShapesResponse
	getInstanceConfigResp core.GetInstanceConfigurationResponse
	requestCount          int
}

func (m *mockShapeClient) ListShapes(_ context.Context, _ core.ListShapesRequest) (core.ListShapesResponse, error) {
	m.requestCount++
	return m.listShapeResponses[m.requestCount-1], m.err
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
	listShapeResponses: []core.ListShapesResponse{
		{
			Items: []core.Shape{
				{
					Shape:       common.String("VM.Standard2.8"),
					Ocpus:       common.Float32(8),
					MemoryInGBs: common.Float32(120),
				},
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
		listShapeResponses: []core.ListShapesResponse{
			{
				Items: []core.Shape{
					{
						Shape:       common.String("VM.Standard1.1"),
						Ocpus:       common.Float32(2),
						MemoryInGBs: common.Float32(16),
					},
				},
				OpcNextPage: common.String("nextPage"),
			},
			{
				Items: []core.Shape{
					{
						Shape:       common.String("VM.Standard1.2"),
						Ocpus:       common.Float32(2),
						MemoryInGBs: common.Float32(16),
					},
				},
			},
		},
	}

	testCases := map[string]struct {
		shape       string
		shapeConfig *oke.NodeShapeConfig
		expected    *Shape
	}{
		"basic x86 shape": {
			shape: "VM.Standard1.2",
			expected: &Shape{
				Name:                    "VM.Standard1.2",
				CPU:                     4,
				MemoryInBytes:           16 * 1024 * 1024 * 1024,
				GPU:                     0,
				EphemeralStorageInBytes: -1,
			},
		},
		"flex x86 shape": {
			shape: "VM.Standard.E3.Flex",
			shapeConfig: &oke.NodeShapeConfig{
				Ocpus:       common.Float32(4),
				MemoryInGBs: common.Float32(64),
			},
			expected: &Shape{
				Name:                    "VM.Standard.E3.Flex",
				CPU:                     8,
				MemoryInBytes:           4 * 16 * 1024 * 1024 * 1024,
				GPU:                     0,
				EphemeralStorageInBytes: -1,
			},
		},
		"flex A1 shape uses 1:1 OCPU-to-vCPU": {
			shape: "VM.Standard.A1.Flex",
			shapeConfig: &oke.NodeShapeConfig{
				Ocpus:       common.Float32(4),
				MemoryInGBs: common.Float32(24),
			},
			expected: &Shape{
				Name:                    "VM.Standard.A1.Flex",
				CPU:                     4,
				MemoryInBytes:           24 * 1024 * 1024 * 1024,
				GPU:                     0,
				EphemeralStorageInBytes: -1,
			},
		},
		"flex A2 shape uses 1:2 OCPU-to-vCPU": {
			shape: "VM.Standard.A2.Flex",
			shapeConfig: &oke.NodeShapeConfig{
				Ocpus:       common.Float32(4),
				MemoryInGBs: common.Float32(24),
			},
			expected: &Shape{
				Name:                    "VM.Standard.A2.Flex",
				CPU:                     8,
				MemoryInBytes:           24 * 1024 * 1024 * 1024,
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
				CPU:           16,
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

func TestOcpuToVCPU(t *testing.T) {
	testCases := map[string]struct {
		shape    string
		ocpus    float32
		expected float32
	}{
		"x86 E3":  {shape: "VM.Standard.E3.Flex", ocpus: 4, expected: 8},
		"x86 E4":  {shape: "VM.Standard.E4.Flex", ocpus: 8, expected: 16},
		"x86 E5":  {shape: "VM.Standard.E5.Flex", ocpus: 2, expected: 4},
		"ARM A1":  {shape: "VM.Standard.A1.Flex", ocpus: 4, expected: 4},
		"ARM A2":  {shape: "VM.Standard.A2.Flex", ocpus: 4, expected: 8},
		"ARM A4":  {shape: "VM.Standard.A4.Flex", ocpus: 4, expected: 8},
		"BM A1":   {shape: "BM.Standard.A1.160", ocpus: 160, expected: 160},
		"GPU A10": {shape: "BM.GPU.A10.4", ocpus: 64, expected: 128},
		"fixed":   {shape: "VM.Standard2.8", ocpus: 8, expected: 16},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := ocpuToVCPU(tc.shape, tc.ocpus)
			if got != tc.expected {
				t.Errorf("ocpuToVCPU(%q, %v) = %v, want %v", tc.shape, tc.ocpus, got, tc.expected)
			}
		})
	}
}
