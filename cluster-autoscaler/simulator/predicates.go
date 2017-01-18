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

package simulator

import (
	"fmt"

	kube_util "k8s.io/contrib/cluster-autoscaler/utils/kubernetes"
	apiv1 "k8s.io/kubernetes/pkg/api/v1"
	kube_client "k8s.io/kubernetes/pkg/client/clientset_generated/release_1_5"
	"k8s.io/kubernetes/plugin/pkg/scheduler/algorithm"
	"k8s.io/kubernetes/plugin/pkg/scheduler/algorithm/predicates"
	"k8s.io/kubernetes/plugin/pkg/scheduler/factory"
	"k8s.io/kubernetes/plugin/pkg/scheduler/schedulercache"

	// We need to import provider to intialize default scheduler.
	_ "k8s.io/kubernetes/plugin/pkg/scheduler/algorithmprovider"
)

// PredicateChecker checks whether all required predicates are matched for given Pod and Node
type PredicateChecker struct {
	predicates map[string]algorithm.FitPredicate
}

// NewPredicateChecker builds PredicateChecker.
func NewPredicateChecker(kubeClient kube_client.Interface) (*PredicateChecker, error) {
	provider, err := factory.GetAlgorithmProvider(factory.DefaultProvider)
	if err != nil {
		return nil, err
	}
	schedulerConfigFactory := factory.NewConfigFactory(kubeClient, "", apiv1.DefaultHardPodAffinitySymmetricWeight, apiv1.DefaultFailureDomains)
	predicates, err := schedulerConfigFactory.GetPredicates(provider.FitPredicateKeys)
	predicates["ready"] = isNodeReadyAndSchedulablePredicate
	if err != nil {
		return nil, err
	}
	schedulerConfigFactory.Run()
	return &PredicateChecker{
		predicates: predicates,
	}, nil
}

func isNodeReadyAndSchedulablePredicate(pod *apiv1.Pod, meta interface{}, nodeInfo *schedulercache.NodeInfo) (bool,
	[]algorithm.PredicateFailureReason, error) {
	ready := kube_util.IsNodeReadyAndSchedulable(nodeInfo.Node())
	if !ready {
		return false, []algorithm.PredicateFailureReason{predicates.NewFailureReason("node is unready")}, nil
	}
	return true, []algorithm.PredicateFailureReason{}, nil
}

// NewTestPredicateChecker builds test version of PredicateChecker.
func NewTestPredicateChecker() *PredicateChecker {
	return &PredicateChecker{
		predicates: map[string]algorithm.FitPredicate{
			"default": predicates.GeneralPredicates,
			"ready":   isNodeReadyAndSchedulablePredicate,
		},
	}
}

// FitsAny checks if the given pod can be place on any of the given nodes.
func (p *PredicateChecker) FitsAny(pod *apiv1.Pod, nodeInfos map[string]*schedulercache.NodeInfo) (string, error) {
	for name, nodeInfo := range nodeInfos {
		// Be sure that the node is schedulable.
		if nodeInfo.Node().Spec.Unschedulable {
			continue
		}
		if err := p.CheckPredicates(pod, nodeInfo); err == nil {
			return name, nil
		}
	}
	return "", fmt.Errorf("cannot put pod %s on any node", pod.Name)
}

// CheckPredicates checks if the given pod can be placed on the given node.
func (p *PredicateChecker) CheckPredicates(pod *apiv1.Pod, nodeInfo *schedulercache.NodeInfo) error {
	for _, predicate := range p.predicates {
		match, failureReason, err := predicate(pod, nil, nodeInfo)

		nodename := "unknown"
		if nodeInfo.Node() != nil {
			nodename = nodeInfo.Node().Name
		}
		if err != nil {
			return fmt.Errorf("cannot put %s on %s due to %v", pod.Name, nodename, err)
		}
		if !match {
			return fmt.Errorf("cannot put %s on %s, reason: %v", pod.Name, nodename, failureReason)
		}
	}
	return nil
}
