/*
Copyright 2018 The Kubernetes Authors.

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
	"fmt"
	"net/http"
	"sync"
	"time"

	"k8s.io/klog/v2"
)

// HealthCheck contains information about last activity time of the monitored component.
// NOTE: This started as a simplified version of ClusterAutoscaler's HealthCheck.
type HealthCheck struct {
	activityTimeout time.Duration
	checkTimeout    bool
	lastActivity    time.Time
	mutex           *sync.Mutex
}

// NewHealthCheck builds new HealthCheck object with given timeout.
func NewHealthCheck(activityTimeout time.Duration) *HealthCheck {
	return &HealthCheck{
		activityTimeout: activityTimeout,
		checkTimeout:    false,
		lastActivity:    time.Now(),
		mutex:           &sync.Mutex{},
	}
}

// StartMonitoring activates checks for the component inactivity.
func (hc *HealthCheck) StartMonitoring() {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	hc.checkTimeout = true
	hc.lastActivity = time.Now()
}

// checkLastActivity returns true if the last activity was too long ago, with duration from it.
func (hc *HealthCheck) checkLastActivity() (bool, time.Duration) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	now := time.Now()
	lastActivity := hc.lastActivity
	activityTimedOut := now.After(lastActivity.Add(hc.activityTimeout))
	timedOut := hc.checkTimeout && activityTimedOut

	return timedOut, now.Sub(lastActivity)
}

// ServeHTTP implements http.Handler interface to provide a health-check endpoint.
func (hc *HealthCheck) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	timedOut, ago := hc.checkLastActivity()
	if timedOut {
		http.Error(w, fmt.Sprintf("Error: last activity more than %v ago", ago), http.StatusInternalServerError)
	} else {
		w.WriteHeader(200)
		_, err := w.Write([]byte("OK"))
		if err != nil {
			klog.Fatalf("Failed to write response message: %v", err)
		}
	}
}

// UpdateLastActivity updates last time of activity to now
func (hc *HealthCheck) UpdateLastActivity() {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	hc.lastActivity = time.Now()
}
