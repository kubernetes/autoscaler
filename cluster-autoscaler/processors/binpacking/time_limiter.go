/*
Copyright 2023 The Kubernetes Authors.

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

package binpacking

import (
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/expander"
	"k8s.io/klog/v2"
)

// TimeLimiter expands binpackingLimiter to limit the total time spends on binpacking
type TimeLimiter struct {
	limiter               BinpackingLimiter
	startTime             time.Time
	maxBinpackingDuration time.Duration
}

// NewTimeLimiter returns an instance of a new TimeLimiter
func NewTimeLimiter(maxBinpackingDuration time.Duration, limiter BinpackingLimiter) *TimeLimiter {
	return &TimeLimiter{
		limiter:               limiter,
		maxBinpackingDuration: maxBinpackingDuration,
	}
}

// InitBinpacking initialises the TimeLimiter.
func (b *TimeLimiter) InitBinpacking(context *context.AutoscalingContext, nodeGroups []cloudprovider.NodeGroup) {
	b.limiter.InitBinpacking(context, nodeGroups)
	b.startTime = time.Now()
}

// MarkProcessed marks the nodegroup as processed.
func (b *TimeLimiter) MarkProcessed(context *context.AutoscalingContext, nodegroupId string) {
	b.limiter.MarkProcessed(context, nodegroupId)
}

// StopBinpacking returns true if the binpacking time exceeds maxBinpackingDuration
func (b *TimeLimiter) StopBinpacking(context *context.AutoscalingContext, evaluatedOptions []expander.Option) bool {
	stopCondition := b.limiter.StopBinpacking(context, evaluatedOptions)
	if time.Now().After(b.startTime.Add(b.maxBinpackingDuration)) {
		klog.Info("Binpacking is cut short due to maxBinpackingDuration reached.")
		return true
	}
	return stopCondition
}

// FinalizeBinpacking is called to finalize the BinpackingLimiter.
func (b *TimeLimiter) FinalizeBinpacking(context *context.AutoscalingContext, finalOptions []expander.Option) {
	b.limiter.FinalizeBinpacking(context, finalOptions)
}
