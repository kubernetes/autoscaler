/*
Copyright 2026 The Kubernetes Authors.

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

package kamatera

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequest_ShouldReuseBodyIfRetryingPost(t *testing.T) {
	t.Parallel()

	var (
		mu            sync.Mutex
		requestBodies []string
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("failed to read request body: %v", err)
			return
		}

		mu.Lock()
		requestBodies = append(requestBodies, string(body))
		attempt := len(requestBodies)
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		if attempt == 1 {
			w.WriteHeader(http.StatusGatewayTimeout)
			_, _ = w.Write([]byte(`{"message":"retry"}`))
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	_, err := request(
		context.Background(),
		ProviderConfig{ApiUrl: server.URL, ApiClientID: mockKamateraClientId, ApiSecret: mockKamateraSecret},
		"POST",
		"/service/server/terminate",
		KamateraServerTerminatePostRequest{ServerName: "test-server", Force: true},
		2,
		0,
	)

	require.NoError(t, err)
	mu.Lock()
	defer mu.Unlock()
	require.Len(t, requestBodies, 2)
	assert.NotEmpty(t, requestBodies[0])
	assert.Equal(t, requestBodies[0], requestBodies[1])
}
