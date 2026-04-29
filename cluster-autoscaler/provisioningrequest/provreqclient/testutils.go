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

package provreqclient

import (
	"context"
	"fmt"
	"testing"
	"time"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/provisioningrequest/autoscaling.x-k8s.io/v1"
	"k8s.io/autoscaler/cluster-autoscaler/apis/provisioningrequest/client/clientset/versioned/fake"
	"k8s.io/autoscaler/cluster-autoscaler/apis/provisioningrequest/client/informers/externalversions"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqwrapper"
	"k8s.io/client-go/informers"
	fake_kubernetes "k8s.io/client-go/kubernetes/fake"
)

// NewFakeProvisioningRequestClient mock ProvisioningRequestClient for tests.
func NewFakeProvisioningRequestClient(ctx context.Context, t *testing.T, prs ...*provreqwrapper.ProvisioningRequest) *ProvisioningRequestClient {
	t.Helper()
	provReqClient := fake.NewSimpleClientset()
	podTemplClient := fake_kubernetes.NewSimpleClientset()
	for _, pr := range prs {
		if pr == nil {
			continue
		}
		if _, err := provReqClient.AutoscalingV1().ProvisioningRequests(pr.Namespace).Create(ctx, pr.ProvisioningRequest, metav1.CreateOptions{}); err != nil {
			t.Errorf("While adding a ProvisioningRequest: %s/%s to fake client, got error: %v", pr.Namespace, pr.Name, err)
		}
		for _, pd := range pr.PodTemplates {
			if _, err := podTemplClient.CoreV1().PodTemplates(pr.Namespace).Create(ctx, pd, metav1.CreateOptions{}); err != nil {
				t.Errorf("While adding a PodTemplate: %s/%s to fake client, got error: %v", pr.Namespace, pd.Name, err)
			}
		}
	}
	prFactory := externalversions.NewSharedInformerFactory(provReqClient, 1*time.Hour)
	provReqLister := prFactory.Autoscaling().V1().ProvisioningRequests().Lister()
	prFactory.Start(ctx.Done())

	podFactory := informers.NewSharedInformerFactory(podTemplClient, 1*time.Hour)
	podTemplLister := podFactory.Core().V1().PodTemplates().Lister()
	podFactory.Start(ctx.Done())

	informersSynced := prFactory.WaitForCacheSync(ctx.Done())
	for _, synced := range informersSynced {
		if !synced {
			t.Fatalf("Failed to sync Provisioning Request informers")
		}
	}

	podInformersSynced := podFactory.WaitForCacheSync(ctx.Done())
	for _, synced := range podInformersSynced {
		if !synced {
			t.Fatalf("Failed to sync Pod Template informers")
		}
	}

	return NewProvisioningRequestClient(
		provReqClient,
		provReqLister,
		podTemplLister,
	)
}

// ProvisioningRequestWrapperForTesting mock ProvisioningRequest for tests.
func ProvisioningRequestWrapperForTesting(namespace, name string) *provreqwrapper.ProvisioningRequest {
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
	v1PR := &v1.ProvisioningRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1.ProvisioningRequestSpec{
			ProvisioningClassName: "test-class",
			PodSets: []v1.PodSet{
				{
					Count: 1,
					PodTemplateRef: v1.Reference{
						Name: podTemplates[0].Name,
					},
				},
			},
		},
		Status: v1.ProvisioningRequestStatus{
			ProvisioningClassDetails: map[string]v1.Detail{},
		},
	}

	pr := provreqwrapper.NewProvisioningRequest(v1PR, podTemplates)
	return pr
}

func podTemplateNameFromName(name string) string {
	return fmt.Sprintf("%s-pod-template", name)
}

// ProvisioningRequestNoCache returns ProvisioningRequest directly from client. For test purposes only.
func (c *ProvisioningRequestClient) ProvisioningRequestNoCache(namespace, name string) (*provreqwrapper.ProvisioningRequest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), provisioningRequestClientCallTimeout)
	defer cancel()
	v1, err := c.client.AutoscalingV1().ProvisioningRequests(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	podTemplates, err := c.FetchPodTemplates(v1)
	if err != nil {
		return nil, err
	}
	return provreqwrapper.NewProvisioningRequest(v1, podTemplates), nil
}

// ProvisioningRequestsNoCache returns all ProvisioningRequests directly from client. For test purposes only.
func (c *ProvisioningRequestClient) ProvisioningRequestsNoCache() ([]*provreqwrapper.ProvisioningRequest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), provisioningRequestClientCallTimeout)
	defer cancel()
	v1s, err := c.client.AutoscalingV1().ProvisioningRequests("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	prs := make([]*provreqwrapper.ProvisioningRequest, 0, len(v1s.Items))
	for _, v1 := range v1s.Items {
		podTemplates, err := c.FetchPodTemplates(&v1)
		if err != nil {
			return nil, err
		}
		prs = append(prs, provreqwrapper.NewProvisioningRequest(&v1, podTemplates))
	}
	return prs, nil
}
