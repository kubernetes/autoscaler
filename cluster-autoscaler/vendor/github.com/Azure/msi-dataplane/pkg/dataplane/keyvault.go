package dataplane

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
)

func ptrTo[o any](s o) *o {
	return &s
}

// IdentifierForManagedIdentityCredentials creates a canonical identifier for a KeyVault item, labelling the
// item as storing managed identity credentials.
func IdentifierForManagedIdentityCredentials(identifier string) string {
	return ManagedIdentityCredentialsStoragePrefix + identifier
}

// IdentifierForUserAssignedIdentityCredentials creates a canonical identifier for a KeyVault item, labelling the
// item as storing user-assigned managed identity credentials.
func IdentifierForUserAssignedIdentityCredentials(identifier string) string {
	return UserAssignedIdentityCredentialsStoragePrefix + identifier
}

// FormatManagedIdentityCredentialsForStorage provides the canonical KeyVault secret parameters for storing
// managed identity credentials, ensuring that appropriate times are recorded for the expiry and notBefore,
// as well as that renewal times are recorded in tags.
func FormatManagedIdentityCredentialsForStorage(identifier string, credentials ManagedIdentityCredentials) (string, azsecrets.SetSecretParameters, error) {
	var rawNotAfter, rawNotBefore, rawRenewAfter, rawCannotRenewAfter *string
	switch len(credentials.ExplicitIdentities) {
	case 0:
		rawNotAfter = credentials.NotAfter
		rawNotBefore = credentials.NotBefore
		rawRenewAfter = credentials.RenewAfter
		rawCannotRenewAfter = credentials.CannotRenewAfter
	case 1:
		rawNotAfter = credentials.ExplicitIdentities[0].NotAfter
		rawNotBefore = credentials.ExplicitIdentities[0].NotBefore
		rawRenewAfter = credentials.ExplicitIdentities[0].RenewAfter
		rawCannotRenewAfter = credentials.ExplicitIdentities[0].CannotRenewAfter
	default:
		return "", azsecrets.SetSecretParameters{}, fmt.Errorf("assumption violated, found %d explicit identities, expected none, or one", len(credentials.ExplicitIdentities))
	}

	parameters, err := keyVaultParameters(credentials, rawNotAfter, rawNotBefore, rawRenewAfter, rawCannotRenewAfter)
	if err != nil {
		return "", azsecrets.SetSecretParameters{}, err
	}

	return IdentifierForManagedIdentityCredentials(identifier), parameters, nil
}

func keyVaultParameters(credentials any, rawNotAfter, rawNotBefore, rawRenewAfter, rawCannotRenewAfter *string) (azsecrets.SetSecretParameters, error) {
	for key, value := range map[string]*string{
		"NotAfter":         rawNotAfter,
		"NotBefore":        rawNotBefore,
		"RenewAfter":       rawRenewAfter,
		"CannotRenewAfter": rawCannotRenewAfter,
	} {
		if value == nil {
			return azsecrets.SetSecretParameters{}, fmt.Errorf("assumption violated, %q was nil", key)
		}
	}

	var notAfter, notBefore time.Time
	for from, to := range map[*string]*time.Time{
		rawNotAfter:  &notAfter,
		rawNotBefore: &notBefore,
	} {
		value, err := time.Parse(time.RFC3339, *from)
		if err != nil {
			return azsecrets.SetSecretParameters{}, err
		}
		*to = value
	}

	raw, err := json.Marshal(credentials)
	if err != nil {
		return azsecrets.SetSecretParameters{}, fmt.Errorf("failed to marshal credentials: %v", err)
	}

	return azsecrets.SetSecretParameters{
		Value: ptrTo(string(raw)),
		SecretAttributes: &azsecrets.SecretAttributes{
			Enabled:   ptrTo(true),
			Expires:   &notAfter,
			NotBefore: &notBefore,
		},
		Tags: map[string]*string{
			RenewAfterKeyVaultTag:       rawRenewAfter,
			CannotRenewAfterKeyVaultTag: rawCannotRenewAfter,
		},
	}, nil
}

// FormatUserAssignedIdentityCredentialsForStorage provides the canonical KeyVault secret parameters for storing
// user-assigned managed identity credentials, ensuring that appropriate times are recorded for the expiry and
// notBefore, as well as that renewal times are recorded in tags.
func FormatUserAssignedIdentityCredentialsForStorage(identifier string, credentials UserAssignedIdentityCredentials) (string, azsecrets.SetSecretParameters, error) {
	parameters, err := keyVaultParameters(credentials, credentials.NotAfter, credentials.NotBefore, credentials.RenewAfter, credentials.CannotRenewAfter)
	if err != nil {
		return "", azsecrets.SetSecretParameters{}, err
	}

	return IdentifierForUserAssignedIdentityCredentials(identifier), parameters, nil
}
