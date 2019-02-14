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
)

type autoscalingOrchestrationClient struct {
	projectId               string

    openstackService        *gophercloud.ServiceClient

	// These can be overridden, e.g. for testing.
	operationWaitTimeout    time.Duration
	operationPollInterval   time.Duration
}

func (client *autoscalingOrchestrationClient) FetchASGTargetSize(asgRef OpenStackRef) (int64, error) {
    //rsrc_result := stackresources.Get(client.openstackService, asgRef.Name, asgRef.Resource, asgRef.Name)
    //if rsrc_result.Err != nil {
    //    return int64(0), rsrc_result.Err
    //}
    //rsrc, err := rsrc_result.Extract()
    //if err != nil {
    //    return int64(0), err
    //}
    //asg_stack_id := rsrc.PhysicalID

    //all_stack_rsrc_pages, err := stackresources.List(client.openstackService, asg_stack_id, asg_stack_id, nil).AllPages()
    //if err != nil {
    //    return int64(0), err
    //}

    //fmt.Println(all_stack_rsrc_pages)

    //all_stack_rsrcs, err := stackresources.ExtractResources(all_stack_rsrc_pages)
    //if err != nil {
    //    return int64(0), err
    //}
    //return int64(len(all_stack_rsrcs)), nil
    return int64(0), nil
}

func (client *autoscalingOrchestrationClient) ResizeASG(asgRef OpenStackRef, size int64) error {
    //TODO resize ASG
    // Fetch Template, change template, do stack update
    return nil
}

func (client *autoscalingOrchestrationClient) DeleteInstances(asgRef OpenStackRef, instances []*OpenStackRef) error {
    // TODO delete instances for ASG
    // Mark as unhealthy then trigger update to delete from ASG
    return nil
}

func (client *autoscalingOrchestrationClient) FetchASGInstances(asgRef OpenStackRef) ([]OpenStackRef, error) {
	//instances, err
    // TODO get instance list for ASG
    //do resource list for ASG, get all type with OS::Nova::Server and make a list
	//refs := []OpenStackRef{}
	//return refs, nil
	return nil, nil
}

func (client *autoscalingOrchestrationClient) FetchASGBasename(asgRef OpenStackRef) (string, error) {
    // TODO get base name for ASG
    return "", nil
}


func (client *autoscalingOrchestrationClient) FetchASGsWithName(name *regexp.Regexp) ([]string, error) {
	links := make([]string, 0)
    //TODO get links from recurivly stack list
    // A link should looks like openstack://<project-id>/<root_stack_id>/<stack_id>/<name>
    // TODO a link can be some other format like <project-id>,<root_stack_id>,<stack_id>,<name>
	return links, nil
}
