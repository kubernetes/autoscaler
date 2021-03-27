/*
Copyright 2018 The Kubernetes Authors.

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

package processors

import (
	"k8s.io/autoscaler/cluster-autoscaler/processors/customresources"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroupconfig"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroups"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroupset"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodeinfos"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodes"
	"k8s.io/autoscaler/cluster-autoscaler/processors/pods"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
)

// AutoscalingProcessors are a set of customizable processors used for encapsulating
// various heuristics used in different parts of Cluster Autoscaler code.
type AutoscalingProcessors struct {
	// PodListProcessor is used to process list of unschedulable pods before autoscaling.
	PodListProcessor pods.PodListProcessor
	// NodeGroupListProcessor is used to process list of NodeGroups that can be used in scale-up.
	NodeGroupListProcessor nodegroups.NodeGroupListProcessor
	// NodeGroupSetProcessor is used to divide scale-up between similar NodeGroups.
	NodeGroupSetProcessor nodegroupset.NodeGroupSetProcessor
	// ScaleUpStatusProcessor is used to process the state of the cluster after a scale-up.
	ScaleUpStatusProcessor status.ScaleUpStatusProcessor
	// ScaleDownNodeProcessor is used to process the nodes of the cluster before scale-down.
	ScaleDownNodeProcessor nodes.ScaleDownNodeProcessor
	// ScaleDownStatusProcessor is used to process the state of the cluster after a scale-down.
	ScaleDownStatusProcessor status.ScaleDownStatusProcessor
	// AutoscalingStatusProcessor is used to process the state of the cluster after each autoscaling iteration.
	AutoscalingStatusProcessor status.AutoscalingStatusProcessor
	// NodeGroupManager is responsible for creating/deleting node groups.
	NodeGroupManager nodegroups.NodeGroupManager
	// NodeInfoProcessor is used to process nodeInfos after they're created.
	NodeInfoProcessor nodeinfos.NodeInfoProcessor
	// NodeGroupConfigProcessor provides config option for each NodeGroup.
	NodeGroupConfigProcessor nodegroupconfig.NodeGroupConfigProcessor
	// CustomResourcesProcessor is interface defining handling custom resources
	CustomResourcesProcessor customresources.CustomResourcesProcessor
}

// DefaultProcessors returns default set of processors.
func DefaultProcessors() *AutoscalingProcessors {
	return &AutoscalingProcessors{
		PodListProcessor:           pods.NewDefaultPodListProcessor(),
		NodeGroupListProcessor:     nodegroups.NewDefaultNodeGroupListProcessor(),
		NodeGroupSetProcessor:      nodegroupset.NewDefaultNodeGroupSetProcessor([]string{}),
		ScaleUpStatusProcessor:     status.NewDefaultScaleUpStatusProcessor(),
		ScaleDownNodeProcessor:     nodes.NewPreFilteringScaleDownNodeProcessor(),
		ScaleDownStatusProcessor:   status.NewDefaultScaleDownStatusProcessor(),
		AutoscalingStatusProcessor: status.NewDefaultAutoscalingStatusProcessor(),
		NodeGroupManager:           nodegroups.NewDefaultNodeGroupManager(),
		NodeInfoProcessor:          nodeinfos.NewDefaultNodeInfoProcessor(),
		NodeGroupConfigProcessor:   nodegroupconfig.NewDefaultNodeGroupConfigProcessor(),
		CustomResourcesProcessor:   customresources.NewDefaultCustomResourcesProcessor(),
	}
}

// CleanUp cleans up the processors' internal structures.
func (ap *AutoscalingProcessors) CleanUp() {
	ap.PodListProcessor.CleanUp()
	ap.NodeGroupListProcessor.CleanUp()
	ap.NodeGroupSetProcessor.CleanUp()
	ap.ScaleUpStatusProcessor.CleanUp()
	ap.ScaleDownStatusProcessor.CleanUp()
	ap.AutoscalingStatusProcessor.CleanUp()
	ap.NodeGroupManager.CleanUp()
	ap.ScaleDownNodeProcessor.CleanUp()
	ap.NodeInfoProcessor.CleanUp()
	ap.NodeGroupConfigProcessor.CleanUp()
	ap.CustomResourcesProcessor.CleanUp()
}
