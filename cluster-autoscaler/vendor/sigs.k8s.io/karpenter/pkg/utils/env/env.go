/*
Copyright The Kubernetes Authors.

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

package env

import (
	"os"
	"regexp"
	"runtime/debug"
	"strconv"
	"time"
)

// WithDefaultInt returns the int value of the supplied environment variable or, if not present,
// the supplied default value. If the int conversion fails, returns the default
func WithDefaultInt(key string, def int) int {
	val, ok := os.LookupEnv(key)
	if !ok {
		return def
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return def
	}
	return i
}

// WithDefaultInt64 returns the int value of the supplied environment variable or, if not present,
// the supplied default value. If the int conversion fails, returns the default
func WithDefaultInt64(key string, def int64) int64 {
	val, ok := os.LookupEnv(key)
	if !ok {
		return def
	}
	i, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return def
	}
	return i
}

// WithDefaultString returns the string value of the supplied environment variable or, if not present,
// the supplied default value.
func WithDefaultString(key string, def string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return def
	}
	return val
}

// WithDefaultBool returns the boolean value of the supplied environment variable or, if not present,
// the supplied default value.
func WithDefaultBool(key string, def bool) bool {
	val, ok := os.LookupEnv(key)
	if !ok {
		return def
	}
	parsedVal, err := strconv.ParseBool(val)
	if err != nil {
		return def
	}
	return parsedVal
}

// WithDefaultDuration returns the duration value of the supplied environment variable or, if not present,
// the supplied default value.
func WithDefaultDuration(key string, def time.Duration) time.Duration {
	val, ok := os.LookupEnv(key)
	if !ok {
		return def
	}
	parsedVal, err := time.ParseDuration(val)
	if err != nil {
		return def
	}
	return parsedVal
}

// GetRevision function is based on the function defined under https://pkg.go.dev/knative.dev/pkg@v0.0.0-20240815051656-89743d9bbf7c/changeset
// at https://github.com/knative/pkg/blob/89743d9bbf7c/changeset/commit.go#L51
func GetRevision() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}

	var revision string
	var modified bool

	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			revision = s.Value
		case "vcs.modified":
			modified, _ = strconv.ParseBool(s.Value)
		}
	}

	if revision == "" {
		return "unknown"
	}

	if regexp.MustCompile(`^[a-f0-9]{40,64}$`).MatchString(revision) {
		revision = revision[:7]
	}

	if modified {
		revision += "-dirty"
	}

	return revision
}
