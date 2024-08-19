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
	azclients "sigs.k8s.io/cloud-provider-azure/pkg/azureclients"
	providerazureconsts "sigs.k8s.io/cloud-provider-azure/pkg/consts"
	providerazureconfig "sigs.k8s.io/cloud-provider-azure/pkg/provider/config"
)

func TestCloudProviderAzureConsts(t *testing.T) {
	// Just detect user-facing breaking changes from cloud-provider-azure.
	// Shouldn't really change a lot, but just in case.
	assert.Equal(t, "vmss", providerazureconsts.VMTypeVMSS)
	assert.Equal(t, "standard", providerazureconsts.VMTypeStandard)
}

func TestInitializeCloudProviderRateLimitConfigWithNoRateLimitSettingsReturnsDefaults(t *testing.T) {
	emptyConfig := &providerazureconfig.CloudProviderRateLimitConfig{}
	providerazureconfig.InitializeCloudProviderRateLimitConfig(emptyConfig)

	assert.InDelta(t, emptyConfig.CloudProviderRateLimitQPS, providerazureconsts.RateLimitQPSDefault, 0.0001)
	assert.InDelta(t, emptyConfig.CloudProviderRateLimitBucket, providerazureconsts.RateLimitBucketDefault, 0.0001)
	assert.InDelta(t, emptyConfig.CloudProviderRateLimitQPSWrite, providerazureconsts.RateLimitQPSDefault, 0.0001)
	assert.InDelta(t, emptyConfig.CloudProviderRateLimitBucketWrite, providerazureconsts.RateLimitBucketDefault, 0.0001)
}

func TestInitializeCloudProviderRateLimitConfigWithReadRateLimitSettings(t *testing.T) {
	var rateLimitReadQPS float32 = 3.0
	rateLimitReadBuckets := 10

	cfg := &providerazureconfig.CloudProviderRateLimitConfig{
		RateLimitConfig: azclients.RateLimitConfig{
			CloudProviderRateLimitQPS:    rateLimitReadQPS,
			CloudProviderRateLimitBucket: rateLimitReadBuckets,
		},
	}
	//t.Setenv("RATE_LIMIT_READ_QPS", fmt.Sprintf("%.1f", rateLimitReadQPS)
	//t.Setenv("RATE_LIMIT_READ_BUCKETS", fmt.Sprintf("%d", rateLimitReadBuckets)

	providerazureconfig.InitializeCloudProviderRateLimitConfig(cfg)
	assert.InDelta(t, cfg.CloudProviderRateLimitQPS, rateLimitReadQPS, 0.0001)
	assert.InDelta(t, cfg.CloudProviderRateLimitBucket, rateLimitReadBuckets, 0.0001)
	assert.InDelta(t, cfg.CloudProviderRateLimitQPSWrite, rateLimitReadQPS, 0.0001)
	assert.InDelta(t, cfg.CloudProviderRateLimitBucketWrite, rateLimitReadBuckets, 0.0001)
	assert.InDelta(t, cfg.InterfaceRateLimit.CloudProviderRateLimitQPS, rateLimitReadQPS, 0.0001)
	assert.InDelta(t, cfg.InterfaceRateLimit.CloudProviderRateLimitBucket, rateLimitReadBuckets, 0.0001)
	assert.InDelta(t, cfg.InterfaceRateLimit.CloudProviderRateLimitQPSWrite, rateLimitReadQPS, 0.0001)
	assert.InDelta(t, cfg.InterfaceRateLimit.CloudProviderRateLimitBucketWrite, rateLimitReadBuckets, 0.0001)
	assert.InDelta(t, cfg.VirtualMachineRateLimit.CloudProviderRateLimitQPS, rateLimitReadQPS, 0.0001)
	assert.InDelta(t, cfg.VirtualMachineRateLimit.CloudProviderRateLimitBucket, rateLimitReadBuckets, 0.0001)
	assert.InDelta(t, cfg.VirtualMachineRateLimit.CloudProviderRateLimitQPSWrite, rateLimitReadQPS, 0.0001)
	assert.InDelta(t, cfg.VirtualMachineRateLimit.CloudProviderRateLimitBucketWrite, rateLimitReadBuckets, 0.0001)
	assert.InDelta(t, cfg.StorageAccountRateLimit.CloudProviderRateLimitQPS, rateLimitReadQPS, 0.0001)
	assert.InDelta(t, cfg.StorageAccountRateLimit.CloudProviderRateLimitBucket, rateLimitReadBuckets, 0.0001)
	assert.InDelta(t, cfg.StorageAccountRateLimit.CloudProviderRateLimitQPSWrite, rateLimitReadQPS, 0.0001)
	assert.InDelta(t, cfg.StorageAccountRateLimit.CloudProviderRateLimitBucketWrite, rateLimitReadBuckets, 0.0001)
	assert.InDelta(t, cfg.VirtualMachineScaleSetRateLimit.CloudProviderRateLimitQPS, rateLimitReadQPS, 0.0001)
	assert.InDelta(t, cfg.VirtualMachineScaleSetRateLimit.CloudProviderRateLimitBucket, rateLimitReadBuckets, 0.0001)
	assert.InDelta(t, cfg.VirtualMachineScaleSetRateLimit.CloudProviderRateLimitQPSWrite, rateLimitReadQPS, 0.0001)
	assert.InDelta(t, cfg.VirtualMachineScaleSetRateLimit.CloudProviderRateLimitBucketWrite, rateLimitReadBuckets, 0.0001)
}

