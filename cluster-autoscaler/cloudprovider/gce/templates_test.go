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
	"strings"
	"testing"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	gpuUtils "k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/autoscaler/cluster-autoscaler/utils/units"

	gce "google.golang.org/api/compute/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	kubeletapis "k8s.io/kubernetes/pkg/kubelet/apis"
	quota "k8s.io/kubernetes/pkg/quota/v1"

	"github.com/stretchr/testify/assert"
)

func TestBuildNodeFromTemplateSetsResources(t *testing.T) {
	type testCase struct {
		scenario         string
		kubeEnv          string
		accelerators     []*gce.AcceleratorConfig
		mig              Mig
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
			accelerators: []*gce.AcceleratorConfig{
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
					Metadata: &gce.Metadata{
						Items: []*gce.MetadataItems{{Key: "kube-env", Value: &tc.kubeEnv}},
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

func TestBuildGenericLabels(t *testing.T) {
	labels, err := BuildGenericLabels(GceRef{
		Name:    "kubernetes-minion-group",
		Project: "mwielgus-proj",
		Zone:    "us-central1-b"},
		"n1-standard-8", "sillyname")
	assert.Nil(t, err)
	assert.Equal(t, "us-central1", labels[apiv1.LabelZoneRegion])
	assert.Equal(t, "us-central1-b", labels[apiv1.LabelZoneFailureDomain])
	assert.Equal(t, "sillyname", labels[apiv1.LabelHostname])
	assert.Equal(t, "n1-standard-8", labels[apiv1.LabelInstanceType])
	assert.Equal(t, cloudprovider.DefaultArch, labels[kubeletapis.LabelArch])
	assert.Equal(t, cloudprovider.DefaultOS, labels[kubeletapis.LabelOS])
}

func TestCalculateAllocatable(t *testing.T) {
	type testCase struct {
		scenario          string
		capacityCpu       string
		reservedCpu       string
		allocatableCpu    string
		capacityMemory    string
		reservedMemory    string
		allocatableMemory string
	}
	testCases := []testCase{
		{
			scenario:          "no reservations",
			capacityCpu:       "8",
			reservedCpu:       "0",
			allocatableCpu:    "8",
			capacityMemory:    fmt.Sprintf("%v", 200*units.MiB),
			reservedMemory:    "0",
			allocatableMemory: fmt.Sprintf("%v", 200*units.MiB-KubeletEvictionHardMemory),
		},
		{
			scenario:          "reserved cpu and memory",
			capacityCpu:       "8",
			reservedCpu:       "1000m",
			allocatableCpu:    "7000m",
			capacityMemory:    fmt.Sprintf("%v", 200*units.MiB),
			reservedMemory:    fmt.Sprintf("%v", 50*units.MiB),
			allocatableMemory: fmt.Sprintf("%v", 150*units.MiB-KubeletEvictionHardMemory),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.scenario, func(t *testing.T) {
			tb := GceTemplateBuilder{}
			capacity, err := makeResourceList(tc.capacityCpu, tc.capacityMemory, 0)
			assert.NoError(t, err)
			reserved, err := makeResourceList(tc.reservedCpu, tc.reservedMemory, 0)
			assert.NoError(t, err)
			expectedAllocatable, err := makeResourceList(tc.allocatableCpu, tc.allocatableMemory, 0)
			assert.NoError(t, err)
			allocatable := tb.CalculateAllocatable(capacity, reserved)
			assertEqualResourceLists(t, "Allocatable", expectedAllocatable, allocatable)
		})
	}
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
		tb := GceTemplateBuilder{}
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
		physicalMemory int64
		capacityMemory int64
		physicalCpu    int64
	}
	testCases := []testCase{
		{
			physicalMemory: 2 * units.GiB,
			capacityMemory: 2*units.GiB - 32*units.MiB - kernelReservedMemory,
			physicalCpu:    1,
		},
		{
			physicalMemory: 4 * units.GiB,
			capacityMemory: 4*units.GiB - 64*units.MiB - kernelReservedMemory - swiotlbReservedMemory,
			physicalCpu:    2,
		},
		{
			physicalMemory: 128 * units.GiB,
			capacityMemory: 128*units.GiB - 2*units.GiB - kernelReservedMemory - swiotlbReservedMemory,
			physicalCpu:    32,
		},
	}
	for idx, tc := range testCases {
		t.Run(fmt.Sprintf("%v", idx), func(t *testing.T) {
			tb := GceTemplateBuilder{}
			capacity, err := tb.BuildCapacity(tc.physicalCpu, tc.physicalMemory, make([]*gce.AcceleratorConfig, 0))
			assert.NoError(t, err)
			expected, err := makeResourceList2(tc.physicalCpu, tc.capacityMemory, 0, 110)
			assert.NoError(t, err)
			assertEqualResourceLists(t, "Capacity", capacity, expected)
		})
	}
}

func TestExtractAutoscalerVarFromKubeEnv(t *testing.T) {
	cases := []struct {
		desc   string
		name   string
		env    string
		expect string
		err    error
	}{
		{
			desc:   "node_labels",
			name:   "node_labels",
			env:    "AUTOSCALER_ENV_VARS: node_labels=a=b,c=d;node_taints=a=b:c,d=e:f\n",
			expect: "a=b,c=d",
		},
		{
			desc:   "node_taints",
			name:   "node_taints",
			env:    "AUTOSCALER_ENV_VARS: node_labels=a=b,c=d;node_taints=a=b:c,d=e:f\n",
			expect: "a=b:c,d=e:f",
		},
		{
			desc: "malformed node_labels",
			name: "node_labels",
			env:  "AUTOSCALER_ENV_VARS: node_labels;node_taints=a=b:c,d=e:f\n",
			err:  fmt.Errorf("malformed autoscaler var: node_labels"),
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			vals, err := extractAutoscalerVarFromKubeEnv(c.env, c.name)
			assert.Equal(t, c.err, err)
			if err != nil {
				return
			}
			assert.Equal(t, c.expect, vals)
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
				"kube_reserved=cpu=1000m,memory=300000Mi\n" +
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
				"kube_reserved=cpu=1000m,memory=300000Mi\n",
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
				"kube_reserved=cpu=1000m,memory=300000Mi\n" +
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
		result[gpuUtils.ResourceNvidiaGPU] = resultGpu
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
