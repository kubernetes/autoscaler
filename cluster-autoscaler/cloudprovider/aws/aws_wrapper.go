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
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/eks"
	apiv1 "k8s.io/api/core/v1"
	klog "k8s.io/klog/v2"
)

// autoScalingI is the interface abstracting specific API calls of the auto-scaling service provided by AWS SDK for use in CA
type autoScalingI interface {
	DescribeAutoScalingGroupsPages(input *autoscaling.DescribeAutoScalingGroupsInput, fn func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool) error
	DescribeLaunchConfigurations(*autoscaling.DescribeLaunchConfigurationsInput) (*autoscaling.DescribeLaunchConfigurationsOutput, error)
	DescribeScalingActivities(*autoscaling.DescribeScalingActivitiesInput) (*autoscaling.DescribeScalingActivitiesOutput, error)
	DescribeTagsPages(input *autoscaling.DescribeTagsInput, fn func(*autoscaling.DescribeTagsOutput, bool) bool) error
	SetDesiredCapacity(input *autoscaling.SetDesiredCapacityInput) (*autoscaling.SetDesiredCapacityOutput, error)
	TerminateInstanceInAutoScalingGroup(input *autoscaling.TerminateInstanceInAutoScalingGroupInput) (*autoscaling.TerminateInstanceInAutoScalingGroupOutput, error)
}

// ec2I is the interface abstracting specific API calls of the EC2 service provided by AWS SDK for use in CA
type ec2I interface {
	DescribeLaunchTemplateVersions(input *ec2.DescribeLaunchTemplateVersionsInput) (*ec2.DescribeLaunchTemplateVersionsOutput, error)
}

// eksI is the interface that represents a specific aspect of EKS (Elastic Kubernetes Service) which is provided by AWS SDK for use in CA
type eksI interface {
	DescribeNodegroup(input *eks.DescribeNodegroupInput) (*eks.DescribeNodegroupOutput, error)
}

// awsWrapper provides several utility methods over the services provided by the AWS SDK
type awsWrapper struct {
	autoScalingI
	ec2I
	eksI
}

func (m *awsWrapper) getManagedNodegroupInfo(nodegroupName string, clusterName string) ([]apiv1.Taint, map[string]string, error) {
	params := &eks.DescribeNodegroupInput{
		ClusterName:   &clusterName,
		NodegroupName: &nodegroupName,
	}
	start := time.Now()
	r, err := m.DescribeNodegroup(params)
	observeAWSRequest("DescribeNodegroup", err, start)
	if err != nil {
		return nil, nil, err
	}

	taints := make([]apiv1.Taint, 0)
	labels := make(map[string]string)

	// Labels will include diskSize, amiType, capacityType, version
	if r.Nodegroup.DiskSize != nil {
		labels["diskSize"] = strconv.FormatInt(*r.Nodegroup.DiskSize, 10)
	}

	if r.Nodegroup.AmiType != nil && len(*r.Nodegroup.AmiType) > 0 {
		labels["amiType"] = *r.Nodegroup.AmiType
	}

	if r.Nodegroup.CapacityType != nil && len(*r.Nodegroup.CapacityType) > 0 {
		labels["capacityType"] = *r.Nodegroup.CapacityType
	}

	if r.Nodegroup.Version != nil && len(*r.Nodegroup.Version) > 0 {
		labels["k8sVersion"] = *r.Nodegroup.Version
	}

	if r.Nodegroup.Labels != nil && len(r.Nodegroup.Labels) > 0 {
		labelsMap := r.Nodegroup.Labels
		for k, v := range labelsMap {
			if v != nil {
				labels[k] = *v
			}
		}
	}

	if r.Nodegroup.Taints != nil && len(r.Nodegroup.Taints) > 0 {
		taintList := r.Nodegroup.Taints
		for _, taint := range taintList {
			if taint != nil && taint.Effect != nil && taint.Key != nil && taint.Value != nil {
				taints = append(taints, apiv1.Taint{
					Key:    *taint.Key,
					Value:  *taint.Value,
					Effect: apiv1.TaintEffect(*taint.Effect),
				})
			}
		}
	}

	return taints, labels, nil
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
		start := time.Now()
		r, err := m.DescribeLaunchConfigurations(params)
		observeAWSRequest("DescribeLaunchConfigurations", err, start)
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
		start := time.Now()
		err := m.DescribeAutoScalingGroupsPages(input, func(output *autoscaling.DescribeAutoScalingGroupsOutput, _ bool) bool {
			asgs = append(asgs, output.AutoScalingGroups...)
			// We return true while we want to be called with the next page of
			// results, if any.
			return true
		})
		observeAWSRequest("DescribeAutoScalingGroupsPages", err, start)
		if err != nil {
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
	start := time.Now()
	err := m.DescribeTagsPages(input, func(out *autoscaling.DescribeTagsOutput, _ bool) bool {
		tags = append(tags, out.Tags...)
		// We return true while we want to be called with the next page of
		// results, if any.
		return true
	})
	observeAWSRequest("DescribeTagsPages", err, start)

	if err != nil {
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

	start := time.Now()
	describeData, err := m.DescribeLaunchTemplateVersions(params)
	observeAWSRequest("DescribeLaunchTemplateVersions", err, start)
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

func (m *awsWrapper) getInstanceTypesForAsgs(asgs []*asg) (map[string]string, error) {
	results := map[string]string{}
	launchConfigsToQuery := map[string]string{}
	launchTemplatesToQuery := map[string]*launchTemplate{}

	for _, asg := range asgs {
		name := asg.AwsRef.Name
		if asg.LaunchConfigurationName != "" {
			launchConfigsToQuery[name] = asg.LaunchConfigurationName
		} else if asg.LaunchTemplate != nil {
			launchTemplatesToQuery[name] = asg.LaunchTemplate
		} else if asg.MixedInstancesPolicy != nil {
			if len(asg.MixedInstancesPolicy.instanceTypesOverrides) > 0 {
				results[name] = asg.MixedInstancesPolicy.instanceTypesOverrides[0]
			} else {
				launchTemplatesToQuery[name] = asg.MixedInstancesPolicy.launchTemplate
			}
		}
	}

	klog.V(4).Infof("%d launch configurations to query", len(launchConfigsToQuery))
	klog.V(4).Infof("%d launch templates to query", len(launchTemplatesToQuery))

	// Query these all at once to minimize AWS API calls
	launchConfigNames := make([]*string, 0, len(launchConfigsToQuery))
	for _, cfgName := range launchConfigsToQuery {
		launchConfigNames = append(launchConfigNames, aws.String(cfgName))
	}
	launchConfigs, err := m.getInstanceTypeByLaunchConfigNames(launchConfigNames)
	if err != nil {
		klog.Errorf("Failed to query %d launch configurations", len(launchConfigsToQuery))
		return nil, err
	}

	for asgName, cfgName := range launchConfigsToQuery {
		results[asgName] = launchConfigs[cfgName]
	}
	klog.V(4).Infof("Successfully queried %d launch configurations", len(launchConfigsToQuery))

	// Have to query LaunchTemplates one-at-a-time, since there's no way to query <lt, version> pairs in bulk
	for asgName, lt := range launchTemplatesToQuery {
		instanceType, err := m.getInstanceTypeByLaunchTemplate(lt)
		if err != nil {
			klog.Errorf("Failed to query launch tempate %s: %v", lt.name, err)
			continue
		}
		results[asgName] = instanceType
	}
	klog.V(4).Infof("Successfully queried %d launch templates", len(launchTemplatesToQuery))

	return results, nil
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
