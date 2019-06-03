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

package test

import (
	"fmt"
	"time"

	"net/http"
	"net/http/httptest"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	refv1 "k8s.io/client-go/tools/reference"
	"k8s.io/kubernetes/pkg/api/testapi"

	"github.com/stretchr/testify/mock"
)

// BuildTestPod creates a pod with specified resources.
func BuildTestPod(name string, cpu int64, mem int64) *apiv1.Pod {
	pod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			UID:         types.UID(name),
			Namespace:   "default",
			Name:        name,
			SelfLink:    fmt.Sprintf("/api/v1/namespaces/default/pods/%s", name),
			Annotations: map[string]string{},
		},
		Spec: apiv1.PodSpec{
			Containers: []apiv1.Container{
				{
					Resources: apiv1.ResourceRequirements{
						Requests: apiv1.ResourceList{},
					},
				},
			},
		},
	}

	if cpu >= 0 {
		pod.Spec.Containers[0].Resources.Requests[apiv1.ResourceCPU] = *resource.NewMilliQuantity(cpu, resource.DecimalSI)
	}
	if mem >= 0 {
		pod.Spec.Containers[0].Resources.Requests[apiv1.ResourceMemory] = *resource.NewQuantity(mem, resource.DecimalSI)
	}

	return pod
}

const (
	// cannot use constants from gpu module due to cyclic package import
	resourceNvidiaGPU = "nvidia.com/gpu"
	gpuLabel          = "cloud.google.com/gke-accelerator"
	defaultGPUType    = "nvidia-tesla-k80"
)

// RequestGpuForPod modifies pod's resource requests by adding a number of GPUs to them.
func RequestGpuForPod(pod *apiv1.Pod, gpusCount int64) {
	if pod.Spec.Containers[0].Resources.Limits == nil {
		pod.Spec.Containers[0].Resources.Limits = apiv1.ResourceList{}
	}
	pod.Spec.Containers[0].Resources.Limits[resourceNvidiaGPU] = *resource.NewQuantity(gpusCount, resource.DecimalSI)

	if pod.Spec.Containers[0].Resources.Requests == nil {
		pod.Spec.Containers[0].Resources.Requests = apiv1.ResourceList{}
	}
	pod.Spec.Containers[0].Resources.Requests[resourceNvidiaGPU] = *resource.NewQuantity(gpusCount, resource.DecimalSI)
}

// BuildTestNode creates a node with specified capacity.
func BuildTestNode(name string, millicpu int64, mem int64) *apiv1.Node {
	node := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:     name,
			SelfLink: fmt.Sprintf("/api/v1/nodes/%s", name),
			Labels:   map[string]string{},
		},
		Spec: apiv1.NodeSpec{
			ProviderID: name,
		},
		Status: apiv1.NodeStatus{
			Capacity: apiv1.ResourceList{
				apiv1.ResourcePods: *resource.NewQuantity(100, resource.DecimalSI),
			},
		},
	}

	if millicpu >= 0 {
		node.Status.Capacity[apiv1.ResourceCPU] = *resource.NewMilliQuantity(millicpu, resource.DecimalSI)
	}
	if mem >= 0 {
		node.Status.Capacity[apiv1.ResourceMemory] = *resource.NewQuantity(mem, resource.DecimalSI)
	}

	node.Status.Allocatable = apiv1.ResourceList{}
	for k, v := range node.Status.Capacity {
		node.Status.Allocatable[k] = v
	}

	return node
}

// AddGpusToNode adds GPU capacity to given node. Default accelerator type is used.
func AddGpusToNode(node *apiv1.Node, gpusCount int64) {
	node.Spec.Taints = append(
		node.Spec.Taints,
		apiv1.Taint{
			Key:    resourceNvidiaGPU,
			Value:  "present",
			Effect: "NoSchedule",
		})
	node.Status.Capacity[resourceNvidiaGPU] = *resource.NewQuantity(gpusCount, resource.DecimalSI)
	node.Status.Allocatable[resourceNvidiaGPU] = *resource.NewQuantity(gpusCount, resource.DecimalSI)
	AddGpuLabelToNode(node)
}

