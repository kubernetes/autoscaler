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

package main

import (
	"context"
	"flag"
	"k8s.io/klog/v2"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	balancerclientset "k8s.io/autoscaler/balancer/pkg/client/clientset/versioned"
	balancerinformers "k8s.io/autoscaler/balancer/pkg/client/informers/externalversions"
	"k8s.io/autoscaler/balancer/pkg/controller"
	cacheddiscovery "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
	scaleclient "k8s.io/client-go/scale"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	masterURL                  string
	kubeconfig                 string
	balancerReprocessPeriodSec int
	concurrency                int
)

const (
	defaultResyncPeriod = 60 * time.Second
	mapperResetPeriod   = 30 * time.Second
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	flag.IntVar(&balancerReprocessPeriodSec, "reprocess-period-sec", 15, "How often (in second) balancers are processed")
	flag.IntVar(&concurrency, "concurrency", 3, "How many balancers can be processed in parallel")
}

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	// set up signals so we handle the first shutdown signal gracefully
	// TODO: handle sigints
	stopCh := make(chan struct{})

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		klog.Fatalf("Error building kubeconfig: %s", err.Error())
	}
	klog.V(1).Infof("Starting Balancer for %v", cfg)

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	balancerClient, err := balancerclientset.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building Balancer clientset: %s", err.Error())
	}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, defaultResyncPeriod)
	balancerInformerFactory := balancerinformers.NewSharedInformerFactory(balancerClient, defaultResyncPeriod)

	cachedClient := cacheddiscovery.NewMemCacheClient(kubeClient.Discovery())
	restMapper := restmapper.NewDeferredDiscoveryRESTMapper(cachedClient)

	// Synchronize mapper until stopCh is closed.
	go wait.Until(func() {
		restMapper.Reset()
	}, mapperResetPeriod, stopCh)

	scaleKindResolver := scaleclient.NewDiscoveryScaleKindResolver(kubeClient.Discovery())
	scaleClient, err := scaleclient.NewForConfig(cfg, restMapper, dynamic.LegacyAPIPathResolverFunc, scaleKindResolver)

	podInformer := kubeInformerFactory.Core().V1().Pods()
	core := controller.NewCore(controller.NewScaleClient(context.TODO(), scaleClient, restMapper), podInformer)

	controller := controller.NewController(balancerClient,
		balancerInformerFactory.Balancer().V1alpha1().Balancers(),
		kubeClient.CoreV1().Events(""),
		core,
		time.Duration(balancerReprocessPeriodSec)*time.Second)

	// Start method is non-blocking and runs all registered informers in a dedicated goroutine.
	kubeInformerFactory.Start(stopCh)
	balancerInformerFactory.Start(stopCh)

	controller.Run(concurrency, stopCh)
}
