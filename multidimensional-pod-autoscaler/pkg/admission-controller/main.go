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

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/multidimensional-pod-autoscaler/common"
	"k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/admission-controller/logic"
	mpa "k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/admission-controller/resource/mpa"
	"k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/admission-controller/resource/pod"
	"k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/admission-controller/resource/pod/patch"
	"k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/admission-controller/resource/pod/recommendation"
	mpa_clientset "k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/client/clientset/versioned"
	"k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/target"
	mpa_api_util "k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/utils/mpa"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/limitrange"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics"
	metrics_admission "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics/admission"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/status"
	"k8s.io/client-go/informers"
	kube_client "k8s.io/client-go/kubernetes"
	kube_flag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"
)

const (
	defaultResyncPeriod  = 10 * time.Minute
	statusUpdateInterval = 10 * time.Second
)

var (
	certsConfiguration = &certsConfig{
		clientCaFile:  flag.String("client-ca-file", "/etc/tls-certs/caCert.pem", "Path to CA PEM file."),
		tlsCertFile:   flag.String("tls-cert-file", "/etc/tls-certs/serverCert.pem", "Path to server certificate PEM file."),
		tlsPrivateKey: flag.String("tls-private-key", "/etc/tls-certs/serverKey.pem", "Path to server certificate key PEM file."),
	}

	port               = flag.Int("port", 8000, "The port to listen on.")
	address            = flag.String("address", ":8944", "The address to expose Prometheus metrics.")
	kubeconfig         = flag.String("kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	kubeApiQps         = flag.Float64("kube-api-qps", 5.0, `QPS limit when making requests to Kubernetes apiserver`)
	kubeApiBurst       = flag.Float64("kube-api-burst", 10.0, `QPS burst limit when making requests to Kubernetes apiserver`)
	namespace          = os.Getenv("NAMESPACE")
	serviceName        = flag.String("webhook-service", "mpa-webhook", "Kubernetes service under which webhook is registered. Used when registerByURL is set to false.")
	webhookAddress     = flag.String("webhook-address", "", "Address under which webhook is registered. Used when registerByURL is set to true.")
	webhookPort        = flag.String("webhook-port", "", "Server Port for Webhook")
	webhookTimeout     = flag.Int("webhook-timeout-seconds", 30, "Timeout in seconds that the API server should wait for this webhook to respond before failing.")
	registerWebhook    = flag.Bool("register-webhook", true, "If set to true, admission webhook object will be created on start up to register with the API server.")
	registerByURL      = flag.Bool("register-by-url", false, "If set to true, admission webhook will be registered by URL (webhookAddress:webhookPort) instead of by service name")
	mpaObjectNamespace = flag.String("mpa-object-namespace", apiv1.NamespaceAll, "Namespace to search for MPA objects. Empty means all namespaces will be used.")
)

func main() {
	klog.InitFlags(nil)
	kube_flag.InitFlags()
	klog.V(1).Infof("Multi-dimensional Pod Autoscaler %s Admission Controller", common.MultidimPodAutoscalerVersion)

	healthCheck := metrics.NewHealthCheck(time.Minute, false)
	metrics.Initialize(*address, healthCheck)
	metrics_admission.Register()

	certs := initCerts(*certsConfiguration)
	klog.V(4).Infof("Certificates initialized!")
	config := common.CreateKubeConfigOrDie(*kubeconfig, float32(*kubeApiQps), int(*kubeApiBurst))

	mpaClient := mpa_clientset.NewForConfigOrDie(config)
	mpaLister := mpa_api_util.NewMpasLister(mpaClient, make(chan struct{}), *mpaObjectNamespace)
	kubeClient := kube_client.NewForConfigOrDie(config)
	factory := informers.NewSharedInformerFactory(kubeClient, defaultResyncPeriod)
	targetSelectorFetcher := target.NewMpaTargetSelectorFetcher(config, kubeClient, factory)
	podPreprocessor := pod.NewDefaultPreProcessor()
	mpaPreprocessor := mpa.NewDefaultPreProcessor()
	var limitRangeCalculator limitrange.LimitRangeCalculator
	limitRangeCalculator, err := limitrange.NewLimitsRangeCalculator(factory)
	if err != nil {
		klog.Errorf("Failed to create limitRangeCalculator, falling back to not checking limits. Error message: %s", err)
		limitRangeCalculator = limitrange.NewNoopLimitsCalculator()
	}
	recommendationProvider := recommendation.NewProvider(limitRangeCalculator, mpa_api_util.NewCappingRecommendationProcessor(limitRangeCalculator))
	vpaMatcher := mpa.NewMatcher(mpaLister, targetSelectorFetcher)

	hostname, err := os.Hostname()
	if err != nil {
		klog.Fatalf("Unable to get hostname: %v", err)
	}

	statusNamespace := status.AdmissionControllerStatusNamespace
	if namespace != "" {
		statusNamespace = namespace
	}
	stopCh := make(chan struct{})
	statusUpdater := status.NewUpdater(
		kubeClient,
		status.AdmissionControllerStatusName,
		statusNamespace,
		statusUpdateInterval,
		hostname,
	)
	defer close(stopCh)

	calculators := []patch.Calculator{patch.NewResourceUpdatesCalculator(recommendationProvider), patch.NewObservedContainersCalculator()}
	as := logic.NewAdmissionServer(podPreprocessor, mpaPreprocessor, limitRangeCalculator, vpaMatcher, calculators)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		as.Serve(w, r)
		healthCheck.UpdateLastActivity()
	})
	server := &http.Server{
		Addr:      fmt.Sprintf(":%d", *port),
		TLSConfig: configTLS(certs.serverCert, certs.serverKey),
	}
	url := fmt.Sprintf("%v:%v", *webhookAddress, *webhookPort)
	go func() {
		if *registerWebhook {
			selfRegistration(kubeClient, certs.caCert, namespace, *serviceName, url, *registerByURL, int32(*webhookTimeout))
		}
		// Start status updates after the webhook is initialized.
		statusUpdater.Run(stopCh)
	}()

	if err = server.ListenAndServeTLS("", ""); err != nil {
		klog.Fatalf("HTTPS Error: %s", err)
	}
}
