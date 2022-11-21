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
	"math/rand"
	"strconv"
	"strings"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	huaweicloudsdkas "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/services/as/v1"
	huaweicloudsdkasmodel "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/services/as/v1/model"
	huaweicloudsdkecs "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2"
	huaweicloudsdkecsmodel "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2/model"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
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

	// GetAsgForInstance returns auto scaling group for the given instance.
	GetAsgForInstance(instanceID string) (*AutoScalingGroup, error)

	// RegisterAsg registers auto scaling group to manager
	RegisterAsg(asg *AutoScalingGroup)

	// DeleteScalingInstances is used to delete instances from auto scaling group by instanceIDs.
	DeleteScalingInstances(groupID string, instanceIds []string) error

	// Get default auto scaling group template
	getAsgTemplate(groupID string) (*asgTemplate, error)

	// buildNodeFromTemplate returns template from instance flavor
	buildNodeFromTemplate(asgName string, template *asgTemplate) (*apiv1.Node, error)
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
	asgs             *autoScalingGroupCache
}

type asgTemplate struct {
	name   string
	vcpu   int64
	ram    int64
	gpu    int64
	region string
	zone   string
	tags   map[string]string
}

func newCloudServiceManager(cloudConfig *CloudConfig) *cloudServiceManager {
	csm := &cloudServiceManager{
		cloudConfig:      cloudConfig,
		getECSClientFunc: cloudConfig.getECSClient,
		getASClientFunc:  cloudConfig.getASClient,
		asgs:             newAutoScalingGroupCache(),
	}

	csm.asgs.generateCache(csm)

	return csm
}

func (csm *cloudServiceManager) GetAsgForInstance(instanceID string) (*AutoScalingGroup, error) {
	return csm.asgs.FindForInstance(instanceID, csm)
}

func (csm *cloudServiceManager) RegisterAsg(asg *AutoScalingGroup) {
	csm.asgs.Register(asg)
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
	Instances, err := csm.GetInstances(groupID)
	if err != nil {
		klog.Errorf("failed to list scaling group instances. group: %s, error: %v", groupID, err)
		return 0, fmt.Errorf("failed to get instance list")
	}

	// Get desire instance number by total instance number minus the one which is in deleting state.
	desireInstanceNumber := len(Instances)
	for _, instance := range Instances {
		if instance.Status.State == cloudprovider.InstanceDeleting {
			desireInstanceNumber--
		}
	}

	return desireInstanceNumber, nil
}

func (csm *cloudServiceManager) GetInstances(groupID string) ([]cloudprovider.Instance, error) {
	scalingGroupInstances, err := csm.ListScalingInstances(groupID)
	if err != nil {
		return nil, err
	}

	instances := make([]cloudprovider.Instance, 0, len(scalingGroupInstances))
	for _, sgi := range scalingGroupInstances {
		// When a new instance joining to the scaling group, the instance id maybe empty(nil).
		if sgi.InstanceId == nil {
			klog.Infof("ignore instance without instance id, maybe instance is joining.")
			continue
		}
		instance := cloudprovider.Instance{
			Id:     *sgi.InstanceId,
			Status: csm.transformInstanceState(*sgi.LifeCycleState, *sgi.HealthStatus),
		}
		instances = append(instances, instance)
	}

	return instances, nil
}

