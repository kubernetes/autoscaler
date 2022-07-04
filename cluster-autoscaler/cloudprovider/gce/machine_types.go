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
}

// IsCustomMachine checks if a machine type is custom or predefined.
func IsCustomMachine(name string) bool {
	return strings.HasPrefix(name, "custom-")
}

// NewMachineTypeFromAPI creates a MachineType object based on machine type representation
// from GCE API client.
func NewMachineTypeFromAPI(name string, mt *gce_api.MachineType) (MachineType, error) {
	if mt == nil {
		return MachineType{}, fmt.Errorf("Failed to create MachineType %s from empty API object", name)
	}
	return MachineType{
		Name:   name,
		CPU:    mt.GuestCpus,
		Memory: mt.MemoryMb * units.MiB,
	}, nil
}

// NewCustomMachineType creates a MachineType object describing a custom GCE machine.
// CPU and Memory are based on parsing custom machine name.
func NewCustomMachineType(name string) (MachineType, error) {
	// example custom-2-2816
	var cpu, mem int64
	var count int
	count, err := fmt.Sscanf(name, "custom-%d-%d", &cpu, &mem)
	if err != nil {
		return MachineType{}, err
	}
	if count != 2 {
		return MachineType{}, fmt.Errorf("failed to parse all params in %s", name)
	}
	mem = mem * units.MiB
	return MachineType{
		Name:   name,
		CPU:    cpu,
		Memory: mem,
	}, nil
}
