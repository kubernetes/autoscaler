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
	"k8s.io/autoscaling/balancer/pkg/apis/balancer.x-k8s.io/v1alpha1"
	"k8s.io/autoscaling/balancer/pkg/pods"
)

const (
	// maxReplicas is the maximum number of replicas supported by the algorithm.
	// It is used as a sentinel value to simplify the code.
	maxReplicas = 1000000000
)

// buildTargetMap build map of BalancerTargets, keyed with target's name.
func buildTargetMap(targets []v1alpha1.BalancerTarget) map[string]v1alpha1.BalancerTarget {
	result := make(map[string]v1alpha1.BalancerTarget)
	for _, target := range targets {
		result[target.Name] = target
	}
	return result
}

// ReplicaPlacement defines the number of replicas for each of the BalancerTargets.
type ReplicaPlacement map[string]int32

// targetInfo is the aggregated information about a Balancer target.
type targetInfo struct {
	// min replica count taken from BalancerTarget.
	min int32
	// max replica count taken from BalancerTarget.
	max int32
	// count of pods of given type based on pod listener data.
	summary pods.Summary

	// proportion taken from ProportionalPolicy. 0 for other policies.
	proportion int32
}

// PlacementProblems contains information about replicas that were problematic
// when applying placement policy and constraints.
type PlacementProblems struct {
	// MissingReplicas is the number of replicas that had to be added
	// on top of the desired count to match the constraint and policy.
	MissingReplicas int32
	// OverflowReplicas is the number of replicas that could not be placed
	// due to constraints and policy.
	OverflowReplicas int32
}

// buildTargetInfoMap aggregates information scattered across multiple
// maps into a single one. It assumes that all inputs are already validated
// and consistent.
func buildTargetInfoMapForProportional(
	targetMap map[string]v1alpha1.BalancerTarget,
	summaryMap map[string]pods.Summary,
	proportions map[string]int32) map[string]*targetInfo {

	result := make(map[string]*targetInfo)
	for name, target := range targetMap {
		summary := summaryMap[name]
		prop := proportions[name]
		result[name] = &targetInfo{
			proportion: prop,
			summary:    summary,
			min:        0,
			max:        maxReplicas,
		}
		if target.MinReplicas != nil {
			result[name].min = *target.MinReplicas
		}
		if target.MaxReplicas != nil {
			result[name].max = *target.MaxReplicas
		}
	}
	return result
}

// buildTargetInfoMap aggregates information scattered across multiple
// maps into a single one. It assumes that all inputs are already validated
// and consistent.
func buildTargetInfoMapForPriority(
	targetMap map[string]v1alpha1.BalancerTarget,
	summaryMap map[string]pods.Summary) map[string]*targetInfo {

	return buildTargetInfoMapForProportional(targetMap, summaryMap, map[string]int32{})
}
