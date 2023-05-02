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

package volcengine

import (
	"fmt"
	"math/rand"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volcengine-go-sdk/volcengine"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/klog/v2"
)

// VolcengineManager define the interface that implements Cloud Provider and Node Group
type VolcengineManager interface {
	// RegisterAsg registers the given ASG with the manager.
	RegisterAsg(asg *AutoScalingGroup)

	// GetAsgForInstance returns the ASG of the given instance.
	GetAsgForInstance(instanceId string) (*AutoScalingGroup, error)

	// GetAsgById returns the ASG of the given id.
	GetAsgById(id string) (*AutoScalingGroup, error)

	// GetAsgDesireCapacity returns the desired capacity of the given ASG.
	GetAsgDesireCapacity(asgId string) (int, error)

	// SetAsgTargetSize sets the target size of the given ASG.
	SetAsgTargetSize(asgId string, targetSize int) error

	// DeleteScalingInstances deletes the given instances from the given ASG.
	DeleteScalingInstances(asgId string, instanceIds []string) error

	// GetAsgNodes returns the scaling instance ids of the given ASG.
	GetAsgNodes(asgId string) ([]cloudprovider.Instance, error)

	// SetAsgDesireCapacity sets the desired capacity of the given ASG.
	SetAsgDesireCapacity(groupId string, desireCapacity int) error

	// getAsgTemplate returns the scaling configuration of the given ASG.
	getAsgTemplate(groupId string) (*asgTemplate, error)

	// buildNodeFromTemplateName builds a node object from the given template.
	buildNodeFromTemplateName(asgName string, template *asgTemplate) (*apiv1.Node, error)
}

type asgTemplate struct {
	vcpu         int64
	memInMB      int64
	gpu          int64
	region       string
	zone         string
	instanceType string
	tags         map[string]string
}

// volcengineManager handles volcengine service communication.
type volcengineManager struct {
	cloudConfig *cloudConfig
	asgs        *autoScalingGroupsCache

	asgService AutoScalingService
	ecsService EcsService
}

func (v *volcengineManager) SetAsgDesireCapacity(groupId string, desireCapacity int) error {
	return v.asgService.SetAsgDesireCapacity(groupId, desireCapacity)
}

func (v *volcengineManager) GetAsgDesireCapacity(asgId string) (int, error) {
	group, err := v.asgService.GetScalingGroupById(asgId)
	if err != nil {
		klog.Errorf("failed to get scaling group by id %s: %v", asgId, err)
		return 0, err
	}
	return int(volcengine.Int32Value(group.DesireInstanceNumber)), nil
}

func (v *volcengineManager) SetAsgTargetSize(asgId string, targetSize int) error {
	return v.asgService.SetAsgTargetSize(asgId, targetSize)
}

func (v *volcengineManager) DeleteScalingInstances(asgId string, instanceIds []string) error {
	if len(instanceIds) == 0 {
		klog.Infof("no instances to delete from scaling group %s", asgId)
		return nil
	}
	klog.Infof("deleting instances %v from scaling group %s", instanceIds, asgId)
	return v.asgService.RemoveInstances(asgId, instanceIds)
}

func (v *volcengineManager) GetAsgNodes(asgId string) ([]cloudprovider.Instance, error) {
	scalingInstances, err := v.asgService.ListScalingInstancesByGroupId(asgId)
	if err != nil {
		return nil, err
	}

	instances := make([]cloudprovider.Instance, 0, len(scalingInstances))
	for _, scalingInstance := range scalingInstances {
		if scalingInstance.InstanceId == nil {
			klog.Warningf("scaling instance has no instance id")
			continue
		}

		instances = append(instances, cloudprovider.Instance{
			Id: getNodeProviderId(volcengine.StringValue(scalingInstance.InstanceId)),
		})
	}
	return instances, nil
}

func getNodeProviderId(instanceId string) string {
	return fmt.Sprintf("volcengine://%s", instanceId)
}

