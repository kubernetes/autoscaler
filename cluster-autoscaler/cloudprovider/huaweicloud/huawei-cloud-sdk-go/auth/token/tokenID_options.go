/*
Copyright 2020 The Kubernetes Authors.

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

package token

// TokenIdOptions presents the required information for token ID auth
type TokenIdOptions struct {
	// IdentityEndpoint specifies the HTTP endpoint that is required to work with
	// the Identity API of the appropriate version. While it's ultimately needed by
	// all of the identity services, it will often be populated by a provider-level
	// function.
	//
	// The IdentityEndpoint is typically referred to as the "auth_url" or
	// "OS_AUTH_URL" in the information provided by the cloud operator.
	IdentityEndpoint string `json:"-" required:"true"`

	// AuthToken allows users to authenticate (possibly as another user) with an
	// authentication token ID.
	AuthToken string `json:"-"`

	// user project id
	ProjectID string

	DomainID string `json:"-" required:"true"`
}

// GetIdentityEndpoint Implements the method of AuthOptionsProvider
func (opts TokenIdOptions) GetIdentityEndpoint() string {
	return opts.IdentityEndpoint
}

//GetProjectId Implements the method of AuthOptionsProvider
func (opts TokenIdOptions) GetProjectId() string {
	return opts.ProjectID
}

// GetDomainId Implements the method of AuthOptionsProvider
func (opts TokenIdOptions) GetDomainId() string {
	return opts.DomainID
}
