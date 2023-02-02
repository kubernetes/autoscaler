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

package policy

import (
	"fmt"
	"k8s.io/autoscaling/balancer/pkg/apis/balancer.x-k8s.io/v1alpha1"
	"k8s.io/autoscaling/balancer/pkg/pods"
)

// GetPlacement calculates the placement for the given balancer and pod summary
// information for individual balancer targets.
func GetPlacement(balancer *v1alpha1.Balancer, summaries map[string]pods.Summary) (ReplicaPlacement, PlacementProblems, error) {
	targetMap := buildTargetMap(balancer.Spec.Targets)
	switch balancer.Spec.Policy.PolicyName {
	case v1alpha1.PriorityPolicyName:
		if balancer.Spec.Policy.Priorities == nil {
			return nil, PlacementProblems{}, fmt.Errorf("incomplete policy definition: missing priorities")
		}
		if balancer.Spec.Policy.Priorities.TargetOrder == nil {
			return nil, PlacementProblems{}, fmt.Errorf("incomplete policy definition: missing targetOrder")
		}
		infos := buildTargetInfoMapForPriority(targetMap, summaries)
		placement, problems := distributeByPriority(balancer.Spec.Replicas, balancer.Spec.Policy.Priorities.TargetOrder, infos)
		return placement, problems, nil

	case v1alpha1.ProportionalPolicyName:
		if balancer.Spec.Policy.Proportions == nil {
			return nil, PlacementProblems{}, fmt.Errorf("incomplete policy definition: missing proportions")
		}
		if balancer.Spec.Policy.Proportions.TargetProportions == nil {
			return nil, PlacementProblems{}, fmt.Errorf("incomplete policy definition: missing targetProportions")
		}
		infos := buildTargetInfoMapForProportional(targetMap, summaries, balancer.Spec.Policy.Proportions.TargetProportions)
		placement, problems := distributeByProportions(balancer.Spec.Replicas, infos)
		return placement, problems, nil

	default:
		return nil, PlacementProblems{}, fmt.Errorf("policy not supported: %v", balancer.Spec.Policy.PolicyName)
	}
}
