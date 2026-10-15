/*
Copyright 2022 The Kubernetes Authors.

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

package healthcheck

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"k8s.io/klog/v2"
)

// HealthCheck contains information about last activity time of the monitored component.
// NOTE: This started as a simplified version of VPA's HealthCheck.
type HealthCheck struct {
	// configuration
	address         string
	activityTimeout time.Duration
	// current state
	mutex        sync.Mutex
	lastActivity time.Time
	// set to true when it starts polling API server
	checkTimeout bool
}

// NewHealthCheck builds new HealthCheck object with given timeout.
func NewHealthCheck(address string, activityTimeout time.Duration) *HealthCheck {
	return &HealthCheck{
		address:         address,
		activityTimeout: activityTimeout,
		mutex:           sync.Mutex{},
		lastActivity:    time.Now(),
		checkTimeout:    false,
	}
}

// Serve sets up healthCheck handler on the given address
func (hc *HealthCheck) Serve() {
	go func() {
		http.Handle("/health-check", hc)
		err := http.ListenAndServe(hc.address, nil)
		klog.Fatalf("Failed to start health check: %v", err)
	}()
}

// checkLastActivity returns true if the last activity was too long ago, with duration from it.
func (hc *HealthCheck) checkLastActivity() (bool, time.Duration) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	now := time.Now()
	lastActivity := hc.lastActivity
	timedOut := hc.checkTimeout && now.After(lastActivity.Add(hc.activityTimeout))

	return timedOut, now.Sub(lastActivity)
}

// ServeHTTP implements http.Handler interface to provide a health-check endpoint.
func (hc *HealthCheck) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if timedOut, ago := hc.checkLastActivity(); timedOut {
		http.Error(w, fmt.Sprintf("Error: last activity more than %v ago (threshold is %v)", ago, hc.activityTimeout), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	_, err := w.Write([]byte("OK"))
	if err != nil {
		klog.Fatalf("Failed to write response message: %v", err)
	}
}

// StartMonitoring updates checkTimeout to true.
func (hc *HealthCheck) StartMonitoring() {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	hc.checkTimeout = true
}

// UpdateLastActivity updates last time of activity to now
func (hc *HealthCheck) UpdateLastActivity() {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	hc.lastActivity = time.Now()
}
