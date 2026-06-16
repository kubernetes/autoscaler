/*
Copyright The Kubernetes Authors.

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

package logic

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource"
)

// InfiniteReader simulates an endless stream of bytes to trigger an oversized request error.
type InfiniteReader struct{}

func (*InfiniteReader) Read(p []byte) (n int, err error) {
	for i := range p {
		p[i] = 'a'
	}
	return len(p), nil
}

func (*InfiniteReader) Close() error {
	return nil
}

// FailureReader simulates a read failure that is not an oversized error.
type FailureReader struct{}

func (*FailureReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("simulated read error")
}

func (*FailureReader) Close() error {
	return nil
}

func TestServePayloadLimit(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    *bytes.Buffer
		isEndless      bool
		isFailure      bool
		expectedStatus int
	}{
		{
			name:           "Small valid JSON payload",
			requestBody:    bytes.NewBufferString(`{"kind":"AdmissionReview","apiVersion":"admission.k8s.io/v1","request":{"uid":"123","resource":{"group":"autoscaling.k8s.io","version":"v1","resource":"verticalpodautoscalers"},"requestKind":{"group":"autoscaling.k8s.io","version":"v1","kind":"VerticalPodAutoscaler"}}}`),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Oversized payload exceeding limit",
			isEndless:      true,
			expectedStatus: http.StatusRequestEntityTooLarge,
		},
		{
			name:           "General read error",
			isFailure:      true,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := &AdmissionServer{
				resourceHandlers: make(map[metav1.GroupResource]resource.Handler), // Unused in basic HTTP parse step
			}

			var req *http.Request
			if tc.isEndless {
				req = httptest.NewRequest(http.MethodPost, "/admit", &InfiniteReader{})
			} else if tc.isFailure {
				req = httptest.NewRequest(http.MethodPost, "/admit", &FailureReader{})
			} else {
				req = httptest.NewRequest(http.MethodPost, "/admit", tc.requestBody)
			}
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			server.Serve(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			if tc.expectedStatus == http.StatusOK {
				var ar admissionv1.AdmissionReview
				if err := json.Unmarshal(w.Body.Bytes(), &ar); assert.NoError(t, err) {
					if assert.NotNil(t, ar.Response) {
						assert.True(t, ar.Response.Allowed)
						if tc.isFailure {
							assert.Empty(t, ar.Response.UID)
						} else {
							assert.Equal(t, "123", string(ar.Response.UID))
						}
					}
				}
			}
		})
	}
}
