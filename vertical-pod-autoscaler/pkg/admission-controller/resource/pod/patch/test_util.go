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
	"testing"

	"github.com/stretchr/testify/assert"
	resource_admission "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource"
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
