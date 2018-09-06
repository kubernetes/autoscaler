/*
Copyright 2019 The Kubernetes Authors.

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

package openstack

import (
	"regexp"
	"time"

    "github.com/gophercloud/gophercloud"
    "github.com/gophercloud/gophercloud/openstack"
)

const (
	defaultOperationWaitTimeout  = 5 * time.Second
	defaultOperationPollInterval = 100 * time.Millisecond
)

// AutoscalingOpenStackClient is used for communicating with OpenStack Go SDK (gophercloud).
type AutoscalingOpenStackClient interface {
	// reading resources
	FetchASGTargetSize(OpenStackRef) (int64, error)
    FetchASGBasename(OpenStackRef) (string, error)
	FetchASGInstances(OpenStackRef) ([]OpenStackRef, error)
	FetchASGsWithName(filter *regexp.Regexp) ([]string, error)

	// modifying resources
	ResizeASG(OpenStackRef, int64) error
	DeleteInstances(asgRef OpenStackRef, instances []*OpenStackRef) error
}

// NewAutoscalingOrchestrationClient creates a new client for communicating with Heat service through OpenStack go sdk.
func NewAutoscalingOrchestrationClient(authOpts gophercloud.AuthOptions,  endpointOpts gophercloud.EndpointOpts) (*autoscalingOrchestrationClient, error) {
    var auth_provider, err = openstack.AuthenticatedClient(authOpts)
	if err != nil {
		return nil, err
	}
    orchestrationClient, err := openstack.NewOrchestrationV1(auth_provider,  endpointOpts)
	if err != nil {
		return nil, err
	}

	return &autoscalingOrchestrationClient{
		projectId:             authOpts.TenantID,
		openstackService:      orchestrationClient,
		operationWaitTimeout:  defaultOperationWaitTimeout,
		operationPollInterval: defaultOperationPollInterval,
	}, nil
}


// NewMagnumClient creates a new client for communicating with Magnum service through OpenStack go sdk.
func NewMagnumClient(authOpts gophercloud.AuthOptions,  endpointOpts gophercloud.EndpointOpts) (*MagnumClient, error) {
    var auth_provider, err = openstack.AuthenticatedClient(authOpts)
	if err != nil {
		return nil, err
	}
    containerInfraClient, err := openstack.NewContainerInfraV1(auth_provider,  endpointOpts)
	if err != nil {
		return nil, err
	}

	return &MagnumClient{
		projectId:             authOpts.TenantID,
		openstackService:      containerInfraClient,
		operationWaitTimeout:  defaultOperationWaitTimeout,
		operationPollInterval: defaultOperationPollInterval,
	}, nil
}
