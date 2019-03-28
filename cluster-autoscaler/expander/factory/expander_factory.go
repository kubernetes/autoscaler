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
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/expander"
	"k8s.io/autoscaler/cluster-autoscaler/expander/mostpods"
	"k8s.io/autoscaler/cluster-autoscaler/expander/price"
	"k8s.io/autoscaler/cluster-autoscaler/expander/priority"
	"k8s.io/autoscaler/cluster-autoscaler/expander/random"
	"k8s.io/autoscaler/cluster-autoscaler/expander/waste"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"

	kube_client "k8s.io/client-go/kubernetes"
)

// ExpanderStrategyFromString creates an expander.Strategy according to its name
func ExpanderStrategyFromString(expanderFlag string, cloudProvider cloudprovider.CloudProvider,
	autoscalingKubeClients *context.AutoscalingKubeClients, kubeClient kube_client.Interface,
	configNamespace string) (expander.Strategy, errors.AutoscalerError) {
	switch expanderFlag {
	case expander.RandomExpanderName:
		return random.NewStrategy(), nil
	case expander.MostPodsExpanderName:
		return mostpods.NewStrategy(), nil
	case expander.LeastWasteExpanderName:
		return waste.NewStrategy(), nil
	case expander.PriceBasedExpanderName:
		if _, err := cloudProvider.Pricing(); err != nil {
			return nil, err
		}
		return price.NewStrategy(cloudProvider,
			price.NewSimplePreferredNodeProvider(autoscalingKubeClients.AllNodeLister()),
			price.SimpleNodeUnfitness), nil
	case expander.PriorityBasedExpanderName:
		// It seems other listers do the same here - they never receive the termination msg on the ch.
		// This should be currently OK.
		stopChannel := make(chan struct{})
		lister := kubernetes.NewConfigMapListerForNamespace(kubeClient, stopChannel, configNamespace)
		return priority.NewStrategy(lister.ConfigMaps(configNamespace), autoscalingKubeClients.Recorder)
	}
	return nil, errors.NewAutoscalerError(errors.InternalError, "Expander %s not supported", expanderFlag)
}
