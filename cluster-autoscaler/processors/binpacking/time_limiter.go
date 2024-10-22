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

// TimeLimiter limits binpacking based on the total time spends on binpacking.
type TimeLimiter struct {
	startTime             time.Time
	maxBinpackingDuration time.Duration
}

// NewTimeLimiter returns an instance of a new TimeLimiter.
func NewTimeLimiter(maxBinpackingDuration time.Duration) *TimeLimiter {
	return &TimeLimiter{
		maxBinpackingDuration: maxBinpackingDuration,
	}
}

// InitBinpacking initialises the TimeLimiter.
func (b *TimeLimiter) InitBinpacking(context *context.AutoscalingContext, nodeGroups []cloudprovider.NodeGroup) {
	b.startTime = time.Now()
}

// MarkProcessed marks the nodegroup as processed.
func (b *TimeLimiter) MarkProcessed(context *context.AutoscalingContext, nodegroupId string) {
}

// StopBinpacking returns true if the binpacking time exceeds maxBinpackingDuration.
func (b *TimeLimiter) StopBinpacking(context *context.AutoscalingContext, evaluatedOptions []expander.Option) bool {
	now := time.Now()
	if now.After(b.startTime.Add(b.maxBinpackingDuration)) {
		klog.Infof("Binpacking is cut short after %v seconds due to exceeding maxBinpackingDuration", now.Sub(b.startTime).Seconds())
	}
	return false
}

// FinalizeBinpacking is called to finalize the BinpackingLimiter.
func (b *TimeLimiter) FinalizeBinpacking(context *context.AutoscalingContext, finalOptions []expander.Option) {
}
