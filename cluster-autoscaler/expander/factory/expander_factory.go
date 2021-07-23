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
	"k8s.io/autoscaler/cluster-autoscaler/expander/grpcplugin"
	"k8s.io/autoscaler/cluster-autoscaler/expander/mostpods"
	"k8s.io/autoscaler/cluster-autoscaler/expander/price"
	"k8s.io/autoscaler/cluster-autoscaler/expander/priority"
	"k8s.io/autoscaler/cluster-autoscaler/expander/random"
	"k8s.io/autoscaler/cluster-autoscaler/expander/waste"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/klog/v2"

	kube_client "k8s.io/client-go/kubernetes"
)

// ExpanderStrategyFromStrings creates an expander.Strategy according to the names of the expanders passed in
// take in whole opts and access stuff here
func ExpanderStrategyFromStrings(expanderFlags []string, cloudProvider cloudprovider.CloudProvider,
	autoscalingKubeClients *context.AutoscalingKubeClients, kubeClient kube_client.Interface,
	configNamespace string, GRPCExpanderCert string, GRPCExpanderURL string) (expander.Strategy, errors.AutoscalerError) {
	var filters []expander.Filter
	seenExpanders := map[string]struct{}{}
	strategySeen := false
	for i, expanderFlag := range expanderFlags {
		if _, ok := seenExpanders[expanderFlag]; ok {
			return nil, errors.NewAutoscalerError(errors.InternalError, "Expander %s was specified multiple times, each expander must not be specified more than once", expanderFlag)
		}
		if strategySeen {
			return nil, errors.NewAutoscalerError(errors.InternalError, "Expander %s came after an expander %s that will always return only one result, this is not allowed since %s will never be used", expanderFlag, expanderFlags[i-1], expanderFlag)
		}
		seenExpanders[expanderFlag] = struct{}{}

		switch expanderFlag {
		case expander.RandomExpanderName:
			filters = append(filters, random.NewFilter())
		case expander.MostPodsExpanderName:
			filters = append(filters, mostpods.NewFilter())
		case expander.LeastWasteExpanderName:
			filters = append(filters, waste.NewFilter())
		case expander.PriceBasedExpanderName:
			if _, err := cloudProvider.Pricing(); err != nil {
				return nil, err
			}
			filters = append(filters, price.NewFilter(cloudProvider,
				price.NewSimplePreferredNodeProvider(autoscalingKubeClients.AllNodeLister()),
				price.SimpleNodeUnfitness))
		case expander.PriorityBasedExpanderName:
			// It seems other listers do the same here - they never receive the termination msg on the ch.
			// This should be currently OK.
			stopChannel := make(chan struct{})
			lister := kubernetes.NewConfigMapListerForNamespace(kubeClient, stopChannel, configNamespace)
			filters = append(filters, priority.NewFilter(lister.ConfigMaps(configNamespace), autoscalingKubeClients.Recorder))
		case expander.GRPCExpanderName:
			klog.V(1).Info("GRPC expander chosen")
			filters = append(filters, grpcplugin.NewFilter(GRPCExpanderCert, GRPCExpanderURL))
		default:
			return nil, errors.NewAutoscalerError(errors.InternalError, "Expander %s not supported", expanderFlag)
		}
		if _, ok := filters[len(filters)-1].(expander.Strategy); ok {
			strategySeen = true
		}
	}
	return newChainStrategy(filters, random.NewStrategy()), nil
}
