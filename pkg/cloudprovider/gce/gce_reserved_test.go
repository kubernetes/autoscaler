/*
Copyright 2016 The Kubernetes Authors.

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
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateKernelReservedLinux(t *testing.T) {
	type testCase struct {
		physicalMemory int64
		reservedMemory int64
		osDistribution OperatingSystemDistribution
		arch           SystemArchitecture
	}
	testCases := []testCase{
		{
			physicalMemory: 256 * MiB,
			reservedMemory: 4*MiB + kernelReservedMemory + lowMemoryOffset,
			osDistribution: OperatingSystemDistributionCOS,
		},
		{
			physicalMemory: 2 * GiB,
			reservedMemory: 32*MiB + kernelReservedMemory + lowMemoryOffset,
			osDistribution: OperatingSystemDistributionCOS,
		},
		{
			physicalMemory: 3 * GiB,
			reservedMemory: 48*MiB + kernelReservedMemory + lowMemoryOffset,
			osDistribution: OperatingSystemDistributionCOS,
		},
		{
			physicalMemory: 3.25 * GiB,
			reservedMemory: 52*MiB + kernelReservedMemory + swiotlbReservedMemory + lowMemoryOffset,
			osDistribution: OperatingSystemDistributionCOS,
		},
		{
			physicalMemory: 4 * GiB,
			reservedMemory: 64*MiB + kernelReservedMemory + swiotlbReservedMemory + lowMemoryOffset,
			osDistribution: OperatingSystemDistributionCOS,
		},
		{
			physicalMemory: 128 * GiB,
			reservedMemory: 2*GiB + kernelReservedMemory + swiotlbReservedMemory,
			osDistribution: OperatingSystemDistributionUbuntu,
		},
		{
			physicalMemory: 256 * MiB,
			reservedMemory: 4*MiB + kernelReservedMemory + lowMemoryOffset,
			osDistribution: OperatingSystemDistributionUbuntu,
		},
		{
			physicalMemory: 2 * GiB,
			reservedMemory: 32*MiB + kernelReservedMemory + lowMemoryOffset,
			osDistribution: OperatingSystemDistributionUbuntu,
		},
		{
			physicalMemory: 3 * GiB,
			reservedMemory: 48*MiB + kernelReservedMemory + lowMemoryOffset,
			osDistribution: OperatingSystemDistributionUbuntu,
		},
		{
			physicalMemory: 3.25 * GiB,
			reservedMemory: 52*MiB + kernelReservedMemory + swiotlbReservedMemory + lowMemoryOffset,
			osDistribution: OperatingSystemDistributionUbuntu,
		},
		{
			physicalMemory: 4 * GiB,
			reservedMemory: 64*MiB + kernelReservedMemory + swiotlbReservedMemory + lowMemoryOffset,
			osDistribution: OperatingSystemDistributionUbuntu,
		},
		{
			physicalMemory: 128 * GiB,
			reservedMemory: 2*GiB + kernelReservedMemory + swiotlbReservedMemory,
			osDistribution: OperatingSystemDistributionUbuntu,
		},
		{
			physicalMemory: 3 * GiB,
			reservedMemory: 48*MiB + kernelReservedMemory + lowMemoryOffset,
			osDistribution: OperatingSystemDistributionCOS,
			arch:           Arm64,
		},
		{
			physicalMemory: 3.25 * GiB,
			reservedMemory: 52*MiB + kernelReservedMemory + swiotlbReservedMemory + lowMemoryOffset,
			osDistribution: OperatingSystemDistributionCOS,
			arch:           Arm64,
		},
	}
	for idx, tc := range testCases {
		r := &GceReserved{}
		t.Run(fmt.Sprintf("%v", idx), func(t *testing.T) {
			m := NewMigOsInfo(OperatingSystemLinux, tc.osDistribution, tc.arch)
			reserved := r.CalculateKernelReserved(m, tc.physicalMemory)
			if tc.osDistribution == OperatingSystemDistributionUbuntu {
				assert.Equal(t, tc.reservedMemory+int64(math.Min(correctionConstant*float64(tc.physicalMemory), maximumCorrectionValue)+ubuntuSpecificOffset), reserved)
			} else if tc.osDistribution == OperatingSystemDistributionCOS {
				val := tc.reservedMemory + int64(math.Min(correctionConstant*float64(tc.physicalMemory), maximumCorrectionValue))
				if tc.arch == Arm64 {
					val = int64(math.Round(float64(val) * architectureOffsetRatio))
				}
				assert.Equal(t, val, reserved)
			}
		})
	}
}

func TestEphemeralStorageOnLocalSSDFilesystemOverheadInBytes(t *testing.T) {
	type testCase struct {
		scenario       string
		diskCount      int64
		osDistribution OperatingSystemDistribution
		expected       int64
	}
	testCases := []testCase{
		{
			scenario:       "measured disk count and OS (cos)",
			diskCount:      1,
			osDistribution: OperatingSystemDistributionCOS,
			expected:       7289472 * KiB,
		},
		{
			scenario:       "measured disk count and OS (ubuntu)",
			diskCount:      1,
			osDistribution: OperatingSystemDistributionUbuntu,
			expected:       7219840 * KiB,
		},
		{
			scenario:       "mapped disk count",
			diskCount:      10,
			osDistribution: OperatingSystemDistributionCOS,
			expected:       52837800 * KiB, // value measured for 16 disks
		},
	}
	for _, tc := range testCases {
		t.Run(tc.scenario, func(t *testing.T) {
			actual := EphemeralStorageOnLocalSSDFilesystemOverheadInBytes(tc.diskCount, tc.osDistribution)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
