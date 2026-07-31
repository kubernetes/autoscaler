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
	"time"

	"sigs.k8s.io/cluster-autoscaler/pkg/config"
	ca_context "sigs.k8s.io/cluster-autoscaler/pkg/context"
	"sigs.k8s.io/cluster-autoscaler/pkg/core/podlistprocessor"
	"sigs.k8s.io/cluster-autoscaler/pkg/observers/nodegroupchange"
	"sigs.k8s.io/cluster-autoscaler/pkg/processors"
	"sigs.k8s.io/cluster-autoscaler/pkg/processors/actionablecluster"
	"sigs.k8s.io/cluster-autoscaler/pkg/processors/binpacking"
	"sigs.k8s.io/cluster-autoscaler/pkg/processors/customresources"
	"sigs.k8s.io/cluster-autoscaler/pkg/processors/nodegroupconfig"
	"sigs.k8s.io/cluster-autoscaler/pkg/processors/nodegroups"
	"sigs.k8s.io/cluster-autoscaler/pkg/processors/nodegroups/asyncnodegroups"
	"sigs.k8s.io/cluster-autoscaler/pkg/processors/nodegroupset"
	"sigs.k8s.io/cluster-autoscaler/pkg/processors/nodeinfosprovider"
	"sigs.k8s.io/cluster-autoscaler/pkg/processors/nodes"
	"sigs.k8s.io/cluster-autoscaler/pkg/processors/pods"
	"sigs.k8s.io/cluster-autoscaler/pkg/processors/scaledowncandidates"
	"sigs.k8s.io/cluster-autoscaler/pkg/processors/status"
	"sigs.k8s.io/cluster-autoscaler/pkg/simulator/scheduling"
)

// NewTestProcessors returns a set of simple processors for use in tests.
func NewTestProcessors(options config.AutoscalingOptions) (*processors.AutoscalingProcessors, ca_context.TemplateNodeInfoRegistry) {
	templateNodeInfoProvider := nodeinfosprovider.NewDefaultTemplateNodeInfoProvider(nil, false)
	templateNodeInfoRegistry := nodeinfosprovider.NewTemplateNodeInfoRegistry(templateNodeInfoProvider)
	if options.PendingPodsBatchingTimeout == 0 {
		// This disables graceful degradation in tests if not set explicitly.
		options.PendingPodsBatchingTimeout = 24 * time.Hour
	}

	return &processors.AutoscalingProcessors{
		PodListProcessor:       podlistprocessor.NewDefaultPodListProcessor(scheduling.ScheduleAnywhere),
		NodeGroupListProcessor: &nodegroups.NoOpNodeGroupListProcessor{},
		BinpackingLimiter:      binpacking.NewTimeLimiter(options.MaxNodeGroupBinpackingDuration),
		NodeGroupSetProcessor:  nodegroupset.NewDefaultNodeGroupSetProcessor([]string{}, config.NodeGroupDifferenceRatios{}),
		ScaleDownSetProcessor:  nodes.NewAtomicResizeFilteringProcessor(),
		// TODO(bskiba): change scale up test so that this can be a NoOpProcessor
		ScaleUpStatusProcessor:      &status.EventingScaleUpStatusProcessor{},
		ScaleDownStatusProcessor:    &status.NoOpScaleDownStatusProcessor{},
		AutoscalingStatusProcessor:  &status.NoOpAutoscalingStatusProcessor{},
		NodeGroupManager:            nodegroups.NewDefaultNodeGroupManager(),
		TemplateNodeInfoProvider:    templateNodeInfoProvider,
		NodeGroupConfigProcessor:    nodegroupconfig.NewDefaultNodeGroupConfigProcessor(options.NodeGroupDefaults),
		CustomResourcesProcessor:    customresources.NewDefaultCustomResourcesProcessor(true, false),
		ActionableClusterProcessor:  actionablecluster.NewDefaultActionableClusterProcessor(),
		ScaleDownCandidatesNotifier: scaledowncandidates.NewObserversList(),
		ScaleStateNotifier:          nodegroupchange.NewNodeGroupChangeObserversList(),
		AsyncNodeGroupStateChecker:  asyncnodegroups.NewDefaultAsyncNodeGroupStateChecker(),
		ScaleUpEnforcer:             pods.NewDefaultScaleUpEnforcer(),
	}, templateNodeInfoRegistry
}
