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
	"context"
	"fmt"
	"maps"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	autoscalingtypes "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	ekstypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
	apiv1 "k8s.io/api/core/v1"
	klog "k8s.io/klog/v2"
)

// autoScalingI is the interface abstracting specific API calls of the auto-scaling service provided by AWS SDK for use in CA
type autoScalingI interface {
	autoscaling.DescribeAutoScalingGroupsAPIClient
	autoscaling.DescribeLaunchConfigurationsAPIClient
	autoscaling.DescribeScalingActivitiesAPIClient
	//DescribeAutoScalingGroupsPages(input *autoscaling.DescribeAutoScalingGroupsInput, fn func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool) error
	//DescribeLaunchConfigurations(*autoscaling.DescribeLaunchConfigurationsInput) (*autoscaling.DescribeLaunchConfigurationsOutput, error)
	//DescribeScalingActivities(*autoscaling.DescribeScalingActivitiesInput) (*autoscaling.DescribeScalingActivitiesOutput, error)
	SetDesiredCapacity(ctx context.Context, input *autoscaling.SetDesiredCapacityInput, opts ...func(options *autoscaling.Options)) (*autoscaling.SetDesiredCapacityOutput, error)
	TerminateInstanceInAutoScalingGroup(ctx context.Context, input *autoscaling.TerminateInstanceInAutoScalingGroupInput, opts ...func(options *autoscaling.Options)) (*autoscaling.TerminateInstanceInAutoScalingGroupOutput, error)
}

// ec2I is the interface abstracting specific API calls of the EC2 service provided by AWS SDK for use in CA
type ec2I interface {
	ec2.DescribeImagesAPIClient
	ec2.DescribeLaunchTemplateVersionsAPIClient
	ec2.GetInstanceTypesFromInstanceRequirementsAPIClient
	//DescribeImages(input *ec2.DescribeImagesInput) (*ec2.DescribeImagesOutput, error)
	//DescribeLaunchTemplateVersions(input *ec2.DescribeLaunchTemplateVersionsInput) (*ec2.DescribeLaunchTemplateVersionsOutput, error)
	//GetInstanceTypesFromInstanceRequirementsPages(input *ec2.GetInstanceTypesFromInstanceRequirementsInput, fn func(*ec2.GetInstanceTypesFromInstanceRequirementsOutput, bool) bool) error
}

// eksI is the interface that represents a specific aspect of EKS (Elastic Kubernetes Service) which is provided by AWS SDK for use in CA
type eksI interface {
	eks.DescribeNodegroupAPIClient
	//DescribeNodegroup(input *eks.DescribeNodegroupInput) (*eks.DescribeNodegroupOutput, error)
}

// awsWrapper provides several utility methods over the services provided by the AWS SDK
type awsWrapper struct {
	autoScalingI
	ec2I
	eksI
}

func (m *awsWrapper) getManagedNodegroupInfo(nodegroupName string, clusterName string) ([]apiv1.Taint, map[string]string, map[string]string, error) {
	params := &eks.DescribeNodegroupInput{
		ClusterName:   &clusterName,
		NodegroupName: &nodegroupName,
	}
	start := time.Now()
	r, err := m.DescribeNodegroup(context.Background(), params)
	observeAWSRequest("DescribeNodegroup", err, start)
	if err != nil {
		return nil, nil, nil, err
	}

	klog.V(6).Infof("DescribeNodegroup output : %+v\n", r)

	taints := make([]apiv1.Taint, 0)
	labels := make(map[string]string)
	tags := make(map[string]string)

	// Labels will include diskSize, amiType, capacityType, version
	if r.Nodegroup.DiskSize != nil {
		labels["diskSize"] = strconv.FormatInt(int64(*r.Nodegroup.DiskSize), 10)
	}

	if r.Nodegroup.AmiType != "" {
		labels["amiType"] = string(r.Nodegroup.AmiType)
	}

	if r.Nodegroup.CapacityType != "" {
		labels["eks.amazonaws.com/capacityType"] = string(r.Nodegroup.CapacityType)
	}

	if r.Nodegroup.Version != nil && len(*r.Nodegroup.Version) > 0 {
		labels["k8sVersion"] = *r.Nodegroup.Version
	}

	if r.Nodegroup.NodegroupName != nil && len(*r.Nodegroup.NodegroupName) > 0 {
		labels["eks.amazonaws.com/nodegroup"] = *r.Nodegroup.NodegroupName
	}

	if r.Nodegroup.Labels != nil && len(r.Nodegroup.Labels) > 0 {
		labelsMap := r.Nodegroup.Labels
		maps.Copy(labels, labelsMap)
	}

	if r.Nodegroup.Tags != nil && len(r.Nodegroup.Tags) > 0 {
		tagsMap := r.Nodegroup.Tags
		maps.Copy(tags, tagsMap)
	}

	if r.Nodegroup.Taints != nil && len(r.Nodegroup.Taints) > 0 {
		taintList := r.Nodegroup.Taints
		for _, taint := range taintList {
			if taint.Key != nil && taint.Value != nil {
				formattedEffect, err := taintEksTranslator(taint)
				if err != nil {
					return nil, nil, nil, err
				}
				taints = append(taints, apiv1.Taint{
					Key:    *taint.Key,
					Value:  *taint.Value,
					Effect: apiv1.TaintEffect(formattedEffect),
				})
			}
		}
	}

	return taints, labels, tags, nil
}

