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
	"testing"

	"github.com/stretchr/testify/assert"
	gce "google.golang.org/api/compute/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	kubeletapis "k8s.io/kubernetes/pkg/kubelet/apis"
	"k8s.io/kubernetes/pkg/quota"
)

func TestBuildNodeFromTemplateSetsResources(t *testing.T) {
	type testCase struct {
		kubeEnv           string
		name              string
		machineType       string
		accelerators      []*gce.AcceleratorConfig
		mig               *Mig
		capacityCpu       string
		capacityMemory    string
		allocatableCpu    string
		allocatableMemory string
		gpuCount          int64
		expectedErr       bool
	}
	testCases := []testCase{{
		kubeEnv: "ENABLE_NODE_PROBLEM_DETECTOR: 'daemonset'\n" +
			"NODE_LABELS: a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true\n" +
			"DNS_SERVER_IP: '10.0.0.10'\n" +
			fmt.Sprintf("KUBELET_TEST_ARGS: --experimental-allocatable-ignore-eviction --kube-reserved=cpu=1000m,memory=%v\n", 1024*1024) +
			"NODE_TAINTS: 'dedicated=ml:NoSchedule,test=dev:PreferNoSchedule,a=b:c'\n",
		name:        "nodeName",
		machineType: "custom-8-2",
		accelerators: []*gce.AcceleratorConfig{
			{AcceleratorType: "nvidia-tesla-k80", AcceleratorCount: 3},
			{AcceleratorType: "nvidia-tesla-p100", AcceleratorCount: 8},
		},
		mig: &Mig{GceRef: GceRef{
			Name:    "some-name",
			Project: "some-proj",
			Zone:    "us-central1-b"}},
		capacityCpu:       "8000m",
		capacityMemory:    fmt.Sprintf("%v", 2*1024*1024),
		allocatableCpu:    "7000m",
		allocatableMemory: fmt.Sprintf("%v", 1024*1024),
		gpuCount:          11,
		expectedErr:       false,
	}, {
		kubeEnv: "ENABLE_NODE_PROBLEM_DETECTOR: 'daemonset'\n" +
			"NODE_LABELS: a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true\n" +
			"DNS_SERVER_IP: '10.0.0.10'\n" +
			"NODE_TAINTS: 'dedicated=ml:NoSchedule,test=dev:PreferNoSchedule,a=b:c'\n",
		name:        "nodeName",
		machineType: "custom-8-2",
		mig: &Mig{GceRef: GceRef{
			Name:    "some-name",
			Project: "some-proj",
			Zone:    "us-central1-b"}},
		capacityCpu:       "8000m",
		capacityMemory:    fmt.Sprintf("%v", 2*1024*1024),
		allocatableCpu:    "8000m",
		allocatableMemory: fmt.Sprintf("%v", 2*1024*1024),
		expectedErr:       false,
	}, {
		kubeEnv:     "This kube-env is totally messed up",
		name:        "nodeName",
		machineType: "custom-8-2",
		mig: &Mig{GceRef: GceRef{
			Name:    "some-name",
			Project: "some-proj",
			Zone:    "us-central1-b"}},
		expectedErr: true,
	},
	}
	for _, tc := range testCases {
		tb := &templateBuilder{}
		template := &gce.InstanceTemplate{
			Name: tc.name,
			Properties: &gce.InstanceProperties{
				GuestAccelerators: tc.accelerators,
				Metadata: &gce.Metadata{
					Items: []*gce.MetadataItems{{Key: "kube-env", Value: &tc.kubeEnv}},
				},
				MachineType: tc.machineType,
			},
		}
		node, err := tb.buildNodeFromTemplate(tc.mig, template)
		if tc.expectedErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			podsQuantity, _ := resource.ParseQuantity("110")
			capacity, err := makeResourceList(tc.capacityCpu, tc.capacityMemory, tc.gpuCount)
			capacity[apiv1.ResourcePods] = podsQuantity
			assert.NoError(t, err)
			allocatable, err := makeResourceList(tc.allocatableCpu, tc.allocatableMemory, tc.gpuCount)
			allocatable[apiv1.ResourcePods] = podsQuantity
			assert.NoError(t, err)
			assertEqualResourceLists(t, "Capacity", capacity, node.Status.Capacity)
			assertEqualResourceLists(t, "Allocatable", allocatable, node.Status.Allocatable)
		}
	}
}

func TestBuildGenericLabels(t *testing.T) {
	labels, err := buildGenericLabels(GceRef{
		Name:    "kubernetes-minion-group",
		Project: "mwielgus-proj",
		Zone:    "us-central1-b"},
		"n1-standard-8", "sillyname")
	assert.Nil(t, err)
	assert.Equal(t, "us-central1", labels[kubeletapis.LabelZoneRegion])
	assert.Equal(t, "us-central1-b", labels[kubeletapis.LabelZoneFailureDomain])
	assert.Equal(t, "sillyname", labels[kubeletapis.LabelHostname])
	assert.Equal(t, "n1-standard-8", labels[kubeletapis.LabelInstanceType])
	assert.Equal(t, cloudprovider.DefaultArch, labels[kubeletapis.LabelArch])
	assert.Equal(t, cloudprovider.DefaultOS, labels[kubeletapis.LabelOS])
}

