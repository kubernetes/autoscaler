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

package framework

import (
	"fmt"

	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

// DelegatingSchedulerSharedLister is an implementation of scheduler.SharedLister which
// passes logic to delegate. Delegate can be updated.
type DelegatingSchedulerSharedLister struct {
	delegate schedulerframework.SharedLister
}

// NewDelegatingSchedulerSharedLister creates new NewDelegatingSchedulerSharedLister
func NewDelegatingSchedulerSharedLister() *DelegatingSchedulerSharedLister {
	return &DelegatingSchedulerSharedLister{
		delegate: unsetSharedListerSingleton,
	}
}

// NodeInfos returns a NodeInfoLister.
func (lister *DelegatingSchedulerSharedLister) NodeInfos() schedulerframework.NodeInfoLister {
	return lister.delegate.NodeInfos()
}

// StorageInfos returns a StorageInfoLister
func (lister *DelegatingSchedulerSharedLister) StorageInfos() schedulerframework.StorageInfoLister {
	return lister.delegate.StorageInfos()
}

// UpdateDelegate updates the delegate
func (lister *DelegatingSchedulerSharedLister) UpdateDelegate(delegate schedulerframework.SharedLister) {
	lister.delegate = delegate
}

// ResetDelegate resets delegate to
func (lister *DelegatingSchedulerSharedLister) ResetDelegate() {
	lister.delegate = unsetSharedListerSingleton
}

type unsetSharedLister struct{}
type unsetNodeInfoLister unsetSharedLister
type unsetStorageInfoLister unsetSharedLister

// List always returns an error
func (lister *unsetNodeInfoLister) List() ([]*schedulerframework.NodeInfo, error) {
	return nil, fmt.Errorf("lister not set in delegate")
}

// HavePodsWithAffinityList always returns an error
func (lister *unsetNodeInfoLister) HavePodsWithAffinityList() ([]*schedulerframework.NodeInfo, error) {
	return nil, fmt.Errorf("lister not set in delegate")
}

// HavePodsWithRequiredAntiAffinityList always returns an error.
func (lister *unsetNodeInfoLister) HavePodsWithRequiredAntiAffinityList() ([]*schedulerframework.NodeInfo, error) {
	return nil, fmt.Errorf("lister not set in delegate")
}

// Get always returns an error
func (lister *unsetNodeInfoLister) Get(nodeName string) (*schedulerframework.NodeInfo, error) {
	return nil, fmt.Errorf("lister not set in delegate")
}

func (lister *unsetStorageInfoLister) IsPVCUsedByPods(key string) bool {
	return false
}

// NodeInfos: Pods returns a fake NodeInfoLister which always returns an error
func (lister *unsetSharedLister) NodeInfos() schedulerframework.NodeInfoLister {
	return (*unsetNodeInfoLister)(lister)
}

// StorageInfos: Pods returns a fake StorageInfoLister which always returns an error
func (lister *unsetSharedLister) StorageInfos() schedulerframework.StorageInfoLister {
	return (*unsetStorageInfoLister)(lister)
}

var unsetSharedListerSingleton *unsetSharedLister
