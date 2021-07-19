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
	"github.com/aws/aws-sdk-go/service/ec2"
)

// autoScaling is the interface represents a specific aspect of the auto-scaling service provided by AWS SDK for use in CA
type autoScalingI interface {
	DescribeAutoScalingGroupsPages(input *autoscaling.DescribeAutoScalingGroupsInput, fn func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool) error
	DescribeLaunchConfigurations(*autoscaling.DescribeLaunchConfigurationsInput) (*autoscaling.DescribeLaunchConfigurationsOutput, error)
	DescribeTagsPages(input *autoscaling.DescribeTagsInput, fn func(*autoscaling.DescribeTagsOutput, bool) bool) error
	SetDesiredCapacity(input *autoscaling.SetDesiredCapacityInput) (*autoscaling.SetDesiredCapacityOutput, error)
	TerminateInstanceInAutoScalingGroup(input *autoscaling.TerminateInstanceInAutoScalingGroupInput) (*autoscaling.TerminateInstanceInAutoScalingGroupOutput, error)
}

type ec2I interface {
	DescribeLaunchTemplateVersions(input *ec2.DescribeLaunchTemplateVersionsInput) (*ec2.DescribeLaunchTemplateVersionsOutput, error)
}

// awsWrapper provides several utility methods over the services provided by the AWS SDK
type awsWrapper struct {
	autoScalingI
	ec2I
}

func (m *awsWrapper) getInstanceTypeByLaunchConfigNames(launchConfigToQuery []*string) (map[string]string, error) {
	launchConfigurationsToInstanceType := map[string]string{}

	for i := 0; i < len(launchConfigToQuery); i += 50 {
		end := i + 50

		if end > len(launchConfigToQuery) {
			end = len(launchConfigToQuery)
		}
		params := &autoscaling.DescribeLaunchConfigurationsInput{
			LaunchConfigurationNames: launchConfigToQuery[i:end],
			MaxRecords:               aws.Int64(50),
		}
		r, err := m.DescribeLaunchConfigurations(params)
		if err != nil {
			return nil, err
		}
		for _, lc := range r.LaunchConfigurations {
			launchConfigurationsToInstanceType[*lc.LaunchConfigurationName] = *lc.InstanceType
		}
	}
	return launchConfigurationsToInstanceType, nil
}

func (m *awsWrapper) getAutoscalingGroupsByNames(names []string) ([]*autoscaling.Group, error) {
	if len(names) == 0 {
		return nil, nil
	}

	asgs := make([]*autoscaling.Group, 0)

	// AWS only accepts up to 50 ASG names as input, describe them in batches
	for i := 0; i < len(names); i += maxAsgNamesPerDescribe {
		end := i + maxAsgNamesPerDescribe

		if end > len(names) {
			end = len(names)
		}

		input := &autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: aws.StringSlice(names[i:end]),
			MaxRecords:            aws.Int64(maxRecordsReturnedByAPI),
		}
		if err := m.DescribeAutoScalingGroupsPages(input, func(output *autoscaling.DescribeAutoScalingGroupsOutput, _ bool) bool {
			asgs = append(asgs, output.AutoScalingGroups...)
			// We return true while we want to be called with the next page of
			// results, if any.
			return true
		}); err != nil {
			return nil, err
		}
	}

	return asgs, nil
}

func (m *awsWrapper) getAutoscalingGroupNamesByTags(kvs map[string]string) ([]string, error) {
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

	return asgNames, nil
}

func (m *awsWrapper) getInstanceTypeByLaunchTemplate(launchTemplate *launchTemplate) (string, error) {
	params := &ec2.DescribeLaunchTemplateVersionsInput{
		LaunchTemplateName: aws.String(launchTemplate.name),
		Versions:           []*string{aws.String(launchTemplate.version)},
	}

	describeData, err := m.DescribeLaunchTemplateVersions(params)
	if err != nil {
		return "", err
	}

	if len(describeData.LaunchTemplateVersions) == 0 {
		return "", fmt.Errorf("unable to find template versions")
	}

	lt := describeData.LaunchTemplateVersions[0]
	instanceType := lt.LaunchTemplateData.InstanceType

	if instanceType == nil {
		return "", fmt.Errorf("unable to find instance type within launch template")
	}

	return aws.StringValue(instanceType), nil
}

func buildLaunchTemplateFromSpec(ltSpec *autoscaling.LaunchTemplateSpecification) *launchTemplate {
	// NOTE(jaypipes): The LaunchTemplateSpecification.Version is a pointer to
	// string. When the pointer is nil, EC2 AutoScaling API considers the value
	// to be "$Default", however aws.StringValue(ltSpec.Version) will return an
	// empty string (which is not considered the same as "$Default" or a nil
	// string pointer. So, in order to not pass an empty string as the version
	// for the launch template when we communicate with the EC2 AutoScaling API
	// using the information in the launchTemplate, we store the string
	// "$Default" here when the ltSpec.Version is a nil pointer.
	//
	// See:
	//
	// https://github.com/kubernetes/autoscaler/issues/1728
	// https://github.com/aws/aws-sdk-go/blob/81fad3b797f4a9bd1b452a5733dd465eefef1060/service/autoscaling/api.go#L10666-L10671
	//
	// A cleaner alternative might be to make launchTemplate.version a string
	// pointer instead of a string, or even store the aws-sdk-go's
	// LaunchTemplateSpecification structs directly.
	var version string
	if ltSpec.Version == nil {
		version = "$Default"
	} else {
		version = aws.StringValue(ltSpec.Version)
	}
	return &launchTemplate{
		name:    aws.StringValue(ltSpec.LaunchTemplateName),
		version: version,
	}
}
