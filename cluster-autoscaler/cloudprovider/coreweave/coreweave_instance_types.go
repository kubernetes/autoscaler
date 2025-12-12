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

package coreweave

import "fmt"

// InstanceType represents the resource specifications for a CoreWeave instance type.
// Units are chosen to match the Kubernetes Node resource representation.
type InstanceType struct {
	// VCPU is the number of virtual CPU cores
	VCPU int64
	// MemoryKi is the amount of memory in kibibytes (1 Ki = 1024 bytes)
	MemoryKi int64
	// GPU is the number of GPUs
	GPU int64
	// EphemeralStorageKi is the amount of ephemeral storage in kibibytes (1 Ki = 1024 bytes)
	EphemeralStorageKi int64
	// Architecture is the CPU architecture (e.g., "amd64", "arm64")
	Architecture string
	// MaxPods is the maximum number of pods that can run on this instance type
	MaxPods int64
}

// InstanceTypes is a map of CoreWeave instance type names to their specifications.
// This map should be populated with the actual instance types supported by CoreWeave.
var InstanceTypes = map[string]*InstanceType{
	"b200-8x": {
		VCPU:               128,
		MemoryKi:           2112277172,
		GPU:                8,
		EphemeralStorageKi: 30003181568,
		Architecture:       "amd64",
		MaxPods:            110,
	},
	"cd-gp-a192-genoa": {
		VCPU:               192,
		MemoryKi:           1583811548,
		GPU:                0,
		EphemeralStorageKi: 7499230528,
		Architecture:       "amd64",
		MaxPods:            110,
	},
	"cd-gp-l-a192-genoa": {
		VCPU:               192,
		MemoryKi:           1583796048,
		GPU:                0,
		EphemeralStorageKi: 30003181568,
		Architecture:       "amd64",
		MaxPods:            110,
	},
	"cd-gp-i64-erapids": {
		VCPU:               64,
		MemoryKi:           526674536,
		GPU:                0,
		EphemeralStorageKi: 7499230528,
		Architecture:       "amd64",
		MaxPods:            110,
	},
	"cd-gp-l-i64-erapids": {
		VCPU:               64,
		MemoryKi:           526668108,
		GPU:                0,
		EphemeralStorageKi: 15000547328,
		Architecture:       "amd64",
		MaxPods:            110,
	},
	"cd-gp-i96-icelake": {
		VCPU:               96,
		MemoryKi:           394209340,
		GPU:                0,
		EphemeralStorageKi: 6248987968,
		Architecture:       "amd64",
		MaxPods:            110,
	},
	"cd-hc-a384ib-genoa": {
		VCPU:               384,
		MemoryKi:           1583672504,
		GPU:                0,
		EphemeralStorageKi: 30003181568,
		Architecture:       "amd64",
		MaxPods:            110,
	},
	"cd-hc-a384-genoa": {
		VCPU:               384,
		MemoryKi:           1583673336,
		GPU:                0,
		EphemeralStorageKi: 30003181568,
		Architecture:       "amd64",
		MaxPods:            300,
	},
	"cd-hp-a96-genoa": {
		VCPU:               96,
		MemoryKi:           791111968,
		GPU:                0,
		EphemeralStorageKi: 7499230528,
		Architecture:       "amd64",
		MaxPods:            110,
	},
	"gd-1xgh200": {
		VCPU:               72,
		MemoryKi:           600218240,
		GPU:                1,
		EphemeralStorageKi: 7499362648,
		Architecture:       "arm64",
		MaxPods:            110,
	},
	"gd-8xa100-i128": {
		VCPU:               128,
		MemoryKi:           2112249840,
		GPU:                8,
		EphemeralStorageKi: 7499362648,
		Architecture:       "amd64",
		MaxPods:            110,
	},
	"gd-8xh100ib-i128": {
		VCPU:               128,
		MemoryKi:           2112109804,
		GPU:                8,
		EphemeralStorageKi: 30003181568,
		Architecture:       "amd64",
		MaxPods:            110,
	},
	"gd-8xh200ib-i128": {
		VCPU:               128,
		MemoryKi:           2112109800,
		GPU:                8,
		EphemeralStorageKi: 30003181568,
		Architecture:       "amd64",
		MaxPods:            110,
	},
	"gd-8xl40-i128": {
		VCPU:               128,
		MemoryKi:           1055335508,
		GPU:                8,
		EphemeralStorageKi: 7499362648,
		Architecture:       "amd64",
		MaxPods:            110,
	},
	"gd-8xl40s-i128": {
		VCPU:               128,
		MemoryKi:           1055337468,
		GPU:                8,
		EphemeralStorageKi: 7499362648,
		Architecture:       "amd64",
		MaxPods:            110,
	},
	"rtxp6000-8x": {
		VCPU:               128,
		MemoryKi:           1055335468,
		GPU:                8,
		EphemeralStorageKi: 30003181568,
		Architecture:       "amd64",
		MaxPods:            110,
	},
	"turin-gp-l": {
		VCPU:               192,
		MemoryKi:           1583282436,
		GPU:                0,
		EphemeralStorageKi: 30003181568,
		Architecture:       "amd64",
		MaxPods:            110,
	},
	"turin-gp": {
		VCPU:               192,
		MemoryKi:           1583297960,
		GPU:                0,
		EphemeralStorageKi: 7499230528,
		Architecture:       "amd64",
		MaxPods:            110,
	},
}

// GetInstanceType returns the InstanceType for the given instance type name.
// It returns an error if the instance type is not found in the InstanceTypes map.
func GetInstanceType(instanceTypeName string) (*InstanceType, error) {
	instanceType, exists := InstanceTypes[instanceTypeName]
	if !exists {
		return nil, fmt.Errorf("unknown instance type: %s", instanceTypeName)
	}
	return instanceType, nil
}
