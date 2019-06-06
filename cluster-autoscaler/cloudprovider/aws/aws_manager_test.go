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
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	kubeletapis "k8s.io/kubernetes/pkg/kubelet/apis"
	provider_aws "k8s.io/legacy-cloud-providers/aws"
)

// resetAWSRegion resets AWS_REGION environment variable key to its pre-test
// value, but only if it was originally present among environment variables.
func resetAWSRegion(value string, present bool) {
	os.Unsetenv("AWS_REGION")
	if present {
		os.Setenv("AWS_REGION", value)
	}
}

// TestGetRegion ensures correct source supplies AWS Region.
func TestGetRegion(t *testing.T) {
	key := "AWS_REGION"
	defer resetAWSRegion(os.LookupEnv(key))
	// Ensure environment variable retains precedence.
	expected1 := "the-shire-1"
	os.Setenv(key, expected1)
	assert.Equal(t, expected1, getRegion())
	// Ensure without environment variable, EC2 Metadata used... and it merely
	// chops the last character off the Availability Zone.
	expected2 := "mordor-2"
	expected2a := expected2 + "a"
	os.Unsetenv(key)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(expected2a))
	}))
	cfg := aws.NewConfig().WithEndpoint(server.URL)
	assert.Equal(t, expected2, getRegion(cfg))
}

func TestBuildGenericLabels(t *testing.T) {
	labels := buildGenericLabels(&asgTemplate{
		InstanceType: &instanceType{
			InstanceType: "c4.large",
			VCPU:         2,
			MemoryMb:     3840,
		},
		Region: "us-east-1",
	}, "sillyname")
	assert.Equal(t, "us-east-1", labels[apiv1.LabelZoneRegion])
	assert.Equal(t, "sillyname", labels[apiv1.LabelHostname])
	assert.Equal(t, "c4.large", labels[apiv1.LabelInstanceType])
	assert.Equal(t, cloudprovider.DefaultArch, labels[kubeletapis.LabelArch])
	assert.Equal(t, cloudprovider.DefaultOS, labels[kubeletapis.LabelOS])
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
	}

	labels := extractAllocatableResourcesFromAsg(tags)

	assert.Equal(t, resource.NewMilliQuantity(100, resource.DecimalSI).String(), labels["cpu"].String())
	expectedMemory := resource.MustParse("100M")
	assert.Equal(t, (&expectedMemory).String(), labels["memory"].String())
	expectedEphemeralStorage := resource.MustParse("20G")
	assert.Equal(t, (&expectedEphemeralStorage).String(), labels["ephemeral-storage"].String())
}

func TestExtractLabelsFromAsg(t *testing.T) {
	tags := []*autoscaling.TagDescription{
		{
			Key:   aws.String("k8s.io/cluster-autoscaler/node-template/label/foo"),
			Value: aws.String("bar"),
		},
		{
			Key:   aws.String("bar"),
			Value: aws.String("baz"),
		},
	}

	labels := extractLabelsFromAsg(tags)

	assert.Equal(t, 1, len(labels))
	assert.Equal(t, "bar", labels["foo"])
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

	s := &AutoScalingMock{}
	s.On("DescribeAutoScalingGroups", &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{aws.String(groupname)},
		MaxRecords:            aws.Int64(1),
	}).Return(&autoscaling.DescribeAutoScalingGroupsOutput{
		AutoScalingGroups: []*autoscaling.Group{
			{AutoScalingGroupName: aws.String(groupname)},
		},
	})

	s.On("DescribeAutoScalingGroupsPages",
		&autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: aws.StringSlice([]string{groupname}),
			MaxRecords:            aws.Int64(maxRecordsReturnedByAPI),
		},
		mock.AnythingOfType("func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool"),
	).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool)
		fn(&autoscaling.DescribeAutoScalingGroupsOutput{
			AutoScalingGroups: []*autoscaling.Group{
				{AutoScalingGroupName: aws.String(groupname)},
			}}, false)
	}).Return(nil)

	do := cloudprovider.NodeGroupDiscoveryOptions{
		// Register the same node group twice with different max nodes.
		// The intention is to test that the asgs.Register method will update
		// the node group instead of registering it twice.
		NodeGroupSpecs: []string{
			fmt.Sprintf("%d:%d:%s", min, max, groupname),
			fmt.Sprintf("%d:%d:%s", min, max-1, groupname),
		},
	}
	// #1449 Without AWS_REGION getRegion() lookup runs till timeout during tests.
	defer resetAWSRegion(os.LookupEnv("AWS_REGION"))
	os.Setenv("AWS_REGION", "fanghorn")
	// fetchExplicitASGs is called at manager creation time.
	m, err := createAWSManagerInternal(nil, do, &autoScalingWrapper{s, map[string]string{}}, nil)
	assert.NoError(t, err)

	asgs := m.asgCache.Get()
	assert.Equal(t, 1, len(asgs))
	validateAsg(t, asgs[0], groupname, min, max)
}

