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

package apis

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEmbeddedCRDs(t *testing.T) {
	// Map filename to embedded bytes
	expectedCRDs := map[string][]byte{
		"autoscaling.x-k8s.io_capacitybuffers.yaml":      CapacityBufferCRD,
		"autoscaling.x-k8s.io_capacityquotas.yaml":       CapacityQuotaCRD,
		"autoscaling.x-k8s.io_provisioningrequests.yaml": ProvisioningRequestCRD,
	}

	// 1. Verify that all embedded variables are non-empty
	for name, content := range expectedCRDs {
		if len(content) == 0 {
			t.Errorf("Embedded variable for %s is empty", name)
		}
	}

	// 2. Read the crd/ directory and make sure every YAML file is in expectedCRDs
	files, err := os.ReadDir("config/crd")
	if err != nil {
		t.Fatalf("Failed to read crd/ directory: %v", err)
	}

	foundFiles := make(map[string]bool)
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		ext := filepath.Ext(file.Name())
		if ext != ".yaml" && ext != ".yml" {
			continue
		}

		name := file.Name()
		foundFiles[name] = true

		embeddedBytes, ok := expectedCRDs[name]
		if !ok {
			t.Errorf("Found new CRD file %q in config/crd/ directory, but it is not exposed as an embedded variable in apis.go. Please add it.", name)
			continue
		}

		// 3. Verify that the embedded bytes match the on-disk file contents exactly
		onDiskBytes, err := os.ReadFile(filepath.Join("config/crd", name))
		if err != nil {
			t.Errorf("Failed to read on-disk file %q: %v", name, err)
			continue
		}

		if string(onDiskBytes) != string(embeddedBytes) {
			t.Errorf("Embedded content for %q does not match the on-disk file contents", name)
		}
	}

	// 4. Verify that we don't have extra files in expectedCRDs that don't exist on disk
	for name := range expectedCRDs {
		if !foundFiles[name] {
			t.Errorf("Config contains embedded variable for %q, but the file was not found in config/crd/ directory", name)
		}
	}
}
