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
	"fmt"
	"time"

	"github.com/golang/glog"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/api"
)

const (
	lookupWindow = time.Minute * 30
	cacheWindow  = time.Hour * 24 * 3
	cacheMaxAge  = time.Minute * 5
)

type instanceTypeInZone struct {
	instanceType     string
	availabilityZone string
}

type spotPriceBucket map[instanceTypeInZone]*History

// Descriptor describes the price interface
type Descriptor interface {
	Price(instanceType string, bidPrice float64, availabilityZones ...string) (float64, error)
}

// NewDescriptor is the constructor of the constructor
func NewDescriptor(spotPriceAPI api.SpotPriceHistoryDescriber) *descriptor {
	return &descriptor{
		bucket: make(spotPriceBucket),
		api:    spotPriceAPI,
	}
}

type descriptor struct {
	bucket spotPriceBucket
	api    api.SpotPriceHistoryDescriber
}

// Price returns the current price, average over the availability zones but the max value within 30 minutes of a zone,
// of the given instanceType.
// It returns an error if the current price is greater than the bid price.
func (d *descriptor) Price(instanceType string, bidPrice float64, availabilityZones ...string) (float64, error) {
	var avgPrice float64
	prices := make([]float64, len(availabilityZones))

	for i, zone := range availabilityZones {
		maxPrice, err := d.maxSpotPriceForDuration(instanceType, zone, lookupWindow)
		if err != nil {
			return avgPrice, err
		}

		if maxPrice == 0.0 {
			return avgPrice, fmt.Errorf("got invalid spot price of 0.0 for instance type %s in availability zone %s", instanceType, zone)
		}

		if maxPrice > bidPrice {
			return 0, fmt.Errorf("spot price bid of %.4f lower than current offer of %.4f at %s", bidPrice, maxPrice, zone)
		}

		prices[i] = maxPrice
	}

	var sum float64
	for _, price := range prices {
		sum += price
	}

	avgPrice = sum / float64(len(prices))

	return avgPrice, nil
}

// cumulatedSpotPriceLastHour collects the spot price history of the last hour for every AZ provided.
// It takes the highest spot price of every AZ within the last hour and returns the average across the AZs.
func (d *descriptor) maxSpotPriceForDuration(instanceType string, availabilityZone string, lookupWindow time.Duration) (float64, error) {
	var maxPrice float64

	history, err := d.spotPriceHistory(instanceType, availabilityZone)
	if err != nil {
		return maxPrice, err
	}

	if history.Empty() {
		return maxPrice, fmt.Errorf("no spot price information for instance %s in availability zone %s", instanceType, availabilityZone)
	}

	startTime := time.Now().Truncate(lookupWindow)

	for _, price := range history.Slice() {
		if price.Timestamp.Before(startTime) {
			continue
		}

		if maxPrice < price.Price {
			maxPrice = price.Price
		}
	}

	// The case when there are no new price information within the requested time window.
	if maxPrice == 0.0 {
		item, err := history.LastItem()
		if err != nil {
			return maxPrice, fmt.Errorf("failed to fetch last history item: %v", err)
		}

		glog.Warningf(
			"no spot price information newer than %s, using last known price of %f which is %s old",
			lookupWindow,
			item.Price,
			time.Now().Sub(item.Timestamp),
		)
		maxPrice = item.Price
	}

	return maxPrice, nil
}

func (d *descriptor) spotPriceHistory(instanceType, availabilityZone string) (*History, error) {
	if d.syncRequired(instanceType, availabilityZone) {
		if err := d.syncSpotPriceHistory(instanceType, availabilityZone); err != nil {
			return nil, fmt.Errorf("spot price sync failed: %v", err)
		}
	}

	instanceZone := instanceTypeInZone{instanceType: instanceType, availabilityZone: availabilityZone}
	return d.bucket[instanceZone], nil
}
func (d *descriptor) syncRequired(instanceType, availabilityZone string) bool {
	instanceZone := instanceTypeInZone{instanceType: instanceType, availabilityZone: availabilityZone}

	history, found := d.bucket[instanceZone]
	if !found {
		return true
	}

	return history.LastSync().Before(time.Now().Truncate(cacheMaxAge))
}

func (d *descriptor) syncSpotPriceHistory(instanceType string, availabilityZone string) error {
	instanceZone := instanceTypeInZone{instanceType: instanceType, availabilityZone: availabilityZone}
	var begin time.Time

	history, found := d.bucket[instanceZone]
	if found {
		begin = history.LastSync()
	} else {
		begin = time.Now().Truncate(cacheWindow)

		history = &History{maxAge: cacheWindow, items: make(api.SpotPriceItems, 0)}
		d.bucket[instanceZone] = history
	}

	defer history.Housekeep()
	defer history.SetLastSync()

	res, err := d.api.DescribeSpotPriceHistory(instanceType, availabilityZone, begin)
	if err != nil {
		return err
	}

	history.Add(res.HistoryItems)

	return nil
}
