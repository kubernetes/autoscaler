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
	"strconv"

    "github.com/gophercloud/gophercloud"
    // Magnum
    "github.com/gophercloud/gophercloud/openstack/containerinfra/v1/clusters"

)

type MagnumClient struct {
	projectId               string

    openstackService        *gophercloud.ServiceClient

	// These can be overridden, e.g. for testing.
	operationWaitTimeout    time.Duration
	operationPollInterval   time.Duration
}

func (client *MagnumClient) FetchASGTargetSize(asgRef OpenStackRef) (int64, error) {
    cluster_result, err := clusters.Get(client.openstackService, asgRef.Resource).Extract()
    if err != nil {
        return int64(0), err
    }
    return int64(cluster_result.NodeCount), nil
}

func (client *MagnumClient) ResizeASG(asgRef OpenStackRef, size int64) error {
    // resize Magnum Cluster
    updateOpts := []clusters.UpdateOptsBuilder{
        clusters.UpdateOpts{
            Op:    clusters.ReplaceOp,
            Path:  "/node_cunt",
            Value: strconv.FormatInt(size, 10),
        },
    }
    _, err := clusters.Update(client.openstackService, asgRef.Resource, updateOpts).Extract()
    return err
}

func (client *MagnumClient) DeleteInstances(asgRef OpenStackRef, instances []*OpenStackRef) error {
    // delete instances for Cluster
    return nil
}

func (client *MagnumClient) FetchASGInstances(asgRef OpenStackRef) ([]OpenStackRef, error) {
    // TODO get instance list for ASG
	return nil, nil
}

func (client *MagnumClient) FetchASGBasename(asgRef OpenStackRef) (string, error) {
    // TODO get base name for ASG
    return "", nil
}


func (client *MagnumClient) FetchASGsWithName(name *regexp.Regexp) ([]string, error) {
	links := make([]string, 0)
    //TODO get links from recurivly stack list
    // A link should looks like openstack://<project-id>/<root_stack_id>/<stack_id>/<name>
    // TODO a link can be some other format like <project-id>,<root_stack_id>,<stack_id>,<name>
	return links, nil
}
