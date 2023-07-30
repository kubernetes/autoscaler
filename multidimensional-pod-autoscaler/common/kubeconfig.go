/*
Copyright 2020 The Kubernetes Authors.

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

package common

import (
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

// CreateKubeConfigOrDie builds and returns a kubeconfig from file or in-cluster configuration.
func CreateKubeConfigOrDie(kubeconfig string, kubeApiQps float32, kubeApiBurst int) *rest.Config {
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

	config.QPS = kubeApiQps
	config.Burst = kubeApiBurst

	return config
}
