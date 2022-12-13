/*
Copyright 2021 The Kubernetes Authors.

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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/aws"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/service/ec2"
	"k8s.io/client-go/tools/cache"
	test_clock "k8s.io/utils/clock/testing"
)

func TestInstanceTypeCache(t *testing.T) {
	c := newAsgInstanceTypeCache(nil)
	err := c.Add(instanceTypeCachedObject{
		name:         "123",
		instanceType: "t2.medium",
	})
	require.NoError(t, err)
	obj, ok, err := c.GetByKey("123")
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "t2.medium", obj.(instanceTypeCachedObject).instanceType)
}

func TestLTVersionChange(t *testing.T) {
	asgName, ltName := "testasg", "launcher"
	ltVersions := []*string{aws.String("1"), aws.String("2")}
	instanceTypes := []*string{aws.String("t2.large"), aws.String("m4.xlarge")}

	a := &autoScalingMock{}
	e := &ec2Mock{}

	for i := 0; i < 2; i++ {
		e.On("DescribeLaunchTemplateVersions", &ec2.DescribeLaunchTemplateVersionsInput{
			LaunchTemplateName: aws.String(ltName),
			Versions:           []*string{ltVersions[i]},
		}).Return(&ec2.DescribeLaunchTemplateVersionsOutput{
			LaunchTemplateVersions: []*ec2.LaunchTemplateVersion{
				{
					LaunchTemplateData: &ec2.ResponseLaunchTemplateData{
						InstanceType: instanceTypes[i],
					},
				},
			},
		})
	}

	fakeClock := test_clock.NewFakeClock(time.Unix(0, 0))
	fakeStore := cache.NewFakeExpirationStore(
		func(obj interface{}) (s string, e error) {
			return obj.(instanceTypeCachedObject).name, nil
		},
		nil,
		&cache.TTLPolicy{
			TTL:   asgInstanceTypeCacheTTL,
			Clock: fakeClock,
		},
		fakeClock,
	)
	m := newAsgInstanceTypeCacheWithClock(&awsWrapper{a, e, nil}, fakeClock, fakeStore)

	for i := 0; i < 2; i++ {
		asgRef := AwsRef{Name: asgName}
		err := m.populate(map[AwsRef]*asg{
			asgRef: {
				AwsRef: asgRef,
				LaunchTemplate: &launchTemplate{
					name:    ltName,
					version: aws.StringValue(ltVersions[i]),
				},
			},
		})
		assert.NoError(t, err)

		result, found, err := m.GetByKey(asgName)
		assert.NoError(t, err)
		assert.Truef(t, found, "%s did not find asg (iteration %d)", asgName, i)

		foundInstanceType := result.(instanceTypeCachedObject).instanceType
		assert.Equalf(t, foundInstanceType, *instanceTypes[i], "%s had %s, expected %s (iteration %d)", asgName, foundInstanceType, *instanceTypes[i], i)

		// Expire the first instance
		fakeClock.SetTime(time.Now().Add(asgInstanceTypeCacheTTL + 10*time.Minute))
	}
}
