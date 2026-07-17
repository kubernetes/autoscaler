/*
Copyright 2019 The Kubernetes Authors.

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

package scheduling

import (
	"reflect"

	"k8s.io/autoscaler/cluster-autoscaler/utils"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/utils/drain"
	pod_utils "k8s.io/autoscaler/cluster-autoscaler/utils/pod"
)

// SimilarPodsSchedulingInfo data structure is used to avoid running predicates #pending_pods * #nodes
// times (which turned out to be very expensive if there are thousands of pending pods).
// This optimization is based on the assumption that if there are that many pods they're
// likely created by controllers (deployment, replication controller, ...).
// So instead of running all predicates for every pod we first check whether we've
// already seen identical pod (in this step we're not binpacking, just checking if
// the pod would fit anywhere right now) and if so we use the result we already
// calculated.
// To decide if two pods are similar enough we check if they have identical label
// and spec and are owned by the same controller. The problem is the whole
// SimilarPodsSchedulingInfo struct is not hashable and keeping a list and running deep
// equality checks would likely also be expensive. So instead we use controller
// UID as a key in initial lookup and only run full comparison on a set of
// SimilarPodsSchedulingInfo created for pods owned by this controller.
type SimilarPodsSchedulingInfo struct {
	spec   apiv1.PodSpec
	labels map[string]string
}

// Match tests if given pod matches SimilarPodsSchedulingInfo
func (psi *SimilarPodsSchedulingInfo) Match(pod *apiv1.Pod) bool {
	return reflect.DeepEqual(pod.Labels, psi.labels) && utils.PodSpecSemanticallyEqual(pod.Spec, psi.spec)
}

const maxPodsPerOwnerRef = 10

// SimilarPodsScheduling stores mapping from controller ref to SimilarPodsSchedulingInfo
type SimilarPodsScheduling struct {
	items                  map[string][]SimilarPodsSchedulingInfo
	overflowingControllers map[string]bool
}

// NewSimilarPodsScheduling creates a new SimilarPodsScheduling
func NewSimilarPodsScheduling() *SimilarPodsScheduling {
	return &SimilarPodsScheduling{
		items:                  make(map[string][]SimilarPodsSchedulingInfo),
		overflowingControllers: make(map[string]bool),
	}
}

// IsSimilarUnschedulable returns scheduling info for given pod if matching one exists in SimilarPodsScheduling
func (p *SimilarPodsScheduling) IsSimilarUnschedulable(pod *apiv1.Pod) bool {
	ref := drain.ControllerRef(pod)
	if ref == nil {
		return false
	}
	uid := string(ref.UID)
	if infos, found := p.items[uid]; found {
		for _, info := range infos {
			if info.Match(pod) {
				return true
			}
		}
	}
	return false
}

// SetUnschedulable sets scheduling info for given pod in SimilarPodsScheduling
func (p *SimilarPodsScheduling) SetUnschedulable(pod *apiv1.Pod) {
	ref := drain.ControllerRef(pod)
	if ref == nil || pod_utils.IsDaemonSetPod(pod) {
		return
	}
	uid := string(ref.UID)
	pm := p.items[uid]
	if len(pm) >= maxPodsPerOwnerRef {
		// Too many different pods per owner reference. Don't cache the
		// entry to avoid O(N) search in Get(). It would defeat the
		// benefits from caching anyway.
		p.overflowingControllers[uid] = true
		return
	}
	p.items[uid] = append(pm, SimilarPodsSchedulingInfo{
		spec:   pod.Spec,
		labels: pod.Labels,
	})
}

// OverflowingControllerCount returns the number of controllers that had too
// many different pods to be effectively cached.
func (p *SimilarPodsScheduling) OverflowingControllerCount() int {
	return len(p.overflowingControllers)
}