func (m *awsWrapper) getInstanceTypeByLaunchConfigNames(launchConfigToQuery []string) (map[string]string, error) {
	launchConfigurationsToInstanceType := map[string]string{}

	for i := 0; i < len(launchConfigToQuery); i += 50 {
		end := min(i+50, len(launchConfigToQuery))
		params := &autoscaling.DescribeLaunchConfigurationsInput{
			LaunchConfigurationNames: launchConfigToQuery[i:end],
			MaxRecords:               aws.Int32(50),
		}
		start := time.Now()
		r, err := m.DescribeLaunchConfigurations(context.Background(), params)
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

func (m *awsWrapper) getAutoscalingGroupsByNames(names []string) ([]autoscalingtypes.AutoScalingGroup, error) {
	asgs := make([]autoscalingtypes.AutoScalingGroup, 0)
	if len(names) == 0 {
		return asgs, nil
	}

	// AWS only accepts up to 100 ASG names as input, describe them in batches
	for i := 0; i < len(names); i += maxAsgNamesPerDescribe {
		end := min(i+maxAsgNamesPerDescribe, len(names))

		input := &autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: names[i:end],
			MaxRecords:            aws.Int32(maxRecordsReturnedByAPI),
		}
		start := time.Now()
		var err error
		paginator := autoscaling.NewDescribeAutoScalingGroupsPaginator(m, input)
		for paginator.HasMorePages() {
			var page *autoscaling.DescribeAutoScalingGroupsOutput
			page, err = paginator.NextPage(context.Background())
			if err != nil {
				break
			}
			asgs = append(asgs, page.AutoScalingGroups...)
		}
		observeAWSRequest("DescribeAutoScalingGroupsPages", err, start)
		if err != nil {
			return nil, err
		}
	}

	return asgs, nil
}

func (m *awsWrapper) getAutoscalingGroupsByTags(tags map[string]string) ([]autoscalingtypes.AutoScalingGroup, error) {
	asgs := make([]autoscalingtypes.AutoScalingGroup, 0)
	if len(tags) == 0 {
		return asgs, nil
	}

	filters := make([]autoscalingtypes.Filter, 0)
	for key, value := range tags {
		if value != "" {
			filters = append(filters, autoscalingtypes.Filter{
				Name:   aws.String(fmt.Sprintf("tag:%s", key)),
				Values: []string{value},
			})
		} else {
			filters = append(filters, autoscalingtypes.Filter{
				Name:   aws.String("tag-key"),
				Values: []string{key},
			})
		}
	}

	input := &autoscaling.DescribeAutoScalingGroupsInput{
		Filters:    filters,
		MaxRecords: aws.Int32(maxRecordsReturnedByAPI),
	}
	start := time.Now()
	var err error
	paginator := autoscaling.NewDescribeAutoScalingGroupsPaginator(m, input)
	for paginator.HasMorePages() {
		var page *autoscaling.DescribeAutoScalingGroupsOutput
		page, err = paginator.NextPage(context.Background())
		if err != nil {
			break
		}
		asgs = append(asgs, page.AutoScalingGroups...)
	}
	observeAWSRequest("DescribeAutoScalingGroupsPages", err, start)
	if err != nil {
		return nil, err
	}

	return asgs, nil
}

func (m *awsWrapper) getInstanceTypeByLaunchTemplate(launchTemplate *launchTemplate) (string, error) {
	templateData, err := m.getLaunchTemplateData(launchTemplate.name, launchTemplate.version)
	if err != nil {
		return "", err
	}

	instanceType := ""
	if templateData.InstanceType != "" {
		instanceType = string(templateData.InstanceType)
	} else if templateData.InstanceRequirements != nil && templateData.ImageId != nil {
		requirementsRequest, err := m.getRequirementsRequestFromEC2(templateData.InstanceRequirements)
		if err != nil {
			return "", fmt.Errorf("unable to get instance requirements request")
		}
		instanceType, err = m.getInstanceTypeFromInstanceRequirements(*templateData.ImageId, requirementsRequest)
		if err != nil {
			return "", err
		}
	}
	if len(instanceType) == 0 {
		return "", fmt.Errorf("unable to find instance type using launch template")
	}

	return instanceType, nil
}

func (m *awsWrapper) getInstanceTypeFromRequirementsOverrides(policy *mixedInstancesPolicy) (string, error) {
	if policy.launchTemplate == nil {
		return "", fmt.Errorf("no launch template found for mixed instances policy")
	}

	templateData, err := m.getLaunchTemplateData(policy.launchTemplate.name, policy.launchTemplate.version)
	if err != nil {
		return "", err
	}

	requirements, err := m.getRequirementsRequestFromAutoscaling(policy.instanceRequirementsOverrides)
	if err != nil {
		return "", err
	}
	instanceType, err := m.getInstanceTypeFromInstanceRequirements(*templateData.ImageId, requirements)
	if err != nil {
		return "", err
	}

	return string(instanceType), nil
}

func (m *awsWrapper) getLaunchTemplateData(templateName string, templateVersion string) (*ec2types.ResponseLaunchTemplateData, error) {
	describeTemplateInput := &ec2.DescribeLaunchTemplateVersionsInput{
		LaunchTemplateName: aws.String(templateName),
		Versions:           []string{templateVersion},
	}

	start := time.Now()
	describeData, err := m.DescribeLaunchTemplateVersions(context.Background(), describeTemplateInput)
	observeAWSRequest("DescribeLaunchTemplateVersions", err, start)
	if err != nil {
		return nil, err
	}
	if describeData == nil || len(describeData.LaunchTemplateVersions) == 0 {
		return nil, fmt.Errorf("unable to find template versions for launch template %s", templateName)
	}
	if describeData.LaunchTemplateVersions[0].LaunchTemplateData == nil {
		return nil, fmt.Errorf("no data found for launch template %s, version %s", templateName, templateVersion)
	}

	return describeData.LaunchTemplateVersions[0].LaunchTemplateData, nil
}

func (m *awsWrapper) getInstanceTypeFromInstanceRequirements(imageId string, requirementsRequest *ec2types.InstanceRequirementsRequest) (string, error) {
	describeImagesInput := &ec2.DescribeImagesInput{
		ImageIds: []string{imageId},
	}

	start := time.Now()
	describeImagesOutput, err := m.DescribeImages(context.Background(), describeImagesInput)
	observeAWSRequest("DescribeImages", err, start)
	if err != nil {
		return "", err
	}

	var imageArchitectures []ec2types.ArchitectureType
	var imageVirtualizationTypes []ec2types.VirtualizationType
	for _, image := range describeImagesOutput.Images {
		imageArchitectures = append(imageArchitectures, ec2types.ArchitectureType(image.Architecture))
		imageVirtualizationTypes = append(imageVirtualizationTypes, image.VirtualizationType)
	}

	requirementsInput := &ec2.GetInstanceTypesFromInstanceRequirementsInput{
		ArchitectureTypes:    imageArchitectures,
		InstanceRequirements: requirementsRequest,
		VirtualizationTypes:  imageVirtualizationTypes,
	}

	start = time.Now()
	var instanceTypes []string
	paginator := ec2.NewGetInstanceTypesFromInstanceRequirementsPaginator(m, requirementsInput)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.Background())
		if err != nil {
			break
		}
		for _, instanceType := range page.InstanceTypes {
			instanceTypes = append(instanceTypes, *instanceType.InstanceType)
		}
	}
	observeAWSRequest("GetInstanceTypesFromInstanceRequirements", err, start)
	if err != nil {
		return "", fmt.Errorf("unable to get instance types from requirements: %w", err)
	}

	if len(instanceTypes) == 0 {
		return "", fmt.Errorf("no instance types found for requirements")
	}
	return instanceTypes[0], nil
}

