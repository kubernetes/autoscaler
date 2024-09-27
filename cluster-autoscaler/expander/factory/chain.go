/*
Copyright 2021 The Kubernetes Authors.

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

package factory

import (
	"k8s.io/autoscaler/cluster-autoscaler/expander"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
)

type chainStrategy struct {
	filters  []expander.Filter
	fallback expander.Strategy
}

func newChainStrategy(filters []expander.Filter, fallback expander.Strategy) expander.Strategy {
	return &chainStrategy{
		filters:  filters,
		fallback: fallback,
	}
}

func (c *chainStrategy) BestOption(options []expander.Option, nodeInfo map[string]*framework.NodeInfo) *expander.Option {
	filteredOptions := options
	for _, filter := range c.filters {
		filteredOptions = filter.BestOptions(filteredOptions, nodeInfo)
		if len(filteredOptions) == 1 {
			return &filteredOptions[0]
		}
	}
	return c.fallback.BestOption(filteredOptions, nodeInfo)
}
