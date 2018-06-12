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

	"github.com/golang/glog"
	kube_flag "k8s.io/apiserver/pkg/util/flag"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/logic"
	"k8s.io/autoscaler/vertical-pod-autoscaler/common"
	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	vpa_lister "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/listers/poc.autoscaling.k8s.io/v1alpha1"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
	"k8s.io/client-go/rest"
)

var (
	certsDir = flag.String("certs-dir", "/etc/tls-certs", `Where the TLS cert files are stored.`)
)

func newReadyVPALister(stopChannel <-chan struct{}) vpa_lister.VerticalPodAutoscalerLister {
	config, err := rest.InClusterConfig()
	if err != nil {
		glog.Fatal(err)
	}
	vpaClient := vpa_clientset.NewForConfigOrDie(config)
	return vpa_api_util.NewAllVpasLister(vpaClient, stopChannel)
}

func main() {
	kube_flag.InitFlags()
	glog.V(1).Infof("Vertical Pod Autoscaler %s Admission Controller", common.VerticalPodAutoscalerVersion)
	certs := initCerts(*certsDir)
	stopChannel := make(chan struct{})
	vpaLister := newReadyVPALister(stopChannel)
	as := logic.NewAdmissionServer(logic.NewRecommendationProvider(vpaLister, vpa_api_util.NewCappingRecommendationProcessor()))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		as.Serve(w, r)
	})
	clientset := getClient()
	server := &http.Server{
		Addr:      ":8000",
		TLSConfig: configTLS(clientset, certs.serverCert, certs.serverKey),
	}
	go selfRegistration(clientset, certs.caCert)
	server.ListenAndServeTLS("", "")
}
