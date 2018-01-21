/*
Copyright 2017 The Kubernetes Authors.

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

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCPUAmountFromCores(t *testing.T) {
	// Given
	cores := float64(1.333)
	// When
	resourceAmount := CPUAmountFromCores(cores)
	// Then
	assert.Equal(t, ResourceAmount(1333), resourceAmount)
}

func TestCoresFromCPUAmount(t *testing.T) {
	// Given
	cpuAmount := ResourceAmount(1333)
	// When
	cores := CoresFromCPUAmount(cpuAmount)
	// Then
	assert.Equal(t, float64(1.333), cores)
}

func TestMemoryAmountFromBytes(t *testing.T) {
	// Given
	bytes := float64(1024)
	// When
	memoryAmount := MemoryAmountFromBytes(bytes)
	// Then
	assert.Equal(t, ResourceAmount(1024), memoryAmount)
}

func TestBytesFromMemoryAmount(t *testing.T) {
	// Given
	memoryAmount := ResourceAmount(1024)
	// When
	bytes := BytesFromMemoryAmount(memoryAmount)
	// Then
	assert.Equal(t, float64(1024), bytes)
}

func TestUsageFromResourceAmount(t *testing.T) {
	// Given
	memoryAmount, memoryName := ResourceAmount(1024), ResourceMemory
	cpuAmount, cpuName := ResourceAmount(1333), ResourceCPU
	// When
	memoryUsage, memoryError := UsageFromResourceAmount(memoryName, memoryAmount)
	cpuUsage, cpuError := UsageFromResourceAmount(cpuName, cpuAmount)
	// Then
	assert.NoError(t, memoryError)
	assert.NoError(t, cpuError)
	assert.Equal(t, BytesFromMemoryAmount(memoryAmount), memoryUsage)
	assert.Equal(t, CoresFromCPUAmount(cpuAmount), cpuUsage)
}
