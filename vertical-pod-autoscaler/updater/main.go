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

package main

import (
	"flag"
	"github.com/golang/glog"
	kube_flag "k8s.io/apiserver/pkg/util/flag"
	kube_restclient "k8s.io/client-go/rest"
	kube_client "k8s.io/kubernetes/pkg/client/clientset_generated/clientset"
	"time"
)

var (
	updaterInterval = flag.Duration("updater-interval", 1*time.Minute,
		`How often updater should run`)

	recommendationsCacheTtl = flag.Duration("recommendation-cache-ttl", 2*time.Minute,
		`TTL for cached VPA recommendations`)

	minReplicas = flag.Int("min-replicas", 2,
		`Minimum number of replicas to perform update`)

	evictionToleranceFraction = flag.Float64("eviction-tolerance", 0.5,
		`Fraction of replica count that can be evicted for update, if more than one pod can be evicted.`)
)

func main() {
	glog.Infof("Running VPA Updater")
	kube_flag.InitFlags()

	// TODO monitoring

	kubeClient := createKubeClient()
	updater := NewUpdater(kubeClient, *recommendationsCacheTtl, *minReplicas, *evictionToleranceFraction)
	for {
		select {
		case <-time.After(*updaterInterval):
			{
				updater.RunOnce()
			}
		}
	}
}

func createKubeClient() kube_client.Interface {
	config, err := kube_restclient.InClusterConfig()
	if err != nil {
		glog.Fatalf("Failed to build Kuberentes client : fail to create config: %v", err)
	}
	return kube_client.NewForConfigOrDie(config)
}
