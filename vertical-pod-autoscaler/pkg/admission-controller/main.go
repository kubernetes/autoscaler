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
	"strings"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	kube_client "k8s.io/client-go/kubernetes"
	kube_flag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"

	"k8s.io/autoscaler/vertical-pod-autoscaler/common"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/logic"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource/pod"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource/pod/patch"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource/pod/recommendation"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource/vpa"
	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target"
	controllerfetcher "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target/controller_fetcher"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/limitrange"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics"
	metrics_admission "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics/admission"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/server"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/status"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
)

const (
	defaultResyncPeriod                        = 10 * time.Minute
	statusUpdateInterval                       = 10 * time.Second
	scaleCacheEntryLifetime      time.Duration = time.Hour
	scaleCacheEntryFreshnessTime time.Duration = 10 * time.Minute
	scaleCacheEntryJitterFactor  float64       = 1.
	webHookDelay                               = 10 * time.Second
)

var (
	certsConfiguration = &certsConfig{
		clientCaFile:  flag.String("client-ca-file", "/etc/tls-certs/caCert.pem", "Path to CA PEM file."),
		tlsCertFile:   flag.String("tls-cert-file", "/etc/tls-certs/serverCert.pem", "Path to server certificate PEM file."),
		tlsPrivateKey: flag.String("tls-private-key", "/etc/tls-certs/serverKey.pem", "Path to server certificate key PEM file."),
		reload:        flag.Bool("reload-cert", false, "If set to true, reload leaf certificate."),
	}
	ciphers       = flag.String("tls-ciphers", "", "A comma-separated or colon-separated list of ciphers to accept.  Only works when min-tls-version is set to tls1_2.")
	minTlsVersion = flag.String("min-tls-version", "tls1_2", "The minimum TLS version to accept.  Must be set to either tls1_2 (default) or tls1_3.")

	port                       = flag.Int("port", 8000, "The port to listen on.")
	address                    = flag.String("address", ":8944", "The address to expose Prometheus metrics.")
	kubeconfig                 = flag.String("kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	kubeApiQps                 = flag.Float64("kube-api-qps", 5.0, `QPS limit when making requests to Kubernetes apiserver`)
	kubeApiBurst               = flag.Float64("kube-api-burst", 10.0, `QPS burst limit when making requests to Kubernetes apiserver`)
	enableProfiling            = flag.Bool("profiling", false, "Is debug/pprof endpoint enabled")
	namespace                  = os.Getenv("NAMESPACE")
	serviceName                = flag.String("webhook-service", "vpa-webhook", "Kubernetes service under which webhook is registered. Used when registerByURL is set to false.")
	webhookAddress             = flag.String("webhook-address", "", "Address under which webhook is registered. Used when registerByURL is set to true.")
	webhookPort                = flag.String("webhook-port", "", "Server Port for Webhook")
	webhookTimeout             = flag.Int("webhook-timeout-seconds", 30, "Timeout in seconds that the API server should wait for this webhook to respond before failing.")
	webHookFailurePolicy       = flag.Bool("webhook-failure-policy-fail", false, "If set to true, will configure the admission webhook failurePolicy to \"Fail\". Use with caution.")
	registerWebhook            = flag.Bool("register-webhook", true, "If set to true, admission webhook object will be created on start up to register with the API server.")
	registerByURL              = flag.Bool("register-by-url", false, "If set to true, admission webhook will be registered by URL (webhookAddress:webhookPort) instead of by service name")
	vpaObjectNamespace         = flag.String("vpa-object-namespace", apiv1.NamespaceAll, "Namespace to search for VPA objects. Empty means all namespaces will be used. Must not be used if ignored-vpa-object-namespaces is set.")
	ignoredVpaObjectNamespaces = flag.String("ignored-vpa-object-namespaces", "", "Comma separated list of namespaces to ignore. Must not be used if vpa-object-namespace is used.")
)

func main() {
	klog.InitFlags(nil)
	kube_flag.InitFlags()
	klog.V(1).Infof("Vertical Pod Autoscaler %s Admission Controller", common.VerticalPodAutoscalerVersion)

	if len(*vpaObjectNamespace) > 0 && len(*ignoredVpaObjectNamespaces) > 0 {
		klog.Fatalf("--vpa-object-namespace and --ignored-vpa-object-namespaces are mutually exclusive and can't be set together.")
	}

	healthCheck := metrics.NewHealthCheck(time.Minute)
	metrics_admission.Register()
	server.Initialize(enableProfiling, healthCheck, address)

	config := common.CreateKubeConfigOrDie(*kubeconfig, float32(*kubeApiQps), int(*kubeApiBurst))

	vpaClient := vpa_clientset.NewForConfigOrDie(config)
	vpaLister := vpa_api_util.NewVpasLister(vpaClient, make(chan struct{}), *vpaObjectNamespace)
	kubeClient := kube_client.NewForConfigOrDie(config)
	factory := informers.NewSharedInformerFactory(kubeClient, defaultResyncPeriod)
	targetSelectorFetcher := target.NewVpaTargetSelectorFetcher(config, kubeClient, factory)
	controllerFetcher := controllerfetcher.NewControllerFetcher(config, kubeClient, factory, scaleCacheEntryFreshnessTime, scaleCacheEntryLifetime, scaleCacheEntryJitterFactor)
	podPreprocessor := pod.NewDefaultPreProcessor()
	vpaPreprocessor := vpa.NewDefaultPreProcessor()
	var limitRangeCalculator limitrange.LimitRangeCalculator
	limitRangeCalculator, err := limitrange.NewLimitsRangeCalculator(factory)
	if err != nil {
		klog.ErrorS(err, "Failed to create limitRangeCalculator, falling back to not checking limits.")
		limitRangeCalculator = limitrange.NewNoopLimitsCalculator()
	}
	recommendationProvider := recommendation.NewProvider(limitRangeCalculator, vpa_api_util.NewCappingRecommendationProcessor(limitRangeCalculator))
	vpaMatcher := vpa.NewMatcher(vpaLister, targetSelectorFetcher, controllerFetcher)

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
	as := logic.NewAdmissionServer(podPreprocessor, vpaPreprocessor, limitRangeCalculator, vpaMatcher, calculators)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		as.Serve(w, r)
		healthCheck.UpdateLastActivity()
	})
	server := &http.Server{
		Addr:      fmt.Sprintf(":%d", *port),
		TLSConfig: configTLS(*certsConfiguration, *minTlsVersion, *ciphers, stopCh),
	}
	url := fmt.Sprintf("%v:%v", *webhookAddress, *webhookPort)
	ignoredNamespaces := strings.Split(*ignoredVpaObjectNamespaces, ",")
	go func() {
		if *registerWebhook {
			selfRegistration(kubeClient, readFile(*certsConfiguration.clientCaFile), webHookDelay, namespace, *serviceName, url, *registerByURL, int32(*webhookTimeout), *vpaObjectNamespace, ignoredNamespaces, *webHookFailurePolicy)
		}
		// Start status updates after the webhook is initialized.
		statusUpdater.Run(stopCh)
	}()

	if err = server.ListenAndServeTLS("", ""); err != nil {
		klog.Fatalf("HTTPS Error: %s", err)
	}
}
