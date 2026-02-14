/*
Copyright The Kubernetes Authors.

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

package config

import (
	"flag"
	"os"

	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/api/resource"
	kube_flag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"

	"k8s.io/autoscaler/vertical-pod-autoscaler/common"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/features"
)

// CertsConfig holds configuration related to TLS certificates
type CertsConfig struct {
	ClientCaFile  string
	TlsCertFile   string
	TlsPrivateKey string
	Reload        bool
}

// AdmissionControllerConfig holds all configuration for the admission controller component
type AdmissionControllerConfig struct {
	// Common flags
	CommonFlags *common.CommonFlags

	CertsConfiguration *CertsConfig

	Ciphers              string
	MinTlsVersion        string
	Port                 int
	Address              string
	Namespace            string
	ServiceName          string
	WebhookAddress       string
	WebhookPort          string
	WebhookTimeout       int
	WebhookFailurePolicy bool
	RegisterWebhook      bool
	WebhookLabels        string
	RegisterByURL        bool

	MaxAllowedCPUBoost resource.QuantityValue
}

// DefaultAdmissionControllerConfig returns a AdmissionControllerConfig with default values
func DefaultAdmissionControllerConfig() *AdmissionControllerConfig {
	return &AdmissionControllerConfig{
		CommonFlags: common.DefaultCommonConfig(),
		CertsConfiguration: &CertsConfig{
			ClientCaFile:  "/etc/tls-certs/caCert.pem",
			TlsCertFile:   "/etc/tls-certs/serverCert.pem",
			TlsPrivateKey: "/etc/tls-certs/serverKey.pem",
			Reload:        false,
		},
		Ciphers:              "",
		MinTlsVersion:        "tls1_2",
		Port:                 8000,
		Address:              ":8944",
		Namespace:            os.Getenv("NAMESPACE"),
		ServiceName:          "vpa-webhook",
		WebhookAddress:       "",
		WebhookPort:          "",
		WebhookTimeout:       30,
		WebhookFailurePolicy: false,
		RegisterWebhook:      true,
		WebhookLabels:        "",
		RegisterByURL:        false,

		MaxAllowedCPUBoost: resource.QuantityValue{},
	}
}

// InitAdmissionControllerFlags initializes the flags for the admission controller component
func InitAdmissionControllerFlags() *AdmissionControllerConfig {
	config := DefaultAdmissionControllerConfig()
	config.CommonFlags = common.InitCommonFlags()

	flag.StringVar(&config.CertsConfiguration.ClientCaFile, "client-ca-file", config.CertsConfiguration.ClientCaFile, "Path to CA PEM file.")
	flag.StringVar(&config.CertsConfiguration.TlsCertFile, "tls-cert-file", config.CertsConfiguration.TlsCertFile, "Path to server certificate PEM file.")
	flag.StringVar(&config.CertsConfiguration.TlsPrivateKey, "tls-private-key", config.CertsConfiguration.TlsPrivateKey, "Path to server certificate key PEM file.")
	flag.BoolVar(&config.CertsConfiguration.Reload, "reload-cert", config.CertsConfiguration.Reload, "If set to true, reload leaf and CA certificates when changed.")

	flag.StringVar(&config.Ciphers, "tls-ciphers", config.Ciphers, "A comma-separated or colon-separated list of ciphers to accept.  Only works when min-tls-version is set to tls1_2.")
	flag.StringVar(&config.MinTlsVersion, "min-tls-version", config.MinTlsVersion, "The minimum TLS version to accept.  Must be set to either tls1_2 (default) or tls1_3.")
	flag.IntVar(&config.Port, "port", config.Port, "The port to listen on.")
	flag.StringVar(&config.Address, "address", config.Address, "The address to expose Prometheus metrics.")
	flag.StringVar(&config.ServiceName, "webhook-service", config.ServiceName, "Kubernetes service under which webhook is registered. Used when registerByURL is set to false.")
	flag.StringVar(&config.WebhookAddress, "webhook-address", config.WebhookAddress, "Address under which webhook is registered. Used when registerByURL is set to true.")
	flag.StringVar(&config.WebhookPort, "webhook-port", config.WebhookPort, "Server Port for Webhook")
	flag.IntVar(&config.WebhookTimeout, "webhook-timeout-seconds", config.WebhookTimeout, "Timeout in seconds that the API server should wait for this webhook to respond before failing.")
	flag.BoolVar(&config.WebhookFailurePolicy, "webhook-failure-policy-fail", config.WebhookFailurePolicy, "If set to true, will configure the admission webhook failurePolicy to \"Fail\". Use with caution.")
	flag.BoolVar(&config.RegisterWebhook, "register-webhook", config.RegisterWebhook, "If set to true, admission webhook object will be created on start up to register with the API server.")
	flag.StringVar(&config.WebhookLabels, "webhook-labels", config.WebhookLabels, "Comma separated list of labels to add to the webhook object. Format: key1:value1,key2:value2")
	flag.BoolVar(&config.RegisterByURL, "register-by-url", config.RegisterByURL, "If set to true, admission webhook will be registered by URL (webhookAddress:webhookPort) instead of by service name")

	flag.Var(&config.MaxAllowedCPUBoost, "max-allowed-cpu-boost", "Maximum amount of CPU that will be applied for a container with boost.")

	// These need to happen last. kube_flag.InitFlags() synchronizes and parses
	// flags from the flag package to pflag, so feature gates must be added to
	// pflag before InitFlags() is called.
	klog.InitFlags(nil)
	common.InitLoggingFlags()
	features.MutableFeatureGate.AddFlag(pflag.CommandLine)
	kube_flag.InitFlags()

	ValidateAdmissionControllerConfig(config)

	return config
}

// ValidateAdmissionControllerConfig performs validation of the admission-controller flags
func ValidateAdmissionControllerConfig(config *AdmissionControllerConfig) {
	common.ValidateCommonConfig(config.CommonFlags)
}
