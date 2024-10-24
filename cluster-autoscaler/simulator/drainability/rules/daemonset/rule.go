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

package daemonset

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/drainability"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	pod_util "k8s.io/autoscaler/cluster-autoscaler/utils/pod"
)

// Rule is a drainability rule on how to handle daemon set pods.
type Rule struct{}

// New creates a new Rule.
func New() *Rule {
	return &Rule{}
}

// Name returns the name of the rule.
func (r *Rule) Name() string {
	return "DaemonSet"
}

// Drainable decides what to do with daemon set pods on node drain.
func (r *Rule) Drainable(drainCtx *drainability.DrainContext, pod *apiv1.Pod, _ *framework.NodeInfo) drainability.Status {
	if pod_util.IsDaemonSetPod(pod) {
		return drainability.NewDrainableStatus()
	}
	return drainability.NewUndefinedStatus()
}
