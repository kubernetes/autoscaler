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
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog/v2"
)

// There should be no imports as it is used standalone in e2e tests

const (
	// KiB - KibiByte size (2^10)
	KiB = 1024
	// MiB - MebiByte size (2^20)
	MiB = 1024 * 1024
	// GiB - GibiByte size (2^30)
	GiB = 1024 * 1024 * 1024

	// MemoryEvictionHardTag tag passed by kubelet used to determine evictionHard values
	MemoryEvictionHardTag = "memory.available"
	// EphemeralStorageEvictionHardTag tag passed by kubelet used to determine evictionHard values
	EphemeralStorageEvictionHardTag = "nodefs.available"

	// defaultKubeletEvictionHardMemory is subtracted from capacity
	// when calculating allocatable (on top of kube-reserved).
	// Equals kubelet "evictionHard: {MemoryEvictionHardTag}"
	// It is hardcoded as a fallback when it is not passed by kubelet.
	defaultKubeletEvictionHardMemory = 100 * MiB

	// defaultKubeletEvictionHardEphemeralStorageRatio is the ratio of disk size to be blocked for eviction
	// subtracted from capacity when calculating allocatable (on top of kube-reserved).
	// Equals kubelet "evictionHard: {EphemeralStorageEvictionHardTag}"
	// It is hardcoded as a fallback when it is not passed by kubelet.
	defaultKubeletEvictionHardEphemeralStorageRatio = 0.1

	// Kernel reserved memory is subtracted when calculating total memory.
	kernelReservedRatio  = 64
	kernelReservedMemory = 16 * MiB
	// Reserved memory for software IO TLB
	swiotlbReservedMemory  = 64 * MiB
	swiotlbThresholdMemory = 3 * GiB

	// Memory Estimation Correction
	// correctionConstant is a linear constant for additional reserved memory
	correctionConstant = 0.00175
	// maximumCorrectionValue is the max-cap for additional reserved memory
	maximumCorrectionValue = 248 * MiB
	// ubuntuSpecificOffset is a constant value that is additionally added to Ubuntu
	// based distributions as reserved memory
	ubuntuSpecificOffset = 4 * MiB
	// lowMemoryOffset is an additional offset added for lower memory sized machines
	lowMemoryOffset = 8 * MiB
	// lowMemoryThreshold is the threshold to apply lowMemoryOffset
	lowMemoryThreshold = 8 * GiB
	// architectureOffsetRatio is the offset value used for arm architecture
	architectureOffsetRatio = 0.9971
)

// EvictionHard is the struct used to keep parsed values for eviction
type EvictionHard struct {
	MemoryEvictionQuantity        int64
	EphemeralStorageEvictionRatio float64
}

// GceReserved implement Reserved interface
type GceReserved struct{}

// CalculateKernelReserved computes how much memory Linux kernel will reserve.
// TODO(jkaniuk): account for crashkernel reservation on RHEL / CentOS
func (r *GceReserved) CalculateKernelReserved(m MigOsInfo, physicalMemory int64) int64 {
	os := m.Os()
	osDistribution := m.OsDistribution()
	arch := m.Arch()
	switch os {
	case OperatingSystemLinux:
		// Account for memory reserved by kernel
		reserved := int64(physicalMemory / kernelReservedRatio)
		reserved += kernelReservedMemory
		// Account for software IO TLB allocation if memory requires 64bit addressing
		if physicalMemory > swiotlbThresholdMemory {
			reserved += swiotlbReservedMemory
		}
		if physicalMemory <= lowMemoryThreshold {
			reserved += lowMemoryOffset
		}

		// Additional reserved memory to correct estimation
		// The reason for this value is we detected additional reservation, but we were
		// unable to find the root cause. Hence, we added a best estimated formula that was
		// statistically developed.
		if osDistribution == OperatingSystemDistributionCOS {
			reserved += int64(math.Min(correctionConstant*float64(physicalMemory), maximumCorrectionValue))
			if arch == Arm64 {
				reserved = int64(math.Round(float64(reserved) * architectureOffsetRatio))
			}
		} else if osDistribution == OperatingSystemDistributionUbuntu {
			reserved += int64(math.Min(correctionConstant*float64(physicalMemory), maximumCorrectionValue))
			reserved += ubuntuSpecificOffset
		}

		return reserved
	case OperatingSystemWindows:
		return 0
	default:
		klog.Errorf("CalculateKernelReserved called for unknown operating system %v", os)
		return 0
	}
}

