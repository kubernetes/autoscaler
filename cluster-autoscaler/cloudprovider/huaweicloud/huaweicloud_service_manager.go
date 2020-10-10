/*
Copyright 2020 The Kubernetes Authors.

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

package huaweicloud

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	huaweicloudsdkas "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/services/as/v1"
	huaweicloudsdkasmodel "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/services/as/v1/model"
	huaweicloudsdkecs "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2"
	huaweicloudsdkecsmodel "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2/model"
	"k8s.io/klog/v2"
)

// ElasticCloudServerService represents the elastic cloud server interfaces.
// It should contains all request against elastic cloud server service.
type ElasticCloudServerService interface {
	// DeleteServers deletes a group of server by ID.
	DeleteServers(serverIDs []string) error
}

// AutoScalingService represents the auto scaling service interfaces.
// It should contains all request against auto scaling service.
type AutoScalingService interface {
	// ListScalingGroups list all scaling groups.
	ListScalingGroups() ([]AutoScalingGroup, error)

	// GetDesireInstanceNumber gets the desire instance number of specific auto scaling group.
	GetDesireInstanceNumber(groupID string) (int, error)

	// GetInstances gets the instances in an auto scaling group.
	GetInstances(groupID string) ([]cloudprovider.Instance, error)

	// IncreaseSizeInstance increases the instance number of specific auto scaling group.
	// The delta should be non-negative.
	// IncreaseSizeInstance wait until instance number is updated.
	IncreaseSizeInstance(groupID string, delta int) error
}

// CloudServiceManager represents the cloud service interfaces.
// It should contains all requests against cloud services.
type CloudServiceManager interface {
	// ElasticCloudServerService represents the elastic cloud server interfaces.
	ElasticCloudServerService

	// AutoScalingService represents the auto scaling service interfaces.
	AutoScalingService
}

type cloudServiceManager struct {
	cloudConfig      *CloudConfig
	getECSClientFunc func() *huaweicloudsdkecs.EcsClient
	getASClientFunc  func() *huaweicloudsdkas.AsClient
}

func newCloudServiceManager(cloudConfig *CloudConfig) *cloudServiceManager {
	return &cloudServiceManager{
		cloudConfig:      cloudConfig,
		getECSClientFunc: cloudConfig.getECSClient,
		getASClientFunc:  cloudConfig.getASClient,
	}
}

// DeleteServers deletes a group of server by ID.
func (csm *cloudServiceManager) DeleteServers(serverIDs []string) error {
	ecsClient := csm.getECSClientFunc()
	if ecsClient == nil {
		return fmt.Errorf("failed to delete servers due to can not get ecs client")
	}

	servers := make([]huaweicloudsdkecsmodel.ServerId, 0, len(serverIDs))
	for i := range serverIDs {
		s := huaweicloudsdkecsmodel.ServerId{
			Id: serverIDs[i],
		}
		servers = append(servers, s)
	}

	deletePublicIP := false
	deleteVolume := false
	opts := &huaweicloudsdkecsmodel.DeleteServersRequest{
		Body: &huaweicloudsdkecsmodel.DeleteServersRequestBody{
			DeletePublicip: &deletePublicIP,
			DeleteVolume:   &deleteVolume,
			Servers:        servers,
		},
	}
	deleteResponse, err := ecsClient.DeleteServers(opts)
	if err != nil {
		return fmt.Errorf("failed to delete servers. error: %v", err)
	}
	jobID := deleteResponse.JobId

	err = wait.Poll(5*time.Second, 300*time.Second, func() (bool, error) {
		showJobOpts := &huaweicloudsdkecsmodel.ShowJobRequest{
			JobId: *jobID,
		}
		showJobResponse, err := ecsClient.ShowJob(showJobOpts)
		if err != nil {
			return false, err
		}

		jobStatusEnum := huaweicloudsdkecsmodel.GetShowJobResponseStatusEnum()
		if *showJobResponse.Status == jobStatusEnum.FAIL {
			errCode := *showJobResponse.ErrorCode
			failReason := *showJobResponse.FailReason
			return false, fmt.Errorf("job failed. error code: %s, error msg: %s", errCode, failReason)
		} else if *showJobResponse.Status == jobStatusEnum.SUCCESS {
			return true, nil
		}

		return true, nil
	})

	if err != nil {
		klog.Warningf("failed to delete servers, error: %v", err)
		return err
	}

	return nil
}

func (csm *cloudServiceManager) GetDesireInstanceNumber(groupID string) (int, error) {
	// TODO(RainbowMango) finish implementation later

	return 0, nil
}

func (csm *cloudServiceManager) GetInstances(groupID string) ([]cloudprovider.Instance, error) {
	asClient := csm.getASClientFunc()
	if asClient == nil {
		return nil, fmt.Errorf("failed to list scaling groups due to can not get as client")
	}

	// SDK 'ListScalingInstances' only return no more than 20 instances.
	// If there is a need in the future, need to retrieve by pages.
	opts := &huaweicloudsdkasmodel.ListScalingInstancesRequest{
		ScalingGroupId: groupID,
	}
	response, err := asClient.ListScalingInstances(opts)
	if err != nil {
		klog.Errorf("failed to list scaling group instances. group: %s, error: %v", groupID, err)
		return nil, err
	}
	if response == nil || response.ScalingGroupInstances == nil {
		klog.Infof("no instance in scaling group: %s", groupID)
		return nil, nil
	}

	instances := make([]cloudprovider.Instance, 0, len(*response.ScalingGroupInstances))
	for _, sgi := range *response.ScalingGroupInstances {
		instance := cloudprovider.Instance{
			Id:     *sgi.InstanceId,
			Status: csm.transformInstanceState(*sgi.LifeCycleState, *sgi.HealthStatus),
		}
		instances = append(instances, instance)
	}

	return instances, nil
}

func (csm *cloudServiceManager) IncreaseSizeInstance(groupID string, delta int) error {
	// TODO(RainbowMango) finish implementation later
	return nil
}

func (csm *cloudServiceManager) ListScalingGroups() ([]AutoScalingGroup, error) {
	asClient := csm.getASClientFunc()
	if asClient == nil {
		return nil, fmt.Errorf("failed to list scaling groups due to can not get as client")
	}

	requiredState := huaweicloudsdkasmodel.GetListScalingGroupsRequestScalingGroupStatusEnum().INSERVICE
	opts := &huaweicloudsdkasmodel.ListScalingGroupsRequest{
		ScalingGroupStatus: &requiredState,
	}
	response, err := asClient.ListScalingGroups(opts)
	if err != nil {
		klog.Errorf("failed to list scaling groups. error: %v", err)
		return nil, err
	}

	if response == nil || response.ScalingGroups == nil {
		klog.Info("no scaling group yet.")
		return nil, nil
	}

	autoScalingGroups := make([]AutoScalingGroup, 0, len(*response.ScalingGroups))
	for _, sg := range *response.ScalingGroups {
		autoScalingGroup := newAutoScalingGroup(csm, sg)
		autoScalingGroups = append(autoScalingGroups, autoScalingGroup)
		klog.Infof("found autoscaling group: %s", autoScalingGroup.groupName)
	}

	return autoScalingGroups, nil
}

func (csm *cloudServiceManager) transformInstanceState(lifeCycleState huaweicloudsdkasmodel.ScalingGroupInstanceLifeCycleState,
	healthStatus huaweicloudsdkasmodel.ScalingGroupInstanceHealthStatus) *cloudprovider.InstanceStatus {
	instanceStatus := &cloudprovider.InstanceStatus{}

	lifeCycleStateEnum := huaweicloudsdkasmodel.GetScalingGroupInstanceLifeCycleStateEnum()
	switch lifeCycleState {
	case lifeCycleStateEnum.INSERVICE:
		instanceStatus.State = cloudprovider.InstanceRunning
	case lifeCycleStateEnum.PENDING:
		instanceStatus.State = cloudprovider.InstanceCreating
	case lifeCycleStateEnum.PENDING_WAIT:
		instanceStatus.State = cloudprovider.InstanceCreating
	case lifeCycleStateEnum.REMOVING:
		instanceStatus.State = cloudprovider.InstanceDeleting
	case lifeCycleStateEnum.REMOVING_WAIT:
		instanceStatus.State = cloudprovider.InstanceDeleting
	default:
		instanceStatus.ErrorInfo = &cloudprovider.InstanceErrorInfo{
			ErrorClass:   cloudprovider.OtherErrorClass,
			ErrorMessage: fmt.Sprintf("invalid instance lifecycle state: %v", lifeCycleState),
		}
		return instanceStatus
	}

	healthStatusEnum := huaweicloudsdkasmodel.GetScalingGroupInstanceHealthStatusEnum()
	switch healthStatus {
	case healthStatusEnum.NORMAL:
	case healthStatusEnum.INITAILIZING:
	case healthStatusEnum.ERROR:
		instanceStatus.ErrorInfo = &cloudprovider.InstanceErrorInfo{
			ErrorClass:   cloudprovider.OtherErrorClass,
			ErrorMessage: fmt.Sprintf("%v", healthStatus),
		}
		return instanceStatus
	default:
		instanceStatus.ErrorInfo = &cloudprovider.InstanceErrorInfo{
			ErrorClass:   cloudprovider.OtherErrorClass,
			ErrorMessage: fmt.Sprintf("invalid instance health state: %v", healthStatus),
		}
		return instanceStatus
	}

	return instanceStatus
}
