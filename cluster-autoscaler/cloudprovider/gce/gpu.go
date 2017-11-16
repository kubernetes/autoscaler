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

package gce

import (
	"strconv"
	"strings"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

var (
	// TODO(maciekpytel): get this from API
	gpuZones = map[string]map[string]bool{
		"nvidia-tesla-k80": {
			"europe-west1-b": true,
			"europe-west1-d": true,
			"asia-east1-a":   true,
			"asia-east1-b":   true,
			"us-east1-c":     true,
			"us-east1-d":     true,
			"us-west1-b":     true,
		},
		"nvidia-tesla-p100": {
			"europe-west1-b": true,
			"europe-west1-d": true,
			"asia-east1-a":   true,
			"us-east1-c":     true,
			"us-west1-b":     true,
		},
	}

	maxGpuCount = map[string]int64{
		"nvidia-tesla-k80":  8,
		"nvidia-tesla-p100": 4,
	}

	maxCpuCount = map[string]map[int64]int{
		"nvidia-tesla-k80": {
			1: 8,
			2: 16,
			4: 32,
			8: 64,
		},
		"nvidia-tesla-p100": {
			1: 16,
			2: 32,
			4: 64,
		},
	}
)

func validateGpuConfig(gpuType string, gpuCount int64, zone, machineType string) error {
	zoneInfo, found := gpuZones[gpuType]
	if !found {
		return cloudprovider.ErrIllegalConfiguration
	}
	if allowed := zoneInfo[zone]; !allowed {
		return cloudprovider.ErrIllegalConfiguration
	}

	maxGpu, found := maxGpuCount[gpuType]
	if !found || gpuCount > maxGpu {
		return cloudprovider.ErrIllegalConfiguration
	}

	parts := strings.Split(machineType, "-")
	cpus, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		return cloudprovider.ErrIllegalConfiguration
	}
	maxCpuInfo, found := maxCpuCount[gpuType]
	if !found {
		return cloudprovider.ErrIllegalConfiguration
	}
	maxCpus, found := maxCpuInfo[gpuCount]
	if !found || cpus > maxCpus {
		return cloudprovider.ErrIllegalConfiguration
	}

	return nil
}

func getNormalizedGpuCount(initialCount int64) (int64, error) {
	for i := int64(1); i <= int64(8); i = 2 * i {
		if i >= initialCount {
			return i, nil
		}
	}
	return 0, cloudprovider.ErrIllegalConfiguration
}
