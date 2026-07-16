package dataplane

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

var (
	errDecodeClientSecret = errors.New("failed to decode client secret")
	errParseCertificate   = errors.New("failed to parse certificate")
	errNilField           = errors.New("expected non nil field in identity")
)

// Get an AzIdentity credential for the given nested credential object
// Clients can use the credential to get a token for the user-assigned identity
func GetCredential(clientOpts azcore.ClientOptions, credential UserAssignedIdentityCredentials) (*azidentity.ClientCertificateCredential, error) {
	// Double check nil pointers so we don't panic
	fieldsToCheck := map[string]*string{
		"clientID":               credential.ClientID,
		"tenantID":               credential.TenantID,
		"clientSecret":           credential.ClientSecret,
		"authenticationEndpoint": credential.AuthenticationEndpoint,
	}
	missing := make([]string, 0)
	for field, val := range fieldsToCheck {
		if val == nil {
			missing = append(missing, field)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("%w: %s", errNilField, strings.Join(missing, ","))
	}

	opts := &azidentity.ClientCertificateCredentialOptions{
		ClientOptions: clientOpts,

		// x5c header required: https://eng.ms/docs/products/arm/rbac/managed_identities/msionboardingrequestingatoken
		SendCertificateChain: true,

		// Disable instance discovery because MSI credential may have regional AAD endpoint that instance discovery endpoint doesn't support
		// e.g. when MSI credential has westus2.logicredential.microsoft.com, it will cause instance discovery to fail with HTTP 400
		DisableInstanceDiscovery: true,
	}

	// Set the regional AAD endpoint
	// https://eng.ms/docs/products/arm/rbac/managed_identities/msionboardingcredentialapiversion2019-08-31
	opts.Cloud.ActiveDirectoryAuthorityHost = *credential.AuthenticationEndpoint

	// Parse the certificate and private key from the base64 encoded secret
	decodedSecret, err := base64.StdEncoding.DecodeString(*credential.ClientSecret)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errDecodeClientSecret, err)
	}
	// Note - ParseCertificates does not currently support pkcs12 SHA256 MAC certs, so if
	// managed identity team changes the cert format, double check this code
	crt, key, err := azidentity.ParseCertificates(decodedSecret, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errParseCertificate, err)
	}
	return azidentity.NewClientCertificateCredential(*credential.TenantID, *credential.ClientID, crt, key, opts)
}
