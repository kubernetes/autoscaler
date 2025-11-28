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

package dra

import (
	"fmt"
	"strings"
)

const (
	// DriverNvidiaGPU is the standard driver name for Nvidia GPUs.
	DriverNvidiaGPUName = "gpu.nvidia.com"
	// DriverNvidiaGPUUid is the standard device UID attribute for Nvidia GPUs, in order to differentiate between different GPUs.
	DriverNvidiaGPUUid = "uuid"
	// DriverNvidiaGPUProductName is the standard device product name attribute for Nvidia GPUs, this will be the field with which we will filter resources by.
	DriverNvidiaGPUProductName = "productName"
)

// GetDraResourceName returns the standardized resource name for a DRA resource given its driver and product name.
func GetDraResourceName(driver string, productName string) string {
	return fmt.Sprintf("%s:%s", strings.ToLower(driver), strings.ToLower(productName))
}
