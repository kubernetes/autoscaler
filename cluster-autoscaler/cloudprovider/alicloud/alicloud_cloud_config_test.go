/*
Copyright 2018 The Kubernetes Authors.

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

package alicloud

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAccessKeyCloudConfigIsValid(t *testing.T) {
	t.Setenv(accessKeyId, "id")
	t.Setenv(accessKeySecret, "secret")
	t.Setenv(regionId, "cn-hangzhou")

	cfg := &cloudConfig{}
	assert.True(t, cfg.isValid())
	assert.False(t, cfg.RRSAEnabled)
}

func TestRRSACloudConfigIsValid(t *testing.T) {
	t.Setenv(oidcProviderARN, "acs:ram::12345:oidc-provider/ack-rrsa-cb123")
	t.Setenv(oidcTokenFilePath, "/var/run/secrets/tokens/oidc-token")
	t.Setenv(roleARN, "acs:ram::12345:role/autoscaler-role")
	t.Setenv(roleSessionName, "session")
	t.Setenv(regionId, "cn-hangzhou")

	cfg := &cloudConfig{}
	assert.True(t, cfg.isValid())
	assert.True(t, cfg.RRSAEnabled)
}

func TestOldRRSACloudConfigIsValid(t *testing.T) {
	t.Setenv(oldOidcProviderARN, "acs:ram::12345:oidc-provider/ack-rrsa-cb123")
	t.Setenv(oldOidcTokenFilePath, "/var/run/secrets/tokens/oidc-token")
	t.Setenv(oldRoleARN, "acs:ram::12345:role/autoscaler-role")
	t.Setenv(oldRoleSessionName, "session")
	t.Setenv(regionId, "cn-hangzhou")

	cfg := &cloudConfig{}
	assert.True(t, cfg.isValid())
	assert.True(t, cfg.RRSAEnabled)
}

func TestFirstNotEmpty(t *testing.T) {
	// Test case where the first non-empty string is at the beginning
	result := firstNotEmpty("hello", "world", "test")
	assert.Equal(t, "hello", result)

	// Test case where the first non-empty string is in the middle
	result = firstNotEmpty("", "foo", "bar")
	assert.Equal(t, "foo", result)

	// Test case where the first non-empty string is at the end
	result = firstNotEmpty("", "", "baz")
	assert.Equal(t, "baz", result)

	// Test case where all strings are empty
	result = firstNotEmpty("", "", "")
	assert.Equal(t, "", result)

	// Test case with no arguments
	result = firstNotEmpty()
	assert.Equal(t, "", result)
}
