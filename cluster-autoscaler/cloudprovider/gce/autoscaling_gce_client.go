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
	"strings"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/utils/klogx"

	gce "google.golang.org/api/compute/v1"
	klog "k8s.io/klog/v2"
)

const (
	defaultOperationWaitTimeout  = 20 * time.Second
	defaultOperationPollInterval = 100 * time.Millisecond

	// ErrorCodeQuotaExceeded is error code used in InstanceErrorInfo if quota exceeded error occurs.
	ErrorCodeQuotaExceeded = "QUOTA_EXCEEDED"

	// ErrorCodeResourcePoolExhausted is error code used in InstanceErrorInfo if requested resources
	// cannot be provisioned by cloud provider.
	ErrorCodeResourcePoolExhausted = "RESOURCE_POOL_EXHAUSTED"

	// ErrorCodeOther is error code used in InstanceErrorInfo if other error occurs.
	ErrorCodeOther = "OTHER"
)

// AutoscalingGceClient is used for communicating with GCE API.
type AutoscalingGceClient interface {
	// reading resources
	FetchMachineType(zone, machineType string) (*gce.MachineType, error)
	FetchMachineTypes(zone string) ([]*gce.MachineType, error)
	FetchAllMigs(zone string) ([]*gce.InstanceGroupManager, error)
	FetchMigTargetSize(GceRef) (int64, error)
	FetchMigBasename(GceRef) (string, error)
	FetchMigInstances(GceRef) ([]cloudprovider.Instance, error)
	FetchMigTemplate(GceRef) (*gce.InstanceTemplate, error)
	FetchMigsWithName(zone string, filter *regexp.Regexp) ([]string, error)
	FetchZones(region string) ([]string, error)
	FetchAvailableCpuPlatforms() (map[string][]string, error)

	// modifying resources
	ResizeMig(GceRef, int64) error
	DeleteInstances(migRef GceRef, instances []GceRef) error
}

type autoscalingGceClientV1 struct {
	gceService *gce.Service

	projectId string

	// These can be overridden, e.g. for testing.
	operationWaitTimeout  time.Duration
	operationPollInterval time.Duration
}

// NewAutoscalingGceClientV1 creates a new client for communicating with GCE v1 API.
func NewAutoscalingGceClientV1(client *http.Client, projectId string) (*autoscalingGceClientV1, error) {
	gceService, err := gce.New(client)
	if err != nil {
		return nil, err
	}

	return &autoscalingGceClientV1{
		projectId:             projectId,
		gceService:            gceService,
		operationWaitTimeout:  defaultOperationWaitTimeout,
		operationPollInterval: defaultOperationPollInterval,
	}, nil
}

// NewCustomAutoscalingGceClientV1 creates a new client using custom server url and timeouts
// for communicating with GCE v1 API.
func NewCustomAutoscalingGceClientV1(client *http.Client, projectId, serverUrl string,
	waitTimeout, pollInterval time.Duration) (*autoscalingGceClientV1, error) {
	gceService, err := gce.New(client)
	if err != nil {
		return nil, err
	}
	gceService.BasePath = serverUrl

	return &autoscalingGceClientV1{
		projectId:             projectId,
		gceService:            gceService,
		operationWaitTimeout:  waitTimeout,
		operationPollInterval: pollInterval,
	}, nil
}

func (client *autoscalingGceClientV1) FetchMachineType(zone, machineType string) (*gce.MachineType, error) {
	registerRequest("machine_types", "get")
	return client.gceService.MachineTypes.Get(client.projectId, zone, machineType).Do()
}

func (client *autoscalingGceClientV1) FetchMachineTypes(zone string) ([]*gce.MachineType, error) {
	registerRequest("machine_types", "list")
	var machineTypes []*gce.MachineType
	err := client.gceService.MachineTypes.List(client.projectId, zone).Pages(
		context.TODO(),
		func(page *gce.MachineTypeList) error {
			machineTypes = append(machineTypes, page.Items...)
			return nil
		})
	if err != nil {
		return nil, err
	}
	return machineTypes, nil
}

func (client *autoscalingGceClientV1) FetchAllMigs(zone string) ([]*gce.InstanceGroupManager, error) {
	registerRequest("instance_group_managers", "list")
	var migs []*gce.InstanceGroupManager
	err := client.gceService.InstanceGroupManagers.List(client.projectId, zone).Pages(
		context.TODO(),
		func(page *gce.InstanceGroupManagerList) error {
			migs = append(migs, page.Items...)
			return nil
		})
	if err != nil {
		return nil, err
	}
	return migs, nil
}

