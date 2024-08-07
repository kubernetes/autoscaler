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
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	admissionv1 "k8s.io/api/admission/v1"
	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	resource_admission "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource/pod/patch"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
)

type fakePodPreProcessor struct {
	err error
}

func (fpp *fakePodPreProcessor) Process(pod apiv1.Pod) (apiv1.Pod, error) {
	return pod, fpp.err
}

type fakeVpaMatcher struct {
	vpa *vpa_types.VerticalPodAutoscaler
}

func (m *fakeVpaMatcher) GetMatchingVPA(_ context.Context, _ *apiv1.Pod) *vpa_types.VerticalPodAutoscaler {
	return m.vpa
}

type fakePatchCalculator struct {
	patches []resource_admission.PatchRecord
	err     error
}

func (c *fakePatchCalculator) CalculatePatches(_ *apiv1.Pod, _ *vpa_types.VerticalPodAutoscaler) (
	[]resource_admission.PatchRecord, error) {
	return c.patches, c.err
}

func TestGetPatches(t *testing.T) {
	testVpa := test.VerticalPodAutoscaler().WithName("name").WithContainer("testy-container").Get()
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
		vpa                  *vpa_types.VerticalPodAutoscaler
		podPreProcessorError error
		calculators          []patch.Calculator
		expectPatches        []resource_admission.PatchRecord
		expectError          error
	}{
		{
			name:                 "invalid JSON",
			podJson:              []byte("{"),
			namespace:            "default",
			vpa:                  testVpa,
			podPreProcessorError: nil,
			expectError:          fmt.Errorf("unexpected end of JSON input"),
		},
		{
			name:                 "invalid pod",
			podJson:              []byte("{}"),
			namespace:            "default",
			vpa:                  testVpa,
			podPreProcessorError: fmt.Errorf("bad pod"),
			expectError:          fmt.Errorf("bad pod"),
		},
		{
			name:                 "no vpa found",
			podJson:              []byte("{}"),
			namespace:            "test",
			vpa:                  nil,
			podPreProcessorError: nil,
			expectError:          nil,
			expectPatches:        []resource_admission.PatchRecord{},
		},
		{
			name:      "calculator returns error",
			podJson:   []byte("{}"),
			namespace: "test",
			vpa:       testVpa,
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
			vpa:       testVpa,
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
			vpa:       testVpa,
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
			vpa:       testVpa,
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
			fvm := &fakeVpaMatcher{vpa: tc.vpa}
			h := NewResourceHandler(fppp, fvm, tc.calculators)
			patches, err := h.GetPatches(context.Background(), &admissionv1.AdmissionRequest{
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