func TestBuildLabelsForAutoscaledMigOK(t *testing.T) {
	labels, err := buildLablesForAutoprovisionedMig(
		&Mig{
			autoprovisioned: true,
			spec: &autoprovisioningSpec{
				machineType: "n1-standard-8",
				labels: map[string]string{
					"A": "B",
				},
			},
			GceRef: GceRef{
				Name:    "kubernetes-minion-autoprovisioned-group",
				Project: "mwielgus-proj",
				Zone:    "us-central1-b"}},
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

func TestBuildLabelsForAutoscaledMigConflict(t *testing.T) {
	_, err := buildLablesForAutoprovisionedMig(
		&Mig{
			autoprovisioned: true,
			spec: &autoprovisioningSpec{
				machineType: "n1-standard-8",
				labels: map[string]string{
					kubeletapis.LabelOS: "windows",
				},
			},
			GceRef: GceRef{
				Name:    "kubernetes-minion-autoprovisioned-group",
				Project: "mwielgus-proj",
				Zone:    "us-central1-b"}},
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
		gcuCount       int64
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
		expectedMemory: "400000Mi",
		gcuCount:       10,
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
		capacity, err := makeResourceList(tc.capacityCpu, tc.capacityMemory, tc.gcuCount)
		assert.NoError(t, err)
		tb := templateBuilder{}
		allocatable, err := tb.buildAllocatableFromKubeEnv(capacity, tc.kubeEnv)
		if tc.expectedErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			expectedResources, err := makeResourceList(tc.expectedCpu, tc.expectedMemory, tc.gcuCount)
			assert.NoError(t, err)
			assert.Equal(t, expectedResources, allocatable)
		}
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
		tb := templateBuilder{}
		assert.Equal(t, tc.count, tb.getAcceleratorCount(tc.accelerators))
	}
}

func TestBuildAllocatableFromCapacity(t *testing.T) {
	type testCase struct {
		capacityCpu       string
		capacityMemory    string
		allocatableCpu    string
		allocatableMemory string
		gpuCount          int64
	}
	testCases := []testCase{{
		capacityCpu:       "16000m",
		capacityMemory:    fmt.Sprintf("%v", 1*1024*1024*1024),
		allocatableCpu:    "15890m",
		allocatableMemory: fmt.Sprintf("%v", 0.75*1024*1024*1024),
		gpuCount:          1,
	}, {
		capacityCpu:       "500m",
		capacityMemory:    fmt.Sprintf("%v", 200*1000*1024*1024),
		allocatableCpu:    "470m",
		allocatableMemory: fmt.Sprintf("%v", (200*1000-10760)*1024*1024),
	}}
	for _, tc := range testCases {
		tb := templateBuilder{}
		capacity, err := makeResourceList(tc.capacityCpu, tc.capacityMemory, tc.gpuCount)
		assert.NoError(t, err)
		expectedAllocatable, err := makeResourceList(tc.allocatableCpu, tc.allocatableMemory, tc.gpuCount)
		assert.NoError(t, err)
		allocatable := tb.buildAllocatableFromCapacity(capacity)
		assertEqualResourceLists(t, "Allocatable", expectedAllocatable, allocatable)
	}
}

func TestExtractLabelsFromKubeEnv(t *testing.T) {
	kubeenv := "ENABLE_NODE_PROBLEM_DETECTOR: 'daemonset'\n" +
		"NODE_LABELS: a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true\n" +
		"DNS_SERVER_IP: '10.0.0.10'\n"

	labels, err := extractLabelsFromKubeEnv(kubeenv)
	assert.Nil(t, err)
	assert.Equal(t, 4, len(labels))
	assert.Equal(t, "b", labels["a"])
	assert.Equal(t, "d", labels["c"])
	assert.Equal(t, "pool-3", labels["cloud.google.com/gke-nodepool"])
	assert.Equal(t, "true", labels["cloud.google.com/gke-preemptible"])
}

func TestExtractTaintsFromKubeEnv(t *testing.T) {
	kubeenv := "ENABLE_NODE_PROBLEM_DETECTOR: 'daemonset'\n" +
		"NODE_LABELS: a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true\n" +
		"DNS_SERVER_IP: '10.0.0.10'\n" +
		"NODE_TAINTS: 'dedicated=ml:NoSchedule,test=dev:PreferNoSchedule,a=b:c'\n"

	expectedTaints := []apiv1.Taint{
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
	}

	taints, err := extractTaintsFromKubeEnv(kubeenv)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(taints))
	assert.Equal(t, makeTaintSet(expectedTaints), makeTaintSet(taints))

}

