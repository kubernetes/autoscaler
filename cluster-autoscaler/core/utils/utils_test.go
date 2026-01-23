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

package utils

import (
	"errors"
	"testing"
	"time"

	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	mock_cloudprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/mocks"
	caerrors "k8s.io/autoscaler/cluster-autoscaler/utils/errors"
)

func TestGetNodeResource(t *testing.T) {
	node := BuildTestNode("n1", 1000, 2*MiB)

	cores := getNodeResource(node, apiv1.ResourceCPU)
	assert.Equal(t, int64(1), cores)

	memory := getNodeResource(node, apiv1.ResourceMemory)
	assert.Equal(t, int64(2*MiB), memory)

	unknownResourceValue := getNodeResource(node, "unknown resource")
	assert.Equal(t, int64(0), unknownResourceValue)

	// if we have no resources in capacity we expect getNodeResource to return 0
	nodeWithMissingCapacity := BuildTestNode("n1", 1000, 2*MiB)
	nodeWithMissingCapacity.Status.Capacity = apiv1.ResourceList{}

	cores = getNodeResource(nodeWithMissingCapacity, apiv1.ResourceCPU)
	assert.Equal(t, int64(0), cores)

	memory = getNodeResource(nodeWithMissingCapacity, apiv1.ResourceMemory)
	assert.Equal(t, int64(0), memory)

	// if we have negative values in resources we expect getNodeResource to return 0
	nodeWithNegativeCapacity := BuildTestNode("n1", -1000, -2*MiB)
	nodeWithNegativeCapacity.Status.Capacity = apiv1.ResourceList{}

	cores = getNodeResource(nodeWithNegativeCapacity, apiv1.ResourceCPU)
	assert.Equal(t, int64(0), cores)

	memory = getNodeResource(nodeWithNegativeCapacity, apiv1.ResourceMemory)
	assert.Equal(t, int64(0), memory)

}

func TestGetNodeCoresAndMemory(t *testing.T) {
	node := BuildTestNode("n1", 2000, 2048*MiB)

	cores, memory := GetNodeCoresAndMemory(node)
	assert.Equal(t, int64(2), cores)
	assert.Equal(t, int64(2048*MiB), memory)

	// if we have no cpu/memory defined in capacity we expect getNodeCoresAndMemory to return 0s
	nodeWithMissingCapacity := BuildTestNode("n1", 1000, 2*MiB)
	nodeWithMissingCapacity.Status.Capacity = apiv1.ResourceList{}

	cores, memory = GetNodeCoresAndMemory(nodeWithMissingCapacity)
	assert.Equal(t, int64(0), cores)
	assert.Equal(t, int64(0), memory)

	// if we have negative capacity defined in capacity we expect getNodeCoresAndMemory to return 0s
	nodeWithNegativeCapacity := BuildTestNode("n1", -1000, -2*MiB)
	cores, memory = GetNodeCoresAndMemory(nodeWithNegativeCapacity)
	assert.Equal(t, int64(0), cores)
	assert.Equal(t, int64(0), memory)
}

func TestGetOldestPod(t *testing.T) {
	p1 := BuildTestPod("p1", 500, 1000)
	p1.CreationTimestamp = metav1.NewTime(time.Now().Add(-1 * time.Minute))
	p2 := BuildTestPod("p2", 500, 1000)
	p2.CreationTimestamp = metav1.NewTime(time.Now().Add(+1 * time.Minute))
	p3 := BuildTestPod("p3", 500, 1000)
	p3.CreationTimestamp = metav1.NewTime(time.Now())

	assert.Equal(t, p1.CreationTimestamp.Time, GetOldestCreateTime([]*apiv1.Pod{p1, p2, p3}))
	assert.Equal(t, p1.CreationTimestamp.Time, GetOldestCreateTime([]*apiv1.Pod{p3, p2, p1}))
}

func TestVirtualKubeletNodeFilter(t *testing.T) {
	type args struct {
		node *apiv1.Node
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "returns false when node is nil",
			args: args{
				node: nil,
			},
			want: false,
		},
		{
			name: "returns false when node does not have virtual-kubelet label",
			args: args{
				node: &apiv1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"type": "not-virtual-kubelet",
						},
					},
				},
			},
			want: false,
		},
		{
			name: "returns true when node has virtual-kubelet label",
			args: args{
				node: &apiv1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"type": VirtualKubeletNodeLabelValue,
						},
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := VirtualKubeletNodeFilter{}
			assert.Equalf(t, tt.want, f.ExcludeFromTracking(tt.args.node), "VirtualKubeletNodeFilter.ExcludeFromTracking(%v)", tt.args.node)
		})
	}
}

