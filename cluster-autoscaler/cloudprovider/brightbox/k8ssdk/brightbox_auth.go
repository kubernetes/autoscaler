// Copyright 2020 Brightbox Systems Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package k8ssdk

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/brightbox/k8ssdk/cached"
	klog "k8s.io/klog/v2"
)

const (
	defaultClientID     = "app-dkmch"
	defaultClientSecret = "uogoelzgt0nwawb"
	clientEnvVar        = "BRIGHTBOX_CLIENT"
	clientSecretEnvVar  = "BRIGHTBOX_CLIENT_SECRET"
	usernameEnvVar      = "BRIGHTBOX_USER_NAME"
	passwordEnvVar      = "BRIGHTBOX_PASSWORD"
	accountEnvVar       = "BRIGHTBOX_ACCOUNT"
	apiURLEnvVar        = "BRIGHTBOX_API_URL"

	defaultTimeoutSeconds = 10

	ValidAcmeDomainStatus = "valid"
)

var infrastructureScope = []string{"infrastructure"}

type authdetails struct {
	APIClient string
	APISecret string
	UserName  string
	password  string
	Account   string
	APIURL    string
}

// obtainCloudClient creates a new Brightbox client using details from
// the environment
func obtainCloudClient() (CloudAccess, error) {
	klog.V(4).Infof("obtainCloudClient")
	config := &authdetails{
		APIClient: getenvWithDefault(clientEnvVar,
			defaultClientID),
		APISecret: getenvWithDefault(clientSecretEnvVar,
			defaultClientSecret),
		UserName: os.Getenv(usernameEnvVar),
		password: os.Getenv(passwordEnvVar),
		Account:  os.Getenv(accountEnvVar),
		APIURL:   os.Getenv(apiURLEnvVar),
	}
	err := config.validateConfig()
	if err != nil {
		return nil, err
	}
	return config.authenticatedClient()
}

// Validate account config entries
func (authd *authdetails) validateConfig() error {
	klog.V(4).Infof("validateConfig")
	if authd.APIClient == defaultClientID &&
		authd.APISecret == defaultClientSecret {
		if authd.Account == "" {
			return fmt.Errorf("must specify Account with User Credentials")
		}
	} else {
		if authd.UserName != "" || authd.password != "" {
			return fmt.Errorf("User Credentials not used with API Client")
		}
	}
	return nil
}

// Authenticate the details and return a client
func (authd *authdetails) authenticatedClient() (CloudAccess, error) {
	ctx := context.Background()
	switch {
	case authd.UserName != "" || authd.password != "":
		return authd.tokenisedAuth(ctx)
	default:
		return authd.apiClientAuth(ctx)
	}
}

func (authd *authdetails) tokenURL() string {
	return authd.APIURL + "/token"
}

func (authd *authdetails) tokenisedAuth(ctx context.Context) (CloudAccess, error) {
	conf := oauth2.Config{
		ClientID:     authd.APIClient,
		ClientSecret: authd.APISecret,
		Scopes:       infrastructureScope,
		Endpoint: oauth2.Endpoint{
			TokenURL:  authd.tokenURL(),
			AuthStyle: oauth2.AuthStyleInHeader,
		},
	}
	klog.V(4).Infof("Obtaining authentication for user %s", authd.UserName)
	klog.V(4).Infof("Speaking to %s", authd.tokenURL())
	token, err := conf.PasswordCredentialsToken(ctx, authd.UserName, authd.password)
	if err != nil {
		return nil, err
	}
	klog.V(4).Infof("Refreshing current token as required")
	oauthConnection := conf.Client(ctx, token)
	return cached.NewClient(authd.APIURL, authd.Account, oauthConnection)
}

func (authd *authdetails) apiClientAuth(ctx context.Context) (CloudAccess, error) {
	conf := clientcredentials.Config{
		ClientID:     authd.APIClient,
		ClientSecret: authd.APISecret,
		Scopes:       infrastructureScope,
		TokenURL:     authd.tokenURL(),
	}
	klog.V(4).Infof("Obtaining API client authorisation for client %s", authd.APIClient)
	klog.V(4).Infof("Speaking to %s", authd.tokenURL())
	oauthConnection := conf.Client(ctx)
	return cached.NewClient(authd.APIURL, authd.Account, oauthConnection)
}
