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
	"encoding/json"
	"fmt"
	"testing"

	resource_admission "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource"

	"github.com/stretchr/testify/assert"
)

// EqPatch returns true if patches are equal by comparing their
// marshalling result.
func EqPatch(a, b resource_admission.PatchRecord) bool {
	aJson, aErr := json.Marshal(a)
	bJson, bErr := json.Marshal(b)
	return string(aJson) == string(bJson) && aErr == bErr
}

// AssertEqPatch asserts patches are equal.
func AssertEqPatch(t *testing.T, got, want resource_admission.PatchRecord) {
	assert.True(t, EqPatch(got, want), "got %+v, want: %+v", got, want)
}

// AssertPatchOneOf asserts patch is one of possible expected patches.
func AssertPatchOneOf(t *testing.T, got resource_admission.PatchRecord, want []resource_admission.PatchRecord) {
	for _, wanted := range want {
		if EqPatch(got, wanted) {
			return
		}
	}
	msg := fmt.Sprintf("got: %+v, expected one of %+v", got, want)
	assert.Fail(t, msg)
}
