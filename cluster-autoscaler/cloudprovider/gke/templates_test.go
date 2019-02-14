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

package gke

import (
	"fmt"
	"strings"
	"testing"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/gce"
	gpuUtils "k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/autoscaler/cluster-autoscaler/utils/units"

	gce_api "google.golang.org/api/compute/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	kubeletapis "k8s.io/kubernetes/pkg/kubelet/apis"
	"k8s.io/kubernetes/pkg/quota"

	"github.com/stretchr/testify/assert"
)

func TestBuildNodeFromTemplateSetsResources(t *testing.T) {
	type testCase struct {
		scenario         string
		kubeEnv          string
		accelerators     []*gce_api.AcceleratorConfig
		mig              gce.Mig
		physicalCpu      int64
		physicalMemory   int64
		kubeReserved     bool
		reservedCpu      string
		reservedMemory   string
		expectedGpuCount int64
		expectedErr      bool
	}
	testCases := []testCase{
		{
			scenario: "kube-reserved present in kube-env",
			kubeEnv: "ENABLE_NODE_PROBLEM_DETECTOR: 'daemonset'\n" +
				"NODE_LABELS: a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true\n" +
				"DNS_SERVER_IP: '10.0.0.10'\n" +
				fmt.Sprintf("KUBELET_TEST_ARGS: --experimental-allocatable-ignore-eviction --kube-reserved=cpu=1000m,memory=%v\n", 1*units.MiB) +
				"NODE_TAINTS: 'dedicated=ml:NoSchedule,test=dev:PreferNoSchedule,a=b:c'\n",
			accelerators: []*gce_api.AcceleratorConfig{
				{AcceleratorType: "nvidia-tesla-k80", AcceleratorCount: 3},
				{AcceleratorType: "nvidia-tesla-p100", AcceleratorCount: 8},
			},
			physicalCpu:      8,
			physicalMemory:   200 * units.MiB,
			kubeReserved:     true,
			reservedCpu:      "1000m",
			reservedMemory:   fmt.Sprintf("%v", 1*units.MiB),
			expectedGpuCount: 11,
			expectedErr:      false,
		},
		{
			scenario: "no kube-reserved in kube-env",
			kubeEnv: "ENABLE_NODE_PROBLEM_DETECTOR: 'daemonset'\n" +
				"NODE_LABELS: a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true\n" +
				"DNS_SERVER_IP: '10.0.0.10'\n" +
				"NODE_TAINTS: 'dedicated=ml:NoSchedule,test=dev:PreferNoSchedule,a=b:c'\n",
			physicalCpu:      8,
			physicalMemory:   200 * units.MiB,
			kubeReserved:     false,
			expectedGpuCount: 11,
			expectedErr:      false,
		}, {
			scenario:    "totally messed up kube-env",
			kubeEnv:     "This kube-env is totally messed up",
			expectedErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.scenario, func(t *testing.T) {
			tb := &GkeTemplateBuilder{}
			mig := &GkeMig{
				gceRef: gce.GceRef{
					Name:    "some-name",
					Project: "some-proj",
					Zone:    "us-central1-b",
				},
			}
			template := &gce_api.InstanceTemplate{
				Name: "node-name",
				Properties: &gce_api.InstanceProperties{
					GuestAccelerators: tc.accelerators,
					Metadata: &gce_api.Metadata{
						Items: []*gce_api.MetadataItems{{Key: "kube-env", Value: &tc.kubeEnv}},
					},
					MachineType: "irrelevant-type",
				},
			}
			node, err := tb.BuildNodeFromTemplate(mig, template, tc.physicalCpu, tc.physicalMemory)
			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				capacity, err := tb.BuildCapacity(tc.physicalCpu, tc.physicalMemory, tc.accelerators)
				assert.NoError(t, err)
				assertEqualResourceLists(t, "Capacity", capacity, node.Status.Capacity)
				if !tc.kubeReserved {
					assertEqualResourceLists(t, "Allocatable", capacity, node.Status.Allocatable)
				} else {
					reserved, err := makeResourceList(tc.reservedCpu, tc.reservedMemory, 0)
					assert.NoError(t, err)
					allocatable := tb.CalculateAllocatable(capacity, reserved)
					assertEqualResourceLists(t, "Allocatable", allocatable, node.Status.Allocatable)
				}
			}
		})
	}
}

