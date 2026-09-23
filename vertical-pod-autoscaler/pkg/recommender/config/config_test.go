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

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateRecommenderConfigLoadsBearerTokenFromFile(t *testing.T) {
	t.Parallel()

	config := DefaultRecommenderConfig()
	tokenFilePath := filepath.Join(t.TempDir(), "token")
	if err := os.WriteFile(tokenFilePath, []byte("\nmy-secret-token\n"), 0600); err != nil {
		t.Fatalf("failed to create token file: %v", err)
	}
	config.PrometheusBearerTokenFile = tokenFilePath

	ValidateRecommenderConfig(config)

	if got, want := config.PrometheusBearerToken, "my-secret-token"; got != want {
		t.Fatalf("PrometheusBearerToken = %q, want %q", got, want)
	}
}
