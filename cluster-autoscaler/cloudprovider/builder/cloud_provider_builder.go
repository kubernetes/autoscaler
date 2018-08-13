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
	"io"
	"os"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/azure"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/gce"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/gke"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/kubemark"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/client-go/informers"
	kubeclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	kubemarkcontroller "k8s.io/kubernetes/pkg/kubemark"

	"github.com/golang/glog"
)

// AvailableCloudProviders supported by the cloud provider builder.
var AvailableCloudProviders = []string{
	aws.ProviderName,
	azure.ProviderName,
	gce.ProviderNameGCE,
	gke.ProviderNameGKE,
	kubemark.ProviderName,
}

// DefaultCloudProvider is GCE.
const DefaultCloudProvider = gce.ProviderNameGCE

// NewCloudProvider builds a cloud provider from provided parameters.
func NewCloudProvider(opts config.AutoscalingOptions) cloudprovider.CloudProvider {
	glog.V(1).Infof("Building %s cloud provider.", opts.CloudProviderName)

	do := cloudprovider.NodeGroupDiscoveryOptions{
		NodeGroupSpecs:              opts.NodeGroups,
		NodeGroupAutoDiscoverySpecs: opts.NodeGroupAutoDiscovery,
	}
	rl := context.NewResourceLimiterFromAutoscalingOptions(opts)

	switch opts.CloudProviderName {
	case gce.ProviderNameGCE:
		return buildGCE(opts, do, rl, gce.ModeGCE)
	case gke.ProviderNameGKE:
		if do.DiscoverySpecified() {
			glog.Fatalf("GKE gets nodegroup specification via API, command line specs are not allowed")
		}
		if opts.NodeAutoprovisioningEnabled {
			return buildGKE(opts, do, rl, gke.ModeGKENAP)
		}
		return buildGKE(opts, do, rl, gke.ModeGKE)
	case aws.ProviderName:
		return buildAWS(opts, do, rl)
	case azure.ProviderName:
		return buildAzure(opts, do, rl)
	case kubemark.ProviderName:
		return buildKubemark(opts, do, rl)
	case "":
		// Ideally this would be an error, but several unit tests of the
		// StaticAutoscaler depend on this behaviour.
		glog.Warning("Returning a nil cloud provider")
		return nil
	}

	glog.Fatalf("Unknown cloud provider: %s", opts.CloudProviderName)
	return nil // This will never happen because the Fatalf will os.Exit
}

func buildGCE(opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter, mode gce.GcpCloudProviderMode) cloudprovider.CloudProvider {
	var config io.ReadCloser
	if opts.CloudConfig != "" {
		var err error
		config, err = os.Open(opts.CloudConfig)
		if err != nil {
			glog.Fatalf("Couldn't open cloud provider configuration %s: %#v", opts.CloudConfig, err)
		}
		defer config.Close()
	}

	manager, err := gce.CreateGceManager(config, mode, opts.ClusterName, do, opts.Regional)
	if err != nil {
		glog.Fatalf("Failed to create GCE Manager: %v", err)
	}

	provider, err := gce.BuildGceCloudProvider(manager, rl)
	if err != nil {
		glog.Fatalf("Failed to create GCE cloud provider: %v", err)
	}
	return provider
}

func buildGKE(opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter, mode gke.GcpCloudProviderMode) cloudprovider.CloudProvider {
	var config io.ReadCloser
	if opts.CloudConfig != "" {
		var err error
		config, err = os.Open(opts.CloudConfig)
		if err != nil {
			glog.Fatalf("Couldn't open cloud provider configuration %s: %#v", opts.CloudConfig, err)
		}
		defer config.Close()
	}

	manager, err := gke.CreateGkeManager(config, mode, opts.ClusterName, do, opts.Regional)
	if err != nil {
		glog.Fatalf("Failed to create GKE Manager: %v", err)
	}

	provider, err := gke.BuildGkeCloudProvider(manager, rl)
	if err != nil {
		glog.Fatalf("Failed to create GKE cloud provider: %v", err)
	}
	return provider
}

func buildAWS(opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
	var config io.ReadCloser
	if opts.CloudConfig != "" {
		var err error
		config, err = os.Open(opts.CloudConfig)
		if err != nil {
			glog.Fatalf("Couldn't open cloud provider configuration %s: %#v", opts.CloudConfig, err)
		}
		defer config.Close()
	}

	manager, err := aws.CreateAwsManager(config, do)
	if err != nil {
		glog.Fatalf("Failed to create AWS Manager: %v", err)
	}

	provider, err := aws.BuildAwsCloudProvider(manager, rl)
	if err != nil {
		glog.Fatalf("Failed to create AWS cloud provider: %v", err)
	}
	return provider
}

func buildAzure(opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
	var config io.ReadCloser
	if opts.CloudConfig != "" {
		glog.Info("Creating Azure Manager using cloud-config file: %v", opts.CloudConfig)
		var err error
		config, err := os.Open(opts.CloudConfig)
		if err != nil {
			glog.Fatalf("Couldn't open cloud provider configuration %s: %#v", opts.CloudConfig, err)
		}
		defer config.Close()
	} else {
		glog.Info("Creating Azure Manager with default configuration.")
	}
	manager, err := azure.CreateAzureManager(config, do)
	if err != nil {
		glog.Fatalf("Failed to create Azure Manager: %v", err)
	}
	provider, err := azure.BuildAzureCloudProvider(manager, rl)
	if err != nil {
		glog.Fatalf("Failed to create Azure cloud provider: %v", err)
	}
	return provider
}

func buildKubemark(opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
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

	provider, err := kubemark.BuildKubemarkCloudProvider(kubemarkController, do.NodeGroupSpecs, rl)
	if err != nil {
		glog.Fatalf("Failed to create Kubemark cloud provider: %v", err)
	}
	return provider
}
