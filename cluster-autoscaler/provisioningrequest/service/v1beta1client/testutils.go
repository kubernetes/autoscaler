/*
Copyright 2023 The Kubernetes Authors.

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

package v1beta1client

import (
	"context"
	"fmt"
	"testing"
	"time"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/apis/autoscaling.x-k8s.io/v1beta1"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/client/clientset/versioned/fake"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqwrapper"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	fake_kubernetes "k8s.io/client-go/kubernetes/fake"
	v1 "k8s.io/client-go/listers/core/v1"
	klog "k8s.io/klog/v2"
)

// NewFakeProvisioningRequestClient mock ProvisioningRequestClient for tests.
func NewFakeProvisioningRequestClient(ctx context.Context, t *testing.T, prs ...*provreqwrapper.ProvisioningRequest) (*ProvisioningRequestClient, *FakeProvisioningRequestForceClient) {
	t.Helper()
	provReqClient := fake.NewSimpleClientset()
	podTemplClient := fake_kubernetes.NewSimpleClientset()
	for _, pr := range prs {
		if pr == nil {
			continue
		}
		if _, err := provReqClient.AutoscalingV1beta1().ProvisioningRequests(pr.Namespace()).Create(ctx, pr.V1Beta1(), metav1.CreateOptions{}); err != nil {
			t.Errorf("While adding a ProvisioningRequest: %s/%s to fake client, got error: %v", pr.Namespace(), pr.Name(), err)
		}
		for _, pd := range pr.PodTemplates() {
			if _, err := podTemplClient.CoreV1().PodTemplates(pr.Namespace()).Create(ctx, pd, metav1.CreateOptions{}); err != nil {
				t.Errorf("While adding a PodTemplate: %s/%s to fake client, got error: %v", pr.Namespace(), pd.Name, err)
			}
		}
	}
	provReqLister, err := newPRsLister(provReqClient, make(chan struct{}))
	if err != nil {
		t.Fatalf("Failed to create Provisioning Request lister. Error was: %v", err)
	}
	podTemplLister, err := newFakePodTemplatesLister(t, podTemplClient, make(chan struct{}))
	if err != nil {
		t.Fatalf("Failed to create Provisioning Request lister. Error was: %v", err)
	}
	return &ProvisioningRequestClient{
			client:         provReqClient,
			provReqLister:  provReqLister,
			podTemplLister: podTemplLister,
		}, &FakeProvisioningRequestForceClient{
			client: provReqClient,
		}
}

// FakeProvisioningRequestForceClient that allows to skip cache.
type FakeProvisioningRequestForceClient struct {
	client *fake.Clientset
}

// ProvisioningRequest gets a specific ProvisioningRequest CR, skipping cache.
func (c *FakeProvisioningRequestForceClient) ProvisioningRequest(namespace, name string) (*v1beta1.ProvisioningRequest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), provisioningRequestClientCallTimeout)
	defer cancel()
	return c.client.AutoscalingV1beta1().ProvisioningRequests(namespace).Get(ctx, name, metav1.GetOptions{})
}

// newFakePodTemplatesLister creates a fake lister for the Pod Templates in the cluster.
func newFakePodTemplatesLister(t *testing.T, client kubernetes.Interface, channel <-chan struct{}) (v1.PodTemplateLister, error) {
	t.Helper()
	factory := informers.NewSharedInformerFactory(client, 1*time.Hour)
	podTemplLister := factory.Core().V1().PodTemplates().Lister()
	factory.Start(channel)
	informersSynced := factory.WaitForCacheSync(channel)
	for _, synced := range informersSynced {
		if !synced {
			return nil, fmt.Errorf("can't create Pod Template lister")
		}
	}
	klog.V(2).Info("Successful initial Pod Template sync")
	return podTemplLister, nil
}

func provisioningRequestBetaForTests(namespace, name string) *provreqwrapper.ProvisioningRequest {
	if namespace == "" {
		namespace = "default"
	}
	podTemplates := []*apiv1.PodTemplate{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      podTemplateNameFromName(name),
				Namespace: namespace,
			},
			Template: apiv1.PodTemplateSpec{
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "test-container",
							Image: "test-image",
						},
					},
				},
			},
		},
	}
	v1Beta1PR := &v1beta1.ProvisioningRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1beta1.ProvisioningRequestSpec{
			ProvisioningClassName: "test-class",
			PodSets: []v1beta1.PodSet{
				{
					Count: 1,
					PodTemplateRef: v1beta1.Reference{
						Name: podTemplates[0].Name,
					},
				},
			},
		},
		Status: v1beta1.ProvisioningRequestStatus{
			ProvisioningClassDetails: map[string]v1beta1.Detail{},
		},
	}

	pr := provreqwrapper.NewV1Beta1ProvisioningRequest(v1Beta1PR, podTemplates)
	return pr
}

func podTemplateNameFromName(name string) string {
	return fmt.Sprintf("%s-pod-template", name)
}
