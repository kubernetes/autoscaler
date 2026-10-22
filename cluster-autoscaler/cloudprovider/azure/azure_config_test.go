/*
Copyright 2020 The Kubernetes Authors.

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

package azure

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	providerazureconsts "sigs.k8s.io/cloud-provider-azure/pkg/consts"
)

func TestCloudProviderAzureConsts(t *testing.T) {
	// Just detect user-facing breaking changes from cloud-provider-azure.
	// Shouldn't really change a lot, but just in case.
	assert.Equal(t, "vmss", providerazureconsts.VMTypeVMSS)
	assert.Equal(t, "standard", providerazureconsts.VMTypeStandard)
}

func TestBuildAzureConfig_StaleInstanceCleanup(t *testing.T) {
	t.Setenv("ARM_SUBSCRIPTION_ID", "test-subscription")
	t.Setenv("ARM_RESOURCE_GROUP", "test-resource-group")
	t.Setenv("ARM_VM_TYPE", providerazureconsts.VMTypeVMSS)
	t.Setenv("ARM_USE_MANAGED_IDENTITY_EXTENSION", "true")
	t.Setenv("ARM_USE_FEDERATED_WORKLOAD_IDENTITY_EXTENSION", "false")
	t.Setenv("ARM_USE_WORKLOAD_IDENTITY_EXTENSION", "false")

	for _, tc := range []struct {
		name    string
		config  string
		enabled bool
		dryRun  bool
		minAge  int
	}{
		{name: "defaults", config: `{}`, dryRun: true, minAge: 5},
		{
			name:    "active cleanup",
			config:  `{"enableStaleInstanceCleanup":true,"staleInstanceCleanupDryRun":false,"staleInstanceMinAgeMinutes":7}`,
			enabled: true, minAge: 7,
		},
		{
			name:    "explicit dry run",
			config:  `{"enableStaleInstanceCleanup":true,"staleInstanceCleanupDryRun":true,"staleInstanceMinAgeMinutes":1}`,
			enabled: true, dryRun: true, minAge: 1,
		},
		{name: "zero age defaults", config: `{"staleInstanceMinAgeMinutes":0}`, dryRun: true, minAge: 5},
		{name: "negative age defaults", config: `{"staleInstanceMinAgeMinutes":-1}`, dryRun: true, minAge: 5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := BuildAzureConfig(strings.NewReader(tc.config))
			require.NoError(t, err)
			assert.Equal(t, tc.enabled, cfg.EnableStaleInstanceCleanup)
			assert.Equal(t, tc.dryRun, cfg.StaleInstanceCleanupDryRun)
			assert.Equal(t, tc.minAge, cfg.StaleInstanceMinAgeMinutes)
		})
	}
}

// Note: The previous tests for InitializeCloudProviderRateLimitConfig were removed
// because that function no longer exists in cloud-provider-azure v1.32.0+.
// The rate limit configuration has been restructured to use
// ratelimit.CloudProviderRateLimitConfig from pkg/azclient/policy/ratelimit
// with an Entries map for per-client configuration.
