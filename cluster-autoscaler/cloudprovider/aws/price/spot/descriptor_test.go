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

	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/api"
)

func TestDescriptor_syncSpotPriceHistory(t *testing.T) {
	cases := []struct {
		name          string
		local         spotPriceBucket
		aws           spotPriceBucket
		expected      api.SpotPriceItems
		errorExpected bool
	}{
		{
			name:  "empty history for unknown type - no change",
			local: spotPriceBucket{},
			aws: spotPriceBucket{
				instanceTypeInZone{"m4.xlarge", "us-east-1"}: newHistoryMock(time.Time{}, 1*time.Second, api.SpotPriceItems{
					{
						Timestamp: testTimestamp,
						Price:     float64(0.5678),
					},
					{
						Timestamp: testTimestamp.Add(-5 * time.Hour),
						Price:     float64(0.6054),
					},
				}),
			},
			expected:      api.SpotPriceItems{},
			errorExpected: true,
		},
		{
			name:  "empty history for known type - updated",
			local: spotPriceBucket{},
			aws: spotPriceBucket{
				instanceTypeInZone{"m4.2xlarge", "us-east-1a"}: newHistoryMock(time.Time{}, 1*time.Second, api.SpotPriceItems{
					{
						Timestamp: testTimestamp.Add(-5 * time.Minute),
						Price:     float64(1.034),
					},
					{
						Timestamp: testTimestamp.Add(-1 * time.Hour),
						Price:     float64(1.203),
					},
				}),
			},
			expected: api.SpotPriceItems{
				{
					Timestamp: testTimestamp.Add(-1 * time.Hour),
					Price:     float64(1.203),
				},
				{
					Timestamp: testTimestamp.Add(-5 * time.Minute),
					Price:     float64(1.034),
				},
			},
			errorExpected: false,
		},
		{
			name: "history updated despite all entries being outside the cache window (fallback)",
			local: spotPriceBucket{
				instanceTypeInZone{"m4.2xlarge", "us-east-1a"}: newHistoryMock(testTimestamp.Add(-30*time.Minute), 1*time.Second, api.SpotPriceItems{
					{
						Timestamp: testTimestamp.Add(-1 * time.Hour),
						Price:     float64(1.203),
					},
				}),
			},
			aws: spotPriceBucket{
				instanceTypeInZone{"m4.2xlarge", "us-east-1a"}: newHistoryMock(time.Time{}, 1*time.Second, api.SpotPriceItems{
					{
						Timestamp: testTimestamp.Add(-5 * time.Minute),
						Price:     float64(1.034),
					},
					{
						Timestamp: testTimestamp.Add(-1 * time.Hour),
						Price:     float64(1.203),
					},
				}),
			},
			expected: api.SpotPriceItems{
				{
					Timestamp: testTimestamp.Add(-5 * time.Minute),
					Price:     float64(1.034),
				},
			},
			errorExpected: false,
		},
		{
			name: "history updated within the cache window",
			local: spotPriceBucket{
				instanceTypeInZone{"m4.2xlarge", "us-east-1a"}: newHistoryMock(testTimestamp.Add(-30*time.Minute), 10*time.Hour, api.SpotPriceItems{
					{
						Timestamp: testTimestamp.Add(-1 * time.Hour),
						Price:     float64(1.203),
					},
				}),
			},
			aws: spotPriceBucket{
				instanceTypeInZone{"m4.2xlarge", "us-east-1a"}: newHistoryMock(time.Time{}, 1*time.Second, api.SpotPriceItems{
					{
						Timestamp: testTimestamp.Add(-5 * time.Minute),
						Price:     float64(1.034),
					},
					{
						Timestamp: testTimestamp.Add(-1 * time.Hour),
						Price:     float64(1.203),
					},
				}),
			},
			expected: api.SpotPriceItems{
				{
					Timestamp: testTimestamp.Add(-1 * time.Hour),
					Price:     float64(1.203),
				},
				{
					Timestamp: testTimestamp.Add(-5 * time.Minute),
					Price:     float64(1.034),
				},
			},
			errorExpected: false,
		},
		{
			name: "no new records found since last update - history unchanged",
			local: spotPriceBucket{
				instanceTypeInZone{"m4.2xlarge", "us-east-1a"}: newHistoryMock(testTimestamp, 10*time.Hour, api.SpotPriceItems{
					{
						Timestamp: testTimestamp.Add(-1 * time.Hour),
						Price:     float64(1.203),
					},
					{
						Timestamp: testTimestamp.Add(-5 * time.Minute),
						Price:     float64(1.034),
					},
				}),
			},
			aws: spotPriceBucket{
				instanceTypeInZone{"m4.2xlarge", "us-east-1a"}: newHistoryMock(time.Time{}, 1*time.Second, api.SpotPriceItems{
					{
						Timestamp: testTimestamp.Add(-5 * time.Minute),
						Price:     float64(1.034),
					},
					{
						Timestamp: testTimestamp.Add(-1 * time.Hour),
						Price:     float64(1.203),
					},
				}),
			},
			expected: api.SpotPriceItems{
				{
					Timestamp: testTimestamp.Add(-1 * time.Hour),
					Price:     float64(1.203),
				},
				{
					Timestamp: testTimestamp.Add(-5 * time.Minute),
					Price:     float64(1.034),
				},
			},
			errorExpected: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(tt *testing.T) {
			awsApi := newFakeSpotPriceHistoryDescriber(c.aws)
			desc := newDescriptorMock(c.local, awsApi)

			itz := instanceTypeInZone{"m4.2xlarge", "us-east-1a"}

			err := desc.syncSpotPriceHistory(itz.instanceType, itz.availabilityZone)

			if c.errorExpected {
				assert.NotNil(tt, err, "an error should have been returned")
			} else {
				assert.Nil(tt, err, "no error should have been returned")
			}

			assert.WithinDuration(tt, time.Now(), desc.bucket[itz].LastSync(), 10*time.Millisecond, "last sync time too old")
			assert.Equal(tt, c.expected, desc.bucket[itz].items, "bucked has not been filled with the expected entries")
		})
	}
}

