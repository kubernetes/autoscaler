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

package simulator

import (
	"context"
	"fmt"
	"time"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	kube_client "k8s.io/client-go/kubernetes"

	scheduler_apis_config "k8s.io/kubernetes/pkg/scheduler/apis/config"
	scheduler_plugins "k8s.io/kubernetes/pkg/scheduler/framework/plugins"
	scheduler_framework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"
	scheduler_listers "k8s.io/kubernetes/pkg/scheduler/listers"
	scheduler_nodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"
	scheduler_volumebinder "k8s.io/kubernetes/pkg/scheduler/volumebinder"
	// We need to import provider to initialize default scheduler.
	"k8s.io/kubernetes/pkg/scheduler/algorithmprovider"
)

// PredicateChecker checks whether all required predicates pass for given Pod and Node.
type PredicateChecker struct {
	framework scheduler_framework.Framework
	snapshot  scheduler_listers.SharedLister
}

// We run some predicates first as they are cheap to check and they should be enough
// to fail predicates in most of our simulations (especially binpacking).
// There are no const arrays in Go, this is meant to be used as a const.
var priorityPredicates = []string{"PodFitsResources", "PodToleratesNodeTaints", "GeneralPredicates", "ready"}

// NoOpEventRecorder is a noop implementation of EventRecorder
type NoOpEventRecorder struct{}

// Event is a noop method implementation
func (NoOpEventRecorder) Event(object runtime.Object, eventtype, reason, message string) {
}

// Eventf is a noop method implementation
func (NoOpEventRecorder) Eventf(regarding runtime.Object, related runtime.Object, eventtype, reason, action, note string, args ...interface{}) {
}

// PastEventf is a noop method implementation
func (NoOpEventRecorder) PastEventf(object runtime.Object, timestamp metav1.Time, eventtype, reason, messageFmt string, args ...interface{}) {
}

// AnnotatedEventf is a noop method implementation
func (NoOpEventRecorder) AnnotatedEventf(object runtime.Object, annotations map[string]string, eventtype, reason, messageFmt string, args ...interface{}) {
}

// NewPredicateChecker builds PredicateChecker.
func NewPredicateChecker(kubeClient kube_client.Interface, stop <-chan struct{}) (*PredicateChecker, error) {
	informerFactory := informers.NewSharedInformerFactory(kubeClient, 0)
	providerRegistry := algorithmprovider.NewRegistry(1) // 1 here is hardPodAffinityWeight not relevant for CA
	config := providerRegistry[scheduler_apis_config.SchedulerDefaultProviderName]
	snapshot := NewEmptySnapshot()

	volumeBinder := scheduler_volumebinder.NewVolumeBinder(
		kubeClient,
		informerFactory.Core().V1().Nodes(),
		informerFactory.Storage().V1().CSINodes(),
		informerFactory.Core().V1().PersistentVolumeClaims(),
		informerFactory.Core().V1().PersistentVolumes(),
		informerFactory.Storage().V1().StorageClasses(),
		time.Duration(10)*time.Second,
	)

	framework, err := scheduler_framework.NewFramework(
		scheduler_plugins.NewInTreeRegistry(),
		config.FrameworkPlugins,
		config.FrameworkPluginConfig,
		scheduler_framework.WithInformerFactory(informerFactory),
		scheduler_framework.WithSnapshotSharedLister(snapshot),
		scheduler_framework.WithVolumeBinder(volumeBinder),
	)

	if err != nil {
		return nil, fmt.Errorf("couldn't create scheduler framework; %v", err)
	}

	// TODO(scheduler_framework) How do I update the snapshot? Every loop?

	// this MUST be called after all the informers/listers are acquired via the
	// informerFactory....Lister()/informerFactory....Informer() methods
	informerFactory.Start(stop)

	return &PredicateChecker{
		framework: framework,
		snapshot:  snapshot,
	}, nil
}

// NewTestPredicateChecker builds test version of PredicateChecker.
func NewTestPredicateChecker() *PredicateChecker {
	// TODO(scheduler_framework)
	return nil
}

// SnapshotClusterState updates cluster snapshot used by the predicate checker.
// It should be called every CA loop iteration.
func (p *PredicateChecker) SnapshotClusterState() error {
	// TODO(scheduler_framework rebuild snapshot
	return nil
}

// FitsAny checks if the given pod can be place on any of the given nodes.
func (p *PredicateChecker) FitsAny(pod *apiv1.Pod, nodeInfos map[string]*scheduler_nodeinfo.NodeInfo) (string, error) {
	// TODO(scheduler_framework) run prefilter only once
	for name, nodeInfo := range nodeInfos {
		// Be sure that the node is schedulable.
		if nodeInfo.Node().Spec.Unschedulable {
			continue
		}
		if err := p.CheckPredicates(pod, nodeInfo); err == nil {
			return name, nil
		}
	}
	return "", fmt.Errorf("cannot put pod %s on any node", pod.Name)
}

// CheckPredicates checks if the given pod can be placed on the given node.
func (p *PredicateChecker) CheckPredicates(pod *apiv1.Pod, nodeInfo *scheduler_nodeinfo.NodeInfo) *PredicateError {
	state := scheduler_framework.NewCycleState()
	preFilterStatus := p.framework.RunPreFilterPlugins(context.TODO(), state, pod)
	if !preFilterStatus.IsSuccess() {
		return NewPredicateError(
			InternalPredicateError,
			"",
			preFilterStatus.Message(),
			preFilterStatus.Reasons(),
			emptyString)
	}

	filterStatuses := p.framework.RunFilterPlugins(context.TODO(), state, pod, nodeInfo)
	for filterName, filterStatus := range filterStatuses {
		if !filterStatus.IsSuccess() {
			if filterStatus.IsUnschedulable() {
				return NewPredicateError(
					NotSchedulablePredicateError,
					filterName,
					filterStatus.Message(),
					filterStatus.Reasons(),
					p.buildDebugInfo(filterName, nodeInfo))
			}
			return NewPredicateError(
				InternalPredicateError,
				filterName,
				filterStatus.Message(),
				filterStatus.Reasons(),
				p.buildDebugInfo(filterName, nodeInfo))
		}
	}
	return nil
}

func (p *PredicateChecker) buildDebugInfo(filterName string, nodeInfo *scheduler_nodeinfo.NodeInfo) func() string {
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
