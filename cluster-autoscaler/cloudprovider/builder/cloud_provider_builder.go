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

package builder

import (
	"github.com/golang/glog"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/gce"
	// Placeholder
	_ "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/kubemark"
	"os"
)

// CloudProviderBuilder builds a cloud provider from all the necessary parameters including the name of a cloud provider e.g. aws, gce
// and the path to a config file
type CloudProviderBuilder struct {
	cloudProviderFlag string
	cloudConfig       string
}

// NewCloudProviderBuilder builds a new builder from static settings
func NewCloudProviderBuilder(cloudProviderFlag string, cloudConfig string) CloudProviderBuilder {
	return CloudProviderBuilder{
		cloudProviderFlag: cloudProviderFlag,
		cloudConfig:       cloudConfig,
	}
}

// Build a cloud provider from static settings contained in the builder and dynamic settings passed via args
func (b CloudProviderBuilder) Build(discoveryOpts cloudprovider.NodeGroupDiscoveryOptions) cloudprovider.CloudProvider {
	var err error
	var cloudProvider cloudprovider.CloudProvider

	nodeGroupsFlag := discoveryOpts.NodeGroupSpecs

	if b.cloudProviderFlag == "gce" {
		// GCE Manager
		var gceManager *gce.GceManager
		var gceError error
		if b.cloudConfig != "" {
			config, fileErr := os.Open(b.cloudConfig)
			if fileErr != nil {
				glog.Fatalf("Couldn't open cloud provider configuration %s: %#v", b.cloudConfig, err)
			}
			defer config.Close()
			gceManager, gceError = gce.CreateGceManager(config)
		} else {
			gceManager, gceError = gce.CreateGceManager(nil)
		}
		if gceError != nil {
			glog.Fatalf("Failed to create GCE Manager: %v", err)
		}
		cloudProvider, err = gce.BuildGceCloudProvider(gceManager, nodeGroupsFlag)
		if err != nil {
			glog.Fatalf("Failed to create GCE cloud provider: %v", err)
		}
	}

	if b.cloudProviderFlag == "aws" {
		var awsManager *aws.AwsManager
		var awsError error
		if b.cloudConfig != "" {
			config, fileErr := os.Open(b.cloudConfig)
			if fileErr != nil {
				glog.Fatalf("Couldn't open cloud provider configuration %s: %#v", b.cloudConfig, err)
			}
			defer config.Close()
			awsManager, awsError = aws.CreateAwsManager(config)
		} else {
			awsManager, awsError = aws.CreateAwsManager(nil)
		}
		if awsError != nil {
			glog.Fatalf("Failed to create AWS Manager: %v", err)
		}
		cloudProvider, err = aws.BuildAwsCloudProvider(awsManager, discoveryOpts)
		if err != nil {
			glog.Fatalf("Failed to create AWS cloud provider: %v", err)
		}
	}
	return cloudProvider
}
