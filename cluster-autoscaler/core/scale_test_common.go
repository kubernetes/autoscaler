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
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate/utils"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	"k8s.io/autoscaler/cluster-autoscaler/expander/random"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	processor_callbacks "k8s.io/autoscaler/cluster-autoscaler/processors/callbacks"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroups"
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
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"
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
	name   string
	cpu    int64
	memory int64
	gpu    int64
	node   string
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

	//expectedScaleUpOptions []groupSizeChange // we expect that all those options should be included in expansion options passed to expander strategy
	//expectedFinalScaleUp   groupSizeChange   // we expect this to be delivered via scale-up event
	expectedScaleDowns []string
}

type scaleTestResults struct {
	expansionOptions []groupSizeChange
	finalOption      groupSizeChange
	scaleUpStatus    *status.ScaleUpStatus
	noScaleUpReason  string
	finalScaleDowns  []string
	events           []string
}

// NewScaleTestAutoscalingContext creates a new test autoscaling context for scaling tests.
func NewScaleTestAutoscalingContext(
	options config.AutoscalingOptions, fakeClient kube_client.Interface,
	listers kube_util.ListerRegistry, provider cloudprovider.CloudProvider,
	processorCallbacks processor_callbacks.ProcessorCallbacks) context.AutoscalingContext {
	fakeRecorder := kube_record.NewFakeRecorder(5)
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", fakeRecorder, false)
	// Ignoring error here is safe - if a test doesn't specify valid estimatorName,
	// it either doesn't need one, or should fail when it turns out to be nil.
	estimatorBuilder, _ := estimator.NewEstimatorBuilder(options.EstimatorName)
	return context.AutoscalingContext{
		AutoscalingOptions: options,
		AutoscalingKubeClients: context.AutoscalingKubeClients{
			ClientSet:      fakeClient,
			Recorder:       fakeRecorder,
			LogRecorder:    fakeLogRecorder,
			ListerRegistry: listers,
		},
		CloudProvider:      provider,
		PredicateChecker:   simulator.NewTestPredicateChecker(),
		ExpanderStrategy:   random.NewStrategy(),
		EstimatorBuilder:   estimatorBuilder,
		ProcessorCallbacks: processorCallbacks,
	}
}

type mockAutoprovisioningNodeGroupManager struct {
	t *testing.T
}

func (p *mockAutoprovisioningNodeGroupManager) CreateNodeGroup(context *context.AutoscalingContext, nodeGroup cloudprovider.NodeGroup) (nodegroups.CreateNodeGroupResult, errors.AutoscalerError) {
	newNodeGroup, err := nodeGroup.Create()
	assert.NoError(p.t, err)
	metrics.RegisterNodeGroupCreation()
	result := nodegroups.CreateNodeGroupResult{
		MainCreatedNodeGroup: newNodeGroup,
	}
	return result, nil
}

func (p *mockAutoprovisioningNodeGroupManager) RemoveUnneededNodeGroups(context *context.AutoscalingContext) error {
	if !context.AutoscalingOptions.NodeAutoprovisioningEnabled {
		return nil
	}
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
	}
	return nil
}

func (p *mockAutoprovisioningNodeGroupManager) CleanUp() {
}

type mockAutoprovisioningNodeGroupListProcessor struct {
	t *testing.T
}

func (p *mockAutoprovisioningNodeGroupListProcessor) Process(context *context.AutoscalingContext, nodeGroups []cloudprovider.NodeGroup, nodeInfos map[string]*schedulernodeinfo.NodeInfo,
	unschedulablePods []*apiv1.Pod) ([]cloudprovider.NodeGroup, map[string]*schedulernodeinfo.NodeInfo, error) {

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
