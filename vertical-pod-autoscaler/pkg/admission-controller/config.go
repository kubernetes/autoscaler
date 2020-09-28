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
	"time"

	admissionregistration "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
)

const (
	webhookConfigName = "vpa-webhook-config"
)

// get a clientset with in-cluster config.
func getClient() *kubernetes.Clientset {
	config, err := rest.InClusterConfig()
	if err != nil {
		klog.Fatal(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatal(err)
	}
	return clientset
}

func configTLS(clientset *kubernetes.Clientset, serverCert, serverKey []byte) *tls.Config {
	sCert, err := tls.X509KeyPair(serverCert, serverKey)
	if err != nil {
		klog.Fatal(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{sCert},
	}
}

// register this webhook admission controller with the kube-apiserver
// by creating MutatingWebhookConfiguration.
func selfRegistration(clientset *kubernetes.Clientset, caCert []byte, namespace, serviceName, url string, registerByURL bool, timeoutSeconds int32) {
	time.Sleep(10 * time.Second)
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
	failurePolicy := admissionregistration.Ignore
	RegisterClientConfig.CABundle = caCert
	webhookConfig := &admissionregistration.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: webhookConfigName,
		},
		Webhooks: []admissionregistration.MutatingWebhook{
			{
				Name:                    "vpa.k8s.io",
				AdmissionReviewVersions: []string{"v1beta1"},
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
				FailurePolicy:  &failurePolicy,
				ClientConfig:   RegisterClientConfig,
				SideEffects:    &sideEffects,
				TimeoutSeconds: &timeoutSeconds,
			},
		},
	}
	if _, err := client.Create(context.TODO(), webhookConfig, metav1.CreateOptions{}); err != nil {
		klog.Fatal(err)
	} else {
		klog.V(3).Info("Self registration as MutatingWebhook succeeded.")
	}
}
