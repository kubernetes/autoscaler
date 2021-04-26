/*
Copyright 2021 The Kubernetes Authors.

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

package rancher

import (
	"log"
	"net/http"
	"net/http/httputil"
)

type transport struct {
	Transport   http.RoundTripper
	LogRequest  func(req *http.Request)
	LogResponse func(resp *http.Response)
}

var loggerTransport = &transport{
	Transport: http.DefaultTransport,
}

// DefaultLogRequest is used if transport.LogRequest is not set.
var DefaultLogRequest = func(req *http.Request) {
	a, _ := httputil.DumpRequestOut(req, true)
	log.Printf("-->\n%s", a)
}

// DefaultLogResponse is used if transport.LogResponse is not set.
var DefaultLogResponse = func(resp *http.Response) {
	a, _ := httputil.DumpResponse(resp, true)
	log.Printf("<--\n%s", a)
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.logRequest(req)

	resp, err := t.transport().RoundTrip(req)
	if err != nil {
		return resp, err
	}

	t.logResponse(resp)
	return resp, err
}

func (t *transport) logRequest(req *http.Request) {
	if t.LogRequest != nil {
		t.LogRequest(req)
	} else {
		DefaultLogRequest(req)
	}
}

func (t *transport) logResponse(resp *http.Response) {
	if t.LogResponse != nil {
		t.LogResponse(resp)
	} else {
		DefaultLogResponse(resp)
	}
}

func (t *transport) transport() http.RoundTripper {
	if t.Transport != nil {
		return t.Transport
	}

	return http.DefaultTransport
}
