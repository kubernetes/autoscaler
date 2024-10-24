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

package random

import (
	"math/rand"

	"k8s.io/autoscaler/cluster-autoscaler/expander"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
)

type random struct {
}

// NewFilter returns an expansion filter that randomly picks between node groups
func NewFilter() expander.Filter {
	return &random{}
}

// NewStrategy returns an expansion strategy that randomly picks between node groups
func NewStrategy() expander.Strategy {
	return &random{}
}

// BestOptions selects from the expansion options at random
func (r *random) BestOptions(expansionOptions []expander.Option, nodeInfo map[string]*framework.NodeInfo) []expander.Option {
	best := r.BestOption(expansionOptions, nodeInfo)
	if best == nil {
		return nil
	}
	return []expander.Option{*best}
}

// BestOption selects from the expansion options at random
func (r *random) BestOption(expansionOptions []expander.Option, nodeInfo map[string]*framework.NodeInfo) *expander.Option {
	if len(expansionOptions) <= 0 {
		return nil
	}

	pos := rand.Int31n(int32(len(expansionOptions)))
	return &expansionOptions[pos]
}
