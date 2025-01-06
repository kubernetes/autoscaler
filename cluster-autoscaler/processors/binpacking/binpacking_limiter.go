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

// BinpackingLimiter processes expansion options to stop binpacking early.
type BinpackingLimiter interface {
	InitBinpacking(autoscalingContext *ca_context.AutoscalingContext, nodeGroups []cloudprovider.NodeGroup)
	MarkProcessed(autoscalingContext *ca_context.AutoscalingContext, nodegroupId string)
	StopBinpacking(autoscalingContext *ca_context.AutoscalingContext, evaluatedOptions []expander.Option) bool
	FinalizeBinpacking(autoscalingContext *ca_context.AutoscalingContext, finalOptions []expander.Option)
}
