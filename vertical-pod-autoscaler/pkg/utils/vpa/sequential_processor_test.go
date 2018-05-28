/*
Copyright 2017 The Kubernetes Authors.

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

package api

import (
	"testing"

	"k8s.io/api/core/v1"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/poc.autoscaling.k8s.io/v1alpha1"

	"github.com/stretchr/testify/assert"
)

type fakeProcessor struct {
	message string
}

func (p *fakeProcessor) Apply(podRecommendation *vpa_types.RecommendedPodResources, policy *vpa_types.PodResourcePolicy, pod *v1.Pod) (*vpa_types.RecommendedPodResources, error) {
	result := podRecommendation.DeepCopy()
	result.ContainerRecommendations[0].Name += p.message
	return result, nil
}

func TestSequentialProcessor(t *testing.T) {
	name1 := "processor1"
	name2 := "processor2"
	tested := NewSequentialProcessor([]RecommendationProcessor{&fakeProcessor{name1}, &fakeProcessor{name2}})

	rec1 := &vpa_types.RecommendedPodResources{
		ContainerRecommendations: []vpa_types.RecommendedContainerResources{
			{
				Name: "",
			},
		}}
	result, _ := tested.Apply(rec1, nil, nil)
	assert.Equal(t, name1+name2, result.ContainerRecommendations[0].Name)
}
