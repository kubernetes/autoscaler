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

package inplace

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource/pod/patch"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
)

func TestCalculatePatches_ManagedLabel(t *testing.T) {
	c := NewManagedLabelCalculator()

	t.Run("pod with labels", func(t *testing.T) {
		patches, err := c.CalculatePatches(test.Pod().WithLabels(map[string]string{"app": "foo"}).Get(), nil)
		assert.NoError(t, err)
		if assert.Len(t, patches, 1, "Unexpected number of patches.") {
			patch.AssertEqPatch(t, patch.GetAddLabelPatch(patch.VpaManagedLabel, "true"), patches[0])
		}
	})

	t.Run("pod without labels", func(t *testing.T) {
		pod := test.Pod().Get()
		pod.Labels = nil
		patches, err := c.CalculatePatches(pod, nil)
		assert.NoError(t, err)
		if assert.Len(t, patches, 2, "Unexpected number of patches.") {
			patch.AssertEqPatch(t, patch.GetAddEmptyLabelsPatch(), patches[0])
			patch.AssertEqPatch(t, patch.GetAddLabelPatch(patch.VpaManagedLabel, "true"), patches[1])
		}
	})
}
