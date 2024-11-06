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
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/expander"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	klog "k8s.io/klog/v2"
)

type leastwaste struct {
}

// NewFilter returns a filter that selects the best scale up option based on which node group returns the least waste
func NewFilter() expander.Filter {
	return &leastwaste{}
}

// BestOption Finds the option that wastes the least fraction of CPU and Memory
func (l *leastwaste) BestOptions(expansionOptions []expander.Option, nodeInfo map[string]*framework.NodeInfo) []expander.Option {
	var leastWastedScore float64
	var leastWastedOptions []expander.Option

	for _, option := range expansionOptions {
		requestedCPU, requestedMemory := resourcesForPods(option.Pods)
		node, found := nodeInfo[option.NodeGroup.Id()]
		if !found {
			klog.Errorf("No node info for: %s", option.NodeGroup.Id())
			continue
		}

		nodeCPU, nodeMemory := resourcesForNode(node.Node())
		availCPU := nodeCPU.MilliValue() * int64(option.NodeCount)
		availMemory := nodeMemory.Value() * int64(option.NodeCount)
		wastedCPU := float64(availCPU-requestedCPU.MilliValue()) / float64(availCPU)
		wastedMemory := float64(availMemory-requestedMemory.Value()) / float64(availMemory)
		wastedScore := wastedCPU + wastedMemory

		klog.V(1).Infof("Expanding Node Group %s would waste %0.2f%% CPU, %0.2f%% Memory, %0.2f%% Blended\n", option.NodeGroup.Id(), wastedCPU*100.0, wastedMemory*100.0, wastedScore*50.0)

		if wastedScore == leastWastedScore {
			leastWastedOptions = append(leastWastedOptions, option)
		}

		if leastWastedOptions == nil || wastedScore < leastWastedScore {
			leastWastedScore = wastedScore
			leastWastedOptions = []expander.Option{option}
		}
	}

	if len(leastWastedOptions) == 0 {
		return nil
	}

	return leastWastedOptions
}

func resourcesForPods(pods []*apiv1.Pod) (cpu resource.Quantity, memory resource.Quantity) {
	for _, pod := range pods {
		for _, container := range pod.Spec.Containers {
			if request, ok := container.Resources.Requests[apiv1.ResourceCPU]; ok {
				cpu.Add(request)
			}
			if request, ok := container.Resources.Requests[apiv1.ResourceMemory]; ok {
				memory.Add(request)
			}
		}
	}

	return cpu, memory
}

func resourcesForNode(node *apiv1.Node) (cpu resource.Quantity, memory resource.Quantity) {
	cpu = node.Status.Capacity[apiv1.ResourceCPU]
	memory = node.Status.Capacity[apiv1.ResourceMemory]

	return cpu, memory
}
