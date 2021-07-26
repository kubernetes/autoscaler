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

package schema

// LoadBalancerType defines the schema of a LoadBalancer type.
type LoadBalancerType struct {
	ID                      int                            `json:"id"`
	Name                    string                         `json:"name"`
	Description             string                         `json:"description"`
	MaxConnections          int                            `json:"max_connections"`
	MaxServices             int                            `json:"max_services"`
	MaxTargets              int                            `json:"max_targets"`
	MaxAssignedCertificates int                            `json:"max_assigned_certificates"`
	Prices                  []PricingLoadBalancerTypePrice `json:"prices"`
}

// LoadBalancerTypeListResponse defines the schema of the response when
// listing LoadBalancer types.
type LoadBalancerTypeListResponse struct {
	LoadBalancerTypes []LoadBalancerType `json:"load_balancer_types"`
}

// LoadBalancerTypeGetResponse defines the schema of the response when
// retrieving a single LoadBalancer type.
type LoadBalancerTypeGetResponse struct {
	LoadBalancerType LoadBalancerType `json:"load_balancer_type"`
}
