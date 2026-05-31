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
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/pflag"
	"k8s.io/autoscaler/vertical-pod-autoscaler/common"
)

// resetFlags clears the global flag registries between t.Run test blocks.
func resetFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
}

func Test_InitRecommenderFlags(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		setupExpected func(*RecommenderConfig) // Directly uses the struct without config. prefix
	}{
		{
			name: "base defaults",
			args: []string{},
			setupExpected: func(c *RecommenderConfig) {
				// Base case expects exactly what DefaultRecommenderConfig() produces
			},
		},
		{
			name: "custom recommender name and address",
			args: []string{
				"--recommender-name=custom-vpa-engine",
				"--address=:9090",
			},
			setupExpected: func(c *RecommenderConfig) {
				c.RecommenderName = "custom-vpa-engine"
				c.Address = ":9090"
			},
		},
		{
			name: "custom cluster flags configuration",
			args: []string{
				"--cluster-name=production-main",
				"--cluster-name-label=k8s_cluster_id",
			},
			setupExpected: func(c *RecommenderConfig) {
				c.ClusterName = "production-main"
				c.ClusterNameLabel = "k8s_cluster_id"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetFlags()

			os.Args = append([]string{"vpa-recommender-binary"}, tt.args...)

			expectedConfig := DefaultRecommenderConfig()
			expectedConfig.CommonFlags = common.DefaultCommonConfig()

			if tt.setupExpected != nil {
				tt.setupExpected(expectedConfig)
			}

			actualConfig := InitRecommenderFlags()

			if diff := cmp.Diff(expectedConfig, actualConfig); diff != "" {
				t.Errorf("InitRecommenderFlags mismatch (-expected +actual):\n%s", diff)
			}
		})
	}
}

func TestValidateRecommenderConfig(t *testing.T) {
	tests := []struct {
		name        string
		setupInput  func(*RecommenderConfig)
		expectError bool
	}{
		{
			name: "Cluster Name Label and Cluster Name, both not set",
			setupInput: func(c *RecommenderConfig) {
				c.ClusterName = ""
				c.ClusterNameLabel = ""
			},
			expectError: false,
		},
		{
			name: "Cluster Name Label and Cluster Name, both are set",
			setupInput: func(c *RecommenderConfig) {
				c.ClusterName = "prod-cluster"
				c.ClusterNameLabel = "cluster_id"
			},
			expectError: false,
		},
		{
			name: "Cluster Name Label missing but Cluster Name is set",
			setupInput: func(c *RecommenderConfig) {
				c.ClusterName = "prod-cluster"
				c.ClusterNameLabel = ""
			},
			expectError: true,
		},
		{
			name: "Cluster Name Label is set but Cluster Name is missing",
			setupInput: func(c *RecommenderConfig) {
				c.ClusterName = ""
				c.ClusterNameLabel = "cluster_id"
			},
			expectError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			inputConfig := DefaultRecommenderConfig()
			inputConfig.CommonFlags = common.DefaultCommonConfig()

			if tt.setupInput != nil {
				tt.setupInput(inputConfig)
			}

			err := ValidateRecommenderConfig(inputConfig)

			hasError := (err != nil)
			if hasError != tt.expectError {
				t.Errorf("Validation mismatch for %s: expected error status %v, got error: %v",
					tt.name, tt.expectError, err)
			}
		})
	}
}

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
