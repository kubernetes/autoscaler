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
	"context"
	"fmt"
	"hash/fnv"
	"time"

	"k8s.io/klog/v2"
	"sigs.k8s.io/cluster-autoscaler/pkg/metrics"
	"sigs.k8s.io/cluster-autoscaler/pkg/utils/errors"
)

var startTimeHash string

func init() {
	h := fnv.New32a()
	h.Write([]byte(time.Now().String()))
	startTimeHash = fmt.Sprintf("%08x", h.Sum32())
}

type autoscaler interface {
	// RunOnce represents an iteration in the control-loop of CA.
	RunOnce(ctx context.Context, currentTime time.Time) errors.AutoscalerError
}

// RunAutoscalerOnce triggers a single autoscaling iteration.
func RunAutoscalerOnce(ctx context.Context, autoscaler autoscaler, healthCheck *metrics.HealthCheck, loopStart time.Time, iteration int) {
	iterationID := fmt.Sprintf("%s-%d", startTimeHash, iteration)
	logger := klog.FromContext(ctx).WithValues("iterationId", iterationID)
	ctx = klog.NewContext(ctx, logger)

	metrics.UpdateLastTime(metrics.Main, loopStart)
	healthCheck.UpdateLastActivity(loopStart)

	err := autoscaler.RunOnce(ctx, loopStart)
	if err != nil && err.Type() != errors.TransientError {
		metrics.RegisterError(err)
	} else {
		var successTime = time.Now()
		healthCheck.UpdateLastSuccessfulRun(successTime)
		metrics.UpdateLastTime(metrics.MainSuccessful, successTime)
	}

	metrics.UpdateDurationFromStart(ctx, metrics.Main, loopStart)
}
