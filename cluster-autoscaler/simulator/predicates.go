/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package simulator

import (
	"fmt"

	kube_api "k8s.io/kubernetes/pkg/api"
	kube_client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/plugin/pkg/scheduler/algorithm"
	// We need to import provider to intialize default scheduler.
	_ "k8s.io/kubernetes/plugin/pkg/scheduler/algorithmprovider"
	"k8s.io/kubernetes/plugin/pkg/scheduler/factory"
	"k8s.io/kubernetes/plugin/pkg/scheduler/schedulercache"
)

// PredicateChecker checks whether all required predicates are matched for given Pod and Node
type PredicateChecker struct {
	predicates map[string]algorithm.FitPredicate
}

// NewPredicateChecker builds PredicateChecker.
func NewPredicateChecker(kubeClient *kube_client.Client) (*PredicateChecker, error) {
	provider, err := factory.GetAlgorithmProvider(factory.DefaultProvider)
	if err != nil {
		return nil, err
	}
	schedulerConfigFactory := factory.NewConfigFactory(kubeClient, "", kube_api.DefaultHardPodAffinitySymmetricWeight, kube_api.DefaultFailureDomains)
	predicates, err := schedulerConfigFactory.GetPredicates(provider.FitPredicateKeys)
	if err != nil {
		return nil, err
	}
	schedulerConfigFactory.Run()
	return &PredicateChecker{
		predicates: predicates,
	}, nil
}

// CheckPredicates Checks if the given pod can be placed on the given node.
func (p *PredicateChecker) CheckPredicates(pod *kube_api.Pod, nodeInfo *schedulercache.NodeInfo) error {
	for _, predicate := range p.predicates {
		match, err := predicate(pod, nodeInfo)
		nodename := "unknown"
		if nodeInfo.Node() != nil {
			nodename = nodeInfo.Node().Name
		}
		if err != nil {
			return fmt.Errorf("cannot put %s on %s due to %v", pod.Name, nodename, err)
		}
		if !match {
			return fmt.Errorf("cannot put %s on %s", pod.Name, nodename)
		}
	}
	return nil
}
