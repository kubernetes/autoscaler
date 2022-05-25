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
