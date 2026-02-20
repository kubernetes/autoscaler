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
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/mock"
	apiv1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kube_types "k8s.io/kubernetes/pkg/kubelet/types"
)

// BuildTestPod creates a pod with specified resources.
func BuildTestPod(name string, cpu int64, mem int64, options ...func(*apiv1.Pod)) *apiv1.Pod {
	startTime := metav1.Unix(0, 0)
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
		Status: apiv1.PodStatus{
			StartTime: &startTime,
		},
	}

	if cpu >= 0 {
		pod.Spec.Containers[0].Resources.Requests[apiv1.ResourceCPU] = *resource.NewMilliQuantity(cpu, resource.DecimalSI)
	}
	if mem >= 0 {
		pod.Spec.Containers[0].Resources.Requests[apiv1.ResourceMemory] = *resource.NewQuantity(mem, resource.DecimalSI)
	}
	for _, o := range options {
		o(pod)
	}
	return pod
}

// MarkUnschedulable marks pod as unschedulable.
func MarkUnschedulable() func(*apiv1.Pod) {
	return func(pod *apiv1.Pod) {
		pod.Status.Conditions = []apiv1.PodCondition{
			{
				Status: apiv1.ConditionFalse,
				Type:   apiv1.PodScheduled,
				Reason: apiv1.PodReasonUnschedulable,
			},
		}
	}
}

// AddSchedulerName adds scheduler name to a pod.
func AddSchedulerName(schedulerName string) func(*apiv1.Pod) {
	return func(pod *apiv1.Pod) {
		pod.Spec.SchedulerName = schedulerName
	}
}

// WithResourceClaim adds a reference to the given resource claim/claim template to a pod.
func WithResourceClaim(refName, claimName, templateName string) func(*apiv1.Pod) {
	return func(pod *apiv1.Pod) {
		claimRef := apiv1.PodResourceClaim{
			Name: refName,
		}
		claimStatus := apiv1.PodResourceClaimStatus{
			Name: refName,
		}

		if templateName != "" {
			claimRef.ResourceClaimTemplateName = &templateName
			claimStatus.ResourceClaimName = &claimName
		} else {
			claimRef.ResourceClaimName = &claimName
		}

		pod.Spec.ResourceClaims = append(pod.Spec.ResourceClaims, claimRef)
		pod.Status.ResourceClaimStatuses = append(pod.Status.ResourceClaimStatuses, claimStatus)
	}
}

// WithControllerOwnerRef sets an owner reference to the pod.
func WithControllerOwnerRef(name, kind string, uid types.UID) func(*apiv1.Pod) {
	return func(pod *apiv1.Pod) {
		pod.OwnerReferences = GenerateOwnerReferences(name, kind, "apps/v1", uid)
	}
}

// WithDSController creates a daemonSet owner ref for the pod.
func WithDSController() func(*apiv1.Pod) {
	return func(pod *apiv1.Pod) {
		pod.OwnerReferences = GenerateOwnerReferences("ds", "DaemonSet", "apps/v1", "some-uid")
	}
}

// WithNodeName sets a node name to the pod.
func WithNodeName(nodeName string) func(*apiv1.Pod) {
	return func(pod *apiv1.Pod) {
		pod.Spec.NodeName = nodeName
	}
}

// WithNamespace sets a namespace to the pod.
func WithNamespace(namespace string) func(*apiv1.Pod) {
	return func(pod *apiv1.Pod) {
		pod.ObjectMeta.Namespace = namespace
	}
}

// WithLabels sets a Labels to the pod.
func WithLabels(labels map[string]string) func(*apiv1.Pod) {
	return func(pod *apiv1.Pod) {
		pod.ObjectMeta.Labels = labels
	}
}

// WithHostPort sets a namespace to the pod.
func WithHostPort(hostport int32) func(*apiv1.Pod) {
	return func(pod *apiv1.Pod) {
		if hostport > 0 {
			pod.Spec.Containers[0].Ports = []apiv1.ContainerPort{
				{
					HostPort: hostport,
				},
			}
		}
	}
}

// WithMaxSkew sets a topology spread constraint to the pod.
func WithMaxSkew(maxSkew int32, topologySpreadingKey string, minDomains int32) func(*apiv1.Pod) {
	return func(pod *apiv1.Pod) {
		if maxSkew > 0 {
			pod.Spec.TopologySpreadConstraints = []apiv1.TopologySpreadConstraint{
				{
					MaxSkew:           maxSkew,
					TopologyKey:       topologySpreadingKey,
					WhenUnsatisfiable: "DoNotSchedule",
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "estimatee",
						},
					},
					MinDomains: &minDomains,
				},
			}
		}
	}
}

