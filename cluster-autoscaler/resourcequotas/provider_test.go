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

	quotasProvider := NewCloudQuotasProvider(cloudProvider)
	quotas, err := quotasProvider.Quotas()
	if err != nil {
		t.Errorf("failed to get quotas: %v", err)
	}
	if len(quotas) != 1 {
		t.Errorf("got %d quotas, expected 1", len(quotas))
	}
	quota := quotas[0]
	if diff := cmp.Diff(maxLimits, quota.Limits()); diff != "" {
		t.Errorf("Limits() mismatch (-want +got):\n%s", diff)
	}
}
