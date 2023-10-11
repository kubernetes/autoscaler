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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/drainability"
	"k8s.io/autoscaler/cluster-autoscaler/utils/drain"
	pod_util "k8s.io/autoscaler/cluster-autoscaler/utils/pod"
)

// Rule is a drainability rule on how to handle replicated pods.
type Rule struct {
	skipNodesWithCustomControllerPods bool
	minReplicaCount                   int
}

// New creates a new Rule.
func New(skipNodesWithCustomControllerPods bool, minReplicaCount int) *Rule {
	return &Rule{
		skipNodesWithCustomControllerPods: skipNodesWithCustomControllerPods,
		minReplicaCount:                   minReplicaCount,
	}
}

// Drainable decides what to do with replicated pods on node drain.
func (r *Rule) Drainable(drainCtx *drainability.DrainContext, pod *apiv1.Pod) drainability.Status {
	controllerRef := drain.ControllerRef(pod)
	replicated := controllerRef != nil

	if r.skipNodesWithCustomControllerPods {
		// TODO(vadasambar): remove this branch when we get rid of skipNodesWithCustomControllerPods.
		status := legacyStatus(drainCtx, pod, r.minReplicaCount)
		if status.Outcome != drainability.UndefinedOutcome {
			return status
		}
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

func legacyStatus(drainCtx *drainability.DrainContext, pod *apiv1.Pod, minReplicaCount int) drainability.Status {
	if drainCtx.Listers == nil {
		return drainability.NewUndefinedStatus()
	}

	// For now, owner controller must be in the same namespace as the pod
	// so OwnerReference doesn't have its own Namespace field.
	controllerNamespace := pod.Namespace

	controllerRef := drain.ControllerRef(pod)
	if controllerRef == nil {
		return drainability.NewUndefinedStatus()
	}
	refKind := controllerRef.Kind

	if refKind == "ReplicationController" {
		rc, err := drainCtx.Listers.ReplicationControllerLister().ReplicationControllers(controllerNamespace).Get(controllerRef.Name)
		// Assume RC is either gone/missing or has too few replicas configured.
		if err != nil || rc == nil {
			return drainability.NewBlockedStatus(drain.ControllerNotFound, fmt.Errorf("replication controller for %s/%s is not available, err: %v", pod.Namespace, pod.Name, err))
		}

		// TODO: Replace the minReplica check with PDB.
		if rc.Spec.Replicas != nil && int(*rc.Spec.Replicas) < minReplicaCount {
			return drainability.NewBlockedStatus(drain.MinReplicasReached, fmt.Errorf("replication controller for %s/%s has too few replicas spec: %d min: %d", pod.Namespace, pod.Name, rc.Spec.Replicas, minReplicaCount))
		}
	} else if pod_util.IsDaemonSetPod(pod) {
		if refKind != "DaemonSet" {
			// We don't have a listener for the other DaemonSet kind.
			// TODO: Use a generic client for checking the reference.
			return drainability.NewUndefinedStatus()
		}

		_, err := drainCtx.Listers.DaemonSetLister().DaemonSets(controllerNamespace).Get(controllerRef.Name)
		if err != nil {
			if apierrors.IsNotFound(err) {
				return drainability.NewBlockedStatus(drain.ControllerNotFound, fmt.Errorf("daemonset for %s/%s is not present, err: %v", pod.Namespace, pod.Name, err))
			}
			return drainability.NewBlockedStatus(drain.UnexpectedError, fmt.Errorf("error when trying to get daemonset for %s/%s , err: %v", pod.Namespace, pod.Name, err))
		}
	} else if refKind == "Job" {
		job, err := drainCtx.Listers.JobLister().Jobs(controllerNamespace).Get(controllerRef.Name)

		if err != nil || job == nil {
			// Assume the only reason for an error is because the Job is gone/missing.
			return drainability.NewBlockedStatus(drain.ControllerNotFound, fmt.Errorf("job for %s/%s is not available: err: %v", pod.Namespace, pod.Name, err))
		}
	} else if refKind == "ReplicaSet" {
		rs, err := drainCtx.Listers.ReplicaSetLister().ReplicaSets(controllerNamespace).Get(controllerRef.Name)

		if err == nil && rs != nil {
			// Assume the only reason for an error is because the RS is gone/missing.
			if rs.Spec.Replicas != nil && int(*rs.Spec.Replicas) < minReplicaCount {
				return drainability.NewBlockedStatus(drain.MinReplicasReached, fmt.Errorf("replication controller for %s/%s has too few replicas spec: %d min: %d", pod.Namespace, pod.Name, rs.Spec.Replicas, minReplicaCount))
			}
		} else {
			return drainability.NewBlockedStatus(drain.ControllerNotFound, fmt.Errorf("replication controller for %s/%s is not available, err: %v", pod.Namespace, pod.Name, err))
		}
	} else if refKind == "StatefulSet" {
		ss, err := drainCtx.Listers.StatefulSetLister().StatefulSets(controllerNamespace).Get(controllerRef.Name)

		if err != nil && ss == nil {
			// Assume the only reason for an error is because the SS is gone/missing.
			return drainability.NewBlockedStatus(drain.ControllerNotFound, fmt.Errorf("statefulset for %s/%s is not available: err: %v", pod.Namespace, pod.Name, err))
		}
	}

	return drainability.NewUndefinedStatus()
}
