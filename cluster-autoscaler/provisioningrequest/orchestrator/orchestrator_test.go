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

package orchestrator

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/apis/provisioningrequest/autoscaling.x-k8s.io/v1beta1"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	. "k8s.io/autoscaler/cluster-autoscaler/core/test"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/checkcapacity"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/pods"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqclient"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqwrapper"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

func TestScaleUp(t *testing.T) {
	// Set up a cluster with 200 nodes:
	// - 100 nodes with high cpu, low memory
	// - 100 nodes with high memory, low cpu
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

	// Active check capacity requests.
	newCheckCapacityCpuProvReq := provreqwrapper.BuildValidTestProvisioningRequestFromOptions(
		provreqwrapper.TestProvReqOptions{
			Name:     "newCheckCapacityCpuProvReq",
			CPU:      "5m",
			Memory:   "5",
			PodCount: int32(100),
			Class:    v1beta1.ProvisioningClassCheckCapacity,
		})

	newCheckCapacityMemProvReq := provreqwrapper.BuildValidTestProvisioningRequestFromOptions(
		provreqwrapper.TestProvReqOptions{
			Name:     "newCheckCapacityMemProvReq",
			CPU:      "1m",
			Memory:   "100",
			PodCount: int32(100),
			Class:    v1beta1.ProvisioningClassCheckCapacity,
		})

	// Active atomic scale up request.
	atomicScaleUpProvReq := provreqwrapper.BuildValidTestProvisioningRequestFromOptions(
		provreqwrapper.TestProvReqOptions{
			Name:     "atomicScaleUpProvReq",
			CPU:      "1",
			Memory:   "1",
			PodCount: int32(5),
			Class:    v1beta1.ProvisioningClassAtomicScaleUp,
		})

	// Already provisioned provisioning request - capacity should be booked before processing a new request.
	bookedCapacityProvReq := provreqwrapper.BuildValidTestProvisioningRequestFromOptions(
		provreqwrapper.TestProvReqOptions{
			Name:     "bookedCapacityProvReq",
			CPU:      "1m",
			Memory:   "200",
			PodCount: int32(100),
			Class:    v1beta1.ProvisioningClassCheckCapacity,
		})
	bookedCapacityProvReq.SetConditions([]metav1.Condition{{Type: v1beta1.Provisioned, Status: metav1.ConditionTrue, LastTransitionTime: metav1.Now()}})

	// Expired provisioning request - should be ignored.
	expiredProvReq := provreqwrapper.BuildValidTestProvisioningRequestFromOptions(
		provreqwrapper.TestProvReqOptions{
			Name:     "expiredProvReq",
			CPU:      "1m",
			Memory:   "200",
			PodCount: int32(100),
			Class:    v1beta1.ProvisioningClassCheckCapacity,
		})
	expiredProvReq.SetConditions([]metav1.Condition{{Type: v1beta1.BookingExpired, Status: metav1.ConditionTrue, LastTransitionTime: metav1.Now()}})

	// Unsupported provisioning request - should be ignored.
	unsupportedProvReq := provreqwrapper.BuildValidTestProvisioningRequestFromOptions(
		provreqwrapper.TestProvReqOptions{
			Name:     "unsupportedProvReq",
			CPU:      "1",
			Memory:   "1",
			PodCount: int32(5),
			Class:    "very much unsupported",
		})

	testCases := []struct {
		name             string
		provReqs         []*provreqwrapper.ProvisioningRequest
		provReqToScaleUp *provreqwrapper.ProvisioningRequest
		scaleUpResult    status.ScaleUpResult
		err              bool
	}{
		{
			name:          "no ProvisioningRequests",
			provReqs:      []*provreqwrapper.ProvisioningRequest{},
			scaleUpResult: status.ScaleUpNotTried,
		},
		{
			name:             "one ProvisioningRequest of check capacity class",
			provReqs:         []*provreqwrapper.ProvisioningRequest{newCheckCapacityCpuProvReq},
			provReqToScaleUp: newCheckCapacityCpuProvReq,
			scaleUpResult:    status.ScaleUpSuccessful,
		},
		{
			name:             "capacity in the cluster is booked",
			provReqs:         []*provreqwrapper.ProvisioningRequest{newCheckCapacityMemProvReq, bookedCapacityProvReq},
			provReqToScaleUp: newCheckCapacityMemProvReq,
			scaleUpResult:    status.ScaleUpNoOptionsAvailable,
		},
		{
			name:             "unsupported ProvisioningRequest is ignored",
			provReqs:         []*provreqwrapper.ProvisioningRequest{newCheckCapacityCpuProvReq, bookedCapacityProvReq, atomicScaleUpProvReq, unsupportedProvReq},
			provReqToScaleUp: unsupportedProvReq,
			scaleUpResult:    status.ScaleUpNotTried,
		},
		{
			name:             "some capacity is pre-booked, successful capacity check",
			provReqs:         []*provreqwrapper.ProvisioningRequest{newCheckCapacityCpuProvReq, bookedCapacityProvReq, atomicScaleUpProvReq},
			provReqToScaleUp: newCheckCapacityCpuProvReq,
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

			client := provreqclient.NewFakeProvisioningRequestClient(context.Background(), t, tc.provReqs...)
			orchestrator := &provReqOrchestrator{
				client:              client,
				provisioningClasses: []provisioningClass{checkcapacity.New(client)},
			}
			orchestrator.Initialize(&autoscalingContext, nil, nil, nil, taints.TaintConfig{})
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
