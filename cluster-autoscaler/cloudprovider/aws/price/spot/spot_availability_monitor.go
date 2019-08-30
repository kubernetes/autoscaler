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

package spot

import (
	"sync"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/api"
	"k8s.io/klog"
)

// AsgAvailabilityChecker provides an interface to check for ASG availability
type AsgAvailabilityChecker interface {
	AsgAvailability(name, iamInstanceProfile, availabilityZone, instanceType string) bool
}

var _ AsgAvailabilityChecker = &spotAvailabilityMonitor{}

// NewSpotAvailabilityMonitor returns an instance of the spot ASG availability monitor
func NewSpotAvailabilityMonitor(requestLister api.AwsEC2SpotRequestManager, checkInterval, exclusionPeriod time.Duration) *spotAvailabilityMonitor {
	return &spotAvailabilityMonitor{
		requestService:  api.NewEC2SpotRequestManager(requestLister),
		exclusionPeriod: exclusionPeriod,
		checkInterval:   checkInterval,
		mux:             sync.RWMutex{},
		requestCache: &spotRequestCache{
			createTime: time.Now(),
			cache:      make([]*api.SpotRequest, 0),
			mux:        sync.RWMutex{},
		},
		statusCache: &asgStatusCache{
			asgNames: make([]api.AWSAsgName, 0),
			cache:    make(map[api.AWSAsgName]*asgSpotStatus, 0),
			mux:      sync.RWMutex{},
		},
	}
}

type asgSpotStatus struct {
	AsgName            api.AWSAsgName
	AvailabilityZone   api.AWSAvailabilityZone
	IamInstanceProfile api.AWSIamInstanceProfile
	InstanceType       api.AWSInstanceType
	Available          bool
	statusChangeTime   time.Time
}

type spotAvailabilityMonitor struct {
	requestService  api.SpotRequestManager
	checkInterval   time.Duration
	mux             sync.RWMutex
	requestCache    *spotRequestCache
	statusCache     *asgStatusCache
	exclusionPeriod time.Duration
}

// Run starts the monitor's check cycle
func (m *spotAvailabilityMonitor) Run() {
	klog.Info("spot availability monitoring started")
	// monitor ad infinitum.
	for {
		select {
		case <-time.After(m.checkInterval):
			{
				err := m.roundtrip()
				if err != nil {
					klog.Errorf("spot availability check roundtrip failed: %v", err)
				} else {
					klog.V(2).Info("successful spot availability check roundtrip")
				}
			}
		}
	}
}

func (m *spotAvailabilityMonitor) roundtrip() error {
	err := m.updateRequestCache()
	if err != nil {
		return err
	}

	asgNames := m.statusCache.asgNameList()

	for _, asgName := range asgNames {
		asgStatus := m.statusCache.get(asgName)

		asgRequests := m.requestCache.findRequests(asgStatus.AvailabilityZone, asgStatus.IamInstanceProfile, asgStatus.InstanceType)

		status := m.requestsAllValid(asgRequests)

		if asgStatus.Available != status {
			if status == true {
				restExclusionTime := m.calculateRestExclusionTime(asgStatus.statusChangeTime)
				if restExclusionTime > 0 {
					// an ASG remains unavailable for a fixed period of time
					klog.V(2).Infof("spot ASG for type %v is flagged unavailable for another %v", asgStatus.InstanceType, restExclusionTime)
					continue
				}
			} else {
				klog.V(2).Infof("spot ASG for type %v has been flagged unavailable for %v", asgStatus.InstanceType, m.exclusionPeriod)
				err := m.requestService.CancelRequests(asgRequests)
				if err != nil {
					return err
				}
			}

			klog.V(2).Infof("spot ASG for type %v as an availability state transition from %v to %v", asgStatus.InstanceType, asgStatus.Available, status)
			m.statusCache.update(asgName, status)
		}
	}

	return nil
}

func (m *spotAvailabilityMonitor) calculateRestExclusionTime(exclusionStart time.Time) time.Duration {
	return m.exclusionPeriod - time.Now().Sub(exclusionStart)
}

func (m *spotAvailabilityMonitor) updateRequestCache() error {
	spotRequests, err := m.requestService.List()
	if err != nil {
		return err
	}

	m.requestCache.refresh(spotRequests)

	return nil
}

// AsgAvailability checks for a given ASG if it is available or not
func (m *spotAvailabilityMonitor) AsgAvailability(name, iamInstanceProfile, availabilityZone, instanceType string) bool {
	asgStatus := m.asgStatus(name, iamInstanceProfile, availabilityZone, instanceType)
	return asgStatus.Available
}

