/*
Copyright 2019 The Kubernetes Authors.

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

package verda

import (
	"runtime/debug"
	"strings"
)

const (
	// fallbackVersion is used when build info is not available (e.g., during development)
	fallbackVersion = "1.2.1"

	// sdkName is the identifier for this SDK
	sdkName = "verdacloud-sdk-go"
)

// SDKVersion returns the version of the SDK.
// It attempts to get the version from Go module build info (which works when
// the SDK is used as a dependency). Falls back to a hardcoded version for
// development/testing scenarios.
func SDKVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		// When used as a dependency, find our module in the deps
		for _, dep := range info.Deps {
			if strings.HasSuffix(dep.Path, sdkName) {
				return strings.TrimPrefix(dep.Version, "v")
			}
		}
		// When running from within the module itself, use Main
		if strings.HasSuffix(info.Main.Path, sdkName) && info.Main.Version != "(devel)" {
			return strings.TrimPrefix(info.Main.Version, "v")
		}
	}
	return fallbackVersion
}

// DefaultUserAgent returns the default User-Agent string for the SDK.
func DefaultUserAgent() string {
	return sdkName + "/" + SDKVersion()
}

// BuildUserAgent constructs the full User-Agent string.
// If customUA is provided, it prepends it to the SDK's default User-Agent.
// If customUA is empty, it returns just the SDK's default User-Agent.
func BuildUserAgent(customUA string) string {
	defaultUA := DefaultUserAgent()
	if customUA == "" {
		return defaultUA
	}
	return customUA + " " + defaultUA
}
