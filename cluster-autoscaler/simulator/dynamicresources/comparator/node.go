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
	resourceapi "k8s.io/api/resource/v1"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/klog/v2"
)

const (
	// countOfDiscrepanciesEstimate is an estimate of the number of discrepancies for
	// a single report cycle
	countOfDiscrepanciesEstimate = 10
)

// metricsEmitter interface is used to emit gauge metrics for DRA discrepancies.
type metricsEmitter interface {
	SetNodeTemplateResourcesMismatch(driver string, mismatchType metrics.ResourceMismatchType, value uint32)
}

// NodeResourcesComparator detects discrepancies between expected and actual resource topologies.
// For more details on the algorithm - see resourcePoolComparator.CompareResourcePools.
type NodeResourcesComparator struct {
	// Reusable map to store intermediate results while comparing nodes
	discrepanciesPerDriver map[string]driverDiscrepancy
	// Reusable buffer to store intermediate per node resource deltas
	// while comparing nodes
	deltas []resourceDelta

	metrics    metricsEmitter
	comparator resourcePoolComparator
	sampler    loggingSampler
}

// NewNodeResourcesComparator returns a new stateful NodeResourcesComparator.
func NewNodeResourcesComparator(metric metricsEmitter) *NodeResourcesComparator {
	return &NodeResourcesComparator{
		metrics:                metric,
		comparator:             newResourcePoolComparator(),
		sampler:                newLoggingSampler(),
		deltas:                 make([]resourceDelta, 0, countOfDiscrepanciesEstimate),
		discrepanciesPerDriver: make(map[string]driverDiscrepancy, countOfDriversEstimate),
	}
}

// newNodeResourcesComparatorWithLogger returns a new NodeResourcesComparator with a custom logger.
func newNodeResourcesComparatorWithLogger(metric metricsEmitter, logger logger) *NodeResourcesComparator {
	return &NodeResourcesComparator{
		metrics:                metric,
		comparator:             newResourcePoolComparator(),
		sampler:                newLoggingSamplerWithLogger(logger),
		deltas:                 make([]resourceDelta, 0, countOfDiscrepanciesEstimate),
		discrepanciesPerDriver: make(map[string]driverDiscrepancy, countOfDriversEstimate),
	}
}

// driverDiscrepancy holds aggregate discrepancies for a single driver.
type driverDiscrepancy struct {
	missing  uint32
	extra    uint32
	mismatch uint32
	unknown  uint32
}

// emitMetrics emits the aggregated discrepancies to the metrics emitter.
func (c *NodeResourcesComparator) emitMetrics() {
	for driver, disc := range c.discrepanciesPerDriver {
		if disc.missing > 0 {
			c.metrics.SetNodeTemplateResourcesMismatch(driver, metrics.ResourceMismatchTypeMissing, disc.missing)
		}
		if disc.extra > 0 {
			c.metrics.SetNodeTemplateResourcesMismatch(driver, metrics.ResourceMismatchTypeExtra, disc.extra)
		}
		if disc.mismatch > 0 {
			c.metrics.SetNodeTemplateResourcesMismatch(driver, metrics.ResourceMismatchTypeMismatch, disc.mismatch)
		}
		if disc.unknown > 0 {
			c.metrics.SetNodeTemplateResourcesMismatch(driver, metrics.ResourceMismatchTypeUnknown, disc.unknown)
		}
	}
}

// reset resets the comparator to its initial state making it ready for the next batch of nodes.
func (c *NodeResourcesComparator) reset() {
	c.sampler.Reset()
	c.deltas = c.deltas[:0]
	clear(c.discrepanciesPerDriver)
}

// ReportResourceDiscrepancies compares DRA resources for a batch of nodes,
// aggregates discrepancies across all drivers, emits metrics, and logs a summary report.
//
// Function assumes that nodeNames, templateSlices, and nodeSlices have the same length,
// and aborts execution if they don't.
func (c *NodeResourcesComparator) ReportResourceDiscrepancies(
	nodeNames []string,
	templateSlices [][]*resourceapi.ResourceSlice,
	nodeSlices [][]*resourceapi.ResourceSlice,
) {
	c.reset()

	if len(nodeNames) == 0 {
		return
	}

	if len(nodeNames) != len(templateSlices) || len(nodeNames) != len(nodeSlices) {
		klog.Errorf("NodeResourcesComparator: nodeNames, templateSlices, and nodeSlices must have the same length")
		return
	}

	for nodeIndex := range nodeNames {
		// No slices to compare, delta is missing
		if len(templateSlices[nodeIndex]) == 0 && len(nodeSlices[nodeIndex]) == 0 {
			continue
		}

		c.deltas = c.comparator.CompareResourcePools(templateSlices[nodeIndex], nodeSlices[nodeIndex], c.deltas)
		if len(c.deltas) == 0 {
			continue
		}

		c.updateDiscrepanciesPerDriver()
		c.sampler.Sample(nodeNames[nodeIndex], c.deltas)

		// Reset the buffer for the next iteration
		c.deltas = c.deltas[:0]
	}

	if len(c.discrepanciesPerDriver) == 0 {
		return
	}

	c.emitMetrics()
	c.sampler.LogSampled()
}

// updateDiscrepanciesPerDriver iterates through comparator found deltas and increases
// disrepancy counters accordingly to deltas found during the last comparison.
func (c *NodeResourcesComparator) updateDiscrepanciesPerDriver() {
	for _, delta := range c.deltas {
		disc := c.discrepanciesPerDriver[delta.Driver]
		switch delta.Type() {
		case resourceDeltaTypeMissing:
			disc.missing++
		case resourceDeltaTypeExtra:
			disc.extra++
		case resourceDeltaTypeMismatch:
			disc.mismatch++
		case resourceDeltaTypeUnknown:
			disc.unknown++
		}
		c.discrepanciesPerDriver[delta.Driver] = disc
	}
}
