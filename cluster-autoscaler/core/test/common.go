/*
Copyright 2018 The Kubernetes Authors.

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
	"fmt"
	"reflect"
	"testing"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	testcloudprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate/utils"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/deletiontracker"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/pdb"
	"k8s.io/autoscaler/cluster-autoscaler/debuggingsnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/expander"
	"k8s.io/autoscaler/cluster-autoscaler/expander/random"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	processor_callbacks "k8s.io/autoscaler/cluster-autoscaler/processors/callbacks"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroups"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot/testsnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/predicatechecker"
	"k8s.io/autoscaler/cluster-autoscaler/utils/backoff"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/autoscaler/cluster-autoscaler/utils/labels"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	kube_client "k8s.io/client-go/kubernetes"
	kube_record "k8s.io/client-go/tools/record"
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
	NodeInfo      *framework.NodeInfo
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

// NewScaleTestAutoscalingContext creates a new test autoscaling context for scaling tests.
func NewScaleTestAutoscalingContext(
	options config.AutoscalingOptions, fakeClient kube_client.Interface,
	listers kube_util.ListerRegistry, provider cloudprovider.CloudProvider,
	processorCallbacks processor_callbacks.ProcessorCallbacks, debuggingSnapshotter debuggingsnapshot.DebuggingSnapshotter,
) (context.AutoscalingContext, error) {
	// Not enough buffer space causes the test to hang without printing any logs.
	// This is not useful.
	fakeRecorder := kube_record.NewFakeRecorder(100)
	fakeLogRecorder, err := utils.NewStatusMapRecorder(fakeClient, "kube-system", fakeRecorder, false, "my-cool-configmap")
	if err != nil {
		return context.AutoscalingContext{}, err
	}
	predicateChecker, err := predicatechecker.NewTestPredicateChecker()
	if err != nil {
		return context.AutoscalingContext{}, err
	}
	remainingPdbTracker := pdb.NewBasicRemainingPdbTracker()
	if debuggingSnapshotter == nil {
		debuggingSnapshotter = debuggingsnapshot.NewDebuggingSnapshotter(false)
	}
	clusterSnapshot, err := testsnapshot.NewTestSnapshot()
	if err != nil {
		return context.AutoscalingContext{}, err
	}
	return context.AutoscalingContext{
		AutoscalingOptions: options,
		AutoscalingKubeClients: context.AutoscalingKubeClients{
			ClientSet:      fakeClient,
			Recorder:       fakeRecorder,
			LogRecorder:    fakeLogRecorder,
			ListerRegistry: listers,
		},
		CloudProvider:        provider,
		PredicateChecker:     predicateChecker,
		ClusterSnapshot:      clusterSnapshot,
		ExpanderStrategy:     random.NewStrategy(),
		ProcessorCallbacks:   processorCallbacks,
		DebuggingSnapshotter: debuggingSnapshotter,
		RemainingPdbTracker:  remainingPdbTracker,
	}, nil
}

// MockAutoprovisioningNodeGroupManager is a mock node group manager to be used in tests
type MockAutoprovisioningNodeGroupManager struct {
	T           *testing.T
	ExtraGroups int
}

// CreateNodeGroup creates a new node group
func (p *MockAutoprovisioningNodeGroupManager) CreateNodeGroup(context *context.AutoscalingContext, nodeGroup cloudprovider.NodeGroup) (nodegroups.CreateNodeGroupResult, errors.AutoscalerError) {
	return p.createNodeGroup(context, nodeGroup)
}

// CreateNodeGroupAsync simulates async node group creation. Returns upcoming node groups, never calls initializer.
func (p *MockAutoprovisioningNodeGroupManager) CreateNodeGroupAsync(context *context.AutoscalingContext, nodeGroup cloudprovider.NodeGroup, nodeGroupInitializer nodegroups.AsyncNodeGroupInitializer) (nodegroups.CreateNodeGroupResult, errors.AutoscalerError) {
	return p.createNodeGroup(context, nodeGroup)
}

func (p *MockAutoprovisioningNodeGroupManager) createNodeGroup(context *context.AutoscalingContext, nodeGroup cloudprovider.NodeGroup) (nodegroups.CreateNodeGroupResult, errors.AutoscalerError) {
	newNodeGroup, err := nodeGroup.Create()
	assert.NoError(p.T, err)
	metrics.RegisterNodeGroupCreation()
	extraGroups := []cloudprovider.NodeGroup{}
	testGroup, ok := nodeGroup.(*testcloudprovider.TestNodeGroup)
	if !ok {
		return nodegroups.CreateNodeGroupResult{}, errors.ToAutoscalerError(errors.InternalError, fmt.Errorf("expected test node group, found %v", reflect.TypeOf(nodeGroup)))
	}
	testCloudProvider, ok := context.CloudProvider.(*testcloudprovider.TestCloudProvider)
	if !ok {
		return nodegroups.CreateNodeGroupResult{}, errors.ToAutoscalerError(errors.InternalError, fmt.Errorf("expected test CloudProvider, found %v", reflect.TypeOf(context.CloudProvider)))
	}
	for i := 0; i < p.ExtraGroups; i++ {
		extraNodeGroup := testCloudProvider.BuildNodeGroup(fmt.Sprintf("autoprovisioned-%s-%d", testGroup.MachineType(), i+1), 0, 1000, 0, false, true, testGroup.MachineType(), nil)
		assert.NoError(p.T, err)
		extraGroup, err := extraNodeGroup.Create()
		assert.NoError(p.T, err)
		metrics.RegisterNodeGroupCreation()
		extraGroups = append(extraGroups, extraGroup)
	}
	result := nodegroups.CreateNodeGroupResult{
		MainCreatedNodeGroup:   newNodeGroup,
		ExtraCreatedNodeGroups: extraGroups,
	}
	return result, nil
}

// RemoveUnneededNodeGroups removes uneeded node groups
func (p *MockAutoprovisioningNodeGroupManager) RemoveUnneededNodeGroups(context *context.AutoscalingContext) (removedNodeGroups []cloudprovider.NodeGroup, err error) {
	if !context.AutoscalingOptions.NodeAutoprovisioningEnabled {
		return nil, nil
	}
	removedNodeGroups = make([]cloudprovider.NodeGroup, 0)
	nodeGroups := context.CloudProvider.NodeGroups()
	for _, nodeGroup := range nodeGroups {
		if !nodeGroup.Autoprovisioned() {
			continue
		}
		targetSize, err := nodeGroup.TargetSize()
		assert.NoError(p.T, err)
		if targetSize > 0 {
			continue
		}
		nodes, err := nodeGroup.Nodes()
		assert.NoError(p.T, err)
		if len(nodes) > 0 {
			continue
		}
		err = nodeGroup.Delete()
		assert.NoError(p.T, err)
		removedNodeGroups = append(removedNodeGroups, nodeGroup)
	}
	return removedNodeGroups, nil
}

// CleanUp doesn't do anything; it's here to satisfy the interface
func (p *MockAutoprovisioningNodeGroupManager) CleanUp() {
}

// MockAutoprovisioningNodeGroupListProcessor is a fake node group list processor to be used in tests
type MockAutoprovisioningNodeGroupListProcessor struct {
	T *testing.T
}

// Process extends the list of node groups
func (p *MockAutoprovisioningNodeGroupListProcessor) Process(context *context.AutoscalingContext, nodeGroups []cloudprovider.NodeGroup, nodeInfos map[string]*framework.NodeInfo,
	unschedulablePods []*apiv1.Pod,
) ([]cloudprovider.NodeGroup, map[string]*framework.NodeInfo, error) {
	machines, err := context.CloudProvider.GetAvailableMachineTypes()
	assert.NoError(p.T, err)

	bestLabels := labels.BestLabelSet(unschedulablePods)
	for _, machineType := range machines {
		nodeGroup, err := context.CloudProvider.NewNodeGroup(machineType, bestLabels, map[string]string{}, []apiv1.Taint{}, map[string]resource.Quantity{})
		assert.NoError(p.T, err)
		nodeInfo, err := nodeGroup.TemplateNodeInfo()
		assert.NoError(p.T, err)
		nodeInfos[nodeGroup.Id()] = nodeInfo
		nodeGroups = append(nodeGroups, nodeGroup)
	}
	return nodeGroups, nodeInfos, nil
}

// CleanUp doesn't do anything; it's here to satisfy the interface
func (p *MockAutoprovisioningNodeGroupListProcessor) CleanUp() {
}

// MockBinpackingLimiter is a fake BinpackingLimiter to be used in tests.
type MockBinpackingLimiter struct {
	requiredExpansionOptions int
}

// InitBinpacking initialises the MockBinpackingLimiter and sets requiredExpansionOptions to 1.
func (p *MockBinpackingLimiter) InitBinpacking(context *context.AutoscalingContext, nodeGroups []cloudprovider.NodeGroup) {
	p.requiredExpansionOptions = 1
}

// MarkProcessed is here to satisfy the interface.
func (p *MockBinpackingLimiter) MarkProcessed(context *context.AutoscalingContext, nodegroupId string) {
}

// StopBinpacking stops the binpacking early, if we already have requiredExpansionOptions i.e. 1.
func (p *MockBinpackingLimiter) StopBinpacking(context *context.AutoscalingContext, evaluatedOptions []expander.Option) bool {
	return len(evaluatedOptions) == p.requiredExpansionOptions
}

// FinalizeBinpacking is here to satisfy the interface.
func (p *MockBinpackingLimiter) FinalizeBinpacking(context *context.AutoscalingContext, finalOptions []expander.Option) {
}

// NewBackoff creates a new backoff object
func NewBackoff() backoff.Backoff {
	return backoff.NewIdBasedExponentialBackoff(5*time.Minute, /*InitialNodeGroupBackoffDuration*/
		30*time.Minute /*MaxNodeGroupBackoffDuration*/, 3*time.Hour /*NodeGroupBackoffResetTimeout*/)
}

