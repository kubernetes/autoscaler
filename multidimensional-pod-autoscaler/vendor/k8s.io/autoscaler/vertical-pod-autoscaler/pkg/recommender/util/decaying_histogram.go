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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
)

var (
	// When the decay factor exceeds 2^maxDecayExponent the histogram is
	// renormalized by shifting the decay start time forward.
	maxDecayExponent = 100
)

// A histogram that gives newer samples a higher weight than the old samples,
// gradually decaying ("forgetting") the past samples. The weight of each sample
// is multiplied by the factor of 2^((sampleTime - referenceTimestamp) / halfLife).
// This means that the sample loses half of its weight ("importance") with
// each halfLife period.
// Since only relative (and not absolute) weights of samples matter, the
// referenceTimestamp can be shifted at any time, which is equivalent to multiplying all
// weights by a constant. In practice the referenceTimestamp is shifted forward whenever
// the exponents become too large, to avoid floating point arithmetics overflow.
type decayingHistogram struct {
	histogram
	// Decay half life period.
	halfLife time.Duration
	// Reference time for determining the relative age of samples.
	// It is always an integer multiple of halfLife.
	referenceTimestamp time.Time
}

// NewDecayingHistogram returns a new DecayingHistogram instance using given options.
func NewDecayingHistogram(options HistogramOptions, halfLife time.Duration) Histogram {
	return &decayingHistogram{
		histogram:          *NewHistogram(options).(*histogram),
		halfLife:           halfLife,
		referenceTimestamp: time.Time{},
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
	// Align the older referenceTimestamp with the younger one.
	if h.referenceTimestamp.Before(o.referenceTimestamp) {
		h.shiftReferenceTimestamp(o.referenceTimestamp)
	} else if o.referenceTimestamp.Before(h.referenceTimestamp) {
		o.shiftReferenceTimestamp(h.referenceTimestamp)
	}
	h.histogram.Merge(&o.histogram)
}

func (h *decayingHistogram) Equals(other Histogram) bool {
	h2, typesMatch := (other).(*decayingHistogram)
	return typesMatch && h.halfLife == h2.halfLife && h.referenceTimestamp == h2.referenceTimestamp && h.histogram.Equals(&h2.histogram)
}

func (h *decayingHistogram) IsEmpty() bool {
	return h.histogram.IsEmpty()
}

func (h *decayingHistogram) String() string {
	return fmt.Sprintf("referenceTimestamp: %v, halfLife: %v\n%s", h.referenceTimestamp, h.halfLife, h.histogram.String())
}

func (h *decayingHistogram) shiftReferenceTimestamp(newreferenceTimestamp time.Time) {
	// Make sure the decay start is an integer multiple of halfLife.
	newreferenceTimestamp = newreferenceTimestamp.Round(h.halfLife)
	exponent := round(float64(h.referenceTimestamp.Sub(newreferenceTimestamp)) / float64(h.halfLife))
	h.histogram.scale(math.Ldexp(1., exponent)) // Scale all weights by 2^exponent.
	h.referenceTimestamp = newreferenceTimestamp
}

func (h *decayingHistogram) decayFactor(timestamp time.Time) float64 {
	// Max timestamp before the exponent grows too large.
	maxAllowedTimestamp := h.referenceTimestamp.Add(
		time.Duration(int64(h.halfLife) * int64(maxDecayExponent)))
	if timestamp.After(maxAllowedTimestamp) {
		// The exponent has grown too large. Renormalize the histogram by
		// shifting the referenceTimestamp to the current timestamp and rescaling
		// the weights accordingly.
		h.shiftReferenceTimestamp(timestamp)
	}
	return math.Exp2(float64(timestamp.Sub(h.referenceTimestamp)) / float64(h.halfLife))
}

func (h *decayingHistogram) SaveToChekpoint() (*vpa_types.HistogramCheckpoint, error) {
	checkpoint, err := h.histogram.SaveToChekpoint()
	if err != nil {
		return checkpoint, err
	}
	checkpoint.ReferenceTimestamp = metav1.NewTime(h.referenceTimestamp)
	return checkpoint, nil
}

func (h *decayingHistogram) LoadFromCheckpoint(checkpoint *vpa_types.HistogramCheckpoint) error {
	err := h.histogram.LoadFromCheckpoint(checkpoint)
	if err != nil {
		return err
	}
	h.referenceTimestamp = checkpoint.ReferenceTimestamp.Time
	return nil
}

func round(x float64) int {
	return int(math.Floor(x + 0.5))
}
