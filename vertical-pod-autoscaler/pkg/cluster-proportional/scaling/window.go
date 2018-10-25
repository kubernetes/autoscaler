package scaling

import (
	"math"
	"time"

	"k8s.io/apimachinery/pkg/util/clock"
)

type windowValues struct {
	// Start is the time at which we started this window
	Start time.Time

	// Retention determines for how long we will retain values (though we always retain the latest)
	Retention time.Duration

	// values records all values we have recorded.  Values older than Window will be removed (lazily)
	values []windowValue
}

type windowValue struct {
	t     time.Time
	value float64
}

func (w *windowValues) Reset(clock clock.Clock, retention time.Duration) {
	w.Start = clock.Now()
	w.Retention = retention
	w.values = nil
}

func (w *windowValues) addObservation(now time.Time, value float64) {
	var filtered []windowValue
	// We always record the latest value, regardless of retention
	filtered = append(filtered, windowValue{t: now, value: value})
	for _, o := range w.values {
		if now.Sub(o.t) <= w.Retention {
			filtered = append(filtered, o)
		}
	}
	w.values = filtered
}

type windowStats struct {
	N   int
	Min float64
	Max float64

	LatestValue     float64
	HasLatest       bool
	LatestTimestamp time.Time
}

func (w *windowValues) stats(now time.Time, window time.Duration) windowStats {
	var stats windowStats
	stats.Min = math.MaxFloat64
	stats.Max = -math.MaxFloat64
	stats.N = 0

	for _, v := range w.values {
		// We always return the latest value, if we have one
		if stats.LatestTimestamp.IsZero() || stats.LatestTimestamp.After(v.t) {
			stats.LatestTimestamp = v.t
			stats.LatestValue = v.value
			stats.HasLatest = true
		}

		if now.Sub(v.t) > window {
			continue
		}
		if v.value < stats.Min {
			stats.Min = v.value
		}
		if v.value > stats.Max {
			stats.Max = v.value
		}
		stats.N++
	}
	return stats
}
