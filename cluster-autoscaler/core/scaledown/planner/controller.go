/*
Copyright 2019 The Kubernetes Authors.

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

package planner

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
)

type controllerCalculatorImpl struct {
	listers kubernetes.ListerRegistry
}

func newControllerReplicasCalculator(listers kubernetes.ListerRegistry) controllerReplicasCalculator {
	return &controllerCalculatorImpl{listers: listers}
}

func (c *controllerCalculatorImpl) getReplicas(ownerRef metav1.OwnerReference, namespace string) (*replicasInfo, error) {
	result := &replicasInfo{}
	switch ownerRef.Kind {
	case "StatefulSet":
		sSet, err := c.listers.StatefulSetLister().StatefulSets(namespace).Get(ownerRef.Name)
		if err != nil {
			return nil, err
		}
		result.currentReplicas = sSet.Status.CurrentReplicas
		if sSet.Spec.Replicas != nil {
			result.targetReplicas = *sSet.Spec.Replicas
		} else {
			result.targetReplicas = 1
		}
	case "ReplicaSet":
		rSet, err := c.listers.ReplicaSetLister().ReplicaSets(namespace).Get(ownerRef.Name)
		if err != nil {
			return nil, err
		}
		result.currentReplicas = rSet.Status.Replicas
		if rSet.Spec.Replicas != nil {
			result.targetReplicas = *rSet.Spec.Replicas
		} else {
			result.targetReplicas = 1
		}
	case "ReplicationController":
		rController, err := c.listers.ReplicationControllerLister().ReplicationControllers(namespace).Get(ownerRef.Name)
		if err != nil {
			return nil, err
		}
		result.currentReplicas = rController.Status.Replicas
		if rController.Spec.Replicas != nil {
			result.targetReplicas = *rController.Spec.Replicas
		} else {
			result.targetReplicas = 1
		}
	case "Job":
		job, err := c.listers.JobLister().Jobs(namespace).Get(ownerRef.Name)
		if err != nil {
			return nil, err
		}
		result.currentReplicas = job.Status.Active
		if job.Spec.Parallelism != nil {
			result.targetReplicas = *job.Spec.Parallelism
		} else {
			result.targetReplicas = 1
		}
		if job.Spec.Completions != nil && *job.Spec.Completions-job.Status.Succeeded < result.targetReplicas {
			result.targetReplicas = *job.Spec.Completions - job.Status.Succeeded
		}
	default:
		return nil, fmt.Errorf("unhandled controller type: %s", ownerRef.Kind)
	}
	return result, nil
}
