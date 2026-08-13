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

package rules

import (
	"context"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/cluster-autoscaler/pkg/core/scaledown/pdb"
	"sigs.k8s.io/cluster-autoscaler/pkg/simulator/drainability"
	"sigs.k8s.io/cluster-autoscaler/pkg/simulator/drainability/rules/daemonset"
	"sigs.k8s.io/cluster-autoscaler/pkg/simulator/drainability/rules/localstorage"
	"sigs.k8s.io/cluster-autoscaler/pkg/simulator/drainability/rules/longterminating"
	"sigs.k8s.io/cluster-autoscaler/pkg/simulator/drainability/rules/mirror"
	"sigs.k8s.io/cluster-autoscaler/pkg/simulator/drainability/rules/notsafetoevict"
	"sigs.k8s.io/cluster-autoscaler/pkg/simulator/drainability/rules/oncompletion"
	pdbrule "sigs.k8s.io/cluster-autoscaler/pkg/simulator/drainability/rules/pdb"
	"sigs.k8s.io/cluster-autoscaler/pkg/simulator/drainability/rules/replicacount"
	"sigs.k8s.io/cluster-autoscaler/pkg/simulator/drainability/rules/replicated"
	"sigs.k8s.io/cluster-autoscaler/pkg/simulator/drainability/rules/safetoevict"
	"sigs.k8s.io/cluster-autoscaler/pkg/simulator/drainability/rules/system"
	"sigs.k8s.io/cluster-autoscaler/pkg/simulator/drainability/rules/terminal"
	"sigs.k8s.io/cluster-autoscaler/pkg/simulator/framework"
	"sigs.k8s.io/cluster-autoscaler/pkg/simulator/options"
)

// Rule determines whether a given pod can be drained or not.
type Rule interface {
	// The name of the rule.
	Name() string
	// Drainable determines whether a given pod is drainable according to
	// the specific Rule.
	//
	// DrainContext cannot be nil.
	Drainable(*drainability.DrainContext, *apiv1.Pod, *framework.NodeInfo) drainability.Status
}

// Default returns the default list of Rules.
func Default(deleteOptions options.NodeDeleteOptions) Rules {
	var rules Rules
	for _, r := range []struct {
		rule Rule
		skip bool
	}{
		{rule: mirror.New()},
		{rule: longterminating.New()},
		{rule: replicacount.New(deleteOptions.MinReplicaCount), skip: !deleteOptions.SkipNodesWithCustomControllerPods},

		// Interrupting checks
		{rule: daemonset.New()},
		{rule: safetoevict.New()},
		{rule: terminal.New()},
		{rule: oncompletion.New()},

		// Blocking checks
		{rule: replicated.New(deleteOptions.SkipNodesWithCustomControllerPods)},
		{rule: system.New(deleteOptions.BspDisruptionTimeout), skip: !deleteOptions.SkipNodesWithSystemPods},
		{rule: notsafetoevict.New()},
		{rule: localstorage.New(), skip: !deleteOptions.SkipNodesWithLocalStorage},
		{rule: pdbrule.New()},
	} {
		if !r.skip {
			rules = append(rules, r.rule)
		}
	}
	return rules
}

// Rules defines operations on a collections of rules.
type Rules []Rule

// Drainable determines whether a given pod is drainable according to the
// specified set of rules.
func (rs Rules) Drainable(ctx context.Context, drainCtx *drainability.DrainContext, pod *apiv1.Pod, nodeInfo *framework.NodeInfo) drainability.Status {
	logger := klog.FromContext(ctx)
	if drainCtx == nil {
		drainCtx = &drainability.DrainContext{}
	}
	if drainCtx.RemainingPdbTracker == nil {
		drainCtx.RemainingPdbTracker = pdb.NewBasicRemainingPdbTracker()
	}

	var candidates []overrideCandidate

	for _, r := range rs {
		status := r.Drainable(drainCtx, pod, nodeInfo)
		if len(status.Overrides) > 0 {
			candidates = append(candidates, overrideCandidate{r.Name(), status})
			continue
		}
		for _, candidate := range candidates {
			for _, override := range candidate.status.Overrides {
				if status.Outcome == override {
					logger.V(5).Info("Overriding pod drainability rule", "pod", klog.KObj(pod), "overriddenRule", r.Name(), "overridingRule", candidate.name, "outcome", candidate.status.Outcome)
					return candidate.status
				}
			}
		}
		if status.Outcome != drainability.UndefinedOutcome {
			return status
		}
	}
	return drainability.NewUndefinedStatus()
}

type overrideCandidate struct {
	name   string
	status drainability.Status
}
