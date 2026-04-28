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

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/klog/v2"
)

// FileSink implements DiagnosticSink by saving reports to the local filesystem.
type FileSink struct {
	directory string
}

// NewFileSink creates a new FileSink that stores reports in the specified directory.
func NewFileSink(directory string) (*FileSink, error) {
	if err := os.MkdirAll(directory, 0755); err != nil {
		return nil, fmt.Errorf("failed to create diagnostics directory %s: %v", directory, err)
	}
	return &FileSink{directory: directory}, nil
}

// Store saves each profile in the report as a separate file.
func (s *FileSink) Store(ctx context.Context, report *DiagnosticReport) error {
	for profileType, data := range report.Profiles {
		filename := fmt.Sprintf("%s.%s", report.ID, profileType)
		path := filepath.Join(s.directory, filename)

		if err := os.WriteFile(path, data, 0644); err != nil {
			klog.ErrorS(err, "Failed to save diagnostic profile", "path", path)
			continue
		}
		klog.V(2).InfoS("Saved diagnostic profile", "path", path)
	}
	return nil
}
