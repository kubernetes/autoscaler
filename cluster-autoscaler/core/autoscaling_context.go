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

package core

import (
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/builder"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate/utils"
	"k8s.io/autoscaler/cluster-autoscaler/expander"
	"k8s.io/autoscaler/cluster-autoscaler/expander/factory"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	kube_client "k8s.io/client-go/kubernetes"
	kube_record "k8s.io/client-go/tools/record"
)

// AutoscalingContext contains user-configurable constant and configuration-related objects passed to
// scale up/scale down functions.
type AutoscalingContext struct {
	// Options to customize how autoscaling works
	AutoscalingOptions
	// CloudProvider used in CA.
	CloudProvider cloudprovider.CloudProvider
	// ClientSet interface.
	ClientSet kube_client.Interface
	// ClusterState for maintaining the state of custer nodes.
	ClusterStateRegistry *clusterstate.ClusterStateRegistry
	// Recorder for recording events.
	Recorder kube_record.EventRecorder
	// PredicateChecker to check if a pod can fit into a node.
	PredicateChecker *simulator.PredicateChecker
	// ExpanderStrategy is the strategy used to choose which node group to expand when scaling up
	ExpanderStrategy expander.Strategy
	// LogRecorder can be used to collect log messages to expose via Events on some central object.
	LogRecorder *utils.LogEventRecorder
}

// AutoscalingOptions contain various options to customize how autoscaling works
type AutoscalingOptions struct {
	// MaxEmptyBulkDelete is a number of empty nodes that can be removed at the same time.
	MaxEmptyBulkDelete int
	// ScaleDownUtilizationThreshold sets threshold for nodes to be considered for scale down.
	// Well-utilized nodes are not touched.
	ScaleDownUtilizationThreshold float64
	// ScaleDownUnneededTime sets the duration CA expects a node to be unneeded/eligible for removal
	// before scaling down the node.
	ScaleDownUnneededTime time.Duration
	// ScaleDownUnreadyTime represents how long an unready node should be unneeded before it is eligible for scale down
	ScaleDownUnreadyTime time.Duration
	// MaxNodesTotal sets the maximum number of nodes in the whole cluster
	MaxNodesTotal int
	// MaxCoresTotal sets the maximum number of cores in the whole cluster
	MaxCoresTotal int64
	// MinCoresTotal sets the minimum number of cores in the whole cluster
	MinCoresTotal int64
	// MaxMemoryTotal sets the maximum memory (in megabytes) in the whole cluster
	MaxMemoryTotal int64
	// MinMemoryTotal sets the maximum memory (in megabytes) in the whole cluster
	MinMemoryTotal int64
	// NodeGroupAutoDiscovery represents one or more definition(s) of node group auto-discovery
	NodeGroupAutoDiscovery []string
	// EstimatorName is the estimator used to estimate the number of needed nodes in scale up.
	EstimatorName string
	// ExpanderName sets the type of node group expander to be used in scale up
	ExpanderName string
	// MaxGracefulTerminationSec is maximum number of seconds scale down waits for pods to terminate before
	// removing the node from cloud provider.
	MaxGracefulTerminationSec int
	//  Maximum time CA waits for node to be provisioned
	MaxNodeProvisionTime time.Duration
	// MaxTotalUnreadyPercentage is the maximum percentage of unready nodes after which CA halts operations
	MaxTotalUnreadyPercentage float64
	// OkTotalUnreadyCount is the number of allowed unready nodes, irrespective of max-total-unready-percentage
	OkTotalUnreadyCount int
	// CloudConfig is the path to the cloud provider configuration file. Empty string for no configuration file.
	CloudConfig string
	// CloudProviderName sets the type of the cloud provider CA is about to run in. Allowed values: gce, aws
	CloudProviderName string
	// NodeGroups is the list of node groups a.k.a autoscaling targets
	NodeGroups []string
	// ScaleDownEnabled is used to allow CA to scale down the cluster
	ScaleDownEnabled bool
	// ScaleDownDelayAfterAdd sets the duration from the last scale up to the time when CA starts to check scale down options
	ScaleDownDelayAfterAdd time.Duration
	// ScaleDownDelayAfterDelete sets the duration between scale down attempts if scale down removes one or more nodes
	ScaleDownDelayAfterDelete time.Duration
	// ScaleDownDelayAfterFailure sets the duration before the next scale down attempt if scale down results in an error
	ScaleDownDelayAfterFailure time.Duration
	// ScaleDownNonEmptyCandidatesCount is the maximum number of non empty nodes
	// considered at once as candidates for scale down.
	ScaleDownNonEmptyCandidatesCount int
	// ScaleDownCandidatesPoolRatio is a ratio of nodes that are considered
	// as additional non empty candidates for scale down when some candidates from
	// previous iteration are no longer valid.
	ScaleDownCandidatesPoolRatio float64
	// ScaleDownCandidatesPoolMinCount is the minimum number of nodes that are
	// considered as additional non empty candidates for scale down when some
	// candidates from previous iteration are no longer valid.
	// The formula to calculate additional candidates number is following:
	// max(#nodes * ScaleDownCandidatesPoolRatio, ScaleDownCandidatesPoolMinCount)
	ScaleDownCandidatesPoolMinCount int
	// WriteStatusConfigMap tells if the status information should be written to a ConfigMap
	WriteStatusConfigMap bool
	// BalanceSimilarNodeGroups enables logic that identifies node groups with similar machines and tries to balance node count between them.
	BalanceSimilarNodeGroups bool
	// ConfigNamespace is the namespace cluster-autoscaler is running in and all related configmaps live in
	ConfigNamespace string
	// ClusterName if available
	ClusterName string
	// NodeAutoprovisioningEnabled tells whether the node auto-provisioning is enabled for this cluster.
	NodeAutoprovisioningEnabled bool
	// MaxAutoprovisionedNodeGroupCount is the maximum number of autoprovisioned groups in the cluster.
	MaxAutoprovisionedNodeGroupCount int
	// Pods with priority below cutoff are expendable. They can be killed without any consideration during scale down and they don't cause scale up.
	// Pods with null priority (PodPriority disabled) are non expendable.
	ExpendablePodsPriorityCutoff int
}

