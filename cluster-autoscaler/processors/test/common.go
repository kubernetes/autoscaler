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

package test

import (
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/podlistprocessor"
	"k8s.io/autoscaler/cluster-autoscaler/observers/nodegroupchange"
	"k8s.io/autoscaler/cluster-autoscaler/processors"
	"k8s.io/autoscaler/cluster-autoscaler/processors/actionablecluster"
	"k8s.io/autoscaler/cluster-autoscaler/processors/binpacking"
	"k8s.io/autoscaler/cluster-autoscaler/processors/customresources"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroupconfig"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroups"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroups/asyncnodegroups"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroupset"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodeinfosprovider"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodes"
	"k8s.io/autoscaler/cluster-autoscaler/processors/scaledowncandidates"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/scheduling"
)

// NewTestProcessors returns a set of simple processors for use in tests.
func NewTestProcessors(context *context.AutoscalingContext) *processors.AutoscalingProcessors {
	return &processors.AutoscalingProcessors{
		PodListProcessor:       podlistprocessor.NewDefaultPodListProcessor(scheduling.ScheduleAnywhere),
		NodeGroupListProcessor: &nodegroups.NoOpNodeGroupListProcessor{},
		BinpackingLimiter:      binpacking.NewTimeLimiter(context.MaxNodeGroupBinpackingDuration),
		NodeGroupSetProcessor:  nodegroupset.NewDefaultNodeGroupSetProcessor([]string{}, config.NodeGroupDifferenceRatios{}),
		ScaleDownSetProcessor:  nodes.NewAtomicResizeFilteringProcessor(),
		// TODO(bskiba): change scale up test so that this can be a NoOpProcessor
		ScaleUpStatusProcessor:      &status.EventingScaleUpStatusProcessor{},
		ScaleDownStatusProcessor:    &status.NoOpScaleDownStatusProcessor{},
		AutoscalingStatusProcessor:  &status.NoOpAutoscalingStatusProcessor{},
		NodeGroupManager:            nodegroups.NewDefaultNodeGroupManager(),
		TemplateNodeInfoProvider:    nodeinfosprovider.NewDefaultTemplateNodeInfoProvider(nil, false),
		NodeGroupConfigProcessor:    nodegroupconfig.NewDefaultNodeGroupConfigProcessor(context.NodeGroupDefaults),
		CustomResourcesProcessor:    customresources.NewDefaultCustomResourcesProcessor(),
		ActionableClusterProcessor:  actionablecluster.NewDefaultActionableClusterProcessor(),
		ScaleDownCandidatesNotifier: scaledowncandidates.NewObserversList(),
		ScaleStateNotifier:          nodegroupchange.NewNodeGroupChangeObserversList(),
		AsyncNodeGroupStateChecker:  asyncnodegroups.NewDefaultAsyncNodeGroupStateChecker(),
	}
}
