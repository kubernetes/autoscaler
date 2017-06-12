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
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/expander"
	"k8s.io/autoscaler/cluster-autoscaler/expander/mostpods"
	"k8s.io/autoscaler/cluster-autoscaler/expander/price"
	"k8s.io/autoscaler/cluster-autoscaler/expander/random"
	"k8s.io/autoscaler/cluster-autoscaler/expander/waste"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"

	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
)

// ExpanderStrategyFromString creates an expander.Strategy according to its name
func ExpanderStrategyFromString(expanderFlag string, cloudProvider cloudprovider.CloudProvider,
	nodeLister kube_util.NodeLister) (expander.Strategy, errors.AutoscalerError) {
	switch expanderFlag {
	case expander.RandomExpanderName:
		return random.NewStrategy(), nil
	case expander.MostPodsExpanderName:
		return mostpods.NewStrategy(), nil
	case expander.LeastWasteExpanderName:
		return waste.NewStrategy(), nil
	case expander.PriceBasedExpanderName:
		pricing, err := cloudProvider.Pricing()
		if err != nil {
			return nil, err
		}
		return price.NewStrategy(pricing,
			price.NewSimplePreferredNodeProvider(nodeLister),
			price.SimpleNodeUnfitness), nil
	}
	return nil, errors.NewAutoscalerError(errors.InternalError, "Expander %s not supported", expanderFlag)
}
