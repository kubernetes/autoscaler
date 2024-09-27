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
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/labels"
	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/provisioningrequest/autoscaling.x-k8s.io/v1"
	"k8s.io/autoscaler/cluster-autoscaler/apis/provisioningrequest/client/clientset/versioned"
	"k8s.io/autoscaler/cluster-autoscaler/apis/provisioningrequest/client/informers/externalversions"
	listers "k8s.io/autoscaler/cluster-autoscaler/apis/provisioningrequest/client/listers/autoscaling.x-k8s.io/v1"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqwrapper"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"

	klog "k8s.io/klog/v2"
)

const (
	provisioningRequestClientCallTimeout = 4 * time.Second
)

// ProvisioningRequestClient represents client for v1 ProvReq CRD.
type ProvisioningRequestClient struct {
	client         versioned.Interface
	provReqLister  listers.ProvisioningRequestLister
	podTemplLister corev1.PodTemplateLister
}

// NewProvisioningRequestClient configures and returns a provisioningRequestClient.
func NewProvisioningRequestClient(kubeConfig *rest.Config) (*ProvisioningRequestClient, error) {
	prClient, err := newPRClient(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Provisioning Request client: %v", err)
	}

	provReqLister, err := newPRsLister(prClient, make(chan struct{}))
	if err != nil {
		return nil, err
	}

	podTemplateClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Pod Template client: %v", err)
	}

	podTemplLister, err := newPodTemplatesLister(podTemplateClient, make(chan struct{}))
	if err != nil {
		return nil, err
	}

	return &ProvisioningRequestClient{
		client:         prClient,
		provReqLister:  provReqLister,
		podTemplLister: podTemplLister,
	}, nil
}

// ProvisioningRequest gets a specific ProvisioningRequest CR.
func (c *ProvisioningRequestClient) ProvisioningRequest(namespace, name string) (*provreqwrapper.ProvisioningRequest, error) {
	v1PR, err := c.provReqLister.ProvisioningRequests(namespace).Get(name)
	if err != nil {
		return nil, err
	}
	podTemplates, err := c.FetchPodTemplates(v1PR)
	if err != nil {
		return nil, fmt.Errorf("while fetching pod templates for Get Provisioning Request %s/%s got error: %v", namespace, name, err)
	}
	return provreqwrapper.NewProvisioningRequest(v1PR, podTemplates), nil
}

// ProvisioningRequests gets all ProvisioningRequest CRs.
func (c *ProvisioningRequestClient) ProvisioningRequests() ([]*provreqwrapper.ProvisioningRequest, error) {
	v1PRs, err := c.provReqLister.List(labels.Everything())
	if err != nil {
		return nil, fmt.Errorf("error fetching provisioningRequests: %w", err)
	}
	prs := make([]*provreqwrapper.ProvisioningRequest, 0, len(v1PRs))
	for _, v1PR := range v1PRs {
		podTemplates, errPodTemplates := c.FetchPodTemplates(v1PR)
		if errPodTemplates != nil {
			return nil, fmt.Errorf("while fetching pod templates for List Provisioning Request %s/%s got error: %v", v1PR.Namespace, v1PR.Name, errPodTemplates)
		}
		prs = append(prs, provreqwrapper.NewProvisioningRequest(v1PR, podTemplates))
	}
	return prs, nil
}

// FetchPodTemplates fetches PodTemplates referenced by the Provisioning Request.
func (c *ProvisioningRequestClient) FetchPodTemplates(pr *v1.ProvisioningRequest) ([]*apiv1.PodTemplate, error) {
	podTemplates := make([]*apiv1.PodTemplate, 0, len(pr.Spec.PodSets))
	for _, podSpec := range pr.Spec.PodSets {
		podTemplate, err := c.podTemplLister.PodTemplates(pr.Namespace).Get(podSpec.PodTemplateRef.Name)
		if errors.IsNotFound(err) {
			klog.Infof("While fetching Pod Template for Provisioning Request %s/%s received not found error", pr.Namespace, pr.Name)
			continue
		} else if err != nil {
			return nil, err
		}
		podTemplates = append(podTemplates, podTemplate)
	}
	return podTemplates, nil
}

// UpdateProvisioningRequest updates the given ProvisioningRequest CR by propagating the changes using the ProvisioningRequestInterface and returns the updated instance or the original one in case of an error.
func (c *ProvisioningRequestClient) UpdateProvisioningRequest(pr *v1.ProvisioningRequest) (*v1.ProvisioningRequest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), provisioningRequestClientCallTimeout)
	defer cancel()

	updatedPr, err := c.client.AutoscalingV1().ProvisioningRequests(pr.Namespace).UpdateStatus(ctx, pr, metav1.UpdateOptions{})
	if err != nil {
		return pr, err
	}
	klog.V(4).Infof("Updated ProvisioningRequest %s/%s,  status: %q,", updatedPr.Namespace, updatedPr.Name, updatedPr.Status)
	return updatedPr, nil
}

