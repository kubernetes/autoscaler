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

// MigOsInfo store os parameters.
type MigOsInfo interface {
	// Os return operating system.
	Os() OperatingSystem
	// OsDistribution return operating system distribution.
	OsDistribution() OperatingSystemDistribution
	// Arch return system architecture.
	Arch() SystemArchitecture
}

// OsReservedCalculator calculates the OS reserved values.
type OsReservedCalculator interface {
	// CalculateKernelReserved computes how much memory OS kernel will reserve.
	// NodeVersion parameter is optional. If empty string is passed a result calculated using default node version will be returned.
	CalculateKernelReserved(m MigOsInfo, physicalMemory int64) int64

	// CalculateOSReservedEphemeralStorage estimates how much ephemeral storage OS will reserve and eviction threshold.
	// NodeVersion parameter is optional. If empty string is passed a result calculated using default node version will be returned.
	CalculateOSReservedEphemeralStorage(m MigOsInfo, diskSize int64) int64
}
