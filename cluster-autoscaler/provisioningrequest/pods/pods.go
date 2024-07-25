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

package pods

import (
	"fmt"

	"google.golang.org/protobuf/proto"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqwrapper"
	"k8s.io/kubernetes/pkg/controller"
)

const (
	// ProvisioningRequestPodAnnotationKey is a key used to annotate pods consuming provisioning request.
	ProvisioningRequestPodAnnotationKey = "cluster-autoscaler.kubernetes.io/consume-provisioning-request"
	// ProvisioningClassPodAnnotationKey is a key used to add annotation about Provisioning Class
	ProvisioningClassPodAnnotationKey = "cluster-autoscaler.kubernetes.io/provisioning-class-name"
)

// PodsForProvisioningRequest returns a list of pods for which Provisioning
// Request needs to provision resources.
func PodsForProvisioningRequest(pr *provreqwrapper.ProvisioningRequest) ([]*v1.Pod, error) {
	if pr == nil {
		return nil, nil
	}
	podSets, err := pr.PodSets()
	if err != nil {
		return nil, err
	}
	pods := make([]*v1.Pod, 0)
	for i, podSet := range podSets {
		for j := 0; j < int(podSet.Count); j++ {
			pod, err := controller.GetPodFromTemplate(&podSet.PodTemplate, pr.ProvisioningRequest, ownerReference(pr))
			if err != nil {
				return nil, fmt.Errorf("while creating pod for pr: %s/%s podSet: %d, got error: %w", pr.Namespace, pr.Name, i, err)
			}
			populatePodFields(pr, pod, i, j)
			pods = append(pods, pod)
		}
	}
	return pods, nil
}

// ownerReference injects owner reference that points to the ProvReq object.
// This allows CA to group the pods as coming from one controller and simplifies
// the scale-up simulation logic and number of logs lines emitted.
func ownerReference(pr *provreqwrapper.ProvisioningRequest) *metav1.OwnerReference {
	return &metav1.OwnerReference{
		APIVersion: pr.APIVersion,
		Kind:       pr.Kind,
		Name:       pr.Name,
		UID:        pr.UID,
		Controller: proto.Bool(true),
	}
}

func populatePodFields(pr *provreqwrapper.ProvisioningRequest, pod *v1.Pod, i, j int) {
	pod.Name = fmt.Sprintf("%s%d-%d", pod.GenerateName, i, j)
	pod.Namespace = pr.Namespace
	if pod.Annotations == nil {
		pod.Annotations = make(map[string]string)
	}
	pod.Annotations[ProvisioningRequestPodAnnotationKey] = pr.Name
	pod.Annotations[ProvisioningClassPodAnnotationKey] = pr.Spec.ProvisioningClassName
	pod.UID = types.UID(fmt.Sprintf("%s/%s", pod.Namespace, pod.Name))
	pod.CreationTimestamp = pr.CreationTimestamp
}
