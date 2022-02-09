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

package kubernetes

import (
	"fmt"
	"time"

	apiv1 "k8s.io/api/core/v1"
)

// NodeNotReadyReason reprents a reason for node to be unready. While it is
// simply a string on the node object, custom type ensures no one accidentally
// performs any string operation on variables of this type and allows them to
// be treated as enums.
type NodeNotReadyReason string

const (
	// ResourceUnready is a fake identifier used internally by Cluster Autoscaler
	// to indicate nodes that appear Ready in the API, but are treated as
	// still upcoming due to a missing resource (e.g. GPU).
	ResourceUnready NodeNotReadyReason = "cluster-autoscaler.kubernetes.io/resource-not-ready"

	// IgnoreTaint is a fake identifier used internally by Cluster Autoscaler
	// to indicate nodes that appear Ready in the API, but are treated as
	// still upcoming due to applied ignore taint.
	IgnoreTaint NodeNotReadyReason = "cluster-autoscaler.kubernetes.io/ignore-taint"
)

// IsNodeReadyAndSchedulable returns true if the node is ready and schedulable.
func IsNodeReadyAndSchedulable(node *apiv1.Node) bool {
	ready, _, _ := GetReadinessState(node)
	if !ready {
		return false
	}
	// Ignore nodes that are marked unschedulable
	if node.Spec.Unschedulable {
		return false
	}
	return true
}

// NodeReadiness represents the last known node readiness.
type NodeReadiness struct {
	// Is the node ready or not.
	Ready bool
	// Time of the last state transition related to readiness.
	LastTransitionTime time.Time
	// Reason for the node to be unready. Defined only when Ready is false.
	Reason NodeNotReadyReason
}

// GetReadinessState gets readiness state for the node
//
// Deprecated: Use GetNodeReadiness instead.
func GetReadinessState(node *apiv1.Node) (isNodeReady bool, lastTransitionTime time.Time, err error) {
	nr, err := GetNodeReadiness(node)
	return nr.Ready, nr.LastTransitionTime, err
}

// GetNodeReadiness gets readiness for the node
func GetNodeReadiness(node *apiv1.Node) (NodeReadiness, error) {
	canNodeBeReady, readyFound := true, false
	lastTransitionTime := time.Time{}
	var reason NodeNotReadyReason

	for _, cond := range node.Status.Conditions {
		switch cond.Type {
		case apiv1.NodeReady:
			readyFound = true
			if cond.Status == apiv1.ConditionFalse || cond.Status == apiv1.ConditionUnknown {
				canNodeBeReady = false
				reason = NodeNotReadyReason(cond.Reason)
			}
			if lastTransitionTime.Before(cond.LastTransitionTime.Time) {
				lastTransitionTime = cond.LastTransitionTime.Time
			}
		case apiv1.NodeDiskPressure:
			if cond.Status == apiv1.ConditionTrue {
				canNodeBeReady = false
			}
			if lastTransitionTime.Before(cond.LastTransitionTime.Time) {
				lastTransitionTime = cond.LastTransitionTime.Time
			}
		case apiv1.NodeNetworkUnavailable:
			if cond.Status == apiv1.ConditionTrue {
				canNodeBeReady = false
			}
			if lastTransitionTime.Before(cond.LastTransitionTime.Time) {
				lastTransitionTime = cond.LastTransitionTime.Time
			}
		}
	}

	notReadyTaints := map[string]bool{
		apiv1.TaintNodeNotReady:           true,
		apiv1.TaintNodeDiskPressure:       true,
		apiv1.TaintNodeNetworkUnavailable: true,
	}
	for _, taint := range node.Spec.Taints {
		if notReadyTaints[taint.Key] {
			canNodeBeReady = false
			if taint.TimeAdded != nil && lastTransitionTime.Before(taint.TimeAdded.Time) {
				lastTransitionTime = taint.TimeAdded.Time
			}
		}
	}

	if !readyFound {
		return NodeReadiness{}, fmt.Errorf("readiness information not found")
	}
	return NodeReadiness{
		Ready:              canNodeBeReady,
		LastTransitionTime: lastTransitionTime,
		Reason:             reason,
	}, nil
}

// GetUnreadyNodeCopy create a copy of the given node and override its NodeReady condition to False
func GetUnreadyNodeCopy(node *apiv1.Node, reason NodeNotReadyReason) *apiv1.Node {
	newNode := node.DeepCopy()
	newReadyCondition := apiv1.NodeCondition{
		Type:               apiv1.NodeReady,
		Status:             apiv1.ConditionFalse,
		LastTransitionTime: node.CreationTimestamp,
		Reason:             string(reason),
	}
	newNodeConditions := []apiv1.NodeCondition{newReadyCondition}
	for _, condition := range newNode.Status.Conditions {
		if condition.Type != apiv1.NodeReady {
			newNodeConditions = append(newNodeConditions, condition)
		}
	}
	newNode.Status.Conditions = newNodeConditions
	return newNode
}
