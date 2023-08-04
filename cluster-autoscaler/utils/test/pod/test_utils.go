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

package pod

import (
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	common "k8s.io/autoscaler/cluster-autoscaler/utils/test/common"
	kube_types "k8s.io/kubernetes/pkg/kubelet/types"
)

const (
	// cannot use constants from gpu module due to cyclic package import
	resourceNvidiaGPU = "nvidia.com/gpu"
	gpuLabel          = "cloud.google.com/gke-accelerator"
	defaultGPUType    = "nvidia-tesla-k80"
)

func basePod(name string) *apiv1.Pod {
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

	return pod
}

// WithCPU adds requests.cpu (unit: number of CPUs) to all containers in the pod
func WithCPU(cpu int64) func(*apiv1.Pod) {
	return func(pod *apiv1.Pod) {

		for i := range pod.Spec.Containers {
			pod.Spec.Containers[i].Resources.Requests[apiv1.ResourceCPU] = *resource.NewQuantity(cpu, resource.DecimalSI)
		}
	}
}

// WithMilliCPU adds requests.cpu (unit: number of milli CPUs) to all containers in the pod
func WithMilliCPU(cpu int64) func(*apiv1.Pod) {
	return func(pod *apiv1.Pod) {

		for i := range pod.Spec.Containers {
			pod.Spec.Containers[i].Resources.Requests[apiv1.ResourceCPU] = *resource.NewMilliQuantity(cpu, resource.DecimalSI)
		}
	}
}

// WithMemory adds requests.memory to all the containers in the pod
func WithMemory(mem int64) func(*apiv1.Pod) {

	return func(pod *apiv1.Pod) {

		for i := range pod.Spec.Containers {
			pod.Spec.Containers[i].Resources.Requests[apiv1.ResourceMemory] = *resource.NewQuantity(mem, resource.DecimalSI)
		}
	}
}

// WithGPU adds requests.nvidia.com/gpu and limits.nvidia.com/gpu
// to all the containers in the pod
func WithGPU(gpusCount int64) func(*apiv1.Pod) {

	return func(pod *apiv1.Pod) {

		for i := range pod.Spec.Containers {
			pod.Spec.Containers[i].Resources.Requests[resourceNvidiaGPU] = *resource.NewQuantity(gpusCount, resource.DecimalSI)
			pod.Spec.Containers[i].Resources.Limits[resourceNvidiaGPU] = *resource.NewQuantity(gpusCount, resource.DecimalSI)
		}
	}
}

// WithGPUToleration adds nvidia.com/gpu:Exists toleration to the pod
func WithGPUToleration() func(*apiv1.Pod) {

	return func(pod *apiv1.Pod) {
		pod.Spec.Tolerations = append(pod.Spec.Tolerations, apiv1.Toleration{Key: resourceNvidiaGPU, Operator: apiv1.TolerationOpExists})
	}
}

// WithAnnotations adds annotations to the pods
// note: if a key already exists, it will be overwritten
func WithAnnotations(anns map[string]string) func(*apiv1.Pod) {

	return func(pod *apiv1.Pod) {

		for k, v := range anns {
			pod.Annotations[k] = v
		}
	}
}

// WithOwnerRef adds owner ref to pod
func WithOwnerRef(ownerRefs []metav1.OwnerReference) func(*apiv1.Pod) {

	return func(pod *apiv1.Pod) {
		pod.OwnerReferences = append(pod.OwnerReferences, ownerRefs...)
	}
}

// WithEphemeralStorage adds requests.ephemeral-storage to all the containers in the pod
func WithEphemeralStorage(ephemeralStorage int64) func(*apiv1.Pod) {

	return func(pod *apiv1.Pod) {
		for i := range pod.Spec.Containers {
			pod.Spec.Containers[i].Resources.Requests[apiv1.ResourceEphemeralStorage] = *resource.NewQuantity(ephemeralStorage, resource.DecimalSI)
		}
	}
}

// ScheduledOnNode sets pod's spec.nodeName
func ScheduledOnNode(nodeName string) func(*apiv1.Pod) {

	return func(pod *apiv1.Pod) {
		pod.Spec.NodeName = nodeName
	}
}

// WithStaticPodAnnotation adds 'kubernetes.io/config.source: file' annotation to the pod
func WithStaticPodAnnotation() func(*apiv1.Pod) {

	return func(pod *apiv1.Pod) {
		pod.Annotations[kube_types.ConfigSourceAnnotationKey] = kube_types.FileSource
	}
}

// AsDaemonSetPod adds daemonset owner ref to the pod
func AsDaemonSetPod(uid string) func(*apiv1.Pod) {

	return func(pod *apiv1.Pod) {
		pod.OwnerReferences = common.GenerateOwnerReferences("ds", "DaemonSet", "apps/v1", types.UID(uid))
	}
}

// AsReplicaSetPod adds replicaset owner ref to the pod
func AsReplicaSetPod(uid string, rsName string) func(*apiv1.Pod) {

	return func(pod *apiv1.Pod) {
		pod.OwnerReferences = common.GenerateOwnerReferences(rsName, "ReplicaSet", "extensions/v1beta1", types.UID(uid))
	}
}

// WithMirrorPodAnnotation adds 'kubernetes.io/config.mirror: mirror' annotation to the pod
func WithMirrorPodAnnotation() func(*apiv1.Pod) {

	return func(pod *apiv1.Pod) {
		pod.ObjectMeta.Annotations[kube_types.ConfigMirrorAnnotationKey] = "mirror"
	}
}

// NewTestPod creates pod with options passed as functions
func NewTestPod(name string, options ...func(*apiv1.Pod)) *apiv1.Pod {
	pod := basePod(name)

	for _, o := range options {
		o(pod)
	}

	return pod
}
