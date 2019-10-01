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
	apiv1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/utils/drain"
	"k8s.io/autoscaler/cluster-autoscaler/utils/glogx"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"
	"reflect"

	"k8s.io/klog"
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

// PodsSchedulableOnNodeChecker allows for querying what subset of pods from a set is schedulable on given node.
// Pods set is give at creation time and then multiple nodes can be used for querying.
type PodsSchedulableOnNodeChecker struct {
	context               *context.AutoscalingContext
	pods                  []*apiv1.Pod
	podsEquivalenceGroups map[types.UID]equivalenceGroupId
}

// NewPodsSchedulableOnNodeChecker creates an instance of PodsSchedulableOnNodeChecker
func NewPodsSchedulableOnNodeChecker(context *context.AutoscalingContext, pods []*apiv1.Pod) *PodsSchedulableOnNodeChecker {
	checker := PodsSchedulableOnNodeChecker{
		context:               context,
		pods:                  pods,
		podsEquivalenceGroups: make(map[types.UID]equivalenceGroupId),
	}

	// compute the podsEquivalenceGroups
	var nextGroupId equivalenceGroupId
	type equivalanceGroup struct {
		id           equivalenceGroupId
		representant *apiv1.Pod
	}

	equivalenceGroupsByController := make(map[types.UID][]equivalanceGroup)

	for _, pod := range pods {
		controllerRef := drain.ControllerRef(pod)
		if controllerRef == nil {
			checker.podsEquivalenceGroups[pod.UID] = nextGroupId
			nextGroupId++
			continue
		}

		matchingFound := false
		for _, g := range equivalenceGroupsByController[controllerRef.UID] {
			if reflect.DeepEqual(pod.Labels, g.representant.Labels) && apiequality.Semantic.DeepEqual(pod.Spec, g.representant.Spec) {
				matchingFound = true
				checker.podsEquivalenceGroups[pod.UID] = g.id
				break
			}
		}

		if !matchingFound {
			newGroup := equivalanceGroup{
				id:           nextGroupId,
				representant: pod,
			}
			equivalenceGroupsByController[controllerRef.UID] = append(equivalenceGroupsByController[controllerRef.UID], newGroup)
			checker.podsEquivalenceGroups[pod.UID] = newGroup.id
			nextGroupId++
		}
	}

	return &checker
}

// CheckPodsSchedulableOnNode checks if pods can be scheduled on the given node.
func (c *PodsSchedulableOnNodeChecker) CheckPodsSchedulableOnNode(nodeGroupId string, nodeInfo *schedulernodeinfo.NodeInfo) map[*apiv1.Pod]*simulator.PredicateError {
	loggingQuota := glogx.PodsLoggingQuota()
	schedulingErrors := make(map[equivalenceGroupId]*simulator.PredicateError)

	for _, pod := range c.pods {
		equivalenceGroup := c.podsEquivalenceGroups[pod.UID]
		err, found := schedulingErrors[equivalenceGroup]
		if found && err != nil {
			glogx.V(2).UpTo(loggingQuota).Infof("Pod %s can't be scheduled on %s. Used cached predicate check results", pod.Name, nodeGroupId)
		}
		// Not found in cache, have to run the predicates.
		if !found {
			err = c.context.PredicateChecker.CheckPredicates(pod, nil, nodeInfo)
			schedulingErrors[equivalenceGroup] = err
			if err != nil {
				// Always log for the first pod in a controller.
				klog.V(2).Infof("Pod %s can't be scheduled on %s, predicate failed: %v", pod.Name, nodeGroupId, err.VerboseError())
			}
		}
	}
	glogx.V(2).Over(loggingQuota).Infof("%v other pods can't be scheduled on %s.", -loggingQuota.Left(), nodeGroupId)

	schedulingErrorsByPod := make(map[*apiv1.Pod]*simulator.PredicateError)
	for _, pod := range c.pods {
		schedulingErrorsByPod[pod] = schedulingErrors[c.podsEquivalenceGroups[pod.UID]]
	}
	return schedulingErrorsByPod
}
