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
	"k8s.io/autoscaler/cluster-autoscaler/expander/leastnodes"
	"k8s.io/autoscaler/cluster-autoscaler/expander/mostpods"
	"k8s.io/autoscaler/cluster-autoscaler/expander/price"
	"k8s.io/autoscaler/cluster-autoscaler/expander/priority"
	"k8s.io/autoscaler/cluster-autoscaler/expander/random"
	"k8s.io/autoscaler/cluster-autoscaler/expander/waste"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"

	kube_client "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

// Factory can create expander.Strategy based on provided expander names.
type Factory struct {
	createFunc map[string]func() expander.Filter
}

// NewFactory returns a new Factory.
func NewFactory() *Factory {
	return &Factory{
		createFunc: make(map[string]func() expander.Filter),
	}
}

// RegisterFilter registers a function that can provision a new expander.Filter under the specified name.
func (f *Factory) RegisterFilter(name string, createFunc func() expander.Filter) {
	f.createFunc[name] = createFunc
}

// Build creates a new expander.Strategy based on a list of expander.Filter names.
func (f *Factory) Build(names []string) (expander.Strategy, errors.AutoscalerError) {
	var filters []expander.Filter
	seenExpanders := map[string]struct{}{}
	strategySeen := false
	for i, name := range names {
		if _, ok := seenExpanders[name]; ok {
			return nil, errors.NewAutoscalerErrorf(errors.InternalError, "Expander %s was specified multiple times, each expander must not be specified more than once", name)
		}
		if strategySeen {
			return nil, errors.NewAutoscalerErrorf(errors.InternalError, "Expander %s came after an expander %s that will always return only one result, this is not allowed since %s will never be used", name, names[i-1], name)
		}
		seenExpanders[name] = struct{}{}

		create, known := f.createFunc[name]
		if known {
			filters = append(filters, create())
		} else {
			return nil, errors.NewAutoscalerErrorf(errors.InternalError, "Expander %s not supported", name)
		}
		if _, ok := filters[len(filters)-1].(expander.Strategy); ok {
			strategySeen = true
		}
	}
	return newChainStrategy(filters, random.NewStrategy()), nil
}

// RegisterDefaultExpanders is a convenience function, registering all known expanders in the Factory.
func (f *Factory) RegisterDefaultExpanders(cloudProvider cloudprovider.CloudProvider, autoscalingKubeClients *context.AutoscalingKubeClients, kubeClient kube_client.Interface, configNamespace string, GRPCExpanderCert string, GRPCExpanderURL string) {
	f.RegisterFilter(expander.RandomExpanderName, random.NewFilter)
	f.RegisterFilter(expander.MostPodsExpanderName, mostpods.NewFilter)
	f.RegisterFilter(expander.LeastWasteExpanderName, waste.NewFilter)
	f.RegisterFilter(expander.LeastNodesExpanderName, leastnodes.NewFilter)
	f.RegisterFilter(expander.PriceBasedExpanderName, func() expander.Filter {
		if _, err := cloudProvider.Pricing(); err != nil {
			klog.Fatalf("Couldn't access cloud provider pricing for %s expander: %v", expander.PriceBasedExpanderName, err)
		}
		return price.NewFilter(cloudProvider, price.NewSimplePreferredNodeProvider(autoscalingKubeClients.AllNodeLister()), price.SimpleNodeUnfitness)
	})
	f.RegisterFilter(expander.PriorityBasedExpanderName, func() expander.Filter {
		// It seems other listers do the same here - they never receive the termination msg on the ch.
		// This should be currently OK.
		stopChannel := make(chan struct{})
		lister := kubernetes.NewConfigMapListerForNamespace(kubeClient, stopChannel, configNamespace)
		return priority.NewFilter(lister.ConfigMaps(configNamespace), autoscalingKubeClients.Recorder)
	})
	f.RegisterFilter(expander.GRPCExpanderName, func() expander.Filter { return grpcplugin.NewFilter(GRPCExpanderCert, GRPCExpanderURL) })
}
