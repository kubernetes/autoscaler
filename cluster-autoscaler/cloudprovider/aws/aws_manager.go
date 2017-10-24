/*
Copyright 2016 The Kubernetes Authors.

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

//go:generate go run ec2_instance_types/gen.go

package aws

import (
	"fmt"
	"io"
	"math/rand"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/golang/glog"
	"gopkg.in/gcfg.v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	apiv1 "k8s.io/kubernetes/pkg/api/v1"
	provider_aws "k8s.io/kubernetes/pkg/cloudprovider/providers/aws"
	kubeletapis "k8s.io/kubernetes/pkg/kubelet/apis"
)

const (
	operationWaitTimeout    = 5 * time.Second
	operationPollInterval   = 100 * time.Millisecond
	maxRecordsReturnedByAPI = 100
)

type asgInformation struct {
	config   *Asg
	basename string
}

// AwsManager is handles aws communication and data caching.
type AwsManager struct {
	service   autoScalingWrapper
	asgs      *autoScalingGroups
	interrupt chan struct{}
}

type asgTemplate struct {
	InstanceType *instanceType
	Region       string
	Zone         string
	Tags         []*autoscaling.TagDescription
}

// CreateAwsManager constructs awsManager object.
func CreateAwsManager(configReader io.Reader) (*AwsManager, error) {
	if configReader != nil {
		var cfg provider_aws.CloudConfig
		if err := gcfg.ReadInto(&cfg, configReader); err != nil {
			glog.Errorf("Couldn't read config: %v", err)
			return nil, err
		}
	}

	service := autoScalingWrapper{
		autoscaling.New(session.New()),
	}
	manager := &AwsManager{
		asgs:      newAutoScalingGroups(service),
		service:   service,
		interrupt: make(chan struct{}),
	}

	go wait.Until(func() {
		manager.asgs.cacheMutex.Lock()
		defer manager.asgs.cacheMutex.Unlock()
		if err := manager.asgs.regenerateCache(); err != nil {
			glog.Errorf("Error while regenerating Asg cache: %v", err)
		}
	}, time.Hour, manager.interrupt)

	return manager, nil
}

func (m *AwsManager) Cleanup() {
	m.interrupt <- struct{}{}
}

// RegisterAsg registers asg in Aws Manager.
func (m *AwsManager) RegisterAsg(asg *Asg) {
	m.asgs.Register(asg)
}

// GetAsgForInstance returns AsgConfig of the given Instance
func (m *AwsManager) GetAsgForInstance(instance *AwsRef) (*Asg, error) {
	return m.asgs.FindForInstance(instance)
}

func (m *AwsManager) getAutoscalingGroupsByTags(keys []string) ([]*autoscaling.Group, error) {
	return m.service.getAutoscalingGroupsByTags(keys)
}

// GetAsgSize gets ASG size.
func (m *AwsManager) GetAsgSize(asgConfig *Asg) (int64, error) {
	params := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{aws.String(asgConfig.Name)},
		MaxRecords:            aws.Int64(1),
	}
	groups, err := m.service.DescribeAutoScalingGroups(params)

	if err != nil {
		return -1, err
	}

	if len(groups.AutoScalingGroups) < 1 {
		return -1, fmt.Errorf("Unable to get first autoscaling.Group for %s", asgConfig.Name)
	}
	asg := *groups.AutoScalingGroups[0]
	return *asg.DesiredCapacity, nil
}

// SetAsgSize sets ASG size.
func (m *AwsManager) SetAsgSize(asg *Asg, size int64) error {
	params := &autoscaling.SetDesiredCapacityInput{
		AutoScalingGroupName: aws.String(asg.Name),
		DesiredCapacity:      aws.Int64(size),
		HonorCooldown:        aws.Bool(false),
	}
	glog.V(0).Infof("Setting asg %s size to %d", asg.Id(), size)
	_, err := m.service.SetDesiredCapacity(params)
	if err != nil {
		return err
	}
	return nil
}

// DeleteInstances deletes the given instances. All instances must be controlled by the same ASG.
func (m *AwsManager) DeleteInstances(instances []*AwsRef) error {
	if len(instances) == 0 {
		return nil
	}
	commonAsg, err := m.asgs.FindForInstance(instances[0])
	if err != nil {
		return err
	}
	for _, instance := range instances {
		asg, err := m.asgs.FindForInstance(instance)
		if err != nil {
			return err
		}
		if asg != commonAsg {
			return fmt.Errorf("Connot delete instances which don't belong to the same ASG.")
		}
	}

	for _, instance := range instances {
		params := &autoscaling.TerminateInstanceInAutoScalingGroupInput{
			InstanceId:                     aws.String(instance.Name),
			ShouldDecrementDesiredCapacity: aws.Bool(true),
		}
		resp, err := m.service.TerminateInstanceInAutoScalingGroup(params)
		if err != nil {
			return err
		}
		glog.V(4).Infof(*resp.Activity.Description)
	}

	return nil
}

// GetAsgNodes returns Asg nodes.
func (m *AwsManager) GetAsgNodes(asg *Asg) ([]string, error) {
	result := make([]string, 0)
	group, err := m.service.getAutoscalingGroupByName(asg.Name)
	if err != nil {
		return []string{}, err
	}
	for _, instance := range group.Instances {
		result = append(result,
			fmt.Sprintf("aws:///%s/%s", *instance.AvailabilityZone, *instance.InstanceId))
	}
	return result, nil
}

func (m *AwsManager) getAsgTemplate(name string) (*asgTemplate, error) {
	asg, err := m.service.getAutoscalingGroupByName(name)
	if err != nil {
		return nil, err
	}

	instanceTypeName, err := m.service.getInstanceTypeByLCName(*asg.LaunchConfigurationName)
	if err != nil {
		return nil, err
	}

	if len(asg.AvailabilityZones) < 1 {
		return nil, fmt.Errorf("Unable to get first AvailabilityZone for %s", name)
	}

	az := *asg.AvailabilityZones[0]
	region := az[0 : len(az)-1]

	if len(asg.AvailabilityZones) > 1 {
		glog.Warningf("Found multiple availability zones, using %s\n", az)
	}

	return &asgTemplate{
		InstanceType: InstanceTypes[instanceTypeName],
		Region:       region,
		Zone:         az,
		Tags:         asg.Tags,
	}, nil
}

func (m *AwsManager) buildNodeFromTemplate(asg *Asg, template *asgTemplate) (*apiv1.Node, error) {
	node := apiv1.Node{}
	nodeName := fmt.Sprintf("%s-asg-%d", asg.Name, rand.Int63())

	node.ObjectMeta = metav1.ObjectMeta{
		Name:     nodeName,
		SelfLink: fmt.Sprintf("/api/v1/nodes/%s", nodeName),
		Labels:   map[string]string{},
	}

	node.Status = apiv1.NodeStatus{
		Capacity: apiv1.ResourceList{},
	}

	// TODO: get a real value.
	node.Status.Capacity[apiv1.ResourcePods] = *resource.NewQuantity(110, resource.DecimalSI)
	node.Status.Capacity[apiv1.ResourceCPU] = *resource.NewQuantity(template.InstanceType.VCPU, resource.DecimalSI)
	node.Status.Capacity[apiv1.ResourceMemory] = *resource.NewQuantity(template.InstanceType.MemoryMb*1024*1024, resource.DecimalSI)

	// TODO: use proper allocatable!!
	node.Status.Allocatable = node.Status.Capacity

	// NodeLabels
	node.Labels = cloudprovider.JoinStringMaps(node.Labels, extractLabelsFromAsg(template.Tags))
	// GenericLabels
	node.Labels = cloudprovider.JoinStringMaps(node.Labels, buildGenericLabels(template, nodeName))

	node.Spec.Taints = extractTaintsFromAsg(template.Tags)

	node.Status.Conditions = cloudprovider.BuildReadyConditions()
	return &node, nil
}

func buildGenericLabels(template *asgTemplate, nodeName string) map[string]string {
	result := make(map[string]string)
	// TODO: extract it somehow
	result[kubeletapis.LabelArch] = cloudprovider.DefaultArch
	result[kubeletapis.LabelOS] = cloudprovider.DefaultOS

	result[kubeletapis.LabelInstanceType] = template.InstanceType.InstanceType

	result[kubeletapis.LabelZoneRegion] = template.Region
	result[kubeletapis.LabelZoneFailureDomain] = template.Zone
	result[kubeletapis.LabelHostname] = nodeName
	return result
}

func extractLabelsFromAsg(tags []*autoscaling.TagDescription) map[string]string {
	result := make(map[string]string)

	for _, tag := range tags {
		k := *tag.Key
		v := *tag.Value
		splits := strings.Split(k, "k8s.io/cluster-autoscaler/node-template/label/")
		if len(splits) > 1 {
			label := splits[1]
			if label != "" {
				result[label] = v
			}
		}
	}

	return result
}

func extractTaintsFromAsg(tags []*autoscaling.TagDescription) []apiv1.Taint {
	taints := make([]apiv1.Taint, 0)

	for _, tag := range tags {
		k := *tag.Key
		v := *tag.Value
		splits := strings.Split(k, "k8s.io/cluster-autoscaler/node-template/taint/")
		if len(splits) > 1 {
			values := strings.SplitN(v, ":", 2)
			taints = append(taints, apiv1.Taint{
				Key:    splits[1],
				Value:  values[0],
				Effect: apiv1.TaintEffect(values[1]),
			})
		}
	}
	return taints
}
