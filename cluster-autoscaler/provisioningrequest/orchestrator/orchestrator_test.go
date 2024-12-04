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
	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/provisioningrequest/autoscaling.x-k8s.io/v1"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	. "k8s.io/autoscaler/cluster-autoscaler/core/test"
	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroupconfig"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodeinfosprovider"
	"k8s.io/autoscaler/cluster-autoscaler/processors/provreq"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	processorstest "k8s.io/autoscaler/cluster-autoscaler/processors/test"
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
	"k8s.io/client-go/kubernetes/fake"
	clocktesting "k8s.io/utils/clock/testing"
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

	anotherCheckCapacityCpuProvReq := provreqwrapper.BuildValidTestProvisioningRequestFromOptions(
		provreqwrapper.TestProvReqOptions{
			Name:     "anotherCheckCapacityCpuProvReq",
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
	impossibleCheckCapacityReq := provreqwrapper.BuildValidTestProvisioningRequestFromOptions(
		provreqwrapper.TestProvReqOptions{
			Name:     "impossibleCheckCapacityRequest",
			CPU:      "1m",
			Memory:   "1",
			PodCount: int32(5001),
			Class:    v1.ProvisioningClassCheckCapacity,
		})

	anotherImpossibleCheckCapacityReq := provreqwrapper.BuildValidTestProvisioningRequestFromOptions(
		provreqwrapper.TestProvReqOptions{
			Name:     "anotherImpossibleCheckCapacityRequest",
			CPU:      "1m",
			Memory:   "1",
			PodCount: int32(5001),
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
		name                string
		provReqs            []*provreqwrapper.ProvisioningRequest
		provReqToScaleUp    *provreqwrapper.ProvisioningRequest
		scaleUpResult       status.ScaleUpResult
		autoprovisioning    bool
		err                 bool
		batchProcessing     bool
		maxBatchSize        int
		batchTimebox        time.Duration
		numProvisionedTrue  int
		numProvisionedFalse int
		numFailedTrue       int
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
			name: "impossible check-capacity, with noRetry parameter",
			provReqs: []*provreqwrapper.ProvisioningRequest{
				impossibleCheckCapacityReq.CopyWithParameters(map[string]v1.Parameter{"noRetry": "true"}),
			},
			provReqToScaleUp: impossibleCheckCapacityReq,
			scaleUpResult:    status.ScaleUpNoOptionsAvailable,
			numFailedTrue:    1,
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
		// Batch processing tests
		{
			name:               "batch processing of check capacity requests with one request",
			provReqs:           []*provreqwrapper.ProvisioningRequest{newCheckCapacityCpuProvReq},
			provReqToScaleUp:   newCheckCapacityCpuProvReq,
			scaleUpResult:      status.ScaleUpSuccessful,
			batchProcessing:    true,
			maxBatchSize:       3,
			batchTimebox:       5 * time.Minute,
			numProvisionedTrue: 1,
		},
		{
			name:               "batch processing of check capacity requests with less requests than max batch size",
			provReqs:           []*provreqwrapper.ProvisioningRequest{newCheckCapacityCpuProvReq, newCheckCapacityMemProvReq},
			provReqToScaleUp:   newCheckCapacityCpuProvReq,
			scaleUpResult:      status.ScaleUpSuccessful,
			batchProcessing:    true,
			maxBatchSize:       3,
			batchTimebox:       5 * time.Minute,
			numProvisionedTrue: 2,
		},
		{
			name:               "batch processing of check capacity requests with requests equal to max batch size",
			provReqs:           []*provreqwrapper.ProvisioningRequest{newCheckCapacityCpuProvReq, newCheckCapacityMemProvReq},
			provReqToScaleUp:   newCheckCapacityCpuProvReq,
			scaleUpResult:      status.ScaleUpSuccessful,
			batchProcessing:    true,
			maxBatchSize:       2,
			batchTimebox:       5 * time.Minute,
			numProvisionedTrue: 2,
		},
		{
			name:               "batch processing of check capacity requests with more requests than max batch size",
			provReqs:           []*provreqwrapper.ProvisioningRequest{newCheckCapacityCpuProvReq, newCheckCapacityMemProvReq, anotherCheckCapacityCpuProvReq},
			provReqToScaleUp:   newCheckCapacityCpuProvReq,
			scaleUpResult:      status.ScaleUpSuccessful,
			batchProcessing:    true,
			maxBatchSize:       2,
			batchTimebox:       5 * time.Minute,
			numProvisionedTrue: 2,
		},
		{
			name:               "batch processing of check capacity requests where cluster contains already provisioned requests",
			provReqs:           []*provreqwrapper.ProvisioningRequest{newCheckCapacityCpuProvReq, bookedCapacityProvReq, anotherCheckCapacityCpuProvReq},
			provReqToScaleUp:   newCheckCapacityCpuProvReq,
			scaleUpResult:      status.ScaleUpSuccessful,
			batchProcessing:    true,
			maxBatchSize:       2,
			batchTimebox:       5 * time.Minute,
			numProvisionedTrue: 3,
		},
		{
			name:               "batch processing of check capacity requests where timebox is exceeded",
			provReqs:           []*provreqwrapper.ProvisioningRequest{newCheckCapacityCpuProvReq, newCheckCapacityMemProvReq},
			provReqToScaleUp:   newCheckCapacityCpuProvReq,
			scaleUpResult:      status.ScaleUpSuccessful,
			batchProcessing:    true,
			maxBatchSize:       5,
			batchTimebox:       0 * time.Nanosecond,
			numProvisionedTrue: 1,
		},
		{
			name:               "batch processing of check capacity requests where max batch size is invalid",
			provReqs:           []*provreqwrapper.ProvisioningRequest{newCheckCapacityCpuProvReq, newCheckCapacityMemProvReq},
			provReqToScaleUp:   newCheckCapacityCpuProvReq,
			scaleUpResult:      status.ScaleUpSuccessful,
			batchProcessing:    true,
			maxBatchSize:       0,
			batchTimebox:       5 * time.Minute,
			numProvisionedTrue: 1,
		},
		{
			name:               "batch processing of check capacity requests where best effort atomic scale-up request is also present in cluster",
			provReqs:           []*provreqwrapper.ProvisioningRequest{newCheckCapacityMemProvReq, newCheckCapacityCpuProvReq, atomicScaleUpProvReq},
			provReqToScaleUp:   newCheckCapacityMemProvReq,
			scaleUpResult:      status.ScaleUpSuccessful,
			batchProcessing:    true,
			maxBatchSize:       2,
			batchTimebox:       5 * time.Minute,
			numProvisionedTrue: 2,
		},
		{
			name:               "process atomic scale-up requests where batch processing of check capacity requests is enabled",
			provReqs:           []*provreqwrapper.ProvisioningRequest{possibleAtomicScaleUpReq},
			provReqToScaleUp:   possibleAtomicScaleUpReq,
			scaleUpResult:      status.ScaleUpSuccessful,
			batchProcessing:    true,
			maxBatchSize:       3,
			batchTimebox:       5 * time.Minute,
			numProvisionedTrue: 1,
		},
		{
			name:               "process atomic scale-up requests where batch processing of check capacity requests is enabled and check capacity requests are present in cluster",
			provReqs:           []*provreqwrapper.ProvisioningRequest{newCheckCapacityMemProvReq, newCheckCapacityCpuProvReq, atomicScaleUpProvReq},
			provReqToScaleUp:   atomicScaleUpProvReq,
			scaleUpResult:      status.ScaleUpNotNeeded,
			batchProcessing:    true,
			maxBatchSize:       3,
			batchTimebox:       5 * time.Minute,
			numProvisionedTrue: 1,
		},
		{
			name:                "batch processing of check capacity requests where some requests' capacity is not available",
			provReqs:            []*provreqwrapper.ProvisioningRequest{newCheckCapacityMemProvReq, impossibleCheckCapacityReq, newCheckCapacityCpuProvReq},
			provReqToScaleUp:    newCheckCapacityMemProvReq,
			scaleUpResult:       status.ScaleUpSuccessful,
			batchProcessing:     true,
			maxBatchSize:        3,
			batchTimebox:        5 * time.Minute,
			numProvisionedTrue:  2,
			numProvisionedFalse: 1,
		},
		{
			name:                "batch processing of check capacity requests where all requests' capacity is not available",
			provReqs:            []*provreqwrapper.ProvisioningRequest{impossibleCheckCapacityReq, anotherImpossibleCheckCapacityReq},
			provReqToScaleUp:    impossibleCheckCapacityReq,
			scaleUpResult:       status.ScaleUpNoOptionsAvailable,
			batchProcessing:     true,
			maxBatchSize:        3,
			batchTimebox:        5 * time.Minute,
			numProvisionedFalse: 2,
		},
	}
	for _, tc := range testCases {
		tc := tc

		nodes := []*apiv1.Node{}
		for _, n := range allNodes {
			nodes = append(nodes, n.DeepCopy())
		}

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

			testProvReqs := []*provreqwrapper.ProvisioningRequest{}
			for _, pr := range tc.provReqs {
				testProvReqs = append(testProvReqs, &provreqwrapper.ProvisioningRequest{ProvisioningRequest: pr.DeepCopy(), PodTemplates: pr.PodTemplates})
			}

			client := provreqclient.NewFakeProvisioningRequestClient(context.Background(), t, testProvReqs...)
			orchestrator, nodeInfos := setupTest(t, client, nodes, onScaleUpFunc, tc.autoprovisioning, tc.batchProcessing, tc.maxBatchSize, tc.batchTimebox)

			st, err := orchestrator.ScaleUp(prPods, []*apiv1.Node{}, []*appsv1.DaemonSet{}, nodeInfos, false)
			if !tc.err {
				assert.NoError(t, err)
				if tc.scaleUpResult != st.Result && len(st.PodsRemainUnschedulable) > 0 {
					// We expected all pods to be scheduled, but some remain unschedulable.
					// Let's add the reason groups were rejected to errors. This is useful for debugging.
					t.Errorf("noScaleUpInfo: %#v", st.PodsRemainUnschedulable[0].RejectedNodeGroups)
				}
				assert.Equal(t, tc.scaleUpResult, st.Result)

				provReqsAfterScaleUp, err := client.ProvisioningRequestsNoCache()
				assert.NoError(t, err)
				assert.Equal(t, len(tc.provReqs), len(provReqsAfterScaleUp))
				assert.Equal(t, tc.numFailedTrue, NumProvisioningRequestsWithCondition(provReqsAfterScaleUp, v1.Failed, metav1.ConditionTrue))

				if tc.batchProcessing {
					// Since batch processing returns aggregated result, we need to check the number of provisioned requests which have the provisioned condition.
					assert.Equal(t, tc.numProvisionedTrue, NumProvisioningRequestsWithCondition(provReqsAfterScaleUp, v1.Provisioned, metav1.ConditionTrue))
					assert.Equal(t, tc.numProvisionedFalse, NumProvisioningRequestsWithCondition(provReqsAfterScaleUp, v1.Provisioned, metav1.ConditionFalse))
				}
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func setupTest(t *testing.T, client *provreqclient.ProvisioningRequestClient, nodes []*apiv1.Node, onScaleUpFunc func(string, int) error, autoprovisioning bool, batchProcessing bool, maxBatchSize int, batchTimebox time.Duration) (*provReqOrchestrator, map[string]*framework.NodeInfo) {
	provider := testprovider.NewTestCloudProvider(onScaleUpFunc, nil)
	clock := clocktesting.NewFakePassiveClock(time.Now())
	now := clock.Now()
	if autoprovisioning {
		machineTypes := []string{"large-machine"}
		template := BuildTestNode("large-node-template", 100, 100)
		SetNodeReadyState(template, true, now)
		nodeInfoTemplate := framework.NewTestNodeInfo(template)
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

	options := config.AutoscalingOptions{}
	if batchProcessing {
		options.CheckCapacityBatchProcessing = true
		options.CheckCapacityProvisioningRequestMaxBatchSize = maxBatchSize
		options.CheckCapacityProvisioningRequestBatchTimebox = batchTimebox
	}

	autoscalingContext, err := NewScaleTestAutoscalingContext(options, &fake.Clientset{}, listers, provider, nil, nil)
	assert.NoError(t, err)

	clustersnapshot.InitializeClusterSnapshotOrDie(t, autoscalingContext.ClusterSnapshot, nodes, nil)
	processors := processorstest.NewTestProcessors(&autoscalingContext)
	if autoprovisioning {
		processors.NodeGroupListProcessor = &MockAutoprovisioningNodeGroupListProcessor{T: t}
		processors.NodeGroupManager = &MockAutoprovisioningNodeGroupManager{T: t, ExtraGroups: 2}
	}
	nodeInfos, err := nodeinfosprovider.NewDefaultTemplateNodeInfoProvider(nil, false).Process(&autoscalingContext, nodes, []*appsv1.DaemonSet{}, taints.TaintConfig{}, now)
	assert.NoError(t, err)

	estimatorBuilder, _ := estimator.NewEstimatorBuilder(
		estimator.BinpackingEstimatorName,
		estimator.NewThresholdBasedEstimationLimiter(nil),
		estimator.NewDecreasingPodOrderer(),
		nil,
	)

	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, autoscalingContext.LogRecorder, NewBackoff(), nodegroupconfig.NewDefaultNodeGroupConfigProcessor(autoscalingContext.NodeGroupDefaults), processors.AsyncNodeGroupStateChecker)
	clusterState.UpdateNodes(nodes, nodeInfos, now)

	var injector *provreq.ProvisioningRequestPodsInjector
	if batchProcessing {
		injector = provreq.NewFakePodsInjector(client, clocktesting.NewFakePassiveClock(now))
	}

	orchestrator := &provReqOrchestrator{
		client:              client,
		provisioningClasses: []ProvisioningClass{checkcapacity.New(client, injector), besteffortatomic.New(client)},
	}

	orchestrator.Initialize(&autoscalingContext, processors, clusterState, estimatorBuilder, taints.TaintConfig{})
	return orchestrator, nodeInfos
}

func NumProvisioningRequestsWithCondition(prList []*provreqwrapper.ProvisioningRequest, conditionType string, conditionStatus metav1.ConditionStatus) int {
	count := 0

	for _, pr := range prList {
		for _, c := range pr.Status.Conditions {
			if c.Type == conditionType && c.Status == conditionStatus {
				count++
				break
			}
		}
	}

	return count
}
