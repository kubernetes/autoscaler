/*
Copyright 2024 The Kubernetes Authors.

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

package forcescaledown

import (
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/drainability"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	"k8s.io/klog/v2"
)

// Rule is a drainability rule on how to handle pods on scale-down-requested nodes.
type Rule struct{}

// New creates a new Rule.
func New() *Rule {
	return &Rule{}
}

// Name returns the name of the rule.
func (r *Rule) Name() string {
	return "FoceScaleDown"
}

// Drainable decides what to do with force-scale-down pods on node drain.
func (r *Rule) Drainable(drainCtx *drainability.DrainContext, pod *apiv1.Pod) drainability.Status {
	podDesc := fmt.Sprintf("pod %s/%s on node %s", pod.Namespace, pod.Name, pod.Spec.NodeName)
	if drainCtx.Listers == nil || drainCtx.Listers.AllNodeLister() == nil {
		return drainability.NewUndefinedStatus()
	}
	if len(pod.Spec.NodeName) == 0 {
		klog.V(6).Infof("Skipped force-scale-down drainable check for %s, because the pod does not have a node yet", podDesc)
		return drainability.NewUndefinedStatus()
	}
	node, err := drainCtx.Listers.AllNodeLister().Get(pod.Spec.NodeName)
	if err != nil {
		klog.Warningf("Skipped force-scale-down drainable check for %s, because failed to get node: %v", podDesc, err)
		return drainability.NewUndefinedStatus()
	}
	if !taints.HasForceScaleDownTaint(node) {
		klog.V(6).Infof("Skipped force-scale-down drainable check for %s, because the node does not have force-scale-down taint", podDesc)
		return drainability.NewUndefinedStatus()
	}
	deadline := taints.GetForceScaleDownDeadline(node)
	if deadline == nil {
		klog.V(2).Infof("Marking %s as drainable, because the tainted node does not have a force-scale-down deadline", podDesc)
		return drainability.NewDrainableStatus()
	}
	if drainCtx.Timestamp.Before(*deadline) {
		klog.V(2).Infof("Marking %s as drainable, because the node has force-scale-down taint, and it is before the deadline %v", podDesc, *deadline)
		return drainability.NewDrainableStatus()
	}
	klog.V(2).Infof("Marking %s as drain skip, because the node has force-scale-down taint, and it exceeded the deadline %v", podDesc, *deadline)
	return drainability.NewSkipStatus()
}
