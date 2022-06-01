/*
Copyright 2017 The Kubernetes Authors.

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

package labels

import (
	"reflect"
	"sort"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	kubeletapis "k8s.io/kubelet/pkg/apis"
)

var (
	// cpu amount used for account pods that don't specify cpu requests
	defaultMinCPU        = *resource.NewMilliQuantity(50, resource.DecimalSI)
	infrastructureLabels = []string{"kubernetes.io", "cloud.google.com"}
)

type nodeSelectorStats struct {
	nodeSelector map[string]string
	totalCpu     resource.Quantity
}

// BestLabelSet returns a set of labels for nodes that will allow to schedule the pods that
// requested the most cpu.
func BestLabelSet(pods []*apiv1.Pod) map[string]string {
	nodeSelectors := calculateNodeSelectorStats(pods)
	sortNodeSelectorStats(nodeSelectors)
	// Take labels from the selector that covers most of the pods (in terms of requested cpu).
	selector := nodeSelectors[0].nodeSelector

	// Expand the list of labels so that the other pods can fit as well. However as infrastructure
	// related labels might not be compatible with each other let's skip these selectors that
	// require new infrastructure labels (like kubernetes.io/preemptive=true). New generic
	// labels that are unlikely to cause problems when mixed are ok. And obviously skip pods that
	// require conflicting labels.
statloop:
	for _, nss := range nodeSelectors[1:] {
		for k, v := range nss.nodeSelector {
			currentValue, found := selector[k]

			if found && currentValue != v {
				continue statloop
			}
			if !found {
				for _, infraLabel := range infrastructureLabels {
					if strings.Contains(k, infraLabel) {
						continue statloop
					}
				}
			}
		}
		// All labels are non-infra and/or, can be added.
		for k, v := range nss.nodeSelector {
			selector[k] = v
		}
	}
	return selector
}

func sortNodeSelectorStats(stats []nodeSelectorStats) {
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].totalCpu.MilliValue() > stats[j].totalCpu.MilliValue()
	})
}

func calculateNodeSelectorStats(pods []*apiv1.Pod) []nodeSelectorStats {
	stats := make([]nodeSelectorStats, 0)
	for _, pod := range pods {
		var podCpu resource.Quantity
		for _, container := range pod.Spec.Containers {
			if container.Resources.Requests != nil {
				containerCpu := container.Resources.Requests[apiv1.ResourceCPU]
				podCpu.Add(containerCpu)
			}
		}
		if podCpu.MilliValue() == 0 {
			podCpu = defaultMinCPU
		}

		found := false
		nodeSelector := pod.Spec.NodeSelector
		if nodeSelector == nil {
			nodeSelector = map[string]string{}
		}

		for i := range stats {
			if reflect.DeepEqual(stats[i].nodeSelector, nodeSelector) {
				found = true
				stats[i].totalCpu.Add(podCpu)
				break
			}
		}
		if !found {
			stats = append(stats, nodeSelectorStats{
				nodeSelector: nodeSelector,
				totalCpu:     podCpu,
			})
		}
	}
	return stats
}

// UpdateDeprecatedLabels updates beta and deprecated labels from stable labels
func UpdateDeprecatedLabels(labels map[string]string) {
	if v, ok := labels[apiv1.LabelArchStable]; ok {
		labels[kubeletapis.LabelArch] = v
	}
	if v, ok := labels[apiv1.LabelOSStable]; ok {
		labels[kubeletapis.LabelOS] = v
	}
	if v, ok := labels[apiv1.LabelInstanceTypeStable]; ok {
		labels[apiv1.LabelInstanceType] = v
	}
	if v, ok := labels[apiv1.LabelTopologyRegion]; ok {
		labels[apiv1.LabelZoneRegion] = v
	}
	if v, ok := labels[apiv1.LabelTopologyZone]; ok {
		labels[apiv1.LabelZoneFailureDomain] = v
	}
}