func (csm *cloudServiceManager) ListScalingInstances(groupID string) ([]huaweicloudsdkasmodel.ScalingGroupInstance, error) {
	asClient := csm.getASClientFunc()
	if asClient == nil {
		return nil, fmt.Errorf("failed to list scaling groups due to can not get as client")
	}

	var scalingGroupInstances []huaweicloudsdkasmodel.ScalingGroupInstance
	var startNumber int32 = 0
	for {
		opts := &huaweicloudsdkasmodel.ListScalingInstancesRequest{
			ScalingGroupId: groupID,
			StartNumber:    &startNumber,
		}
		response, err := asClient.ListScalingInstances(opts)
		if err != nil {
			klog.Errorf("failed to list scaling group instances. group: %s, error: %v", groupID, err)
			return nil, err
		}
		if response == nil {
			klog.Errorf("Unexpected get nil response when listing instances.")
			return nil, fmt.Errorf("unexpected get nil response when listing instances")
		}
		if response.ScalingGroupInstances == nil {
			klog.Errorf("Unexpected get nil scaling group instances")
			return nil, fmt.Errorf("unexpected get nil scaling group instances")
		}

		klog.Infof("Got %d instances from scaling group, total instances: %d", len(*response.ScalingGroupInstances), *response.TotalNumber)
		scalingGroupInstances = append(scalingGroupInstances, *response.ScalingGroupInstances...)

		// break once we get all instances
		if response.TotalNumber != nil && len(scalingGroupInstances) == int(*response.TotalNumber) {
			break
		}

		// get ready to request next page
		startNumber += int32(len(*response.ScalingGroupInstances))
	}

	return scalingGroupInstances, nil
}

func (csm *cloudServiceManager) DeleteScalingInstances(groupID string, instanceIds []string) error {
	asClient := csm.getASClientFunc()

	instanceDeleteBatchRemoveInstancesOption := huaweicloudsdkasmodel.GetBatchRemoveInstancesOptionInstanceDeleteEnum().YES

	opts := &huaweicloudsdkasmodel.BatchRemoveScalingInstancesRequest{
		ScalingGroupId: groupID,
		Body: &huaweicloudsdkasmodel.BatchRemoveInstancesOption{
			InstancesId:    instanceIds,
			InstanceDelete: &instanceDeleteBatchRemoveInstancesOption,
			Action:         huaweicloudsdkasmodel.GetBatchRemoveInstancesOptionActionEnum().REMOVE,
		},
	}

	_, err := asClient.BatchRemoveScalingInstances(opts)

	if err != nil {
		klog.Errorf("failed to delete scaling instances. group: %s, error: %v", groupID, err)
		return err
	}

	return nil
}

// IncreaseSizeInstance increases a scaling group's instance size.
// The workflow works as follows:
// 1. create scaling policy with scheduled type.
// 2. execute the scaling policy immediately(not waiting the policy's launch time).
// 3. wait for the instance number be increased and remove the scaling policy.
func (csm *cloudServiceManager) IncreaseSizeInstance(groupID string, delta int) error {
	originalInstanceSize, err := csm.GetDesireInstanceNumber(groupID)
	if err != nil {
		return err
	}

	// create a scaling policy
	addOperation := huaweicloudsdkasmodel.GetScalingPolicyActionV1OperationEnum().ADD
	instanceNum := int32(delta)
	opts := &huaweicloudsdkasmodel.CreateScalingPolicyRequest{
		Body: &huaweicloudsdkasmodel.CreateScalingPolicyOption{
			// It's not mandatory for AS service to set a unique policy name.
			ScalingPolicyName: "huaweicloudautoscaler",
			ScalingGroupId:    groupID,
			ScalingPolicyType: huaweicloudsdkasmodel.GetCreateScalingPolicyOptionScalingPolicyTypeEnum().SCHEDULED,
			ScheduledPolicy: &huaweicloudsdkasmodel.ScheduledPolicy{
				LaunchTime: time.Now().Add(time.Hour).UTC().Format("2006-01-02T15:04Z"),
			},
			ScalingPolicyAction: &huaweicloudsdkasmodel.ScalingPolicyActionV1{
				Operation:      &addOperation,
				InstanceNumber: &instanceNum,
			},
		},
	}

	spID, err := csm.createScalingPolicy(opts)
	if err != nil {
		return err
	}

	// make sure scaling policy will be cleaned up.
	deletePolicyOps := &huaweicloudsdkasmodel.DeleteScalingPolicyRequest{
		ScalingPolicyId: spID,
	}
	defer csm.deleteScalingPolicy(deletePolicyOps)

	// execute policy immediately
	executeAction := huaweicloudsdkasmodel.GetExecuteScalingPolicyOptionActionEnum().EXECUTE
	executeOpts := &huaweicloudsdkasmodel.ExecuteScalingPolicyRequest{
		ScalingPolicyId: spID,
		Body: &huaweicloudsdkasmodel.ExecuteScalingPolicyOption{
			Action: executeAction,
		},
	}
	err = csm.executeScalingPolicy(executeOpts)
	if err != nil {
		return err
	}

	// wait for instance number indeed be increased
	return wait.Poll(5*time.Second, 30*time.Minute, func() (done bool, err error) {
		currentInstanceSize, err := csm.GetDesireInstanceNumber(groupID)
		if err != nil {
			return false, err
		}

		if currentInstanceSize == originalInstanceSize+delta {
			return true, nil
		}
		klog.V(1).Infof("waiting instance increase from %d to %d, now is: %d", originalInstanceSize, originalInstanceSize+delta, currentInstanceSize)

		return false, nil
	})
}

