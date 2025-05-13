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

package equinixmetal

import (
	"math"
	"time"

	apiv1 "k8s.io/api/core/v1"
	podutils "k8s.io/autoscaler/cluster-autoscaler/utils/pod"
	"k8s.io/autoscaler/cluster-autoscaler/utils/units"
)

// Price implements Price interface for Equinix Metal.
type Price struct {
}

const (
	cpuPricePerHour         = 0.005208
	memoryPricePerHourPerGb = 0.003815
)

var instancePrices = map[string]float64{
	"a3.large.x86":   7.5000,
	"c2.medium.x86":  1.0000,
	"c3.large.arm64": 2.5000,
	"c3.medium.x86":  1.3500,
	"c3.small.x86":   0.7500,
	"g2.large.x86":   5.0000,
	"m2.xlarge.x86":  2.9000,
	"m3.large.x86":   3.1000,
	"m3.small.x86":   1.0500,
	"n2.xlarge.x86":  3.2500,
	"n3.xlarge.x86":  4.5000,
	"s3.xlarge.x86":  2.9500,
	"t3.small.x86":   0.3500,
	"x2.xlarge.x86":  2.5000,
}

// NodePrice returns a price of running the given node for a given period of time.
// All prices are in USD.
func (model *Price) NodePrice(node *apiv1.Node, startTime time.Time, endTime time.Time) (float64, error) {
	price := 0.0
	if node.Labels != nil {
		if machineType, found := node.Labels[apiv1.LabelInstanceType]; found {
			if pricePerHour, found := instancePrices[machineType]; found {
				price = pricePerHour * getHours(startTime, endTime)
			}
		}
	}
	return price, nil
}

func getHours(startTime time.Time, endTime time.Time) float64 {
	minutes := math.Ceil(float64(endTime.Sub(startTime)) / float64(time.Minute))
	hours := minutes / 60.0
	return hours
}

// PodPrice returns a theoretical minimum price of running a pod for a given
// period of time on a perfectly matching machine.
func (model *Price) PodPrice(pod *apiv1.Pod, startTime time.Time, endTime time.Time) (float64, error) {
	podRequests := podutils.PodRequests(pod)
	return getBasePrice(podRequests, startTime, endTime), nil
}

func getBasePrice(resources apiv1.ResourceList, startTime time.Time, endTime time.Time) float64 {
	if len(resources) == 0 {
		return 0
	}
	hours := getHours(startTime, endTime)
	price := 0.0
	cpu := resources[apiv1.ResourceCPU]
	mem := resources[apiv1.ResourceMemory]
	price += float64(cpu.MilliValue()) / 1000.0 * cpuPricePerHour * hours
	price += float64(mem.Value()) / float64(units.GiB) * memoryPricePerHourPerGb * hours
	return price
}
