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

package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
)

func TestGetContainersMetricsReturnsEmptyList(t *testing.T) {
	tc := newEmptyMetricsClientTestCase()
	emptyMetricsClient := tc.createFakeMetricsClient()

	containerMetricsSnapshots, err := emptyMetricsClient.GetContainersMetrics(make(map[model.PodID]bool), true)

	assert.NoError(t, err)
	assert.Empty(t, containerMetricsSnapshots, "should be empty for empty MetricsGetter")
}

func TestGetContainersMetricsReturnsResults(t *testing.T) {
	tc, podList := newMetricsClientTestCase()
	fakeMetricsClient := tc.createFakeMetricsClient()

	snapshots, err := fakeMetricsClient.GetContainersMetrics(podList, true)

	assert.NoError(t, err)
	assert.Len(t, snapshots, len(tc.getAllSnaps()), "It should return right number of snapshots")
	for _, snap := range snapshots {
		assert.Contains(t, tc.getAllSnaps(), snap, "One of returned ContainerMetricsSnapshot is different then expected ")
	}
}