// newPRClient creates a new Provisioning Request client from the given config.
func newPRClient(kubeConfig *rest.Config) (*versioned.Clientset, error) {
	return versioned.NewForConfig(kubeConfig)
}

// newPRsLister creates a lister for the Provisioning Requests in the cluster.
func newPRsLister(prClient versioned.Interface, stopChannel <-chan struct{}) (listers.ProvisioningRequestLister, error) {
	factory := externalversions.NewSharedInformerFactory(prClient, 1*time.Hour)
	provReqLister := factory.Autoscaling().V1().ProvisioningRequests().Lister()
	factory.Start(stopChannel)
	informersSynced := factory.WaitForCacheSync(stopChannel)
	for _, synced := range informersSynced {
		if !synced {
			return nil, fmt.Errorf("can't create Provisioning Request lister")
		}
	}
	klog.V(2).Info("Successful initial Provisioning Request sync")
	return provReqLister, nil
}

// newPodTemplatesLister creates a lister for the Pod Templates in the cluster.
func newPodTemplatesLister(client *kubernetes.Clientset, stopChannel <-chan struct{}) (corev1.PodTemplateLister, error) {
	factory := informers.NewSharedInformerFactory(client, 1*time.Hour)
	podTemplLister := factory.Core().V1().PodTemplates().Lister()
	factory.Start(stopChannel)
	informersSynced := factory.WaitForCacheSync(stopChannel)
	for _, synced := range informersSynced {
		if !synced {
			return nil, fmt.Errorf("can't create Pod Template lister")
		}
	}
	klog.V(2).Info("Successful initial Pod Template sync")
	return podTemplLister, nil
}

// ProvisioningRequestsForPods check that all pods belong to one ProvisioningRequest and return it.
func ProvisioningRequestsForPods(client *ProvisioningRequestClient, unschedulablePods []*apiv1.Pod) []*provreqwrapper.ProvisioningRequest {
	prMap := make(map[string]*provreqwrapper.ProvisioningRequest)
	prList := []*provreqwrapper.ProvisioningRequest{}
	if len(unschedulablePods) == 0 {
		return prList
	}
	for _, pod := range unschedulablePods {
		if len(pod.OwnerReferences) == 0 {
			klog.Errorf("pod %s has no OwnerReference", pod.Name)
			continue
		}
		provReq, err := client.ProvisioningRequest(pod.Namespace, pod.OwnerReferences[0].Name)
		if err != nil {
			klog.Errorf("failed to retrieve ProvisioningRequest from unschedulable pod, err: %v", err)
			continue
		}
		if _, found := prMap[provReq.Name]; !found {
			prMap[provReq.Name] = provReq
		}
	}
	for _, pr := range prMap {
		prList = append(prList, pr)
	}
	return prList
}

// SegregatePodsByProvisioningRequest segregates pods by ProvisioningRequest.
func SegregatePodsByProvisioningRequest(client *ProvisioningRequestClient, pods []*apiv1.Pod) map[string][]*apiv1.Pod {
	podsByPR := make(map[string][]*apiv1.Pod)
	for _, pod := range pods {
		if len(pod.OwnerReferences) == 0 {
			klog.Errorf("pod %s has no OwnerReference", pod.Name)
			continue
		}

		provReq := fmt.Sprintf("%s/%s", pod.Namespace, pod.OwnerReferences[0].Name)
		podsByPR[provReq] = append(podsByPR[provReq], pod)
	}
	return podsByPR
}

// DeleteProvisioningRequest deletes the given ProvisioningRequest CR using the ProvisioningRequestInterface and returns an error in case of failure.
func (c *ProvisioningRequestClient) DeleteProvisioningRequest(pr *v1.ProvisioningRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), provisioningRequestClientCallTimeout)
	defer cancel()

	err := c.client.AutoscalingV1().ProvisioningRequests(pr.Namespace).Delete(ctx, pr.Name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("error deleting ProvisioningRequest %s/%s: %w", pr.Namespace, pr.Name, err)
	}
	klog.V(4).Infof("Deleted ProvisioningRequest %s/%s", pr.Namespace, pr.Name)
	return nil
}

// FilterOutProvisioningClass filters out ProvReqs that belongs to certain Provisioning Class
func FilterOutProvisioningClass(prList []*provreqwrapper.ProvisioningRequest, class string) []*provreqwrapper.ProvisioningRequest {
	newPrList := []*provreqwrapper.ProvisioningRequest{}
	for _, pr := range prList {
		if pr.Spec.ProvisioningClassName == class {
			newPrList = append(newPrList, pr)
		}
	}
	return newPrList
}