func (csm *cloudServiceManager) ListScalingGroups() ([]AutoScalingGroup, error) {
	asClient := csm.getASClientFunc()
	if asClient == nil {
		return nil, fmt.Errorf("failed to list scaling groups due to can not get as client")
	}

	// requiredState := huaweicloudsdkasmodel.GetListScalingGroupsRequestScalingGroupStatusEnum().INSERVICE
	opts := &huaweicloudsdkasmodel.ListScalingGroupsRequest{
		// ScalingGroupStatus: &requiredState,
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

func (csm *cloudServiceManager) createScalingPolicy(opts *huaweicloudsdkasmodel.CreateScalingPolicyRequest) (scalingPolicyID string, err error) {
	asClient := csm.getASClientFunc()
	if asClient == nil {
		return "", fmt.Errorf("failed to get as client")
	}

	response, err := asClient.CreateScalingPolicy(opts)
	if err != nil {
		klog.Warningf("create scaling policy failed. policy: %s, error: %v", opts.String(), err)
		return "", err
	}

	klog.V(1).Infof("create scaling policy succeed. policy id: %s", *(response.ScalingPolicyId))

	return *(response.ScalingPolicyId), nil
}

func (csm *cloudServiceManager) executeScalingPolicy(opts *huaweicloudsdkasmodel.ExecuteScalingPolicyRequest) error {
	asClient := csm.getASClientFunc()
	if asClient == nil {
		return fmt.Errorf("failed to get as client")
	}

	_, err := asClient.ExecuteScalingPolicy(opts)
	if err != nil {
		klog.Warningf("execute scaling policy failed. policy id: %s, error: %v", opts.ScalingPolicyId, err)
		return err
	}

	klog.V(1).Infof("execute scaling policy succeed. policy id: %s", opts.ScalingPolicyId)
	return nil
}

func (csm *cloudServiceManager) deleteScalingPolicy(opts *huaweicloudsdkasmodel.DeleteScalingPolicyRequest) error {
	asClient := csm.getASClientFunc()
	if asClient == nil {
		return fmt.Errorf("failed to get as client")
	}

	_, err := asClient.DeleteScalingPolicy(opts)
	if err != nil {
		klog.Warningf("failed to delete scaling policy. policy id: %s, error: %v", opts.ScalingPolicyId, err)
		return err
	}

	klog.V(1).Infof("delete scaling policy succeed. policy id: %s", opts.ScalingPolicyId)
	return nil
}

func (csm *cloudServiceManager) getScalingGroupByID(groupID string) (*huaweicloudsdkasmodel.ScalingGroups, error) {
	asClient := csm.getASClientFunc()
	opts := &huaweicloudsdkasmodel.ShowScalingGroupRequest{
		ScalingGroupId: groupID,
	}
	response, err := asClient.ShowScalingGroup(opts)
	if err != nil {
		klog.Errorf("failed to show scaling group info. group: %s, error: %v", groupID, err)
		return nil, err
	}
	if response == nil || response.ScalingGroup == nil {
		return nil, fmt.Errorf("no scaling group found: %s", groupID)
	}

	return response.ScalingGroup, nil
}

func (csm *cloudServiceManager) listScalingTagsByID(groupID string) (map[string]string, error) {
	asClient := csm.getASClientFunc()
	opts := &huaweicloudsdkasmodel.ListScalingTagInfosByResourceIdRequest{
		ResourceType: huaweicloudsdkasmodel.GetListScalingTagInfosByResourceIdRequestResourceTypeEnum().SCALING_GROUP_TAG,
		ResourceId:   groupID,
	}
	response, err := asClient.ListScalingTagInfosByResourceId(opts)
	if err != nil {
		klog.Errorf("failed to list scaling group tags. scaling group id: %s, error: %v", groupID, err)
		return nil, err
	}
	if response == nil || response.Tags == nil {
		klog.Infof("Not tags found for scaling group by id:%s", groupID)
		return nil, nil
	}

	tags := make(map[string]string)
	for _, tag := range *response.Tags {
		if tag.Value != nil {
			tags[tag.Key] = *tag.Value
			continue
		}
		tags[tag.Key] = ""
	}

	return tags, nil
}

func (csm *cloudServiceManager) getScalingGroupConfigByID(groupID, configID string) (*huaweicloudsdkasmodel.ScalingConfiguration, error) {
	asClient := csm.getASClientFunc()
	opts := &huaweicloudsdkasmodel.ShowScalingConfigRequest{
		ScalingConfigurationId: configID,
	}
	response, err := asClient.ShowScalingConfig(opts)
	if err != nil {
		klog.Errorf("failed to show scaling group config. config id: %s, error: %v", configID, err)
		return nil, err
	}
	if response == nil || response.ScalingConfiguration == nil {
		return nil, fmt.Errorf("no scaling configuration found, groupID: %s, configID: %s", groupID, configID)
	}
	return response.ScalingConfiguration, nil
}

func (csm *cloudServiceManager) listFlavors(az string) (*[]huaweicloudsdkecsmodel.Flavor, error) {
	ecsClient := csm.getECSClientFunc()
	opts := &huaweicloudsdkecsmodel.ListFlavorsRequest{
		AvailabilityZone: &az,
	}
	response, err := ecsClient.ListFlavors(opts)
	if err != nil {
		klog.Errorf("failed to list flavors. availability zone: %s", az)
		return nil, err
	}

	return response.Flavors, nil
}

func (csm *cloudServiceManager) getAsgTemplate(groupID string) (*asgTemplate, error) {
	sg, err := csm.getScalingGroupByID(groupID)
	if err != nil {
		klog.Errorf("failed to get ASG by id:%s,because of %s", groupID, err.Error())
		return nil, err
	}

	configuration, err := csm.getScalingGroupConfigByID(groupID, *sg.ScalingConfigurationId)
	if err != nil {
		klog.Errorf("failed to get scaling group config by id:%s", *sg.ScalingConfigurationId)
		return nil, err
	}
	tags, err := csm.listScalingTagsByID(groupID)
	if err != nil {
		klog.Errorf("failed to list scaling tags by id:%s", groupID)
		return nil, err
	}
	for _, az := range *sg.AvailableZones {
		flavors, err := csm.listFlavors(az)
		if err != nil {
			klog.Errorf("failed to list flavors, available zone is: %s, error: %v", az, err)
			return nil, err
		}

		for _, flavor := range *flavors {
			// FlavorRef is in format like: c7.4xlarge.2,c7.4xlarge.4,c7.6xlarge.2
			if !strings.Contains(*configuration.InstanceConfig.FlavorRef, flavor.Name) {
				continue
			}

			vcpus, _ := strconv.ParseInt(flavor.Vcpus, 10, 64)
			return &asgTemplate{
				name: flavor.Name,
				vcpu: vcpus,
				ram:  int64(flavor.Ram),
				zone: az,
				tags: tags,
			}, nil
		}
	}
	return nil, fmt.Errorf("no available instance flavor:%s found in as group:%s", *configuration.InstanceConfig.FlavorRef, groupID)
}

func (csm *cloudServiceManager) buildNodeFromTemplate(asgName string, template *asgTemplate) (*apiv1.Node, error) {
	node := apiv1.Node{}
	nodeName := fmt.Sprintf("%s-asg-%d", asgName, rand.Int63())

	node.ObjectMeta = metav1.ObjectMeta{
		Name:     nodeName,
		SelfLink: fmt.Sprintf("/api/v1/nodes/%s", nodeName),
		Labels:   map[string]string{},
	}

	node.Status = apiv1.NodeStatus{
		Capacity: apiv1.ResourceList{},
	}

	node.Status.Capacity[apiv1.ResourcePods] = *resource.NewQuantity(110, resource.DecimalSI)
	node.Status.Capacity[apiv1.ResourceCPU] = *resource.NewQuantity(template.vcpu, resource.DecimalSI)
	node.Status.Capacity[gpu.ResourceNvidiaGPU] = *resource.NewQuantity(template.gpu, resource.DecimalSI)
	node.Status.Capacity[apiv1.ResourceMemory] = *resource.NewQuantity(template.ram*1024*1024, resource.DecimalSI)

	node.Status.Allocatable = node.Status.Capacity

	node.Labels = cloudprovider.JoinStringMaps(node.Labels, buildGenericLabels(template, nodeName))

	node.Spec.Taints = extractTaintsFromTags(template.tags)

	node.Status.Conditions = cloudprovider.BuildReadyConditions()
	return &node, nil
}

// extractTaintsFromTags extract taints from as group tags.
// The tag is of the format "k8s.io_cluster-autoscaler_node-template_taint_<taint-key>". "<taint-key>" is
// the name of the taint and the value of each tag specifies the taint value and effect with the
// format "<taint-value>:<taint-effect>".
// Example tags: "k8s.io_cluster-autoscaler_node-template_taint_dedicated": "true:NoSchedule"
func extractTaintsFromTags(tags map[string]string) []apiv1.Taint {
	taints := make([]apiv1.Taint, 0)

	for tagKey, tagValue := range tags {
		if !strings.Contains(tagKey, "k8s.io_cluster-autoscaler_node-template_taint_") {
			continue
		}

		splits := strings.Split(tagKey, "k8s.io_cluster-autoscaler_node-template_taint_")
		// If the tagKey is 'k8s.io_cluster-autoscaler_node-template_taint_', the second element is '',
		// this should be ruled out.
		if len(splits) < 2 || splits[1] == "" {
			klog.Warningf("Invalid tag key format:%s", tagKey)
			continue
		}

		values := strings.Split(tagValue, ":")
		if len(values) != 2 {
			klog.Warningf("Invalid tag value format:%s", tagValue)
			continue
		}

		if values[1] != string(apiv1.TaintEffectNoSchedule) &&
			values[1] != string(apiv1.TaintEffectPreferNoSchedule) &&
			values[1] != string(apiv1.TaintEffectNoExecute) {
			klog.Warningf("Invalid tag value format:%s", tagValue)
			continue
		}

		taints = append(taints, apiv1.Taint{
			Key:    splits[1],
			Value:  values[0],
			Effect: apiv1.TaintEffect(values[1]),
		})
		klog.V(6).Infof("Extract taints from tag key/value successfully:%s, %s", tagKey, tagValue)
	}

	return taints
}

func buildGenericLabels(template *asgTemplate, nodeName string) map[string]string {
	result := make(map[string]string)
	result[apiv1.LabelArchStable] = cloudprovider.DefaultArch
	result[apiv1.LabelOSStable] = cloudprovider.DefaultOS

	result[apiv1.LabelInstanceTypeStable] = template.name

	result[apiv1.LabelTopologyRegion] = template.region
	result[apiv1.LabelTopologyZone] = template.zone
	result[apiv1.LabelHostname] = nodeName

	// append custom node labels
	for key, value := range template.tags {
		// ignore the tag which represents a taint
		if strings.Contains(key, "k8s.io_cluster-autoscaler_node-template_taint_") {
			continue
		}
		result[key] = value
	}

	return result
}
