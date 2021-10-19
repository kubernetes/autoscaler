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
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func ResetAuthEnvironment() {
	vars := []string{
		clientEnvVar,
		clientSecretEnvVar,
		usernameEnvVar,
		passwordEnvVar,
		accountEnvVar,
		apiURLEnvVar,
	}
	for _, envvar := range vars {
		os.Unsetenv(envvar)
	}
}

func SetAuthEnvClientID() {
	os.Setenv(clientSecretEnvVar, "not default")
}

func SetAuthEnvUsername() {
	os.Setenv(usernameEnvVar, "itsy@bitzy.com")
}

func SetAuthEnvPassword() {
	os.Setenv(passwordEnvVar, "madeuppassword")
}

func SetAuthEnvAPIURL(value string) {
	os.Setenv(apiURLEnvVar, value)
}

func SetAuthEnvAccount() {
	os.Setenv(accountEnvVar, "acc-testy")
}

func ClearAuthEnvUsername() {
	os.Unsetenv(usernameEnvVar)
}

func GetAuthEnvTokenHandler(t *testing.T) *httptest.Server {
	ResetAuthEnvironment()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		expected := "/token"
		if r.URL.String() != expected {
			t.Errorf("URL = %q; want %q", r.URL, expected)
		}
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Failed reading request body: %s.", err)
		}
		headerContentType := r.Header.Get("Content-Type")
		expected = "application/x-www-form-urlencoded"
		if headerContentType != expected {
			t.Errorf("Content-Type header = %q; want %q", headerContentType, expected)
		}
		headerAuth := r.Header.Get("Authorization")
		expected = "Basic YXBwLWRrbWNoOnVvZ29lbHpndDBud2F3Yg=="
		if headerAuth != expected {
			t.Errorf("Authorization header = %q; want %q", headerAuth, expected)
		}
		switch string(body) {
		case "grant_type=password&password=madeuppassword&scope=infrastructure&username=itsy%40bitzy.com":
			w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
			w.Write([]byte("access_token=90d64460d14870c08c81352a05dedd3465940a7c&scope=user&token_type=bearer"))
		case "grant_type=password&password=&scope=infrastructure&username=itsy%40bitzy.com":
			w.WriteHeader(http.StatusUnauthorized)
		default:
			t.Errorf("Unexpected res.Body = %q", string(body))
			w.WriteHeader(http.StatusUnauthorized)
		}
	}))
	SetAuthEnvAPIURL(ts.URL)
	SetAuthEnvAccount()
	return ts
}

func MakeTestClient(testClient CloudAccess, testMetadata EC2Metadata) *Cloud {
	return &Cloud{
		client:              testClient,
		metadataClientCache: testMetadata,
	}
}
