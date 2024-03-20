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

package checkcapacity

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/apis/provisioningrequest/autoscaling.x-k8s.io/v1beta1"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	. "k8s.io/autoscaler/cluster-autoscaler/core/test"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/pods"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqclient"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqwrapper"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/scheduling"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

func TestScaleUp(t *testing.T) {
	allNodes := []*apiv1.Node{}
	for i := 0; i < 100; i++ {
		name := fmt.Sprintf("test-cpu-node-%d", i)
		node := BuildTestNode(name, 100, 10)
		allNodes = append(allNodes, node)
	}
	for i := 0; i < 100; i++ {
		name := fmt.Sprintf("test-mem-node-%d", i)
		node := BuildTestNode(name, 1, 1000)
		allNodes = append(allNodes, node)
	}
	newCpuProvReq := provreqwrapper.BuildTestProvisioningRequest("ns", "newCpuProvReq", "5m", "5", "", int32(100), false, time.Now(), v1beta1.ProvisioningClassCheckCapacity)
	newMemProvReq := provreqwrapper.BuildTestProvisioningRequest("ns", "newMemProvReq", "1m", "100", "", int32(100), false, time.Now(), v1beta1.ProvisioningClassCheckCapacity)
	bookedCapacityProvReq := provreqwrapper.BuildTestProvisioningRequest("ns", "bookedCapacity", "1m", "200", "", int32(100), false, time.Now(), v1beta1.ProvisioningClassCheckCapacity)
	bookedCapacityProvReq.SetConditions([]metav1.Condition{{Type: v1beta1.Provisioned, Status: metav1.ConditionTrue, LastTransitionTime: metav1.Now()}})
	expiredProvReq := provreqwrapper.BuildTestProvisioningRequest("ns", "bookedCapacity", "1m", "200", "", int32(100), false, time.Now(), v1beta1.ProvisioningClassCheckCapacity)
	expiredProvReq.SetConditions([]metav1.Condition{{Type: v1beta1.BookingExpired, Status: metav1.ConditionTrue, LastTransitionTime: metav1.Now()}})
	differentProvReqClass := provreqwrapper.BuildTestProvisioningRequest("ns", "differentProvReqClass", "1", "1", "", int32(5), false, time.Now(), v1beta1.ProvisioningClassAtomicScaleUp)
	testCases := []struct {
		name             string
		provReqs         []*provreqwrapper.ProvisioningRequest
		provReqToScaleUp *provreqwrapper.ProvisioningRequest
		scaleUpResult    status.ScaleUpResult
		err              bool
	}{
		{
			name:     "no ProvisioningRequests",
			provReqs: []*provreqwrapper.ProvisioningRequest{},
		},
		{
			name:             "one ProvisioningRequest",
			provReqs:         []*provreqwrapper.ProvisioningRequest{newCpuProvReq},
			provReqToScaleUp: newCpuProvReq,
			scaleUpResult:    status.ScaleUpSuccessful,
		},
		{
			name:             "capacity in the cluster is booked",
			provReqs:         []*provreqwrapper.ProvisioningRequest{newMemProvReq, bookedCapacityProvReq},
			provReqToScaleUp: newMemProvReq,
			scaleUpResult:    status.ScaleUpNoOptionsAvailable,
		},
		{
			name:             "pods from different ProvisioningRequest class",
			provReqs:         []*provreqwrapper.ProvisioningRequest{newCpuProvReq, bookedCapacityProvReq, differentProvReqClass},
			provReqToScaleUp: differentProvReqClass,
			err:              true,
		},
		{
			name:             "some capacity is booked, succesfull ScaleUp",
			provReqs:         []*provreqwrapper.ProvisioningRequest{newCpuProvReq, bookedCapacityProvReq, differentProvReqClass},
			provReqToScaleUp: newCpuProvReq,
			scaleUpResult:    status.ScaleUpSuccessful,
		},
	}
	for _, tc := range testCases {
		tc := tc
		allNodes := allNodes
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			provider := testprovider.NewTestCloudProvider(nil, nil)
			autoscalingContext, err := NewScaleTestAutoscalingContext(config.AutoscalingOptions{}, &fake.Clientset{}, nil, provider, nil, nil)
			assert.NoError(t, err)
			clustersnapshot.InitializeClusterSnapshotOrDie(t, autoscalingContext.ClusterSnapshot, allNodes, nil)
			prPods, err := pods.PodsForProvisioningRequest(tc.provReqToScaleUp)
			assert.NoError(t, err)
			orchestrator := &provReqOrchestrator{
				initialized: true,
				context:     &autoscalingContext,
				client:      provreqclient.NewFakeProvisioningRequestClient(context.Background(), t, tc.provReqs...),
				injector:    scheduling.NewHintingSimulator(autoscalingContext.PredicateChecker),
			}
			st, err := orchestrator.ScaleUp(prPods, []*apiv1.Node{}, []*v1.DaemonSet{}, map[string]*framework.NodeInfo{})
			if !tc.err {
				assert.NoError(t, err)
				assert.Equal(t, tc.scaleUpResult, st.Result)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
