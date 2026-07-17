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

package leastnodes

import (
	"math"

	"k8s.io/autoscaler/cluster-autoscaler/expander"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
)

type leastnodes struct {
}

// NewFilter returns a scale up filter that picks the node group that uses the least number of nodes
func NewFilter() expander.Filter {
	return &leastnodes{}
}

// BestOptions selects the expansion option that uses the least number of nodes
func (m *leastnodes) BestOptions(expansionOptions []expander.Option, nodeInfo map[string]*framework.NodeInfo) []expander.Option {
	leastNodes := math.MaxInt
	var leastOptions []expander.Option

	for _, option := range expansionOptions {
		// Don't think this is possible, but just in case
		if option.NodeCount == 0 {
			continue
		}

		if option.NodeCount == leastNodes {
			leastOptions = append(leastOptions, option)
			continue
		}

		if option.NodeCount < leastNodes {
			leastNodes = option.NodeCount
			leastOptions = []expander.Option{option}
		}
	}

	if len(leastOptions) == 0 {
		return nil
	}

	return leastOptions
}
