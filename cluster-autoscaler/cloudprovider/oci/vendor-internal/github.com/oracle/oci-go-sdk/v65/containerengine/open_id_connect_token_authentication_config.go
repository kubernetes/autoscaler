// Copyright (c) 2016, 2018, 2025, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Kubernetes Engine API
//
// API for the Kubernetes Engine service (also known as the Container Engine for Kubernetes service). Use this API to build, deploy,
// and manage cloud-native applications. For more information, see
// Overview of Kubernetes Engine (https://docs.oracle.com/iaas/Content/ContEng/Concepts/contengoverview.htm).
//

package containerengine

import (
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	"strings"
)

// OpenIdConnectTokenAuthenticationConfig The properties that configure OIDC token authentication in kube-apiserver.
// For more information, see Configuring the API Server (https://kubernetes.io/docs/reference/access-authn-authz/authentication/#using-flags).
type OpenIdConnectTokenAuthenticationConfig struct {

	// Whether the cluster has OIDC Auth Config enabled. Defaults to false.
	IsOpenIdConnectAuthEnabled *bool `mandatory:"true" json:"isOpenIdConnectAuthEnabled"`

	// URL of the provider that allows the API server to discover public signing keys.
	// Only URLs that use the https:// scheme are accepted. This is typically the provider's discovery URL,
	// changed to have an empty path.
	IssuerUrl *string `mandatory:"false" json:"issuerUrl"`

	// A client id that all tokens must be issued for.
	ClientId *string `mandatory:"false" json:"clientId"`

	// JWT claim to use as the user name. By default sub, which is expected to be a unique identifier of the end
	// user. Admins can choose other claims, such as email or name, depending on their provider. However, claims
	// other than email will be prefixed with the issuer URL to prevent naming clashes with other plugins.
	UsernameClaim *string `mandatory:"false" json:"usernameClaim"`

	// Prefix prepended to username claims to prevent clashes with existing names (such as system:users).
	// For example, the value oidc: will create usernames like oidc:jane.doe. If this flag isn't provided and
	// --oidc-username-claim is a value other than email the prefix defaults to ( Issuer URL )# where
	// ( Issuer URL ) is the value of --oidc-issuer-url. The value - can be used to disable all prefixing.
	UsernamePrefix *string `mandatory:"false" json:"usernamePrefix"`

	// JWT claim to use as the user's group. If the claim is present it must be an array of strings.
	GroupsClaim *string `mandatory:"false" json:"groupsClaim"`

	// Prefix prepended to group claims to prevent clashes with existing names (such as system:groups).
	GroupsPrefix *string `mandatory:"false" json:"groupsPrefix"`

	// A key=value pair that describes a required claim in the ID Token. If set, the claim is verified to be present
	// in the ID Token with a matching value. Repeat this flag to specify multiple claims.
	RequiredClaims []KeyValue `mandatory:"false" json:"requiredClaims"`

	// A Base64 encoded public RSA or ECDSA certificates used to signed your identity provider's web certificate.
	CaCertificate *string `mandatory:"false" json:"caCertificate"`

	// The signing algorithms accepted. Default is ["RS256"].
	SigningAlgorithms []string `mandatory:"false" json:"signingAlgorithms"`

	// A Base64 encoded string of a Kubernetes OIDC Auth Config file. More info here (https://kubernetes.io/docs/reference/access-authn-authz/authentication/#using-authentication-configuration)
	ConfigurationFile *string `mandatory:"false" json:"configurationFile"`
}

func (m OpenIdConnectTokenAuthenticationConfig) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m OpenIdConnectTokenAuthenticationConfig) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}
