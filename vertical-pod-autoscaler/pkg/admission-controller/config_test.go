/*
Copyright 2024 The Kubernetes Authors.

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
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	admissionregistration "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/config"
)

func TestSelfRegistrationBase(t *testing.T) {
	testClientSet := fake.NewClientset()
	caCert := []byte("fake")
	webHookDelay := 0 * time.Second
	namespace := "default"
	serviceName := "vpa-service"
	url := "http://example.com/"
	registerByURL := true
	timeoutSeconds := int32(32)
	selectedNamespace := ""
	ignoredNamespaces := []string{}

	selfRegistration(testClientSet, caCert, webHookDelay, namespace, serviceName, url, registerByURL, timeoutSeconds, selectedNamespace, ignoredNamespaces, false, "key1:value1,key2:value2")

	webhookConfigInterface := testClientSet.AdmissionregistrationV1().MutatingWebhookConfigurations()
	webhookConfig, err := webhookConfigInterface.Get(context.TODO(), webhookConfigName, metav1.GetOptions{})

	assert.NoError(t, err, "expected no error fetching webhook configuration")
	assert.Equal(t, webhookConfigName, webhookConfig.Name, "expected webhook configuration name to match")
	assert.Equal(t, webhookConfig.Labels, map[string]string{"key1": "value1", "key2": "value2"}, "expected webhook configuration labels to match")

	assert.Len(t, webhookConfig.Webhooks, 1, "expected one webhook configuration")
	webhook := webhookConfig.Webhooks[0]
	assert.Equal(t, "vpa.k8s.io", webhook.Name, "expected webhook name to match")

	podRule := webhook.Rules[0]
	assert.Equal(t, []admissionregistration.OperationType{admissionregistration.Create}, podRule.Operations, "expected operations to match")
	assert.Equal(t, []string{""}, podRule.APIGroups, "expected API groups to match")
	assert.Equal(t, []string{"v1"}, podRule.APIVersions, "expected API versions to match")
	assert.Equal(t, []string{"pods"}, podRule.Resources, "expected resources to match")

	vpaRule := webhook.Rules[1]
	assert.Equal(t, []admissionregistration.OperationType{admissionregistration.Create, admissionregistration.Update}, vpaRule.Operations, "expected operations to match")
	assert.Equal(t, []string{"autoscaling.k8s.io"}, vpaRule.APIGroups, "expected API groups to match")
	assert.Equal(t, []string{"*"}, vpaRule.APIVersions, "ehook.Rulxpected API versions to match")
	assert.Equal(t, []string{"verticalpodautoscalers"}, vpaRule.Resources, "expected resources to match")

	assert.Equal(t, admissionregistration.SideEffectClassNone, *webhook.SideEffects, "expected side effects to match")
	assert.Equal(t, admissionregistration.Ignore, *webhook.FailurePolicy, "expected failure policy to match")
	assert.Equal(t, caCert, webhook.ClientConfig.CABundle, "expected CA bundle to match")
	assert.Equal(t, timeoutSeconds, *webhook.TimeoutSeconds, "expected timeout seconds to match")
}

func TestSelfRegistrationWithURL(t *testing.T) {
	testClientSet := fake.NewClientset()
	caCert := []byte("fake")
	webHookDelay := 0 * time.Second
	namespace := "default"
	serviceName := "vpa-service"
	url := "http://example.com/"
	registerByURL := true
	timeoutSeconds := int32(32)
	selectedNamespace := ""
	ignoredNamespaces := []string{}

	selfRegistration(testClientSet, caCert, webHookDelay, namespace, serviceName, url, registerByURL, timeoutSeconds, selectedNamespace, ignoredNamespaces, false, "")

	webhookConfigInterface := testClientSet.AdmissionregistrationV1().MutatingWebhookConfigurations()
	webhookConfig, err := webhookConfigInterface.Get(context.TODO(), webhookConfigName, metav1.GetOptions{})

	assert.NoError(t, err, "expected no error fetching webhook configuration")

	assert.Len(t, webhookConfig.Webhooks, 1, "expected one webhook configuration")
	webhook := webhookConfig.Webhooks[0]

	assert.Nil(t, webhook.ClientConfig.Service, "expected service reference to be nil")
	assert.NotNil(t, webhook.ClientConfig.URL, "expected URL to be set")
	assert.Equal(t, url, *webhook.ClientConfig.URL, "expected URL to match")
}

func TestSelfRegistrationWithOutURL(t *testing.T) {
	testClientSet := fake.NewClientset()
	caCert := []byte("fake")
	webHookDelay := 0 * time.Second
	namespace := "default"
	serviceName := "vpa-service"
	url := "http://example.com/"
	registerByURL := false
	timeoutSeconds := int32(32)
	selectedNamespace := ""
	ignoredNamespaces := []string{}

	selfRegistration(testClientSet, caCert, webHookDelay, namespace, serviceName, url, registerByURL, timeoutSeconds, selectedNamespace, ignoredNamespaces, false, "")

	webhookConfigInterface := testClientSet.AdmissionregistrationV1().MutatingWebhookConfigurations()
	webhookConfig, err := webhookConfigInterface.Get(context.TODO(), webhookConfigName, metav1.GetOptions{})

	assert.NoError(t, err, "expected no error fetching webhook configuration")

	assert.Len(t, webhookConfig.Webhooks, 1, "expected one webhook configuration")
	webhook := webhookConfig.Webhooks[0]

	assert.NotNil(t, webhook.ClientConfig.Service, "expected service reference to be nil")
	assert.Equal(t, webhook.ClientConfig.Service.Name, serviceName, "expected service name to be equal")
	assert.Equal(t, webhook.ClientConfig.Service.Namespace, namespace, "expected service namespace to be equal")

	assert.Nil(t, webhook.ClientConfig.URL, "expected URL to be set")
}

func TestSelfRegistrationWithIgnoredNamespaces(t *testing.T) {
	testClientSet := fake.NewClientset()
	caCert := []byte("fake")
	webHookDelay := 0 * time.Second
	namespace := "default"
	serviceName := "vpa-service"
	url := "http://example.com/"
	registerByURL := false
	timeoutSeconds := int32(32)
	selectedNamespace := ""
	ignoredNamespaces := []string{"test"}

	selfRegistration(testClientSet, caCert, webHookDelay, namespace, serviceName, url, registerByURL, timeoutSeconds, selectedNamespace, ignoredNamespaces, false, "")

	webhookConfigInterface := testClientSet.AdmissionregistrationV1().MutatingWebhookConfigurations()
	webhookConfig, err := webhookConfigInterface.Get(context.TODO(), webhookConfigName, metav1.GetOptions{})

	assert.NoError(t, err, "expected no error fetching webhook configuration")

	assert.Len(t, webhookConfig.Webhooks, 1, "expected one webhook configuration")
	webhook := webhookConfig.Webhooks[0]

	assert.NotNil(t, webhook.NamespaceSelector.MatchExpressions, "expected namespace selector not to be nil")
	assert.Len(t, webhook.NamespaceSelector.MatchExpressions, 1, "expected one match expression")

	matchExpression := webhook.NamespaceSelector.MatchExpressions[0]
	assert.Equal(t, matchExpression.Operator, metav1.LabelSelectorOpNotIn, "expected namespace operator to be OpNotIn")
	assert.Equal(t, matchExpression.Values, ignoredNamespaces, "expected namespace selector match expression to be equal")
}

func TestSelfRegistrationWithSelectedNamespaces(t *testing.T) {
	testClientSet := fake.NewClientset()
	caCert := []byte("fake")
	webHookDelay := 0 * time.Second
	namespace := "default"
	serviceName := "vpa-service"
	url := "http://example.com/"
	registerByURL := false
	timeoutSeconds := int32(32)
	selectedNamespace := "test"
	ignoredNamespaces := []string{}

	selfRegistration(testClientSet, caCert, webHookDelay, namespace, serviceName, url, registerByURL, timeoutSeconds, selectedNamespace, ignoredNamespaces, false, "")

	webhookConfigInterface := testClientSet.AdmissionregistrationV1().MutatingWebhookConfigurations()
	webhookConfig, err := webhookConfigInterface.Get(context.TODO(), webhookConfigName, metav1.GetOptions{})

	assert.NoError(t, err, "expected no error fetching webhook configuration")

	assert.Len(t, webhookConfig.Webhooks, 1, "expected one webhook configuration")
	webhook := webhookConfig.Webhooks[0]

	assert.NotNil(t, webhook.NamespaceSelector.MatchExpressions, "expected namespace selector not to be nil")
	assert.Len(t, webhook.NamespaceSelector.MatchExpressions, 1, "expected one match expression")

	matchExpression := webhook.NamespaceSelector.MatchExpressions[0]
	assert.Equal(t, metav1.LabelSelectorOpIn, matchExpression.Operator, "expected namespace operator to be OpIn")
	assert.Equal(t, matchExpression.Operator, metav1.LabelSelectorOpIn, "expected namespace operator to be OpIn")
	assert.Equal(t, matchExpression.Values, []string{selectedNamespace}, "expected namespace selector match expression to be equal")
}

func TestSelfRegistrationWithFailurePolicy(t *testing.T) {
	testClientSet := fake.NewClientset()
	caCert := []byte("fake")
	webHookDelay := 0 * time.Second
	namespace := "default"
	serviceName := "vpa-service"
	url := "http://example.com/"
	registerByURL := false
	timeoutSeconds := int32(32)
	selectedNamespace := "test"
	ignoredNamespaces := []string{}

	selfRegistration(testClientSet, caCert, webHookDelay, namespace, serviceName, url, registerByURL, timeoutSeconds, selectedNamespace, ignoredNamespaces, true, "")

	webhookConfigInterface := testClientSet.AdmissionregistrationV1().MutatingWebhookConfigurations()
	webhookConfig, err := webhookConfigInterface.Get(context.TODO(), webhookConfigName, metav1.GetOptions{})

	assert.NoError(t, err, "expected no error fetching webhook configuration")

	assert.Len(t, webhookConfig.Webhooks, 1, "expected one webhook configuration")
	webhook := webhookConfig.Webhooks[0]

	assert.NotNil(t, *webhook.FailurePolicy, "expected failurePolicy not to be nil")
	assert.Equal(t, *webhook.FailurePolicy, admissionregistration.Fail, "expected failurePolicy to be Fail")
}

func TestSelfRegistrationWithOutFailurePolicy(t *testing.T) {
	testClientSet := fake.NewClientset()
	caCert := []byte("fake")
	webHookDelay := 0 * time.Second
	namespace := "default"
	serviceName := "vpa-service"
	url := "http://example.com/"
	registerByURL := false
	timeoutSeconds := int32(32)
	selectedNamespace := "test"
	ignoredNamespaces := []string{}

	selfRegistration(testClientSet, caCert, webHookDelay, namespace, serviceName, url, registerByURL, timeoutSeconds, selectedNamespace, ignoredNamespaces, false, "")

	webhookConfigInterface := testClientSet.AdmissionregistrationV1().MutatingWebhookConfigurations()
	webhookConfig, err := webhookConfigInterface.Get(context.TODO(), webhookConfigName, metav1.GetOptions{})

	assert.NoError(t, err, "expected no error fetching webhook configuration")

	assert.Len(t, webhookConfig.Webhooks, 1, "expected one webhook configuration")
	webhook := webhookConfig.Webhooks[0]

	assert.NotNil(t, *webhook.FailurePolicy, "expected namespace selector not to be nil")
	assert.Equal(t, *webhook.FailurePolicy, admissionregistration.Ignore, "expected failurePolicy to be Ignore")
}

func TestSelfRegistrationWithInvalidLabels(t *testing.T) {
	testClientSet := fake.NewClientset()
	caCert := []byte("fake")
	webHookDelay := 0 * time.Second
	namespace := "default"
	serviceName := "vpa-service"
	url := "http://example.com/"
	registerByURL := true
	timeoutSeconds := int32(32)
	selectedNamespace := ""
	ignoredNamespaces := []string{}

	selfRegistration(testClientSet, caCert, webHookDelay, namespace, serviceName, url, registerByURL, timeoutSeconds, selectedNamespace, ignoredNamespaces, false, "foo,bar")

	webhookConfigInterface := testClientSet.AdmissionregistrationV1().MutatingWebhookConfigurations()
	webhookConfig, err := webhookConfigInterface.Get(context.TODO(), webhookConfigName, metav1.GetOptions{})

	assert.NoError(t, err, "expected no error fetching webhook configuration")
	assert.Equal(t, webhookConfigName, webhookConfig.Name, "expected webhook configuration name to match")
	assert.Equal(t, webhookConfig.Labels, map[string]string{}, "expected invalid webhook configuration labels to match")

	assert.Len(t, webhookConfig.Webhooks, 1, "expected one webhook configuration")
	webhook := webhookConfig.Webhooks[0]
	assert.Equal(t, "vpa.k8s.io", webhook.Name, "expected webhook name to match")

	podRule := webhook.Rules[0]
	assert.Equal(t, []admissionregistration.OperationType{admissionregistration.Create}, podRule.Operations, "expected operations to match")
	assert.Equal(t, []string{""}, podRule.APIGroups, "expected API groups to match")
	assert.Equal(t, []string{"v1"}, podRule.APIVersions, "expected API versions to match")
	assert.Equal(t, []string{"pods"}, podRule.Resources, "expected resources to match")

	vpaRule := webhook.Rules[1]
	assert.Equal(t, []admissionregistration.OperationType{admissionregistration.Create, admissionregistration.Update}, vpaRule.Operations, "expected operations to match")
	assert.Equal(t, []string{"autoscaling.k8s.io"}, vpaRule.APIGroups, "expected API groups to match")
	assert.Equal(t, []string{"*"}, vpaRule.APIVersions, "ehook.Rulxpected API versions to match")
	assert.Equal(t, []string{"verticalpodautoscalers"}, vpaRule.Resources, "expected resources to match")

	assert.Equal(t, admissionregistration.SideEffectClassNone, *webhook.SideEffects, "expected side effects to match")
	assert.Equal(t, admissionregistration.Ignore, *webhook.FailurePolicy, "expected failure policy to match")
	assert.Equal(t, caCert, webhook.ClientConfig.CABundle, "expected CA bundle to match")
	assert.Equal(t, timeoutSeconds, *webhook.TimeoutSeconds, "expected timeout seconds to match")
}

func TestConvertLabelsToMap(t *testing.T) {
	testCases := []struct {
		desc           string
		labels         string
		expectedOutput map[string]string
		expectedError  bool
	}{
		{
			desc:           "should return empty map when tag is empty",
			labels:         "",
			expectedOutput: map[string]string{},
			expectedError:  false,
		},
		{
			desc:   "single valid tag should be converted",
			labels: "key:value",
			expectedOutput: map[string]string{
				"key": "value",
			},
			expectedError: false,
		},
		{
			desc:   "multiple valid labels should be converted",
			labels: "key1:value1,key2:value2",
			expectedOutput: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			expectedError: false,
		},
		{
			desc:   "whitespaces should be trimmed",
			labels: "key1:value1, key2:value2",
			expectedOutput: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			expectedError: false,
		},
		{
			desc:   "whitespaces between keys and values should be trimmed",
			labels: "key1 : value1,key2 : value2",
			expectedOutput: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			expectedError: false,
		},
		{
			desc:           "should return error for invalid format",
			labels:         "foo,bar",
			expectedOutput: nil,
			expectedError:  true,
		},
		{
			desc:           "should return error for when key is missed",
			labels:         "key1:value1,:bar",
			expectedOutput: nil,
			expectedError:  true,
		},
		{
			desc:   "should strip additional quotes",
			labels: "\"key1:value1,key2:value2\"",
			expectedOutput: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			expectedError: false,
		},
	}

	for i, c := range testCases {
		m, err := convertLabelsToMap(c.labels)
		if c.expectedError {
			assert.NotNil(t, err, "TestCase[%d]: %s", i, c.desc)
		} else {
			assert.Nil(t, err, "TestCase[%d]: %s", i, c.desc)
			assert.Equal(t, m, c.expectedOutput, "expected labels map")
		}
	}
}

func writeStaticCertFiles(t *testing.T) config.CertsConfig {
	tempDir := t.TempDir()
	caCert := &x509.Certificate{
		SerialNumber: big.NewInt(0),
		Subject: pkix.Name{
			Organization: []string{"ca"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(2, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	caKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		t.Fatal(err)
	}
	pub, priv := generateCerts(t, "server", caCert, caKey)

	certPath := path.Join(tempDir, "cert.crt")
	if err := os.WriteFile(certPath, pub, 0666); err != nil {
		t.Fatal(err)
	}
	keyPath := path.Join(tempDir, "cert.key")
	if err := os.WriteFile(keyPath, priv, 0666); err != nil {
		t.Fatal(err)
	}

	return config.CertsConfig{TlsCertFile: certPath, TlsPrivateKey: keyPath}
}

func TestConfigTLSMinVersion(t *testing.T) {
	cfg := writeStaticCertFiles(t)

	testCases := []struct {
		minTlsVersion string
		wantVersion   uint16
	}{
		{"", tls.VersionTLS12},
		{"tls1_2", tls.VersionTLS12},
		{"tls1_3", tls.VersionTLS13},
	}

	for _, c := range testCases {
		tlsConfig := configTLS(cfg, c.minTlsVersion, "", nil, nil)
		assert.Equal(t, c.wantVersion, tlsConfig.MinVersion, "minTlsVersion %q", c.minTlsVersion)
		assert.Len(t, tlsConfig.Certificates, 1, "minTlsVersion %q", c.minTlsVersion)
	}
}

func TestConfigTLSCiphers(t *testing.T) {
	cfg := writeStaticCertFiles(t)

	tlsConfig := configTLS(cfg, "tls1_2", "TLS_AES_128_GCM_SHA256:not-a-real-cipher", nil, nil)
	assert.Equal(t, []uint16{tls.TLS_AES_128_GCM_SHA256}, tlsConfig.CipherSuites)

	tlsConfig = configTLS(cfg, "tls1_2", "not-a-real-cipher", nil, nil)
	assert.Nil(t, tlsConfig.CipherSuites, "unknown ciphers should be dropped, leaving the suite list empty")

	tlsConfig = configTLS(cfg, "tls1_2", "", nil, nil)
	assert.Nil(t, tlsConfig.CipherSuites, "an empty ciphers flag should leave the suite list unset")
}
