/*
Copyright 2018 The Kubernetes Authors.

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

package azure

import (
	"os"
	"testing"

	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/stretchr/testify/assert"
	azclient "sigs.k8s.io/cloud-provider-azure/pkg/azclient"
	providerazure "sigs.k8s.io/cloud-provider-azure/pkg/provider"
	providerazureconfig "sigs.k8s.io/cloud-provider-azure/pkg/provider/config"
)

func TestGetServicePrincipalTokenFromCertificate(t *testing.T) {
	config := &Config{
		Config: providerazure.Config{
			AzureAuthConfig: providerazureconfig.AzureAuthConfig{
				ARMClientConfig: azclient.ARMClientConfig{
					TenantID: "TenantID",
				},
				AzureAuthConfig: azclient.AzureAuthConfig{
					AADClientID:           "AADClientID",
					AADClientCertPath:     "./testdata/test.pfx",
					AADClientCertPassword: "id",
				},
			},
		},
	}
	env := &azure.PublicCloud
	token, err := providerazureconfig.GetServicePrincipalToken(&config.AzureAuthConfig, env, "")
	assert.NoError(t, err)

	oauthConfig, err := adal.NewOAuthConfig(env.ActiveDirectoryEndpoint, config.TenantID)
	assert.NoError(t, err)
	pfxContent, err := os.ReadFile("./testdata/test.pfx")
	assert.NoError(t, err)
	certificate, privateKey, err := adal.DecodePfxCertificateData(pfxContent, "id")
	assert.NoError(t, err)
	spt, err := adal.NewServicePrincipalTokenFromCertificate(
		*oauthConfig, config.AADClientID, certificate, privateKey, env.ServiceManagementEndpoint)
	assert.NoError(t, err)
	assert.Equal(t, token.Token(), spt.Token())
}

func TestGetServicePrincipalTokenFromCertificateWithoutPassword(t *testing.T) {
	config := &Config{
		Config: providerazure.Config{
			AzureAuthConfig: providerazureconfig.AzureAuthConfig{
				ARMClientConfig: azclient.ARMClientConfig{
					TenantID: "TenantID",
				},
				AzureAuthConfig: azclient.AzureAuthConfig{
					AADClientID:       "AADClientID",
					AADClientCertPath: "./testdata/testnopassword.pfx",
				},
			},
		},
	}
	env := &azure.PublicCloud
	token, err := providerazureconfig.GetServicePrincipalToken(&config.AzureAuthConfig, env, "")
	assert.NoError(t, err)

	oauthConfig, err := adal.NewOAuthConfig(env.ActiveDirectoryEndpoint, config.TenantID)
	assert.NoError(t, err)
	pfxContent, err := os.ReadFile("./testdata/testnopassword.pfx")
	assert.NoError(t, err)
	certificate, privateKey, err := adal.DecodePfxCertificateData(pfxContent, "")
	assert.NoError(t, err)
	spt, err := adal.NewServicePrincipalTokenFromCertificate(
		*oauthConfig, config.AADClientID, certificate, privateKey, env.ServiceManagementEndpoint)
	assert.NoError(t, err)
	assert.Equal(t, token.Token(), spt.Token())
}
