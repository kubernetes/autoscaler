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

package priority

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
)

type countingPodEvictionAdmission struct {
	admit         bool
	loopInitCalls int
	admitCalls    int
	cleanUpCalls  int
}

func (c *countingPodEvictionAdmission) LoopInit(allLivePods []*corev1.Pod, vpaControlledPods map[*vpa_types.VerticalPodAutoscaler][]*corev1.Pod) {
	c.loopInitCalls++
}

func (c *countingPodEvictionAdmission) Admit(pod *corev1.Pod, recommendation *vpa_types.RecommendedPodResources) bool {
	c.admitCalls++
	return c.admit
}

func (c *countingPodEvictionAdmission) CleanUp() {
	c.cleanUpCalls++
}

func TestNoopPodEvictionAdmissionAdmitsEverything(t *testing.T) {
	admission := NewDefaultPodEvictionAdmission()
	assert.True(t, admission.Admit(&corev1.Pod{}, nil))
}

func TestSequentialPodEvictionAdmissionAdmitsOnlyIfAllAdmit(t *testing.T) {
	cases := []struct {
		name   string
		admits []bool
		want   bool
	}{
		{"empty chain admits", nil, true},
		{"all admit", []bool{true, true, true}, true},
		{"one rejects", []bool{true, false, true}, false},
		{"first rejects", []bool{false, true, true}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var admissions []PodEvictionAdmission
			for _, admit := range tc.admits {
				admissions = append(admissions, &countingPodEvictionAdmission{admit: admit})
			}
			sequential := NewSequentialPodEvictionAdmission(admissions)
			got := sequential.Admit(&corev1.Pod{}, nil)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestSequentialPodEvictionAdmissionShortCircuitsOnFirstRejection(t *testing.T) {
	first := &countingPodEvictionAdmission{admit: false}
	second := &countingPodEvictionAdmission{admit: true}
	sequential := NewSequentialPodEvictionAdmission([]PodEvictionAdmission{first, second})

	got := sequential.Admit(&corev1.Pod{}, nil)

	assert.False(t, got)
	assert.Equal(t, 1, first.admitCalls)
	assert.Equal(t, 0, second.admitCalls, "should not call Admit on subsequent admissions once one rejects")
}

func TestSequentialPodEvictionAdmissionPropagatesLoopInitAndCleanUp(t *testing.T) {
	first := &countingPodEvictionAdmission{admit: true}
	second := &countingPodEvictionAdmission{admit: true}
	sequential := NewSequentialPodEvictionAdmission([]PodEvictionAdmission{first, second})

	sequential.LoopInit(nil, nil)
	sequential.CleanUp()

	assert.Equal(t, 1, first.loopInitCalls)
	assert.Equal(t, 1, second.loopInitCalls)
	assert.Equal(t, 1, first.cleanUpCalls)
	assert.Equal(t, 1, second.cleanUpCalls)
}
