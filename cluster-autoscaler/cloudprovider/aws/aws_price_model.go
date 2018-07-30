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

package aws

import (
	"math"
	"time"

	"fmt"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/price"
)

const (
	// Price calculation is based on m4.* machines (general purpose instances)
	// Base price m4.large $ 0.111 (2 CPU, 8 GB Ram)
	// Price share 50%/50% CPU/Ram => 0.0555 / 0.0555
	cpuPricePerHour         = 0.111 * 0.5 / 2
	memoryPricePerHourPerGb = 0.111 * 0.5 / 8

	// Reference is g3.4xlarge $ 1.21 (16CPU, 122GB Ram, 1GPU)
	// To compensate the gap between the different kind of instances, we just substrate the cpu and memory price
	// of the m4 instance by 80% of its initial value.
	gpuPricePerHour = 1.21 - (((16 * cpuPricePerHour) + (122 * memoryPricePerHourPerGb)) * 0.2)

	gigabyte = 1024.0 * 1024.0 * 1024.0

	resourceNvidiaGPU = "nvidia.com/gpu"
)

type instanceByASGFinder interface {
	GetAsgForInstance(instance AwsInstanceRef) *asg
}

// NewPriceModel is the constructor of priceModel which provides general access to price information
func NewPriceModel(asgs instanceByASGFinder, pd price.ShapeDescriptor) *priceModel {
	return &priceModel{
		asgs:            asgs,
		priceDescriptor: pd,
	}
}

// priceModel is responsible to provide pod as well as node price information.
// It implements the cloudprovider.PricingModel interface.
type priceModel struct {
	asgs            instanceByASGFinder
	priceDescriptor price.ShapeDescriptor
}

// NodePrice returns a price of running the given node for a given period of time.
// All prices are in USD.
func (pm *priceModel) NodePrice(node *apiv1.Node, startTime time.Time, endTime time.Time) (float64, error) {
	var (
		asgName string
		found   bool
	)

	if asgName, found = node.ObjectMeta.Annotations[nodeTemplateASGAnnotation]; !found {
		instanceRef, err := AwsRefFromProviderId(node.Spec.ProviderID)
		if err != nil {
			return 0, err
		}

		asg := pm.asgs.GetAsgForInstance(*instanceRef)
		if asg == nil {
			return 0, fmt.Errorf("asg for instance %s (%s) not found", instanceRef.String(), node.Spec.ProviderID)
		}

		asgName = asg.Name
	}

	hourlyPrice, err := pm.priceDescriptor.Price(asgName)
	if err != nil {
		return 0, fmt.Errorf("failed to describe price for asg %s: %v", asgName, err)
	}

	hours := getHours(startTime, endTime)

	return hourlyPrice * hours, nil
}

// PodPrice calculates the operating costs for the given pod under perfect conditions.
func (pm *priceModel) PodPrice(pod *apiv1.Pod, startTime time.Time, endTime time.Time) (float64, error) {
	sum := 0.0
	for _, container := range pod.Spec.Containers {
		sum += getBasePrice(container.Resources.Requests, startTime, endTime)
	}
	return sum, nil
}

func getBasePrice(resources apiv1.ResourceList, startTime time.Time, endTime time.Time) float64 {
	if len(resources) == 0 {
		return 0
	}
	hours := getHours(startTime, endTime)
	sum := 0.0
	cpu := resources[apiv1.ResourceCPU]
	mem := resources[apiv1.ResourceMemory]
	gpu := resources[resourceNvidiaGPU]
	sum += float64(cpu.MilliValue()) / 1000.0 * cpuPricePerHour * hours
	sum += float64(gpu.MilliValue()) / 1000.0 * gpuPricePerHour * hours
	sum += float64(mem.Value()) / gigabyte * memoryPricePerHourPerGb * hours

	return sum
}

func getHours(startTime time.Time, endTime time.Time) float64 {
	minutes := math.Ceil(float64(endTime.Sub(startTime)) / float64(time.Minute))
	hours := minutes / 60.0
	return hours
}
