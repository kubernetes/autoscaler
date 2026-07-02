/*
Copyright 2016 The Kubernetes Authors.

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

package clusterstate

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	mockprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/mocks"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate/utils"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroupconfig"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroups/asyncnodegroups"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/client-go/kubernetes/fake"
	kube_record "k8s.io/client-go/tools/record"
)

// TestHandleInstanceCreationErrorsStartDeallocatedFailed verifies that deallocated VMs
// that fail to start trigger backoff with the expected Azure-specific error code.
func TestHandleInstanceCreationErrorsStartDeallocatedFailed(t *testing.T) {
	now := time.Now()

	provider := testprovider.NewTestCloudProviderBuilder().Build()
	mockedNodeGroup := &mockprovider.NodeGroup{}
	mockedNodeGroup.On("Id").Return("deallocate-pool")
	mockedNodeGroup.On("Nodes").Return([]cloudprovider.Instance{
		{
			Id: "vm1",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceCreating,
				ErrorInfo: &cloudprovider.InstanceErrorInfo{
					ErrorClass:   cloudprovider.OutOfResourcesErrorClass,
					ErrorCode:    "start-deallocated-failed",
					ErrorMessage: "Failed to start deallocated VM",
				},
			},
		},
		{
			Id: "vm2",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceCreating,
				ErrorInfo: &cloudprovider.InstanceErrorInfo{
					ErrorClass:   cloudprovider.OutOfResourcesErrorClass,
					ErrorCode:    "start-deallocated-failed",
					ErrorMessage: "Failed to start deallocated VM",
				},
			},
		},
	}, nil)
	mockedNodeGroup.On("Autoprovisioned").Return(false)
	mockedNodeGroup.On("TargetSize").Return(2, nil)
	node := BuildTestNode("deallocate-pool_1", 1000, 1000)
	mockedNodeGroup.On("TemplateNodeInfo").Return(framework.NewTestNodeInfo(node), nil)
	mockedNodeGroup.On("GetOptions", mock.Anything).Return(&config.NodeGroupAutoscalingOptions{}, nil)
	provider.InsertNodeGroup(mockedNodeGroup)

	fakeClient := &fake.Clientset{}
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false, "my-cool-configmap")
	mockMetrics := &mockMetrics{}
	mockMetrics.On("RegisterFailedScaleUp", mock.Anything, mock.Anything, mock.Anything).Return()
	clusterstate := newClusterStateRegistry(provider, ClusterStateRegistryConfig{}, fakeLogRecorder, newBackoff(), nodegroupconfig.NewDefaultNodeGroupConfigProcessor(config.NodeGroupAutoscalingOptions{MaxNodeProvisionTime: 15 * time.Minute}), asyncnodegroups.NewDefaultAsyncNodeGroupStateChecker(), mockMetrics)
	clusterstate.RegisterScaleUp(mockedNodeGroup, 2, now)

	// UpdateNodes will trigger handleInstanceCreationErrors.
	err := clusterstate.UpdateNodes([]*apiv1.Node{}, nil, now)
	assert.NoError(t, err)

	// Verify metrics were recorded.
	mockMetrics.AssertCalled(t, "RegisterFailedScaleUp", metrics.FailedScaleUpReason("start-deallocated-failed"), "", "")

	// Verify the node group is in backoff with the correct error code.
	safety := clusterstate.NodeGroupScaleUpSafety(mockedNodeGroup, now)
	assert.False(t, safety.SafeToScale, "node group should not be safe to scale")
	assert.True(t, safety.BackoffStatus.IsBackedOff, "node group should be backed off after start-deallocated-failed")
	assert.Equal(t, "start-deallocated-failed", safety.BackoffStatus.ErrorInfo.ErrorCode)
	assert.Equal(t, cloudprovider.OutOfResourcesErrorClass, safety.BackoffStatus.ErrorInfo.ErrorClass)
}
