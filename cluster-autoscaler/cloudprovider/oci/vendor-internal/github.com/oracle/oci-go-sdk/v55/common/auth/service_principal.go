// Copyright (c) 2016, 2018, 2022, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.

package auth

import (
	"crypto/rsa"
	"fmt"
	"io/ioutil"
	"path"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v55/common"
)

type servicePrincipalKeyProvider struct {
	federationClient federationClient
}

func newServicePrincipalKeyProvider(tenancyID, region string, cert, key []byte, intermediates [][]byte, passphrase []byte, modifier func(common.HTTPRequestDispatcher) (common.HTTPRequestDispatcher, error)) (provider *servicePrincipalKeyProvider, err error) {
	clientModifier := newDispatcherModifier(modifier)

	leafCertificateRetriever := newStaticX509CertificateRetriever(cert, key, passphrase)

	var intermediateCertificateRetrievers []x509CertificateRetriever
	for _, intermediate := range intermediates {
		intermediateCertificateRetrievers =
			append(intermediateCertificateRetrievers, newStaticX509CertificateRetriever(intermediate, key, passphrase))
	}

	federationClient, err := newX509FederationClientWithPurpose(
		common.Region(region), tenancyID, leafCertificateRetriever,
		intermediateCertificateRetrievers, true, *clientModifier, defaultTokenPurpose)

	if err != nil {
		err = fmt.Errorf("failed to create federation client: %s", err.Error())
		return nil, err
	}

	provider = &servicePrincipalKeyProvider{federationClient: federationClient}
	return
}

func (p *servicePrincipalKeyProvider) PrivateRSAKey() (privateKey *rsa.PrivateKey, err error) {
	if privateKey, err = p.federationClient.PrivateKey(); err != nil {
		err = fmt.Errorf("failed to get private key: %s", err.Error())
		return nil, err
	}
	return privateKey, nil
}

func (p *servicePrincipalKeyProvider) KeyID() (string, error) {
	var securityToken string
	var err error
	if securityToken, err = p.federationClient.SecurityToken(); err != nil {
		return "", fmt.Errorf("failed to get security token: %s", err.Error())
	}

	return fmt.Sprintf("ST$%s", securityToken), nil
}

func (p *servicePrincipalKeyProvider) AuthType() (common.AuthConfig, error) {
	return common.AuthConfig{common.UnknownAuthenticationType, false, nil},
		fmt.Errorf("unsupported, keep the interface")
}

type servicePrincipalConfigurationProvider struct {
	keyProvider       *servicePrincipalKeyProvider
	tenancyID, region string
}

// NewServicePrincipalConfigurationProvider will create a new service principal configuration provider
func NewServicePrincipalConfigurationProvider(tenancyID, region string, cert, key []byte, intermediates [][]byte, passphrase []byte) (common.ConfigurationProvider, error) {
	return NewServicePrincipalConfigurationProviderWithCustomClient(nil, tenancyID, region, cert, key, intermediates, passphrase)
}

// NewServicePrincipalConfigurationProviderWithCustomClient will create a new service principal configuration provider using a modifier function to modify the HTTPRequestDispatcher
func NewServicePrincipalConfigurationProviderWithCustomClient(modifier func(common.HTTPRequestDispatcher) (common.HTTPRequestDispatcher, error), tenancyID, region string, cert, key []byte, intermediates [][]byte, passphrase []byte) (common.ConfigurationProvider, error) {
	var err error
	var keyProvider *servicePrincipalKeyProvider
	if keyProvider, err = newServicePrincipalKeyProvider(tenancyID, region, cert, key, intermediates, passphrase, modifier); err != nil {
		return nil, fmt.Errorf("failed to create a new key provider: %s", err.Error())
	}
	return servicePrincipalConfigurationProvider{keyProvider: keyProvider, region: region, tenancyID: tenancyID}, nil
}

