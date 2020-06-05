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

import klog "k8s.io/klog/v2"

// There should be no imports as it is used standalone in e2e tests

const (
	// MiB - MebiByte size (2^20)
	MiB = 1024 * 1024
	// GiB - GibiByte size (2^30)
	GiB = 1024 * 1024 * 1024

	// KubeletEvictionHardMemory is subtracted from capacity
	// when calculating allocatable (on top of kube-reserved).
	// Equals kubelet "evictionHard: {memory.available}"
	// We don't have a good place to get it from, but it has been hard-coded
	// to 100Mi since at least k8s 1.4.
	KubeletEvictionHardMemory = 100 * MiB

	// Kernel reserved memory is subtracted when calculating total memory.
	kernelReservedRatio  = 64
	kernelReservedMemory = 16 * MiB
	// Reserved memory for software IO TLB
	swiotlbReservedMemory  = 64 * MiB
	swiotlbThresholdMemory = 3 * GiB
)

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
		klog.Errorf("CalculateKernelReserved called for unknown operatin system %v", os)
		return 0
	}
}
