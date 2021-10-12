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

package ionoscloud

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
)

func TestAPIClientFor(t *testing.T) {
	apiClientFactory = func(_, _ string, _ bool) APIClient { return &MockAPIClient{} }
	defer func() { apiClientFactory = NewAPIClient }()

	cases := []struct {
		name         string
		token        string
		cachedTokens map[string]string
		expectClient APIClient
		expectError  bool
	}{
		{
			name:         "from cached client",
			token:        "token",
			expectClient: &MockAPIClient{},
		},
		{
			name:         "from token cache",
			cachedTokens: map[string]string{"test": "token"},
			expectClient: &MockAPIClient{},
		},
		{
			name:         "not in token cache",
			cachedTokens: map[string]string{"notfound": "token"},
			expectError:  true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			apiClientFactory = func(_, _ string, _ bool) APIClient { return &MockAPIClient{} }
			defer func() { apiClientFactory = NewAPIClient }()

			client, _ := NewAutoscalingClient(&Config{
				Token:    c.token,
				Endpoint: "https://api.cloud.ionos.com/v6",
				Insecure: true,
			})
			client.tokens = c.cachedTokens
			apiClient, err := client.apiClientFor("test")
			require.Equalf(t, c.expectError, err != nil, "expected error: %t, got: %v", c.expectError, err)
			require.EqualValues(t, c.expectClient, apiClient)
		})
	}
}

func TestLoadTokensFromFilesystem_OK(t *testing.T) {
	tempDir, err := ioutil.TempDir("", t.Name())
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	uuid1, uuid2, uuid3 := uuid.NewV4().String(), uuid.NewV4().String(), uuid.NewV4().String()

	input := map[string]string{
		uuid1: "token1",
		uuid2: "token2",
		uuid3: "token3",
	}
	expect := map[string]string{
		uuid1: "token1",
		uuid2: "token2",
		uuid3: "token3",
	}

	for name, token := range input {
		require.NoError(t, ioutil.WriteFile(filepath.Join(tempDir, name), []byte(token), 0600))
	}
	require.NoError(t, ioutil.WriteFile(filepath.Join(tempDir, "..somfile"), []byte("foobar"), 0600))

	client, err := NewAutoscalingClient(&Config{TokensPath: tempDir})
	require.NoError(t, err)
	require.Equal(t, expect, client.tokens)
}

func TestLoadTokensFromFilesystem_ReadError(t *testing.T) {
	tempDir, err := ioutil.TempDir("", t.Name())
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	require.NoError(t, os.Mkdir(filepath.Join(tempDir, uuid.NewV4().String()), 0755))
	client, err := NewAutoscalingClient(&Config{TokensPath: tempDir})
	require.Error(t, err)
	require.Nil(t, client)
}

func TestLoadTokensFromFilesystem_NoValidToken(t *testing.T) {
	tempDir, err := ioutil.TempDir("", t.Name())
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	for i := 0; i < 10; i++ {
		path := filepath.Join(tempDir, fmt.Sprintf("notauuid%d", i))
		require.NoError(t, ioutil.WriteFile(path, []byte("token"), 0600))
		path = filepath.Join(tempDir, fmt.Sprintf("foo.bar.notauuid%d", i))
		require.NoError(t, ioutil.WriteFile(path, []byte("token"), 0600))
	}
	client, err := NewAutoscalingClient(&Config{TokensPath: tempDir})
	require.Error(t, err)
	require.Nil(t, client)
}