// ParseEvictionHardOrGetDefault tries to parse evictionHard map, else fills defaults to be used.
func ParseEvictionHardOrGetDefault(evictionHard map[string]string) *EvictionHard {
	if evictionHard == nil {
		return &EvictionHard{
			MemoryEvictionQuantity:        defaultKubeletEvictionHardMemory,
			EphemeralStorageEvictionRatio: defaultKubeletEvictionHardEphemeralStorageRatio,
		}
	}
	evictionReturn := EvictionHard{}

	// CalculateOrDefault for Memory
	memory, found := evictionHard[MemoryEvictionHardTag]
	if !found {
		klog.V(4).Info("evictionHard memory tag not found, using default")
		evictionReturn.MemoryEvictionQuantity = defaultKubeletEvictionHardMemory
	} else {
		memQuantity, err := resource.ParseQuantity(memory)
		if err != nil {
			evictionReturn.MemoryEvictionQuantity = defaultKubeletEvictionHardMemory
		} else {
			value, possible := memQuantity.AsInt64()
			if !possible {
				klog.Errorf("unable to parse eviction ratio for memory: %v", err)
				evictionReturn.MemoryEvictionQuantity = defaultKubeletEvictionHardMemory
			} else {
				evictionReturn.MemoryEvictionQuantity = value
			}
		}
	}

	// CalculateOrDefault for Ephemeral Storage
	ephRatio, found := evictionHard[EphemeralStorageEvictionHardTag]
	if !found {
		klog.V(4).Info("evictionHard ephemeral storage tag not found, using default")
		evictionReturn.EphemeralStorageEvictionRatio = defaultKubeletEvictionHardEphemeralStorageRatio
	} else {
		value, err := parsePercentageToRatio(ephRatio)
		if err != nil {
			klog.Errorf("unable to parse eviction ratio for ephemeral storage: %v", err)
			evictionReturn.EphemeralStorageEvictionRatio = defaultKubeletEvictionHardEphemeralStorageRatio
		} else {
			if value < 0.0 || value > 1.0 {
				evictionReturn.EphemeralStorageEvictionRatio = defaultKubeletEvictionHardEphemeralStorageRatio
			} else {
				evictionReturn.EphemeralStorageEvictionRatio = value
			}
		}
	}

	return &evictionReturn
}

// GetKubeletEvictionHardForMemory calculates the evictionHard value for Memory.
func GetKubeletEvictionHardForMemory(evictionHard *EvictionHard) int64 {
	if evictionHard == nil {
		return defaultKubeletEvictionHardMemory
	}
	return evictionHard.MemoryEvictionQuantity
}

// GetKubeletEvictionHardForEphemeralStorage calculates the evictionHard value for Ephemeral Storage.
func GetKubeletEvictionHardForEphemeralStorage(diskSize int64, evictionHard *EvictionHard) float64 {
	if diskSize <= 0 {
		return 0
	}
	if evictionHard == nil {
		return defaultKubeletEvictionHardEphemeralStorageRatio * float64(diskSize)
	}
	return evictionHard.EphemeralStorageEvictionRatio * float64(diskSize)
}

func parsePercentageToRatio(percentString string) (float64, error) {
	i := strings.Index(percentString, "%")
	if i < 0 || i != len(percentString)-1 {
		return 0, fmt.Errorf("parsePercentageRatio: percentage sign not found")
	}
	percentVal, err := strconv.ParseFloat(percentString[:i], 64)
	if err != nil {
		return 0, err
	}
	return percentVal / 100, nil
}

