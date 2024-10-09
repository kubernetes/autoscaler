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
	"time"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/apis/provisioningrequest/autoscaling.x-k8s.io/v1"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	. "k8s.io/autoscaler/cluster-autoscaler/core/test"
	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroupconfig"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodeinfosprovider"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/besteffortatomic"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/checkcapacity"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/pods"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqclient"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqwrapper"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/autoscaler/cluster-autoscaler/utils/units"
	"k8s.io/client-go/kubernetes/fake"
)

func TestScaleUp(t *testing.T) {
	// Set up a cluster with 200 nodes:
	// - 100 nodes with high cpu, low memory in autoscaled group with max 150
	// - 100 nodes with high memory, low cpu not in autoscaled group
	now := time.Now()
	allNodes := []*apiv1.Node{}
	for i := 0; i < 100; i++ {
		name := fmt.Sprintf("test-cpu-node-%d", i)
		node := BuildTestNode(name, 100, 10)
		SetNodeReadyState(node, true, now.Add(-2*time.Minute))
		allNodes = append(allNodes, node)
	}
	for i := 0; i < 100; i++ {
		name := fmt.Sprintf("test-mem-node-%d", i)
		node := BuildTestNode(name, 1, 1000)
		SetNodeReadyState(node, true, now.Add(-2*time.Minute))
		allNodes = append(allNodes, node)
	}

	// Active check capacity requests.
	newCheckCapacityCpuProvReq := provreqwrapper.BuildValidTestProvisioningRequestFromOptions(
		provreqwrapper.TestProvReqOptions{
			Name:     "newCheckCapacityCpuProvReq",
			CPU:      "5m",
			Memory:   "5",
			PodCount: int32(100),
			Class:    v1.ProvisioningClassCheckCapacity,
		})

	newCheckCapacityMemProvReq := provreqwrapper.BuildValidTestProvisioningRequestFromOptions(
		provreqwrapper.TestProvReqOptions{
			Name:     "newCheckCapacityMemProvReq",
			CPU:      "1m",
			Memory:   "100",
			PodCount: int32(100),
			Class:    v1.ProvisioningClassCheckCapacity,
		})

	// Active atomic scale up requests.
	atomicScaleUpProvReq := provreqwrapper.BuildValidTestProvisioningRequestFromOptions(
		provreqwrapper.TestProvReqOptions{
			Name:     "atomicScaleUpProvReq",
			CPU:      "5m",
			Memory:   "5",
			PodCount: int32(5),
			Class:    v1.ProvisioningClassBestEffortAtomicScaleUp,
		})
	largeAtomicScaleUpProvReq := provreqwrapper.BuildValidTestProvisioningRequestFromOptions(
		provreqwrapper.TestProvReqOptions{
			Name:     "largeAtomicScaleUpProvReq",
			CPU:      "1m",
			Memory:   "100",
			PodCount: int32(100),
			Class:    v1.ProvisioningClassBestEffortAtomicScaleUp,
		})
	impossibleAtomicScaleUpReq := provreqwrapper.BuildValidTestProvisioningRequestFromOptions(
		provreqwrapper.TestProvReqOptions{
			Name:     "impossibleAtomicScaleUpRequest",
			CPU:      "1m",
			Memory:   "1",
			PodCount: int32(5001),
			Class:    v1.ProvisioningClassBestEffortAtomicScaleUp,
		})
	possibleAtomicScaleUpReq := provreqwrapper.BuildValidTestProvisioningRequestFromOptions(
		provreqwrapper.TestProvReqOptions{
			Name:     "possibleAtomicScaleUpReq",
			CPU:      "100m",
			Memory:   "1",
			PodCount: int32(120),
			Class:    v1.ProvisioningClassBestEffortAtomicScaleUp,
		})
	autoprovisioningAtomicScaleUpReq := provreqwrapper.BuildValidTestProvisioningRequestFromOptions(
		provreqwrapper.TestProvReqOptions{
			Name:     "autoprovisioningAtomicScaleUpReq",
			CPU:      "100m",
			Memory:   "100",
			PodCount: int32(5),
			Class:    v1.ProvisioningClassBestEffortAtomicScaleUp,
		})

	// Already provisioned provisioning request - capacity should be booked before processing a new request.
	// Books 20 out of 100 high-memory nodes.
	bookedCapacityProvReq := provreqwrapper.BuildValidTestProvisioningRequestFromOptions(
		provreqwrapper.TestProvReqOptions{
			Name:     "bookedCapacityProvReq",
			CPU:      "1m",
			Memory:   "200",
			PodCount: int32(100),
			Class:    v1.ProvisioningClassCheckCapacity,
		})
	bookedCapacityProvReq.SetConditions([]metav1.Condition{{Type: v1.Provisioned, Status: metav1.ConditionTrue, LastTransitionTime: metav1.Now()}})

	// Expired provisioning request - should be ignored.
	expiredProvReq := provreqwrapper.BuildValidTestProvisioningRequestFromOptions(
		provreqwrapper.TestProvReqOptions{
			Name:     "expiredProvReq",
			CPU:      "1m",
			Memory:   "200",
			PodCount: int32(100),
			Class:    v1.ProvisioningClassCheckCapacity,
		})
	expiredProvReq.SetConditions([]metav1.Condition{{Type: v1.BookingExpired, Status: metav1.ConditionTrue, LastTransitionTime: metav1.Now()}})

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
		autoprovisioning bool
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
			name:             "one ProvisioningRequest of atomic scale up class",
			provReqs:         []*provreqwrapper.ProvisioningRequest{atomicScaleUpProvReq},
			provReqToScaleUp: atomicScaleUpProvReq,
			scaleUpResult:    status.ScaleUpNotNeeded,
		},
		{
			name:             "capacity is there, check-capacity class",
			provReqs:         []*provreqwrapper.ProvisioningRequest{newCheckCapacityMemProvReq},
			provReqToScaleUp: newCheckCapacityMemProvReq,
			scaleUpResult:    status.ScaleUpSuccessful,
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
		{
			name:             "some capacity is pre-booked, atomic scale-up not needed",
			provReqs:         []*provreqwrapper.ProvisioningRequest{bookedCapacityProvReq, atomicScaleUpProvReq},
			provReqToScaleUp: atomicScaleUpProvReq,
			scaleUpResult:    status.ScaleUpNotNeeded,
		},
		{
			name:             "capacity is there, large atomic scale-up request doesn't require scale-up",
			provReqs:         []*provreqwrapper.ProvisioningRequest{largeAtomicScaleUpProvReq},
			provReqToScaleUp: largeAtomicScaleUpProvReq,
			scaleUpResult:    status.ScaleUpNotNeeded,
		},
		{
			name:             "impossible atomic scale-up request doesn't trigger scale-up",
			provReqs:         []*provreqwrapper.ProvisioningRequest{impossibleAtomicScaleUpReq},
			provReqToScaleUp: impossibleAtomicScaleUpReq,
			scaleUpResult:    status.ScaleUpNoOptionsAvailable,
		},
		{
			name:             "possible atomic scale-up request triggers scale-up",
			provReqs:         []*provreqwrapper.ProvisioningRequest{possibleAtomicScaleUpReq},
			provReqToScaleUp: possibleAtomicScaleUpReq,
			scaleUpResult:    status.ScaleUpSuccessful,
		},
		{
			name:             "autoprovisioning atomic scale-up request triggers scale-up",
			provReqs:         []*provreqwrapper.ProvisioningRequest{autoprovisioningAtomicScaleUpReq},
			provReqToScaleUp: autoprovisioningAtomicScaleUpReq,
			autoprovisioning: true,
			scaleUpResult:    status.ScaleUpSuccessful,
		},
	}
	for _, tc := range testCases {
		tc := tc
		allNodes := allNodes
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			prPods, err := pods.PodsForProvisioningRequest(tc.provReqToScaleUp)
			assert.NoError(t, err)

			onScaleUpFunc := func(name string, n int) error {
				if tc.scaleUpResult == status.ScaleUpSuccessful {
					return nil
				}
				return fmt.Errorf("unexpected scale-up of %s by %d", name, n)
			}
			orchestrator, nodeInfos := setupTest(t, allNodes, tc.provReqs, onScaleUpFunc, tc.autoprovisioning)

			st, err := orchestrator.ScaleUp(prPods, []*apiv1.Node{}, []*appsv1.DaemonSet{}, nodeInfos, false)
			if !tc.err {
				assert.NoError(t, err)
				if tc.scaleUpResult != st.Result && len(st.PodsRemainUnschedulable) > 0 {
					// We expected all pods to be scheduled, but some remain unschedulable.
					// Let's add the reason groups were rejected to errors. This is useful for debugging.
					t.Errorf("noScaleUpInfo: %#v", st.PodsRemainUnschedulable[0].RejectedNodeGroups)
				}
				assert.Equal(t, tc.scaleUpResult, st.Result)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func setupTest(t *testing.T, nodes []*apiv1.Node, prs []*provreqwrapper.ProvisioningRequest, onScaleUpFunc func(string, int) error, autoprovisioning bool) (*provReqOrchestrator, map[string]*framework.NodeInfo) {
	provider := testprovider.NewTestCloudProvider(onScaleUpFunc, nil)
	if autoprovisioning {
		machineTypes := []string{"large-machine"}
		template := BuildTestNode("large-node-template", 100, 100)
		SetNodeReadyState(template, true, time.Now())
		nodeInfoTemplate := framework.NewNodeInfo(template, nil)
		machineTemplates := map[string]*framework.NodeInfo{
			"large-machine": nodeInfoTemplate,
		}
		onNodeGroupCreateFunc := func(name string) error { return nil }
		provider = testprovider.NewTestAutoprovisioningCloudProvider(onScaleUpFunc, nil, onNodeGroupCreateFunc, nil, machineTypes, machineTemplates)
	}

	provider.AddNodeGroup("test-cpu", 50, 150, 100)
	for _, n := range nodes[:100] {
		provider.AddNode("test-cpu", n)
	}

	podLister := kube_util.NewTestPodLister(nil)
	listers := kube_util.NewListerRegistry(nil, nil, podLister, nil, nil, nil, nil, nil, nil)
	autoscalingContext, err := NewScaleTestAutoscalingContext(config.AutoscalingOptions{}, &fake.Clientset{}, listers, provider, nil, nil)
	assert.NoError(t, err)

	clustersnapshot.InitializeClusterSnapshotOrDie(t, autoscalingContext.ClusterSnapshot, nodes, nil)
	client := provreqclient.NewFakeProvisioningRequestClient(context.Background(), t, prs...)
	processors := NewTestProcessors(&autoscalingContext)
	if autoprovisioning {
		processors.NodeGroupListProcessor = &MockAutoprovisioningNodeGroupListProcessor{T: t}
		processors.NodeGroupManager = &MockAutoprovisioningNodeGroupManager{T: t, ExtraGroups: 2}
	}

	now := time.Now()
	nodeInfos, err := nodeinfosprovider.NewDefaultTemplateNodeInfoProvider(nil, false).Process(&autoscalingContext, nodes, []*appsv1.DaemonSet{}, taints.TaintConfig{}, now)
	assert.NoError(t, err)

	options := config.AutoscalingOptions{
		EstimatorName:                    estimator.BinpackingEstimatorName,
		MaxCoresTotal:                    config.DefaultMaxClusterCores,
		MaxMemoryTotal:                   config.DefaultMaxClusterMemory * units.GiB,
		MinCoresTotal:                    0,
		MinMemoryTotal:                   0,
		NodeAutoprovisioningEnabled:      autoprovisioning,
		MaxAutoprovisionedNodeGroupCount: 10,
	}
	estimatorBuilder, _ := estimator.NewEstimatorBuilder(
		estimator.BinpackingEstimatorName,
		estimator.NewThresholdBasedEstimationLimiter(nil),
		estimator.NewDecreasingPodOrderer(),
		nil,
	)

	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, autoscalingContext.LogRecorder, NewBackoff(), nodegroupconfig.NewDefaultNodeGroupConfigProcessor(options.NodeGroupDefaults), processors.AsyncNodeGroupStateChecker)
	clusterState.UpdateNodes(nodes, nodeInfos, now)

	orchestrator := &provReqOrchestrator{
		client:              client,
		provisioningClasses: []ProvisioningClass{checkcapacity.New(client), besteffortatomic.New(client)},
	}
	orchestrator.Initialize(&autoscalingContext, processors, clusterState, estimatorBuilder, taints.TaintConfig{})
	return orchestrator, nodeInfos
}
