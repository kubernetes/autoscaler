/*
Copyright 2020 The Kubernetes Authors.

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
	"strings"
	"testing"

	core "k8s.io/api/core/v1"
	resource_admission "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/annotations"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"

	"github.com/stretchr/testify/assert"
)

func addVpaObservedContainersPatch(containerNames []string) resource_admission.PatchRecord {
	return GetAddAnnotationPatch(
		annotations.VpaObservedContainersLabel,
		strings.Join(containerNames, ", "),
	)
}

func TestCalculatePatches_ObservedContainers(t *testing.T) {
	tests := []struct {
		name          string
		pod           *core.Pod
		expectedPatch resource_admission.PatchRecord
	}{
		{
			name: "create vpa observed containers annotation",
			pod: test.Pod().AddContainer(test.Container().WithName("test1").Get()).
				AddContainer(test.Container().WithName("test2").Get()).Get(),
			expectedPatch: addVpaObservedContainersPatch([]string{"test1", "test2"}),
		},
		{
			name:          "create vpa observed containers annotation with no containers",
			pod:           test.Pod().Get(),
			expectedPatch: addVpaObservedContainersPatch([]string{}),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := NewObservedContainersCalculator()
			patches, err := c.CalculatePatches(tc.pod, nil)
			assert.NoError(t, err)
			if assert.Len(t, patches, 1, "Unexpected number of patches.") {
				AssertEqPatch(t, tc.expectedPatch, patches[0])
			}
		})
	}
}