func (v *volcengineManager) getAsgTemplate(groupId string) (*asgTemplate, error) {
	group, err := v.asgService.GetScalingGroupById(groupId)
	if err != nil {
		klog.Errorf("failed to get scaling group by id %s: %v", groupId, err)
		return nil, err
	}

	configuration, err := v.asgService.GetScalingConfigurationById(volcengine.StringValue(group.ActiveScalingConfigurationId))
	if err != nil {
		klog.Errorf("failed to get scaling configuration by id %s: %v", volcengine.StringValue(group.ActiveScalingConfigurationId), err)
		return nil, err
	}

	instanceType, err := v.ecsService.GetInstanceTypeById(volcengine.StringValue(configuration.InstanceTypes[0]))
	if err != nil {
		klog.Errorf("failed to get instance type by id %s: %v", volcengine.StringValue(configuration.InstanceTypes[0]), err)
		return nil, err
	}

	return &asgTemplate{
		vcpu:         int64(volcengine.Int32Value(instanceType.Processor.Cpus)),
		memInMB:      int64(volcengine.Int32Value(instanceType.Memory.Size)),
		region:       v.cloudConfig.getRegion(),
		instanceType: volcengine.StringValue(instanceType.InstanceTypeId),
		tags:         map[string]string{}, // TODO read tags from configuration
	}, nil
}

func (v *volcengineManager) buildNodeFromTemplateName(asgName string, template *asgTemplate) (*apiv1.Node, error) {
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
	node.Status.Capacity[apiv1.ResourceMemory] = *resource.NewQuantity(template.memInMB*1024*1024, resource.DecimalSI)
	node.Status.Capacity[gpu.ResourceNvidiaGPU] = *resource.NewQuantity(template.gpu, resource.DecimalSI)

	node.Status.Allocatable = node.Status.Capacity

	node.Labels = cloudprovider.JoinStringMaps(node.Labels, buildGenericLabels(template, nodeName))

	node.Status.Conditions = cloudprovider.BuildReadyConditions()
	return &node, nil
}

func buildGenericLabels(template *asgTemplate, nodeName string) map[string]string {
	result := make(map[string]string)
	result[apiv1.LabelArchStable] = cloudprovider.DefaultArch
	result[apiv1.LabelOSStable] = cloudprovider.DefaultOS

	result[apiv1.LabelInstanceTypeStable] = template.instanceType

	result[apiv1.LabelTopologyRegion] = template.region
	result[apiv1.LabelTopologyZone] = template.zone
	result[apiv1.LabelHostname] = nodeName

	// append custom node labels
	for key, value := range template.tags {
		result[key] = value
	}

	return result
}

func (v *volcengineManager) GetAsgById(id string) (*AutoScalingGroup, error) {
	asg, err := v.asgService.GetScalingGroupById(id)
	if err != nil {
		return nil, err
	}
	return &AutoScalingGroup{
		manager:           v,
		asgId:             volcengine.StringValue(asg.ScalingGroupId),
		minInstanceNumber: int(volcengine.Int32Value(asg.MinInstanceNumber)),
		maxInstanceNumber: int(volcengine.Int32Value(asg.MaxInstanceNumber)),
	}, nil
}

func (v *volcengineManager) GetAsgForInstance(instanceId string) (*AutoScalingGroup, error) {
	return v.asgs.FindForInstance(instanceId)
}

func (v *volcengineManager) RegisterAsg(asg *AutoScalingGroup) {
	v.asgs.Register(asg)
}

// CreateVolcengineManager returns the VolcengineManager interface implementation
func CreateVolcengineManager(cloudConfig *cloudConfig) (VolcengineManager, error) {
	asgCloudService := newAutoScalingService(cloudConfig)
	return &volcengineManager{
		cloudConfig: cloudConfig,
		asgs:        newAutoScalingGroupsCache(asgCloudService),
		asgService:  asgCloudService,
		ecsService:  newEcsService(cloudConfig),
	}, nil
}
