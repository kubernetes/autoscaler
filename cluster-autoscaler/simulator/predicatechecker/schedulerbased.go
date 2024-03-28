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

package predicatechecker

import (
	"context"
	"fmt"

	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	v1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/events"
	klog "k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/apis/config"
	scheduler_config "k8s.io/kubernetes/pkg/scheduler/apis/config/latest"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
	scheduler_plugins "k8s.io/kubernetes/pkg/scheduler/framework/plugins"
	schedulerframeworkruntime "k8s.io/kubernetes/pkg/scheduler/framework/runtime"
	scheduler_profile "k8s.io/kubernetes/pkg/scheduler/profile"
)

// SchedulerBasedPredicateChecker checks whether all required predicates pass for given Pod and Node.
// The verification is done by calling out to scheduler code.
type SchedulerBasedPredicateChecker struct {
	profiles               scheduler_profile.Map
	defaultSchedulerName   string
	delegatingSharedLister *DelegatingSchedulerSharedLister
	nodeLister             v1listers.NodeLister
	podLister              v1listers.PodLister
	lastIndex              int
}

// NewSchedulerBasedPredicateChecker builds scheduler based PredicateChecker.
func NewSchedulerBasedPredicateChecker(informerFactory informers.SharedInformerFactory, schedConfig *config.KubeSchedulerConfiguration) (*SchedulerBasedPredicateChecker, error) {
	if schedConfig == nil {
		var err error
		schedConfig, err = scheduler_config.Default()
		if err != nil {
			return nil, fmt.Errorf("couldn't create scheduler config: %v", err)
		}
	}

	if len(schedConfig.Profiles) == 0 {
		return nil, fmt.Errorf("unexpected scheduler config: expected one scheduler profile only (found %d profiles)", len(schedConfig.Profiles))
	}
	sharedLister := NewDelegatingSchedulerSharedLister()

	recorderFactory := func(string) events.EventRecorder { return nil }
	profiles, err := scheduler_profile.NewMap(context.TODO(), schedConfig.Profiles, scheduler_plugins.NewInTreeRegistry(), recorderFactory,
		schedulerframeworkruntime.WithInformerFactory(informerFactory),
		schedulerframeworkruntime.WithSnapshotSharedLister(sharedLister),
	)

	if err != nil {
		return nil, fmt.Errorf("couldn't create scheduler framework; %v", err)
	}

	checker := &SchedulerBasedPredicateChecker{
		profiles:               profiles,
		defaultSchedulerName:   schedConfig.Profiles[0].SchedulerName,
		delegatingSharedLister: sharedLister,
	}

	return checker, nil
}

// FitsAnyNode checks if the given pod can be placed on any of the given nodes.
func (p *SchedulerBasedPredicateChecker) FitsAnyNode(clusterSnapshot clustersnapshot.ClusterSnapshot, pod *apiv1.Pod) (string, error) {
	return p.FitsAnyNodeMatching(clusterSnapshot, pod, func(*schedulerframework.NodeInfo) bool {
		return true
	})
}

// FitsAnyNodeMatching checks if the given pod can be placed on any of the given nodes matching the provided function.
func (p *SchedulerBasedPredicateChecker) FitsAnyNodeMatching(clusterSnapshot clustersnapshot.ClusterSnapshot, pod *apiv1.Pod, nodeMatches func(*schedulerframework.NodeInfo) bool) (string, error) {
	if clusterSnapshot == nil {
		return "", fmt.Errorf("ClusterSnapshot not provided")
	}

	nodeInfosList, err := clusterSnapshot.NodeInfos().List()
	if err != nil {
		// This should never happen.
		//
		// Scheduler requires interface returning error, but no implementation
		// of ClusterSnapshot ever does it.
		klog.Errorf("Error obtaining nodeInfos from schedulerLister")
		return "", fmt.Errorf("error obtaining nodeInfos from schedulerLister")
	}

	p.delegatingSharedLister.UpdateDelegate(clusterSnapshot)
	defer p.delegatingSharedLister.ResetDelegate()

	fwk, err := p.frameworkForPod(pod)
	if err != nil {
		// This shouldn't happen, because we only accept for scheduling the pods
		// which specify a scheduler name that matches one of the profiles.
		klog.Errorf("Error obtaining framework for pod %s: %v", pod.Name, err)
		return "", fmt.Errorf("error obtaining framework for pod")
	}

	state := schedulerframework.NewCycleState()
	preFilterResult, preFilterStatus := fwk.RunPreFilterPlugins(context.TODO(), state, pod)
	if !preFilterStatus.IsSuccess() {
		return "", fmt.Errorf("error running pre filter plugins for pod %s; %s", pod.Name, preFilterStatus.Message())
	}

	for i := range nodeInfosList {
		nodeInfo := nodeInfosList[(p.lastIndex+i)%len(nodeInfosList)]
		if !nodeMatches(nodeInfo) {
			continue
		}

		if !preFilterResult.AllNodes() && !preFilterResult.NodeNames.Has(nodeInfo.Node().Name) {
			continue
		}

		// Be sure that the node is schedulable.
		if nodeInfo.Node().Spec.Unschedulable {
			continue
		}

		filterStatus := fwk.RunFilterPlugins(context.TODO(), state, pod, nodeInfo)
		if filterStatus.IsSuccess() {
			p.lastIndex = (p.lastIndex + i + 1) % len(nodeInfosList)
			return nodeInfo.Node().Name, nil
		}
	}
	return "", fmt.Errorf("cannot put pod %s on any node", pod.Name)
}

