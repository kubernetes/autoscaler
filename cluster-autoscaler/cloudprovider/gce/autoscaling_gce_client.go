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

	"google.golang.org/api/googleapi"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/klogx"

	gce "google.golang.org/api/compute/v1"
	klog "k8s.io/klog/v2"
)

const (
	defaultOperationWaitTimeout          = 20 * time.Second
	defaultOperationPollInterval         = 100 * time.Millisecond
	defaultOperationDeletionPollInterval = 1 * time.Second
	// ErrorCodeQuotaExceeded is an error code used in InstanceErrorInfo if quota exceeded error occurs.
	ErrorCodeQuotaExceeded = "QUOTA_EXCEEDED"

	// ErrorCodeResourcePoolExhausted is an error code used in InstanceErrorInfo if requested resources
	// cannot be provisioned by cloud provider.
	ErrorCodeResourcePoolExhausted = "RESOURCE_POOL_EXHAUSTED"

	// ErrorIPSpaceExhausted is an error code used in InstanceErrorInfo if the IP space has been
	// exhausted.
	ErrorIPSpaceExhausted = "IP_SPACE_EXHAUSTED"

	// ErrorCodePermissions is an error code used in InstanceErrorInfo if the user is facing
	// permissions error
	ErrorCodePermissions = "PERMISSIONS_ERROR"

	// ErrorCodeVmExternalIpAccessPolicyConstraint is an error code in InstanceErrorInfo if the user
	// is facing errors caused by vmExternalIpAccess policy constraint misconfiguration.
	ErrorCodeVmExternalIpAccessPolicyConstraint = "VM_EXTERNAL_IP_ACCESS_POLICY_CONSTRAINT"

	// ErrorCodeOther is an error code used in InstanceErrorInfo if other error occurs.
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
	FetchMigTemplateName(migRef GceRef) (string, error)
	FetchMigTemplate(migRef GceRef, templateName string) (*gce.InstanceTemplate, error)
	FetchMigsWithName(zone string, filter *regexp.Regexp) ([]string, error)
	FetchZones(region string) ([]string, error)
	FetchAvailableCpuPlatforms() (map[string][]string, error)

	// modifying resources
	ResizeMig(GceRef, int64) error
	DeleteInstances(migRef GceRef, instances []GceRef) error
	CreateInstances(GceRef, string, int64, []string) error
}

type autoscalingGceClientV1 struct {
	gceService *gce.Service

	projectId string

	// These can be overridden, e.g. for testing.
	operationWaitTimeout          time.Duration
	operationPollInterval         time.Duration
	operationDeletionPollInterval time.Duration
}

// NewAutoscalingGceClientV1 creates a new client for communicating with GCE v1 API.
func NewAutoscalingGceClientV1(client *http.Client, projectId string, userAgent string) (*autoscalingGceClientV1, error) {
	gceService, err := gce.New(client)
	if err != nil {
		return nil, err
	}
	gceService.UserAgent = userAgent

	return &autoscalingGceClientV1{
		projectId:                     projectId,
		gceService:                    gceService,
		operationWaitTimeout:          defaultOperationWaitTimeout,
		operationPollInterval:         defaultOperationPollInterval,
		operationDeletionPollInterval: defaultOperationDeletionPollInterval,
	}, nil
}

// NewCustomAutoscalingGceClientV1 creates a new client using custom server url and timeouts
// for communicating with GCE v1 API.
func NewCustomAutoscalingGceClientV1(client *http.Client, projectId, serverUrl, userAgent string,
	waitTimeout, pollInterval time.Duration, deletionPollInterval time.Duration) (*autoscalingGceClientV1, error) {
	gceService, err := gce.New(client)
	if err != nil {
		return nil, err
	}
	gceService.BasePath = serverUrl
	gceService.UserAgent = userAgent

	return &autoscalingGceClientV1{
		projectId:                     projectId,
		gceService:                    gceService,
		operationWaitTimeout:          waitTimeout,
		operationPollInterval:         pollInterval,
		operationDeletionPollInterval: deletionPollInterval,
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
		if err, ok := err.(*googleapi.Error); ok {
			if err.Code == http.StatusNotFound {
				return 0, errors.NewAutoscalerError(errors.NodeGroupDoesNotExistError, "%s", err.Error())
			}
		}
		return 0, err
	}
	return igm.TargetSize, nil
}