// To implement expander.Strategy, BestOption method must have a struct receiver.
// This prevents it from modifying fields of reportingStrategy, so we need a thin
// pointer wrapper for mutable parts.
type expanderResults struct {
	inputOptions []GroupSizeChange
}

// MockReportingStrategy implements expander.Strategy
type MockReportingStrategy struct {
	defaultStrategy expander.Strategy
	optionToChoose  *GroupSizeChange
	t               *testing.T
	results         *expanderResults
}

// NewMockRepotingStrategy creates an expander strategy with reporting and mocking capabilities.
func NewMockRepotingStrategy(t *testing.T, optionToChoose *GroupSizeChange) *MockReportingStrategy {
	return &MockReportingStrategy{
		defaultStrategy: random.NewStrategy(),
		results:         &expanderResults{},
		optionToChoose:  optionToChoose,
		t:               t,
	}
}

// LastInputOptions provides access to expansion options passed as an input in recent strategy execution
func (r *MockReportingStrategy) LastInputOptions() []GroupSizeChange {
	return r.results.inputOptions
}

// BestOption satisfies the Strategy interface. Picks the best option from those passed as an argument.
// When parameter optionToChoose is defined, it's picked as the best one.
// Otherwise, random option is used.
func (r *MockReportingStrategy) BestOption(options []expander.Option, nodeInfo map[string]*framework.NodeInfo) *expander.Option {
	r.results.inputOptions = expanderOptionsToGroupSizeChanges(options)
	if r.optionToChoose == nil {
		return r.defaultStrategy.BestOption(options, nodeInfo)
	}
	for _, option := range options {
		groupSizeChange := expanderOptionToGroupSizeChange(option)
		if groupSizeChange == *r.optionToChoose {
			return &option
		}
	}
	assert.Fail(r.t, "did not find expansionOptionToChoose %+v", r.optionToChoose)
	return nil
}

func expanderOptionsToGroupSizeChanges(options []expander.Option) []GroupSizeChange {
	groupSizeChanges := make([]GroupSizeChange, 0, len(options))
	for _, option := range options {
		groupSizeChange := expanderOptionToGroupSizeChange(option)
		groupSizeChanges = append(groupSizeChanges, groupSizeChange)
	}
	return groupSizeChanges
}

func expanderOptionToGroupSizeChange(option expander.Option) GroupSizeChange {
	groupName := option.NodeGroup.Id()
	groupSizeIncrement := option.NodeCount
	scaleUpOption := GroupSizeChange{GroupName: groupName, SizeChange: groupSizeIncrement}
	return scaleUpOption
}
