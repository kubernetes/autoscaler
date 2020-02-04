/*
Copyright 2020 The Kubernetes Authors.

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
	"fmt"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	scheduler_listers "k8s.io/kubernetes/pkg/scheduler/listers"
	scheduler_nodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"
)

// DelegatingSchedulerSharedLister is an implementation of scheduler.SharedLister which
// passes logic to delegate. Delegate can be updated.
type DelegatingSchedulerSharedLister struct {
	delegate scheduler_listers.SharedLister
}

// NewDelegatingSchedulerSharedLister creates new NewDelegatingSchedulerSharedLister
func NewDelegatingSchedulerSharedLister() *DelegatingSchedulerSharedLister {
	return &DelegatingSchedulerSharedLister{
		delegate: unsetSharedListerSingleton,
	}
}

// Pods returns a PodLister
func (lister *DelegatingSchedulerSharedLister) Pods() scheduler_listers.PodLister {
	return lister.delegate.Pods()
}

// NodeInfos returns a NodeInfoLister.
func (lister *DelegatingSchedulerSharedLister) NodeInfos() scheduler_listers.NodeInfoLister {
	return lister.delegate.NodeInfos()
}

// UpdateDelegate updates the delegate
func (lister *DelegatingSchedulerSharedLister) UpdateDelegate(delegate scheduler_listers.SharedLister) {
	lister.delegate = delegate
}

// ResetDelegate resets delegate to
func (lister *DelegatingSchedulerSharedLister) ResetDelegate() {
	lister.delegate = unsetSharedListerSingleton
}

type unsetSharedLister struct{}
type unsetPodLister unsetSharedLister
type unsetNodeInfoLister unsetSharedLister

// List always returns an error
func (lister *unsetPodLister) List(labels.Selector) ([]*apiv1.Pod, error) {
	return nil, fmt.Errorf("lister not set in delegate")
}

// FilteredList always returns an error
func (lister *unsetPodLister) FilteredList(podFilter scheduler_listers.PodFilter, selector labels.Selector) ([]*apiv1.Pod, error) {
	return nil, fmt.Errorf("lister not set in delegate")
}

// List always returns an error
func (lister *unsetNodeInfoLister) List() ([]*scheduler_nodeinfo.NodeInfo, error) {
	return nil, fmt.Errorf("lister not set in delegate")
}

// HavePodsWithAffinityList always returns an error
func (lister *unsetNodeInfoLister) HavePodsWithAffinityList() ([]*scheduler_nodeinfo.NodeInfo, error) {
	return nil, fmt.Errorf("lister not set in delegate")
}

// Get always returns an error
func (lister *unsetNodeInfoLister) Get(nodeName string) (*scheduler_nodeinfo.NodeInfo, error) {
	return nil, fmt.Errorf("lister not set in delegate")
}

// Pods returns a fake PodLister which always returns an error
func (lister *unsetSharedLister) Pods() scheduler_listers.PodLister {
	return (*unsetPodLister)(lister)
}

// Pods returns a fake NodeInfoLister which always returns an error
func (lister *unsetSharedLister) NodeInfos() scheduler_listers.NodeInfoLister {
	return (*unsetNodeInfoLister)(lister)
}

var unsetSharedListerSingleton *unsetSharedLister
