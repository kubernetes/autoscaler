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

package utilization

import (
	"testing"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/kubernetes/pkg/kubelet/types"

	"github.com/stretchr/testify/assert"
)

// TODO(DRA): Add DRA-specific test cases.

func TestCalculate(t *testing.T) {
	testTime := time.Date(2020, time.December, 18, 17, 0, 0, 0, time.UTC)
	pod := BuildTestPod("p1", 100, 200000)
	pod2 := BuildTestPod("p2", -1, -1)

	podWithInitContainers := BuildTestPod("p-init", 100, 200000)
	restartAlways := apiv1.ContainerRestartPolicyAlways
	podWithInitContainers.Spec.InitContainers = []apiv1.Container{
		// restart always
		{
			Resources: apiv1.ResourceRequirements{
				Requests: apiv1.ResourceList{
					apiv1.ResourceCPU:    *resource.NewMilliQuantity(50, resource.DecimalSI),
					apiv1.ResourceMemory: *resource.NewQuantity(100000, resource.DecimalSI),
				},
			},
			RestartPolicy: &restartAlways,
		},
		// non-restartable, should be excluded from calculations
		{
			Resources: apiv1.ResourceRequirements{
				Requests: apiv1.ResourceList{
					apiv1.ResourceCPU:    *resource.NewMilliQuantity(5, resource.DecimalSI),
					apiv1.ResourceMemory: *resource.NewQuantity(100, resource.DecimalSI),
				},
			},
		},
	}

	podWithLargeNonRestartableInitContainers := BuildTestPod("p-large-init", 100, 200000)
	podWithLargeNonRestartableInitContainers.Spec.InitContainers = []apiv1.Container{
		// large non-restartable should "overwhelm" the pod utilization calc
		// see formula: https://github.com/kubernetes/enhancements/tree/master/keps/sig-node/753-sidecar-containers#resources-calculation-for-scheduling-and-pod-admission
		{
			Resources: apiv1.ResourceRequirements{
				Requests: apiv1.ResourceList{
					apiv1.ResourceCPU:    *resource.NewMilliQuantity(50000, resource.DecimalSI),
					apiv1.ResourceMemory: *resource.NewQuantity(100000000, resource.DecimalSI),
				},
			},
		},
	}
	node := BuildTestNode("node1", 2000, 2000000)
	SetNodeReadyState(node, true, time.Time{})
	nodeInfo := framework.NewTestNodeInfo(node, pod, pod, pod2)

	gpuConfig := GetGpuConfigFromNode(nodeInfo.Node())
	utilInfo, err := Calculate(nodeInfo, false, false, false, gpuConfig, testTime)
	assert.NoError(t, err)
	assert.InEpsilon(t, 2.0/10, utilInfo.Utilization, 0.01)
	assert.Equal(t, 0.1, utilInfo.CpuUtil)

	node2 := BuildTestNode("node2", 2000, -1)
	nodeInfo = framework.NewTestNodeInfo(node2, pod, pod, pod2)

	gpuConfig = GetGpuConfigFromNode(nodeInfo.Node())
	_, err = Calculate(nodeInfo, false, false, false, gpuConfig, testTime)
	assert.Error(t, err)

	node3 := BuildTestNode("node3", 2000, 2000000)
	SetNodeReadyState(node3, true, time.Time{})
	nodeInfo = framework.NewTestNodeInfo(node3, pod, podWithInitContainers, podWithLargeNonRestartableInitContainers)

	gpuConfig = GetGpuConfigFromNode(nodeInfo.Node())
	utilInfo, err = Calculate(nodeInfo, false, false, false, gpuConfig, testTime)
	assert.NoError(t, err)
	assert.InEpsilon(t, 50.25, utilInfo.Utilization, 0.01)
	assert.Equal(t, 25.125, utilInfo.CpuUtil)

	daemonSetPod3 := BuildTestPod("p3", 100, 200000)
	daemonSetPod3.OwnerReferences = GenerateOwnerReferences("ds", "DaemonSet", "apps/v1", "")

	daemonSetPod4 := BuildTestPod("p4", 100, 200000)
	daemonSetPod4.OwnerReferences = GenerateOwnerReferences("ds", "CustomDaemonSet", "crd/v1", "")
	daemonSetPod4.Annotations = map[string]string{"cluster-autoscaler.kubernetes.io/daemonset-pod": "true"}

	nodeInfo = framework.NewTestNodeInfo(node, pod, pod, pod2, daemonSetPod3, daemonSetPod4)
	gpuConfig = GetGpuConfigFromNode(nodeInfo.Node())
	utilInfo, err = Calculate(nodeInfo, true, false, false, gpuConfig, testTime)
	assert.NoError(t, err)
	assert.InEpsilon(t, 2.5/10, utilInfo.Utilization, 0.01)

	nodeInfo = framework.NewTestNodeInfo(node, pod, pod2, daemonSetPod3)
	gpuConfig = GetGpuConfigFromNode(nodeInfo.Node())
	utilInfo, err = Calculate(nodeInfo, false, false, false, gpuConfig, testTime)
	assert.NoError(t, err)
	assert.InEpsilon(t, 2.0/10, utilInfo.Utilization, 0.01)

	terminatedPod := BuildTestPod("podTerminated", 100, 200000)
	terminatedPod.DeletionTimestamp = &metav1.Time{Time: testTime.Add(-10 * time.Minute)}
	nodeInfo = framework.NewTestNodeInfo(node, pod, pod, pod2, terminatedPod)
	gpuConfig = GetGpuConfigFromNode(nodeInfo.Node())
	utilInfo, err = Calculate(nodeInfo, false, false, false, gpuConfig, testTime)
	assert.NoError(t, err)
	assert.InEpsilon(t, 2.0/10, utilInfo.Utilization, 0.01)

	mirrorPod := BuildTestPod("p4", 100, 200000)
	mirrorPod.Annotations = map[string]string{
		types.ConfigMirrorAnnotationKey: "",
	}

	nodeInfo = framework.NewTestNodeInfo(node, pod, pod, pod2, mirrorPod)
	gpuConfig = GetGpuConfigFromNode(nodeInfo.Node())
	utilInfo, err = Calculate(nodeInfo, false, true, false, gpuConfig, testTime)
	assert.NoError(t, err)
	assert.InEpsilon(t, 2.0/9.0, utilInfo.Utilization, 0.01)

	nodeInfo = framework.NewTestNodeInfo(node, pod, pod2, mirrorPod)
	gpuConfig = GetGpuConfigFromNode(nodeInfo.Node())
	utilInfo, err = Calculate(nodeInfo, false, false, false, gpuConfig, testTime)
	assert.NoError(t, err)
	assert.InEpsilon(t, 2.0/10, utilInfo.Utilization, 0.01)

	nodeInfo = framework.NewTestNodeInfo(node, pod, mirrorPod, daemonSetPod3)
	gpuConfig = GetGpuConfigFromNode(nodeInfo.Node())
	utilInfo, err = Calculate(nodeInfo, true, true, false, gpuConfig, testTime)
	assert.NoError(t, err)
	assert.InEpsilon(t, 1.0/8.0, utilInfo.Utilization, 0.01)

	gpuNode := BuildTestNode("gpu_node", 2000, 2000000)
	AddGpusToNode(gpuNode, 1)
	gpuPod := BuildTestPod("gpu_pod", 100, 200000)
	RequestGpuForPod(gpuPod, 1)
	TolerateGpuForPod(gpuPod)
	nodeInfo = framework.NewTestNodeInfo(gpuNode, pod, pod, gpuPod)
	gpuConfig = GetGpuConfigFromNode(nodeInfo.Node())
	utilInfo, err = Calculate(nodeInfo, false, false, false, gpuConfig, testTime)
	assert.NoError(t, err)
	assert.InEpsilon(t, 1/1, utilInfo.Utilization, 0.01)

	// Node with Unready GPU
	gpuNode = BuildTestNode("gpu_node", 2000, 2000000)
	AddGpuLabelToNode(gpuNode)
	nodeInfo = framework.NewTestNodeInfo(gpuNode, pod, pod)
	gpuConfig = GetGpuConfigFromNode(nodeInfo.Node())
	utilInfo, err = Calculate(nodeInfo, false, false, false, gpuConfig, testTime)
	assert.NoError(t, err)
	assert.Zero(t, utilInfo.Utilization)
}
