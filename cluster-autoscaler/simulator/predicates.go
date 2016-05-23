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
	"k8s.io/kubernetes/plugin/pkg/scheduler/algorithm/predicates"
	"k8s.io/kubernetes/plugin/pkg/scheduler/schedulercache"
)

// PredicateChecker checks whether all required predicates are matched for given Pod and Node
type PredicateChecker struct {
}

// NewPredicateChecker builds PredicateChecker.
func NewPredicateChecker() *PredicateChecker {
	// TODO(fgrzadkowsi): Get a full list of all predicates.
	return &PredicateChecker{}
}

// FitsAny checks if the given pod can be place on any of the given nodes.
func (p *PredicateChecker) FitsAny(pod *kube_api.Pod, nodeInfos map[string]*schedulercache.NodeInfo) (string, error) {
	for name, nodeInfo := range nodeInfos {
		if err := p.CheckPredicates(pod, nodeInfo); err == nil {
			return name, nil
		}
	}
	return "", fmt.Errorf("cannot put pod %s on any node", pod.Name)
}

// CheckPredicates checks if the given pod can be placed on the given node.
func (p *PredicateChecker) CheckPredicates(pod *kube_api.Pod, nodeInfo *schedulercache.NodeInfo) error {
	// TODO(fgrzadkowski): Use full list of predicates.
	match, err := predicates.GeneralPredicates(pod, nodeInfo)
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
	return nil
}
