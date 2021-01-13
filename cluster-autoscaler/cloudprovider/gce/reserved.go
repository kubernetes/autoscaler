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
	klog "k8s.io/klog/v2"
)

// There should be no imports as it is used standalone in e2e tests

const (
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
)

// EvictionHard is the struct used to keep parsed values for eviction
type EvictionHard struct {
	MemoryEvictionQuantity        int64
	EphemeralStorageEvictionRatio float64
}

// CalculateKernelReserved computes how much memory Linux kernel will reserve.
// TODO(jkaniuk): account for crashkernel reservation on RHEL / CentOS
func CalculateKernelReserved(physicalMemory int64, os OperatingSystem) int64 {
	switch os {
	case OperatingSystemLinux:
		// Account for memory reserved by kernel
		reserved := int64(physicalMemory / kernelReservedRatio)
		reserved += kernelReservedMemory
		// Account for software IO TLB allocation if memory requires 64bit addressing
		if physicalMemory > swiotlbThresholdMemory {
			reserved += swiotlbReservedMemory
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

// CalculateOSReservedEphemeralStorage estimates how much ephemeral storage OS will reserve and eviction threshold
func CalculateOSReservedEphemeralStorage(diskSize int64, osDistribution OperatingSystemDistribution) int64 {
	switch osDistribution {
	case OperatingSystemDistributionCOS, OperatingSystemDistributionCOSContainerd:
		storage := int64(math.Ceil(0.015635*float64(diskSize))) + int64(math.Ceil(4.148*GiB)) // os partition estimation
		storage += int64(math.Min(100*MiB, math.Ceil(0.001*float64(diskSize))))               // over-provisioning buffer
		return storage
	case OperatingSystemDistributionUbuntu, OperatingSystemDistributionUbuntuContainerd:
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
