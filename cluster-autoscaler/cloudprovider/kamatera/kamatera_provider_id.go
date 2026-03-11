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

import "strings"

const defaultKamateraProviderIDPrefix = "kamatera://"

func normalizeKamateraProviderIDPrefix(providerIDPrefix string) string {
	if providerIDPrefix == "" {
		return defaultKamateraProviderIDPrefix
	}
	return providerIDPrefix
}

func formatKamateraProviderID(providerIDPrefix, serverName string) string {
	if serverName == "" {
		return ""
	}
	return normalizeKamateraProviderIDPrefix(providerIDPrefix) + serverName
}

func parseKamateraProviderID(providerIDPrefix, providerID string) string {
	if providerID == "" {
		return ""
	}

	prefix := normalizeKamateraProviderIDPrefix(providerIDPrefix)
	trimmed := strings.TrimPrefix(providerID, prefix)
	if trimmed != providerID {
		return trimmed
	}

	// Backwards compatibility: accept the old default prefix even if the configured prefix differs.
	if prefix != defaultKamateraProviderIDPrefix {
		trimmed = strings.TrimPrefix(providerID, defaultKamateraProviderIDPrefix)
		if trimmed != providerID {
			return trimmed
		}
	}

	return providerID
}
