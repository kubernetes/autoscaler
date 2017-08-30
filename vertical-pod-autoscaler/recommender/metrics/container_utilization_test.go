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
	"k8s.io/apimachinery/pkg/api/resource"
	clientapiv1 "k8s.io/client-go/pkg/api/v1"
	k8sapiv1 "k8s.io/kubernetes/pkg/api/v1"
)

type containerUtilizationTestCase struct {
	id1, id2 containerID
}

func newContainerUtilizationTestCase() *containerUtilizationTestCase {
	return &containerUtilizationTestCase{
		id1: containerID{"a", "b", "c"},
		id2: containerID{"a", "b", "cs"},
	}
}

func (tc *containerUtilizationTestCase) newContainerSpec() *containerSpec {
	return &containerSpec{ID: tc.id1}
}

func (tc *containerUtilizationTestCase) newMatchingSnap() *containerUsageSnapshot {
	return &containerUsageSnapshot{ID: tc.id1}
}

func (tc *containerUtilizationTestCase) newNonMatchingSnap() *containerUsageSnapshot {
	return &containerUsageSnapshot{ID: tc.id2}
}

func TestCreatingUtilizationSnapshotFromDifferentContainers(t *testing.T) {
	tc := newContainerUtilizationTestCase()

	_, err := NewContainerUtilizationSnapshot(tc.newNonMatchingSnap(), tc.newContainerSpec())

	assert.Error(t, err)
}
func TestCreatingUtilizationSnapshotFromSameContainer(t *testing.T) {
	tc := newContainerUtilizationTestCase()

	_, err := NewContainerUtilizationSnapshot(tc.newMatchingSnap(), tc.newContainerSpec())

	assert.NoError(t, err)
}

func TestTypeConversionToGoClient(t *testing.T) {
	quantity1 := *resource.NewMilliQuantity(1234, resource.BinarySI)
	quantity2 := *resource.NewQuantity(9, resource.DecimalSI)
	input := make(k8sapiv1.ResourceList, 2)
	input[k8sapiv1.ResourceCPU] = quantity1
	input[k8sapiv1.ResourceMemory] = quantity2

	output := convertResourceListToClientAPIType(input)

	assert.Equal(t, quantity1, output[clientapiv1.ResourceCPU])
	assert.Equal(t, quantity2, output[clientapiv1.ResourceMemory])
	assert.Len(t, output, 2)
}
