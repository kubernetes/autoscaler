/*
Copyright The Kubernetes Authors.

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

package comparator

import (
	"math/rand/v2"
	"slices"
	"strconv"
	"strings"

	"k8s.io/klog/v2"
)

const (
	// loggedSampleSize is the max number of deltas to log in a single entry
	loggedSampleSize = 5
	// attributesCountEstimate is an estimate of the number of attributes per resource pool
	attributesCountEstimate = 20
	// bytesPerDeltaSummaryEstimate is an estimate of the number of bytes per delta summary
	bytesPerDeltaSummaryEstimate = 256
)

// logger is a function that logs messages, used for testing purposes.
type logger func(format string, args ...any)

// loggingSampler is aggregator that collects a representative subset of resource
// discrepancies. To avoid being spammy and to reduce performance and memory footprint
// of the component - sampler hard caps the amount of logged entry nodes and randomly
// picks ones from a dataset, to guarantee selection fairness - it uses reservoir sampling
type loggingSampler struct {
	// summaryBuilder is used to build the summary of the sampled deltas
	summaryBuilder strings.Builder
	// attributeBuffers is used to store the signatures of the sampled deltas
	attributeBuffers []string
	// logger is the logger to use for logging the sampled deltas
	logger logger
	// sampledNodes is the list of node names that have been sampled.
	sampledNodes []string
	// sampledDeltas is the list of resource deltas that have been sampled.
	sampledDeltas [][]resourceDelta
	// deltasProcessed is the number of deltas that have been processed by sampler.
	deltasProcessed uint64
}

// newLoggingSampler creates a new logging sampler.
func newLoggingSampler() loggingSampler {
	return newLoggingSamplerWithLogger(klog.Warningf)
}

// newLoggingSamplerWithLogger creates a new logging sampler with the given logger.
func newLoggingSamplerWithLogger(logger logger) loggingSampler {
	return loggingSampler{
		attributeBuffers: make([]string, 0, attributesCountEstimate),
		summaryBuilder:   strings.Builder{},
		logger:           logger,
		sampledNodes:     make([]string, 0, loggedSampleSize),
		sampledDeltas:    make([][]resourceDelta, 0, loggedSampleSize),
	}
}

// Sample samples the given resource deltas for the given node name.
// It uses reservoir sampling to select a random sample of deltas to log,
// ensuring that each delta has an equal probability of being selected.
func (s *loggingSampler) Sample(nodeName string, deltas []resourceDelta) {
	s.deltasProcessed++

	if len(s.sampledNodes) < loggedSampleSize {
		s.sampledNodes = append(s.sampledNodes, nodeName)
		s.sampledDeltas = append(s.sampledDeltas, slices.Clone(deltas))
		return
	}

	// Only swap if the random index hits our target range
	// to guarantee fairness of the sampling
	j := rand.Uint64N(s.deltasProcessed)
	if j < loggedSampleSize {
		s.sampledNodes[j] = nodeName
		s.sampledDeltas[j] = slices.Clone(deltas)
	}
}

// Reset resets sampler internal buffers and trackers to initial values without
// dropping already allocated capacity
func (s *loggingSampler) Reset() {
	s.deltasProcessed = 0
	s.summaryBuilder.Reset()
	s.attributeBuffers = s.attributeBuffers[:0]
	s.sampledNodes = s.sampledNodes[:0]
	s.sampledDeltas = s.sampledDeltas[:0]
}

// LogSampled flushes the collected samples to the logger without
// clearing internal buffers.
func (s *loggingSampler) LogSampled() {
	if s.deltasProcessed == 0 {
		return
	}

	deltasInSample := 0
	for i := range s.sampledDeltas {
		deltasInSample += len(s.sampledDeltas[i])
	}

	s.summaryBuilder.Grow(bytesPerDeltaSummaryEstimate * deltasInSample)
	for i, nodeName := range s.sampledNodes {
		s.summaryBuilder.WriteString("- ")
		s.summaryBuilder.WriteString(nodeName)
		s.summaryBuilder.WriteString(": ")

		deltas := s.sampledDeltas[i]
		for j := range deltas {
			if j > 0 {
				s.summaryBuilder.WriteString(", ")
			}
			s.writeDeltaSummary(&deltas[j])
		}

		if i < len(s.sampledNodes)-1 {
			s.summaryBuilder.WriteString("\n")
		}
	}

	if len(s.sampledNodes) > 0 {
		s.logger("DRA Resource Discrepancies detected between node templates and actual nodes:\n%s", s.summaryBuilder.String())
	}
}

// writeDeltaSummary writes the resource delta representation to the internal string builder.
func (s *loggingSampler) writeDeltaSummary(d *resourceDelta) {
	builder := &s.summaryBuilder

	builder.WriteString(`ResourceDelta{Type="`)
	builder.WriteString(d.Type().String())

	builder.WriteString(`", Driver="`)
	builder.WriteString(d.Driver)

	builder.WriteString(`", TemplatePool="`)
	builder.WriteString(d.TemplateResourcePool)

	builder.WriteString(`", NodePool="`)
	builder.WriteString(d.NodeResourcePool)

	builder.WriteString(`", DeviceCountDelta="`)
	builder.WriteString(strconv.FormatInt(d.DeviceCountDelta, 10))

	builder.WriteString(`", TemplateSignature="`)
	s.writeAttributes(d.TemplateSignatureMap)

	builder.WriteString(`", NodeSignature="`)
	s.writeAttributes(d.NodeSignatureMap)

	builder.WriteString(`", MissingAttributes="`)
	s.writeAttributesDifference(d.TemplateSignatureMap, d.NodeSignatureMap)

	builder.WriteString(`", ExtraAttributes="`)
	s.writeAttributesDifference(d.NodeSignatureMap, d.TemplateSignatureMap)

	builder.WriteString(`"}`)
}

// writeAttributes writes given attribute keys to the internal string builder
func (s *loggingSampler) writeAttributes(m attributesMap) {
	s.attributeBuffers = s.attributeBuffers[:0]
	for k := range m {
		s.attributeBuffers = append(s.attributeBuffers, string(k))
	}
	slices.Sort(s.attributeBuffers)

	builder := &s.summaryBuilder
	for i, k := range s.attributeBuffers {
		if i > 0 {
			builder.WriteString(";")
		}
		builder.WriteString(k)
	}
}

// writeAttributesDifference writes given a difference of given attribute keys to the internal string builder
func (s *loggingSampler) writeAttributesDifference(base, subtract attributesMap) {
	s.attributeBuffers = s.attributeBuffers[:0]
	for k := range base {
		if _, ok := subtract[k]; !ok {
			s.attributeBuffers = append(s.attributeBuffers, string(k))
		}
	}
	slices.Sort(s.attributeBuffers)

	builder := &s.summaryBuilder
	for i, k := range s.attributeBuffers {
		if i > 0 {
			builder.WriteString(";")
		}
		builder.WriteString(k)
	}
}
