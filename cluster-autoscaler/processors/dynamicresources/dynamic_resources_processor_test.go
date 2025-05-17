/*
Copyright 2025 The Kubernetes Authors.

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

package dynamicresources

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	apiv1 "k8s.io/api/core/v1"
	resourceapi "k8s.io/api/resource/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	ca_context "k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	testutils "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	klog "k8s.io/klog/v2"
)

func TestDynamicResourcesProcessor(t *testing.T) {
	start := time.Now()
	later := start.Add(10 * time.Minute)
	onScaleUpMock := &onScaleUpMock{}
	onScaleDownMock := &onScaleDownMock{}
	onNodeGroupCreateMock := &onNodeGroupCreateMock{}
	onNodeGroupDeleteMock := &onNodeGroupDeleteMock{}
	deleteFinished := make(chan bool, 1)

	tn1 := testutils.BuildTestNode(
		"readynodewithreadyresources",
		1000,
		1000,
	)
	tn1.Status.Conditions = []apiv1.NodeCondition{
		{
			Type:              apiv1.NodeReady,
			Status:            apiv1.ConditionTrue,
			LastHeartbeatTime: metav1.NewTime(later),
		},
	}
	tn2 := testutils.BuildTestNode(
		"readynodewithunreadyresources",
		1000,
		1000,
	)
	tn2.Status.Conditions = []apiv1.NodeCondition{
		{
			Type:              apiv1.NodeReady,
			Status:            apiv1.ConditionTrue,
			LastHeartbeatTime: metav1.NewTime(later),
		},
	}
	tn3 := testutils.BuildTestNode(
		"readynodewithoutresources",
		1000,
		1000,
	)
	tn3.Status.Conditions = []apiv1.NodeCondition{
		{
			Type:              apiv1.NodeReady,
			Status:            apiv1.ConditionTrue,
			LastHeartbeatTime: metav1.NewTime(later),
		},
	}
	tn4 := testutils.BuildTestNode(
		"unreadynodewithoutresources",
		1000,
		1000,
	)
	tn4.Status.Conditions = []apiv1.NodeCondition{
		{
			Type:              apiv1.NodeReady,
			Status:            apiv1.ConditionFalse,
			LastHeartbeatTime: metav1.NewTime(later),
		},
	}
	tn5 := testutils.BuildTestNode(
		"nodenothandledbyautoscaler",
		1000,
		1000,
	)
	tn5.Status.Conditions = []apiv1.NodeCondition{
		{
			Type:              apiv1.NodeReady,
			Status:            apiv1.ConditionTrue,
			LastHeartbeatTime: metav1.NewTime(later),
		},
	}

	expectedReadiness := make(map[string]bool)
	expectedReadiness["readynodewithreadyresources"] = true
	expectedReadiness["readynodewithunreadyresources"] = false
	expectedReadiness["readynodewithoutresources"] = true
	expectedReadiness["unreadynodewithoutresources"] = false
	expectedReadiness["nodenothandledbyautoscaler"] = true

	testResourceSlices := []*resourceapi.ResourceSlice{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "slice1",
			},
			Spec: resourceapi.ResourceSliceSpec{
				NodeName: "readynodewithreadyresources",
			},
		},
	}

	tni1 := framework.NewTestNodeInfo(tn1)
	tni3 := framework.NewTestNodeInfo(tn3)

	tni1.LocalResourceSlices = testResourceSlices

	provider := testprovider.NewTestAutoprovisioningCloudProvider(
		func(id string, delta int) error {
			return onScaleUpMock.ScaleUp(id, delta)
		}, func(id string, name string) error {
			ret := onScaleDownMock.ScaleDown(id, name)
			deleteFinished <- true
			return ret
		}, func(id string) error {
			return onNodeGroupCreateMock.Create(id)
		}, func(id string) error {
			return onNodeGroupDeleteMock.Delete(id)
		},
		[]string{"tni1", "tni3"},
		map[string]*framework.NodeInfo{"ng1": tni1, "ng3": tni3},
	)

	provider.AddAutoprovisionedNodeGroup("ng1", 0, 10, 2, "ng1")
	provider.AddAutoprovisionedNodeGroup("ng3", 0, 10, 2, "ng3")
	provider.AddNode("ng1", tn1)
	provider.AddNode("ng1", tn2)
	provider.AddNode("ng3", tn3)
	provider.AddNode("ng3", tn4)

	autoscalingContext := ca_context.AutoscalingContext{
		CloudProvider: provider,
	}

	processor := NewDefaultDynamicResourcesProcessor()

	newAllNodes, newReadyNodes, err := processor.FilterOutNodesWithUnreadyResources(
		&autoscalingContext,
		[]*apiv1.Node{tn1, tn2, tn3, tn4, tn5},
		[]*apiv1.Node{tn1, tn2, tn3, tn5},
		testResourceSlices,
	)

	assert.NoError(t, err)

	foundInReady := make(map[string]bool)
	for _, node := range newReadyNodes {
		foundInReady[node.Name] = true
		assert.True(t, expectedReadiness[node.Name], fmt.Sprintf("Node %s found in ready nodes list (it shouldn't be there)", node.Name))
	}
	for nodeName, expected := range expectedReadiness {
		if expected {
			assert.True(t, foundInReady[nodeName], fmt.Sprintf("Node %s expected ready, but not found in ready nodes list", nodeName))
		}
	}
	for _, node := range newAllNodes {
		assert.Equal(t, len(node.Status.Conditions), 1)
		if expectedReadiness[node.Name] {
			assert.Equal(t, node.Status.Conditions[0].Status, apiv1.ConditionTrue, fmt.Sprintf("Unexpected ready condition value for node %s", node.Name))
		} else {
			assert.Equal(t, node.Status.Conditions[0].Status, apiv1.ConditionFalse, fmt.Sprintf("Unexpected ready condition value for node %s", node.Name))
		}
	}
}

type onScaleUpMock struct {
	mock.Mock
}

func (m *onScaleUpMock) ScaleUp(id string, delta int) error {
	klog.Infof("Scale up: %v %v", id, delta)
	args := m.Called(id, delta)
	return args.Error(0)
}

type onScaleDownMock struct {
	mock.Mock
}

func (m *onScaleDownMock) ScaleDown(id string, name string) error {
	klog.Infof("Scale down: %v %v", id, name)
	args := m.Called(id, name)
	return args.Error(0)
}

type onNodeGroupCreateMock struct {
	mock.Mock
}

func (m *onNodeGroupCreateMock) Create(id string) error {
	klog.Infof("Create group: %v", id)
	args := m.Called(id)
	return args.Error(0)
}

type onNodeGroupDeleteMock struct {
	mock.Mock
}

func (m *onNodeGroupDeleteMock) Delete(id string) error {
	klog.Infof("Delete group: %v", id)
	args := m.Called(id)
	return args.Error(0)
}
