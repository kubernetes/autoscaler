/*
Copyright 2023 The Kubernetes Authors.

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

package pdb

import (
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/drainability"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/utils/drain"
)

// Rule is a drainability rule on how to handle pods with pdbs.
type Rule struct{}

// New creates a new Rule.
func New() *Rule {
	return &Rule{}
}

// Name returns the name of the rule.
func (r *Rule) Name() string {
	return "PDB"
}

// Drainable decides how to handle pods with pdbs on node drain.
func (Rule) Drainable(drainCtx *drainability.DrainContext, pod *apiv1.Pod, _ *framework.NodeInfo) drainability.Status {
	for _, pdb := range drainCtx.RemainingPdbTracker.MatchingPdbs(pod) {
		if pdb.Status.DisruptionsAllowed < 1 {
			return drainability.NewBlockedStatus(drain.NotEnoughPdb, fmt.Errorf("not enough pod disruption budget to move %s/%s", pod.Namespace, pod.Name))
		}
	}
	return drainability.NewUndefinedStatus()
}
