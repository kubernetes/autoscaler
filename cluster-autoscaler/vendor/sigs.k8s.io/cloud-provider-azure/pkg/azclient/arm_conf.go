/*
Copyright 2023 The Kubernetes Authors.

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

package azclient

import (
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/utils"
)

type ARMClientConfig struct {
	// The cloud environment identifier. Takes values from https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/azcore@v1.6.0/cloud
	Cloud string `json:"cloud,omitempty" yaml:"cloud,omitempty"`
	// The user agent for Azure customer usage attribution
	UserAgent string `json:"userAgent,omitempty" yaml:"userAgent,omitempty"`
	// ResourceManagerEndpoint is the cloud's resource manager endpoint. If set, cloud provider queries this endpoint
	// in order to generate an autorest.Environment instance instead of using one of the pre-defined Environments.
	ResourceManagerEndpoint string `json:"resourceManagerEndpoint,omitempty" yaml:"resourceManagerEndpoint,omitempty"`
	// The AAD Tenant ID for the Subscription that the cluster is deployed in
	TenantID string `json:"tenantId,omitempty" yaml:"tenantId,omitempty"`
	// The AAD Tenant ID for the Subscription that the network resources are deployed in.
	NetworkResourceTenantID string `json:"networkResourceTenantID,omitempty" yaml:"networkResourceTenantID,omitempty"`
}

func (config *ARMClientConfig) GetTenantID() string {
	// these environment variables are injected by workload identity webhook
	if tenantID := os.Getenv(utils.AzureTenantID); tenantID != "" {
		return tenantID
	}
	return config.TenantID
}

func GetAzCoreClientOption(armConfig *ARMClientConfig) (*policy.ClientOptions, error) {
	//Get default settings
	azCoreClientConfig := utils.GetDefaultAzCoreClientOption()
	if armConfig != nil {
		//update user agent header
		azCoreClientConfig.Telemetry.ApplicationID = strings.TrimSpace(armConfig.UserAgent)
		//set cloud
		cloudConfig, err := GetAzureCloudConfig(armConfig)
		if err != nil {
			return nil, err
		}
		azCoreClientConfig.Cloud = *cloudConfig
	}
	return &azCoreClientConfig, nil
}

func IsMultiTenant(armConfig *ARMClientConfig) bool {
	return armConfig != nil && armConfig.NetworkResourceTenantID != "" && !strings.EqualFold(armConfig.NetworkResourceTenantID, armConfig.GetTenantID())
}
