/*
Copyright The Kubernetes Authors.

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

package nodegroupchange

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	apiv1 "k8s.io/api/core/v1"
	resourceapi "k8s.io/api/resource/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
)

type mockMetrics struct {
	mock.Mock
}

func (m *mockMetrics) RegisterFailedScaleUp(reason metrics.FailedScaleUpReason, gpuResourceName, gpuType, draDrivers string) {
	m.Called(reason, gpuResourceName, gpuType, draDrivers)
}

func (m *mockMetrics) RegisterFailedNodeCreations(reason metrics.FailedScaleUpReason, nodesCount int) {
	m.Called(reason, nodesCount)
}

func TestRegisterFailedScaleUpDirectCall(t *testing.T) {
	now := time.Now()
	gpuNode := &apiv1.Node{
		Status: apiv1.NodeStatus{
			Capacity: apiv1.ResourceList{
				apiv1.ResourceName("nvidia.com/gpu"): resource.MustParse("1"),
			},
		},
	}
	gpuNode.Labels = map[string]string{
		"TestGPULabel/accelerator": "nvidia-tesla-k80",
	}

	slices := []*resourceapi.ResourceSlice{
		{
			Spec: resourceapi.ResourceSliceSpec{
				Driver: "dra.net",
			},
		},
	}

	provider := testprovider.NewTestCloudProviderBuilder().Build()
	provider.AddNodeGroup("ng1", 1, 10, 1)
	provider.SetMachineTemplates(map[string]*framework.NodeInfo{
		"ng1": framework.NewNodeInfo(gpuNode, slices),
	})

	mockMetricsObj := &mockMetrics{}
	mockMetricsObj.On("RegisterFailedScaleUp", metrics.FailedScaleUpReason("OUT_OF_RESOURCES"), "nvidia.com/gpu", "nvidia-tesla-k80", "dra.net").Return()
	mockMetricsObj.On("RegisterFailedNodeCreations", metrics.FailedScaleUpReason("OUT_OF_RESOURCES"), 3).Return()

	producer := NewNodeGroupChangeMetricsProducer(provider, mockMetricsObj)

	nodeGroup := provider.GetNodeGroup("ng1")
	errorInfo := cloudprovider.InstanceErrorInfo{
		ErrorCode: "OUT_OF_RESOURCES",
	}
	producer.RegisterFailedScaleUp(nodeGroup, 3, errorInfo, now)

	mockMetricsObj.AssertCalled(t, "RegisterFailedScaleUp", metrics.FailedScaleUpReason("OUT_OF_RESOURCES"), "nvidia.com/gpu", "nvidia-tesla-k80", "dra.net")
	mockMetricsObj.AssertCalled(t, "RegisterFailedNodeCreations", metrics.FailedScaleUpReason("OUT_OF_RESOURCES"), 3)
}

func TestRegisterFailedScaleUpCallViaObserversList(t *testing.T) {
	now := time.Now()

	provider := testprovider.NewTestCloudProviderBuilder().Build()
	provider.AddNodeGroup("ng2", 1, 10, 1)
	// No template node info for ng2, meaning GPU and DRA extraction should fall back to empty strings.
	provider.SetMachineTemplates(map[string]*framework.NodeInfo{})

	// Setup Mock Metrics Observer
	mockMetricsObj := &mockMetrics{}
	mockMetricsObj.On("RegisterFailedScaleUp", metrics.FailedScaleUpReason("TIMEOUT"), "", "", "").Return()
	mockMetricsObj.On("RegisterFailedNodeCreations", metrics.FailedScaleUpReason("TIMEOUT"), 5).Return()

	producer := NewNodeGroupChangeMetricsProducer(provider, mockMetricsObj)

	// Register with ObserversList
	list := NewNodeGroupChangeObserversList()
	list.Register(producer)

	// Trigger via list
	nodeGroup := provider.GetNodeGroup("ng2")
	errorInfo := cloudprovider.InstanceErrorInfo{
		ErrorCode: "TIMEOUT",
	}
	list.RegisterFailedScaleUp(nodeGroup, 5, errorInfo, now)

	// Assertions
	mockMetricsObj.AssertCalled(t, "RegisterFailedScaleUp", metrics.FailedScaleUpReason("TIMEOUT"), "", "", "")
	mockMetricsObj.AssertCalled(t, "RegisterFailedNodeCreations", metrics.FailedScaleUpReason("TIMEOUT"), 5)
}