// WithCreationTimestamp sets creation timestamp to the pod.
func WithCreationTimestamp(timestamp time.Time) func(*apiv1.Pod) {
	return func(pod *apiv1.Pod) {
		pod.CreationTimestamp = metav1.Time{Time: timestamp}
	}
}

// WithDeletionTimestamp sets deletion timestamp to the pod.
func WithDeletionTimestamp(deletionTimestamp time.Time) func(*apiv1.Pod) {
	return func(pod *apiv1.Pod) {
		pod.DeletionTimestamp = &metav1.Time{Time: deletionTimestamp}
	}
}

// WithNodeNamesAffinity sets pod's affinity for specific nodes.
func WithNodeNamesAffinity(nodeNames ...string) func(*apiv1.Pod) {
	return func(pod *apiv1.Pod) {
		pod.Spec.Affinity = &apiv1.Affinity{
			NodeAffinity: &apiv1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &apiv1.NodeSelector{
					NodeSelectorTerms: []apiv1.NodeSelectorTerm{
						{
							MatchFields: []apiv1.NodeSelectorRequirement{
								{
									Key:      metav1.ObjectNameField,
									Operator: apiv1.NodeSelectorOpIn,
									Values:   nodeNames,
								},
							},
						},
					},
				},
			},
		}
	}
}

// WithPodHostnameAntiAffinity sets pod's anti-affinity for pods matching the given labels at hostname topology level.
func WithPodHostnameAntiAffinity(labels map[string]string) func(*apiv1.Pod) {
	return func(pod *apiv1.Pod) {
		if pod.Spec.Affinity == nil {
			pod.Spec.Affinity = &apiv1.Affinity{}
		}
		pod.Spec.Affinity.PodAntiAffinity = &apiv1.PodAntiAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: []apiv1.PodAffinityTerm{
				{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: labels,
					},
					TopologyKey: "kubernetes.io/hostname",
				},
			},
		}
	}
}

// BuildTestPodWithEphemeralStorage creates a pod with cpu, memory and ephemeral storage resources.
func BuildTestPodWithEphemeralStorage(name string, cpu, mem, ephemeralStorage int64) *apiv1.Pod {
	startTime := metav1.Unix(0, 0)
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
		Status: apiv1.PodStatus{
			StartTime: &startTime,
		},
	}

	if cpu >= 0 {
		pod.Spec.Containers[0].Resources.Requests[apiv1.ResourceCPU] = *resource.NewMilliQuantity(cpu, resource.DecimalSI)
	}
	if mem >= 0 {
		pod.Spec.Containers[0].Resources.Requests[apiv1.ResourceMemory] = *resource.NewQuantity(mem, resource.DecimalSI)
	}
	if ephemeralStorage >= 0 {
		pod.Spec.Containers[0].Resources.Requests[apiv1.ResourceEphemeralStorage] = *resource.NewQuantity(ephemeralStorage, resource.DecimalSI)
	}

	return pod
}

// BuildScheduledTestPod builds a scheduled test pod with a given spec
func BuildScheduledTestPod(name string, cpu, memory int64, nodeName string) *apiv1.Pod {
	p := BuildTestPod(name, cpu, memory)
	p.Spec.NodeName = nodeName
	return p
}

// SetStaticPodSpec sets pod spec to make it a static pod
func SetStaticPodSpec(pod *apiv1.Pod) *apiv1.Pod {
	pod.Annotations[kube_types.ConfigSourceAnnotationKey] = kube_types.FileSource
	return pod
}

// SetMirrorPodSpec sets pod spec to make it a mirror pod
func SetMirrorPodSpec(pod *apiv1.Pod) *apiv1.Pod {
	pod.ObjectMeta.Annotations[kube_types.ConfigMirrorAnnotationKey] = "mirror"
	return pod
}

// SetDSPodSpec sets pod spec to make it a DS pod
func SetDSPodSpec(pod *apiv1.Pod) *apiv1.Pod {
	pod.OwnerReferences = GenerateOwnerReferences("ds", "DaemonSet", "apps/v1", "api/v1/namespaces/default/daemonsets/ds")
	return pod
}

