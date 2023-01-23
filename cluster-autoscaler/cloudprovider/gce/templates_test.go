/*
Copyright 2016 The Kubernetes Authors.

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

package gce

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"

	"k8s.io/autoscaler/cluster-autoscaler/config"
	gpuUtils "k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/autoscaler/cluster-autoscaler/utils/units"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	gce "google.golang.org/api/compute/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	quota "k8s.io/apiserver/pkg/quota/v1"
)

// TestBuildNodeFromTemplateSetsResources tests that capacity and allocatable
// are loaded into the node template status, a few error scenarios, and physical
// ephemeral storage (an intermediate result), but it doesn't test that capacity
// and allocatable are computed correctly, (the test itself calls
// GceTemplateBuilder.BuildCapacity, GceTemplateBuilder.CalculateAllocatable,
// and ParseEvictionHardOrGetDefault to compute expected values); computations
// are tested separately.
func TestBuildNodeFromTemplateSetsResources(t *testing.T) {
	var thirtyPodsPerNode int64 = 30
	type testCase struct {
		scenario string
		// test inputs
		kubeEnv               string
		accelerators          []*gce.AcceleratorConfig
		attachedLocalSSDCount int64
		pods                  *int64
		// other test inputs (constant across test cases, because they are test invariants for now)
		physicalCpu     int64
		physicalMemory  int64
		bootDiskSizeGiB int64
		// dependent inputs, should match kubeEnv, used to compute expected capacity and allocatable, out of test scope
		kubeReserved                  bool
		reservedCpu                   string
		reservedMemory                string
		reservedEphemeralStorage      string
		isEphemeralStorageBlocked     bool
		ephemeralStorageLocalSSDCount int64
		extendedResources             apiv1.ResourceList
		// test outputs
		expectedMigInfoErr      bool
		expectedNodeTemplateErr bool
	}
	testCases := []testCase{
		{
			scenario: "kube-reserved present in kube-env",
			kubeEnv: "ENABLE_NODE_PROBLEM_DETECTOR: 'daemonset'\n" +
				"NODE_LABELS: a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true\n" +
				"DNS_SERVER_IP: '10.0.0.10'\n" +
				fmt.Sprintf("KUBELET_TEST_ARGS: --experimental-allocatable-ignore-eviction --kube-reserved=cpu=1000m,memory=%v,ephemeral-storage=30Gi\n", 1*units.MiB) +
				"NODE_TAINTS: 'dedicated=ml:NoSchedule,test=dev:PreferNoSchedule,a=b:c'\n",
			accelerators: []*gce.AcceleratorConfig{
				{AcceleratorType: "nvidia-tesla-k80", AcceleratorCount: 3},
				{AcceleratorType: "nvidia-tesla-p100", AcceleratorCount: 8},
			},
			physicalCpu:              8,
			physicalMemory:           200 * units.MiB,
			bootDiskSizeGiB:          300,
			kubeReserved:             true,
			reservedCpu:              "1000m",
			reservedMemory:           fmt.Sprintf("%v", 1*units.MiB),
			reservedEphemeralStorage: "30Gi",
		},
		{
			scenario: "no kube-reserved in kube-env",
			kubeEnv: "ENABLE_NODE_PROBLEM_DETECTOR: 'daemonset'\n" +
				"NODE_LABELS: a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true\n" +
				"DNS_SERVER_IP: '10.0.0.10'\n" +
				"NODE_TAINTS: 'dedicated=ml:NoSchedule,test=dev:PreferNoSchedule,a=b:c'\n",
			physicalCpu:    8,
			physicalMemory: 200 * units.MiB,
			kubeReserved:   false,
		}, {
			scenario:       "no kube-env at all",
			kubeEnv:        "",
			physicalCpu:    8,
			physicalMemory: 200 * units.MiB,
			kubeReserved:   false,
		}, {
			scenario:           "totally messed up kube-env",
			kubeEnv:            "This kube-env is totally messed up",
			expectedMigInfoErr: true,
		}, {
			scenario:       "max pods per node specified",
			kubeEnv:        "",
			physicalCpu:    8,
			physicalMemory: 200 * units.MiB,
			pods:           &thirtyPodsPerNode,
			kubeReserved:   false,
		},
		{
			scenario: "BLOCK_EPH_STORAGE_BOOT_DISK in kube-env",
			kubeEnv: "ENABLE_NODE_PROBLEM_DETECTOR: 'daemonset'\n" +
				"NODE_LABELS: a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true\n" +
				"DNS_SERVER_IP: '10.0.0.10'\n" +
				"AUTOSCALER_ENV_VARS: os_distribution=cos;os=linux;kube_reserved=cpu=0,memory=0,ephemeral-storage=0;BLOCK_EPH_STORAGE_BOOT_DISK=true\n" +
				"NODE_TAINTS: 'dedicated=ml:NoSchedule,test=dev:PreferNoSchedule,a=b:c'\n",
			physicalCpu:               8,
			physicalMemory:            200 * units.MiB,
			bootDiskSizeGiB:           300,
			reservedCpu:               "0m",
			reservedMemory:            fmt.Sprintf("%v", 0*units.MiB),
			reservedEphemeralStorage:  "0Gi",
			kubeReserved:              true,
			isEphemeralStorageBlocked: true,
		},
		{
			scenario: "BLOCK_EPH_STORAGE_BOOT_DISK is false in kube-env",
			kubeEnv: "ENABLE_NODE_PROBLEM_DETECTOR: 'daemonset'\n" +
				"NODE_LABELS: a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true\n" +
				"DNS_SERVER_IP: '10.0.0.10'\n" +
				"AUTOSCALER_ENV_VARS: os_distribution=cos;os=linux;kube_reserved=cpu=0,memory=0,ephemeral-storage=0;BLOCK_EPH_STORAGE_BOOT_DISK=false\n" +
				"NODE_TAINTS: 'dedicated=ml:NoSchedule,test=dev:PreferNoSchedule,a=b:c'\n",
			reservedCpu:              "0m",
			reservedMemory:           fmt.Sprintf("%v", 0*units.MiB),
			reservedEphemeralStorage: "0Gi",
			kubeReserved:             true,
		},
		{
			scenario:                      "more local SSDs requested for ephemeral storage than attached",
			kubeEnv:                       "AUTOSCALER_ENV_VARS: os_distribution=cos;os=linux;ephemeral_storage_local_ssd_count=1\n",
			ephemeralStorageLocalSSDCount: 1,
			attachedLocalSSDCount:         0,
			expectedNodeTemplateErr:       true,
		},
		{
			scenario:                      "all attached local SSDs requested for ephemeral storage",
			kubeEnv:                       "AUTOSCALER_ENV_VARS: os_distribution=cos;os=linux;ephemeral_storage_local_ssd_count=2\n",
			physicalCpu:                   8,
			physicalMemory:                200 * units.MiB,
			bootDiskSizeGiB:               300,
			ephemeralStorageLocalSSDCount: 2,
			attachedLocalSSDCount:         2,
		},
		{
			scenario:                      "more local SSDs attached than requested for ephemeral storage",
			kubeEnv:                       "AUTOSCALER_ENV_VARS: os_distribution=cos;os=linux;ephemeral_storage_local_ssd_count=2\n",
			physicalCpu:                   8,
			physicalMemory:                200 * units.MiB,
			ephemeralStorageLocalSSDCount: 2,
			attachedLocalSSDCount:         4,
		},
		{
			scenario:                      "ephemeral storage on local SSDs with kube-reserved",
			kubeEnv:                       "AUTOSCALER_ENV_VARS: kube_reserved=cpu=0,memory=0,ephemeral-storage=10Gi;os_distribution=cos;os=linux;ephemeral_storage_local_ssd_count=2\n",
			physicalCpu:                   8,
			physicalMemory:                200 * units.MiB,
			ephemeralStorageLocalSSDCount: 2,
			kubeReserved:                  true,
			reservedCpu:                   "0m",
			reservedMemory:                fmt.Sprintf("%v", 0*units.MiB),
			reservedEphemeralStorage:      "10Gi",
			attachedLocalSSDCount:         4,
		},
		{
			scenario:                      "extended_resources present in kube-env",
			kubeEnv:                       "AUTOSCALER_ENV_VARS: kube_reserved=cpu=0,memory=0,ephemeral-storage=10Gi;os_distribution=cos;os=linux;ephemeral_storage_local_ssd_count=2;extended_resources=someResource=2,anotherResource=1G\n",
			physicalCpu:                   8,
			physicalMemory:                200 * units.MiB,
			ephemeralStorageLocalSSDCount: 2,
			kubeReserved:                  true,
			reservedCpu:                   "0m",
			reservedMemory:                fmt.Sprintf("%v", 0*units.MiB),
			reservedEphemeralStorage:      "10Gi",
			attachedLocalSSDCount:         4,
			extendedResources: apiv1.ResourceList{
				apiv1.ResourceName("someResource"):    *resource.NewQuantity(2, resource.DecimalSI),
				apiv1.ResourceName("anotherResource"): *resource.NewQuantity(1*units.GB, resource.DecimalSI),
			},
		},
		{
			scenario:                      "malformed extended_resources in kube-env",
			kubeEnv:                       "AUTOSCALER_ENV_VARS: kube_reserved=cpu=0,memory=0,ephemeral-storage=10Gi;os_distribution=cos;os=linux;ephemeral_storage_local_ssd_count=2;extended_resources=someResource\n",
			physicalCpu:                   8,
			physicalMemory:                200 * units.MiB,
			ephemeralStorageLocalSSDCount: 2,
			kubeReserved:                  true,
			reservedCpu:                   "0m",
			reservedMemory:                fmt.Sprintf("%v", 0*units.MiB),
			reservedEphemeralStorage:      "10Gi",
			attachedLocalSSDCount:         4,
			extendedResources:             apiv1.ResourceList{},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.scenario, func(t *testing.T) {
			tb := &GceTemplateBuilder{}
			mig := &gceMig{
				gceRef: GceRef{
					Name:    "some-name",
					Project: "some-proj",
					Zone:    "us-central1-b",
				},
			}
			template := &gce.InstanceTemplate{
				Name: "node-name",
				Properties: &gce.InstanceProperties{
					GuestAccelerators: tc.accelerators,
					Metadata:          &gce.Metadata{},
					MachineType:       "irrelevant-type",
					Disks: []*gce.AttachedDisk{
						{
							Boot: true,
							InitializeParams: &gce.AttachedDiskInitializeParams{
								DiskSizeGb: tc.bootDiskSizeGiB,
							},
						},
					},
				},
			}
			for i := int64(0); i < tc.attachedLocalSSDCount; i++ {
				template.Properties.Disks = append(template.Properties.Disks, &gce.AttachedDisk{
					Type: "SCRATCH",
					InitializeParams: &gce.AttachedDiskInitializeParams{
						DiskType: "local-ssd",
					},
				})
			}
			if tc.kubeEnv != "" {
				template.Properties.Metadata.Items = []*gce.MetadataItems{{Key: "kube-env", Value: &tc.kubeEnv}}
			}
			migOsInfo, err := tb.MigOsInfo(mig.Id(), template)
			if tc.expectedMigInfoErr {
				assert.Error(t, err)
				return
			}
			node, err := tb.BuildNodeFromTemplate(mig, migOsInfo, template, tc.physicalCpu, tc.physicalMemory, tc.pods, &GceReserved{})
			if tc.expectedNodeTemplateErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, node)
				assert.NotNil(t, node.Status)
				assert.NotNil(t, node.Status.Capacity)
				assert.NotNil(t, node.Status.Allocatable)
				if tc.bootDiskSizeGiB > 0 && !tc.isEphemeralStorageBlocked {
					val, ok := node.Annotations[BootDiskSizeAnnotation]
					if !ok {
						t.Errorf("Expected to have boot disk size annotation, have nil")
					}
					assert.Equal(t, val, strconv.FormatInt(tc.bootDiskSizeGiB, 10))
				} else if tc.attachedLocalSSDCount > 0 {
					val, ok := node.Annotations[LocalSsdCountAnnotation]
					if !ok {
						t.Errorf("Expected to have local SSD count annotation")
					}
					assert.Equal(t, val, strconv.FormatInt(tc.attachedLocalSSDCount, 10))
				}
				// this logic is a duplicate of logic under test and would best be captured by
				// specifying physicalEphemeralStorageGiB in the testCase struct
				physicalEphemeralStorageGiB := tc.bootDiskSizeGiB
				if tc.ephemeralStorageLocalSSDCount > 0 {
					physicalEphemeralStorageGiB = tc.ephemeralStorageLocalSSDCount * LocalSSDDiskSizeInGiB
				} else if tc.isEphemeralStorageBlocked {
					physicalEphemeralStorageGiB = 0
				}
				migOsInfo := NewMigOsInfo(OperatingSystemLinux, OperatingSystemDistributionCOS, "")
				capacity, err := tb.BuildCapacity(migOsInfo, tc.physicalCpu, tc.physicalMemory, tc.accelerators, physicalEphemeralStorageGiB*units.GiB, tc.ephemeralStorageLocalSSDCount, tc.pods, &GceReserved{}, tc.extendedResources)
				assert.NoError(t, err)
				assertEqualResourceLists(t, "Capacity", capacity, node.Status.Capacity)
				if !tc.kubeReserved {
					assertEqualResourceLists(t, "Allocatable", capacity, node.Status.Allocatable)
				} else {
					reserved, err := makeResourceList(tc.reservedCpu, tc.reservedMemory, 0, tc.reservedEphemeralStorage)
					assert.NoError(t, err)
					allocatable := tb.CalculateAllocatable(capacity, reserved, ParseEvictionHardOrGetDefault(nil))
					assertEqualResourceLists(t, "Allocatable", allocatable, node.Status.Allocatable)
				}
			}
		})
	}
}

func TestBuildGenericLabels(t *testing.T) {
	type testCase struct {
		name              string
		os                OperatingSystem
		arch              SystemArchitecture
		expectedOsLabel   string
		expectedArchLabel string
		expectedError     bool
	}
	testCases := []testCase{
		{
			name:              "os linux",
			os:                OperatingSystemLinux,
			arch:              Amd64,
			expectedArchLabel: "amd64",
			expectedOsLabel:   "linux",
			expectedError:     false,
		},
		{
			name:            "os windows",
			os:              OperatingSystemWindows,
			expectedOsLabel: "windows",
			expectedError:   false,
		},
		{
			name:            "os unknown",
			os:              OperatingSystemUnknown,
			expectedOsLabel: "",
			expectedError:   true,
		},
		{
			name:              "os linux arm",
			os:                OperatingSystemLinux,
			arch:              Arm64,
			expectedArchLabel: "arm64",
			expectedOsLabel:   "linux",
			expectedError:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expectedLabels := map[string]string{
				apiv1.LabelTopologyRegion:     "us-central1",
				apiv1.LabelTopologyZone:       "us-central1-b",
				gceCSITopologyKeyZone:         "us-central1-b",
				apiv1.LabelHostname:           "sillyname",
				apiv1.LabelInstanceTypeStable: "n1-standard-8",
				apiv1.LabelArchStable:         tc.expectedArchLabel,
				apiv1.LabelOSStable:           tc.expectedOsLabel,
			}
			labels, err := BuildGenericLabels(GceRef{
				Name:    "kubernetes-minion-group",
				Project: "mwielgus-proj",
				Zone:    "us-central1-b"},
				"n1-standard-8",
				"sillyname",
				tc.os,
				tc.arch)
			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, expectedLabels, labels)
			}
		})
	}
}

func TestCalculateAllocatable(t *testing.T) {
	type testCase struct {
		scenario                    string
		capacityCpu                 string
		reservedCpu                 string
		allocatableCpu              string
		capacityMemory              string
		reservedMemory              string
		allocatableMemory           string
		capacityEphemeralStorage    string
		reservedEphemeralStorage    string
		allocatableEphemeralStorage string
	}
	testCases := []testCase{
		{
			scenario:                    "no reservations",
			capacityCpu:                 "8",
			reservedCpu:                 "0",
			allocatableCpu:              "8",
			capacityMemory:              fmt.Sprintf("%v", 200*units.MiB),
			reservedMemory:              "0",
			allocatableMemory:           fmt.Sprintf("%v", 200*units.MiB-GetKubeletEvictionHardForMemory(nil)),
			capacityEphemeralStorage:    fmt.Sprintf("%v", 200*units.GiB),
			reservedEphemeralStorage:    "0",
			allocatableEphemeralStorage: fmt.Sprintf("%v", 200*units.GiB-GetKubeletEvictionHardForEphemeralStorage(200*GiB, nil)),
		},
		{
			scenario:                    "reserved cpu, memory and ephemeral storage",
			capacityCpu:                 "8",
			reservedCpu:                 "1000m",
			allocatableCpu:              "7000m",
			capacityMemory:              fmt.Sprintf("%v", 200*units.MiB),
			reservedMemory:              fmt.Sprintf("%v", 50*units.MiB),
			allocatableMemory:           fmt.Sprintf("%v", 150*units.MiB-GetKubeletEvictionHardForMemory(nil)),
			capacityEphemeralStorage:    fmt.Sprintf("%v", 200*units.GiB),
			reservedEphemeralStorage:    fmt.Sprintf("%v", 40*units.GiB),
			allocatableEphemeralStorage: fmt.Sprintf("%v", 160*units.GiB-GetKubeletEvictionHardForEphemeralStorage(200*GiB, nil)),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.scenario, func(t *testing.T) {
			tb := GceTemplateBuilder{}
			capacity, err := makeResourceList(tc.capacityCpu, tc.capacityMemory, 0, tc.capacityEphemeralStorage)
			assert.NoError(t, err)
			reserved, err := makeResourceList(tc.reservedCpu, tc.reservedMemory, 0, tc.reservedEphemeralStorage)
			assert.NoError(t, err)
			expectedAllocatable, err := makeResourceList(tc.allocatableCpu, tc.allocatableMemory, 0, tc.allocatableEphemeralStorage)
			assert.NoError(t, err)
			allocatable := tb.CalculateAllocatable(capacity, reserved, ParseEvictionHardOrGetDefault(nil))
			assertEqualResourceLists(t, "Allocatable", expectedAllocatable, allocatable)
		})
	}
}

func TestBuildAllocatableFromKubeEnv(t *testing.T) {
	type testCase struct {
		kubeEnv                  string
		capacityCpu              string
		capacityMemory           string
		capacityEphemeralStorage string
		expectedCpu              string
		expectedMemory           string
		expectedEphemeralStorage string
		gpuCount                 int64
		expectedErr              bool
	}
	testCases := []testCase{{
		kubeEnv: "ENABLE_NODE_PROBLEM_DETECTOR: 'daemonset'\n" +
			"NODE_LABELS: a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true\n" +
			"DNS_SERVER_IP: '10.0.0.10'\n" +
			"KUBELET_TEST_ARGS: --experimental-allocatable-ignore-eviction --kube-reserved=cpu=1000m,memory=300000Mi,ephemeral-storage=30Gi\n" +
			"NODE_TAINTS: 'dedicated=ml:NoSchedule,test=dev:PreferNoSchedule,a=b:c'\n",
		capacityCpu:              "4000m",
		capacityMemory:           "700000Mi",
		capacityEphemeralStorage: "100Gi",
		expectedCpu:              "3000m",
		expectedMemory:           "399900Mi", // capacityMemory-kube_reserved-DefaultKubeletEvictionHardMemory
		expectedEphemeralStorage: "60Gi",     // capacityEphemeralStorage-kube_reserved-DefaultKubeletEvictionHardMemory
		gpuCount:                 10,
		expectedErr:              false,
	}, {
		kubeEnv: "ENABLE_NODE_PROBLEM_DETECTOR: 'daemonset'\n" +
			"NODE_LABELS: a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true\n" +
			"DNS_SERVER_IP: '10.0.0.10'\n" +
			"NODE_TAINTS: 'dedicated=ml:NoSchedule,test=dev:PreferNoSchedule,a=b:c'\n",
		capacityCpu:    "4000m",
		capacityMemory: "700000Mi",
		expectedErr:    true,
	}}
	for _, tc := range testCases {
		capacity, err := makeResourceList(tc.capacityCpu, tc.capacityMemory, tc.gpuCount, tc.capacityEphemeralStorage)
		assert.NoError(t, err)
		tb := GceTemplateBuilder{}
		allocatable, err := tb.BuildAllocatableFromKubeEnv(capacity, tc.kubeEnv, ParseEvictionHardOrGetDefault(nil))
		if tc.expectedErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			expectedResources, err := makeResourceList(tc.expectedCpu, tc.expectedMemory, tc.gpuCount, tc.expectedEphemeralStorage)
			assert.NoError(t, err)
			for res, expectedQty := range expectedResources {
				qty, found := allocatable[res]
				assert.True(t, found)
				assert.Equal(t, qty.Value(), expectedQty.Value())
			}
		}
	}
}

func TestParseEvictionHard(t *testing.T) {
	type testCase struct {
		memory                        string
		ephemeralStorage              string
		memoryExpected                int64 // bytes
		ephemeralStorageRatioExpected float64
	}
	testCases := []testCase{{
		memory:                        "200Mi",
		ephemeralStorage:              "15%",
		memoryExpected:                200 * MiB,
		ephemeralStorageRatioExpected: 0.15,
	}, {
		memory:                        "2Gi",
		ephemeralStorage:              "11.5%",
		memoryExpected:                2 * GiB,
		ephemeralStorageRatioExpected: 0.115,
	}, {
		memory:                        "",
		ephemeralStorage:              "", // empty string, fallback to default
		memoryExpected:                100 * MiB,
		ephemeralStorageRatioExpected: 0.1,
	}, {
		memory:                        "110292",
		ephemeralStorage:              "11", // percentage missing, should fallback to default
		memoryExpected:                110292,
		ephemeralStorageRatioExpected: 0.1,
	}, {
		memory:                        "abcb12", // unparsable, fallback to default
		ephemeralStorage:              "-11%",   // negative percentage, should fallback to default
		memoryExpected:                100 * MiB,
		ephemeralStorageRatioExpected: 0.1,
	}}
	for _, tc := range testCases {
		test := map[string]string{
			MemoryEvictionHardTag:           tc.memory,
			EphemeralStorageEvictionHardTag: tc.ephemeralStorage,
		}
		actualOutput := ParseEvictionHardOrGetDefault(test)
		assert.EqualValues(t, tc.memoryExpected, actualOutput.MemoryEvictionQuantity, "TestParseEviction Failed Memory. %v expected does not match %v actual.", tc.memoryExpected, actualOutput.MemoryEvictionQuantity)
		assert.EqualValues(t, tc.ephemeralStorageRatioExpected, actualOutput.EphemeralStorageEvictionRatio, "TestParseEviction Failed Ephemeral Storage. %v expected does not match %v actual.", tc.memoryExpected, actualOutput.EphemeralStorageEvictionRatio)
	}
}

func TestGetAcceleratorCount(t *testing.T) {
	testCases := []struct {
		accelerators []*gce.AcceleratorConfig
		count        int64
	}{{
		accelerators: []*gce.AcceleratorConfig{},
		count:        0,
	}, {
		accelerators: []*gce.AcceleratorConfig{
			{AcceleratorType: "nvidia-tesla-k80", AcceleratorCount: 3},
		},
		count: 3,
	}, {
		accelerators: []*gce.AcceleratorConfig{
			{AcceleratorType: "nvidia-tesla-k80", AcceleratorCount: 3},
			{AcceleratorType: "nvidia-tesla-p100", AcceleratorCount: 8},
		},
		count: 11,
	}, {
		accelerators: []*gce.AcceleratorConfig{
			{AcceleratorType: "other-type", AcceleratorCount: 3},
			{AcceleratorType: "nvidia-tesla-p100", AcceleratorCount: 8},
		},
		count: 8,
	}}

	for _, tc := range testCases {
		tb := GceTemplateBuilder{}
		assert.Equal(t, tc.count, tb.getAcceleratorCount(tc.accelerators))
	}
}

func TestBuildCapacityMemory(t *testing.T) {
	type testCase struct {
		physicalCpu            int64
		physicalMemory         int64
		os                     OperatingSystem
		expectedCapacityMemory int64
	}
	testCases := []testCase{
		{
			physicalCpu:            1,
			physicalMemory:         2 * units.GiB,
			os:                     OperatingSystemLinux,
			expectedCapacityMemory: 2*units.GiB - 32*units.MiB - kernelReservedMemory - int64(math.Min(correctionConstant*float64(2*units.GiB), maximumCorrectionValue)) - lowMemoryOffset,
		},
		{
			physicalCpu:            2,
			physicalMemory:         4 * units.GiB,
			os:                     OperatingSystemLinux,
			expectedCapacityMemory: 4*units.GiB - 64*units.MiB - kernelReservedMemory - swiotlbReservedMemory - int64(math.Min(correctionConstant*float64(4*units.GiB), maximumCorrectionValue)) - lowMemoryOffset,
		},
		{
			physicalCpu:            32,
			physicalMemory:         128 * units.GiB,
			os:                     OperatingSystemLinux,
			expectedCapacityMemory: 128*units.GiB - 2*units.GiB - kernelReservedMemory - swiotlbReservedMemory - int64(math.Min(correctionConstant*float64(128*units.GiB), maximumCorrectionValue)),
		},
		{
			physicalCpu:            2,
			physicalMemory:         4 * units.GiB,
			os:                     OperatingSystemWindows,
			expectedCapacityMemory: 4 * units.GiB,
		},
	}
	for idx, tc := range testCases {
		t.Run(fmt.Sprintf("%v", idx), func(t *testing.T) {
			tb := GceTemplateBuilder{}
			noAccelerators := make([]*gce.AcceleratorConfig, 0)
			migOsInfo := NewMigOsInfo(tc.os, OperatingSystemDistributionCOS, "")
			buildCapacity, err := tb.BuildCapacity(migOsInfo, tc.physicalCpu, tc.physicalMemory, noAccelerators, -1, 0, nil, &GceReserved{}, apiv1.ResourceList{})
			assert.NoError(t, err)
			expectedCapacity, err := makeResourceList2(tc.physicalCpu, tc.expectedCapacityMemory, 0, 110)
			assert.NoError(t, err)
			assertEqualResourceLists(t, "Capacity", expectedCapacity, buildCapacity)
		})
	}
}

func TestExtractAutoscalingOptionsFromKubeEnv(t *testing.T) {
	cases := []struct {
		desc          string
		env           string
		expectedValue map[string]string
		expectedErr   bool
	}{
		{
			desc:          "autoscaling_options not specified",
			env:           "AUTOSCALER_ENV_VARS: node_labels=a=b,c=d;node_taints=a=b:c,d=e:f\n",
			expectedValue: map[string]string{},
			expectedErr:   false,
		},
		{
			desc:          "empty KubeEnv",
			env:           "",
			expectedValue: map[string]string{},
			expectedErr:   false,
		},
		{
			desc:          "unparsable KubeEnv",
			env:           "AUTOSCALER_ENV_VARS",
			expectedValue: nil,
			expectedErr:   true,
		},
		{
			desc: "partial option set",
			env:  "AUTOSCALER_ENV_VARS: node_labels=a=b;autoscaling_options=scaledownunreadytime=1h",
			expectedValue: map[string]string{
				config.DefaultScaleDownUnreadyTimeKey: "1h",
			},
			expectedErr: false,
		},
		{
			desc: "full option set",
			env:  "AUTOSCALER_ENV_VARS: node_labels=a,b;autoscaling_options=scaledownutilizationthreshold=0.4,scaledowngpuutilizationthreshold=0.5,scaledownunneededtime=30m,scaledownunreadytime=1h",
			expectedValue: map[string]string{
				config.DefaultScaleDownUtilizationThresholdKey:    "0.4",
				config.DefaultScaleDownGpuUtilizationThresholdKey: "0.5",
				config.DefaultScaleDownUnneededTimeKey:            "30m",
				config.DefaultScaleDownUnreadyTimeKey:             "1h",
			},
			expectedErr: false,
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			value, err := extractAutoscalingOptionsFromKubeEnv(c.env)
			assert.Equal(t, c.expectedValue, value)
			if c.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExtractAutoscalerVarFromKubeEnv(t *testing.T) {
	cases := []struct {
		desc          string
		name          string
		env           string
		expectedValue string
		expectedFound bool
		expectedErr   error
	}{
		{
			desc:          "node_labels",
			name:          "node_labels",
			env:           "AUTOSCALER_ENV_VARS: node_labels=a=b,c=d;node_taints=a=b:c,d=e:f\n",
			expectedValue: "a=b,c=d",
			expectedFound: true,
			expectedErr:   nil,
		},
		{
			desc:          "node_labels not found",
			name:          "node_labels",
			env:           "AUTOSCALER_ENV_VARS: node_taints=a=b:c,d=e:f\n",
			expectedValue: "",
			expectedFound: false,
			expectedErr:   nil,
		},
		{
			desc:          "node_labels empty",
			name:          "node_labels",
			env:           "AUTOSCALER_ENV_VARS: node_labels=;node_taints=a=b:c,d=e:f\n",
			expectedValue: "",
			expectedFound: true,
			expectedErr:   nil,
		},
		{
			desc:          "node_taints",
			name:          "node_taints",
			env:           "AUTOSCALER_ENV_VARS: node_labels=a=b,c=d;node_taints=a=b:c,d=e:f\n",
			expectedValue: "a=b:c,d=e:f",
			expectedFound: true,
			expectedErr:   nil,
		},
		{
			desc:          "malformed node_labels",
			name:          "node_labels",
			env:           "AUTOSCALER_ENV_VARS: node_labels;node_taints=a=b:c,d=e:f\n",
			expectedValue: "",
			expectedFound: false,
			expectedErr:   fmt.Errorf("malformed autoscaler var: node_labels"),
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			value, found, err := extractAutoscalerVarFromKubeEnv(c.env, c.name)
			assert.Equal(t, c.expectedValue, value)
			assert.Equal(t, c.expectedFound, found)
			assert.Equal(t, c.expectedErr, err)
		})
	}
}

func TestExtractLabelsFromKubeEnv(t *testing.T) {
	poolLabel := "cloud.google.com/gke-nodepool"
	preemptibleLabel := "cloud.google.com/gke-preemptible"
	expectedLabels := map[string]string{
		"a":              "b",
		"c":              "d",
		poolLabel:        "pool-3",
		preemptibleLabel: "true",
	}
	cases := []struct {
		desc   string
		env    string
		expect map[string]string
		err    error
	}{
		{
			desc: "from NODE_LABELS",
			env: "ENABLE_NODE_PROBLEM_DETECTOR: 'daemonset'\n" +
				"NODE_LABELS: a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true\n" +
				"DNS_SERVER_IP: '10.0.0.10'\n",
			expect: expectedLabels,
			err:    nil,
		},
		{
			desc: "from AUTOSCALER_ENV_VARS.node_labels",
			env: "ENABLE_NODE_PROBLEM_DETECTOR: 'daemonset'\n" +
				"AUTOSCALER_ENV_VARS: node_labels=a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true;" +
				"node_taints='dedicated=ml:NoSchedule,test=dev:PreferNoSchedule,a=b:c';" +
				"kube_reserved=cpu=1000m,memory=300000Mi;" +
				"os=linux\n" +
				"DNS_SERVER_IP: '10.0.0.10'\n",
			expect: expectedLabels,
			err:    nil,
		},
		{
			desc: "malformed key-value in AUTOSCALER_ENV_VARS.node_labels",
			env:  "AUTOSCALER_ENV_VARS: node_labels=ab,c=d\n",
			err:  fmt.Errorf("error while parsing key-value list, val: ab"),
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			labels, err := extractLabelsFromKubeEnv(c.env)
			assert.Equal(t, c.err, err)
			if c.err != nil {
				return
			}
			assert.Equal(t, c.expect, labels)

			template := &gce.InstanceTemplate{
				Properties: &gce.InstanceProperties{
					Metadata: &gce.Metadata{
						Items: []*gce.MetadataItems{
							{
								Key:   "kube-env",
								Value: &c.env,
							},
						},
					},
				},
			}

			labels, err = GetLabelsFromTemplate(template)
			assert.Equal(t, c.err, err)
			if c.err != nil {
				return
			}
			assert.Equal(t, c.expect, labels)
		})
	}
}

func TestExtractTaintsFromKubeEnv(t *testing.T) {
	expectedTaints := makeTaintSet([]apiv1.Taint{
		{
			Key:    "dedicated",
			Value:  "ml",
			Effect: apiv1.TaintEffectNoSchedule,
		},
		{
			Key:    "test",
			Value:  "dev",
			Effect: apiv1.TaintEffectPreferNoSchedule,
		},
		{
			Key:    "a",
			Value:  "b",
			Effect: apiv1.TaintEffect("c"),
		},
	})

	cases := []struct {
		desc   string
		env    string
		expect map[apiv1.Taint]bool
		err    error
	}{
		{
			desc: "from NODE_TAINTS",
			env: "ENABLE_NODE_PROBLEM_DETECTOR: 'daemonset'\n" +
				"NODE_LABELS: a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true\n" +
				"DNS_SERVER_IP: '10.0.0.10'\n" +
				"NODE_TAINTS: 'dedicated=ml:NoSchedule,test=dev:PreferNoSchedule,a=b:c'\n",
			expect: expectedTaints,
		},
		{
			desc: "from AUTOSCALER_ENV_VARS.node_taints",
			env: "ENABLE_NODE_PROBLEM_DETECTOR: 'daemonset'\n" +
				"DNS_SERVER_IP: '10.0.0.10'\n" +
				"AUTOSCALER_ENV_VARS: node_labels=a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true;" +
				"node_taints='dedicated=ml:NoSchedule,test=dev:PreferNoSchedule,a=b:c';" +
				"kube_reserved=cpu=1000m,memory=3000`00Mi;" +
				"os=linux\n",
			expect: expectedTaints,
		},
		{
			desc: "from empty AUTOSCALER_ENV_VARS.node_taints",
			env: "ENABLE_NODE_PROBLEM_DETECTOR: 'daemonset'\n" +
				"DNS_SERVER_IP: '10.0.0.10'\n" +
				"AUTOSCALER_ENV_VARS: node_labels=a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true;" +
				"node_taints=\n",
			expect: makeTaintSet([]apiv1.Taint{}),
		},
		{
			desc: "malformed key-value in AUTOSCALER_ENV_VARS.node_taints",
			env:  "AUTOSCALER_ENV_VARS: node_taints='dedicatedml:NoSchedule,test=dev:PreferNoSchedule,a=b:c'\n",
			err:  fmt.Errorf("error while parsing key-value list, val: dedicatedml:NoSchedule"),
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			taints, err := extractTaintsFromKubeEnv(c.env)
			assert.Equal(t, c.err, err)
			if c.err != nil {
				return
			}
			assert.Equal(t, c.expect, makeTaintSet(taints))

			template := &gce.InstanceTemplate{
				Properties: &gce.InstanceProperties{
					Metadata: &gce.Metadata{
						Items: []*gce.MetadataItems{
							{
								Key:   "kube-env",
								Value: &c.env,
							},
						},
					},
				},
			}

			taints, err = GetTaintsFromTemplate(template)
			assert.Equal(t, c.err, err)
			if c.err != nil {
				return
			}
			assert.Equal(t, c.expect, makeTaintSet(taints))
		})
	}

}

func TestExtractKubeReservedFromKubeEnv(t *testing.T) {
	type testCase struct {
		kubeEnv          string
		expectedReserved string
		expectedErr      bool
	}

	testCases := []testCase{
		{
			kubeEnv: "ENABLE_NODE_PROBLEM_DETECTOR: 'daemonset'\n" +
				"NODE_LABELS: a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true\n" +
				"DNS_SERVER_IP: '10.0.0.10'\n" +
				"KUBELET_TEST_ARGS: --experimental-allocatable-ignore-eviction --kube-reserved=cpu=1000m,memory=300000Mi\n" +
				"NODE_TAINTS: 'dedicated=ml:NoSchedule,test=dev:PreferNoSchedule,a=b:c'\n",
			expectedReserved: "cpu=1000m,memory=300000Mi",
			expectedErr:      false,
		},
		{
			kubeEnv: "ENABLE_NODE_PROBLEM_DETECTOR: 'daemonset'\n" +
				"DNS_SERVER_IP: '10.0.0.10'\n" +
				"AUTOSCALER_ENV_VARS: node_labels=a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true;" +
				"node_taints='dedicated=ml:NoSchedule,test=dev:PreferNoSchedule,a=b:c';" +
				"kube_reserved=cpu=1000m,memory=300000Mi;" +
				"os=linux\n" +
				"KUBELET_TEST_ARGS: --experimental-allocatable-ignore-eviction\n",
			expectedReserved: "cpu=1000m,memory=300000Mi",
			expectedErr:      false,
		},
		{
			// Multi-line KUBELET_ARGS
			kubeEnv: "ENABLE_NODE_PROBLEM_DETECTOR: 'daemonset'\n" +
				"DNS_SERVER_IP: '10.0.0.10'\n" +
				"AUTOSCALER_ENV_VARS: node_labels=a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true;" +
				"node_taints='dedicated=ml:NoSchedule,test=dev:PreferNoSchedule,a=b:c';" +
				"kube_reserved=cpu=1000m,memory=300000Mi\n" +
				"KUBELET_ARGS: --experimental-allocatable-ignore-eviction\n" +
				" --kube_reserved=cpu=1000m,memory=300000Mi\n",
			expectedReserved: "cpu=1000m,memory=300000Mi",
			expectedErr:      false,
		},
		{
			kubeEnv: "ENABLE_NODE_PROBLEM_DETECTOR: 'daemonset'\n" +
				"NODE_LABELS: a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true\n" +
				"DNS_SERVER_IP: '10.0.0.10'\n" +
				"KUBELET_TEST_ARGS: --experimental-allocatable-ignore-eviction\n" +
				"NODE_TAINTS: 'dedicated=ml:NoSchedule,test=dev:PreferNoSchedule,a=b:c'\n",
			expectedReserved: "",
			expectedErr:      true,
		},
		{
			kubeEnv: "ENABLE_NODE_PROBLEM_DETECTOR: 'daemonset'\n" +
				"NODE_LABELS: a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true\n" +
				"DNS_SERVER_IP: '10.0.0.10'\n" +
				"NODE_TAINTS: 'dedicated=ml:NoSchedule,test=dev:PreferNoSchedule,a=b:c'\n",
			expectedReserved: "",
			expectedErr:      true,
		},
	}

	for _, tc := range testCases {
		reserved, err := extractKubeReservedFromKubeEnv(tc.kubeEnv)
		assert.Equal(t, tc.expectedReserved, reserved)
		if tc.expectedErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
	}
}

func TestExtractOperatingSystemFromKubeEnv(t *testing.T) {
	type testCase struct {
		name                    string
		kubeEnv                 string
		expectedOperatingSystem OperatingSystem
	}

	testCases := []testCase{
		{
			name: "linux",
			kubeEnv: "ENABLE_NODE_PROBLEM_DETECTOR: 'daemonset'\n" +
				"DNS_SERVER_IP: '10.0.0.10'\n" +
				"AUTOSCALER_ENV_VARS: node_labels=a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true;" +
				"node_taints='dedicated=ml:NoSchedule,test=dev:PreferNoSchedule,a=b:c';" +
				"kube_reserved=cpu=1000m,memory=300000Mi;" +
				"os=linux\n" +
				"KUBELET_TEST_ARGS: --experimental-allocatable-ignore-eviction\n",
			expectedOperatingSystem: OperatingSystemLinux,
		},
		{
			name: "windows",
			kubeEnv: "ENABLE_NODE_PROBLEM_DETECTOR: 'daemonset'\n" +
				"DNS_SERVER_IP: '10.0.0.10'\n" +
				"AUTOSCALER_ENV_VARS: node_labels=a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true" +
				"node_taints='dedicated=ml:NoSchedule,test=dev:PreferNoSchedule,a=b:c';" +
				"kube_reserved=cpu=1000m,memory=300000Mi;" +
				"os=windows\n" +
				"KUBELET_TEST_ARGS: --experimental-allocatable-ignore-eviction\n",
			expectedOperatingSystem: OperatingSystemWindows,
		},
		{
			name: "no AUTOSCALER_ENV_VARS",
			kubeEnv: "ENABLE_NODE_PROBLEM_DETECTOR: 'daemonset'\n" +
				"NODE_LABELS: a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true\n" +
				"DNS_SERVER_IP: '10.0.0.10'\n" +
				"KUBELET_TEST_ARGS: --experimental-allocatable-ignore-eviction --kube-reserved=cpu=1000m,memory=300000Mi\n" +
				"NODE_TAINTS: 'dedicated=ml:NoSchedule,test=dev:PreferNoSchedule,a=b:c'\n",
			expectedOperatingSystem: OperatingSystemDefault,
		},
		{
			name: "no os defined",
			kubeEnv: "ENABLE_NODE_PROBLEM_DETECTOR: 'daemonset'\n" +
				"DNS_SERVER_IP: '10.0.0.10'\n" +
				"AUTOSCALER_ENV_VARS: node_labels=a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true;" +
				"node_taints='dedicated=ml:NoSchedule,test=dev:PreferNoSchedule,a=b:c';" +
				"kube_reserved=cpu=1000m,memory=300000Mi\n" +
				"KUBELET_TEST_ARGS: --experimental-allocatable-ignore-eviction\n",
			expectedOperatingSystem: OperatingSystemDefault,
		},
		{
			name: "os is empty",
			kubeEnv: "ENABLE_NODE_PROBLEM_DETECTOR: 'daemonset'\n" +
				"DNS_SERVER_IP: '10.0.0.10'\n" +
				"AUTOSCALER_ENV_VARS: node_labels=a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true;" +
				"node_taints='dedicated=ml:NoSchedule,test=dev:PreferNoSchedule,a=b:c';" +
				"kube_reserved=cpu=1000m,memory=300000Mi;" +
				"os=\n" +
				"KUBELET_TEST_ARGS: --experimental-allocatable-ignore-eviction\n",
			expectedOperatingSystem: OperatingSystemUnknown,
		},
		{
			name: "unknown (macos)",
			kubeEnv: "ENABLE_NODE_PROBLEM_DETECTOR: 'daemonset'\n" +
				"DNS_SERVER_IP: '10.0.0.10'\n" +
				"AUTOSCALER_ENV_VARS: node_labels=a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true" +
				"node_taints='dedicated=ml:NoSchedule,test=dev:PreferNoSchedule,a=b:c';" +
				"kube_reserved=cpu=1000m,memory=300000Mi;" +
				"os=macos\n" +
				"KUBELET_TEST_ARGS: --experimental-allocatable-ignore-eviction\n",
			expectedOperatingSystem: OperatingSystemUnknown,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualOperatingSystem := extractOperatingSystemFromKubeEnv(tc.kubeEnv)
			assert.Equal(t, tc.expectedOperatingSystem, actualOperatingSystem)
		})
	}
}

func TestExtractOperatingSystemDistributionFromKubeEnv(t *testing.T) {
	type testCase struct {
		name                                string
		kubeEnv                             string
		expectedOperatingSystemDistribution OperatingSystemDistribution
	}

	testCases := []testCase{
		{
			name: "cos",
			kubeEnv: "ENABLE_NODE_PROBLEM_DETECTOR: 'daemonset'\n" +
				"DNS_SERVER_IP: '10.0.0.10'\n" +
				"AUTOSCALER_ENV_VARS: node_labels=a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true;" +
				"node_taints='dedicated=ml:NoSchedule,test=dev:PreferNoSchedule,a=b:c';" +
				"kube_reserved=cpu=1000m,memory=300000Mi;" +
				"os_distribution=cos\n" +
				"KUBELET_TEST_ARGS: --experimental-allocatable-ignore-eviction\n",
			expectedOperatingSystemDistribution: OperatingSystemDistributionCOS,
		},
		{
			name: "cos containerd",
			kubeEnv: "ENABLE_NODE_PROBLEM_DETECTOR: 'daemonset'\n" +
				"DNS_SERVER_IP: '10.0.0.10'\n" +
				"AUTOSCALER_ENV_VARS: node_labels=a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true;" +
				"node_taints='dedicated=ml:NoSchedule,test=dev:PreferNoSchedule,a=b:c';" +
				"kube_reserved=cpu=1000m,memory=300000Mi;" +
				"os_distribution=cos_containerd\n" +
				"KUBELET_TEST_ARGS: --experimental-allocatable-ignore-eviction\n",
			expectedOperatingSystemDistribution: OperatingSystemDistributionCOS,
		},
		{
			name: "ubuntu containerd",
			kubeEnv: "ENABLE_NODE_PROBLEM_DETECTOR: 'daemonset'\n" +
				"DNS_SERVER_IP: '10.0.0.10'\n" +
				"AUTOSCALER_ENV_VARS: node_labels=a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true;" +
				"node_taints='dedicated=ml:NoSchedule,test=dev:PreferNoSchedule,a=b:c';" +
				"kube_reserved=cpu=1000m,memory=300000Mi;" +
				"os_distribution=ubuntu_containerd\n" +
				"KUBELET_TEST_ARGS: --experimental-allocatable-ignore-eviction\n",
			expectedOperatingSystemDistribution: OperatingSystemDistributionUbuntu,
		},
		{
			name: "ubuntu",
			kubeEnv: "ENABLE_NODE_PROBLEM_DETECTOR: 'daemonset'\n" +
				"DNS_SERVER_IP: '10.0.0.10'\n" +
				"AUTOSCALER_ENV_VARS: node_labels=a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true;" +
				"node_taints='dedicated=ml:NoSchedule,test=dev:PreferNoSchedule,a=b:c';" +
				"kube_reserved=cpu=1000m,memory=300000Mi;" +
				"os_distribution=ubuntu\n" +
				"KUBELET_TEST_ARGS: --experimental-allocatable-ignore-eviction\n",
			expectedOperatingSystemDistribution: OperatingSystemDistributionUbuntu,
		},
		{
			name: "windows ltsc",
			kubeEnv: "ENABLE_NODE_PROBLEM_DETECTOR: 'daemonset'\n" +
				"DNS_SERVER_IP: '10.0.0.10'\n" +
				"AUTOSCALER_ENV_VARS: node_labels=a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true" +
				"node_taints='dedicated=ml:NoSchedule,test=dev:PreferNoSchedule,a=b:c';" +
				"kube_reserved=cpu=1000m,memory=300000Mi;" +
				"os_distribution=windows_ltsc\n" +
				"KUBELET_TEST_ARGS: --experimental-allocatable-ignore-eviction\n",
			expectedOperatingSystemDistribution: OperatingSystemDistributionWindowsLTSC,
		},
		{
			name: "windows sac",
			kubeEnv: "ENABLE_NODE_PROBLEM_DETECTOR: 'daemonset'\n" +
				"DNS_SERVER_IP: '10.0.0.10'\n" +
				"AUTOSCALER_ENV_VARS: node_labels=a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true" +
				"node_taints='dedicated=ml:NoSchedule,test=dev:PreferNoSchedule,a=b:c';" +
				"kube_reserved=cpu=1000m,memory=300000Mi;" +
				"os_distribution=windows_sac\n" +
				"KUBELET_TEST_ARGS: --experimental-allocatable-ignore-eviction\n",
			expectedOperatingSystemDistribution: OperatingSystemDistributionWindowsSAC,
		},
		{
			name: "no AUTOSCALER_ENV_VARS",
			kubeEnv: "ENABLE_NODE_PROBLEM_DETECTOR: 'daemonset'\n" +
				"NODE_LABELS: a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true\n" +
				"DNS_SERVER_IP: '10.0.0.10'\n" +
				"KUBELET_TEST_ARGS: --experimental-allocatable-ignore-eviction --kube-reserved=cpu=1000m,memory=300000Mi\n" +
				"NODE_TAINTS: 'dedicated=ml:NoSchedule,test=dev:PreferNoSchedule,a=b:c'\n",
			expectedOperatingSystemDistribution: OperatingSystemDistributionDefault,
		},
		{
			name: "no os distribution defined",
			kubeEnv: "ENABLE_NODE_PROBLEM_DETECTOR: 'daemonset'\n" +
				"DNS_SERVER_IP: '10.0.0.10'\n" +
				"AUTOSCALER_ENV_VARS: node_labels=a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true;" +
				"node_taints='dedicated=ml:NoSchedule,test=dev:PreferNoSchedule,a=b:c';" +
				"kube_reserved=cpu=1000m,memory=300000Mi\n" +
				"KUBELET_TEST_ARGS: --experimental-allocatable-ignore-eviction\n",
			expectedOperatingSystemDistribution: OperatingSystemDistributionDefault,
		},
		{
			name: "os distribution is empty",
			kubeEnv: "ENABLE_NODE_PROBLEM_DETECTOR: 'daemonset'\n" +
				"DNS_SERVER_IP: '10.0.0.10'\n" +
				"AUTOSCALER_ENV_VARS: node_labels=a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true;" +
				"node_taints='dedicated=ml:NoSchedule,test=dev:PreferNoSchedule,a=b:c';" +
				"kube_reserved=cpu=1000m,memory=300000Mi;" +
				"os_distribution=\n" +
				"KUBELET_TEST_ARGS: --experimental-allocatable-ignore-eviction\n",
			expectedOperatingSystemDistribution: OperatingSystemDistributionUnknown,
		},
		{
			name: "unknown (macos)",
			kubeEnv: "ENABLE_NODE_PROBLEM_DETECTOR: 'daemonset'\n" +
				"DNS_SERVER_IP: '10.0.0.10'\n" +
				"AUTOSCALER_ENV_VARS: node_labels=a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true" +
				"node_taints='dedicated=ml:NoSchedule,test=dev:PreferNoSchedule,a=b:c';" +
				"kube_reserved=cpu=1000m,memory=300000Mi;" +
				"os_distribution=macos\n" +
				"KUBELET_TEST_ARGS: --experimental-allocatable-ignore-eviction\n",
			expectedOperatingSystemDistribution: OperatingSystemDistributionUnknown,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualOperatingSystem := extractOperatingSystemDistributionFromKubeEnv(tc.kubeEnv)
			assert.Equal(t, tc.expectedOperatingSystemDistribution, actualOperatingSystem)
		})
	}
}

func TestExtractExtendedResourcesFromKubeEnv(t *testing.T) {
	type testCase struct {
		name                      string
		kubeEnv                   string
		expectedExtendedResources apiv1.ResourceList
		expectedErr               bool
	}

	testCases := []testCase{
		{
			name: "numeric value",
			kubeEnv: "AUTOSCALER_ENV_VARS: node_labels=a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true;" +
				"node_taints='dedicated=ml:NoSchedule,test=dev:PreferNoSchedule,a=b:c';" +
				"kube_reserved=cpu=1000m,memory=300000Mi;" +
				"extended_resources=foo=10",
			expectedExtendedResources: apiv1.ResourceList{
				apiv1.ResourceName("foo"): *resource.NewQuantity(10, resource.DecimalSI),
			},
			expectedErr: false,
		},
		{
			name: "numeric value with quantity suffix",
			kubeEnv: "AUTOSCALER_ENV_VARS: node_labels=a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true;" +
				"node_taints='dedicated=ml:NoSchedule,test=dev:PreferNoSchedule,a=b:c';" +
				"kube_reserved=cpu=1000m,memory=300000Mi;" +
				"extended_resources=foo=10G",
			expectedExtendedResources: apiv1.ResourceList{
				apiv1.ResourceName("foo"): *resource.NewQuantity(10*units.GB, resource.DecimalSI),
			},
			expectedErr: false,
		},
		{
			name: "multiple extended_resources definition",
			kubeEnv: "AUTOSCALER_ENV_VARS: node_labels=a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true;" +
				"node_taints='dedicated=ml:NoSchedule,test=dev:PreferNoSchedule,a=b:c';" +
				"kube_reserved=cpu=1000m,memory=300000Mi;" +
				"extended_resources=foo=10G,bar=230",
			expectedExtendedResources: apiv1.ResourceList{
				apiv1.ResourceName("foo"): *resource.NewQuantity(10*units.GB, resource.DecimalSI),
				apiv1.ResourceName("bar"): *resource.NewQuantity(230, resource.DecimalSI),
			},
			expectedErr: false,
		},
		{
			name: "invalid value",
			kubeEnv: "AUTOSCALER_ENV_VARS: node_labels=a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true;" +
				"node_taints='dedicated=ml:NoSchedule,test=dev:PreferNoSchedule,a=b:c';" +
				"kube_reserved=cpu=1000m,memory=300000Mi;" +
				"extended_resources=foo=bar",
			expectedExtendedResources: apiv1.ResourceList{},
			expectedErr:               false,
		},
		{
			name: "both valid and invalid values",
			kubeEnv: "AUTOSCALER_ENV_VARS: node_labels=a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true;" +
				"node_taints='dedicated=ml:NoSchedule,test=dev:PreferNoSchedule,a=b:c';" +
				"kube_reserved=cpu=1000m,memory=300000Mi;" +
				"extended_resources=foo=bar,baz=10G",
			expectedExtendedResources: apiv1.ResourceList{
				apiv1.ResourceName("baz"): *resource.NewQuantity(10*units.GB, resource.DecimalSI),
			},
			expectedErr: false,
		},
		{
			name: "invalid quantity suffix",
			kubeEnv: "AUTOSCALER_ENV_VARS: node_labels=a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true;" +
				"node_taints='dedicated=ml:NoSchedule,test=dev:PreferNoSchedule,a=b:c';" +
				"kube_reserved=cpu=1000m,memory=300000Mi;" +
				"extended_resources=foo=10Wi",
			expectedExtendedResources: apiv1.ResourceList{},
			expectedErr:               false,
		},
		{
			name: "malformed extended_resources map",
			kubeEnv: "AUTOSCALER_ENV_VARS: node_labels=a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true;" +
				"node_taints='dedicated=ml:NoSchedule,test=dev:PreferNoSchedule,a=b:c';" +
				"kube_reserved=cpu=1000m,memory=300000Mi;" +
				"extended_resources=foo",
			expectedExtendedResources: apiv1.ResourceList{},
			expectedErr:               true,
		},
		{
			name: "malformed extended_resources definition",
			kubeEnv: "AUTOSCALER_ENV_VARS: node_labels=a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true;" +
				"node_taints='dedicated=ml:NoSchedule,test=dev:PreferNoSchedule,a=b:c';" +
				"kube_reserved=cpu=1000m,memory=300000Mi;" +
				"extended_resources/",
			expectedExtendedResources: apiv1.ResourceList{},
			expectedErr:               true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			extendedResources, err := extractExtendedResourcesFromKubeEnv(tc.kubeEnv)
			assertEqualResourceLists(t, "Resources", tc.expectedExtendedResources, extendedResources)
			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}

}

func TestParseKubeReserved(t *testing.T) {
	type testCase struct {
		reserved                 string
		expectedCpu              string
		expectedMemory           string
		expectedEphemeralStorage string
		expectedErr              bool
	}
	testCases := []testCase{{
		reserved:                 "cpu=1000m,memory=300000Mi,ephemeral-storage=100Gi",
		expectedCpu:              "1000m",
		expectedMemory:           "300000Mi",
		expectedEphemeralStorage: "100Gi",
		expectedErr:              false,
	}, {
		reserved:                 "cpu=1000m,ignored=300Mi,memory=0,ephemeral-storage=10Gi",
		expectedCpu:              "1000m",
		expectedMemory:           "0",
		expectedEphemeralStorage: "10Gi",
		expectedErr:              false,
	}, {
		reserved:    "This is a wrong reserved",
		expectedErr: true,
	}}
	for _, tc := range testCases {
		resources, err := parseKubeReserved(tc.reserved)
		if tc.expectedErr {
			assert.Error(t, err)
			assert.Nil(t, resources)
		} else {
			assert.NoError(t, err)
			expectedResources, err := makeResourceList(tc.expectedCpu, tc.expectedMemory, 0, tc.expectedEphemeralStorage)
			assert.NoError(t, err)
			assertEqualResourceLists(t, "Resources", expectedResources, resources)
		}
	}
}

func TestToSystemArchitecture(t *testing.T) {
	for tn, tc := range map[string]struct {
		archName string
		wantArch SystemArchitecture
	}{
		"valid architecture is converted": {
			archName: "amd64",
			wantArch: Amd64,
		},
		"invalid architecture results in UnknownArchitecture": {
			archName: "some-arch",
			wantArch: UnknownArch,
		},
	} {
		t.Run(tn, func(t *testing.T) {
			gotArch := ToSystemArchitecture(tc.archName)
			if diff := cmp.Diff(tc.wantArch, gotArch); diff != "" {
				t.Errorf("ToSystemArchitecture diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestExtractSystemArchitectureFromKubeEnv(t *testing.T) {
	for tn, tc := range map[string]struct {
		kubeEnv  string
		wantArch SystemArchitecture
		wantErr  error
	}{
		"valid arch defined in AUTOSCALER_ENV_VARS": {
			kubeEnv:  "AUTOSCALER_ENV_VARS: os_distribution=cos;arch=arm64;os=linux\n",
			wantArch: Arm64,
		},
		"invalid arch defined in AUTOSCALER_ENV_VARS": {
			kubeEnv:  "AUTOSCALER_ENV_VARS: os_distribution=cos;arch=blah;os=linux\n",
			wantArch: UnknownArch,
			wantErr:  cmpopts.AnyError,
		},
		"empty arch defined in AUTOSCALER_ENV_VARS": {
			kubeEnv:  "AUTOSCALER_ENV_VARS: os_distribution=cos;arch=;os=linux\n",
			wantArch: UnknownArch,
			wantErr:  cmpopts.AnyError,
		},

		"no arch defined in AUTOSCALER_ENV_VARS": {
			kubeEnv:  "AUTOSCALER_ENV_VARS: os_distribution=cos;os=linux\n",
			wantArch: UnknownArch,
			wantErr:  cmpopts.AnyError,
		},
		"KUBE_ENV parsing error": {
			kubeEnv:  "some-invalid-string",
			wantArch: UnknownArch,
			wantErr:  cmpopts.AnyError,
		},
	} {
		t.Run(tn, func(t *testing.T) {
			gotArch, gotErr := extractSystemArchitectureFromKubeEnv(tc.kubeEnv)
			if diff := cmp.Diff(tc.wantArch, gotArch); diff != "" {
				t.Errorf("extractSystemArchitectureFromKubeEnv diff (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tc.wantErr, gotErr, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("extractSystemArchitectureFromKubeEnv error diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestBuildNodeFromTemplateArch(t *testing.T) {
	for tn, tc := range map[string]struct {
		kubeEnv  string
		wantArch SystemArchitecture
	}{
		"valid arch defined in KUBE_ENV is passed through": {
			kubeEnv:  "AUTOSCALER_ENV_VARS: os_distribution=cos;arch=arm64;os=linux\n",
			wantArch: Arm64,
		},
		"invalid arch defined in KUBE_ENV is defaulted to the default arch": {
			kubeEnv:  "AUTOSCALER_ENV_VARS: os_distribution=cos;arch=;os=linux\n",
			wantArch: DefaultArch,
		},
	} {
		t.Run(tn, func(t *testing.T) {
			mig := &gceMig{gceRef: GceRef{Name: "some-name", Project: "some-proj", Zone: "us-central1-b"}}
			template := &gce.InstanceTemplate{
				Name: "node-name",
				Properties: &gce.InstanceProperties{
					Metadata: &gce.Metadata{
						Items: []*gce.MetadataItems{{Key: "kube-env", Value: &tc.kubeEnv}},
					},
					Disks: []*gce.AttachedDisk{},
				},
			}
			tb := &GceTemplateBuilder{}
			migOsInfo, gotErr := tb.MigOsInfo(mig.Id(), template)
			if gotErr != nil {
				t.Fatalf("MigOsInfo unexpected error: %v", gotErr)
			}
			gotNode, gotErr := tb.BuildNodeFromTemplate(mig, migOsInfo, template, 16, 128, nil, &GceReserved{})
			if gotErr != nil {
				t.Fatalf("BuildNodeFromTemplate unexpected error: %v", gotErr)
			}
			gotArch := gotNode.Labels[apiv1.LabelArchStable]
			if diff := cmp.Diff(tc.wantArch.Name(), gotArch); diff != "" {
				t.Errorf("BuildNodeFromTemplate arch label diff (-want +got):\n%s", diff)
			}
		})
	}
}

func makeTaintSet(taints []apiv1.Taint) map[apiv1.Taint]bool {
	set := make(map[apiv1.Taint]bool)
	for _, taint := range taints {
		set[taint] = true
	}
	return set
}

func makeResourceList(cpu string, memory string, gpu int64, ephemeralStorage string) (apiv1.ResourceList, error) {
	result := apiv1.ResourceList{}
	resultCpu, err := resource.ParseQuantity(cpu)
	if err != nil {
		return nil, err
	}
	result[apiv1.ResourceCPU] = resultCpu
	resultMemory, err := resource.ParseQuantity(memory)
	if err != nil {
		return nil, err
	}
	result[apiv1.ResourceMemory] = resultMemory
	if gpu > 0 {
		resultGpu := *resource.NewQuantity(gpu, resource.DecimalSI)
		if err != nil {
			return nil, err
		}
		result[gpuUtils.ResourceNvidiaGPU] = resultGpu
	}
	if len(ephemeralStorage) != 0 {
		resultEphemeralStorage, err := resource.ParseQuantity(ephemeralStorage)
		if err != nil {
			return nil, err
		}
		result[apiv1.ResourceEphemeralStorage] = resultEphemeralStorage
	}
	return result, nil
}

func makeResourceList2(cpu int64, memory int64, gpu int64, pods int64) (apiv1.ResourceList, error) {
	result := apiv1.ResourceList{}
	result[apiv1.ResourceCPU] = *resource.NewQuantity(cpu, resource.DecimalSI)
	result[apiv1.ResourceMemory] = *resource.NewQuantity(memory, resource.BinarySI)
	if gpu > 0 {
		result[gpuUtils.ResourceNvidiaGPU] = *resource.NewQuantity(gpu, resource.DecimalSI)
	}
	if pods > 0 {
		result[apiv1.ResourcePods] = *resource.NewQuantity(pods, resource.DecimalSI)
	}
	return result, nil
}

func assertEqualResourceLists(t *testing.T, name string, expected, actual apiv1.ResourceList) {
	t.Helper()
	assert.True(t, quota.Equals(expected, actual),
		"%q unequal:\nExpected: %v\nActual:   %v", name, stringifyResourceList(expected), stringifyResourceList(actual))
}

func stringifyResourceList(resourceList apiv1.ResourceList) string {
	resourceNames := []apiv1.ResourceName{
		apiv1.ResourcePods, apiv1.ResourceCPU, gpuUtils.ResourceNvidiaGPU, apiv1.ResourceMemory, apiv1.ResourceEphemeralStorage}
	var results []string
	for _, name := range resourceNames {
		quantity, found := resourceList[name]
		if found {
			value := quantity.Value()
			if name == apiv1.ResourceCPU {
				value = quantity.MilliValue()
			}
			results = append(results, fmt.Sprintf("%v: %v", string(name), value))
		}
	}
	return strings.Join(results, ", ")
}
