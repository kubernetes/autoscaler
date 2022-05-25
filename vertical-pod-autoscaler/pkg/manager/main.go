/*
Copyright 2018 The Kubernetes Authors.

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

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/manager/common"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/manager/resource"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	versioned "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	externalversions "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/informers/externalversions"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

const (
	defaultResyncPeriod  = 10 * time.Minute
	statusUpdateInterval = 1 * time.Minute
)

var (
	kubeconfig        = flag.String("kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	defaultUpdateMode = flag.String("default-update-mode", string(vpa_types.UpdateModeInitial), "Pod default update policy mode, default initial")
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()
	config := common.CreateKubeConfigOrDie(*kubeconfig)
	kubeClient := kubernetes.NewForConfigOrDie(config)

	factory := informers.NewSharedInformerFactory(kubeClient, time.Minute)
	stopCh := make(chan struct{})

	vpaClient := versioned.NewForConfigOrDie(config)
	vpaFactory := externalversions.NewSharedInformerFactory(vpaClient, defaultResyncPeriod)

	fetch := resource.NewFetcherOrDie(kubeClient, factory, vpaFactory, vpaClient, statusUpdateInterval, vpa_types.UpdateMode(*defaultUpdateMode), stopCh)
	factory.Start(stopCh)
	factory.WaitForCacheSync(stopCh)

	vpaFactory.Start(stopCh)
	vpaFactory.WaitForCacheSync(stopCh)

	go fetch.Run(stopCh)
	// listening OS shutdown singal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

}
