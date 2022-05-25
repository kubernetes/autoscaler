package common

import (
	"math/rand"
	"time"

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

func RandString(len int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		b := r.Intn(26) + 65
		bytes[i] = byte(b)
	}
	return string(bytes)
}
