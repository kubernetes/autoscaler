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

package openstack

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"time"

	"github.com/golang/glog"
    "github.com/gophercloud/gophercloud/openstack"
    "github.com/gophercloud/gophercloud/openstack/orchestration/v1/stacks"
    "github.com/gophercloud/gophercloud/openstack/orchestration/v1/stackresources"
)

const (
	defaultOperationWaitTimeout  = 5 * time.Second
	defaultOperationPollInterval = 100 * time.Millisecond
)

// AutoscalingOpenStackClient is used for communicating with OpenStack Go SDK (gophercloud).
type AutoscalingOpenStackClient interface {
	// reading resources
	FetchASGTargetSize(OpenStackRef) (int64, error)
	FetchASGInstances(OpenStackRef) ([]OpenStackRef, error)
	FetchASGsWithName(filter *regexp.Regexp) ([]string, error)

	// modifying resources
	ResizeASG(OpenStackRef, int64) error
	DeleteInstances(asgRef OpenStackRef, instances []*OpenStackRef) error
}

type autoscalingOrchestrationClient struct {
	openstackService *openstack.Service

	projectId string

	// These can be overridden, e.g. for testing.
	operationWaitTimeout  time.Duration
	operationPollInterval time.Duration
}

// NewAutoscalingOrchestrationClient creates a new client for communicating with Heat service through OpenStack go sdk.
func NewAutoscalingOrchestrationClient(authOpts gophercloud.AuthOptions,  endpointOpts gophercloud.EndpointOpts) (*autoscalingOrchestrationClient, error) {
    var auth_provider, err = openstack.AuthenticatedClient(authOpts)
	if err != nil {
		return nil, err
	}
    orchestrationClient, err := openstack.NewOrchestrationV1(auth_provider,  endpoint_options)
	if err != nil {
		return nil, err
	}

	return &autoscalingOrchestrationClient{
		projectId:             projectId,
		openstackService:      orchestrationClient,
		operationWaitTimeout:  defaultOperationWaitTimeout,
		operationPollInterval: defaultOperationPollInterval,
	}, nil
}

func (client *autoscalingOrchestrationClient) FetchASGTargetSize(asgRef OpenStackRef) (int64, error) {
    rsrc_result := stackresources.Get(client.openStackService, asgRef.Stack, asgRef.Name)
    if rsrc_result.Err != nil {
        return _, rsrc_result.Err
    }
    rsrc, err := rsrc_result.Extract()
    if err != nil {
        return _, err
    }
    asg_stack_id := rsrc.PhysicalID

    all_stack_rsrc_pages, err := stackresources.List(client, asg_stack_id, nil).AllPages()
    if err != nil {
        return _, err
    }

    fmt.Println(all_stack_rsrc_pages)

    all_stack_rsrcs, err := stackresources.ExtractResources(all_stack_rsrc_pages)
    if err != nil {
        return _, err
    }
    return len(all_stack_rsrcs), nil
}

func (client *autoscalingOrchestrationClient) ResizeASG(asgRef OpenStackRef, size int64) error {
    //TODO resize ASG
    // Fetch Template, change template, do stack update
}

func (client *autoscalingOrchestrationClient) DeleteInstances(asgRef OpenStackRef, instances []*OpenStackRef) error {
    // TODO delete instances for ASG
    // Mark as unhealthy then trigger update to delete from ASG
}

func (client *autoscalingOrchestrationClient) FetchASGInstances(asgRef OpenStackRef) ([]OpenStackRef, error) {
	//instances, err
    // TODO get instance list for ASG
    //do resource list for ASG, get all type with OS::Nova::Server and make a list
	//refs := []OpenStackRef{}
	return refs, nil
}

func (client *autoscalingOrchestrationClient) FetchASGsWithName(name *regexp.Regexp) ([]string, error) {
	links := make([]string, 0)
    //TODO get links from recurivly stack list
    // A link should looks like openstack://<project-id>/<root_stack_id>/<stack_id>/<name>
	return links, nil
}