func TestBuildLabelsForAutoprovisionedMigOK(t *testing.T) {
	labels, err := buildLabelsForAutoprovisionedMig(
		&GkeMig{
			gceRef: gce.GceRef{
				Name:    "kubernetes-minion-autoprovisioned-group",
				Project: "mwielgus-proj",
				Zone:    "us-central1-b",
			},
			autoprovisioned: true,
			spec: &MigSpec{
				MachineType: "n1-standard-8",
				Labels: map[string]string{
					"A": "B",
				},
			},
		},
		"sillyname",
	)

	assert.Nil(t, err)
	assert.Equal(t, "B", labels["A"])
	assert.Equal(t, "us-central1", labels[kubeletapis.LabelZoneRegion])
	assert.Equal(t, "us-central1-b", labels[kubeletapis.LabelZoneFailureDomain])
	assert.Equal(t, "sillyname", labels[kubeletapis.LabelHostname])
	assert.Equal(t, "n1-standard-8", labels[kubeletapis.LabelInstanceType])
	assert.Equal(t, cloudprovider.DefaultArch, labels[kubeletapis.LabelArch])
	assert.Equal(t, cloudprovider.DefaultOS, labels[kubeletapis.LabelOS])
}

func TestBuildLabelsForAutoprovisionedMigConflict(t *testing.T) {
	_, err := buildLabelsForAutoprovisionedMig(
		&GkeMig{
			gceRef: gce.GceRef{
				Name:    "kubernetes-minion-autoprovisioned-group",
				Project: "mwielgus-proj",
				Zone:    "us-central1-b",
			},
			autoprovisioned: true,
			spec: &MigSpec{
				MachineType: "n1-standard-8",
				Labels: map[string]string{
					kubeletapis.LabelOS: "windows",
				},
			},
		},
		"sillyname",
	)
	assert.Error(t, err)
}

func TestBuildAllocatableFromKubeEnv(t *testing.T) {
	type testCase struct {
		kubeEnv        string
		capacityCpu    string
		capacityMemory string
		expectedCpu    string
		expectedMemory string
		gpuCount       int64
		expectedErr    bool
	}
	testCases := []testCase{{
		kubeEnv: "ENABLE_NODE_PROBLEM_DETECTOR: 'daemonset'\n" +
			"NODE_LABELS: a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true\n" +
			"DNS_SERVER_IP: '10.0.0.10'\n" +
			"KUBELET_TEST_ARGS: --experimental-allocatable-ignore-eviction --kube-reserved=cpu=1000m,memory=300000Mi\n" +
			"NODE_TAINTS: 'dedicated=ml:NoSchedule,test=dev:PreferNoSchedule,a=b:c'\n",
		capacityCpu:    "4000m",
		capacityMemory: "700000Mi",
		expectedCpu:    "3000m",
		expectedMemory: "399900Mi", // capacityMemory-kube_reserved-kubeletEvictionHardMemory
		gpuCount:       10,
		expectedErr:    false,
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
		capacity, err := makeResourceList(tc.capacityCpu, tc.capacityMemory, tc.gpuCount)
		assert.NoError(t, err)
		tb := GkeTemplateBuilder{}
		allocatable, err := tb.BuildAllocatableFromKubeEnv(capacity, tc.kubeEnv)
		if tc.expectedErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			expectedResources, err := makeResourceList(tc.expectedCpu, tc.expectedMemory, tc.gpuCount)
			assert.NoError(t, err)
			for res, expectedQty := range expectedResources {
				qty, found := allocatable[res]
				assert.True(t, found)
				assert.Equal(t, qty.Value(), expectedQty.Value())
			}
		}
	}
}

func TestBuildKubeReserved(t *testing.T) {
	type testCase struct {
		physicalCpu    int64
		reservedCpu    string
		physicalMemory int64
		reservedMemory string
	}
	testCases := []testCase{{
		physicalCpu: 16,
		reservedCpu: "110m",
		// Below threshold for reserving memory
		physicalMemory: units.GB,
		reservedMemory: fmt.Sprintf("%v", 255*units.MiB),
	}, {
		physicalCpu: 1,
		reservedCpu: "60m",
		// 10760Mi = 0.25*4000Mi + 0.2*4000Mi + 0.1*8000Mi + 0.06*112000Mi + 0.02*72000Mi
		physicalMemory: 200 * 1000 * units.MiB,
		reservedMemory: fmt.Sprintf("%v", 10760*units.MiB),
	}}
	for _, tc := range testCases {
		tb := GkeTemplateBuilder{}
		expectedReserved, err := makeResourceList(tc.reservedCpu, tc.reservedMemory, 0)
		assert.NoError(t, err)
		kubeReserved := tb.BuildKubeReserved(tc.physicalCpu, tc.physicalMemory)
		assertEqualResourceLists(t, "Kube reserved", expectedReserved, kubeReserved)
	}
}

func makeResourceList(cpu string, memory string, gpu int64) (apiv1.ResourceList, error) {
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
	return result, nil
}

func assertEqualResourceLists(t *testing.T, name string, expected, actual apiv1.ResourceList) {
	t.Helper()
	assert.True(t, quota.V1Equals(expected, actual),
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