func TestDescriptor_syncRequired(t *testing.T) {
	cases := []struct {
		name      string
		bucketKey instanceTypeInZone
		local     spotPriceBucket
		expected  bool
	}{
		{
			name:      "unknown bucket: sync required",
			bucketKey: instanceTypeInZone{"m4.xlarge", "us-east-1"},
			local:     spotPriceBucket{},
			expected:  true,
		},
		{
			name:      "deprecated cache: sync required",
			bucketKey: instanceTypeInZone{"m4.2xlarge", "us-east-1a"},
			local: spotPriceBucket{
				instanceTypeInZone{"m4.2xlarge", "us-east-1a"}: newHistoryMock(testTimestamp.Add(-cacheMaxAge), 10*time.Hour, api.SpotPriceItems{
					{
						Timestamp: testTimestamp.Add(-1 * time.Hour),
						Price:     float64(1.203),
					},
					{
						Timestamp: testTimestamp.Add(-5 * time.Minute),
						Price:     float64(1.034),
					},
				}),
			},
			expected: true,
		},
		{
			name:      "valid cache age: sync not required",
			bucketKey: instanceTypeInZone{"m4.2xlarge", "us-east-1a"},
			local: spotPriceBucket{
				instanceTypeInZone{"m4.2xlarge", "us-east-1a"}: newHistoryMock(time.Now(), 10*time.Hour, api.SpotPriceItems{
					{
						Timestamp: testTimestamp.Add(-1 * time.Hour),
						Price:     float64(1.203),
					},
					{
						Timestamp: testTimestamp.Add(-5 * time.Minute),
						Price:     float64(1.034),
					},
				}),
			},
			expected: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(tt *testing.T) {
			awsApi := newFakeSpotPriceHistoryDescriber(spotPriceBucket{})
			desc := newDescriptorMock(c.local, awsApi)

			assert.Equal(tt, c.expected, desc.syncRequired(c.bucketKey.instanceType, c.bucketKey.availabilityZone))
		})
	}
}

