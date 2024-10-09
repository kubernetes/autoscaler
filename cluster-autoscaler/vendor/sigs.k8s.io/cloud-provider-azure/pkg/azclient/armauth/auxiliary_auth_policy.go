/*
Copyright 2024 The Kubernetes Authors.

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

package armauth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
)

const (
	HeaderAuthorizationAuxiliary = "x-ms-authorization-auxiliary"
)

type AuxiliaryAuthPolicy struct {
	credentials []azcore.TokenCredential
	scope       string
}

func NewAuxiliaryAuthPolicy(credentials []azcore.TokenCredential, scope string) *AuxiliaryAuthPolicy {
	return &AuxiliaryAuthPolicy{
		credentials: credentials,
		scope:       scope,
	}
}

func (p *AuxiliaryAuthPolicy) Do(req *policy.Request) (*http.Response, error) {
	tokens := make([]string, 0, len(p.credentials))

	for _, cred := range p.credentials {
		token, err := cred.GetToken(context.TODO(), policy.TokenRequestOptions{
			Scopes: []string{p.scope},
		})
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, fmt.Sprintf("Bearer %s", token.Token))
	}
	req.Raw().Header.Set(HeaderAuthorizationAuxiliary, strings.Join(tokens, ", "))
	return req.Next()
}
