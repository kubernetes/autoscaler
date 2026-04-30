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

package metrics

import (
	"strconv"
	"time"

	"github.com/blang/semver/v4"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1beta1"
	cbclient "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/client"
	filters "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/filters"
	"k8s.io/autoscaler/cluster-autoscaler/utils/klogx"
	"k8s.io/component-base/metrics/legacyregistry"
	"k8s.io/klog/v2"
	"k8s.io/utils/clock"
)

const (
	metricFQName = "capacity_buffer_oldest_reconciliation_timestamp_seconds"
	loggingQuota = 5
)

var (
	oldestReconciliationTimestampMetricDesc = prometheus.NewDesc(
		prometheus.BuildFQName("cluster_autoscaler", "", metricFQName),
		"Unix timestamp of the oldest capacity buffer last reconciliation time.",
		[]string{"new_buffer"},
		nil,
	)
)

type capacityBufferClient interface {
	ListCapacityBuffers(namespace string) ([]*v1beta1.CapacityBuffer, error)
}

// reconciliationTimestampCollector is a prometheus.Collector that collects capacity buffer reconciliation interval metrics.
type reconciliationTimestampCollector struct {
	client                 capacityBufferClient
	reconciledBuffers      *ReconciliationCache
	supportedBuffersFilter filters.Filter
	clock                  clock.Clock
}

// NewReconciliationTimestampCollector creates a new collector instance.
func NewReconciliationTimestampCollector(client capacityBufferClient, strategies []string, reconciledBuffers *ReconciliationCache, clock clock.Clock) *reconciliationTimestampCollector {
	return &reconciliationTimestampCollector{
		client:                 client,
		reconciledBuffers:      reconciledBuffers,
		supportedBuffersFilter: filters.NewStrategyFilter(strategies),
		clock:                  clock,
	}
}

// RegisterReconciliationTimestampCollector registers the reconciliation timestamp collector.
func RegisterReconciliationTimestampCollector(client *cbclient.CapacityBufferClient, strategies []string, reconciledBuffers *ReconciliationCache, clock clock.Clock) {
	collector := NewReconciliationTimestampCollector(client, strategies, reconciledBuffers, clock)
	legacyregistry.MustRegister(collector)
}

// Describe implements the prometheus.Collector interface.
func (c *reconciliationTimestampCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- oldestReconciliationTimestampMetricDesc
}

// Collect implements the prometheus.Collector interface.
func (c *reconciliationTimestampCollector) Collect(ch chan<- prometheus.Metric) {
	// List all capacity buffers
	buffers, err := c.client.ListCapacityBuffers("")
	if err != nil {
		klog.Errorf("Failed to list capacity buffers with error: %v", err.Error())
		return
	}

	// Delete buffers that no longer exist from cache
	c.reconciledBuffers.Prune(buffers)

	supportedBuffers, _ := c.supportedBuffersFilter.Filter(buffers)
	reconciledBuffersSnapshot := c.reconciledBuffers.Snapshot()

	now := c.clock.Now()
	oldestNewBufferTimestamp := now
	oldestReconciledBufferTimestamp := now

	loggingQuota := klogx.NewLoggingQuota(loggingQuota)
	for _, buffer := range supportedBuffers {
		lastReconciliationTime, isNewBuffer, skipped := calculateBufferLastReconciliationTime(buffer, reconciledBuffersSnapshot)
		if skipped {
			klogx.V(2).UpTo(loggingQuota).Infof("Skipping capacity buffer %s/%s from %s metric: "+
				"buffer is not registered in reconciled buffers cache while having conditions, expected when CA restarts.", buffer.Namespace, buffer.Name, metricFQName)
			continue
		}

		if isNewBuffer {
			oldestNewBufferTimestamp = updateOldest(oldestNewBufferTimestamp, lastReconciliationTime)
		} else {
			oldestReconciledBufferTimestamp = updateOldest(oldestReconciledBufferTimestamp, lastReconciliationTime)
		}
	}
	klogx.V(2).Over(loggingQuota).Infof("Skipped %d other capacity buffers from %s metric "+
		"for not being registered in reconciled buffers cache while having conditions, expected when CA restarts.", -loggingQuota.Left(), metricFQName)

	c.reportMetric(ch, oldestNewBufferTimestamp, true)
	c.reportMetric(ch, oldestReconciledBufferTimestamp, false)
}

func updateOldest(oldest, current time.Time) time.Time {
	if current.Before(oldest) {
		return current
	}
	return oldest
}

func (c *reconciliationTimestampCollector) reportMetric(ch chan<- prometheus.Metric, timestamp time.Time, isNew bool) {
	ch <- prometheus.MustNewConstMetric(
		oldestReconciliationTimestampMetricDesc,
		prometheus.GaugeValue,
		float64(timestamp.Unix()),
		strconv.FormatBool(isNew),
	)
}

// Create implements the k8smetrics.Registerable interface.
func (c *reconciliationTimestampCollector) Create(version *semver.Version) bool {
	return true
}

// ClearState implements the k8smetrics.Registerable interface.
func (c *reconciliationTimestampCollector) ClearState() {
	// No-op for stateless collector
}

// FQName implements the k8smetrics.Registerable interface.
func (c *reconciliationTimestampCollector) FQName() string {
	return metricFQName
}

func calculateBufferLastReconciliationTime(buffer *v1beta1.CapacityBuffer, reconciledBuffersSnapshot map[types.UID]time.Time) (time.Time, bool, bool) {
	lastReconciliationTime, exists := reconciledBuffersSnapshot[buffer.UID]
	if exists {
		return lastReconciliationTime, false, false
	}

	if len(buffer.Status.Conditions) == 0 {
		// non-cached buffers with no conditions are considered new buffers
		return buffer.CreationTimestamp.Time, true, false
	}

	// non-cached buffers with conditions are skipped (CA restart)
	return time.Time{}, false, true
}
