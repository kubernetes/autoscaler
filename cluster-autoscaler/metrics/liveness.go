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
	"fmt"
	"net/http"
	"sync"
	"time"
)

// HealthCheck contains information about last time of autoscaler activity and timeout
type HealthCheck struct {
	lastActivity      time.Time
	lastSuccessfulRun time.Time
	mutex             *sync.Mutex
	activityTimeout   time.Duration
	successTimeout    time.Duration
	checkTimeout      bool
}

// NewHealthCheck builds new HealthCheck object with given timeout
func NewHealthCheck(activityTimeout, successTimeout time.Duration) *HealthCheck {
	now := time.Now()
	return &HealthCheck{
		lastActivity:      now,
		lastSuccessfulRun: now,
		mutex:             &sync.Mutex{},
		activityTimeout:   activityTimeout,
		successTimeout:    successTimeout,
		checkTimeout:      false,
	}
}

// StartMonitoring activates checks for autoscaler inactivity
func (hc *HealthCheck) StartMonitoring() {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()
	hc.checkTimeout = true
	now := time.Now()
	if now.After(hc.lastActivity) {
		hc.lastActivity = now
	}
	if now.After(hc.lastSuccessfulRun) {
		hc.lastSuccessfulRun = now
	}
}

// ServeHTTP implements http.Handler interface to provide a health-check endpoint
func (hc *HealthCheck) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hc.mutex.Lock()

	lastActivity := hc.lastActivity
	lastSuccessfulRun := hc.lastSuccessfulRun
	now := time.Now()
	activityTimedOut := now.After(lastActivity.Add(hc.activityTimeout))
	successTimedOut := now.After(lastSuccessfulRun.Add(hc.successTimeout))
	timedOut := hc.checkTimeout && (activityTimedOut || successTimedOut)

	hc.mutex.Unlock()

	if timedOut {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("Error: last activity more %v ago, last success more than %v ago", time.Now().Sub(lastActivity).String(), time.Now().Sub(lastSuccessfulRun).String())))
	} else {
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	}
}

// UpdateLastActivity updates last time of activity
func (hc *HealthCheck) UpdateLastActivity(timestamp time.Time) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()
	if timestamp.After(hc.lastActivity) {
		hc.lastActivity = timestamp
	}
}

// UpdateLastSuccessfulRun updates last time of successful (i.e. not ending in error) activity
func (hc *HealthCheck) UpdateLastSuccessfulRun(timestamp time.Time) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()
	if timestamp.After(hc.lastSuccessfulRun) {
		hc.lastSuccessfulRun = timestamp
	}
	// finishing successful run is also a sign of activity
	if timestamp.After(hc.lastActivity) {
		hc.lastActivity = timestamp
	}
}
