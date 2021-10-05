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

package waste

import (
	"testing"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/expander"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

type FakeNodeGroup struct {
	id string
}

func (f *FakeNodeGroup) MaxSize() int                       { return 2 }
func (f *FakeNodeGroup) MinSize() int                       { return 1 }
func (f *FakeNodeGroup) TargetSize() (int, error)           { return 2, nil }
func (f *FakeNodeGroup) IncreaseSize(delta int) error       { return nil }
func (f *FakeNodeGroup) DecreaseTargetSize(delta int) error { return nil }
func (f *FakeNodeGroup) DeleteNodes([]*apiv1.Node) error    { return nil }
func (f *FakeNodeGroup) Id() string                         { return f.id }
func (f *FakeNodeGroup) Debug() string                      { return f.id }
func (f *FakeNodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	return []cloudprovider.Instance{}, nil
}
func (f *FakeNodeGroup) TemplateNodeInfo() (*schedulerframework.NodeInfo, error) {
	return nil, cloudprovider.ErrNotImplemented
}
func (f *FakeNodeGroup) Exist() bool { return true }
func (f *FakeNodeGroup) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrAlreadyExist
}
func (f *FakeNodeGroup) Delete() error         { return cloudprovider.ErrNotImplemented }
func (f *FakeNodeGroup) Autoprovisioned() bool { return false }

func makeNodeInfo(cpu int64, memory int64, pods int64) *schedulerframework.NodeInfo {
	node := &apiv1.Node{
		Status: apiv1.NodeStatus{
			Capacity: apiv1.ResourceList{
				apiv1.ResourceCPU:    *resource.NewMilliQuantity(cpu, resource.DecimalSI),
				apiv1.ResourceMemory: *resource.NewQuantity(memory, resource.DecimalSI),
				apiv1.ResourcePods:   *resource.NewQuantity(pods, resource.DecimalSI),
			},
		},
	}
	node.Status.Allocatable = node.Status.Capacity
	SetNodeReadyState(node, true, time.Time{})

	nodeInfo := schedulerframework.NewNodeInfo()
	nodeInfo.SetNode(node)

	return nodeInfo
}

func TestLeastWaste(t *testing.T) {
	cpuPerPod := int64(500)
	memoryPerPod := int64(1000 * 1024 * 1024)
	e := NewStrategy()
	balancedNodeInfo := makeNodeInfo(16*cpuPerPod, 16*memoryPerPod, 100)
	nodeMap := map[string]*schedulerframework.NodeInfo{"balanced": balancedNodeInfo}
	balancedOption := expander.Option{NodeGroup: &FakeNodeGroup{"balanced"}, NodeCount: 1}

	// Test without any pods, one node info
	ret := e.BestOption([]expander.Option{balancedOption}, nodeMap)
	assert.Equal(t, *ret, balancedOption)

	pod := &apiv1.Pod{
		Spec: apiv1.PodSpec{
			Containers: []apiv1.Container{
				{
					Resources: apiv1.ResourceRequirements{
						Requests: apiv1.ResourceList{
							apiv1.ResourceCPU:    *resource.NewMilliQuantity(cpuPerPod, resource.DecimalSI),
							apiv1.ResourceMemory: *resource.NewQuantity(memoryPerPod, resource.DecimalSI),
						},
					},
				},
			},
		},
	}

	// Test with one pod, one node info
	balancedOption.Pods = []*apiv1.Pod{pod}
	ret = e.BestOption([]expander.Option{balancedOption}, nodeMap)
	assert.Equal(t, *ret, balancedOption)

	// Test with one pod, two node infos, one that has lots of RAM one that has less
	highmemNodeInfo := makeNodeInfo(16*cpuPerPod, 32*memoryPerPod, 100)
	nodeMap["highmem"] = highmemNodeInfo
	highmemOption := expander.Option{NodeGroup: &FakeNodeGroup{"highmem"}, NodeCount: 1, Pods: []*apiv1.Pod{pod}}
	ret = e.BestOption([]expander.Option{balancedOption, highmemOption}, nodeMap)
	assert.Equal(t, *ret, balancedOption)

	// Test with one pod, three node infos, one that has lots of RAM one that has less, and one that has less CPU
	lowcpuNodeInfo := makeNodeInfo(8*cpuPerPod, 16*memoryPerPod, 100)
	nodeMap["lowcpu"] = lowcpuNodeInfo
	lowcpuOption := expander.Option{NodeGroup: &FakeNodeGroup{"lowcpu"}, NodeCount: 1, Pods: []*apiv1.Pod{pod}}
	ret = e.BestOption([]expander.Option{balancedOption, highmemOption, lowcpuOption}, nodeMap)
	assert.Equal(t, *ret, lowcpuOption)
}
