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

package egoscale

// optionalString returns the dereferenced string value of v if not nil, otherwise an empty string.
func optionalString(v *string) string {
	if v != nil {
		return *v
	}

	return ""
}

// optionalInt64 returns the dereferenced int64 value of v if not nil, otherwise 0.
func optionalInt64(v *int64) int64 {
	if v != nil {
		return *v
	}

	return 0
}