// ephemeralStorageOnLocalSSDFilesystemOverheadInKiBByOSAndDiskCount was
// measured by creating 1-node nodepools in a GKE cluster with ephemeral
// storage on N local SSDs, measuring for each node
// N * 375GiB - .status.capacity["ephemeral-storage"]
var ephemeralStorageOnLocalSSDFilesystemOverheadInKiBByOSAndDiskCount = map[OperatingSystemDistribution]map[int64]int64{
	OperatingSystemDistributionCOS: {
		1:  7289472,
		2:  13725224,
		3:  20031312,
		4:  26332924,
		5:  32634536,
		6:  38946604,
		7:  45254008,
		8:  51556096,
		16: 52837800,
		24: 78686620,
	},
	OperatingSystemDistributionUbuntu: {
		1:  7219840,
		2:  13651496,
		3:  19953488,
		4:  26255100,
		5:  32556712,
		6:  38860588,
		7:  45163896,
		8:  51465984,
		16: 52747688,
		24: 78601704,
	},
}

// EphemeralStorageOnLocalSSDFilesystemOverheadInBytes estimates the difference
// between the total physical capacity of the local SSDs and the ephemeral
// storage filesystem capacity. It uses experimental values measured for all
// possible disk counts in GKE. Custom Kubernetes on GCE may allow intermediate
// counts, attaching the measured count, but not using it all for ephemeral
// storage. In that case, the difference in overhead between GKE and custom node
// images may be higher than the difference in overhead between two disk counts,
// so interpolating wouldn't make much sense. Instead, we use the next count for
// which we measured a filesystem overhead, which is a safer approximation
// (better to reserve more and not scale up than not enough and not schedule).
func EphemeralStorageOnLocalSSDFilesystemOverheadInBytes(diskCount int64, osDistribution OperatingSystemDistribution) int64 {
	var measuredCount int64
	if diskCount <= 8 {
		measuredCount = diskCount
	} else if diskCount <= 16 {
		measuredCount = 16
	} else {
		measuredCount = 24 // max attachable
	}

	o, ok := ephemeralStorageOnLocalSSDFilesystemOverheadInKiBByOSAndDiskCount[osDistribution]
	if !ok {
		klog.Errorf("Ephemeral storage backed by local SSDs is not supported for image family %v", osDistribution)
		return 0
	}
	return o[measuredCount] * KiB
}

// CalculateOSReservedEphemeralStorage estimates how much ephemeral storage OS will reserve and eviction threshold
func (r *GceReserved) CalculateOSReservedEphemeralStorage(m MigOsInfo, diskSize int64) int64 {
	osDistribution := m.OsDistribution()
	arch := m.Arch()
	switch osDistribution {
	case OperatingSystemDistributionCOS:
		storage := int64(math.Ceil(0.015635*float64(diskSize))) + int64(math.Ceil(4.148*GiB)) // os partition estimation
		storage += int64(math.Min(100*MiB, math.Ceil(0.001*float64(diskSize))))               // over-provisioning buffer
		if arch == Arm64 {
			storage += 65536 * KiB
		}
		return storage
	case OperatingSystemDistributionUbuntu:
		storage := int64(math.Ceil(0.03083*float64(diskSize))) + int64(math.Ceil(0.171*GiB)) // os partition estimation
		storage += int64(math.Min(100*MiB, math.Ceil(0.001*float64(diskSize))))              // over-provisioning buffer
		return storage
	case OperatingSystemDistributionWindowsLTSC, OperatingSystemDistributionWindowsSAC:
		storage := int64(math.Ceil(0.1133 * GiB)) // os partition estimation
		storage += int64(math.Ceil(0.010 * GiB))  // over-provisioning buffer
		return storage
	default:
		klog.Errorf("CalculateReservedAndEvictionEphemeralStorage called for unknown os distribution %v", osDistribution)
		return 0
	}
}

// GceMigOsInfo contains os details of nodes in gce mig.
type GceMigOsInfo struct {
	os             OperatingSystem
	osDistribution OperatingSystemDistribution
	arch           SystemArchitecture
}

// Os return operating system.
func (m *GceMigOsInfo) Os() OperatingSystem {
	return m.os
}

// OsDistribution return operating system distribution.
func (m *GceMigOsInfo) OsDistribution() OperatingSystemDistribution {
	return m.osDistribution
}

// Arch return system architecture.
func (m *GceMigOsInfo) Arch() SystemArchitecture {
	return m.arch
}

// NewMigOsInfo return gce implementation of MigOsInfo interface.
func NewMigOsInfo(os OperatingSystem, osDistribution OperatingSystemDistribution, arch SystemArchitecture) MigOsInfo {
	return &GceMigOsInfo{os, osDistribution, arch}
}
