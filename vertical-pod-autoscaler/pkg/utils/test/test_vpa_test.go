/*
Copyright The Kubernetes Authors.

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

package test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
)

func TestVerticalPodAutoscalerBuilderCopies(t *testing.T) {
	base := VerticalPodAutoscaler().WithContainer("container").WithUpdateMode(vpa_types.UpdateModeRecreate)

	_ = base.
		WithMinAllowed("container", "1", "1Gi").
		WithMaxAllowed("container", "2", "2Gi").
		WithControlledValues("container", vpa_types.ContainerControlledValuesRequestsOnly).
		WithScalingMode("container", vpa_types.ContainerScalingModeOff).
		WithContainerCPUStartupBoost("container", vpa_types.FactorStartupBoostType, new(int32(2)), nil, 10).
		WithUpdateMode(vpa_types.UpdateModeOff).
		WithMinReplicas(new(int32(2))).
		AppendCondition(vpa_types.RecommendationProvided, corev1.ConditionTrue, "", "", time.Now()).
		AppendRecommendation(vpa_types.RecommendedContainerResources{ContainerName: "other"})

	vpa := base.Get()
	policy := vpa.Spec.ResourcePolicy.ContainerPolicies[0]
	assert.Nil(t, policy.MinAllowed)
	assert.Nil(t, policy.MaxAllowed)
	assert.Nil(t, policy.ControlledValues)
	assert.Equal(t, vpa_types.ContainerScalingModeAuto, *policy.Mode)
	assert.Nil(t, policy.StartupBoost)
	assert.Equal(t, vpa_types.UpdateModeRecreate, *vpa.Spec.UpdatePolicy.UpdateMode)
	assert.Nil(t, vpa.Spec.UpdatePolicy.MinReplicas)
	assert.Empty(t, vpa.Status.Conditions)
	assert.Empty(t, vpa.Status.Recommendation.ContainerRecommendations)
}