func TestBuildInstanceType(t *testing.T) {
	ltName, ltVersion, instanceType := "launcher", "1", "t2.large"

	s := &EC2Mock{}
	s.On("DescribeLaunchTemplateVersions", &ec2.DescribeLaunchTemplateVersionsInput{
		LaunchTemplateName: aws.String(ltName),
		Versions:           []*string{aws.String(ltVersion)},
	}).Return(&ec2.DescribeLaunchTemplateVersionsOutput{
		LaunchTemplateVersions: []*ec2.LaunchTemplateVersion{
			{
				LaunchTemplateData: &ec2.ResponseLaunchTemplateData{
					InstanceType: aws.String(instanceType),
				},
			},
		},
	})

	// #1449 Without AWS_REGION getRegion() lookup runs till timeout during tests.
	defer resetAWSRegion(os.LookupEnv("AWS_REGION"))
	os.Setenv("AWS_REGION", "fanghorn")
	m, err := createAWSManagerInternal(nil, cloudprovider.NodeGroupDiscoveryOptions{}, nil, &ec2Wrapper{s})
	assert.NoError(t, err)

	asg := asg{
		LaunchTemplateName:    ltName,
		LaunchTemplateVersion: ltVersion,
	}

	builtInstanceType, err := m.buildInstanceType(&asg)

	assert.NoError(t, err)
	assert.Equal(t, instanceType, builtInstanceType)
}

func TestGetASGTemplate(t *testing.T) {
	const (
		knownInstanceType = "t3.micro"
		region            = "us-east-1"
		az                = region + "a"
		ltName            = "launcher"
		ltVersion         = "1"
	)

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
			s := &EC2Mock{}
			s.On("DescribeLaunchTemplateVersions", &ec2.DescribeLaunchTemplateVersionsInput{
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

			// #1449 Without AWS_REGION getRegion() lookup runs till timeout during tests.
			defer resetAWSRegion(os.LookupEnv("AWS_REGION"))
			os.Setenv("AWS_REGION", "fanghorn")
			m, err := createAWSManagerInternal(nil, cloudprovider.NodeGroupDiscoveryOptions{}, nil, &ec2Wrapper{s})
			assert.NoError(t, err)

			asg := &asg{
				AwsRef:                AwsRef{Name: "sample"},
				AvailabilityZones:     test.availabilityZones,
				LaunchTemplateName:    ltName,
				LaunchTemplateVersion: ltVersion,
				Tags:                  tags,
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

	s := &AutoScalingMock{}
	// Lookup groups associated with tags
	expectedTagsInput := &autoscaling.DescribeTagsInput{
		Filters: []*autoscaling.Filter{
			{Name: aws.String("key"), Values: aws.StringSlice([]string{tags[0]})},
			{Name: aws.String("key"), Values: aws.StringSlice([]string{tags[1]})},
		},
		MaxRecords: aws.Int64(maxRecordsReturnedByAPI),
	}
	// Use MatchedBy pattern to avoid list order issue https://github.com/kubernetes/autoscaler/issues/1346
	s.On("DescribeTagsPages", mock.MatchedBy(tagsMatcher(expectedTagsInput)),
		mock.AnythingOfType("func(*autoscaling.DescribeTagsOutput, bool) bool"),
	).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(*autoscaling.DescribeTagsOutput, bool) bool)
		fn(&autoscaling.DescribeTagsOutput{
			Tags: []*autoscaling.TagDescription{
				{ResourceId: aws.String(groupname)},
				{ResourceId: aws.String(groupname)},
			}}, false)
	}).Return(nil).Once()

	// Describe the group to register it, then again to generate the instance
	// cache.
	s.On("DescribeAutoScalingGroupsPages",
		&autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: aws.StringSlice([]string{groupname}),
			MaxRecords:            aws.Int64(maxRecordsReturnedByAPI),
		},
		mock.AnythingOfType("func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool"),
	).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool)
		fn(&autoscaling.DescribeAutoScalingGroupsOutput{
			AutoScalingGroups: []*autoscaling.Group{{
				AutoScalingGroupName: aws.String(groupname),
				MinSize:              aws.Int64(int64(min)),
				MaxSize:              aws.Int64(int64(max)),
			}}}, false)
	}).Return(nil).Twice()

	do := cloudprovider.NodeGroupDiscoveryOptions{
		NodeGroupAutoDiscoverySpecs: []string{fmt.Sprintf("asg:tag=%s", strings.Join(tags, ","))},
	}

	// #1449 Without AWS_REGION getRegion() lookup runs till timeout during tests.
	defer resetAWSRegion(os.LookupEnv("AWS_REGION"))
	os.Setenv("AWS_REGION", "fanghorn")
	// fetchAutoASGs is called at manager creation time, via forceRefresh
	m, err := createAWSManagerInternal(nil, do, &autoScalingWrapper{s, map[string]string{}}, nil)
	assert.NoError(t, err)

	asgs := m.asgCache.Get()
	assert.Equal(t, 1, len(asgs))
	validateAsg(t, asgs[0], groupname, min, max)

	// Simulate the previously discovered ASG disappearing
	s.On("DescribeTagsPages", mock.MatchedBy(tagsMatcher(expectedTagsInput)),
		mock.AnythingOfType("func(*autoscaling.DescribeTagsOutput, bool) bool"),
	).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(*autoscaling.DescribeTagsOutput, bool) bool)
		fn(&autoscaling.DescribeTagsOutput{Tags: []*autoscaling.TagDescription{}}, false)
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

