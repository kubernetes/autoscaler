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
	"github.com/stretchr/testify/assert"
	azclients "k8s.io/legacy-cloud-providers/azure/clients"
	"os"
	"testing"
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
	os.Setenv(rateLimitReadQPSEnvVar, fmt.Sprintf("%.1f", rateLimitReadQPS))
	os.Setenv(rateLimitReadBucketsEnvVar, fmt.Sprintf("%d", rateLimitReadBuckets))

	err := initializeCloudProviderRateLimitConfig(emptyConfig)
	assert.NoError(t, err)
	assert.Equal(t, emptyConfig.CloudProviderRateLimitQPS, rateLimitReadQPS)
	assert.Equal(t, emptyConfig.CloudProviderRateLimitBucket, rateLimitReadBuckets)
	assert.Equal(t, emptyConfig.CloudProviderRateLimitQPSWrite, rateLimitReadQPS)
	assert.Equal(t, emptyConfig.CloudProviderRateLimitBucketWrite, rateLimitReadBuckets)

	os.Unsetenv(rateLimitReadBucketsEnvVar)
	os.Unsetenv(rateLimitReadQPSEnvVar)
}

func TestInitializeCloudProviderRateLimitConfigWithReadAndWriteRateLimitSettingsFromEnv(t *testing.T) {
	emptyConfig := &CloudProviderRateLimitConfig{}
	var rateLimitReadQPS float32 = 3.0
	rateLimitReadBuckets := 10
	var rateLimitWriteQPS float32 = 6.0
	rateLimitWriteBuckets := 20

	os.Setenv(rateLimitReadQPSEnvVar, fmt.Sprintf("%.1f", rateLimitReadQPS))
	os.Setenv(rateLimitReadBucketsEnvVar, fmt.Sprintf("%d", rateLimitReadBuckets))
	os.Setenv(rateLimitWriteQPSEnvVar, fmt.Sprintf("%.1f", rateLimitWriteQPS))
	os.Setenv(rateLimitWriteBucketsEnvVar, fmt.Sprintf("%d", rateLimitWriteBuckets))

	err := initializeCloudProviderRateLimitConfig(emptyConfig)

	assert.NoError(t, err)
	assert.Equal(t, emptyConfig.CloudProviderRateLimitQPS, rateLimitReadQPS)
	assert.Equal(t, emptyConfig.CloudProviderRateLimitBucket, rateLimitReadBuckets)
	assert.Equal(t, emptyConfig.CloudProviderRateLimitQPSWrite, rateLimitWriteQPS)
	assert.Equal(t, emptyConfig.CloudProviderRateLimitBucketWrite, rateLimitWriteBuckets)

	os.Unsetenv(rateLimitReadQPSEnvVar)
	os.Unsetenv(rateLimitReadBucketsEnvVar)
	os.Unsetenv(rateLimitWriteQPSEnvVar)
	os.Unsetenv(rateLimitWriteBucketsEnvVar)
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

	os.Setenv(rateLimitReadQPSEnvVar, "99")
	os.Setenv(rateLimitReadBucketsEnvVar, "99")
	os.Setenv(rateLimitWriteQPSEnvVar, "99")
	os.Setenv(rateLimitWriteBucketsEnvVar, "99")

	err := initializeCloudProviderRateLimitConfig(configWithRateLimits)

	assert.NoError(t, err)
	assert.Equal(t, configWithRateLimits.CloudProviderRateLimitQPS, rateLimitReadQPS)
	assert.Equal(t, configWithRateLimits.CloudProviderRateLimitBucket, rateLimitReadBuckets)
	assert.Equal(t, configWithRateLimits.CloudProviderRateLimitQPSWrite, rateLimitWriteQPS)
	assert.Equal(t, configWithRateLimits.CloudProviderRateLimitBucketWrite, rateLimitWriteBuckets)

	os.Unsetenv(rateLimitReadQPSEnvVar)
	os.Unsetenv(rateLimitReadBucketsEnvVar)
	os.Unsetenv(rateLimitWriteQPSEnvVar)
	os.Unsetenv(rateLimitWriteBucketsEnvVar)
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
			os.Setenv(rateLimitReadQPSEnvVar, invalidSetting)
		} else {
			os.Setenv(rateLimitReadQPSEnvVar, fmt.Sprintf("%.1f", rateLimitReadQPS))
		}
		if test.isInvalidRateLimitReadBucketsEnvVar {
			os.Setenv(rateLimitReadBucketsEnvVar, invalidSetting)
		} else {
			os.Setenv(rateLimitReadBucketsEnvVar, fmt.Sprintf("%d", rateLimitReadBuckets))
		}
		if test.isInvalidRateLimitWriteQPSEnvVar {
			os.Setenv(rateLimitWriteQPSEnvVar, invalidSetting)
		} else {
			os.Setenv(rateLimitWriteQPSEnvVar, fmt.Sprintf("%.1f", rateLimitWriteQPS))
		}
		if test.isInvalidRateLimitWriteBucketsEnvVar {
			os.Setenv(rateLimitWriteBucketsEnvVar, invalidSetting)
		} else {
			os.Setenv(rateLimitWriteBucketsEnvVar, fmt.Sprintf("%d", rateLimitWriteBuckets))
		}

		err := initializeCloudProviderRateLimitConfig(emptyConfig)

		assert.Equal(t, test.expectedErr, err != nil, "TestCase[%d]: %s, return error: %v", i, test.desc, err)
		assert.Equal(t, test.expectedErrMsg, err, "TestCase[%d]: %s, expected: %v, return: %v", i, test.desc, test.expectedErrMsg, err)

		os.Unsetenv(rateLimitReadQPSEnvVar)
		os.Unsetenv(rateLimitReadBucketsEnvVar)
		os.Unsetenv(rateLimitWriteQPSEnvVar)
		os.Unsetenv(rateLimitWriteBucketsEnvVar)
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

func TestGetSubscriptionIdFromInstanceMetadata(t *testing.T) {
	// metadataURL in azure_manager.go is not available for our tests, expect fail.
	result, err := getSubscriptionIdFromInstanceMetadata()
	expected := ""
	assert.NotNil(t, err.Error())
	assert.Equal(t, expected, result, "Verify return result failed, expected: %v, actual: %v", expected, result)
}
