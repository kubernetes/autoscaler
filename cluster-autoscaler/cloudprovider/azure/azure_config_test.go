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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	azclients "sigs.k8s.io/cloud-provider-azure/pkg/azureclients"
)

func TestInitializeCloudProviderRateLimitConfigWithNoConfigReturnsNoError(t *testing.T) {
	err := initializeCloudProviderRateLimitConfig(nil)
	assert.Nil(t, err, "err should be nil")
}

func TestInitializeCloudProviderRateLimitConfigWithNoRateLimitSettingsReturnsDefaults(t *testing.T) {
	emptyConfig := &CloudProviderRateLimitConfig{}
	err := initializeCloudProviderRateLimitConfig(emptyConfig)

	assert.NoError(t, err)
	assert.Equal(t, emptyConfig.CloudProviderRateLimitQPS, rateLimitQPSDefault)
	assert.Equal(t, emptyConfig.CloudProviderRateLimitBucket, rateLimitBucketDefault)
	assert.Equal(t, emptyConfig.CloudProviderRateLimitQPSWrite, rateLimitQPSDefault)
	assert.Equal(t, emptyConfig.CloudProviderRateLimitBucketWrite, rateLimitBucketDefault)
}

func TestInitializeCloudProviderRateLimitConfigWithReadRateLimitSettingsFromEnv(t *testing.T) {
	emptyConfig := &CloudProviderRateLimitConfig{}
	var rateLimitReadQPS float32 = 3.0
	rateLimitReadBuckets := 10
	t.Setenv(rateLimitReadQPSEnvVar, fmt.Sprintf("%.1f", rateLimitReadQPS))
	t.Setenv(rateLimitReadBucketsEnvVar, fmt.Sprintf("%d", rateLimitReadBuckets))

	err := initializeCloudProviderRateLimitConfig(emptyConfig)
	assert.NoError(t, err)
	assert.Equal(t, emptyConfig.CloudProviderRateLimitQPS, rateLimitReadQPS)
	assert.Equal(t, emptyConfig.CloudProviderRateLimitBucket, rateLimitReadBuckets)
	assert.Equal(t, emptyConfig.CloudProviderRateLimitQPSWrite, rateLimitReadQPS)
	assert.Equal(t, emptyConfig.CloudProviderRateLimitBucketWrite, rateLimitReadBuckets)
}

func TestInitializeCloudProviderRateLimitConfigWithReadAndWriteRateLimitSettingsFromEnv(t *testing.T) {
	emptyConfig := &CloudProviderRateLimitConfig{}
	var rateLimitReadQPS float32 = 3.0
	rateLimitReadBuckets := 10
	var rateLimitWriteQPS float32 = 6.0
	rateLimitWriteBuckets := 20

	t.Setenv(rateLimitReadQPSEnvVar, fmt.Sprintf("%.1f", rateLimitReadQPS))
	t.Setenv(rateLimitReadBucketsEnvVar, fmt.Sprintf("%d", rateLimitReadBuckets))
	t.Setenv(rateLimitWriteQPSEnvVar, fmt.Sprintf("%.1f", rateLimitWriteQPS))
	t.Setenv(rateLimitWriteBucketsEnvVar, fmt.Sprintf("%d", rateLimitWriteBuckets))

	err := initializeCloudProviderRateLimitConfig(emptyConfig)

	assert.NoError(t, err)
	assert.Equal(t, emptyConfig.CloudProviderRateLimitQPS, rateLimitReadQPS)
	assert.Equal(t, emptyConfig.CloudProviderRateLimitBucket, rateLimitReadBuckets)
	assert.Equal(t, emptyConfig.CloudProviderRateLimitQPSWrite, rateLimitWriteQPS)
	assert.Equal(t, emptyConfig.CloudProviderRateLimitBucketWrite, rateLimitWriteBuckets)
}

