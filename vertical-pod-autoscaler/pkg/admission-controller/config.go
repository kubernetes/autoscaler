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
	"crypto/tls"
	"time"

	"k8s.io/api/admissionregistration/v1beta1"
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
func selfRegistration(clientset *kubernetes.Clientset, caCert []byte, namespace *string, url string, registerByURL bool) {
	time.Sleep(10 * time.Second)
	client := clientset.AdmissionregistrationV1beta1().MutatingWebhookConfigurations()
	_, err := client.Get(webhookConfigName, metav1.GetOptions{})
	if err == nil {
		if err2 := client.Delete(webhookConfigName, nil); err2 != nil {
			klog.Fatal(err2)
		}
	}
	RegisterClientConfig := v1beta1.WebhookClientConfig{}
	if !registerByURL {
		RegisterClientConfig.Service = &v1beta1.ServiceReference{
			Namespace: *namespace,
			Name:      "vpa-webhook",
		}
	} else {
		RegisterClientConfig.URL = &url
	}
	RegisterClientConfig.CABundle = caCert
	webhookConfig := &v1beta1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: webhookConfigName,
		},
		Webhooks: []v1beta1.Webhook{
			{
				Name: "vpa.k8s.io",
				Rules: []v1beta1.RuleWithOperations{
					{
						Operations: []v1beta1.OperationType{v1beta1.Create},
						Rule: v1beta1.Rule{
							APIGroups:   []string{""},
							APIVersions: []string{"v1"},
							Resources:   []string{"pods"},
						},
					},
					{
						Operations: []v1beta1.OperationType{v1beta1.Create, v1beta1.Update},
						Rule: v1beta1.Rule{
							APIGroups:   []string{"autoscaling.k8s.io"},
							APIVersions: []string{"v1beta2"},
							Resources:   []string{"verticalpodautoscalers"},
						},
					},
					{
						Operations: []v1beta1.OperationType{v1beta1.Create, v1beta1.Update},
						Rule: v1beta1.Rule{
							APIGroups:   []string{"autoscaling.k8s.io"},
							APIVersions: []string{"v1"},
							Resources:   []string{"verticalpodautoscalers"},
						},
					},
				},
				ClientConfig: RegisterClientConfig,
			},
		},
	}
	if _, err := client.Create(webhookConfig); err != nil {
		klog.Fatal(err)
	} else {
		klog.V(3).Info("Self registration as MutatingWebhook succeeded.")
	}
}
