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
	"context"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	klog "k8s.io/klog/v2"
	"sigs.k8s.io/cluster-autoscaler/pkg/expander"
	"sigs.k8s.io/cluster-autoscaler/pkg/simulator/framework"
	podutils "sigs.k8s.io/cluster-autoscaler/pkg/utils/pod"
)

type leastwaste struct {
}

// NewFilter returns a filter that selects the best scale up option based on which node group returns the least waste
func NewFilter() expander.Filter {
	return &leastwaste{}
}

// BestOption Finds the option that wastes the least fraction of CPU and Memory
func (l *leastwaste) BestOptions(ctx context.Context, expansionOptions []expander.Option, nodeInfo map[string]*framework.NodeInfo) []expander.Option {
	logger := klog.FromContext(ctx)
	var leastWastedScore float64
	var leastWastedOptions []expander.Option

	for _, option := range expansionOptions {
		requestedCPU, requestedMemory := resourcesForPods(option.Pods)
		node, found := nodeInfo[option.NodeGroup.Id()]
		if !found {
			logger.Error(nil, "No node info for node group", "nodeGroupId", option.NodeGroup.Id())
			continue
		}

		nodeCPU, nodeMemory := resourcesForNode(node.Node())
		availCPU := nodeCPU.MilliValue() * int64(option.NodeCount)
		availMemory := nodeMemory.Value() * int64(option.NodeCount)
		wastedCPU := float64(availCPU-requestedCPU.MilliValue()) / float64(availCPU)
		wastedMemory := float64(availMemory-requestedMemory.Value()) / float64(availMemory)
		wastedScore := wastedCPU + wastedMemory
		logger.V(1).Info("Expanding Node Group would waste resources", "nodeGroupId", option.NodeGroup.Id(), "wastedCpuPercent", wastedCPU*100.0, "wastedMemoryPercent", wastedMemory*100.0, "wastedResourcesMeanPercent", wastedScore*50.0)

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
		podRequests := podutils.PodRequests(pod)
		cpu.Add(podRequests[apiv1.ResourceCPU])
		memory.Add(podRequests[apiv1.ResourceMemory])
	}

	return cpu, memory
}

func resourcesForNode(node *apiv1.Node) (cpu resource.Quantity, memory resource.Quantity) {
	cpu = node.Status.Capacity[apiv1.ResourceCPU]
	memory = node.Status.Capacity[apiv1.ResourceMemory]

	return cpu, memory
}
