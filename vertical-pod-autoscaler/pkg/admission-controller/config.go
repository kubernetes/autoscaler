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
	"crypto/x509"
	"fmt"
	"time"

	"k8s.io/api/admissionregistration/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/golang/glog"
)

const (
	webhookConfigName = "vpa-webhook-config"
)

// get a clientset with in-cluster config.
func getClient() *kubernetes.Clientset {
	config, err := rest.InClusterConfig()
	if err != nil {
		glog.Fatal(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Fatal(err)
	}
	return clientset
}

// retrieve the CA cert that will signed the cert used by the
// "GenericAdmissionWebhook" plugin admission controller.
func getAPIServerCert(clientset *kubernetes.Clientset) []byte {
	c, err := clientset.CoreV1().ConfigMaps("kube-system").Get("extension-apiserver-authentication", metav1.GetOptions{})
	if err != nil {
		glog.Fatal(err)
	}

	pem, ok := c.Data["requestheader-client-ca-file"]
	if !ok {
		glog.Fatalf(fmt.Sprintf("cannot find the ca.crt in the configmap, configMap.Data is %#v", c.Data))
	}
	glog.V(4).Info("client-ca-file=", pem)
	return []byte(pem)
}

func configTLS(clientset *kubernetes.Clientset, serverCert, serverKey []byte) *tls.Config {
	cert := getAPIServerCert(clientset)
	apiserverCA := x509.NewCertPool()
	apiserverCA.AppendCertsFromPEM(cert)

	sCert, err := tls.X509KeyPair(serverCert, serverKey)
	if err != nil {
		glog.Fatal(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{sCert},
		ClientCAs:    apiserverCA,
		// Consider changing to tls.RequireAndVerifyClientCert.
		ClientAuth: tls.NoClientCert,
	}
}

// register this webhook admission controller with the kube-apiserver
// by creating MutatingWebhookConfiguration.
func selfRegistration(clientset *kubernetes.Clientset, caCert []byte) {
	time.Sleep(10 * time.Second)
	client := clientset.AdmissionregistrationV1beta1().MutatingWebhookConfigurations()
	_, err := client.Get(webhookConfigName, metav1.GetOptions{})
	if err == nil {
		if err2 := client.Delete(webhookConfigName, nil); err2 != nil {
			glog.Fatal(err2)
		}
	}
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
							APIGroups:   []string{"poc.autoscaling.k8s.io"},
							APIVersions: []string{"v1alpha1"},
							Resources:   []string{"verticalpodautoscalers"},
						},
					}},
				ClientConfig: v1beta1.WebhookClientConfig{
					Service: &v1beta1.ServiceReference{
						Namespace: "kube-system",
						Name:      "vpa-webhook",
					},
					CABundle: caCert,
				},
			},
		},
	}
	if _, err := client.Create(webhookConfig); err != nil {
		glog.Fatal(err)
	} else {
		glog.V(3).Info("Self registration as MutatingWebhook succeeded.")
	}
}
