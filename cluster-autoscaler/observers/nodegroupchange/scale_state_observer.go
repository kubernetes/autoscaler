/*
Copyright 2023 The Kubernetes Authors.

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

package nodegroupchange

import (
	"sync"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

// NodeGroupChangeObserver is an observer of:
// * scale-up(s) for a nodegroup
// * scale-down(s) for a nodegroup
// * scale-up failure(s) for a nodegroup
// * scale-down failure(s) for a nodegroup
type NodeGroupChangeObserver interface {
	// RegisterScaleUp records scale up for a nodegroup.
	RegisterScaleUp(nodeGroup cloudprovider.NodeGroup, delta int, currentTime time.Time)
	// RegisterScaleDowns records scale down for a nodegroup.
	RegisterScaleDown(nodeGroup cloudprovider.NodeGroup, nodeName string, currentTime time.Time, expectedDeleteTime time.Time)
	// RegisterFailedScaleUp records failed scale-up for a nodegroup.
	// reason denotes optional reason for failed scale-up
	// errMsg denotes the actual error message
	RegisterFailedScaleUp(nodeGroup cloudprovider.NodeGroup, reason string, errMsg string, gpuResourceName, gpuType string, currentTime time.Time)
	// RegisterFailedScaleDown records failed scale-down for a nodegroup.
	RegisterFailedScaleDown(nodeGroup cloudprovider.NodeGroup, reason string, currentTime time.Time)
}

// NodeGroupChangeObserversList is a slice of observers
// of state of scale up/down in the cluster
type NodeGroupChangeObserversList struct {
	observers []NodeGroupChangeObserver
	// TODO(vadasambar): consider using separate mutexes for functions not related to each other
	mutex sync.Mutex
}

// Register adds new observer to the list.
func (l *NodeGroupChangeObserversList) Register(o NodeGroupChangeObserver) {
	l.observers = append(l.observers, o)
}

// RegisterScaleUp calls RegisterScaleUp for each observer.
func (l *NodeGroupChangeObserversList) RegisterScaleUp(nodeGroup cloudprovider.NodeGroup,
	delta int, currentTime time.Time) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	for _, observer := range l.observers {
		observer.RegisterScaleUp(nodeGroup, delta, currentTime)
	}
}

// RegisterScaleDown calls RegisterScaleDown for each observer.
func (l *NodeGroupChangeObserversList) RegisterScaleDown(nodeGroup cloudprovider.NodeGroup,
	nodeName string, currentTime time.Time, expectedDeleteTime time.Time) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	for _, observer := range l.observers {
		observer.RegisterScaleDown(nodeGroup, nodeName, currentTime, expectedDeleteTime)
	}
}

// RegisterFailedScaleUp calls RegisterFailedScaleUp for each observer.
func (l *NodeGroupChangeObserversList) RegisterFailedScaleUp(nodeGroup cloudprovider.NodeGroup,
	reason string, errMsg, gpuResourceName, gpuType string, currentTime time.Time) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	for _, observer := range l.observers {
		observer.RegisterFailedScaleUp(nodeGroup, reason, errMsg, gpuResourceName, gpuType, currentTime)
	}
}

// RegisterFailedScaleDown records failed scale-down for a nodegroup.
func (l *NodeGroupChangeObserversList) RegisterFailedScaleDown(nodeGroup cloudprovider.NodeGroup,
	reason string, currentTime time.Time) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	for _, observer := range l.observers {
		observer.RegisterFailedScaleDown(nodeGroup, reason, currentTime)
	}
}

// NewNodeGroupChangeObserversList return empty list of scale state observers.
func NewNodeGroupChangeObserversList() *NodeGroupChangeObserversList {
	return &NodeGroupChangeObserversList{}
}
