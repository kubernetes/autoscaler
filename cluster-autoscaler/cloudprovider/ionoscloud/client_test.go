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
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCustomGetClient(t *testing.T) {
	tokensPath := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tokensPath, "..ignoreme"), []byte(`{"invalid"}`), 0o600))

	client := NewAutoscalingClient(&Config{
		TokensPath:        tokensPath,
		Endpoint:          "https://api.ionos.com",
		Insecure:          true,
		AdditionalHeaders: map[string]string{"Foo": "Bar"},
	}, "test")

	_, err := client.getClient()
	require.Error(t, err)

	require.NoError(t, os.WriteFile(filepath.Join(tokensPath, "a"), []byte(`{"tokens":["token1"]}`), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(tokensPath, "b"), []byte(`{"tokens":["token2"]}`), 0o600))

	c, err := client.getClient()
	require.NoError(t, err)
	require.NotNil(t, c)
}
