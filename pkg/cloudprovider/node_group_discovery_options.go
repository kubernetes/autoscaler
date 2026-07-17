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

package cloudprovider

// NodeGroupDiscoveryOptions contains various options to configure how a cloud provider discovers node groups
type NodeGroupDiscoveryOptions struct {
	// NodeGroupSpecs is specified to statically discover node groups listed in it
	NodeGroupSpecs []string
	// NodeGroupAutoDiscoverySpec is specified for automatically discovering node groups according to the specs
	NodeGroupAutoDiscoverySpecs []string
}

// StaticDiscoverySpecified returns true only when there are 1 or more --nodes flags specified
func (o NodeGroupDiscoveryOptions) StaticDiscoverySpecified() bool {
	return len(o.NodeGroupSpecs) > 0
}

// AutoDiscoverySpecified returns true only when there are 1 or more --node-group-auto-discovery flags specified
func (o NodeGroupDiscoveryOptions) AutoDiscoverySpecified() bool {
	return len(o.NodeGroupAutoDiscoverySpecs) > 0
}

// DiscoverySpecified returns true when at least one of the --nodes or
// --node-group-auto-discovery flags specified.
func (o NodeGroupDiscoveryOptions) DiscoverySpecified() bool {
	return o.StaticDiscoverySpecified() || o.AutoDiscoverySpecified()
}