func (m *awsWrapper) getRequirementsRequestFromAutoscaling(requirements *autoscalingtypes.InstanceRequirements) (*ec2types.InstanceRequirementsRequest, error) {
	requirementsRequest := ec2types.InstanceRequirementsRequest{}

	// required instance requirements
	requirementsRequest.MemoryMiB = &ec2types.MemoryMiBRequest{
		Min: requirements.MemoryMiB.Min,
		Max: requirements.MemoryMiB.Max,
	}

	requirementsRequest.VCpuCount = &ec2types.VCpuCountRangeRequest{
		Min: requirements.VCpuCount.Min,
		Max: requirements.VCpuCount.Max,
	}

	// optional instance requirements
	if requirements.AcceleratorCount != nil {
		requirementsRequest.AcceleratorCount = &ec2types.AcceleratorCountRequest{
			Min: requirements.AcceleratorCount.Min,
			Max: requirements.AcceleratorCount.Max,
		}
	}

	if requirements.AcceleratorManufacturers != nil {
		var acceleratorManufacturers []ec2types.AcceleratorManufacturer
		for _, manufacturer := range requirements.AcceleratorManufacturers {
			acceleratorManufacturers = append(acceleratorManufacturers, ec2types.AcceleratorManufacturer(manufacturer))
		}
		requirementsRequest.AcceleratorManufacturers = acceleratorManufacturers
	}

	if requirements.AcceleratorNames != nil {
		var acceleratorNames []ec2types.AcceleratorName
		for _, name := range requirements.AcceleratorNames {
			acceleratorNames = append(acceleratorNames, ec2types.AcceleratorName(name))
		}
		requirementsRequest.AcceleratorNames = acceleratorNames
	}

	if requirements.AcceleratorTotalMemoryMiB != nil {
		requirementsRequest.AcceleratorTotalMemoryMiB = &ec2types.AcceleratorTotalMemoryMiBRequest{
			Min: requirements.AcceleratorTotalMemoryMiB.Min,
			Max: requirements.AcceleratorTotalMemoryMiB.Max,
		}
	}

	if requirements.AcceleratorTypes != nil {
		var acceleratorTypes []ec2types.AcceleratorType
		for _, acceleratorType := range requirements.AcceleratorTypes {
			acceleratorTypes = append(acceleratorTypes, ec2types.AcceleratorType(acceleratorType))
		}
		requirementsRequest.AcceleratorTypes = acceleratorTypes
	}

	if requirements.BareMetal != "" {
		requirementsRequest.BareMetal = ec2types.BareMetal(requirements.BareMetal)
	}

	if requirements.BaselineEbsBandwidthMbps != nil {
		requirementsRequest.BaselineEbsBandwidthMbps = &ec2types.BaselineEbsBandwidthMbpsRequest{
			Min: requirements.BaselineEbsBandwidthMbps.Min,
			Max: requirements.BaselineEbsBandwidthMbps.Max,
		}
	}

	if requirements.BurstablePerformance != "" {
		requirementsRequest.BurstablePerformance = ec2types.BurstablePerformance(requirements.BurstablePerformance)
	}

	if requirements.CpuManufacturers != nil {
		var cpuManufacturers []ec2types.CpuManufacturer
		for _, manufacturer := range requirements.CpuManufacturers {
			cpuManufacturers = append(cpuManufacturers, ec2types.CpuManufacturer(manufacturer))
		}
		requirementsRequest.CpuManufacturers = cpuManufacturers
	}

	if requirements.ExcludedInstanceTypes != nil {
		requirementsRequest.ExcludedInstanceTypes = requirements.ExcludedInstanceTypes
	}

	if requirements.InstanceGenerations != nil {
		var instanceGenerations []ec2types.InstanceGeneration
		for _, generation := range requirements.InstanceGenerations {
			instanceGenerations = append(instanceGenerations, ec2types.InstanceGeneration(generation))
		}
		requirementsRequest.InstanceGenerations = instanceGenerations
	}

	if requirements.LocalStorage != "" {
		requirementsRequest.LocalStorage = ec2types.LocalStorage(requirements.LocalStorage)
	}

	if requirements.LocalStorageTypes != nil {
		var localStorageTypes []ec2types.LocalStorageType
		for _, localStorageType := range requirements.LocalStorageTypes {
			localStorageTypes = append(localStorageTypes, ec2types.LocalStorageType(localStorageType))
		}
		requirementsRequest.LocalStorageTypes = localStorageTypes
	}

	if requirements.MemoryGiBPerVCpu != nil {
		requirementsRequest.MemoryGiBPerVCpu = &ec2types.MemoryGiBPerVCpuRequest{
			Min: requirements.MemoryGiBPerVCpu.Min,
			Max: requirements.MemoryGiBPerVCpu.Max,
		}
	}

	if requirements.NetworkInterfaceCount != nil {
		requirementsRequest.NetworkInterfaceCount = &ec2types.NetworkInterfaceCountRequest{
			Min: requirements.NetworkInterfaceCount.Min,
			Max: requirements.NetworkInterfaceCount.Max,
		}
	}

	if requirements.OnDemandMaxPricePercentageOverLowestPrice != nil {
		requirementsRequest.OnDemandMaxPricePercentageOverLowestPrice = requirements.OnDemandMaxPricePercentageOverLowestPrice
	}

	if requirements.RequireHibernateSupport != nil {
		requirementsRequest.RequireHibernateSupport = requirements.RequireHibernateSupport
	}

	if requirements.SpotMaxPricePercentageOverLowestPrice != nil {
		requirementsRequest.SpotMaxPricePercentageOverLowestPrice = requirements.SpotMaxPricePercentageOverLowestPrice
	}

	if requirements.TotalLocalStorageGB != nil {
		requirementsRequest.TotalLocalStorageGB = &ec2types.TotalLocalStorageGBRequest{
			Min: requirements.TotalLocalStorageGB.Min,
			Max: requirements.TotalLocalStorageGB.Max,
		}
	}

	return &requirementsRequest, nil
}

