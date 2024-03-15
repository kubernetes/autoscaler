/*
Copyright 2024 The Kubernetes Authors.

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

package loop

import (
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
)

type autoscaler interface {
	// RunOnce represents an iteration in the control-loop of CA.
	RunOnce(currentTime time.Time) errors.AutoscalerError
}

// RunAutoscalerOnce triggers a single autoscaling iteration.
func RunAutoscalerOnce(autoscaler autoscaler, healthCheck *metrics.HealthCheck, loopStart time.Time) {
	metrics.UpdateLastTime(metrics.Main, loopStart)
	healthCheck.UpdateLastActivity(loopStart)

	err := autoscaler.RunOnce(loopStart)
	if err != nil && err.Type() != errors.TransientError {
		metrics.RegisterError(err)
	} else {
		healthCheck.UpdateLastSuccessfulRun(time.Now())
	}

	metrics.UpdateDurationFromStart(metrics.Main, loopStart)
}
