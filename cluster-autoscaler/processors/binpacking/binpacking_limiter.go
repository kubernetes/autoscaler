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
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/expander"
)

// BinpackingLimiter processes expansion options to stop binpacking early.
type BinpackingLimiter interface {
	InitBinpacking(context *context.AutoscalingContext, nodeGroups []cloudprovider.NodeGroup)
	MarkProcessed(context *context.AutoscalingContext, nodegroupId string)
	StopBinpacking(context *context.AutoscalingContext, evaluatedOptions []expander.Option) bool
	FinalizeBinpacking(context *context.AutoscalingContext, finalOptions []expander.Option)
}

// NoOpBinpackingLimiter returns true without processing expansion options.
type NoOpBinpackingLimiter struct {
}

// NewDefaultBinpackingLimiter creates an instance of NoOpBinpackingLimiter.
func NewDefaultBinpackingLimiter() BinpackingLimiter {
	return &NoOpBinpackingLimiter{}
}

// InitBinpacking initialises the BinpackingLimiter.
func (p *NoOpBinpackingLimiter) InitBinpacking(context *context.AutoscalingContext, nodeGroups []cloudprovider.NodeGroup) {
}

// MarkProcessed marks the nodegroup as processed.
func (p *NoOpBinpackingLimiter) MarkProcessed(context *context.AutoscalingContext, nodegroupId string) {
}

// StopBinpacking is used to make decsions on the evaluated expansion options.
func (p *NoOpBinpackingLimiter) StopBinpacking(context *context.AutoscalingContext, evaluatedOptions []expander.Option) bool {
	return false
}

// FinalizeBinpacking is called to finalize the BinpackingLimiter.
func (p *NoOpBinpackingLimiter) FinalizeBinpacking(context *context.AutoscalingContext, finalOptions []expander.Option) {
}
