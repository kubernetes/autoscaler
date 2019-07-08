/*
Copyright 2018 The Kubernetes Authors.

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

package priority

import (
	apiv1 "k8s.io/api/core/v1"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
)

// PodEvictionAdmission controls evictions of pods.
type PodEvictionAdmission interface {
	// LoopInit initializes PodEvictionAdmission for next Updater loop with the live pods and
	// pods currently controlled by VPA in this cluster.
	LoopInit(allLivePods []*apiv1.Pod, vpaControlledPods map[*vpa_types.VerticalPodAutoscaler][]*apiv1.Pod)
	// Admit returns true if PodEvictionAdmission decides that pod can be evicted with given recommendation.
	Admit(pod *apiv1.Pod, recommendation *vpa_types.RecommendedPodResources) bool
	// CleanUp cleans up any state that PodEvictionAdmission may keep. Called
	// when no VPA objects are present in the cluster.
	CleanUp()
}

// NewDefaultPodEvictionAdmission constructs new PodEvictionAdmission that admits all pods.
func NewDefaultPodEvictionAdmission() PodEvictionAdmission {
	return &noopPodEvictionAdmission{}
}

// NewSequentialPodEvictionAdmission constructs PodEvictionAdmission that will chain provided PodEvictionAdmission objects
func NewSequentialPodEvictionAdmission(admissions []PodEvictionAdmission) PodEvictionAdmission {
	return &sequentialPodEvictionAdmission{admissions: admissions}
}

type sequentialPodEvictionAdmission struct {
	admissions []PodEvictionAdmission
}

func (a *sequentialPodEvictionAdmission) LoopInit(allLivePods []*apiv1.Pod, vpaControlledPods map[*vpa_types.VerticalPodAutoscaler][]*apiv1.Pod) {
	for _, admission := range a.admissions {
		admission.LoopInit(allLivePods, vpaControlledPods)
	}
}

func (a *sequentialPodEvictionAdmission) Admit(pod *apiv1.Pod, recommendation *vpa_types.RecommendedPodResources) bool {
	for _, admission := range a.admissions {
		admit := admission.Admit(pod, recommendation)
		if !admit {
			return false
		}
	}
	return true
}

func (a *sequentialPodEvictionAdmission) CleanUp() {
	for _, admission := range a.admissions {
		admission.CleanUp()
	}
}

type noopPodEvictionAdmission struct{}

func (n *noopPodEvictionAdmission) LoopInit(allLivePods []*apiv1.Pod, vpaControlledPods map[*vpa_types.VerticalPodAutoscaler][]*apiv1.Pod) {
}
func (n *noopPodEvictionAdmission) Admit(pod *apiv1.Pod, recommendation *vpa_types.RecommendedPodResources) bool {
	return true
}
func (n *noopPodEvictionAdmission) CleanUp() {
}