func (m *awsWrapper) getRequirementsRequestFromEC2(requirements *ec2types.InstanceRequirements) (*ec2types.InstanceRequirementsRequest, error) {
	requirementsRequest := ec2types.InstanceRequirementsRequest{}

	// required instance requirements
	requirementsRequest.MemoryMiB = &ec2types.MemoryMiBRequest{
		Min: requirements.MemoryMiB.Min,
		Max: requirements.MemoryMiB.Max,
	}

	requirementsRequest.VCpuCount = &ec2types.VCpuCountRangeRequest{
		Min: requirements.VCpuCount.Min,
		Max: requirements.VCpuCount.Max,
	}

	// optional instance requirements
	if requirements.AcceleratorCount != nil {
		requirementsRequest.AcceleratorCount = &ec2types.AcceleratorCountRequest{
			Min: requirements.AcceleratorCount.Min,
			Max: requirements.AcceleratorCount.Max,
		}
	}

	if requirements.AcceleratorManufacturers != nil {
		requirementsRequest.AcceleratorManufacturers = requirements.AcceleratorManufacturers
	}

	if requirements.AcceleratorNames != nil {
		requirementsRequest.AcceleratorNames = requirements.AcceleratorNames
	}

	if requirements.AcceleratorTotalMemoryMiB != nil {
		requirementsRequest.AcceleratorTotalMemoryMiB = &ec2types.AcceleratorTotalMemoryMiBRequest{
			Min: requirements.AcceleratorTotalMemoryMiB.Min,
			Max: requirements.AcceleratorTotalMemoryMiB.Max,
		}
	}

	if requirements.AcceleratorTypes != nil {
		requirementsRequest.AcceleratorTypes = requirements.AcceleratorTypes
	}

	if requirements.BareMetal != "" {
		requirementsRequest.BareMetal = requirements.BareMetal
	}

	if requirements.BaselineEbsBandwidthMbps != nil {
		requirementsRequest.BaselineEbsBandwidthMbps = &ec2types.BaselineEbsBandwidthMbpsRequest{
			Min: requirements.BaselineEbsBandwidthMbps.Min,
			Max: requirements.BaselineEbsBandwidthMbps.Max,
		}
	}

	if requirements.BurstablePerformance != "" {
		requirementsRequest.BurstablePerformance = requirements.BurstablePerformance
	}

	if requirements.CpuManufacturers != nil {
		requirementsRequest.CpuManufacturers = requirements.CpuManufacturers
	}

	if requirements.ExcludedInstanceTypes != nil {
		requirementsRequest.ExcludedInstanceTypes = requirements.ExcludedInstanceTypes
	}

	if requirements.InstanceGenerations != nil {
		requirementsRequest.InstanceGenerations = requirements.InstanceGenerations
	}

	if requirements.LocalStorage != "" {
		requirementsRequest.LocalStorage = requirements.LocalStorage
	}

	if requirements.LocalStorageTypes != nil {
		requirementsRequest.LocalStorageTypes = requirements.LocalStorageTypes
	}

	if requirements.MemoryGiBPerVCpu != nil {
		requirementsRequest.MemoryGiBPerVCpu = &ec2types.MemoryGiBPerVCpuRequest{
			Min: requirements.MemoryGiBPerVCpu.Min,
			Max: requirements.MemoryGiBPerVCpu.Max,
		}
	}

	if requirements.NetworkInterfaceCount != nil {
		requirementsRequest.NetworkInterfaceCount = &ec2types.NetworkInterfaceCountRequest{
			Min: requirements.NetworkInterfaceCount.Min,
			Max: requirements.NetworkInterfaceCount.Max,
		}
	}

	if requirements.OnDemandMaxPricePercentageOverLowestPrice != nil {
		requirementsRequest.OnDemandMaxPricePercentageOverLowestPrice = requirements.OnDemandMaxPricePercentageOverLowestPrice
	}

	if requirements.RequireHibernateSupport != nil {
		requirementsRequest.RequireHibernateSupport = requirements.RequireHibernateSupport
	}

	if requirements.SpotMaxPricePercentageOverLowestPrice != nil {
		requirementsRequest.SpotMaxPricePercentageOverLowestPrice = requirements.SpotMaxPricePercentageOverLowestPrice
	}

	if requirements.TotalLocalStorageGB != nil {
		requirementsRequest.TotalLocalStorageGB = &ec2types.TotalLocalStorageGBRequest{
			Min: requirements.TotalLocalStorageGB.Min,
			Max: requirements.TotalLocalStorageGB.Max,
		}
	}

	return &requirementsRequest, nil
}

