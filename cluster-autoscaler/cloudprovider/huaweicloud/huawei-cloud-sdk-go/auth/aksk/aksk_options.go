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

package aksk

// AKSKOptions presents the required information for AK/SK auth
type AKSKOptions struct {
	IdentityEndpoint string `json:"-" required:"true"` // HTTP endpoint for identity API.
	ProjectID        string // user project id
	DomainID         string `json:"-" required:"true"` // Huawei cloud account id
	Region           string // eg. "cn-north-1" for "Beijing1", "cn-north-4" for "Beijing4"
	Domain           string // Cloud name
	Cloud            string // Cloud name
	AccessKey        string // Access Key
	SecretKey        string // Secret key
	SecurityToken    string // If AK/SK is used, token won't be used.
}

// GetIdentityEndpoint implements the method of AuthOptionsProvider
func (opts AKSKOptions) GetIdentityEndpoint() string {
	return opts.IdentityEndpoint
}

//GetProjectId implements the method of AuthOptionsProvider
func (opts AKSKOptions) GetProjectId() string {
	return opts.ProjectID
}

// GetDomainId implements the method of AuthOptionsProvider
func (opts AKSKOptions) GetDomainId() string {
	return opts.DomainID
}