func TestDescriptor_spotPriceHistory(t *testing.T) {
	cases := []struct {
		name          string
		local         spotPriceBucket
		aws           spotPriceBucket
		expected      api.SpotPriceItems
		errorExpected bool
	}{
		{
			name:  "empty history for unknown type - no change",
			local: spotPriceBucket{},
			aws: spotPriceBucket{
				instanceTypeInZone{"m4.xlarge", "us-east-1"}: newHistoryMock(time.Time{}, 1*time.Second, api.SpotPriceItems{
					{
						Timestamp: testTimestamp,
						Price:     float64(0.5678),
					},
					{
						Timestamp: testTimestamp.Add(-5 * time.Hour),
						Price:     float64(0.6054),
					},
				}),
			},
			expected:      api.SpotPriceItems{},
			errorExpected: true,
		},
		{
			name:  "empty history for known type - updated",
			local: spotPriceBucket{},
			aws: spotPriceBucket{
				instanceTypeInZone{"m4.2xlarge", "us-east-1a"}: newHistoryMock(time.Time{}, 1*time.Second, api.SpotPriceItems{
					{
						Timestamp: testTimestamp.Add(-5 * time.Minute),
						Price:     float64(1.034),
					},
					{
						Timestamp: testTimestamp.Add(-1 * time.Hour),
						Price:     float64(1.203),
					},
				}),
			},
			expected: api.SpotPriceItems{
				{
					Timestamp: testTimestamp.Add(-1 * time.Hour),
					Price:     float64(1.203),
				},
				{
					Timestamp: testTimestamp.Add(-5 * time.Minute),
					Price:     float64(1.034),
				},
			},
			errorExpected: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(tt *testing.T) {
			awsApi := newFakeSpotPriceHistoryDescriber(c.aws)
			desc := newDescriptorMock(c.local, awsApi)

			itz := instanceTypeInZone{"m4.2xlarge", "us-east-1a"}

			res, err := desc.spotPriceHistory(itz.instanceType, itz.availabilityZone)

			if c.errorExpected {
				assert.NotNil(tt, err, "an error should have been returned")
				assert.Nil(tt, res, "no history should have been returned")
			} else {
				assert.Nil(tt, err, "no error should have been returned")
				assert.Equal(tt, c.expected, res.items, "history is not filled with the expected entries")
			}

			assert.WithinDuration(tt, time.Now(), desc.bucket[itz].LastSync(), 10*time.Millisecond, "last sync time too old")
		})
	}
}

func TestDescriptor_maxSpotPriceForDuration(t *testing.T) {
	actualTime := time.Now()

	cases := []struct {
		name          string
		local         spotPriceBucket
		lookupWindow  time.Duration
		expected      float64
		errorExpected bool
		expectedError string
	}{
		{
			name: "empty history for unknown type - returns 0.0",
			local: spotPriceBucket{
				instanceTypeInZone{"m4.xlarge", "us-east-1"}: newHistoryMock(actualTime, 10*time.Hour, api.SpotPriceItems{
					{
						Timestamp: actualTime.Add(-5 * time.Minute),
						Price:     float64(1.034),
					},
					{
						Timestamp: actualTime.Add(-1 * time.Hour),
						Price:     float64(1.203),
					},
				}),
			},
			lookupWindow:  30 * time.Minute,
			expected:      0.0,
			errorExpected: true,
			expectedError: "spot price sync failed: instance info not available for instance type m4.2xlarge in AZ us-east-1a",
		},
		{
			name: "empty history for known type - returns 0.0",
			local: spotPriceBucket{
				instanceTypeInZone{"m4.2xlarge", "us-east-1a"}: newHistoryMock(actualTime, 10*time.Hour, api.SpotPriceItems{}),
			},
			lookupWindow:  30 * time.Minute,
			expected:      0.0,
			errorExpected: true,
			expectedError: "no spot price information for instance m4.2xlarge in availability zone us-east-1a",
		},
		{
			name: "history outside lookup window - returns last price",
			local: spotPriceBucket{
				instanceTypeInZone{"m4.2xlarge", "us-east-1a"}: newHistoryMock(actualTime, 10*time.Hour, api.SpotPriceItems{
					{
						Timestamp: actualTime.Add(-1 * time.Hour),
						Price:     float64(1.203),
					},
					{
						Timestamp: actualTime.Add(-5 * time.Minute),
						Price:     float64(1.034),
					},
				}),
			},
			lookupWindow:  1 * time.Minute,
			expected:      1.034,
			errorExpected: false,
		},
		{
			name: "history with short lookup window - returns max price",
			local: spotPriceBucket{
				instanceTypeInZone{"m4.2xlarge", "us-east-1a"}: newHistoryMock(actualTime, 10*time.Hour, api.SpotPriceItems{
					{
						Timestamp: actualTime.Add(-1 * time.Hour),
						Price:     float64(1.203),
					},
					{
						Timestamp: actualTime.Add(-5 * time.Minute),
						Price:     float64(1.034),
					},
				}),
			},
			lookupWindow:  30 * time.Minute,
			expected:      1.034,
			errorExpected: false,
		},
		{
			name: "history with long lookup window - returns max price",
			local: spotPriceBucket{
				instanceTypeInZone{"m4.2xlarge", "us-east-1a"}: newHistoryMock(actualTime, 10*time.Hour, api.SpotPriceItems{
					{
						Timestamp: actualTime.Add(-5 * time.Minute),
						Price:     float64(1.034),
					},
					{
						Timestamp: actualTime.Add(-1 * time.Hour),
						Price:     float64(1.203),
					},
				}),
			},
			lookupWindow:  3 * time.Hour,
			expected:      1.203,
			errorExpected: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(tt *testing.T) {
			awsApi := newFakeSpotPriceHistoryDescriber(spotPriceBucket{})
			desc := newDescriptorMock(c.local, awsApi)

			itz := instanceTypeInZone{"m4.2xlarge", "us-east-1a"}

			res, err := desc.maxSpotPriceForDuration(itz.instanceType, itz.availabilityZone, c.lookupWindow)

			if c.errorExpected {
				assert.NotNil(tt, err, "an error should have been returned")
				assert.Equal(tt, c.expectedError, err.Error(), "false error message returned")
			} else {
				assert.Nil(tt, err, "no error should have been returned")
			}

			assert.Equal(tt, c.expected, res, "false max price for duration returned")
		})
	}
}