func TestFilterOutNodesFromNotAutoscaledGroups(t *testing.T) {
	nodeInGroup := BuildTestNode("node-in-group", 1000, 2*MiB)
	nodeNotInGroup := BuildTestNode("node-not-in-group", 1000, 2*MiB)
	nodeError := BuildTestNode("node-error", 1000, 2*MiB)
	nodeVirtual := BuildTestNode("node-virtual", 1000, 2*MiB)
	nodeVirtual.Labels["type"] = VirtualKubeletNodeLabelValue

	mockProvider := &mock_cloudprovider.CloudProvider{}
	mockProvider.On("NodeGroupForNode", nodeInGroup).Return(new(mock_cloudprovider.NodeGroup), nil)
	mockProvider.On("NodeGroupForNode", nodeNotInGroup).Return(nil, nil)
	mockProvider.On("NodeGroupForNode", nodeError).Return(nil, errors.New("some error"))

	t.Run("ignores virtual nodes", func(t *testing.T) {
		result, err := FilterOutNodesFromNotAutoscaledGroups([]*apiv1.Node{nodeVirtual}, mockProvider)
		assert.NoError(t, err)
		assert.Empty(t, result)

		mockProvider.AssertNotCalled(t, "NodeGroupForNode", nodeVirtual)
	})

	t.Run("returns error if cloud provider fails", func(t *testing.T) {
		result, err := FilterOutNodesFromNotAutoscaledGroups([]*apiv1.Node{nodeError}, mockProvider)
		assert.Error(t, err)
		assert.Equal(t, caerrors.CloudProviderError, err.Type())
		assert.Empty(t, result)
	})

	t.Run("filters out nodes in autoscaled groups", func(t *testing.T) {
		result, err := FilterOutNodesFromNotAutoscaledGroups([]*apiv1.Node{nodeInGroup, nodeNotInGroup}, mockProvider)
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "node-not-in-group", result[0].Name)
	})

	mockProvider.AssertExpectations(t)
}

func TestHasHardInterPodAffinity(t *testing.T) {
	type args struct {
		affinity *apiv1.Affinity
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "returns false when affinity is nil",
			args: args{
				affinity: nil,
			},
			want: false,
		},
		{
			name: "returns false when affinity is empty",
			args: args{
				affinity: &apiv1.Affinity{},
			},
			want: false,
		},
		{
			name: "returns true when PodAffinity has RequiredDuringScheduling terms",
			args: args{
				affinity: &apiv1.Affinity{
					PodAffinity: &apiv1.PodAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: []apiv1.PodAffinityTerm{
							{LabelSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "test"}}},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "returns false when PodAffinity has empty RequiredDuringScheduling",
			args: args{
				affinity: &apiv1.Affinity{
					PodAffinity: &apiv1.PodAffinity{
						PreferredDuringSchedulingIgnoredDuringExecution: []apiv1.WeightedPodAffinityTerm{
							{Weight: 10, PodAffinityTerm: apiv1.PodAffinityTerm{}},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "returns true when PodAntiAffinity has RequiredDuringScheduling terms",
			args: args{
				affinity: &apiv1.Affinity{
					PodAntiAffinity: &apiv1.PodAntiAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: []apiv1.PodAffinityTerm{
							{LabelSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "test"}}},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "returns false when PodAntiAffinity has empty RequiredDuringScheduling",
			args: args{
				affinity: &apiv1.Affinity{
					PodAntiAffinity: &apiv1.PodAntiAffinity{
						PreferredDuringSchedulingIgnoredDuringExecution: []apiv1.WeightedPodAffinityTerm{
							{Weight: 10, PodAffinityTerm: apiv1.PodAffinityTerm{}},
						},
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, hasHardInterPodAffinity(tt.args.affinity), "hasHardInterPodAffinity(%v)", tt.args.affinity)
		})
	}
}

func TestGetOldestCreateTimeWithGpu(t *testing.T) {
	fixedTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	p1 := BuildTestPod("p1", 500, 1000)
	p1.CreationTimestamp = metav1.NewTime(fixedTime.Add(-1 * time.Minute))
	p2 := BuildTestPod("p2", 500, 1000)
	p2.CreationTimestamp = metav1.NewTime(fixedTime.Add(-2 * time.Minute))
	RequestGpuForPod(p2, 1)
	p3 := BuildTestPod("p3", 500, 1000)
	p3.CreationTimestamp = metav1.NewTime(fixedTime.Add(-3 * time.Minute))
	RequestGpuForPod(p3, 1)

	t.Run("returns false when no pods", func(t *testing.T) {
		found, _ := GetOldestCreateTimeWithGpu([]*apiv1.Pod{})
		assert.False(t, found)
	})

	t.Run("returns false when no GPU pods", func(t *testing.T) {
		found, _ := GetOldestCreateTimeWithGpu([]*apiv1.Pod{p1})
		assert.False(t, found)
	})

	t.Run("returns creation time when single GPU pod", func(t *testing.T) {
		found, ts := GetOldestCreateTimeWithGpu([]*apiv1.Pod{p2})
		assert.True(t, found)
		assert.Equal(t, p2.CreationTimestamp.Time, ts)
	})

	t.Run("returns oldest time among multiple GPU pods", func(t *testing.T) {
		found, ts := GetOldestCreateTimeWithGpu([]*apiv1.Pod{p1, p2, p3})
		assert.True(t, found)
		assert.Equal(t, p3.CreationTimestamp.Time, ts)
	})
}