func TestInitializeCloudProviderRateLimitConfigWithReadAndWriteRateLimitSettings(t *testing.T) {
	var rateLimitReadQPS float32 = 3.0
	rateLimitReadBuckets := 10
	var rateLimitWriteQPS float32 = 6.0
	rateLimitWriteBuckets := 20

	cfg := &providerazureconfig.CloudProviderRateLimitConfig{
		RateLimitConfig: azclients.RateLimitConfig{
			CloudProviderRateLimitBucket:      rateLimitReadBuckets,
			CloudProviderRateLimitBucketWrite: rateLimitWriteBuckets,
			CloudProviderRateLimitQPS:         rateLimitReadQPS,
			CloudProviderRateLimitQPSWrite:    rateLimitWriteQPS,
		},
	}

	//t.Setenv("RATE_LIMIT_READ_QPS", fmt.Sprintf("%.1f", rateLimitReadQPS)
	//t.Setenv("RATE_LIMIT_READ_BUCKETS", fmt.Sprintf("%d", rateLimitReadBuckets)
	//t.Setenv("RATE_LIMIT_WRITE_QPS", fmt.Sprintf("%.1f", rateLimitWriteQPS)
	//t.Setenv("RATE_LIMIT_WRITE_BUCKETS", fmt.Sprintf("%d", rateLimitWriteBuckets)

	providerazureconfig.InitializeCloudProviderRateLimitConfig(cfg)

	assert.InDelta(t, cfg.CloudProviderRateLimitQPS, rateLimitReadQPS, 0.0001)
	assert.InDelta(t, cfg.CloudProviderRateLimitBucket, rateLimitReadBuckets, 0.0001)
	assert.InDelta(t, cfg.CloudProviderRateLimitQPSWrite, rateLimitWriteQPS, 0.0001)
	assert.InDelta(t, cfg.CloudProviderRateLimitBucketWrite, rateLimitWriteBuckets, 0.0001)
	assert.InDelta(t, cfg.InterfaceRateLimit.CloudProviderRateLimitQPS, rateLimitReadQPS, 0.0001)
	assert.InDelta(t, cfg.InterfaceRateLimit.CloudProviderRateLimitBucket, rateLimitReadBuckets, 0.0001)
	assert.InDelta(t, cfg.InterfaceRateLimit.CloudProviderRateLimitQPSWrite, rateLimitWriteQPS, 0.0001)
	assert.InDelta(t, cfg.InterfaceRateLimit.CloudProviderRateLimitBucketWrite, rateLimitWriteBuckets, 0.0001)
	assert.InDelta(t, cfg.VirtualMachineRateLimit.CloudProviderRateLimitQPS, rateLimitReadQPS, 0.0001)
	assert.InDelta(t, cfg.VirtualMachineRateLimit.CloudProviderRateLimitBucket, rateLimitReadBuckets, 0.0001)
	assert.InDelta(t, cfg.VirtualMachineRateLimit.CloudProviderRateLimitQPSWrite, rateLimitWriteQPS, 0.0001)
	assert.InDelta(t, cfg.VirtualMachineRateLimit.CloudProviderRateLimitBucketWrite, rateLimitWriteBuckets, 0.0001)
	assert.InDelta(t, cfg.StorageAccountRateLimit.CloudProviderRateLimitQPS, rateLimitReadQPS, 0.0001)
	assert.InDelta(t, cfg.StorageAccountRateLimit.CloudProviderRateLimitBucket, rateLimitReadBuckets, 0.0001)
	assert.InDelta(t, cfg.StorageAccountRateLimit.CloudProviderRateLimitQPSWrite, rateLimitWriteQPS, 0.0001)
	assert.InDelta(t, cfg.StorageAccountRateLimit.CloudProviderRateLimitBucketWrite, rateLimitWriteBuckets, 0.0001)
	assert.InDelta(t, cfg.VirtualMachineScaleSetRateLimit.CloudProviderRateLimitQPS, rateLimitReadQPS, 0.0001)
	assert.InDelta(t, cfg.VirtualMachineScaleSetRateLimit.CloudProviderRateLimitBucket, rateLimitReadBuckets, 0.0001)
	assert.InDelta(t, cfg.VirtualMachineScaleSetRateLimit.CloudProviderRateLimitQPSWrite, rateLimitWriteQPS, 0.0001)
	assert.InDelta(t, cfg.VirtualMachineScaleSetRateLimit.CloudProviderRateLimitBucketWrite, rateLimitWriteBuckets, 0.0001)
}