func TestInitializeCloudProviderRateLimitConfigWithReadAndWriteRateLimitAlreadySetInConfig(t *testing.T) {
	var rateLimitReadQPS float32 = 3.0
	rateLimitReadBuckets := 10
	var rateLimitWriteQPS float32 = 6.0
	rateLimitWriteBuckets := 20

	configWithRateLimits := &CloudProviderRateLimitConfig{
		RateLimitConfig: azclients.RateLimitConfig{
			CloudProviderRateLimitBucket:      rateLimitReadBuckets,
			CloudProviderRateLimitBucketWrite: rateLimitWriteBuckets,
			CloudProviderRateLimitQPS:         rateLimitReadQPS,
			CloudProviderRateLimitQPSWrite:    rateLimitWriteQPS,
		},
	}

	t.Setenv(rateLimitReadQPSEnvVar, "99")
	t.Setenv(rateLimitReadBucketsEnvVar, "99")
	t.Setenv(rateLimitWriteQPSEnvVar, "99")
	t.Setenv(rateLimitWriteBucketsEnvVar, "99")

	err := initializeCloudProviderRateLimitConfig(configWithRateLimits)

	assert.NoError(t, err)
	assert.Equal(t, configWithRateLimits.CloudProviderRateLimitQPS, rateLimitReadQPS)
	assert.Equal(t, configWithRateLimits.CloudProviderRateLimitBucket, rateLimitReadBuckets)
	assert.Equal(t, configWithRateLimits.CloudProviderRateLimitQPSWrite, rateLimitWriteQPS)
	assert.Equal(t, configWithRateLimits.CloudProviderRateLimitBucketWrite, rateLimitWriteBuckets)
}

func TestInitializeCloudProviderRateLimitConfigWithInvalidReadAndWriteRateLimitSettingsFromEnv(t *testing.T) {
	emptyConfig := &CloudProviderRateLimitConfig{}
	var rateLimitReadQPS float32 = 3.0
	rateLimitReadBuckets := 10
	var rateLimitWriteQPS float32 = 6.0
	rateLimitWriteBuckets := 20

	invalidSetting := "invalid"
	testCases := []struct {
		desc                                 string
		isInvalidRateLimitReadQPSEnvVar      bool
		isInvalidRateLimitReadBucketsEnvVar  bool
		isInvalidRateLimitWriteQPSEnvVar     bool
		isInvalidRateLimitWriteBucketsEnvVar bool
		expectedErr                          bool
		expectedErrMsg                       error
	}{
		{
			desc:                            "an error shall be returned if invalid rateLimitReadQPSEnvVar",
			isInvalidRateLimitReadQPSEnvVar: true,
			expectedErr:                     true,
			expectedErrMsg:                  fmt.Errorf("failed to parse %s: %q, strconv.ParseFloat: parsing \"invalid\": invalid syntax", rateLimitReadQPSEnvVar, invalidSetting),
		},
		{
			desc:                                "an error shall be returned if invalid rateLimitReadBucketsEnvVar",
			isInvalidRateLimitReadBucketsEnvVar: true,
			expectedErr:                         true,
			expectedErrMsg:                      fmt.Errorf("failed to parse %s: %q, strconv.ParseInt: parsing \"invalid\": invalid syntax", rateLimitReadBucketsEnvVar, invalidSetting),
		},
		{
			desc:                             "an error shall be returned if invalid rateLimitWriteQPSEnvVar",
			isInvalidRateLimitWriteQPSEnvVar: true,
			expectedErr:                      true,
			expectedErrMsg:                   fmt.Errorf("failed to parse %s: %q, strconv.ParseFloat: parsing \"invalid\": invalid syntax", rateLimitWriteQPSEnvVar, invalidSetting),
		},
		{
			desc:                                 "an error shall be returned if invalid rateLimitWriteBucketsEnvVar",
			isInvalidRateLimitWriteBucketsEnvVar: true,
			expectedErr:                          true,
			expectedErrMsg:                       fmt.Errorf("failed to parse %s: %q, strconv.ParseInt: parsing \"invalid\": invalid syntax", rateLimitWriteBucketsEnvVar, invalidSetting),
		},
	}

	for i, test := range testCases {
		if test.isInvalidRateLimitReadQPSEnvVar {
			t.Setenv(rateLimitReadQPSEnvVar, invalidSetting)
		} else {
			t.Setenv(rateLimitReadQPSEnvVar, fmt.Sprintf("%.1f", rateLimitReadQPS))
		}
		if test.isInvalidRateLimitReadBucketsEnvVar {
			t.Setenv(rateLimitReadBucketsEnvVar, invalidSetting)
		} else {
			t.Setenv(rateLimitReadBucketsEnvVar, fmt.Sprintf("%d", rateLimitReadBuckets))
		}
		if test.isInvalidRateLimitWriteQPSEnvVar {
			t.Setenv(rateLimitWriteQPSEnvVar, invalidSetting)
		} else {
			t.Setenv(rateLimitWriteQPSEnvVar, fmt.Sprintf("%.1f", rateLimitWriteQPS))
		}
		if test.isInvalidRateLimitWriteBucketsEnvVar {
			t.Setenv(rateLimitWriteBucketsEnvVar, invalidSetting)
		} else {
			t.Setenv(rateLimitWriteBucketsEnvVar, fmt.Sprintf("%d", rateLimitWriteBuckets))
		}

		err := initializeCloudProviderRateLimitConfig(emptyConfig)

		assert.Equal(t, test.expectedErr, err != nil, "TestCase[%d]: %s, return error: %v", i, test.desc, err)
		assert.Equal(t, test.expectedErrMsg, err, "TestCase[%d]: %s, expected: %v, return: %v", i, test.desc, test.expectedErrMsg, err)
	}
}

