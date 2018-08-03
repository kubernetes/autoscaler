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

package gce

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"time"

	"github.com/golang/glog"
	gce "google.golang.org/api/compute/v1"
)

const (
	operationWaitTimeout  = 5 * time.Second
	operationPollInterval = 100 * time.Millisecond
)

// AutoscalingGceClient is used for communicating with GCE API.
type AutoscalingGceClient interface {
	// reading resources
	FetchMachineType(zone, machineType string) (*gce.MachineType, error)
	FetchMachineTypes(zone string) ([]*gce.MachineType, error)
	FetchMigTargetSize(GceRef) (int64, error)
	FetchMigBasename(GceRef) (string, error)
	FetchMigInstances(GceRef) ([]GceRef, error)
	FetchMigTemplate(GceRef) (*gce.InstanceTemplate, error)
	FetchMigs(zone string, filter *regexp.Regexp) ([]string, error)
	FetchZones(region string) ([]string, error)

	// modifying resources
	ResizeMig(GceRef, int64) error
	DeleteInstances(migRef GceRef, instances []*GceRef) error
}

type autoscalingGceClientV1 struct {
	gceService *gce.Service

	projectId string
}

// NewAutoscalingGceClientV1 creates a new client for communicating with GCE v1 API.
func NewAutoscalingGceClientV1(client *http.Client, projectId string) (*autoscalingGceClientV1, error) {
	gceService, err := gce.New(client)
	if err != nil {
		return nil, err
	}

	return &autoscalingGceClientV1{
		projectId:  projectId,
		gceService: gceService,
	}, nil
}

func (client *autoscalingGceClientV1) FetchMachineType(zone, machineType string) (*gce.MachineType, error) {
	return client.gceService.MachineTypes.Get(client.projectId, zone, machineType).Do()
}

func (client *autoscalingGceClientV1) FetchMachineTypes(zone string) ([]*gce.MachineType, error) {
	machines, err := client.gceService.MachineTypes.List(client.projectId, zone).Do()
	if err != nil {
		return nil, err
	}
	return machines.Items, nil
}

func (client *autoscalingGceClientV1) FetchMigTargetSize(migRef GceRef) (int64, error) {
	igm, err := client.gceService.InstanceGroupManagers.Get(migRef.Project, migRef.Zone, migRef.Name).Do()
	if err != nil {
		return 0, err
	}
	return igm.TargetSize, nil
}

func (client *autoscalingGceClientV1) FetchMigBasename(migRef GceRef) (string, error) {
	igm, err := client.gceService.InstanceGroupManagers.Get(migRef.Project, migRef.Zone, migRef.Name).Do()
	if err != nil {
		return "", err
	}
	return igm.BaseInstanceName, nil
}

func (client *autoscalingGceClientV1) ResizeMig(migRef GceRef, size int64) error {
	op, err := client.gceService.InstanceGroupManagers.Resize(migRef.Project, migRef.Zone, migRef.Name, size).Do()
	if err != nil {
		return err
	}
	return client.waitForOp(op, migRef.Project, migRef.Zone)
}

func (client *autoscalingGceClientV1) waitForOp(operation *gce.Operation, project, zone string) error {
	for start := time.Now(); time.Since(start) < operationWaitTimeout; time.Sleep(operationPollInterval) {
		glog.V(4).Infof("Waiting for operation %s %s %s", project, zone, operation.Name)
		if op, err := client.gceService.ZoneOperations.Get(project, zone, operation.Name).Do(); err == nil {
			glog.V(4).Infof("Operation %s %s %s status: %s", project, zone, operation.Name, op.Status)
			if op.Status == "DONE" {
				return nil
			}
		} else {
			glog.Warningf("Error while getting operation %s on %s: %v", operation.Name, operation.TargetLink, err)
		}
	}
	return fmt.Errorf("Timeout while waiting for operation %s on %s to complete.", operation.Name, operation.TargetLink)
}

func (client *autoscalingGceClientV1) DeleteInstances(migRef GceRef, instances []*GceRef) error {
	req := gce.InstanceGroupManagersDeleteInstancesRequest{
		Instances: []string{},
	}
	for _, i := range instances {
		req.Instances = append(req.Instances, GenerateInstanceUrl(*i))
	}
	op, err := client.gceService.InstanceGroupManagers.DeleteInstances(migRef.Project, migRef.Zone, migRef.Name, &req).Do()
	if err != nil {
		return err
	}
	return client.waitForOp(op, migRef.Project, migRef.Zone)
}

func (client *autoscalingGceClientV1) FetchMigInstances(migRef GceRef) ([]GceRef, error) {
	instances, err := client.gceService.InstanceGroupManagers.ListManagedInstances(migRef.Project, migRef.Zone, migRef.Name).Do()
	if err != nil {
		glog.V(4).Infof("Failed MIG info request for %s %s %s: %v", migRef.Project, migRef.Zone, migRef.Name, err)
		return nil, err
	}
	refs := []GceRef{}
	for _, i := range instances.ManagedInstances {
		ref, err := ParseInstanceUrlRef(i.Instance)
		if err != nil {
			return nil, err
		}
		refs = append(refs, ref)
	}
	return refs, nil
}

func (client *autoscalingGceClientV1) FetchZones(region string) ([]string, error) {
	r, err := client.gceService.Regions.Get(client.projectId, region).Do()
	if err != nil {
		return nil, fmt.Errorf("cannot get zones for GCE region %s: %v", region, err)
	}
	zones := make([]string, len(r.Zones))
	for i, link := range r.Zones {
		zones[i] = path.Base(link)
	}
	return zones, nil
}

func (client *autoscalingGceClientV1) FetchMigTemplate(migRef GceRef) (*gce.InstanceTemplate, error) {
	igm, err := client.gceService.InstanceGroupManagers.Get(migRef.Project, migRef.Zone, migRef.Name).Do()
	if err != nil {
		return nil, err
	}
	templateUrl, err := url.Parse(igm.InstanceTemplate)
	if err != nil {
		return nil, err
	}
	_, templateName := path.Split(templateUrl.EscapedPath())
	return client.gceService.InstanceTemplates.Get(migRef.Project, templateName).Do()
}

func (client *autoscalingGceClientV1) FetchMigs(zone string, name *regexp.Regexp) ([]string, error) {
	filter := fmt.Sprintf("name eq %s", name)
	links := make([]string, 0)
	req := client.gceService.InstanceGroups.List(client.projectId, zone).Filter(filter)
	if err := req.Pages(context.TODO(), func(page *gce.InstanceGroupList) error {
		for _, ig := range page.Items {
			links = append(links, ig.SelfLink)
			glog.V(3).Infof("autodiscovered managed instance group %s using regexp %s", ig.Name, name)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("cannot list managed instance groups: %v", err)
	}
	return links, nil
}
