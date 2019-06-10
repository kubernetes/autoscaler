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
	"fmt"

	"net/http"
	"os"
	"time"

	"k8s.io/autoscaler/vertical-pod-autoscaler/common"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/logic"
	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/limitrange"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics"
	metrics_admission "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics/admission"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
	"k8s.io/client-go/informers"
	kube_client "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	kube_flag "k8s.io/component-base/cli/flag"
	"k8s.io/klog"
)

const (
	defaultResyncPeriod time.Duration = 10 * time.Minute
)

var (
	certsConfiguration = &certsConfig{
		clientCaFile:  flag.String("client-ca-file", "/etc/tls-certs/caCert.pem", "Path to CA PEM file."),
		tlsCertFile:   flag.String("tls-cert-file", "/etc/tls-certs/serverCert.pem", "Path to server certificate PEM file."),
		tlsPrivateKey: flag.String("tls-private-key", "/etc/tls-certs/serverKey.pem", "Path to server certificate key PEM file."),
	}

	port           = flag.Int("port", 8000, "The port to listen on.")
	address        = flag.String("address", ":8944", "The address to expose Prometheus metrics.")
	namespace      = os.Getenv("NAMESPACE")
	webhookAddress = flag.String("webhook-address", "", "Address under which webhook is registered. Used when registerByURL is set to true.")
	webhookPort    = flag.String("webhook-port", "", "Server Port for Webhook")
	registerByURL  = flag.Bool("register-by-url", false, "If set to true, admission webhook will be registered by URL (webhookAddress:webhookPort) instead of by service name")
)

func main() {
	klog.InitFlags(nil)
	kube_flag.InitFlags()
	klog.V(1).Infof("Vertical Pod Autoscaler %s Admission Controller", common.VerticalPodAutoscalerVersion)

	healthCheck := metrics.NewHealthCheck(time.Minute, false)
	metrics.Initialize(*address, healthCheck)
	metrics_admission.Register()

	certs := initCerts(*certsConfiguration)

	config, err := rest.InClusterConfig()
	if err != nil {
		klog.Fatal(err)
	}

	vpaClient := vpa_clientset.NewForConfigOrDie(config)
	vpaLister := vpa_api_util.NewAllVpasLister(vpaClient, make(chan struct{}))
	kubeClient := kube_client.NewForConfigOrDie(config)
	factory := informers.NewSharedInformerFactory(kubeClient, defaultResyncPeriod)
	targetSelectorFetcher := target.NewVpaTargetSelectorFetcher(config, kubeClient, factory)
	podPreprocessor := logic.NewDefaultPodPreProcessor()
	vpaPreprocessor := logic.NewDefaultVpaPreProcessor()
	var limitRangeCalculator limitrange.LimitRangeCalculator
	limitRangeCalculator, err = limitrange.NewLimitsRangeCalculator(factory)
	if err != nil {
		klog.Errorf("Failed to create limitRangeCalculator, falling back to not checking limits. Error message: %s", err)
		limitRangeCalculator = limitrange.NewNoopLimitsCalculator()
	}
	recommendationProvider := logic.NewRecommendationProvider(limitRangeCalculator, vpa_api_util.NewCappingRecommendationProcessor(limitRangeCalculator), targetSelectorFetcher, vpaLister)

	as := logic.NewAdmissionServer(recommendationProvider, podPreprocessor, vpaPreprocessor, limitRangeCalculator)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		as.Serve(w, r)
		healthCheck.UpdateLastActivity()
	})
	clientset := getClient()
	server := &http.Server{
		Addr:      fmt.Sprintf(":%d", *port),
		TLSConfig: configTLS(clientset, certs.serverCert, certs.serverKey),
	}
	url := fmt.Sprintf("%v:%v", webhookAddress, webhookPort)
	go selfRegistration(clientset, certs.caCert, &namespace, url, *registerByURL)
	server.ListenAndServeTLS("", "")
}
