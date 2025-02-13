package provisioningrequest

import (
	"testing"

	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/provisioningrequest/autoscaling.x-k8s.io/v1"
)

func TestSupportedProvisioningClass(t *testing.T) {
	tests := []struct {
		name                                            string
		provisioningClassName                           string
		processorInstance                               v1.Parameter
		checkCapacityProvisioningClassProcessorInstance string
		want                                            bool
	}{
		{
			name:                  "Check capacity without instance",
			provisioningClassName: v1.ProvisioningClassCheckCapacity,
			want:                  true,
		},
		{
			name:                  "Check capacity with matching instance param",
			provisioningClassName: v1.ProvisioningClassCheckCapacity,
			processorInstance:     "instance",
			checkCapacityProvisioningClassProcessorInstance: "instance",
			want: true,
		},
		{
			name:                  "Check capacity with not matching instance param",
			provisioningClassName: v1.ProvisioningClassCheckCapacity,
			processorInstance:     "instance2",
			checkCapacityProvisioningClassProcessorInstance: "instance",
			want: false,
		},
		{
			name:                  "Check capacity with matching instance prefix",
			provisioningClassName: "instance" + v1.ProvisioningClassCheckCapacity,
			checkCapacityProvisioningClassProcessorInstance: "instance",
			want: true,
		},
		{
			name:                  "Check capacity with not matching instance prefix",
			provisioningClassName: "instance2" + v1.ProvisioningClassCheckCapacity,
			checkCapacityProvisioningClassProcessorInstance: "instance",
			want: false,
		},
		{
			name:                  "Best effort atomic",
			provisioningClassName: v1.ProvisioningClassBestEffortAtomicScaleUp,
			want:                  true,
		},
		{
			name:                  "Best effort atomic with instance ignored",
			provisioningClassName: v1.ProvisioningClassBestEffortAtomicScaleUp,
			checkCapacityProvisioningClassProcessorInstance: "instance",
			want: true,
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
			got := SupportedProvisioningClass(pr, test.checkCapacityProvisioningClassProcessorInstance)
			if test.want != got {
				t.Errorf("Expected SupportedProvisioningClass result: %v, got: %v", test.want, got)
			}
		})
	}
}
