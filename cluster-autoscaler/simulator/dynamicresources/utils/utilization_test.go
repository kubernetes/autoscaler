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
	"k8s.io/utils/ptr"
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
				testPodsWithClaims(fooDriver, "pool1", "node", 6, 1, false, 0)...,
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
					testPodsWithClaims(fooDriver, "pool1", "node", 6, 2, false, 0),
					testPodsWithClaims(fooDriver, "pool2", "node", 18, 3, false, 0),
					testPodsWithClaims(barDriver, "pool1", "node", 2, 1, false, 0),
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
				testPodsWithClaims(fooDriver, "pool1", "node", 6, 2, false, 0)...,
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
				testPodsWithClaims(fooDriver, "pool1", "node", 6, 2, false, 0)...,
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
				testPodsWithClaims(fooDriver, "pool1", "node", 6, 2, false, 0)...,
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
				testResourceSlicesWithPartitionableDevices(fooDriver, "pool1", "node", 1, 4, 2, "gpu-0"),
				testPodsWithCustomClaims("pod", fooDriver, "pool1", "node", []string{"gpu-0-partition-0", "gpu-0-partition-1"})...,
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
			testName: "partitionable devices, multiple pods, 2/4 partitions used",
			nodeInfo: framework.NewNodeInfo(node,
				testResourceSlicesWithPartitionableDevices(fooDriver, "pool1", "node", 1, 4, 2, "gpu-0"),
				mergeLists(
					testPodsWithCustomClaims("pod-0", fooDriver, "pool1", "node", []string{"gpu-0-partition-0"}),
					testPodsWithCustomClaims("pod-1", fooDriver, "pool1", "node", []string{"gpu-0-partition-1"}),
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
			testName: "partitionable devices, multiple pools",
			nodeInfo: framework.NewNodeInfo(node,
				mergeLists(
					testResourceSlicesWithPartitionableDevices(fooDriver, "pool0", "node", 1, 4, 2, "gpu-0"),
					testResourceSlicesWithPartitionableDevices(fooDriver, "pool1", "node", 1, 4, 2, "gpu-0"),
					testResourceSlicesWithPartitionableDevices(fooDriver, "pool2", "node", 1, 4, 2, "gpu-1"),
				),
				mergeLists(
					testPodsWithCustomClaims("pod0", fooDriver, "pool0", "node", []string{"gpu-0-partition-0", "gpu-0-partition-1"}),
					testPodsWithCustomClaims("pod1", fooDriver, "pool1", "node", []string{"gpu-0"}),
				)...,
			),
			wantUtilization: map[string]map[string]float64{
				fooDriver: {
					"pool0": 0.5,
					"pool1": 1,
					"pool2": 0,
				},
			},
			wantHighestUtilization:     1,
			wantHighestUtilizationName: apiv1.ResourceName(fmt.Sprintf("%s/%s", fooDriver, "pool1")),
		},
		{
			testName: "multi-counter partitionable devices",
			nodeInfo: framework.NewNodeInfo(node,
				addPartitionableDeviceToResourceSlices(
					testResourceSlicesWithPartitionableDevices(fooDriver, "pool1", "node", 1, 4, 2, "gpu-0"),
					fooDriver,
					"pool1",
					"gpu-0-uneven-partition",
					map[string]resource.Quantity{
						"memory": resource.MustParse("1Gi"),
						"cpu":    resource.MustParse("500m"),
					},
					// cpu is a new counter type not tracked by the pool yet; total=1 makes the 500m consumption a 50% ratio.
					map[string]resource.Quantity{
						"cpu": resource.MustParse("1"),
					},
				),
				testPodsWithCustomClaims("pod", fooDriver, "pool1", "node", []string{"gpu-0-uneven-partition"})...,
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
				testResourceSlicesWithPartitionableDevices(fooDriver, "pool1", "node", 1, 4, 2, "gpu-0", "gpu-1"),
				testPodsWithCustomClaims("pod", fooDriver, "pool1", "node", []string{"gpu-0-partition-0", "gpu-0-partition-1"})...,
			),
			wantUtilization: map[string]map[string]float64{
				fooDriver: {
					"pool1": 0.25,
				},
			},
			wantHighestUtilization:     0.25,
			wantHighestUtilizationName: apiv1.ResourceName(fmt.Sprintf("%s/%s", fooDriver, "pool1")),
		},
		{
			testName: "Single Pod with AdminAccess ResourceClaim doesn't count for utilization",
			nodeInfo: framework.NewNodeInfo(node,
				testResourceSlices(fooDriver, "pool1", "node", 0, 10, 1),
				testPodsWithClaims(fooDriver, "pool1", "node", 6, 1, true, 0)...,
			),
			wantUtilization: map[string]map[string]float64{
				fooDriver: {
					"pool1": 0,
				},
			},
			wantHighestUtilization:     0,
			wantHighestUtilizationName: apiv1.ResourceName(fmt.Sprintf("%s/%s", fooDriver, "pool1")),
		},
		{
			testName: "Multiple Pods with AdminAccess ResourceClaims don't count for utilization",
			nodeInfo: framework.NewNodeInfo(node,
				testResourceSlices(fooDriver, "pool1", "node", 0, 10, 1),
				mergeLists(
					testPodsWithClaims(fooDriver, "pool1", "node", 6, 1, true, 0),
					testPodsWithClaims(fooDriver, "pool1", "node", 6, 1, false, 6),
				)...,
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
			testName: "Pod with mixed AdminAccess requests counts only non-admin",
			nodeInfo: framework.NewNodeInfo(node,
				testResourceSlices(fooDriver, "pool1", "node", 0, 10, 1),
				testPodsWithMixedAdminAccessClaims(fooDriver, "pool1", "node", 6, []bool{true, true, false, false, false, false}, 0)...,
			),
			wantUtilization: map[string]map[string]float64{
				fooDriver: {
					"pool1": 0.4,
				},
			},
			wantHighestUtilization:     0.4,
			wantHighestUtilizationName: apiv1.ResourceName(fmt.Sprintf("%s/%s", fooDriver, "pool1")),
		},
	} {
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

func TestCalculateAtomicDevicesPoolUtilization(t *testing.T) {
	for _, tc := range []struct {
		testName        string
		allocated       []resourceapi.Device
		unallocated     []resourceapi.Device
		wantUtilization float64
	}{
		{
			testName:        "no devices",
			allocated:       []resourceapi.Device{},
			unallocated:     []resourceapi.Device{},
			wantUtilization: 0,
		},
		{
			testName:        "no allocated devices",
			allocated:       []resourceapi.Device{},
			unallocated:     []resourceapi.Device{{Name: "dev-0"}, {Name: "dev-1"}},
			wantUtilization: 0,
		},
		{
			testName:        "all devices allocated",
			allocated:       []resourceapi.Device{{Name: "dev-0"}, {Name: "dev-1"}},
			unallocated:     []resourceapi.Device{},
			wantUtilization: 1,
		},
		{
			testName:        "partial allocation",
			allocated:       []resourceapi.Device{{Name: "dev-0"}},
			unallocated:     []resourceapi.Device{{Name: "dev-1"}, {Name: "dev-2"}},
			wantUtilization: 1.0 / 3.0,
		},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			total := append(tc.allocated, tc.unallocated...)
			utilization := calculateAtomicDevicesPoolUtilization(tc.allocated, total)
			if utilization != tc.wantUtilization {
				t.Errorf("calculateAtomicDevicesPoolUtilization() = %v, want %v", utilization, tc.wantUtilization)
			}
		})
	}
}

