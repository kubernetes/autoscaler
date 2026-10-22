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
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestEmbeddedCRDs(t *testing.T) {
	// expectedCRDs maps CRD filenames in config/crd/ to their corresponding embedded byte variables.
	// When adding new CRD definitions to config/crd/, please ensure you:
	// 1. Expose them as an embedded []byte variable in embedcrds.go.
	// 2. Add the variable to AllCRDs in embedcrds.go.
	// 3. Register the new mapping here in expectedCRDs.
	expectedCRDs := map[string][]byte{
		"autoscaling.x-k8s.io_capacitybuffers.yaml":      CapacityBufferCRD,
		"autoscaling.x-k8s.io_capacityquotas.yaml":       CapacityQuotaCRD,
		"autoscaling.x-k8s.io_provisioningrequests.yaml": ProvisioningRequestCRD,
	}

	// Read the crd/ directory and make sure every YAML file is in expectedCRDs
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
			t.Errorf("Found new CRD file %q in config/crd/ directory, but it is not exposed as an embedded variable in embedcrds.go. Please add it.", name)
			continue
		}

		// Verify that the embedded bytes match the on-disk file contents exactly
		onDiskBytes, err := os.ReadFile(filepath.Join("config/crd", name))
		if err != nil {
			t.Errorf("Failed to read on-disk file %q: %v", name, err)
			continue
		}

		if string(onDiskBytes) != string(embeddedBytes) {
			t.Errorf("Embedded content for %q does not match the on-disk file contents", name)
		}
	}

	// Verify that we don't have extra files in expectedCRDs that don't exist on disk
	for name := range expectedCRDs {
		if !foundFiles[name] {
			t.Errorf("Config contains embedded variable for %q, but the file was not found in config/crd/ directory", name)
		}
	}

	// Verify that AllCRDs contains all expected CRDs and its length matches the number of CRD files found on disk
	if len(AllCRDs) != len(foundFiles) {
		t.Errorf("AllCRDs has length %d, but found %d CRD files in config/crd/", len(AllCRDs), len(foundFiles))
	}
	for name, expectedBytes := range expectedCRDs {
		found := false
		for _, crd := range AllCRDs {
			if bytes.Equal(crd, expectedBytes) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("AllCRDs does not contain the CRD defined in %q", name)
		}
	}
}