func TestOverridesActiveConfig(t *testing.T) {
	tests := []struct {
		name string

		reader io.Reader
		aws    provider_aws.Services

		expectError        bool
		active             bool
		servicesOverridden []ServiceDescriptor
	}{
		{
			"No overrides",
			strings.NewReader(`
				[global]
				`),
			nil,
			false, false,
			[]ServiceDescriptor{},
		},
		{
			"Missing Service Name",
			strings.NewReader(`
				[global]
				[ServiceOverride "1"]
				Region=sregion
				URL=https://s3.foo.bar
				SigningRegion=sregion
				SigningMethod = sign
				`),
			nil,
			true, false,
			[]ServiceDescriptor{},
		},
		{
			"Missing Service Region",
			strings.NewReader(`
				[global]
				[ServiceOverride "1"]
				Service=s3
				URL=https://s3.foo.bar
				SigningRegion=sregion
				SigningMethod = sign
				`),
			nil,
			true, false,
			[]ServiceDescriptor{},
		},
		{
			"Missing URL",
			strings.NewReader(`
				[global]
				[ServiceOverride "1"]
				Service="s3"
				Region=sregion
				SigningRegion=sregion
				SigningMethod = sign
				`),
			nil,
			true, false,
			[]ServiceDescriptor{},
		},
		{
			"Missing Signing Region",
			strings.NewReader(`
				[global]
				[ServiceOverride "1"]
				Service=s3
				Region=sregion
				URL=https://s3.foo.bar
				SigningMethod = sign
				`),
			nil,
			true, false,
			[]ServiceDescriptor{},
		},
		{
			"Active Overrides",
			strings.NewReader(`
				[Global]
				[ServiceOverride "1"]
				Service = "s3      "
				Region = sregion
				URL = https://s3.foo.bar
				SigningRegion = sregion
				SigningMethod = v4
				`),
			nil,
			false, true,
			[]ServiceDescriptor{{name: "s3", region: "sregion", signingRegion: "sregion", signingMethod: "v4"}},
		},
		{
			"Multiple Overridden Services",
			strings.NewReader(`
				[Global]
				vpc = vpc-abc1234567
				[ServiceOverride "1"]
				Service=s3
				Region=sregion1
				URL=https://s3.foo.bar
				SigningRegion=sregion1
				SigningMethod = v4
				[ServiceOverride "2"]
				Service=ec2
				Region=sregion2
				URL=https://ec2.foo.bar
				SigningRegion=sregion2
				SigningMethod = v4
				`),
			nil,
			false, true,
			[]ServiceDescriptor{{name: "s3", region: "sregion1", signingRegion: "sregion1", signingMethod: "v4"},
				{name: "ec2", region: "sregion2", signingRegion: "sregion2", signingMethod: "v4"}},
		},
		{
			"Duplicate Services",
			strings.NewReader(`
				[Global]
				vpc = vpc-abc1234567
				[ServiceOverride "1"]
				Service=s3
				Region=sregion1
				URL=https://s3.foo.bar
				SigningRegion=sregion
				SigningMethod = sign
				[ServiceOverride "2"]
				Service=s3
				Region=sregion1
				URL=https://s3.foo.bar
				SigningRegion=sregion
				SigningMethod = sign
				`),
			nil,
			true, false,
			[]ServiceDescriptor{},
		},
		{
			"Multiple Overridden Services in Multiple regions",
			strings.NewReader(`
				[global]
				[ServiceOverride "1"]
			 	Service=s3
				Region=region1
				URL=https://s3.foo.bar
				SigningRegion=sregion1
				[ServiceOverride "2"]
				Service=ec2
				Region=region2
				URL=https://ec2.foo.bar
				SigningRegion=sregion
				SigningMethod = v4
				`),
			nil,
			false, true,
			[]ServiceDescriptor{{name: "s3", region: "region1", signingRegion: "sregion1", signingMethod: ""},
				{name: "ec2", region: "region2", signingRegion: "sregion", signingMethod: "v4"}},
		},
		{
			"Multiple regions, Same Service",
			strings.NewReader(`
				[global]
				[ServiceOverride "1"]
				Service=s3
				Region=region1
				URL=https://s3.foo.bar
				SigningRegion=sregion1
				SigningMethod = v3
				[ServiceOverride "2"]
				Service=s3
				Region=region2
				URL=https://s3.foo.bar
				SigningRegion=sregion1
				SigningMethod = v4
				SigningName = "name"
				`),
			nil,
			false, true,
			[]ServiceDescriptor{{name: "s3", region: "region1", signingRegion: "sregion1", signingMethod: "v3"},
				{name: "s3", region: "region2", signingRegion: "sregion1", signingMethod: "v4", signingName: "name"}},
		},
	}

	for _, test := range tests {
		t.Logf("Running test case %s", test.name)
		cfg, err := readAWSCloudConfig(test.reader)
		if err == nil {
			err = validateOverrides(cfg)
		}
		if test.expectError {
			if err == nil {
				t.Errorf("Should error for case %s (cfg=%v)", test.name, cfg)
			}
		} else {
			if err != nil {
				t.Errorf("Should succeed for case: %s, got %v", test.name, err)
			}

			if len(cfg.ServiceOverride) != len(test.servicesOverridden) {
				t.Errorf("Expected %d overridden services, received %d for case %s",
					len(test.servicesOverridden), len(cfg.ServiceOverride), test.name)
			} else {
				for _, sd := range test.servicesOverridden {
					var found *struct {
						Service       string
						Region        string
						URL           string
						SigningRegion string
						SigningMethod string
						SigningName   string
					}
					for _, v := range cfg.ServiceOverride {
						if v.Service == sd.name && v.Region == sd.region {
							found = v
							break
						}
					}
					if found == nil {
						t.Errorf("Missing override for service %s in case %s",
							sd.name, test.name)
					} else {
						if found.SigningRegion != sd.signingRegion {
							t.Errorf("Expected signing region '%s', received '%s' for case %s",
								sd.signingRegion, found.SigningRegion, test.name)
						}
						if found.SigningMethod != sd.signingMethod {
							t.Errorf("Expected signing method '%s', received '%s' for case %s",
								sd.signingMethod, found.SigningRegion, test.name)
						}
						targetName := fmt.Sprintf("https://%s.foo.bar", sd.name)
						if found.URL != targetName {
							t.Errorf("Expected Endpoint '%s', received '%s' for case %s",
								targetName, found.URL, test.name)
						}
						if found.SigningName != sd.signingName {
							t.Errorf("Expected signing name '%s', received '%s' for case %s",
								sd.signingName, found.SigningName, test.name)
						}

						fn := getResolver(cfg)
						ep1, e := fn(sd.name, sd.region, nil)
						if e != nil {
							t.Errorf("Expected a valid endpoint for %s in case %s",
								sd.name, test.name)
						} else {
							targetName := fmt.Sprintf("https://%s.foo.bar", sd.name)
							if ep1.URL != targetName {
								t.Errorf("Expected endpoint url: %s, received %s in case %s",
									targetName, ep1.URL, test.name)
							}
							if ep1.SigningRegion != sd.signingRegion {
								t.Errorf("Expected signing region '%s', received '%s' in case %s",
									sd.signingRegion, ep1.SigningRegion, test.name)
							}
							if ep1.SigningMethod != sd.signingMethod {
								t.Errorf("Expected signing method '%s', received '%s' in case %s",
									sd.signingMethod, ep1.SigningRegion, test.name)
							}
						}
					}
				}
			}
		}
	}
}

func tagsMatcher(expected *autoscaling.DescribeTagsInput) func(*autoscaling.DescribeTagsInput) bool {
	return func(actual *autoscaling.DescribeTagsInput) bool {
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
