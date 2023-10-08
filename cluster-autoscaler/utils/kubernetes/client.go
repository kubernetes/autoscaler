/*
Copyright 2023 The Kubernetes Authors.

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

package kubernetes

import (
	"net/url"

	"k8s.io/autoscaler/cluster-autoscaler/config"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

// GetKubeConfig returns the rest config from AutoscalingOptions.
func GetKubeConfig(opts config.AutoscalingOptions) *rest.Config {
	var kubeConfig *rest.Config
	var err error

	if opts.KubeConfigPath != "" {
		klog.V(1).Infof("Using kubeconfig file: %s", opts.KubeConfigPath)
		// use the current context in kubeconfig
		kubeConfig, err = clientcmd.BuildConfigFromFlags("", opts.KubeConfigPath)
		if err != nil {
			klog.Fatalf("Failed to build config: %v", err)
		}
	} else {
		url, err := url.Parse(opts.Kubernetes)
		if err != nil {
			klog.Fatalf("Failed to parse Kubernetes url: %v", err)
		}

		kubeConfig, err = config.GetKubeClientConfig(url)
		if err != nil {
			klog.Fatalf("Failed to build Kubernetes client configuration: %v", err)
		}
	}

	kubeConfig.ContentType = opts.KubeAPIContentType

	return kubeConfig
}
