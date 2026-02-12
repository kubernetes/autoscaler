/*
Copyright 2017 The Kubernetes Authors.

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

package restriction

import (
	corev1 "k8s.io/api/core/v1"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
)

// FakePodsRestrictionFactory is a fake implementation of the PodsRestrictionFactory interface.
type FakePodsRestrictionFactory struct {
	// Eviction is the fake eviction restriction.
	Eviction PodsEvictionRestriction
	// InPlace is the fake in-place restriction.
	InPlace PodsInPlaceRestriction
}

// NewPodsEvictionRestriction returns the fake eviction restriction.
func (f *FakePodsRestrictionFactory) NewPodsEvictionRestriction(creatorToSingleGroupStatsMap map[podReplicaCreator]singleGroupStats, podToReplicaCreatorMap map[string]podReplicaCreator) PodsEvictionRestriction {
	return f.Eviction
}

// NewPodsInPlaceRestriction returns the fake in-place restriction.
func (f *FakePodsRestrictionFactory) NewPodsInPlaceRestriction(creatorToSingleGroupStatsMap map[podReplicaCreator]singleGroupStats, podToReplicaCreatorMap map[string]podReplicaCreator) PodsInPlaceRestriction {
	return f.InPlace
}

// GetCreatorMaps returns nil maps.
func (f *FakePodsRestrictionFactory) GetCreatorMaps(pods []*corev1.Pod, vpa *vpa_types.VerticalPodAutoscaler) (map[podReplicaCreator]singleGroupStats, map[string]podReplicaCreator, error) {
	return nil, nil, nil
}
