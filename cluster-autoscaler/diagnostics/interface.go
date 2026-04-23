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

package diagnostics

import "context"

// DiagnosticReport represents a collection of runtime profiles and associated metadata.
type DiagnosticReport struct {
	// ID is a unique identifier for the event that triggered this report.
	ID string
	// Profiles contains the collected profile data, keyed by profile type (e.g., "cpu", "trace", "heap").
	Profiles map[string][]byte
	// Metadata contains contextual tags related to the report.
	Metadata map[string]string
}

// DiagnosticSink defines the interface for storing diagnostic reports.
type DiagnosticSink interface {
	// Store saves the diagnostic report.
	Store(ctx context.Context, report *DiagnosticReport) error
}

// ProfileCollector defines the interface for collecting diagnostic profiles.
type ProfileCollector interface {
	// Collect gathers a profile.
	// For "cpu" and "trace", it should respect the context's deadline for collection duration.
	Collect(ctx context.Context) ([]byte, error)
}
