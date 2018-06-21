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
	"k8s.io/autoscaler/vertical-pod-autoscaler/common"
	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	updater "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/updater/logic"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
	kube_client "k8s.io/client-go/kubernetes"
	kube_restclient "k8s.io/client-go/rest"
	"time"
)

var (
	updaterInterval = flag.Duration("updater-interval", 1*time.Minute,
		`How often updater should run`)

	minReplicas = flag.Int("min-replicas", 2,
		`Minimum number of replicas to perform update`)

	evictionToleranceFraction = flag.Float64("eviction-tolerance", 0.5,
		`Fraction of replica count that can be evicted for update, if more than one pod can be evicted.`)
)

func main() {
	kube_flag.InitFlags()
	glog.V(1).Infof("Vertical Pod Autoscaler %s Updater", common.VerticalPodAutoscalerVersion)

	// TODO monitoring

	kubeClient, vpaClient := createKubeClients()
	updater := updater.NewUpdater(kubeClient, vpaClient, *minReplicas, *evictionToleranceFraction, vpa_api_util.NewCappingRecommendationProcessor(), nil)
	for {
		select {
		case <-time.After(*updaterInterval):
			{
				updater.RunOnce()
			}
		}
	}
}

func createKubeClients() (kube_client.Interface, *vpa_clientset.Clientset) {
	config, err := kube_restclient.InClusterConfig()
	if err != nil {
		glog.Fatalf("Failed to build Kubernetes client : fail to create config: %v", err)
	}
	return kube_client.NewForConfigOrDie(config), vpa_clientset.NewForConfigOrDie(config)
}
