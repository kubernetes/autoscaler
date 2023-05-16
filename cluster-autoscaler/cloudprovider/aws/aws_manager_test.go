/*
Copyright 2017 The Kubernetes Authors.

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
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/aws"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/service/autoscaling"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/service/ec2"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
)

func TestJoinNodeLabelsChoosingUserValuesOverAPIValues(t *testing.T) {
	extractedLabels := make(map[string]string)
	mngLabels := make(map[string]string)

	extractedLabels["key1"] = "value1extracted"
	extractedLabels["key2"] = "value2extracted"

	mngLabels["key3"] = "value3mng"
	mngLabels["key2"] = "value2mng"

	result := joinNodeLabelsChoosingUserValuesOverAPIValues(extractedLabels, mngLabels)

	// Make sure any duplicate keys keep the extractedLabels value
	assert.Equal(t, result["key1"], "value1extracted")
	assert.Equal(t, result["key2"], "value2mng")
	assert.Equal(t, result["key3"], "value3mng")
}

func TestBuildGenericLabels(t *testing.T) {
	labels := buildGenericLabels(&asgTemplate{
		InstanceType: &InstanceType{
			InstanceType: "c4.large",
			VCPU:         2,
			MemoryMb:     3840,
			Architecture: cloudprovider.DefaultArch,
		},
		Region: "us-east-1",
	}, "sillyname")
	assert.Equal(t, "us-east-1", labels[apiv1.LabelZoneRegionStable])
	assert.Equal(t, "sillyname", labels[apiv1.LabelHostname])
	assert.Equal(t, "c4.large", labels[apiv1.LabelInstanceTypeStable])
	assert.Equal(t, cloudprovider.DefaultArch, labels[apiv1.LabelArchStable])
	assert.Equal(t, cloudprovider.DefaultOS, labels[apiv1.LabelOSStable])
}

func TestExtractAllocatableResourcesFromAsg(t *testing.T) {
	tags := []*autoscaling.TagDescription{
		{
			Key:   aws.String("k8s.io/cluster-autoscaler/node-template/resources/cpu"),
			Value: aws.String("100m"),
		},
		{
			Key:   aws.String("k8s.io/cluster-autoscaler/node-template/resources/memory"),
			Value: aws.String("100M"),
		},
		{
			Key:   aws.String("k8s.io/cluster-autoscaler/node-template/resources/ephemeral-storage"),
			Value: aws.String("20G"),
		},
		{
			Key:   aws.String("k8s.io/cluster-autoscaler/node-template/resources/custom-resource"),
			Value: aws.String("5"),
		},
	}

	labels := extractAllocatableResourcesFromAsg(tags)

	assert.Equal(t, resource.NewMilliQuantity(100, resource.DecimalSI).String(), labels["cpu"].String())
	expectedMemory := resource.MustParse("100M")
	assert.Equal(t, (&expectedMemory).String(), labels["memory"].String())
	expectedEphemeralStorage := resource.MustParse("20G")
	assert.Equal(t, (&expectedEphemeralStorage).String(), labels["ephemeral-storage"].String())
	assert.Equal(t, resource.NewQuantity(5, resource.DecimalSI).String(), labels["custom-resource"].String())
}

func TestGetAsgOptions(t *testing.T) {
	defaultOptions := config.NodeGroupAutoscalingOptions{
		ScaleDownUtilizationThreshold:    0.1,
		ScaleDownGpuUtilizationThreshold: 0.2,
		ScaleDownUnneededTime:            time.Second,
		ScaleDownUnreadyTime:             time.Minute,
	}

	tests := []struct {
		description string
		tags        map[string]string
		expected    *config.NodeGroupAutoscalingOptions
	}{
		{
			description: "use defaults on unspecified tags",
			tags:        make(map[string]string),
			expected:    &defaultOptions,
		},
		{
			description: "keep defaults on invalid tags values",
			tags: map[string]string{
				"scaledownutilizationthreshold": "not-a-float",
				"scaledownunneededtime":         "not-a-duration",
				"ScaleDownUnreadyTime":          "",
			},
			expected: &defaultOptions,
		},
		{
			description: "use provided tags and fill missing with defaults",
			tags: map[string]string{
				"scaledownutilizationthreshold": "0.42",
				"scaledownunneededtime":         "1h",
			},
			expected: &config.NodeGroupAutoscalingOptions{
				ScaleDownUtilizationThreshold:    0.42,
				ScaleDownGpuUtilizationThreshold: defaultOptions.ScaleDownGpuUtilizationThreshold,
				ScaleDownUnneededTime:            time.Hour,
				ScaleDownUnreadyTime:             defaultOptions.ScaleDownUnreadyTime,
			},
		},
		{
			description: "ignore unknown tags",
			tags: map[string]string{
				"scaledownutilizationthreshold":    "0.6",
				"scaledowngpuutilizationthreshold": "0.7",
				"scaledownunneededtime":            "1m",
				"scaledownunreadytime":             "1h",
				"notyetspecified":                  "42",
			},
			expected: &config.NodeGroupAutoscalingOptions{
				ScaleDownUtilizationThreshold:    0.6,
				ScaleDownGpuUtilizationThreshold: 0.7,
				ScaleDownUnneededTime:            time.Minute,
				ScaleDownUnreadyTime:             time.Hour,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			testAsg := asg{AwsRef: AwsRef{Name: "testAsg"}}
			cache, _ := newASGCache(nil, []string{}, []asgAutoDiscoveryConfig{})
			cache.autoscalingOptions[testAsg.AwsRef] = tt.tags
			awsManager := &AwsManager{asgCache: cache}

			actual := awsManager.GetAsgOptions(testAsg, defaultOptions)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestBuildNodeFromTemplateWithManagedNodegroup(t *testing.T) {
	mngCache := newManagedNodeGroupCache(nil)
	awsManager := &AwsManager{managedNodegroupCache: mngCache}
	asg := &asg{AwsRef: AwsRef{Name: "test-auto-scaling-group"}}
	c5Instance := &InstanceType{
		InstanceType: "c5.xlarge",
		VCPU:         4,
		MemoryMb:     8192,
		GPU:          0,
	}

	ngNameLabelValue := "nodegroup-1"
	clusterNameLabelValue := "cluster-1"

	labelKey1 := "labelKey 1"
	labelKey2 := "labelKey 2"
	labelValue1 := "testValue 1"
	labelValue2 := "testValue 2"
	labelValue3 := "testValue 3"

	taintEffect1 := "effect 1"
	taintKey1 := "key 1"
	taintValue1 := "value 1"
	taint1 := apiv1.Taint{
		Effect: apiv1.TaintEffect(taintEffect1),
		Key:    taintKey1,
		Value:  taintValue1,
	}

	taintEffect2 := "effect 2"
	taintKey2 := "key 2"
	taintValue2 := "value 2"
	taint2 := apiv1.Taint{
		Effect: apiv1.TaintEffect(taintEffect2),
		Key:    taintKey2,
		Value:  taintValue2,
	}

	err := mngCache.Add(managedNodegroupCachedObject{
		name:        ngNameLabelValue,
		clusterName: clusterNameLabelValue,
		taints:      []apiv1.Taint{taint1, taint2},
		labels:      map[string]string{labelKey1: labelValue1, labelKey2: labelValue2},
	})
	require.NoError(t, err)

	// Node with EKS labels
	observedNode, observedErr := awsManager.buildNodeFromTemplate(asg, &asgTemplate{
		InstanceType: c5Instance,
		Tags: []*autoscaling.TagDescription{
			{
				Key:   aws.String("eks:nodegroup-name"),
				Value: aws.String(ngNameLabelValue),
			},
			{
				Key:   aws.String("eks:cluster-name"),
				Value: aws.String(clusterNameLabelValue),
			},
			{
				Key:   aws.String("k8s.io/cluster-autoscaler/node-template/label/" + labelKey2),
				Value: aws.String(labelValue3),
			},
		},
	})
	assert.NoError(t, observedErr)
	assert.GreaterOrEqual(t, len(observedNode.Labels), 4)
	ngNameValue, ngLabelExist := observedNode.Labels["nodegroup-name"]
	assert.True(t, ngLabelExist)
	assert.Equal(t, ngNameLabelValue, ngNameValue)
	clusterNameValue, clusterLabelExist := observedNode.Labels["cluster-name"]
	assert.True(t, clusterLabelExist)
	assert.Equal(t, clusterNameValue, clusterNameLabelValue)
	labelKeyValue1, labelKeyExist1 := observedNode.Labels[labelKey1]
	assert.True(t, labelKeyExist1)
	assert.Equal(t, labelKeyValue1, labelValue1)
	labelKeyValue2, labelKeyExist2 := observedNode.Labels[labelKey2]
	assert.True(t, labelKeyExist2)
	// Check the value specified in the ASG tag is kept, instead of the EKS API value
	assert.Equal(t, labelKeyValue2, labelValue2)
	assert.Equal(t, len(observedNode.Spec.Taints), 2)
	assert.Equal(t, observedNode.Spec.Taints[0].Effect, apiv1.TaintEffect(taintEffect1))
	assert.Equal(t, observedNode.Spec.Taints[0].Key, taintKey1)
	assert.Equal(t, observedNode.Spec.Taints[0].Value, taintValue1)
	assert.Equal(t, observedNode.Spec.Taints[1].Effect, apiv1.TaintEffect(taintEffect2))
	assert.Equal(t, observedNode.Spec.Taints[1].Key, taintKey2)
	assert.Equal(t, observedNode.Spec.Taints[1].Value, taintValue2)
}

func TestBuildNodeFromTemplateWithManagedNodegroupNoLabelsOrTaints(t *testing.T) {
	mngCache := newManagedNodeGroupCache(nil)
	awsManager := &AwsManager{managedNodegroupCache: mngCache}
	asg := &asg{AwsRef: AwsRef{Name: "test-auto-scaling-group"}}
	c5Instance := &InstanceType{
		InstanceType: "c5.xlarge",
		VCPU:         4,
		MemoryMb:     8192,
		GPU:          0,
	}

	ngNameLabelValue := "nodegroup-1"
	clusterNameLabelValue := "cluster-1"

	err := mngCache.Add(managedNodegroupCachedObject{
		name:        ngNameLabelValue,
		clusterName: clusterNameLabelValue,
		taints:      make([]apiv1.Taint, 0),
		labels:      make(map[string]string),
	})
	require.NoError(t, err)

	// Node with EKS labels
	observedNode, observedErr := awsManager.buildNodeFromTemplate(asg, &asgTemplate{
		InstanceType: c5Instance,
		Tags: []*autoscaling.TagDescription{
			{
				Key:   aws.String("eks:nodegroup-name"),
				Value: aws.String(ngNameLabelValue),
			},
			{
				Key:   aws.String("eks:cluster-name"),
				Value: aws.String(clusterNameLabelValue),
			},
		},
	})
	assert.NoError(t, observedErr)
	assert.GreaterOrEqual(t, len(observedNode.Labels), 2)
	ngNameValue, ngLabelExist := observedNode.Labels["nodegroup-name"]
	assert.True(t, ngLabelExist)
	assert.Equal(t, ngNameLabelValue, ngNameValue)
	clusterNameValue, clusterLabelExist := observedNode.Labels["cluster-name"]
	assert.True(t, clusterLabelExist)
	assert.Equal(t, clusterNameValue, clusterNameLabelValue)
	assert.Equal(t, len(observedNode.Spec.Taints), 0)
}

func TestBuildNodeFromTemplateWithManagedNodegroupNilLabelsOrTaints(t *testing.T) {
	mngCache := newManagedNodeGroupCache(nil)
	awsManager := &AwsManager{managedNodegroupCache: mngCache}
	asg := &asg{AwsRef: AwsRef{Name: "test-auto-scaling-group"}}
	c5Instance := &InstanceType{
		InstanceType: "c5.xlarge",
		VCPU:         4,
		MemoryMb:     8192,
		GPU:          0,
	}

	ngNameLabelValue := "nodegroup-1"
	clusterNameLabelValue := "cluster-1"

	err := mngCache.Add(managedNodegroupCachedObject{
		name:        ngNameLabelValue,
		clusterName: clusterNameLabelValue,
		taints:      nil,
		labels:      nil,
	})
	require.NoError(t, err)

	// Node with EKS labels
	observedNode, observedErr := awsManager.buildNodeFromTemplate(asg, &asgTemplate{
		InstanceType: c5Instance,
		Tags: []*autoscaling.TagDescription{
			{
				Key:   aws.String("eks:nodegroup-name"),
				Value: aws.String(ngNameLabelValue),
			},
			{
				Key:   aws.String("eks:cluster-name"),
				Value: aws.String(clusterNameLabelValue),
			},
		},
	})
	assert.NoError(t, observedErr)
	assert.GreaterOrEqual(t, len(observedNode.Labels), 2)
	ngNameValue, ngLabelExist := observedNode.Labels["nodegroup-name"]
	assert.True(t, ngLabelExist)
	assert.Equal(t, ngNameLabelValue, ngNameValue)
	clusterNameValue, clusterLabelExist := observedNode.Labels["cluster-name"]
	assert.True(t, clusterLabelExist)
	assert.Equal(t, clusterNameValue, clusterNameLabelValue)
	assert.Equal(t, len(observedNode.Spec.Taints), 0)
}

func TestBuildNodeFromTemplate(t *testing.T) {
	awsManager := &AwsManager{}
	asg := &asg{AwsRef: AwsRef{Name: "test-auto-scaling-group"}}
	c5Instance := &InstanceType{
		InstanceType: "c5.xlarge",
		VCPU:         4,
		MemoryMb:     8192,
		GPU:          0,
	}

	// Node with custom resource
	ephemeralStorageKey := "ephemeral-storage"
	ephemeralStorageValue := int64(20)
	customResourceKey := "custom-resource"
	customResourceValue := int64(5)
	vpcIPKey := "vpc.amazonaws.com/PrivateIPv4Address"
	observedNode, observedErr := awsManager.buildNodeFromTemplate(asg, &asgTemplate{
		InstanceType: c5Instance,
		Tags: []*autoscaling.TagDescription{
			{
				Key:   aws.String(fmt.Sprintf("k8s.io/cluster-autoscaler/node-template/resources/%s", ephemeralStorageKey)),
				Value: aws.String(strconv.FormatInt(ephemeralStorageValue, 10)),
			},
			{
				Key:   aws.String(fmt.Sprintf("k8s.io/cluster-autoscaler/node-template/resources/%s", customResourceKey)),
				Value: aws.String(strconv.FormatInt(customResourceValue, 10)),
			},
		},
	})
	assert.NoError(t, observedErr)
	esValue, esExist := observedNode.Status.Capacity[apiv1.ResourceName(ephemeralStorageKey)]
	assert.True(t, esExist)
	assert.Equal(t, int64(20), esValue.Value())
	crValue, crExist := observedNode.Status.Capacity[apiv1.ResourceName(customResourceKey)]
	assert.True(t, crExist)
	assert.Equal(t, int64(5), crValue.Value())
	_, ipExist := observedNode.Status.Capacity[apiv1.ResourceName(vpcIPKey)]
	assert.False(t, ipExist)

	// Node with labels
	GPULabelValue := "nvidia-telsa-v100"
	observedNode, observedErr = awsManager.buildNodeFromTemplate(asg, &asgTemplate{
		InstanceType: c5Instance,
		Tags: []*autoscaling.TagDescription{
			{
				Key:   aws.String(fmt.Sprintf("k8s.io/cluster-autoscaler/node-template/label/%s", GPULabel)),
				Value: aws.String(GPULabelValue),
			},
		},
	})
	assert.NoError(t, observedErr)
	gpuValue, gpuLabelExist := observedNode.Labels[GPULabel]
	assert.True(t, gpuLabelExist)
	assert.Equal(t, GPULabelValue, gpuValue)

	// Node with EKS labels
	ngNameLabelValue := "nodegroup-1"
	observedNode, observedErr = awsManager.buildNodeFromTemplate(asg, &asgTemplate{
		InstanceType: c5Instance,
		Tags: []*autoscaling.TagDescription{
			{
				Key:   aws.String("eks:nodegroup-name"),
				Value: aws.String(ngNameLabelValue),
			},
		},
	})
	assert.NoError(t, observedErr)
	ngNameValue, ngLabelExist := observedNode.Labels["nodegroup-name"]
	assert.True(t, ngLabelExist)
	assert.Equal(t, ngNameLabelValue, ngNameValue)

	// Node with taints
	gpuTaint := apiv1.Taint{
		Key:    "nvidia.com/gpu",
		Value:  "present",
		Effect: "NoSchedule",
	}
	observedNode, observedErr = awsManager.buildNodeFromTemplate(asg, &asgTemplate{
		InstanceType: c5Instance,
		Tags: []*autoscaling.TagDescription{
			{
				Key:   aws.String(fmt.Sprintf("k8s.io/cluster-autoscaler/node-template/taint/%s", gpuTaint.Key)),
				Value: aws.String(fmt.Sprintf("%s:%s", gpuTaint.Value, gpuTaint.Effect)),
			},
		},
	})

	assert.NoError(t, observedErr)
	observedTaints := observedNode.Spec.Taints
	assert.Equal(t, 1, len(observedTaints))
	assert.Equal(t, gpuTaint, observedTaints[0])

	// Node with instance requirements
	asg.MixedInstancesPolicy = &mixedInstancesPolicy{
		instanceRequirementsOverrides: &autoscaling.InstanceRequirements{
			VCpuCount: &autoscaling.VCpuCountRequest{
				Min: aws.Int64(4),
				Max: aws.Int64(8),
			},
			MemoryMiB: &autoscaling.MemoryMiBRequest{
				Min: aws.Int64(4),
				Max: aws.Int64(8),
			},
			AcceleratorTypes:         []*string{aws.String(autoscaling.AcceleratorTypeGpu)},
			AcceleratorManufacturers: []*string{aws.String(autoscaling.AcceleratorManufacturerNvidia)},
			AcceleratorCount: &autoscaling.AcceleratorCountRequest{
				Min: aws.Int64(4),
				Max: aws.Int64(8),
			},
		},
	}
	observedNode, observedErr = awsManager.buildNodeFromTemplate(asg, &asgTemplate{
		InstanceType: c5Instance,
	})

	assert.NoError(t, observedErr)
	observedMemoryRequirement := observedNode.Status.Capacity[apiv1.ResourceMemory]
	assert.Equal(t, int64(4*1024*1024), observedMemoryRequirement.Value())
	observedVCpuRequirement := observedNode.Status.Capacity[apiv1.ResourceCPU]
	assert.Equal(t, int64(4), observedVCpuRequirement.Value())
	observedGpuRequirement := observedNode.Status.Capacity[gpu.ResourceNvidiaGPU]
	assert.Equal(t, int64(4), observedGpuRequirement.Value())
}

func TestExtractLabelsFromAsg(t *testing.T) {
	tags := []*autoscaling.TagDescription{
		{
			Key:   aws.String("k8s.io/cluster-autoscaler/node-template/label/foo"),
			Value: aws.String("bar"),
		},
		{
			Key:   aws.String("eks:nodegroup-name"),
			Value: aws.String("bar2"),
		},
		{
			Key:   aws.String("eks:cluster-name"),
			Value: aws.String("bar4"),
		},
		{
			Key:   aws.String("bar"),
			Value: aws.String("baz"),
		},
	}

	labels := extractLabelsFromAsg(tags)

	assert.Equal(t, 3, len(labels))
	assert.Equal(t, "bar", labels["foo"])
	assert.Equal(t, "bar2", labels["nodegroup-name"])
	assert.Equal(t, "bar4", labels["cluster-name"])
}

func TestExtractTaintsFromAsg(t *testing.T) {
	tags := []*autoscaling.TagDescription{
		{
			Key:   aws.String("k8s.io/cluster-autoscaler/node-template/taint/dedicated"),
			Value: aws.String("foo:NoSchedule"),
		},
		{
			Key:   aws.String("k8s.io/cluster-autoscaler/node-template/taint/group"),
			Value: aws.String("bar:NoExecute"),
		},
		{
			Key:   aws.String("k8s.io/cluster-autoscaler/node-template/taint/app"),
			Value: aws.String("fizz:PreferNoSchedule"),
		},
		{
			Key:   aws.String("bar"),
			Value: aws.String("baz"),
		},
		{
			Key:   aws.String("k8s.io/cluster-autoscaler/node-template/taint/blank"),
			Value: aws.String(""),
		},
		{
			Key:   aws.String("k8s.io/cluster-autoscaler/node-template/taint/nosplit"),
			Value: aws.String("some_value"),
		},
	}

	expectedTaints := []apiv1.Taint{
		{
			Key:    "dedicated",
			Value:  "foo",
			Effect: apiv1.TaintEffectNoSchedule,
		},
		{
			Key:    "group",
			Value:  "bar",
			Effect: apiv1.TaintEffectNoExecute,
		},
		{
			Key:    "app",
			Value:  "fizz",
			Effect: apiv1.TaintEffectPreferNoSchedule,
		},
	}

	taints := extractTaintsFromAsg(tags)
	assert.Equal(t, 3, len(taints))
	assert.Equal(t, makeTaintSet(expectedTaints), makeTaintSet(taints))
}

func makeTaintSet(taints []apiv1.Taint) map[apiv1.Taint]bool {
	set := make(map[apiv1.Taint]bool)
	for _, taint := range taints {
		set[taint] = true
	}
	return set
}

func TestFetchExplicitAsgs(t *testing.T) {
	min, max, groupname := 1, 10, "coolasg"
	asgRef := AwsRef{Name: groupname}

	a := &autoScalingMock{}
	a.On("DescribeAutoScalingGroups", &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{aws.String(groupname)},
		MaxRecords:            aws.Int64(1),
	}).Return(&autoscaling.DescribeAutoScalingGroupsOutput{
		AutoScalingGroups: []*autoscaling.Group{
			{AutoScalingGroupName: aws.String(groupname)},
		},
	})

	a.On("DescribeAutoScalingGroupsPages",
		&autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: aws.StringSlice([]string{groupname}),
			MaxRecords:            aws.Int64(maxRecordsReturnedByAPI),
		},
		mock.AnythingOfType("func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool"),
	).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool)
		zone := "test-1a"
		fn(&autoscaling.DescribeAutoScalingGroupsOutput{
			AutoScalingGroups: []*autoscaling.Group{
				{
					AvailabilityZones:    []*string{&zone},
					AutoScalingGroupName: aws.String(groupname),
					MinSize:              aws.Int64(int64(min)),
					MaxSize:              aws.Int64(int64(max)),
					DesiredCapacity:      aws.Int64(int64(min)),
				},
			}}, false)
	}).Return(nil)

	a.On("DescribeScalingActivities",
		&autoscaling.DescribeScalingActivitiesInput{
			AutoScalingGroupName: aws.String("coolasg"),
		},
	).Return(&autoscaling.DescribeScalingActivitiesOutput{}, nil)

	do := cloudprovider.NodeGroupDiscoveryOptions{
		// Register the same node group twice with different max nodes.
		// The intention is to test that the asgs.Register method will update
		// the node group instead of registering it twice.
		NodeGroupSpecs: []string{
			fmt.Sprintf("%d:%d:%s", min, max, groupname),
			fmt.Sprintf("%d:%d:%s", min, max-1, groupname),
		},
	}
	t.Setenv("AWS_REGION", "fanghorn")
	instanceTypes, _ := GetStaticEC2InstanceTypes()
	m, err := createAWSManagerInternal(nil, do, &awsWrapper{a, nil, nil}, instanceTypes)
	assert.NoError(t, err)

	asgs := m.asgCache.Get()
	assert.Equal(t, 1, len(asgs))
	validateAsg(t, asgs[asgRef], groupname, min, max)
}

func TestGetASGTemplate(t *testing.T) {
	const (
		asgName           = "sample"
		knownInstanceType = "t3.micro"
		region            = "us-east-1"
		az                = region + "a"
		ltName            = "launcher"
		ltVersion         = "1"
	)

	asgRef := AwsRef{Name: asgName}

	tags := []*autoscaling.TagDescription{
		{
			Key:   aws.String("k8s.io/cluster-autoscaler/node-template/taint/dedicated"),
			Value: aws.String("foo:NoSchedule"),
		},
	}

	tests := []struct {
		description       string
		instanceType      string
		availabilityZones []string
		error             bool
	}{
		{"insufficient availability zones",
			knownInstanceType, []string{}, true},
		{"single availability zone",
			knownInstanceType, []string{az}, false},
		{"multiple availability zones",
			knownInstanceType, []string{az, "us-west-1b"}, false},
		{"unknown instance type",
			"nonexistent.xlarge", []string{az}, true},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			e := &ec2Mock{}
			e.On("DescribeLaunchTemplateVersions", &ec2.DescribeLaunchTemplateVersionsInput{
				LaunchTemplateName: aws.String(ltName),
				Versions:           []*string{aws.String(ltVersion)},
			}).Return(&ec2.DescribeLaunchTemplateVersionsOutput{
				LaunchTemplateVersions: []*ec2.LaunchTemplateVersion{
					{
						LaunchTemplateData: &ec2.ResponseLaunchTemplateData{
							InstanceType: aws.String(test.instanceType),
						},
					},
				},
			})

			t.Setenv("AWS_REGION", "fanghorn")
			instanceTypes, _ := GetStaticEC2InstanceTypes()
			do := cloudprovider.NodeGroupDiscoveryOptions{}

			m, err := createAWSManagerInternal(nil, do, &awsWrapper{nil, e, nil}, instanceTypes)
			origGetInstanceTypeFunc := getInstanceTypeForAsg
			defer func() { getInstanceTypeForAsg = origGetInstanceTypeFunc }()
			getInstanceTypeForAsg = func(m *asgCache, asg *asg) (string, error) {
				return test.instanceType, nil
			}
			assert.NoError(t, err)

			asg := &asg{
				AwsRef:            asgRef,
				AvailabilityZones: test.availabilityZones,
				LaunchTemplate: &launchTemplate{
					name:    ltName,
					version: ltVersion},
				Tags: tags,
			}

			template, err := m.getAsgTemplate(asg)
			if test.error {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if assert.NotNil(t, template) {
					assert.Equal(t, test.instanceType, template.InstanceType.InstanceType)
					assert.Equal(t, region, template.Region)
					assert.Equal(t, test.availabilityZones[0], template.Zone)
					assert.Equal(t, tags, template.Tags)
				}
			}
		})
	}
}

func TestFetchAutoAsgs(t *testing.T) {
	min, max := 1, 10
	groupname, tags := "coolasg", []string{"tag", "anothertag"}
	asgRef := AwsRef{Name: groupname}

	a := &autoScalingMock{}

	// Describe the group to register it, then again to generate the instance
	// cache.
	a.On("DescribeAutoScalingGroupsPages",
		&autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: aws.StringSlice([]string{groupname}),
			MaxRecords:            aws.Int64(maxRecordsReturnedByAPI),
		},
		mock.AnythingOfType("func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool"),
	).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool)
		zone := "test-1a"
		fn(&autoscaling.DescribeAutoScalingGroupsOutput{
			AutoScalingGroups: []*autoscaling.Group{{
				AvailabilityZones:    []*string{&zone},
				AutoScalingGroupName: aws.String(groupname),
				MinSize:              aws.Int64(int64(min)),
				MaxSize:              aws.Int64(int64(max)),
				DesiredCapacity:      aws.Int64(int64(min)),
			}}}, false)
	}).Return(nil).Once()

	expectedGroupsInputWithTags := &autoscaling.DescribeAutoScalingGroupsInput{
		Filters: []*autoscaling.Filter{
			{Name: aws.String("tag-key"), Values: aws.StringSlice([]string{tags[0]})},
			{Name: aws.String("tag-key"), Values: aws.StringSlice([]string{tags[1]})},
		},
		MaxRecords: aws.Int64(maxRecordsReturnedByAPI),
	}
	a.On("DescribeAutoScalingGroupsPages",
		mock.MatchedBy(tagsMatcher(expectedGroupsInputWithTags)),
		mock.AnythingOfType("func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool"),
	).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool)
		zone := "test-1a"
		fn(&autoscaling.DescribeAutoScalingGroupsOutput{
			AutoScalingGroups: []*autoscaling.Group{{
				AvailabilityZones:    []*string{&zone},
				AutoScalingGroupName: aws.String(groupname),
				MinSize:              aws.Int64(int64(min)),
				MaxSize:              aws.Int64(int64(max)),
				DesiredCapacity:      aws.Int64(int64(min)),
			}}}, false)
	}).Return(nil).Once()

	a.On("DescribeScalingActivities",
		&autoscaling.DescribeScalingActivitiesInput{
			AutoScalingGroupName: aws.String("coolasg"),
		},
	).Return(&autoscaling.DescribeScalingActivitiesOutput{}, nil)

	do := cloudprovider.NodeGroupDiscoveryOptions{
		NodeGroupAutoDiscoverySpecs: []string{fmt.Sprintf("asg:tag=%s", strings.Join(tags, ","))},
	}

	t.Setenv("AWS_REGION", "fanghorn")
	// fetchAutoASGs is called at manager creation time, via forceRefresh
	instanceTypes, _ := GetStaticEC2InstanceTypes()
	m, err := createAWSManagerInternal(nil, do, &awsWrapper{a, nil, nil}, instanceTypes)
	assert.NoError(t, err)

	asgs := m.asgCache.Get()
	assert.Equal(t, 1, len(asgs))
	validateAsg(t, asgs[asgRef], groupname, min, max)

	// Simulate the previously discovered ASG disappearing
	a.On("DescribeAutoScalingGroupsPages",
		mock.MatchedBy(tagsMatcher(expectedGroupsInputWithTags)),
		mock.AnythingOfType("func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool"),
	).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool)
		fn(&autoscaling.DescribeAutoScalingGroupsOutput{
			AutoScalingGroups: []*autoscaling.Group{}}, false)
	}).Return(nil).Once()

	err = m.asgCache.regenerate()
	assert.NoError(t, err)
	assert.Empty(t, m.asgCache.Get())
}

type ServiceDescriptor struct {
	name                         string
	region                       string
	signingRegion, signingMethod string
	signingName                  string
}

func tagsMatcher(expected *autoscaling.DescribeAutoScalingGroupsInput) func(*autoscaling.DescribeAutoScalingGroupsInput) bool {
	return func(actual *autoscaling.DescribeAutoScalingGroupsInput) bool {
		expectedTags := flatTagSlice(expected.Filters)
		actualTags := flatTagSlice(actual.Filters)

		return *expected.MaxRecords == *actual.MaxRecords && reflect.DeepEqual(expectedTags, actualTags)
	}
}

func flatTagSlice(filters []*autoscaling.Filter) []string {
	tags := []string{}
	for _, filter := range filters {
		tags = append(tags, aws.StringValueSlice(filter.Values)...)
	}
	// Sort slice for compare
	sort.Strings(tags)
	return tags
}

func TestParseASGAutoDiscoverySpecs(t *testing.T) {
	cases := []struct {
		name    string
		specs   []string
		want    []asgAutoDiscoveryConfig
		wantErr bool
	}{
		{
			name: "GoodSpecs",
			specs: []string{
				"asg:tag=tag,anothertag",
				"asg:tag=cooltag,anothertag",
				"asg:tag=label=value,anothertag",
				"asg:tag=my:label=value,my:otherlabel=othervalue",
			},
			want: []asgAutoDiscoveryConfig{
				{Tags: map[string]string{"tag": "", "anothertag": ""}},
				{Tags: map[string]string{"cooltag": "", "anothertag": ""}},
				{Tags: map[string]string{"label": "value", "anothertag": ""}},
				{Tags: map[string]string{"my:label": "value", "my:otherlabel": "othervalue"}},
			},
		},
		{
			name:    "MissingASGType",
			specs:   []string{"tag=tag,anothertag"},
			wantErr: true,
		},
		{
			name:    "WrongType",
			specs:   []string{"mig:tag=tag,anothertag"},
			wantErr: true,
		},
		{
			name:    "KeyMissingValue",
			specs:   []string{"asg:tag="},
			wantErr: true,
		},
		{
			name:    "ValueMissingKey",
			specs:   []string{"asg:=tag"},
			wantErr: true,
		},
		{
			name:    "KeyMissingSeparator",
			specs:   []string{"asg:tag"},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			do := cloudprovider.NodeGroupDiscoveryOptions{NodeGroupAutoDiscoverySpecs: tc.specs}
			got, err := parseASGAutoDiscoverySpecs(do)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.True(t, assert.ObjectsAreEqualValues(tc.want, got), "\ngot: %#v\nwant: %#v", got, tc.want)
		})
	}
}