func TestOverrideDefaultRateLimitConfig(t *testing.T) {
	var rateLimitReadQPS float32 = 3.0
	rateLimitReadBuckets := 10
	var rateLimitWriteQPS float32 = 6.0
	rateLimitWriteBuckets := 20

	defaultConfigWithRateLimits := &CloudProviderRateLimitConfig{
		RateLimitConfig: azclients.RateLimitConfig{
			CloudProviderRateLimitBucket:      rateLimitReadBuckets,
			CloudProviderRateLimitBucketWrite: rateLimitWriteBuckets,
			CloudProviderRateLimitQPS:         rateLimitReadQPS,
			CloudProviderRateLimitQPSWrite:    rateLimitWriteQPS,
		},
	}

	configWithRateLimits := &CloudProviderRateLimitConfig{
		RateLimitConfig: azclients.RateLimitConfig{
			CloudProviderRateLimit:            true,
			CloudProviderRateLimitBucket:      0,
			CloudProviderRateLimitBucketWrite: 0,
			CloudProviderRateLimitQPS:         0,
			CloudProviderRateLimitQPSWrite:    0,
		},
	}

	newconfig := overrideDefaultRateLimitConfig(&defaultConfigWithRateLimits.RateLimitConfig, &configWithRateLimits.RateLimitConfig)

	assert.Equal(t, defaultConfigWithRateLimits.CloudProviderRateLimitQPS, newconfig.CloudProviderRateLimitQPS)
	assert.Equal(t, defaultConfigWithRateLimits.CloudProviderRateLimitBucket, newconfig.CloudProviderRateLimitBucket)
	assert.Equal(t, defaultConfigWithRateLimits.CloudProviderRateLimitQPSWrite, newconfig.CloudProviderRateLimitQPSWrite)
	assert.Equal(t, defaultConfigWithRateLimits.CloudProviderRateLimitBucketWrite, newconfig.CloudProviderRateLimitBucketWrite)

	falseCloudProviderRateLimit := &CloudProviderRateLimitConfig{
		RateLimitConfig: azclients.RateLimitConfig{
			CloudProviderRateLimit: false,
		},
	}
	newconfig = overrideDefaultRateLimitConfig(&defaultConfigWithRateLimits.RateLimitConfig, &falseCloudProviderRateLimit.RateLimitConfig)
	assert.Equal(t, &falseCloudProviderRateLimit.RateLimitConfig, newconfig)
}
