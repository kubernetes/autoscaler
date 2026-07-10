/*
Copyright 2023 The Kubernetes Authors.

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

package useragent

import (
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
)

type CustomUserAgentPolicy struct {
	CustomUserAgent string
}

const HeaderUserAgent = "User-Agent"

func NewCustomUserAgentPolicy(customUserAgent string) policy.Policy {
	return &CustomUserAgentPolicy{
		CustomUserAgent: customUserAgent,
	}
}

func (p CustomUserAgentPolicy) Do(req *policy.Request) (*http.Response, error) {
	if p.CustomUserAgent == "" {
		return req.Next()
	}
	// preserve the existing User-Agent string
	if ua := req.Raw().Header.Get(HeaderUserAgent); ua == "" {
		req.Raw().Header.Set(HeaderUserAgent, p.CustomUserAgent)
	}
	return req.Next()
}
