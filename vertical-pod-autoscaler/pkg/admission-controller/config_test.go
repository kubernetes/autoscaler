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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	admissionregistration "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestSelfRegistrationBase(t *testing.T) {

	testClientSet := fake.NewSimpleClientset()
	caCert := []byte("fake")
	webHookDelay := 0 * time.Second
	namespace := "default"
	serviceName := "vpa-service"
	url := "http://example.com/"
	registerByURL := true
	timeoutSeconds := int32(32)
	selectedNamespace := ""
	ignoredNamespaces := []string{}

	selfRegistration(testClientSet, caCert, webHookDelay, namespace, serviceName, url, registerByURL, timeoutSeconds, selectedNamespace, ignoredNamespaces, false)

	webhookConfigInterface := testClientSet.AdmissionregistrationV1().MutatingWebhookConfigurations()
	webhookConfig, err := webhookConfigInterface.Get(context.TODO(), webhookConfigName, metav1.GetOptions{})

	assert.NoError(t, err, "expected no error fetching webhook configuration")
	assert.Equal(t, webhookConfigName, webhookConfig.Name, "expected webhook configuration name to match")

	assert.Len(t, webhookConfig.Webhooks, 1, "expected one webhook configuration")
	webhook := webhookConfig.Webhooks[0]
	assert.Equal(t, "vpa.k8s.io", webhook.Name, "expected webhook name to match")

	PodRule := webhook.Rules[0]
	assert.Equal(t, []admissionregistration.OperationType{admissionregistration.Create}, PodRule.Operations, "expected operations to match")
	assert.Equal(t, []string{""}, PodRule.APIGroups, "expected API groups to match")
	assert.Equal(t, []string{"v1"}, PodRule.APIVersions, "expected API versions to match")
	assert.Equal(t, []string{"pods"}, PodRule.Resources, "expected resources to match")

	VPARule := webhook.Rules[1]
	assert.Equal(t, []admissionregistration.OperationType{admissionregistration.Create, admissionregistration.Update}, VPARule.Operations, "expected operations to match")
	assert.Equal(t, []string{"autoscaling.k8s.io"}, VPARule.APIGroups, "expected API groups to match")
	assert.Equal(t, []string{"*"}, VPARule.APIVersions, "ehook.Rulxpected API versions to match")
	assert.Equal(t, []string{"verticalpodautoscalers"}, VPARule.Resources, "expected resources to match")

	assert.Equal(t, admissionregistration.SideEffectClassNone, *webhook.SideEffects, "expected side effects to match")
	assert.Equal(t, admissionregistration.Ignore, *webhook.FailurePolicy, "expected failure policy to match")
	assert.Equal(t, caCert, webhook.ClientConfig.CABundle, "expected CA bundle to match")
	assert.Equal(t, timeoutSeconds, *webhook.TimeoutSeconds, "expected timeout seconds to match")
}

func TestSelfRegistrationWithURL(t *testing.T) {

	testClientSet := fake.NewSimpleClientset()
	caCert := []byte("fake")
	webHookDelay := 0 * time.Second
	namespace := "default"
	serviceName := "vpa-service"
	url := "http://example.com/"
	registerByURL := true
	timeoutSeconds := int32(32)
	selectedNamespace := ""
	ignoredNamespaces := []string{}

	selfRegistration(testClientSet, caCert, webHookDelay, namespace, serviceName, url, registerByURL, timeoutSeconds, selectedNamespace, ignoredNamespaces, false)

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

	testClientSet := fake.NewSimpleClientset()
	caCert := []byte("fake")
	webHookDelay := 0 * time.Second
	namespace := "default"
	serviceName := "vpa-service"
	url := "http://example.com/"
	registerByURL := false
	timeoutSeconds := int32(32)
	selectedNamespace := ""
	ignoredNamespaces := []string{}

	selfRegistration(testClientSet, caCert, webHookDelay, namespace, serviceName, url, registerByURL, timeoutSeconds, selectedNamespace, ignoredNamespaces, false)

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

	testClientSet := fake.NewSimpleClientset()
	caCert := []byte("fake")
	webHookDelay := 0 * time.Second
	namespace := "default"
	serviceName := "vpa-service"
	url := "http://example.com/"
	registerByURL := false
	timeoutSeconds := int32(32)
	selectedNamespace := ""
	ignoredNamespaces := []string{"test"}

	selfRegistration(testClientSet, caCert, webHookDelay, namespace, serviceName, url, registerByURL, timeoutSeconds, selectedNamespace, ignoredNamespaces, false)

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

	testClientSet := fake.NewSimpleClientset()
	caCert := []byte("fake")
	webHookDelay := 0 * time.Second
	namespace := "default"
	serviceName := "vpa-service"
	url := "http://example.com/"
	registerByURL := false
	timeoutSeconds := int32(32)
	selectedNamespace := "test"
	ignoredNamespaces := []string{}

	selfRegistration(testClientSet, caCert, webHookDelay, namespace, serviceName, url, registerByURL, timeoutSeconds, selectedNamespace, ignoredNamespaces, false)

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

	testClientSet := fake.NewSimpleClientset()
	caCert := []byte("fake")
	webHookDelay := 0 * time.Second
	namespace := "default"
	serviceName := "vpa-service"
	url := "http://example.com/"
	registerByURL := false
	timeoutSeconds := int32(32)
	selectedNamespace := "test"
	ignoredNamespaces := []string{}

	selfRegistration(testClientSet, caCert, webHookDelay, namespace, serviceName, url, registerByURL, timeoutSeconds, selectedNamespace, ignoredNamespaces, true)

	webhookConfigInterface := testClientSet.AdmissionregistrationV1().MutatingWebhookConfigurations()
	webhookConfig, err := webhookConfigInterface.Get(context.TODO(), webhookConfigName, metav1.GetOptions{})

	assert.NoError(t, err, "expected no error fetching webhook configuration")

	assert.Len(t, webhookConfig.Webhooks, 1, "expected one webhook configuration")
	webhook := webhookConfig.Webhooks[0]

	assert.NotNil(t, *webhook.FailurePolicy, "expected failurePolicy not to be nil")
	assert.Equal(t, *webhook.FailurePolicy, admissionregistration.Fail, "expected failurePolicy to be Fail")
}

func TestSelfRegistrationWithOutFailurePolicy(t *testing.T) {

	testClientSet := fake.NewSimpleClientset()
	caCert := []byte("fake")
	webHookDelay := 0 * time.Second
	namespace := "default"
	serviceName := "vpa-service"
	url := "http://example.com/"
	registerByURL := false
	timeoutSeconds := int32(32)
	selectedNamespace := "test"
	ignoredNamespaces := []string{}

	selfRegistration(testClientSet, caCert, webHookDelay, namespace, serviceName, url, registerByURL, timeoutSeconds, selectedNamespace, ignoredNamespaces, false)

	webhookConfigInterface := testClientSet.AdmissionregistrationV1().MutatingWebhookConfigurations()
	webhookConfig, err := webhookConfigInterface.Get(context.TODO(), webhookConfigName, metav1.GetOptions{})

	assert.NoError(t, err, "expected no error fetching webhook configuration")

	assert.Len(t, webhookConfig.Webhooks, 1, "expected one webhook configuration")
	webhook := webhookConfig.Webhooks[0]

	assert.NotNil(t, *webhook.FailurePolicy, "expected namespace selector not to be nil")
	assert.Equal(t, *webhook.FailurePolicy, admissionregistration.Ignore, "expected failurePolicy to be Ignore")
}
