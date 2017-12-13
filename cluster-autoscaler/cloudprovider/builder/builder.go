/*
Copyright 2017 The Kubernetes Authors.

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

package builder

// CloudProviderBuilder builds a cloud provider from all the necessary parameters including the name of a cloud provider e.g. aws, gce
// and the path to a config file
type CloudProviderBuilder struct {
	cloudProviderFlag       string
	cloudConfig             string
	clusterName             string
	autoprovisioningEnabled bool
}

// NewCloudProviderBuilder builds a new builder from static settings
func NewCloudProviderBuilder(cloudProviderFlag string, cloudConfig string, clusterName string, autoprovisioningEnabled bool) CloudProviderBuilder {
	return CloudProviderBuilder{
		cloudProviderFlag:       cloudProviderFlag,
		cloudConfig:             cloudConfig,
		clusterName:             clusterName,
		autoprovisioningEnabled: autoprovisioningEnabled,
	}
}
