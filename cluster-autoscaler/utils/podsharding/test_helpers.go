/*
Copyright 2023 The Kubernetes Authors.

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

package podsharding

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apitypes "k8s.io/apimachinery/pkg/types"

	"github.com/stretchr/testify/assert"
)

func podForUID(uid apitypes.UID) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			UID: apitypes.UID(uid),
		},
	}
}

func assertNodeGroupDescriptorEqual(t *testing.T, expected, actual NodeGroupDescriptor) {
	t.Helper()
	if len(expected.Labels) != 0 || len(actual.Labels) != 0 {
		assert.Equal(t, expected.Labels, actual.Labels, "Labels")
	}
	if len(expected.SystemLabels) != 0 || len(actual.SystemLabels) != 0 {
		assert.Equal(t, expected.SystemLabels, actual.SystemLabels, "SystemLabels")
	}
	assert.ElementsMatch(t, expected.Taints, actual.Taints, "Taints")
	if len(expected.ExtraResources) != 0 || len(actual.ExtraResources) != 0 {
		assert.Equal(t, resourcesToIntMap(expected.ExtraResources), resourcesToIntMap(actual.ExtraResources), "ExtraResources")
	}
	assert.Equal(t, expected.ProvisioningClassName, actual.ProvisioningClassName, "ProvisioningClassName")
}

func resourcesToIntMap(quantityMap map[string]resource.Quantity) map[string]int64 {
	result := make(map[string]int64)
	for k, v := range quantityMap {
		valInt64, _ := v.AsInt64()
		result[k] = valInt64
	}
	return result
}
