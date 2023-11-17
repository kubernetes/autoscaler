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

package provreqservice

import (
	"fmt"

	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqwrapper"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/service/v1beta1client"
	"k8s.io/client-go/rest"
)

// ProvisioningRequestService represents the service that is able to list,
// access and delete different Provisioning Requests.
type ProvisioningRequestService struct {
	provReqV1Beta1Client *v1beta1client.ProvisioningRequestClient
}

// NewProvisioningRequestService returns new service for interacting with ProvisioningRequests.
func NewProvisioningRequestService(kubeConfig *rest.Config) (*ProvisioningRequestService, error) {
	v1Beta1Client, err := v1beta1client.NewProvisioningRequestClient(kubeConfig)
	if err != nil {
		return nil, err
	}
	return &ProvisioningRequestService{
		provReqV1Beta1Client: v1Beta1Client,
	}, nil
}

// ProvisioningRequest gets a specific ProvisioningRequest CR.
func (s *ProvisioningRequestService) ProvisioningRequest(namespace, name string) (*provreqwrapper.ProvisioningRequest, error) {
	v1Beta1PR, err := s.provReqV1Beta1Client.ProvisioningRequest(namespace, name)
	if err == nil {
		podTemplates, errPodTemplates := s.provReqV1Beta1Client.FetchPodTemplates(v1Beta1PR)
		if errPodTemplates != nil {
			return nil, fmt.Errorf("while fetching pod templates for Get Provisioning Request %s/%s got error: %v", namespace, name, errPodTemplates)
		}
		return provreqwrapper.NewV1Beta1ProvisioningRequest(v1Beta1PR, podTemplates), nil
	}
	return nil, err
}

// ProvisioningRequests gets all Queued ProvisioningRequest CRs.
func (s *ProvisioningRequestService) ProvisioningRequests() ([]*provreqwrapper.ProvisioningRequest, error) {
	v1Beta1PRs, err := s.provReqV1Beta1Client.ProvisioningRequests()
	if err != nil {
		return nil, err
	}
	prs := make([]*provreqwrapper.ProvisioningRequest, 0, len(v1Beta1PRs))
	for _, v1Beta1PR := range v1Beta1PRs {
		podTemplates, errPodTemplates := s.provReqV1Beta1Client.FetchPodTemplates(v1Beta1PR)
		if errPodTemplates != nil {
			return nil, fmt.Errorf("while fetching pod templates for List Provisioning Request %s/%s got error: %v", v1Beta1PR.Namespace, v1Beta1PR.Name, errPodTemplates)
		}
		prs = append(prs, provreqwrapper.NewV1Beta1ProvisioningRequest(v1Beta1PR, podTemplates))
	}
	return prs, nil
}
