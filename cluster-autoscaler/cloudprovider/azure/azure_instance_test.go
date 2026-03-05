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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeArch(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "x64 to amd64",
			input:    "x64",
			expected: "amd64",
		},
		{
			name:     "X64 to amd64 (case insensitive)",
			input:    "X64",
			expected: "amd64",
		},
		{
			name:     "Arm64 to arm64",
			input:    "Arm64",
			expected: "arm64",
		},
		{
			name:     "ARM64 to arm64 (case insensitive)",
			input:    "ARM64",
			expected: "arm64",
		},
		{
			name:     "arm64 already normalized",
			input:    "arm64",
			expected: "arm64",
		},
		{
			name:     "unknown arch passed through",
			input:    "riscv64",
			expected: "riscv64",
		},
		{
			name:     "empty string passed through",
			input:    "",
			expected: "",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, NormalizeArch(tc.input))
		})
	}
}
