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

package kwok

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	listersv1 "k8s.io/client-go/listers/core/v1"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
)

// KwokCloudProvider implements CloudProvider interface for kwok
type KwokCloudProvider struct {
	nodeGroups      []*NodeGroup
	config          *KwokProviderConfig
	resourceLimiter *cloudprovider.ResourceLimiter
	// kubeClient is to be used only for create, delete and update
	kubeClient kubernetes.Interface
	//allNodesLister is a lister to list all nodes in cluster
	allNodesLister listersv1.NodeLister
}

type kwokOptions struct {
	kubeClient      kubernetes.Interface
	autoscalingOpts *config.AutoscalingOptions
	discoveryOpts   *cloudprovider.NodeGroupDiscoveryOptions
	resourceLimiter *cloudprovider.ResourceLimiter
	// TODO(vadasambar): look into abstracting kubeClient
	// and lister into a single client
	// allNodeLister lists all the nodes in the cluster
	allNodesLister listersv1.NodeLister
	// nodeLister lists all nodes managed by kwok for a specific nodegroup
	ngNodeListerFn listerFn
}

// NodeGroup implements NodeGroup interface.
type NodeGroup struct {
	name         string
	kubeClient   kubernetes.Interface
	lister       kube_util.NodeLister
	nodeTemplate *apiv1.Node
	minSize      int
	targetSize   int
	maxSize      int
}

// NodegroupsConfig defines options for creating nodegroups
type NodegroupsConfig struct {
	FromNodeLabelKey      string `json:"fromNodeLabelKey" yaml:"fromNodeLabelKey"`
	FromNodeAnnotationKey string `json:"fromNodeAnnotationKey" yaml:"fromNodeAnnotationKey"`
}

// NodeConfig defines config options for the nodes
type NodeConfig struct {
	GPUConfig *GPUConfig `json:"gpuConfig" yaml:"gpuConfig"`
	SkipTaint bool       `json:"skipTaint" yaml:"skipTaint"`
}

// ConfigMapConfig allows setting the kwok provider configmap name
type ConfigMapConfig struct {
	Name string `json:"name" yaml:"name"`
	Key  string `json:"key" yaml:"key"`
}

// GPUConfig defines GPU related config for the node
type GPUConfig struct {
	GPULabelKey       string              `json:"gpuLabelKey" yaml:"gpuLabelKey"`
	AvailableGPUTypes map[string]struct{} `json:"availableGPUTypes" yaml:"availableGPUTypes"`
}

// KwokConfig is the struct to define kwok specific config
// (needs to be implemented; currently empty)
type KwokConfig struct {
}

// KwokProviderConfig is the struct to hold kwok provider config
type KwokProviderConfig struct {
	APIVersion    string            `json:"apiVersion" yaml:"apiVersion"`
	ReadNodesFrom string            `json:"readNodesFrom" yaml:"readNodesFrom"`
	Nodegroups    *NodegroupsConfig `json:"nodegroups" yaml:"nodegroups"`
	Nodes         *NodeConfig       `json:"nodes" yaml:"nodes"`
	ConfigMap     *ConfigMapConfig  `json:"configmap" yaml:"configmap"`
	Kwok          *KwokConfig       `json:"kwok" yaml:"kwok"`
	status        *GroupingConfig
}

// GroupingConfig defines different
type GroupingConfig struct {
	groupNodesBy      string              // [annotation, label]
	key               string              // annotation or label key
	gpuLabel          string              // gpu label key
	availableGPUTypes map[string]struct{} // available gpu types
}