// CheckPredicates checks if the given pod can be placed on the given node.
func (p *SchedulerBasedPredicateChecker) CheckPredicates(clusterSnapshot clustersnapshot.ClusterSnapshot, pod *apiv1.Pod, nodeName string) *PredicateError {
	if clusterSnapshot == nil {
		return NewPredicateError(InternalPredicateError, "", "ClusterSnapshot not provided", nil, emptyString)
	}
	nodeInfo, err := clusterSnapshot.NodeInfos().Get(nodeName)
	if err != nil {
		errorMessage := fmt.Sprintf("Error obtaining NodeInfo for name %s; %v", nodeName, err)
		return NewPredicateError(InternalPredicateError, "", errorMessage, nil, emptyString)
	}

	p.delegatingSharedLister.UpdateDelegate(clusterSnapshot)
	defer p.delegatingSharedLister.ResetDelegate()

	fwk, err := p.frameworkForPod(pod)
	if err != nil {
		klog.Errorf("Error obtaining framework for pod %s: %v", pod.Name, err)
		return NewPredicateError(InternalPredicateError, "", "error obtaining framework for pod", nil, emptyString)
	}

	state := schedulerframework.NewCycleState()
	_, preFilterStatus := fwk.RunPreFilterPlugins(context.TODO(), state, pod)
	if !preFilterStatus.IsSuccess() {
		return NewPredicateError(
			InternalPredicateError,
			"",
			preFilterStatus.Message(),
			preFilterStatus.Reasons(),
			emptyString)
	}

	filterStatus := fwk.RunFilterPlugins(context.TODO(), state, pod, nodeInfo)

	if !filterStatus.IsSuccess() {
		filterName := filterStatus.Plugin()
		filterMessage := filterStatus.Message()
		filterReasons := filterStatus.Reasons()
		if filterStatus.IsRejected() {
			return NewPredicateError(
				NotSchedulablePredicateError,
				filterName,
				filterMessage,
				filterReasons,
				p.buildDebugInfo(filterName, nodeInfo))
		}
		return NewPredicateError(
			InternalPredicateError,
			filterName,
			filterMessage,
			filterReasons,
			p.buildDebugInfo(filterName, nodeInfo))
	}

	return nil
}

func (p *SchedulerBasedPredicateChecker) frameworkForPod(pod *apiv1.Pod) (schedulerframework.Framework, error) {
	fwk, ok := p.profiles[p.schedulerNameForPod(pod)]
	if !ok {
		return nil, fmt.Errorf("profile not found for scheduler name %q", pod.Spec.SchedulerName)
	}
	return fwk, nil
}

func (p *SchedulerBasedPredicateChecker) schedulerNameForPod(pod *apiv1.Pod) string {
	if len(pod.Spec.SchedulerName) == 0 {
		return p.defaultSchedulerName
	}
	return pod.Spec.SchedulerName
}

func (p *SchedulerBasedPredicateChecker) buildDebugInfo(filterName string, nodeInfo *schedulerframework.NodeInfo) func() string {
	switch filterName {
	case "TaintToleration":
		taints := nodeInfo.Node().Spec.Taints
		return func() string {
			return fmt.Sprintf("taints on node: %#v", taints)
		}
	default:
		return emptyString
	}
}
