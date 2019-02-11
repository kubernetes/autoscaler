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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateKernelReserved(t *testing.T) {
	type testCase struct {
		physicalMemory int64
		reservedMemory int64
	}
	testCases := []testCase{
		{
			physicalMemory: 256 * MiB,
			reservedMemory: 4*MiB + kernelReservedMemory,
		},
		{
			physicalMemory: 2 * GiB,
			reservedMemory: 32*MiB + kernelReservedMemory,
		},
		{
			physicalMemory: 3 * GiB,
			reservedMemory: 48*MiB + kernelReservedMemory,
		},
		{
			physicalMemory: 3.25 * GiB,
			reservedMemory: 52*MiB + kernelReservedMemory + swiotlbReservedMemory,
		},
		{
			physicalMemory: 4 * GiB,
			reservedMemory: 64*MiB + kernelReservedMemory + swiotlbReservedMemory,
		},
		{
			physicalMemory: 128 * GiB,
			reservedMemory: 2*GiB + kernelReservedMemory + swiotlbReservedMemory,
		},
	}
	for idx, tc := range testCases {
		t.Run(fmt.Sprintf("%v", idx), func(t *testing.T) {
			reserved := CalculateKernelReserved(tc.physicalMemory)
			assert.Equal(t, tc.reservedMemory, reserved)
		})
	}
}
