package resourcequotas

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/utils/units"
)

func TestCloudLimitersProvider(t *testing.T) {
	cloudProvider := test.NewTestCloudProviderBuilder().Build()
	maxLimits := map[string]int64{"cpu": 4, "memory": 16 * units.GiB}
	resourceLimiter := cloudprovider.NewResourceLimiter(nil, maxLimits)
	cloudProvider.SetResourceLimiter(resourceLimiter)

	limitsProvider := NewCloudQuotasProvider(cloudProvider)
	limiters, err := limitsProvider.Quotas()
	if err != nil {
		t.Errorf("failed to get quotas: %v", err)
	}
	if len(limiters) != 1 {
		t.Errorf("got %d quotas, expected 1", len(limiters))
	}
	limiter := limiters[0]
	if diff := cmp.Diff(maxLimits, limiter.Limits()); diff != "" {
		t.Errorf("Limits() mismatch (-want +got):\n%s", diff)
	}
}
