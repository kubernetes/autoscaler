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
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

var (
	// AvailableExpanders is a list of available expander options
	AvailableExpanders = []string{RandomExpanderName, MostPodsExpanderName, LeastWasteExpanderName, PriceBasedExpanderName, PriorityBasedExpanderName}
	// RandomExpanderName selects a node group at random
	RandomExpanderName = "random"
	// MostPodsExpanderName selects a node group that fits the most pods
	MostPodsExpanderName = "most-pods"
	// LeastWasteExpanderName selects a node group that leaves the least fraction of CPU and Memory
	LeastWasteExpanderName = "least-waste"
	// PriceBasedExpanderName selects a node group that is the most cost-effective and consistent with
	// the preferred node size for the cluster
	PriceBasedExpanderName = "price"
	// PriorityBasedExpanderName selects a node group based on a user-configured priorities assigned to group names
	PriorityBasedExpanderName = "priority"
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
	BestOption(options []Option, nodeInfo map[string]*schedulerframework.NodeInfo) *Option
}
