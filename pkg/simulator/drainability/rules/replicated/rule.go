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

package replicated

import (
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/drainability"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/utils/drain"
)

// Rule is a drainability rule on how to handle replicated pods.
type Rule struct {
	skipNodesWithCustomControllerPods bool
}

// New creates a new Rule.
func New(skipNodesWithCustomControllerPods bool) *Rule {
	return &Rule{
		skipNodesWithCustomControllerPods: skipNodesWithCustomControllerPods,
	}
}

// Name returns the name of the rule.
func (r *Rule) Name() string {
	return "Replicated"
}

// Drainable decides what to do with replicated pods on node drain.
func (r *Rule) Drainable(drainCtx *drainability.DrainContext, pod *apiv1.Pod, _ *framework.NodeInfo) drainability.Status {
	controllerRef := drain.ControllerRef(pod)
	replicated := controllerRef != nil

	if r.skipNodesWithCustomControllerPods {
		// TODO(vadasambar): remove this when we get rid of skipNodesWithCustomControllerPods
		replicated = replicated && replicatedKind[controllerRef.Kind]
	}

	if !replicated {
		return drainability.NewBlockedStatus(drain.NotReplicated, fmt.Errorf("%s/%s is not replicated", pod.Namespace, pod.Name))
	}
	return drainability.NewUndefinedStatus()
}

// replicatedKind returns true if this kind has replicates pods.
var replicatedKind = map[string]bool{
	"ReplicationController": true,
	"Job":                   true,
	"ReplicaSet":            true,
	"StatefulSet":           true,
}
