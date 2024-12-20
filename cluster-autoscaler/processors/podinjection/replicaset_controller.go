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

func createReplicaSetControllers(ctx *context.AutoscalingContext) []controller {
	var controllers []controller
	replicaSets, err := ctx.ListerRegistry.ReplicaSetLister().List(labels.Everything())
	if err != nil {
		klog.Errorf("Failed to list replicaSets: %v", err)
		return controllers
	}
	for _, replicaSet := range replicaSets {
		controllers = append(controllers, controller{uid: replicaSet.UID, desiredReplicas: desiredReplicasFromReplicaSet(replicaSet)})
	}
	return controllers
}

func desiredReplicasFromReplicaSet(replicaSet *appsv1.ReplicaSet) int {
	if replicaSet.Spec.Replicas == nil {
		return 0
	}
	return int(*replicaSet.Spec.Replicas)
}
