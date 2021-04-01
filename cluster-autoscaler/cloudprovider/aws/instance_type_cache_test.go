package aws

import (
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/clock"
	"k8s.io/client-go/tools/cache"
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

func TestPopulate(t *testing.T) {
	asgName, ltName, ltVersion, instanceType := "testasg", "launcher", "1", "t2.large"
	ltSpec := autoscaling.LaunchTemplateSpecification{
		LaunchTemplateName: aws.String(ltName),
		Version:            aws.String(ltVersion),
	}

	a := &autoScalingMock{}
	e := &ec2Mock{}
	a.On("DescribeLaunchConfigurations", &autoscaling.DescribeLaunchConfigurationsInput{
		LaunchConfigurationNames: []*string{aws.String(ltName)},
		MaxRecords:               aws.Int64(50),
	}).Return(&autoscaling.DescribeLaunchConfigurationsOutput{
		LaunchConfigurations: []*autoscaling.LaunchConfiguration{
			{
				LaunchConfigurationName: aws.String(ltName),
				InstanceType:            aws.String(instanceType),
			},
		},
	})
	e.On("DescribeLaunchTemplateVersions", &ec2.DescribeLaunchTemplateVersionsInput{
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

	cases := []struct {
		name                    string
		launchConfigurationName *string
		launchTemplate          *autoscaling.LaunchTemplateSpecification
		mixedInstancesPolicy    *autoscaling.MixedInstancesPolicy
	}{
		{
			"AsgWithLaunchConfiguration",
			aws.String(ltName),
			nil,
			nil,
		},
		{
			"AsgWithLaunchTemplate",
			nil,
			&ltSpec,
			nil,
		},
		{
			"AsgWithLaunchTemplateMixedInstancePolicyOverride",
			nil,
			nil,
			&autoscaling.MixedInstancesPolicy{
				LaunchTemplate: &autoscaling.LaunchTemplate{
					Overrides: []*autoscaling.LaunchTemplateOverrides{
						{
							InstanceType: aws.String(instanceType),
						},
					},
				},
			},
		},
		{
			"AsgWithLaunchTemplateMixedInstancePolicyNoOverride",
			nil,
			nil,
			&autoscaling.MixedInstancesPolicy{
				LaunchTemplate: &autoscaling.LaunchTemplate{
					LaunchTemplateSpecification: &ltSpec,
					Overrides:                   []*autoscaling.LaunchTemplateOverrides{},
				},
			},
		},
	}

	for _, tc := range cases {
		m := newAsgInstanceTypeCache(&awsWrapper{a, e})

		err := m.populate([]*autoscaling.Group{
			{
				AutoScalingGroupName:    aws.String(asgName),
				LaunchConfigurationName: tc.launchConfigurationName,
				LaunchTemplate:          tc.launchTemplate,
				MixedInstancesPolicy:    tc.mixedInstancesPolicy,
			},
		})
		assert.NoError(t, err)

		result, found, err := m.GetByKey(asgName)
		assert.NoErrorf(t, err, "%s had error %v", tc.name, err)
		assert.Truef(t, found, "%s did not find asg", tc.name)

		foundInstanceType := result.(instanceTypeCachedObject).instanceType
		assert.Equalf(t, foundInstanceType, instanceType, "%s had %s, expected %s", tc.name, foundInstanceType, instanceType)
	}
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

	fakeClock := clock.NewFakeClock(time.Unix(0, 0))
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
	m := newAsgInstanceTypeCacheWithClock(&awsWrapper{a, e}, fakeClock, fakeStore)

	for i := 0; i < 2; i++ {
		err := m.populate([]*autoscaling.Group{
			{
				AutoScalingGroupName: aws.String(asgName),
				LaunchTemplate: &autoscaling.LaunchTemplateSpecification{
					LaunchTemplateName: aws.String(ltName),
					Version:            ltVersions[i],
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
