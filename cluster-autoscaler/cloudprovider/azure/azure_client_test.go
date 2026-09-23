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
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	armpolicy "github.com/Azure/azure-sdk-for-go/sdk/azcore/arm/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	azcorepolicy "github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	armcompute "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v7"
	autorestazure "github.com/Azure/go-autorest/autorest/azure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/virtualmachinescalesetclient"
)

// Note: The previous tests for GetServicePrincipalToken were removed
// because that function no longer exists in cloud-provider-azure v1.32.0+.
// The authentication mechanism has been changed to use Azure Identity SDK
// instead of the older ADAL library.

type staticTokenCredential struct{}

func (staticTokenCredential) GetToken(context.Context, azcorepolicy.TokenRequestOptions) (azcore.AccessToken, error) {
	return azcore.AccessToken{Token: "token", ExpiresOn: time.Now().Add(time.Hour)}, nil
}

type recordingTransport struct {
	request *http.Request
}

func (transport *recordingTransport) Do(request *http.Request) (*http.Response, error) {
	transport.request = request
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(`{"value":[]}`)),
		Request:    request,
	}, nil
}

func TestNewARMClientConfigControlsAzureStackVMSSDeleteAPIVersion(t *testing.T) {
	testCases := []struct {
		name                   string
		disableAzureStackCloud bool
		wantAzureStackVersion  bool
	}{
		{
			name:                  "Azure Stack API version enabled",
			wantAzureStackVersion: true,
		},
		{
			name:                   "Azure Stack API version disabled",
			disableAzureStackCloud: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			cfg := &Config{}
			cfg.Cloud = "AzureStackCloud"
			cfg.DisableAzureStackCloud = testCase.disableAzureStackCloud
			armConfig := newARMClientConfig(cfg, &autorestazure.Environment{})

			transport := &recordingTransport{}
			factory, err := azclient.NewClientFactory(
				&azclient.ClientFactoryConfig{SubscriptionID: "subscription"},
				armConfig,
				cloud.Configuration{
					ActiveDirectoryAuthorityHost: "https://login.microsoftonline.com/",
					Services: map[cloud.ServiceName]cloud.ServiceConfiguration{
						cloud.ResourceManager: {
							Endpoint: "https://management.test/",
							Audience: "https://management.test/",
						},
					},
				},
				staticTokenCredential{},
				func(options *armpolicy.ClientOptions) {
					options.Transport = transport
				},
			)
			require.NoError(t, err)

			deleteClient := NewVMSSDeleteClient(factory.GetVirtualMachineScaleSetClient())
			require.NotNil(t, deleteClient)
			instanceID := "0"
			forceDeletion := true
			_, err = deleteClient.BeginDeleteInstances(
				context.Background(),
				"resource-group",
				"scale-set",
				armcompute.VirtualMachineScaleSetVMInstanceRequiredIDs{InstanceIDs: []*string{&instanceID}},
				&armcompute.VirtualMachineScaleSetsClientBeginDeleteInstancesOptions{ForceDeletion: &forceDeletion},
			)
			require.NoError(t, err)
			require.NotNil(t, transport.request)

			assert.Equal(t, http.MethodPost, transport.request.Method)
			assert.Equal(t, "/subscriptions/subscription/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set/delete", transport.request.URL.Path)
			assert.Equal(t, "true", transport.request.URL.Query().Get("forceDeletion"))
			apiVersion := transport.request.URL.Query().Get("api-version")
			if testCase.wantAzureStackVersion {
				assert.Equal(t, virtualmachinescalesetclient.AzureStackCloudAPIVersion, apiVersion)
			} else {
				assert.NotEqual(t, virtualmachinescalesetclient.AzureStackCloudAPIVersion, apiVersion)
			}
		})
	}
}