func (client *autoscalingGceClientV1) FetchMigBasename(migRef GceRef) (string, error) {
	registerRequest("instance_group_managers", "get")
	igm, err := client.gceService.InstanceGroupManagers.Get(migRef.Project, migRef.Zone, migRef.Name).Do()
	if err != nil {
		if err, ok := err.(*googleapi.Error); ok && err.Code == http.StatusNotFound {
			return "", errors.NewAutoscalerError(errors.NodeGroupDoesNotExistError, "%s", err.Error())
		}
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
	return client.waitForOp(op, migRef.Project, migRef.Zone, false)
}

func (client *autoscalingGceClientV1) CreateInstances(migRef GceRef, baseName string, delta int64, existingInstanceProviderIds []string) error {
	registerRequest("instance_group_managers", "create_instances")
	req := gce.InstanceGroupManagersCreateInstancesRequest{}
	instanceNames := instanceIdsToNamesMap(existingInstanceProviderIds)
	req.Instances = make([]*gce.PerInstanceConfig, 0, delta)
	for i := int64(0); i < delta; i++ {
		newInstanceName := generateInstanceName(baseName, instanceNames)
		instanceNames[newInstanceName] = true
		req.Instances = append(req.Instances, &gce.PerInstanceConfig{Name: newInstanceName})
	}
	op, err := client.gceService.InstanceGroupManagers.CreateInstances(migRef.Project, migRef.Zone, migRef.Name, &req).Do()
	if err != nil {
		return err
	}
	return client.waitForOp(op, migRef.Project, migRef.Zone, false)
}

func instanceIdsToNamesMap(instanceProviderIds []string) map[string]bool {
	instanceNames := make(map[string]bool, len(instanceProviderIds))
	for _, inst := range instanceProviderIds {
		ref, err := GceRefFromProviderId(inst)
		if err != nil {
			klog.Warningf("Failed to extract instance name from %q: %v", inst, err)
		} else {
			inst = ref.Name
		}
		instanceNames[inst] = true
	}
	return instanceNames
}

func (client *autoscalingGceClientV1) waitForOp(operation *gce.Operation, project, zone string, isDeletion bool) error {
	pollInterval := client.operationPollInterval
	if isDeletion {
		pollInterval = client.operationDeletionPollInterval
	}
	for start := time.Now(); time.Since(start) < client.operationWaitTimeout; time.Sleep(pollInterval) {
		klog.V(4).Infof("Waiting for operation %s %s %s", project, zone, operation.Name)
		registerRequest("zone_operations", "get")
		if op, err := client.gceService.ZoneOperations.Get(project, zone, operation.Name).Do(); err == nil {
			klog.V(4).Infof("Operation %s %s %s status: %s", project, zone, operation.Name, op.Status)
			if op.Status == "DONE" {
				if op.Error != nil {
					errBytes, err := op.Error.MarshalJSON()
					if err != nil {
						errBytes = []byte(fmt.Sprintf("operation failed, but error couldn't be recovered: %v", err))
					}
					return fmt.Errorf("error while getting operation %s on %s: %s", operation.Name, operation.TargetLink, errBytes)
				}

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
		Instances:                      []string{},
		SkipInstancesOnValidationError: true,
	}
	for _, i := range instances {
		req.Instances = append(req.Instances, GenerateInstanceUrl(i))
	}
	op, err := client.gceService.InstanceGroupManagers.DeleteInstances(migRef.Project, migRef.Zone, migRef.Name, &req).Do()
	if err != nil {
		return err
	}
	return client.waitForOp(op, migRef.Project, migRef.Zone, true)
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
			klog.Errorf("Received error while parsing of the instance url: %v", err)
			continue
		}

		instance := cloudprovider.Instance{
			Id: ref.ToProviderId(),
			Status: &cloudprovider.InstanceStatus{
				State: getInstanceState(gceInstance.CurrentAction),
			},
		}

		if instance.Status.State == cloudprovider.InstanceCreating {
			var errorInfo *cloudprovider.InstanceErrorInfo
			errorMessages := []string{}
			lastAttemptErrors := getLastAttemptErrors(gceInstance)
			for _, instanceError := range lastAttemptErrors {
				errorCodeCounts[instanceError.Code]++
				if newErrorInfo := GetErrorInfo(instanceError.Code, instanceError.Message, gceInstance.InstanceStatus, errorInfo); newErrorInfo != nil {
					// override older error
					errorInfo = newErrorInfo
				} else {
					// no error
					continue
				}

				if instanceError.Message != "" {
					errorMessages = append(errorMessages, instanceError.Message)
				}
			}
			if errorInfo != nil {
				errorInfo.ErrorMessage = strings.Join(errorMessages, "; ")
				instance.Status.ErrorInfo = errorInfo
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
		klog.Warningf("Spotted following instance creation error codes: %#v", errorCodeCounts)
	}
	return infos, nil
}

// GetErrorInfo maps the error code, error message and instance status to CA instance error info
func GetErrorInfo(errorCode, errorMessage, instanceStatus string, previousErrorInfo *cloudprovider.InstanceErrorInfo) *cloudprovider.InstanceErrorInfo {
	if isResourcePoolExhaustedErrorCode(errorCode) {
		return &cloudprovider.InstanceErrorInfo{
			ErrorClass: cloudprovider.OutOfResourcesErrorClass,
			ErrorCode:  ErrorCodeResourcePoolExhausted,
		}
	} else if isQuotaExceededErrorCode(errorCode) {
		return &cloudprovider.InstanceErrorInfo{
			ErrorClass: cloudprovider.OutOfResourcesErrorClass,
			ErrorCode:  ErrorCodeQuotaExceeded,
		}
	} else if isIPSpaceExhaustedErrorCode(errorCode) {
		return &cloudprovider.InstanceErrorInfo{
			ErrorClass: cloudprovider.OtherErrorClass,
			ErrorCode:  ErrorIPSpaceExhausted,
		}
	} else if isPermissionsError(errorCode) {
		return &cloudprovider.InstanceErrorInfo{
			ErrorClass: cloudprovider.OtherErrorClass,
			ErrorCode:  ErrorCodePermissions,
		}
	} else if isVmExternalIpAccessPolicyConstraintError(errorCode, errorMessage) {
		return &cloudprovider.InstanceErrorInfo{
			ErrorClass: cloudprovider.OtherErrorClass,
			ErrorCode:  ErrorCodeVmExternalIpAccessPolicyConstraint,
		}
	} else if isInstanceStatusNotRunningYet(instanceStatus) {
		if previousErrorInfo != nil {
			// keep the current error
			return previousErrorInfo
		}
		return &cloudprovider.InstanceErrorInfo{
			ErrorClass: cloudprovider.OtherErrorClass,
			ErrorCode:  ErrorCodeOther,
		}
	}
	return nil
}

func getInstanceState(currentAction string) cloudprovider.InstanceState {
	switch currentAction {
	case "CREATING", "RECREATING", "CREATING_WITHOUT_RETRIES":
		return cloudprovider.InstanceCreating
	case "ABANDONING", "DELETING":
		return cloudprovider.InstanceDeleting
	default:
		return cloudprovider.InstanceRunning
	}
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

func isQuotaExceededErrorCode(errorCode string) bool {
	return strings.Contains(errorCode, "QUOTA")
}

func isIPSpaceExhaustedErrorCode(errorCode string) bool {
	return strings.Contains(errorCode, "IP_SPACE_EXHAUSTED")
}

func isPermissionsError(errorCode string) bool {
	return strings.Contains(errorCode, "PERMISSIONS_ERROR")
}

func isVmExternalIpAccessPolicyConstraintError(errorCode, errorMessage string) bool {
	regexProjectPolicyConstraint := regexp.MustCompile(`Constraint constraints/compute.vmExternalIpAccess violated for project`)
	return strings.Contains(errorCode, "CONDITION_NOT_MET") && regexProjectPolicyConstraint.MatchString(errorMessage)
}

func isInstanceStatusNotRunningYet(instanceStatus string) bool {
	return instanceStatus == "" || instanceStatus == "PROVISIONING" || instanceStatus == "STAGING"
}

func generateInstanceName(baseName string, existingNames map[string]bool) string {
	for i := 0; i < 100; i++ {
		name := fmt.Sprintf("%v-%v", baseName, rand.String(4))
		if ok, _ := existingNames[name]; !ok {
			return name
		}
	}
	klog.Warning("Unable to create unique name for a new instance, duplicate name might occur")
	name := fmt.Sprintf("%v-%v", baseName, rand.String(4))
	return name
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

func (client *autoscalingGceClientV1) FetchMigTemplateName(migRef GceRef) (string, error) {
	registerRequest("instance_group_managers", "get")
	igm, err := client.gceService.InstanceGroupManagers.Get(migRef.Project, migRef.Zone, migRef.Name).Do()
	if err != nil {
		if err, ok := err.(*googleapi.Error); ok {
			if err.Code == http.StatusNotFound {
				return "", errors.NewAutoscalerError(errors.NodeGroupDoesNotExistError, "%s", err.Error())
			}
		}
		return "", err
	}
	templateUrl, err := url.Parse(igm.InstanceTemplate)
	if err != nil {
		return "", err
	}
	_, templateName := path.Split(templateUrl.EscapedPath())
	return templateName, nil
}

func (client *autoscalingGceClientV1) FetchMigTemplate(migRef GceRef, templateName string) (*gce.InstanceTemplate, error) {
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
