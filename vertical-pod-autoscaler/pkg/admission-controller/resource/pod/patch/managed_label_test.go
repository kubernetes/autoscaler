/*
Copyright 2025 The Kubernetes Authors.

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

package patch

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
)

func TestCalculatePatches_ManagedLabel(t *testing.T) {
	c := NewManagedLabelCalculator()
	patches, err := c.CalculatePatches(test.Pod().Get(), nil)
	assert.NoError(t, err)
	if assert.Len(t, patches, 1, "Unexpected number of patches.") {
		AssertEqPatch(t, GetAddLabelPatch(VpaManagedLabel, "true"), patches[0])
		assert.Equal(t, "/metadata/labels/vpa-autoscaler.k8s.io~1managed", patches[0].Path)
	}
}
