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
	"testing"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/gce"
	gpuUtils "k8s.io/autoscaler/cluster-autoscaler/utils/gpu"

	gce_api "google.golang.org/api/compute/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	kubeletapis "k8s.io/kubernetes/pkg/kubelet/apis"
	quota "k8s.io/kubernetes/pkg/quota/v1"

	"github.com/stretchr/testify/assert"
)

func TestBuildNodeFromTemplateSetsResources(t *testing.T) {
	type testCase struct {
		kubeEnv           string
		name              string
		machineType       string
		accelerators      []*gce_api.AcceleratorConfig
		mig               gce.Mig
		capacityCpu       int64
		capacityMemory    int64
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
		accelerators: []*gce_api.AcceleratorConfig{
			{AcceleratorType: "nvidia-tesla-k80", AcceleratorCount: 3},
			{AcceleratorType: "nvidia-tesla-p100", AcceleratorCount: 8},
		},
		mig: &GkeMig{
			gceRef: gce.GceRef{
				Name:    "some-name",
				Project: "some-proj",
				Zone:    "us-central1-b",
			},
		},
		capacityCpu:       8,
		capacityMemory:    200 * 1024 * 1024,
		allocatableCpu:    "7000m",
		allocatableMemory: fmt.Sprintf("%v", 99*1024*1024),
		gpuCount:          11,
		expectedErr:       false,
	},
		{
			kubeEnv: "ENABLE_NODE_PROBLEM_DETECTOR: 'daemonset'\n" +
				"NODE_LABELS: a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true\n" +
				"DNS_SERVER_IP: '10.0.0.10'\n" +
				"NODE_TAINTS: 'dedicated=ml:NoSchedule,test=dev:PreferNoSchedule,a=b:c'\n",
			name:        "nodeName",
			machineType: "custom-8-2",
			mig: &GkeMig{
				gceRef: gce.GceRef{
					Name:    "some-name",
					Project: "some-proj",
					Zone:    "us-central1-b",
				},
			},
			capacityCpu:       8,
			capacityMemory:    2 * 1024 * 1024,
			allocatableCpu:    "8000m",
			allocatableMemory: fmt.Sprintf("%v", 2*1024*1024),
			expectedErr:       false,
		}, {
			kubeEnv:     "This kube-env is totally messed up",
			name:        "nodeName",
			machineType: "custom-8-2",
			mig: &GkeMig{
				gceRef: gce.GceRef{
					Name:    "some-name",
					Project: "some-proj",
					Zone:    "us-central1-b",
				},
			},
			expectedErr: true,
		},
	}
	for _, tc := range testCases {
		tb := &GkeTemplateBuilder{}
		template := &gce_api.InstanceTemplate{
			Name: tc.name,
			Properties: &gce_api.InstanceProperties{
				GuestAccelerators: tc.accelerators,
				Metadata: &gce_api.Metadata{
					Items: []*gce_api.MetadataItems{{Key: "kube-env", Value: &tc.kubeEnv}},
				},
				MachineType: tc.machineType,
			},
		}
		node, err := tb.BuildNodeFromTemplate(tc.mig, template, tc.capacityCpu, tc.capacityMemory)
		if tc.expectedErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			podsQuantity, _ := resource.ParseQuantity("110")
			capacity, err := makeResourceList(fmt.Sprintf("%dm", tc.capacityCpu*1000), fmt.Sprintf("%v", tc.capacityMemory), tc.gpuCount)
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

func TestBuildAllocatableFromCapacity(t *testing.T) {
	type testCase struct {
		capacityCpu       string
		capacityMemory    string
		allocatableCpu    string
		allocatableMemory string
		gpuCount          int64
	}
	testCases := []testCase{{
		capacityCpu:    "16000m",
		capacityMemory: fmt.Sprintf("%v", 1*mbPerGB*bytesPerMB),
		allocatableCpu: "15890m",
		// Below threshold for reserving memory
		allocatableMemory: fmt.Sprintf("%v", 1*mbPerGB*bytesPerMB-kubeletEvictionHardMemory),
		gpuCount:          1,
	}, {
		capacityCpu:    "500m",
		capacityMemory: fmt.Sprintf("%v", 1.1*mbPerGB*bytesPerMB),
		allocatableCpu: "470m",
		// Final 1024*1024 because we're duplicating upstream bug using MB as MiB
		allocatableMemory: fmt.Sprintf("%v", 1.1*mbPerGB*bytesPerMB-0.25*1.1*mbPerGB*1024*1024-kubeletEvictionHardMemory),
	}}
	for _, tc := range testCases {
		tb := GkeTemplateBuilder{}
		capacity, err := makeResourceList(tc.capacityCpu, tc.capacityMemory, tc.gpuCount)
		assert.NoError(t, err)
		expectedAllocatable, err := makeResourceList(tc.allocatableCpu, tc.allocatableMemory, tc.gpuCount)
		assert.NoError(t, err)
		allocatable := tb.BuildAllocatableFromCapacity(capacity)
		assertEqualResourceLists(t, "Allocatable", expectedAllocatable, allocatable)
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
	assert.True(t, quota.V1Equals(expected, actual), "%q unequal:\nExpected:%v\nActual:%v", name, expected, actual)
}
