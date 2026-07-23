/*
Copyright The Kubernetes Authors.

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

package efficiency

import (
	"fmt"
	"strings"

	apiv1 "k8s.io/api/core/v1"
)

type pricingInfo struct {
	CPUPerHour         float64
	MemoryPerHourPerGB float64
}

// getCPUPrice returns the CPU core hourly price.
func (p pricingInfo) getCPUPrice() float64 {
	return p.CPUPerHour
}

// getMemoryPrice returns the memory GiB hourly price.
func (p pricingInfo) getMemoryPrice() float64 {
	return p.MemoryPerHourPerGB
}

// machineFamily represents a cloud instances family and its pricing details.
type machineFamily struct {
	FamilyName  string
	PricingInfo pricingInfo
}

// machineFamilies maps the machine family prefix to pricing information.
// Fixed ratio of 1 CPU/h = 7.5*mem/h/GB with imaginary prices.
var machineFamilies = map[string]*machineFamily{
	"s": {
		FamilyName: "s",
		PricingInfo: pricingInfo{
			CPUPerHour:         1,
			MemoryPerHourPerGB: 0.13,
		},
	},
	"m": {
		FamilyName: "m",
		PricingInfo: pricingInfo{
			CPUPerHour:         2.5,
			MemoryPerHourPerGB: 0.33,
		},
	},
	"l": {
		FamilyName: "l",
		PricingInfo: pricingInfo{
			CPUPerHour:         4.2,
			MemoryPerHourPerGB: 0.56,
		},
	},
	"xl": {
		FamilyName: "xl",
		PricingInfo: pricingInfo{
			CPUPerHour:         8,
			MemoryPerHourPerGB: 1.06,
		},
	},
}

// getMachineFamily retrieves the machine family of a node based on the labels.
// if "node.kubernetes.io/instance-type" not in labels => try "beta.kubernetes.io/instance-type"
func getMachineFamily(node *apiv1.Node) (*machineFamily, error) {
	instanceType, ok := node.Labels[apiv1.LabelInstanceTypeStable]
	if !ok {
		instanceType, ok = node.Labels[apiv1.LabelInstanceType]
		if !ok {
			return nil, fmt.Errorf("node %s has no instance type label", node.Name)
		}
	}
	chunks := strings.Split(instanceType, "-")
	if len(chunks) == 0 {
		return nil, fmt.Errorf("invalid instance type label value: %s for node %s", instanceType, node.Name)
	}
	familyName := chunks[0]
	family, ok := machineFamilies[familyName]
	if !ok {
		return nil, fmt.Errorf("unknown pricing for machine family %s parsed from instance type %s, node %s", familyName, instanceType, node.Name)
	}
	return family, nil
}

// calculateNodeCost returns the hourly cost of the node as node_CPU_allocatable*CPU_cost + node_mem_allocatable*mem_cost.
func calculateNodeCost(node *apiv1.Node) (float64, error) {
	family, err := getMachineFamily(node)
	if err != nil {
		return 0, err
	}
	nodeAllocatableCPU := node.Status.Allocatable[apiv1.ResourceCPU]
	nodeAllocatableMem := node.Status.Allocatable[apiv1.ResourceMemory]
	cpuCores := float64(nodeAllocatableCPU.MilliValue()) / 1000.0
	memGiB := float64(nodeAllocatableMem.MilliValue()) / (1024.0 * 1024.0 * 1024.0)
	cpuCost := cpuCores * family.PricingInfo.getCPUPrice()
	memCost := memGiB * family.PricingInfo.getMemoryPrice()
	return cpuCost + memCost, nil
}
