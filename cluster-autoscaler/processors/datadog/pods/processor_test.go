/*
Copyright 2021 The Kubernetes Authors.

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

package pods

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	proc "k8s.io/autoscaler/cluster-autoscaler/processors/pods"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Test processors frontend by ensuring pipelines are properly wired and chained.
// Effective filters and transforms are tested individually in other files.

func TestFilteringPodListProcessor(t *testing.T) {
	fp := filteringPodListProcessor{
		transforms: []proc.PodListProcessor{
			&testTransformProcessor{
				suffix: "-renamed",
			},
			&testTransformProcessor{
				suffix: "-again",
			},
		},
		filters: []proc.PodListProcessor{
			&testFilterProcessor{
				filteredName: "p2-renamed-again",
			},
		},
	}

	ctx := &context.AutoscalingContext{}

	in := []*apiv1.Pod{newPod("p1"), newPod("p2")}
	expected := []*apiv1.Pod{newPod("p1-renamed-again")}

	actual, err := fp.Process(ctx, in)
	assert.NoError(t, err)
	assert.ElementsMatch(t, actual, expected, "filtered pods differ")
}

func newPod(name string) *apiv1.Pod {
	return &apiv1.Pod{ObjectMeta: metav1.ObjectMeta{Name: name}}
}

type testTransformProcessor struct {
	suffix string
}

func (p *testTransformProcessor) Process(ctx *context.AutoscalingContext, pending []*apiv1.Pod) ([]*apiv1.Pod, error) {
	for _, pod := range pending {
		pod.SetName(fmt.Sprintf("%s%s", pod.GetName(), p.suffix))
	}
	return pending, nil
}

func (p *testTransformProcessor) CleanUp() {}

type testFilterProcessor struct {
	filteredName string
}

func (p *testFilterProcessor) Process(ctx *context.AutoscalingContext, pending []*apiv1.Pod) ([]*apiv1.Pod, error) {
	var result []*apiv1.Pod
	for _, pod := range pending {
		if pod.GetName() != p.filteredName {
			result = append(result, pod)
		}
	}
	return result, nil
}

func (p *testFilterProcessor) CleanUp() {}
