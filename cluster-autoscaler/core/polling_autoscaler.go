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

package core

import (
	"time"

	"github.com/golang/glog"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
)

// PollingAutoscaler is a variant of autoscaler which polls the source-of-truth every time RunOnce is invoked
type PollingAutoscaler struct {
	autoscaler        Autoscaler
	autoscalerBuilder AutoscalerBuilder
}

// NewPollingAutoscaler builds a PollingAutoscaler from required parameters
func NewPollingAutoscaler(autoscalerBuilder AutoscalerBuilder) *PollingAutoscaler {
	return &PollingAutoscaler{
		autoscaler:        autoscalerBuilder.Build(),
		autoscalerBuilder: autoscalerBuilder,
	}
}

// CleanUp does the work required before all the iterations of a polling autoscaler run
func (a *PollingAutoscaler) CleanUp() {
	a.autoscaler.CleanUp()
}

// ExitCleanUp cleans-up after autoscaler, so no mess remains after process termination.
func (a *PollingAutoscaler) ExitCleanUp() {
	a.autoscaler.ExitCleanUp()
}

// RunOnce represents a single iteration of a polling autoscaler inside the CA's control-loop
func (a *PollingAutoscaler) RunOnce(currentTime time.Time) {
	reconfigureStart := time.Now()
	metrics.UpdateLastTime("poll", reconfigureStart)
	if err := a.Poll(); err != nil {
		glog.Errorf("Failed to poll : %v", err)
	}
	metrics.UpdateDuration("poll", reconfigureStart)
	a.autoscaler.RunOnce(currentTime)
}

// Poll latest data from cloud provider to recreate this autoscaler
func (a *PollingAutoscaler) Poll() error {
	// For safety, any config change should stop and recreate all the stuff running in CA hence recreating all the Autoscaler instance here
	// See https://github.com/kubernetes/contrib/pull/2226#discussion_r94126064
	a.autoscaler = a.autoscalerBuilder.Build()
	glog.V(4).Infof("Poll finished")

	return nil
}
