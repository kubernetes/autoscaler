package main

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/autoscaler/vertical-pod-autoscaler/recommender/metrics"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/contrib/service-loadbalancer/Godeps/_workspace/src/github.com/golang/glog"
	apiv1 "k8s.io/kubernetes/pkg/api/v1"
	kube_client "k8s.io/kubernetes/pkg/client/clientset_generated/clientset"
	v1lister "k8s.io/kubernetes/pkg/client/listers/core/v1"
	resourceclient "k8s.io/metrics/pkg/client/clientset_generated/clientset/typed/metrics/v1alpha1"
)

// Recommender recommend resources for certain containers, based on utilization periodically got from metrics api.
type Recommender interface {
	Run()
}

type recommender struct {
	metricsClient           metrics.Client
	metricsFetchingInterval time.Duration
}

func (r *recommender) Run() {
	utilizations, err := r.metricsClient.GetContainersUtilization()
	if err != nil {
		glog.Errorf("Cannot get containers utilization. Reason: %+v", err)
	}
	for n, utilization := range utilizations {
		fmt.Printf("Utilization #%v: %+v", n, utilization)
	}
}

func NewRecommender(config *rest.Config, metricsFetcherInterval time.Duration) Recommender {
	return &recommender{
		metricsClient:           newMetricsClient(config),
		metricsFetchingInterval: metricsFetcherInterval,
	}
}

func newMetricsClient(config *rest.Config) metrics.Client {
	kubeClient := kube_client.NewForConfigOrDie(config)

	metricsGetter := resourceclient.NewForConfigOrDie(config)
	podLister := newPodLister(kubeClient)
	namespaceLister := newNamespaceLister(kubeClient)

	return metrics.NewClient(metricsGetter, podLister, namespaceLister)
}

func newPodLister(kubeClient kube_client.Interface) v1lister.PodLister {
	podListWatch := cache.NewListWatchFromClient(kubeClient.CoreV1().RESTClient(), "pods", apiv1.NamespaceAll, fields.Everything())
	store := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	podLister := v1lister.NewPodLister(store)
	podReflector := cache.NewReflector(podListWatch, &apiv1.Pod{}, store, time.Hour)
	podReflector.Run()
	return podLister
}

func newNamespaceLister(kubeClient kube_client.Interface) v1lister.NamespaceLister {
	namespaceListWatch := cache.NewListWatchFromClient(kubeClient.CoreV1().RESTClient(), "namespaces", apiv1.NamespaceAll, fields.Everything())
	store := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	namespaceLister := v1lister.NewNamespaceLister(store)
	podReflector := cache.NewReflector(namespaceListWatch, &apiv1.Namespace{}, store, time.Hour)
	podReflector.Run()
	return namespaceLister
}
