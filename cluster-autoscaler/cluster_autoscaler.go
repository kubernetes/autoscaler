/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package main

import (
	"flag"
	"net/url"
	"time"

	"k8s.io/contrib/cluster-autoscaler/config"
	kube_api "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/cache"
	kube_client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"

	"github.com/golang/glog"
)

var (
	migConfig  config.MigConfigFlag
	kubernetes = flag.String("kubernetes", "", "Kuberentes master location. Leave blank for default")
)

func main() {
	flag.Var(&migConfig, "nodes", "sets min,max size and url of a MIG to be controlled by Cluster Autoscaler. "+
		"Can be used multiple times. Format: <min>:<max>:<migurl>")
	flag.Parse()

	glog.Infof("MIG: %s\n", migConfig.String())

	url, err := url.Parse(*kubernetes)
	if err != nil {
		glog.Fatalf("Failed to parse Kuberentes url: %v", err)
	}
	kubeConfig, err := config.GetKubeClientConfig(url)
	if err != nil {
		glog.Fatalf("Failed to build Kuberentes client configuration: %v", err)
	}

	kubeClient := kube_client.NewOrDie(kubeConfig)

	// watch unscheduled pods
	selector := fields.ParseSelectorOrDie("spec.nodeName==" + "" + ",status.phase!=" +
		string(kube_api.PodSucceeded) + ",status.phase!=" + string(kube_api.PodFailed))
	podListWatch := cache.NewListWatchFromClient(kubeClient, "pods", kube_api.NamespaceAll, selector)
	podListener := &cache.StoreToPodLister{Store: cache.NewStore(cache.MetaNamespaceKeyFunc)}
	podReflector := cache.NewReflector(podListWatch, &kube_api.Pod{}, podListener.Store, time.Hour)
	podReflector.Run()

	for {
		select {
		case <-time.After(time.Minute):
			{
				pods, err := podListener.List(labels.Everything())
				if err != nil {
					glog.Errorf("Failed to list pods: %v", err)
					break
				}

				for _, pod := range pods {
					glog.Infof("Pod %s/%s is not scheduled", pod.Namespace, pod.Name)
				}
			}
		}
	}
}
