package spot

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/api"
)

var testTimestamp = time.Now().Add(-5 * time.Minute)

func TestHistory_Slice(t *testing.T) {

	cases := []struct {
		name     string
		history  *History
		expected api.SpotPriceItems
	}{
		{
			name:     "empty history",
			history:  newHistoryMock(time.Time{}, 1*time.Second, api.SpotPriceItems{}),
			expected: api.SpotPriceItems{},
		},
		{
			name: "filled history",
			history: newHistoryMock(time.Time{}, 1*time.Second, api.SpotPriceItems{
				{
					Timestamp: testTimestamp,
					Price:     float64(0.5678),
				},
				{
					Timestamp: testTimestamp.Add(-5 * time.Hour),
					Price:     float64(0.6054),
				},
			}),
			expected: api.SpotPriceItems{
				{
					Timestamp: testTimestamp,
					Price:     float64(0.5678),
				},
				{
					Timestamp: testTimestamp.Add(-5 * time.Hour),
					Price:     float64(0.6054),
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(tt *testing.T) {
			res := c.history.Slice()
			assert.Equal(tt, c.expected, res)
		})
	}
}

func TestHistory_Add(t *testing.T) {
	cases := []struct {
		name     string
		history  *History
		items    api.SpotPriceItems
		expected api.SpotPriceItems
	}{
		{
			name:     "adding empty item list to empty history",
			history:  newHistoryMock(time.Time{}, 1*time.Second, api.SpotPriceItems{}),
			items:    api.SpotPriceItems{},
			expected: api.SpotPriceItems{},
		},
		{
			name:    "adding filled item list to empty history",
			history: newHistoryMock(time.Time{}, 1*time.Second, api.SpotPriceItems{}),
			items: api.SpotPriceItems{
				{
					Timestamp: testTimestamp,
					Price:     float64(0.5678),
				},
				{
					Timestamp: testTimestamp.Add(-5 * time.Hour),
					Price:     float64(0.6054),
				},
			},
			expected: api.SpotPriceItems{
				{
					Timestamp: testTimestamp.Add(-5 * time.Hour),
					Price:     float64(0.6054),
				},
				{
					Timestamp: testTimestamp,
					Price:     float64(0.5678),
				},
			},
		},
		{
			name: "adding empty item list to filled history",
			history: newHistoryMock(time.Time{}, 1*time.Second, api.SpotPriceItems{
				{
					Timestamp: testTimestamp,
					Price:     float64(0.5678),
				},
				{
					Timestamp: testTimestamp.Add(-5 * time.Hour),
					Price:     float64(0.6054),
				},
			}),
			items: api.SpotPriceItems{},
			expected: api.SpotPriceItems{
				{
					Timestamp: testTimestamp.Add(-5 * time.Hour),
					Price:     float64(0.6054),
				},
				{
					Timestamp: testTimestamp,
					Price:     float64(0.5678),
				},
			},
		},
		{
			name: "adding filled item list to filled history",
			history: newHistoryMock(testTimestamp, 1*time.Second, api.SpotPriceItems{
				{
					Timestamp: testTimestamp,
					Price:     float64(0.5678),
				},
				{
					Timestamp: testTimestamp.Add(-5 * time.Hour),
					Price:     float64(0.6054),
				},
			}),
			items: api.SpotPriceItems{
				{
					Timestamp: testTimestamp.Add(-3 * time.Hour),
					Price:     float64(0.5783),
				},
				{
					Timestamp: testTimestamp.Add(-2 * time.Hour),
					Price:     float64(0.5609),
				},
			},
			expected: api.SpotPriceItems{
				{
					Timestamp: testTimestamp.Add(-5 * time.Hour),
					Price:     float64(0.6054),
				},
				{
					Timestamp: testTimestamp.Add(-3 * time.Hour),
					Price:     float64(0.5783),
				},
				{
					Timestamp: testTimestamp.Add(-2 * time.Hour),
					Price:     float64(0.5609),
				},
				{
					Timestamp: testTimestamp,
					Price:     float64(0.5678),
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(tt *testing.T) {
			c.history.Add(c.items)
			res := c.history.Slice()

			assert.WithinDuration(tt, time.Now(), c.history.LastSync(), 10*time.Millisecond, "last sync time too old")
			assert.Equal(tt, c.expected, res)
		})
	}
}

func TestHistory_Empty(t *testing.T) {
	cases := []struct {
		name     string
		history  *History
		expected bool
	}{
		{
			name:     "empty history: return true",
			history:  newHistoryMock(time.Time{}, 1*time.Second, api.SpotPriceItems{}),
			expected: true,
		},
		{
			name: "filled history: return false",
			history: newHistoryMock(time.Time{}, 1*time.Second, api.SpotPriceItems{
				{
					Timestamp: time.Now(),
					Price:     float64(0.5678),
				},
			}),
			expected: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(tt *testing.T) {
			res := c.history.Empty()
			assert.Equal(tt, c.expected, res)
		})
	}
}

func TestHistory_LastItem(t *testing.T) {
	cases := []struct {
		name          string
		history       *History
		expected      api.SpotPriceItem
		errorExpected bool
	}{
		{
			name:          "no last item found",
			history:       newHistoryMock(time.Time{}, 1*time.Second, api.SpotPriceItems{}),
			expected:      api.EmptySpotPriceItem,
			errorExpected: true,
		},
		{
			name: "last item found",
			history: newHistoryMock(time.Time{}, 1*time.Second, api.SpotPriceItems{
				{
					Timestamp: testTimestamp.Add(-5 * time.Hour),
					Price:     float64(0.6054),
				},
				{
					Timestamp: testTimestamp,
					Price:     float64(0.5678),
				},
			}),
			expected: api.SpotPriceItem{
				Timestamp: testTimestamp,
				Price:     float64(0.5678),
			},
			errorExpected: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(tt *testing.T) {
			res, err := c.history.LastItem()

			if c.errorExpected {
				assert.NotNil(tt, err)
				assert.Equal(tt, ErrEmptySpotPriceHistory, err)
				assert.Equal(tt, c.expected, res)
			} else {
				assert.Nil(tt, err)
				assert.Equal(tt, c.expected, res)
			}

		})
	}
}

func TestHistory_Housekeep(t *testing.T) {
	cases := []struct {
		name     string
		history  *History
		expected api.SpotPriceItems
	}{
		{
			name:     "empty history - no change",
			history:  newHistoryMock(time.Time{}, 1*time.Second, api.SpotPriceItems{}),
			expected: api.SpotPriceItems{},
		},
		{
			name: "some entries deprecated",
			history: newHistoryMock(time.Time{}, 3*time.Hour+30*time.Minute, api.SpotPriceItems{
				{
					Timestamp: testTimestamp.Add(-5 * time.Hour),
					Price:     float64(0.6054),
				},
				{
					Timestamp: testTimestamp.Add(-4 * time.Hour),
					Price:     float64(0.5783),
				},
				{
					Timestamp: testTimestamp.Add(-3 * time.Hour),
					Price:     float64(0.5783),
				},
				{
					Timestamp: testTimestamp.Add(-2 * time.Hour),
					Price:     float64(0.5609),
				},
				{
					Timestamp: testTimestamp,
					Price:     float64(0.5678),
				},
			}),
			expected: api.SpotPriceItems{
				{
					Timestamp: testTimestamp.Add(-3 * time.Hour),
					Price:     float64(0.5783),
				},
				{
					Timestamp: testTimestamp.Add(-2 * time.Hour),
					Price:     float64(0.5609),
				},
				{
					Timestamp: testTimestamp,
					Price:     float64(0.5678),
				},
			},
		},
		{
			name: "all entries deprecated, history empty after cleanup - last item still present",
			history: newHistoryMock(time.Time{}, 1*time.Second, api.SpotPriceItems{
				{
					Timestamp: testTimestamp.Add(-5 * time.Hour),
					Price:     float64(0.6054),
				},
				{
					Timestamp: testTimestamp.Add(-3 * time.Hour),
					Price:     float64(0.5783),
				},
				{
					Timestamp: testTimestamp.Add(-2 * time.Hour),
					Price:     float64(0.5609),
				},
				{
					Timestamp: testTimestamp,
					Price:     float64(0.5678),
				},
			}),
			expected: api.SpotPriceItems{
				{
					Timestamp: testTimestamp,
					Price:     float64(0.5678),
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(tt *testing.T) {
			c.history.Housekeep()
			res := c.history.Slice()

			assert.Equal(tt, c.expected, res)
		})
	}
}

func TestHistory_SetLastSync(t *testing.T) {
	cases := []struct {
		name    string
		history *History
	}{
		{
			name:    "zero sync time: set actual time",
			history: newHistoryMock(time.Time{}, 1*time.Second, api.SpotPriceItems{}),
		},
		{
			name:    "existing sync time: set actual time",
			history: newHistoryMock(testTimestamp, 1*time.Second, api.SpotPriceItems{}),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(tt *testing.T) {
			c.history.SetLastSync()
			assert.WithinDuration(tt, time.Now(), c.history.LastSync(), 10*time.Millisecond)
		})
	}
}

func newHistoryMock(lastSync time.Time, maxAge time.Duration, items api.SpotPriceItems) *History {
	history := new(History)

	history.lastSync = lastSync
	history.maxAge = maxAge
	history.items = items

	return history
}
