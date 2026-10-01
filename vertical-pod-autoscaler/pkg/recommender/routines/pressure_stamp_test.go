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

package routines

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
)

func TestSetLastPressureTime(t *testing.T) {
	enabled, disabled := vpa_types.PressureDetectionEnabled, vpa_types.PressureDetectionDisabled
	t0 := time.Date(2026, 9, 29, 10, 0, 0, 0, time.UTC)
	older, newer := metav1.NewTime(t0.Add(-time.Minute)), metav1.NewTime(t0.Add(time.Minute))
	tests := []struct {
		name     string
		mode     *vpa_types.PressureDetectionMode
		observed *metav1.Time // stamp already in status
		accepted time.Time    // aggregate's last accepted pressure sample
		want     *metav1.Time
	}{
		{name: "new sample stamps", mode: &enabled, accepted: t0, want: ptr.To(metav1.NewTime(t0))},
		{name: "newer sample replaces an older stamp", mode: &enabled, observed: &older, accepted: t0, want: ptr.To(metav1.NewTime(t0))},
		{name: "older sample keeps the newer stamp", mode: &enabled, observed: &newer, accepted: t0, want: &newer},
		{name: "no sample carries the stamp forward", mode: &enabled, observed: &older, want: &older},
		{name: "opt-out clears the stamp", mode: &disabled, observed: &older, accepted: t0, want: nil},
		{name: "no policy clears the stamp", observed: &older, accepted: t0, want: nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			vpa := &model.Vpa{ResourcePolicy: &vpa_types.PodResourcePolicy{}}
			if tc.mode != nil {
				vpa.ResourcePolicy.ContainerPolicies = []vpa_types.ContainerResourcePolicy{{ContainerName: "c", PressureDetection: tc.mode}}
			}
			agg := model.NewAggregateContainerState()
			agg.LastPressureTime = tc.accepted
			observed := &vpa_types.VerticalPodAutoscaler{}
			observed.Status.Recommendation = &vpa_types.RecommendedPodResources{ContainerRecommendations: []vpa_types.RecommendedContainerResources{
				{ContainerName: "c", LastPressureTime: tc.observed}}}
			recs := &vpa_types.RecommendedPodResources{ContainerRecommendations: []vpa_types.RecommendedContainerResources{{ContainerName: "c"}}}
			setLastPressureTime(vpa, model.ContainerNameToAggregateStateMap{"c": agg}, observed, recs)
			got := recs.ContainerRecommendations[0].LastPressureTime
			if tc.want == nil {
				assert.Nil(t, got)
				return
			}
			if assert.NotNil(t, got) {
				assert.True(t, tc.want.Equal(got), "want %v got %v", tc.want, got)
			}
		})
	}
}
