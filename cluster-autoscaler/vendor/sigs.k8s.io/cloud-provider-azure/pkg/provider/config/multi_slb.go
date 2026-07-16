/*
Copyright 2024 The Kubernetes Authors.

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

package config

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	utilsets "sigs.k8s.io/cloud-provider-azure/pkg/util/sets"
)

// MultipleStandardLoadBalancerConfiguration stores the properties regarding multiple standard load balancers.
type MultipleStandardLoadBalancerConfiguration struct {
	// Name of the public load balancer. There will be an internal load balancer
	// created if needed, and the name will be `<name>-internal`. The internal lb
	// shares the same configurations as the external one. The internal lbs
	// are not needed to be included in `MultipleStandardLoadBalancerConfigurations`.
	// There must be a name of "<clustername>" in the load balancer configuration list.
	Name string `json:"name" yaml:"name"`

	MultipleStandardLoadBalancerConfigurationSpec

	MultipleStandardLoadBalancerConfigurationStatus
}

// MultipleStandardLoadBalancerConfigurationSpec stores the properties regarding multiple standard load balancers.
type MultipleStandardLoadBalancerConfigurationSpec struct {
	// This load balancer can have services placed on it. Defaults to true,
	// can be set to false to drain and eventually remove a load balancer.
	// This only affects services that will be using the LB. For services
	// that is currently using the LB, they will not be affected.
	AllowServicePlacement *bool `json:"allowServicePlacement" yaml:"allowServicePlacement"`

	// A string value that must specify the name of an existing vmSet.
	// All nodes in the given vmSet will always be added to this load balancer.
	// A vmSet can only be the primary vmSet for a single load balancer.
	PrimaryVMSet string `json:"primaryVMSet" yaml:"primaryVMSet"`

	// Services that must match this selector can be placed on this load balancer. If not supplied,
	// services with any labels can be created on the load balancer.
	// A ServiceLabelSelector with empty matchLabels and matchExpressions will match all services, but
	// only works if no non-empty ServiceLabelSelector has matched the service.
	ServiceLabelSelector *metav1.LabelSelector `json:"serviceLabelSelector" yaml:"serviceLabelSelector"`

	// Services created in namespaces with the supplied label will be allowed to select that load balancer.
	// If not supplied, services created in any namespaces can be created on that load balancer.
	// A ServiceNamespaceSelector with empty matchLabels and matchExpressions will match all nodes, but
	// only works if no non-empty ServiceNamespaceSelector has matched the service.
	ServiceNamespaceSelector *metav1.LabelSelector `json:"serviceNamespaceSelector" yaml:"serviceNamespaceSelector"`

	// Nodes matching this selector will be preferentially added to the load balancers that
	// they match selectors for. NodeSelector does not override primaryAgentPool for node allocation.
	// A NodeSelector with empty matchLabels and matchExpressions will match all nodes, but
	// only works if no non-empty NodeSelector has matched the node.
	NodeSelector *metav1.LabelSelector `json:"nodeSelector" yaml:"nodeSelector"`
}

// MultipleStandardLoadBalancerConfigurationStatus stores the properties regarding multiple standard load balancers.
type MultipleStandardLoadBalancerConfigurationStatus struct {
	// ActiveServices stores the services that are supposed to use the load balancer.
	ActiveServices *utilsets.IgnoreCaseSet `json:"activeServices" yaml:"activeServices"`

	// ActiveNodes stores the nodes that are supposed to be in the load balancer.
	// It will be used in EnsureHostsInPool to make sure the given ones are in the backend pool.
	ActiveNodes *utilsets.IgnoreCaseSet `json:"activeNodes" yaml:"activeNodes"`
}