func TestCalculatePartitionableDevicesPoolUtilization(t *testing.T) {
	for _, tc := range []struct {
		testName        string
		allocated       []resourceapi.Device
		unallocated     []resourceapi.Device
		sharedCounters  []resourceapi.CounterSet
		wantUtilization float64
	}{
		{
			testName:        "no devices",
			allocated:       []resourceapi.Device{},
			unallocated:     []resourceapi.Device{},
			sharedCounters:  []resourceapi.CounterSet{},
			wantUtilization: 0,
		},
		{
			testName:  "no allocated devices",
			allocated: []resourceapi.Device{},
			unallocated: []resourceapi.Device{
				{
					Name: "dev-0",
					ConsumesCounters: []resourceapi.DeviceCounterConsumption{
						{
							CounterSet: "counter-set-0",
							Counters: map[string]resourceapi.Counter{
								"memory": {Value: resource.MustParse("10Gi")},
							},
						},
					},
				},
			},
			sharedCounters: []resourceapi.CounterSet{
				{
					Name: "counter-set-0",
					Counters: map[string]resourceapi.Counter{
						"memory": {Value: resource.MustParse("10Gi")},
					},
				},
			},
			wantUtilization: 0,
		},
		{
			testName: "partial allocation",
			allocated: []resourceapi.Device{
				{
					Name: "dev-0-partition-0",
					ConsumesCounters: []resourceapi.DeviceCounterConsumption{
						{
							CounterSet: "counter-set-0",
							Counters: map[string]resourceapi.Counter{
								"memory": {Value: resource.MustParse("5Gi")},
							},
						},
					},
				},
			},
			unallocated: []resourceapi.Device{
				{
					Name: "dev-0",
					ConsumesCounters: []resourceapi.DeviceCounterConsumption{
						{
							CounterSet: "counter-set-0",
							Counters: map[string]resourceapi.Counter{
								"memory": {Value: resource.MustParse("10Gi")},
							},
						},
					},
				},
				{
					Name: "dev-0-partition-1",
					ConsumesCounters: []resourceapi.DeviceCounterConsumption{
						{
							CounterSet: "counter-set-0",
							Counters: map[string]resourceapi.Counter{
								"memory": {Value: resource.MustParse("5Gi")},
							},
						},
					},
				},
			},
			sharedCounters: []resourceapi.CounterSet{
				{
					Name: "counter-set-0",
					Counters: map[string]resourceapi.Counter{
						"memory": {Value: resource.MustParse("10Gi")},
					},
				},
			},
			wantUtilization: 0.5,
		},
		{
			testName: "full allocation of one device, all partitions used",
			allocated: []resourceapi.Device{
				{
					Name: "dev-0-partition-0",
					ConsumesCounters: []resourceapi.DeviceCounterConsumption{
						{
							CounterSet: "counter-set-0",
							Counters: map[string]resourceapi.Counter{
								"memory": {Value: resource.MustParse("5Gi")},
							},
						},
					},
				},
				{
					Name: "dev-0-partition-1",
					ConsumesCounters: []resourceapi.DeviceCounterConsumption{
						{
							CounterSet: "counter-set-0",
							Counters: map[string]resourceapi.Counter{
								"memory": {Value: resource.MustParse("5Gi")},
							},
						},
					},
				},
			},
			unallocated: []resourceapi.Device{
				{
					Name: "dev-0",
					ConsumesCounters: []resourceapi.DeviceCounterConsumption{
						{
							CounterSet: "counter-set-0",
							Counters: map[string]resourceapi.Counter{
								"memory": {Value: resource.MustParse("10Gi")},
							},
						},
					},
				},
			},
			sharedCounters: []resourceapi.CounterSet{
				{
					Name: "counter-set-0",
					Counters: map[string]resourceapi.Counter{
						"memory": {Value: resource.MustParse("10Gi")},
					},
				},
			},
			wantUtilization: 1,
		},
		{
			testName: "full allocation of one device, 1 device used",
			allocated: []resourceapi.Device{
				{
					Name: "dev-0",
					ConsumesCounters: []resourceapi.DeviceCounterConsumption{
						{
							CounterSet: "counter-set-0",
							Counters: map[string]resourceapi.Counter{
								"memory": {Value: resource.MustParse("10Gi")},
							},
						},
					},
				},
			},
			unallocated: []resourceapi.Device{
				{
					Name: "dev-0-partition-0",
					ConsumesCounters: []resourceapi.DeviceCounterConsumption{
						{
							CounterSet: "counter-set-0",
							Counters: map[string]resourceapi.Counter{
								"memory": {Value: resource.MustParse("5Gi")},
							},
						},
					},
				},
				{
					Name: "dev-0-partition-1",
					ConsumesCounters: []resourceapi.DeviceCounterConsumption{
						{
							CounterSet: "counter-set-0",
							Counters: map[string]resourceapi.Counter{
								"memory": {Value: resource.MustParse("5Gi")},
							},
						},
					},
				},
			},
			sharedCounters: []resourceapi.CounterSet{
				{
					Name: "counter-set-0",
					Counters: map[string]resourceapi.Counter{
						"memory": {Value: resource.MustParse("10Gi")},
					},
				},
			},
			wantUtilization: 1,
		},
		{
			testName: "full allocation of one device, multiple counters used",
			allocated: []resourceapi.Device{
				{
					Name: "dev-0",
					ConsumesCounters: []resourceapi.DeviceCounterConsumption{
						{
							CounterSet: "counter-set-0",
							Counters: map[string]resourceapi.Counter{
								"memory": {Value: resource.MustParse("10Gi")},
								"cores":  {Value: resource.MustParse("8")},
							},
						},
					},
				},
			},
			unallocated: []resourceapi.Device{
				{
					Name: "dev-0-partition-0",
					ConsumesCounters: []resourceapi.DeviceCounterConsumption{
						{
							CounterSet: "counter-set-0",
							Counters: map[string]resourceapi.Counter{
								"memory": {Value: resource.MustParse("5Gi")},
								"cores":  {Value: resource.MustParse("6")},
							},
						},
					},
				},
				{
					Name: "dev-0-partition-1",
					ConsumesCounters: []resourceapi.DeviceCounterConsumption{
						{
							CounterSet: "counter-set-0",
							Counters: map[string]resourceapi.Counter{
								"memory": {Value: resource.MustParse("5Gi")},
								"cores":  {Value: resource.MustParse("2")},
							},
						},
					},
				},
			},
			sharedCounters: []resourceapi.CounterSet{
				{
					Name: "counter-set-0",
					Counters: map[string]resourceapi.Counter{
						"memory": {Value: resource.MustParse("10Gi")},
						"cores":  {Value: resource.MustParse("16")},
					},
				},
			},
			wantUtilization: 1,
		},
		{
			testName: "partial allocation of one device, multiple counters used",
			allocated: []resourceapi.Device{
				{
					Name: "dev-0-partition-0",
					ConsumesCounters: []resourceapi.DeviceCounterConsumption{
						{
							CounterSet: "counter-set-0",
							Counters: map[string]resourceapi.Counter{
								"memory": {Value: resource.MustParse("5Gi")},
								"cores":  {Value: resource.MustParse("6")},
							},
						},
					},
				},
			},
			unallocated: []resourceapi.Device{
				{
					Name: "dev-0",
					ConsumesCounters: []resourceapi.DeviceCounterConsumption{
						{
							CounterSet: "counter-set-0",
							Counters: map[string]resourceapi.Counter{
								"memory": {Value: resource.MustParse("10Gi")},
								"cores":  {Value: resource.MustParse("8")},
							},
						},
					},
				},
				{
					Name: "dev-0-partition-1",
					ConsumesCounters: []resourceapi.DeviceCounterConsumption{
						{
							CounterSet: "counter-set-0",
							Counters: map[string]resourceapi.Counter{
								"memory": {Value: resource.MustParse("5Gi")},
								"cores":  {Value: resource.MustParse("2")},
							},
						},
					},
				},
			},
			sharedCounters: []resourceapi.CounterSet{
				{
					Name: "counter-set-0",
					Counters: map[string]resourceapi.Counter{
						"memory": {Value: resource.MustParse("10Gi")},
						"cores":  {Value: resource.MustParse("8")},
					},
				},
			},
			wantUtilization: 0.75,
		},
		{
			testName: "complex allocation of multiple devices, multiple counters used",
			allocated: []resourceapi.Device{
				{
					Name: "dev-0-partition-0",
					ConsumesCounters: []resourceapi.DeviceCounterConsumption{
						{
							CounterSet: "counter-set-0",
							Counters: map[string]resourceapi.Counter{
								"memory": {Value: resource.MustParse("25Gi")},
								"cores":  {Value: resource.MustParse("8")},
							},
						},
					},
				},
				{
					Name: "dev-0-partition-2",
					ConsumesCounters: []resourceapi.DeviceCounterConsumption{
						{
							CounterSet: "counter-set-0",
							Counters: map[string]resourceapi.Counter{
								"memory": {Value: resource.MustParse("50Gi")},
								"cores":  {Value: resource.MustParse("8")},
							},
						},
					},
				},
				{
					Name: "dev-1-partition-0",
					ConsumesCounters: []resourceapi.DeviceCounterConsumption{
						{
							CounterSet: "counter-set-1",
							Counters: map[string]resourceapi.Counter{
								"memory": {Value: resource.MustParse("50Gi")},
								"cores":  {Value: resource.MustParse("16")},
							},
						},
					},
				},
				{
					Name: "dev-1-partition-1",
					ConsumesCounters: []resourceapi.DeviceCounterConsumption{
						{
							CounterSet: "counter-set-1",
							Counters: map[string]resourceapi.Counter{
								"memory": {Value: resource.MustParse("50Gi")},
								"cores":  {Value: resource.MustParse("16")},
							},
						},
					},
				},
				{
					Name: "dev-2",
					ConsumesCounters: []resourceapi.DeviceCounterConsumption{
						{
							CounterSet: "counter-set-2",
							Counters: map[string]resourceapi.Counter{
								"memory": {Value: resource.MustParse("100Gi")},
								"cores":  {Value: resource.MustParse("32")},
							},
						},
					},
				},
			},
			unallocated: []resourceapi.Device{
				{
					Name: "dev-0",
					ConsumesCounters: []resourceapi.DeviceCounterConsumption{
						{
							CounterSet: "counter-set-0",
							Counters: map[string]resourceapi.Counter{
								"memory": {Value: resource.MustParse("100Gi")},
								"cores":  {Value: resource.MustParse("32")},
							},
						},
					},
				},
				{
					Name: "dev-0-partition-1",
					ConsumesCounters: []resourceapi.DeviceCounterConsumption{
						{
							CounterSet: "counter-set-0",
							Counters: map[string]resourceapi.Counter{
								"memory": {Value: resource.MustParse("25Gi")},
								"cores":  {Value: resource.MustParse("16")},
							},
						},
					},
				},
				{
					Name: "dev-1",
					ConsumesCounters: []resourceapi.DeviceCounterConsumption{
						{
							CounterSet: "counter-set-1",
							Counters: map[string]resourceapi.Counter{
								"memory": {Value: resource.MustParse("100Gi")},
								"cores":  {Value: resource.MustParse("32")},
							},
						},
					},
				},
				{
					Name: "dev-2",
					ConsumesCounters: []resourceapi.DeviceCounterConsumption{
						{
							CounterSet: "counter-set-2",
							Counters: map[string]resourceapi.Counter{
								"memory": {Value: resource.MustParse("100Gi")},
								"cores":  {Value: resource.MustParse("32")},
							},
						},
					},
				},
				{
					Name: "dev-3",
					ConsumesCounters: []resourceapi.DeviceCounterConsumption{
						{
							CounterSet: "counter-set-3",
							Counters: map[string]resourceapi.Counter{
								"memory": {Value: resource.MustParse("100Gi")},
								"cores":  {Value: resource.MustParse("32")},
							},
						},
					},
				},
			},
			sharedCounters: []resourceapi.CounterSet{
				{
					Name: "counter-set-0",
					Counters: map[string]resourceapi.Counter{
						"memory": {Value: resource.MustParse("100Gi")},
						"cores":  {Value: resource.MustParse("32")},
					},
				},
				{
					Name: "counter-set-1",
					Counters: map[string]resourceapi.Counter{
						"memory": {Value: resource.MustParse("100Gi")},
						"cores":  {Value: resource.MustParse("32")},
					},
				},
				{
					Name: "counter-set-2",
					Counters: map[string]resourceapi.Counter{
						"memory": {Value: resource.MustParse("100Gi")},
						"cores":  {Value: resource.MustParse("32")},
					},
				},
				{
					Name: "counter-set-3",
					Counters: map[string]resourceapi.Counter{
						"memory": {Value: resource.MustParse("100Gi")},
						"cores":  {Value: resource.MustParse("32")},
					},
				},
			},
			wantUtilization: 0.6875,
		},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			total := append(tc.allocated, tc.unallocated...)
			utilization := calculatePartitionableDevicesPoolUtilization(tc.allocated, total, tc.sharedCounters)
			if utilization != tc.wantUtilization {
				t.Errorf("calculatePartitionableDevicesPoolUtilization() = %v, want %v", utilization, tc.wantUtilization)
			}
		})
	}
}

