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
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"k8s.io/apimachinery/pkg/util/clock"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/tools/cache"
	klog "k8s.io/klog/v2"
)

const (
	launchConfigurationCachedTTL = time.Minute * 20
	cacheMinTTL                  = 120
	cacheMaxTTL                  = 600
)

// autoScaling is the interface represents a specific aspect of the auto-scaling service provided by AWS SDK for use in CA
type autoScaling interface {
	DescribeAutoScalingGroupsPages(input *autoscaling.DescribeAutoScalingGroupsInput, fn func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool) error
	DescribeLaunchConfigurations(*autoscaling.DescribeLaunchConfigurationsInput) (*autoscaling.DescribeLaunchConfigurationsOutput, error)
	DescribeTagsPages(input *autoscaling.DescribeTagsInput, fn func(*autoscaling.DescribeTagsOutput, bool) bool) error
	SetDesiredCapacity(input *autoscaling.SetDesiredCapacityInput) (*autoscaling.SetDesiredCapacityOutput, error)
	TerminateInstanceInAutoScalingGroup(input *autoscaling.TerminateInstanceInAutoScalingGroupInput) (*autoscaling.TerminateInstanceInAutoScalingGroupOutput, error)
}

// autoScalingWrapper provides several utility methods over the auto-scaling service provided by AWS SDK
type autoScalingWrapper struct {
	autoScaling
	launchConfigurationInstanceTypeCache *expirationStore
}

// expirationStore cache the launch configuration with their instance type.
// The store expires its keys based on a TTL. This TTL can have a jitter applied to it.
// This allows to get a better repartition of the AWS queries.
type expirationStore struct {
	cache.Store
	jitterClock *jitterClock
}

type instanceTypeCachedObject struct {
	name         string
	instanceType string
}

type jitterClock struct {
	clock.Clock

	jitter bool
	sync.RWMutex
}

func newLaunchConfigurationInstanceTypeCache() *expirationStore {
	jc := &jitterClock{}
	return &expirationStore{
		cache.NewExpirationStore(func(obj interface{}) (s string, e error) {
			return obj.(instanceTypeCachedObject).name, nil
		}, &cache.TTLPolicy{
			TTL:   launchConfigurationCachedTTL,
			Clock: jc,
		}),
		jc,
	}
}

func (c *jitterClock) Since(ts time.Time) time.Duration {
	since := time.Since(ts)
	c.RLock()
	defer c.RUnlock()
	if c.jitter {
		return since + (time.Second * time.Duration(rand.IntnRange(cacheMinTTL, cacheMaxTTL)))
	}
	return since
}

func (m autoScalingWrapper) getInstanceTypeByLCNames(launchConfigToQuery []*string) ([]*autoscaling.LaunchConfiguration, error) {
	var launchConfigurations []*autoscaling.LaunchConfiguration

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
		launchConfigurations = append(launchConfigurations, r.LaunchConfigurations...)
		for _, lc := range r.LaunchConfigurations {
			_ = m.launchConfigurationInstanceTypeCache.Add(instanceTypeCachedObject{
				name:         *lc.LaunchConfigurationName,
				instanceType: *lc.InstanceType,
			})
		}
	}
	return launchConfigurations, nil
}

func (m autoScalingWrapper) getInstanceTypeByLCName(name string) (string, error) {
	if obj, found, _ := m.launchConfigurationInstanceTypeCache.GetByKey(name); found {
		return obj.(instanceTypeCachedObject).instanceType, nil
	}

	launchConfigs, err := m.getInstanceTypeByLCNames([]*string{aws.String(name)})
	if err != nil {
		klog.Errorf("Failed to query the launch configuration %s to get the instance type: %v", name, err)
		return "", err
	}
	if len(launchConfigs) < 1 || launchConfigs[0].InstanceType == nil {
		return "", fmt.Errorf("unable to get first LaunchConfiguration for %s", name)
	}
	return *launchConfigs[0].InstanceType, nil
}

func (m *autoScalingWrapper) getAutoscalingGroupsByNames(names []string) ([]*autoscaling.Group, error) {
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

func (m autoScalingWrapper) populateLaunchConfigurationInstanceTypeCache(autoscalingGroups []*autoscaling.Group) error {
	var launchConfigToQuery []*string

	m.launchConfigurationInstanceTypeCache.jitterClock.Lock()
	m.launchConfigurationInstanceTypeCache.jitterClock.jitter = true
	m.launchConfigurationInstanceTypeCache.jitterClock.Unlock()
	for _, asg := range autoscalingGroups {
		if asg == nil {
			continue
		}
		if asg.LaunchConfigurationName == nil {
			continue
		}
		_, found, _ := m.launchConfigurationInstanceTypeCache.GetByKey(*asg.LaunchConfigurationName)
		if found {
			continue
		}
		launchConfigToQuery = append(launchConfigToQuery, asg.LaunchConfigurationName)
	}
	m.launchConfigurationInstanceTypeCache.jitterClock.Lock()
	m.launchConfigurationInstanceTypeCache.jitterClock.jitter = false
	m.launchConfigurationInstanceTypeCache.jitterClock.Unlock()

	// List expire old entries
	_ = m.launchConfigurationInstanceTypeCache.List()

	if len(launchConfigToQuery) == 0 {
		klog.V(4).Infof("%d launch configurations already in cache", len(autoscalingGroups))
		return nil
	}
	klog.V(4).Infof("%d launch configurations to query", len(launchConfigToQuery))

	_, err := m.getInstanceTypeByLCNames(launchConfigToQuery)
	if err != nil {
		klog.Errorf("Failed to query %d launch configurations", len(launchConfigToQuery))
		return err
	}

	klog.V(4).Infof("Successfully query %d launch configurations", len(launchConfigToQuery))
	return nil
}

func (m *autoScalingWrapper) getAutoscalingGroupNamesByTags(kvs map[string]string) ([]string, error) {
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
