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

func getTestResponse(start time.Time, activityTimeout, successTimeout time.Duration, checkMonitoring bool) *httptest.ResponseRecorder {
	req := httptest.NewRequest("GET", "/health-check", nil)
	w := httptest.NewRecorder()
	healthCheck := NewHealthCheck(activityTimeout, successTimeout)
	if checkMonitoring {
		healthCheck.StartMonitoring()
	}
	healthCheck.lastActivity = start
	healthCheck.lastSuccessfulRun = start
	healthCheck.ServeHTTP(w, req)
	return w
}

func TestOkServeHTTP(t *testing.T) {
	w := getTestResponse(time.Now(), time.Second, time.Second, true)
	assert.Equal(t, 200, w.Code)
}

func TestFailTimeoutServeHTTP(t *testing.T) {
	w := getTestResponse(time.Now().Add(time.Second*-2), time.Second, time.Second, true)
	assert.Equal(t, 500, w.Code)
}

func TestMonitoringOffAfterTimeout(t *testing.T) {
	w := getTestResponse(time.Now().Add(time.Second*-2), time.Second, time.Second, false)
	assert.Equal(t, 200, w.Code)
}

func TestMonitoringOffBeforeTimeout(t *testing.T) {
	w := getTestResponse(time.Now().Add(time.Second*2), time.Second, time.Second, false)
	assert.Equal(t, 200, w.Code)
}

func TestUpdateLastActivity(t *testing.T) {
	timeout := time.Second
	start := time.Now().Add(timeout * -2)
	// to make sure it doesn't cause health check failure
	lastSuccess := time.Now().Add(timeout * 10)

	req := httptest.NewRequest("GET", "/health-check", nil)
	healthCheck := NewHealthCheck(timeout, timeout)
	healthCheck.StartMonitoring()
	healthCheck.lastActivity = start
	healthCheck.lastSuccessfulRun = lastSuccess

	w := httptest.NewRecorder()
	healthCheck.ServeHTTP(w, req)
	assert.Equal(t, 500, w.Code)

	w = httptest.NewRecorder()
	healthCheck.UpdateLastActivity(time.Now())
	healthCheck.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestUpdateActivityAtUpdateLastSuccessfulRun(t *testing.T) {
	timeout := time.Second
	start := time.Now().Add(timeout * -2)
	// to make sure it doesn't cause health check failure
	lastSuccess := time.Now().Add(timeout * 10)

	req := httptest.NewRequest("GET", "/health-check", nil)
	healthCheck := NewHealthCheck(timeout, timeout)
	healthCheck.StartMonitoring()
	healthCheck.lastActivity = start
	healthCheck.lastSuccessfulRun = lastSuccess

	w := httptest.NewRecorder()
	healthCheck.ServeHTTP(w, req)
	assert.Equal(t, 500, w.Code)

	w = httptest.NewRecorder()
	healthCheck.UpdateLastSuccessfulRun(time.Now())
	healthCheck.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	// verify last successful run from the future wasn't overwritten
	assert.Equal(t, true, healthCheck.lastSuccessfulRun.After(healthCheck.lastActivity))
}

func TestUpdateLastSuccessfulRun(t *testing.T) {
	timeout := time.Second
	start := time.Now().Add(timeout * -2)
	// to make sure it doesn't cause health check failure
	lastActivity := time.Now().Add(timeout * 10)

	req := httptest.NewRequest("GET", "/health-check", nil)
	healthCheck := NewHealthCheck(timeout, timeout)
	healthCheck.StartMonitoring()
	healthCheck.lastActivity = lastActivity
	healthCheck.lastSuccessfulRun = start

	w := httptest.NewRecorder()
	healthCheck.ServeHTTP(w, req)
	assert.Equal(t, 500, w.Code)

	w = httptest.NewRecorder()
	healthCheck.UpdateLastSuccessfulRun(time.Now())
	healthCheck.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	// verify last activity timestamp from the future wasn't overwritten
	assert.Equal(t, true, healthCheck.lastActivity.After(healthCheck.lastSuccessfulRun))
}
