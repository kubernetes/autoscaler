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

package aws

import (
	"fmt"
	"io"
	"time"

	"gopkg.in/gcfg.v1"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/golang/glog"
	provider_aws "k8s.io/kubernetes/pkg/cloudprovider/providers/aws"
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
	service autoScalingWrapper
	asgs    *autoScalingGroups
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
		asgs:    newAutoScalingGroups(service),
		service: service,
	}

	return manager, nil
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
