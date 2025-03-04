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
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/aws"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/service/autoscaling"
)

func TestBuildAsg(t *testing.T) {
	asgCache := &asgCache{}

	asg, err := asgCache.buildAsgFromSpec("1:5:test-asg")
	assert.NoError(t, err)
	assert.Equal(t, asg.minSize, 1)
	assert.Equal(t, asg.maxSize, 5)
	assert.Equal(t, asg.Name, "test-asg")

	_, err = asgCache.buildAsgFromSpec("a")
	assert.Error(t, err)
	_, err = asgCache.buildAsgFromSpec("a:b:c")
	assert.Error(t, err)
	_, err = asgCache.buildAsgFromSpec("1:")
	assert.Error(t, err)
	_, err = asgCache.buildAsgFromSpec("1:2:")
	assert.Error(t, err)
}

func TestBuildAsgTags(t *testing.T) {

	cases := []struct {
		name    string
		specs   []string
		want    []map[string]string
		wantErr bool
	}{
		{
			name: "MultiSpecs",
			specs: []string{
				"asg:tag=tag,anothertag",
				"asg:tag=cooltag,anothertag",
				"asg:tag=label=value,anothertag",
				"asg:tag=my:label=value,my:otherlabel=othervalue",
			},
			want: []map[string]string{
				{"tag": "", "anothertag": ""},
				{"cooltag": "", "anothertag": ""},
				{"label": "value", "anothertag": ""},
				{"my:label": "value", "my:otherlabel": "othervalue"},
			},
		},
		{
			name: "SingleSpec",
			specs: []string{
				"asg:tag=mylabel=value,myotherlabel=othervalue",
			},
			want: []map[string]string{
				{"mylabel": "value", "myotherlabel": "othervalue"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			do := cloudprovider.NodeGroupDiscoveryOptions{NodeGroupAutoDiscoverySpecs: tc.specs}
			specs, _ := parseASGAutoDiscoverySpecs(do)
			asgCache := &asgCache{
				asgAutoDiscoverySpecs: specs,
			}
			got := asgCache.buildAsgTags()
			assert.True(t, assert.ObjectsAreEqualValues(tc.want, got), "\ngot: %#v\nwant: %#v", got, tc.want)
		})
	}
}

func validateAsg(t *testing.T, asg *asg, name string, minSize int, maxSize int) {
	assert.Equal(t, name, asg.Name)
	assert.Equal(t, minSize, asg.minSize)
	assert.Equal(t, maxSize, asg.maxSize)
}

func TestCreatePlaceholders(t *testing.T) {
	registeredAsgName := aws.String("test-asg")
	registeredAsgRef := AwsRef{Name: *registeredAsgName}

	cases := []struct {
		name                string
		desiredCapacity     *int64
		activities          []*autoscaling.Activity
		groupLastUpdateTime time.Time
		describeErr         error
		asgToCheck          *string
	}{
		{
			name:            "add placeholders successful",
			desiredCapacity: aws.Int64(10),
		},
		{
			name:            "no placeholders needed",
			desiredCapacity: aws.Int64(0),
		},
		{
			name:            "DescribeScalingActivities failed",
			desiredCapacity: aws.Int64(1),
			describeErr:     errors.New("timeout"),
		},
		{
			name:            "early abort if AWS scaling up fails",
			desiredCapacity: aws.Int64(1),
			activities: []*autoscaling.Activity{
				{
					StatusCode: aws.String("Failed"),
					StartTime:  aws.Time(time.Unix(10, 0)),
				},
			},
			groupLastUpdateTime: time.Unix(9, 0),
		},
		{
			name:            "AWS scaling failed event before CA scale_up",
			desiredCapacity: aws.Int64(1),
			activities: []*autoscaling.Activity{
				{
					StatusCode: aws.String("Failed"),
					StartTime:  aws.Time(time.Unix(9, 0)),
				},
			},
			groupLastUpdateTime: time.Unix(10, 0),
		},
		{
			name:            "asg not registered",
			desiredCapacity: aws.Int64(10),
			activities: []*autoscaling.Activity{
				{
					StatusCode: aws.String("Failed"),
					StartTime:  aws.Time(time.Unix(10, 0)),
				},
			},
			groupLastUpdateTime: time.Unix(9, 0),
			asgToCheck:          aws.String("unregisteredAsgName"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			shouldCallDescribeScalingActivities := true
			if *tc.desiredCapacity == int64(0) {
				shouldCallDescribeScalingActivities = false
			}

			asgName := registeredAsgName
			if tc.asgToCheck != nil {
				asgName = tc.asgToCheck
			}

			a := &autoScalingMock{}
			if shouldCallDescribeScalingActivities {
				a.On("DescribeScalingActivities", &autoscaling.DescribeScalingActivitiesInput{
					AutoScalingGroupName: asgName,
				}).Return(
					&autoscaling.DescribeScalingActivitiesOutput{Activities: tc.activities},
					tc.describeErr,
				).Once()
			}

			asgCache := &asgCache{
				awsService: &awsWrapper{
					autoScalingI: a,
					ec2I:         nil,
				},
				registeredAsgs: map[AwsRef]*asg{
					registeredAsgRef: {
						AwsRef:         registeredAsgRef,
						lastUpdateTime: tc.groupLastUpdateTime,
					},
				},
			}

			groups := []*autoscaling.Group{
				{
					AutoScalingGroupName: asgName,
					AvailabilityZones:    []*string{aws.String("westeros-1a")},
					DesiredCapacity:      tc.desiredCapacity,
					Instances:            []*autoscaling.Instance{},
				},
			}
			asgCache.createPlaceholdersForDesiredNonStartedInstances(groups)
			assert.Equal(t, int64(len(groups[0].Instances)), *tc.desiredCapacity)
			if tc.activities != nil && *tc.activities[0].StatusCode == "Failed" && tc.activities[0].StartTime.After(tc.groupLastUpdateTime) && asgName == registeredAsgName {
				assert.Equal(t, *groups[0].Instances[0].HealthStatus, placeholderUnfulfillableStatus)
			} else if len(groups[0].Instances) > 0 {
				assert.Equal(t, *groups[0].Instances[0].HealthStatus, "")
			}
			a.AssertExpectations(t)
		})
	}
}
