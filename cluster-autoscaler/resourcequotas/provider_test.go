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
	"errors"
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

func TestCombinedQuotasProvider(t *testing.T) {
	q1 := &FakeQuota{Name: "quota1"}
	q2 := &FakeQuota{Name: "quota2"}
	q3 := &FakeQuota{Name: "quota3"}
	providerErr := errors.New("test error")

	p1 := NewFakeProvider([]Quota{q1})
	p2 := NewFakeProvider([]Quota{q2, q3})
	pErr := NewFailingProvider(providerErr)

	testCases := []struct {
		name       string
		providers  []Provider
		wantQuotas []Quota
		wantErr    error
	}{
		{
			name:       "no providers",
			providers:  []Provider{},
			wantQuotas: nil,
		},
		{
			name:       "one provider",
			providers:  []Provider{p1},
			wantQuotas: []Quota{q1},
		},
		{
			name:       "multiple providers",
			providers:  []Provider{p1, p2},
			wantQuotas: []Quota{q1, q2, q3},
		},
		{
			name:       "provider with error",
			providers:  []Provider{p1, pErr},
			wantQuotas: nil,
			wantErr:    providerErr,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := NewCombinedQuotasProvider(tc.providers)
			quotas, err := provider.Quotas()

			if !errors.Is(err, tc.wantErr) {
				t.Errorf("Quotas() err mismatch: got %v, want %v", err, tc.wantErr)
			}
			if diff := cmp.Diff(tc.wantQuotas, quotas, cmp.Comparer(func(q1, q2 Quota) bool {
				return q1.ID() == q2.ID()
			})); diff != "" {
				t.Errorf("Quotas() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
