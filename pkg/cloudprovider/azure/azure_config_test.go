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
	"testing"

	"github.com/stretchr/testify/assert"
	providerazureconsts "sigs.k8s.io/cloud-provider-azure/pkg/consts"
)

func TestCloudProviderAzureConsts(t *testing.T) {
	// Just detect user-facing breaking changes from cloud-provider-azure.
	// Shouldn't really change a lot, but just in case.
	assert.Equal(t, "vmss", providerazureconsts.VMTypeVMSS)
	assert.Equal(t, "standard", providerazureconsts.VMTypeStandard)
}

// Note: The previous tests for InitializeCloudProviderRateLimitConfig were removed
// because that function no longer exists in cloud-provider-azure v1.32.0+.
// The rate limit configuration has been restructured to use
// ratelimit.CloudProviderRateLimitConfig from pkg/azclient/policy/ratelimit
// with an Entries map for per-client configuration.