func TestDescriptor_Price(t *testing.T) {
	actualTime := time.Now()

	cases := []struct {
		name              string
		bidPrice          float64
		availabilityZones []string
		local             spotPriceBucket
		expected          float64
		errorExpected     bool
		expectedError     string
	}{
		{
			name:              "instance type unknown in all AZ - returns 0.0",
			bidPrice:          1.8,
			availabilityZones: []string{"us-east-1a", "us-east-1b"},
			local: spotPriceBucket{
				instanceTypeInZone{"m4.xlarge", "us-east-1"}: newHistoryMock(actualTime, 10*time.Hour, api.SpotPriceItems{
					{
						Timestamp: actualTime.Add(-5 * time.Minute),
						Price:     float64(1.034),
					},
					{
						Timestamp: actualTime.Add(-1 * time.Minute),
						Price:     float64(1.203),
					},
				}),
			},
			expected:      0.0,
			errorExpected: true,
			expectedError: "got invalid spot price of 0.0 for instance type m4.2xlarge in availability zones [us-east-1a us-east-1b]",
		},
		{
			name:              "empty history for known type in all AZ - returns 0.0",
			bidPrice:          1.8,
			availabilityZones: []string{"us-east-1a", "us-east-1b"},
			local: spotPriceBucket{
				instanceTypeInZone{"m4.2xlarge", "us-east-1a"}: newHistoryMock(actualTime, 10*time.Hour, api.SpotPriceItems{}),
			},
			expected:      0.0,
			errorExpected: true,
			expectedError: "got invalid spot price of 0.0 for instance type m4.2xlarge in availability zones [us-east-1a us-east-1b]",
		},
		{
			name:              "empty history for known type in all AZ - returns 0.0",
			bidPrice:          1.8,
			availabilityZones: []string{"us-east-1a", "us-east-1b"},
			local: spotPriceBucket{
				instanceTypeInZone{"m4.2xlarge", "us-east-1a"}: newHistoryMock(actualTime, 10*time.Hour, api.SpotPriceItems{
					{
						Timestamp: actualTime.Add(-5 * time.Minute),
						Price:     float64(2.056),
					},
					{
						Timestamp: actualTime.Add(-1 * time.Minute),
						Price:     float64(1.203),
					},
				}),
			},
			expected:      0.0,
			errorExpected: true,
			expectedError: "spot price bid of 1.8000 lower than current offer of 2.0560 at us-east-1a",
		},
		{
			name:              "history outside lookup window - returns last price",
			bidPrice:          1.8,
			availabilityZones: []string{"us-east-1a", "us-east-1b"},
			local: spotPriceBucket{
				instanceTypeInZone{"m4.2xlarge", "us-east-1a"}: newHistoryMock(actualTime, 10*time.Hour, api.SpotPriceItems{
					{
						Timestamp: actualTime.Add(-5 * time.Hour),
						Price:     float64(1.034),
					},
					{
						Timestamp: actualTime.Add(-1 * time.Hour),
						Price:     float64(1.203),
					},
				}),
			},
			expected:      1.203,
			errorExpected: false,
		},
		{
			name:              "single AZ price with short lookup window - returns max average",
			bidPrice:          1.8,
			availabilityZones: []string{"us-east-1a", "us-east-1b"},
			local: spotPriceBucket{
				instanceTypeInZone{"m4.2xlarge", "us-east-1a"}: newHistoryMock(actualTime, 10*time.Hour, api.SpotPriceItems{
					{
						Timestamp: actualTime.Add(-5 * time.Minute),
						Price:     float64(1.034),
					},
					{
						Timestamp: actualTime.Add(-1 * time.Minute),
						Price:     float64(1.203),
					},
				}),
			},
			expected:      1.203,
			errorExpected: false,
		},
		{
			name:              "single AZ price with long lookup window - returns max average",
			bidPrice:          1.8,
			availabilityZones: []string{"us-east-1a", "us-east-1b"},
			local: spotPriceBucket{
				instanceTypeInZone{"m4.2xlarge", "us-east-1a"}: newHistoryMock(actualTime, 10*time.Hour, api.SpotPriceItems{
					{
						Timestamp: actualTime.Add(-5 * time.Minute),
						Price:     float64(1.034),
					},
					{
						Timestamp: actualTime.Add(-1 * time.Minute),
						Price:     float64(1.203),
					},
				}),
			},
			expected:      1.203,
			errorExpected: false,
		},
		{
			name:              "multi AZ price - returns max average",
			bidPrice:          1.8,
			availabilityZones: []string{"us-east-1a", "us-east-1b"},
			local: spotPriceBucket{
				instanceTypeInZone{"m4.2xlarge", "us-east-1a"}: newHistoryMock(actualTime, 10*time.Hour, api.SpotPriceItems{
					{
						Timestamp: actualTime.Add(-5 * time.Minute),
						Price:     float64(1.034),
					},
					{
						Timestamp: actualTime.Add(-1 * time.Minute),
						Price:     float64(1.203),
					},
				}),
				instanceTypeInZone{"m4.2xlarge", "us-east-1b"}: newHistoryMock(actualTime, 10*time.Hour, api.SpotPriceItems{
					{
						Timestamp: actualTime.Add(-12 * time.Minute),
						Price:     float64(1.389),
					},
					{
						Timestamp: actualTime.Add(-2 * time.Minute),
						Price:     float64(1.123),
					},
				}),
			},
			expected:      1.296,
			errorExpected: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(tt *testing.T) {
			awsApi := newFakeSpotPriceHistoryDescriber(spotPriceBucket{})
			desc := newDescriptorMock(c.local, awsApi)

			res, err := desc.Price("m4.2xlarge", c.bidPrice, c.availabilityZones...)

			if c.errorExpected {
				assert.NotNil(tt, err, "an error should have been returned")
				assert.Equal(tt, c.expectedError, err.Error(), "false error message returned")
			} else {
				assert.Nil(tt, err, "no error should have been returned")
			}

			assert.Equal(tt, c.expected, res, "false max price for duration returned")
		})
	}
}