func (m *spotAvailabilityMonitor) asgStatus(name, iamInstanceProfile, availabilityZone, instanceType string) asgSpotStatus {
	castedName := api.AWSAsgName(name)

	var asgStatus *asgSpotStatus

	exists := m.statusCache.exists(castedName)

	if !exists {
		asgStatus = &asgSpotStatus{
			AsgName:            castedName,
			AvailabilityZone:   api.AWSAvailabilityZone(availabilityZone),
			IamInstanceProfile: api.AWSIamInstanceProfile(iamInstanceProfile),
			InstanceType:       api.AWSInstanceType(instanceType),
			Available:          true,
			statusChangeTime:   time.Time{},
		}

		asgRequests := m.requestCache.findRequests(asgStatus.AvailabilityZone, asgStatus.IamInstanceProfile, asgStatus.InstanceType)
		asgStatus.Available = m.requestsAllValid(asgRequests)

		m.statusCache.add(castedName, asgStatus)
		klog.V(4).Infof("added spot ASG availability status (%v) for group %s", asgStatus.Available, name)
	} else {
		asgStatus = m.statusCache.get(castedName)
	}

	return *asgStatus
}

// requestsAllValid checks for unwanted spot request states,
// if no requests are provided, the response is "true"
// see https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/spot-bid-status.html
func (m *spotAvailabilityMonitor) requestsAllValid(asgRequests []*api.SpotRequest) bool {
	if len(asgRequests) > 0 {
		for _, request := range asgRequests {
			if request.State == api.AWSSpotRequestStateFailed {
				klog.V(4).Infof("spot request %v has invalid state %v", request.ID, request.State)
				return false
			}

			switch request.Status {
			case api.AWSSpotRequestStatusNotAvailable:
				fallthrough
			case api.AWSSpotRequestStatusNotFulfillable:
				fallthrough
			case api.AWSSpotRequestStatusOversubscribed:
				fallthrough
			case api.AWSSpotRequestStatusPriceToLow:
				klog.V(4).Infof("spot request %v has invalid status %v", request.ID, request.Status)
				return false
			}
		}
	}

	return true
}

type asgStatusCache struct {
	asgNames []api.AWSAsgName
	cache    map[api.AWSAsgName]*asgSpotStatus
	mux      sync.RWMutex
}

func (c *asgStatusCache) asgNameList() []api.AWSAsgName {
	c.mux.RLock()
	defer c.mux.RUnlock()
	return c.asgNames
}

func (c *asgStatusCache) exists(asgName api.AWSAsgName) bool {
	c.mux.RLock()
	_, ok := c.cache[asgName]
	c.mux.RUnlock()

	return ok
}

func (c *asgStatusCache) get(asgName api.AWSAsgName) *asgSpotStatus {
	var asgStatus *asgSpotStatus

	c.mux.RLock()
	if actualStatus, exists := c.cache[asgName]; exists {
		asgStatus = actualStatus
	}
	c.mux.RUnlock()

	return asgStatus
}

func (c *asgStatusCache) add(asgName api.AWSAsgName, status *asgSpotStatus) {
	c.mux.Lock()
	if _, exists := c.cache[asgName]; !exists {
		c.asgNames = append(c.asgNames, asgName)
		c.cache[asgName] = status
	}
	c.mux.Unlock()
}

func (c *asgStatusCache) update(asgName api.AWSAsgName, status bool) {
	c.mux.Lock()
	if _, exists := c.cache[asgName]; exists {
		c.cache[asgName].Available = status
		c.cache[asgName].statusChangeTime = time.Now()
	}
	c.mux.Unlock()
}

type spotRequestCache struct {
	createTime time.Time
	cache      []*api.SpotRequest
	mux        sync.RWMutex
}

func (c *spotRequestCache) refresh(requests []*api.SpotRequest) {
	c.mux.Lock()
	if len(requests) > 0 {
		c.cache = requests
		c.createTime = time.Now()
	}
	c.mux.Unlock()
}

func (c *spotRequestCache) findRequests(availabilityZone api.AWSAvailabilityZone, iamInstanceProfile api.AWSIamInstanceProfile, instanceType api.AWSInstanceType) []*api.SpotRequest {
	requests := make([]*api.SpotRequest, 0, len(c.cache))

	c.mux.RLock()
	for _, request := range c.cache {
		if availabilityZone == request.AvailabilityZone && iamInstanceProfile == request.InstanceProfile && instanceType == request.InstanceType {
			requests = append(requests, request)
		}
	}
	c.mux.RUnlock()

	return requests
}
