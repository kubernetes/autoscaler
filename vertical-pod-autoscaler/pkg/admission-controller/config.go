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
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	admissionregistration "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

const (
	webhookConfigName = "vpa-webhook-config"
)

func configTLS(cfg certsConfig, minTlsVersion, ciphers string, stop <-chan struct{}) *tls.Config {
	var tlsVersion uint16
	var ciphersuites []uint16
	reverseCipherMap := make(map[string]uint16)

	for _, c := range tls.CipherSuites() {
		reverseCipherMap[c.Name] = c.ID
	}
	for _, c := range strings.Split(strings.ReplaceAll(ciphers, ",", ":"), ":") {
		cipher, ok := reverseCipherMap[c]
		if ok {
			ciphersuites = append(ciphersuites, cipher)
		}
	}
	if len(ciphersuites) == 0 {
		ciphersuites = nil
	}

	switch minTlsVersion {
	case "":
		fallthrough
	case "tls1_2":
		tlsVersion = tls.VersionTLS12
	case "tls1_3":
		tlsVersion = tls.VersionTLS13
	default:
		klog.Fatal(fmt.Errorf("Unable to determine value for --min-tls-version (%s), must be either tls1_2 or tls1_3", minTlsVersion))
	}

	config := &tls.Config{
		MinVersion:   tlsVersion,
		CipherSuites: ciphersuites,
	}
	if *cfg.reload {
		cr := certReloader{
			tlsCertPath: *cfg.tlsCertFile,
			tlsKeyPath:  *cfg.tlsPrivateKey,
		}
		if err := cr.load(); err != nil {
			klog.Fatal(err)
		}
		if err := cr.start(stop); err != nil {
			klog.Fatal(err)
		}
		config.GetCertificate = cr.getCertificate
	} else {
		cert, err := tls.LoadX509KeyPair(*cfg.tlsCertFile, *cfg.tlsPrivateKey)
		if err != nil {
			klog.Fatal(err)
		}
		config.Certificates = []tls.Certificate{cert}
	}
	return config
}

// register this webhook admission controller with the kube-apiserver
// by creating MutatingWebhookConfiguration.
func selfRegistration(clientset kubernetes.Interface, caCert []byte, webHookDelay time.Duration, namespace, serviceName, url string, registerByURL bool, timeoutSeconds int32, selectedNamespace string, ignoredNamespaces []string, webHookFailurePolicy bool) {
	time.Sleep(webHookDelay)
	client := clientset.AdmissionregistrationV1().MutatingWebhookConfigurations()
	_, err := client.Get(context.TODO(), webhookConfigName, metav1.GetOptions{})
	if err == nil {
		if err2 := client.Delete(context.TODO(), webhookConfigName, metav1.DeleteOptions{}); err2 != nil {
			klog.Fatal(err2)
		}
	}
	RegisterClientConfig := admissionregistration.WebhookClientConfig{}
	if !registerByURL {
		RegisterClientConfig.Service = &admissionregistration.ServiceReference{
			Namespace: namespace,
			Name:      serviceName,
		}
	} else {
		RegisterClientConfig.URL = &url
	}
	sideEffects := admissionregistration.SideEffectClassNone

	var failurePolicy admissionregistration.FailurePolicyType
	if webHookFailurePolicy {
		failurePolicy = admissionregistration.Fail
	} else {
		failurePolicy = admissionregistration.Ignore
	}

	RegisterClientConfig.CABundle = caCert

	var namespaceSelector metav1.LabelSelector
	if len(ignoredNamespaces) > 0 {
		namespaceSelector = metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{
				{
					Key:      "kubernetes.io/metadata.name",
					Operator: metav1.LabelSelectorOpNotIn,
					Values:   ignoredNamespaces,
				},
			},
		}
	} else if len(selectedNamespace) > 0 {
		namespaceSelector = metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{
				{
					Key:      "kubernetes.io/metadata.name",
					Operator: metav1.LabelSelectorOpIn,
					Values:   []string{selectedNamespace},
				},
			},
		}
	}
	webhookConfig := &admissionregistration.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: webhookConfigName,
		},
		Webhooks: []admissionregistration.MutatingWebhook{
			{
				Name:                    "vpa.k8s.io",
				AdmissionReviewVersions: []string{"v1"},
				Rules: []admissionregistration.RuleWithOperations{
					{
						Operations: []admissionregistration.OperationType{admissionregistration.Create},
						Rule: admissionregistration.Rule{
							APIGroups:   []string{""},
							APIVersions: []string{"v1"},
							Resources:   []string{"pods"},
						},
					},
					{
						Operations: []admissionregistration.OperationType{admissionregistration.Create, admissionregistration.Update},
						Rule: admissionregistration.Rule{
							APIGroups:   []string{"autoscaling.k8s.io"},
							APIVersions: []string{"*"},
							Resources:   []string{"verticalpodautoscalers"},
						},
					},
				},
				FailurePolicy:     &failurePolicy,
				ClientConfig:      RegisterClientConfig,
				SideEffects:       &sideEffects,
				TimeoutSeconds:    &timeoutSeconds,
				NamespaceSelector: &namespaceSelector,
			},
		},
	}
	if _, err := client.Create(context.TODO(), webhookConfig, metav1.CreateOptions{}); err != nil {
		klog.Fatal(err)
	} else {
		klog.V(3).Info("Self registration as MutatingWebhook succeeded.")
	}
}
