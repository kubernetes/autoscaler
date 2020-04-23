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
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"
)

type random struct {
}

// NewStrategy returns an expansion strategy that randomly picks between node groups
func NewStrategy() expander.Strategy {
	return &random{}
}

// RandomExpansion Selects from the expansion options at random
func (r *random) BestOption(expansionOptions []expander.Option, nodeInfo map[string]*schedulerframework.NodeInfo) *expander.Option {
	if len(expansionOptions) <= 0 {
		return nil
	}

	pos := rand.Int31n(int32(len(expansionOptions)))
	return &expansionOptions[pos]
}
