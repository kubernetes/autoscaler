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

package podinjection

import (
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/klog/v2"
)

func createStatefulSetControllers(ctx *context.AutoscalingContext) []controller {
	var controllers []controller
	statefulSets, err := ctx.ListerRegistry.StatefulSetLister().List(labels.Everything())
	if err != nil {
		klog.Errorf("Failed to list statefulsets: %v", err)
		return controllers
	}
	for _, statefulSet := range statefulSets {
		// Non parallel pod management (OrderedReadyPodManagement) waits for a pod to be ready to create another one
		// Which becomes very slow (can take 2+ mins per pod)
		// Making fast scale ups not a priority and fake pod injection costly
		if statefulSet.Spec.PodManagementPolicy == appsv1.ParallelPodManagement {
			controllers = append(controllers, controller{uid: statefulSet.UID, desiredReplicas: desiredReplicasFromStatefulSet(statefulSet)})
		}
	}
	return controllers
}

func desiredReplicasFromStatefulSet(statefulSet *appsv1.StatefulSet) int {
	if statefulSet.Spec.Replicas == nil {
		return 0
	}
	return int(*statefulSet.Spec.Replicas)
}
