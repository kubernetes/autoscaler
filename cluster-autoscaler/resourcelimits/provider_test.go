package resourcelimits

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/utils/units"
)

func TestCloudLimitersProvider(t *testing.T) {
	cloudProvider := test.NewTestCloudProviderBuilder().Build()
	minLimits := map[string]int64{"cpu": 2, "memory": 8 * units.GiB}
	maxLimits := map[string]int64{"cpu": 4, "memory": 16 * units.GiB}
	resourceLimiter := cloudprovider.NewResourceLimiter(minLimits, maxLimits)
	cloudProvider.SetResourceLimiter(resourceLimiter)

	limitsProvider := NewCloudLimitersProvider(cloudProvider)
	limiters, err := limitsProvider.AllLimiters()
	if err != nil {
		t.Errorf("failed to get limiters: %v", err)
	}
	if len(limiters) != 1 {
		t.Errorf("got %d limiters, expected 1", len(limiters))
	}
	limiter := limiters[0]
	if diff := cmp.Diff(minLimits, limiter.MinLimits()); diff != "" {
		t.Errorf("MinLimits() mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(maxLimits, limiter.MaxLimits()); diff != "" {
		t.Errorf("MaxLimits() mismatch (-want +got):\n%s", diff)
	}
}
