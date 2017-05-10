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

package expander

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	apiv1 "k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/plugin/pkg/scheduler/schedulercache"
)

var (
	// AvailableExpanders is a list of avaialble expander options
	AvailableExpanders = []string{NodeCountBalanceExpanderName, RandomExpanderName, MostPodsExpanderName, LeastWasteExpanderName}
	// NodeCountBalanceExpanderName is the name of expander which selects a node group to balance node groups in their node counts
	NodeCountBalanceExpanderName = "node-count-balance"
	// RandomExpanderName is the name of expander which selects a node group at random
	RandomExpanderName = "random"
	// MostPodsExpanderName is the name of expander which selects a node group that fits the most pods
	MostPodsExpanderName = "most-pods"
	// LeastWasteExpanderName is the name of expander which selects a node group that leaves the least fraction of CPU and Memory
	LeastWasteExpanderName = "least-waste"
)

// Option describes an option to expand the cluster.
type Option struct {
	NodeGroup cloudprovider.NodeGroup
	NodeCount int
	Debug     string
	Pods      []*apiv1.Pod
}

// Strategy describes an interface for selecting the best option when scaling up
type Strategy interface {
	BestOption(options []Option, nodeInfo map[string]*schedulercache.NodeInfo) *Option
}
