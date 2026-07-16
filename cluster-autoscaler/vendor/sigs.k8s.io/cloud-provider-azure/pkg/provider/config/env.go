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

package config

import (
	"io"
	"os"
	"strings"

	"sigs.k8s.io/yaml"
)

// ParseConfig returns a parsed configuration for an Azure cloudprovider config file
func ParseConfig(configReader io.Reader) (*Config, error) {
	var config Config
	if configReader == nil {
		return nil, nil
	}

	configContents, err := io.ReadAll(configReader)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(configContents, &config)
	if err != nil {
		return nil, err
	}

	// The resource group name may be in different cases from different Azure APIs, hence it is converted to lower here.
	// See more context at https://github.com/kubernetes/kubernetes/issues/71994.
	config.ResourceGroup = strings.ToLower(config.ResourceGroup)

	// these environment variables are injected by workload identity webhook
	if tenantID := os.Getenv("AZURE_TENANT_ID"); tenantID != "" {
		config.TenantID = tenantID
	}
	if clientID := os.Getenv("AZURE_CLIENT_ID"); clientID != "" {
		config.AADClientID = clientID
	}
	if federatedTokenFile := os.Getenv("AZURE_FEDERATED_TOKEN_FILE"); federatedTokenFile != "" {
		config.AADFederatedTokenFile = federatedTokenFile
		config.UseFederatedWorkloadIdentityExtension = true
	}
	return &config, nil
}
