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
)

type containerUtilizationTestCase struct {
	id1, id2                      containerID
	containerSpec                 *basicContainerSpec
	matchingSnap, nonMatchingSnap *containerMetricsSnapshot
}

func newContainerUtilizationTestCase() *containerUtilizationTestCase {
	id1 := containerID{"a", "b", "c"}
	id2 := containerID{"a", "b", "cs"}

	return &containerUtilizationTestCase{
		id1:             id1,
		id2:             id2,
		containerSpec:   &basicContainerSpec{ID: id1},
		matchingSnap:    &containerMetricsSnapshot{ID: id1},
		nonMatchingSnap: &containerMetricsSnapshot{ID: id2},
	}
}

func TestCreatingUtilizationSnapshotFromDifferentContainers(t *testing.T) {
	tc := newContainerUtilizationTestCase()

	_, err := NewContainerUtilizationSnapshot(tc.nonMatchingSnap, tc.containerSpec)

	assert.Error(t, err)
}

func TestCreatingUtilizationSnapshotFromSameContainer(t *testing.T) {
	tc := newContainerUtilizationTestCase()

	_, err := NewContainerUtilizationSnapshot(tc.matchingSnap, tc.containerSpec)

	assert.NoError(t, err)
}
