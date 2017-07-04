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
	"bytes"
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	kube_client "k8s.io/kubernetes/pkg/client/clientset_generated/clientset"
	informers "k8s.io/kubernetes/pkg/client/informers/informers_generated/externalversions"
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
func NewPredicateChecker(kubeClient kube_client.Interface, stop <-chan struct{}) (*PredicateChecker, error) {
	provider, err := factory.GetAlgorithmProvider(factory.DefaultProvider)
	if err != nil {
		return nil, err
	}
	informerFactory := informers.NewSharedInformerFactory(kubeClient, 0)

	schedulerConfigFactory := factory.NewConfigFactory(
		"cluster-autoscaler",
		kubeClient,
		informerFactory.Core().V1().Nodes(),
		informerFactory.Core().V1().Pods(),
		informerFactory.Core().V1().PersistentVolumes(),
		informerFactory.Core().V1().PersistentVolumeClaims(),
		informerFactory.Core().V1().ReplicationControllers(),
		informerFactory.Extensions().V1beta1().ReplicaSets(),
		informerFactory.Apps().V1beta1().StatefulSets(),
		informerFactory.Core().V1().Services(),
		apiv1.DefaultHardPodAffinitySymmetricWeight,
	)

	informerFactory.Start(stop)

	predicates, err := schedulerConfigFactory.GetPredicates(provider.FitPredicateKeys)
	predicates["ready"] = isNodeReadyAndSchedulablePredicate
	if err != nil {
		return nil, err
	}
	// TODO: Verify that run is not needed anymore.
	// schedulerConfigFactory.Run()
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
	for name, predicate := range p.predicates {
		match, failureReason, err := predicate(pod, nil, nodeInfo)

		nodename := "unknown"
		if nodeInfo.Node() != nil {
			nodename = nodeInfo.Node().Name
		}
		if err != nil {
			return fmt.Errorf("%s predicate error, cannot put %s/%s on %s due to, error %v", name, pod.Namespace,
				pod.Name, nodename, err)
		}
		if !match {
			var buffer bytes.Buffer
			for i, reason := range failureReason {
				if i > 0 {
					buffer.WriteString(",")
				}
				buffer.WriteString(reason.GetReason())
			}
			return fmt.Errorf("%s predicate mismatch, cannot put %s/%s on %s, reason: %s", name, pod.Namespace,
				pod.Name, nodename, buffer.String())
		}
	}
	return nil
}
