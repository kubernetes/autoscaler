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
	"sync"
	"time"

	"gopkg.in/gcfg.v1"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/golang/glog"
	provider_aws "k8s.io/kubernetes/pkg/cloudprovider/providers/aws"
	"k8s.io/kubernetes/pkg/util/wait"
)

const (
	operationWaitTimeout  = 5 * time.Second
	operationPollInterval = 100 * time.Millisecond
)

type asgInformation struct {
	config   *Asg
	basename string
}

type autoScaling interface {
	DescribeAutoScalingGroups(input *autoscaling.DescribeAutoScalingGroupsInput) (*autoscaling.DescribeAutoScalingGroupsOutput, error)
	SetDesiredCapacity(input *autoscaling.SetDesiredCapacityInput) (*autoscaling.SetDesiredCapacityOutput, error)
	TerminateInstanceInAutoScalingGroup(input *autoscaling.TerminateInstanceInAutoScalingGroupInput) (*autoscaling.TerminateInstanceInAutoScalingGroupOutput, error)
}

// AwsManager is handles aws communication and data caching.
type AwsManager struct {
	asgs     []*asgInformation
	asgCache map[AwsRef]*Asg

	service    autoScaling
	cacheMutex sync.Mutex
}

// CreateAwsManager constructs awsManager object.
func CreateAwsManager(configReader io.Reader) (*AwsManager, error) {
	if configReader != nil {
		var cfg provider_aws.AWSCloudConfig
		if err := gcfg.ReadInto(&cfg, configReader); err != nil {
			glog.Errorf("Couldn't read config: %v", err)
			return nil, err
		}
	}

	service := autoscaling.New(session.New())
	manager := &AwsManager{
		asgs:     make([]*asgInformation, 0),
		service:  service,
		asgCache: make(map[AwsRef]*Asg),
	}

	go wait.Forever(func() {
		manager.cacheMutex.Lock()
		defer manager.cacheMutex.Unlock()
		if err := manager.regenerateCache(); err != nil {
			glog.Errorf("Error while regenerating Asg cache: %v", err)
		}
	}, time.Hour)

	return manager, nil
}

// RegisterAsg registers asg in Aws Manager.
func (m *AwsManager) RegisterAsg(asg *Asg) {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()

	m.asgs = append(m.asgs, &asgInformation{
		config: asg,
	})
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
	commonAsg, err := m.GetAsgForInstance(instances[0])
	if err != nil {
		return err
	}
	for _, instance := range instances {
		asg, err := m.GetAsgForInstance(instance)
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

// GetAsgForInstance returns AsgConfig of the given Instance
func (m *AwsManager) GetAsgForInstance(instance *AwsRef) (*Asg, error) {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()
	if config, found := m.asgCache[*instance]; found {
		return config, nil
	}
	if err := m.regenerateCache(); err != nil {
		return nil, fmt.Errorf("Error while looking for ASG for instance %+v, error: %v", *instance, err)
	}
	if config, found := m.asgCache[*instance]; found {
		return config, nil
	}
	// instance does not belong to any configured ASG
	return nil, nil
}

func (m *AwsManager) regenerateCache() error {
	newCache := make(map[AwsRef]*Asg)

	for _, asg := range m.asgs {
		glog.V(4).Infof("Regenerating ASG information for %s", asg.config.Name)

		params := &autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: []*string{aws.String(asg.config.Name)},
			MaxRecords:            aws.Int64(1),
		}
		groups, err := m.service.DescribeAutoScalingGroups(params)
		if err != nil {
			glog.V(4).Infof("Failed ASG info request for %s: %v", asg.config.Name, err)
			return err
		}
		if len(groups.AutoScalingGroups) < 1 {
			return fmt.Errorf("Unable to get first autoscaling.Group for %s", asg.config.Name)
		}
		group := *groups.AutoScalingGroups[0]

		for _, instance := range group.Instances {
			ref := AwsRef{Name: *instance.InstanceId}
			newCache[ref] = asg.config
		}
	}

	m.asgCache = newCache
	return nil
}
