/*
Copyright 2020 The Kubernetes Authors.

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

// PARTIAL COPY OF https://github.com/kubernetes/kubernetes/blob/master/test/e2e/apimachinery/webhook.go

package utils

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/onsi/ginkgo"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/test/e2e/framework"
	e2edeploy "k8s.io/kubernetes/test/e2e/framework/deployment"
	"k8s.io/utils/pointer"
)

const (
	// WebhookServiceName is the webhook service name.
	WebhookServiceName = "e2e-test-webhook"

	roleBindingName = "webhook-auth-reader"
	secretName      = "sample-webhook-secret"
	deploymentName  = "sample-webhook-deployment"
)

func strPtr(s string) *string { return &s }

// LabelNamespace applies unique label to the namespace.
func LabelNamespace(f *framework.Framework, namespace string) {
	client := f.ClientSet

	// Add a unique label to the namespace
	ns, err := client.CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{})
	framework.ExpectNoError(err, "error getting namespace %s", namespace)
	if ns.Labels == nil {
		ns.Labels = map[string]string{}
	}
	ns.Labels[f.UniqueName] = "true"
	_, err = client.CoreV1().Namespaces().Update(context.TODO(), ns, metav1.UpdateOptions{})
	framework.ExpectNoError(err, "error labeling namespace %s", namespace)
}

// CreateWebhookConfigurationReadyNamespace creates a separate namespace for webhook configuration ready markers to
// prevent cross-talk with webhook configurations being tested.
func CreateWebhookConfigurationReadyNamespace(f *framework.Framework) {
	ns, err := f.ClientSet.CoreV1().Namespaces().Create(context.TODO(), &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   f.Namespace.Name + "-markers",
			Labels: map[string]string{f.UniqueName + "-markers": "true"},
		},
	}, metav1.CreateOptions{})
	framework.ExpectNoError(err, "creating namespace for webhook configuration ready markers")
	f.AddNamespacesToDelete(ns)
}

// RegisterMutatingWebhookForPod creates mutation webhook configuration
// and applies it to the cluster.
func RegisterMutatingWebhookForPod(f *framework.Framework, configName string, certContext *certContext, servicePort int32) func() {
	client := f.ClientSet
	ginkgo.By("Registering the mutating pod webhook via the AdmissionRegistration API")

	namespace := f.Namespace.Name
	sideEffectsNone := admissionregistrationv1beta1.SideEffectClassNone

	_, err := createMutatingWebhookConfiguration(f, &admissionregistrationv1beta1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: configName,
		},
		Webhooks: []admissionregistrationv1beta1.MutatingWebhook{
			{
				Name: "adding-init-container.k8s.io",
				Rules: []admissionregistrationv1beta1.RuleWithOperations{{
					Operations: []admissionregistrationv1beta1.OperationType{admissionregistrationv1beta1.Create},
					Rule: admissionregistrationv1beta1.Rule{
						APIGroups:   []string{""},
						APIVersions: []string{"v1"},
						Resources:   []string{"pods"},
					},
				}},
				ClientConfig: admissionregistrationv1beta1.WebhookClientConfig{
					Service: &admissionregistrationv1beta1.ServiceReference{
						Namespace: namespace,
						Name:      WebhookServiceName,
						Path:      strPtr("/mutating-pods-sidecar"),
						Port:      pointer.Int32Ptr(servicePort),
					},
					CABundle: certContext.signingCert,
				},
				SideEffects:             &sideEffectsNone,
				AdmissionReviewVersions: []string{"v1", "v1beta1"},
				// Scope the webhook to just this namespace
				NamespaceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{f.UniqueName: "true"},
				},
			},
			// Register a webhook that can be probed by marker requests to detect when the configuration is ready.
			newMutatingIsReadyWebhookFixture(f, certContext, servicePort),
		},
	})
	framework.ExpectNoError(err, "registering mutating webhook config %s with namespace %s", configName, namespace)

	err = waitWebhookConfigurationReady(f)
	framework.ExpectNoError(err, "waiting for webhook configuration to be ready")

	return func() {
		client.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Delete(context.TODO(), configName, metav1.DeleteOptions{})
	}
}

// createMutatingWebhookConfiguration ensures the webhook config scopes object or namespace selection
// to avoid interfering with other tests, then creates the config.
func createMutatingWebhookConfiguration(f *framework.Framework, config *admissionregistrationv1beta1.MutatingWebhookConfiguration) (*admissionregistrationv1beta1.MutatingWebhookConfiguration, error) {
	for _, webhook := range config.Webhooks {
		if webhook.NamespaceSelector != nil && webhook.NamespaceSelector.MatchLabels[f.UniqueName] == "true" {
			continue
		}
		if webhook.ObjectSelector != nil && webhook.ObjectSelector.MatchLabels[f.UniqueName] == "true" {
			continue
		}
		framework.Failf(`webhook %s in config %s has no namespace or object selector with %s="true", and can interfere with other tests`, webhook.Name, config.Name, f.UniqueName)
	}
	return f.ClientSet.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Create(context.TODO(), config, metav1.CreateOptions{})
}

// newMutatingIsReadyWebhookFixture creates a mutating webhook that can be added to a webhook configuration and then probed
// with "marker" requests via waitWebhookConfigurationReady to wait for a webhook configuration to be ready.
func newMutatingIsReadyWebhookFixture(f *framework.Framework, certContext *certContext, servicePort int32) admissionregistrationv1beta1.MutatingWebhook {
	sideEffectsNone := admissionregistrationv1beta1.SideEffectClassNone
	failOpen := admissionregistrationv1beta1.Ignore
	return admissionregistrationv1beta1.MutatingWebhook{
		Name: "mutating-is-webhook-configuration-ready.k8s.io",
		Rules: []admissionregistrationv1beta1.RuleWithOperations{{
			Operations: []admissionregistrationv1beta1.OperationType{admissionregistrationv1beta1.Create},
			Rule: admissionregistrationv1beta1.Rule{
				APIGroups:   []string{""},
				APIVersions: []string{"v1"},
				Resources:   []string{"configmaps"},
			},
		}},
		ClientConfig: admissionregistrationv1beta1.WebhookClientConfig{
			Service: &admissionregistrationv1beta1.ServiceReference{
				Namespace: f.Namespace.Name,
				Name:      WebhookServiceName,
				Path:      strPtr("/always-deny"),
				Port:      pointer.Int32Ptr(servicePort),
			},
			CABundle: certContext.signingCert,
		},
		// network failures while the service network routing is being set up should be ignored by the marker
		FailurePolicy:           &failOpen,
		SideEffects:             &sideEffectsNone,
		AdmissionReviewVersions: []string{"v1", "v1beta1"},
		// Scope the webhook to just the markers namespace
		NamespaceSelector: &metav1.LabelSelector{
			MatchLabels: map[string]string{f.UniqueName + "-markers": "true"},
		},
		// appease createMutatingWebhookConfiguration isolation requirements
		ObjectSelector: &metav1.LabelSelector{
			MatchLabels: map[string]string{f.UniqueName: "true"},
		},
	}
}

// waitWebhookConfigurationReady sends "marker" requests until a webhook configuration is ready.
// A webhook created with newValidatingIsReadyWebhookFixture or newMutatingIsReadyWebhookFixture should first be added to
// the webhook configuration.
func waitWebhookConfigurationReady(f *framework.Framework) error {
	cmClient := f.ClientSet.CoreV1().ConfigMaps(f.Namespace.Name + "-markers")
	return wait.PollImmediate(100*time.Millisecond, 30*time.Second, func() (bool, error) {
		marker := &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name: string(uuid.NewUUID()),
				Labels: map[string]string{
					f.UniqueName: "true",
				},
			},
		}
		_, err := cmClient.Create(context.TODO(), marker, metav1.CreateOptions{})
		if err != nil {
			// The always-deny webhook does not provide a reason, so check for the error string we expect
			if strings.Contains(err.Error(), "denied") {
				return true, nil
			}
			return false, err
		}
		// best effort cleanup of markers that are no longer needed
		_ = cmClient.Delete(context.TODO(), marker.GetName(), metav1.DeleteOptions{})
		framework.Logf("Waiting for webhook configuration to be ready...")
		return false, nil
	})
}

// CreateAuthReaderRoleBinding creates the role binding to allow the webhook read
// the extension-apiserver-authentication configmap.
func CreateAuthReaderRoleBinding(f *framework.Framework, namespace string) {
	ginkgo.By("Create role binding to let webhook read extension-apiserver-authentication")
	client := f.ClientSet
	_, err := client.RbacV1().RoleBindings("kube-system").Create(context.TODO(), &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: roleBindingName,
			Annotations: map[string]string{
				rbacv1.AutoUpdateAnnotationKey: "true",
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "",
			Kind:     "Role",
			Name:     "extension-apiserver-authentication-reader",
		},
		// Webhook uses the default service account.
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "default",
				Namespace: namespace,
			},
		},
	}, metav1.CreateOptions{})
	if err != nil && apierrors.IsAlreadyExists(err) {
		framework.Logf("role binding %s already exists", roleBindingName)
	} else {
		framework.ExpectNoError(err, "creating role binding %s:webhook to access configMap", namespace)
	}
}

// DeployWebhookAndService creates a webhook with a corresponding service.
func DeployWebhookAndService(f *framework.Framework, image string, certContext *certContext, servicePort int32,
	containerPort int32, params ...string) {
	ginkgo.By("Deploying the webhook pod")
	client := f.ClientSet

	// Creating the secret that contains the webhook's cert.
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
		},
		Type: v1.SecretTypeOpaque,
		Data: map[string][]byte{
			"tls.crt": certContext.cert,
			"tls.key": certContext.key,
		},
	}
	namespace := f.Namespace.Name
	_, err := client.CoreV1().Secrets(namespace).Create(context.TODO(), secret, metav1.CreateOptions{})
	framework.ExpectNoError(err, "creating secret %q in namespace %q", secretName, namespace)

	// Create the deployment of the webhook
	podLabels := map[string]string{"app": "sample-webhook", "webhook": "true"}
	replicas := int32(1)
	zero := int64(0)
	mounts := []v1.VolumeMount{
		{
			Name:      "webhook-certs",
			ReadOnly:  true,
			MountPath: "/webhook.local.config/certificates",
		},
	}
	volumes := []v1.Volume{
		{
			Name: "webhook-certs",
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{SecretName: secretName},
			},
		},
	}
	containers := []v1.Container{
		{
			Name:         "sample-webhook",
			VolumeMounts: mounts,
			Args: append([]string{
				"webhook",
				"--tls-cert-file=/webhook.local.config/certificates/tls.crt",
				"--tls-private-key-file=/webhook.local.config/certificates/tls.key",
				"--alsologtostderr",
				"-v=4",
				// Use a non-default port for containers.
				fmt.Sprintf("--port=%d", containerPort),
			}, params...),
			ReadinessProbe: &v1.Probe{
				Handler: v1.Handler{
					HTTPGet: &v1.HTTPGetAction{
						Scheme: v1.URISchemeHTTPS,
						Port:   intstr.FromInt(int(containerPort)),
						Path:   "/readyz",
					},
				},
				PeriodSeconds:    1,
				SuccessThreshold: 1,
				FailureThreshold: 30,
			},
			Image: image,
			Ports: []v1.ContainerPort{{ContainerPort: containerPort}},
		},
	}
	d := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:   deploymentName,
			Labels: podLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: podLabels,
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: podLabels,
				},
				Spec: v1.PodSpec{
					TerminationGracePeriodSeconds: &zero,
					Containers:                    containers,
					Volumes:                       volumes,
				},
			},
		},
	}
	deployment, err := client.AppsV1().Deployments(namespace).Create(context.TODO(), d, metav1.CreateOptions{})
	framework.ExpectNoError(err, "creating deployment %s in namespace %s", deploymentName, namespace)
	ginkgo.By("Wait for the deployment to be ready")
	err = e2edeploy.WaitForDeploymentRevisionAndImage(client, namespace, deploymentName, "1", image)
	framework.ExpectNoError(err, "waiting for the deployment of image %s in %s in %s to complete", image, deploymentName, namespace)
	err = e2edeploy.WaitForDeploymentComplete(client, deployment)
	framework.ExpectNoError(err, "waiting for the deployment status valid", image, deploymentName, namespace)

	ginkgo.By("Deploying the webhook service")

	serviceLabels := map[string]string{"webhook": "true"}
	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      WebhookServiceName,
			Labels:    map[string]string{"test": "webhook"},
		},
		Spec: v1.ServiceSpec{
			Selector: serviceLabels,
			Ports: []v1.ServicePort{
				{
					Protocol:   "TCP",
					Port:       servicePort,
					TargetPort: intstr.FromInt(int(containerPort)),
				},
			},
		},
	}
	_, err = client.CoreV1().Services(namespace).Create(context.TODO(), service, metav1.CreateOptions{})
	framework.ExpectNoError(err, "creating service %s in namespace %s", WebhookServiceName, namespace)

	ginkgo.By("Verifying the service has paired with the endpoint")
	err = framework.WaitForServiceEndpointsNum(client, namespace, WebhookServiceName, 1, 1*time.Second, 30*time.Second)
	framework.ExpectNoError(err, "waiting for service %s/%s have %d endpoint", namespace, WebhookServiceName, 1)
}

// CleanWebhookTest cleans after a webhook test.
func CleanWebhookTest(client clientset.Interface, namespaceName string) {
	_ = client.CoreV1().Services(namespaceName).Delete(context.TODO(), WebhookServiceName, metav1.DeleteOptions{})
	_ = client.AppsV1().Deployments(namespaceName).Delete(context.TODO(), deploymentName, metav1.DeleteOptions{})
	_ = client.CoreV1().Secrets(namespaceName).Delete(context.TODO(), WebhookServiceName, metav1.DeleteOptions{})
	_ = client.RbacV1().RoleBindings("kube-system").Delete(context.TODO(), roleBindingName, metav1.DeleteOptions{})
}
