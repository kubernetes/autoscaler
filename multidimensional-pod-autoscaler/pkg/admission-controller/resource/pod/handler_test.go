/*
Copyright 2019 The Kubernetes Authors.

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

package pod

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	admissionv1 "k8s.io/api/admission/v1"
	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/admission-controller/resource/pod/patch"
	mpa_types "k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1alpha1"
	"k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/utils/test"
	resource_admission "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource"
)

type fakePodPreProcessor struct {
	err error
}

func (fpp *fakePodPreProcessor) Process(pod apiv1.Pod) (apiv1.Pod, error) {
	return pod, fpp.err
}

type fakeMpaMatcher struct {
	mpa *mpa_types.MultidimPodAutoscaler
}

func (m *fakeMpaMatcher) GetMatchingMPA(_ *apiv1.Pod) *mpa_types.MultidimPodAutoscaler {
	return m.mpa
}

type fakePatchCalculator struct {
	patches []resource_admission.PatchRecord
	err     error
}

func (c *fakePatchCalculator) CalculatePatches(_ *apiv1.Pod, _ *mpa_types.MultidimPodAutoscaler) (
	[]resource_admission.PatchRecord, error) {
	return c.patches, c.err
}

func TestGetPatches(t *testing.T) {
	testMpa := test.MultidimPodAutoscaler().WithName("name").WithContainer("testy-container").Get()
	testPatchRecord := resource_admission.PatchRecord{
		Op:    "add",
		Path:  "some/path",
		Value: "much",
	}
	testPatchRecord2 := resource_admission.PatchRecord{
		Op:    "add",
		Path:  "other/path",
		Value: "not so much",
	}
	tests := []struct {
		name                 string
		podJson              []byte
		namespace            string
		mpa                  *mpa_types.MultidimPodAutoscaler
		podPreProcessorError error
		calculators          []patch.Calculator
		expectPatches        []resource_admission.PatchRecord
		expectError          error
	}{
		{
			name:                 "invalid JSON",
			podJson:              []byte("{"),
			namespace:            "default",
			mpa:                  testMpa,
			podPreProcessorError: nil,
			expectError:          fmt.Errorf("unexpected end of JSON input"),
		},
		{
			name:                 "invalid pod",
			podJson:              []byte("{}"),
			namespace:            "default",
			mpa:                  testMpa,
			podPreProcessorError: fmt.Errorf("bad pod"),
			expectError:          fmt.Errorf("bad pod"),
		},
		{
			name:                 "no vpa found",
			podJson:              []byte("{}"),
			namespace:            "test",
			mpa:                  nil,
			podPreProcessorError: nil,
			expectError:          nil,
			expectPatches:        []resource_admission.PatchRecord{},
		},
		{
			name:      "calculator returns error",
			podJson:   []byte("{}"),
			namespace: "test",
			mpa:       testMpa,
			calculators: []patch.Calculator{&fakePatchCalculator{
				[]resource_admission.PatchRecord{}, fmt.Errorf("Can't calculate this"),
			}},
			podPreProcessorError: nil,
			expectError:          fmt.Errorf("Can't calculate this"),
			expectPatches:        []resource_admission.PatchRecord{},
		},
		{
			name:      "second calculator returns error",
			podJson:   []byte("{}"),
			namespace: "test",
			mpa:       testMpa,
			calculators: []patch.Calculator{
				&fakePatchCalculator{[]resource_admission.PatchRecord{
					testPatchRecord,
				}, nil},
				&fakePatchCalculator{
					[]resource_admission.PatchRecord{}, fmt.Errorf("Can't calculate this"),
				}},
			podPreProcessorError: nil,
			expectError:          fmt.Errorf("Can't calculate this"),
			expectPatches:        []resource_admission.PatchRecord{},
		},
		{
			name:      "patches returned correctly",
			podJson:   []byte("{}"),
			namespace: "test",
			mpa:       testMpa,
			calculators: []patch.Calculator{
				&fakePatchCalculator{[]resource_admission.PatchRecord{
					testPatchRecord,
					testPatchRecord2,
				}, nil}},
			podPreProcessorError: nil,
			expectError:          nil,
			expectPatches: []resource_admission.PatchRecord{
				patch.GetAddEmptyAnnotationsPatch(),
				testPatchRecord,
				testPatchRecord2,
			},
		},
		{
			name:      "patches returned correctly for multiple calculators",
			podJson:   []byte("{}"),
			namespace: "test",
			mpa:       testMpa,
			calculators: []patch.Calculator{
				&fakePatchCalculator{[]resource_admission.PatchRecord{
					testPatchRecord,
				}, nil},
				&fakePatchCalculator{[]resource_admission.PatchRecord{
					testPatchRecord2,
				}, nil}},
			podPreProcessorError: nil,
			expectError:          nil,
			expectPatches: []resource_admission.PatchRecord{
				patch.GetAddEmptyAnnotationsPatch(),
				testPatchRecord,
				testPatchRecord2,
			},
		},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("test case: %s", tc.name), func(t *testing.T) {
			fppp := &fakePodPreProcessor{tc.podPreProcessorError}
			fvm := &fakeMpaMatcher{mpa: tc.mpa}
			h := NewResourceHandler(fppp, fvm, tc.calculators)
			patches, err := h.GetPatches(&admissionv1.AdmissionRequest{
				Resource: v1.GroupVersionResource{
					Version: "v1",
				},
				Namespace: tc.namespace,
				Object: runtime.RawExtension{
					Raw: tc.podJson,
				},
			})
			if tc.expectError == nil {
				assert.NoError(t, err)
			} else {
				if assert.Error(t, err) {
					assert.Equal(t, tc.expectError.Error(), err.Error())
				}
			}
			if assert.Equal(t, len(tc.expectPatches), len(patches), fmt.Sprintf("got %+v, want %+v", patches, tc.expectPatches)) {
				for i, gotPatch := range patches {
					if !patch.EqPatch(gotPatch, tc.expectPatches[i]) {
						t.Errorf("Expected patch at position %d to be %+v, got %+v", i, tc.expectPatches[i], gotPatch)
					}
				}
			}
		})
	}
}
