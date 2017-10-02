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
	"os"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/gce"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/kubemark"
	"k8s.io/client-go/informers"
	kubeclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	kubemarkcontroller "k8s.io/kubernetes/pkg/kubemark"

	"github.com/golang/glog"
)

// CloudProviderBuilder builds a cloud provider from all the necessary parameters including the name of a cloud provider e.g. aws, gce
// and the path to a config file
type CloudProviderBuilder struct {
	cloudProviderFlag string
	cloudConfig       string
	clusterName       string
}

// NewCloudProviderBuilder builds a new builder from static settings
func NewCloudProviderBuilder(cloudProviderFlag string, cloudConfig string, clusterName string) CloudProviderBuilder {
	return CloudProviderBuilder{
		cloudProviderFlag: cloudProviderFlag,
		cloudConfig:       cloudConfig,
		clusterName:       clusterName,
	}
}

// Build a cloud provider from static settings contained in the builder and dynamic settings passed via args
func (b CloudProviderBuilder) Build(discoveryOpts cloudprovider.NodeGroupDiscoveryOptions) cloudprovider.CloudProvider {
	var err error
	var cloudProvider cloudprovider.CloudProvider

	nodeGroupsFlag := discoveryOpts.NodeGroupSpecs

	if b.cloudProviderFlag == "gce" || b.cloudProviderFlag == "gke" {
		// GCE Manager
		var gceManager gce.GceManager
		var gceError error
		mode := gce.ModeGCE
		if b.cloudProviderFlag == "gke" {
			mode = gce.ModeGKE
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

	if b.cloudProviderFlag == kubemark.ProviderName {
		glog.V(1).Infof("Building kubemark cloud provider.")
		externalConfig, err := rest.InClusterConfig()
		if err != nil {
			glog.Fatalf("Failed to get kubeclient config for external cluster: %v", err)
		}

		kubemarkConfig, err := clientcmd.BuildConfigFromFlags("", "/kubeconfig/cluster_autoscaler.kubeconfig")
		if err != nil {
			glog.Fatalf("Failed to get kubeclient config for kubemark cluster: %v", err)
		}

		stop := make(chan struct{})

		externalClient := kubeclient.NewForConfigOrDie(externalConfig)
		kubemarkClient := kubeclient.NewForConfigOrDie(kubemarkConfig)

		externalInformerFactory := informers.NewSharedInformerFactory(externalClient, 0)
		kubemarkInformerFactory := informers.NewSharedInformerFactory(kubemarkClient, 0)
		kubemarkNodeInformer := kubemarkInformerFactory.Core().V1().Nodes()
		go kubemarkNodeInformer.Informer().Run(stop)

		kubemarkController, err := kubemarkcontroller.NewKubemarkController(externalClient, externalInformerFactory,
			kubemarkClient, kubemarkNodeInformer)
		if err != nil {
			glog.Fatalf("Failed to create Kubemark cloud provider: %v", err)
		}

		externalInformerFactory.Start(stop)
		if !kubemarkController.WaitForCacheSync(stop) {
			glog.Fatalf("Failed to sync caches for kubemark controller")
		}
		go kubemarkController.Run(stop)

		cloudProvider, err = kubemark.BuildKubemarkCloudProvider(kubemarkController, nodeGroupsFlag)
		if err != nil {
			glog.Fatalf("Failed to create Kubemark cloud provider: %v", err)
		}
	}

	return cloudProvider
}
