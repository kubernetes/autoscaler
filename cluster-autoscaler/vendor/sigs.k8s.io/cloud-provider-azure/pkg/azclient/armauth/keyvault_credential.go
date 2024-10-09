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
	"encoding/json"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets"
)

type KeyVaultCredential struct {
	secretClient *azsecrets.Client
	secretPath   string

	token *azcore.AccessToken
}

type KeyVaultCredentialSecret struct {
	AccessToken string    `json:"access_token"`
	ExpiresOn   time.Time `json:"expires_on"`
}

func NewKeyVaultCredential(
	msiCredential azcore.TokenCredential,
	keyVaultURL string,
	secretName string,
) (*KeyVaultCredential, error) {
	cli, err := azsecrets.NewClient(keyVaultURL, msiCredential, nil)
	if err != nil {
		return nil, fmt.Errorf("create KeyVault client: %w", err)
	}

	rv := &KeyVaultCredential{
		secretClient: cli,
		secretPath:   secretName,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := rv.refreshToken(ctx); err != nil {
		return nil, fmt.Errorf("refresh token: %w", err)
	}

	return rv, nil
}

func (c *KeyVaultCredential) refreshToken(ctx context.Context) error {
	const LatestVersion = ""

	resp, err := c.secretClient.GetSecret(ctx, c.secretPath, LatestVersion, nil)
	if err != nil {
		return err
	}
	if resp.Value == nil {
		return fmt.Errorf("secret value is nil")
	}

	var secret KeyVaultCredentialSecret
	if err := json.Unmarshal([]byte(*resp.Value), &secret); err != nil {
		return fmt.Errorf("unmarshal secret value `%s`: %w", *resp.Value, err)
	}

	c.token = &azcore.AccessToken{
		Token:     secret.AccessToken,
		ExpiresOn: secret.ExpiresOn,
	}

	return nil
}

func (c *KeyVaultCredential) GetToken(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
	const RefreshTokenOffset = 5 * time.Minute

	if c.token != nil && c.token.ExpiresOn.Add(RefreshTokenOffset).Before(time.Now()) {
		return *c.token, nil
	}

	if err := c.refreshToken(ctx); err != nil {
		return azcore.AccessToken{}, fmt.Errorf("refresh token: %w", err)
	}

	return *c.token, nil
}