func (m *awsWrapper) getEC2RequirementsFromAutoscaling(autoscalingRequirements *autoscalingtypes.InstanceRequirements) (*ec2types.InstanceRequirements, error) {
	ec2Requirements := ec2types.InstanceRequirements{}

	// required instance requirements
	ec2Requirements.MemoryMiB = &ec2types.MemoryMiB{
		Min: autoscalingRequirements.MemoryMiB.Min,
		Max: autoscalingRequirements.MemoryMiB.Max,
	}

	ec2Requirements.VCpuCount = &ec2types.VCpuCountRange{
		Min: autoscalingRequirements.VCpuCount.Min,
		Max: autoscalingRequirements.VCpuCount.Max,
	}

	// optional instance requirements
	if autoscalingRequirements.AcceleratorCount != nil {
		ec2Requirements.AcceleratorCount = &ec2types.AcceleratorCount{
			Min: autoscalingRequirements.AcceleratorCount.Min,
			Max: autoscalingRequirements.AcceleratorCount.Max,
		}
	}

	if autoscalingRequirements.AcceleratorManufacturers != nil {
		var acceleratorManufacturers []ec2types.AcceleratorManufacturer
		for _, manufacturer := range autoscalingRequirements.AcceleratorManufacturers {
			acceleratorManufacturers = append(acceleratorManufacturers, ec2types.AcceleratorManufacturer(manufacturer))
		}
		ec2Requirements.AcceleratorManufacturers = acceleratorManufacturers
	}

	if autoscalingRequirements.AcceleratorNames != nil {
		var acceleratorNames []ec2types.AcceleratorName
		for _, acceleratorName := range autoscalingRequirements.AcceleratorNames {
			acceleratorNames = append(acceleratorNames, ec2types.AcceleratorName(acceleratorName))
		}
		ec2Requirements.AcceleratorNames = acceleratorNames
	}

	if autoscalingRequirements.AcceleratorTotalMemoryMiB != nil {
		ec2Requirements.AcceleratorTotalMemoryMiB = &ec2types.AcceleratorTotalMemoryMiB{
			Min: autoscalingRequirements.AcceleratorTotalMemoryMiB.Min,
			Max: autoscalingRequirements.AcceleratorTotalMemoryMiB.Max,
		}
	}

	if autoscalingRequirements.AcceleratorTypes != nil {
		var acceleratorTypes []ec2types.AcceleratorType
		for _, acceleratorType := range autoscalingRequirements.AcceleratorTypes {
			acceleratorTypes = append(acceleratorTypes, ec2types.AcceleratorType(acceleratorType))
		}
		ec2Requirements.AcceleratorTypes = acceleratorTypes
	}

	if autoscalingRequirements.BareMetal != "" {
		ec2Requirements.BareMetal = ec2types.BareMetal(autoscalingRequirements.BareMetal)
	}

	if autoscalingRequirements.BaselineEbsBandwidthMbps != nil {
		ec2Requirements.BaselineEbsBandwidthMbps = &ec2types.BaselineEbsBandwidthMbps{
			Min: autoscalingRequirements.BaselineEbsBandwidthMbps.Min,
			Max: autoscalingRequirements.BaselineEbsBandwidthMbps.Max,
		}
	}

	if autoscalingRequirements.BurstablePerformance != "" {
		ec2Requirements.BurstablePerformance = ec2types.BurstablePerformance(autoscalingRequirements.BurstablePerformance)
	}

	if autoscalingRequirements.CpuManufacturers != nil {
		var cpuManufacturers []ec2types.CpuManufacturer
		for _, manufacturer := range autoscalingRequirements.CpuManufacturers {
			cpuManufacturers = append(cpuManufacturers, ec2types.CpuManufacturer(manufacturer))
		}
		ec2Requirements.CpuManufacturers = cpuManufacturers
	}

	if autoscalingRequirements.ExcludedInstanceTypes != nil {
		ec2Requirements.ExcludedInstanceTypes = autoscalingRequirements.ExcludedInstanceTypes
	}

	if autoscalingRequirements.InstanceGenerations != nil {
		var instanceGenerations []ec2types.InstanceGeneration
		for _, generation := range autoscalingRequirements.InstanceGenerations {
			instanceGenerations = append(instanceGenerations, ec2types.InstanceGeneration(generation))
		}
		ec2Requirements.InstanceGenerations = instanceGenerations
	}

	if autoscalingRequirements.LocalStorage != "" {
		ec2Requirements.LocalStorage = ec2types.LocalStorage(autoscalingRequirements.LocalStorage)
	}

	if autoscalingRequirements.LocalStorageTypes != nil {
		var localStorageTypes []ec2types.LocalStorageType
		for _, localStorageType := range autoscalingRequirements.LocalStorageTypes {
			localStorageTypes = append(localStorageTypes, ec2types.LocalStorageType(localStorageType))
		}
		ec2Requirements.LocalStorageTypes = localStorageTypes
	}

	if autoscalingRequirements.MemoryGiBPerVCpu != nil {
		ec2Requirements.MemoryGiBPerVCpu = &ec2types.MemoryGiBPerVCpu{
			Min: autoscalingRequirements.MemoryGiBPerVCpu.Min,
			Max: autoscalingRequirements.MemoryGiBPerVCpu.Max,
		}
	}

	if autoscalingRequirements.NetworkInterfaceCount != nil {
		ec2Requirements.NetworkInterfaceCount = &ec2types.NetworkInterfaceCount{
			Min: autoscalingRequirements.NetworkInterfaceCount.Min,
			Max: autoscalingRequirements.NetworkInterfaceCount.Max,
		}
	}

	if autoscalingRequirements.OnDemandMaxPricePercentageOverLowestPrice != nil {
		ec2Requirements.OnDemandMaxPricePercentageOverLowestPrice = autoscalingRequirements.OnDemandMaxPricePercentageOverLowestPrice
	}

	if autoscalingRequirements.RequireHibernateSupport != nil {
		ec2Requirements.RequireHibernateSupport = autoscalingRequirements.RequireHibernateSupport
	}

	if autoscalingRequirements.SpotMaxPricePercentageOverLowestPrice != nil {
		ec2Requirements.SpotMaxPricePercentageOverLowestPrice = autoscalingRequirements.SpotMaxPricePercentageOverLowestPrice
	}

	if autoscalingRequirements.TotalLocalStorageGB != nil {
		ec2Requirements.TotalLocalStorageGB = &ec2types.TotalLocalStorageGB{
			Min: autoscalingRequirements.TotalLocalStorageGB.Min,
			Max: autoscalingRequirements.TotalLocalStorageGB.Max,
		}
	}

	return &ec2Requirements, nil
}

