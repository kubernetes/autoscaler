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

package provisioningrequest

import (
	"testing"

	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/provisioningrequest/autoscaling.x-k8s.io/v1"
)

func TestSupportedProvisioningClass(t *testing.T) {
	tests := []struct {
		name                           string
		provisioningClassName          string
		processorInstance              v1.Parameter
		checkCapacityProcessorInstance string
		want                           bool
	}{
		{
			name:                  "Check capacity without instance",
			provisioningClassName: v1.ProvisioningClassCheckCapacity,
			want:                  true,
		},
		{
			name:                           "Check capacity with matching instance param",
			provisioningClassName:          v1.ProvisioningClassCheckCapacity,
			processorInstance:              "instance",
			checkCapacityProcessorInstance: "instance",
			want:                           true,
		},
		{
			name:                           "Check capacity with not matching instance param",
			provisioningClassName:          v1.ProvisioningClassCheckCapacity,
			processorInstance:              "instance2",
			checkCapacityProcessorInstance: "instance",
			want:                           false,
		},
		{
			name:                           "Check capacity with matching instance prefix",
			provisioningClassName:          "instance" + v1.ProvisioningClassCheckCapacity,
			checkCapacityProcessorInstance: "instance",
			want:                           true,
		},
		{
			name:                           "Check capacity with not matching instance prefix",
			provisioningClassName:          "instance2" + v1.ProvisioningClassCheckCapacity,
			checkCapacityProcessorInstance: "instance",
			want:                           false,
		},
		{
			name:                  "Best effort atomic",
			provisioningClassName: v1.ProvisioningClassBestEffortAtomicScaleUp,
			want:                  true,
		},
		{
			name:                           "Best effort atomic with any instance",
			provisioningClassName:          v1.ProvisioningClassBestEffortAtomicScaleUp,
			checkCapacityProcessorInstance: "instance",
			want:                           false,
		},
		{
			name:                  "Invalid class name",
			provisioningClassName: "invalid",
			want:                  false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pr := &v1.ProvisioningRequest{
				Spec: v1.ProvisioningRequestSpec{
					ProvisioningClassName: test.provisioningClassName,
					Parameters: map[string]v1.Parameter{
						CheckCapacityProcessorInstanceKey: test.processorInstance,
					},
				},
			}
			got := SupportedProvisioningClass(pr, test.checkCapacityProcessorInstance)
			if test.want != got {
				t.Errorf("Expected SupportedProvisioningClass result: %v, got: %v", test.want, got)
			}
		})
	}
}