func (client *autoscalingGceClientV1) FetchMigTargetSize(migRef GceRef) (int64, error) {
	registerRequest("instance_group_managers", "get")
	igm, err := client.gceService.InstanceGroupManagers.Get(migRef.Project, migRef.Zone, migRef.Name).Do()
	if err != nil {
		return 0, err
	}
	return igm.TargetSize, nil
}

func (client *autoscalingGceClientV1) FetchMigBasename(migRef GceRef) (string, error) {
	registerRequest("instance_group_managers", "get")
	igm, err := client.gceService.InstanceGroupManagers.Get(migRef.Project, migRef.Zone, migRef.Name).Do()
	if err != nil {
		return "", err
	}
	return igm.BaseInstanceName, nil
}

func (client *autoscalingGceClientV1) ResizeMig(migRef GceRef, size int64) error {
	registerRequest("instance_group_managers", "resize")
	op, err := client.gceService.InstanceGroupManagers.Resize(migRef.Project, migRef.Zone, migRef.Name, size).Do()
	if err != nil {
		return err
	}
	return client.waitForOp(op, migRef.Project, migRef.Zone)
}

func (client *autoscalingGceClientV1) waitForOp(operation *gce.Operation, project, zone string) error {
	for start := time.Now(); time.Since(start) < client.operationWaitTimeout; time.Sleep(client.operationPollInterval) {
		klog.V(4).Infof("Waiting for operation %s %s %s", project, zone, operation.Name)
		registerRequest("zone_operations", "get")
		if op, err := client.gceService.ZoneOperations.Get(project, zone, operation.Name).Do(); err == nil {
			klog.V(4).Infof("Operation %s %s %s status: %s", project, zone, operation.Name, op.Status)
			if op.Status == "DONE" {
				return nil
			}
		} else {
			klog.Warningf("Error while getting operation %s on %s: %v", operation.Name, operation.TargetLink, err)
		}
	}
	return fmt.Errorf("timeout while waiting for operation %s on %s to complete.", operation.Name, operation.TargetLink)
}

func (client *autoscalingGceClientV1) DeleteInstances(migRef GceRef, instances []GceRef) error {
	registerRequest("instance_group_managers", "delete_instances")
	req := gce.InstanceGroupManagersDeleteInstancesRequest{
		Instances: []string{},
	}
	for _, i := range instances {
		req.Instances = append(req.Instances, GenerateInstanceUrl(i))
	}
	op, err := client.gceService.InstanceGroupManagers.DeleteInstances(migRef.Project, migRef.Zone, migRef.Name, &req).Do()
	if err != nil {
		return err
	}
	return client.waitForOp(op, migRef.Project, migRef.Zone)
}

func (client *autoscalingGceClientV1) FetchMigInstances(migRef GceRef) ([]cloudprovider.Instance, error) {
	registerRequest("instance_group_managers", "list_managed_instances")
	gceInstances, err := client.gceService.InstanceGroupManagers.ListManagedInstances(migRef.Project, migRef.Zone, migRef.Name).Do()
	if err != nil {
		klog.V(4).Infof("Failed MIG info request for %s %s %s: %v", migRef.Project, migRef.Zone, migRef.Name, err)
		return nil, err
	}
	infos := []cloudprovider.Instance{}
	errorCodeCounts := make(map[string]int)
	errorLoggingQuota := klogx.NewLoggingQuota(100)
	for _, gceInstance := range gceInstances.ManagedInstances {
		ref, err := ParseInstanceUrlRef(gceInstance.Instance)
		if err != nil {
			return nil, err
		}

		instance := cloudprovider.Instance{
			Id:     ref.ToProviderId(),
			Status: &cloudprovider.InstanceStatus{},
		}

		switch gceInstance.CurrentAction {
		case "CREATING", "RECREATING", "CREATING_WITHOUT_RETRIES":
			instance.Status.State = cloudprovider.InstanceCreating
		case "ABANDONING", "DELETING":
			instance.Status.State = cloudprovider.InstanceDeleting
		default:
			instance.Status.State = cloudprovider.InstanceRunning
		}

		if instance.Status.State == cloudprovider.InstanceCreating {
			var errorInfo cloudprovider.InstanceErrorInfo
			errorMessages := []string{}
			errorFound := false
			lastAttemptErrors := getLastAttemptErrors(gceInstance)
			for _, instanceError := range lastAttemptErrors {
				errorCodeCounts[instanceError.Code]++
				if isResourcePoolExhaustedErrorCode(instanceError.Code) {
					errorInfo.ErrorClass = cloudprovider.OutOfResourcesErrorClass
					errorInfo.ErrorCode = ErrorCodeResourcePoolExhausted
				} else if isQuotaExceededErrorCoce(instanceError.Code) {
					errorInfo.ErrorClass = cloudprovider.OutOfResourcesErrorClass
					errorInfo.ErrorCode = ErrorCodeQuotaExceeded
				} else if isInstanceNotRunningYet(gceInstance) {
					if !errorFound {
						// do not override error code with OTHER
						errorInfo.ErrorClass = cloudprovider.OtherErrorClass
						errorInfo.ErrorCode = ErrorCodeOther
					}
				} else {
					// no error
					continue
				}
				errorFound = true
				if instanceError.Message != "" {
					errorMessages = append(errorMessages, instanceError.Message)
				}
			}
			errorInfo.ErrorMessage = strings.Join(errorMessages, "; ")
			if errorFound {
				instance.Status.ErrorInfo = &errorInfo
			}

			if len(lastAttemptErrors) > 0 {
				gceInstanceJSONBytes, err := gceInstance.MarshalJSON()
				var gceInstanceJSON string
				if err != nil {
					gceInstanceJSON = fmt.Sprintf("Got error from MarshalJSON; %v", err)
				} else {
					gceInstanceJSON = string(gceInstanceJSONBytes)
				}
				klogx.V(4).UpTo(errorLoggingQuota).Infof("Got GCE instance which is being created and has lastAttemptErrors; gceInstance=%v; errorInfo=%#v", gceInstanceJSON, errorInfo)
			}
		}
		infos = append(infos, instance)
	}
	klogx.V(4).Over(errorLoggingQuota).Infof("Got %v other GCE instances being created with lastAttemptErrors", -errorLoggingQuota.Left())
	if len(errorCodeCounts) > 0 {
		klog.V(4).Infof("Spotted following instance creation error codes: %#v", errorCodeCounts)
	}
	return infos, nil
}

