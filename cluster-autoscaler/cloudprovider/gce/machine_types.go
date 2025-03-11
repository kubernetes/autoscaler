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
	"fmt"
	"strconv"
	"strings"

	"k8s.io/autoscaler/cluster-autoscaler/utils/units"

	gce_api "google.golang.org/api/compute/v1"
)

// MachineType represents information about underlying GCE machine used by one
// or more MIGs.
type MachineType struct {
	// Name is the name of the particular MachineType (ex. e2-standard-4)
	Name string
	// CPU is the number of cores the machine has.
	CPU int64
	// Memory is the memory capacity of the machine in bytes.
	Memory int64
	// MaxDisk is the maximum total persistent disks size (GB) allowed.
	MaxDiskSizeGb int64
}

// IsCustomMachine checks if a machine type is custom or predefined.
func IsCustomMachine(machineType string) bool {
	// Custom types are in the form "<family>-custom-<num_cpu>-<memory_mb>", with the "<family>-" prefix being optional for the N1
	// family. Examples: n2-custom-48-184320, custom-48-184320 (equivalent to n1-custom-48-184320).
	// Docs: https://cloud.google.com/compute/docs/instances/creating-instance-with-custom-machine-type#gcloud.
	parts := strings.Split(machineType, "-")
	if len(parts) < 2 {
		return false
	}
	return parts[0] == "custom" || parts[1] == "custom"
}

// GetMachineFamily gets the machine family from the machine type. Machine family is determined as everything before the first
// dash character, unless this first part is "custom", which actually means "n1":
// https://cloud.google.com/compute/docs/instances/creating-instance-with-custom-machine-type#gcloud.
func GetMachineFamily(machineType string) (string, error) {
	parts := strings.Split(machineType, "-")
	if len(parts) < 2 {
		return "", fmt.Errorf("unable to parse machine type %q", machineType)
	}
	if parts[0] == "custom" || parts[0] == "f1" || parts[0] == "g1" {
		return "n1", nil
	}
	return parts[0], nil
}

// NewMachineTypeFromAPI creates a MachineType object based on machine type representation
// from GCE API client.
func NewMachineTypeFromAPI(name string, mt *gce_api.MachineType) (MachineType, error) {
	if mt == nil {
		return MachineType{}, fmt.Errorf("Failed to create MachineType %s from empty API object", name)
	}
	return MachineType{
		Name:          name,
		CPU:           mt.GuestCpus,
		Memory:        mt.MemoryMb * units.MiB,
		MaxDiskSizeGb: mt.MaximumPersistentDisksSizeGb,
	}, nil
}

// NewCustomMachineType creates a MachineType object describing a custom GCE machine.
// CPU and Memory are based on parsing custom machine name.
func NewCustomMachineType(name string) (MachineType, error) {
	if !IsCustomMachine(name) {
		return MachineType{}, fmt.Errorf("%q is not a valid custom machine type", name)
	}

	// Identify the "custom" part of the name, assume the next part is the CPU count, and the one after that is the memory amount.
	// This should work if the type name has a "custom-*-*" infix, regardless of the rest of the name.
	parts := strings.Split(name, "-")
	customPartIndex := -1
	for i, part := range parts {
		if part == "custom" {
			customPartIndex = i
			break
		}
	}
	if customPartIndex == -1 || customPartIndex+2 >= len(parts) {
		return MachineType{}, fmt.Errorf("unable to parse custom machine type %q", name)
	}
	cpuPart := parts[customPartIndex+1]
	memPart := parts[customPartIndex+2]

	cpu, err := parseCustomCpu(name, cpuPart)
	if err != nil {
		return MachineType{}, fmt.Errorf("unable to parse CPU %q from machine type %q: %v", cpuPart, name, err)
	}
	memBytes, err := strconv.ParseInt(memPart, 10, 64)
	if err != nil {
		return MachineType{}, fmt.Errorf("unable to parse memory %q from machine type %q: %v", memPart, name, err)
	}

	return MachineType{
		Name:   name,
		CPU:    cpu,
		Memory: memBytes * units.MiB,
	}, nil
}

func parseCustomCpu(machineType string, cpuPart string) (int64, error) {
	// We need to identify the family because some e2 machine types have special names.
	family, err := GetMachineFamily(machineType)
	if err != nil {
		return 0, fmt.Errorf("unable to get family while parsing custom machine type %q: %v", machineType, err)
	}

	// There are e2-custom-micro-*, e2-custom-small-*, e2-custom-medium-* custom machine types in the e2 family. They all have 2 guestCpus
	// in the API.
	if family == "e2" && (cpuPart == "micro" || cpuPart == "small" || cpuPart == "medium") {
		return 2, nil
	}

	return strconv.ParseInt(cpuPart, 10, 64)
}