func TestExtractKubeReservedFromKubeEnv(t *testing.T) {
	type testCase struct {
		kubeEnv          string
		expectedReserved string
		expectedErr      bool
	}

	testCases := []testCase{{
		kubeEnv: "ENABLE_NODE_PROBLEM_DETECTOR: 'daemonset'\n" +
			"NODE_LABELS: a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true\n" +
			"DNS_SERVER_IP: '10.0.0.10'\n" +
			"KUBELET_TEST_ARGS: --experimental-allocatable-ignore-eviction --kube-reserved=cpu=1000m,memory=300000Mi\n" +
			"NODE_TAINTS: 'dedicated=ml:NoSchedule,test=dev:PreferNoSchedule,a=b:c'\n",
		expectedReserved: "cpu=1000m,memory=300000Mi",
		expectedErr:      false,
	}, {
		kubeEnv: "ENABLE_NODE_PROBLEM_DETECTOR: 'daemonset'\n" +
			"NODE_LABELS: a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true\n" +
			"DNS_SERVER_IP: '10.0.0.10'\n" +
			"KUBELET_TEST_ARGS: --experimental-allocatable-ignore-eviction\n" +
			"NODE_TAINTS: 'dedicated=ml:NoSchedule,test=dev:PreferNoSchedule,a=b:c'\n",
		expectedReserved: "",
		expectedErr:      true,
	}, {
		kubeEnv: "ENABLE_NODE_PROBLEM_DETECTOR: 'daemonset'\n" +
			"NODE_LABELS: a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true\n" +
			"DNS_SERVER_IP: '10.0.0.10'\n" +
			"NODE_TAINTS: 'dedicated=ml:NoSchedule,test=dev:PreferNoSchedule,a=b:c'\n",
		expectedReserved: "",
		expectedErr:      true}}

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

func TestParseKubeReserved(t *testing.T) {
	type testCase struct {
		reserved       string
		expectedCpu    string
		expectedMemory string
		expectedErr    bool
	}
	testCases := []testCase{{
		reserved:       "cpu=1000m,memory=300000Mi",
		expectedCpu:    "1000m",
		expectedMemory: "300000Mi",
		expectedErr:    false,
	}, {
		reserved:       "cpu=1000m,ignored=300Mi,memory=0",
		expectedCpu:    "1000m",
		expectedMemory: "0",
		expectedErr:    false,
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
			expectedResources, err := makeResourceList(tc.expectedCpu, tc.expectedMemory, 0)
			assert.NoError(t, err)
			assertEqualResourceLists(t, "Resources", expectedResources, resources)
		}
	}
}

func TestCalculateReserved(t *testing.T) {
	type testCase struct {
		name             string
		function         func(capacity int64) int64
		capacity         int64
		expectedReserved int64
	}
	testCases := []testCase{
		{
			name:             "zero memory capacity",
			function:         memoryReservedMB,
			capacity:         0,
			expectedReserved: 0,
		},
		{
			name:             "between memory thresholds",
			function:         memoryReservedMB,
			capacity:         2 * mbPerGB,
			expectedReserved: 500, // 0.5 Gb
		},
		{
			name:             "at a memory threshold boundary",
			function:         memoryReservedMB,
			capacity:         8 * mbPerGB,
			expectedReserved: 1800, // 1.8 Gb
		},
		{
			name:             "exceeds highest memory threshold",
			function:         memoryReservedMB,
			capacity:         200 * mbPerGB,
			expectedReserved: 10760, // 10.8 Gb
		},
		{
			name:             "cpu sanity check",
			function:         cpuReservedMillicores,
			capacity:         4 * millicoresPerCore,
			expectedReserved: 80,
		},
	}
	for _, tc := range testCases {
		if actualReserved := tc.function(tc.capacity); actualReserved != tc.expectedReserved {
			t.Errorf("Test case: %s, Got f(%d Mb) = %d.  Want %d", tc.name, tc.capacity, actualReserved, tc.expectedReserved)
		}
	}
}

func makeTaintSet(taints []apiv1.Taint) map[apiv1.Taint]bool {
	set := make(map[apiv1.Taint]bool)
	for _, taint := range taints {
		set[taint] = true
	}
	return set
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
		result[apiv1.ResourceNvidiaGPU] = resultGpu
	}
	return result, nil
}

func assertEqualResourceLists(t *testing.T, name string, expected, actual apiv1.ResourceList) {
	assert.True(t, quota.V1Equals(expected, actual), "%q unequal:\nExpected:%v\nActual:%v", name, expected, actual)
}

func TestParseCustomMachineType(t *testing.T) {
	cpu, mem, err := parseCustomMachineType("custom-2-2816")
	assert.NoError(t, err)
	assert.Equal(t, int64(2), cpu)
	assert.Equal(t, int64(2816*1024*1024), mem)
	cpu, mem, err = parseCustomMachineType("other-a2-2816")
	assert.Error(t, err)
	cpu, mem, err = parseCustomMachineType("other-2-2816")
	assert.Error(t, err)
}
