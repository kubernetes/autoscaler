/*
Copyright 2016 The Kubernetes Authors.

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

package test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/stretchr/testify/mock"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// BuildServiceTokenProjectedVolumeSource returns a ProjectedVolumeSource with SA token
// projection
func BuildServiceTokenProjectedVolumeSource(path string) *apiv1.ProjectedVolumeSource {
	return &apiv1.ProjectedVolumeSource{
		Sources: []apiv1.VolumeProjection{
			{
				ServiceAccountToken: &apiv1.ServiceAccountTokenProjection{
					Path: path,
				},
			},
		},
	}
}

const (
	// cannot use constants from gpu module due to cyclic package import
	ResourceNvidiaGPU = "nvidia.com/gpu"
	GPULabel          = "cloud.google.com/gke-accelerator"
	DefaultGPUType    = "nvidia-tesla-k80"
)

// GenerateOwnerReferences builds OwnerReferences with a single reference
func GenerateOwnerReferences(name, kind, api string, uid types.UID) []metav1.OwnerReference {
	return []metav1.OwnerReference{
		{
			APIVersion:         api,
			Kind:               kind,
			Name:               name,
			BlockOwnerDeletion: boolptr(true),
			Controller:         boolptr(true),
			UID:                uid,
		},
	}
}

func boolptr(val bool) *bool {
	b := val
	return &b
}

// HttpServerMock mocks server HTTP.
//
// Example:
// // Create HttpServerMock.
// server := NewHttpServerMock()
// defer server.Close()
// // Use server.URL to point your code to HttpServerMock.
// g := newTestGceManager(t, server.URL, ModeGKE)
// // Declare handled urls and results for them.
// server.On("handle", "/project1/zones/us-central1-b/listManagedInstances").Return("<managedInstances>").Once()
// // Call http server in your code.
// instances, err := g.GetManagedInstances()
// // Check if expected calls were executed.
//
//	mock.AssertExpectationsForObjects(t, server)
//
// Note: to provide a content type, you may pass in the desired
// fields:
// server := NewHttpServerMock(MockFieldContentType, MockFieldResponse)
// ...
// server.On("handle", "/project1/zones/us-central1-b/listManagedInstances").Return("<content type>", "<response>").Once()
// The order of the return objects must match that of the HttpServerMockField constants passed to NewHttpServerMock()
type HttpServerMock struct {
	mock.Mock
	*httptest.Server
	fields []HttpServerMockField
}

// HttpServerMockField specifies a type of field.
type HttpServerMockField int

const (
	// MockFieldResponse represents a string response.
	MockFieldResponse HttpServerMockField = iota
	// MockFieldStatusCode represents an integer HTTP response code.
	MockFieldStatusCode
	// MockFieldContentType represents a string content type.
	MockFieldContentType
	// MockFieldUserAgent represents a string user agent.
	MockFieldUserAgent
)

// NewHttpServerMock creates new HttpServerMock.
func NewHttpServerMock(fields ...HttpServerMockField) *HttpServerMock {
	if len(fields) == 0 {
		fields = []HttpServerMockField{MockFieldResponse}
	}
	foundResponse := false
	for _, field := range fields {
		if field == MockFieldResponse {
			foundResponse = true
			break
		}
	}
	if !foundResponse {
		panic("Must use MockFieldResponse.")
	}
	httpServerMock := &HttpServerMock{fields: fields}
	mux := http.NewServeMux()
	mux.HandleFunc("/",
		func(w http.ResponseWriter, req *http.Request) {
			result := httpServerMock.handle(req, w, httpServerMock)
			_, _ = w.Write([]byte(result))
		})

	server := httptest.NewServer(mux)
	httpServerMock.Server = server
	return httpServerMock
}

func (l *HttpServerMock) handle(req *http.Request, w http.ResponseWriter, serverMock *HttpServerMock) string {
	url := req.URL.Path
	var response string
	args := l.Called(url)
	for i, field := range l.fields {
		switch field {
		case MockFieldResponse:
			response = args.String(i)
		case MockFieldContentType:
			w.Header().Set("Content-Type", args.String(i))
		case MockFieldStatusCode:
			w.WriteHeader(args.Int(i))
		case MockFieldUserAgent:
			gotUserAgent := req.UserAgent()
			expectedUserAgent := args.String(i)
			if !strings.Contains(gotUserAgent, expectedUserAgent) {
				panic(fmt.Sprintf("Error handling URL %s, expected user agent %s but got %s.", url, expectedUserAgent, gotUserAgent))
			}
		}
	}
	return response
}
