package aws

import (
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
	asgInstanceTypeCacheTTL = time.Minute * 20
	cacheMinTTL             = 120
	cacheMaxTTL             = 600
)

// instanceTypeExpirationStore caches the canonical instance type for an ASG.
// The store expires its keys based on a TTL. This TTL can have a jitter applied to it.
// This allows to get a better repartition of the AWS queries.
type instanceTypeExpirationStore struct {
	cache.Store
	jitterClock clock.Clock
	awsService  *awsWrapper
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

func newAsgInstanceTypeCache(awsService *awsWrapper) *instanceTypeExpirationStore {
	jc := &jitterClock{}
	return newAsgInstanceTypeCacheWithClock(
		awsService,
		jc,
		cache.NewExpirationStore(func(obj interface{}) (s string, e error) {
			return obj.(instanceTypeCachedObject).name, nil
		}, &cache.TTLPolicy{
			TTL:   asgInstanceTypeCacheTTL,
			Clock: jc,
		}),
	)
}

func newAsgInstanceTypeCacheWithClock(awsService *awsWrapper, jc clock.Clock, store cache.Store) *instanceTypeExpirationStore {
	return &instanceTypeExpirationStore{
		store,
		jc,
		awsService,
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

func (m instanceTypeExpirationStore) populate(autoscalingGroups []*autoscaling.Group) error {
	launchConfigsToQuery := map[string]*string{}
	launchTemplatesToQuery := map[string]*autoscaling.LaunchTemplateSpecification{}

	if c, ok := m.jitterClock.(*jitterClock); ok {
		c.Lock()
		c.jitter = true
		c.Unlock()
	}

	for _, asg := range autoscalingGroups {
		name := aws.StringValue(asg.AutoScalingGroupName)

		if asg == nil {
			continue
		}
		_, found, _ := m.GetByKey(name)
		if found {
			continue
		}

		if asg.LaunchConfigurationName != nil {
			launchConfigsToQuery[name] = asg.LaunchConfigurationName
		} else if asg.LaunchTemplate != nil {
			launchTemplatesToQuery[name] = asg.LaunchTemplate
		} else if asg.MixedInstancesPolicy != nil {
			if len(asg.MixedInstancesPolicy.LaunchTemplate.Overrides) > 0 {
				m.Add(instanceTypeCachedObject{
					name:         name,
					instanceType: aws.StringValue(asg.MixedInstancesPolicy.LaunchTemplate.Overrides[0].InstanceType),
				})
			} else {
				launchTemplatesToQuery[name] = asg.MixedInstancesPolicy.LaunchTemplate.LaunchTemplateSpecification
			}
		}
	}

	if c, ok := m.jitterClock.(*jitterClock); ok {
		c.Lock()
		c.jitter = false
		c.Unlock()
	}

	// List expires old entries
	_ = m.List()

	klog.V(4).Infof("%d launch configurations to query", len(launchConfigsToQuery))
	klog.V(4).Infof("%d launch templates to query", len(launchTemplatesToQuery))

	// Query these all at once to minimize AWS API calls
	launchConfigNames := make([]*string, 0, len(launchConfigsToQuery))
	for _, cfgName := range launchConfigsToQuery {
		launchConfigNames = append(launchConfigNames, cfgName)
	}
	launchConfigs, err := m.awsService.getInstanceTypeByLaunchConfigNames(launchConfigNames)
	if err != nil {
		klog.Errorf("Failed to query %d launch configurations", len(launchConfigsToQuery))
		return err
	}

	for asgName, cfgName := range launchConfigsToQuery {
		m.Add(instanceTypeCachedObject{
			name:         asgName,
			instanceType: launchConfigs[*cfgName],
		})
	}
	klog.V(4).Infof("Successfully query %d launch configurations", len(launchConfigsToQuery))

	// Have to query LaunchTemplates one-at-a-time, since there's no way to query <lt, version> pairs in bulk
	for asgName, launchTemplateSpec := range launchTemplatesToQuery {
		launchTemplate := buildLaunchTemplateFromSpec(launchTemplateSpec)
		instanceType, err := m.awsService.getInstanceTypeByLaunchTemplate(launchTemplate)
		if err != nil {
			klog.Error("Failed to query launch tempate %s", launchTemplate.name)
			continue
		}
		m.Add(instanceTypeCachedObject{
			name:         asgName,
			instanceType: instanceType,
		})
	}
	klog.V(4).Infof("Successfully query %d launch templates", len(launchTemplatesToQuery))

	return nil
}