// AddGpuLabelToNode adds GPULabel to give node. This is used to mock intermediate result that GPU on node is not ready
func AddGpuLabelToNode(node *apiv1.Node) {
	node.Labels[gpuLabel] = defaultGPUType
}

// GetGPULabel return GPULabel on the node. This is only used in unit tests.
func GetGPULabel() string {
	return gpuLabel
}

// SetNodeReadyState sets node ready state to either ConditionTrue or ConditionFalse.
func SetNodeReadyState(node *apiv1.Node, ready bool, lastTransition time.Time) {
	if ready {
		SetNodeCondition(node, apiv1.NodeReady, apiv1.ConditionTrue, lastTransition)
	} else {
		SetNodeCondition(node, apiv1.NodeReady, apiv1.ConditionFalse, lastTransition)
	}
}

// SetNodeCondition sets node condition.
func SetNodeCondition(node *apiv1.Node, conditionType apiv1.NodeConditionType, status apiv1.ConditionStatus, lastTransition time.Time) {
	for i := range node.Status.Conditions {
		if node.Status.Conditions[i].Type == conditionType {
			node.Status.Conditions[i].LastTransitionTime = metav1.Time{Time: lastTransition}
			node.Status.Conditions[i].Status = status
			return
		}
	}
	// Condition doesn't exist yet.
	condition := apiv1.NodeCondition{
		Type:               conditionType,
		Status:             status,
		LastTransitionTime: metav1.Time{Time: lastTransition},
	}
	node.Status.Conditions = append(node.Status.Conditions, condition)
}

// RefJSON builds string reference to
func RefJSON(o runtime.Object) string {
	ref, err := refv1.GetReference(scheme.Scheme, o)
	if err != nil {
		panic(err)
	}

	codec := testapi.Default.Codec()
	json := runtime.EncodeOrDie(codec, &apiv1.SerializedReference{Reference: *ref})
	return string(json)
}

// GenerateOwnerReferences builds OwnerReferences with a single reference
func GenerateOwnerReferences(name, kind, api string, uid types.UID) []metav1.OwnerReference {
	return []metav1.OwnerReference{
		{
			APIVersion:         api,
			Kind:               kind,
			Name:               name,
			BlockOwnerDeletion: boolptr(true),
			Controller:         boolptr(true),
			UID:                uid,
		},
	}
}

func boolptr(val bool) *bool {
	b := val
	return &b
}

// HttpServerMock mocks server HTTP.
//
// Example:
// // Create HttpServerMock.
// server := NewHttpServerMock()
// defer server.Close()
// // Use server.URL to point your code to HttpServerMock.
// g := newTestGceManager(t, server.URL, ModeGKE)
// // Declare handled urls and results for them.
// server.On("handle", "/project1/zones/us-central1-b/listManagedInstances").Return("<managedInstances>").Once()
// // Call http server in your code.
// instances, err := g.GetManagedInstances()
// // Check if expected calls were executed.
// 	mock.AssertExpectationsForObjects(t, server)
type HttpServerMock struct {
	mock.Mock
	*httptest.Server
}

// NewHttpServerMock creates new HttpServerMock.
func NewHttpServerMock() *HttpServerMock {
	httpServerMock := &HttpServerMock{}
	mux := http.NewServeMux()
	mux.HandleFunc("/",
		func(w http.ResponseWriter, req *http.Request) {
			result := httpServerMock.handle(req.URL.Path)
			w.Write([]byte(result))
		})

	server := httptest.NewServer(mux)
	httpServerMock.Server = server
	return httpServerMock
}

func (l *HttpServerMock) handle(url string) string {
	args := l.Called(url)
	return args.String(0)
}
