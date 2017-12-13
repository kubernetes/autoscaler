// +build gke

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

import (
	"os"

	"github.com/golang/glog"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/gce"
)

// Build a cloud provider from static settings contained in the builder and dynamic settings passed via args
func (b CloudProviderBuilder) Build(discoveryOpts cloudprovider.NodeGroupDiscoveryOptions, resourceLimiter *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
	var err error
	var cloudProvider cloudprovider.CloudProvider

	nodeGroupsFlag := discoveryOpts.NodeGroupSpecs

	if b.cloudProviderFlag == "gce" || b.cloudProviderFlag == "gke" {
		// GCE Manager
		var gceManager gce.GceManager
		var gceError error
		mode := gce.ModeGCE
		if b.cloudProviderFlag == "gke" {
			if b.autoprovisioningEnabled {
				mode = gce.ModeGKENAP
			} else {
				mode = gce.ModeGKE
			}
		}

		if b.cloudConfig != "" {
			config, fileErr := os.Open(b.cloudConfig)
			if fileErr != nil {
				glog.Fatalf("Couldn't open cloud provider configuration %s: %#v", b.cloudConfig, err)
			}
			defer config.Close()
			gceManager, gceError = gce.CreateGceManager(config, mode, b.clusterName)
		} else {
			gceManager, gceError = gce.CreateGceManager(nil, mode, b.clusterName)
		}
		if gceError != nil {
			glog.Fatalf("Failed to create GCE Manager: %v", gceError)
		}
		cloudProvider, err = gce.BuildGceCloudProvider(gceManager, nodeGroupsFlag, resourceLimiter)
		if err != nil {
			glog.Fatalf("Failed to create GCE cloud provider: %v", err)
		}
	}
	return cloudProvider
}
