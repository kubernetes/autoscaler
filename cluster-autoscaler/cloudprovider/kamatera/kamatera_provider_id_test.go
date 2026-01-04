/*
Copyright 2024 The Kubernetes Authors.

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

package kamatera

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatKamateraProviderID(t *testing.T) {
	// test empty string returns empty string
	result := formatKamateraProviderID("")
	assert.Equal(t, "", result)

	// test server name is prefixed with kamatera://
	serverName := "my-server-123"
	result = formatKamateraProviderID(serverName)
	assert.Equal(t, "kamatera://my-server-123", result)
}

func TestParseKamateraProviderID(t *testing.T) {
	// test empty string returns empty string
	result := parseKamateraProviderID("")
	assert.Equal(t, "", result)

	// test provider ID with kamatera:// prefix returns server name
	providerID := "kamatera://my-server-123"
	result = parseKamateraProviderID(providerID)
	assert.Equal(t, "my-server-123", result)

	// test provider ID without prefix returns the same string (for backwards compatibility)
	providerIDWithoutPrefix := "my-server-456"
	result = parseKamateraProviderID(providerIDWithoutPrefix)
	assert.Equal(t, "my-server-456", result)
}

func TestFormatAndParseRoundTrip(t *testing.T) {
	// test that formatting and then parsing returns the original server name
	serverName := mockKamateraServerName()
	providerID := formatKamateraProviderID(serverName)
	parsedServerName := parseKamateraProviderID(providerID)
	assert.Equal(t, serverName, parsedServerName)
}
