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

package api

import (
	"sort"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"k8s.io/klog"

	"github.com/aws/aws-sdk-go/service/ec2"
)

// EmptySpotPriceItem is an empty struct for that could be used for default values
var EmptySpotPriceItem SpotPriceItem

type awsSpotPriceHistoryService interface {
	DescribeSpotPriceHistory(input *ec2.DescribeSpotPriceHistoryInput) (*ec2.DescribeSpotPriceHistoryOutput, error)
}

// SpotPriceHistory is the output returned by DescribeSpotPriceHistory
type SpotPriceHistory struct {
	// HistoryItems is a slice of spot price items that implements sort.Interface
	HistoryItems SpotPriceItems
}

// NewEC2SpotPriceService is the constructor of spotPriceHistoryService
func NewEC2SpotPriceService(awsEC2Service awsSpotPriceHistoryService) *spotPriceHistoryService {
	return &spotPriceHistoryService{service: awsEC2Service}
}

type spotPriceHistoryService struct {
	service awsSpotPriceHistoryService
}

// DescribeSpotPriceHistory returns the spot price history for given instance type
func (spd *spotPriceHistoryService) DescribeSpotPriceHistory(instanceType string, availabilityZone string, startTime time.Time) (*SpotPriceHistory, error) {
	req := new(ec2.DescribeSpotPriceHistoryInput)
	req.SetInstanceTypes(aws.StringSlice([]string{instanceType}))
	req.SetAvailabilityZone(availabilityZone)
	req.SetProductDescriptions(aws.StringSlice([]string{"Linux/UNIX"}))

	if startTime.IsZero() {
		klog.V(5).Info("initial history loading - retrieve only the last 10 prices")
		req.SetMaxResults(10)
	} else {
		req.SetStartTime(startTime)
	}

	prices := make(SpotPriceItems, 0)

	for {
		res, err := spd.service.DescribeSpotPriceHistory(req)
		if err != nil {
			return nil, err
		}

		prices = append(prices, convertSpotPriceItems(res.SpotPriceHistory...)...)

		req.NextToken = res.NextToken
		if req.NextToken == nil || len(*req.NextToken) == 0 {
			klog.V(6).Info("breaking history loop after pagination record")
			break
		}
		if startTime.IsZero() {
			klog.V(6).Info("breaking history loop after retrieving last 10 prices")
			break
		}
	}

	sort.Sort(prices)

	return &SpotPriceHistory{HistoryItems: prices}, nil
}

func newSpotPriceItem(price float64, ts time.Time) SpotPriceItem {
	return SpotPriceItem{
		Timestamp: ts,
		Price:     price,
	}
}

// SpotPriceItem consists of a timestamp and a price
type SpotPriceItem struct {
	// Timestamp indicating the occurrence of the price change
	Timestamp time.Time
	// Price of the spot instance at given time
	Price float64
}

// SpotPriceItems is a list of SpotPriceItem
// Implements sort.Interface
type SpotPriceItems []SpotPriceItem

// Len of spot price items
func (sps SpotPriceItems) Len() int {
	return len(sps)
}

// Less returns true if the spot price on the left side is younger than the right one
func (sps SpotPriceItems) Less(i, j int) bool {
	return sps[i].Timestamp.Before(sps[j].Timestamp)
}

// Swap the spot price elements of the given idx
func (sps SpotPriceItems) Swap(i, j int) {
	sps[i], sps[j] = sps[j], sps[i]
}

func convertSpotPriceItems(in ...*ec2.SpotPrice) SpotPriceItems {
	prices := make(SpotPriceItems, len(in))

	for i, item := range in {
		priceValue := aws.StringValue(item.SpotPrice)
		price, err := strconv.ParseFloat(priceValue, 64)
		if err != nil {
			klog.Warningf("Failed to parse aws spot price '%s' to float: %v", priceValue, err)
			continue
		}

		prices[i] = newSpotPriceItem(price, *item.Timestamp)
	}

	return prices
}