// SetRSPodSpec sets pod spec to make it a RS pod
func SetRSPodSpec(pod *apiv1.Pod, rsName string) *apiv1.Pod {
	pod.OwnerReferences = GenerateOwnerReferences(rsName, "ReplicaSet", "extensions/v1beta1", types.UID(rsName))
	return pod
}

// BuildServiceTokenProjectedVolumeSource returns a ProjectedVolumeSource with SA token
// projection
func BuildServiceTokenProjectedVolumeSource(path string) *apiv1.ProjectedVolumeSource {
	return &apiv1.ProjectedVolumeSource{
		Sources: []apiv1.VolumeProjection{
			{
				ServiceAccountToken: &apiv1.ServiceAccountTokenProjection{
					Path: path,
				},
			},
		},
	}
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

// TolerateGpuForPod adds toleration for nvidia.com/gpu to Pod
func TolerateGpuForPod(pod *apiv1.Pod) {
	pod.Spec.Tolerations = append(pod.Spec.Tolerations, apiv1.Toleration{Key: resourceNvidiaGPU, Operator: apiv1.TolerationOpExists})
}

// NodeOption is a function that modifies a Node during construction.
type NodeOption func(*apiv1.Node)

// IsReady sets the node to a Ready state.
func IsReady(ready bool) NodeOption {
	return func(node *apiv1.Node) {
		SetNodeReadyState(node, ready, time.Now())
	}
}

// BuildTestNode creates a node with specified capacity.
func BuildTestNode(name string, millicpuCapacity int64, memCapacity int64, opts ...NodeOption) *apiv1.Node {
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

	if millicpuCapacity >= 0 {
		node.Status.Capacity[apiv1.ResourceCPU] = *resource.NewMilliQuantity(millicpuCapacity, resource.DecimalSI)
	}
	if memCapacity >= 0 {
		node.Status.Capacity[apiv1.ResourceMemory] = *resource.NewQuantity(memCapacity, resource.DecimalSI)
	}

	node.Status.Allocatable = apiv1.ResourceList{}
	for k, v := range node.Status.Capacity {
		node.Status.Allocatable[k] = v
	}

	for _, opt := range opts {
		opt(node)
	}

	return node
}

// BuildCSINode returns a CSINode object given a node object
func BuildCSINode(node *apiv1.Node) *storagev1.CSINode {
	return &storagev1.CSINode{
		ObjectMeta: metav1.ObjectMeta{
			Name: node.Name,
			UID:  types.UID("csi-node" + node.UID),
		},
		Spec: storagev1.CSINodeSpec{
			Drivers: []storagev1.CSINodeDriver{
				{
					Name:         "ebs.csi.aws.com",
					NodeID:       string(node.UID),
					TopologyKeys: []string{"topology.ebs.csi.aws.com/zone"},
				},
			},
		},
	}
}

// WithSchedulingGatedStatus upserts the condition with type PodScheduled to be of status false
// and reason PodReasonSchedulingGated
func WithSchedulingGatedStatus(pod *apiv1.Pod) *apiv1.Pod {
	gatedPodCondition := apiv1.PodCondition{
		Type:   apiv1.PodScheduled,
		Status: apiv1.ConditionFalse,
		Reason: apiv1.PodReasonSchedulingGated,
	}
	for index := range pod.Status.Conditions {
		if pod.Status.Conditions[index].Type == apiv1.PodScheduled {
			pod.Status.Conditions[index] = gatedPodCondition
			return pod
		}
	}
	pod.Status.Conditions = append(pod.Status.Conditions, gatedPodCondition)
	return pod
}

// WithAllocatable adds specified milliCpu and memory to Allocatable of the node in-place.
func WithAllocatable(node *apiv1.Node, millicpuAllocatable, memAllocatable int64) *apiv1.Node {
	node.Status.Allocatable[apiv1.ResourceCPU] = *resource.NewMilliQuantity(millicpuAllocatable, resource.DecimalSI)
	node.Status.Allocatable[apiv1.ResourceMemory] = *resource.NewQuantity(memAllocatable, resource.DecimalSI)
	return node
}

// AddEphemeralStorageToNode adds ephemeral storage capacity to a given node.
func AddEphemeralStorageToNode(node *apiv1.Node, eph int64) *apiv1.Node {
	if eph >= 0 {
		node.Status.Capacity[apiv1.ResourceEphemeralStorage] = *resource.NewQuantity(eph, resource.DecimalSI)
		node.Status.Allocatable[apiv1.ResourceEphemeralStorage] = *resource.NewQuantity(eph, resource.DecimalSI)
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
		node.Spec.Taints = append(node.Spec.Taints, apiv1.Taint{
			Key:    "node.kubernetes.io/not-ready",
			Value:  "true",
			Effect: apiv1.TaintEffectNoSchedule,
		})
	}
}

// SetNodeNotReadyTaint sets the not ready taint on node.
func SetNodeNotReadyTaint(node *apiv1.Node) {
	node.Spec.Taints = append(node.Spec.Taints, apiv1.Taint{Key: apiv1.TaintNodeNotReady, Effect: apiv1.TaintEffectNoSchedule})
}

// RemoveNodeNotReadyTaint removes the not ready taint.
func RemoveNodeNotReadyTaint(node *apiv1.Node) {
	var final []apiv1.Taint
	for i := range node.Spec.Taints {
		if node.Spec.Taints[i].Key == apiv1.TaintNodeNotReady {
			continue
		}
		final = append(final, node.Spec.Taints[i])
	}
	node.Spec.Taints = final
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
//
//	mock.AssertExpectationsForObjects(t, server)
//
// Note: to provide a content type, you may pass in the desired
// fields:
// server := NewHttpServerMock(MockFieldContentType, MockFieldResponse)
// ...
// server.On("handle", "/project1/zones/us-central1-b/listManagedInstances").Return("<content type>", "<response>").Once()
// The order of the return objects must match that of the HttpServerMockField constants passed to NewHttpServerMock()
type HttpServerMock struct {
	mock.Mock
	*httptest.Server
	fields []HttpServerMockField
}

// HttpServerMockField specifies a type of field.
type HttpServerMockField int

const (
	// MockFieldResponse represents a string response.
	MockFieldResponse HttpServerMockField = iota
	// MockFieldStatusCode represents an integer HTTP response code.
	MockFieldStatusCode
	// MockFieldContentType represents a string content type.
	MockFieldContentType
	// MockFieldUserAgent represents a string user agent.
	MockFieldUserAgent
)

// NewHttpServerMock creates new HttpServerMock.
func NewHttpServerMock(fields ...HttpServerMockField) *HttpServerMock {
	if len(fields) == 0 {
		fields = []HttpServerMockField{MockFieldResponse}
	}
	foundResponse := false
	for _, field := range fields {
		if field == MockFieldResponse {
			foundResponse = true
			break
		}
	}
	if !foundResponse {
		panic("Must use MockFieldResponse.")
	}
	httpServerMock := &HttpServerMock{fields: fields}
	mux := http.NewServeMux()
	mux.HandleFunc("/",
		func(w http.ResponseWriter, req *http.Request) {
			result := httpServerMock.handle(req, w, httpServerMock)
			_, _ = w.Write([]byte(result))
		})

	server := httptest.NewServer(mux)
	httpServerMock.Server = server
	return httpServerMock
}

func (l *HttpServerMock) handle(req *http.Request, w http.ResponseWriter, serverMock *HttpServerMock) string {
	url := req.URL.Path
	query := req.URL.Query()
	var response string
	var args mock.Arguments
	if query.Has("pageToken") {
		args = l.Called(url, query.Get("pageToken"))
	} else {
		args = l.Called(url)
	}
	for i, field := range l.fields {
		switch field {
		case MockFieldResponse:
			response = args.String(i)
		case MockFieldContentType:
			w.Header().Set("Content-Type", args.String(i))
		case MockFieldStatusCode:
			w.WriteHeader(args.Int(i))
		case MockFieldUserAgent:
			gotUserAgent := req.UserAgent()
			expectedUserAgent := args.String(i)
			if !strings.Contains(gotUserAgent, expectedUserAgent) {
				panic(fmt.Sprintf("Error handling URL %s, expected user agent %s but got %s.", url, expectedUserAgent, gotUserAgent))
			}
		}
	}
	return response
}

// IgnoreObjectOrder returns a cmp.Option that ignores the order of elements when comparing slices of K8s objects of type T,
// depending on their GetName() function for sorting.
func IgnoreObjectOrder[T interface{ GetName() string }]() cmp.Option {
	return cmpopts.SortSlices(func(c1, c2 T) bool {
		return c1.GetName() < c2.GetName()
	})
}
