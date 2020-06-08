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

package alicloud

import (
	"errors"
	"fmt"
	"gopkg.in/gcfg.v1"
	"io"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/alicloud/alibaba-cloud-sdk-go/services/ess"
	klog "k8s.io/klog/v2"
	kubeletapis "k8s.io/kubernetes/pkg/kubelet/apis"
	"math/rand"
	"time"
)

const (
	sdkCoolDownTimeout     = 200 * time.Millisecond
	defaultPodAmountsLimit = 110
	//ResourceGPU GPU resource type
	ResourceGPU apiv1.ResourceName = "nvidia.com/gpu"
)

type asgInformation struct {
	config   *Asg
	basename string
}

// AliCloudManager handles alicloud service communication.
type AliCloudManager struct {
	cfg      *cloudConfig
	aService *autoScalingWrapper
	iService *instanceWrapper
	asgs     *autoScalingGroups
}

type sgTemplate struct {
	InstanceType *instanceType
	Region       string
	Zone         string
	Tags         map[string]string
}

// CreateAliCloudManager constructs aliCloudManager object.
func CreateAliCloudManager(configReader io.Reader) (*AliCloudManager, error) {
	cfg := &cloudConfig{}
	if configReader != nil {
		if err := gcfg.ReadInto(cfg, configReader); err != nil {
			klog.Errorf("couldn't read config: %v", err)
			return nil, err
		}
	}
	if cfg.isValid() == false {
		return nil, errors.New("please check whether you have provided correct AccessKeyId,AccessKeySecret,RegionId or STS Token")
	}

	asw, err := newAutoScalingWrapper(cfg)
	if err != nil {
		klog.Errorf("failed to create NewAutoScalingWrapper because of %s", err.Error())
		return nil, err
	}
	iw, err := newInstanceWrapper(cfg)
	if err != nil {
		klog.Errorf("failed to create NewInstanceWrapper because of %s", err.Error())
		return nil, err
	}

	manager := &AliCloudManager{
		cfg:      cfg,
		asgs:     newAutoScalingGroups(asw),
		aService: asw,
		iService: iw,
	}
	return manager, nil
}

// RegisterAsg registers asg in AliCloud Manager.
func (m *AliCloudManager) RegisterAsg(asg *Asg) {
	m.asgs.Register(asg)
}

// GetAsgForInstance returns AsgConfig of the given Instance
func (m *AliCloudManager) GetAsgForInstance(instanceId string) (*Asg, error) {
	return m.asgs.FindForInstance(instanceId)
}

// GetAsgSize gets ASG size.
func (m *AliCloudManager) GetAsgSize(asgConfig *Asg) (int64, error) {
	sg, err := m.aService.getScalingGroupByID(asgConfig.id)
	if err != nil {
		return -1, fmt.Errorf("failed to describe ASG %s,Because of %s", asgConfig.id, err.Error())
	}
	return int64(sg.ActiveCapacity + sg.PendingCapacity), nil
}

// SetAsgSize sets ASG size.
func (m *AliCloudManager) SetAsgSize(asg *Asg, size int64) error {
	return m.aService.setCapcityInstanceSize(asg.id, size)
}

// DeleteInstances deletes the given instances. All instances must be controlled by the same ASG.
func (m *AliCloudManager) DeleteInstances(instanceIds []string) error {
	klog.Infof("start to remove Instances from ASG %v", instanceIds)
	if len(instanceIds) == 0 {
		klog.Warningf("you don't provide any instanceIds to remove")
		return nil
	}
	// Check whether instances are in the same group
	// TODO: remove or provide more meaningful check method.
	commonAsg, err := m.asgs.FindForInstance(instanceIds[0])
	if err != nil {
		klog.Errorf("failed to find instance:%s in ASG", instanceIds[0])
		return err
	}
	for _, instanceId := range instanceIds {
		asg, err := m.asgs.FindForInstance(instanceId)
		if err != nil {
			klog.Errorf("failed to find instanceId:%s from ASG and exit", instanceId)
			return err
		}
		if asg != commonAsg {
			return fmt.Errorf("cannot delete instances which doesn't belong to the same ASG")
		}
	}
	// Remove instance from ASG in loop
	for _, instanceId := range instanceIds {
		req := ess.CreateRemoveInstancesRequest()
		req.ScalingGroupId = commonAsg.id
		req.InstanceId1 = instanceId

		resp, err := m.aService.RemoveInstances(req)
		if err != nil {
			klog.Errorf("failed to remove instance from scaling group %s,because of %s", commonAsg.id, err.Error())
			continue
		}
		klog.Infof("remove instances successfully with response: %s", resp.GetHttpContentString())
		// prevent from triggering api flow control
		time.Sleep(sdkCoolDownTimeout)
	}
	return nil
}

