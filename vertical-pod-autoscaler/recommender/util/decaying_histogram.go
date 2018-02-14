/*
Copyright 2018 The Kubernetes Authors.

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

package util

import (
	"fmt"
	"math"
	"time"
)

var (
	// When the decay factor exceeds 2^maxDecayExponent the histogram is
	// renormalized by shifting the decay start time.
	maxDecayExponent = 100
)

// A histogram that gives newer samples a higher weight than the old samples,
// gradually "forgetting" the past samples. The weight of each sample is
// multiplied by the factor of 2^((sampleTime - decayStart) / halfLife).
// This means that the sample loses half of its weight ("importance") after
// each halfLife period.
// Since only relative (and not absolute) weights of samples matter, the
// decayStart can be shifted at any time, which is equivalent to multiplying all
// weights by a constant. In fact this happens whenever the exponents become too
// large, to stay in the range of the floating point arithmentics.
type decayingHistogram struct {
	histogram
	// Decay half life period.
	halfLife time.Duration
	// Reference time for determining the relative age of samples.
	// It is always an integer multiple of halfLife.
	decayStart time.Time
}

// NewDecayingHistogram returns a new DecayingHistogram instance using given options.
func NewDecayingHistogram(options *HistogramOptions, halfLife time.Duration) Histogram {
	return &decayingHistogram{
		histogram:  *NewHistogram(options).(*histogram),
		halfLife:   halfLife,
		decayStart: time.Time{},
	}
}

func (h *decayingHistogram) Percentile(percentile float64) float64 {
	return h.histogram.Percentile(percentile)
}

func (h *decayingHistogram) AddSample(value float64, weight float64, time time.Time) {
	h.histogram.AddSample(value, weight*h.decayFactor(time), time)
}

func (h *decayingHistogram) SubtractSample(value float64, weight float64, time time.Time) {
	h.histogram.SubtractSample(value, weight*h.decayFactor(time), time)
}

func (h *decayingHistogram) Merge(other Histogram) {
	o := other.(*decayingHistogram)
	if h.halfLife != o.halfLife {
		panic("can't merge decaying histograms with different half life periods")
	}
	// Align the older decayStart with the younger one.
	if h.decayStart.Before(o.decayStart) {
		h.shiftDecayStart(o.decayStart)
	} else if o.decayStart.Before(h.decayStart) {
		o.shiftDecayStart(h.decayStart)
	}
	h.histogram.Merge(&o.histogram)
}

func (h *decayingHistogram) Equals(other Histogram) bool {
	h2, typesMatch := (other).(*decayingHistogram)
	return typesMatch && h.halfLife == h2.halfLife && h.decayStart == h2.decayStart && h.histogram.Equals(&h2.histogram)
}

func (h *decayingHistogram) IsEmpty() bool {
	return h.histogram.IsEmpty()
}

func (h *decayingHistogram) String() string {
	return fmt.Sprintf("decayStart: %v, halfLife: %v\n%s", h.decayStart, h.halfLife, h.histogram.String())
}

func (h *decayingHistogram) shiftDecayStart(newDecayStart time.Time) {
	newDecayStart = newDecayStart.Round(h.halfLife)
	exponent := round(float64(h.decayStart.Sub(newDecayStart)) / float64(h.halfLife))
	for bucket := h.histogram.minBucket; bucket <= h.histogram.maxBucket; bucket++ {
		h.histogram.bucketWeight[bucket] = math.Ldexp(h.histogram.bucketWeight[bucket], exponent)
	}
	h.histogram.totalWeight = math.Ldexp(h.histogram.totalWeight, exponent)
	h.decayStart = newDecayStart
	h.updateMinAndMaxBucket()
}

func (h *decayingHistogram) decayFactor(timestamp time.Time) float64 {
	// Max timestamp before the exponent grows too large.
	maxAllowedTimestamp := h.decayStart.Add(
		time.Duration(int64(h.halfLife) * int64(maxDecayExponent)))
	if timestamp.After(maxAllowedTimestamp) {
		// The exponent has grown too large. Renormalize the histogram by
		// shifting the decayStart to the current timestamp and rescaling
		// the weights accordingly.
		h.shiftDecayStart(timestamp)
	}
	return math.Exp2(float64(timestamp.Sub(h.decayStart)) / float64(h.halfLife))
}

func round(x float64) int {
	return int(math.Floor(x + 0.5))
}
