/*
Copyright 2017 The Kubernetes Authors.

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

package metrics

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func getTestResponse(start time.Time, timeout time.Duration, checkMonitoring bool) *httptest.ResponseRecorder {
	req := httptest.NewRequest("GET", "/health-check", nil)
	w := httptest.NewRecorder()
	healthCheck := NewHealthCheck(timeout)
	if checkMonitoring {
		healthCheck.StartMonitoring()
	}
	healthCheck.lastActivity = start
	healthCheck.ServeHTTP(w, req)
	return w
}

func TestOkServeHTTP(t *testing.T) {
	w := getTestResponse(time.Now(), time.Second, true)
	assert.Equal(t, 200, w.Code)
}

func TestFailServeHTTP(t *testing.T) {
	w := getTestResponse(time.Now().Add(time.Second*-2), time.Second, true)
	assert.Equal(t, 500, w.Code)
}

func TestMonitoringOffAfterTimeout(t *testing.T) {
	w := getTestResponse(time.Now(), time.Second, false)
	assert.Equal(t, 200, w.Code)
}

func TestMonitoringOffBeforeTimeout(t *testing.T) {
	w := getTestResponse(time.Now().Add(time.Second*-2), time.Second, false)
	assert.Equal(t, 200, w.Code)
}

func TestUpdateLastActivity(t *testing.T) {
	timeout := time.Second
	start := time.Now().Add(timeout * -2)

	req := httptest.NewRequest("GET", "/health-check", nil)
	w := httptest.NewRecorder()
	healthCheck := NewHealthCheck(timeout)
	healthCheck.StartMonitoring()
	healthCheck.lastActivity = start

	healthCheck.ServeHTTP(w, req)
	assert.Equal(t, 500, w.Code)

	w = httptest.NewRecorder()
	healthCheck.UpdateLastActivity(time.Now())
	healthCheck.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}
