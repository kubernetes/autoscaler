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
	"net/http"
	"time"

	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/autoscaler/vertical-pod-autoscaler/admission-controller/logic"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/poc.autoscaling.k8s.io/v1alpha1"
	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	vpa_lister "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/listers/poc.autoscaling.k8s.io/v1alpha1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

var (
	certsDir = *flag.String("certs-dir", "/etc/tls-certs", `Where the TLS cert files are stored.`)
)

func newReadyVPALister(stopChannel <-chan struct{}) vpa_lister.VerticalPodAutoscalerLister {
	config, err := rest.InClusterConfig()
	if err != nil {
		glog.Fatal(err)
	}
	listWatcher := cache.NewListWatchFromClient(
		vpa_clientset.NewForConfigOrDie(config).PocV1alpha1().RESTClient(),
		"verticalpodautoscalers", v1.NamespaceAll, fields.Everything())
	store := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	lister := vpa_lister.NewVerticalPodAutoscalerLister(store)
	reflector := cache.NewReflector(listWatcher, &v1alpha1.VerticalPodAutoscaler{}, store, time.Hour)
	go reflector.Run(stopChannel)
	return lister
}

func main() {
	flag.Parse()
	initCerts(certsDir)
	stopChannel := make(chan struct{})
	vpaLister := newReadyVPALister(stopChannel)
	as := &admissionServer{logic.NewRecommendationProvider(vpaLister)}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		as.serve(w, r)
	})
	clientset := getClient()
	server := &http.Server{
		Addr:      ":8000",
		TLSConfig: configTLS(clientset),
	}
	go selfRegistration(clientset, caCert)
	server.ListenAndServeTLS("", "")
}
