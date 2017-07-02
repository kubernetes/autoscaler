package main

import (
	"k8s.io/autoscaler/vertical-pod-autoscaler/recommender/metrics"
	kube_client "k8s.io/kubernetes/pkg/client/clientset_generated/clientset"
	"time"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
	apiv1 "k8s.io/kubernetes/pkg/api/v1"
	v1lister "k8s.io/kubernetes/pkg/client/listers/core/v1"
	resourceclient "k8s.io/metrics/pkg/client/clientset_generated/clientset/typed/metrics/v1alpha1"
	"k8s.io/client-go/rest"
	"fmt"
	"k8s.io/contrib/service-loadbalancer/Godeps/_workspace/src/github.com/golang/glog"
)

type Recommender interface {
	Run()
}

type recommender struct {
	metricsClient metrics.MetricsClient
	metricsFetcherInterval time.Duration
}

func (r *recommender) Run() {
	utilizations, err:= r.metricsClient.GetContainersUtilization();
	if err != nil {
		glog.Error("Cannot get containers utilization. Reason: %+v", err)
	}else {
		for n, utilization:= range utilizations {
			fmt.Printf("Utilization #%v: %+v", n, utilization)
		}
	}
}

func NewRecommender(config *rest.Config, metricsFetcherInterval time.Duration) Recommender {
	return &recommender{
		metricsClient: newMetricsClient(config),
		metricsFetcherInterval: metricsFetcherInterval,
	}
}

func newMetricsClient(config *rest.Config) metrics.MetricsClient{
	kubeClient:= kube_client.NewForConfigOrDie(config)

	metricsGetter:= resourceclient.NewForConfigOrDie(config)
	podLister:= newPodLister(kubeClient)
	namespaceLister:= newNamespaceLister(kubeClient)

	return metrics.NewMetricsClient(metricsGetter, podLister, namespaceLister)
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
	namespaceLister:= v1lister.NewNamespaceLister(store)
	podReflector := cache.NewReflector(namespaceListWatch, &apiv1.Namespace{}, store, time.Hour)
	podReflector.Run()
	return namespaceLister
}