// NewAutoscalingContext returns an autoscaling context from all the necessary parameters passed via arguments
func NewAutoscalingContext(options AutoscalingOptions, predicateChecker *simulator.PredicateChecker,
	kubeClient kube_client.Interface, kubeEventRecorder kube_record.EventRecorder,
	logEventRecorder *utils.LogEventRecorder, listerRegistry kube_util.ListerRegistry) (*AutoscalingContext, errors.AutoscalerError) {

	cloudProviderBuilder := builder.NewCloudProviderBuilder(options.CloudProviderName, options.CloudConfig, options.ClusterName, options.NodeAutoprovisioningEnabled)
	cloudProvider := cloudProviderBuilder.Build(cloudprovider.NodeGroupDiscoveryOptions{
		NodeGroupSpecs:              options.NodeGroups,
		NodeGroupAutoDiscoverySpecs: options.NodeGroupAutoDiscovery,
		KubeClient:                  kubeClient,
	},
		cloudprovider.NewResourceLimiter(
			map[string]int64{cloudprovider.ResourceNameCores: int64(options.MinCoresTotal), cloudprovider.ResourceNameMemory: options.MinMemoryTotal},
			map[string]int64{cloudprovider.ResourceNameCores: options.MaxCoresTotal, cloudprovider.ResourceNameMemory: options.MaxMemoryTotal}))

	expanderStrategy, err := factory.ExpanderStrategyFromString(options.ExpanderName,
		cloudProvider, listerRegistry.AllNodeLister())
	if err != nil {
		return nil, err
	}

	clusterStateConfig := clusterstate.ClusterStateRegistryConfig{
		MaxTotalUnreadyPercentage: options.MaxTotalUnreadyPercentage,
		OkTotalUnreadyCount:       options.OkTotalUnreadyCount,
		MaxNodeProvisionTime:      options.MaxNodeProvisionTime,
	}
	clusterStateRegistry := clusterstate.NewClusterStateRegistry(cloudProvider, clusterStateConfig, logEventRecorder)

	autoscalingContext := AutoscalingContext{
		AutoscalingOptions:   options,
		CloudProvider:        cloudProvider,
		ClusterStateRegistry: clusterStateRegistry,
		ClientSet:            kubeClient,
		Recorder:             kubeEventRecorder,
		PredicateChecker:     predicateChecker,
		ExpanderStrategy:     expanderStrategy,
		LogRecorder:          logEventRecorder,
	}

	return &autoscalingContext, nil
}