func (m *awsWrapper) getInstanceTypesForAsgs(asgs []*asg) (map[string]string, error) {
	results := map[string]string{}
	launchConfigsToQuery := map[string]string{}
	launchTemplatesToQuery := map[string]*launchTemplate{}
	mixedInstancesPoliciesToQuery := map[string]*mixedInstancesPolicy{}

	for _, asg := range asgs {
		name := asg.AwsRef.Name
		if asg.LaunchConfigurationName != "" {
			launchConfigsToQuery[name] = asg.LaunchConfigurationName
		} else if asg.LaunchTemplate != nil {
			launchTemplatesToQuery[name] = asg.LaunchTemplate
		} else if asg.MixedInstancesPolicy != nil {
			if len(asg.MixedInstancesPolicy.instanceTypesOverrides) > 0 {
				results[name] = asg.MixedInstancesPolicy.instanceTypesOverrides[0]
			} else if asg.MixedInstancesPolicy.instanceRequirementsOverrides != nil {
				mixedInstancesPoliciesToQuery[name] = asg.MixedInstancesPolicy
			} else {
				launchTemplatesToQuery[name] = asg.MixedInstancesPolicy.launchTemplate
			}
		}
	}

	klog.V(4).Infof("%d launch configurations to query", len(launchConfigsToQuery))
	klog.V(4).Infof("%d launch templates to query", len(launchTemplatesToQuery))

	// Query these all at once to minimize AWS API calls
	launchConfigNames := make([]string, 0, len(launchConfigsToQuery))
	for _, cfgName := range launchConfigsToQuery {
		launchConfigNames = append(launchConfigNames, cfgName)
	}
	launchConfigs, err := m.getInstanceTypeByLaunchConfigNames(launchConfigNames)
	if err != nil {
		klog.Errorf("Failed to query %d launch configurations", len(launchConfigsToQuery))
		return nil, err
	}

	for asgName, cfgName := range launchConfigsToQuery {
		if instanceType, ok := launchConfigs[cfgName]; !ok || instanceType == "" {
			klog.Warningf("Could not fetch %q launch configuration for ASG %q", cfgName, asgName)
			continue
		}
		results[asgName] = launchConfigs[cfgName]
	}
	klog.V(4).Infof("Successfully queried %d launch configurations", len(launchConfigs))

	// Have to query LaunchTemplates one-at-a-time, since there's no way to query <lt, version> pairs in bulk
	for asgName, lt := range launchTemplatesToQuery {
		instanceType, err := m.getInstanceTypeByLaunchTemplate(lt)
		if err != nil {
			klog.Errorf("Failed to query launch template %s: %v", lt.name, err)
			continue
		}
		results[asgName] = instanceType
	}
	klog.V(4).Infof("Successfully queried %d launch templates", len(launchTemplatesToQuery))

	// Have to match Instance Requirements one-at-a-time, since they are configured per asg and can't be queried in bulk
	for asgName, policy := range mixedInstancesPoliciesToQuery {
		instanceType, err := m.getInstanceTypeFromRequirementsOverrides(policy)
		if err != nil {
			klog.Errorf("Failed to query instance requirements for ASG %s: %v", asgName, err)
			continue
		}
		results[asgName] = instanceType
	}
	klog.V(4).Infof("Successfully queried instance requirements for %d ASGs", len(mixedInstancesPoliciesToQuery))

	return results, nil
}

