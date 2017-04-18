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

package factory

import (
	"k8s.io/autoscaler/cluster-autoscaler/expander"
	"k8s.io/autoscaler/cluster-autoscaler/expander/mostpods"
	"k8s.io/autoscaler/cluster-autoscaler/expander/random"
	"k8s.io/autoscaler/cluster-autoscaler/expander/waste"
)

// ExpanderStrategyFromString creates an expander.Strategy according to its name
func ExpanderStrategyFromString(expanderFlag string) expander.Strategy {
	var expanderStrategy expander.Strategy
	{
		switch expanderFlag {
		case expander.RandomExpanderName:
			expanderStrategy = random.NewStrategy()
		case expander.MostPodsExpanderName:
			expanderStrategy = mostpods.NewStrategy()
		case expander.LeastWasteExpanderName:
			expanderStrategy = waste.NewStrategy()
		}
	}
	return expanderStrategy
}
