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
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	ca_context "k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/expander"
)

// CombinedLimiter combines the outcome of multiple limiters. It will limit
// binpacking when at least one limiter meets the stop condition.
type CombinedLimiter struct {
	limiters []BinpackingLimiter
}

// NewCombinedLimiter returns an instance of a new CombinedLimiter.
func NewCombinedLimiter(limiters []BinpackingLimiter) *CombinedLimiter {
	return &CombinedLimiter{
		limiters: limiters,
	}
}

// InitBinpacking initialises all the underline limiters.
func (l *CombinedLimiter) InitBinpacking(autoscalingContext *ca_context.AutoscalingContext, nodeGroups []cloudprovider.NodeGroup) {
	for _, limiter := range l.limiters {
		limiter.InitBinpacking(autoscalingContext, nodeGroups)
	}
}

// MarkProcessed marks the nodegroup as processed in all underline limiters.
func (l *CombinedLimiter) MarkProcessed(autoscalingContext *ca_context.AutoscalingContext, nodegroupId string) {
	for _, limiter := range l.limiters {
		limiter.MarkProcessed(autoscalingContext, nodegroupId)
	}
}

// StopBinpacking returns true if at least one of the underline limiter met the stop condition.
func (l *CombinedLimiter) StopBinpacking(autoscalingContext *ca_context.AutoscalingContext, evaluatedOptions []expander.Option) bool {
	stopCondition := false
	for _, limiter := range l.limiters {
		stopCondition = limiter.StopBinpacking(autoscalingContext, evaluatedOptions) || stopCondition
	}
	return stopCondition
}

// FinalizeBinpacking will call FinalizeBinpacking for all the underline limiters.
func (l *CombinedLimiter) FinalizeBinpacking(autoscalingContext *ca_context.AutoscalingContext, finalOptions []expander.Option) {
	for _, limiter := range l.limiters {
		limiter.FinalizeBinpacking(autoscalingContext, finalOptions)
	}
}