// NewServicePrincipalWithInstancePrincipalConfigurationProvider create a S2S configuration provider by acquiring credentials via instance principals
func NewServicePrincipalWithInstancePrincipalConfigurationProvider(region common.Region) (common.ConfigurationProvider, error) {
	return newInstancePrincipalConfigurationProvider(region, servicePrincipalTokenPurpose, nil)
}

// NewServicePrincipalConfigurationWithCerts returns a configuration for service principals with a given region and hardcoded certificates in lieu of metadata service certs
func NewServicePrincipalConfigurationWithCerts(region common.Region, leafCertificate, leafPassphrase, leafPrivateKey []byte, intermediateCertificates [][]byte) (common.ConfigurationProvider, error) {
	leafCertificateRetriever := staticCertificateRetriever{Passphrase: leafPassphrase, CertificatePem: leafCertificate, PrivateKeyPem: leafPrivateKey}

	//The .Refresh() call actually reads the certificates from the inputs
	err := leafCertificateRetriever.Refresh()
	if err != nil {
		return nil, err
	}
	certificate := leafCertificateRetriever.Certificate()
	tenancyID := extractTenancyIDFromCertificate(certificate)
	fedClient, err := newX509FederationClientWithCerts(region, tenancyID, leafCertificate, leafPassphrase, leafPrivateKey, intermediateCertificates, *newDispatcherModifier(nil), "")
	if err != nil {
		return nil, err
	}
	keyProvider := servicePrincipalKeyProvider{federationClient: fedClient}
	return servicePrincipalConfigurationProvider{keyProvider: &keyProvider, region: string(region), tenancyID: tenancyID}, nil
}

// NewServicePrincipalConfigurationProviderFromHostCerts returns a configuration for service principals,
// given the region and a pathname to the host's service principal certificate directory.
// The pathname generally follows the pattern "/var/certs/hostclass/${hostclass}/${servicePrincipalName}-identity"
func NewServicePrincipalConfigurationProviderFromHostCerts(region common.Region, certDir string) (common.ConfigurationProvider, error) {
	if certDir == "" {
		return nil, fmt.Errorf("empty input string")
	}
	// Read certs from substrate host.
	leafKey, err := ioutil.ReadFile(path.Join(certDir, "key.pem"))
	if err != nil {
		return nil, fmt.Errorf("error reading leafPrivateKey")
	}
	leafCert, err := ioutil.ReadFile(path.Join(certDir, "cert.pem"))
	if err != nil {
		return nil, fmt.Errorf("error reading leafCertificate")
	}
	interCert, err := ioutil.ReadFile(path.Join(certDir, "intermediates.pem"))
	if err != nil {
		return nil, fmt.Errorf("error reading intermediateCertificate")
	}
	var interCerts [][]byte
	interCerts = append(interCerts, interCert)
	var leafPass = []byte("")
	return NewServicePrincipalConfigurationWithCerts(region, leafCert, leafPass, leafKey, interCerts)
}

func (p servicePrincipalConfigurationProvider) PrivateRSAKey() (*rsa.PrivateKey, error) {
	return p.keyProvider.PrivateRSAKey()
}

func (p servicePrincipalConfigurationProvider) KeyID() (string, error) {
	return p.keyProvider.KeyID()
}

func (p servicePrincipalConfigurationProvider) TenancyOCID() (string, error) {
	return p.tenancyID, nil
}

func (p servicePrincipalConfigurationProvider) UserOCID() (string, error) {
	return "", nil
}

func (p servicePrincipalConfigurationProvider) KeyFingerprint() (string, error) {
	return "", nil
}

func (p servicePrincipalConfigurationProvider) Region() (string, error) {
	return p.region, nil
}

func (p servicePrincipalConfigurationProvider) AuthType() (common.AuthConfig, error) {
	return common.AuthConfig{common.UnknownAuthenticationType, false, nil},
		fmt.Errorf("unsupported, keep the interface")

}
