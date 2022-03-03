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
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/tools/cache"
	"k8s.io/utils/clock"
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

func (es instanceTypeExpirationStore) populate(autoscalingGroups map[AwsRef]*asg) error {
	asgsToQuery := []*asg{}

	if c, ok := es.jitterClock.(*jitterClock); ok {
		c.Lock()
		c.jitter = true
		c.Unlock()
	}

	for _, asg := range autoscalingGroups {
		if asg == nil {
			continue
		}
		_, found, _ := es.GetByKey(asg.AwsRef.Name)
		if found {
			continue
		}
		asgsToQuery = append(asgsToQuery, asg)
	}

	if c, ok := es.jitterClock.(*jitterClock); ok {
		c.Lock()
		c.jitter = false
		c.Unlock()
	}

	// List expires old entries
	_ = es.List()

	instanceTypesByAsg, err := es.awsService.getInstanceTypesForAsgs(asgsToQuery)
	if err != nil {
		return err
	}

	for asgName, instanceType := range instanceTypesByAsg {
		es.Add(instanceTypeCachedObject{
			name:         asgName,
			instanceType: instanceType,
		})
	}
	return nil
}
