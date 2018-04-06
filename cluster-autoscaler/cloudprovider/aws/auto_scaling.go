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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/golang/glog"
)

// autoScaling is the interface represents a specific aspect of the auto-scaling service provided by AWS SDK for use in CA
type autoScaling interface {
	DescribeAutoScalingGroups(input *autoscaling.DescribeAutoScalingGroupsInput) (*autoscaling.DescribeAutoScalingGroupsOutput, error)
	DescribeAutoScalingGroupsPages(input *autoscaling.DescribeAutoScalingGroupsInput, fn func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool) error
	DescribeLaunchConfigurations(*autoscaling.DescribeLaunchConfigurationsInput) (*autoscaling.DescribeLaunchConfigurationsOutput, error)
	DescribeTagsPages(input *autoscaling.DescribeTagsInput, fn func(*autoscaling.DescribeTagsOutput, bool) bool) error
	SetDesiredCapacity(input *autoscaling.SetDesiredCapacityInput) (*autoscaling.SetDesiredCapacityOutput, error)
	TerminateInstanceInAutoScalingGroup(input *autoscaling.TerminateInstanceInAutoScalingGroupInput) (*autoscaling.TerminateInstanceInAutoScalingGroupOutput, error)
}

// autoScalingWrapper provides several utility methods over the auto-scaling service provided by AWS SDK
type autoScalingWrapper struct {
	autoScaling
}

func (m autoScalingWrapper) getInstanceTypeByLCName(name string) (string, error) {
	params := &autoscaling.DescribeLaunchConfigurationsInput{
		LaunchConfigurationNames: []*string{aws.String(name)},
		MaxRecords:               aws.Int64(1),
	}
	launchConfigurations, err := m.DescribeLaunchConfigurations(params)
	if err != nil {
		glog.V(4).Infof("Failed LaunchConfiguration info request for %s: %v", name, err)
		return "", err
	}
	if len(launchConfigurations.LaunchConfigurations) < 1 {
		return "", fmt.Errorf("Unable to get first LaunchConfiguration for %s", name)
	}

	return *launchConfigurations.LaunchConfigurations[0].InstanceType, nil
}

func (m autoScalingWrapper) getAutoscalingGroupByName(name string) (*autoscaling.Group, error) {
	params := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{aws.String(name)},
		MaxRecords:            aws.Int64(1),
	}
	groups, err := m.DescribeAutoScalingGroups(params)
	if err != nil {
		glog.V(4).Infof("Failed ASG info request for %s: %v", name, err)
		return nil, err
	}
	if len(groups.AutoScalingGroups) < 1 {
		return nil, fmt.Errorf("Unable to get first autoscaling.Group for %s", name)
	}
	return groups.AutoScalingGroups[0], nil
}

func (m *autoScalingWrapper) getAutoscalingGroupsByNames(names []string) ([]*autoscaling.Group, error) {
	if len(names) == 0 {
		return nil, nil
	}
	input := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: aws.StringSlice(names),
		MaxRecords:            aws.Int64(maxRecordsReturnedByAPI),
	}
	asgs := make([]*autoscaling.Group, 0)
	if err := m.DescribeAutoScalingGroupsPages(input, func(output *autoscaling.DescribeAutoScalingGroupsOutput, _ bool) bool {
		asgs = append(asgs, output.AutoScalingGroups...)
		// We return true while we want to be called with the next page of
		// results, if any.
		return true
	}); err != nil {
		return nil, err
	}
	return asgs, nil
}

func (m *autoScalingWrapper) getAutoscalingGroupsByTags(kvs map[string]string) ([]*autoscaling.Group, error) {
	// DescribeTags does an OR query when multiple filters on different tags are
	// specified. In other words, DescribeTags returns [asg1, asg1] for keys
	// [t1, t2] when there's only one asg tagged both t1 and t2.
	filters := []*autoscaling.Filter{}
	for key, value := range kvs {
		filter := &autoscaling.Filter{
			Name:   aws.String("key"),
			Values: []*string{aws.String(key)},
		}
		filters = append(filters, filter)
		if value != "" {
			filters = append(filters, &autoscaling.Filter{
				Name:   aws.String("value"),
				Values: []*string{aws.String(value)},
			})
		}
	}

	tags := []*autoscaling.TagDescription{}
	input := &autoscaling.DescribeTagsInput{
		Filters:    filters,
		MaxRecords: aws.Int64(maxRecordsReturnedByAPI),
	}
	if err := m.DescribeTagsPages(input, func(out *autoscaling.DescribeTagsOutput, _ bool) bool {
		tags = append(tags, out.Tags...)
		// We return true while we want to be called with the next page of
		// results, if any.
		return true
	}); err != nil {
		return nil, err
	}

	// According to how DescribeTags API works, the result contains ASGs which
	// not all but only subset of tags are associated. Explicitly select ASGs to
	// which all the tags are associated so that we won't end up calling
	// DescribeAutoScalingGroups API multiple times on an ASG.
	asgNames := []string{}
	asgNameOccurrences := make(map[string]int)
	for _, t := range tags {
		asgName := aws.StringValue(t.ResourceId)
		occurrences := asgNameOccurrences[asgName] + 1
		if occurrences >= len(kvs) {
			asgNames = append(asgNames, asgName)
		}
		asgNameOccurrences[asgName] = occurrences
	}

	return m.getAutoscalingGroupsByNames(asgNames)
}
