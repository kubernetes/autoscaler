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
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"

	resourceapi "k8s.io/api/resource/v1"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/klog/v2"
)

const (
	maxLoggedSampleSize          = 5
	countOfDiscrepanciesEstimate = 10
)

// NodeComparisonData encapsulates the resource slices for a node and its template.
type NodeComparisonData struct {
	NodeName       string
	TemplateSlices []*resourceapi.ResourceSlice
	NodeSlices     []*resourceapi.ResourceSlice
}

// metricsEmitter interface is used to emit gauge metrics for DRA discrepancies.
type metricsEmitter interface {
	SetNodeTemplateResourcesMismatch(driver string, mismatchType metrics.ResourceMismatchType, value uint32)
}

// logger is a function that logs messages, used for testing purposes.
type logger func(format string, args ...any)

// NodeResourcesComparator detects discrepancies between expected and actual resource topologies.
type NodeResourcesComparator struct {
	metrics metricsEmitter
	logger  logger
}

// NewNodeResourcesComparator returns a new stateful NodeResourcesComparator.
func NewNodeResourcesComparator(metric metricsEmitter) *NodeResourcesComparator {
	return &NodeResourcesComparator{
		metrics: metric,
		logger:  klog.Warningf,
	}
}

type driverDiscrepancy struct {
	missing  uint32
	extra    uint32
	mismatch uint32
	unknown  uint32
}

func (c *NodeResourcesComparator) logSampledDiscrepancies(sampledNodes []string, sampledDeltas [][]resourceDelta) {
	if len(sampledNodes) == 0 {
		return
	}

	nodeReports := make([]string, len(sampledNodes))
	for i, nodeName := range sampledNodes {
		deltas := sampledDeltas[i]
		deltaSummaries := make([]string, len(deltas))
		for j, delta := range deltas {
			deltaSummaries[j] = delta.Summary()
		}
		nodeReports[i] = fmt.Sprintf("- %s: %s", nodeName, strings.Join(deltaSummaries, ", "))
	}

	c.logger("DRA Resource Discrepancies detected between node templates and actual nodes:\n%s", strings.Join(nodeReports, "\n"))
}

func (c *NodeResourcesComparator) emitMetrics(aggregatedDiscrepancies map[string]driverDiscrepancy) {
	for driver, disc := range aggregatedDiscrepancies {
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


// ReportResourceDiscrepancies compares DRA resources for a batch of nodes,
// aggregates discrepancies across all drivers, emits metrics, and logs a summary report.
func (c *NodeResourcesComparator) ReportResourceDiscrepancies(data []NodeComparisonData) {
	if len(data) == 0 {
		return
	}

	aggregatedDiscrepancies := make(map[string]driverDiscrepancy, countOfDriversEstimate)
	sampledNodes := make([]string, 0, maxLoggedSampleSize)
	sampledDeltas := make([][]resourceDelta, 0, maxLoggedSampleSize)
	deltas := make([]resourceDelta, 0, countOfDiscrepanciesEstimate)
	discrepantNodesCount := 0
	for nodeIndex := range data {
		nodeData := &data[nodeIndex]
		// No slices to compare, delta is missing
		if len(nodeData.TemplateSlices) == 0 && len(nodeData.NodeSlices) == 0 {
			continue
		}

		deltas = compareDraResources(nodeData.TemplateSlices, nodeData.NodeSlices, deltas)
		if len(deltas) == 0 {
			continue
		}

		for _, delta := range deltas {
			disc := aggregatedDiscrepancies[delta.Driver]
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
			aggregatedDiscrepancies[delta.Driver] = disc
		}

		discrepantNodesCount++
		if len(sampledNodes) < maxLoggedSampleSize {
			sampledNodes = append(sampledNodes, nodeData.NodeName)
			// Sampled deltas need to be cloned because the deltas slice is reused
			sampledDeltas = append(sampledDeltas, slices.Clone(deltas))
		} else {
			// Only swap if the random index hits our target range
			// to guarantee fairness of the sampling
			j := rand.IntN(discrepantNodesCount)
			if j < maxLoggedSampleSize {
				sampledNodes[j] = nodeData.NodeName
				// Sampled deltas need to be cloned because the deltas slice is reused
				sampledDeltas[j] = slices.Clone(deltas)
			}
		}

		// Reset the buffer for the next iteration
		deltas = deltas[:0]
	}

	if len(aggregatedDiscrepancies) == 0 {
		return
	}

	c.emitMetrics(aggregatedDiscrepancies)
	c.logSampledDiscrepancies(sampledNodes, sampledDeltas)
}
