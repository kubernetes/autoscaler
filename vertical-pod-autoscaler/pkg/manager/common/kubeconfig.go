package common

import (
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

// CreateKubeConfigOrDie builds and returns a kubeconfig from file or in-cluster configuration.
func CreateKubeConfigOrDie(kubeconfig string) *rest.Config {
	var config *rest.Config
	var err error
	if len(kubeconfig) > 0 {
		klog.V(1).Infof("Using kubeconfig file: %s", kubeconfig)
		// use the current context in kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			klog.Fatalf("Failed to build kubeconfig from file: %v", err)
		}
	} else {
		config, err = rest.InClusterConfig()
		if err != nil {
			klog.Fatalf("Failed to create config: %v", err)
		}
	}
	return config
}
