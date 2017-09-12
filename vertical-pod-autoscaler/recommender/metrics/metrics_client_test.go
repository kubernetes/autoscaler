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
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEmptyContainersUtilization(t *testing.T) {
	tc := newEmptyMetricsClientTestCase()
	metricsClient := tc.createFakeMetricsClient()

	snapshots, err := metricsClient.GetContainersUtilization()

	assert.NoError(t, err)
	assert.Empty(t, snapshots, "should be empty for empty Client")
}

func TestGetEmptyNamespaces(t *testing.T) {
	tc := newEmptyMetricsClientTestCase()
	metricsClient := tc.createFakeMetricsClient().(*metricsClient)

	namespaces, err := metricsClient.getAllNamespaces()

	assert.NoError(t, err)
	assert.Empty(t, namespaces, "should be empty for empty NamespaceLister")
}

func TestGetEmptyContainersSpec(t *testing.T) {
	tc := newEmptyMetricsClientTestCase()
	metricsClient := tc.createFakeMetricsClient().(*metricsClient)

	specs, err := metricsClient.getContainersSpec()

	assert.NoError(t, err)
	assert.Empty(t, specs, "should be empty for empty PodLister")
}

func TestGetEmptyContainersMetrics(t *testing.T) {
	tc := newEmptyMetricsClientTestCase()
	metricsClient := tc.createFakeMetricsClient().(*metricsClient)

	containerMetricsSnapshots, err := metricsClient.getContainersMetrics()

	assert.NoError(t, err)
	assert.Empty(t, containerMetricsSnapshots, "should be empty for empty MetricsGetter")
}

func TestGetContainersUtilization(t *testing.T) {
	tc := newMetricsClientTestCase()
	metricsClient := tc.createFakeMetricsClient()

	snapshots, err := metricsClient.GetContainersUtilization()

	assert.NoError(t, err)
	assert.Len(t, snapshots, len(tc.getAllSnaps()), "It should return right number of snapshots")
	for _, expectedSnap := range tc.getAllSnaps() {
		contains := false
		for _, actualSnap := range snapshots {
			if reflect.DeepEqual(actualSnap, expectedSnap) {
				contains = true
			}

		}
		if !contains {
			assert.Fail(t, "xxx", "Expected snapshot: %+v not found in results: %+v", expectedSnap, snapshots)
		}
	}
}

func TestGetNamespaces(t *testing.T) {
	tc := newMetricsClientTestCase()
	metricsClient := tc.createFakeMetricsClient().(*metricsClient)

	namespaces, err := metricsClient.getAllNamespaces()

	assert.NoError(t, err)
	assert.Len(t, namespaces, len(tc.getFakeNamespaces()), "It should return right number of Namespaces")
}

func TestGetContainersSpec(t *testing.T) {
	tc := newMetricsClientTestCase()
	metricsClient := tc.createFakeMetricsClient().(*metricsClient)

	specs, err := metricsClient.getContainersSpec()

	assert.NoError(t, err)
	assert.Len(t, specs, len(tc.getAllSnaps()), "It should return right number of ContainerSpecs")
}

func TestGetContainersMetrics(t *testing.T) {
	tc := newMetricsClientTestCase()
	metricsClient := tc.createFakeMetricsClient().(*metricsClient)

	containerMetricsSnapshots, err := metricsClient.getContainersMetrics()

	assert.NoError(t, err)
	assert.Len(t, containerMetricsSnapshots, len(tc.getAllSnaps()), "It should return right number of containerMetricsSnapshots")
}