func newDescriptorMock(bucket spotPriceBucket, spotApi *fakeSpotPriceHistoryDescriber) *descriptor {
	d := NewDescriptor(spotApi)

	d.bucket = bucket

	return d
}

func newFakeSpotPriceHistoryDescriber(buckets spotPriceBucket) *fakeSpotPriceHistoryDescriber {
	f := new(fakeSpotPriceHistoryDescriber)

	f.c = map[instanceTypeInZone]*api.SpotPriceHistory{}

	for itz, history := range buckets {
		f.c[itz] = &api.SpotPriceHistory{
			history.Slice(),
		}
	}

	return f
}

type fakeSpotPriceHistoryDescriber struct {
	c map[instanceTypeInZone]*api.SpotPriceHistory
}

func (i *fakeSpotPriceHistoryDescriber) DescribeSpotPriceHistory(instanceType string, availabilityZone string, startTime time.Time) (*api.SpotPriceHistory, error) {
	itz := instanceTypeInZone{instanceType, availabilityZone}

	if history, found := i.c[itz]; found {
		filtered := new(api.SpotPriceHistory)
		filtered.HistoryItems = make(api.SpotPriceItems, 0, len(history.HistoryItems))

		for _, item := range history.HistoryItems {
			if item.Timestamp.After(startTime) {
				filtered.HistoryItems = append(filtered.HistoryItems, item)
			}
		}

		return filtered, nil
	}

	return nil, fmt.Errorf("instance info not available for instance type %s in AZ %s", instanceType, availabilityZone)
}
