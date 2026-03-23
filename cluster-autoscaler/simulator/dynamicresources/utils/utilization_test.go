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
