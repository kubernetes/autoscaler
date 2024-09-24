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
	testcloudprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/podlistprocessor"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/deletiontracker"
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
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"

	apiv1 "k8s.io/api/core/v1"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

// NodeConfig is a node config used in tests
type NodeConfig struct {
	Name   string
	Cpu    int64
	Memory int64
	Gpu    int64
	Ready  bool
	Group  string
}

// PodConfig is a pod config used in tests
type PodConfig struct {
	Name         string
	Cpu          int64
	Memory       int64
	Gpu          int64
	Node         string
	ToleratesGpu bool
}

// GroupSizeChange represents a change in group size
type GroupSizeChange struct {
	GroupName  string
	SizeChange int
}

// ScaleTestConfig represents a config of a scale test
type ScaleTestConfig struct {
	Nodes                   []NodeConfig
	Pods                    []PodConfig
	ExtraPods               []PodConfig
	Options                 config.AutoscalingOptions
	NodeDeletionTracker     *deletiontracker.NodeDeletionTracker
	ExpansionOptionToChoose GroupSizeChange // this will be selected by assertingStrategy.BestOption

	ExpectedScaleDowns     []string
	ExpectedScaleDownCount int
}

// NodeGroupConfig is a node group config used in tests
type NodeGroupConfig struct {
	Name    string
	MinSize int
	MaxSize int
}

// NodeTemplateConfig is a structure to provide node info in tests
type NodeTemplateConfig struct {
	MachineType   string
	NodeInfo      *schedulerframework.NodeInfo
	NodeGroupName string
}

// ScaleUpTestConfig represents a config of a scale test
type ScaleUpTestConfig struct {
	Groups                  []NodeGroupConfig
	Nodes                   []NodeConfig
	Pods                    []PodConfig
	ExtraPods               []PodConfig
	OnScaleUp               testcloudprovider.OnScaleUpFunc
	OnCreateGroup           testcloudprovider.OnNodeGroupCreateFunc
	ExpansionOptionToChoose *GroupSizeChange
	Options                 *config.AutoscalingOptions
	NodeTemplateConfigs     map[string]*NodeTemplateConfig
	EnableAutoprovisioning  bool
	AllOrNothing            bool
}

// ScaleUpTestResult represents a node groups scale up result
type ScaleUpTestResult struct {
	ScaleUpError     errors.AutoscalerError
	ScaleUpStatus    ScaleUpStatusInfo
	GroupSizeChanges []GroupSizeChange
	ExpansionOptions []GroupSizeChange
	Events           []string
	GroupTargetSizes map[string]int
}

// ScaleTestResults contains results of a scale test
type ScaleTestResults struct {
	ExpansionOptions []GroupSizeChange
	GroupTargetSizes map[string]int
	FinalOption      GroupSizeChange
	NoScaleUpReason  string
	FinalScaleDowns  []string
	Events           []string
	ScaleUpStatus    ScaleUpStatusInfo
}

// ScaleUpStatusInfo is a simplified form of a ScaleUpStatus, to avoid mocking actual NodeGroup and Pod objects in test config.
type ScaleUpStatusInfo struct {
	Result                  status.ScaleUpResult
	PodsTriggeredScaleUp    []string
	PodsRemainUnschedulable []string
	PodsAwaitEvaluation     []string
}

// WasSuccessful returns true iff scale up was successful
func (s *ScaleUpStatusInfo) WasSuccessful() bool {
	return s.Result == status.ScaleUpSuccessful
}

// ExtractPodNames extract pod names from a list of pods
func ExtractPodNames(pods []*apiv1.Pod) []string {
	podNames := []string{}
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}

// NewTestProcessors returns a set of simple processors for use in tests.
func NewTestProcessors(context *context.AutoscalingContext) *processors.AutoscalingProcessors {
	return &processors.AutoscalingProcessors{
		PodListProcessor:       podlistprocessor.NewDefaultPodListProcessor(context.PredicateChecker, scheduling.ScheduleAnywhere),
		NodeGroupListProcessor: &nodegroups.NoOpNodeGroupListProcessor{},
		BinpackingLimiter:      binpacking.NewTimeLimiter(context.MaxNodeGroupBinpackingDuration),
		NodeGroupSetProcessor:  nodegroupset.NewDefaultNodeGroupSetProcessor([]string{}, config.NodeGroupDifferenceRatios{}),
		ScaleDownSetProcessor: nodes.NewCompositeScaleDownSetProcessor([]nodes.ScaleDownSetProcessor{
			nodes.NewMaxNodesProcessor(),
			nodes.NewAtomicResizeFilteringProcessor(),
		}),
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
