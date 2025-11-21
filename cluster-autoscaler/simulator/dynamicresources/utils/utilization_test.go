/*
Copyright 2024 The Kubernetes Authors.

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

package utils

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	apiv1 "k8s.io/api/core/v1"
	resourceapi "k8s.io/api/resource/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

func TestDynamicResourceUtilization(t *testing.T) {
	fooDriver := "driver.foo.com"
	barDriver := "driver.bar.com"
	node := test.BuildTestNode("node", 1000, 1000)
	noDraPod := test.BuildTestPod("noDraPod", 500, 500, test.WithNodeName("node"))

	for _, tc := range []struct {
		testName                   string
		nodeInfo                   *framework.NodeInfo
		wantUtilization            map[string]map[string]float64
		wantHighestUtilizationName apiv1.ResourceName
		wantHighestUtilization     float64
		wantErr                    error
	}{
		{
			testName:                   "no DRA resources",
			nodeInfo:                   framework.NewTestNodeInfo(node, noDraPod),
			wantUtilization:            map[string]map[string]float64{},
			wantHighestUtilization:     0,
			wantHighestUtilizationName: "",
			wantErr:                    nil,
		},
		{
			testName: "single slice, single pool, 6/10 devices used",
			nodeInfo: framework.NewNodeInfo(node,
				testResourceSlices(fooDriver, "pool1", "node", 0, 10, 1),
				testPodsWithClaims(fooDriver, "pool1", "node", 6, 1)...,
			),
			wantUtilization: map[string]map[string]float64{
				fooDriver: {
					"pool1": 0.6,
				},
			},
			wantHighestUtilization:     0.6,
			wantHighestUtilizationName: apiv1.ResourceName(fmt.Sprintf("%s/%s", fooDriver, "pool1")),
		},
		{
			testName: "multiple slices, multiple pools, multiple drivers",
			nodeInfo: framework.NewNodeInfo(node,
				mergeLists(
					testResourceSlices(fooDriver, "pool1", "node", 0, 10, 2),
					testResourceSlices(fooDriver, "pool2", "node", 0, 20, 5),
					testResourceSlices(barDriver, "pool1", "node", 0, 8, 2),
				),
				mergeLists(
					testPodsWithClaims(fooDriver, "pool1", "node", 6, 2),
					testPodsWithClaims(fooDriver, "pool2", "node", 18, 3),
					testPodsWithClaims(barDriver, "pool1", "node", 2, 1),
				)...,
			),
			wantUtilization: map[string]map[string]float64{
				fooDriver: {
					"pool1": 0.6,
					"pool2": 0.9,
				},
				barDriver: {
					"pool1": 0.25,
				},
			},
			wantHighestUtilization:     0.9,
			wantHighestUtilizationName: apiv1.ResourceName(fmt.Sprintf("%s/%s", fooDriver, "pool2")),
		},
		{
			testName: "old pool generations are ignored",
			nodeInfo: framework.NewNodeInfo(node,
				mergeLists(
					testResourceSlices(fooDriver, "pool1", "node", 0, 10, 2),
					testResourceSlices(fooDriver, "pool1", "node", 1, 20, 2),
					testResourceSlices(fooDriver, "pool1", "node", 2, 30, 2),
				),
				testPodsWithClaims(fooDriver, "pool1", "node", 6, 2)...,
			),
			wantUtilization: map[string]map[string]float64{
				fooDriver: {
					"pool1": 0.2,
				},
			},
			wantHighestUtilization:     0.2,
			wantHighestUtilizationName: apiv1.ResourceName(fmt.Sprintf("%s/%s", fooDriver, "pool1")),
		},
		{
			testName: "incomplete newest pool generation is an error",
			nodeInfo: framework.NewNodeInfo(node,
				mergeLists(
					testResourceSlices(fooDriver, "pool1", "node", 0, 10, 2),
					testResourceSlices(fooDriver, "pool1", "node", 1, 20, 2),
					testResourceSlices(fooDriver, "pool1", "node", 2, 30, 2)[:14],
				),
				testPodsWithClaims(fooDriver, "pool1", "node", 6, 2)...,
			),
			wantErr: cmpopts.AnyError,
		},
		{
			testName: "incomplete older pool generations are not an error",
			nodeInfo: framework.NewNodeInfo(node,
				mergeLists(
					testResourceSlices(fooDriver, "pool1", "node", 0, 10, 2)[:3],
					testResourceSlices(fooDriver, "pool1", "node", 1, 20, 2)[:7],
					testResourceSlices(fooDriver, "pool1", "node", 2, 30, 2),
				),
				testPodsWithClaims(fooDriver, "pool1", "node", 6, 2)...,
			),
			wantUtilization: map[string]map[string]float64{
				fooDriver: {
					"pool1": 0.2,
				},
			},
			wantHighestUtilization:     0.2,
			wantHighestUtilizationName: apiv1.ResourceName(fmt.Sprintf("%s/%s", fooDriver, "pool1")),
		},
		{
			testName: "partitionable devices, 2/4 partitions used",
			nodeInfo: framework.NewNodeInfo(node,
				mergeLists(
					testResourceSlicesWithPartionableDevices(fooDriver, "pool1", "gpu-0", "node", 2, 4),
				),
				mergeLists(
					testPodsWithCustomClaims(fooDriver, "pool1", "node", []string{"gpu-0-partition-0", "gpu-0-partition-1"}),
				)...,
			),
			wantUtilization: map[string]map[string]float64{
				fooDriver: {
					"pool1": 0.5,
				},
			},
			wantHighestUtilization:     0.5,
			wantHighestUtilizationName: apiv1.ResourceName(fmt.Sprintf("%s/%s", fooDriver, "pool1")),
		},
		{
			testName: "multi-GPU partitionable devices, 2/8 partitions used",
			nodeInfo: framework.NewNodeInfo(node,
				mergeLists(
					testResourceSlicesWithPartionableDevices(fooDriver, "pool1", "gpu-0", "node", 2, 4),
					testResourceSlicesWithPartionableDevices(fooDriver, "pool1", "gpu-1", "node", 0, 4),
				),
				mergeLists(
					testPodsWithCustomClaims(fooDriver, "pool1", "node", []string{"gpu-0-partition-0", "gpu-0-partition-1"}),
				)...,
			),
			wantUtilization: map[string]map[string]float64{
				fooDriver: {
					"pool1": 0.25,
				},
			},
			wantHighestUtilization:     0.25,
			wantHighestUtilizationName: apiv1.ResourceName(fmt.Sprintf("%s/%s", fooDriver, "pool1")),
		},
	} {
		if tc.testName != "" {
			continue
		}
		t.Run(tc.testName, func(t *testing.T) {
			utilization, err := CalculateDynamicResourceUtilization(tc.nodeInfo)
			if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("CalculateDynamicResourceUtilization(): unexpected error (-want +got): %s", diff)
			}
			if diff := cmp.Diff(tc.wantUtilization, utilization, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("CalculateDynamicResourceUtilization(): unexpected output (-want +got): %s", diff)
			}

			highestUtilName, highestUtil, err := HighestDynamicResourceUtilization(tc.nodeInfo)
			if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("HighestDynamicResourceUtilization(): unexpected error (-want +got): %s", diff)
			}
			if tc.wantHighestUtilizationName != highestUtilName {
				t.Errorf("HighestDynamicResourceUtilization(): unexpected resource name: want %q, got %q", tc.wantHighestUtilizationName, highestUtilName)
			}
			if tc.wantHighestUtilization != highestUtil {
				t.Errorf("HighestDynamicResourceUtilization(): unexpected utilization: want %v, got %v", tc.wantHighestUtilization, highestUtil)
			}
		})
	}
}

func testResourceSlices(driverName, poolName, nodeName string, poolGen, deviceCount, devicesPerSlice int64) []*resourceapi.ResourceSlice {
	sliceCount := deviceCount / devicesPerSlice

	deviceIndex := 0
	var result []*resourceapi.ResourceSlice
	for sliceIndex := range sliceCount {
		sliceName := fmt.Sprintf("%s-%s-slice-%d", driverName, poolName, sliceIndex)
		var devices []resourceapi.Device
		for range devicesPerSlice {
			devices = append(devices, resourceapi.Device{Name: fmt.Sprintf("%s-%s-dev-%d", driverName, poolName, deviceIndex)})
			deviceIndex++
		}
		result = append(result, &resourceapi.ResourceSlice{
			ObjectMeta: metav1.ObjectMeta{Name: sliceName, UID: types.UID(sliceName)},
			Spec: resourceapi.ResourceSliceSpec{
				Driver:   driverName,
				NodeName: &nodeName,
				Pool:     resourceapi.ResourcePool{Name: poolName, Generation: poolGen, ResourceSliceCount: sliceCount},
				Devices:  devices,
			},
		})
	}
	return result
}

func testResourceSlicesWithPartionableDevices(driverName, poolName, deviceName, nodeName string, poolGen, partitionCount int) []*resourceapi.ResourceSlice {
	sliceName := fmt.Sprintf("%s-%s-slice", driverName, poolName)
	var devices []resourceapi.Device
	for i := 0; i < partitionCount; i++ {
		devices = append(
			devices,
			resourceapi.Device{
				Name: fmt.Sprintf("%s-partition-%d", deviceName, i),
				Capacity: map[resourceapi.QualifiedName]resourceapi.DeviceCapacity{
					"memory": {
						Value: resource.MustParse("10Gi"),
					},
				},
				ConsumesCounters: []resourceapi.DeviceCounterConsumption{
					{
						CounterSet: fmt.Sprintf("%s-counter-set", deviceName),
						Counters: map[string]resourceapi.Counter{
							"memory": {
								Value: resource.MustParse("10Gi"),
							},
						},
					},
				},
			},
		)
	}
	devices = append(devices,
		resourceapi.Device{
			Name: deviceName,
			Capacity: map[resourceapi.QualifiedName]resourceapi.DeviceCapacity{
				"memory": {
					Value: resource.MustParse(fmt.Sprintf("%dGi", 10*partitionCount)),
				},
			},
			ConsumesCounters: []resourceapi.DeviceCounterConsumption{
				{
					CounterSet: fmt.Sprintf("%s-counter-set", deviceName),
					Counters: map[string]resourceapi.Counter{
						"memory": {
							Value: resource.MustParse(fmt.Sprintf("%dGi", 10*partitionCount)),
						},
					},
				},
			},
		},
	)
	resourceSlice := &resourceapi.ResourceSlice{
		ObjectMeta: metav1.ObjectMeta{Name: sliceName, UID: types.UID(sliceName)},
		Spec: resourceapi.ResourceSliceSpec{
			Driver:   driverName,
			NodeName: &nodeName,
			Pool:     resourceapi.ResourcePool{Name: poolName, Generation: int64(poolGen), ResourceSliceCount: 1},
			Devices:  devices,
			SharedCounters: []resourceapi.CounterSet{
				{
					Name: "gpu-0-counter-set",
					Counters: map[string]resourceapi.Counter{
						"memory": {
							Value: resource.MustParse(fmt.Sprintf("%dGi", 10*partitionCount)),
						},
					},
				},
			},
		},
	}
	return []*resourceapi.ResourceSlice{resourceSlice}
}

func testPodsWithClaims(driverName, poolName, nodeName string, deviceCount, devicesPerPod int64) []*framework.PodInfo {
	podCount := deviceCount / devicesPerPod

	deviceIndex := 0
	var result []*framework.PodInfo
	for podIndex := range podCount {
		pod := test.BuildTestPod(fmt.Sprintf("%s-%s-pod-%d", driverName, poolName, podIndex), 1, 1)
		var claims []*resourceapi.ResourceClaim
		for podDevIndex := range devicesPerPod {
			claimName := fmt.Sprintf("%s-claim-%d", pod.Name, podDevIndex)
			devName := fmt.Sprintf("%s-%s-dev-%d", driverName, poolName, deviceIndex)
			claims = append(claims, &resourceapi.ResourceClaim{
				ObjectMeta: metav1.ObjectMeta{Name: claimName, UID: types.UID(claimName)},
				Status: resourceapi.ResourceClaimStatus{
					Allocation: &resourceapi.AllocationResult{
						Devices: resourceapi.DeviceAllocationResult{
							Results: []resourceapi.DeviceRequestAllocationResult{
								{Request: fmt.Sprintf("request-%d", podDevIndex), Driver: driverName, Pool: poolName, Device: devName},
							},
						},
					},
				},
			})
			deviceIndex++
		}
		result = append(result, framework.NewPodInfo(pod, claims))
	}
	return result
}

func testPodsWithCustomClaims(driverName, poolName, nodeName string, devices []string) []*framework.PodInfo {
	deviceIndex := 0
	var result []*framework.PodInfo
	pod := test.BuildTestPod(fmt.Sprintf("%s-%s-pod", driverName, poolName), 1, 1)
	var claims []*resourceapi.ResourceClaim
	var results []resourceapi.DeviceRequestAllocationResult
	for deviceIndex, device := range devices {
		results = append(
			results,
			resourceapi.DeviceRequestAllocationResult{
				Request: fmt.Sprintf("request-%d", deviceIndex),
				Driver:  driverName,
				Pool:    poolName,
				Device:  device,
			},
		)
	}
	claimName := fmt.Sprintf("%s-claim", pod.Name)
	claims = append(claims, &resourceapi.ResourceClaim{
		ObjectMeta: metav1.ObjectMeta{Name: claimName, UID: types.UID(claimName)},
		Status: resourceapi.ResourceClaimStatus{
			Allocation: &resourceapi.AllocationResult{
				Devices: resourceapi.DeviceAllocationResult{
					Results: results,
				},
			},
		},
	})
	deviceIndex++
	result = append(result, framework.NewPodInfo(pod, claims))
	return result
}

func mergeLists[T any](sliceLists ...[]T) []T {
	var result []T
	for _, sliceList := range sliceLists {
		for _, slice := range sliceList {
			result = append(result, slice)
		}
	}
	return result
}

func TestGetUniquePartitionableDevicesCount(t *testing.T) {
	for _, tc := range []struct {
		testName  string
		devices   []resourceapi.Device
		wantCount int
	}{
		{
			testName:  "no devices",
			devices:   []resourceapi.Device{},
			wantCount: 0,
		},
		{
			testName: "single atomic device (no counters)",
			devices: []resourceapi.Device{
				{Name: "gpu-0"},
			},
			wantCount: 0,
		},
		{
			testName: "multiple atomic devices (no counters)",
			devices: []resourceapi.Device{
				{Name: "gpu-0"},
				{Name: "gpu-1"},
				{Name: "cpu-0"},
			},
			wantCount: 0,
		},
		{
			testName: "single partitionable device",
			devices: []resourceapi.Device{
				{
					Name: "gpu-0-whole",
					ConsumesCounters: []resourceapi.DeviceCounterConsumption{
						{CounterSet: "gpu-0-counters"},
					},
				},
			},
			wantCount: 1,
		},
		{
			testName: "single GPU partitioned into 4 devices",
			devices: []resourceapi.Device{
				{
					Name: "gpu-0-whole",
					ConsumesCounters: []resourceapi.DeviceCounterConsumption{
						{CounterSet: "gpu-0-counters"},
					},
				},
				{
					Name: "gpu-0-half-1",
					ConsumesCounters: []resourceapi.DeviceCounterConsumption{
						{CounterSet: "gpu-0-counters"},
					},
				},
				{
					Name: "gpu-0-half-2",
					ConsumesCounters: []resourceapi.DeviceCounterConsumption{
						{CounterSet: "gpu-0-counters"},
					},
				},
				{
					Name: "gpu-0-quarter-1",
					ConsumesCounters: []resourceapi.DeviceCounterConsumption{
						{CounterSet: "gpu-0-counters"},
					},
				},
			},
			wantCount: 1,
		},
		{
			testName: "two GPUs each partitioned into 2 devices",
			devices: []resourceapi.Device{
				{
					Name: "gpu-0-whole",
					ConsumesCounters: []resourceapi.DeviceCounterConsumption{
						{CounterSet: "gpu-0-counters"},
					},
				},
				{
					Name: "gpu-0-half",
					ConsumesCounters: []resourceapi.DeviceCounterConsumption{
						{CounterSet: "gpu-0-counters"},
					},
				},
				{
					Name: "gpu-1-whole",
					ConsumesCounters: []resourceapi.DeviceCounterConsumption{
						{CounterSet: "gpu-1-counters"},
					},
				},
				{
					Name: "gpu-1-half",
					ConsumesCounters: []resourceapi.DeviceCounterConsumption{
						{CounterSet: "gpu-1-counters"},
					},
				},
			},
			wantCount: 2,
		},
		{
			testName: "mixed atomic and partitionable devices",
			devices: []resourceapi.Device{
				{Name: "cpu-0"}, // Atomic
				{
					Name: "gpu-0-whole",
					ConsumesCounters: []resourceapi.DeviceCounterConsumption{
						{CounterSet: "gpu-0-counters"},
					},
				},
				{
					Name: "gpu-0-half",
					ConsumesCounters: []resourceapi.DeviceCounterConsumption{
						{CounterSet: "gpu-0-counters"},
					},
				},
			},
			wantCount: 1,
		},
		{
			testName: "device with multiple counter sets",
			devices: []resourceapi.Device{
				{
					Name: "gpu-0-whole",
					ConsumesCounters: []resourceapi.DeviceCounterConsumption{
						{CounterSet: "gpu-0-memory"},
						{CounterSet: "gpu-0-compute"},
					},
				},
				{
					Name: "gpu-0-half",
					ConsumesCounters: []resourceapi.DeviceCounterConsumption{
						{CounterSet: "gpu-0-memory"},
						{CounterSet: "gpu-0-compute"},
					},
				},
			},
			wantCount: 1,
		},
		{
			testName: "device with overlapping counter sets",
			devices: []resourceapi.Device{
				{
					Name: "gpu-0-partition-1",
					ConsumesCounters: []resourceapi.DeviceCounterConsumption{
						{CounterSet: "gpu-0-memory"},
					},
				},
				{
					Name: "gpu-0-partition-2",
					ConsumesCounters: []resourceapi.DeviceCounterConsumption{
						{CounterSet: "gpu-0-memory"},
						{CounterSet: "gpu-0-compute"},
					},
				},
			},
			wantCount: 1,
		},
		{
			testName: "complex multi-GPU scenario",
			devices: []resourceapi.Device{
				{
					Name: "gpu-0-whole",
					ConsumesCounters: []resourceapi.DeviceCounterConsumption{
						{CounterSet: "gpu-0-counters"},
					},
				},
				{
					Name: "gpu-0-half-1",
					ConsumesCounters: []resourceapi.DeviceCounterConsumption{
						{CounterSet: "gpu-0-counters"},
					},
				},
				{
					Name: "gpu-0-quarter-1",
					ConsumesCounters: []resourceapi.DeviceCounterConsumption{
						{CounterSet: "gpu-0-counters"},
					},
				},
				{
					Name: "gpu-1-whole",
					ConsumesCounters: []resourceapi.DeviceCounterConsumption{
						{CounterSet: "gpu-1-counters"},
					},
				},
				{
					Name: "gpu-1-half-1",
					ConsumesCounters: []resourceapi.DeviceCounterConsumption{
						{CounterSet: "gpu-1-counters"},
					},
				},
				{
					Name: "gpu-2-whole",
					ConsumesCounters: []resourceapi.DeviceCounterConsumption{
						{CounterSet: "gpu-2-counters"},
					},
				},
			},
			wantCount: 3,
		},
		{
			testName: "device with no counter consumption entries",
			devices: []resourceapi.Device{
				{
					Name:             "device-1",
					ConsumesCounters: []resourceapi.DeviceCounterConsumption{},
				},
			},
			wantCount: 1,
		},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			count := getUniquePartitionableDevicesCount(tc.devices)

			if count != tc.wantCount {
				t.Errorf("getUniquePartitionableDevicesCount() = %v, want %v", count, tc.wantCount)
			}
		})
	}
}
