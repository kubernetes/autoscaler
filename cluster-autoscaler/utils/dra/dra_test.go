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

package dra

import (
"testing"

"github.com/stretchr/testify/assert"
)

func TestGetDraResourceName(t *testing.T) {
	testCases := []struct {
		name           string
		driver         string
		productName    string
		expectedOutput string
	}{
		{
			name:           "nvidia GPU H100",
			driver:         "gpu.nvidia.com",
			productName:    "H100",
			expectedOutput: "gpu.nvidia.com:h100",
		},
		{
			name:           "nvidia GPU A100",
			driver:         "gpu.nvidia.com",
			productName:    "A100",
			expectedOutput: "gpu.nvidia.com:a100",
		},
		{
			name:           "nvidia GPU with mixed case",
			driver:         "GPU.NVIDIA.COM",
			productName:    "Tesla-K80",
			expectedOutput: "gpu.nvidia.com:tesla-k80",
		},
		{
			name:           "already lowercase",
			driver:         "gpu.nvidia.com",
			productName:    "a100",
			expectedOutput: "gpu.nvidia.com:a100",
		},
		{
			name:           "different driver",
			driver:         "custom.driver.io",
			productName:    "CustomDevice",
			expectedOutput: "custom.driver.io:customdevice",
		},
		{
			name:           "product name with spaces",
			driver:         "gpu.nvidia.com",
			productName:    "Tesla K80",
			expectedOutput: "gpu.nvidia.com:tesla k80",
		},
		{
			name:           "product name with special characters",
			driver:         "gpu.nvidia.com",
			productName:    "H100-PCIe",
			expectedOutput: "gpu.nvidia.com:h100-pcie",
		},
		{
			name:           "empty product name",
			driver:         "gpu.nvidia.com",
			productName:    "",
			expectedOutput: "gpu.nvidia.com:",
		},
		{
			name:           "empty driver",
			driver:         "",
			productName:    "H100",
			expectedOutput: ":h100",
		},
		{
			name:           "both empty",
			driver:         "",
			productName:    "",
			expectedOutput: ":",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
result := GetDraResourceName(tc.driver, tc.productName)
assert.Equal(t, tc.expectedOutput, result)
})
	}
}
