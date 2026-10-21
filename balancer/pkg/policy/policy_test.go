/*
Copyright 2026 The Kubernetes Authors.

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
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/balancer/pkg/apis/balancer.x-k8s.io/v1alpha1"
	"k8s.io/autoscaler/balancer/pkg/pods"
)

func TestGetPlacementPriorityTargetOrder(t *testing.T) {
	newBalancer := func(targetOrder []string) *v1alpha1.Balancer {
		return &v1alpha1.Balancer{
			Spec: v1alpha1.BalancerSpec{
				Replicas: 5,
				Targets: []v1alpha1.BalancerTarget{
					{Name: "a"},
					{Name: "b"},
				},
				Policy: v1alpha1.BalancerPolicy{
					PolicyName: v1alpha1.PriorityPolicyName,
					Priorities: &v1alpha1.PriorityPolicy{TargetOrder: targetOrder},
				},
			},
		}
	}
	summaries := map[string]pods.Summary{}

	placement, _, err := GetPlacement(newBalancer([]string{"a", "b"}), summaries)
	assert.NoError(t, err)
	assert.Equal(t, ReplicaPlacement{"a": 5, "b": 0}, placement)

	// An unknown target used to dereference a nil targetInfo.
	assert.NotPanics(t, func() {
		_, _, err = GetPlacement(newBalancer([]string{"a", "c"}), summaries)
	})
	assert.ErrorContains(t, err, `unknown target "c"`)
}