func testResourceSlices(driverName, poolName, nodeName string, poolGen, deviceCount, devicesPerSlice int64) []*resourceapi.ResourceSlice {
	sliceCount := deviceCount / devicesPerSlice

	var result []*resourceapi.ResourceSlice
	for sliceIndex := range sliceCount {
		sliceName := fmt.Sprintf("%s-%s-slice-%d", driverName, poolName, sliceIndex)
		var devices []resourceapi.Device
		for sliceDeviceIndex := range devicesPerSlice {
			devices = append(devices, resourceapi.Device{Name: fmt.Sprintf("%s-%s-dev-%d", driverName, poolName, sliceIndex*devicesPerSlice+sliceDeviceIndex)})
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

func testResourceSlicesWithPartitionableDevices(driverName, poolName, nodeName string, poolGen, partitionCount, slicecount int, deviceNames ...string) []*resourceapi.ResourceSlice {
	sliceName := fmt.Sprintf("%s-%s-slice", driverName, poolName)
	var devices []resourceapi.Device
	for _, deviceName := range deviceNames {
		for i := 0; i < partitionCount; i++ {
			devices = append(
				devices,
				resourceapi.Device{
					Name: fmt.Sprintf("%s-partition-%d", deviceName, i),
					Capacity: map[resourceapi.QualifiedName]resourceapi.DeviceCapacity{
						"memory": {
							Value: resource.MustParse("10Gi"),
						},
						"cpu": {
							Value: resource.MustParse("1000m"),
						},
					},
					ConsumesCounters: []resourceapi.DeviceCounterConsumption{
						{
							CounterSet: fmt.Sprintf("%s-counter-set", deviceName),
							Counters: map[string]resourceapi.Counter{
								"memory": {
									Value: resource.MustParse("10Gi"),
								},
								"cpu": {
									Value: resource.MustParse("100m"),
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
					"cpu": {
						Value: resource.MustParse(fmt.Sprintf("%dm", 100*partitionCount)),
					},
				},
				ConsumesCounters: []resourceapi.DeviceCounterConsumption{
					{
						CounterSet: fmt.Sprintf("%s-counter-set", deviceName),
						Counters: map[string]resourceapi.Counter{
							"memory": {
								Value: resource.MustParse(fmt.Sprintf("%dGi", 10*partitionCount)),
							},
							"cpu": {
								Value: resource.MustParse(fmt.Sprintf("%dm", 100*partitionCount)),
							},
						},
					},
				},
			},
		)
	}
	resourceSlice := &resourceapi.ResourceSlice{
		ObjectMeta: metav1.ObjectMeta{Name: sliceName, UID: types.UID(sliceName)},
		Spec: resourceapi.ResourceSliceSpec{
			Driver:   driverName,
			NodeName: &nodeName,
			Pool:     resourceapi.ResourcePool{Name: poolName, Generation: int64(poolGen), ResourceSliceCount: int64(slicecount)},
			Devices:  devices,
		},
	}

	sharedCounters := make([]resourceapi.CounterSet, 0, len(deviceNames))
	for _, deviceName := range deviceNames {
		sharedCounters = append(sharedCounters, resourceapi.CounterSet{
			Name: fmt.Sprintf("%s-counter-set", deviceName),
			Counters: map[string]resourceapi.Counter{
				"memory": {
					Value: resource.MustParse(fmt.Sprintf("%dGi", 10*partitionCount)),
				},
				"cpu": {
					Value: resource.MustParse(fmt.Sprintf("%dm", 100*partitionCount)),
				},
			},
		})
	}

	countersSlice := &resourceapi.ResourceSlice{
		ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-counters", sliceName), UID: types.UID(fmt.Sprintf("%s-counters", sliceName))},
		Spec: resourceapi.ResourceSliceSpec{
			Driver:         driverName,
			NodeName:       &nodeName,
			Pool:           resourceapi.ResourcePool{Name: poolName, Generation: int64(poolGen), ResourceSliceCount: int64(slicecount)},
			SharedCounters: sharedCounters,
		},
	}

	return []*resourceapi.ResourceSlice{resourceSlice, countersSlice}
}

// addPartitionableDeviceToResourceSlices appends a new partition device to an existing set of partitionable resource slices.
// The device will consume from the existing pool's CounterSet.
// additionalSharedCounters adds new counter types (with their totals) to the existing SharedCounterSet,
// which is necessary when the new device consumes counter types not already tracked by the pool.
func addPartitionableDeviceToResourceSlices(resourceSlices []*resourceapi.ResourceSlice, driverName, poolName, deviceName string, consumedCounters map[string]resource.Quantity, additionalSharedCounters map[string]resource.Quantity) []*resourceapi.ResourceSlice {
	devicesResourceSlice := resourceSlices[0]
	countersResourceSlice := resourceSlices[1]

	// Use the existing pool's CounterSet name so the device is counted as a partition of the same physical device.
	counterSetName := countersResourceSlice.Spec.SharedCounters[0].Name

	counters := make(map[string]resourceapi.Counter, len(consumedCounters))
	for name, qty := range consumedCounters {
		counters[name] = resourceapi.Counter{Value: qty.DeepCopy()}
	}
	device := resourceapi.Device{
		Name: deviceName,
		ConsumesCounters: []resourceapi.DeviceCounterConsumption{
			{
				CounterSet: counterSetName,
				Counters:   counters,
			},
		},
	}
	devicesResourceSlice.Spec.Devices = append(devicesResourceSlice.Spec.Devices, device)

	// Add any new counter types (with their totals) to the existing SharedCounterSet.
	for i, cs := range countersResourceSlice.Spec.SharedCounters {
		if cs.Name == counterSetName {
			for name, qty := range additionalSharedCounters {
				countersResourceSlice.Spec.SharedCounters[i].Counters[name] = resourceapi.Counter{Value: qty.DeepCopy()}
			}
			break
		}
	}

	return []*resourceapi.ResourceSlice{devicesResourceSlice, countersResourceSlice}
}

func testPodsWithClaims(driverName, poolName, nodeName string, deviceCount, devicesPerPod int64, adminAccess bool, podIndexOffset int) []*framework.PodInfo {
	podCount := deviceCount / devicesPerPod

	deviceIndex := 0
	var result []*framework.PodInfo
	for podIndex := range podCount {
		pod := test.BuildTestPod(fmt.Sprintf("%s-%s-pod-%d", driverName, poolName, podIndex+int64(podIndexOffset)), 1, 1)
		var claims []*resourceapi.ResourceClaim
		for podDevIndex := range devicesPerPod {
			claimName := fmt.Sprintf("%s-claim-%d", pod.Name, podDevIndex)
			devName := fmt.Sprintf("%s-%s-dev-%d", driverName, poolName, deviceIndex)
			devReqName := fmt.Sprintf("request-%d", podDevIndex)
			devReq := resourceapi.DeviceRequest{
				Name: devReqName,
				Exactly: &resourceapi.ExactDeviceRequest{
					AdminAccess: ptr.To(adminAccess),
				},
			}
			claims = append(claims, &resourceapi.ResourceClaim{
				ObjectMeta: metav1.ObjectMeta{Name: claimName, UID: types.UID(claimName)},
				Spec: resourceapi.ResourceClaimSpec{
					Devices: resourceapi.DeviceClaim{
						Requests: []resourceapi.DeviceRequest{devReq},
					},
				},
				Status: resourceapi.ResourceClaimStatus{
					Allocation: &resourceapi.AllocationResult{
						Devices: resourceapi.DeviceAllocationResult{
							Results: []resourceapi.DeviceRequestAllocationResult{
								{Request: devReqName, Driver: driverName, Pool: poolName, Device: devName, AdminAccess: ptr.To(adminAccess)},
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

func testPodsWithCustomClaims(podName, driverName, poolName, nodeName string, devices []string) []*framework.PodInfo {
	pod := test.BuildTestPod(fmt.Sprintf("%s-%s-%s-pod", podName, driverName, poolName), 1, 1)
	claimName := fmt.Sprintf("%s-claim", pod.Name)
	claim := &resourceapi.ResourceClaim{
		ObjectMeta: metav1.ObjectMeta{Name: claimName, UID: types.UID(claimName)},
		Status: resourceapi.ResourceClaimStatus{
			Allocation: &resourceapi.AllocationResult{
				Devices: resourceapi.DeviceAllocationResult{},
			},
		},
	}
	for i, device := range devices {
		claim.Status.Allocation.Devices.Results = append(claim.Status.Allocation.Devices.Results, resourceapi.DeviceRequestAllocationResult{
			Request: fmt.Sprintf("request-%d", i),
			Driver:  driverName,
			Pool:    poolName,
			Device:  device,
		})
	}
	return []*framework.PodInfo{framework.NewPodInfo(pod, []*resourceapi.ResourceClaim{claim})}
}

// testPodsWithMixedAdminAccessClaims creates Pods with ResourceClaims that have mixed AdminAccess settings.
func testPodsWithMixedAdminAccessClaims(driverName, poolName, nodeName string, deviceCount int64, adminAccessPattern []bool, podIndexOffset int) []*framework.PodInfo {
	pis := testPodsWithClaims(driverName, poolName, nodeName, deviceCount, int64(len(adminAccessPattern)), true, podIndexOffset)
	for i, podInfo := range pis {
		claims := podInfo.NeededResourceClaims
		for claimsIndex, claim := range claims {
			claim.Status.Allocation.Devices.Results[0].AdminAccess = ptr.To(adminAccessPattern[claimsIndex])
		}
		pi := framework.NewPodInfo(podInfo.Pod, claims)
		pis[i] = pi
	}
	return pis
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
			wantCount: 0, // Should be treated as atomic, not partitionable
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
