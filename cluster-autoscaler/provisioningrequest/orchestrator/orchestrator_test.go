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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/provisioningrequest/autoscaling.x-k8s.io/v1"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	. "k8s.io/autoscaler/cluster-autoscaler/core/test"
	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroupconfig"
	"k8s.io/autoscaler/cluster-autoscaler/processors/provreq"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	processorstest "k8s.io/autoscaler/cluster-autoscaler/processors/test"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/besteffortatomic"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/checkcapacity"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/pods"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqclient"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqwrapper"
	"k8s.io/autoscaler/cluster-autoscaler/resourcequotas"
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
			Name:     "new-check-capacity-cpu-provreq",
			CPU:      "5m",
			Memory:   "5",
			PodCount: int32(100),
			Class:    v1.ProvisioningClassCheckCapacity,
		})

	anotherCheckCapacityCpuProvReq := provreqwrapper.BuildValidTestProvisioningRequestFromOptions(
		provreqwrapper.TestProvReqOptions{
			Name:     "another-check-capacity-cpu-provreq",
			CPU:      "5m",
			Memory:   "5",
			PodCount: int32(100),
			Class:    v1.ProvisioningClassCheckCapacity,
		})

	newCheckCapacityMemProvReq := provreqwrapper.BuildValidTestProvisioningRequestFromOptions(
		provreqwrapper.TestProvReqOptions{
			Name:     "new-check-capacity-mem-provreq",
			CPU:      "1m",
			Memory:   "100",
			PodCount: int32(100),
			Class:    v1.ProvisioningClassCheckCapacity,
		})
	impossibleCheckCapacityReq := provreqwrapper.BuildValidTestProvisioningRequestFromOptions(
		provreqwrapper.TestProvReqOptions{
			Name:     "impossible-check-capacity-request",
			CPU:      "1m",
			Memory:   "1",
			PodCount: int32(5001),
			Class:    v1.ProvisioningClassCheckCapacity,
		})

	anotherImpossibleCheckCapacityReq := provreqwrapper.BuildValidTestProvisioningRequestFromOptions(
		provreqwrapper.TestProvReqOptions{
			Name:     "another-impossible-check-capacity-request",
			CPU:      "1m",
			Memory:   "1",
			PodCount: int32(5001),
			Class:    v1.ProvisioningClassCheckCapacity,
		})

	// Partial capacity check requests.
	// Total cluster capacity: 100 CPU nodes (10 bytes each) + 100 memory nodes (1000 bytes each) = 101,000 bytes
	// Request 300 pods * 400 bytes = 120,000 bytes, which exceeds capacity
	// Expected: ~252 pods can fit (101,000 / 400)
	partialCapacityProvReq := provreqwrapper.BuildValidTestProvisioningRequestFromOptions(
		provreqwrapper.TestProvReqOptions{
			Name:     "partial-capacity-provreq",
			CPU:      "1m",
			Memory:   "400",
			PodCount: int32(300),
			Class:    v1.ProvisioningClassCheckCapacity,
		})

	// Another partial capacity request for batch testing
	// Request 280 pods * 400 bytes = 112,000 bytes
	// Expected: ~252 pods can fit (101,000 / 400)
	anotherPartialCapacityProvReq := provreqwrapper.BuildValidTestProvisioningRequestFromOptions(
		provreqwrapper.TestProvReqOptions{
			Name:     "another-partial-capacity-provreq",
			CPU:      "1m",
			Memory:   "400",
			PodCount: int32(280),
			Class:    v1.ProvisioningClassCheckCapacity,
		})

	// Check capacity request that fits completely, even with partialCapacityCheck enabled
	completeCapacityProvReq := provreqwrapper.BuildValidTestProvisioningRequestFromOptions(
		provreqwrapper.TestProvReqOptions{
			Name:     "complete-capacity-provreq",
			CPU:      "5m",
			Memory:   "5",
			PodCount: int32(50),
			Class:    v1.ProvisioningClassCheckCapacity,
		})

	// Check capacity request that has 0 pods that can fit
	impossibleResourceRequest := provreqwrapper.BuildValidTestProvisioningRequestFromOptions(
		provreqwrapper.TestProvReqOptions{
			Name:     "impossible-resource-request",
			CPU:      "101m", // More than any single node has (100m max)
			Memory:   "1001", // More than any single node has (1000 max)
			PodCount: int32(5),
			Class:    v1.ProvisioningClassCheckCapacity,
		})

	// Multi-PodSet ProvReq: first PodSet (small pods) fits, second PodSet (impossible pods) doesn't.
	// Used to test checkOnly vs bookPartial divergence, which only occurs with multiple PodSets.
	multiPodSetProvReq := provreqwrapper.NewProvisioningRequest(
		&v1.ProvisioningRequest{
			ObjectMeta: metav1.ObjectMeta{Name: "multi-podset-provreq", Namespace: "default",
				CreationTimestamp: metav1.NewTime(time.Now())},
			Spec: v1.ProvisioningRequestSpec{
				ProvisioningClassName: v1.ProvisioningClassCheckCapacity,
				PodSets: []v1.PodSet{
					{PodTemplateRef: v1.Reference{Name: "small-template"}, Count: 10},
					{PodTemplateRef: v1.Reference{Name: "large-template"}, Count: 5},
				},
			},
			Status: v1.ProvisioningRequestStatus{Conditions: []metav1.Condition{}},
		},
		[]*apiv1.PodTemplate{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "small-template", Namespace: "default"},
				Template: apiv1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": "test-app"}},
					Spec: apiv1.PodSpec{Containers: []apiv1.Container{
						{Name: "c", Image: "img", Resources: apiv1.ResourceRequirements{
							Requests: apiv1.ResourceList{apiv1.ResourceCPU: resource.MustParse("5m"), apiv1.ResourceMemory: resource.MustParse("5")},
							Limits:   apiv1.ResourceList{apiv1.ResourceCPU: resource.MustParse("5m"), apiv1.ResourceMemory: resource.MustParse("5")},
						}},
					}},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{Name: "large-template", Namespace: "default"},
				Template: apiv1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": "test-app"}},
					Spec: apiv1.PodSpec{Containers: []apiv1.Container{
						{Name: "c", Image: "img", Resources: apiv1.ResourceRequirements{
							Requests: apiv1.ResourceList{apiv1.ResourceCPU: resource.MustParse("101m"), apiv1.ResourceMemory: resource.MustParse("1001")},
							Limits:   apiv1.ResourceList{apiv1.ResourceCPU: resource.MustParse("101m"), apiv1.ResourceMemory: resource.MustParse("1001")},
						}},
					}},
				},
			},
		},
	)

	// Active atomic scale up requests.
	atomicScaleUpProvReq := provreqwrapper.BuildValidTestProvisioningRequestFromOptions(
		provreqwrapper.TestProvReqOptions{
			Name:     "atomic-scale-up-provreq",
			CPU:      "5m",
			Memory:   "5",
			PodCount: int32(5),
			Class:    v1.ProvisioningClassBestEffortAtomicScaleUp,
		})
	largeAtomicScaleUpProvReq := provreqwrapper.BuildValidTestProvisioningRequestFromOptions(
		provreqwrapper.TestProvReqOptions{
			Name:     "large-atomic-scale-up-provreq",
			CPU:      "1m",
			Memory:   "100",
			PodCount: int32(100),
			Class:    v1.ProvisioningClassBestEffortAtomicScaleUp,
		})
	impossibleAtomicScaleUpReq := provreqwrapper.BuildValidTestProvisioningRequestFromOptions(
		provreqwrapper.TestProvReqOptions{
			Name:     "impossible-atomic-scale-up-request",
			CPU:      "1m",
			Memory:   "1",
			PodCount: int32(5001),
			Class:    v1.ProvisioningClassBestEffortAtomicScaleUp,
		})
	possibleAtomicScaleUpReq := provreqwrapper.BuildValidTestProvisioningRequestFromOptions(
		provreqwrapper.TestProvReqOptions{
			Name:     "possible-atomic-scale-up-req",
			CPU:      "100m",
			Memory:   "1",
			PodCount: int32(120),
			Class:    v1.ProvisioningClassBestEffortAtomicScaleUp,
		})
	autoprovisioningAtomicScaleUpReq := provreqwrapper.BuildValidTestProvisioningRequestFromOptions(
		provreqwrapper.TestProvReqOptions{
			Name:     "autoprovisioning-atomic-scale-up-req",
			CPU:      "100m",
			Memory:   "100",
			PodCount: int32(5),
			Class:    v1.ProvisioningClassBestEffortAtomicScaleUp,
		})

	// Already provisioned provisioning request - capacity should be booked before processing a new request.
	// Books 20 out of 100 high-memory nodes.
	bookedCapacityProvReq := provreqwrapper.BuildValidTestProvisioningRequestFromOptions(
		provreqwrapper.TestProvReqOptions{
			Name:     "booked-capacity-provreq",
			CPU:      "1m",
			Memory:   "200",
			PodCount: int32(100),
			Class:    v1.ProvisioningClassCheckCapacity,
		})
	bookedCapacityProvReq.SetConditions([]metav1.Condition{{Type: v1.Provisioned, Status: metav1.ConditionTrue, LastTransitionTime: metav1.Now()}})

	// Expired provisioning request - should be ignored.
	expiredProvReq := provreqwrapper.BuildValidTestProvisioningRequestFromOptions(
		provreqwrapper.TestProvReqOptions{
			Name:     "expired-provreq",
			CPU:      "1m",
			Memory:   "200",
			PodCount: int32(100),
			Class:    v1.ProvisioningClassCheckCapacity,
		})
	expiredProvReq.SetConditions([]metav1.Condition{{Type: v1.BookingExpired, Status: metav1.ConditionTrue, LastTransitionTime: metav1.Now()}})

	// Unsupported provisioning request - should be ignored.
	unsupportedProvReq := provreqwrapper.BuildValidTestProvisioningRequestFromOptions(
		provreqwrapper.TestProvReqOptions{
			Name:     "unsupported-provreq",
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
			scaleUpResult:      status.ScaleUpSuccessful,
			batchProcessing:    true,
			maxBatchSize:       3,
			batchTimebox:       5 * time.Minute,
			numProvisionedTrue: 2,
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
		// Partial capacity check tests (non-batch)
		{
			name: "partial capacity check enabled, some pods fit",
			provReqs: []*provreqwrapper.ProvisioningRequest{
				partialCapacityProvReq.CopyWithParameters(map[string]v1.Parameter{checkcapacity.PartialCapacityCheckKey: checkcapacity.PartialCapacityCheckBookPartial}),
			},
			provReqToScaleUp: partialCapacityProvReq,
			// Single PodSet where not all pods fit → podset not schedulable → NoOptionsAvailable
			scaleUpResult: status.ScaleUpNoOptionsAvailable,
		},
		{
			name: "partial capacity check enabled, all pods fit",
			provReqs: []*provreqwrapper.ProvisioningRequest{
				completeCapacityProvReq.CopyWithParameters(map[string]v1.Parameter{checkcapacity.PartialCapacityCheckKey: checkcapacity.PartialCapacityCheckBookPartial}),
			},
			provReqToScaleUp: completeCapacityProvReq,
			scaleUpResult:    status.ScaleUpSuccessful,
		},
		{
			name: "partial capacity check disabled, partial capacity available but returns not available",
			provReqs: []*provreqwrapper.ProvisioningRequest{
				partialCapacityProvReq.CopyWithParameters(map[string]v1.Parameter{}),
			},
			provReqToScaleUp: partialCapacityProvReq,
			scaleUpResult:    status.ScaleUpNoOptionsAvailable,
		},
		{
			name: "partial capacity check with noRetry parameter, some pods fit",
			provReqs: []*provreqwrapper.ProvisioningRequest{
				partialCapacityProvReq.CopyWithParameters(map[string]v1.Parameter{checkcapacity.PartialCapacityCheckKey: checkcapacity.PartialCapacityCheckBookPartial, "noRetry": "true"}),
			},
			provReqToScaleUp: partialCapacityProvReq,
			// Single PodSet where not all pods fit → noRetry triggers Failed=true
			scaleUpResult: status.ScaleUpNoOptionsAvailable,
			numFailedTrue: 1,
		},
		{
			name: "partial capacity check enabled, zero pods fit due to resource constraints",
			provReqs: []*provreqwrapper.ProvisioningRequest{
				impossibleResourceRequest.CopyWithParameters(map[string]v1.Parameter{checkcapacity.PartialCapacityCheckKey: checkcapacity.PartialCapacityCheckBookPartial}),
			},
			provReqToScaleUp: impossibleResourceRequest,
			scaleUpResult:    status.ScaleUpNoOptionsAvailable,
		},
		{
			name: "partial capacity check param value invalid, use default behavior despite partial capacity existing",
			provReqs: []*provreqwrapper.ProvisioningRequest{
				partialCapacityProvReq.CopyWithParameters(map[string]v1.Parameter{checkcapacity.PartialCapacityCheckKey: "invalid-value"}),
			},
			provReqToScaleUp: partialCapacityProvReq,
			// Returns NoOptionsAvailable since partial capacity check is not enabled
			scaleUpResult: status.ScaleUpNoOptionsAvailable,
		},
		// Batch processing with partial capacity check
		{
			name: "batch processing with partial capacity check, first request gets partial, second gets none",
			provReqs: []*provreqwrapper.ProvisioningRequest{
				// First request: 300 pods, single PodSet, not all fit → not schedulable
				partialCapacityProvReq.CopyWithParameters(map[string]v1.Parameter{checkcapacity.PartialCapacityCheckKey: checkcapacity.PartialCapacityCheckBookPartial}),
				// Second request: 280 pods, single PodSet, not all fit → not schedulable
				anotherPartialCapacityProvReq.CopyWithParameters(map[string]v1.Parameter{checkcapacity.PartialCapacityCheckKey: checkcapacity.PartialCapacityCheckBookPartial}),
			},
			provReqToScaleUp:    partialCapacityProvReq,
			scaleUpResult:       status.ScaleUpNoOptionsAvailable,
			batchProcessing:     true,
			maxBatchSize:        3,
			batchTimebox:        5 * time.Minute,
			numProvisionedTrue:  0, // Neither single-PodSet request fully fits
			numProvisionedFalse: 2, // Both get Provisioned=False
		},
		{
			name: "batch processing with mixed partial capacity check settings",
			provReqs: []*provreqwrapper.ProvisioningRequest{
				// Single PodSet, not all fit → NoOptionsAvailable, Provisioned=False
				partialCapacityProvReq.CopyWithParameters(map[string]v1.Parameter{checkcapacity.PartialCapacityCheckKey: checkcapacity.PartialCapacityCheckBookPartial}),
				// No partial check, not all fit → NoOptionsAvailable, Provisioned=False
				anotherPartialCapacityProvReq.CopyWithParameters(map[string]v1.Parameter{}),
				// Single PodSet, all 50 pods fit → Successful, Provisioned=True
				completeCapacityProvReq.CopyWithParameters(map[string]v1.Parameter{checkcapacity.PartialCapacityCheckKey: checkcapacity.PartialCapacityCheckBookPartial}),
			},
			provReqToScaleUp:    partialCapacityProvReq,
			scaleUpResult:       status.ScaleUpSuccessful,
			batchProcessing:     true,
			maxBatchSize:        3,
			batchTimebox:        5 * time.Minute,
			numProvisionedTrue:  1,
			numProvisionedFalse: 2,
		},
		// Multi-PodSet checkOnly vs bookPartial tests
		{
			name: "checkOnly mode, multi-PodSet, partial capacity reports Provisioned=false",
			provReqs: []*provreqwrapper.ProvisioningRequest{
				multiPodSetProvReq.CopyWithParameters(map[string]v1.Parameter{
					checkcapacity.PartialCapacityCheckKey: checkcapacity.PartialCapacityCheckCheckOnly,
				}),
			},
			provReqToScaleUp:    multiPodSetProvReq,
			scaleUpResult:       status.ScaleUpSuccessful,
			numProvisionedFalse: 1, // checkOnly: partial capacity found but not booked
		},
		{
			name: "bookPartial mode, multi-PodSet, partial capacity reports Provisioned=true",
			provReqs: []*provreqwrapper.ProvisioningRequest{
				multiPodSetProvReq.CopyWithParameters(map[string]v1.Parameter{
					checkcapacity.PartialCapacityCheckKey: checkcapacity.PartialCapacityCheckBookPartial,
				}),
			},
			provReqToScaleUp:   multiPodSetProvReq,
			scaleUpResult:      status.ScaleUpSuccessful,
			numProvisionedTrue: 1, // bookPartial: partial capacity found and booked
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

				if tc.batchProcessing || tc.numProvisionedTrue > 0 || tc.numProvisionedFalse > 0 {
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
	provider := testprovider.NewTestCloudProviderBuilder().WithOnScaleUp(onScaleUpFunc).Build()
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
		provider = testprovider.NewTestCloudProviderBuilder().WithOnScaleUp(onScaleUpFunc).WithOnNodeGroupCreate(onNodeGroupCreateFunc).WithMachineTypes(machineTypes).WithMachineTemplates(machineTemplates).Build()
	}

	provider.AddNodeGroup("test-cpu", 50, 150, 100)
	for _, n := range nodes[:100] {
		provider.AddNode("test-cpu", n)
	}

	podLister := kube_util.NewTestPodLister(nil)
	listers := kube_util.NewListerRegistry(nil, nil, podLister, nil, nil, nil, nil, nil, nil)

	options := config.AutoscalingOptions{
		MaxNodeGroupBinpackingDuration: 1 * time.Second,
	}
	if batchProcessing {
		options.CheckCapacityBatchProcessing = true
		options.CheckCapacityProvisioningRequestMaxBatchSize = maxBatchSize
		options.CheckCapacityProvisioningRequestBatchTimebox = batchTimebox
	}

	processors, templateNodeInfoRegistry := processorstest.NewTestProcessors(options)
	autoscalingCtx, err := NewScaleTestAutoscalingContext(options, &fake.Clientset{}, listers, provider, nil, nil, templateNodeInfoRegistry)
	assert.NoError(t, err)

	clustersnapshot.InitializeClusterSnapshotOrDie(t, autoscalingCtx.ClusterSnapshot, nodes, nil)
	if autoprovisioning {
		processors.NodeGroupListProcessor = &MockAutoprovisioningNodeGroupListProcessor{T: t}
		processors.NodeGroupManager = &MockAutoprovisioningNodeGroupManager{T: t, ExtraGroups: 2}
	}
	err = autoscalingCtx.TemplateNodeInfoRegistry.Recompute(&autoscalingCtx, nodes, []*appsv1.DaemonSet{}, taints.TaintConfig{}, now)
	assert.NoError(t, err)
	nodeInfos := autoscalingCtx.TemplateNodeInfoRegistry.GetNodeInfos()

	estimatorBuilder, _ := estimator.NewEstimatorBuilder(
		estimator.BinpackingEstimatorName,
		estimator.NewThresholdBasedEstimationLimiter(nil),
		estimator.NewDecreasingPodOrderer(),
		nil,
		false,
	)

	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, autoscalingCtx.LogRecorder, NewBackoff(), nodegroupconfig.NewDefaultNodeGroupConfigProcessor(autoscalingCtx.NodeGroupDefaults), processors.AsyncNodeGroupStateChecker)
	clusterState.UpdateNodes(nodes, nodeInfos, now)

	var injector *provreq.ProvisioningRequestPodsInjector
	if batchProcessing {
		injector = provreq.NewFakePodsInjector(client, clocktesting.NewFakePassiveClock(now))
	}

	quotasTrackerFactory := resourcequotas.NewTrackerFactory(resourcequotas.TrackerOptions{
		QuotaProvider:            resourcequotas.NewFakeProvider(nil),
		CustomResourcesProcessor: processors.CustomResourcesProcessor,
	})
	orchestrator := &provReqOrchestrator{
		client:              client,
		provisioningClasses: []ProvisioningClass{checkcapacity.New(client, injector), besteffortatomic.New(client)},
	}
	orchestrator.Initialize(&autoscalingCtx, processors, clusterState, estimatorBuilder, taints.TaintConfig{}, quotasTrackerFactory)
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
