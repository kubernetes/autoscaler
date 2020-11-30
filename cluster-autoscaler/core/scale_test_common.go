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

package core

import (
	"fmt"
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	testcloudprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate/utils"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	"k8s.io/autoscaler/cluster-autoscaler/expander/random"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/processors"
	processor_callbacks "k8s.io/autoscaler/cluster-autoscaler/processors/callbacks"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroups"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroupset"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodeinfos"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/autoscaler/cluster-autoscaler/utils/labels"

	"github.com/stretchr/testify/assert"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/utils/backoff"
	kube_client "k8s.io/client-go/kubernetes"
	kube_record "k8s.io/client-go/tools/record"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

type nodeConfig struct {
	name   string
	cpu    int64
	memory int64
	gpu    int64
	ready  bool
	group  string
}

type podConfig struct {
	name         string
	cpu          int64
	memory       int64
	gpu          int64
	node         string
	toleratesGpu bool
}

type groupSizeChange struct {
	groupName  string
	sizeChange int
}

type scaleTestConfig struct {
	nodes                   []nodeConfig
	pods                    []podConfig
	extraPods               []podConfig
	options                 config.AutoscalingOptions
	nodeDeletionTracker     *NodeDeletionTracker
	expansionOptionToChoose groupSizeChange // this will be selected by assertingStrategy.BestOption

	expectedScaleDowns     []string
	expectedScaleDownCount int
}

type scaleTestResults struct {
	expansionOptions []groupSizeChange
	finalOption      groupSizeChange
	noScaleUpReason  string
	finalScaleDowns  []string
	events           []string
	scaleUpStatus    scaleUpStatusInfo
}

// scaleUpStatusInfo is a simplified form of a ScaleUpStatus, to avoid mocking actual NodeGroup and Pod objects in test config.
type scaleUpStatusInfo struct {
	result                  status.ScaleUpResult
	podsTriggeredScaleUp    []string
	podsRemainUnschedulable []string
	podsAwaitEvaluation     []string
}

func (s *scaleUpStatusInfo) WasSuccessful() bool {
	return s.result == status.ScaleUpSuccessful
}

func extractPodNames(pods []*apiv1.Pod) []string {
	podNames := []string{}
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}

func simplifyScaleUpStatus(scaleUpStatus *status.ScaleUpStatus) scaleUpStatusInfo {
	remainUnschedulable := []string{}
	for _, nsi := range scaleUpStatus.PodsRemainUnschedulable {
		remainUnschedulable = append(remainUnschedulable, nsi.Pod.Name)
	}
	return scaleUpStatusInfo{
		result:                  scaleUpStatus.Result,
		podsTriggeredScaleUp:    extractPodNames(scaleUpStatus.PodsTriggeredScaleUp),
		podsRemainUnschedulable: remainUnschedulable,
		podsAwaitEvaluation:     extractPodNames(scaleUpStatus.PodsAwaitEvaluation),
	}
}

// NewTestProcessors returns a set of simple processors for use in tests.
func NewTestProcessors() *processors.AutoscalingProcessors {
	return &processors.AutoscalingProcessors{
		PodListProcessor:       NewFilterOutSchedulablePodListProcessor(),
		NodeGroupListProcessor: &nodegroups.NoOpNodeGroupListProcessor{},
		NodeGroupSetProcessor:  nodegroupset.NewDefaultNodeGroupSetProcessor([]string{}),
		// TODO(bskiba): change scale up test so that this can be a NoOpProcessor
		ScaleUpStatusProcessor:     &status.EventingScaleUpStatusProcessor{},
		ScaleDownStatusProcessor:   &status.NoOpScaleDownStatusProcessor{},
		AutoscalingStatusProcessor: &status.NoOpAutoscalingStatusProcessor{},
		NodeGroupManager:           nodegroups.NewDefaultNodeGroupManager(),
		NodeInfoProcessor:          nodeinfos.NewDefaultNodeInfoProcessor(),
	}
}

// NewScaleTestAutoscalingContext creates a new test autoscaling context for scaling tests.
func NewScaleTestAutoscalingContext(
	options config.AutoscalingOptions, fakeClient kube_client.Interface,
	listers kube_util.ListerRegistry, provider cloudprovider.CloudProvider,
	processorCallbacks processor_callbacks.ProcessorCallbacks) (context.AutoscalingContext, error) {
	// Not enough buffer space causes the test to hang without printing any logs.
	// This is not useful.
	fakeRecorder := kube_record.NewFakeRecorder(100)
	fakeLogRecorder, err := utils.NewStatusMapRecorder(fakeClient, "kube-system", fakeRecorder, false)
	if err != nil {
		return context.AutoscalingContext{}, err
	}
	// Ignoring error here is safe - if a test doesn't specify valid estimatorName,
	// it either doesn't need one, or should fail when it turns out to be nil.
	estimatorBuilder, _ := estimator.NewEstimatorBuilder(options.EstimatorName)
	predicateChecker, err := simulator.NewTestPredicateChecker()
	if err != nil {
		return context.AutoscalingContext{}, err
	}
	clusterSnapshot := simulator.NewBasicClusterSnapshot()
	return context.AutoscalingContext{
		AutoscalingOptions: options,
		AutoscalingKubeClients: context.AutoscalingKubeClients{
			ClientSet:      fakeClient,
			Recorder:       fakeRecorder,
			LogRecorder:    fakeLogRecorder,
			ListerRegistry: listers,
		},
		CloudProvider:      provider,
		PredicateChecker:   predicateChecker,
		ClusterSnapshot:    clusterSnapshot,
		ExpanderStrategy:   random.NewStrategy(),
		EstimatorBuilder:   estimatorBuilder,
		ProcessorCallbacks: processorCallbacks,
	}, nil
}

type mockAutoprovisioningNodeGroupManager struct {
	t           *testing.T
	extraGroups int
}

func (p *mockAutoprovisioningNodeGroupManager) CreateNodeGroup(context *context.AutoscalingContext, nodeGroup cloudprovider.NodeGroup) (nodegroups.CreateNodeGroupResult, errors.AutoscalerError) {
	newNodeGroup, err := nodeGroup.Create()
	assert.NoError(p.t, err)
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
	for i := 0; i < p.extraGroups; i++ {
		extraNodeGroup, err := testCloudProvider.NewNodeGroupWithId(
			testGroup.MachineType(),
			testGroup.Labels(),
			map[string]string{},
			[]apiv1.Taint{},
			map[string]resource.Quantity{},
			fmt.Sprintf("%d", i+1),
		)
		assert.NoError(p.t, err)
		extraGroup, err := extraNodeGroup.Create()
		assert.NoError(p.t, err)
		metrics.RegisterNodeGroupCreation()
		extraGroups = append(extraGroups, extraGroup)
	}
	result := nodegroups.CreateNodeGroupResult{
		MainCreatedNodeGroup:   newNodeGroup,
		ExtraCreatedNodeGroups: extraGroups,
	}
	return result, nil
}

func (p *mockAutoprovisioningNodeGroupManager) RemoveUnneededNodeGroups(context *context.AutoscalingContext) (removedNodeGroups []cloudprovider.NodeGroup, err error) {
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
		assert.NoError(p.t, err)
		if targetSize > 0 {
			continue
		}
		nodes, err := nodeGroup.Nodes()
		assert.NoError(p.t, err)
		if len(nodes) > 0 {
			continue
		}
		err = nodeGroup.Delete()
		assert.NoError(p.t, err)
		removedNodeGroups = append(removedNodeGroups, nodeGroup)
	}
	return removedNodeGroups, nil
}

