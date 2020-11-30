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

package mostpods

import (
	"k8s.io/autoscaler/cluster-autoscaler/expander"
	"k8s.io/autoscaler/cluster-autoscaler/expander/random"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

type mostpods struct {
	fallbackStrategy expander.Strategy
}

// NewStrategy returns a scale up strategy (expander) that picks the node group that can schedule the most pods
func NewStrategy() expander.Strategy {
	return &mostpods{random.NewStrategy()}
}

// BestOption Selects the expansion option that schedules the most pods
func (m *mostpods) BestOption(expansionOptions []expander.Option, nodeInfo map[string]*schedulerframework.NodeInfo) *expander.Option {
	var maxPods int
	var maxOptions []expander.Option

	for _, option := range expansionOptions {
		if len(option.Pods) == maxPods {
			maxOptions = append(maxOptions, option)
		}

		if len(option.Pods) > maxPods {
			maxPods = len(option.Pods)
			maxOptions = []expander.Option{option}
		}
	}

	if len(maxOptions) == 0 {
		return nil
	}

	return m.fallbackStrategy.BestOption(maxOptions, nodeInfo)
}
