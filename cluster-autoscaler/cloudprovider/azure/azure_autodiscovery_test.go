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

package azure

import (
	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"testing"
)

func TestParseLabelAutoDiscoverySpecs(t *testing.T) {
	testCases := []struct {
		name        string
		specs       []string
		expected    []labelAutoDiscoveryConfig
		expectedErr bool
	}{
		{
			name: "ValidSpec",
			specs: []string{
				"label:cluster-autoscaler-enabled=true,cluster-autoscaler-name=fake-cluster",
				"label:test-tag=test-value,another-test-tag=another-test-value",
			},
			expected: []labelAutoDiscoveryConfig{
				{Selector: map[string]string{"cluster-autoscaler-enabled": "true", "cluster-autoscaler-name": "fake-cluster"}},
				{Selector: map[string]string{"test-tag": "test-value", "another-test-tag": "another-test-value"}},
			},
		},
		{
			name:        "MissingAutoDiscoverLabel",
			specs:       []string{"test-tag=test-value,another-test-tag"},
			expectedErr: true,
		},
		{
			name:        "InvalidAutoDiscoerLabel",
			specs:       []string{"invalid:test-tag=test-value,another-test-tag"},
			expectedErr: true,
		},
		{
			name:        "MissingValue",
			specs:       []string{"label:test-tag="},
			expectedErr: true,
		},
		{
			name:        "MissingKey",
			specs:       []string{"label:=test-val"},
			expectedErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ngdo := cloudprovider.NodeGroupDiscoveryOptions{NodeGroupAutoDiscoverySpecs: tc.specs}
			actual, err := ParseLabelAutoDiscoverySpecs(ngdo)
			if tc.expectedErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.True(t, assert.ObjectsAreEqualValues(tc.expected, actual), "expected %#v, but found: %#v", tc.expected, actual)
		})
	}
}