func (p *mockAutoprovisioningNodeGroupManager) CleanUp() {
}

type mockAutoprovisioningNodeGroupListProcessor struct {
	t *testing.T
}

func (p *mockAutoprovisioningNodeGroupListProcessor) Process(context *context.AutoscalingContext, nodeGroups []cloudprovider.NodeGroup, nodeInfos map[string]*schedulerframework.NodeInfo,
	unschedulablePods []*apiv1.Pod) ([]cloudprovider.NodeGroup, map[string]*schedulerframework.NodeInfo, error) {

	machines, err := context.CloudProvider.GetAvailableMachineTypes()
	assert.NoError(p.t, err)

	bestLabels := labels.BestLabelSet(unschedulablePods)
	for _, machineType := range machines {
		nodeGroup, err := context.CloudProvider.NewNodeGroup(machineType, bestLabels, map[string]string{}, []apiv1.Taint{}, map[string]resource.Quantity{})
		assert.NoError(p.t, err)
		nodeInfo, err := nodeGroup.TemplateNodeInfo()
		assert.NoError(p.t, err)
		nodeInfos[nodeGroup.Id()] = nodeInfo
		nodeGroups = append(nodeGroups, nodeGroup)
	}
	return nodeGroups, nodeInfos, nil
}

func (p *mockAutoprovisioningNodeGroupListProcessor) CleanUp() {
}

func newBackoff() backoff.Backoff {
	return backoff.NewIdBasedExponentialBackoff(clusterstate.InitialNodeGroupBackoffDuration, clusterstate.MaxNodeGroupBackoffDuration, clusterstate.NodeGroupBackoffResetTimeout)
}
