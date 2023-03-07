// Copyright (c) 2016, 2018, 2022, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.

package auth

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v55/common"
	"net/http"
	"strings"
	"testing"
)

type fakeHTTPCaller struct {
	Body            string
	validateRequest func(response *http.Request) error
}

func (f fakeHTTPCaller) Do(req *http.Request) (*http.Response, error) {
	if f.validateRequest != nil && f.validateRequest(req) != nil {
		return nil, f.validateRequest(req)
	}

	response := http.Response{}
	response.Body = ioutil.NopCloser(bytes.NewBufferString(f.Body))
	return &response, nil
}

func fakeInstanceProvider(region common.Region, tenancyID string) (*instancePrincipalConfigurationProvider, error) {

	modifier := func(dispatcher common.HTTPRequestDispatcher) (common.HTTPRequestDispatcher, error) {
		body := fmt.Sprintf(`{"token":"%s"}`, expectedSecurityToken)
		return &fakeHTTPCaller{Body: body}, nil
	}

	passPhrase := []byte("")
	intermediateCerts := [][]byte{[]byte(intermediateCertPem)}

	fedClient, err :=
		newX509FederationClientWithCerts(region, tenancyID,
			[]byte(leafCertPem), passPhrase, []byte(leafCertPrivateKeyPem), intermediateCerts, *newDispatcherModifier(modifier), "")
	if err != nil {
		return nil, err
	}

	provider := &instancePrincipalConfigurationProvider{
		keyProvider: instancePrincipalKeyProvider{
			Region:           region,
			FederationClient: fedClient,
			TenancyID:        tenancyID,
		},
		region: &region,
	}
	return provider, nil
}

func TestNewResourcePrincipalConfigurationProvider(t *testing.T) {
	testRegion := common.RegionFRA
	tenancyID := tenancyID
	instanceID := "sdf"

	instanceProvider, e := fakeInstanceProvider(testRegion, tenancyID)
	assert.NoError(t, e)

	rpTkClient, e := common.NewClientWithConfig(instanceProvider)
	assert.NoError(t, e)
	rpTkClient.BasePath = "/some/path"
	rpTkClient.Host = "https://somehost"
	rpTkClient.HTTPClient = fakeHTTPCaller{
		Body: `{"resourcePrincipalToken": "T1","servicePrincipalSessionToken":"S1"}`,
	}

	rpSessionClient, e := common.NewClientWithConfig(instanceProvider)
	assert.NoError(t, e)
	rpSessionClient.BasePath = identityResourcePrincipalSessionTokenPath
	rpSessionClient.Host = "https://someotherhost"
	rpSessionClient.HTTPClient = fakeHTTPCaller{
		Body: fmt.Sprintf(`{"token":"%s"}`, expectedSecurityToken),
		validateRequest: func(req *http.Request) error {
			if !strings.Contains(req.URL.Path, identityResourcePrincipalSessionTokenPath) {
				return fmt.Errorf("request path: %v needs to contain: %v", req.URL.Path, identityResourcePrincipalSessionTokenPath)
			}
			return nil
		},
	}

	provider, e := resourcePrincipalConfigurationProviderForInstanceWithClients(*instanceProvider, rpTkClient, rpSessionClient, instanceID, identityResourcePrincipalSessionTokenPath)
	assert.NoError(t, e)
	assert.NotNil(t, provider)

	s, e := provider.KeyID()
	assert.NoError(t, e)
	assert.Equal(t, "ST$"+expectedSecurityToken, s)
}

func TestNewServicePrincipalConfigurationProvider(t *testing.T) {
	t.Skip("Skipping integration test for resource principals as it requires service support")

	//Set endpoints in for resource principal
	resourcePrincipalTokenEndpoint := "https://someservice.region/endpoint"
	resourcePrincipalTokenSessionEndpoint := "https://auth.someregion.oraclecloud.com"

	//Create an instance principal
	instancePrincipalProvider, err := InstancePrincipalConfigurationProvider()
	assert.NoError(t, err)

	//Define an interceptor(optional)
	//an interceptor is call back made just before signing and it can
	//be used to apply arbitrary transformations to the request
	interceptor := func(request *http.Request) error {
		request.URL.Query().Set("param1", "value1")
		return nil
	}

	provider, err := ResourcePrincipalConfigurationProviderWithInterceptor(instancePrincipalProvider, resourcePrincipalTokenEndpoint,
		resourcePrincipalTokenSessionEndpoint, interceptor)

	assert.NoError(t, err)
	valid, err := common.IsConfigurationProviderValid(provider)
	assert.NoError(t, err)
	assert.True(t, valid)

	// Finally, use the provider to create a client
	// client, _ := containerengine.NewContainerEngineClientWithConfigurationProvider(provider)

}
