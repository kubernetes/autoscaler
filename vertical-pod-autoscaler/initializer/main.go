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
	"os"
	"os/signal"
	"syscall"
	"time"

	"k8s.io/autoscaler/vertical-pod-autoscaler/initializer/core"

	"github.com/golang/glog"
	kube_flag "k8s.io/apiserver/pkg/util/flag"
	kube_client "k8s.io/client-go/kubernetes"
	kube_restclient "k8s.io/client-go/rest"
)

var (
	recommendationsCacheTTL = flag.Duration("recommendation-cache-ttl", 2*time.Minute,
		`TTL for cached VPA recommendations`)
)

func main() {
	glog.Infof("starting VPA Initializer")
	kube_flag.InitFlags()

	kubeClient := createKubeClient()
	i := core.NewInitializer(kubeClient, *recommendationsCacheTTL)

	stop := make(chan struct{})
	go i.Run(stop)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	close(stop)
}

func createKubeClient() kube_client.Interface {
	config, err := kube_restclient.InClusterConfig()
	if err != nil {
		glog.Fatalf("failed to build Kuberentes client : failed to create config: %v", err)
	}
	return kube_client.NewForConfigOrDie(config)
}