func getLastAttemptErrors(instance *gce.ManagedInstance) []*gce.ManagedInstanceLastAttemptErrorsErrors {
	if instance.LastAttempt != nil && instance.LastAttempt.Errors != nil {
		return instance.LastAttempt.Errors.Errors
	}
	return nil
}

func isResourcePoolExhaustedErrorCode(errorCode string) bool {
	return errorCode == "RESOURCE_POOL_EXHAUSTED" || errorCode == "ZONE_RESOURCE_POOL_EXHAUSTED" || errorCode == "ZONE_RESOURCE_POOL_EXHAUSTED_WITH_DETAILS"
}

func isQuotaExceededErrorCoce(errorCode string) bool {
	return strings.Contains(errorCode, "QUOTA")
}

func isInstanceNotRunningYet(gceInstance *gce.ManagedInstance) bool {
	return gceInstance.InstanceStatus == "" || gceInstance.InstanceStatus == "PROVISIONING" || gceInstance.InstanceStatus == "STAGING"
}

func (client *autoscalingGceClientV1) FetchZones(region string) ([]string, error) {
	registerRequest("regions", "get")
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

func (client *autoscalingGceClientV1) FetchAvailableCpuPlatforms() (map[string][]string, error) {
	availableCpuPlatforms := make(map[string][]string)
	err := client.gceService.Zones.List(client.projectId).Pages(
		context.TODO(),
		func(zones *gce.ZoneList) error {
			for _, zone := range zones.Items {
				availableCpuPlatforms[zone.Name] = zone.AvailableCpuPlatforms
			}
			return nil
		})
	if err != nil {
		return nil, err
	}
	return availableCpuPlatforms, nil
}

func (client *autoscalingGceClientV1) FetchMigTemplate(migRef GceRef) (*gce.InstanceTemplate, error) {
	registerRequest("instance_group_managers", "get")
	igm, err := client.gceService.InstanceGroupManagers.Get(migRef.Project, migRef.Zone, migRef.Name).Do()
	if err != nil {
		return nil, err
	}
	templateUrl, err := url.Parse(igm.InstanceTemplate)
	if err != nil {
		return nil, err
	}
	_, templateName := path.Split(templateUrl.EscapedPath())
	registerRequest("instance_templates", "get")
	return client.gceService.InstanceTemplates.Get(migRef.Project, templateName).Do()
}

func (client *autoscalingGceClientV1) FetchMigsWithName(zone string, name *regexp.Regexp) ([]string, error) {
	filter := fmt.Sprintf("name eq %s", name)
	links := make([]string, 0)
	registerRequest("instance_groups", "list")
	req := client.gceService.InstanceGroups.List(client.projectId, zone).Filter(filter)
	if err := req.Pages(context.TODO(), func(page *gce.InstanceGroupList) error {
		for _, ig := range page.Items {
			links = append(links, ig.SelfLink)
			klog.V(3).Infof("found managed instance group %s matching regexp %s", ig.Name, name)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("cannot list managed instance groups: %v", err)
	}
	return links, nil
}
