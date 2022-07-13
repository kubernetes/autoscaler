/*
Copyright 2022 The Kubernetes Authors.

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

package gce

import (
	"testing"

	"k8s.io/autoscaler/cluster-autoscaler/utils/units"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
)

func TestNewCustomMachineType(t *testing.T) {
	testCases := []struct {
		name         string
		expectCustom bool
		expectCPU    int64
		expectMemory int64
	}{
		{
			name:         "custom-2-2816",
			expectCustom: true,
			expectCPU:    2,
			expectMemory: 2816 * units.MiB,
		},
		{
			name:         "n2-custom-2-2816",
			expectCustom: true,
			expectCPU:    2,
			expectMemory: 2816 * units.MiB,
		},
		{
			name: "other-a2-2816",
		},
		{
			name: "other-2-2816",
		},
		{
			name: "n1-standard-8",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectCustom, IsCustomMachine(tc.name))
			m, err := NewCustomMachineType(tc.name)
			if tc.expectCustom {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectCPU, m.CPU)
				assert.Equal(t, tc.expectMemory, m.Memory)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestGetMachineFamily(t *testing.T) {
	for tn, tc := range map[string]struct {
		machineType string
		wantFamily  string
		wantErr     error
	}{
		"predefined machine type": {
			machineType: "n1-standard-8",
			wantFamily:  "n1",
		},
		"predefined short machine type": {
			machineType: "e2-small",
			wantFamily:  "e2",
		},
		"custom machine type with family prefix": {
			machineType: "n2-custom-2-2816",
			wantFamily:  "n2",
		},
		"custom machine type without family prefix": {
			machineType: "custom-2-2816",
			wantFamily:  "n1",
		},
		"invalid machine type": {
			machineType: "nodashes",
			wantErr:     cmpopts.AnyError,
		},
	} {
		t.Run(tn, func(t *testing.T) {
			gotFamily, gotErr := GetMachineFamily(tc.machineType)
			if diff := cmp.Diff(tc.wantFamily, gotFamily); diff != "" {
				t.Errorf("GetMachineFamily(%q): diff (-want +got):\n%s", tc.machineType, diff)
			}
			if diff := cmp.Diff(tc.wantErr, gotErr, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("GetMachineFamily(%q): err diff (-want +got):\n%s", tc.machineType, diff)
			}
		})
	}
}
