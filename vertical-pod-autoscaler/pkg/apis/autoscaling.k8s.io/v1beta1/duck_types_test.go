/*
Copyright 2018 The Kubernetes Authors.

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

package v1beta1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestGetContainerResourcePolicy(t *testing.T) {
	containerPolicy1 := ContainerResourcePolicy{
		ContainerName: "container1",
		MinAllowed: core.ResourceList{
			core.ResourceCPU: *resource.NewScaledQuantity(10, 1),
		},
	}
	containerPolicy2 := ContainerResourcePolicy{
		ContainerName: "container2",
		MaxAllowed: core.ResourceList{
			core.ResourceMemory: *resource.NewScaledQuantity(100, 1),
		},
	}
	policy := VerticalPodAutoscaler{
		Spec: VerticalPodAutoscalerSpec{
			ResourcePolicy: &PodResourcePolicy{
				ContainerPolicies: []ContainerResourcePolicy{
					containerPolicy1, containerPolicy2,
				},
			},
		},
	}
	assert.Equal(t, &containerPolicy1, GetContainerResourcePolicy("container1", &policy))
	assert.Equal(t, &containerPolicy2, GetContainerResourcePolicy("container2", &policy))
	assert.Nil(t, GetContainerResourcePolicy("container3", &policy))

	// Add the wildcard ("*") policy.
	defaultPolicy := ContainerResourcePolicy{
		ContainerName: "*",
		MinAllowed: core.ResourceList{
			core.ResourceCPU: *resource.NewScaledQuantity(20, 1),
		},
	}
	policy = VerticalPodAutoscaler{
		Spec: VerticalPodAutoscalerSpec{
			ResourcePolicy: &PodResourcePolicy{
				ContainerPolicies: []ContainerResourcePolicy{
					containerPolicy1, defaultPolicy, containerPolicy2,
				},
			},
		},
	}
	assert.Equal(t, &containerPolicy1, GetContainerResourcePolicy("container1", &policy))
	assert.Equal(t, &containerPolicy2, GetContainerResourcePolicy("container2", &policy))
	assert.Equal(t, &defaultPolicy, GetContainerResourcePolicy("container3", &policy))
}
