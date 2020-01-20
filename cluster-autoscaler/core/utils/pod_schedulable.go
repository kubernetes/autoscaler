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

package utils

import (
	"reflect"

	apiv1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/utils/drain"
)

// PodSchedulableInfo data structure is used to avoid running predicates #pending_pods * #nodes
// times (which turned out to be very expensive if there are thousands of pending pods).
// This optimization is based on the assumption that if there are that many pods they're
// likely created by controllers (deployment, replication controller, ...).
// So instead of running all predicates for every pod we first check whether we've
// already seen identical pod (in this step we're not binpacking, just checking if
// the pod would fit anywhere right now) and if so we use the result we already
// calculated.
// To decide if two pods are similar enough we check if they have identical label
// and spec and are owned by the same controller. The problem is the whole
// PodSchedulableInfo struct is not hashable and keeping a list and running deep
// equality checks would likely also be expensive. So instead we use controller
// UID as a key in initial lookup and only run full comparison on a set of
// podSchedulableInfos created for pods owned by this controller.
type PodSchedulableInfo struct {
	spec            apiv1.PodSpec
	labels          map[string]string
	schedulingError *simulator.PredicateError
}

// PodSchedulableMap stores mapping from controller ref to PodSchedulableInfo
type PodSchedulableMap map[string][]PodSchedulableInfo

// Match tests if given pod matches PodSchedulableInfo
func (psi *PodSchedulableInfo) Match(pod *apiv1.Pod) bool {
	return reflect.DeepEqual(pod.Labels, psi.labels) && apiequality.Semantic.DeepEqual(pod.Spec, psi.spec)
}

// Get returns scheduling info for given pod if matching one exists in PodSchedulableMap
func (podMap PodSchedulableMap) Get(pod *apiv1.Pod) (*simulator.PredicateError, bool) {
	ref := drain.ControllerRef(pod)
	if ref == nil {
		return nil, false
	}
	uid := string(ref.UID)
	if infos, found := podMap[uid]; found {
		for _, info := range infos {
			if info.Match(pod) {
				return info.schedulingError, true
			}
		}
	}
	return nil, false
}

// Set sets scheduling info for given pod in PodSchedulableMap
func (podMap PodSchedulableMap) Set(pod *apiv1.Pod, err *simulator.PredicateError) {
	ref := drain.ControllerRef(pod)
	if ref == nil {
		return
	}
	uid := string(ref.UID)
	podMap[uid] = append(podMap[uid], PodSchedulableInfo{
		spec:            pod.Spec,
		labels:          pod.Labels,
		schedulingError: err,
	})
}
