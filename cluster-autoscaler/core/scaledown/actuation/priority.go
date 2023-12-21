/*
Copyright 2022 The Kubernetes Authors.

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

package actuation

import (
	"math"
	"sort"
	"strconv"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	kubelet_config "k8s.io/kubernetes/pkg/kubelet/apis/config"
)

func groupByPriority(shutdownGracePeriodByPodPriority []kubelet_config.ShutdownGracePeriodByPodPriority, fullEvictionPods, bestEffortEvictionPods []*apiv1.Pod) []podEvictionGroup {
	groups := make([]podEvictionGroup, 0, len(shutdownGracePeriodByPodPriority))
	for _, period := range shutdownGracePeriodByPodPriority {
		groups = append(groups, podEvictionGroup{
			ShutdownGracePeriodByPodPriority: period,
		})
	}

	for _, pod := range fullEvictionPods {
		index := groupIndex(pod, groups)
		groups[index].FullEvictionPods = append(groups[index].FullEvictionPods, pod)
	}
	for _, pod := range bestEffortEvictionPods {
		index := groupIndex(pod, groups)
		groups[index].BestEffortEvictionPods = append(groups[index].BestEffortEvictionPods, pod)
	}
	return groups
}

func groupIndex(pod *apiv1.Pod, groups []podEvictionGroup) int {
	var priority int32
	if pod.Spec.Priority != nil {
		priority = *pod.Spec.Priority
	}

	// Find the group index according to the priority.
	index := sort.Search(len(groups), func(i int) bool {
		return (groups)[i].Priority >= priority
	})

	// 1. Those higher than the highest priority default to the highest priority
	// 2. Those lower than the lowest priority default to the lowest priority
	// 3. Those boundary priority default to the lower priority
	// if priority of pod is:
	//   groups[index-1].Priority <= pod priority < groups[index].Priority
	// in which case we want to pick lower one (i.e. index-1)
	if index == len(groups) {
		index = len(groups) - 1
	} else if index < 0 {
		index = 0
	} else if index > 0 && (groups)[index].Priority > priority {
		index--
	}
	return index
}

// ParseShutdownGracePeriodsAndPriorities parse priorityGracePeriodStr and returns an array of ShutdownGracePeriodByPodPriority if succeeded.
// Otherwise, returns an empty list
func ParseShutdownGracePeriodsAndPriorities(priorityGracePeriodStr string) []kubelet_config.ShutdownGracePeriodByPodPriority {
	var priorityGracePeriodMap, emptyMap []kubelet_config.ShutdownGracePeriodByPodPriority

	if priorityGracePeriodStr == "" {
		return emptyMap
	}
	priorityGracePeriodStrArr := strings.Split(priorityGracePeriodStr, ",")
	for _, item := range priorityGracePeriodStrArr {
		priorityAndPeriod := strings.Split(item, ":")
		if len(priorityAndPeriod) != 2 {
			klog.Errorf("Parsing shutdown grace periods failed because '%s' is not a priority and grace period couple separated by ':'", item)
			return emptyMap
		}
		priority, err := strconv.Atoi(priorityAndPeriod[0])
		if err != nil {
			klog.Errorf("Parsing shutdown grace periods and priorities failed: %v", err)
			return emptyMap
		}
		shutDownGracePeriod, err := strconv.Atoi(priorityAndPeriod[1])
		if err != nil {
			klog.Errorf("Parsing shutdown grace periods and priorities failed: %v", err)
			return emptyMap
		}
		priorityGracePeriodMap = append(priorityGracePeriodMap, kubelet_config.ShutdownGracePeriodByPodPriority{
			Priority:                   int32(priority),
			ShutdownGracePeriodSeconds: int64(shutDownGracePeriod),
		})
	}
	return priorityGracePeriodMap
}

// SingleRuleDrainConfig returns an array of ShutdownGracePeriodByPodPriority with a single ShutdownGracePeriodByPodPriority
func SingleRuleDrainConfig(shutdownGracePeriodSeconds int) []kubelet_config.ShutdownGracePeriodByPodPriority {
	return []kubelet_config.ShutdownGracePeriodByPodPriority{
		{
			Priority:                   math.MaxInt32,
			ShutdownGracePeriodSeconds: int64(shutdownGracePeriodSeconds),
		},
	}
}
