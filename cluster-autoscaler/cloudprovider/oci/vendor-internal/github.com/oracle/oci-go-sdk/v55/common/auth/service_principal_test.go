// Copyright (c) 2016, 2018, 2022, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.

package auth

import (
	"crypto/rsa"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v55/common"
)

func TestServicePrincipalKeyProvider(t *testing.T) {
	key, _ := common.PrivateKeyFromBytes([]byte(leafCertPrivateKeyPem), nil)
	testIO := []struct {
		name               string
		tenancy, region    string
		cert, key          []byte
		intermediates      [][]byte
		passphrase         []byte
		expectErrorKey     error
		expectErrorToken   error
		expectedPrivateKey *rsa.PrivateKey
		expectSecToken     string
	}{
		{
			name:               "Should crate a service principal with no failure",
			tenancy:            tenancyID,
			region:             "anyRegion",
			cert:               []byte(leafCertPem),
			key:                []byte(leafCertPrivateKeyPem),
			intermediates:      [][]byte{[]byte(intermediateCertPem)},
			passphrase:         nil,
			expectedPrivateKey: key,
			expectSecToken:     "token",
		},
		{
			name:               "Should create a service principal even if, skipping tenancy verification",
			tenancy:            "random ocid",
			region:             "anyRegion",
			cert:               []byte(leafCertPem),
			key:                []byte(leafCertPrivateKeyPem),
			intermediates:      [][]byte{[]byte(intermediateCertPem)},
			passphrase:         nil,
			expectedPrivateKey: key,
			expectSecToken:     "token",
		},
		{
			name:               "Should create fail if there is an error returning the sec token",
			tenancy:            "random ocid",
			region:             "anyRegion",
			cert:               []byte(leafCertPem),
			key:                []byte(leafCertPrivateKeyPem),
			intermediates:      [][]byte{[]byte(intermediateCertPem)},
			passphrase:         nil,
			expectedPrivateKey: key,
			expectErrorToken:   assert.AnError,
		},
		{
			name:           "Should create fail if there is an error returning private key",
			tenancy:        "random ocid",
			region:         "anyRegion",
			cert:           []byte(leafCertPem),
			key:            []byte(leafCertPrivateKeyPem),
			intermediates:  [][]byte{[]byte(intermediateCertPem)},
			passphrase:     nil,
			expectErrorKey: assert.AnError,
		},
	}
	for _, test := range testIO {
		t.Run(test.name, func(t *testing.T) {
			mockFederationClient := new(mockFederationClient)
			mockFederationClient.On("PrivateKey").Return(test.expectedPrivateKey, test.expectErrorKey).Once()
			mockFederationClient.On("SecurityToken").Return(test.expectSecToken, test.expectErrorToken).Once()

			keyProvider, _ := newServicePrincipalKeyProvider(test.tenancy,
				test.region, test.cert, test.key, test.intermediates, test.passphrase, nil)

			keyProvider.federationClient = mockFederationClient
			actualPrivateKey, errKey := keyProvider.PrivateRSAKey()
			actualToken, errToken := keyProvider.KeyID()

			if test.expectErrorKey != nil {
				assert.Equal(t, fmt.Errorf("failed to get private key: %s", test.expectErrorKey.Error()), errKey)
			}
			if test.expectErrorToken != nil {
				assert.Equal(t, fmt.Errorf("failed to get security token: %s", test.expectErrorToken.Error()), errToken)
			}
			assert.Equal(t, test.expectedPrivateKey, actualPrivateKey)
			if test.expectSecToken != "" {
				assert.Equal(t, "ST$"+test.expectSecToken, actualToken)
			}
			mockFederationClient.AssertExpectations(t)
		})
	}
}