// GetAsgNodes returns Asg nodes.
func (m *AliCloudManager) GetAsgNodes(sg *Asg) ([]string, error) {
	result := make([]string, 0)
	instances, err := m.aService.getScalingInstancesByGroup(sg.id)
	if err != nil {
		return []string{}, err
	}
	for _, instance := range instances {
		result = append(result, getNodeProviderID(instance.InstanceId, sg.RegionId()))
	}
	return result, nil
}

// getNodeProviderID build provider id from ecs id and region
func getNodeProviderID(id, region string) string {
	return fmt.Sprintf("%s.%s", region, id)
}

func (m *AliCloudManager) getAsgTemplate(asgId string) (*sgTemplate, error) {
	sg, err := m.aService.getScalingGroupByID(asgId)
	if err != nil {
		klog.Errorf("failed to get ASG by id:%s,because of %s", asgId, err.Error())
		return nil, err
	}

	configuration, err := m.aService.getScalingGroupConfigurationByID(sg.ActiveScalingConfigurationId, asgId)
	if err != nil {
		return nil, err
	}

	instanceType, err := m.iService.getInstanceTypeById(configuration.InstanceType)
	if err != nil {
		klog.Errorf("failed to get instanceType by Id:%s,because of %s", configuration.InstanceType, err.Error())
		return nil, err
	}

	tags, err := m.iService.getInstanceTags(configuration.Tags)
	if err != nil {
		klog.Errorf("failed to getInstanceTags from scalingGroup %s,because of %s", asgId, err.Error())
		return nil, err
	}

	return &sgTemplate{
		InstanceType: instanceType,
		Region:       sg.RegionId,
		Tags:         tags,
	}, nil
}

func (m *AliCloudManager) buildNodeFromTemplate(sg *Asg, template *sgTemplate) (*apiv1.Node, error) {
	node := apiv1.Node{}
	nodeName := fmt.Sprintf("%s-asg-%d", sg.id, rand.Int63())

	node.ObjectMeta = metav1.ObjectMeta{
		Name:     nodeName,
		SelfLink: fmt.Sprintf("/api/v1/nodes/%s", nodeName),
		Labels:   map[string]string{},
	}

	node.Status = apiv1.NodeStatus{
		Capacity: apiv1.ResourceList{},
	}

	node.Status.Capacity[apiv1.ResourcePods] = *resource.NewQuantity(defaultPodAmountsLimit, resource.DecimalSI)
	node.Status.Capacity[apiv1.ResourceCPU] = *resource.NewQuantity(template.InstanceType.vcpu, resource.DecimalSI)
	node.Status.Capacity[apiv1.ResourceMemory] = *resource.NewQuantity(template.InstanceType.memoryInBytes, resource.DecimalSI)
	//add gpu resource support
	node.Status.Capacity[ResourceGPU] = *resource.NewQuantity(template.InstanceType.gpu, resource.DecimalSI)

	node.Status.Allocatable = node.Status.Capacity

	node.Labels = cloudprovider.JoinStringMaps(node.Labels, buildGenericLabels(template, nodeName))

	node.Status.Conditions = cloudprovider.BuildReadyConditions()
	return &node, nil
}

func buildGenericLabels(template *sgTemplate, nodeName string) map[string]string {
	result := make(map[string]string)
	result[kubeletapis.LabelArch] = cloudprovider.DefaultArch
	result[kubeletapis.LabelOS] = cloudprovider.DefaultOS

	result[apiv1.LabelInstanceType] = template.InstanceType.instanceTypeID

	result[apiv1.LabelZoneRegion] = template.Region
	result[apiv1.LabelZoneFailureDomain] = template.Zone
	result[apiv1.LabelHostname] = nodeName

	// append custom node labels
	for key, value := range template.Tags {
		result[key] = value
	}

	return result
}
