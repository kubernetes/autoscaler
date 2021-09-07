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
	"testing"

	"github.com/stretchr/testify/assert"

	sdkaws "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
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

func TestUnregister(t *testing.T) {
	asgCache := new(asgCache)
	asgCache.registeredAsgs = []*asg{
		{AwsRef: AwsRef{Name: "test-asg-1"}},
		{AwsRef: AwsRef{Name: "test-asg-2"}},
		{AwsRef: AwsRef{Name: "test-asg-1"}},
		{AwsRef: AwsRef{Name: "test-asg-3"}},
	}
	t.Run("unregister non-existing asg", func(t *testing.T) {
		n := len(asgCache.registeredAsgs)
		assert.Nil(t, asgCache.unregister(&asg{AwsRef: AwsRef{Name: "test-asg-404"}}))
		assert.Len(t, asgCache.registeredAsgs, n)
		assert.Nil(t, asgCache.unregister(&asg{AwsRef: AwsRef{Name: "non-existing"}}))
		assert.Len(t, asgCache.registeredAsgs, n)
	})
	t.Run("unregister existing asg", func(t *testing.T) {
		asg1 := &asg{AwsRef: AwsRef{Name: "test-asg-1"}}
		assert.Equal(t, asg1, asgCache.unregister(asg1))
		assert.Len(t, asgCache.registeredAsgs, 2)
		asg3 := &asg{AwsRef: AwsRef{Name: "test-asg-3"}}
		assert.Equal(t, asg3, asgCache.unregister(asg3))
		assert.Len(t, asgCache.registeredAsgs, 1)
		asg2 := &asg{AwsRef: AwsRef{Name: "test-asg-2"}}
		assert.Equal(t, asg2, asgCache.unregister(asg2))
		assert.Empty(t, asgCache.registeredAsgs)
	})

}

func validateAsg(t *testing.T, asg *asg, name string, minSize int, maxSize int) {
	assert.Equal(t, name, asg.Name)
	assert.Equal(t, minSize, asg.minSize)
	assert.Equal(t, maxSize, asg.maxSize)
}

func TestBuildLaunchTemplateFromSpec(t *testing.T) {
	assert := assert.New(t)

	units := []struct {
		name string
		in   *autoscaling.LaunchTemplateSpecification
		exp  *launchTemplate
	}{
		{
			name: "non-default, specified version",
			in: &autoscaling.LaunchTemplateSpecification{
				LaunchTemplateName: sdkaws.String("foo"),
				Version:            sdkaws.String("1"),
			},
			exp: &launchTemplate{
				name:    "foo",
				version: "1",
			},
		},
		{
			name: "non-default, specified $Latest",
			in: &autoscaling.LaunchTemplateSpecification{
				LaunchTemplateName: sdkaws.String("foo"),
				Version:            sdkaws.String("$Latest"),
			},
			exp: &launchTemplate{
				name:    "foo",
				version: "$Latest",
			},
		},
		{
			name: "specified $Default",
			in: &autoscaling.LaunchTemplateSpecification{
				LaunchTemplateName: sdkaws.String("foo"),
				Version:            sdkaws.String("$Default"),
			},
			exp: &launchTemplate{
				name:    "foo",
				version: "$Default",
			},
		},
		{
			name: "no version specified",
			in: &autoscaling.LaunchTemplateSpecification{
				LaunchTemplateName: sdkaws.String("foo"),
				Version:            nil,
			},
			exp: &launchTemplate{
				name:    "foo",
				version: "$Default",
			},
		},
	}

	cache := &asgCache{}
	for _, unit := range units {
		got := cache.buildLaunchTemplateFromSpec(unit.in)
		assert.Equal(unit.exp, got)
	}
}
