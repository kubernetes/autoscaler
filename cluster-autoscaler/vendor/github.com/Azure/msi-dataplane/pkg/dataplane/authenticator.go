package dataplane

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"

	"github.com/Azure/msi-dataplane/pkg/dataplane/internal/challenge"
)

var (
	errInvalidAuthHeader = errors.New("could not parse the provided WWW-Authenticate header")
)

// Authenticating with MSI: https://eng.ms/docs/products/arm/rbac/managed_identities/msionboardinginteractionwithmsi .
func newAuthenticatorPolicy(cred azcore.TokenCredential, audience string) policy.Policy {
	return runtime.NewBearerTokenPolicy(cred, nil, &policy.BearerTokenOptions{
		AuthorizationHandler: policy.AuthorizationHandler{
			// Make an unauthenticated request
			OnRequest: func(*policy.Request, func(policy.TokenRequestOptions) error) error {
				return nil
			},
			// Inspect WWW-Authenticate header returned from challenge
			OnChallenge: func(req *policy.Request, resp *http.Response, authenticateAndAuthorize func(policy.TokenRequestOptions) error) error {
				// we expect 'Bearer authorization="https://login.windows-ppe.net/5D929AE3-B37C-46AA-A3C8-C1558902F101"'
				authParam, err := parseChallengeHeader(resp.Header)
				if err != nil {
					return err
				}

				u, err := url.Parse(authParam)
				if err != nil {
					return fmt.Errorf("%w: %w", errInvalidAuthHeader, err)
				}
				tenantID := strings.ToLower(strings.Trim(u.Path, "/"))

				req.Raw().Context()

				// Note: "In api versions prior to 2023-09-30, the audience is included in the bearer challenge, but we recommend that partners
				// rely on hard-configuring the explicit values above for security reasons."

				// Authenticate from tenantID and audience
				return authenticateAndAuthorize(policy.TokenRequestOptions{
					Scopes:   []string{audience + "/.default"},
					TenantID: tenantID,
				})
			},
		},
	})
}

func parseChallengeHeader(headers http.Header) (string, error) {
	challenges, err := challenge.Parse(headers)
	if err != nil {
		return "", fmt.Errorf("%w: %w", errInvalidAuthHeader, err)
	}
	if len(challenges) == 0 {
		return "", fmt.Errorf("%w: %s", errInvalidAuthHeader, "no challenges found")
	}
	var bearer *challenge.Challenge
	for _, c := range challenges {
		if c.Scheme == "Bearer" {
			bearer = &c
		}
	}
	if bearer == nil {
		return "", fmt.Errorf("%w: %s", errInvalidAuthHeader, "no bearer challenge found")
	}
	authParam, provided := bearer.Parameters["authorization"]
	if !provided {
		return "", fmt.Errorf("%w: %s", errInvalidAuthHeader, "no authorization parameter in bearer challenge")
	}
	return authParam, nil
}