func buildLaunchTemplateFromSpec(ltSpec *autoscalingtypes.LaunchTemplateSpecification) *launchTemplate {
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
	// https://github.com/aws/aws-sdk-go-v2/blob/25f26c231f47889249f6bf9b6b11ed817097abdf/service/autoscaling/types/types.go#L1897-L1905
	//
	// A cleaner alternative might be to make launchTemplate.version a string
	// pointer instead of a string, or even store the aws-sdk-go's
	// LaunchTemplateSpecification structs directly.
	var version string
	if ltSpec.Version == nil {
		version = "$Default"
	} else {
		version = aws.ToString(ltSpec.Version)
	}
	return &launchTemplate{
		name:    aws.ToString(ltSpec.LaunchTemplateName),
		version: version,
	}
}

func taintEksTranslator(t ekstypes.Taint) (apiv1.TaintEffect, error) {
	// Translation between AWS EKS and Kubernetes taints
	//
	// See:
	//
	// https://docs.aws.amazon.com/eks/latest/APIReference/API_Taint.html
	switch effect := t.Effect; effect {
	case ekstypes.TaintEffectNoSchedule:
		return apiv1.TaintEffectNoSchedule, nil
	case ekstypes.TaintEffectNoExecute:
		return apiv1.TaintEffectNoExecute, nil
	case ekstypes.TaintEffectPreferNoSchedule:
		return apiv1.TaintEffectPreferNoSchedule, nil
	default:
		return "", fmt.Errorf("couldn't translate EKS DescribeNodegroup response taint %s into Kubernetes format", effect)
	}
}
